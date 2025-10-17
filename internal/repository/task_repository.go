package repository

import (
	"context"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/pkg/database"
)

// TaskRepository 任务仓储接口
type TaskRepository interface {
	// 基础 CRUD 方法由 database.Repository[T] 提供
	Create(ctx context.Context, task *model.Task) error
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Task, error)

	// 业务特定查询方法
	FindByTaskID(ctx context.Context, taskID string) (*model.Task, error)
	ListByStatus(ctx context.Context, status string, pagination *database.Pagination) ([]model.Task, error)
	ListByUser(ctx context.Context, userID uint, pagination *database.Pagination) ([]model.Task, error)
	ListByType(ctx context.Context, taskType string, pagination *database.Pagination) ([]model.Task, error)
	ListByUserAndStatus(ctx context.Context, userID uint, status string, pagination *database.Pagination) ([]model.Task, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
	UpdateStatus(ctx context.Context, taskID string, status string, err string) error
}

// taskRepository 任务仓储实现
type taskRepository struct {
	*database.Repository[model.Task]
}

// NewTaskRepository 创建任务仓储
func NewTaskRepository(db *database.Database) TaskRepository {
	return &taskRepository{
		Repository: database.NewRepository[model.Task](db.DB),
	}
}

// FindByTaskID 根据任务ID查找任务
func (r *taskRepository) FindByTaskID(ctx context.Context, taskID string) (*model.Task, error) {
	return r.Repository.FindOne(ctx, "task_id = ?", taskID)
}

// ListByStatus 根据状态查询任务列表
func (r *taskRepository) ListByStatus(ctx context.Context, status string, pagination *database.Pagination) ([]model.Task, error) {
	query := "status = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, status)
}

// ListByUser 查询用户的任务列表
func (r *taskRepository) ListByUser(ctx context.Context, userID uint, pagination *database.Pagination) ([]model.Task, error) {
	query := "user_id = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, userID)
}

// ListByType 根据任务类型查询列表
func (r *taskRepository) ListByType(ctx context.Context, taskType string, pagination *database.Pagination) ([]model.Task, error) {
	query := "type = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, taskType)
}

// ListByUserAndStatus 查询用户特定状态的任务
func (r *taskRepository) ListByUserAndStatus(ctx context.Context, userID uint, status string, pagination *database.Pagination) ([]model.Task, error) {
	query := "user_id = ? AND status = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, userID, status)
}

// CountByStatus 统计特定状态的任务数量
func (r *taskRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.Repository.DB().WithContext(ctx).
		Model(&model.Task{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

// CountByUser 统计用户任务数量
func (r *taskRepository) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.Repository.DB().WithContext(ctx).
		Model(&model.Task{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// UpdateStatus 更新任务状态
func (r *taskRepository) UpdateStatus(ctx context.Context, taskID string, status string, errMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errMsg != "" {
		updates["error"] = errMsg
	}
	return r.Repository.DB().WithContext(ctx).
		Model(&model.Task{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}
