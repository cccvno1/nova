package handler

import (
	"net/http"
	"strconv"

	"github.com/cccvno1/nova/internal/service"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/middleware"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

// UserRoleHandler 用户角色管理处理器
type UserRoleHandler struct {
	rbacService service.RBACService
}

// NewUserRoleHandler 创建用户角色处理器
func NewUserRoleHandler(rbacService service.RBACService) *UserRoleHandler {
	return &UserRoleHandler{
		rbacService: rbacService,
	}
}

// AssignRolesRequest 分配角色请求
type AssignRolesRequest struct {
	RoleIDs []uint `json:"role_ids" validate:"required,min=1"`
	Domain  string `json:"domain" validate:"required,min=1,max=100"`
}

// AssignRolesToUser 给用户分配角色
// POST /api/v1/user-roles/user/:userId/roles
func (h *UserRoleHandler) AssignRolesToUser(c echo.Context) error {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid user id")
	}

	var req AssignRolesRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// 获取当前操作用户ID
	operatorID := middleware.GetUserID(c)

	// 🔒 安全检查：只能给用户分配比自己等级低的角色
	if err := h.rbacService.CheckRolesLevelPermission(c.Request().Context(), operatorID, req.RoleIDs, req.Domain); err != nil {
		return errors.New(errors.ErrForbidden, err.Error())
	}

	if err := h.rbacService.AssignRolesToUser(c.Request().Context(), uint(userID), req.RoleIDs, req.Domain, operatorID); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "角色分配成功", nil)
}

// RevokeRolesRequest 撤销角色请求
type RevokeRolesRequest struct {
	RoleIDs []uint `json:"role_ids" validate:"required,min=1"`
	Domain  string `json:"domain" validate:"required,min=1,max=100"`
}

// RevokeRolesFromUser 撤销用户的角色
func (h *UserRoleHandler) RevokeRolesFromUser(c echo.Context) error {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid user id")
	}

	var req RevokeRolesRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// 获取当前操作用户ID
	operatorID := middleware.GetUserID(c)

	// 🔒 安全检查：只能撤销比自己等级低的角色
	if err := h.rbacService.CheckRolesLevelPermission(c.Request().Context(), operatorID, req.RoleIDs, req.Domain); err != nil {
		return errors.New(errors.ErrForbidden, err.Error())
	}

	if err := h.rbacService.RevokeRolesFromUser(c.Request().Context(), uint(userID), req.RoleIDs, req.Domain); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "角色撤销成功", nil)
}

// GetUserRoles 获取用户的角色列表
// GET /api/v1/user-roles/user/:userId
func (h *UserRoleHandler) GetUserRoles(c echo.Context) error {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid user id")
	}

	domain := c.QueryParam("domain")
	if domain == "" {
		domain = "default"
	}

	roles, err := h.rbacService.GetUserRoles(c.Request().Context(), uint(userID), domain)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.Success(c, roles)
}

// GetUserPermissions 获取用户的权限列表（包括通过角色继承的权限）
// GET /api/v1/user-roles/user/:userId/permissions
// 重要：此接口供前端调用，用于生成动态路由和菜单
func (h *UserRoleHandler) GetUserPermissions(c echo.Context) error {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid user id")
	}

	domain := c.QueryParam("domain")
	if domain == "" {
		domain = "default"
	}

	permissions, err := h.rbacService.GetUserPermissions(c.Request().Context(), uint(userID), domain)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.Success(c, permissions)
}

// CheckUserPermission 检查用户是否拥有指定权限
// POST /api/v1/user-roles/check
func (h *UserRoleHandler) CheckUserPermission(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return errors.New(errors.ErrUnauthorized, "user not authenticated")
	}

	domain := c.QueryParam("domain")
	if domain == "" {
		domain = "default"
	}
	resource := c.QueryParam("resource")
	action := c.QueryParam("action")

	if resource == "" || action == "" {
		return errors.New(errors.ErrInvalidParams, "resource and action are required")
	}

	allowed, err := h.rbacService.CheckPermission(c.Request().Context(), userID, domain, resource, action)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"allowed": allowed,
		},
	})
}
