package queue

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cccvno1/nova/pkg/cache"
	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/logger"
)

// Worker 队列 Worker
type Worker struct {
	client     *Client
	handlers   map[string]HandlerFunc
	workerNum  int
	maxRetry   int
	retryDelay time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// NewWorker 创建 Worker
func NewWorker(cfg *config.QueueConfig) *Worker {
	client := NewClient(cfg.RedisPrefix)
	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		client:     client,
		handlers:   make(map[string]HandlerFunc),
		workerNum:  cfg.Workers,
		maxRetry:   cfg.MaxRetry,
		retryDelay: time.Duration(cfg.RetryDelay) * time.Second,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Register 注册任务处理器
func (w *Worker) Register(name string, handler HandlerFunc) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[name] = handler
	logger.Info("registered task handler", slog.String("task", name))
}

// Start 启动 Worker
func (w *Worker) Start() error {
	logger.Info("starting queue workers", slog.Int("workers", w.workerNum))

	// 启动多个 Worker 协程
	for i := 0; i < w.workerNum; i++ {
		w.wg.Add(1)
		go w.work(i)
	}

	// 启动延迟任务调度器
	w.wg.Add(1)
	go w.scheduleDelayedTasks()

	return nil
}

// Stop 停止 Worker
func (w *Worker) Stop() error {
	logger.Info("stopping queue workers")
	w.cancel()
	w.wg.Wait()
	logger.Info("queue workers stopped")
	return nil
}

// work Worker 工作循环
func (w *Worker) work(id int) {
	defer w.wg.Done()

	logger.Info("worker started", slog.Int("worker_id", id))

	for {
		select {
		case <-w.ctx.Done():
			logger.Info("worker stopped", slog.Int("worker_id", id))
			return
		default:
			// 从队列中弹出任务（阻塞 5 秒）
			result, err := cache.BRPop(w.ctx, 5*time.Second, w.client.GetQueueKey())
			if err != nil {
				// 超时或其他错误，继续循环
				continue
			}

			if len(result) < 2 {
				continue
			}

			// result[0] 是队列名，result[1] 是任务数据
			taskData := result[1]

			// 反序列化任务
			task, err := UnmarshalTask([]byte(taskData))
			if err != nil {
				logger.Error("failed to unmarshal task",
					slog.String("error", err.Error()),
					slog.Int("worker_id", id))
				continue
			}

			// 执行任务
			w.processTask(task, id)
		}
	}
}

// processTask 处理任务
func (w *Worker) processTask(task *Task, workerID int) {
	logger.Info("processing task",
		slog.String("task_id", task.ID),
		slog.String("task_name", task.Name),
		slog.Int("worker_id", workerID))

	// 获取处理器
	w.mu.RLock()
	handler, exists := w.handlers[task.Name]
	w.mu.RUnlock()

	if !exists {
		logger.Error("task handler not found",
			slog.String("task_id", task.ID),
			slog.String("task_name", task.Name))
		return
	}

	// 执行处理器
	startTime := time.Now()
	err := handler(task)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error("task failed",
			slog.String("task_id", task.ID),
			slog.String("task_name", task.Name),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration),
			slog.Int("retry_count", task.RetryCount))

		// 重试逻辑
		if task.RetryCount < task.MaxRetry {
			task.RetryCount++
			logger.Info("retrying task",
				slog.String("task_id", task.ID),
				slog.Int("retry_count", task.RetryCount),
				slog.Duration("delay", w.retryDelay))

			// 延迟重新提交任务
			go func() {
				time.Sleep(w.retryDelay)
				data, _ := task.Marshal()
				if err := cache.LPush(w.ctx, w.client.GetQueueKey(), data); err != nil {
					logger.Error("failed to retry task",
						slog.String("task_id", task.ID),
						slog.String("error", err.Error()))
				}
			}()
		} else {
			logger.Error("task failed after max retries",
				slog.String("task_id", task.ID),
				slog.String("task_name", task.Name),
				slog.Int("max_retry", task.MaxRetry))
		}
	} else {
		logger.Info("task completed",
			slog.String("task_id", task.ID),
			slog.String("task_name", task.Name),
			slog.Duration("duration", duration))
	}
}

// scheduleDelayedTasks 调度延迟任务
func (w *Worker) scheduleDelayedTasks() {
	defer w.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logger.Info("delayed task scheduler started")

	for {
		select {
		case <-w.ctx.Done():
			logger.Info("delayed task scheduler stopped")
			return
		case <-ticker.C:
			w.checkDelayedTasks()
		}
	}
}

// checkDelayedTasks 检查并移动到期的延迟任务
func (w *Worker) checkDelayedTasks() {
	now := float64(time.Now().Unix())

	// 获取分数小于当前时间的任务（即已到期的任务）
	tasks, err := cache.ZRange(w.ctx, w.client.GetDelayKey(), 0, -1)
	if err != nil {
		return
	}

	for _, taskData := range tasks {
		task, err := UnmarshalTask([]byte(taskData))
		if err != nil {
			continue
		}

		// 检查是否到期
		if float64(task.ExecuteAt.Unix()) <= now {
			// 移动到主队列
			if err := cache.LPush(w.ctx, w.client.GetQueueKey(), taskData); err != nil {
				logger.Error("failed to move delayed task to queue",
					slog.String("task_id", task.ID),
					slog.String("error", err.Error()))
				continue
			}

			// 从延迟队列中移除
			if err := cache.ZRem(w.ctx, w.client.GetDelayKey(), taskData); err != nil {
				logger.Error("failed to remove delayed task",
					slog.String("task_id", task.ID),
					slog.String("error", err.Error()))
			}

			logger.Info("moved delayed task to queue",
				slog.String("task_id", task.ID),
				slog.String("task_name", task.Name))
		}
	}
}

// GetClient 获取队列客户端
func (w *Worker) GetClient() *Client {
	return w.client
}

// HealthCheck 健康检查
func (w *Worker) HealthCheck() error {
	return w.client.HealthCheck(w.ctx)
}

// Stats 获取队列统计信息
func (w *Worker) Stats() (map[string]interface{}, error) {
	queueLen, err := w.client.GetQueueLength(w.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue length: %w", err)
	}

	stats := map[string]interface{}{
		"workers":     w.workerNum,
		"queue_len":   queueLen,
		"max_retry":   w.maxRetry,
		"retry_delay": w.retryDelay.Seconds(),
	}

	return stats, nil
}
