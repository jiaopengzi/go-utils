//
// FilePath    : go-utils\boot\boot.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : boot 引导程序公共工具——子进程启动、日志尾缓存、端口占用检测.
//

package boot

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// ChildProcessLogTailLimit 定义子进程日志尾部缓存的最大字节数, 避免过多内存占用.
const ChildProcessLogTailLimit = 8 * 1024

// TailBuffer 仅保留最近写入的一段日志, 用于在子进程异常退出时输出尾部内容.
type TailBuffer struct {
	mu    sync.Mutex
	limit int
	data  []byte
}

// NewTailBuffer 创建一个仅保留固定字节数尾部内容的日志缓冲区.
func NewTailBuffer(limit int) *TailBuffer {
	return &TailBuffer{limit: limit}
}

// Write 写入日志内容, 超出限制时仅保留最新的尾部内容.
func (b *TailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.limit <= 0 {
		return len(p), nil
	}

	if len(p) >= b.limit {
		b.data = append(b.data[:0], p[len(p)-b.limit:]...)
		return len(p), nil
	}

	remaining := len(b.data) + len(p) - b.limit
	if remaining > 0 {
		b.data = append([]byte(nil), b.data[remaining:]...)
	}

	b.data = append(b.data, p...)

	return len(p), nil
}

// String 返回当前缓存的日志尾部内容.
func (b *TailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return string(b.data)
}

// IsPortBindError 判断子进程输出是否为端口占用错误（避免无限重拉）.
func IsPortBindError(errText string) bool {
	lower := strings.ToLower(errText)
	return strings.Contains(lower, "端口已被占用") ||
		strings.Contains(lower, "address already in use") ||
		strings.Contains(lower, "only one usage")
}

// Start 启动子进程并在异常退出时自动重启.
// 当检测到子进程因端口占用退出时, 停止重拉.
func Start(name string, arg ...string) {
	for {
		//nolint:gosec // 命令及参数来自受控配置与调用方，不拼接 shell 字符串
		cmd := exec.Command(name, arg...)
		stdoutTail := NewTailBuffer(ChildProcessLogTailLimit)
		stderrTail := NewTailBuffer(ChildProcessLogTailLimit)

		// 将命令输出透传到当前进程, 同时保留最近一段日志供异常退出时打印.
		cmd.Stdout = io.MultiWriter(os.Stdout, stdoutTail)
		cmd.Stderr = io.MultiWriter(os.Stderr, stderrTail)

		// 启动命令
		if err := cmd.Start(); err != nil {
			zap.L().Fatal("Failed to cmd.Start", zap.Error(err))
			return
		}

		// 等待命令执行完毕
		if err := cmd.Wait(); err != nil {
			fields := []zap.Field{zap.Error(err)}

			if exitErr, ok := err.(*exec.ExitError); ok {
				fields = append(fields, zap.Int("exit_code", exitErr.ExitCode()))
			}

			combinedTail := stdoutTail.String() + stderrTail.String()

			if combined := strings.TrimSpace(combinedTail); combined != "" {
				fields = append(fields, zap.String("output_tail", combined))
			}

			zap.L().Error("Failed to cmd.Wait, will restart", fields...)

			// 端口占用时不重拉
			if IsPortBindError(combinedTail) {
				zap.L().Fatal("端口被占用，已停止重拉", zap.String("appPath", name))
				return
			}
		} else {
			// 子进程正常退出(exit code 0), 不再重启
			zap.L().Info("App exited normally")
			return
		}
	}
}
