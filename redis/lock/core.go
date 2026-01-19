//
// FilePath    : go-utils\redis\lock\core.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 业务锁，入口文件.
//

// Package lock 业务锁，负责业务锁的管理
package lock

import (
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// Key 锁的Key
type Key string

// Locker 封装锁
type Locker struct {
	Rs          *redsync.Redsync // redsync 实例
	LockTimeout int              // 锁的超时时间(毫秒)
}

// MyLocker 全局变量锁实例
var MyLocker *Locker

// Init 初始化 Locker
// - delegate: Redis 客户端
// - lockTimeout: 锁的超时时间(毫秒)
func Init(delegate redis.UniversalClient, lockTimeout int) {
	pool := goredis.NewPool(delegate)
	rs := redsync.New(pool)
	MyLocker = &Locker{
		Rs:          rs,
		LockTimeout: lockTimeout,
	}
}

// KeepLockAlive 为锁续期
//   - mutex: 互斥锁
//   - stopChan: 停止信号
func KeepLockAlive(mutex *redsync.Mutex, stopChan chan struct{}) error {
	// 定时器
	ticker := time.NewTicker(time.Duration(MyLocker.LockTimeout) * time.Millisecond / 3) // 续期时间

	// 延迟关闭定时器
	defer ticker.Stop()

	// 循环续期
	for {
		select {
		case <-ticker.C:
			if _, err := mutex.Extend(); err != nil {
				return err // 续期失败，返回错误
			}
		case <-stopChan:
			return nil // 接收到停止信号，退出
		}
	}
}
