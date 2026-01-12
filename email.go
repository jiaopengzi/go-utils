//
// FilePath    : go-utils\email.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 邮件工具
//

package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/jordan-wright/email"
)

// EmailConfig 邮件配置结构体
type EmailConfig struct {
	Host         string `mapstructure:"host" json:"host" binding:"required" example:"localhost"`                     // 邮箱服务器地址
	UserName     string `mapstructure:"user_name" json:"user_name"  binding:"required" example:"jiaopengzi"`         // 发件人用户名
	From         string `mapstructure:"from" json:"from" binding:"required,email" example:"name@example.com"`        // 发件人邮箱账号
	Password     string `mapstructure:"password" json:"password"  binding:"required" example:"Pwd12356"`             // 发件人邮箱密码
	Port         int    `mapstructure:"port" json:"port"  binding:"required,gte=1,lte=65535" example:"587"`          // 邮箱服务器端口
	MaxSendCount int    `mapstructure:"max_send_count" json:"max_send_count" binding:"required,gte=1" example:"100"` // 单次发送邮件的最大数量
	SendInterval int    `mapstructure:"send_interval" json:"send_interval" binding:"required,gte=1" example:"60"`    // 邮件发送间隔时间 单位：秒 默认 60 秒
	PoolSize     int    `mapstructure:"pool_size" json:"pool_size" binding:"required,gte=1" example:"4"`             // 连接池大小(工作协程数)
}

// EmailMeta 邮件元数据结构体定义
type EmailMeta[T any] struct {
	Config       *EmailConfig
	To           []string // 收件人邮箱账号
	Subject      string   // 邮件主题
	TemplatePath string   // 邮件模板路径
	Data         *T       // 邮件模板数据
}

// SendEmail 发送邮件：
// - 解析并渲染模板为邮件正文
// - 构建 `email.Email` 对象(处理 To/Bcc)
// - 获取合适的连接池(STARTTLS 或 隐式 TLS)并通过池发送以复用连接
func SendEmail[T any](meta *EmailMeta[T]) error {
	// 加载指定模板
	t, err := template.ParseFiles(meta.TemplatePath)
	if err != nil {
		return err
	}

	// 将数据渲染到模板
	var body bytes.Buffer
	if err = t.Execute(&body, meta.Data); err != nil {
		return err
	}

	// 构建邮件
	e := email.NewEmail()
	e.From = meta.Config.From
	e.To = meta.To
	e.Subject = meta.Subject
	e.HTML = body.Bytes() // 使用HTML

	// 如果有 Bcc 则添加到 Bcc 保证邮件发送时不会显示所有收件人
	if len(meta.To) > 0 {
		// 只显示第一位收件人
		e.To = []string{meta.To[0]}
		// 其余收件人全部放入 Bcc
		if len(meta.To) > 1 {
			e.Bcc = meta.To[1:]
		}
	}

	// 使用连接池复用 SMTP/TLS 连接, 避免每次发送都进行握手开销
	p, err := GetPool(&SMTPPool{
		Addr: fmt.Sprintf("%s:%d", meta.Config.Host, meta.Config.Port),
		Auth: smtp.PlainAuth("", meta.Config.From, meta.Config.Password, meta.Config.Host),
		Size: meta.Config.PoolSize,
	})
	if err != nil {
		return err
	}

	// 发送邮件, 超时时间 30 秒
	return p.Send(e, 30*time.Second)
}

// IsEmail 判断字符串是否为有效的Email地址
func IsEmail(s string) bool {
	if s == "" {
		return false
	}

	// 使用正则表达式验证Email格式
	const emailRegex = `^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`

	matched, err := regexp.MatchString(emailRegex, s)
	if err != nil {
		return false
	}

	return matched
}

// FilterStrIsEmail 过滤字符串中的Email地址
// 返回两个切片：一个包含有效的Email地址, 另一个包含无效的
func FilterStrIsEmail(str string, delimiter ...string) ([]string, []string) {
	// 初始化两个切片, 用于存储有效和无效的Email地址
	var (
		emails    []string
		notEmails []string
	)

	// 如果没有提供分隔符, 则默认为逗号
	if len(delimiter) == 0 {
		delimiter = append(delimiter, ",")
	}

	// 使用提供的分隔符分割字符串
	sep := delimiter[0]

	// 首先将字符串按照逗号分割成字符串切片
	strList := strings.Split(str, sep)

	if len(strList) == 0 {
		return emails, notEmails
	}

	for _, item := range strList {
		if IsEmail(item) {
			emails = append(emails, item)
		} else {
			notEmails = append(notEmails, item)
		}
	}

	return emails, notEmails
}
