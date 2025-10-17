package model

import (
	"github.com/cccvno1/nova/pkg/database"
)

// Role 角色模型（用于 UI 管理和元数据存储）
// 实际权限验证由 Casbin 处理，这个模型主要用于：
// 1. 提供友好的 UI 展示（中文名称、描述）
// 2. 管理角色的元数据
// 3. 角色分类和组织
type Role struct {
	database.Model
	Name        string `json:"name" gorm:"size:100;not null;uniqueIndex:idx_role_domain"`         // 角色标识，如 "admin", "editor"
	DisplayName string `json:"display_name" gorm:"size:100;not null"`                             // 显示名称，如 "系统管理员"
	Description string `json:"description" gorm:"size:500"`                                       // 角色描述
	Domain      string `json:"domain" gorm:"size:100;not null;uniqueIndex:idx_role_domain;index"` // 所属域/租户
	Category    string `json:"category" gorm:"size:50"`                                           // 角色分类，如 "system", "business"
	IsSystem    bool   `json:"is_system" gorm:"default:false;index"`                              // 是否系统角色（不可删除）
	Level       int    `json:"level" gorm:"default:10;index"`                                     // 角色等级(1-100)，数字越大权限越高，用于防止权限越级
	Sort        int    `json:"sort" gorm:"default:0"`                                             // 排序权重
	Status      int8   `json:"status" gorm:"default:1;index"`                                     // 状态：1=启用，0=禁用

	// 关联关系（可选，用于前端展示）
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// RolePermission 角色-权限关联表（可选，用于快速查询UI展示）
// 注意：实际权限验证仍使用 Casbin 的策略表
type RolePermission struct {
	database.Model
	RoleID       uint `json:"role_id" gorm:"not null;index"`
	PermissionID uint `json:"permission_id" gorm:"not null;index"`
}

// TableName 指定表名
func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserRole 用户-角色关联（辅助模型，可选）
// 实际的用户角色关系存储在 Casbin 的 g 表中
// 这个表主要用于：
// 1. 添加额外的元数据（如分配时间、分配人）
// 2. 快速查询用户角色列表
type UserRole struct {
	database.Model
	UserID     uint   `json:"user_id" gorm:"not null;index"`
	RoleID     uint   `json:"role_id" gorm:"not null;index"`
	Domain     string `json:"domain" gorm:"size:100;not null;index"` // 域/租户
	AssignedBy uint   `json:"assigned_by" gorm:"default:0"`          // 分配人ID

	// 关联关系
	Role *Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_roles"
}
