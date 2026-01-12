//
// FilePath    : go-utils\ip_platform.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : IP地址归属地
//

package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Location IP地址归属地
type Location struct {
	Status     string `json:"status"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
}

// GetIPLocation 根据 ip 地址, 获取 ip 地址归属地 Location
func GetIPLocation(ip string) (*Location, error) {
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			zap.L().Error("关闭响应体失败", zap.Error(err))
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var loc Location

	err = json.Unmarshal(body, &loc)
	if err != nil {
		return nil, err
	}

	return &loc, nil
}

// ClientInfo 日志内容
type ClientInfo struct {
	IP        string // IP
	UserAgent string // 用户浏览器信息
}

// GetClientInfo 从 *gin.Context 中获取客户端信息 ClientInfo
func GetClientInfo(c *gin.Context) ClientInfo {
	return ClientInfo{
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
}
