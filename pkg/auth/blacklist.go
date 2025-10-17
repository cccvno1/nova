package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/cccvno1/nova/pkg/cache"
)

const (
	tokenBlacklistPrefix = "token:blacklist"
)

// TokenBlacklist Token 黑名单管理
type TokenBlacklist struct {
	jwtAuth *JWTAuth
}

// NewTokenBlacklist 创建 Token 黑名单管理器
func NewTokenBlacklist(jwtAuth *JWTAuth) *TokenBlacklist {
	return &TokenBlacklist{
		jwtAuth: jwtAuth,
	}
}

// AddToBlacklist 将 token 加入黑名单
func (tb *TokenBlacklist) AddToBlacklist(ctx context.Context, token string) error {
	// 解析 token 获取过期时间
	claims, err := tb.jwtAuth.ValidateToken(token)
	if err != nil {
		// 如果 token 已过期，不需要加入黑名单
		if err == ErrExpiredToken {
			return nil
		}
		// 如果 token 无效，也加入黑名单（防止重复验证）
	}

	// 计算剩余有效期
	var ttl time.Duration
	if claims != nil && claims.ExpiresAt != nil {
		ttl = time.Until(claims.ExpiresAt.Time)
		if ttl <= 0 {
			return nil // 已过期，无需加入黑名单
		}
	} else {
		// 如果无法获取过期时间，使用默认值
		ttl = 24 * time.Hour
	}

	// 加入黑名单
	key := fmt.Sprintf("%s:%s", tokenBlacklistPrefix, token)
	return cache.Set(ctx, key, "1", ttl)
}

// IsInBlacklist 检查 token 是否在黑名单中
func (tb *TokenBlacklist) IsInBlacklist(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("%s:%s", tokenBlacklistPrefix, token)
	count, err := cache.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RemoveFromBlacklist 从黑名单中移除 token（管理功能）
func (tb *TokenBlacklist) RemoveFromBlacklist(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s:%s", tokenBlacklistPrefix, token)
	return cache.Del(ctx, key)
}

// AddUserToBlacklist 将用户的所有 token 加入黑名单（强制下线）
func (tb *TokenBlacklist) AddUserToBlacklist(ctx context.Context, userID uint, duration time.Duration) error {
	key := fmt.Sprintf("%s:user:%d", tokenBlacklistPrefix, userID)
	return cache.Set(ctx, key, "1", duration)
}

// IsUserInBlacklist 检查用户是否被强制下线
func (tb *TokenBlacklist) IsUserInBlacklist(ctx context.Context, userID uint) (bool, error) {
	key := fmt.Sprintf("%s:user:%d", tokenBlacklistPrefix, userID)
	count, err := cache.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RemoveUserFromBlacklist 解除用户强制下线
func (tb *TokenBlacklist) RemoveUserFromBlacklist(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("%s:user:%d", tokenBlacklistPrefix, userID)
	return cache.Del(ctx, key)
}
