package tools

import (
	"context"
	"github.com/gofrs/uuid"
	"time"
)

const (
	DefaultExp = 10 * time.Second
	SleepDur   = 10 * time.Millisecond
)

type RedisLock struct {
	client RedisClient
	uuid   string
	key    string
	cancel context.CancelFunc
}

func NewRedisLock(client RedisClient, key string) (*RedisLock, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &RedisLock{
		client: client,
		uuid:   uu.String(),
		key:    key,
	}, nil
}

func (lock *RedisLock) TryLock(ctx context.Context) (bool, error) {

	suc, err := lock.client.SetNX(ctx, lock.key, "", DefaultExp)
	if err != nil || suc != true {
		return false, err
	}
	_, cancelFunc := context.WithCancel(ctx)

	lock.cancel = cancelFunc
	return suc, nil
}

func (lock *RedisLock) LoopRetryLock(ctx context.Context, times int) (bool, error) {

	for i := 0; i < times; i++ {
		suc, err := lock.TryLock(ctx)
		if err != nil {
			return suc, err
		}
		if suc {
			return suc, nil
		}
		time.Sleep(SleepDur)
	}
	return false, err
}

func (lock *RedisLock) Unlock(ctx context.Context) (bool, error) {
	res, err := NewTools(lock.client).Cad(ctx, lock.key, lock.uuid)
	if err != nil {
		return false, err
	}
	if res {
		lock.cancelFunc()
	}
	return res, nil
}

func (lock *RedisLock) Refresh(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(DefaultExp / 4)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lock.client.Expire(ctx, lock.key, DefaultExp)
			}
		}
	}()
}
