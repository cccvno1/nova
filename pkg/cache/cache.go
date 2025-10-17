package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var (
	// ErrCacheMiss 缓存未命中
	ErrCacheMiss = errors.New("cache miss")
	// ErrCacheNil 缓存空值（防穿透标记）
	ErrCacheNil = errors.New("cache nil value")
)

const (
	// DefaultExpiration 默认过期时间
	DefaultExpiration = 5 * time.Minute
	// NilValueExpiration 空值过期时间（防穿透）
	NilValueExpiration = 1 * time.Minute
	// LockExpiration 分布式锁过期时间
	LockExpiration = 10 * time.Second
	// LockRetryDelay 锁重试延迟
	LockRetryDelay = 50 * time.Millisecond
	// LockMaxRetries 锁最大重试次数
	LockMaxRetries = 20
)

// CacheManager 缓存管理器
type CacheManager struct {
	rdb *redis.Client
	sg  singleflight.Group
}

// NewCacheManager 创建缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		rdb: GetClient(),
	}
}

// LoadFunc 从数据库加载数据的函数类型
type LoadFunc func(ctx context.Context) (interface{}, error)

// GetWithCacheAside Cache-Aside 模式获取数据
// 1. 先查缓存
// 2. 缓存未命中，加载数据
// 3. 设置缓存
func (cm *CacheManager) GetWithCacheAside(ctx context.Context, key string, dest interface{}, ttl time.Duration, loadFunc LoadFunc) error {
	// 1. 尝试从缓存获取
	val, err := Get(ctx, key)
	if err == nil {
		// 检查是否为空值标记（防穿透）
		if val == "nil" {
			return ErrCacheNil
		}
		// 反序列化到目标对象
		return json.Unmarshal([]byte(val), dest)
	}

	if err != redis.Nil {
		// Redis 错误，从数据库加载
		data, loadErr := loadFunc(ctx)
		if loadErr != nil {
			return loadErr
		}
		if data != nil {
			jsonData, _ := json.Marshal(data)
			_ = json.Unmarshal(jsonData, dest)
		}
		return nil
	}

	// 2. 缓存未命中，使用 singleflight 防止缓存击穿
	result, err, _ := cm.sg.Do(key, func() (interface{}, error) {
		data, loadErr := loadFunc(ctx)
		if loadErr != nil {
			return nil, loadErr
		}

		// 数据为空，设置空值标记（防穿透）
		if data == nil {
			_ = Set(ctx, key, "nil", NilValueExpiration)
			return nil, ErrCacheNil
		}

		// 序列化数据
		jsonData, err := json.Marshal(data)
		if err != nil {
			return data, err
		}

		// 设置缓存，添加随机过期时间（防雪崩）
		expiration := cm.addJitter(ttl)
		_ = Set(ctx, key, string(jsonData), expiration)

		return data, nil
	})

	if err != nil {
		return err
	}

	// 将 result 赋值给 dest
	if result != nil {
		data, _ := json.Marshal(result)
		return json.Unmarshal(data, dest)
	}

	return nil
}

// addJitter 添加随机时间偏移（防雪崩）
// 在原有 TTL 基础上增加 0-20% 的随机时间
func (cm *CacheManager) addJitter(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		ttl = DefaultExpiration
	}
	jitter := time.Duration(rand.Int63n(int64(ttl) / 5)) // 0-20%
	return ttl + jitter
}

// SetObject 设置对象到缓存
func (cm *CacheManager) SetObject(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if value == nil {
		return Set(ctx, key, "nil", NilValueExpiration)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	expiration := cm.addJitter(ttl)
	return Set(ctx, key, string(data), expiration)
}

// GetObject 从缓存获取对象
func (cm *CacheManager) GetObject(ctx context.Context, key string, dest interface{}) error {
	val, err := Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}

	if val == "nil" {
		return ErrCacheNil
	}

	return json.Unmarshal([]byte(val), dest)
}

// DeleteByPattern 根据模式删除缓存
func (cm *CacheManager) DeleteByPattern(ctx context.Context, pattern string) error {
	fullPattern := BuildKey(pattern)
	iter := cm.rdb.Scan(ctx, 0, fullPattern, 0).Iterator()

	pipe := cm.rdb.Pipeline()
	count := 0

	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
		count++

		// 批量提交，避免阻塞
		if count%100 == 0 {
			if _, err := pipe.Exec(ctx); err != nil {
				return err
			}
			pipe = cm.rdb.Pipeline()
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	// 提交剩余的
	if count%100 != 0 {
		_, err := pipe.Exec(ctx)
		return err
	}

	return nil
}

// Lock 分布式锁
type Lock struct {
	key    string
	token  string
	client *redis.Client
}

// AcquireLock 获取分布式锁
func (cm *CacheManager) AcquireLock(ctx context.Context, key string) (*Lock, error) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	lockKey := BuildKey("lock", key)

	for i := 0; i < LockMaxRetries; i++ {
		ok, err := cm.rdb.SetNX(ctx, lockKey, token, LockExpiration).Result()
		if err != nil {
			return nil, err
		}

		if ok {
			return &Lock{
				key:    lockKey,
				token:  token,
				client: cm.rdb,
			}, nil
		}

		// 等待后重试
		time.Sleep(LockRetryDelay)
	}

	return nil, errors.New("failed to acquire lock")
}

// Release 释放锁
func (l *Lock) Release(ctx context.Context) error {
	// Lua 脚本确保原子性：只有持有锁的才能释放
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return l.client.Eval(ctx, script, []string{l.key}, l.token).Err()
}

// Refresh 刷新锁过期时间
func (l *Lock) Refresh(ctx context.Context, ttl time.Duration) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	return l.client.Eval(ctx, script, []string{l.key}, l.token, int(ttl.Seconds())).Err()
}

// BatchGet 批量获取缓存
func (cm *CacheManager) BatchGet(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = BuildKey(key)
	}

	values, err := cm.rdb.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for i, val := range values {
		if val != nil {
			if str, ok := val.(string); ok && str != "nil" {
				result[keys[i]] = str
			}
		}
	}

	return result, nil
}

// BatchSet 批量设置缓存
func (cm *CacheManager) BatchSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := cm.rdb.Pipeline()
	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			continue
		}
		expiration := cm.addJitter(ttl)
		pipe.Set(ctx, BuildKey(key), string(data), expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}
