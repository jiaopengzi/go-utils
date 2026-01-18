//
// FilePath    : go-utils\date_time.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 时间日期工具
//

package utils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// StrToSQLNullTime 将字符串时间转换为 sql.NullTime
func StrToSQLNullTime(str string) sql.NullTime {
	if str == "" {
		return sql.NullTime{}
	}

	t, err := time.Parse(time.RFC3339, str) // "2006-01-02T15:04:05Z07:00" 解析格式要对齐
	if err != nil {
		zap.L().Error("StrToSQLNullTime failed", zap.Error(err))
	}

	return sql.NullTime{Time: t, Valid: true}
}

// GetDisableExpiresAtSeconds 获取禁用过期时间, 返回禁用时间 单位秒 0 未被禁用
func GetDisableExpiresAtSeconds(disableExpiresAt any) uint64 {
	var expires sql.NullTime // 禁用过期时间

	// 断言
	if _, ok := disableExpiresAt.(string); ok {
		// 判断 disableExpiresAt 类型 如果是字符串就转成 sql.NullTime
		str, ok := disableExpiresAt.(string)
		if !ok {
			return 0
		}

		expires = StrToSQLNullTime(str)
	} else if nt, ok := disableExpiresAt.(*sql.NullTime); ok {
		// 判断 disableExpiresAt 类型 如果是 sql.NullTime 就直接赋值
		if nt != nil {
			expires = *nt
		}
	}

	// 获取当前服务器时间
	currentTime := time.Now()

	// 如果 expires 不为空，且大于当前时间，则禁用
	if expires.Valid && expires.Time.After(currentTime) {
		duration := uint64(expires.Time.Sub(currentTime).Seconds()) // 计算禁用时间 单位秒
		return duration                                             // 用户被禁用
	}

	return 0 // 用户未被禁用
}

// GetFutureTime 获取未来时间, 返回未来时间, 参数为秒
func GetFutureTime(seconds int64) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}

// YM 日期时间格式化接口
type YM interface {
	// 生成横线连接的日期别名 YYYY-MM
	GenDateSlug() string
}

// YearMonth 用于传递年份和月份的参数
type YearMonth struct {
	Year  int
	Month int
}

// GenDateSlug 生成日期别名
func (ym YearMonth) GenDateSlug() string {
	return fmt.Sprintf("%04d-%02d", ym.Year, ym.Month)
}

type YearMonthLastModTime struct {
	YearMonth
	LastModTime time.Time
}

// GenDateSlug 生成日期别名
func (ym YearMonthLastModTime) GenDateSlug() string {
	return fmt.Sprintf("%04d-%02d", ym.Year, ym.Month)
}

// YearMonthSelect 选择年月区间, 月不填写为参数年整年，年月都不填写为 当前日期整年;返回开始时间和结束时间(比较大小的时候为左闭右开)
func YearMonthSelect(params *YearMonth) (startDate, endDate time.Time) {
	currentYear := time.Now().Year()

	// 检查是否提供了年份，如果没有提供，使用当前年份
	year := currentYear
	if params.Year != 0 {
		year = params.Year
	}

	if params.Month == 0 {
		// 如果没有提供月份，默认查询整个年份
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		// 如果提供了月份，查询指定月份
		startDate = time.Date(year, time.Month(params.Month), 1, 0, 0, 0, 0, time.UTC)

		if params.Month == 12 {
			endDate = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
		} else {
			endDate = time.Date(year, time.Month(params.Month+1), 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return
}

// ValidateYearMonthFormat 验证年月格式
func ValidateYearMonthFormat(ym string) (bool, *YearMonth) {
	if len(ym) != 7 || ym[4] != '-' {
		return false, nil
	}

	// 将年月拆分为年和月转成整数再验证
	var year, month int

	_, err := fmt.Sscanf(ym, "%04d-%02d", &year, &month)
	if err != nil {
		return false, nil
	}

	if year < 1900 || year > 2100 || month < 1 || month > 12 {
		return false, nil
	}

	return true, &YearMonth{Year: year, Month: month}
}

// SleepWithContext 在上下文中等待
func SleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer func() {
		// 保证 timer 被停止，并将可能残留的事件从 timer.C 中 drain 掉
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err() // 上下文取消，返回错误
	case <-timer.C:
		return nil // 等待完成
	}
}

// GetCurrentYear
func GetCurrentYear() int {
	return time.Now().Year()
}

// GetCurrentTimestampNano 获取当前时间戳(纳秒)
func GetCurrentTimestampNano() int64 {
	return time.Now().UnixNano()
}

// GetCurrentTimestampNanoStr 获取当前时间戳字符串(纳秒)
func GetCurrentTimestampNanoStr() string {
	return fmt.Sprintf("%d", GetCurrentTimestampNano())
}
