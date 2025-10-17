package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/cccvno1/nova/pkg/cache"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Client 队列客户端
type Client struct {
	queueKey string
	delayKey string
}

// NewClient 创建队列客户端
func NewClient(prefix string) *Client {
	return &Client{
		queueKey: prefix + ":tasks",
		delayKey: prefix + ":delayed_tasks",
	}
}

// Submit 提交任务到队列
func (c *Client) Submit(ctx context.Context, name string, payload map[string]interface{}, maxRetry int) (string, error) {
	// 生成任务 ID
	taskID := uuid.New().String()

	// 创建任务
	task := &Task{
		ID:         taskID,
		Name:       name,
		Payload:    payload,
		RetryCount: 0,
		MaxRetry:   maxRetry,
		CreatedAt:  time.Now(),
		ExecuteAt:  time.Now(),
	}

	// 序列化任务
	data, err := task.Marshal()
	if err != nil {
		return "", errors.Wrap(errors.ErrInternalServer, err)
	}

	// 推入队列
	if err := cache.LPush(ctx, c.queueKey, data); err != nil {
		return "", errors.Wrap(errors.ErrInternalServer, err)
	}

	return taskID, nil
}

// SubmitIn 延迟提交任务（在指定时间后执行）
func (c *Client) SubmitIn(ctx context.Context, name string, payload map[string]interface{}, maxRetry int, delay time.Duration) (string, error) {
	// 生成任务 ID
	taskID := uuid.New().String()

	// 创建任务
	executeAt := time.Now().Add(delay)
	task := &Task{
		ID:         taskID,
		Name:       name,
		Payload:    payload,
		RetryCount: 0,
		MaxRetry:   maxRetry,
		CreatedAt:  time.Now(),
		ExecuteAt:  executeAt,
	}

	// 序列化任务
	data, err := task.Marshal()
	if err != nil {
		return "", errors.Wrap(errors.ErrInternalServer, err)
	}

	// 使用 ZSet 存储延迟任务（以执行时间为分数）
	score := float64(executeAt.Unix())
	z := redis.Z{
		Score:  score,
		Member: string(data),
	}
	if err := cache.ZAdd(ctx, c.delayKey, z); err != nil {
		return "", errors.Wrap(errors.ErrInternalServer, err)
	}

	return taskID, nil
}

// GetQueueLength 获取队列长度
func (c *Client) GetQueueLength(ctx context.Context) (int64, error) {
	length, err := cache.LLen(ctx, c.queueKey)
	if err != nil {
		return 0, errors.Wrap(errors.ErrInternalServer, err)
	}
	return length, nil
}

// GetQueueKey 获取队列键
func (c *Client) GetQueueKey() string {
	return c.queueKey
}

// GetDelayKey 获取延迟队列键
func (c *Client) GetDelayKey() string {
	return c.delayKey
}

// ClearQueue 清空队列（用于测试）
func (c *Client) ClearQueue(ctx context.Context) error {
	if err := cache.Del(ctx, c.queueKey); err != nil {
		return errors.Wrap(errors.ErrInternalServer, err)
	}
	if err := cache.Del(ctx, c.delayKey); err != nil {
		return errors.Wrap(errors.ErrInternalServer, err)
	}
	return nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	if err := cache.HealthCheck(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}
