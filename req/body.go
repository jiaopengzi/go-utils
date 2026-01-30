//
// FilePath    : go-utils\req\body.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 请求体相关的工具
//

package req

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

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
func DecryptAndSetBody[T any](c *gin.Context, opt *SignOptions) error {
	// 无加密数据时跳过
	if opt.EncryptedData == "" {
		return nil
	}

	var decryptedData T

	if err := DecryptJSON(opt.EncryptedData, opt.CertKey, &decryptedData); err != nil {
		return err
	}

	// 将解密后的数据序列化为 JSON, 需要使用指针
	jsonData, err := json.Marshal(&decryptedData)
	if err != nil {
		return err
	}

	// 将解密后的数据写回 body 供后续 ShouldBindJSON 使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonData))

	return nil
}
