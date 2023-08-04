package tools

import (
	"context"
	"time"
)

type RedisClient interface {
	Expire(context context.Context, key string, expire time.Duration)
	SetNX(context context.Context, key string, args interface{}, expireTime time.Duration) (bool, error)
	Delete(context context.Context)
	Eval(context context.Context, scripts string, key string, expireTime time.Duration)
}
