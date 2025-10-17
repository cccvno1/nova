package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/cccvno1/nova/pkg/cache"
	"github.com/redis/go-redis/v9"
)

// TokenBucketLimiter 令牌桶限流器
// 特点：允许突发流量，平滑限流
type TokenBucketLimiter struct {
	capacity    int           // 桶容量（最大令牌数）
	rate        int           // 令牌生成速率（每秒）
	keyPrefix   string        // Redis 键前缀
	ttl         time.Duration // 键过期时间
	redisClient *redis.Client
}

// NewTokenBucketLimiter 创建令牌桶限流器
// capacity: 桶容量，允许的突发请求数
// rate: 每秒生成的令牌数，即稳定速率
func NewTokenBucketLimiter(capacity, rate int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:    capacity,
		rate:        rate,
		keyPrefix:   "ratelimit:token_bucket",
		ttl:         time.Duration(capacity/rate+10) * time.Second, // 确保足够的 TTL
		redisClient: cache.GetClient(),
	}
}

// Allow 判断是否允许请求
// key: 限流维度的唯一标识（如 IP、UserID）
// 返回: 是否允许，剩余令牌数
func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, int, error) {
	fullKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)
	now := time.Now().Unix()

	// Lua 脚本实现令牌桶算法（原子性）
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local ttl = tonumber(ARGV[4])
		
		-- 获取桶信息
		local bucket = redis.call('hmget', key, 'tokens', 'last_time')
		local tokens = tonumber(bucket[1])
		local last_time = tonumber(bucket[2])
		
		-- 初始化桶
		if tokens == nil then
			tokens = capacity
			last_time = now
		else
			-- 计算新增令牌
			local delta = math.max(0, now - last_time)
			local new_tokens = math.min(capacity, tokens + delta * rate)
			tokens = new_tokens
			last_time = now
		end
		
		-- 尝试消费一个令牌
		local allowed = 0
		if tokens >= 1 then
			tokens = tokens - 1
			allowed = 1
		end
		
		-- 更新桶状态
		redis.call('hmset', key, 'tokens', tokens, 'last_time', last_time)
		redis.call('expire', key, ttl)
		
		return {allowed, math.floor(tokens)}
	`

	result, err := l.redisClient.Eval(ctx, script,
		[]string{fullKey},
		l.capacity, l.rate, now, int(l.ttl.Seconds())).Result()

	if err != nil {
		return false, 0, err
	}

	values := result.([]interface{})
	allowed := values[0].(int64) == 1
	remaining := int(values[1].(int64))

	return allowed, remaining, nil
}

// AllowN 判断是否允许消费 N 个令牌
func (l *TokenBucketLimiter) AllowN(ctx context.Context, key string, n int) (bool, int, error) {
	fullKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)
	now := time.Now().Unix()

	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local ttl = tonumber(ARGV[4])
		local cost = tonumber(ARGV[5])
		
		local bucket = redis.call('hmget', key, 'tokens', 'last_time')
		local tokens = tonumber(bucket[1])
		local last_time = tonumber(bucket[2])
		
		if tokens == nil then
			tokens = capacity
			last_time = now
		else
			local delta = math.max(0, now - last_time)
			local new_tokens = math.min(capacity, tokens + delta * rate)
			tokens = new_tokens
			last_time = now
		end
		
		local allowed = 0
		if tokens >= cost then
			tokens = tokens - cost
			allowed = 1
		end
		
		redis.call('hmset', key, 'tokens', tokens, 'last_time', last_time)
		redis.call('expire', key, ttl)
		
		return {allowed, math.floor(tokens)}
	`

	result, err := l.redisClient.Eval(ctx, script,
		[]string{fullKey},
		l.capacity, l.rate, now, int(l.ttl.Seconds()), n).Result()

	if err != nil {
		return false, 0, err
	}

	values := result.([]interface{})
	allowed := values[0].(int64) == 1
	remaining := int(values[1].(int64))

	return allowed, remaining, nil
}

// GetRemaining 获取剩余令牌数
func (l *TokenBucketLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	fullKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)

	bucket, err := l.redisClient.HMGet(ctx, fullKey, "tokens", "last_time").Result()
	if err != nil {
		return 0, err
	}

	if bucket[0] == nil {
		return l.capacity, nil // 未使用过，返回满容量
	}

	tokens, _ := bucket[0].(string)
	lastTime, _ := bucket[1].(string)

	var tokensFloat float64
	var lastTimeInt int64
	fmt.Sscanf(tokens, "%f", &tokensFloat)
	fmt.Sscanf(lastTime, "%d", &lastTimeInt)

	// 计算当前令牌数
	now := time.Now().Unix()
	delta := now - lastTimeInt
	newTokens := tokensFloat + float64(delta*int64(l.rate))
	if newTokens > float64(l.capacity) {
		newTokens = float64(l.capacity)
	}

	return int(newTokens), nil
}

// Reset 重置限流器
func (l *TokenBucketLimiter) Reset(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)
	return l.redisClient.Del(ctx, fullKey).Err()
}
