package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/internal/repository"
	"github.com/cccvno1/nova/pkg/cache"
	"github.com/cccvno1/nova/pkg/casbin"
	"github.com/cccvno1/nova/pkg/database"
	"gorm.io/gorm"
)

// RBACService RBAC服务接口
// 提供角色和权限管理的完整功能，包括：
// - 角色管理（CRUD）
// - 权限管理（CRUD）
// - 角色-权限关联管理
// - 用户-角色关联管理
// - 权限验证（基于Casbin）
// - 策略管理（高级功能）
type RBACService interface {
	// 角色管理
	CreateRole(ctx context.Context, role *model.Role) error
	UpdateRole(ctx context.Context, role *model.Role) error
	DeleteRole(ctx context.Context, id uint) error
	GetRole(ctx context.Context, id uint) (*model.Role, error)
	ListRoles(ctx context.Context, domain string, pagination *database.Pagination) ([]model.Role, error)
	ListRolesFiltered(ctx context.Context, operatorID uint, domain string, pagination *database.Pagination) ([]model.Role, error) // 新增：带等级过滤
	SearchRoles(ctx context.Context, keyword, domain string, pagination *database.Pagination) ([]model.Role, error)
	SearchRolesFiltered(ctx context.Context, operatorID uint, keyword, domain string, pagination *database.Pagination) ([]model.Role, error) // 新增：带等级过滤

	// 权限管理
	CreatePermission(ctx context.Context, permission *model.Permission) error
	UpdatePermission(ctx context.Context, permission *model.Permission) error
	DeletePermission(ctx context.Context, id uint) error
	GetPermission(ctx context.Context, id uint) (*model.Permission, error)
	ListPermissions(ctx context.Context, domain string, pagination *database.Pagination) ([]model.Permission, error)
	ListPermissionsByType(ctx context.Context, permType model.PermissionType, domain string) ([]model.Permission, error)
	ListPermissionsTree(ctx context.Context, domain string) ([]model.Permission, error)
	SearchPermissions(ctx context.Context, keyword, domain string, pagination *database.Pagination) ([]model.Permission, error)

	// 角色-权限管理
	UpdateRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint, domain string, preview bool) (interface{}, error)
	AssignPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint, domain string) error
	RevokePermissionsFromRole(ctx context.Context, roleID uint, permissionIDs []uint, domain string) error
	GetRolePermissions(ctx context.Context, roleID uint, domain string) ([]model.Permission, error)

	// 用户-角色管理
	AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint, domain string, assignedBy uint) error
	RevokeRolesFromUser(ctx context.Context, userID uint, roleIDs []uint, domain string) error
	GetUserRoles(ctx context.Context, userID uint, domain string) ([]model.Role, error)
	GetRoleUsers(ctx context.Context, roleID uint) ([]model.UserRole, error)

	// 权限验证
	CheckPermission(ctx context.Context, userID uint, domain, resource, action string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uint, domain string) ([]model.Permission, error)

	// 策略管理（高级用户使用）
	AddPolicy(ctx context.Context, sub, dom, obj, act string) error
	RemovePolicy(ctx context.Context, sub, dom, obj, act string) error
	ListPolicies(ctx context.Context, domain string) ([][]string, error)

	// 安全检查（权限越级保护）
	GetUserMaxRoleLevel(ctx context.Context, userID uint, domain string) (int, error)
	CheckRoleLevelPermission(ctx context.Context, operatorID uint, targetRoleID uint, domain string) error
	CheckRolesLevelPermission(ctx context.Context, operatorID uint, targetRoleIDs []uint, domain string) error
}

// rbacService RBAC服务实现
type rbacService struct {
	enforcer     *casbin.Enforcer                // Casbin权限执行器（保留但不使用，方案A已改为直接查询RBAC表）
	roleRepo     repository.RoleRepository       // 角色数据仓储
	permRepo     repository.PermissionRepository // 权限数据仓储
	userRoleRepo repository.UserRoleRepository   // 用户角色关联仓储
	db           *database.Database              // 数据库实例（用于直接操作关联表）
	cache        *cache.CacheManager             // Redis缓存管理器
	logger       *slog.Logger                    // 日志记录器
}

const (
	// 缓存key前缀
	cacheKeyUserPermissions = "rbac:user:permissions:%d:%s" // user_id:domain
	cacheKeyRolePermissions = "rbac:role:permissions:%d:%s" // role_id:domain
	// 缓存TTL
	cacheTTLPermissions = 10 * time.Minute
)

// NewRBACService 创建RBAC服务实例
func NewRBACService(
	enforcer *casbin.Enforcer,
	roleRepo repository.RoleRepository,
	permRepo repository.PermissionRepository,
	userRoleRepo repository.UserRoleRepository,
	db *database.Database,
	logger *slog.Logger,
) RBACService {
	return &rbacService{
		enforcer:     enforcer,
		roleRepo:     roleRepo,
		permRepo:     permRepo,
		userRoleRepo: userRoleRepo,
		db:           db,
		cache:        cache.NewCacheManager(),
		logger:       logger,
	}
}

// ============================
// 角色管理
// ============================

// CreateRole 创建角色
func (s *rbacService) CreateRole(ctx context.Context, role *model.Role) error {
	// 检查角色名称是否已存在
	exists, err := s.roleRepo.ExistsByName(ctx, role.Name, role.Domain, 0)
	if err != nil {
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	if exists {
		return fmt.Errorf("role name %s already exists in domain %s", role.Name, role.Domain)
	}

	// 创建角色
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	s.logger.Info("role created",
		"role_id", role.ID,
		"role_name", role.Name,
		"domain", role.Domain,
	)

	return nil
}

// UpdateRole 更新角色
func (s *rbacService) UpdateRole(ctx context.Context, role *model.Role) error {
	// 检查角色是否存在
	_, err := s.roleRepo.FindByID(ctx, role.ID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 检查角色名称是否重复
	exists, err := s.roleRepo.ExistsByName(ctx, role.Name, role.Domain, role.ID)
	if err != nil {
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	if exists {
		return fmt.Errorf("role name %s already exists in domain %s", role.Name, role.Domain)
	}

	// 更新角色
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	s.logger.Info("role updated",
		"role_id", role.ID,
		"role_name", role.Name,
		"domain", role.Domain,
	)

	return nil
}

// DeleteRole 删除角色
func (s *rbacService) DeleteRole(ctx context.Context, id uint) error {
	// 检查角色是否存在
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 检查是否为系统角色
	if role.IsSystem {
		return fmt.Errorf("cannot delete system role")
	}

	// 删除角色的所有权限策略
	roleName := strconv.FormatUint(uint64(id), 10)
	if _, err := s.enforcer.RemoveAllPoliciesForRole(roleName, role.Domain); err != nil {
		return fmt.Errorf("failed to remove role policies: %w", err)
	}

	// 删除所有用户的这个角色分配
	users, err := s.enforcer.GetUsersForRole(roleName, role.Domain)
	if err != nil {
		return fmt.Errorf("failed to get role users: %w", err)
	}
	for _, user := range users {
		if _, err := s.enforcer.DeleteRoleForUser(user, roleName, role.Domain); err != nil {
			s.logger.Error("failed to remove user role",
				"user", user,
				"role", roleName,
				"error", err,
			)
		}
	}

	// 删除角色记录
	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	s.logger.Info("role deleted",
		"role_id", id,
		"role_name", role.Name,
		"domain", role.Domain,
	)

	return nil
}

// GetRole 获取角色详情
func (s *rbacService) GetRole(ctx context.Context, id uint) (*model.Role, error) {
	return s.roleRepo.FindByID(ctx, id)
}

// ListRoles 查询角色列表
// ListRoles 列出所有角色（无权限过滤，仅供内部使用）
func (s *rbacService) ListRoles(ctx context.Context, domain string, pagination *database.Pagination) ([]model.Role, error) {
	return s.roleRepo.List(ctx, domain, pagination)
}

// ListRolesFiltered 获取角色列表（带权限过滤：只返回等级低于操作者的角色）
// 🔒 安全加固：实现行级安全（Row Level Security）
func (s *rbacService) ListRolesFiltered(ctx context.Context, operatorID uint, domain string, pagination *database.Pagination) ([]model.Role, error) {
	// 1. 获取操作者的最高等级
	operatorLevel, err := s.GetUserMaxRoleLevel(ctx, operatorID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator level: %w", err)
	}

	s.logger.Info("ListRolesFiltered: 开始过滤",
		"operatorID", operatorID,
		"operatorLevel", operatorLevel,
		"domain", domain,
	)

	// 2. 查询所有角色
	roles, err := s.roleRepo.List(ctx, domain, pagination)
	if err != nil {
		return nil, err
	}

	s.logger.Info("ListRolesFiltered: 查询到角色",
		"totalRoles", len(roles),
	)

	// 3. 过滤：只返回等级严格低于操作者的角色
	filteredRoles := make([]model.Role, 0)
	for _, role := range roles {
		s.logger.Info("ListRolesFiltered: 检查角色",
			"roleName", role.Name,
			"roleLevel", role.Level,
			"operatorLevel", operatorLevel,
			"willInclude", role.Level < operatorLevel,
		)
		if role.Level < operatorLevel {
			filteredRoles = append(filteredRoles, role)
		}
	}

	s.logger.Info("ListRolesFiltered: 过滤完成",
		"filteredCount", len(filteredRoles),
	)

	// 4. 更新分页信息（过滤后的总数）
	pagination.Total = int64(len(filteredRoles))

	return filteredRoles, nil
}

// SearchRoles 搜索角色（无权限过滤，仅供内部使用）
func (s *rbacService) SearchRoles(ctx context.Context, keyword, domain string, pagination *database.Pagination) ([]model.Role, error) {
	return s.roleRepo.Search(ctx, keyword, domain, pagination)
}

// SearchRolesFiltered 搜索角色（带权限过滤：只返回等级低于操作者的角色）
// 🔒 安全加固：实现行级安全（Row Level Security）
func (s *rbacService) SearchRolesFiltered(ctx context.Context, operatorID uint, keyword, domain string, pagination *database.Pagination) ([]model.Role, error) {
	// 1. 获取操作者的最高等级
	operatorLevel, err := s.GetUserMaxRoleLevel(ctx, operatorID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator level: %w", err)
	}

	s.logger.Info("SearchRolesFiltered: 开始过滤",
		"operatorID", operatorID,
		"operatorLevel", operatorLevel,
		"keyword", keyword,
		"domain", domain,
	)

	// 2. 搜索所有匹配的角色
	roles, err := s.roleRepo.Search(ctx, keyword, domain, pagination)
	if err != nil {
		return nil, err
	}

	s.logger.Info("SearchRolesFiltered: 查询到角色",
		"totalRoles", len(roles),
	)

	// 3. 过滤：只返回等级严格低于操作者的角色
	filteredRoles := make([]model.Role, 0)
	for _, role := range roles {
		s.logger.Info("SearchRolesFiltered: 检查角色",
			"roleName", role.Name,
			"roleLevel", role.Level,
			"operatorLevel", operatorLevel,
			"willInclude", role.Level < operatorLevel,
		)
		if role.Level < operatorLevel {
			filteredRoles = append(filteredRoles, role)
		}
	}

	s.logger.Info("SearchRolesFiltered: 过滤完成",
		"filteredCount", len(filteredRoles),
	)

	// 4. 更新分页信息（过滤后的总数）
	pagination.Total = int64(len(filteredRoles))

	return filteredRoles, nil
}

// ============================
// 权限管理
// ============================

// CreatePermission 创建权限
func (s *rbacService) CreatePermission(ctx context.Context, permission *model.Permission) error {
	// 检查权限名称是否已存在
	exists, err := s.permRepo.ExistsByName(ctx, permission.Name, permission.Domain, 0)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}
	if exists {
		return fmt.Errorf("permission name %s already exists in domain %s", permission.Name, permission.Domain)
	}

	// 创建权限
	if err := s.permRepo.Create(ctx, permission); err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	s.logger.Info("permission created",
		"permission_id", permission.ID,
		"permission_name", permission.Name,
		"domain", permission.Domain,
	)

	return nil
}

// UpdatePermission 更新权限
func (s *rbacService) UpdatePermission(ctx context.Context, permission *model.Permission) error {
	// 检查权限是否存在
	oldPerm, err := s.permRepo.FindByID(ctx, permission.ID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// 检查权限名称是否重复
	exists, err := s.permRepo.ExistsByName(ctx, permission.Name, permission.Domain, permission.ID)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}
	if exists {
		return fmt.Errorf("permission name %s already exists in domain %s", permission.Name, permission.Domain)
	}

	// 如果资源或操作发生变化，需要更新 Casbin 策略
	if oldPerm.Resource != permission.Resource || oldPerm.Action != permission.Action {
		// 获取所有使用该权限的角色
		policies, err := s.enforcer.GetPolicy()
		if err != nil {
			return fmt.Errorf("failed to get policies: %w", err)
		}

		// 更新策略
		for _, policy := range policies {
			if len(policy) >= 4 && policy[1] == permission.Domain &&
				policy[2] == oldPerm.Resource && policy[3] == oldPerm.Action {
				// 删除旧策略
				if _, err := s.enforcer.RemovePolicy(policy[0], policy[1], policy[2], policy[3]); err != nil {
					s.logger.Error("failed to remove old policy", "error", err)
				}
				// 添加新策略
				if _, err := s.enforcer.AddPolicy(policy[0], permission.Domain, permission.Resource, permission.Action); err != nil {
					s.logger.Error("failed to add new policy", "error", err)
				}
			}
		}
	}

	// 更新权限记录
	if err := s.permRepo.Update(ctx, permission); err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	s.logger.Info("permission updated",
		"permission_id", permission.ID,
		"permission_name", permission.Name,
		"domain", permission.Domain,
	)

	return nil
}

// DeletePermission 删除权限
func (s *rbacService) DeletePermission(ctx context.Context, id uint) error {
	// 检查权限是否存在
	permission, err := s.permRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// 检查是否为系统权限
	if permission.IsSystem {
		return fmt.Errorf("cannot delete system permission")
	}

	// 删除所有使用该权限的策略
	policies, err := s.enforcer.GetPolicy()
	if err != nil {
		return fmt.Errorf("failed to get policies: %w", err)
	}

	for _, policy := range policies {
		if len(policy) >= 4 && policy[1] == permission.Domain &&
			policy[2] == permission.Resource && policy[3] == permission.Action {
			if _, err := s.enforcer.RemovePolicy(policy[0], policy[1], policy[2], policy[3]); err != nil {
				s.logger.Error("failed to remove policy", "error", err)
			}
		}
	}

	// 删除权限记录
	if err := s.permRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	s.logger.Info("permission deleted",
		"permission_id", id,
		"permission_name", permission.Name,
		"domain", permission.Domain,
	)

	return nil
}

// GetPermission 获取权限详情
func (s *rbacService) GetPermission(ctx context.Context, id uint) (*model.Permission, error) {
	return s.permRepo.FindByID(ctx, id)
}

// ListPermissions 查询权限列表
func (s *rbacService) ListPermissions(ctx context.Context, domain string, pagination *database.Pagination) ([]model.Permission, error) {
	return s.permRepo.List(ctx, domain, pagination)
}

// ListPermissionTree 查询权限树
func (s *rbacService) ListPermissionTree(ctx context.Context, domain string) ([]model.Permission, error) {
	return s.permRepo.ListTree(ctx, domain)
}

// ListPermissionsByType 根据类型查询权限
func (s *rbacService) ListPermissionsByType(ctx context.Context, permType model.PermissionType, domain string) ([]model.Permission, error) {
	return s.permRepo.ListByType(ctx, permType, domain)
}

// ListPermissionsTree 查询权限树（别名方法，用于兼容）
func (s *rbacService) ListPermissionsTree(ctx context.Context, domain string) ([]model.Permission, error) {
	return s.permRepo.ListTree(ctx, domain)
}

// SearchPermissions 搜索权限
func (s *rbacService) SearchPermissions(ctx context.Context, keyword, domain string, pagination *database.Pagination) ([]model.Permission, error) {
	return s.permRepo.Search(ctx, keyword, domain, pagination)
}

// ============================
// 角色-权限管理
// ============================
// 角色权限管理（方案C：Diff API）
// ============================

// UpdateRolePermissions 更新角色权限（支持预览和执行）
// preview=true: 返回PermissionDiff（预览变更）
// preview=false: 返回ChangeResult（执行变更）
func (s *rbacService) UpdateRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint, domain string, preview bool) (interface{}, error) {
	// 1. 检查角色是否存在
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	if role.Domain != domain {
		return nil, fmt.Errorf("role domain mismatch")
	}

	// 2. 获取当前权限列表
	currentPerms, err := s.GetRolePermissions(ctx, roleID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get current permissions: %w", err)
	}

	// 3. 构建当前权限ID集合
	currentIDs := make(map[uint]bool)
	for _, p := range currentPerms {
		currentIDs[p.ID] = true
	}

	// 4. 构建新权限ID集合
	newIDs := make(map[uint]bool)
	for _, id := range permissionIDs {
		newIDs[id] = true
	}

	// 5. 计算差异
	var toAddIDs, toRemoveIDs []uint

	// 找出要删除的权限（在当前中但不在新列表中）
	for _, p := range currentPerms {
		if !newIDs[p.ID] {
			toRemoveIDs = append(toRemoveIDs, p.ID)
		}
	}

	// 找出要添加的权限（在新列表中但不在当前中）
	for id := range newIDs {
		if !currentIDs[id] {
			toAddIDs = append(toAddIDs, id)
		}
	}

	// 6. 如果是预览模式，返回差异信息
	if preview {
		var added, removed, kept []model.Permission

		// 获取要添加的权限详情
		if len(toAddIDs) > 0 {
			added, err = s.permRepo.ListByIDs(ctx, toAddIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to get added permissions: %w", err)
			}
			// 验证域匹配
			for _, perm := range added {
				if perm.Domain != domain {
					return nil, fmt.Errorf("permission domain mismatch: %s (expected: %s, got: %s)", perm.Name, domain, perm.Domain)
				}
			}
		}

		// 获取要删除的权限详情
		if len(toRemoveIDs) > 0 {
			removed, err = s.permRepo.ListByIDs(ctx, toRemoveIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to get removed permissions: %w", err)
			}
		}

		// 保留的权限
		for _, p := range currentPerms {
			if newIDs[p.ID] {
				kept = append(kept, p)
			}
		}

		diff := map[string]interface{}{
			"added":   added,
			"removed": removed,
			"kept":    kept,
		}

		s.logger.Info("preview permission changes",
			"role_id", roleID,
			"role_name", role.Name,
			"added_count", len(added),
			"removed_count", len(removed),
			"kept_count", len(kept),
		)

		return diff, nil
	}

	// 7. 执行模式：应用变更
	if len(toAddIDs) == 0 && len(toRemoveIDs) == 0 {
		return map[string]interface{}{
			"added_count":   0,
			"removed_count": 0,
		}, nil
	}

	// 使用事务确保原子性
	err = s.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除权限
		if len(toRemoveIDs) > 0 {
			removePerms, err := s.permRepo.ListByIDs(ctx, toRemoveIDs)
			if err != nil {
				return fmt.Errorf("failed to get permissions to remove: %w", err)
			}
			if err := tx.Model(role).Association("Permissions").Delete(removePerms); err != nil {
				return fmt.Errorf("failed to remove permissions: %w", err)
			}
		}

		// 添加权限
		if len(toAddIDs) > 0 {
			addPerms, err := s.permRepo.ListByIDs(ctx, toAddIDs)
			if err != nil {
				return fmt.Errorf("failed to get permissions to add: %w", err)
			}
			// 验证域匹配
			for _, perm := range addPerms {
				if perm.Domain != domain {
					return fmt.Errorf("permission domain mismatch: %s", perm.Name)
				}
			}
			if err := tx.Model(role).Association("Permissions").Append(addPerms); err != nil {
				return fmt.Errorf("failed to add permissions: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 8. 清理缓存
	roleCacheKey := fmt.Sprintf(cacheKeyRolePermissions, roleID, domain)
	if err := cache.Del(ctx, roleCacheKey); err != nil {
		s.logger.Warn("failed to delete role permissions cache", "error", err)
	}
	s.clearUserPermissionsCacheByRole(ctx, roleID, domain)

	// 9. 返回变更结果
	result := map[string]interface{}{
		"added_count":   len(toAddIDs),
		"removed_count": len(toRemoveIDs),
	}

	s.logger.Info("permissions updated for role",
		"role_id", roleID,
		"role_name", role.Name,
		"added_count", len(toAddIDs),
		"removed_count", len(toRemoveIDs),
		"domain", domain,
	)

	return result, nil
}

// AssignPermissionsToRole 给角色分配权限（已废弃，保留向后兼容）
// @Deprecated 请使用 UpdateRolePermissions 替代
// 方案A实现：直接操作RBAC表（role_permissions），Casbin从RBAC表自动同步
func (s *rbacService) AssignPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint, domain string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if role.Domain != domain {
		return fmt.Errorf("role domain mismatch")
	}

	// 查询权限列表（用于验证）
	permissions, err := s.permRepo.ListByIDs(ctx, permissionIDs)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	// 验证权限域是否匹配
	for _, perm := range permissions {
		if perm.Domain != domain {
			s.logger.Warn("permission domain mismatch, skipped",
				"permission_id", perm.ID,
				"expected_domain", domain,
				"actual_domain", perm.Domain,
			)
			return fmt.Errorf("permission domain mismatch: %s", perm.Name)
		}
	}

	// 直接使用GORM关联更新role_permissions表
	// 使用Association.Replace替换现有的权限关联
	if err := s.db.DB.WithContext(ctx).Model(role).Association("Permissions").Replace(permissions); err != nil {
		return fmt.Errorf("failed to assign permissions: %w", err)
	}

	// 清理角色权限缓存
	roleCacheKey := fmt.Sprintf(cacheKeyRolePermissions, roleID, domain)
	if err := cache.Del(ctx, roleCacheKey); err != nil {
		s.logger.Warn("failed to delete role permissions cache", "error", err)
	}

	// 清理所有拥有该角色的用户的权限缓存
	s.clearUserPermissionsCacheByRole(ctx, roleID, domain)

	s.logger.Info("permissions assigned to role in RBAC table",
		"role_id", roleID,
		"role_name", role.Name,
		"permission_count", len(permissions),
		"domain", domain,
	)

	return nil
}

// RevokePermissionsFromRole 撤销角色的权限
// 方案A实现：直接从RBAC表删除关联，Casbin自动同步
func (s *rbacService) RevokePermissionsFromRole(ctx context.Context, roleID uint, permissionIDs []uint, domain string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if role.Domain != domain {
		return fmt.Errorf("role domain mismatch")
	}

	// 查询权限列表
	permissions, err := s.permRepo.ListByIDs(ctx, permissionIDs)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	// 从GORM关联删除role_permissions表中的记录
	if err := s.db.DB.WithContext(ctx).Model(role).Association("Permissions").Delete(permissions); err != nil {
		return fmt.Errorf("failed to revoke permissions: %w", err)
	}

	// 清理角色权限缓存
	roleCacheKey := fmt.Sprintf(cacheKeyRolePermissions, roleID, domain)
	if err := cache.Del(ctx, roleCacheKey); err != nil {
		s.logger.Warn("failed to delete role permissions cache", "error", err)
	}

	// 清理所有拥有该角色的用户的权限缓存
	s.clearUserPermissionsCacheByRole(ctx, roleID, domain)

	s.logger.Info("permissions revoked from role in RBAC table",
		"role_id", roleID,
		"role_name", role.Name,
		"permission_count", len(permissions),
		"domain", domain,
	)

	return nil
}

// GetRolePermissions 获取角色的所有权限
// 方案A实现：直接从RBAC表（role_permissions）读取
func (s *rbacService) GetRolePermissions(ctx context.Context, roleID uint, domain string) ([]model.Permission, error) {
	// 检查角色是否存在
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	if role.Domain != domain {
		return nil, fmt.Errorf("role domain mismatch")
	}

	// 直接从RBAC表读取角色的权限（通过GORM Preload）
	var roleWithPerms model.Role
	if err := s.db.DB.WithContext(ctx).
		Preload("Permissions").
		Where("id = ?", roleID).
		First(&roleWithPerms).Error; err != nil {
		return nil, fmt.Errorf("failed to load role permissions: %w", err)
	}

	s.logger.Debug("loaded role permissions from RBAC table",
		"role_id", roleID,
		"permission_count", len(roleWithPerms.Permissions),
	)

	return roleWithPerms.Permissions, nil
}

// 继续下一部分...
// ============================
// 用户-角色管理
// ============================

// AssignRolesToUser 给用户分配角色
// 方案A实现：只在user_roles表中记录，Casbin从RBAC表自动同步
func (s *rbacService) AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint, domain string, assignedBy uint) error {
	// 查询角色列表
	roles, err := s.roleRepo.ListByIDs(ctx, roleIDs)
	if err != nil {
		return fmt.Errorf("failed to get roles: %w", err)
	}

	// 构建用户-角色关系
	var userRoles []model.UserRole
	for _, role := range roles {
		if role.Domain != domain {
			s.logger.Warn("role domain mismatch, skipped",
				"role_id", role.ID,
				"expected_domain", domain,
				"actual_domain", role.Domain,
			)
			continue
		}

		userRoles = append(userRoles, model.UserRole{
			UserID:     userID,
			RoleID:     role.ID,
			Domain:     domain,
			AssignedBy: assignedBy,
		})
	}

	// 批量写入user_roles表
	if len(userRoles) > 0 {
		if err := s.userRoleRepo.BatchAssign(ctx, userRoles); err != nil {
			return fmt.Errorf("failed to assign roles: %w", err)
		}
	}

	// 清理用户权限缓存
	userCacheKey := fmt.Sprintf(cacheKeyUserPermissions, userID, domain)
	if err := cache.Del(ctx, userCacheKey); err != nil {
		s.logger.Warn("failed to delete user permissions cache", "error", err)
	}

	s.logger.Info("roles assigned to user in RBAC table",
		"user_id", userID,
		"role_count", len(userRoles),
		"domain", domain,
		"assigned_by", assignedBy,
	)

	return nil
}

// RevokeRolesFromUser 撤销用户的角色
// 方案A实现：只从user_roles表删除，自动清理缓存
func (s *rbacService) RevokeRolesFromUser(ctx context.Context, userID uint, roleIDs []uint, domain string) error {
	// 从user_roles表批量删除
	for _, roleID := range roleIDs {
		if err := s.userRoleRepo.Revoke(ctx, userID, roleID, domain); err != nil {
			s.logger.Error("failed to revoke role from user",
				"user_id", userID,
				"role_id", roleID,
				"error", err,
			)
			return fmt.Errorf("failed to revoke role: %w", err)
		}
	}

	// 清理用户权限缓存
	userCacheKey := fmt.Sprintf(cacheKeyUserPermissions, userID, domain)
	if err := cache.Del(ctx, userCacheKey); err != nil {
		s.logger.Warn("failed to delete user permissions cache", "error", err)
	}

	s.logger.Info("roles revoked from user in RBAC table",
		"user_id", userID,
		"role_count", len(roleIDs),
		"domain", domain,
	)

	return nil
}

// GetUserRoles 获取用户的所有角色
func (s *rbacService) GetUserRoles(ctx context.Context, userID uint, domain string) ([]model.Role, error) {
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	// 从 Casbin 获取用户的角色
	roleIDStrs, err := s.enforcer.GetRolesForUser(userIDStr, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 转换为角色ID
	var roleIDs []uint
	for _, roleIDStr := range roleIDStrs {
		roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
		if err != nil {
			s.logger.Error("invalid role ID", "role_id_str", roleIDStr, "error", err)
			continue
		}
		roleIDs = append(roleIDs, uint(roleID))
	}

	// 查询角色详情
	if len(roleIDs) == 0 {
		return []model.Role{}, nil
	}

	return s.roleRepo.ListByIDs(ctx, roleIDs)
}

// GetRoleUsers 获取拥有某个角色的所有用户
func (s *rbacService) GetRoleUsers(ctx context.Context, roleID uint) ([]model.UserRole, error) {
	return s.userRoleRepo.FindByRole(ctx, roleID)
}

// ============================
// 权限验证
// ============================

// CheckPermission 检查用户是否有权限
// 方案A实现：直接从RBAC表查询权限，支持通配符匹配
func (s *rbacService) CheckPermission(ctx context.Context, userID uint, domain, resource, action string) (bool, error) {
	// 1. 获取用户的所有权限
	permissions, err := s.GetUserPermissions(ctx, userID, domain)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 2. 检查是否有匹配的权限
	// 支持精确匹配和通配符匹配（*）
	for _, perm := range permissions {
		// 精确匹配
		if perm.Resource == resource && perm.Action == action {
			return true, nil
		}

		// 通配符匹配：resource或action为*表示匹配所有
		if (perm.Resource == "*" || perm.Resource == resource) &&
			(perm.Action == "*" || perm.Action == action) {
			return true, nil
		}
	}

	s.logger.Debug("permission check failed",
		"user_id", userID,
		"domain", domain,
		"resource", resource,
		"action", action,
	)

	return false, nil
}

// GetUserPermissions 获取用户的所有权限（包括通过角色继承的）
// 重要：此方法供前端调用，用于生成动态路由和菜单
// 方案A实现：从RBAC表（user_roles + role_permissions + permissions）联表查询 + Redis缓存
func (s *rbacService) GetUserPermissions(ctx context.Context, userID uint, domain string) ([]model.Permission, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf(cacheKeyUserPermissions, userID, domain)
	var cachedPermissions []model.Permission

	err := s.cache.GetObject(ctx, cacheKey, &cachedPermissions)
	if err == nil {
		s.logger.Debug("user permissions loaded from cache",
			"user_id", userID,
			"domain", domain,
			"permission_count", len(cachedPermissions),
		)
		return cachedPermissions, nil
	}

	// 2. 缓存未命中，从数据库查询
	permissions, err := s.loadUserPermissionsFromDB(ctx, userID, domain)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	if err := s.cache.SetObject(ctx, cacheKey, permissions, cacheTTLPermissions); err != nil {
		s.logger.Warn("failed to cache user permissions", "error", err)
	}

	return permissions, nil
}

// loadUserPermissionsFromDB 从数据库加载用户权限
func (s *rbacService) loadUserPermissionsFromDB(ctx context.Context, userID uint, domain string) ([]model.Permission, error) {
	// 1. 查询用户的所有角色（从user_roles表）
	userRoles, err := s.userRoleRepo.FindByUser(ctx, userID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(userRoles) == 0 {
		s.logger.Debug("user has no roles", "user_id", userID, "domain", domain)
		return []model.Permission{}, nil
	}

	// 2. 提取角色ID列表
	roleIDs := make([]uint, 0, len(userRoles))
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	// 3. 联表查询所有权限（去重）
	// SQL: SELECT DISTINCT permissions.* FROM permissions
	//      INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id
	//      WHERE role_permissions.role_id IN (?) AND permissions.domain = ?
	var permissions []model.Permission
	if err := s.db.DB.WithContext(ctx).
		Distinct().
		Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id IN ? AND permissions.domain = ?", roleIDs, domain).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}

	s.logger.Debug("loaded user permissions from RBAC tables",
		"user_id", userID,
		"role_count", len(roleIDs),
		"permission_count", len(permissions),
	)

	return permissions, nil
}

// clearUserPermissionsCacheByRole 清理拥有指定角色的所有用户的权限缓存
func (s *rbacService) clearUserPermissionsCacheByRole(ctx context.Context, roleID uint, domain string) {
	// 查询所有拥有该角色的用户
	userRoles, err := s.userRoleRepo.FindByRole(ctx, roleID)
	if err != nil {
		s.logger.Warn("failed to find users by role", "role_id", roleID, "error", err)
		return
	}

	// 清理每个用户的权限缓存
	for _, ur := range userRoles {
		if ur.Domain == domain {
			userCacheKey := fmt.Sprintf(cacheKeyUserPermissions, ur.UserID, domain)
			if err := cache.Del(ctx, userCacheKey); err != nil {
				s.logger.Warn("failed to delete user permissions cache",
					"user_id", ur.UserID,
					"error", err,
				)
			}
		}
	}

	s.logger.Debug("cleared user permissions cache for role",
		"role_id", roleID,
		"user_count", len(userRoles),
	)
}

// ============================
// 策略管理（高级）
// ============================

// AddPolicy 添加策略
func (s *rbacService) AddPolicy(ctx context.Context, sub, dom, obj, act string) error {
	_, err := s.enforcer.AddPolicy(sub, dom, obj, act)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	s.logger.Info("policy added",
		"subject", sub,
		"domain", dom,
		"object", obj,
		"action", act,
	)

	return nil
}

// RemovePolicy 删除策略
func (s *rbacService) RemovePolicy(ctx context.Context, sub, dom, obj, act string) error {
	_, err := s.enforcer.RemovePolicy(sub, dom, obj, act)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	s.logger.Info("policy removed",
		"subject", sub,
		"domain", dom,
		"object", obj,
		"action", act,
	)

	return nil
}

// ListPolicies 列出所有策略
func (s *rbacService) ListPolicies(ctx context.Context, domain string) ([][]string, error) {
	policies, err := s.enforcer.GetPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	// 如果指定了域，过滤策略
	if domain != "" {
		var filteredPolicies [][]string
		for _, policy := range policies {
			if len(policy) >= 2 && policy[1] == domain {
				filteredPolicies = append(filteredPolicies, policy)
			}
		}
		return filteredPolicies, nil
	}

	return policies, nil
}

// ============================
// 安全辅助函数
// ============================

// GetUserMaxRoleLevel 获取用户在指定域下的最高角色等级
// 用于权限越级检查：操作者只能管理比自己等级低的角色
func (s *rbacService) GetUserMaxRoleLevel(ctx context.Context, userID uint, domain string) (int, error) {
	roles, err := s.GetUserRoles(ctx, userID, domain)
	if err != nil {
		return 0, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(roles) == 0 {
		s.logger.Warn("GetUserMaxRoleLevel: 用户没有任何角色",
			"userID", userID,
			"domain", domain,
		)
		return 0, nil // 用户没有任何角色
	}

	maxLevel := 0
	for _, role := range roles {
		s.logger.Info("GetUserMaxRoleLevel: 检查角色",
			"userID", userID,
			"roleName", role.Name,
			"roleLevel", role.Level,
		)
		if role.Level > maxLevel {
			maxLevel = role.Level
		}
	}

	s.logger.Info("GetUserMaxRoleLevel: 计算完成",
		"userID", userID,
		"maxLevel", maxLevel,
		"rolesCount", len(roles),
	)

	return maxLevel, nil
}

// CheckRoleLevelPermission 检查用户是否有权限操作指定等级的角色
// 规则：只能操作比自己等级严格低的角色（不包括同级）
// 这样可以防止同级用户互相修改权限
func (s *rbacService) CheckRoleLevelPermission(ctx context.Context, operatorID uint, targetRoleID uint, domain string) error {
	// 获取操作者的最高角色等级
	operatorLevel, err := s.GetUserMaxRoleLevel(ctx, operatorID, domain)
	if err != nil {
		return fmt.Errorf("failed to get operator level: %w", err)
	}

	// 获取目标角色
	targetRole, err := s.GetRole(ctx, targetRoleID)
	if err != nil {
		return fmt.Errorf("failed to get target role: %w", err)
	}

	// 检查等级：操作者等级必须严格高于目标角色等级
	if operatorLevel <= targetRole.Level {
		return fmt.Errorf("权限不足：无法操作等级为 %d 的角色（您的等级为 %d），只能管理比您等级低的角色", targetRole.Level, operatorLevel)
	}

	return nil
}

// CheckRolesLevelPermission 批量检查用户是否有权限操作指定的多个角色
// 规则：只有当操作者等级严格高于所有目标角色时才允许操作
func (s *rbacService) CheckRolesLevelPermission(ctx context.Context, operatorID uint, targetRoleIDs []uint, domain string) error {
	// 获取操作者的最高角色等级
	operatorLevel, err := s.GetUserMaxRoleLevel(ctx, operatorID, domain)
	if err != nil {
		return fmt.Errorf("failed to get operator level: %w", err)
	}

	// 检查每个目标角色的等级
	for _, roleID := range targetRoleIDs {
		targetRole, err := s.GetRole(ctx, roleID)
		if err != nil {
			return fmt.Errorf("failed to get target role %d: %w", roleID, err)
		}

		// 检查等级：操作者等级必须严格高于目标角色等级
		if operatorLevel <= targetRole.Level {
			return fmt.Errorf("权限不足：无法分配等级为 %d 的角色 '%s'（您的等级为 %d），只能分配比您等级低的角色", targetRole.Level, targetRole.DisplayName, operatorLevel)
		}
	}

	return nil
}
