package scheduler

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cccvno1/nova/pkg/logger"
	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron *cron.Cron
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
	// 创建支持秒级的 cron（默认只支持到分钟）
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cron: c,
	}
}

// AddFunc 添加定时任务（Cron 表达式）
// Cron 表达式格式: 秒 分 时 日 月 周
// 例如: "0 0 2 * * *" 表示每天凌晨2点执行
// 例如: "*/30 * * * * *" 表示每30秒执行一次
func (s *Scheduler) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	entryID, err := s.cron.AddFunc(spec, func() {
		logger.Info("executing scheduled task", slog.String("spec", spec))
		startTime := time.Now()
		cmd()
		duration := time.Since(startTime)
		logger.Info("scheduled task completed",
			slog.String("spec", spec),
			slog.Duration("duration", duration))
	})

	if err != nil {
		return 0, fmt.Errorf("failed to add scheduled task: %w", err)
	}

	logger.Info("scheduled task added",
		slog.String("spec", spec),
		slog.Int("entry_id", int(entryID)))

	return entryID, nil
}

// AddInterval 添加间隔任务（每隔指定时间执行）
// 例如: AddInterval(30*time.Second, func) 表示每30秒执行一次
func (s *Scheduler) AddInterval(interval time.Duration, cmd func()) (cron.EntryID, error) {
	// 使用 @every 语法
	spec := fmt.Sprintf("@every %s", interval.String())

	entryID, err := s.cron.AddFunc(spec, func() {
		logger.Info("executing interval task", slog.Duration("interval", interval))
		startTime := time.Now()
		cmd()
		duration := time.Since(startTime)
		logger.Info("interval task completed",
			slog.Duration("interval", interval),
			slog.Duration("duration", duration))
	})

	if err != nil {
		return 0, fmt.Errorf("failed to add interval task: %w", err)
	}

	logger.Info("interval task added",
		slog.Duration("interval", interval),
		slog.Int("entry_id", int(entryID)))

	return entryID, nil
}

// AddJob 添加自定义任务
func (s *Scheduler) AddJob(spec string, job cron.Job) (cron.EntryID, error) {
	entryID, err := s.cron.AddJob(spec, job)
	if err != nil {
		return 0, fmt.Errorf("failed to add job: %w", err)
	}

	logger.Info("custom job added",
		slog.String("spec", spec),
		slog.Int("entry_id", int(entryID)))

	return entryID, nil
}

// Remove 移除任务
func (s *Scheduler) Remove(id cron.EntryID) {
	s.cron.Remove(id)
	logger.Info("task removed", slog.Int("entry_id", int(id)))
}

// Start 启动调度器
func (s *Scheduler) Start() {
	logger.Info("starting scheduler")
	s.cron.Start()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	logger.Info("stopping scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("scheduler stopped")
}

// Entries 获取所有任务条目
func (s *Scheduler) Entries() []cron.Entry {
	return s.cron.Entries()
}

// Entry 根据 ID 获取任务条目
func (s *Scheduler) Entry(id cron.EntryID) cron.Entry {
	return s.cron.Entry(id)
}

// Stats 获取调度器统计信息
func (s *Scheduler) Stats() map[string]interface{} {
	entries := s.cron.Entries()

	stats := map[string]interface{}{
		"total_tasks": len(entries),
		"tasks":       make([]map[string]interface{}, 0, len(entries)),
	}

	for _, entry := range entries {
		taskInfo := map[string]interface{}{
			"id":       entry.ID,
			"next_run": entry.Next.Format("2006-01-02 15:04:05"),
			"prev_run": entry.Prev.Format("2006-01-02 15:04:05"),
			"valid":    entry.Valid(),
		}
		stats["tasks"] = append(stats["tasks"].([]map[string]interface{}), taskInfo)
	}

	return stats
}
