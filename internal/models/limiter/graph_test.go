package limiter

import (
	"context"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/redis/go-redis/v9"
	"os"
	"testing"
	"time"
)

func TestGraphEightAcrossInstancesAndDefaultTwo(t *testing.T) {
	l, _, rdb := newTestLimiter(t, time.Minute, 5*time.Millisecond)
	other := NewRedisLimiter(rdb).(*redisLimiter)
	SetGovernor(l, 2)
	t.Cleanup(func() { SetGovernor(nil, 0) })
	ctx := types.WithGraphTask(context.Background())
	releases := []func(){}
	for i := 0; i < 8; i++ {
		var r func()
		var err error
		if i%2 == 0 {
			_, r, err = AcquireGraph(ctx, "m", "model")
		} else {
			_, r, err = other.acquireGraph(ctx, "graph:m")
		}
		if err != nil {
			t.Fatal(err)
		}
		releases = append(releases, r)
		defer r()
	}
	if n := zcard(t, rdb, "graph:m"); n != 8 {
		t.Fatalf("active=%d", n)
	}
	waitCtx, cancel := context.WithTimeout(ctx, 60*time.Millisecond)
	defer cancel()
	if _, r, err := other.acquireGraph(waitCtx, "graph:m"); err == nil {
		r()
		t.Fatal("ninth admitted")
	}
	releases[0]()
	nextCtx, done := context.WithTimeout(ctx, time.Second)
	defer done()
	_, r, err := other.acquireGraph(nextCtx, "graph:m")
	if err != nil {
		t.Fatal(err)
	}
	r()
}
func TestGraphFailsClosed(t *testing.T) {
	l, server, _ := newTestLimiter(t, time.Minute, time.Millisecond)
	SetGovernor(l, 2)
	defer SetGovernor(nil, 0)
	server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, r, err := AcquireGraph(ctx, "m", "model"); err == nil {
		r()
		t.Fatal("redis failure admitted")
	}
}
func TestGraphLeaseLossCancelsProvider(t *testing.T) {
	l, _, rdb := newTestLimiter(t, 300*time.Millisecond, time.Millisecond)
	SetGovernor(l, 2)
	defer SetGovernor(nil, 0)
	ctx, r, err := AcquireGraph(context.Background(), "m", "model")
	if err != nil {
		t.Fatal(err)
	}
	defer r()
	rdb.Del(context.Background(), keyPrefix+"graph:m")
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("provider was not cancelled after lease loss")
	}
}

func TestGraphRealRedisEightAcrossInstances(t *testing.T) {
	addr := os.Getenv("WEKNORA_ENRICHMENT_TEST_REDIS")
	if addr == "" {
		t.Skip("isolated real Redis required")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()
	first := NewRedisLimiter(client).(*redisLimiter)
	second := NewRedisLimiter(client).(*redisLimiter)
	key := "graph:integration-eight"
	client.Del(context.Background(), keyPrefix+key)
	SetGovernor(first, 2)
	defer SetGovernor(nil, 0)
	releases := []func(){}
	for i := 0; i < 8; i++ {
		var r func()
		var err error
		if i%2 == 0 {
			_, r, err = AcquireGraph(context.Background(), "integration-eight", "integration")
		} else {
			_, r, err = second.acquireGraph(context.Background(), key)
		}
		if err != nil {
			t.Fatal(err)
		}
		releases = append(releases, r)
		defer r()
	}
	if got := zcard(t, client, key); got != 8 {
		t.Fatalf("real Redis holders=%d", got)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, r, err := second.acquireGraph(ctx, key); err == nil {
		r()
		t.Fatal("real Redis admitted ninth")
	}
	releases[0]()
	ctx2, done := context.WithTimeout(context.Background(), time.Second)
	defer done()
	_, r, err := second.acquireGraph(ctx2, key)
	if err != nil {
		t.Fatal(err)
	}
	r()
}
