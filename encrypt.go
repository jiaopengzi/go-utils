//
// FilePath    : go-utils\encrypt.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 加密解密.
//

package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GenerateHashedPassword 生成密码的哈希值
func GenerateHashedPassword(password string, bcryptCost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// ComparePasswords 根据哈希值验证密码是否匹配
func ComparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// EncryptAES 加密函数, 对明文字符串进行加密 返回加密后的字符串
//   - plainText: 需要加密的明文字符串
//   - keyStr: 密钥
//   - ivStr: 初始化向量
func EncryptAES(plainText string, keyStr string, ivStr ...string) (string, error) {
	// 将字符串形式的密钥和初始化向量转换为字节切片
	key := []byte(keyStr)

	// 校验初始化向量
	iv, err := validateIV(ivStr)
	if err != nil {
		return "", err
	}

	// 使用密钥创建一个新的AES密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New("创建新密码块失败：" + err.Error())
	}

	// 为明文添加填充，以满足AES加密的块大小要求
	padding := aes.BlockSize - len(plainText)%aes.BlockSize
	padText := append([]byte(plainText), bytes.Repeat([]byte{byte(padding)}, padding)...)

	// 创建一个用于加密的CBC模式实例，并使用初始化向量进行加密操作

	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(padText))
	mode.CryptBlocks(ciphertext, padText)

	// 将加密后的字节数据转换为Base64编码的字符串
	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)

	// 返回加密后的字符串
	return encodedCiphertext, nil
}

// DecryptAES 解密函数, 对加密过的字符串进行解密, 返回解密后的字符串.
//   - encryptStr: 需要解密的字符串
//   - keyStr: 密钥
//   - ivStr: 初始化向量
func DecryptAES(encryptStr string, keyStr string, ivStr ...string) (string, error) {
	// 将字符串形式的密钥和初始化向量转换为字节切片
	key := []byte(keyStr)

	// 校验初始化向量
	iv, err := validateIV(ivStr)
	if err != nil {
		return "", err
	}

	// 对加密过的字符串进行Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptStr)
	if err != nil {
		return "", errors.New("base64 解码失败：" + err.Error())
	}

	// 使用密钥创建一个新的AES密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New("创建新的密码块失败：" + err.Error())
	}

	// 创建一个用于解密的CBC模式实例，并使用初始化向量进行解密操作
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 获取填充长度，以便将解密后的明文恢复为原始数据
	paddingLength := int(plaintext[len(plaintext)-1])

	// 检查填充长度是否有效
	if paddingLength <= 0 || paddingLength > len(plaintext) {
		return "", errors.New("无效的填充长度")
	}

	// 去除填充部分，获取原始明文
	trimmedPlaintext := plaintext[:len(plaintext)-paddingLength]

	// 返回解密后的字符串
	return string(trimmedPlaintext), nil
}

// validateIV 校验初始化向量,返回初始化向量的字节切片
func validateIV(ivArg []string) ([]byte, error) {
	// 初始化向量
	ivByte := []byte("0000000000000000")

	// 如果没有提供初始化向量，则使用"0000000000000000"作为初始化向量
	if len(ivArg) == 0 {
		return ivByte, nil
	}

	// 如果 长度大于 1 报错
	if len(ivArg) != 1 {
		return nil, errors.New("参数数量只有一个")
	}

	iv := ivArg[0]

	// 检查初始化向量是否具有正确的长度 16 位
	if len(iv) != aes.BlockSize {
		return nil, errors.New("无效的初始化向量长度：需要 16 位数字符串")
	}

	return []byte(iv), nil
}

// PlayKeyEncryptAES2Base64 使用 AES 加密算法用 playKeyKey + iv 加密 playKey 生成加密后的密钥 encryptedPlayKeyBase64 字符串; 再拼接 倒序后的playKeyKey + encryptedPlayKeyBase64 + 倒序后的validateIVStr
//   - playKey: 播放密钥
//   - playKeyKey: 播放密钥的加密密钥
//   - ivStr: 初始化向量(可选)
func PlayKeyEncryptAES2Base64(playKey string, playKeyKey string, ivStr ...string) (string, error) {
	// 校验初始化向量
	iv, err := validateIV(ivStr)
	if err != nil {
		return "", err
	}

	// 将 iv 转换为字符串
	validateIVStr := string(iv)

	// 使用 AES 加密算法用 playKeyKey 加密 playKey 生成加密后的密钥 encryptedPlayKeyBase64 字符串
	encryptedPlayKeyBase64, err := EncryptAES(playKey, playKeyKey, validateIVStr)
	if err != nil {
		return "", err
	}

	// 将 playKeyKey  validateIVStr 逆序排列
	playKeyKey = ReverseString(playKeyKey)
	validateIVStr = ReverseString(validateIVStr)

	return playKeyKey + encryptedPlayKeyBase64 + validateIVStr, nil
}

// PlayKeyDecryptAES2String 使用 AES 解密算法用 encryptKey 解密 playKey 生成解密后的密钥 encryptPlayKey 16进制字符串
func PlayKeyDecryptAES2String(playKeyEncrypt string) (string, error) {
	// 获取 playKeyEncrypt 字符长度
	l := len(playKeyEncrypt)

	// 获取 playKeyKey 从 playKeyEncryptAES2Base64 中从左至右截取 32 长度的字符串并逆序排列
	playKeyKey := ReverseString(playKeyEncrypt[:32])

	// 获取 iv 从 playKeyEncryptAES2Base64 中从右至左截取 16 长度的字符串,并逆序排列
	iv := ReverseString(playKeyEncrypt[l-16:])

	// 获取 encryptedPlayKeyBase64 从 playKeyEncrypt 中从 32 开始到 l-16 的字符串
	encryptedPlayKeyBase64 := playKeyEncrypt[32 : l-16]

	// 使用 AES 解密算法用 encryptKey 解密 playKey 生成解密后的密钥 encryptPlayKey 16进制字符串
	playKey, err := DecryptAES(encryptedPlayKeyBase64, playKeyKey, iv)
	if err != nil {
		return "", err
	}

	return playKey, nil
}

// ReverseString 将字符串逆序排列并返回
func ReverseString(str string) string {
	var result string
	for _, v := range str {
		result = string(v) + result
	}

	return result
}

// GenerateAESKeyAndIV 生成 AES 加密算法的 32 位的 key 和 16 位的 iv
func GenerateAESKeyAndIV() (string, string, error) {
	// 生成 32 位的 key 和 16 位的 iv
	key, err := GenerateHexStr(16)
	if err != nil {
		return "", "", err
	}

	iv, err := GenerateHexStr(8)
	if err != nil {
		return "", "", err
	}

	return key, iv, nil
}

// GenerateHexStr 根据 numEvenBytes(必须是偶数) 生成指定长度的十六进制字符串
func GenerateHexStr(numEvenBytes int) (string, error) {
	data, err := GenerateBinaryData(numEvenBytes)
	if err != nil {
		return "", err
	}

	// 转换为十六进制字符串
	return hex.EncodeToString(data), nil
}

// GenerateBinaryData 根据 numEvenBytes(必须是偶数) 生成指定长度的二进制数据
func GenerateBinaryData(numEvenBytes int) ([]byte, error) {
	if numEvenBytes <= 0 {
		return nil, fmt.Errorf("numBytes must be greater than 0")
	}

	if numEvenBytes%2 != 0 {
		return nil, fmt.Errorf("number of characters must be even")
	}

	_bytes := make([]byte, numEvenBytes)

	_, err := rand.Read(_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %v", err)
	}

	return _bytes, nil
}

// GenerateB64Str 根据 numEvenBytes 生成指定长度的 Base64 字符串
func GenerateB64Str(numEvenBytes int) (string, error) {
	data, err := GenerateBinaryData(numEvenBytes)
	if err != nil {
		return "", err
	}

	// 转换为 Base64 字符串
	return base64.StdEncoding.EncodeToString(data), nil
}
