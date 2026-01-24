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
func RegisterAllTasks() error {
	for _, r := range registrars {
		if err := r(); err != nil {
			return fmt.Errorf("注册定时任务失败: %w", err)
		}
	}

	return nil
}

// Init 初始化定时任务
func Init() error {
	// 注册所有任务
	if err := RegisterAllTasks(); err != nil {
		return err
	}

	// 创建任务管理器
	manager := NewTaskManager()

	for _, task := range Tasks {
		// 定时任务的cron表达式配置不能为空
		if task.Spec == "" {
			return fmt.Errorf("定时任务的cron表达式配置不能为空，任务名称：%s", string(task.Name))
		}

		err := manager.AddTask(task)
		if err != nil {
			return fmt.Errorf("添加任务 %s 失败: %w", string(task.Name), err)
		}
	}

	// 启动任务管理器
	manager.Start()

	return nil
}
