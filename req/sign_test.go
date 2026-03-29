//
// FilePath    : go-utils\req\sign_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 请求签名相关功能单元测试
//

package req

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestBuildQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name:   "空参数",
			params: nil,
			want:   "",
		},
		{
			name:   "空 map",
			params: map[string]string{},
			want:   "",
		},
		{
			name:   "单个参数",
			params: map[string]string{"key": "value"},
			want:   "key=value",
		},
		{
			name:   "多个参数按字母排序",
			params: map[string]string{"z": "3", "a": "1", "m": "2"},
			want:   "a=1&m=2&z=3",
		},
		{
			name:   "包含特殊字符",
			params: map[string]string{"name": "alice", "msg": "hello world"},
			want:   "msg=hello world&name=alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildQueryString(tt.params)
			if got != tt.want {
				t.Errorf("BuildQueryString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSignOptions_GetSignData(t *testing.T) {
	tests := []struct {
		name string
		opt  SignOptions
		want string
	}{
		{
			name: "不含版本号",
			opt: SignOptions{
				HTTPMethod:    "POST",
				Path:          "/api/test",
				QueryParams:   map[string]string{"b": "2", "a": "1"},
				AppID:         "app123",
				TimestampNano: "1234567890",
				Nonce:         "nonce123",
				EncryptedData: "encrypted_body",
			},
			want: "POST\n/api/test\na=1&b=2\napp123\n1234567890\nnonce123\nencrypted_body",
		},
		{
			name: "含版本号",
			opt: SignOptions{
				HTTPMethod:    "POST",
				Path:          "/api/test",
				QueryParams:   map[string]string{"b": "2", "a": "1"},
				AppID:         "app123",
				TimestampNano: "1234567890",
				Nonce:         "nonce123",
				EncryptedData: "encrypted_body",
				Version:       "v1.2.3",
			},
			want: "POST\n/api/test\na=1&b=2\napp123\n1234567890\nnonce123\nencrypted_body\nv1.2.3",
		},
		{
			name: "空字段不含版本号",
			opt: SignOptions{
				HTTPMethod:    "GET",
				Path:          "/",
				QueryParams:   nil,
				AppID:         "",
				TimestampNano: "",
				Nonce:         "",
				EncryptedData: "",
			},
			want: "GET\n/\n\n\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.opt.GetSignData())
			if got != tt.want {
				t.Errorf("GetSignData() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSignOptions_Version(t *testing.T) {
	tests := []struct {
		name          string
		versionHeader string
		withVersion   bool
		wantVersion   string
	}{
		{
			name:          "默认不读取版本头",
			versionHeader: "v1.2.3",
			withVersion:   false,
			wantVersion:   "",
		},
		{
			name:          "WithVersion 读取版本头",
			versionHeader: "v1.2.3",
			withVersion:   true,
			wantVersion:   "v1.2.3",
		},
		{
			name:          "WithVersion 但头部为空",
			versionHeader: "",
			withVersion:   true,
			wantVersion:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.versionHeader != "" {
				c.Request.Header.Set(HeaderBlogServerVersion, tt.versionHeader)
			}

			var opts []SignOption
			if tt.withVersion {
				opts = append(opts, WithVersion(c))
			}

			got := BuildSignOptions(c, opts...)
			if got.Version != tt.wantVersion {
				t.Errorf("BuildSignOptions() Version = %q, want %q", got.Version, tt.wantVersion)
			}
		})
	}
}

func TestSignOptions_VerifyTimestamp(t *testing.T) {
	tests := []struct {
		name                   string
		timestampNano          string
		maxTimestampDiffSecond int64
		want                   bool
	}{
		{
			name:                   "空时间戳",
			timestampNano:          "",
			maxTimestampDiffSecond: 60,
			want:                   false,
		},
		{
			name:                   "无效时间戳",
			timestampNano:          "invalid",
			maxTimestampDiffSecond: 60,
			want:                   false,
		},
		{
			name:                   "零时间戳",
			timestampNano:          "0",
			maxTimestampDiffSecond: 60,
			want:                   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := SignOptions{
				TimestampNano:          tt.timestampNano,
				MaxTimestampDiffSecond: tt.maxTimestampDiffSecond,
			}
			got := opt.VerifyTimestamp()
			if got != tt.want {
				t.Errorf("VerifyTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAppIDWithGin(t *testing.T) {
	tests := []struct {
		name      string
		headerVal string
		wantHas   bool
		wantID    string
	}{
		{
			name:      "有 AppID",
			headerVal: "app123",
			wantHas:   true,
			wantID:    "app123",
		},
		{
			name:      "无 AppID",
			headerVal: "",
			wantHas:   false,
			wantID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerVal != "" {
				c.Request.Header.Set(HeaderAppID, tt.headerVal)
			}

			gotHas, gotID := HasAppIDWithGin(c)
			if gotHas != tt.wantHas {
				t.Errorf("HasAppIDWithGin() has = %v, want %v", gotHas, tt.wantHas)
			}
			if gotID != tt.wantID {
				t.Errorf("HasAppIDWithGin() id = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestParseQueryParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want map[string]string
	}{
		{
			name: "无查询参数",
			url:  "/api/test",
			want: nil,
		},
		{
			name: "单个查询参数",
			url:  "/api/test?key=value",
			want: map[string]string{"key": "value"},
		},
		{
			name: "多个查询参数",
			url:  "/api/test?a=1&b=2&c=3",
			want: map[string]string{"a": "1", "b": "2", "c": "3"},
		},
		{
			name: "重复参数取第一个值",
			url:  "/api/test?key=first&key=second",
			want: map[string]string{"key": "first"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tt.url, nil)

			got := ParseQueryParams(c)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ParseQueryParams() = %v, want nil", got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseQueryParams() len = %d, want %d", len(got), len(tt.want))
				return
			}

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ParseQueryParams()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
