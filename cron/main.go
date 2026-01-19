//
// FilePath    : billing-center\cron\main.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2025 by jiaopengzi, All Rights Reserved.
// Description : 定时任务入口
//

// Package cron 定时任务
package cron

import (
	"fmt"

	"go.uber.org/zap"
)

// 定时任务变量
var Tasks []*Task // 存储所有的定时任务

// TaskRegistrar 定义任务注册函数类型
type TaskRegistrar func() error

// 记录所有已注册的任务注册函数
var registrars []TaskRegistrar

// RegisterTask 供各个任务模块调用，注册自己的注册函数
func RegisterTask(r TaskRegistrar) {
	registrars = append(registrars, r)
}

// RegisterAllTasks 执行所有已注册的任务注册函数
func RegisterAllTasks() {
	for _, r := range registrars {
		if err := r(); err != nil {
			zap.L().Error("注册定时任务失败", zap.Error(err))
		}
	}
}

// Init 初始化定时任务
func Init() error {
	// 注册所有任务
	RegisterAllTasks()

	// 创建任务管理器
	manager := NewTaskManager()

	for _, task := range Tasks {
		// 定时任务的cron表达式配置不能为空
		if task.Spec == "" {
			zap.L().Error("定时任务的cron表达式配置不能为空", zap.String("任务名称", string(task.Name)))
			return fmt.Errorf("定时任务的cron表达式配置不能为空，任务名称：%s", string(task.Name))
		}

		err := manager.AddTask(task)
		if err != nil {
			zap.L().Error("添加任务错误", zap.String("任务名称", string(task.Name)), zap.Error(err))
			return err
		}
	}

	// 启动任务管理器
	manager.Start()
	zap.L().Info("定时任务启动成功")

	return nil
}
