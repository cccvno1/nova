package router

import (
	"os"

	"github.com/cccvno1/nova/internal/handler"
	"github.com/cccvno1/nova/internal/repository"
	"github.com/cccvno1/nova/internal/service"
	"github.com/cccvno1/nova/pkg/auth"
	"github.com/cccvno1/nova/pkg/casbin"
	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/logger"
	"github.com/cccvno1/nova/pkg/middleware"
	"github.com/cccvno1/nova/pkg/storage"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Setup(e *echo.Echo, cfg *config.Config, jwtAuth *auth.JWTAuth, blacklist *auth.TokenBlacklist, enforcer *casbin.Enforcer) {
	// Swagger UI 路由
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	healthHandler := handler.NewHealthHandler()

	db := database.GetDB()

	// 用户服务和处理器
	userService := service.NewUserService(db, jwtAuth)
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(userService, blacklist)

	// RBAC 服务和处理器
	roleRepo := repository.NewRoleRepository(database.DB())
	permRepo := repository.NewPermissionRepository(database.DB())
	userRoleRepo := repository.NewUserRoleRepository(database.DB())
	// 方案A：传入database.DB()实例用于直接操作RBAC表
	rbacService := service.NewRBACService(enforcer, roleRepo, permRepo, userRoleRepo, database.DB(), logger.Logger())

	roleHandler := handler.NewRoleHandler(rbacService)
	permissionHandler := handler.NewPermissionHandler(rbacService)
	userRoleHandler := handler.NewUserRoleHandler(rbacService)

	// 文件上传服务和处理器
	fileStorage, err := storage.NewLocalStorage(cfg.Upload.LocalPath, cfg.Upload.LocalURL)
	if err != nil {
		panic("failed to initialize file storage: " + err.Error())
	}
	fileRepo := repository.NewFileRepository(database.DB())
	fileService := service.NewFileService(fileRepo, fileStorage, &cfg.Upload)
	fileHandler := handler.NewFileHandler(fileService)

	// 任务服务和处理器
	taskRepo := repository.NewTaskRepository(database.DB())
	taskHandler := handler.NewTaskHandler(taskRepo)

	// 审计日志服务和处理器
	auditRepo := repository.NewAuditLogRepository(database.DB())
	auditHandler := handler.NewAuditLogHandler(auditRepo)

	// 审计日志中间件
	auditMiddleware := middleware.NewAuditLogMiddleware(&cfg.AuditLog, database.DB())

	api := e.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// 公开路由（IP 限流：每分钟 100 次）
			publicGroup := v1.Group("", middleware.RateLimit(&middleware.RateLimitConfig{
				Enabled:   cfg.RateLimit.Enabled,
				Algorithm: cfg.RateLimit.Algorithm,
				Limit:     cfg.RateLimit.IPLimit,
				Window:    cfg.RateLimit.IPWindow,
				Dimension: "ip",
			}))
			{
				publicGroup.GET("/health", healthHandler.Check)
				publicGroup.POST("/ping", healthHandler.Ping)

				// 认证相关路由
				authGroup := publicGroup.Group("/auth")
				{
					authGroup.POST("/register", authHandler.Register)
					authGroup.POST("/login", authHandler.Login)
					authGroup.POST("/refresh", authHandler.RefreshToken)
					authGroup.POST("/logout", authHandler.Logout, middleware.Auth(jwtAuth, blacklist))
				}
			}

			// 需要认证的路由（用户限流：每分钟 1000 次）
			// 应用审计日志中间件
			authGroup := v1.Group("",
				middleware.Auth(jwtAuth, blacklist),
				middleware.RateLimit(&middleware.RateLimitConfig{
					Enabled:   cfg.RateLimit.Enabled,
					Algorithm: cfg.RateLimit.Algorithm,
					Limit:     cfg.RateLimit.UserLimit,
					Window:    cfg.RateLimit.UserWindow,
					Dimension: "user",
				}),
				auditMiddleware.Handler(), // 添加审计日志中间件
			)
			{
				// 用户管理路由
				users := authGroup.Group("/users")
				{
					users.POST("", userHandler.Create)
					users.GET("", userHandler.List)
					users.GET("/:id", userHandler.GetByID)
					users.PUT("/:id", userHandler.Update)
					users.DELETE("/:id", userHandler.Delete)
				}

				// 角色管理路由
				roles := authGroup.Group("/roles")
				{
					roles.POST("", roleHandler.CreateRole)
					roles.GET("", roleHandler.ListRoles)
					roles.GET("/search", roleHandler.SearchRoles)
					roles.GET("/:id", roleHandler.GetRole)
					roles.PUT("/:id", roleHandler.UpdateRole)
					roles.DELETE("/:id", roleHandler.DeleteRole)

					// 角色权限管理
					roles.POST("/:id/permissions/update", roleHandler.UpdatePermissions) // 新API：支持预览和执行
					roles.POST("/:id/permissions", roleHandler.AssignPermissions)        // 旧API：保留向后兼容
					roles.DELETE("/:id/permissions", roleHandler.RevokePermission)
					roles.GET("/:id/permissions", roleHandler.GetRolePermissions)
					roles.GET("/:id/users", roleHandler.GetRoleUsers)
				}

				// 权限管理路由
				permissions := authGroup.Group("/permissions")
				{
					permissions.POST("", permissionHandler.CreatePermission)
					permissions.GET("", permissionHandler.ListPermissions)
					permissions.GET("/tree", permissionHandler.ListPermissionsTree)
					permissions.GET("/search", permissionHandler.SearchPermissions)
					permissions.GET("/type/:type", permissionHandler.ListPermissionsByType)
					permissions.GET("/:id", permissionHandler.GetPermission)
					permissions.PUT("/:id", permissionHandler.UpdatePermission)
					permissions.DELETE("/:id", permissionHandler.DeletePermission)
				}

				// 用户角色管理路由
				userRoles := authGroup.Group("/user-roles")
				{
					userRoles.POST("", userRoleHandler.AssignRolesToUser)
					userRoles.DELETE("", userRoleHandler.RevokeRolesFromUser)
					userRoles.GET("/user/:userId", userRoleHandler.GetUserRoles)
					userRoles.GET("/user/:userId/permissions", userRoleHandler.GetUserPermissions)
					userRoles.POST("/check", userRoleHandler.CheckUserPermission)
				}

				// 文件管理路由
				files := authGroup.Group("/files")
				{
					files.POST("/upload", fileHandler.Upload)
					files.POST("/upload/avatar", fileHandler.UploadAvatar)
					files.GET("", fileHandler.List)
					files.GET("/search", fileHandler.Search)
					files.GET("/storage-info", fileHandler.GetStorageInfo)
					files.GET("/:id", fileHandler.GetByID)
					files.GET("/:id/download", fileHandler.Download)
					files.DELETE("/:id", fileHandler.Delete)
				}

				// 任务管理路由
				tasks := authGroup.Group("/tasks")
				{
					tasks.GET("", taskHandler.List)
					tasks.GET("/stats", taskHandler.GetStats)
					tasks.GET("/:id", taskHandler.GetByID)
					tasks.GET("/task/:taskId", taskHandler.GetByTaskID)
				}

				// 审计日志路由
				auditLogs := authGroup.Group("/audit-logs")
				{
					auditLogs.GET("", auditHandler.List)
					auditLogs.GET("/stats", auditHandler.GetStats)
					auditLogs.GET("/stats/actions", auditHandler.GetActionStats)
					auditLogs.GET("/stats/users", auditHandler.GetUserStats)
					auditLogs.GET("/stats/resources", auditHandler.GetResourceStats)
					auditLogs.GET("/user/:userId", auditHandler.ListByUser)
					auditLogs.GET("/:id", auditHandler.GetByID)
					auditLogs.DELETE("/clean", auditHandler.CleanOldLogs) // 清理旧日志（需要管理员权限）
				}
			}
		}
	}

	// 静态文件服务（用于本地存储的文件访问）
	e.Static("/uploads", cfg.Upload.LocalPath)

	// 前端静态文件服务（生产环境）
	// 开发环境下前端由 Vite 独立服务，生产环境由后端统一服务
	distPath := "dist"
	if _, err := os.Stat(distPath); err == nil {
		// dist 目录存在，说明是生产环境
		e.Static("/assets", distPath+"/assets")
		e.File("/", distPath+"/index.html")
		// 处理 SPA 路由，所有未匹配的路由都返回 index.html
		e.RouteNotFound("/*", func(c echo.Context) error {
			return c.File(distPath + "/index.html")
		})
	}
}
