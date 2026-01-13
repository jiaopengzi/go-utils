//
// FilePath    : go-utils\encrypt_json.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 加密和解密 JSON 数据
//

package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// EncryptJSON 使用 Base64 的密钥 key, 加密任意结构体 data, 并返回 Base64 编码的密文和 nonce
func EncryptJSON(data any, key string) (string, string, error) {
	// 序列化为 JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", "", fmt.Errorf("marshal json: %w", err)
	}

	// 将 Base64 密钥解码为字节切片
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", "", fmt.Errorf("base64 decode key: %w", err)
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", "", fmt.Errorf("new cipher: %w", err)
	}

	// 使用 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("new gcm: %w", err)
	}

	// 生成随机 nonce (GCM 推荐 12 字节)
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", fmt.Errorf("read nonce: %w", err)
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, jsonBytes, nil)

	// 返回 Base64 编码的密文和 nonce
	return base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(nonce), nil
}

// DecryptJSON 使用 Base64 的密钥 key, 解密 Base64 编码的密文 encryptedB64 到目标结构 dst, dst 应为指针类型.
func DecryptJSON(encryptedB64 string, key string, dst any) error {
	// 如果 dst 不是指针类型，返回错误
	if !IsPointer(dst) {
		return fmt.Errorf("dst %T must be a pointer", dst)
	}

	// 将 Base64 编码的密文解码为字节切片
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return fmt.Errorf("base64 decode: %w", err)
	}

	// 将 Base64 密钥解码为字节切片
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("base64 decode key: %w", err)
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return fmt.Errorf("new cipher: %w", err)
	}

	// 使用 GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("new gcm: %w", err)
	}

	// 提取 nonce 和密文
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	jsonBytes, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	// 反序列化到目标结构
	return json.Unmarshal(jsonBytes, dst)
}
