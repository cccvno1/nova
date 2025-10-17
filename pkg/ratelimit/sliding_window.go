package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/cccvno1/nova/pkg/cache"
	"github.com/redis/go-redis/v9"
)

// SlidingWindowLimiter 滑动窗口限流器
// 特点：精确限流，避免固定窗口的边界突刺问题
type SlidingWindowLimiter struct {
	limit       int           // 窗口内最大请求数
	window      time.Duration // 时间窗口
	keyPrefix   string        // Redis 键前缀
	redisClient *redis.Client
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
// limit: 时间窗口内允许的最大请求数
// window: 时间窗口大小（如 1分钟、1秒）
func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:       limit,
		window:      window,
		keyPrefix:   "ratelimit:sliding_window",
		redisClient: cache.GetClient(),
	}
}

// Allow 判断是否允许请求
// key: 限流维度的唯一标识
// 返回: 是否允许，当前窗口请求数
func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, int, error) {
	fullKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)
	now := time.Now().UnixMilli()
	windowStart := now - l.window.Milliseconds()

	// Lua 脚本实现滑动窗口（原子性）
	script := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_ms = tonumber(ARGV[4])
		
		-- 删除窗口外的记录
		redis.call('zremrangebyscore', key, 0, window_start)
		
		-- 获取当前窗口内的请求数
		local count = redis.call('zcard', key)
		
		local allowed = 0
		if count < limit then
			-- 添加当前请求
			redis.call('zadd', key, now, now)
			allowed = 1
			count = count + 1
		end
		
		-- 设置过期时间（窗口大小 + 1秒）
		redis.call('pexpire', key, window_ms + 1000)
		
		return {allowed, count}
	`

	result, err := l.redisClient.Eval(ctx, script,
		[]string{fullKey},
		now, windowStart, l.limit, l.window.Milliseconds()).Result()

	if err != nil {
		return false, 0, err
	}

	values := result.([]interface{})
	allowed := values[0].(int64) == 1
	count := int(values[1].(int64))

	return allowed, count, nil
}

// GetCount 获取当前窗口内的请求数
func (l *SlidingWindowLimiter) GetCount(ctx context.Context, key string) (int, error) {
	fullKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)
	now := time.Now().UnixMilli()
	windowStart := now - l.window.Milliseconds()

	// 删除窗口外的记录
	err := l.redisClient.ZRemRangeByScore(ctx, fullKey, "0", fmt.Sprint(windowStart)).Err()
	if err != nil {
		return 0, err
	}

	// 获取当前窗口内的请求数
	count, err := l.redisClient.ZCard(ctx, fullKey).Result()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// Reset 重置限流器
func (l *SlidingWindowLimiter) Reset(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)
	return l.redisClient.Del(ctx, fullKey).Err()
}

// GetRemaining 获取剩余配额
func (l *SlidingWindowLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	count, err := l.GetCount(ctx, key)
	if err != nil {
		return 0, err
	}
	remaining := l.limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}
