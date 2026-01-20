//
// FilePath    : go-utils\middleware\gin\request_id.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 请求 ID 中间件
//

package mwgin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jiaopengzi/go-utils/res"
)

// AddRequestID 添加请求 ID 中间件.
func AddRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 将当前请求的 RequestID 信息保存到请求的上下文 c 上
		requestID := uuid.NewString()
		c.Set(res.KeyRequestID, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}
