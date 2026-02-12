//
// FilePath    : go-utils\email\email_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 邮件工具测试
//

package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// IsEmail 测试
// ---------------------------------------------------------------------------

func TestIsEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// 有效邮箱
		{"标准邮箱", "user@example.com", true},
		{"带子域名邮箱", "user@mail.example.com", true},
		{"带数字域名", "user@123.123.123.123", false}, // IP 无方括号, 期望 false
		{"带方括号IP域名", "user@[123.123.123.123]", true},
		{"带点号用户名", "user.name@example.com", true},
		{"带加号用户名", "user+tag@example.com", true},
		{"带连字符域名", "user@my-domain.com", true},
		{"两个字符TLD", "user@example.co", true},
		{"长TLD", "user@example.museum", true},

		// 无效邮箱
		{"空字符串", "", false},
		{"缺少@符号", "userexample.com", false},
		{"缺少域名", "user@", false},
		{"缺少用户名", "@example.com", false},
		{"双@符号", "user@@example.com", false},
		{"域名无TLD", "user@localhost", false},
		{"有空格", "user @example.com", false},
		{"纯文本", "plaintext", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEmail(tt.input)
			assert.Equal(t, tt.want, got, "IsEmail(%q)", tt.input)
		})
	}
}

// ---------------------------------------------------------------------------
// FilterStrIsEmail 测试
// ---------------------------------------------------------------------------

func TestFilterStrIsEmail(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		delimiter     []string
		wantEmails    []string
		wantNotEmails []string
	}{
		{
			name:          "默认逗号分隔-混合",
			input:         "a@b.com,invalid,c@d.org",
			wantEmails:    []string{"a@b.com", "c@d.org"},
			wantNotEmails: []string{"invalid"},
		},
		{
			name:          "自定义分号分隔",
			input:         "a@b.com;invalid;c@d.org",
			delimiter:     []string{";"},
			wantEmails:    []string{"a@b.com", "c@d.org"},
			wantNotEmails: []string{"invalid"},
		},
		{
			name:          "全部有效",
			input:         "a@b.com,c@d.org",
			wantEmails:    []string{"a@b.com", "c@d.org"},
			wantNotEmails: nil,
		},
		{
			name:          "全部无效",
			input:         "foo,bar",
			wantEmails:    nil,
			wantNotEmails: []string{"foo", "bar"},
		},
		{
			name:          "空字符串",
			input:         "",
			wantEmails:    nil,
			wantNotEmails: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEmails, gotNotEmails := FilterStrIsEmail(tt.input, tt.delimiter...)
			assert.Equal(t, tt.wantEmails, gotEmails)
			assert.Equal(t, tt.wantNotEmails, gotNotEmails)
		})
	}
}

// ---------------------------------------------------------------------------
// GetTemplate 模板缓存测试
// ---------------------------------------------------------------------------

func TestGetTemplate_Cache(t *testing.T) {
	// 创建临时模板文件
	dir := t.TempDir()
	path := dir + "/test.html"

	// 写入模板文件
	err := writeTestFile(path, `<p>Hello {{.Name}}</p>`)
	assert.NoError(t, err)

	// 首次加载
	t1, err := GetTemplate(path)
	assert.NoError(t, err)
	assert.NotNil(t, t1)

	// 第二次加载应返回缓存的同一实例
	t2, err := GetTemplate(path)
	assert.NoError(t, err)
	assert.Same(t, t1, t2, "第二次调用应返回缓存的模板实例")
}

func TestGetTemplate_NotFound(t *testing.T) {
	_, err := GetTemplate("/nonexistent/path/template.html")
	assert.Error(t, err)
}
