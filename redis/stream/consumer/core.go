//
// FilePath    : go-utils\redis\stream\consumer\core.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 消费者核心
//

// Package consumer redis stream 消费者核心
package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jiaopengzi/go-utils/redis/stream"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Consumer 消费者接口
type Consumer[T any] interface {
	CreateGroup() error                                               // 创建消息组
	ConsumerNameExists(consumerName string) (bool, error)             // 检查消费者是否存在
	GenerateUniqueConsumerName(k int) (string, error)                 // 生成唯一的消费者名称
	CreateConsumer() error                                            // 创建消费者
	PendingMessage() error                                            // 获取挂起的消息并重新分配
	OnlineMessage() error                                             // 监听在线消息
	RunConsumer() error                                               // 运行消费者
	ProcessMessage(message redis.XMessage) error                      // 处理消息
	AckMessage(messageID string) error                                // 签收消息
	GetConsumersInfo() ([]redis.XInfoConsumer, error)                 // 获取该组所有消费者状态
	GetConsumerInfo(consumerName string) (redis.XInfoConsumer, error) // 获取单个消费者状态
	GetPendingCount() (int64, error)                                  // 获取当前未消费的消息总数(包括所有消费者的待处理消息)
	RemoveConsumer(consumerInfo redis.XInfoConsumer) error            // 移除消费者
}

// MessageStateManager 消息状态管理接口
// 用于管理消息的处理状态和签收状态，实现消费者之间的协调
type MessageStateManager interface {
	// IsProcessing 检查消息是否正在被处理
	// 返回 (正在处理的消费者名称, 是否正在处理)
	IsProcessing(streamName, msgID string) (string, bool)

	// MarkProcessing 标记消息开始处理
	MarkProcessing(streamName, msgID, consumerName string) error

	// ClearProcessing 清除处理标记
	ClearProcessing(streamName, msgID string) error

	// UpdateAckStatus ACK 后更新消息状态
	UpdateAckStatus(streamName, msgID, groupName string, isSuccess bool) error
}

// BaseConsumer 消费者基类
type BaseConsumer[T any] struct {
	StreamName         string                                                 // stream 名称 相同的 stream 名称的消费者将会共享消息
	GroupName          string                                                 // 组名称 相同的组名称的消费者将会共享消息, 但是不同的组名称的消费者不会共享消息。
	Start              string                                                 // 开始位置
	ConsumerName       string                                                 // 消费者名称
	MsgKey             string                                                 // 消息的 key 用于解析消息.
	Ctx                context.Context                                        // context 上下文
	ProcessMessageFunc func(c *BaseConsumer[T], message redis.XMessage) error // 处理消息函数
	Rdb                redis.UniversalClient                                  // Redis 客户端
	StateManager       MessageStateManager                                    // 消息状态管理器
}

// isGroupExistsByError 通过错误信息判断组是否存在, true 存在, false 不存在.
func isGroupExistsByError(err error) (bool, error) {
	if err == nil {
		// 存在
		return true, nil
	} else if errors.Is(err, redis.Nil) {
		// redis.Nil 表示不存在
		return false, err
	}

	// "NOGROUP No such key 'stream:XXX' or consumer group 'group:XXX'" 即该组不存在.
	return false, nil
}

// CreateGroup 实现 Consumer 接口方法, 创建消息组
func (c *BaseConsumer[T]) CreateGroup() error {
	_, err := c.Rdb.XPending(c.Ctx, c.StreamName, c.GroupName).Result() // 检查组是否存在

	// 通过错误信息判断组是否存在
	exists, err := isGroupExistsByError(err)
	if err != nil {
		return err
	}

	// 如果组不存在,则创建组
	if !exists {
		if err = c.Rdb.XGroupCreateMkStream(c.Ctx, c.StreamName, c.GroupName, c.Start).Err(); err != nil {
			return err
		}
	}

	// 如果组存在,则返回 nil
	return nil
}

// ConsumerNameExists 实现 Consumer 接口方法, 检查消费者是否存在; true 存在, false 不存在.
func (c *BaseConsumer[T]) ConsumerNameExists(consumerName string) (bool, error) {
	// 获取消费者状态
	consumers, err := c.GetConsumersInfo()
	if err != nil {
		return false, err
	}

	// 当 consumers 为空时,表示没有消费者
	if len(consumers) == 0 {
		return false, nil
	}

	// 遍历消费者列表
	for _, consumer := range consumers {
		if consumer.Name == consumerName {
			return true, nil
		}
	}

	return false, nil
}

// GenerateUniqueConsumerName 生成唯一的消费者名称,预防消费者名称生成的时候覆盖原来的消费者,k 为同组消费者的ID.
func (c *BaseConsumer[T]) GenerateUniqueConsumerName(k int) error {
	for {
		consumerName := fmt.Sprintf("%s%s%04d", stream.ConsumerNamePrefix, c.MsgKey, k)

		exist, err := c.ConsumerNameExists(consumerName)
		if err != nil {
			return err
		}

		if !exist {
			c.ConsumerName = consumerName
			return nil
		}

		k++

		// 预防死循环添加一个计数器
		if k > stream.GenerateUniqueConsumerNameMaxCount {
			errText := fmt.Sprintf("生成消费者名称失败,超过最大次数:%d.", stream.GenerateUniqueConsumerNameMaxCount)
			return errors.New(errText)
		}
	}
}

// CreateConsumer 实现 Consumer 接口方法, 创建消费者
func (c *BaseConsumer[T]) CreateConsumer() error {
	return c.Rdb.XGroupCreateConsumer(c.Ctx, c.StreamName, c.GroupName, c.ConsumerName).Err()
}

// PendingMessage 实现 Consumer 接口方法, 获取 pending 的消息并重新分配
func (c *BaseConsumer[T]) PendingMessage() error {
	// 限制数量, 避免一次性太多
	const maxPending = int64(10)

	// 获取当前消费者的 pending 消息详情(查询组内 pending, 后续手动过滤)
	pendingExt, err := c.getPendingExt(maxPending)
	if err != nil {
		return err
	}

	// 没有 pending 消息
	if len(pendingExt) == 0 {
		return nil
	}

	// 过滤并提取需要认领的 pending 消息 ID
	minIdle := 2 * time.Second

	msgIDs, err := c.filterClaimableMsgIDs(pendingExt, minIdle)
	if err != nil {
		return err
	}

	// 没有需要认领的消息
	if len(msgIDs) == 0 {
		return nil
	}

	// 使用 XCLAIM 批量认领这些消息
	claimedMessages, err := c.claimMessages(msgIDs, minIdle)
	if err != nil {
		return err
	}

	// 逐个处理认领到的消息
	c.processClaimedMessages(claimedMessages)

	return nil
}

// getPendingExt 获取组内 pending 扩展信息
func (c *BaseConsumer[T]) getPendingExt(maxPending int64) ([]redis.XPendingExt, error) {
	pendingExt, err := c.Rdb.XPendingExt(c.Ctx, &redis.XPendingExtArgs{
		Stream: c.StreamName,
		Group:  c.GroupName,
		Start:  "-",
		End:    "+",
		Count:  maxPending,
	}).Result()

	if err != nil {
		// 没有 pending 消息
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	return pendingExt, nil
}

// filterClaimableMsgIDs 过滤并提取需要认领的消息 ID：
//   - 跳过属于自己的 pending(避免重复处理)
//   - 只认领空闲时间大于阈值的消息(避免正在被其他消费者处理的短暂情况)
//   - 跳过已被标记为正在处理(processing)的消息
func (c *BaseConsumer[T]) filterClaimableMsgIDs(pendingExt []redis.XPendingExt, minIdle time.Duration) ([]string, error) {
	var msgIDs []string

	for _, p := range pendingExt {
		if p.Consumer == c.ConsumerName {
			// 跳过属于自己的 pending
			continue
		}

		// 仅认领空闲时间达到阈值的消息
		if p.Idle < minIdle {
			continue
		}

		// 如果该消息被标记为正在处理(processing), 则跳过认领
		if c.StateManager != nil {
			if _, processing := c.StateManager.IsProcessing(c.StreamName, p.ID); processing {
				// 有消费者正在处理, 跳过
				continue
			}
		}

		msgIDs = append(msgIDs, p.ID)
	}

	return msgIDs, nil
}

// claimMessages 使用 XCLAIM 批量认领消息
func (c *BaseConsumer[T]) claimMessages(msgIDs []string, minIdle time.Duration) ([]redis.XMessage, error) {
	if len(msgIDs) == 0 {
		return nil, nil
	}

	claimedMessages, err := c.Rdb.XClaim(c.Ctx, &redis.XClaimArgs{
		Stream:   c.StreamName,
		Group:    c.GroupName,
		Consumer: c.ConsumerName,
		MinIdle:  minIdle,
		Messages: msgIDs,
	}).Result()

	if err != nil {
		return nil, err
	}

	return claimedMessages, nil
}

// processClaimedMessages 处理认领到的消息
func (c *BaseConsumer[T]) processClaimedMessages(claimedMessages []redis.XMessage) {
	for _, msg := range claimedMessages {
		if err := c.ProcessMessage(msg); err != nil {
			// 只记录错误日志, 继续处理其他消息
			zap.L().Warn("处理 pending 消息失败, 跳过", zap.String("msgID", msg.ID), zap.Error(err))
			continue
		}
	}
}

// OnlineMessage 拉取并处理一批在线消息
func (c *BaseConsumer[T]) OnlineMessage() error {
	ctxWithTimeout, cancel := context.WithTimeout(c.Ctx, 5*time.Second)
	defer cancel()

	// 拉取在线消息
	entries, err := c.Rdb.XReadGroup(ctxWithTimeout, &redis.XReadGroupArgs{
		Group:    c.GroupName,
		Consumer: c.ConsumerName,
		Streams:  []string{c.StreamName, ">"},
		Count:    10, // 一次拉多条, 提升吞吐
		Block:    0,  // 阻塞由 context 控制
		NoAck:    false,
	}).Result()

	if err != nil {
		// 仅在 context 取消或者超时情况下返回错误
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		// 没有新消息
		if errors.Is(err, redis.Nil) {
			return nil
		}

		// 其他错误
		return err
	}

	// 交给 ProcessMessage 处理, 内部可决定是否 Ack
	for _, entry := range entries[0].Messages {
		if err := c.ProcessMessage(entry); err != nil {
			// 只记录错误日志, 继续处理其他消息
			zap.L().Warn("处理在线消息失败, 跳过", zap.String("msgID", entry.ID), zap.String("consumer", c.ConsumerName), zap.Error(err))
		}
	}

	return nil
}

// startPendingLoop 定时检查并处理挂起的消息
func (c *BaseConsumer[T]) startPendingLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// context 被取消, 退出循环
			zap.L().Info("PendingMessage loop stopped", zap.String("consumer", c.ConsumerName))
			return

		case <-ticker.C:
			if err := c.PendingMessage(); err != nil {
				// 只记录错误日志, 不中断循环
				zap.L().Warn("处理 pending 消息失败", zap.String("consumer", c.ConsumerName), zap.Error(err))
			}
		}
	}
}

// startOnlineMessageLoop 循环监听在线消息(带 context 控制)
func (c *BaseConsumer[T]) startOnlineMessageLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			// context 被取消, 退出循环
			zap.L().Info("OnlineMessage loop stopped", zap.String("consumer", c.ConsumerName))
			return nil

		default:
			if err := c.OnlineMessage(); err != nil {
				// 仅在 context 取消或者超时情况下退出
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return nil
				}

				return fmt.Errorf("拉取在线消息失败: consumer=%s; %w", c.ConsumerName, err)
			}
		}
	}
}

// RunConsumer 实现 Consumer 接口方法, 运行消费者
func (c *BaseConsumer[T]) RunConsumer() error {
	// 启动 goroutine 持续处理 pending 消息
	go c.startPendingLoop(c.Ctx)

	// 主循环：持续监听新消息
	return c.startOnlineMessageLoop(c.Ctx)
}

// ProcessMessage 实现 Consumer 接口方法, 处理消息
func (c *BaseConsumer[T]) ProcessMessage(message redis.XMessage) error {
	return c.ProcessMessageFunc(c, message)
}

// AckMessage 实现 Consumer 接口方法, 签收消息
func (c *BaseConsumer[T]) AckMessage(msgID string, valueStruct *T, isSuccess bool) error {
	status := "failed"
	if isSuccess {
		status = "success"
	}

	// 消息信息
	msg := fmt.Sprintf("%s>%s>%s>ack key:%s>ID:%s, ProcessMessage %s", c.StreamName, c.GroupName, c.ConsumerName, c.MsgKey, msgID, status)

	// 签收消息
	if err := c.Rdb.XAck(c.Ctx, c.StreamName, c.GroupName, msgID).Err(); err != nil {
		// 签收失败
		return fmt.Errorf("签收消息失败: msg=%s; value=%+v; %w", msg, valueStruct, err)
	}

	// 在缓存中更新消息 ID 状态为已处理
	if c.StateManager != nil {
		if err := c.StateManager.UpdateAckStatus(c.StreamName, msgID, c.GroupName, isSuccess); err != nil {
			return err
		}
	}

	// 签收成功
	msg = fmt.Sprintf("签收消息成功, %s", msg)
	zap.L().Info(msg, zap.Any("value", valueStruct))

	return nil
}

// GetConsumersInfo 实现 Consumer 接口方法, 获取该组所有消费者状态
func (c *BaseConsumer[T]) GetConsumersInfo() ([]redis.XInfoConsumer, error) {
	// 获取消费者状态
	return c.Rdb.XInfoConsumers(c.Ctx, c.StreamName, c.GroupName).Result()
}

// GetConsumerInfo 实现 Consumer 接口方法, 获取单个消费者状态
func (c *BaseConsumer[T]) GetConsumerInfo(consumerName string) (redis.XInfoConsumer, error) {
	var info redis.XInfoConsumer

	consumers, err := c.GetConsumersInfo()

	if err != nil {
		return info, err
	}

	for _, consumer := range consumers {
		if consumer.Name == consumerName {
			return consumer, nil
		}
	}

	return info, fmt.Errorf("consumer %s not found", c.ConsumerName)
}

// GetPendingCount 获取当前未消费的消息总数(包括所有消费者的待处理消息)
func (c *BaseConsumer[T]) GetPendingCount() (int64, error) {
	// 使用 XPENDING 命令获取汇总信息
	pendingRes, err := c.Rdb.XPending(c.Ctx, c.StreamName, c.GroupName).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) { // 消费组或 stream 不存在的情况
			return 0, fmt.Errorf("stream %s or group %s not found", c.StreamName, c.GroupName)
		}

		return 0, err
	}

	return pendingRes.Count, nil
}

// RemoveConsumer 实现 Consumer 接口方法, 移除消费者
func (c *BaseConsumer[T]) RemoveConsumer(consumerInfo redis.XInfoConsumer) error {
	// 判断当前消费者是否正在消费(是否有待处理的消息)
	noPending := consumerInfo.Pending == 0

	// 判断是否有消费者空闲
	hasIdle := consumerInfo.Idle > 1*time.Second // 1秒

	// 还未消费信息
	noInactive := consumerInfo.Inactive == -1*time.Millisecond // -1 毫秒

	// 已经有成功消费信息的时间(可以理解为空闲时间)
	hasInactive := consumerInfo.Inactive > 1*time.Second // 1秒

	// 如果没有待处理的消息并且有空闲时间并且(没有成功消费信息或者有较长成功消费信息)
	if noPending && hasIdle && (noInactive || hasInactive) {
		return c.Rdb.XGroupDelConsumer(c.Ctx, c.StreamName, c.GroupName, consumerInfo.Name).Err()
	}

	return errors.New("消费者正在消费中")
}
