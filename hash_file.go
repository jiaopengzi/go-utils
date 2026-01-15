//
// FilePath    : go-utils\hash_file.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 文件哈希相关工具
//

package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// GenerateHashByFileContent 通过文件内容 file 生成哈希值, 可通过 WithAlgorithm 选项指定哈希算法
func GenerateHashByFileContent(file *bytes.Reader, opts ...SignOptionFunc) (string, error) {
	// 根据算法生成哈希对象
	hasher := GenerateHasher(opts...)

	// 将指针移动到文件开头
	_, err := file.Seek(0, 0)
	if err != nil {
		return "", err
	}

	// 复制文件内容到哈希对象
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	// 计算哈希值
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}

// GenerateHashByFilePath 通过 filePath 生成哈希值, 可通过 WithAlgorithm 选项指定哈希算法
func GenerateHashByFilePath(filePath string, opts ...SignOptionFunc) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// 根据算法生成哈希对象
	hasher := GenerateHasher(opts...)

	// 复制文件内容到哈希对象
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	// 计算哈希值
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}

// GenerateIncrementalHash 逐个分片复制内容到哈希对象, 然后增量计算哈希值, 可通过 WithAlgorithm 选项指定哈希算法
func GenerateIncrementalHash(chunks []io.Reader, opts ...SignOptionFunc) (string, error) {
	// 根据算法生成哈希对象
	hasher := GenerateHasher(opts...)

	// 逐个分片复制内容到哈希对象
	for _, chunk := range chunks {
		_, err := io.Copy(hasher, chunk)
		if err != nil {
			return "", err
		}
	}

	// 计算哈希值
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}

// GenerateIncrementalHashFromFilePaths 根据 filePaths 逐个分片复制内容到哈希对象, 然后增量计算哈希值, 可通过 WithAlgorithm 选项指定哈希算法
func GenerateIncrementalHashFromFilePaths(filePaths []string, opts ...SignOptionFunc) (string, error) {
	var readers []io.Reader

	var files []*os.File

	// 打开每个文件并添加到 readers 列表
	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}

		files = append(files, file)
		readers = append(readers, file)
	}

	// 调用 GenerateIncrementalHash 计算哈希值
	hashString, err := GenerateIncrementalHash(readers, opts...)
	if err != nil {
		return "", err
	}

	// 在所有文件都被读取后关闭文件
	for _, file := range files {
		errClose := file.Close()
		if errClose != nil {
			return "", errClose
		}
	}

	return hashString, nil
}

// CheckContentHash 根据文件内容 file 检查哈希值是否匹配目标哈希值 targetHash, 可通过 WithAlgorithm 选项指定哈希算法
func CheckContentHash(file *bytes.Reader, targetHash string, opts ...SignOptionFunc) (bool, error) {
	// 生成哈希值
	hashStr, err := GenerateHashByFileContent(file, opts...)
	if err != nil {
		return false, err
	}

	// 检查哈希值
	if hashStr != targetHash {
		return false, fmt.Errorf("文件哈希值不匹配")
	}

	return true, nil
}

// GenerateHashByStrContent 通过字符串 生成哈希值, 可通过 WithAlgorithm 选项指定哈希算法
func GenerateHashByStrContent(str string, opts ...SignOptionFunc) (string, error) {
	// 根据算法生成哈希对象
	hasher := GenerateHasher(opts...)

	// 将字符串转换为字节切片
	strBytes := []byte(str)

	// 计算哈希值
	_, err := hasher.Write(strBytes)
	if err != nil {
		return "", err
	}

	// 计算哈希值
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	// 返回哈希值
	return hashString, nil
}
