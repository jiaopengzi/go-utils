//
// FilePath    : go-utils\email\pool_implicit_tls_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 隐式 TLS 连接池测试(含 mock SMTP 服务器, 测试重连重试逻辑)
//

package email

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	jordanemail "github.com/jordan-wright/email"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 自签名 TLS 证书生成(仅测试使用)
// ---------------------------------------------------------------------------

func generateTestTLSConfig() *tls.Config {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Test"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, _ := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}

// ---------------------------------------------------------------------------
// Mock SMTP 服务器
// ---------------------------------------------------------------------------

// mockSMTPServer 是一个简单的 SMTP 模拟服务器.
type mockSMTPServer struct {
	listener  net.Listener
	addr      string
	tlsConfig *tls.Config
	wg        sync.WaitGroup
	closeOnce sync.Once
	mailCount atomic.Int32 // 成功接收的邮件计数

	// 控制行为: 首次会话中在 MAIL 命令时断开连接, 模拟空闲连接被断开
	failFirstSend atomic.Bool
	failCount     atomic.Int32 // 已失败的次数
}

func newMockSMTPServer(t *testing.T) *mockSMTPServer {
	t.Helper()

	serverTLS := generateTestTLSConfig()
	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverTLS)
	require.NoError(t, err)

	s := &mockSMTPServer{
		listener:  listener,
		addr:      listener.Addr().String(),
		tlsConfig: serverTLS,
	}

	s.wg.Add(1)
	go s.acceptLoop(t)

	return s
}

func (s *mockSMTPServer) acceptLoop(t *testing.T) {
	t.Helper()
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return // listener 已关闭
		}

		s.wg.Add(1)
		go s.handleConn(t, conn)
	}
}

func (s *mockSMTPServer) handleConn(t *testing.T, conn net.Conn) {
	t.Helper()
	defer s.wg.Done()
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// 发送 SMTP 欢迎
	_, _ = fmt.Fprintf(conn, "220 localhost Mock SMTP\r\n")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		cmd := strings.ToUpper(line)

		switch {
		case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
			_, _ = fmt.Fprintf(conn, "250-localhost\r\n250 OK\r\n")

		case strings.HasPrefix(cmd, "MAIL FROM:"):
			// 如果设置了 failFirstSend, 首次 MAIL 命令时断开连接(模拟空闲断连)
			if s.failFirstSend.Load() && s.failCount.Add(1) <= 1 {
				// 直接关闭连接, 模拟底层 TCP 断开
				return
			}
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")

		case strings.HasPrefix(cmd, "RCPT TO:"):
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")

		case cmd == "DATA":
			_, _ = fmt.Fprintf(conn, "354 Start mail input\r\n")
			// 读取数据直到 \r\n.\r\n
			for {
				dataLine, readErr := reader.ReadString('\n')
				if readErr != nil {
					return
				}
				if strings.TrimSpace(dataLine) == "." {
					break
				}
			}
			s.mailCount.Add(1)
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")

		case cmd == "RSET":
			_, _ = fmt.Fprintf(conn, "250 OK\r\n")

		case cmd == "QUIT":
			_, _ = fmt.Fprintf(conn, "221 Bye\r\n")
			return

		default:
			_, _ = fmt.Fprintf(conn, "500 Unknown command\r\n")
		}
	}
}

func (s *mockSMTPServer) close() {
	s.closeOnce.Do(func() {
		_ = s.listener.Close()
	})
	s.wg.Wait()
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

func newTestEmail() *jordanemail.Email {
	e := jordanemail.NewEmail()
	e.From = "sender@example.com"
	e.To = []string{"recipient@example.com"}
	e.Subject = "Test"
	e.Text = []byte("Hello")

	return e
}

func newTestPool(t *testing.T, server *mockSMTPServer, workers int) *implicitTLSPool {
	t.Helper()

	// 客户端跳过证书验证(因为是自签名证书)
	clientTLS := &tls.Config{
		InsecureSkipVerify: true,
	}

	return newImplicitTLSPool(server.addr, "localhost", nil, clientTLS, workers)
}

// ---------------------------------------------------------------------------
// 测试用例
// ---------------------------------------------------------------------------

// TestImplicitTLSPool_SendSuccess 正常发送邮件成功
func TestImplicitTLSPool_SendSuccess(t *testing.T) {
	server := newMockSMTPServer(t)
	defer server.close()

	pool := newTestPool(t, server, 2)
	defer pool.Close()

	e := newTestEmail()

	err := pool.Send(e, 5*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), server.mailCount.Load())
}

// TestImplicitTLSPool_SendMultiple 连续发送多封邮件, 验证连接复用
func TestImplicitTLSPool_SendMultiple(t *testing.T) {
	server := newMockSMTPServer(t)
	defer server.close()

	pool := newTestPool(t, server, 1) // 1 个 worker, 确保连接复用
	defer pool.Close()

	for i := range 5 {
		e := newTestEmail()
		e.Subject = fmt.Sprintf("Test %d", i)
		err := pool.Send(e, 5*time.Second)
		assert.NoError(t, err, "第 %d 封邮件发送应成功", i+1)
	}

	assert.Equal(t, int32(5), server.mailCount.Load())
}

// TestImplicitTLSPool_RetryOnStaleConnection 空闲连接断开时自动重连重试
// 这是此次修复的核心测试用例
func TestImplicitTLSPool_RetryOnStaleConnection(t *testing.T) {
	server := newMockSMTPServer(t)
	defer server.close()

	pool := newTestPool(t, server, 1) // 1 个 worker
	defer pool.Close()

	// 先发送一封成功的邮件, 建立连接
	e1 := newTestEmail()
	e1.Subject = "First"
	err := pool.Send(e1, 5*time.Second)
	require.NoError(t, err, "首封邮件应成功")

	// 设置: 下次 MAIL 命令时服务器断开连接(模拟空闲连接失效)
	server.failFirstSend.Store(true)

	// 发送第二封邮件: worker 应检测到连接断开, 自动重连并重试成功
	e2 := newTestEmail()
	e2.Subject = "After stale"
	err = pool.Send(e2, 5*time.Second)
	assert.NoError(t, err, "空闲连接断开后, 重连重试应成功")

	// 验证: 两封邮件都成功发送
	assert.Equal(t, int32(2), server.mailCount.Load())
}

// TestImplicitTLSPool_ConnectFailure 连接失败时返回错误
func TestImplicitTLSPool_ConnectFailure(t *testing.T) {
	// 使用一个不存在的地址
	clientTLS := &tls.Config{InsecureSkipVerify: true}
	pool := newImplicitTLSPool("127.0.0.1:1", "localhost", nil, clientTLS, 1)
	defer pool.Close()

	e := newTestEmail()
	err := pool.Send(e, 3*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "smtp connect error")
}

// TestImplicitTLSPool_SendTimeout 超时测试
func TestImplicitTLSPool_SendTimeout(t *testing.T) {
	// 创建一个池, 但 channel 很小且没有 worker(模拟队列满/处理慢)
	pool := &implicitTLSPool{
		tasks: make(chan tlsTask, 0), // 无缓冲
	}
	// 不启动 worker, 入队会超时

	e := newTestEmail()
	err := pool.Send(e, 50*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

// TestImplicitTLSPool_Close worker 优雅退出
func TestImplicitTLSPool_Close(t *testing.T) {
	server := newMockSMTPServer(t)
	defer server.close()

	pool := newTestPool(t, server, 2)

	// 先发送一封确保 worker 建立了连接
	e := newTestEmail()
	err := pool.Send(e, 5*time.Second)
	require.NoError(t, err)

	// 关闭池, worker 应优雅退出(不会 panic 或死锁)
	pool.Close()

	// 给 worker goroutine 一点时间退出
	time.Sleep(100 * time.Millisecond)
}

// TestImplicitTLSPool_ConcurrentSend 并发发送
func TestImplicitTLSPool_ConcurrentSend(t *testing.T) {
	server := newMockSMTPServer(t)
	defer server.close()

	pool := newTestPool(t, server, 4)
	defer pool.Close()

	const count = 20
	var wg sync.WaitGroup
	errors := make([]error, count)

	for i := range count {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			e := newTestEmail()
			e.Subject = fmt.Sprintf("Concurrent %d", idx)
			errors[idx] = pool.Send(e, 10*time.Second)
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "并发邮件 %d 应成功", i)
	}
	assert.Equal(t, int32(count), server.mailCount.Load())
}

// TestImplicitTLSPool_SendNoTimeout 无超时模式 (timeout <= 0) 正常发送
func TestImplicitTLSPool_SendNoTimeout(t *testing.T) {
	server := newMockSMTPServer(t)
	defer server.close()

	pool := newTestPool(t, server, 1)
	defer pool.Close()

	e := newTestEmail()

	err := pool.Send(e, 0) // timeout = 0, 无超时
	assert.NoError(t, err)
	assert.Equal(t, int32(1), server.mailCount.Load())
}

// ---------------------------------------------------------------------------
// emailOnly / addressLists 内部函数测试
// ---------------------------------------------------------------------------

func TestEmailOnly(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"纯地址", "bob@example.com", "bob@example.com", false},
		{"带名称", "Bob <bob@example.com>", "bob@example.com", false},
		{"带引号名称", `"Bob Smith" <bob@example.com>`, "bob@example.com", false},
		{"无效地址", "not-an-email", "", true},
		{"空字符串", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := emailOnly(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAddressLists(t *testing.T) {
	tests := []struct {
		name    string
		lists   [][]string
		want    []string
		wantErr bool
	}{
		{
			name:  "混合列表",
			lists: [][]string{{"a@b.com", "Bob <c@d.com>"}, {"e@f.org"}},
			want:  []string{"a@b.com", "c@d.com", "e@f.org"},
		},
		{
			name:  "空列表",
			lists: [][]string{},
			want:  nil,
		},
		{
			name:    "包含无效地址",
			lists:   [][]string{{"a@b.com", "invalid"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addressLists(tt.lists...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
