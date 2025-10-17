package database

import (
	"gorm.io/gorm"
)

type Scope func(*gorm.DB) *gorm.DB

func OrderBy(field string, desc bool) Scope {
	return func(db *gorm.DB) *gorm.DB {
		order := field
		if desc {
			order += " DESC"
		}
		return db.Order(order)
	}
}

func Preload(fields ...string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		for _, field := range fields {
			db = db.Preload(field)
		}
		return db
	}
}

func Select(fields ...string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(fields)
	}
}

func Omit(fields ...string) Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Omit(fields...)
	}
}

func Unscoped() Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}
}

func WithDeleted() Scope {
	return Unscoped()
}

func OnlyDeleted() Scope {
	return func(db *gorm.DB) *gorm.DB {
		return db.Unscoped().Where("deleted_at IS NOT NULL")
	}
}
