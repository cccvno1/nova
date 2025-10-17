package middleware

import (
	"strings"

	"github.com/cccvno1/nova/pkg/auth"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/labstack/echo/v4"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UsernameKey         = "username"
)

func Auth(jwtAuth *auth.JWTAuth, blacklist *auth.TokenBlacklist) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(AuthorizationHeader)
			if authHeader == "" {
				return errors.New(errors.ErrUnauthorized, "")
			}

			if !strings.HasPrefix(authHeader, BearerPrefix) {
				return errors.New(errors.ErrUnauthorized, "")
			}

			tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
			if tokenString == "" {
				return errors.New(errors.ErrUnauthorized, "")
			}

			claims, err := jwtAuth.ValidateToken(tokenString)
			if err != nil {
				if err == auth.ErrExpiredToken {
					return errors.New(errors.ErrTokenExpired, "")
				}
				return errors.New(errors.ErrUnauthorized, "")
			}

			if claims.Type != auth.AccessToken {
				return errors.New(errors.ErrUnauthorized, "")
			}

			// 检查 token 黑名单
			if blacklist != nil {
				inBlacklist, err := blacklist.IsInBlacklist(c.Request().Context(), tokenString)
				if err == nil && inBlacklist {
					return errors.New(errors.ErrUnauthorized, "token has been revoked")
				}

				// 检查用户是否被强制下线
				userBlacklisted, err := blacklist.IsUserInBlacklist(c.Request().Context(), claims.UserID)
				if err == nil && userBlacklisted {
					return errors.New(errors.ErrUnauthorized, "user has been logged out")
				}
			}

			c.Set(UserIDKey, claims.UserID)
			c.Set(UsernameKey, claims.Username)

			return next(c)
		}
	}
}

func GetUserID(c echo.Context) uint {
	userID, ok := c.Get(UserIDKey).(uint)
	if !ok {
		return 0
	}
	return userID
}

func GetUsername(c echo.Context) string {
	username, ok := c.Get(UsernameKey).(string)
	if !ok {
		return ""
	}
	return username
}
