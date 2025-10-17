package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cccvno1/nova/pkg/cache"
	"github.com/cccvno1/nova/pkg/database"
	"gorm.io/gorm"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled   bool          // 是否启用缓存
	TTL       time.Duration // 缓存过期时间
	KeyPrefix string        // 缓存键前缀
}

// CacheStrategy 缓存策略（定义哪些方法需要缓存）
type CacheStrategy struct {
	CacheRead         bool // 是否缓存读操作（FindByID）
	CacheList         bool // 是否缓存列表查询
	InvalidateOnWrite bool // 写操作时是否失效缓存
}

// DefaultCacheStrategy 默认缓存策略
func DefaultCacheStrategy() *CacheStrategy {
	return &CacheStrategy{
		CacheRead:         true,
		CacheList:         false, // 列表默认不缓存
		InvalidateOnWrite: true,
	}
}

// CachedRepository 带缓存的仓储装饰器
// 装饰 database.Repository[T]，为其添加缓存能力
type CachedRepository[T any] struct {
	*database.Repository[T]
	cache    *cache.CacheManager
	config   *CacheConfig
	strategy *CacheStrategy
}

// WithCache 为 database.Repository 添加缓存装饰
func WithCache[T any](baseRepo *database.Repository[T], config *CacheConfig, strategy *CacheStrategy) *CachedRepository[T] {
	if strategy == nil {
		strategy = DefaultCacheStrategy()
	}

	return &CachedRepository[T]{
		Repository: baseRepo,
		cache:      cache.NewCacheManager(),
		config:     config,
		strategy:   strategy,
	}
}

// Create 创建（带缓存失效）
func (r *CachedRepository[T]) Create(ctx context.Context, entity *T) error {
	err := r.Repository.Create(ctx, entity)
	if err != nil {
		return err
	}

	if r.strategy.InvalidateOnWrite {
		go r.invalidateCache(context.Background())
	}

	return nil
}

// Update 更新（带缓存失效）
func (r *CachedRepository[T]) Update(ctx context.Context, entity *T) error {
	err := r.Repository.Update(ctx, entity)
	if err != nil {
		return err
	}

	if r.strategy.InvalidateOnWrite {
		go r.invalidateCache(context.Background())
	}

	return nil
}

// UpdateFields 更新字段（带缓存失效）
func (r *CachedRepository[T]) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	err := r.Repository.UpdateFields(ctx, id, fields)
	if err != nil {
		return err
	}

	if r.strategy.InvalidateOnWrite {
		go r.invalidateCache(context.Background())
	}

	return nil
}

// Delete 删除（带缓存失效）
func (r *CachedRepository[T]) Delete(ctx context.Context, id uint) error {
	err := r.Repository.Delete(ctx, id)
	if err != nil {
		return err
	}

	if r.strategy.InvalidateOnWrite {
		go r.invalidateCache(context.Background())
	}

	return nil
}

// FindByID 根据ID查找（带缓存）
func (r *CachedRepository[T]) FindByID(ctx context.Context, id uint) (*T, error) {
	if !r.strategy.CacheRead {
		return r.Repository.FindByID(ctx, id)
	}

	var entity T
	key := r.buildKey("id", id)

	err := r.cache.GetWithCacheAside(ctx, key, &entity, r.config.TTL, func(ctx context.Context) (interface{}, error) {
		return r.Repository.FindByID(ctx, id)
	})

	if err != nil {
		if err == cache.ErrCacheNil {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &entity, nil
}

// FindWithPagination 分页查询（可选缓存）
func (r *CachedRepository[T]) FindWithPagination(ctx context.Context, pagination *database.Pagination, query interface{}, args ...interface{}) ([]T, error) {
	// 列表查询暂不缓存（查询条件多变，缓存命中率低）
	return r.Repository.FindWithPagination(ctx, pagination, query, args...)
}

// buildKey 构建缓存键
func (r *CachedRepository[T]) buildKey(parts ...interface{}) string {
	key := r.config.KeyPrefix
	for _, part := range parts {
		key += ":" + fmt.Sprint(part)
	}
	return key
}

// invalidateCache 失效缓存（删除该资源的所有缓存）
func (r *CachedRepository[T]) invalidateCache(ctx context.Context) {
	pattern := r.config.KeyPrefix + "*"
	_ = r.cache.DeleteByPattern(ctx, pattern)
}

// InvalidateCacheByID 失效指定 ID 的缓存
func (r *CachedRepository[T]) InvalidateCacheByID(ctx context.Context, id uint) error {
	key := r.buildKey("id", id)
	return cache.Del(ctx, key)
}
