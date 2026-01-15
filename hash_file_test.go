//
// FilePath    : go-utils\hash_file_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 单元测试 - 文件哈希相关工具
//

package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateHashByFileContent(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		opts    []SignOptionFunc
		wantErr bool
	}{
		{
			name:    "empty content",
			content: []byte{},
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "simple content",
			content: []byte("hello world"),
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "binary content",
			content: []byte{0x00, 0x01, 0x02, 0x03},
			opts:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.content)
			hash, err := GenerateHashByFileContent(reader, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateHashByFileContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("GenerateHashByFileContent() returned empty hash")
			}
		})
	}
}

func TestGenerateHashByFileContent_Consistency(t *testing.T) {
	content := []byte("test content for consistency check")
	reader1 := bytes.NewReader(content)
	reader2 := bytes.NewReader(content)

	hash1, err := GenerateHashByFileContent(reader1)
	if err != nil {
		t.Fatalf("GenerateHashByFileContent() error = %v", err)
	}

	hash2, err := GenerateHashByFileContent(reader2)
	if err != nil {
		t.Fatalf("GenerateHashByFileContent() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("GenerateHashByFileContent() inconsistent hashes: %s != %s", hash1, hash2)
	}
}

func TestGenerateHashByFilePath(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	content := []byte("test file content")

	err := os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	hash, err := GenerateHashByFilePath(tmpFile)
	if err != nil {
		t.Errorf("GenerateHashByFilePath() error = %v", err)
		return
	}

	if hash == "" {
		t.Error("GenerateHashByFilePath() returned empty hash")
	}
}

func TestGenerateHashByFilePath_NonExistentFile(t *testing.T) {
	_, err := GenerateHashByFilePath("/non/existent/path/file.txt")
	if err == nil {
		t.Error("GenerateHashByFilePath() expected error for non-existent file")
	}
}

func TestGenerateHashByFilePath_Consistency(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	content := []byte("consistency test content")

	err := os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	hash1, err := GenerateHashByFilePath(tmpFile)
	if err != nil {
		t.Fatalf("GenerateHashByFilePath() error = %v", err)
	}

	hash2, err := GenerateHashByFilePath(tmpFile)
	if err != nil {
		t.Fatalf("GenerateHashByFilePath() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("GenerateHashByFilePath() inconsistent hashes: %s != %s", hash1, hash2)
	}
}

func TestGenerateIncrementalHash(t *testing.T) {
	hash, err := GenerateIncrementalHash([]io.Reader{
		bytes.NewReader([]byte("chunk1")),
		bytes.NewReader([]byte("chunk2")),
		bytes.NewReader([]byte("chunk3")),
	})
	if err != nil {
		t.Errorf("GenerateIncrementalHash() error = %v", err)
		return
	}

	if hash == "" {
		t.Error("GenerateIncrementalHash() returned empty hash")
	}
}

func TestGenerateIncrementalHashFromFilePaths(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	file3 := filepath.Join(tmpDir, "file3.txt")

	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)
	os.WriteFile(file3, []byte("content3"), 0644)

	filePaths := []string{file1, file2, file3}

	hash, err := GenerateIncrementalHashFromFilePaths(filePaths)
	if err != nil {
		t.Errorf("GenerateIncrementalHashFromFilePaths() error = %v", err)
		return
	}

	if hash == "" {
		t.Error("GenerateIncrementalHashFromFilePaths() returned empty hash")
	}
}

func TestGenerateIncrementalHashFromFilePaths_NonExistentFile(t *testing.T) {
	filePaths := []string{"/non/existent/file.txt"}

	_, err := GenerateIncrementalHashFromFilePaths(filePaths)
	if err == nil {
		t.Error("GenerateIncrementalHashFromFilePaths() expected error for non-existent file")
	}
}

func TestCheckContentHash(t *testing.T) {
	content := []byte("test content")
	reader := bytes.NewReader(content)

	// First, generate the hash
	expectedHash, err := GenerateHashByFileContent(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	// Test with matching hash
	match, err := CheckContentHash(reader, expectedHash)
	if err != nil {
		t.Errorf("CheckContentHash() error = %v", err)
		return
	}
	if !match {
		t.Error("CheckContentHash() expected match for correct hash")
	}
}

func TestCheckContentHash_Mismatch(t *testing.T) {
	content := []byte("test content")
	reader := bytes.NewReader(content)

	// Test with wrong hash
	match, err := CheckContentHash(reader, "wronghash")
	if err == nil {
		t.Error("CheckContentHash() expected error for mismatched hash")
	}
	if match {
		t.Error("CheckContentHash() should return false for mismatched hash")
	}
}

func TestGenerateHashByStrContent(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		wantErr bool
	}{
		{
			name:    "empty string",
			str:     "",
			wantErr: false,
		},
		{
			name:    "simple string",
			str:     "hello world",
			wantErr: false,
		},
		{
			name:    "unicode string",
			str:     "你好世界",
			wantErr: false,
		},
		{
			name:    "long string",
			str:     string(make([]byte, 10000)),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := GenerateHashByStrContent(tt.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateHashByStrContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("GenerateHashByStrContent() returned empty hash")
			}
		})
	}
}

func TestGenerateHashByStrContent_Consistency(t *testing.T) {
	str := "consistency test string"

	hash1, err := GenerateHashByStrContent(str)
	if err != nil {
		t.Fatalf("GenerateHashByStrContent() error = %v", err)
	}

	hash2, err := GenerateHashByStrContent(str)
	if err != nil {
		t.Fatalf("GenerateHashByStrContent() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("GenerateHashByStrContent() inconsistent hashes: %s != %s", hash1, hash2)
	}
}

func TestHashConsistencyAcrossMethods(t *testing.T) {
	content := "test content for cross-method consistency"

	// Hash using string method
	strHash, err := GenerateHashByStrContent(content)
	if err != nil {
		t.Fatalf("GenerateHashByStrContent() error = %v", err)
	}

	// Hash using bytes.Reader method
	reader := bytes.NewReader([]byte(content))
	contentHash, err := GenerateHashByFileContent(reader)
	if err != nil {
		t.Fatalf("GenerateHashByFileContent() error = %v", err)
	}

	if strHash != contentHash {
		t.Errorf("Hash mismatch between methods: string=%s, content=%s", strHash, contentHash)
	}

	// Hash using file path method
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	err = os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	fileHash, err := GenerateHashByFilePath(tmpFile)
	if err != nil {
		t.Fatalf("GenerateHashByFilePath() error = %v", err)
	}

	if strHash != fileHash {
		t.Errorf("Hash mismatch between methods: string=%s, file=%s", strHash, fileHash)
	}
}
