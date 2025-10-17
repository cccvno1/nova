package repository

import (
	"context"
	"time"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/pkg/database"
	"gorm.io/gorm"
)

// UserRepository 用户仓储
type UserRepository struct {
	base *database.Repository[model.User]
	repo interface {
		Create(ctx context.Context, entity *model.User) error
		Update(ctx context.Context, entity *model.User) error
		UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
		Delete(ctx context.Context, id uint) error
		FindByID(ctx context.Context, id uint) (*model.User, error)
		FindOne(ctx context.Context, query interface{}, args ...interface{}) (*model.User, error)
		FindByCondition(ctx context.Context, query interface{}, args ...interface{}) ([]model.User, error)
		FindWithPagination(ctx context.Context, pagination *database.Pagination, query interface{}, args ...interface{}) ([]model.User, error)
		Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error)
	}
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
// enableCache: 是否启用缓存
func NewUserRepository(db *gorm.DB, enableCache bool) *UserRepository {
	baseRepo := database.NewRepository[model.User](db)

	repo := &UserRepository{
		base: baseRepo,
		db:   db,
	}

	// 按需装饰缓存层
	if enableCache {
		cachedRepo := WithCache(baseRepo, &CacheConfig{
			Enabled:   true,
			TTL:       30 * time.Minute,
			KeyPrefix: "user",
		}, DefaultCacheStrategy())

		repo.repo = cachedRepo
	} else {
		repo.repo = baseRepo
	}

	return repo
}

// === 代理基础 CRUD 方法 ===

func (r *UserRepository) Create(ctx context.Context, entity *model.User) error {
	return r.repo.Create(ctx, entity)
}

func (r *UserRepository) Update(ctx context.Context, entity *model.User) error {
	return r.repo.Update(ctx, entity)
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.repo.Delete(ctx, id)
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	return r.repo.FindByID(ctx, id)
}

func (r *UserRepository) FindWithPagination(ctx context.Context, pagination *database.Pagination, query interface{}, args ...interface{}) ([]model.User, error) {
	return r.repo.FindWithPagination(ctx, pagination, query, args...)
}

// === 业务专属方法 ===

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.repo.FindOne(ctx, "username = ?", username)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.repo.FindOne(ctx, "email = ?", email)
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.repo.Exists(ctx, "username = ?", username)
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.repo.Exists(ctx, "email = ?", email)
}

func (r *UserRepository) FindActiveUsers(ctx context.Context, pagination *database.Pagination) ([]model.User, error) {
	return r.repo.FindWithPagination(ctx, pagination, "status = ?", 1)
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uint, status int) error {
	return r.repo.UpdateFields(ctx, id, map[string]interface{}{"status": status})
}
