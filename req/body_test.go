//
// FilePath    : go-utils\req\body_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 请求体相关功能单元测试
//

package req

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindEncryptedBody(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		contentLength  int64
		wantErr        bool
		wantCipherText string
	}{
		{
			name:           "正常绑定",
			body:           `{"cipher_text":"encrypted_data_here"}`,
			contentLength:  -1, // 自动计算
			wantErr:        false,
			wantCipherText: "encrypted_data_here",
		},
		{
			name:           "空请求体",
			body:           "",
			contentLength:  0,
			wantErr:        false,
			wantCipherText: "",
		},
		{
			name:           "无效 JSON",
			body:           `{invalid json}`,
			contentLength:  -1,
			wantErr:        true,
			wantCipherText: "",
		},
		{
			name:           "缺少 cipher_text 字段",
			body:           `{"other_field":"value"}`,
			contentLength:  -1,
			wantErr:        false,
			wantCipherText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = bytes.NewBufferString(tt.body)
			} else {
				bodyReader = bytes.NewBuffer(nil)
			}

			c.Request = httptest.NewRequest(http.MethodPost, "/", bodyReader)
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.contentLength >= 0 {
				c.Request.ContentLength = tt.contentLength
			} else {
				c.Request.ContentLength = int64(len(tt.body))
			}

			opt := &SignOptions{}
			err := BindEncryptedBody(c, opt)

			if (err != nil) != tt.wantErr {
				t.Errorf("BindEncryptedBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if opt.EncryptedData != tt.wantCipherText {
				t.Errorf("BindEncryptedBody() EncryptedData = %q, want %q", opt.EncryptedData, tt.wantCipherText)
			}
		})
	}
}

func TestDecryptAndSetBody(t *testing.T) {
	type TestPayload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	for _, cg := range allCertGenerators {
		t.Run(cg.name, func(t *testing.T) {
			certPEM, keyPEM := cg.generate(t)

			// 准备原始数据
			orig := TestPayload{Name: "alice", Age: 30}

			// 加密数据
			encryptedData, _, err := EncryptJSON(orig, certPEM)
			if err != nil {
				t.Fatalf("EncryptJSON error: %v", err)
			}

			// 创建 gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

			// 设置 SignOptions
			opt := &SignOptions{
				EncryptedData: encryptedData,
				CertKey:       keyPEM,
			}

			// 解密并设置 body
			err = DecryptAndSetBody[TestPayload](c, opt)
			if err != nil {
				t.Fatalf("DecryptAndSetBody error: %v", err)
			}

			// 读取解密后的 body
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				t.Fatalf("ReadAll error: %v", err)
			}

			var got TestPayload
			if err := json.Unmarshal(bodyBytes, &got); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			if got.Name != orig.Name || got.Age != orig.Age {
				t.Errorf("DecryptAndSetBody() got = %+v, want %+v", got, orig)
			}
		})
	}
}

func TestDecryptAndSetBody_EmptyEncryptedData(t *testing.T) {
	type TestPayload struct {
		Name string `json:"name"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	opt := &SignOptions{
		EncryptedData: "", // 空加密数据
		CertKey:       "some-key",
	}

	err := DecryptAndSetBody[TestPayload](c, opt)
	if err != nil {
		t.Errorf("DecryptAndSetBody() expected nil error for empty encrypted data, got: %v", err)
	}
}

func TestDecryptAndSetBody_InvalidEncryptedData(t *testing.T) {
	type TestPayload struct {
		Name string `json:"name"`
	}

	for _, cg := range allCertGenerators {
		t.Run(cg.name, func(t *testing.T) {
			_, keyPEM := cg.generate(t)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

			opt := &SignOptions{
				EncryptedData: "invalid-encrypted-data",
				CertKey:       keyPEM,
			}

			err := DecryptAndSetBody[TestPayload](c, opt)
			if err == nil {
				t.Errorf("DecryptAndSetBody() expected error for invalid encrypted data")
			}
		})
	}
}
