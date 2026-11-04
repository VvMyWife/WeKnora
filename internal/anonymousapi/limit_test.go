package anonymousapi

import (
	"context"
	"errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"os"
	"testing"
	"time"
)

func testLimits(t *testing.T, addr string) {
	t.Helper()
	ctx := context.Background()
	a := redis.NewClient(&redis.Options{Addr: addr})
	b := redis.NewClient(&redis.Options{Addr: addr})
	defer a.Close()
	defer b.Close()
	prefix := "weknora:{anonymous-api}:"
	keys, e := a.Keys(ctx, prefix+"*").Result()
	if e != nil {
		t.Fatal(e)
	}
	if len(keys) > 0 {
		a.Del(ctx, keys...)
	}
	c := &Config{GlobalConcurrency: 2, IPConcurrency: 1, RequestsPerMinute: 2, TimeoutSeconds: 2}
	release, e := Acquire(ctx, a, c, "ip-a", "one")
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Acquire(ctx, b, c, "ip-a", "two"); !errors.Is(e, ErrLimited) {
		t.Fatal("IP limit failed", e)
	}
	other, e := Acquire(ctx, b, c, "ip-b", "three")
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Acquire(ctx, a, c, "ip-c", "four"); !errors.Is(e, ErrLimited) {
		t.Fatal("global limit failed", e)
	}
	release()
	other()
	release2, e := Acquire(ctx, b, c, "ip-a", "five")
	if e != nil {
		t.Fatal(e)
	}
	release2()
	if _, e := Acquire(ctx, a, c, "ip-a", "six"); !errors.Is(e, ErrLimited) {
		t.Fatal("rpm limit failed", e)
	}
	cancel, done := context.WithCancel(ctx)
	done()
	if _, e := Acquire(cancel, a, c, "ip-z", "seven"); e == nil {
		t.Fatal("cancelled request admitted")
	}
}
func TestAnonymousSharedLimits(t *testing.T) {
	m := miniredis.RunT(t)
	testLimits(t, m.Addr())
	m.FastForward(61 * time.Second)
}
func TestAnonymousRealRedis(t *testing.T) {
	addr := os.Getenv("ANON_TEST_REDIS")
	if addr == "" {
		t.Skip("isolated Redis only")
	}
	testLimits(t, addr)
}
func TestAnonymousLimiterFailsClosed(t *testing.T) {
	if _, e := Acquire(context.Background(), nil, &Config{}, "x", "x"); e == nil {
		t.Fatal("missing limiter allowed")
	}
}
