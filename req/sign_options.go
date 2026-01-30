//
// FilePath    : go-utils\req\sign_options.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : SignOptions 的 With Option 模式
//

package req

import "github.com/gin-gonic/gin"

// SignOption 定义 SignOptions 的可选配置函数类型
type SignOption func(*SignOptions)

// WithCert 设置证书
func WithCert(cert string) SignOption {
	return func(o *SignOptions) {
		o.Cert = cert
	}
}

// WithCertKey 设置证书密钥
func WithCertKey(certKey string) SignOption {
	return func(o *SignOptions) {
		o.CertKey = certKey
	}
}

// WithMaxTimestampDiffSecond 设置最大时间戳差异(秒)
func WithMaxTimestampDiffSecond(maxDiff int64) SignOption {
	return func(o *SignOptions) {
		o.MaxTimestampDiffSecond = maxDiff
	}
}

// BuildSignOptions 构建 SignOptions 实例
// 可选参数 opts 用于配置 SignOptions 的其他字段
func BuildSignOptions(c *gin.Context, opts ...SignOption) SignOptions {
	opt := SignOptions{
		HTTPMethod:    c.Request.Method,
		Path:          c.Request.URL.Path,
		AppID:         c.GetHeader(HeaderAppID),
		TimestampNano: c.GetHeader(HeaderTimestamp),
		Nonce:         c.GetHeader(HeaderNonce),
		Signature:     c.GetHeader(HeaderSignature),
	}

	// 解析查询参数
	opt.QueryParams = ParseQueryParams(c)

	// 应用可选配置
	for _, apply := range opts {
		// 最大时间戳差异默认值为 60 秒
		if opt.MaxTimestampDiffSecond == 0 {
			opt.MaxTimestampDiffSecond = 60
		}

		apply(&opt)
	}

	return opt
}
