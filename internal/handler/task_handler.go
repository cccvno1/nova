package handler

import (
	"strconv"

	"github.com/cccvno1/nova/internal/repository"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskRepo repository.TaskRepository
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskRepo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{
		taskRepo: taskRepo,
	}
}

// GetByID 根据 ID 获取任务
// @Summary 获取任务详情
// @Description 根据数据库ID获取任务的详细信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=model.Task} "任务详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "任务不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetByID(c echo.Context) error {
	// 获取任务 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid task id")
	}

	// 查询任务
	task, err := h.taskRepo.FindByID(c.Request().Context(), uint(id))
	if err != nil {
		return errors.Wrap(errors.ErrRecordNotFound, err)
	}

	return response.Success(c, task)
}

// GetByTaskID 根据任务ID获取任务
// @Summary 根据任务ID查询
// @Description 根据业务任务ID（task_id字段）查询任务信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param taskId path string true "业务任务ID"
// @Success 200 {object} response.Response{data=model.Task} "任务详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "任务不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /tasks/task/{taskId} [get]
func (h *TaskHandler) GetByTaskID(c echo.Context) error {
	taskID := c.Param("taskId")
	if taskID == "" {
		return errors.New(errors.ErrInvalidParams, "task_id is required")
	}

	task, err := h.taskRepo.FindByTaskID(c.Request().Context(), taskID)
	if err != nil {
		return errors.Wrap(errors.ErrRecordNotFound, err)
	}

	return response.Success(c, task)
}

// List 查询任务列表
// @Summary 获取任务列表
// @Description 查询任务列表，支持按状态和类型过滤，默认查询待处理任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "任务状态" Enums(pending, processing, success, failed)
// @Param type query string false "任务类型"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=[]model.Task} "任务列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /tasks [get]
func (h *TaskHandler) List(c echo.Context) error {
	// 分页参数
	pagination := &database.Pagination{}
	if err := c.Bind(pagination); err != nil {
		return errors.New(errors.ErrBindQuery, "")
	}

	// 获取可选的过滤条件
	status := c.QueryParam("status")
	taskType := c.QueryParam("type")

	var tasks interface{}
	var err error

	if status != "" {
		// 按状态查询
		tasks, err = h.taskRepo.ListByStatus(c.Request().Context(), status, pagination)
	} else if taskType != "" {
		// 按类型查询
		tasks, err = h.taskRepo.ListByType(c.Request().Context(), taskType, pagination)
	} else {
		// 默认按状态pending查询
		tasks, err = h.taskRepo.ListByStatus(c.Request().Context(), "pending", pagination)
	}

	if err != nil {
		return errors.Wrap(errors.ErrDatabase, err)
	}

	return response.SuccessWithPagination(c, tasks, pagination)
}

// GetStats 获取任务统计信息
// @Summary 获取任务统计
// @Description 获取各状态任务的数量统计
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object} "统计信息(包含pending, processing, success, failed, total)"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /tasks/stats [get]
func (h *TaskHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()

	// 统计各状态的任务数量
	pendingCount, _ := h.taskRepo.CountByStatus(ctx, "pending")
	processingCount, _ := h.taskRepo.CountByStatus(ctx, "processing")
	successCount, _ := h.taskRepo.CountByStatus(ctx, "success")
	failedCount, _ := h.taskRepo.CountByStatus(ctx, "failed")

	stats := map[string]interface{}{
		"pending":    pendingCount,
		"processing": processingCount,
		"success":    successCount,
		"failed":     failedCount,
		"total":      pendingCount + processingCount + successCount + failedCount,
	}

	return response.Success(c, stats)
}
