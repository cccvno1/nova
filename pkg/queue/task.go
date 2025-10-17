package queue

import (
	"encoding/json"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"    // 等待中
	TaskStatusProcessing TaskStatus = "processing" // 处理中
	TaskStatusSuccess    TaskStatus = "success"    // 成功
	TaskStatusFailed     TaskStatus = "failed"     // 失败
)

// Task 队列任务
type Task struct {
	ID         string                 `json:"id"`          // 任务 ID
	Name       string                 `json:"name"`        // 任务名称
	Payload    map[string]interface{} `json:"payload"`     // 任务载荷
	RetryCount int                    `json:"retry_count"` // 重试次数
	MaxRetry   int                    `json:"max_retry"`   // 最大重试次数
	CreatedAt  time.Time              `json:"created_at"`  // 创建时间
	ExecuteAt  time.Time              `json:"execute_at"`  // 执行时间
}

// HandlerFunc 任务处理函数
type HandlerFunc func(task *Task) error

// Marshal 序列化任务
func (t *Task) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

// Unmarshal 反序列化任务
func UnmarshalTask(data []byte) (*Task, error) {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}
