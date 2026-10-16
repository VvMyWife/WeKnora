package limiter

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

const GraphConcurrency = 8

// AcquireGraph uses a separate key and never falls back to the ordinary fail-open governor.
func AcquireGraph(ctx context.Context, modelID, name string) (context.Context, func(), error) {
	governorMu.RLock()
	l := governor
	governorMu.RUnlock()
	if modelID == "" {
		return ctx, noop, errors.New("graph limiter: missing model id")
	}
	key := "graph:" + modelID
	switch g := l.(type) {
	case *redisLimiter:
		if g == nil || g.rdb == nil {
			return ctx, noop, errors.New("graph limiter: redis unavailable")
		}
		g.SetModelName(key, name+" (graph)")
		return g.acquireGraph(ctx, key)
	case *localLimiter:
		if err := ctx.Err(); err != nil {
			return ctx, noop, err
		}
		release, err := g.Acquire(ctx, key, GraphConcurrency)
		if err == nil {
			err = ctx.Err()
		}
		if err != nil {
			release()
			return ctx, noop, err
		}
		return ctx, release, nil
	default:
		return ctx, noop, errors.New("graph limiter: no strict limiter installed")
	}
}

var renewGraphLease = redis.NewScript(`
if redis.call('ZSCORE', KEYS[1], ARGV[1]) == false then return 0 end
redis.call('ZADD', KEYS[1], 'XX', ARGV[2], ARGV[1])
redis.call('PEXPIRE', KEYS[1], ARGV[3])
return 1`)

func (l *redisLimiter) acquireGraph(ctx context.Context, key string) (context.Context, func(), error) {
	entry, _ := l.tracked.LoadOrStore(key, &trackedSemaphore{})
	tracked := entry.(*trackedSemaphore)
	tracked.limit.Store(GraphConcurrency)
	tracked.waiting.Add(1)
	defer tracked.waiting.Add(-1)
	token := uuid.NewString()
	zkey := keyPrefix + key
	timer := time.NewTicker(l.pollInterval)
	defer timer.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return ctx, noop, err
		}
		n, err := acquireScript.Run(ctx, l.rdb, []string{zkey}, time.Now().UnixMilli(), GraphConcurrency, token, l.ttl.Milliseconds()).Int()
		if err != nil {
			return ctx, noop, fmt.Errorf("graph limiter acquire: %w", err)
		}
		if n == 1 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx, noop, ctx.Err()
		case <-timer.C:
		}
	}
	callCtx, cancel := context.WithCancel(ctx)
	stop := make(chan struct{})
	var once sync.Once
	go func() {
		tick := time.NewTicker(l.ttl / 3)
		defer tick.Stop()
		for {
			select {
			case <-stop:
				return
			case <-callCtx.Done():
				return
			case <-tick.C:
				renewCtx, done := context.WithTimeout(callCtx, l.ttl/6)
				n, err := renewGraphLease.Run(renewCtx, l.rdb, []string{zkey}, token, time.Now().Add(l.ttl).UnixMilli(), (2 * l.ttl).Milliseconds()).Int()
				done()
				if err != nil || n != 1 {
					cancel()
					return
				}
			}
		}
	}()
	release := func() {
		once.Do(func() {
			close(stop)
			cancel()
			c, done := context.WithTimeout(context.Background(), 5*time.Second)
			defer done()
			_ = l.rdb.ZRem(c, zkey, token).Err()
		})
	}
	return callCtx, release, nil
}
