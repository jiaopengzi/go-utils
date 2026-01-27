//
// FilePath    : go-utils\crypto_json.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 加密和解密 JSON 数据
//

package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// EncryptJSON 使用 Base64 的密钥 key, 加密任意结构体 data, 并返回 Base64 编码的密文和 nonce.
// 如果 data 为 nil, 则返回空密文和有效的 nonce.
func EncryptJSON(data any, key string) (string, string, error) {
	// 将 Base64 密钥解码为字节切片.
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", "", fmt.Errorf("base64 decode key: %w", err)
	}

	// 如果 data 为 nil, 生成 nonce 并返回空密文.
	if IsInterfaceNil(data) {
		nonce, errN := GenerateGCMNonce()
		if errN != nil {
			return "", "", fmt.Errorf("generate nonce: %w", errN)
		}

		return "", base64.StdEncoding.EncodeToString(nonce), nil
	}

	// 序列化为 JSON.
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", "", fmt.Errorf("marshal json: %w", err)
	}

	// 使用 GCM 加密, nonce 前置到密文中.
	ciphertext, nonce, err := GCMEncryptWithNoncePrepended(keyBytes, jsonBytes)
	if err != nil {
		return "", "", fmt.Errorf("encrypt: %w", err)
	}

	// 返回 Base64 编码的密文和 nonce.
	return base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(nonce), nil
}

// DecryptJSON 使用 Base64 的密钥 key, 解密 Base64 编码的密文 encryptedB64 到目标结构 dst, dst 应为指针类型.
// 如果 encryptedB64 为空字符串, 则直接返回 nil.
func DecryptJSON(encryptedB64 string, key string, dst any) error {
	// 如果 dst 不是指针类型, 返回错误.
	if !IsPointer(dst) {
		return fmt.Errorf("dst %T must be a pointer", dst)
	}

	// 如果密文为空, 直接返回 nil.
	if encryptedB64 == "" {
		return nil
	}

	// 将 Base64 编码的密文解码为字节切片.
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return fmt.Errorf("base64 decode: %w", err)
	}

	// 将 Base64 密钥解码为字节切片.
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("base64 decode key: %w", err)
	}

	// 使用 GCM 解密, nonce 从密文前缀提取.
	jsonBytes, err := GCMDecryptWithNoncePrepended(keyBytes, ciphertext)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	// 反序列化到目标结构.
	return json.Unmarshal(jsonBytes, dst)
}
