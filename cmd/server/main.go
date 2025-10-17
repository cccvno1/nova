package main

import (
	"flag"
	"log"
	"time"

	_ "github.com/cccvno1/nova/docs" // Swagger docs
	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/internal/router"
	"github.com/cccvno1/nova/internal/server"
	"github.com/cccvno1/nova/pkg/auth"
	"github.com/cccvno1/nova/pkg/cache"
	"github.com/cccvno1/nova/pkg/casbin"
	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/logger"
	"github.com/cccvno1/nova/pkg/queue"
	"github.com/cccvno1/nova/pkg/scheduler"
)

// @title Nova API
// @version 1.0
// @description Nova 是一个基于 Go + Echo 的现代化 Web 应用框架，提供用户认证、RBAC 权限管理、文件上传、异步任务队列、审计日志等功能。
// @description
// @description 主要特性：
// @description - 用户认证与授权（JWT）
// @description - RBAC 权限管理（基于 Casbin）
// @description - 文件上传与管理（支持本地/OSS/S3）
// @description - 异步任务队列（基于 Redis）
// @description - 定时任务调度（基于 Cron）
// @description - 审计日志（自动记录所有操作）
// @description - 限流保护（IP/用户维度）
// @description - 健康检查与监控

// @contact.name API Support
// @contact.url https://github.com/cccvno1/nova
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

var configFile = flag.String("config", "", "config file path")

func main() {
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	loggerCfg := &logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		FilePath:   cfg.Logger.FilePath,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
	}

	if err := logger.Init(loggerCfg); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	if err := database.Init(&cfg.DB); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer database.Close()

	if err := cache.Init(&cfg.Redis); err != nil {
		log.Fatalf("failed to initialize redis: %v", err)
	}
	defer cache.Close()

	// 自动迁移数据库表
	if err := database.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.UserRole{},
		&model.File{},
		&model.Task{},
		&model.AuditLog{},
	); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	// 初始化 Casbin enforcer
	enforcer, err := casbin.NewEnforcer(
		database.GetDB(),
		casbin.Config{
			ModelPath:    cfg.Casbin.ModelPath,
			AutoSave:     cfg.Casbin.AutoSave,
			AutoLoad:     cfg.Casbin.AutoLoad,
			AutoLoadTick: time.Duration(cfg.Casbin.AutoLoadTick) * time.Second,
		},
		logger.Logger(),
	)
	if err != nil {
		log.Fatalf("failed to initialize casbin enforcer: %v", err)
	}
	defer enforcer.Close()

	jwtAuth := auth.NewJWTAuth(&auth.Config{
		SecretKey:            cfg.Auth.JWTSecret,
		AccessTokenDuration:  time.Duration(cfg.Auth.AccessTokenDuration) * time.Second,
		RefreshTokenDuration: time.Duration(cfg.Auth.RefreshTokenDuration) * time.Second,
		Issuer:               cfg.Auth.Issuer,
	})

	blacklist := auth.NewTokenBlacklist(jwtAuth)

	// 初始化队列 Worker（如果启用）
	var queueWorker *queue.Worker
	if cfg.Queue.Enabled {
		queueWorker = queue.NewWorker(&cfg.Queue)
		// 可以在这里注册任务处理器
		// queueWorker.Register("example_task", func(task *queue.Task) error {
		// 	// 处理任务逻辑
		// 	return nil
		// })
		if err := queueWorker.Start(); err != nil {
			log.Fatalf("failed to start queue worker: %v", err)
		}
		defer queueWorker.Stop()
	}

	// 初始化调度器
	taskScheduler := scheduler.NewScheduler()
	// 可以在这里添加定时任务
	// taskScheduler.AddFunc("0 0 2 * * *", func() {
	// 	// 每天凌晨2点执行的任务
	// })
	// taskScheduler.AddInterval(30*time.Second, func() {
	// 	// 每30秒执行的任务
	// })
	taskScheduler.Start()
	defer taskScheduler.Stop()

	srv := server.New(cfg)

	router.Setup(srv.Echo(), cfg, jwtAuth, blacklist, enforcer)

	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
