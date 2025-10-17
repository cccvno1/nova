package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cccvno1/nova/pkg/logger"
	gormLogger "gorm.io/gorm/logger"
)

type gormLoggerAdapter struct {
	SlowThreshold time.Duration
}

func newGormLogger() gormLogger.Interface {
	return &gormLoggerAdapter{
		SlowThreshold: 200 * time.Millisecond,
	}
}

func (l *gormLoggerAdapter) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return l
}

func (l *gormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.Info(fmt.Sprintf(msg, data...))
}

func (l *gormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.Warn(fmt.Sprintf(msg, data...))
}

func (l *gormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.Error(fmt.Sprintf(msg, data...))
}

func (l *gormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		logger.Error("database query error",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed),
			slog.Any("error", err))
		return
	}

	if elapsed > l.SlowThreshold {
		logger.Warn("slow query detected",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed))
		return
	}

	logger.Debug("database query",
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed))
}
