//
// FilePath    : go-utils\hash.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 哈希相关工具
//

package utils

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

// HashAlgorithm 哈希类型
type HashAlgorithm string

// 定义哈希类型全局常量
const (
	SHA256 HashAlgorithm = "SHA-256"
	SHA384 HashAlgorithm = "SHA-384"
	SHA512 HashAlgorithm = "SHA-512"
)

// HAOption 签名选项
type HAOption struct {
	Algorithm HashAlgorithm
}

// HAOptionFunc 签名选项函数
type HAOptionFunc func(*HAOption)

// WithAlgorithm 设置哈希算法
func WithAlgorithm(alg HashAlgorithm) HAOptionFunc {
	return func(o *HAOption) {
		o.Algorithm = alg
	}
}

// getHashFunc 根据算法类型返回对应的哈希函数
func getHashFunc(alg HashAlgorithm) func() hash.Hash {
	switch alg {
	case SHA384:
		return sha512.New384
	case SHA512:
		return sha512.New
	default:
		return sha256.New
	}
}

// GenerateHasher 生成哈希对象, 可通过 WithAlgorithm 选项指定哈希算法
func GenerateHasher(opts ...HAOptionFunc) hash.Hash {
	opt := &HAOption{
		Algorithm: SHA256, // 默认使用 SHA256
	}

	for _, fn := range opts {
		fn(opt)
	}

	return getHashFunc(opt.Algorithm)()
}
