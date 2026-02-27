//
// FilePath    : go-utils\logger\es.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : es zap 日志库
//

package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// ZapESLogger 是一个实现 elastictransport.Logger 接口的自定义日志记录器
type ZapESLogger struct {
	L                  *zap.Logger // 使用 zap 接管日志
	EnableRequestBody  bool        // 启用请求体打印
	EnableResponseBody bool        // 启用响应体打印
}

// LogRoundTrip 实现 elastictransport.Logger 接口, 记录 Elasticsearch 请求/响应日志 参考官方源码实现过程
func (z *ZapESLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	query, errUnescape := url.QueryUnescape(req.URL.RawQuery)
	if errUnescape != nil {
		z.L.Error("Failed to unescape query", zap.Error(errUnescape))
		return errUnescape
	}

	if query != "" {
		query = "?" + query
	}

	status, color := z.statusColor(res)

	msg := fmt.Sprintf("[ES Request]%6s \x1b[1;4m%s://%s%s\x1b[0m%s %s%s\x1b[0m \x1b[2m%s\x1b[0m",
		req.Method,
		req.URL.Scheme,
		req.URL.Host,
		req.URL.Path,
		query,
		color,
		status,
		dur.Truncate(time.Millisecond),
	)

	z.L.Debug(msg)

	// 处理请求体
	if z.RequestBodyEnabled() && req != nil && req.Body != nil && req.Body != http.NoBody {
		reqBody, errReq := z.readRequestBody(req)
		if errReq != nil {
			z.L.Error("Failed to read request body", zap.Error(errReq))
			return errReq
		}

		reqBodyMsg := fmt.Sprintf("[ES Request Body] \x1b[2m%s\x1b[0m", string(reqBody))
		z.L.Debug(reqBodyMsg)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	// 处理响应体
	if z.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		resBody, errRes := z.readResponseBody(res)
		if errRes != nil {
			z.L.Error("Failed to read response body", zap.Error(errRes))
			return errRes
		}

		resBodyMsg := fmt.Sprintf("[ES Response Body] \x1b[2m%s\x1b[0m", string(resBody))
		z.L.Debug(resBodyMsg)
		res.Body = io.NopCloser(bytes.NewReader(resBody))
	}

	if err != nil {
		errMsg := fmt.Sprintf("ES Request Error \x1b[31;1m» ERROR \x1b[31m%v\x1b[0m", err)
		z.L.Error(errMsg)
	}

	return nil
}

func (z *ZapESLogger) statusColor(res *http.Response) (string, string) {
	if res != nil {
		status := res.Status
		switch {
		case res.StatusCode > 0 && res.StatusCode < 300:
			return status, "\x1b[32m"
		case res.StatusCode > 299 && res.StatusCode < 500:
			return status, "\x1b[33m"
		case res.StatusCode > 499:
			return status, "\x1b[31m"
		default:
			return "ERROR", "\x1b[31;4m"
		}
	}
	return "ERROR", "\x1b[31;4m"
}

func (z *ZapESLogger) readRequestBody(req *http.Request) ([]byte, error) {
	if req.GetBody != nil {
		b, errGet := req.GetBody()
		if errGet != nil {
			return nil, errGet
		}
		return io.ReadAll(b)
	}
	return io.ReadAll(req.Body)
}

func (z *ZapESLogger) readResponseBody(res *http.Response) ([]byte, error) {
	return io.ReadAll(res.Body)
}

// RequestBodyEnabled 实现 elastictransport.Logger 接口, 启用请求体打印
func (z *ZapESLogger) RequestBodyEnabled() bool {
	return z.EnableRequestBody
}

// ResponseBodyEnabled 实现 elastictransport.Logger 接口, 启用响应体打印
func (z *ZapESLogger) ResponseBodyEnabled() bool {
	return z.EnableResponseBody
}

// func resStatusCode(res *http.Response) int {
// 	if res == nil {
// 		return -1
// 	}

// 	return res.StatusCode
// }
