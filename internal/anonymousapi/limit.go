package anonymousapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

// Three keys share a cluster hash tag. Expiring leases bound crash recovery.
const acquire = `
local now=tonumber(ARGV[1]); local expire=tonumber(ARGV[2]);
redis.call('ZREMRANGEBYSCORE',KEYS[1],'-inf',now);redis.call('ZREMRANGEBYSCORE',KEYS[2],'-inf',now)
local n=tonumber(redis.call('GET',KEYS[3]) or '0')
if redis.call('ZCARD',KEYS[1])>=tonumber(ARGV[4]) or redis.call('ZCARD',KEYS[2])>=tonumber(ARGV[5]) or n>=tonumber(ARGV[6]) then return 0 end
redis.call('ZADD',KEYS[1],expire,ARGV[3]);redis.call('ZADD',KEYS[2],expire,ARGV[3]);redis.call('PEXPIRE',KEYS[1],expire-now+1000);redis.call('PEXPIRE',KEYS[2],expire-now+1000)
local v=redis.call('INCR',KEYS[3]);if v==1 then redis.call('PEXPIRE',KEYS[3],60000) end
return 1`

var ErrLimited = errors.New("anonymous API limit exceeded")

func Acquire(ctx context.Context, rdb *redis.Client, c *Config, ip, id string) (func(), error) {
	if rdb == nil {
		return nil, errors.New("anonymous API requires shared limiter")
	}
	hash := sha256.Sum256([]byte(ip))
	prefix := "weknora:{anonymous-api}:"
	host := hex.EncodeToString(hash[:])
	keys := []string{prefix + "active", prefix + "ip:" + host, prefix + "rpm:" + host}
	now := time.Now()
	n, e := rdb.Eval(ctx, acquire, keys, now.UnixMilli(), now.Add(time.Duration(c.TimeoutSeconds+30)*time.Second).UnixMilli(), id, c.GlobalConcurrency, c.IPConcurrency, c.RequestsPerMinute).Int()
	if e != nil {
		return nil, e
	}
	if n != 1 {
		return nil, ErrLimited
	}
	return func() {
		stop, done := context.WithTimeout(context.Background(), 3*time.Second)
		defer done()
		rdb.ZRem(stop, keys[0], id)
		rdb.ZRem(stop, keys[1], id)
	}, nil
}
