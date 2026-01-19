//
// FilePath    : billing-center\cron\core.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2025 by jiaopengzi, All Rights Reserved.
// Description : 定时任务管理器
//

package cron

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Name string

// Task 单独的任务结构体
type Task struct {
	ID         cron.EntryID // 任务ID(由cron生成)
	Name       Name         // 名称(唯一标识)
	StartTime  time.Time    // 开始时间
	ExpireTime time.Time    // 过期时间
	Spec       string       // 定时任务表达式(为空表示仅执行一次)
	Action     func() error // 执行函数
}

// TaskManager 管理任务的添加、删除和更新
type TaskManager struct {
	cron      *cron.Cron
	tasks     map[string]*Task
	taskMutex sync.Mutex // 互斥锁，保护任务列表的并发访问
}

// NewTaskManager 创建一个新的任务管理器
func NewTaskManager() *TaskManager {
	return &TaskManager{
		// 如果不需要秒级别的任务可去掉 WithSeconds
		cron:  cron.New(cron.WithSeconds()),
		tasks: make(map[string]*Task),
	}
}

// AddTask 添加任务
func (tm *TaskManager) AddTask(task *Task) error {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()

	// 检查任务名称是否已存在
	if _, exists := tm.tasks[string(task.Name)]; exists {
		return fmt.Errorf("任务 %s 已存在, 无法添加", task.Name)
	}

	// 如果过期时间已过，不执行
	if !task.ExpireTime.IsZero() && time.Now().After(task.ExpireTime) {
		return fmt.Errorf("任务 %s 已经过期, 不再执行", task.Name)
	}

	// 如 StartTime 未指定, 默认立即开始
	if task.StartTime.IsZero() {
		task.StartTime = time.Now()
	}

	// 根据是否有 Spec 来判定是一次性任务, 还是周期性任务
	if task.Spec == "" {
		return tm.addOneTimeTask(task)
	} else {
		return tm.addRecurringTask(task)
	}
}

// addOneTimeTask 添加一次性任务
func (tm *TaskManager) addOneTimeTask(task *Task) error {
	// 根据 StartTime 生成一个仅执行一次的 cron 表达式
	singleSpec := buildOneTimeSpec(task.StartTime)
	return tm.createTaskExecutor(task, singleSpec, true)
}

// addRecurringTask 添加周期性任务
func (tm *TaskManager) addRecurringTask(task *Task) error {
	return tm.createTaskExecutor(task, task.Spec, false)
}

// createTaskExecutor 工厂函数，用于生成任务执行和添加逻辑
//   - task: 任务对象
//   - spec: cron 表达式
//   - isOneTime: 是否为一次性任务
func (tm *TaskManager) createTaskExecutor(task *Task, spec string, isOneTime bool) error {
	id, err := tm.cron.AddFunc(spec, func() {
		// 检查是否过期
		if !task.ExpireTime.IsZero() && time.Now().After(task.ExpireTime) {
			if err := tm.RemoveTask(string(task.Name)); err != nil {
				zap.L().Error("移除过期任务失败", zap.String("任务名", string(task.Name)), zap.Error(err))
			}

			zap.L().Info("任务已过期，停止执行", zap.String("任务名", string(task.Name)))

			return
		}

		// 执行任务
		if err := task.Action(); err != nil {
			msg := fmt.Sprintf("任务 %s 执行失败，错误信息: %v", task.Name, err)
			zap.L().Error(msg)
		}

		// 如果是一次性任务，执行完成后移除
		if isOneTime {
			if err := tm.RemoveTask(string(task.Name)); err != nil {
				zap.L().Error("移除一次性任务失败", zap.String("任务名", string(task.Name)), zap.Error(err))
			}

			zap.L().Info("一次性任务已执行完毕，停止执行", zap.String("任务名", string(task.Name)))
		}
	})

	if err != nil {
		return fmt.Errorf("添加任务 %s 失败: %v", task.Name, err)
	}

	task.ID = id
	tm.tasks[string(task.Name)] = task

	return nil
}

// buildOneTimeSpec 根据给定时间生成一个仅执行一次的 cron 表达式
// 注意需要 futureTime >= 当前时间，否则生成的表达式无效
func buildOneTimeSpec(futureTime time.Time) string {
	now := time.Now()

	// 如果 futureTime 在当前时间之前, 则将其设置为当前时间 + 1 秒
	if futureTime.Before(now) {
		futureTime = now.Add(1 * time.Second)
	}

	// 构造一个 "秒 分 时 日 月  周" 格式的表达式
	return fmt.Sprintf("%d %d %d %d %d *",
		futureTime.Second(),
		futureTime.Minute(),
		futureTime.Hour(),
		futureTime.Day(),
		futureTime.Month(),
	)
}

// RemoveTask 移除任务
func (tm *TaskManager) RemoveTask(name string) error {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()

	task, exists := tm.tasks[name]
	if !exists {
		return fmt.Errorf("任务 %s 不存在，无法移除", name)
	}

	tm.cron.Remove(task.ID)
	delete(tm.tasks, name)

	return nil
}

// UpdateTask 更新任务
func (tm *TaskManager) UpdateTask(task *Task) error {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()

	existingTask, exists := tm.tasks[string(task.Name)]
	if !exists {
		return fmt.Errorf("任务 %s 不存在，无法更新", task.Name)
	}

	// 如果过期时间已过, 不执行
	if !task.ExpireTime.IsZero() && time.Now().After(task.ExpireTime) {
		return fmt.Errorf("任务 %s 已过期，无法更新", task.Name)
	}

	// 移除旧的
	tm.cron.Remove(existingTask.ID)
	delete(tm.tasks, string(existingTask.Name))

	// 添加新任务
	return tm.AddTask(task)
}

// Start 启动任务管理器
func (tm *TaskManager) Start() {
	tm.cron.Start()
}

// Stop 停止任务管理器，同时清理所有已注册任务
func (tm *TaskManager) Stop() {
	tm.cron.Stop()

	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()

	for name, task := range tm.tasks {
		tm.cron.Remove(task.ID)
		delete(tm.tasks, name)
	}

	zap.L().Info("所有任务已停止")
}
