package handler

import (
	"net/http"
	"strconv"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/internal/service"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/middleware"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

// RoleHandler 角色管理处理器
type RoleHandler struct {
	rbacService service.RBACService
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(rbacService service.RBACService) *RoleHandler {
	return &RoleHandler{
		rbacService: rbacService,
	}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	Domain      string `json:"domain" validate:"required,min=1,max=100"`
	Category    string `json:"category" validate:"omitempty,max=50"`
	Sort        int    `json:"sort"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	Category    string `json:"category" validate:"omitempty,max=50"`
	Sort        int    `json:"sort"`
	Status      *int8  `json:"status" validate:"omitempty,oneof=0 1"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" validate:"required,min=1"`
}

// UpdatePermissionsRequest 更新权限请求（支持预览）
type UpdatePermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" validate:"required"`
	Preview       bool   `json:"preview"` // true=仅预览，false=执行更新
}

// PermissionDiff 权限差异
type PermissionDiff struct {
	Added   []model.Permission `json:"added"`   // 将要添加的权限
	Removed []model.Permission `json:"removed"` // 将要删除的权限
	Kept    []model.Permission `json:"kept"`    // 将要保留的权限
}

// UpdatePermissionsResponse 更新权限响应
type UpdatePermissionsResponse struct {
	Preview *PermissionDiff `json:"preview,omitempty"` // 预览模式返回差异
	Result  *ChangeResult   `json:"result,omitempty"`  // 执行模式返回结果
}

// ChangeResult 变更结果
type ChangeResult struct {
	AddedCount   int `json:"added_count"`   // 添加的权限数量
	RemovedCount int `json:"removed_count"` // 删除的权限数量
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c echo.Context) error {
	var req CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// 获取当前用户ID

	role := &model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Domain:      req.Domain,
		Category:    req.Category,
		Sort:        req.Sort,
		Status:      1,
	}

	if err := h.rbacService.CreateRole(c.Request().Context(), role); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "角色创建成功", role)
}

// UpdateRole 更新角色
func (h *RoleHandler) UpdateRole(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid role id")
	}

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// 获取现有角色
	role, err := h.rbacService.GetRole(c.Request().Context(), uint(id))
	if err != nil {
		return errors.New(errors.ErrNotFound, "角色不存在")
	}

	// 🔒 安全检查：只能修改比自己等级低的角色
	operatorID := middleware.GetUserID(c)
	if operatorID > 0 {
		if err := h.rbacService.CheckRoleLevelPermission(c.Request().Context(), operatorID, uint(id), role.Domain); err != nil {
			return errors.New(errors.ErrForbidden, err.Error())
		}
	}

	// 更新字段
	role.DisplayName = req.DisplayName
	role.Description = req.Description
	role.Category = req.Category
	role.Sort = req.Sort
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := h.rbacService.UpdateRole(c.Request().Context(), role); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "角色更新成功", role)
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid role id")
	}

	// 检查是否是系统角色
	role, err := h.rbacService.GetRole(c.Request().Context(), uint(id))
	if err != nil {
		return errors.New(errors.ErrNotFound, "角色不存在")
	}

	if role.IsSystem {
		return errors.New(errors.ErrInvalidParams, "系统角色不能删除")
	}

	// 🔒 安全检查：只能删除比自己等级低的角色
	operatorID := middleware.GetUserID(c)
	if operatorID > 0 {
		if err := h.rbacService.CheckRoleLevelPermission(c.Request().Context(), operatorID, uint(id), role.Domain); err != nil {
			return errors.New(errors.ErrForbidden, err.Error())
		}
	}

	if err := h.rbacService.DeleteRole(c.Request().Context(), uint(id)); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "角色删除成功", nil)
}

// GetRole 获取角色详情
func (h *RoleHandler) GetRole(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid role id")
	}

	role, err := h.rbacService.GetRole(c.Request().Context(), uint(id))
	if err != nil {
		return errors.New(errors.ErrNotFound, "角色不存在")
	}

	return response.Success(c, role)
}

// ListRoles 获取角色列表
func (h *RoleHandler) ListRoles(c echo.Context) error {
	domain := c.QueryParam("domain")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	pagination := &database.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// 🔒 使用带等级过滤的查询，只返回可管理的角色
	var roles []model.Role
	var err error

	// 使用middleware提供的辅助函数获取user_id
	operatorID := middleware.GetUserID(c)

	if operatorID > 0 {
		// 成功获取到operatorID，使用过滤查询
		roles, err = h.rbacService.ListRolesFiltered(c.Request().Context(), operatorID, domain, pagination)
	} else {
		// ⚠️ 无法获取user_id（未认证或token无效），返回空列表
		roles = []model.Role{}
	}

	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"items":     roles,
			"total":     pagination.Total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// SearchRoles 搜索角色
func (h *RoleHandler) SearchRoles(c echo.Context) error {
	keyword := c.QueryParam("keyword")
	domain := c.QueryParam("domain")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	pagination := &database.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// 🔒 使用带等级过滤的搜索，只返回可管理的角色
	var roles []model.Role
	var err error

	// 使用middleware提供的辅助函数获取user_id
	operatorID := middleware.GetUserID(c)

	if operatorID > 0 {
		roles, err = h.rbacService.SearchRolesFiltered(c.Request().Context(), operatorID, keyword, domain, pagination)
	} else {
		// 如果无法获取用户ID，返回空列表（安全起见）
		roles = []model.Role{}
	}

	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"items":     roles,
			"total":     pagination.Total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetRolePermissions 获取角色的权限列表
func (h *RoleHandler) GetRolePermissions(c echo.Context) error {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid role id")
	}

	// 获取角色信息以确定domain
	role, err := h.rbacService.GetRole(c.Request().Context(), uint(roleID))
	if err != nil {
		return errors.New(errors.ErrNotFound, "角色不存在")
	}

	permissions, err := h.rbacService.GetRolePermissions(c.Request().Context(), uint(roleID), role.Domain)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.Success(c, permissions)
}

// UpdatePermissions 更新角色权限（支持预览和执行）
func (h *RoleHandler) UpdatePermissions(c echo.Context) error {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid role id")
	}

	var req UpdatePermissionsRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// 获取角色信息
	role, err := h.rbacService.GetRole(c.Request().Context(), uint(roleID))
	if err != nil {
		return errors.New(errors.ErrNotFound, "角色不存在")
	}

	// 🔒 安全检查：只能修改比自己等级低的角色的权限
	operatorID := middleware.GetUserID(c)
	if operatorID > 0 {
		if err := h.rbacService.CheckRoleLevelPermission(c.Request().Context(), operatorID, uint(roleID), role.Domain); err != nil {
			return errors.New(errors.ErrForbidden, err.Error())
		}
	}

	// 调用Service层处理
	result, err := h.rbacService.UpdateRolePermissions(
		c.Request().Context(),
		uint(roleID),
		req.PermissionIDs,
		role.Domain,
		req.Preview,
	)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	// 根据是否预览返回不同的响应
	if req.Preview {
		return response.SuccessWithMessage(c, "权限变更预览", result)
	}
	return response.SuccessWithMessage(c, "权限更新成功", result)
}

// AssignPermissions 给角色分配权限（已废弃，保留向后兼容）
// @Deprecated 请使用 UpdatePermissions 替代
func (h *RoleHandler) AssignPermissions(c echo.Context) error {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid role id")
	}

	var req AssignPermissionsRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// 获取角色信息
	role, err := h.rbacService.GetRole(c.Request().Context(), uint(roleID))
	if err != nil {
		return errors.New(errors.ErrNotFound, "角色不存在")
	}

	// 使用新的UpdatePermissions方法，不预览直接执行
	_, err = h.rbacService.UpdateRolePermissions(
		c.Request().Context(),
		uint(roleID),
		req.PermissionIDs,
		role.Domain,
		false,
	)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "权限分配成功", nil)
}

// RevokePermission 撤销角色的权限
func (h *RoleHandler) RevokePermission(c echo.Context) error {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid role id")
	}

	permissionID, err := strconv.ParseUint(c.Param("permission_id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid permission id")
	}

	// 获取角色信息
	role, err := h.rbacService.GetRole(c.Request().Context(), uint(roleID))
	if err != nil {
		return errors.New(errors.ErrNotFound, "角色不存在")
	}

	if err := h.rbacService.RevokePermissionsFromRole(c.Request().Context(), uint(roleID), []uint{uint(permissionID)}, role.Domain); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "权限撤销成功", nil)
}

// GetRoleUsers 获取拥有某个角色的用户列表
func (h *RoleHandler) GetRoleUsers(c echo.Context) error {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid role id")
	}

	userRoles, err := h.rbacService.GetRoleUsers(c.Request().Context(), uint(roleID))
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.Success(c, userRoles)
}
