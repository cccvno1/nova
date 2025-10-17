package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/ratelimit"
	"github.com/labstack/echo/v4"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled   bool                      // 是否启用
	Algorithm string                    // 算法：token_bucket, sliding_window
	Limit     int                       // 限制数量
	Window    int                       // 时间窗口（秒）
	Dimension string                    // 限流维度：ip, user, api
	Skipper   func(c echo.Context) bool // 跳过规则
}

// DefaultRateLimitConfig 默认配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:   true,
		Algorithm: "sliding_window",
		Limit:     100,
		Window:    60,
		Dimension: "ip",
		Skipper:   nil,
	}
}

// RateLimit 限流中间件
func RateLimit(config *RateLimitConfig) echo.MiddlewareFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	// 创建限流器
	var limiter interface {
		Allow(ctx context.Context, key string) (bool, int, error)
		GetRemaining(ctx context.Context, key string) (int, error)
	}

	switch config.Algorithm {
	case "token_bucket":
		// 令牌桶：capacity = limit, rate = limit/window
		capacity := config.Limit
		rate := config.Limit / config.Window
		if rate < 1 {
			rate = 1
		}
		limiter = ratelimit.NewTokenBucketLimiter(capacity, rate)
	case "sliding_window":
		// 滑动窗口
		window := time.Duration(config.Window) * time.Second
		limiter = ratelimit.NewSlidingWindowLimiter(config.Limit, window)
	default:
		// 默认使用滑动窗口
		window := time.Duration(config.Window) * time.Second
		limiter = ratelimit.NewSlidingWindowLimiter(config.Limit, window)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 跳过检查
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			if !config.Enabled {
				return next(c)
			}

			// 构建限流键
			key := buildRateLimitKey(c, config.Dimension)

			// 检查限流
			allowed, current, err := limiter.Allow(c.Request().Context(), key)
			if err != nil {
				// Redis 错误不影响正常请求
				return next(c)
			}

			// 设置响应头
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprint(config.Limit))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprint(config.Limit-current))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprint(time.Now().Add(time.Duration(config.Window)*time.Second).Unix()))

			if !allowed {
				return errors.New(errors.ErrTooManyRequests, "rate limit exceeded")
			}

			return next(c)
		}
	}
}

// buildRateLimitKey 构建限流键
func buildRateLimitKey(c echo.Context, dimension string) string {
	switch dimension {
	case "ip":
		return getRealIP(c)
	case "user":
		// 从上下文获取用户 ID
		userID := GetUserID(c)
		if userID == 0 {
			return getRealIP(c) // 未登录用户使用 IP
		}
		return fmt.Sprintf("user:%d", userID)
	case "api":
		// API 路径 + IP
		return fmt.Sprintf("%s:%s", c.Request().URL.Path, getRealIP(c))
	default:
		return getRealIP(c)
	}
}

// getRealIP 获取真实 IP
func getRealIP(c echo.Context) string {
	// 优先从 X-Forwarded-For 获取
	xff := c.Request().Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 其次从 X-Real-IP 获取
	xri := c.Request().Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 最后使用 RemoteAddr
	ip := c.Request().RemoteAddr
	// 移除端口号
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// RateLimitByIP IP 限流（快捷方式）
func RateLimitByIP(limit, window int) echo.MiddlewareFunc {
	return RateLimit(&RateLimitConfig{
		Enabled:   true,
		Algorithm: "sliding_window",
		Limit:     limit,
		Window:    window,
		Dimension: "ip",
	})
}

// RateLimitByUser 用户限流（快捷方式）
func RateLimitByUser(limit, window int) echo.MiddlewareFunc {
	return RateLimit(&RateLimitConfig{
		Enabled:   true,
		Algorithm: "sliding_window",
		Limit:     limit,
		Window:    window,
		Dimension: "user",
	})
}

// RateLimitByAPI API 限流（快捷方式）
func RateLimitByAPI(limit, window int) echo.MiddlewareFunc {
	return RateLimit(&RateLimitConfig{
		Enabled:   true,
		Algorithm: "sliding_window",
		Limit:     limit,
		Window:    window,
		Dimension: "api",
	})
}
