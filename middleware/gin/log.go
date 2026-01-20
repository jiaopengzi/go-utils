//
// FilePath    : go-utils\middleware\gin\log.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 日志中间件
//

package mwgin

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiaopengzi/go-utils/res"
	"go.uber.org/zap"
)

// ZapLogger 使用 zap 接管 gin 框架默认的日志
func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 格式化日志
		fields := genZapFields(c, start)

		// 根据状态码决定 zap 日志级别：>=500 -> Error, >=400 -> Warn, else Info
		status := c.Writer.Status()

		switch {
		case status >= http.StatusInternalServerError:
			zap.L().Error("[GIN]", fields...)
		case status >= http.StatusBadRequest:
			zap.L().Warn("[GIN]", fields...)
		default:
			zap.L().Info("[GIN]", fields...)
		}
	}
}

// GinRecovery 对 gin 框架的错误恢复中间件 当请求处理过程中发生 panic 时,
// 该中间件可以捕获 panic, 并在日志中记录相关信息, 避免程序异常终止;
// stack 参数决定是否打印堆栈信息
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用 defer 来确保在函数执行完毕之前调用后面的代码块
		defer func() {
			handlePanic(c, recover(), stack)
		}()

		c.Next()
	}
}

// handlePanic 处理 recover 后的错误记录与响应
func handlePanic(c *gin.Context, rec any, stack bool) {
	// 如果 recover 为 nil，则直接返回
	if rec == nil {
		return
	}

	// 尝试将 recover 返回值转换为 error
	err := convertToError(rec)
	if err == nil {
		return
	}

	// 安全地获取请求的 dump, 发生错误时记录错误但继续处理
	httpRequest, dumpErr := dumpRequest(c.Request)
	if dumpErr != nil {
		// 格式化日志
		zap.L().Error("[GIN HandlePanic]", append(genZapFields(c), zap.Any("dump_error", dumpErr))...)
	}

	// 检测是否为断开连接相关错误
	if detectBrokenPipe(err) {
		handleBrokenPipe(c, err, httpRequest)
		return
	}

	// 记录 panic 日志
	logRecovery(err, httpRequest, stack)

	// 终止请求, 并返回 Internal Server Error (500) 状态码
	c.AbortWithStatus(http.StatusInternalServerError)
}

// convertToError 尝试将 recover 返回值转换为 error
func convertToError(rec any) error {
	if rec == nil {
		return nil
	}

	if err, ok := rec.(error); ok {
		return err
	}

	return fmt.Errorf("%v", rec)
}

// detectBrokenPipe 判断错误是否为断开连接相关错误
func detectBrokenPipe(err error) bool {
	var netErr *net.OpError
	if !errors.As(err, &netErr) {
		return false
	}

	var syscallErr *os.SyscallError
	if !errors.As(netErr.Err, &syscallErr) {
		return false
	}

	se := strings.ToLower(syscallErr.Error())

	// 判断是否包含断开连接的关键字
	return strings.Contains(se, "broken pipe") || strings.Contains(se, "connection reset by peer")
}

// dumpRequest 安全地获取请求的 dump, 发生错误时返回错误供调用者记录
func dumpRequest(r *http.Request) ([]byte, error) {
	return httputil.DumpRequest(r, false)
}

// handleBrokenPipe 处理断连场景的日志与请求中断
func handleBrokenPipe(c *gin.Context, err error, httpRequest []byte) {
	// 记录断连日志
	zap.L().Error("[GIN Broken Pipe]", append(genZapFields(c), zap.Any("error", err), zap.String("request", string(httpRequest)))...)

	// 如果连接已经断开, 不在给客户端写入状态信息
	if ginErr := c.Error(err); ginErr != nil {
		zap.L().Error("[GIN Broken Pipe]", append(genZapFields(c), zap.Any("error", ginErr), zap.String("request", string(httpRequest)))...)
	}

	// 中断请求
	c.Abort()
}

// logRecovery 根据 stack 参数决定是否记录堆栈信息
func logRecovery(err error, httpRequest []byte, stack bool) {
	// 统一组装日志字段
	fields := []zap.Field{
		zap.Any("error", err),
		zap.String("request", string(httpRequest)),
	}

	// 如果需要堆栈信息, 则添加堆栈字段
	if stack {
		fields = append(fields, zap.String("stack", string(debug.Stack())))
	}

	zap.L().Error("[GIN Recovery from panic]", fields...)
}

// genZapFields 按照统一格式和顺序生成 zap 日志字段
func genZapFields(c *gin.Context, start ...time.Time) []zap.Field {
	// 检查请求 ID, 并获取相关字段
	fields, _, err := res.CheckRequestID(c)
	if err != nil {
		return nil
	}

	// 保证日志字段顺序一致
	fields = append(fields,
		zap.Int("status", c.Writer.Status()), // 状态码
	)

	// 当且仅当传入 start 参数时, 计算耗时
	if len(start) == 1 {
		cost := time.Since(start[0])
		fields = append(fields, zap.Duration("cost", cost)) // 耗时
	}

	// 保证日志字段顺序一致
	fields = append(fields,
		zap.String("ip", c.ClientIP()),         // 客户端 IP
		zap.String("method", c.Request.Method), // 请求方法
		zap.String("path", c.Request.URL.Path), // 请求路径
	)

	// 判断是否有 query 消息
	if c.Request.URL.RawQuery != "" {
		fields = append(fields, zap.String("query", c.Request.URL.RawQuery))
	}

	fields = append(fields, zap.String("user-agent", c.Request.UserAgent()))

	// 错误信息
	errMsg := c.Errors.ByType(gin.ErrorTypePrivate)
	if errMsg.String() != "" {
		fields = append(fields, zap.String("gin.error", errMsg.String()))
	}

	return fields
}
