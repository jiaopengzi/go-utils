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

	if err := cd.validateRequired(); err != nil {
		return "", err
	}

	cd.applyDefaults(fileNameHashLength)

	layout, err := cd.layoutString()
	if err != nil {
		return "", err
	}

	subDir := cd.buildSubDir(layout)

	if isMkdirLocal {
		if err := createDirIfNeeded(subDir, cd.RootDir, cd.Permission); err != nil {
			return "", err
		}
	}

	return subDir, nil
}

func (cd *ChunkSubDir) validateRequired() error {
	if cd.CreatedAt.IsZero() {
		return fmt.Errorf("date is required")
	}
	if cd.ID == 0 {
		return fmt.Errorf("id is required")
	}
	if cd.HashStr == "" {
		return fmt.Errorf("hash string is required")
	}
	if cd.PartNumbers == 0 {
		return fmt.Errorf("part numbers is required")
	}
	return nil
}

func (cd *ChunkSubDir) applyDefaults(fileNameHashLength int) {
	if cd.Delimiter == "" {
		cd.Delimiter = "-"
	}

	if cd.HashStrPrefixLength == 0 {
		cd.HashStrPrefixLength = fileNameHashLength
	} else if cd.HashStrPrefixLength > len(cd.HashStr) {
		cd.HashStrPrefixLength = len(cd.HashStr)
	}

	if cd.DateTimeLayout == "" {
		cd.DateTimeLayout = DateTimeLayoutDay
	}

	if cd.Permission == 0 {
		cd.Permission = 0700
	}
}

func (cd *ChunkSubDir) layoutString() (string, error) {
	switch cd.DateTimeLayout {
	case DateTimeLayoutYear:
		return "2006", nil
	case DateTimeLayoutMonth:
		return "2006/01", nil
	case DateTimeLayoutDay:
		return "2006/01/02", nil
	case DateTimeLayoutHour:
		return "2006/01/02/15", nil
	case DateTimeLayoutMin:
		return "2006/01/02/15/04", nil
	case DateTimeLayoutSec:
		return "2006/01/02/15/04/05", nil
	default:
		return "", fmt.Errorf("invalid prefix: %s", cd.DateTimeLayout)
	}
}

func (cd *ChunkSubDir) buildSubDir(layout string) string {
	if cd.PartNumbers == 1 {
		return cd.CreatedAt.Format(layout)
	}
	return fmt.Sprintf("%s/%d%s%s", cd.CreatedAt.Format(layout), cd.ID, cd.Delimiter, cd.HashStr[:cd.HashStrPrefixLength])
}

func createDirIfNeeded(subDir, rootDir string, perm os.FileMode) error {
	dir := subDir
	if rootDir != "" {
		dir = filepath.Join(rootDir, dir)
	}
	dirPath := filepath.FromSlash(dir)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, perm); err != nil {
			return err
		}
	}
	return nil
}
