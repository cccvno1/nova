package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var (
	rdb    *redis.Client
	ctx    = context.Background()
	prefix = "nova:"
)

// Init 初始化 Redis 连接
func Init(cfg *config.RedisConfig) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	logger.Info("redis connected",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.Int("db", cfg.DB))

	return nil
}

// GetClient 获取 Redis 客户端
func GetClient() *redis.Client {
	return rdb
}

// Close 关闭 Redis 连接
func Close() error {
	if rdb == nil {
		return nil
	}
	logger.Info("closing redis connection")
	return rdb.Close()
}

// SetPrefix 设置全局键前缀
func SetPrefix(p string) {
	prefix = p
}

// GetPrefix 获取全局键前缀
func GetPrefix() string {
	return prefix
}

// BuildKey 构建带前缀的键
func BuildKey(keys ...string) string {
	key := prefix
	for _, k := range keys {
		key += k + ":"
	}
	// 去掉最后一个冒号
	if len(key) > 0 && key[len(key)-1] == ':' {
		key = key[:len(key)-1]
	}
	return key
}

// HealthCheck Redis 健康检查
func HealthCheck() error {
	if rdb == nil {
		return fmt.Errorf("redis not initialized")
	}
	return rdb.Ping(ctx).Err()
}

// Set 设置键值对
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rdb.Set(ctx, BuildKey(key), value, expiration).Err()
}

// Get 获取键值
func Get(ctx context.Context, key string) (string, error) {
	return rdb.Get(ctx, BuildKey(key)).Result()
}

// GetDel 获取并删除键值
func GetDel(ctx context.Context, key string) (string, error) {
	return rdb.GetDel(ctx, BuildKey(key)).Result()
}

// Exists 判断键是否存在
func Exists(ctx context.Context, keys ...string) (int64, error) {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = BuildKey(key)
	}
	return rdb.Exists(ctx, fullKeys...).Result()
}

// Del 删除键
func Del(ctx context.Context, keys ...string) error {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = BuildKey(key)
	}
	return rdb.Del(ctx, fullKeys...).Err()
}

// Expire 设置键过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rdb.Expire(ctx, BuildKey(key), expiration).Err()
}

// TTL 获取键剩余过期时间
func TTL(ctx context.Context, key string) (time.Duration, error) {
	return rdb.TTL(ctx, BuildKey(key)).Result()
}

// Incr 自增
func Incr(ctx context.Context, key string) (int64, error) {
	return rdb.Incr(ctx, BuildKey(key)).Result()
}

// Decr 自减
func Decr(ctx context.Context, key string) (int64, error) {
	return rdb.Decr(ctx, BuildKey(key)).Result()
}

// SetNX 仅当键不存在时设置（分布式锁）
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return rdb.SetNX(ctx, BuildKey(key), value, expiration).Result()
}

// HSet 设置哈希字段
func HSet(ctx context.Context, key string, values ...interface{}) error {
	return rdb.HSet(ctx, BuildKey(key), values...).Err()
}

// HGet 获取哈希字段
func HGet(ctx context.Context, key, field string) (string, error) {
	return rdb.HGet(ctx, BuildKey(key), field).Result()
}

// HGetAll 获取所有哈希字段
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return rdb.HGetAll(ctx, BuildKey(key)).Result()
}

// HDel 删除哈希字段
func HDel(ctx context.Context, key string, fields ...string) error {
	return rdb.HDel(ctx, BuildKey(key), fields...).Err()
}

// HExists 判断哈希字段是否存在
func HExists(ctx context.Context, key, field string) (bool, error) {
	return rdb.HExists(ctx, BuildKey(key), field).Result()
}

// SAdd 添加集合成员
func SAdd(ctx context.Context, key string, members ...interface{}) error {
	return rdb.SAdd(ctx, BuildKey(key), members...).Err()
}

// SRem 移除集合成员
func SRem(ctx context.Context, key string, members ...interface{}) error {
	return rdb.SRem(ctx, BuildKey(key), members...).Err()
}

// SIsMember 判断是否为集合成员
func SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return rdb.SIsMember(ctx, BuildKey(key), member).Result()
}

// SMembers 获取集合所有成员
func SMembers(ctx context.Context, key string) ([]string, error) {
	return rdb.SMembers(ctx, BuildKey(key)).Result()
}

// ZAdd 添加有序集合成员
func ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return rdb.ZAdd(ctx, BuildKey(key), members...).Err()
}

// ZRem 删除有序集合成员
func ZRem(ctx context.Context, key string, members ...interface{}) error {
	return rdb.ZRem(ctx, BuildKey(key), members...).Err()
}

// ZScore 获取有序集合成员分数
func ZScore(ctx context.Context, key string, member string) (float64, error) {
	return rdb.ZScore(ctx, BuildKey(key), member).Result()
}

// ZRange 获取有序集合指定区间成员
func ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rdb.ZRange(ctx, BuildKey(key), start, stop).Result()
}

// Pipeline 创建管道
func Pipeline() redis.Pipeliner {
	return rdb.Pipeline()
}

// TxPipeline 创建事务管道
func TxPipeline() redis.Pipeliner {
	return rdb.TxPipeline()
}

// LPush 从列表左侧插入元素
func LPush(ctx context.Context, key string, values ...interface{}) error {
	return rdb.LPush(ctx, BuildKey(key), values...).Err()
}

// RPush 从列表右侧插入元素
func RPush(ctx context.Context, key string, values ...interface{}) error {
	return rdb.RPush(ctx, BuildKey(key), values...).Err()
}

// LPop 从列表左侧弹出元素
func LPop(ctx context.Context, key string) (string, error) {
	return rdb.LPop(ctx, BuildKey(key)).Result()
}

// RPop 从列表右侧弹出元素
func RPop(ctx context.Context, key string) (string, error) {
	return rdb.RPop(ctx, BuildKey(key)).Result()
}

// BRPop 阻塞式从列表右侧弹出元素
func BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = BuildKey(key)
	}
	return rdb.BRPop(ctx, timeout, fullKeys...).Result()
}

// LLen 获取列表长度
func LLen(ctx context.Context, key string) (int64, error) {
	return rdb.LLen(ctx, BuildKey(key)).Result()
}

// LRange 获取列表指定区间元素
func LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rdb.LRange(ctx, BuildKey(key), start, stop).Result()
}
