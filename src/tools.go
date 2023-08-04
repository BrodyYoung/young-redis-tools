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
	res, err := t.client.Eval(ctx, compareAndSwapScript, oldValue, newValue)

	if err != nil {
		return false, err
	}
}

func (t RedisTools) CasEx(ctx context.Context, key string, oldValue interface{}, newValue interface{}, expire time.Duration) (bool, error) {
	if expire == 0 {
		return t.Cas(ctx, key, oldValue, newValue)
	}

	if usePrecise(expire) {
		ti := parseMill(expire)
		res, err := t.client.Eval(ctx, fmt.Sprintf(compareAndSwapEXScript, ti), oldValue, newValue)

	} else if expire%time.Second == 0 {
		ti := parseSec(expire)
		res, err := t.client.Eval(ctx, fmt.Sprintf(compareAndSwapEXScript, ti), oldValue, newValue)

	} else {
		res, err := t.client.Eval(ctx, compareAndSwapKeepTTLScript, oldValue, newValue)
	}

}

func (t RedisTools) Cad(ctx context.Context, key string, oldValue interface{}) (bool, error) {

}
