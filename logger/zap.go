//
// FilePath    : go-utils\logger\zap.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : zap 日志工具
//

package logger

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"

	"github.com/jiaopengzi/go-utils"

	"go.uber.org/zap"
)

// MyZapConfig 自定义 zap 配置
type MyZapConfig struct {
	ZapConfig     zap.Config `json:"zapConfig" yaml:"zapConfig"`          // 继承 zap 配置
	BufferSize    int        `json:"bufferSize"  yaml:"bufferSize"`       // 缓冲区大小 单位:字节
	FlushInterval int        `json:"flushInterval"  yaml:"flushInterval"` // 刷新间隔 单位:秒
}

// lumberjackSink 重写 lumberjackSink
type lumberjackSink struct {
	lumberjack.Logger
}

// Sync 让 lumberjackSink 实现 WriteSyncer 接口, 以便注册 sink，无需实际实现 Sync 方法,直接返回 nil 即可.
func (l *lumberjackSink) Sync() error {
	return nil
}

// useDevMode 是否使用生产模式, 默认 false, 即使用生产模式
var useDevMode bool

// SetUseDevMode 设置是否使用开发模式
func SetUseDevMode(dev bool) {
	useDevMode = dev
}

// Init 初始化日志
func Init(confFilePath string) error {
	// 读取配置文件
	cfg, err := cfgUnmarshal(confFilePath)
	if err != nil {
		panic(err)
	}

	// 注册 sink
	if err = registerSink(); err != nil {
		panic(err)
	}

	// 将 info 和 error 分拆不同的文件

	// // 创建 info 级别日志的 Core
	// infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
	// 	return lvl < zapcore.ErrorLevel && lvl >= cfg.ZapConfig.Level.Level() // 应用配置文件中的日志级别
	// })

	// infoCore, err := newCoreCustom(cfg, &infoLevel, cfg.ZapConfig.OutputPaths...)
	// if err != nil {
	// 	panic(err)
	// }

	// // 创建 error 级别日志的 Core
	// errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
	// 	return lvl >= zapcore.ErrorLevel && lvl >= cfg.ZapConfig.Level.Level() // 应用配置文件中的日志级别
	// })

	// errorCore, err := newCoreCustom(cfg, &errorLevel, cfg.ZapConfig.ErrorOutputPaths...)
	// if err != nil {
	// 	panic(err)
	// }

	// // 合并 Core
	// core := zapcore.NewTee(infoCore, errorCore)

	// 不分拆日志文件
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= cfg.ZapConfig.Level.Level() // 应用配置文件中的日志级别
	})

	core, err := newCoreCustom(cfg, &infoLevel, cfg.ZapConfig.OutputPaths...)
	if err != nil {
		panic(err)
	}

	// 创建 logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// 替换全局 logger
	zap.ReplaceGlobals(logger)

	return nil
}

// cfgUnmarshal 解析配置文件, 返回 MyZapConfig 配置信息
func cfgUnmarshal(confFilePath string) (*MyZapConfig, error) {
	// 读取配置文件
	var cfg MyZapConfig

	// 默认配置文件路径
	if confFilePath == "" {
		confFilePath = "./config/log_zap.yaml"
	}

	yamlFile, err := os.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}

	// 解析配置文件
	if err = yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// registerSink  注册 sink, 用于 zap.Config 中的 outputPaths 和 errorOutputPaths
func registerSink() error {
	// 注册 lumberjack sink
	return zap.RegisterSink("lumberjack", func(u *url.URL) (zap.Sink, error) {
		// - "lumberjack://localhost/logs/app.log?max-size=100&max-backups=100&max-age=180&compress=true"
		filename := utils.ParsePath(u, "./logs/app.log")

		// 如果是不是相对路径，加上当前目录 .
		if !utils.IsRelativePath(filename) {
			filename = fmt.Sprintf(".%s", filename)
		}

		// 创建日志目录
		if err := utils.CreateDir(filepath.Dir(filename), 0644); err != nil {
			return nil, err
		}

		// 解析参数
		q := u.Query()

		// 初始化 lumberjackSink
		l := &lumberjackSink{
			Logger: lumberjack.Logger{
				Filename:   filename,
				MaxSize:    utils.ParseNumber(q.Get("max-size"), 100),    // 默认日志文件最大 100M, 可在 log_zap.yaml 中配置
				MaxBackups: utils.ParseNumber(q.Get("max-backups"), 100), // 默认日志文件存储最大 100 个 相当于存储10G,可在 log_zap.yaml 中配置
				MaxAge:     utils.ParseNumber(q.Get("max-age"), 180),     // 默认日志文件存储最大 180 天 一般法规要求6个月,可在 log_zap.yaml 中配置
				Compress:   utils.ParseBool(q.Get("compress"), true),     // 默认日志文件是否压缩,可在 log_zap.yaml 中配置
				LocalTime:  utils.ParseBool(q.Get("local-time"), true),   // 默认日志文件是否使用本地时间,可在 log_zap.yaml 中配置
			},
		}

		return l, nil
	})
}

// newCoreCustom 创建并返回自定义 zap Core, 用于输出日志
//   - cfg: MyZapConfig 配置信息
//   - levelEnablerFunc: 日志级别
//   - paths: 输出路径, 可以有多个
func newCoreCustom(cfg *MyZapConfig, levelEnablerFunc *zap.LevelEnablerFunc, paths ...string) (zapcore.Core, error) {
	// 根据配置文件中的 encoding 字段选择 encoder
	encoder, err := newEncoder(cfg.ZapConfig.EncoderConfig, cfg.ZapConfig.Encoding)
	if err != nil {
		return nil, fmt.Errorf("创建 encoder 失败: %w", err)
	}

	// 创建写入器
	writeSyncer, _, err := zap.Open(paths...)
	if err != nil {
		panic(err)
	}

	// 设置默认值
	if cfg.BufferSize <= 0 || cfg.FlushInterval <= 0 {
		cfg.BufferSize = 256 * 1024 // 256 kB
		cfg.FlushInterval = 5       // 5秒
	}

	// 创建一个缓冲写入器
	buffer := zapcore.AddSync(&zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(writeSyncer),                   // 写入器
		Size:          cfg.BufferSize,                                 // 缓冲区大小
		FlushInterval: time.Duration(cfg.FlushInterval) * time.Second, // 刷新间隔
	})

	// 开发环境使用 writeSyncer, 能够实时看到日志, 不会有延迟.
	if useDevMode {
		return zapcore.NewCore(encoder, writeSyncer, levelEnablerFunc), nil
	}

	// 生产环境使用 buffer, 会有一定的延迟, 但是能够提高性能.
	return zapcore.NewCore(encoder, buffer, levelEnablerFunc), nil
}

// newEncoder 创建并返回 encoder
//   - cfg: zapcore.EncoderConfig 配置信息
//   - encoding: 编码方式 json 或 console
func newEncoder(cfg zapcore.EncoderConfig, encoding string) (zapcore.Encoder, error) {
	switch encoding {
	case "json":
		return zapcore.NewJSONEncoder(cfg), nil // 创建 json encoder
	case "console":
		return zapcore.NewConsoleEncoder(cfg), nil // 创建 console encoder
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}
}
