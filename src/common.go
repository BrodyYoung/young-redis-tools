package tools

import (
	"context"
	"time"
)

type RedisClient interface {
	Expire(context context.Context, key string, expire time.Duration) *redis.BoolCmd
	SetNX(context context.Context, key string, args interface{}, expireTime time.Duration) *redis.IntCmd
	Delete(context context.Context) *redis.boolCmd
	Eval(context context.Context, scripts string, key []string, args ...interface{}) *redis.Cmd
}
