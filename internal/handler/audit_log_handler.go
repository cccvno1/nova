package handler

import (
	"strconv"
	"time"

	"github.com/cccvno1/nova/internal/repository"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

// AuditLogHandler 审计日志处理器
type AuditLogHandler struct {
	auditRepo repository.AuditLogRepository
}

// NewAuditLogHandler 创建审计日志处理器
func NewAuditLogHandler(auditRepo repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{
		auditRepo: auditRepo,
	}
}

// GetByID 根据 ID 获取审计日志
// @Summary 获取审计日志详情
// @Description 根据ID获取审计日志的详细信息
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "审计日志ID"
// @Success 200 {object} response.Response{data=model.AuditLog} "审计日志详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "审计日志不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit-logs/{id} [get]
func (h *AuditLogHandler) GetByID(c echo.Context) error {
	// 获取日志 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid audit log id")
	}

	// 查询审计日志
	log, err := h.auditRepo.FindByID(c.Request().Context(), uint(id))
	if err != nil {
		return errors.Wrap(errors.ErrRecordNotFound, err)
	}

	return response.Success(c, log)
}

// List 查询审计日志列表
// @Summary 获取审计日志列表
// @Description 查询审计日志列表，支持多条件过滤和分页
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "用户ID"
// @Param action query string false "操作类型"
// @Param resource query string false "资源类型"
// @Param ip query string false "IP地址"
// @Param method query string false "HTTP方法"
// @Param path query string false "请求路径"
// @Param status_code query int false "HTTP状态码"
// @Param start_time query string false "开始时间(RFC3339格式)"
// @Param end_time query string false "结束时间(RFC3339格式)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=[]model.AuditLog} "审计日志列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit-logs [get]
func (h *AuditLogHandler) List(c echo.Context) error {
	// 分页参数
	pagination := &database.Pagination{}
	if err := c.Bind(pagination); err != nil {
		return errors.New(errors.ErrBindQuery, "")
	}

	// 获取过滤条件
	filters := make(map[string]interface{})

	if userIDStr := c.QueryParam("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}

	if action := c.QueryParam("action"); action != "" {
		filters["action"] = action
	}

	if resource := c.QueryParam("resource"); resource != "" {
		filters["resource"] = resource
	}

	if ip := c.QueryParam("ip"); ip != "" {
		filters["ip"] = ip
	}

	if method := c.QueryParam("method"); method != "" {
		filters["method"] = method
	}

	if path := c.QueryParam("path"); path != "" {
		filters["path"] = path
	}

	if statusCodeStr := c.QueryParam("status_code"); statusCodeStr != "" {
		if statusCode, err := strconv.Atoi(statusCodeStr); err == nil {
			filters["status_code"] = statusCode
		}
	}

	// 时间范围过滤
	if startTimeStr := c.QueryParam("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filters["start_time"] = startTime
		}
	}

	if endTimeStr := c.QueryParam("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filters["end_time"] = endTime
		}
	}

	// 查询审计日志
	var logs interface{}
	var queryErr error

	if len(filters) > 0 {
		logs, queryErr = h.auditRepo.Search(c.Request().Context(), filters, pagination)
	} else {
		logs, queryErr = h.auditRepo.List(c.Request().Context(), pagination)
	}

	if queryErr != nil {
		return errors.Wrap(errors.ErrDatabase, queryErr)
	}

	return response.SuccessWithPagination(c, logs, pagination)
}

// ListByUser 查询指定用户的审计日志
// @Summary 获取用户审计日志
// @Description 查询指定用户的审计日志列表
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=[]model.AuditLog} "审计日志列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit-logs/user/{userId} [get]
func (h *AuditLogHandler) ListByUser(c echo.Context) error {
	// 获取用户 ID
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid user id")
	}

	// 分页参数
	pagination := &database.Pagination{}
	if err := c.Bind(pagination); err != nil {
		return errors.New(errors.ErrBindQuery, "")
	}

	// 查询日志
	logs, err := h.auditRepo.ListByUser(c.Request().Context(), uint(userID), pagination)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, err)
	}

	return response.SuccessWithPagination(c, logs, pagination)
}

// GetStats 获取审计日志统计信息
// @Summary 获取综合统计
// @Description 获取审计日志的综合统计信息，包括总数、成功/失败数、操作统计、用户统计和资源统计
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string false "开始时间(RFC3339格式)" default(7天前)
// @Param end_time query string false "结束时间(RFC3339格式)" default(当前时间)
// @Success 200 {object} response.Response{data=object} "统计信息"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit-logs/stats [get]
func (h *AuditLogHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()

	// 获取时间范围（默认最近7天）
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)

	if startTimeStr := c.QueryParam("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr := c.QueryParam("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	// 获取总数
	totalCount, _ := h.auditRepo.CountByTimeRange(ctx, startTime, endTime)

	// 获取成功/失败数量
	successCount, _ := h.auditRepo.CountByStatus(ctx, 200)
	errorCount, _ := h.auditRepo.CountByStatus(ctx, 500)

	// 获取操作统计
	actionStats, _ := h.auditRepo.GetActionStats(ctx, startTime, endTime)

	// 获取用户统计（Top 10）
	userStats, _ := h.auditRepo.GetUserStats(ctx, startTime, endTime, 10)

	// 获取资源统计
	resourceStats, _ := h.auditRepo.GetResourceStats(ctx, startTime, endTime)

	stats := map[string]interface{}{
		"time_range": map[string]interface{}{
			"start": startTime.Format(time.RFC3339),
			"end":   endTime.Format(time.RFC3339),
		},
		"total_count":    totalCount,
		"success_count":  successCount,
		"error_count":    errorCount,
		"action_stats":   actionStats,
		"user_stats":     userStats,
		"resource_stats": resourceStats,
	}

	return response.Success(c, stats)
}

// GetActionStats 获取操作统计
// @Summary 获取操作统计
// @Description 获取指定时间范围内的操作类型统计
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string false "开始时间(RFC3339格式)" default(24小时前)
// @Param end_time query string false "结束时间(RFC3339格式)" default(当前时间)
// @Success 200 {object} response.Response{data=[]object} "操作统计列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit-logs/stats/actions [get]
func (h *AuditLogHandler) GetActionStats(c echo.Context) error {
	ctx := c.Request().Context()

	// 获取时间范围（默认最近24小时）
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if startTimeStr := c.QueryParam("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr := c.QueryParam("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	// 获取操作统计
	stats, err := h.auditRepo.GetActionStats(ctx, startTime, endTime)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, err)
	}

	return response.Success(c, stats)
}

// GetUserStats 获取用户操作统计
// @Summary 获取用户统计
// @Description 获取指定时间范围内操作最频繁的用户统计（Top N）
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string false "开始时间(RFC3339格式)" default(7天前)
// @Param end_time query string false "结束时间(RFC3339格式)" default(当前时间)
// @Param limit query int false "返回数量" default(10)
// @Success 200 {object} response.Response{data=[]object} "用户统计列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit-logs/stats/users [get]
func (h *AuditLogHandler) GetUserStats(c echo.Context) error {
	ctx := c.Request().Context()

	// 获取时间范围（默认最近7天）
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)

	if startTimeStr := c.QueryParam("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr := c.QueryParam("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	// 获取限制数量（默认 Top 10）
	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 获取用户统计
	stats, err := h.auditRepo.GetUserStats(ctx, startTime, endTime, limit)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, err)
	}

	return response.Success(c, stats)
}

// GetResourceStats 获取资源操作统计
// @Summary 获取资源统计
// @Description 获取指定时间范围内的资源操作统计
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string false "开始时间(RFC3339格式)" default(7天前)
// @Param end_time query string false "结束时间(RFC3339格式)" default(当前时间)
// @Success 200 {object} response.Response{data=[]object} "资源统计列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit-logs/stats/resources [get]
func (h *AuditLogHandler) GetResourceStats(c echo.Context) error {
	ctx := c.Request().Context()

	// 获取时间范围（默认最近7天）
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)

	if startTimeStr := c.QueryParam("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr := c.QueryParam("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	// 获取资源统计
	stats, err := h.auditRepo.GetResourceStats(ctx, startTime, endTime)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, err)
	}

	return response.Success(c, stats)
}

// CleanOldLogs 清理旧的审计日志
// @Summary 清理旧审计日志
// @Description 清理指定天数之前的审计日志（物理删除）
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "保留天数" default(90)
// @Success 200 {object} response.Response{data=object} "删除结果(包含删除数量和截止时间)"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /audit-logs/clean [delete]
func (h *AuditLogHandler) CleanOldLogs(c echo.Context) error {
	// 获取清理天数（默认保留90天）
	days := 90
	if daysStr := c.QueryParam("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// 计算截止时间
	beforeTime := time.Now().AddDate(0, 0, -days)

	// 删除旧日志
	deletedCount, err := h.auditRepo.DeleteBefore(c.Request().Context(), beforeTime)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, err)
	}

	return response.Success(c, map[string]interface{}{
		"deleted_count": deletedCount,
		"before_time":   beforeTime.Format(time.RFC3339),
	})
}
