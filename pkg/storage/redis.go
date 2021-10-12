package storage

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/flashpaper/pkg/web"
)

type Redis struct {
	*redis.Client

	l hclog.Logger
}

func NewRedis(l hclog.Logger) (web.Storage, error) {
	r := Redis{l: l.Named("redis")}
	r.Client = redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})

	return &r, nil
}

func (r *Redis) PutEx(ctx context.Context, key string, value interface{}, d time.Duration) error {
	return r.SetEX(ctx, key, value, d).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *Redis) Del(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
