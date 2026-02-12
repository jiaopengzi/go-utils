//
// FilePath    : go-utils\email\pool_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 邮件连接池测试
//

package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPool_InvalidAddr(t *testing.T) {
	defer ResetPools()

	_, err := GetPool(&SMTPPool{
		Addr: "invalid-addr-no-port",
		Auth: nil,
		Size: 2,
	})
	assert.Error(t, err, "无效地址应返回错误")
}

func TestGetPool_CachesPool(t *testing.T) {
	defer ResetPools()

	// 使用 465 端口创建隐式 TLS 池(不需要真实服务器, pool 创建时不连接)
	p1, err := GetPool(&SMTPPool{
		Addr: "localhost:465",
		Auth: nil,
		Size: 1,
	})
	assert.NoError(t, err)
	assert.NotNil(t, p1)

	// 第二次获取应返回相同实例
	p2, err := GetPool(&SMTPPool{
		Addr: "localhost:465",
		Auth: nil,
		Size: 1,
	})
	assert.NoError(t, err)
	assert.Equal(t, p1, p2, "相同地址应返回缓存的池实例")
}

func TestCloseAllPools(t *testing.T) {
	defer ResetPools()

	// 创建一个池
	_, err := GetPool(&SMTPPool{
		Addr: "localhost:465",
		Auth: nil,
		Size: 1,
	})
	assert.NoError(t, err)

	// 关闭所有池
	CloseAllPools()

	// 再次获取应创建新池(不是缓存的旧池)
	p, err := GetPool(&SMTPPool{
		Addr: "localhost:465",
		Auth: nil,
		Size: 1,
	})
	assert.NoError(t, err)
	assert.NotNil(t, p)
}
