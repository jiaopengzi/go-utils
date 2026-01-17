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
	HeaderAppID              = "X-App-Id"
	HeaderTimestamp          = "X-Timestamp"
	HeaderNonce              = "X-Nonce"
	HeaderSignatureAlgorithm = "X-Signature-Algorithm"
	HeaderSignature          = "X-Signature"
)

// EncryptedData 加密数据结构
type EncryptedData struct {
	CipherText string `json:"cipher_text" example:"cipher_text"` // 密文
}

// SignOptions 应用签名选项
type SignOptions struct {
	HTTPMethod    string            // GET, POST 等
	URL           string            // 请求 URL
	QueryParams   map[string]string // 查询参数
	AppID         string            // 应用 ID
	Timestamp     string            // 时间戳
	Nonce         string            // 随机数
	EncryptedData string            // 加密数据(请求体)
	AppSecret     string            // 应用密钥
	Signature     string            // 签名
}

// GetSignData 获取用于签名的数据字符串
func (o *SignOptions) GetSignData() string {
	// 使用 \n 连接各个字段
	return o.HTTPMethod + "\n" +
		o.URL + "\n" +
		buildQueryString(o.QueryParams) + "\n" +
		o.AppID + "\n" +
		o.Timestamp + "\n" +
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
func (o *SignOptions) Verify(opts ...SignOptionFunc) bool {
	signData := o.GetSignData()
	// 计算并验证签名
	return VerifySignature(signData, o.AppSecret, o.Signature, opts...)
}

// HasAppID 在 Gin 框架中检查请求头中是否包含 AppID, 返回是否存在的布尔值和 AppID 字符串
func HasAppIDWithGin(c *gin.Context) (bool, string) {
	appID := c.GetHeader(HeaderAppID)
	return appID != "", appID
}

// AuthWithGin 从 Gin 框架的请求上下文中提取签名相关信息并进行校验和设置解密后的数据
func AuthWithGin[T any](c *gin.Context, appSecret string) error {
	// 构建签名选项
	var opt SignOptions

	// 填充签名选项字段
	opt.HTTPMethod = c.Request.Method
	opt.URL = c.Request.URL.Path
	opt.AppID = c.GetHeader(HeaderAppID)
	opt.Timestamp = c.GetHeader(HeaderTimestamp)
	opt.Nonce = c.GetHeader(HeaderNonce)
	opt.AppSecret = appSecret
	opt.Signature = c.GetHeader(HeaderSignature)

	// 解析查询参数
	if query := c.Request.URL.Query(); len(query) > 0 {
		queryParams := make(map[string]string, len(query))

		for key, values := range query {
			if len(values) > 0 {
				queryParams[key] = values[0]
			}
		}

		opt.QueryParams = queryParams
	}

	// 读取请求体数据
	var encryptedData EncryptedData
	if err := c.ShouldBindJSON(&encryptedData); err != nil {
		return err
	}

	opt.EncryptedData = encryptedData.CipherText

	// 验证签名
	if !opt.Verify() {
		return ErrInvalidSignature
	}

	// 解密数据
	var decryptedData T

	err := DecryptJSON(opt.EncryptedData, appSecret, &decryptedData)
	if err != nil {
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
func buildQueryString(params map[string]string) string {
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
