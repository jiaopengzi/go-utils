//
// FilePath    : go-utils\req\sign.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 请求签名相关的工具
//

package req

import (
	"encoding/base64"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jiaopengzi/cert/core"
	"github.com/jiaopengzi/go-utils"
)

// HTTP 相关的签名头部常量a
const (
	HeaderAppID     = "X-App-Id"    // 应用 ID
	HeaderTimestamp = "X-Timestamp" // 时间戳
	HeaderNonce     = "X-Nonce"     // 随机数
	HeaderSignature = "X-Signature" // 签名
)

// EncryptedData 加密数据结构
type EncryptedData struct {
	CipherText string `json:"cipher_text" example:"cipher_text"` // 密文
}

// SignOptions 应用签名选项
type SignOptions struct {
	HTTPMethod             string            // GET, POST 等
	Path                   string            // 请求 Path
	QueryParams            map[string]string // 查询参数
	AppID                  string            // 应用 ID
	TimestampNano          string            // 时间戳(纳秒)
	Nonce                  string            // 随机数
	EncryptedData          string            // 加密数据(请求体)
	Cert                   string            // 证书
	CertKey                string            // 证书密钥
	Signature              string            // 签名
	MaxTimestampDiffSecond int64             // 最大时间戳差异(秒)
}

// GetSignData 获取用于签名的数据字符串
func (o *SignOptions) GetSignData() []byte {
	// 使用 \n 连接各个字段
	return []byte(
		o.HTTPMethod + "\n" +
			o.Path + "\n" +
			BuildQueryString(o.QueryParams) + "\n" +
			o.AppID + "\n" +
			o.TimestampNano + "\n" +
			o.Nonce + "\n" +
			o.EncryptedData,
	)
}

// Sign 生成请求签名并设置到 SignOptions.Signature 字段
func (o *SignOptions) Sign() error {
	// 获取用于签名的数据
	signData := o.GetSignData()

	// 计算签名
	signature, err := core.SignData(o.Cert, signData)
	if err != nil {
		return err
	}

	// 使用 URL 安全的 Base64 编码并去掉填充字符
	o.Signature = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(signature)

	return nil
}

// Verify 验证请求签名是否有效
func (o *SignOptions) Verify() error {
	// 校验时间戳差异
	if !o.VerifyTimestamp() {
		return utils.ErrTimestampDiffExceeded
	}

	// 获取用于签名的数据
	signData := o.GetSignData()

	// 解码签名
	s, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(o.Signature)
	if err != nil {
		return err
	}

	// 计算并验证签名
	return core.VerifySignature(o.Cert, signData, s)
}

// VerifyTimestamp 验证时间戳是否在允许范围内
func (o *SignOptions) VerifyTimestamp() bool {
	// 将时间戳字符串转换为整数
	timestampInt := utils.StrToInt64(o.TimestampNano)
	if timestampInt == 0 {
		return false
	}

	// 获取当前时间戳
	currentTimestamp := utils.GetCurrentTimestampNano()

	// 计算差异的绝对值
	diff := currentTimestamp - timestampInt
	if diff < 0 {
		diff = -diff
	}

	// 将最大允许差异转换为纳秒
	maxDiffNano := o.MaxTimestampDiffSecond * 1e9

	// 验证差异是否在允许范围内
	return diff <= maxDiffNano
}

// HasAppID 在 Gin 框架中检查请求头中是否包含 AppID, 返回是否存在的布尔值和 AppID 字符串
func HasAppIDWithGin(c *gin.Context) (bool, string) {
	appID := c.GetHeader(HeaderAppID)
	return appID != "", appID
}

// ParseQueryParams 解析查询参数
func ParseQueryParams(c *gin.Context) map[string]string {
	query := c.Request.URL.Query()
	if len(query) == 0 {
		return nil
	}

	queryParams := make(map[string]string, len(query))

	for key, values := range query {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	return queryParams
}

// 构建排序后的查询字符串
func BuildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	// 1. 提取所有 key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	// 2. 对 key 进行排序
	slices.Sort(keys)

	// 3. 按排序后的 key 顺序拼接
	var builder strings.Builder

	for i, k := range keys {
		if i > 0 {
			builder.WriteByte('&')
		}

		builder.WriteString(k)
		builder.WriteByte('=')
		builder.WriteString(params[k])
	}

	return builder.String()
}
