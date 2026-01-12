//
// FilePath    : go-utils\dir_generator.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 目录生成器
//

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DateTimeLayout 文件夹的日期时间布局
type DateTimeLayout string

// 定义日期时间布局全局常量
const (
	DateTimeLayoutYear  DateTimeLayout = "year"
	DateTimeLayoutMonth DateTimeLayout = "month"
	DateTimeLayoutDay   DateTimeLayout = "day"
	DateTimeLayoutHour  DateTimeLayout = "hour"
	DateTimeLayoutMin   DateTimeLayout = "min"
	DateTimeLayoutSec   DateTimeLayout = "sec"
)

// DirGenerator 文件夹生成器 接口
type DirGenerator interface {
	GenerateDir() (string, error) // 生成文件夹 返回文件夹路径和错误信息
}

// ChunkSubDir 分片存储子目录
type ChunkSubDir struct {
	// 必填项
	CreatedAt   time.Time // 数据库中的创建日期
	ID          uint64    // ID
	HashStr     string    // hash 字符串
	PartNumbers int64     // 分片数量

	// 可选项
	Delimiter           string         // 分隔符
	DateTimeLayout      DateTimeLayout // 日期时间布局
	HashStrPrefixLength int            // 取用 Hash 字符串前缀长度
	RootDir             string         // 根目录
	Permission          os.FileMode    // 文件夹权限
}

// GenerateDirOptions 生成文件夹选项
type GenerateDirOptions struct {
	IsMkdirLocal       bool // 是否创建本地文件夹
	FileNameHashLength int  // 文件名哈希长度
}

// GenerateDir 根据 opts 选项生成文件夹
func (cd *ChunkSubDir) GenerateDir(opts *GenerateDirOptions) (string, error) {
	var (
		isMkdirLocal       bool
		fileNameHashLength int
	)

	if opts == nil {
		isMkdirLocal = false // 默认值
		fileNameHashLength = 8
	} else {
		isMkdirLocal = opts.IsMkdirLocal
		fileNameHashLength = opts.FileNameHashLength
	}

	// 校验 Date ID HashStr 为必填项
	if cd.CreatedAt.IsZero() {
		return "", fmt.Errorf("date is required")
	}

	if cd.ID == 0 {
		return "", fmt.Errorf("id is required")
	}

	if cd.HashStr == "" {
		return "", fmt.Errorf("hash string is required")
	}

	if cd.PartNumbers == 0 {
		return "", fmt.Errorf("part numbers is required")
	}

	// 设置可选项默认值
	if cd.Delimiter == "" {
		cd.Delimiter = "-"
	}

	if cd.HashStrPrefixLength == 0 {
		cd.HashStrPrefixLength = fileNameHashLength
	} else if cd.HashStrPrefixLength > len(cd.HashStr) {
		// 如果设置的 HashStrPrefixLength 大于 HashStr 的长度，则取 HashStr 的长度
		cd.HashStrPrefixLength = len(cd.HashStr)
	}

	if cd.DateTimeLayout == "" {
		// 默认为天
		cd.DateTimeLayout = DateTimeLayoutDay
	}

	if cd.Permission == 0 {
		// 默认权限 0700 rwx 权限分布 rwx------
		cd.Permission = 0700
	}

	var layout string

	switch cd.DateTimeLayout {
	case DateTimeLayoutYear:
		layout = "2006"
	case DateTimeLayoutMonth:
		layout = "2006/01"
	case DateTimeLayoutDay:
		layout = "2006/01/02"
	case DateTimeLayoutHour:
		layout = "2006/01/02/15"
	case DateTimeLayoutMin:
		layout = "2006/01/02/15/04"
	case DateTimeLayoutSec:
		layout = "2006/01/02/15/04/05"
	default:
		return "", fmt.Errorf("invalid prefix: %s", cd.DateTimeLayout)
	}

	// 拼接目录
	var subDir string
	// 如果分片数量为 1 则不创建子目录
	if cd.PartNumbers == 1 {
		subDir = cd.CreatedAt.Format(layout)
	} else {
		subDir = fmt.Sprintf("%s/%d%s%s", cd.CreatedAt.Format(layout), cd.ID, cd.Delimiter, cd.HashStr[:cd.HashStrPrefixLength])
	}

	// 根据 mkdirLocal 判断是否创建目录
	if isMkdirLocal {
		// 如果设置了 RootDir 则拼接 RootDir 默认为当前工作目录
		dir := subDir

		// 如果设置了 RootDir 则拼接 RootDir 默认为当前工作目录
		if cd.RootDir != "" {
			dir = filepath.Join(cd.RootDir, dir)
		}

		// 将路径中的斜杠替换为当前系统的路径分隔符
		dirPath := filepath.FromSlash(dir)

		// 创建目录 如果目录已经存在则不创建
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, cd.Permission)
			if err != nil {
				return "", err
			}
		}
	}

	return subDir, nil
}
