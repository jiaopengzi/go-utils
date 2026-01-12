//
// FilePath    : go-utils\paginate.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 分页工具
//

package utils

import (
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"gorm.io/gorm"
)

// PaginateConfig 分页查询配置
type PaginateConfig struct {
	PageSizes      []int64  // 分页大小
	SourceIncludes []string // 需要返回的字段
	SourceExcludes []string // 需要排除返回的字段
	Highlight      bool     // 是否高亮
}

// PaginateOption 包含 Apply 方法，用来配置 PaginateConfig
type PaginateOption interface {
	Apply(*PaginateConfig)
}

// PageSizesOption 分页大小选项
type PageSizesOption struct {
	PageSizes []int64
}

// Apply 配置 Config
func (o PageSizesOption) Apply(config *PaginateConfig) {
	config.PageSizes = o.PageSizes
}

// WithPageSizes 设置分页大小
func WithPageSizes(pageSizes []int64) PaginateOption {
	return PageSizesOption{PageSizes: pageSizes}
}

// SourceIncludesOption 需要返回的字段
type SourceIncludesOption struct {
	SourceIncludes []string
}

// Apply 配置 Config
func (o SourceIncludesOption) Apply(config *PaginateConfig) {
	config.SourceIncludes = o.SourceIncludes
}

// WithSourceIncludes 设置需要返回的字段
func WithSourceIncludes(sourceIncludes []string) PaginateOption {
	return SourceIncludesOption{SourceIncludes: sourceIncludes}
}

// SourceExcludesOption 需要排除返回的字段
type SourceExcludesOption struct {
	SourceExcludes []string
}

// Apply 配置 Config
func (o SourceExcludesOption) Apply(config *PaginateConfig) {
	config.SourceExcludes = o.SourceExcludes
}

// WithSourceExcludes 设置需要排除返回的字段
func WithSourceExcludes(sourceExcludes []string) PaginateOption {
	return SourceExcludesOption{SourceExcludes: sourceExcludes}
}

// HighlightOption 高亮选项
type HighlightOption struct {
	Highlight bool
}

// Apply 配置 Config
func (o HighlightOption) Apply(config *PaginateConfig) {
	config.Highlight = o.Highlight
}

// WithHighlight 设置高亮
func WithHighlight(flag bool) PaginateOption {
	return HighlightOption{Highlight: flag}
}

// applyPageSizesOption 应用分页选项
func applyPageSizesOption(opts ...PaginateOption) *PaginateConfig {
	cfg := &PaginateConfig{
		PageSizes:      []int64{10, 20, 50, 100}, // 默认分页大小
		SourceIncludes: nil,                      // 默认返回所有字段
	}

	for _, opt := range opts {
		opt.Apply(cfg)
	}

	return cfg
}

// PageBase 分页基础结构体
type PageBase struct {
	Total       int64   `json:"total"`        // 总记录数
	CurrentPage int64   `json:"current_page"` // 当前页
	PageSize    int64   `json:"page_size"`    // 分页大小
	PageCount   int64   `json:"page_count"`   // 总页数
	PageSizes   []int64 `json:"page_sizes"`   // 分页大小列表 10,20,50,100
}

// CalculatePageParams 实现 PageInterface 接口
func (page *PageBase) CalculatePageParams() {
	maxSize := page.PageSizes[0] // 分页上限

	// 循环 page.PageSizes 切片求最大值 限定分页大小
	for _, size := range page.PageSizes {
		if size > maxSize {
			maxSize = size
		}
	}

	// 如果当前页小于等于0 则设置为1
	if page.CurrentPage <= 0 {
		page.CurrentPage = 1
	}

	// 限定每页的行数
	switch {
	case page.PageSize > maxSize:
		page.PageSize = maxSize
	case page.PageSize <= 0:
		page.PageSize = 10 // 如果分页大小小于等于0 默认设置为10
	}

	// 兜底设置分页大小
	if page.PageSize >= 100 {
		page.PageSize = 100
	}

	// 计算总页数
	page.PageCount = (page.Total + page.PageSize - 1) / page.PageSize
}

// Page 通用泛型分页结构体
type Page[T any] struct {
	*PageBase
	Records []T `json:"records"` // 返回的当前页数据记录
}

// SelectPages [T any] 执行分页查询并返回结果
//   - db: 数据库连接
//   - query: 查询语句
//   - opts: 可选参数，用于配置分页大小
func (page *Page[T]) SelectPages(db, query *gorm.DB, opts ...PaginateOption) error {
	var modelT T

	options := applyPageSizesOption(opts...)

	page.PageSizes = options.PageSizes

	// 使用子查询获取总记录数
	if err := db.Table("(?) as subquery", query).Select("COUNT(*)").Count(&page.Total).Error; err != nil {
		return err
	}

	// 如果总记录数为 0, 则返回 nil
	if page.Total == 0 {
		page.Records = nil
		return nil
	}

	page.CalculatePageParams()

	// 分页查询
	return query.Model(&modelT).Scopes(Paginate(page)).Find(&page.Records).Error
}

// Paginate 返回一个gorm.DB作用域函数，用于应用分页查询
func Paginate[T any](page *Page[T]) func(db *gorm.DB) *gorm.DB {
	size := page.PageSize
	offset := int((page.CurrentPage - 1) * size) // 计算偏移量

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(int(size))
	}
}

// PageES elasticsearch 分页结构体, 结果直接是 json.RawMessage 类型,不用指定具体的类型
type PageES struct {
	*PageBase
	Records   []json.RawMessage     `json:"records"`   // 返回的当前页数据记录
	Highlight []map[string][]string `json:"highlight"` // 高亮字段
}

// SelectPagesES 执行分页查询并返回结果
//   - client: Elasticsearch 客户端
//   - index: 索引名称
//   - query: 查询请求
//   - sourceIncludes: 返回的字段
//   - opts: 可选参数，用于配置分页大小
func (page *PageES) SelectPagesES(client *elasticsearch.TypedClient, index string, req *search.Request, ctx context.Context, opts ...PaginateOption) error {
	options := applyPageSizesOption(opts...)

	page.PageSizes = options.PageSizes
	page.CalculatePageParams()

	// 设置分页参数 es 使用的是 int 类型
	size := int(page.PageSize)
	req.Size = &size // 设置分页大小
	from := int((page.CurrentPage - 1) * page.PageSize)
	req.From = &from // 设置偏移量

	var (
		res *search.Response
		err error
	)

	search := client.Search().Index(index).Request(req)

	switch {
	case len(options.SourceIncludes) > 0:
		// 按照设置的字段返回
		res, err = search.SourceIncludes_(options.SourceIncludes...).Do(ctx)
	case len(options.SourceExcludes) > 0:
		// 按照设置的字段排除返回
		res, err = search.SourceExcludes_(options.SourceExcludes...).Do(ctx)
	default:
		// 返回所有字段
		res, err = search.Do(ctx)
	}

	// 执行查询
	if err != nil {
		return err
	}

	page.Total = res.Hits.Total.Value                                // 总记录数
	page.Records = make([]json.RawMessage, len(res.Hits.Hits))       // 初始化记录切片
	page.Highlight = make([]map[string][]string, len(res.Hits.Hits)) // 初始化高亮字段切片

	// 遍历结果集, 组成 json.RawMessage 切片
	for i, hit := range res.Hits.Hits {
		page.Records[i] = hit.Source_

		// 添加高亮结果
		if options.Highlight && hit.Highlight != nil {
			page.Highlight[i] = hit.Highlight
		}
	}

	// 如果没有高亮字段, 则设置为 nil,避免上述初始化的切片是有多个 null
	if !options.Highlight {
		page.Highlight = nil
	}

	// 计算总页数
	page.PageCount = (page.Total + page.PageSize - 1) / page.PageSize

	return nil
}
