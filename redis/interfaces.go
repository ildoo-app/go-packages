package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// RedisClient is the subset of Redis commands commonly used by application code.
// *Client satisfies it, as does a go-redis ClusterClient / Pipeline where applicable.
type RedisClient interface {
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *goredis.StatusCmd
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	Exists(ctx context.Context, keys ...string) *goredis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *goredis.BoolCmd
	TTL(ctx context.Context, key string) *goredis.DurationCmd
	HSet(ctx context.Context, key string, values ...any) *goredis.IntCmd
}
