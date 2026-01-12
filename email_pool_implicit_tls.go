//
// FilePath    : go-utils\email_pool_implicit_tls.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 隐式 TLS 邮件连接池工具, 参考 jordan-wright/email.Pool 实现
//

package utils

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"
	"go.uber.org/zap"
)

// implicitTLSPool 管理一组 worker, 每个 worker 维护一个通过 tls.Dial 建立的
// 隐式 TLS `smtp.Client`, 用于复用 465 端口的 TLS 会话.
type implicitTLSPool struct {
	addr      string       // SMTP 服务器地址 (host:port)
	host      string       // SMTP 服务器主机名
	auth      smtp.Auth    // SMTP 认证信息
	tlsConfig *tls.Config  // TLS 配置
	tasks     chan tlsTask // 发送任务队列
	workers   int          // 工作协程数
}

// tlsTask 表示发送邮件的请求及其响应通道.
type tlsTask struct {
	e    *email.Email // 要发送的邮件
	resp chan error   // 发送结果响应通道
}

// newImplicitTLSPool 创建并启动隐式 TLS 池的 worker(工作协程).
func newImplicitTLSPool(addr, host string, auth smtp.Auth, tlsCfg *tls.Config, workers int) *implicitTLSPool {
	if workers <= 0 {
		workers = 4 // 默认工作协程数
	}

	// 初始化池并启动 worker goroutines
	p := &implicitTLSPool{
		addr:      addr,
		host:      host,
		auth:      auth,
		tlsConfig: tlsCfg,
		tasks:     make(chan tlsTask, workers*4),
		workers:   workers,
	}

	for range workers {
		go p.workerLoop()
	}

	return p
}

// Send 将发送请求放入隐式 TLS 池的任务队列并等待结果或超时.
func (p *implicitTLSPool) Send(e *email.Email, timeout time.Duration) error {
	req := tlsTask{e: e, resp: make(chan error, 1)}

	if timeout >= 0 {
		select {
		case p.tasks <- req:
		case <-time.After(timeout):
			return fmt.Errorf("timed out enqueuing email send request")
		}
	} else {
		p.tasks <- req
	}

	if timeout >= 0 {
		select {
		case err := <-req.resp:
			return err
		case <-time.After(timeout):
			return fmt.Errorf("timed out waiting for send result")
		}
	}

	return <-req.resp
}

// workerLoop 是每个 worker 的主循环：确保与服务器建立连接,
// 用现有 client 发送邮件, 发生错误时重建连接.
func (p *implicitTLSPool) workerLoop() {
	var (
		client  *smtp.Client
		connErr error
	)

	// 处理任务队列
	for task := range p.tasks {
		// 确保 client 已连接并完成认证
		if client == nil {
			client, connErr = p.connect()
			if connErr != nil {
				task.resp <- connErr
				continue
			}
		}

		// 尝试发送邮件
		err := p.sendWithClient(client, task.e)
		if err != nil {
			// 出错时关闭并重置 client, 记录 Quit 错误, 但仍上报原始发送错误
			if quitErr := client.Quit(); quitErr != nil {
				zap.L().Error("smtp client quit error", zap.Error(quitErr))
			}

			client = nil
			task.resp <- err

			continue
		}

		// 发送成功
		task.resp <- nil
	}
}

// connect 与服务器建立隐式 TLS 连接并返回认证后的 smtp.Client
func (p *implicitTLSPool) connect() (*smtp.Client, error) {
	// 建立 TCP + TLS 连接(隐式 TLS)
	conn, err := tls.Dial("tcp", p.addr, p.tlsConfig)
	if err != nil {
		return nil, err
	}

	// 使用已建立的 TLS 连接创建 smtp.Client
	c, err := smtp.NewClient(conn, p.host)
	if err != nil {
		return nil, err
	}

	// 如需认证则执行 Auth
	if p.auth != nil {
		if err := c.Auth(p.auth); err != nil {
			c.Close()
			return nil, err
		}
	}

	return c, nil
}

// sendWithClient 使用给定的 smtp.Client 发送邮件数据, 并在成功后调用 Reset 以复用会话
func (p *implicitTLSPool) sendWithClient(c *smtp.Client, e *email.Email) error {
	// 组合收件人列表并生成邮件字节流
	recipients, err := addressLists(e.To, e.Cc, e.Bcc)
	if err != nil {
		return err
	}

	msg, err := e.Bytes()
	if err != nil {
		return err
	}

	// 解析发件人地址
	from, err := emailOnly(e.From)
	if err != nil {
		return err
	}

	// 设置 Mail From
	if err = c.Mail(from); err != nil {
		return err
	}

	// 添加每个 RCPT
	for _, r := range recipients {
		if err = c.Rcpt(r); err != nil {
			return err
		}
	}

	// 写入邮件内容
	w, err := c.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write(msg); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	// 重置会话以便连接可被复用用于下条消息
	if err = c.Reset(); err != nil {
		return err
	}

	return nil
}

// emailOnly 从可能带名称的邮件地址中提取纯邮箱地址, 例如 "Bob <bob@example.com>" -> "bob@example.com"
func emailOnly(full string) (string, error) {
	addr, err := mail.ParseAddress(full)
	if err != nil {
		return "", err
	}

	return addr.Address, nil
}

// addressLists 将多个地址切片展平成单个仅包含邮箱地址的切片
func addressLists(lists ...[]string) ([]string, error) {
	var combined []string

	for _, lst := range lists {
		for _, full := range lst {
			addr, err := emailOnly(full)
			if err != nil {
				return nil, err
			}

			combined = append(combined, addr)
		}
	}

	return combined, nil
}
