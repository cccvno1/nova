package model

import (
	"time"

	"github.com/cccvno1/nova/pkg/database"
)

// AuditLog 审计日志模型
type AuditLog struct {
	database.Model
	UserID     uint   `gorm:"index" json:"user_id"`                    // 用户ID
	Username   string `gorm:"size:100;index" json:"username"`          // 用户名
	Action     string `gorm:"not null;size:100;index" json:"action"`   // 操作动作（如：create, update, delete, login）
	Resource   string `gorm:"not null;size:100;index" json:"resource"` // 操作资源（如：user, file, role）
	ResourceID string `gorm:"size:100;index" json:"resource_id"`       // 资源ID
	Method     string `gorm:"not null;size:10" json:"method"`          // HTTP 方法
	Path       string `gorm:"not null;size:500;index" json:"path"`     // 请求路径
	IP         string `gorm:"not null;size:50;index" json:"ip"`        // 客户端 IP
	UserAgent  string `gorm:"size:500" json:"user_agent"`              // User Agent
	Request    string `gorm:"type:text" json:"request,omitempty"`      // 请求体（可选）
	Response   string `gorm:"type:text" json:"response,omitempty"`     // 响应体（可选）
	StatusCode int    `gorm:"not null;index" json:"status_code"`       // HTTP 状态码
	Duration   int64  `gorm:"not null" json:"duration"`                // 请求耗时（毫秒）
	Error      string `gorm:"type:text" json:"error,omitempty"`        // 错误信息
	Extra      string `gorm:"type:jsonb" json:"extra,omitempty"`       // 额外信息（JSON）
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditAction 审计动作常量
const (
	AuditActionCreate = "create" // 创建
	AuditActionRead   = "read"   // 读取
	AuditActionUpdate = "update" // 更新
	AuditActionDelete = "delete" // 删除
	AuditActionLogin  = "login"  // 登录
	AuditActionLogout = "logout" // 登出
	AuditActionExport = "export" // 导出
	AuditActionImport = "import" // 导入
)

// AuditResource 审计资源常量
const (
	AuditResourceUser       = "user"       // 用户
	AuditResourceRole       = "role"       // 角色
	AuditResourcePermission = "permission" // 权限
	AuditResourceFile       = "file"       // 文件
	AuditResourceTask       = "task"       // 任务
	AuditResourceAuditLog   = "audit_log"  // 审计日志
)

// IsSuccess 判断操作是否成功
func (a *AuditLog) IsSuccess() bool {
	return a.StatusCode >= 200 && a.StatusCode < 300
}

// GetDurationMs 获取耗时（毫秒）
func (a *AuditLog) GetDurationMs() float64 {
	return float64(a.Duration) / float64(time.Millisecond)
}
