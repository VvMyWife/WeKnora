// Package processingcontrol implements resumable, knowledge-base-scoped worker admission.
package processingcontrol

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrPaused = errors.New("knowledge base processing is paused")

// RetryDelay keeps parked work out of worker slots without consuming retries.
const RetryDelay = 30 * time.Second

type State struct {
	Supported       bool   `json:"supported"`
	RequestedPaused bool   `json:"requested_paused"`
	State           string `json:"state"`
	ActiveTasks     int64  `json:"active_tasks"`
}

type Controller struct {
	rdb *redis.Client
	ttl time.Duration
}

var installed atomic.Pointer[Controller]

func Install(c *Controller)             { installed.Store(c) }
func Current() *Controller              { return installed.Load() }
func New(rdb *redis.Client) *Controller { return &Controller{rdb: rdb, ttl: 45 * time.Second} }

func (c *Controller) Supported() bool { return c != nil && c.rdb != nil }

func keys(tenant uint64, kb string) []string {
	base := fmt.Sprintf("weknora:processing:{%d:%s}:", tenant, kb)
	return []string{base + "paused", base + "active"}
}

var admit = redis.NewScript(`
redis.call('ZREMRANGEBYSCORE', KEYS[2], '-inf', ARGV[1])
if redis.call('EXISTS', KEYS[1]) == 1 then return 0 end
redis.call('ZADD', KEYS[2], ARGV[1]+ARGV[3], ARGV[2])
redis.call('PEXPIRE', KEYS[2], ARGV[3]*2)
return 1
`)

var stateScript = redis.NewScript(`
if ARGV[2] == 'pause' then redis.call('SET', KEYS[1], '1') end
if ARGV[2] == 'resume' then redis.call('DEL', KEYS[1]) end
redis.call('ZREMRANGEBYSCORE', KEYS[2], '-inf', ARGV[1])
return {redis.call('EXISTS', KEYS[1]), redis.call('ZCARD', KEYS[2])}
`)

func (c *Controller) State(ctx context.Context, tenant uint64, kb string) (State, error) {
	return c.state(ctx, tenant, kb, "read")
}

func (c *Controller) SetPaused(ctx context.Context, tenant uint64, kb string, paused bool) (State, error) {
	action := "resume"
	if paused {
		action = "pause"
	}
	return c.state(ctx, tenant, kb, action)
}

func (c *Controller) state(ctx context.Context, tenant uint64, kb, action string) (State, error) {
	if !c.Supported() {
		return State{State: "unsupported"}, nil
	}
	if tenant == 0 || kb == "" {
		return State{}, errors.New("processing control requires tenant and knowledge base")
	}
	values, err := stateScript.Run(ctx, c.rdb, keys(tenant, kb), time.Now().UnixMilli(), action).Int64Slice()
	if err != nil {
		return State{}, err
	}
	s := State{Supported: true, RequestedPaused: values[0] == 1, ActiveTasks: values[1], State: "running"}
	if s.RequestedPaused {
		s.State = "paused"
		if s.ActiveTasks > 0 {
			s.State = "pausing"
		}
	}
	return s, nil
}

// Acquire admits a whole task atomically with the pause flag. Existing tasks drain
// naturally; no cancellation, fake success, failure span, or counter decrement is synthesized.
func (c *Controller) Acquire(ctx context.Context, tenant uint64, kb string) (func(), error) {
	if !c.Supported() {
		return func() {}, nil
	}
	if tenant == 0 || kb == "" {
		return nil, errors.New("processing control requires tenant and knowledge base")
	}
	ks := keys(tenant, kb)
	token := uuid.NewString()
	allowed, err := admit.Run(ctx, c.rdb, ks, time.Now().UnixMilli(), token, c.ttl.Milliseconds()).Int()
	if err != nil {
		return nil, err
	}
	if allowed == 0 {
		return nil, ErrPaused
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(c.ttl / 3)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				bg, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				pipe := c.rdb.TxPipeline()
				pipe.ZAdd(bg, ks[1], redis.Z{Score: float64(time.Now().Add(c.ttl).UnixMilli()), Member: token})
				pipe.PExpire(bg, ks[1], c.ttl*2)
				_, _ = pipe.Exec(bg)
				cancel()
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(stop)
			<-done
			bg, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = c.rdb.ZRem(bg, ks[1], token).Err()
		})
	}, nil
}
