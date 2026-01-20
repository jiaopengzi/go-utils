//
// FilePath    : go-utils\redis\stream\consumer\factory.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 消费者工厂
//

package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	_stream "github.com/jiaopengzi/go-utils/redis/stream"
)

// ConsumerConfig 通用消费者配置结构体
type ConsumerConfig[T any] struct {
	StreamName         string                                                 // 消息队列名称
	GroupName          string                                                 // 消费者组名称
	MsgKey             string                                                 // 消息键
	ProcessMessageFunc func(c *BaseConsumer[T], message redis.XMessage) error // 处理消息函数
	ConfigCount        int                                                    // 消费者数量
	Ctx                context.Context                                        // context 上下文
	Rdb                redis.UniversalClient                                  // Redis 客户端
	StateManager       MessageStateManager                                    // 消息状态管理器
}

// ManageConsumers 通过配置初始化并管理消费者
func ManageConsumers[T any](config *ConsumerConfig[T]) error {
	count := _stream.ConsumerMinCount // 默认最少需要消费者

	// 根据配置设置消费者数量
	if config.ConfigCount > _stream.ConsumerMinCount && config.ConfigCount <= _stream.ConsumerMaxCount {
		count = config.ConfigCount
	}

	// 消费者结构体
	consumer := &BaseConsumer[T]{
		StreamName:         config.StreamName,
		GroupName:          config.GroupName,
		Start:              _stream.CreateStreamStart,
		MsgKey:             config.MsgKey,
		ProcessMessageFunc: config.ProcessMessageFunc,
		Ctx:                config.Ctx,
		Rdb:                config.Rdb,
		StateManager:       config.StateManager,
	}

	// 创建消费者组
	if err := consumer.CreateGroup(); err != nil {
		return err
	}

	// 在循环中, 根据配置 创建、移除消费者.
	for i := range count {
		if err := manageConsumer(consumer, count, i); err != nil {
			return err
		}
	}

	// 获取最终同组所有消费者信息
	consumerInfos, err := consumer.GetConsumersInfo()
	if err != nil {
		return err
	}

	// 协程异步运行消费者
	for _, consumerInfo := range consumerInfos {
		consumer.ConsumerName = consumerInfo.Name
		go func(c BaseConsumer[T]) {
			if err = c.RunConsumer(); err != nil {
				zap.L().Error("消费者运行错误", zap.Error(err), zap.String("consumerName", c.ConsumerName))
				return
			}
		}(*consumer) // 传递消费者结构体的值,注意不要使用指针,造成指向同一个内存地址,也就是同一个消费者.
	}

	return nil
}

// manageConsumer 根据消费者 consumer 及消费者数量 count 和 消费者索引 i 进行消费者管理.
//
//  1. 如果当前消费者数量等于配置的消费者数量则退出(不再创建消费者).
//
//  2. 如果当前消费者数量小于配置的消费者数量则补充创建消费者(创建消费者).
//
//  3. 如果当前消费者数量大于配置的消费者数量则移除多余的消费者(移除消费者).
func manageConsumer[T any](consumer *BaseConsumer[T], count, i int) error {
	// 获取当前消费者同组所有消费者信息
	consumerInfos, err := consumer.GetConsumersInfo()
	if err != nil {
		return err
	}

	// 获取当前消费者同组消费者数量
	currentCount := len(consumerInfos)

	// 1、如果当前消费者数量等于配置的消费者数量则退出(不再创建消费者).
	if currentCount == count {
		return nil
	}

	// 2、如果当前消费者数量小于配置的消费者数量则补充创建消费者(创建消费者).
	// 根据当前消费者最后最后一位,从低位到高位递补创建.
	if currentCount < count && currentCount <= i {
		return createConsumerIfNeeded(consumer, i)
	}

	// 3、如果当前消费者数量大于配置的消费者数量则移除多余的消费者(移除消费者).
	if currentCount > count {
		return removeExtraConsumers(consumer, consumerInfos, currentCount, count)
	}

	return nil
}

// createConsumerIfNeeded 负责为 consumer 根据索引 i 生成唯一名称并创建消费者。
// 生成唯一名称 -> 创建消费者 -> 记录日志
func createConsumerIfNeeded[T any](consumer *BaseConsumer[T], i int) error {
	// 生成唯一的消费者名称
	if err := consumer.GenerateUniqueConsumerName(i); err != nil {
		return err
	}

	// 创建消费者
	if err := consumer.CreateConsumer(); err != nil {
		return err
	}

	// 记录日志
	zap.L().Info("创建消费者成功", zap.String("consumerName", consumer.ConsumerName))

	return nil
}

// removeExtraConsumers 在消费者数量大于期望时移除多余消费者。
// 参数说明:
//   - consumer: 当前消费者模板
//   - consumerInfos: 当前同组所有消费者信息
//   - currentCount: 当前消费者数量
//   - targetCount: 目标消费者数量
func removeExtraConsumers[T any](consumer *BaseConsumer[T], consumerInfos []redis.XInfoConsumer, currentCount, targetCount int) error {
	// 预防死循环设置一个最大循环次数
	MaxForCount := 0

	// 循环移除多余的消费者,直到消费者数量等于配置的消费者数量.
	for {
		for _, consumerInfo := range consumerInfos {
			// 移除消费者
			if err := consumer.RemoveConsumer(consumerInfo); err != nil {
				// 记录日志, 这里的错误是允许的
				zap.L().Warn("移除消费者错误", zap.Error(err), zap.String("consumerName", consumerInfo.Name))
			} else {
				// 移除成功,当前消费者数量减一
				currentCount--

				// 如果当前消费者数量等于配置的消费者数量则退出
				if currentCount == targetCount {
					return nil
				}

				// 记录日志
				zap.L().Info("移除消费者成功", zap.String("consumerName", consumerInfo.Name))
			}

			MaxForCount++
			if MaxForCount == _stream.RemoveConsumerForMaxCount {
				commandStr := fmt.Sprintf("消费者删除已经执行了 10000 次, 请在 redis 执行命令:XINFO CONSUMERS %s %s 查看具体问题。", consumer.StreamName, consumer.GroupName)
				err := errors.New(commandStr)
				zap.L().Error("移除消费者陷入死循环, 请检查", zap.Error(err))

				return err
			}
		}
	}
}

// parseMessageValue 泛型函数 根据消息 message 和消息中的键 msgKey 解析消息中的值
func parseMessageValue[T any](message redis.XMessage, msgKey string) (*T, error) {
	var valueStruct T // 定义一个结构体

	// 获取消息中的 key
	msgValue, ok := message.Values[msgKey]
	if !ok {
		return nil, fmt.Errorf("message does not have key %s", msgKey)
	}

	// 将消息转换为 string
	msgStr, ok := msgValue.(string)
	if !ok {
		return nil, fmt.Errorf("failed to assert message as string %s", msgStr)
	}

	// 将消息序列化为结构体
	if err := json.Unmarshal([]byte(msgStr), &valueStruct); err != nil {
		return nil, err
	}

	// zap.L().Debug("异步处理消息前", zap.String("msgStr", msgStr))
	return &valueStruct, nil
}

// HandleAndAckMessage 泛型函数 处理消息并签收消息
//   - c: 消费者
//   - message: 消息
//   - msgKey: 消息中的 key
//   - messageHandler: 处理消息的回调函数
func HandleAndAckMessage[T any](c *BaseConsumer[T], message redis.XMessage, msgKey string, messageHandler func(valueStruct *T) error) error {
	// 在处理前标记为正在处理中, 防止其他消费者认领
	if c.StateManager != nil {
		if errSet := c.StateManager.MarkProcessing(c.StreamName, message.ID, c.ConsumerName); errSet != nil {
			zap.L().Warn("set processing flag failed", zap.Error(errSet), zap.String("msgID", message.ID))
		}
	}

	// 解析消息中的值
	valueStruct, err := parseMessageValue[T](message, msgKey)

	// 日志字段记录
	logFields := func(err error) []zap.Field {
		return []zap.Field{
			zap.Error(err),
			zap.String("consumer", c.ConsumerName),
			zap.String("msgKey", msgKey),
			zap.String("msgID", message.ID),
			zap.Any("message.Values", message.Values),
			zap.Any("valueStruct", valueStruct),
		}
	}

	if err != nil {
		zap.L().Error("parseMessageValue() failed", logFields(err)...)
		return err
	}

	// 调用回调函数处理消息
	if err = messageHandler(valueStruct); err != nil {
		zap.L().Error("messageHandler() failed DLQ(Dead Letter Queue, 死信队列)", logFields(err)...)

		// 消费失败 ACK 签收消息
		if err = c.AckMessage(message.ID, valueStruct, false); err != nil {
			zap.L().Error("c.AckMessage() failed", logFields(err)...)
			return err
		}

		// 删除处理标记(无论 Ack 成功与否, 都需要删除标记)
		if c.StateManager != nil {
			if errDel := c.StateManager.ClearProcessing(c.StreamName, message.ID); errDel != nil {
				zap.L().Error("del processing flag failed", zap.Error(errDel), zap.String("msgID", message.ID))
			}
		}

		return err
	}

	// 消费成功 ACK 签收消息
	if err = c.AckMessage(message.ID, valueStruct, true); err != nil {
		zap.L().Error("c.AckMessage() failed", logFields(err)...)
		return err
	}

	// 删除处理标记
	if c.StateManager != nil {
		if errDel := c.StateManager.ClearProcessing(c.StreamName, message.ID); errDel != nil {
			zap.L().Error("del processing flag failed", zap.Error(errDel), zap.String("msgID", message.ID))
		}
	}

	return nil
}
