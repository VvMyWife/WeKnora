package router

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tencent/WeKnora/internal/processingcontrol"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func TestKnowledgePauseMiddlewareDoesNotEnterHandler(t *testing.T) {
	mini := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer rdb.Close()
	c := processingcontrol.New(rdb)
	ctx := context.Background()
	if _, err := c.SetPaused(ctx, 7, "paused-kb", true); err != nil {
		t.Fatal(err)
	}
	called := 0
	handler := knowledgePauseMiddleware(c, nil, nil, nil)(asynq.HandlerFunc(func(context.Context, *asynq.Task) error { called++; return nil }))
	task := asynq.NewTask(types.TypeWikiIngest, []byte(`{"tenant_id":7,"kb_id":"paused-kb"}`))
	if err := handler.ProcessTask(ctx, task); !errors.Is(err, processingcontrol.ErrPaused) {
		t.Fatalf("pause error: %v", err)
	}
	if called != 0 {
		t.Fatal("paused task entered handler, risking failed spans/counter changes")
	}
	if asynqTaskIsFailure(processingcontrol.ErrPaused) {
		t.Fatal("pause consumes retry budget")
	}
	if asynqRetryDelayFunc(25, processingcontrol.ErrPaused, task) != processingcontrol.RetryDelay {
		t.Fatal("pause uses failure backoff")
	}
	other := asynq.NewTask(types.TypeWikiIngest, []byte(`{"tenant_id":7,"kb_id":"short-kb"}`))
	if err := handler.ProcessTask(ctx, other); err != nil || called != 1 {
		t.Fatalf("other KB blocked: called=%d err=%v", called, err)
	}
	if _, err := c.SetPaused(ctx, 7, "paused-kb", false); err != nil {
		t.Fatal(err)
	}
	if err := handler.ProcessTask(ctx, task); err != nil || called != 2 {
		t.Fatalf("resume: called=%d err=%v", called, err)
	}
}

func TestPausedFinalAttemptStaysQueuedAndCanResume(t *testing.T) {
	mini := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer rdb.Close()
	client := asynq.NewClientFromRedisClient(rdb)
	inspector := asynq.NewInspectorFromRedisClient(rdb)
	c := processingcontrol.New(rdb)
	ctx := context.Background()
	if _, err := c.SetPaused(ctx, 7, "paused-kb", true); err != nil {
		t.Fatal(err)
	}
	completed := make(chan struct{}, 1)
	handler := knowledgePauseMiddleware(c, nil, client, inspector)(asynq.HandlerFunc(func(context.Context, *asynq.Task) error { completed <- struct{}{}; return nil }))
	worker := asynq.NewServerFromRedisClient(rdb, asynq.Config{
		Concurrency: 1, Queues: map[string]int{types.QueueWiki: 1},
		TaskCheckInterval: 10 * time.Millisecond, DelayedTaskCheckInterval: time.Hour,
		ShutdownTimeout: time.Second, LogLevel: asynq.FatalLevel,
		RetryDelayFunc: asynqRetryDelayFunc, IsFailure: asynqTaskIsFailure,
	})
	if err := worker.Start(handler); err != nil {
		t.Fatal(err)
	}
	defer worker.Shutdown()
	task := asynq.NewTask(types.TypeWikiIngest, []byte(`{"tenant_id":7,"kb_id":"paused-kb"}`))
	if _, err := client.Enqueue(task, asynq.Queue(types.QueueWiki), asynq.MaxRetry(0)); err != nil {
		t.Fatal(err)
	}
	var parked []*asynq.TaskInfo
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		parked, _ = inspector.ListScheduledTasks(types.QueueWiki)
		if len(parked) == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(parked) != 1 || parked[0].MaxRetry != 0 {
		t.Fatalf("final attempt not parked: %+v", parked)
	}
	archived, err := inspector.ListArchivedTasks(types.QueueWiki)
	if err != nil || len(archived) != 0 {
		t.Fatalf("pause archived a task: count=%d err=%v", len(archived), err)
	}
	select {
	case <-completed:
		t.Fatal("paused handler ran")
	default:
	}
	if _, err := c.SetPaused(ctx, 7, "paused-kb", false); err != nil {
		t.Fatal(err)
	}
	if err := inspector.RunTask(types.QueueWiki, parked[0].ID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-completed:
	case <-time.After(5 * time.Second):
		t.Fatal("resumed task did not execute")
	}
}

func TestPauseDoesNotBlockDeletionOrInteractiveWork(t *testing.T) {
	for _, kind := range []string{types.TypeKBDelete, types.TypeKnowledgeListDelete, types.TypeTemporaryDocumentProcess, types.TypeMemoryExtract} {
		if pausableKnowledgeTask(kind) {
			t.Fatalf("unrelated task is pausable: %s", kind)
		}
	}
	if !asynqTaskIsFailure(errors.New("real graph failure")) {
		t.Fatal("real failures must retain retry accounting")
	}
}
