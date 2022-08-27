package etcdtool

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	lockPrefix = "etcd_lock_"
)

type EtcdLock struct {
	m *concurrency.Mutex
}

// 创建一把分布式锁
func NewLocker(lockKey string, timeout int) (locker *EtcdLock, cancelFunc func(), err error) {
	if lockKey == "" || timeout <= 0 || etcdcli == nil {
		err = errors.Errorf("GetLocker_err args err lockKey = %s , timeout = %s , cli = %+v \n ",
			lockKey, timeout, etcdcli)
		return
	}
	lockKey = lockPrefix + lockKey
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	cancelFunc = cancel
	ss, e := concurrency.NewSession(etcdcli, concurrency.WithContext(ctx), concurrency.WithTTL(1))
	if e != nil {
		err = errors.Wrapf(e, "GetLocker_err NewSession_err")
		return
	}
	myLocker := concurrency.NewMutex(ss, lockKey)
	locker = &EtcdLock{m: myLocker}
	return
}

// 锁住代码
// isWait 是否阻塞等待锁释放
func (locker *EtcdLock) Lock(isWait bool) (err error) {

	var ctx context.Context
	var cancel func()

	if isWait {
		ctx = context.TODO()
	} else {
		// 获取锁的等待时间 ,1 秒未获取到锁，会取消获取
		ctx2, cn := context.WithTimeout(context.Background(), time.Second*1)
		cancel = cn
		ctx = ctx2
	}
	err = locker.m.Lock(ctx)
	if err != nil {
		err = errors.Wrapf(err, "Lock_err")
		if cancel != nil {
			cancel()
		}
		return
	}
	return
}

// 释放锁
func (locker *EtcdLock) UnLock() (err error) {
	err = locker.m.Unlock(context.Background())
	if err != nil {
		err = errors.Wrapf(err, "Lock_err")
		return
	}
	return
}
