package model

import (
	"github.com/cccvno1/nova/pkg/database"
)

// PermissionType 权限类型
type PermissionType string

const (
	PermissionTypeAPI    PermissionType = "api"    // API 权限
	PermissionTypeMenu   PermissionType = "menu"   // 菜单权限
	PermissionTypeButton PermissionType = "button" // 按钮权限
	PermissionTypeData   PermissionType = "data"   // 数据权限
	PermissionTypeField  PermissionType = "field"  // 字段权限
)

// Permission 权限模型（用于 UI 管理和元数据存储）
// 实际权限验证由 Casbin 处理，这个模型主要用于：
// 1. 提供友好的 UI 展示（中文名称、分组、图标）
// 2. 管理权限的元数据
// 3. 权限选择器的数据源
type Permission struct {
	database.Model
	Name        string         `json:"name" gorm:"size:100;not null;uniqueIndex:idx_permission_domain"`         // 权限标识，如 "user:read"
	DisplayName string         `json:"display_name" gorm:"size:100;not null"`                                   // 显示名称，如 "查看用户"
	Description string         `json:"description" gorm:"size:500"`                                             // 权限描述
	Type        PermissionType `json:"type" gorm:"size:20;not null;index"`                                      // 权限类型
	Domain      string         `json:"domain" gorm:"size:100;not null;uniqueIndex:idx_permission_domain;index"` // 所属域/租户

	// 资源信息（对应 Casbin 的 obj 和 act）
	Resource string `json:"resource" gorm:"size:200;not null"` // 资源路径，如 "/api/users", "/menu/system"
	Action   string `json:"action" gorm:"size:50;not null"`    // 操作，如 "read", "write", "delete"

	// 分组和组织
	Category string `json:"category" gorm:"size:50;index"`    // 权限分类，如 "用户管理", "订单管理"
	ParentID uint   `json:"parent_id" gorm:"default:0;index"` // 父权限ID（用于树形结构）

	// 前端相关
	Path      string `json:"path" gorm:"size:200"`      // 前端路由路径（菜单权限用）
	Component string `json:"component" gorm:"size:200"` // 前端组件路径（菜单权限用）
	Icon      string `json:"icon" gorm:"size:50"`       // 图标

	// 其他属性
	IsSystem bool `json:"is_system" gorm:"default:false;index"` // 是否系统权限（不可删除）
	Sort     int  `json:"sort" gorm:"default:0"`                // 排序权重
	Status   int8 `json:"status" gorm:"default:1;index"`        // 状态：1=启用，0=禁用

	// 关联关系
	Children []Permission `json:"children,omitempty" gorm:"foreignKey:ParentID"` // 子权限
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// PermissionGroup 权限分组（可选，用于UI组织）
type PermissionGroup struct {
	database.Model
	Name        string `json:"name" gorm:"size:100;not null;uniqueIndex:idx_group_domain"` // 分组名称
	DisplayName string `json:"display_name" gorm:"size:100;not null"`                      // 显示名称
	Description string `json:"description" gorm:"size:500"`
	Domain      string `json:"domain" gorm:"size:100;not null;uniqueIndex:idx_group_domain;index"` // 所属域
	Icon        string `json:"icon" gorm:"size:50"`
	Sort        int    `json:"sort" gorm:"default:0"`
	Status      int8   `json:"status" gorm:"default:1;index"`
}

// TableName 指定表名
func (PermissionGroup) TableName() string {
	return "permission_groups"
}

// DataScope 数据权限范围（可选，用于数据权限控制）
type DataScope struct {
	database.Model
	Name        string `json:"name" gorm:"size:100;not null;uniqueIndex:idx_scope_domain"` // 范围名称
	DisplayName string `json:"display_name" gorm:"size:100;not null"`                      // 显示名称
	Description string `json:"description" gorm:"size:500"`
	Domain      string `json:"domain" gorm:"size:100;not null;uniqueIndex:idx_scope_domain;index"` // 所属域
	ScopeType   string `json:"scope_type" gorm:"size:50;not null"`                                 // 范围类型：all, custom, dept, self
	ScopeRule   string `json:"scope_rule" gorm:"type:text"`                                        // 范围规则（JSON）
	Status      int8   `json:"status" gorm:"default:1;index"`
}

// TableName 指定表名
func (DataScope) TableName() string {
	return "data_scopes"
}

// RoleDataScope 角色-数据范围关联
type RoleDataScope struct {
	database.Model
	RoleID      uint   `json:"role_id" gorm:"not null;index"`
	DataScopeID uint   `json:"data_scope_id" gorm:"not null;index"`
	Resource    string `json:"resource" gorm:"size:200;not null"` // 应用于哪个资源
}

// TableName 指定表名
func (RoleDataScope) TableName() string {
	return "role_data_scopes"
}
