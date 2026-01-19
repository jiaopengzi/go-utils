//
// FilePath    : go-utils\model\base_model.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 基础模型
//

package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModelNoPrimarykey 基础模型没有主键(需要自己定义主键)
type BaseModelNoPrimarykey struct {
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp(6) with time zone;comment:创建时间" json:"created_at" example:"2025-12-29T09:19:51+08:00"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamp(6) with time zone;comment:更新时间" json:"updated_at" example:"2025-12-29T09:19:51+08:00"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(6) with time zone;index;comment:删除时间" json:"deleted_at"`
}

// BaseModel 基础模型 ID 自增 增加索引
type BaseModel struct {
	// ID 自增主键
	ID uint64 `gorm:"column:id;type:bigint;primarykey;index;autoIncrement:true;not null;comment:自增ID" json:"id,string" example:"1234567890"`
	BaseModelNoPrimarykey
}

// BaseModelSonyflake 基础模型 ID 使用雪花算法生成
type BaseModelSonyflake struct {
	// ID 使用雪花算法生成，不使用 gorm.Model 中的 ID 字段，不使用自增
	ID uint64 `gorm:"column:id;primarykey;type:bigint;index;autoIncrement:false;not null;comment:雪花算法生成的ID" json:"id,string" example:"1234567890123456789"`
	BaseModelNoPrimarykey
}

// TableName 实现 Tabler 接口的方法, 没有实际意义, 为了 GetColumnName 方法能够获取到表名
func (BaseModelNoPrimarykey) TableName() string {
	return "base_model_no_primarykey"
}
