package handler

import (
	"net/http"
	"strconv"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/internal/service"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

// PermissionHandler 权限管理处理器
type PermissionHandler struct {
	rbacService service.RBACService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(rbacService service.RBACService) *PermissionHandler {
	return &PermissionHandler{
		rbacService: rbacService,
	}
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string               `json:"name" validate:"required,min=2,max=100"`
	DisplayName string               `json:"display_name" validate:"required,min=2,max=100"`
	Description string               `json:"description" validate:"max=500"`
	Type        model.PermissionType `json:"type" validate:"required,oneof=api menu button data field"`
	Domain      string               `json:"domain" validate:"required,min=1,max=100"`
	Resource    string               `json:"resource" validate:"required,max=200"`
	Action      string               `json:"action" validate:"required,max=50"`
	Category    string               `json:"category" validate:"omitempty,max=50"`
	ParentID    uint                 `json:"parent_id"`
	Path        string               `json:"path" validate:"omitempty,max=200"`
	Component   string               `json:"component" validate:"omitempty,max=200"`
	Icon        string               `json:"icon" validate:"omitempty,max=50"`
	Sort        int                  `json:"sort"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	DisplayName string               `json:"display_name" validate:"required,min=2,max=100"`
	Description string               `json:"description" validate:"max=500"`
	Type        model.PermissionType `json:"type" validate:"required,oneof=api menu button data field"`
	Resource    string               `json:"resource" validate:"required,max=200"`
	Action      string               `json:"action" validate:"required,max=50"`
	Category    string               `json:"category" validate:"omitempty,max=50"`
	ParentID    *uint                `json:"parent_id"`
	Path        string               `json:"path" validate:"omitempty,max=200"`
	Component   string               `json:"component" validate:"omitempty,max=200"`
	Icon        string               `json:"icon" validate:"omitempty,max=50"`
	Sort        int                  `json:"sort"`
	Status      *int8                `json:"status" validate:"omitempty,oneof=0 1"`
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c echo.Context) error {
	var req CreatePermissionRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	permission := &model.Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Type:        req.Type,
		Domain:      req.Domain,
		Resource:    req.Resource,
		Action:      req.Action,
		Category:    req.Category,
		ParentID:    req.ParentID,
		Path:        req.Path,
		Component:   req.Component,
		Icon:        req.Icon,
		Sort:        req.Sort,
		Status:      1,
	}

	if err := h.rbacService.CreatePermission(c.Request().Context(), permission); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "权限创建成功", permission)
}

// UpdatePermission 更新权限
func (h *PermissionHandler) UpdatePermission(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid permission id")
	}

	var req UpdatePermissionRequest
	if err := c.Bind(&req); err != nil {
		return errors.New(errors.ErrBindJSON, "")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// 获取现有权限
	permission, err := h.rbacService.GetPermission(c.Request().Context(), uint(id))
	if err != nil {
		return errors.New(errors.ErrNotFound, "权限不存在")
	}

	// 更新字段
	permission.DisplayName = req.DisplayName
	permission.Description = req.Description
	permission.Type = req.Type
	permission.Resource = req.Resource
	permission.Action = req.Action
	permission.Category = req.Category
	if req.ParentID != nil {
		permission.ParentID = *req.ParentID
	}
	permission.Path = req.Path
	permission.Component = req.Component
	permission.Icon = req.Icon
	permission.Sort = req.Sort
	if req.Status != nil {
		permission.Status = *req.Status
	}

	if err := h.rbacService.UpdatePermission(c.Request().Context(), permission); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "权限更新成功", permission)
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid permission id")
	}

	// 检查是否是系统权限
	permission, err := h.rbacService.GetPermission(c.Request().Context(), uint(id))
	if err != nil {
		return errors.New(errors.ErrNotFound, "权限不存在")
	}

	if permission.IsSystem {
		return errors.New(errors.ErrInvalidParams, "系统权限不能删除")
	}

	if err := h.rbacService.DeletePermission(c.Request().Context(), uint(id)); err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.SuccessWithMessage(c, "权限删除成功", nil)
}

// GetPermission 获取权限详情
func (h *PermissionHandler) GetPermission(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid permission id")
	}

	permission, err := h.rbacService.GetPermission(c.Request().Context(), uint(id))
	if err != nil {
		return errors.New(errors.ErrNotFound, "权限不存在")
	}

	return response.Success(c, permission)
}

// ListPermissions 获取权限列表
func (h *PermissionHandler) ListPermissions(c echo.Context) error {
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

	permissions, err := h.rbacService.ListPermissions(c.Request().Context(), domain, pagination)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"items":     permissions,
			"total":     pagination.Total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// ListPermissionsByType 根据类型获取权限列表
func (h *PermissionHandler) ListPermissionsByType(c echo.Context) error {
	permType := model.PermissionType(c.QueryParam("type"))
	domain := c.QueryParam("domain")

	if permType == "" {
		return errors.New(errors.ErrInvalidParams, "type is required")
	}

	permissions, err := h.rbacService.ListPermissionsByType(c.Request().Context(), permType, domain)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.Success(c, permissions)
}

// ListPermissionsTree 获取树形权限结构
func (h *PermissionHandler) ListPermissionsTree(c echo.Context) error {
	domain := c.QueryParam("domain")

	permissions, err := h.rbacService.ListPermissionsTree(c.Request().Context(), domain)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return response.Success(c, permissions)
}

// SearchPermissions 搜索权限
func (h *PermissionHandler) SearchPermissions(c echo.Context) error {
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

	permissions, err := h.rbacService.SearchPermissions(c.Request().Context(), keyword, domain, pagination)
	if err != nil {
		return errors.New(errors.ErrDatabase, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"items":     permissions,
			"total":     pagination.Total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
