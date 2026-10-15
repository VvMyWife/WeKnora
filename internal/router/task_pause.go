package router

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Tencent/WeKnora/internal/processingcontrol"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/hibiken/asynq"
)

func asynqTaskIsFailure(err error) bool { return !errors.Is(err, processingcontrol.ErrPaused) }

// Pause is checked outside the failure/dead-letter middleware, before any
// document handler can open a span or touch outstanding-subtask counters.
func knowledgePauseMiddleware(control *processingcontrol.Controller, knowledge interfaces.KnowledgeService, enqueuer interfaces.TaskEnqueuer, inspector *asynq.Inspector) asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			if !control.Supported() || !pausableKnowledgeTask(task.Type()) {
				return next.ProcessTask(ctx, task)
			}
			var p runtimeTaskPayloadProbe
			if err := json.Unmarshal(task.Payload(), &p); err != nil {
				return next.ProcessTask(ctx, task)
			}
			kbID := p.KnowledgeBaseID
			if kbID == "" {
				kbID = p.KBID
			}
			if p.KnowledgeID != "" && knowledge != nil {
				k, err := knowledge.GetKnowledgeByIDOnly(ctx, p.KnowledgeID)
				if err != nil {
					return err
				}
				if k != nil {
					kbID = k.KnowledgeBaseID
					p.TenantID = k.TenantID
				}
			}
			// Legacy/malformed tasks without an owner still go through their usual
			// handler, which is responsible for validation and cleanup.
			if p.TenantID == 0 || kbID == "" {
				return next.ProcessTask(ctx, task)
			}
			release, err := control.Acquire(ctx, p.TenantID, kbID)
			if errors.Is(err, processingcontrol.ErrPaused) {
				return parkPausedTask(ctx, task, enqueuer, inspector)
			}
			if err != nil {
				return err
			}
			defer release()
			return next.ProcessTask(ctx, task)
		})
	}
}

// Asynq checks exhaustion before IsFailure. On an already-final attempt, create
// a replacement with zero remaining failure retries, then revoke only the old
// queue record. The business payload/attempt and timeout policy stay unchanged.
func parkPausedTask(ctx context.Context, task *asynq.Task, enqueuer interfaces.TaskEnqueuer, inspector *asynq.Inspector) error {
	retried, hasRetry := asynq.GetRetryCount(ctx)
	maximum, hasMaximum := asynq.GetMaxRetry(ctx)
	if !hasRetry || !hasMaximum || retried < maximum {
		return processingcontrol.ErrPaused
	}
	queue, _ := asynq.GetQueueName(ctx)
	id, _ := asynq.GetTaskID(ctx)
	if enqueuer == nil || inspector == nil || queue == "" || id == "" {
		return processingcontrol.ErrPaused
	}
	info, err := inspector.GetTaskInfo(queue, id)
	if err != nil {
		return err
	}
	opts := []asynq.Option{asynq.Queue(queue), asynq.MaxRetry(0), asynq.ProcessIn(processingcontrol.RetryDelay)}
	if info.Timeout > 0 {
		opts = append(opts, asynq.Timeout(info.Timeout))
	}
	if !info.Deadline.IsZero() {
		opts = append(opts, asynq.Deadline(info.Deadline))
	}
	if _, err := enqueuer.Enqueue(task, opts...); err != nil {
		return err
	}
	return asynq.RevokeTask
}

func pausableKnowledgeTask(taskType string) bool {
	switch taskType {
	case types.TypeDocumentProcess, types.TypeManualProcess, types.TypeImageMultimodal,
		types.TypeKnowledgePostProcess, types.TypeQuestionGeneration, types.TypeSummaryGeneration,
		types.TypeChunkExtract, types.TypeDataTableSummary, types.TypeKnowledgeAutoTag,
		types.TypeWikiIngest, types.TypeWikiFinalize:
		return true
	default:
		return false
	}
}
