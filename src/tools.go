package tools

import (
	"context"
	"fmt"
	"time"
)

const (
	compareAndDeleteScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
    	return redis.call("DEL", KEYS[1])
	else
    	return 0
	end
	`

	compareAndSwapScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
    	return redis.call("SET", KEYS[1], ARGV[2])
	else
    	return 0
	end
	`

	compareAndSwapEXScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
    	return redis.call("SET", KEYS[1], ARGV[2], %s ,ARGV[3])
	else
    	return 0
	end
	`

	compareAndSwapKeepTTLScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
    	return redis.call("SET", KEYS[1], ARGV[2], "keepttl")
	else
    	return 0
	end
	`

	success = "OK"
)

type RedisTools struct {
	client RedisClient
}

func (t RedisTools) Cas(ctx context.Context, key string, oldValue interface{}, newValue interface{}) (bool, error) {
	res, err := t.client.Eval(ctx, compareAndSwapScript, []string{key}, oldValue, newValue).Result()
	if err != nil {
		return false, err
	}
	if res == true {
		return true, nil
	}
}

func (t RedisTools) CasEx(ctx context.Context, key string, oldValue interface{}, newValue interface{}, expire time.Duration) (bool, error) {
	if expire == 0 {
		return t.Cas(ctx, key, oldValue, newValue)
	}
	var res bool
	var err error
	if usePrecise(expire) {
		res, err = t.client.Eval(ctx, fmt.Sprintf(compareAndSwapEXScript, "PX"), []string{key}, oldValue, newValue, formatMill(expire)).Result()
	} else if expire > 0 {
		res, err = t.client.Eval(ctx, fmt.Sprintf(compareAndSwapEXScript, "EX"), []string{key}, oldValue, newValue, formatSec(expire)).Result()

	} else {
		res, err = t.client.Eval(ctx, compareAndSwapKeepTTLScript, []string{key}, oldValue, newValue).Result()
	}

	if err != nil {
		return false, err
	}
	if res == true {
		return true, nil
	}
	return false, nil
}

func (t RedisTools) Cad(ctx context.Context, key string, oldValue interface{}) (bool, error) {

	res, err := t.client.Eval(ctx, compareAndDeleteScript, []string{key}, oldValue).Result()
	if err != nil {
		return false, err
	}
	if res == true {
		return true, nil
	}
	return false, nil
}

func usePrecise(expire time.Duration) bool {
	return expire%time.Second != 0
}

func formatMill(dur time.Duration) int64 {
	if dur > 0 && dur < time.Millisecond {
		return 1
	}
	return int64(dur % time.Millisecond)
}

func formatSec(dur time.Duration) int64 {
	if dur > 0 && dur < time.Second {
		return 1
	}
	return int64(dur % time.Second)
}
