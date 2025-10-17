package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrTokenClaims  = errors.New("invalid token claims")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Config struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

type Claims struct {
	UserID   uint      `json:"user_id"`
	Username string    `json:"username"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type JWTAuth struct {
	config *Config
}

func NewJWTAuth(config *Config) *JWTAuth {
	if config.AccessTokenDuration == 0 {
		config.AccessTokenDuration = 2 * time.Hour
	}
	if config.RefreshTokenDuration == 0 {
		config.RefreshTokenDuration = 7 * 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "nova"
	}
	return &JWTAuth{config: config}
}

func (j *JWTAuth) GenerateTokenPair(userID uint, username string) (*TokenPair, error) {
	accessToken, err := j.generateToken(userID, username, AccessToken, j.config.AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := j.generateToken(userID, username, RefreshToken, j.config.RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.config.AccessTokenDuration.Seconds()),
	}, nil
}

func (j *JWTAuth) generateToken(userID uint, username string, tokenType TokenType, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

func (j *JWTAuth) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenClaims
	}

	return claims, nil
}

func (j *JWTAuth) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	if claims.Type != RefreshToken {
		return "", ErrInvalidToken
	}

	return j.generateToken(claims.UserID, claims.Username, AccessToken, j.config.AccessTokenDuration)
}

func (j *JWTAuth) GetUserIDFromToken(tokenString string) (uint, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
