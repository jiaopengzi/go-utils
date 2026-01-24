//
// FilePath    : go-utils\res\main.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 响应信息
//

// Package res 响应信息
package res

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiaopengzi/go-utils"
	"github.com/jiaopengzi/go-utils/logger"
	"github.com/jiaopengzi/go-utils/rescode"
	"go.uber.org/zap"
)

// 定义在 gin 上下文中的 key
const (
	KeyRequestID = "RequestID" // 请求ID
	KeyUserID    = "UserID"    // 用户ID
	KeyPostID    = "PostID"    // 文章ID
	KeySecret    = "Secret"    // 密钥
)

// enableResponseBody 是否记录响应体到日志
var enableResponseBody bool

// SetEnableResponseBody 设置是否记录响应体到日志
func SetEnableResponseBody(enable bool) {
	enableResponseBody = enable
}

// DocResponse 由于 Swagger 不支持泛型, DocResponse 仅用于 Swagger 文档生成.
type DocResponse struct {
	RequestID string                 `json:"request_id" example:"request_id"` // 请求ID
	Code      rescode.StatusCodeType `json:"code" example:"10000"`            // 业务状态码
	Msg       string                 `json:"msg" example:"Success"`           // 状态码对应信息
	Data      any                    `json:"data" example:"{}"`               // 无数据时为空
}

// Response 返回信息结构体
type Response[D any] struct {
	RequestID string                 `json:"request_id" example:"request_id"` // 请求ID (必选)
	Code      rescode.StatusCodeType `json:"code" example:"10000"`            // 业务状态码 (必选)
	Msg       string                 `json:"msg" example:"Success"`           // 状态码对应信息 (必选)
	Data      D                      `json:"data" example:"{}"`               // 无数据时为空 (可选)
}

// ResPayNotify 返回信息结构体, 用于支付相通知应答
type ResPayNotify struct {
	IsSuccess bool   `json:"is_success"` // 是否成功 (必选)
	Code      string `json:"code"`       // 状态码 (必选)
	Message   string `json:"message"`    // 状态码对应信息 (必选)
}

// MsgResponse 通过 r 响应信息, c gin 上下文, 统一返回信息的格式，并记录响应信息到日志.
func MsgResponse[D any](r *Response[D], c *gin.Context) {
	// 构建日志字段
	fields, requestID, err := CheckRequestID(c)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, &Response[D]{
		RequestID: requestID,
		Code:      r.Code,
		Msg:       r.Code.Msg(),
		Data:      r.Data,
	})

	fields = append(fields, zap.Any("code", r.Code), zap.String("msg", r.Code.Msg()))

	// 如果配置了 enableResponseBody, 并且 Data 不为 nil, 则记录 Data
	if enableResponseBody && !utils.IsInterfaceNil(r.Data) {
		// 创建 data 的副本
		dataCopy, err := utils.DeepCopy(r.Data)
		if err != nil {
			zap.L().Error("dataCopy, err := utils.DeepCopy[*R](data) failed")
			return
		}

		// 移除敏感字段
		logger.MaskSensitiveFields(&dataCopy, logger.SensitiveFields)
		fields = append(fields, zap.Any("data", &dataCopy))
	}

	zap.L().Info("响应信息", fields...)

	c.Abort()
}

// MsgResPayNotify 通过 r 响应信息, c gin 上下文, 统一返回信息的格式，并记录响应信息到日志.
func MsgResPayNotify(r *ResPayNotify, c *gin.Context) {
	// 构建日志字段
	fields, _, err := CheckRequestID(c)
	if err != nil {
		return
	}

	fields = append(fields, zap.Bool("isSuccess", r.IsSuccess))

	// 处理返回信息
	if !r.IsSuccess {
		c.JSON(http.StatusInternalServerError, r)
		c.Abort()
		zap.L().Warn("应答失败", fields...)

		return
	}

	// 成功应答
	c.JSON(http.StatusOK, r)

	zap.L().Info("响应信息-应答成功", fields...)

	c.Abort()
}

// MsgResXMLResponse 通过 r 响应信息, c gin 上下文, 统一返回 XML 格式的信息，并记录响应信息到日志.
func MsgResXMLResponse(xml []byte, c *gin.Context) {
	// 构建日志字段
	fields, _, err := CheckRequestID(c)
	if err != nil {
		return
	}

	// 如果没有 XML 内容, 返回 204 No Content
	if len(xml) == 0 {
		c.Status(http.StatusNoContent)
		c.Abort()
		zap.L().Warn("响应 XML 为空", fields...)

		return
	}

	// 输出 XML
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xml)

	// 如果配置了 EnableResponseBody, 并且 xml 不为空, 则记录 XML
	if enableResponseBody {
		dataCopy, err := utils.DeepCopy(xml)
		if err != nil {
			zap.L().Error("dataCopy, err := utils.DeepCopy(xml) failed")
			return
		}

		// 尝试对数据做掩码处理（对字符串类型无副作用）
		logger.MaskSensitiveFields(&dataCopy, logger.SensitiveFields)

		if s, ok := any(dataCopy).([]byte); ok {
			fields = append(fields, zap.String("xml", string(s)))
		}
	}

	zap.L().Info("响应信息-XML", fields...)

	c.Abort()
}

// CheckRequestID 检查请求ID是否存在, 并返回日志字段.
func CheckRequestID(c *gin.Context) ([]zap.Field, string, error) {
	// 构建日志字段
	fields := []zap.Field{
		zap.String("requestID", c.GetString(KeyRequestID)), // 请求ID
	}

	requestID := c.GetString(KeyRequestID)

	// 没有获取到请求ID
	if requestID == "" {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		zap.L().Error("获取请求ID失败", fields...)

		return nil, "", utils.ErrRequestIDNotFound
	}

	return fields, requestID, nil
}
