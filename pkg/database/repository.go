package database

import (
	"context"

	"gorm.io/gorm"
)

type Repository[T any] struct {
	db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *Repository[T]) CreateBatch(ctx context.Context, entities []T) error {
	return r.db.WithContext(ctx).Create(&entities).Error
}

func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *Repository[T]) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(fields).Error
}

func (r *Repository[T]) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(new(T), id).Error
}

func (r *Repository[T]) DeleteBatch(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(new(T), ids).Error
}

func (r *Repository[T]) HardDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(new(T), id).Error
}

func (r *Repository[T]) FindByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Repository[T]) FindOne(ctx context.Context, query interface{}, args ...interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Where(query, args...).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Repository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

func (r *Repository[T]) FindByCondition(ctx context.Context, query interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Where(query, args...).Find(&entities).Error
	return entities, err
}

func (r *Repository[T]) FindWithPagination(ctx context.Context, pagination *Pagination, query interface{}, args ...interface{}) ([]T, error) {
	var entities []T

	db := r.db.WithContext(ctx).Model(new(T))
	if query != nil {
		db = db.Where(query, args...)
	}

	if err := db.Count(&pagination.Total).Error; err != nil {
		return nil, err
	}

	err := db.Scopes(Paginate(pagination)).Find(&entities).Error
	return entities, err
}

func (r *Repository[T]) Count(ctx context.Context, query interface{}, args ...interface{}) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(new(T))
	if query != nil {
		db = db.Where(query, args...)
	}
	err := db.Count(&count).Error
	return count, err
}

func (r *Repository[T]) Exists(ctx context.Context, query interface{}, args ...interface{}) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&count).Error
	return count > 0, err
}

func (r *Repository[T]) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *Repository[T]) DB() *gorm.DB {
	return r.db
}

func (r *Repository[T]) WithDB(db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}
