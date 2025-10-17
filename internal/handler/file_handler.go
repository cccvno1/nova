package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/cccvno1/nova/internal/service"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/middleware"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

// FileHandler 文件上传处理器
type FileHandler struct {
	fileService service.FileService
}

// NewFileHandler 创建文件上传处理器
func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// Upload godoc
// @Summary 上传文件
// @Description 上传文件到服务器，支持图片自动生成缩略图
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "要上传的文件"
// @Param category formData string false "文件分类" Enums(avatar, document, image, video, audio, other) default(other)
// @Success 200 {object} response.Response{data=model.File} "上传成功，返回文件信息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 413 {object} response.Response "文件过大"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /files/upload [post]
func (h *FileHandler) Upload(c echo.Context) error {
	// 获取当前用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return errors.New(errors.ErrUnauthorized, "user not authenticated")
	}

	// 获取文件分类（可选）
	category := c.FormValue("category")
	if category == "" {
		category = "other"
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "file is required")
	}

	// 调用服务上传文件
	fileResp, err := h.fileService.Upload(c.Request().Context(), file, category, userID)
	if err != nil {
		return err
	}

	return response.Success(c, fileResp)
}

// Download 下载文件
// @Summary 下载文件
// @Description 根据文件ID下载文件，支持断点续传
// @Tags 文件管理
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Success 200 {file} binary "文件流"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "文件不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /files/{id}/download [get]
func (h *FileHandler) Download(c echo.Context) error {
	// 获取文件 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid file id")
	}

	// 获取当前用户 ID
	userID := middleware.GetUserID(c)

	// 下载文件
	reader, file, err := h.fileService.Download(c.Request().Context(), uint(id), userID)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 设置响应头
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.OriginalName))
	c.Response().Header().Set("Content-Type", file.MimeType)
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	// 返回文件内容
	return c.Stream(http.StatusOK, file.MimeType, reader)
}

// Delete 删除文件
// @Summary 删除文件
// @Description 根据ID删除指定文件（逻辑删除）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "无权限删除该文件"
// @Failure 404 {object} response.Response "文件不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /files/{id} [delete]
func (h *FileHandler) Delete(c echo.Context) error {
	// 获取文件 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid file id")
	}

	// 获取当前用户 ID
	userID := middleware.GetUserID(c)

	// 删除文件
	if err := h.fileService.Delete(c.Request().Context(), uint(id), userID); err != nil {
		return err
	}

	return response.SuccessWithMessage(c, "file deleted successfully", nil)
}

// GetByID 获取文件信息
// @Summary 获取文件详情
// @Description 根据ID获取文件的详细信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文件ID"
// @Success 200 {object} response.Response{data=model.File} "文件详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "文件不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /files/{id} [get]
func (h *FileHandler) GetByID(c echo.Context) error {
	// 获取文件 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "invalid file id")
	}

	// 获取文件信息
	file, err := h.fileService.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return err
	}

	return response.Success(c, file)
}

// List 获取文件列表
// @Summary 获取文件列表
// @Description 获取当前用户的文件列表，支持分类过滤和分页
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category query string false "文件分类" Enums(avatar, document, image, video, audio, other)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=[]model.File} "文件列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /files [get]
func (h *FileHandler) List(c echo.Context) error {
	// 获取当前用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return errors.New(errors.ErrUnauthorized, "user not authenticated")
	}

	// 获取分类过滤（可选）
	category := c.QueryParam("category")

	// 分页参数
	pagination := &database.Pagination{}
	if err := c.Bind(pagination); err != nil {
		return errors.New(errors.ErrBindQuery, "")
	}

	// 查询文件列表
	files, err := h.fileService.List(c.Request().Context(), userID, category, pagination)
	if err != nil {
		return err
	}

	return response.SuccessWithPagination(c, files, pagination)
}

// Search 搜索文件
// @Summary 搜索文件
// @Description 根据关键词搜索文件名，支持分页
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=[]model.File} "搜索结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /files/search [get]
func (h *FileHandler) Search(c echo.Context) error {
	// 获取搜索关键词
	keyword := c.QueryParam("keyword")

	// 分页参数
	pagination := &database.Pagination{}
	if err := c.Bind(pagination); err != nil {
		return errors.New(errors.ErrBindQuery, "")
	}

	// 搜索文件
	files, err := h.fileService.Search(c.Request().Context(), keyword, pagination)
	if err != nil {
		return err
	}

	return response.SuccessWithPagination(c, files, pagination)
}

// GetStorageInfo 获取用户存储信息
// @Summary 获取存储空间信息
// @Description 获取当前用户的存储空间使用情况统计
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object} "存储空间信息"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /files/storage/info [get]
func (h *FileHandler) GetStorageInfo(c echo.Context) error {
	// 获取当前用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return errors.New(errors.ErrUnauthorized, "user not authenticated")
	}

	// 获取存储信息
	info, err := h.fileService.GetUserStorageInfo(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return response.Success(c, info)
}

// UploadAvatar 上传头像（快捷方法）
// @Summary 上传用户头像
// @Description 上传头像文件（仅支持图片格式：jpg, jpeg, png, gif, webp）
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "头像文件"
// @Success 200 {object} response.Response{data=model.File} "上传成功"
// @Failure 400 {object} response.Response "请求参数错误或文件格式不支持"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /files/upload/avatar [post]
func (h *FileHandler) UploadAvatar(c echo.Context) error {
	// 获取当前用户 ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return errors.New(errors.ErrUnauthorized, "user not authenticated")
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		return errors.New(errors.ErrInvalidParams, "file is required")
	}

	// 验证是否为图片
	ext := filepath.Ext(file.Filename)
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isImage := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isImage = true
			break
		}
	}
	if !isImage {
		return errors.New(errors.ErrInvalidParams, "only image files are allowed for avatar")
	}

	// 调用服务上传文件
	fileResp, err := h.fileService.Upload(c.Request().Context(), file, "avatar", userID)
	if err != nil {
		return err
	}

	return response.Success(c, fileResp)
}
