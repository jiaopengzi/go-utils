//
// FilePath    : go-utils\email\pool.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 邮件连接池工具
//

package email

import (
	"crypto/tls"
	"net"
	"net/smtp"
	"sync"
	"time"

	jordanemail "github.com/jordan-wright/email"
)

// 连接池管理(按 addr 缓存 pool)
var (
	poolsMu   sync.Mutex                    // 保护 smtpPools 并发访问
	smtpPools = make(map[string]poolSender) // 按 addr 缓存的连接池实例
)

// poolSender 连接池抽象, 支持不同实现(例如 github.com/jordan-wright/email.Pool 或 隐式 TLS 自实现)
type poolSender interface {
	Send(e *jordanemail.Email, timeout time.Duration) error
}

// poolCloser 带有 Close 方法的池, 用于关闭连接池释放资源.
type poolCloser interface {
	Close()
}

// SMTPPool 管理按 addr 缓存的连接池实例
type SMTPPool struct {
	Addr string    // SMTP 服务器地址 (host:port)
	Auth smtp.Auth // SMTP 认证信息
	Size int       // 连接池大小(工作协程数)
}

// GetPool 返回按 addr 缓存的连接池实例. 对于 465 端口返回隐式 TLS 池,
// 否则返回基于 email.NewPool 的 STARTTLS 池.
func GetPool(smtp *SMTPPool) (poolSender, error) {
	poolsMu.Lock()
	defer poolsMu.Unlock()

	if p, ok := smtpPools[smtp.Addr]; ok {
		return p, nil
	}

	// 判断是否为隐式 TLS (通常为 465 端口)
	host, port, err := net.SplitHostPort(smtp.Addr)
	if err != nil {
		return nil, err
	}

	// 为隐式 TLS 创建自定义工作池
	if port == "465" {
		tlsCfg := &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		}

		p := newImplicitTLSPool(smtp.Addr, host, smtp.Auth, tlsCfg, smtp.Size)
		smtpPools[smtp.Addr] = p

		return p, nil
	}

	// 默认使用 email.Pool(支持 STARTTLS, 如 587)
	p, err := jordanemail.NewPool(smtp.Addr, smtp.Size, smtp.Auth)
	if err != nil {
		return nil, err
	}

	smtpPools[smtp.Addr] = p

	return p, nil
}

// CloseAllPools 关闭并清理所有缓存的连接池.
// 通常在应用优雅关闭时调用, 释放底层 SMTP 连接.
func CloseAllPools() {
	poolsMu.Lock()
	defer poolsMu.Unlock()

	for addr, p := range smtpPools {
		if closer, ok := p.(poolCloser); ok {
			closer.Close()
		}

		delete(smtpPools, addr)
	}
}

// ResetPools 重置连接池缓存, 仅用于测试.
func ResetPools() {
	CloseAllPools()
}
