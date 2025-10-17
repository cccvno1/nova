package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cccvno1/nova/pkg/casbin"
	"github.com/labstack/echo/v4"
)

// PermissionConfig 权限中间件配置
type PermissionConfig struct {
	Enforcer *casbin.Enforcer          // Casbin enforcer
	Domain   string                    // 默认域（可选）
	Skipper  func(c echo.Context) bool // 跳过某些路径
	Logger   *slog.Logger              // 日志记录器
}

// Permission 创建权限验证中间件（自动根据 HTTP 方法映射权限）
func Permission(config PermissionConfig) echo.MiddlewareFunc {
	if config.Enforcer == nil {
		panic("casbin enforcer is required")
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	if config.Skipper == nil {
		config.Skipper = func(c echo.Context) bool { return false }
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 跳过某些路径
			if config.Skipper(c) {
				return next(c)
			}

			// 获取用户 ID（从 JWT 中间件设置的上下文）
			userIDValue := c.Get("user_id")
			if userIDValue == nil {
				config.Logger.Warn("permission check: user_id not found in context")
				return echo.NewHTTPError(http.StatusUnauthorized, "用户未认证")
			}

			userID, ok := userIDValue.(uint)
			if !ok {
				config.Logger.Error("permission check: invalid user_id type", "user_id", userIDValue)
				return echo.NewHTTPError(http.StatusInternalServerError, "内部错误")
			}

			// 获取域（可以从请求头、查询参数或配置中获取）
			domain := getDomain(c, config.Domain)

			// 获取请求的资源和操作
			resource := c.Request().URL.Path
			method := c.Request().Method
			action := mapHTTPMethodToAction(method)

			// 验证权限
			userIDStr := strconv.FormatUint(uint64(userID), 10)
			allowed, err := config.Enforcer.Enforce(userIDStr, domain, resource, action)
			if err != nil {
				config.Logger.Error("permission check failed",
					"user_id", userID,
					"domain", domain,
					"resource", resource,
					"action", action,
					"error", err,
				)
				return echo.NewHTTPError(http.StatusInternalServerError, "权限验证失败")
			}

			if !allowed {
				config.Logger.Warn("permission denied",
					"user_id", userID,
					"domain", domain,
					"resource", resource,
					"action", action,
				)
				return echo.NewHTTPError(http.StatusForbidden, "无权访问该资源")
			}

			// 将域信息存入上下文
			c.Set("domain", domain)

			config.Logger.Debug("permission granted",
				"user_id", userID,
				"domain", domain,
				"resource", resource,
				"action", action,
			)

			return next(c)
		}
	}
}

// RequirePermission 创建指定权限验证中间件
func RequirePermission(config PermissionConfig, resource, action string) echo.MiddlewareFunc {
	if config.Enforcer == nil {
		panic("casbin enforcer is required")
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userIDValue := c.Get("user_id")
			if userIDValue == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "用户未认证")
			}

			userID, ok := userIDValue.(uint)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "内部错误")
			}

			domain := getDomain(c, config.Domain)

			// 验证权限
			userIDStr := strconv.FormatUint(uint64(userID), 10)
			allowed, err := config.Enforcer.Enforce(userIDStr, domain, resource, action)
			if err != nil {
				config.Logger.Error("permission check failed",
					"user_id", userID,
					"domain", domain,
					"resource", resource,
					"action", action,
					"error", err,
				)
				return echo.NewHTTPError(http.StatusInternalServerError, "权限验证失败")
			}

			if !allowed {
				config.Logger.Warn("permission denied",
					"user_id", userID,
					"domain", domain,
					"resource", resource,
					"action", action,
				)
				return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("无权执行 %s 操作", action))
			}

			c.Set("domain", domain)
			return next(c)
		}
	}
}

// CheckPermission 创建权限检查函数（用于在handler中手动检查）
func CheckPermission(enforcer *casbin.Enforcer, logger *slog.Logger) func(c echo.Context, resource, action string) error {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c echo.Context, resource, action string) error {
		userIDValue := c.Get("user_id")
		if userIDValue == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "用户未认证")
		}

		userID, ok := userIDValue.(uint)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "内部错误")
		}

		domain := c.Get("domain")
		if domain == nil {
			domain = "default"
		}
		domainStr, ok := domain.(string)
		if !ok {
			domainStr = "default"
		}

		userIDStr := strconv.FormatUint(uint64(userID), 10)
		allowed, err := enforcer.Enforce(userIDStr, domainStr, resource, action)
		if err != nil {
			logger.Error("permission check failed",
				"user_id", userID,
				"domain", domainStr,
				"resource", resource,
				"action", action,
				"error", err,
			)
			return echo.NewHTTPError(http.StatusInternalServerError, "权限验证失败")
		}

		if !allowed {
			logger.Warn("permission denied",
				"user_id", userID,
				"domain", domainStr,
				"resource", resource,
				"action", action,
			)
			return echo.NewHTTPError(http.StatusForbidden, "无权访问该资源")
		}

		return nil
	}
}

// RequireAnyPermission 要求用户拥有任意一个权限（OR 逻辑）
func RequireAnyPermission(config PermissionConfig, permissions [][2]string) echo.MiddlewareFunc {
	if config.Enforcer == nil {
		panic("casbin enforcer is required")
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userIDValue := c.Get("user_id")
			if userIDValue == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "用户未认证")
			}

			userID := userIDValue.(uint)
			domain := getDomain(c, config.Domain)
			userIDStr := strconv.FormatUint(uint64(userID), 10)

			// 检查是否有任意一个权限
			for _, perm := range permissions {
				resource, action := perm[0], perm[1]
				allowed, err := config.Enforcer.Enforce(userIDStr, domain, resource, action)
				if err != nil {
					config.Logger.Error("permission check failed", "error", err)
					continue
				}
				if allowed {
					c.Set("domain", domain)
					return next(c)
				}
			}

			config.Logger.Warn("permission denied - no matching permission",
				"user_id", userID,
				"domain", domain,
			)

			return echo.NewHTTPError(http.StatusForbidden, "无权访问该资源")
		}
	}
}

// RequireAllPermissions 要求用户拥有所有权限（AND 逻辑）
func RequireAllPermissions(config PermissionConfig, permissions [][2]string) echo.MiddlewareFunc {
	if config.Enforcer == nil {
		panic("casbin enforcer is required")
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userIDValue := c.Get("user_id")
			if userIDValue == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "用户未认证")
			}

			userID := userIDValue.(uint)
			domain := getDomain(c, config.Domain)
			userIDStr := strconv.FormatUint(uint64(userID), 10)

			// 检查是否拥有所有权限
			for _, perm := range permissions {
				resource, action := perm[0], perm[1]
				allowed, err := config.Enforcer.Enforce(userIDStr, domain, resource, action)
				if err != nil {
					config.Logger.Error("permission check failed", "error", err)
					return echo.NewHTTPError(http.StatusInternalServerError, "权限验证失败")
				}
				if !allowed {
					config.Logger.Warn("permission denied - missing permission",
						"user_id", userID,
						"domain", domain,
						"missing_permission", perm,
					)
					return echo.NewHTTPError(http.StatusForbidden, "无权访问该资源")
				}
			}

			c.Set("domain", domain)
			return next(c)
		}
	}
}

// ============================
// 辅助函数
// ============================

// mapHTTPMethodToAction 将 HTTP 方法映射到操作
func mapHTTPMethodToAction(method string) string {
	switch method {
	case http.MethodGet:
		return "read"
	case http.MethodPost:
		return "write"
	case http.MethodPut, http.MethodPatch:
		return "write"
	case http.MethodDelete:
		return "delete"
	default:
		return "read"
	}
}

// getDomain 获取域信息
func getDomain(c echo.Context, defaultDomain string) string {
	// 优先从上下文获取（可能由其他中间件设置）
	if domain := c.Get("domain"); domain != nil {
		if domainStr, ok := domain.(string); ok {
			return domainStr
		}
	}

	// 从请求头获取
	domain := c.Request().Header.Get("X-Domain")
	if domain != "" {
		return domain
	}

	// 从查询参数获取
	domain = c.QueryParam("domain")
	if domain != "" {
		return domain
	}

	// 使用默认域
	if defaultDomain != "" {
		return defaultDomain
	}

	return "default"
}
