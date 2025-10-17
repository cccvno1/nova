package model

import "github.com/cccvno1/nova/pkg/database"

// Task 任务模型
type Task struct {
	database.Model
	TaskID     string `gorm:"not null;size:100;uniqueIndex" json:"task_id"`           // 任务ID（UUID）
	Name       string `gorm:"not null;size:100;index" json:"name"`                    // 任务名称
	Type       string `gorm:"not null;size:50;index" json:"type"`                     // 任务类型: async, scheduled
	Payload    string `gorm:"type:text" json:"payload"`                               // 任务载荷（JSON）
	Status     string `gorm:"not null;size:20;index;default:'pending'" json:"status"` // 任务状态
	RetryCount int    `gorm:"default:0" json:"retry_count"`                           // 重试次数
	MaxRetry   int    `gorm:"default:3" json:"max_retry"`                             // 最大重试次数
	Error      string `gorm:"type:text" json:"error,omitempty"`                       // 错误信息
	UserID     uint   `gorm:"index" json:"user_id,omitempty"`                         // 关联用户ID（可选）
}

func (Task) TableName() string {
	return "tasks"
}

// TaskType 任务类型常量
const (
	TaskTypeAsync     = "async"     // 异步任务
	TaskTypeScheduled = "scheduled" // 定时任务
)

// TaskStatus 任务状态常量
const (
	TaskStatusPending    = "pending"    // 等待中
	TaskStatusProcessing = "processing" // 处理中
	TaskStatusSuccess    = "success"    // 成功
	TaskStatusFailed     = "failed"     // 失败
	TaskStatusCancelled  = "cancelled"  // 已取消
)
