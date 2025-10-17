package handler

import (
	"net/http"
	"strings"

	"github.com/cccvno1/nova/internal/service"
	"github.com/cccvno1/nova/pkg/auth"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

// AuthHandler 认证处理器
// 处理用户注册、登录、登出、令牌刷新等认证相关请求
type AuthHandler struct {
	userService *service.UserService
	blacklist   *auth.TokenBlacklist
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(userService *service.UserService, blacklist *auth.TokenBlacklist) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		blacklist:   blacklist,
	}
}

// RegisterRequest 用户注册请求参数
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"` // 用户名：3-50字符
	Email    string `json:"email" validate:"required,email"`           // 邮箱：必须符合邮箱格式
	Password string `json:"password" validate:"required,min=6"`        // 密码：最少6字符
	Nickname string `json:"nickname" validate:"omitempty,max=50"`      // 昵称：可选，最多50字符
}

// LoginRequest 用户登录请求参数
type LoginRequest struct {
	Username string `json:"username" validate:"required"` // 用户名
	Password string `json:"password" validate:"required"` // 密码
}

// RefreshTokenRequest 刷新令牌请求参数
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"` // 刷新令牌
}

// Register godoc
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} response.Response{data=auth.TokenPair} "注册成功，返回访问令牌和刷新令牌"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "用户名或邮箱已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	userReq := &service.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	}

	tokenPair, err := h.userService.Register(c.Request().Context(), userReq)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, response.Response{
		Code:    errors.Success,
		Message: "success",
		Data:    tokenPair,
	})
}

// Login godoc
// @Summary 用户登录
// @Description 使用用户名和密码登录，返回访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=auth.TokenPair} "登录成功，返回访问令牌和刷新令牌"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	tokenPair, err := h.userService.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return err
	}

	return response.Success(c, tokenPair)
}

// RefreshToken godoc
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} response.Response{data=object} "刷新成功，返回新的访问令牌"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "刷新令牌无效或已过期"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	accessToken, err := h.userService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return err
	}

	return response.Success(c, map[string]string{
		"access_token": accessToken,
	})
}

// Logout godoc
// @Summary 用户登出
// @Description 登出当前用户，将访问令牌加入黑名单
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "登出成功"
// @Failure 401 {object} response.Response "未授权或令牌无效"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	// 从Authorization头提取token
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return errors.New(errors.ErrUnauthorized, "")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return errors.New(errors.ErrUnauthorized, "")
	}

	// 将token加入黑名单
	if err := h.blacklist.AddToBlacklist(c.Request().Context(), token); err != nil {
		return errors.Wrap(errors.ErrInternalServer, err)
	}

	return response.Success(c, nil)
}
