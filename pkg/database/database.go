package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DB     *gorm.DB
	config *config.DBConfig
}

var db *Database

func Init(cfg *config.DBConfig) error {
	dsn := cfg.GetDSN()

	gormConfig := &gorm.Config{
		Logger:                 newGormLogger(),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db = &Database{
		DB:     gormDB,
		config: cfg,
	}

	logger.Info("database connected",
		slog.String("driver", cfg.Driver),
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.DBName))

	return nil
}

func GetDB() *gorm.DB {
	if db == nil {
		panic("database not initialized")
	}
	return db.DB
}

// DB 返回 Database 实例
func DB() *Database {
	if db == nil {
		panic("database not initialized")
	}
	return db
}

func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}

	logger.Info("closing database connection")
	return sqlDB.Close()
}

func WithTransaction(fn func(*gorm.DB) error) error {
	return db.DB.Transaction(fn)
}

func WithContext(ctx context.Context) *gorm.DB {
	return db.DB.WithContext(ctx)
}

func AutoMigrate(models ...interface{}) error {
	logger.Info("running auto migration", slog.Int("models", len(models)))
	return db.DB.AutoMigrate(models...)
}

func HealthCheck() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
