//
// FilePath    : go-utils\dir_file.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 目录和文件工具
//

package utils

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// GetFileName 根据文件路径 path 获取文件名
func GetFileName(path string) string {
	return filepath.Base(path)
}

// GetFileNameNoExt 根据文件路径 path 获取不带扩展名的文件名
func GetFileNameNoExt(path string) string {
	return strings.TrimSuffix(GetFileName(path), filepath.Ext(path))
}

// CreateDir 根据路径 path 创建文件夹, 如果文件夹不存在则创建
func CreateDir(path string, perm os.FileMode) error {
	// 判断路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 不存在则创建文件夹
		if err = os.MkdirAll(path, perm); err != nil {
			return err
		}
	}

	return nil
}

// InitDir 初始化文件夹 path, 如果文件夹不存在则创建, 如果文件夹存在则删除后再创建
func InitDir(path string, perm os.FileMode) error {
	// 判断路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 不存在则创建文件夹
		if err := os.MkdirAll(path, perm); err != nil {
			return err
		}
	} else {
		// 存在则删除文件夹
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		// 重新创建文件夹
		if err := os.MkdirAll(path, perm); err != nil {
			return err
		}
	}

	return nil
}

// IsRelativePath 判断 path 是否为相对路径
func IsRelativePath(path string) bool {
	switch {
	case strings.HasPrefix(path, "/"):
		return false
	case strings.HasPrefix(path, "./"), strings.HasPrefix(path, "../"):
		return true
	default:
		return false
	}
}

// FlattenDirectoryStructure 将目录结构 dir 扁平化, 使用 delimiter 作为分隔符,默认为"-";比如 FlattenDirectoryStructure("a/b/c") => a-b-c
func FlattenDirectoryStructure(dir string, delimiter ...string) string {
	// 如果没有传入分隔符，则使用默认的 -
	if len(delimiter) == 0 {
		delimiter = append(delimiter, "-")
	}

	// 将目录中的 / 替换为 -
	return strings.ReplaceAll(dir, "/", delimiter[0])
}

// RemoveDir 根据 path 删除文件夹
func RemoveDir(path string) error {
	// 删除文件夹
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

// CheckIfDirIsEmpty 检查目录是否为空
func CheckIfDirIsEmpty(dir string) (bool, error) {
	// 打开目录
	file, err := os.Open(dir)
	if err != nil {
		return false, err
	}

	// 使用 defer 关闭
	defer func(f *os.File) {
		if errClose := f.Close(); errClose != nil {
			zap.L().Error("关闭文件错误", zap.Error(errClose))
		}
	}(file)

	// 读取目录
	_, err = file.Readdir(1)
	if err == io.EOF {
		return true, nil // 空目录
	}

	if err != nil {
		return false, err // 错误
	}

	return false, nil // 非空目录
}

// RemoveDirAndRecursiveRemoveEmptyParents 删除目录并向上递归删除空目录,直到遇到 stopAtDir
func RemoveDirAndRecursiveRemoveEmptyParents(dir string, stopAtDir string) error {
	// 判断 stopAtDir 是否有 / 或者 \ 结尾，有的话去掉
	if strings.HasSuffix(stopAtDir, "/") || strings.HasSuffix(stopAtDir, "\\") {
		stopAtDir = stopAtDir[:len(stopAtDir)-1]
	}

	// 删除目录
	err := RemoveDir(dir)
	if err != nil {
		return err
	}

	// 获取父目录
	parentDir := filepath.Dir(dir)

	// 如果父目录等于停止目录则返回
	if parentDir == stopAtDir {
		return nil
	}

	// 检查目录是否为空
	isEmpty, err := CheckIfDirIsEmpty(parentDir)
	if err != nil {
		return err
	}

	// 如果为空则递归删除
	if isEmpty {
		return RemoveDirAndRecursiveRemoveEmptyParents(parentDir, stopAtDir)
	}

	// 遇到非空目录则直接返回
	return nil
}

// IsDirExists 判断目录 path 是否存在
func IsDirExists(path string) bool {
	_, err := os.Stat(path) // 获取文件信息
	if err != nil {
		return os.IsExist(err) // 判断是否存在
	}

	return true
}

// IsFileExists 判断文件 path 是否存在
func IsFileExists(path string) bool {
	_, err := os.Stat(path) // 获取文件信息
	if err != nil {
		return os.IsExist(err) // 判断是否存在
	}

	return true
}

// WriteFile 将 data 写入到 path 文件中, 如果文件不存在则创建, 如果文件存在则覆盖
func WriteFile(path string, data []byte, perm os.FileMode) error {
	// 打开文件 不存在则创建 , 存在则覆盖 , 读写权限
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	// 使用 defer 关闭
	defer func(f *os.File) {
		if errClose := f.Close(); errClose != nil {
			zap.L().Error("关闭文件错误", zap.Error(errClose))
		}
	}(file)

	// 创建一个带缓冲的写入器
	writer := bufio.NewWriter(file)

	// 写入数据
	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	// 确保所有缓冲数据都写入底层文件
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

// AppendToFile 将 data 追加到 path 文件中, 如果文件不存在则创建, 如果文件存在则将给定内容追加到指定文件
func AppendToFile(path string, data []byte, perm os.FileMode) error {
	// 以追加模式打开文件
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, perm)
	if err != nil {
		return err
	}

	// 使用 defer 关闭
	defer func(f *os.File) {
		if errClose := f.Close(); errClose != nil {
			zap.L().Error("关闭文件错误", zap.Error(errClose))
		}
	}(file)

	// 创建一个带缓冲的写入器
	writer := bufio.NewWriter(file)

	// 写入数据
	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	// 确保所有缓冲数据都写入底层文件
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

// DeleteFile 删除文件 path
func DeleteFile(path string) error {
	// 如果文件不存在则直接返回
	if !IsFileExists(path) {
		return nil
	}

	//	删除文件
	return os.Remove(path)
}

// DeleteFileRemoveDirRecursiveRemoveEmptyParents 删除文件并递归删除空目录,直到遇到 stopAtDir
func DeleteFileRemoveDirRecursiveRemoveEmptyParents(path string, stopAtDir string) error {
	// 删除文件
	err := DeleteFile(path)
	if err != nil {
		return err
	}

	// 获取父目录
	parentDir := filepath.Dir(path)

	// 检查目录是否为空
	isEmpty, err := CheckIfDirIsEmpty(parentDir)
	if err != nil {
		return err
	}

	// 如果为空则递归删除
	if isEmpty {
		return RemoveDirAndRecursiveRemoveEmptyParents(parentDir, stopAtDir)
	}

	return nil
}

// ReadFile 读取文件 path 的内容, 返回文件数据和可能的错误
func ReadFile(path string) ([]byte, error) {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// 使用 defer 关闭
	defer func(f *os.File) {
		if errClose := f.Close(); errClose != nil {
			zap.L().Error("关闭文件错误", zap.Error(errClose))
		}
	}(file)

	// 创建一个带缓冲的读取器
	reader := bufio.NewReader(file)

	// 读取数据
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ReadFileToString 读取文件 filePath 到字符串
func ReadFileToString(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	// 使用 defer 关闭
	defer func(f *os.File) {
		if errClose := f.Close(); errClose != nil {
			zap.L().Error("关闭文件错误", zap.Error(errClose))
		}
	}(file)

	var content strings.Builder

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text() + "\n")
	}

	if err = scanner.Err(); err != nil {
		return "", err
	}

	return content.String(), nil
}

// IsFileName 判断 fileName 是否为合法的文件名
func IsFileName(fileName string) bool {
	// 文件名不能包含以下字符: / \ : * ? " < > |  需要包含扩展名
	return !strings.ContainsAny(fileName, "/\\:*?\"<>|") && strings.Contains(fileName, ".")
}
