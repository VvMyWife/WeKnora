package processingcontrol

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func testController(t *testing.T) *Controller {
	t.Helper()
	mini := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return New(rdb)
}

func TestPauseDrainsAndResumePreservesWork(t *testing.T) {
	c := testController(t)
	ctx := context.Background()
	release, err := c.Acquire(ctx, 1, "kb-a")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(release)
	s, err := c.SetPaused(ctx, 1, "kb-a", true)
	if err != nil || s.State != "pausing" || s.ActiveTasks != 1 {
		t.Fatalf("state=%+v err=%v", s, err)
	}
	if _, err := c.Acquire(ctx, 1, "kb-a"); !errors.Is(err, ErrPaused) {
		t.Fatalf("paused admission: %v", err)
	}
	release()
	release() // release is idempotent
	s, err = c.State(ctx, 1, "kb-a")
	if err != nil || s.State != "paused" || s.ActiveTasks != 0 {
		t.Fatalf("state=%+v err=%v", s, err)
	}
	if _, err := c.SetPaused(ctx, 1, "kb-a", false); err != nil {
		t.Fatal(err)
	}
	release, err = c.Acquire(ctx, 1, "kb-a")
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestPauseIsScopedAndSurvivesControllerRestart(t *testing.T) {
	c := testController(t)
	ctx := context.Background()
	if _, err := c.SetPaused(ctx, 1, "kb-a", true); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []struct {
		tenant uint64
		kb     string
	}{{1, "kb-b"}, {2, "kb-a"}} {
		release, err := c.Acquire(ctx, scope.tenant, scope.kb)
		if err != nil {
			t.Fatal(err)
		}
		release()
	}
	restarted := New(c.rdb)
	s, err := restarted.State(ctx, 1, "kb-a")
	if err != nil || s.State != "paused" {
		t.Fatalf("restart state=%+v err=%v", s, err)
	}
}

func TestDeadWorkerLeaseDoesNotBlockPausedState(t *testing.T) {
	c := testController(t)
	ctx := context.Background()
	if _, err := c.SetPaused(ctx, 1, "kb-a", true); err != nil {
		t.Fatal(err)
	}
	if err := c.rdb.ZAdd(ctx, keys(1, "kb-a")[1], redis.Z{Score: float64(time.Now().Add(-time.Second).UnixMilli()), Member: "dead-worker"}).Err(); err != nil {
		t.Fatal(err)
	}
	s, err := c.State(ctx, 1, "kb-a")
	if err != nil || s.State != "paused" || s.ActiveTasks != 0 {
		t.Fatalf("state=%+v err=%v", s, err)
	}
}

func TestUnsupportedBackendDoesNotPretendToPause(t *testing.T) {
	s, err := New(nil).SetPaused(context.Background(), 1, "kb", true)
	if err != nil || s.Supported || s.State != "unsupported" {
		t.Fatalf("state=%+v err=%v", s, err)
	}
}
