//
// FilePath    : go-utils\request.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 请求相关的工具
//

package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// HTTP 相关的签名头部常量a
const (
	HeaderAppID              = "X-App-Id"              // 应用 ID
	HeaderTimestamp          = "X-Timestamp"           // 时间戳
	HeaderNonce              = "X-Nonce"               // 随机数
	HeaderSignatureAlgorithm = "X-Signature-Algorithm" // 签名算法
	HeaderSignature          = "X-Signature"           // 签名
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
	AppSecret              string            // 应用密钥
	Signature              string            // 签名
	MaxTimestampDiffSecond int64             // 最大时间戳差异(秒)
}

// GetSignData 获取用于签名的数据字符串
func (o *SignOptions) GetSignData() string {
	// 使用 \n 连接各个字段
	return o.HTTPMethod + "\n" +
		o.Path + "\n" +
		BuildQueryString(o.QueryParams) + "\n" +
		o.AppID + "\n" +
		o.TimestampNano + "\n" +
		o.Nonce + "\n" +
		o.EncryptedData
}

// Sign 生成请求签名并设置到 SignOptions.Signature 字段
func (o *SignOptions) Sign(opts ...SignOptionFunc) {
	signData := o.GetSignData()
	// 计算签名
	o.Signature = SignData(signData, o.AppSecret, opts...)
}

// Verify 验证请求签名是否有效
func (o *SignOptions) Verify(opts ...SignOptionFunc) error {
	// 校验时间戳差异
	if !o.VerifyTimestamp() {
		return ErrTimestampDiffExceeded
	}

	// 获取用于签名的数据
	signData := o.GetSignData()

	// 计算并验证签名
	return VerifySignature(signData, o.AppSecret, o.Signature, opts...)
}

// VerifyTimestamp 验证时间戳是否在允许范围内
func (o *SignOptions) VerifyTimestamp() bool {
	// 如果没有最大时间戳差异限制，则设置为默认值 60 秒
	if o.MaxTimestampDiffSecond <= 0 {
		o.MaxTimestampDiffSecond = 60
	}

	// 将时间戳字符串转换为整数
	timestampInt := StrToInt64(o.TimestampNano)
	if timestampInt == 0 {
		return false
	}

	// 获取当前时间戳
	currentTimestamp := GetCurrentTimestampNano()

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

// BuildSignOptions 构建签名选项
func BuildSignOptions(c *gin.Context, appSecret string) SignOptions {
	opt := SignOptions{
		HTTPMethod:    c.Request.Method,
		Path:          c.Request.URL.Path,
		AppID:         c.GetHeader(HeaderAppID),
		TimestampNano: c.GetHeader(HeaderTimestamp),
		Nonce:         c.GetHeader(HeaderNonce),
		AppSecret:     appSecret,
		Signature:     c.GetHeader(HeaderSignature),
	}

	// 解析查询参数
	opt.QueryParams = ParseQueryParams(c)

	return opt
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

// BindEncryptedBody 绑定加密的请求体数据
func BindEncryptedBody(c *gin.Context, opt *SignOptions) error {
	// 检查请求体是否为空
	if c.Request.ContentLength <= 0 {
		return nil
	}

	var encryptedData EncryptedData
	if err := c.ShouldBindJSON(&encryptedData); err != nil {
		return err
	}

	opt.EncryptedData = encryptedData.CipherText

	return nil
}

// DecryptAndSetBody 解密数据并写回请求体
func DecryptAndSetBody[T any](c *gin.Context, encryptedData, appSecret string) error {
	// 无加密数据时跳过
	if encryptedData == "" {
		return nil
	}

	var decryptedData T

	if err := DecryptJSON(encryptedData, appSecret, &decryptedData); err != nil {
		return err
	}

	// 将解密后的数据序列化为 JSON
	jsonData, err := json.Marshal(decryptedData)
	if err != nil {
		return err
	}

	// 将解密后的数据写回 body 供后续 ShouldBindJSON 使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonData))

	return nil
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
