package repository

import (
	"context"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/pkg/database"
)

// FileRepository 文件仓储接口
type FileRepository interface {
	// 基础 CRUD 方法由 database.Repository[T] 提供
	Create(ctx context.Context, file *model.File) error
	Update(ctx context.Context, file *model.File) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.File, error)

	// 业务特定查询方法
	FindByHash(ctx context.Context, hash string) (*model.File, error)
	FindBySavedName(ctx context.Context, savedName string) (*model.File, error)
	ListByUser(ctx context.Context, userID uint, pagination *database.Pagination) ([]model.File, error)
	ListByCategory(ctx context.Context, category string, pagination *database.Pagination) ([]model.File, error)
	ListByUserAndCategory(ctx context.Context, userID uint, category string, pagination *database.Pagination) ([]model.File, error)
	Search(ctx context.Context, keyword string, pagination *database.Pagination) ([]model.File, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
	GetUserStorageUsage(ctx context.Context, userID uint) (int64, error)
}

// fileRepository 文件仓储实现
type fileRepository struct {
	*database.Repository[model.File]
}

// NewFileRepository 创建文件仓储
func NewFileRepository(db *database.Database) FileRepository {
	return &fileRepository{
		Repository: database.NewRepository[model.File](db.DB),
	}
}

// FindByHash 根据 Hash 查找文件（用于秒传功能）
func (r *fileRepository) FindByHash(ctx context.Context, hash string) (*model.File, error) {
	return r.Repository.FindOne(ctx, "hash = ? AND status = ?", hash, model.FileStatusNormal)
}

// FindBySavedName 根据保存的文件名查找
func (r *fileRepository) FindBySavedName(ctx context.Context, savedName string) (*model.File, error) {
	return r.Repository.FindOne(ctx, "saved_name = ? AND status = ?", savedName, model.FileStatusNormal)
}

// ListByUser 查询用户的文件列表
func (r *fileRepository) ListByUser(ctx context.Context, userID uint, pagination *database.Pagination) ([]model.File, error) {
	query := "uploaded_by = ? AND status = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, userID, model.FileStatusNormal)
}

// ListByCategory 根据分类查询文件列表
func (r *fileRepository) ListByCategory(ctx context.Context, category string, pagination *database.Pagination) ([]model.File, error) {
	query := "category = ? AND status = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, category, model.FileStatusNormal)
}

// ListByUserAndCategory 查询用户特定分类的文件
func (r *fileRepository) ListByUserAndCategory(ctx context.Context, userID uint, category string, pagination *database.Pagination) ([]model.File, error) {
	query := "uploaded_by = ? AND category = ? AND status = ?"
	return r.Repository.FindWithPagination(ctx, pagination, query, userID, category, model.FileStatusNormal)
}

// Search 搜索文件
func (r *fileRepository) Search(ctx context.Context, keyword string, pagination *database.Pagination) ([]model.File, error) {
	var files []model.File

	db := r.Repository.DB().WithContext(ctx).Model(&model.File{})
	db = db.Where("status = ?", model.FileStatusNormal)

	if keyword != "" {
		db = db.Where("original_name LIKE ? OR saved_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
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

	if err := db.Find(&files).Error; err != nil {
		return nil, err
	}

	return files, nil
}

// CountByUser 统计用户文件数量
func (r *fileRepository) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.Repository.DB().WithContext(ctx).
		Model(&model.File{}).
		Where("uploaded_by = ? AND status = ?", userID, model.FileStatusNormal).
		Count(&count).Error
	return count, err
}

// GetUserStorageUsage 获取用户存储空间使用量（字节）
func (r *fileRepository) GetUserStorageUsage(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.Repository.DB().WithContext(ctx).
		Model(&model.File{}).
		Where("uploaded_by = ? AND status = ?", userID, model.FileStatusNormal).
		Select("COALESCE(SUM(size), 0)").
		Scan(&total).Error
	return total, err
}
