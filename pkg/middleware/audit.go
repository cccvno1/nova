package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/internal/repository"
	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/labstack/echo/v4"
)

// AuditLogMiddleware 审计日志中间件配置
type AuditLogMiddleware struct {
	config   *config.AuditLogConfig
	repo     repository.AuditLogRepository
	disabled bool
}

// NewAuditLogMiddleware 创建审计日志中间件
func NewAuditLogMiddleware(cfg *config.AuditLogConfig, db *database.Database) *AuditLogMiddleware {
	if !cfg.Enabled {
		return &AuditLogMiddleware{disabled: true}
	}

	return &AuditLogMiddleware{
		config:   cfg,
		repo:     repository.NewAuditLogRepository(db),
		disabled: false,
	}
}

// Handler 审计日志中间件处理函数
func (m *AuditLogMiddleware) Handler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 如果中间件被禁用，直接跳过
			if m.disabled {
				return next(c)
			}

			// 检查是否在排除路径中
			if m.isExcluded(c.Request().URL.Path) {
				return next(c)
			}

			// 记录开始时间
			startTime := time.Now()

			// 捕获请求体（如果需要）
			var requestBody string
			if m.config.LogRequest && c.Request().Body != nil {
				bodyBytes, _ := io.ReadAll(c.Request().Body)
				// 恢复请求体供后续使用
				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// 限制记录大小
				if len(bodyBytes) > m.config.MaxBodySize {
					requestBody = string(bodyBytes[:m.config.MaxBodySize]) + "...(truncated)"
				} else {
					requestBody = string(bodyBytes)
				}

				// 脱敏处理
				requestBody = m.maskSensitiveData(requestBody)
			}

			// 创建自定义响应写入器以捕获响应
			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
			c.Response().Writer = writer

			// 执行实际的处理函数
			err := next(c)

			// 计算耗时
			duration := time.Since(startTime)

			// 获取响应体（如果需要）
			var responseBody string
			if m.config.LogResponse {
				respBytes := resBody.Bytes()
				if len(respBytes) > m.config.MaxBodySize {
					responseBody = string(respBytes[:m.config.MaxBodySize]) + "...(truncated)"
				} else {
					responseBody = string(respBytes)
				}
				// 脱敏处理
				responseBody = m.maskSensitiveData(responseBody)
			}

			// 提取操作信息
			action, resource, resourceID := m.extractActionInfo(c)

			// 获取用户信息
			userID := GetUserID(c)
			username := GetUsername(c)

			// 获取错误信息
			var errorMsg string
			if err != nil {
				errorMsg = err.Error()
			}

			// 创建审计日志记录
			auditLog := &model.AuditLog{
				UserID:     userID,
				Username:   username,
				Action:     action,
				Resource:   resource,
				ResourceID: resourceID,
				Method:     c.Request().Method,
				Path:       c.Request().URL.Path,
				IP:         c.RealIP(),
				UserAgent:  c.Request().UserAgent(),
				Request:    requestBody,
				Response:   responseBody,
				StatusCode: c.Response().Status,
				Duration:   int64(duration),
				Error:      errorMsg,
			}

			// 异步保存审计日志（不阻塞主流程）
			go func() {
				if saveErr := m.repo.Create(c.Request().Context(), auditLog); saveErr != nil {
					// 记录日志保存失败，但不影响主流程
					// 可以在这里添加日志记录
				}
			}()

			return err
		}
	}
}

// isExcluded 检查路径是否在排除列表中
func (m *AuditLogMiddleware) isExcluded(path string) bool {
	for _, excludePath := range m.config.ExcludePaths {
		if strings.HasPrefix(path, excludePath) {
			return true
		}
	}
	return false
}

// extractActionInfo 从路径和方法提取操作信息
func (m *AuditLogMiddleware) extractActionInfo(c echo.Context) (action, resource, resourceID string) {
	method := c.Request().Method
	path := c.Request().URL.Path

	// 根据 HTTP 方法映射操作
	switch method {
	case "POST":
		if strings.Contains(path, "/login") {
			action = model.AuditActionLogin
		} else if strings.Contains(path, "/logout") {
			action = model.AuditActionLogout
		} else {
			action = model.AuditActionCreate
		}
	case "GET":
		action = model.AuditActionRead
	case "PUT", "PATCH":
		action = model.AuditActionUpdate
	case "DELETE":
		action = model.AuditActionDelete
	default:
		action = strings.ToLower(method)
	}

	// 从路径提取资源类型
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) >= 3 {
		resource = pathParts[2] // 通常是 /api/v1/{resource}

		// 提取资源ID（如果存在）
		if len(pathParts) >= 4 && pathParts[3] != "" {
			resourceID = pathParts[3]
		}
	}

	// 标准化资源名称（去掉复数）
	if strings.HasSuffix(resource, "s") && len(resource) > 1 {
		resource = resource[:len(resource)-1]
	}

	return action, resource, resourceID
}

// maskSensitiveData 脱敏敏感数据
func (m *AuditLogMiddleware) maskSensitiveData(data string) string {
	if data == "" {
		return data
	}

	// 尝试解析为 JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		// 不是 JSON，返回原始数据
		return data
	}

	// 对敏感字段进行脱敏
	for _, field := range m.config.SensitiveFields {
		if _, exists := jsonData[field]; exists {
			jsonData[field] = "***MASKED***"
		}
	}

	// 序列化回 JSON
	maskedData, err := json.Marshal(jsonData)
	if err != nil {
		return data
	}

	return string(maskedData)
}

// bodyDumpResponseWriter 自定义响应写入器
type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, echo.ErrNotFound
}
