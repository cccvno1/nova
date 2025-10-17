package repository

import (
	"context"
	"time"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/pkg/database"
)

// AuditLogRepository 审计日志仓储接口
type AuditLogRepository interface {
	// 基础 CRUD 方法由 database.Repository[T] 提供
	Create(ctx context.Context, log *model.AuditLog) error
	FindByID(ctx context.Context, id uint) (*model.AuditLog, error)

	// 业务特定查询方法
	List(ctx context.Context, pagination *database.Pagination) ([]model.AuditLog, error)
	ListByUser(ctx context.Context, userID uint, pagination *database.Pagination) ([]model.AuditLog, error)
	ListByAction(ctx context.Context, action string, pagination *database.Pagination) ([]model.AuditLog, error)
	ListByResource(ctx context.Context, resource string, pagination *database.Pagination) ([]model.AuditLog, error)
	ListByIP(ctx context.Context, ip string, pagination *database.Pagination) ([]model.AuditLog, error)
	ListByTimeRange(ctx context.Context, startTime, endTime time.Time, pagination *database.Pagination) ([]model.AuditLog, error)
	Search(ctx context.Context, filters map[string]interface{}, pagination *database.Pagination) ([]model.AuditLog, error)

	// 统计方法
	CountByUser(ctx context.Context, userID uint) (int64, error)
	CountByAction(ctx context.Context, action string) (int64, error)
	CountByStatus(ctx context.Context, statusCode int) (int64, error)
	CountByTimeRange(ctx context.Context, startTime, endTime time.Time) (int64, error)
	GetActionStats(ctx context.Context, startTime, endTime time.Time) ([]map[string]interface{}, error)
	GetUserStats(ctx context.Context, startTime, endTime time.Time, limit int) ([]map[string]interface{}, error)
	GetResourceStats(ctx context.Context, startTime, endTime time.Time) ([]map[string]interface{}, error)

	// 清理方法
	DeleteBefore(ctx context.Context, beforeTime time.Time) (int64, error)
}

// auditLogRepository 审计日志仓储实现
type auditLogRepository struct {
	*database.Repository[model.AuditLog]
}

// NewAuditLogRepository 创建审计日志仓储
func NewAuditLogRepository(db *database.Database) AuditLogRepository {
	return &auditLogRepository{
		Repository: database.NewRepository[model.AuditLog](db.DB),
	}
}

// List 查询审计日志列表
func (r *auditLogRepository) List(ctx context.Context, pagination *database.Pagination) ([]model.AuditLog, error) {
	return r.Repository.FindWithPagination(ctx, pagination, "")
}

// ListByUser 根据用户ID查询审计日志
func (r *auditLogRepository) ListByUser(ctx context.Context, userID uint, pagination *database.Pagination) ([]model.AuditLog, error) {
	query := "user_id = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, userID)
}

// ListByAction 根据操作动作查询审计日志
func (r *auditLogRepository) ListByAction(ctx context.Context, action string, pagination *database.Pagination) ([]model.AuditLog, error) {
	query := "action = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, action)
}

// ListByResource 根据资源类型查询审计日志
func (r *auditLogRepository) ListByResource(ctx context.Context, resource string, pagination *database.Pagination) ([]model.AuditLog, error) {
	query := "resource = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, resource)
}

// ListByIP 根据IP地址查询审计日志
func (r *auditLogRepository) ListByIP(ctx context.Context, ip string, pagination *database.Pagination) ([]model.AuditLog, error) {
	query := "ip = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, ip)
}

// ListByTimeRange 根据时间范围查询审计日志
func (r *auditLogRepository) ListByTimeRange(ctx context.Context, startTime, endTime time.Time, pagination *database.Pagination) ([]model.AuditLog, error) {
	var logs []model.AuditLog

	db := r.Repository.DB().WithContext(ctx).Model(&model.AuditLog{})
	db = db.Where("created_at BETWEEN ? AND ?", startTime, endTime)

	// 统计总数
	if err := db.Count(&pagination.Total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	if pagination.PageSize > 0 {
		offset := (pagination.Page - 1) * pagination.PageSize
		db = db.Offset(offset).Limit(pagination.PageSize)
	}

	// 默认按创建时间倒序
	db = db.Order("created_at DESC")

	if err := db.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// Search 复合条件搜索审计日志
func (r *auditLogRepository) Search(ctx context.Context, filters map[string]interface{}, pagination *database.Pagination) ([]model.AuditLog, error) {
	var logs []model.AuditLog

	db := r.Repository.DB().WithContext(ctx).Model(&model.AuditLog{})

	// 应用过滤条件
	if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		db = db.Where("action = ?", action)
	}
	if resource, ok := filters["resource"].(string); ok && resource != "" {
		db = db.Where("resource = ?", resource)
	}
	if ip, ok := filters["ip"].(string); ok && ip != "" {
		db = db.Where("ip = ?", ip)
	}
	if statusCode, ok := filters["status_code"].(int); ok && statusCode > 0 {
		db = db.Where("status_code = ?", statusCode)
	}
	if method, ok := filters["method"].(string); ok && method != "" {
		db = db.Where("method = ?", method)
	}
	if path, ok := filters["path"].(string); ok && path != "" {
		db = db.Where("path LIKE ?", "%"+path+"%")
	}
	if startTime, ok := filters["start_time"].(time.Time); ok && !startTime.IsZero() {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filters["end_time"].(time.Time); ok && !endTime.IsZero() {
		db = db.Where("created_at <= ?", endTime)
	}

	// 统计总数
	if err := db.Count(&pagination.Total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	if pagination.PageSize > 0 {
		offset := (pagination.Page - 1) * pagination.PageSize
		db = db.Offset(offset).Limit(pagination.PageSize)
	}

	// 默认按创建时间倒序
	db = db.Order("created_at DESC")

	if err := db.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// CountByUser 统计用户的操作次数
func (r *auditLogRepository) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.Repository.DB().WithContext(ctx).
		Model(&model.AuditLog{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByAction 统计指定动作的次数
func (r *auditLogRepository) CountByAction(ctx context.Context, action string) (int64, error) {
	var count int64
	err := r.Repository.DB().WithContext(ctx).
		Model(&model.AuditLog{}).
		Where("action = ?", action).
		Count(&count).Error
	return count, err
}

// CountByStatus 统计指定状态码的次数
func (r *auditLogRepository) CountByStatus(ctx context.Context, statusCode int) (int64, error) {
	var count int64
	err := r.Repository.DB().WithContext(ctx).
		Model(&model.AuditLog{}).
		Where("status_code = ?", statusCode).
		Count(&count).Error
	return count, err
}

// CountByTimeRange 统计时间范围内的操作次数
func (r *auditLogRepository) CountByTimeRange(ctx context.Context, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := r.Repository.DB().WithContext(ctx).
		Model(&model.AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&count).Error
	return count, err
}

// GetActionStats 获取操作统计
func (r *auditLogRepository) GetActionStats(ctx context.Context, startTime, endTime time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	rows, err := r.Repository.DB().WithContext(ctx).
		Model(&model.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Group("action").
		Order("count DESC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int64
		if err := rows.Scan(&action, &count); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"action": action,
			"count":  count,
		})
	}

	return results, nil
}

// GetUserStats 获取用户操作统计（Top N）
func (r *auditLogRepository) GetUserStats(ctx context.Context, startTime, endTime time.Time, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	rows, err := r.Repository.DB().WithContext(ctx).
		Model(&model.AuditLog{}).
		Select("user_id, username, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Group("user_id, username").
		Order("count DESC").
		Limit(limit).
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID uint
		var username string
		var count int64
		if err := rows.Scan(&userID, &username, &count); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"user_id":  userID,
			"username": username,
			"count":    count,
		})
	}

	return results, nil
}

// GetResourceStats 获取资源操作统计
func (r *auditLogRepository) GetResourceStats(ctx context.Context, startTime, endTime time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	rows, err := r.Repository.DB().WithContext(ctx).
		Model(&model.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Group("resource").
		Order("count DESC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var resource string
		var count int64
		if err := rows.Scan(&resource, &count); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"resource": resource,
			"count":    count,
		})
	}

	return results, nil
}

// DeleteBefore 删除指定时间之前的审计日志
func (r *auditLogRepository) DeleteBefore(ctx context.Context, beforeTime time.Time) (int64, error) {
	result := r.Repository.DB().WithContext(ctx).
		Where("created_at < ?", beforeTime).
		Delete(&model.AuditLog{})

	return result.RowsAffected, result.Error
}
