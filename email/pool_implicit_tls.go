//
// FilePath    : go-utils\email\pool_implicit_tls.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 隐式 TLS 邮件连接池工具, 参考 jordan-wright/email.Pool 实现
//

package email

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"time"

	jordanemail "github.com/jordan-wright/email"
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
	e    *jordanemail.Email // 要发送的邮件
	resp chan error         // 发送结果响应通道
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
// 当 timeout > 0 时, 入队与等待结果共享同一个定时器(总超时);
// 当 timeout <= 0 时, 无超时限制, 阻塞直到发送完成.
func (p *implicitTLSPool) Send(e *jordanemail.Email, timeout time.Duration) error {
	req := tlsTask{e: e, resp: make(chan error, 1)}

	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		// 入队(带超时)
		select {
		case p.tasks <- req:
		case <-timer.C:
			return fmt.Errorf("timed out enqueuing email send request")
		}

		// 等待发送结果(共享同一个 timer, 剩余时间内)
		select {
		case err := <-req.resp:
			return err
		case <-timer.C:
			return fmt.Errorf("timed out waiting for send result")
		}
	}

	// 无超时: 阻塞入队并等待结果
	p.tasks <- req

	return <-req.resp
}

// Close 关闭任务队列, 所有 worker 将完成当前任务后退出并释放 SMTP 连接.
func (p *implicitTLSPool) Close() {
	close(p.tasks)
}

// workerLoop 是每个 worker 的主循环：确保与服务器建立连接,
// 用现有 client 发送邮件, 发生错误时自动重连并重试一次.
func (p *implicitTLSPool) workerLoop() {
	var client *smtp.Client

	// 退出时清理 SMTP 连接
	defer func() {
		if client != nil {
			// 优雅关闭: 发送 QUIT 命令; 失败则强制关闭
			if err := client.Quit(); err != nil {
				p.closeClient(client)
			}
		}
	}()

	// 处理任务队列
	for task := range p.tasks {
		// 确保 client 已连接并完成认证
		if err := p.ensureClientConnected(&client); err != nil {
			task.resp <- fmt.Errorf("smtp connect error: %w", err)
			continue
		}

		// 尝试发送邮件并在失败时重试一次
		if err := p.trySendWithRetry(&client, task.e); err != nil {
			task.resp <- err
			continue
		}

		// 发送成功(首次或重试), 重置会话以复用连接
		if err := client.Reset(); err != nil {
			// Reset 失败说明连接已损坏, 丢弃连接, 下次任务将重新连接
			// 但邮件已经发送成功, 不影响本次结果
			zap.L().Warn("smtp session reset failed, discarding connection", zap.Error(err))
			p.closeClient(client)
			client = nil
		}

		task.resp <- nil
	}
}

// ensureClientConnected 确保传入的 client 已连接并认证, 若为 nil 则建立连接
func (p *implicitTLSPool) ensureClientConnected(client **smtp.Client) error {
	if *client == nil {
		c, err := p.connect()
		if err != nil {
			return err
		}
		*client = c
	}
	return nil
}

// trySendWithRetry 使用现有 client 发送, 若失败则重连并重试一次。成功返回 nil, 失败返回具体错误。
func (p *implicitTLSPool) trySendWithRetry(client **smtp.Client, e *jordanemail.Email) error {
	if err := p.sendWithClient(*client, e); err != nil {
		// 首次发送失败(可能是空闲连接被断开), 关闭旧连接并重试一次
		zap.L().Warn("smtp send failed, reconnecting and retrying",
			zap.String("addr", p.addr),
			zap.Error(err),
		)
		p.closeClient(*client)
		*client = nil

		// 重新建立连接
		c, connErr := p.connect()
		if connErr != nil {
			*client = nil
			return fmt.Errorf("smtp send failed: %w; reconnect also failed: %w", err, connErr)
		}

		*client = c

		// 用新连接重试发送
		if retryErr := p.sendWithClient(*client, e); retryErr != nil {
			p.closeClient(*client)
			*client = nil
			return fmt.Errorf("smtp retry send failed: %w (original error: %w)", retryErr, err)
		}
	}

	return nil
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
		if cerr := conn.Close(); cerr != nil {
			zap.L().Debug("tls conn close error", zap.Error(cerr))
		}
		return nil, err
	}

	// 如需认证则执行 Auth
	if p.auth != nil {
		if err := c.Auth(p.auth); err != nil {
			if cerr := c.Close(); cerr != nil {
				zap.L().Debug("smtp client close error", zap.Error(cerr))
			}
			return nil, err
		}
	}

	return c, nil
}

// closeClient 安全关闭 smtp.Client, 仅使用 Close 不发送 QUIT 命令,
// 适用于连接已损坏需要丢弃的场景.
func (p *implicitTLSPool) closeClient(c *smtp.Client) {
	if c != nil {
		if err := c.Close(); err != nil {
			zap.L().Debug("smtp client close error", zap.Error(err))
		}
	}
}

// sendWithClient 使用给定的 smtp.Client 发送邮件数据.
// 注意: Reset 由 workerLoop 在发送成功后单独调用, 避免 Reset 失败被误判为发送失败.
func (p *implicitTLSPool) sendWithClient(c *smtp.Client, e *jordanemail.Email) error {
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
