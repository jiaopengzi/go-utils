//
// FilePath    : go-utils\signature.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 签名和验签
//

package utils

import (
	"crypto/hmac"
	"encoding/base64"
)

// SignData 使用 HMAC 对 data 进行签名, secret 作为密钥
// 默认使用 HMAC-SHA256, 可通过 WithAlgorithm 选项指定其他算法
// 返回 Base64 编码的签名字符串(URL 安全, 无填充)
func SignData(data string, secret string, opts ...SignOptionFunc) string {
	opt := &SignOption{
		Algorithm: SHA256, // 默认使用 SHA256
	}
	for _, fn := range opts {
		fn(opt)
	}

	key := []byte(secret)
	h := hmac.New(getHashFunc(opt.Algorithm), key)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	// 使用 URLEncoding 且不带填充(=), 适合放在 HTTP Header 或 URL
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(signature)
}

// VerifySignature 验证数据的签名是否有效
// 注意: 验签时需要使用与签名时相同的算法选项
func VerifySignature(data string, secret string, signature string, opts ...SignOptionFunc) error {
	expectedSignature := SignData(data, secret, opts...)

	// 比较签名
	if expectedSignature == signature {
		return nil
	}

	return ErrInvalidSignature
}
