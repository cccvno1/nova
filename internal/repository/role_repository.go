package repository

import (
	"context"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/pkg/database"
)

// RoleRepository 角色仓储接口
// 提供角色数据的持久化操作，包括：
// - 基础CRUD操作（Create, Update, Delete, FindByID）
// - 业务查询（按名称、域查询）
// - 分页和搜索功能
type RoleRepository interface {
	// 基础CRUD方法
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Role, error)

	// 业务查询方法
	FindByName(ctx context.Context, name, domain string) (*model.Role, error)                                  // 按名称查询
	List(ctx context.Context, domain string, pagination *database.Pagination) ([]model.Role, error)            // 分页查询
	ListByIDs(ctx context.Context, ids []uint) ([]model.Role, error)                                           // 批量查询
	Search(ctx context.Context, keyword, domain string, pagination *database.Pagination) ([]model.Role, error) // 关键词搜索
	ListByDomain(ctx context.Context, domain string) ([]model.Role, error)                                     // 按域查询所有启用角色
	ExistsByName(ctx context.Context, name, domain string, excludeID uint) (bool, error)                       // 检查名称是否存在
}

// roleRepository 角色仓储实现
// 继承database.Repository提供的基础功能
type roleRepository struct {
	*database.Repository[model.Role]
}

// NewRoleRepository 创建角色仓储实例
func NewRoleRepository(db *database.Database) RoleRepository {
	return &roleRepository{
		Repository: database.NewRepository[model.Role](db.DB),
	}
}

// FindByName 根据角色名称和域查询角色
func (r *roleRepository) FindByName(ctx context.Context, name, domain string) (*model.Role, error) {
	return r.Repository.FindOne(ctx, "name = ? AND domain = ?", name, domain)
}

// List 分页查询角色列表
// 支持按域过滤，domain为空则查询所有域
func (r *roleRepository) List(ctx context.Context, domain string, pagination *database.Pagination) ([]model.Role, error) {
	query := "domain = ?"
	if domain == "" {
		query = "1=1"
		return r.Repository.FindWithPagination(ctx, pagination, query)
	}
	return r.Repository.FindWithPagination(ctx, pagination, query, domain)
}

// ListByIDs 根据ID列表批量查询角色
// 用于Service层的批量角色查询，避免N+1查询
func (r *roleRepository) ListByIDs(ctx context.Context, ids []uint) ([]model.Role, error) {
	if len(ids) == 0 {
		return []model.Role{}, nil
	}
	return r.Repository.FindByCondition(ctx, "id IN ?", ids)
}

// Search 根据关键词搜索角色
// 支持按名称、显示名称、描述模糊查询
func (r *roleRepository) Search(ctx context.Context, keyword, domain string, pagination *database.Pagination) ([]model.Role, error) {
	var roles []model.Role

	db := r.Repository.DB().WithContext(ctx).Model(&model.Role{})
	if domain != "" {
		db = db.Where("domain = ?", domain)
	}
	if keyword != "" {
		db = db.Where("name LIKE ? OR display_name LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	if err := db.Count(&pagination.Total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	err := db.Order("sort DESC, id DESC").
		Scopes(database.Paginate(pagination)).
		Find(&roles).Error
	return roles, err
}

// ListByDomain 根据域查询所有启用的角色（status=1）
func (r *roleRepository) ListByDomain(ctx context.Context, domain string) ([]model.Role, error) {
	return r.Repository.FindByCondition(ctx, "domain = ? AND status = ?", domain, 1)
}

// ExistsByName 检查角色名称是否已存在
// excludeID用于更新时排除自身
func (r *roleRepository) ExistsByName(ctx context.Context, name, domain string, excludeID uint) (bool, error) {
	var count int64
	db := r.Repository.DB().WithContext(ctx).Model(&model.Role{}).
		Where("name = ? AND domain = ?", name, domain)

	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}

	err := db.Count(&count).Error
	return count > 0, err
}

// ============================
// UserRole Repository
// ============================

// UserRoleRepository 用户角色关联仓储接口
// 管理用户和角色之间的多对多关系
type UserRoleRepository interface {
	Assign(ctx context.Context, userRole *model.UserRole) error                           // 分配单个角色
	Revoke(ctx context.Context, userID, roleID uint, domain string) error                 // 撤销单个角色
	RevokeAll(ctx context.Context, userID uint, domain string) error                      // 撤销用户在某域的所有角色
	FindByUser(ctx context.Context, userID uint, domain string) ([]model.UserRole, error) // 查询用户的角色列表
	FindByRole(ctx context.Context, roleID uint) ([]model.UserRole, error)                // 查询拥有某角色的用户列表
	HasRole(ctx context.Context, userID, roleID uint, domain string) (bool, error)        // 检查用户是否拥有角色
	BatchAssign(ctx context.Context, userRoles []model.UserRole) error                    // 批量分配角色
}

// userRoleRepository 用户角色关联仓储实现
type userRoleRepository struct {
	db *database.Database
}

// NewUserRoleRepository 创建用户角色仓储实例
func NewUserRoleRepository(db *database.Database) UserRoleRepository {
	return &userRoleRepository{db: db}
}

// Assign 分配角色给用户
func (r *userRoleRepository) Assign(ctx context.Context, userRole *model.UserRole) error {
	return r.db.DB.WithContext(ctx).Create(userRole).Error
}

// Revoke 撤销用户的角色
func (r *userRoleRepository) Revoke(ctx context.Context, userID, roleID uint, domain string) error {
	return r.db.DB.WithContext(ctx).
		Where("user_id = ? AND role_id = ? AND domain = ?", userID, roleID, domain).
		Delete(&model.UserRole{}).Error
}

// RevokeAll 撤销用户在某个域的所有角色
func (r *userRoleRepository) RevokeAll(ctx context.Context, userID uint, domain string) error {
	return r.db.DB.WithContext(ctx).
		Where("user_id = ? AND domain = ?", userID, domain).
		Delete(&model.UserRole{}).Error
}

// FindByUser 查询用户的所有角色
// 使用Preload预加载角色详情，避免N+1查询
func (r *userRoleRepository) FindByUser(ctx context.Context, userID uint, domain string) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	query := r.db.DB.WithContext(ctx).
		Preload("Role").
		Where("user_id = ?", userID)

	if domain != "" {
		query = query.Where("domain = ?", domain)
	}

	err := query.Find(&userRoles).Error
	return userRoles, err
}

// FindByRole 查询拥有某个角色的所有用户
// 使用Preload预加载用户详情，避免N+1查询
func (r *userRoleRepository) FindByRole(ctx context.Context, roleID uint) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := r.db.DB.WithContext(ctx).
		Preload("User").
		Where("role_id = ?", roleID).
		Find(&userRoles).Error
	return userRoles, err
}

// HasRole 检查用户是否拥有某个角色
func (r *userRoleRepository) HasRole(ctx context.Context, userID, roleID uint, domain string) (bool, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&model.UserRole{}).
		Where("user_id = ? AND role_id = ? AND domain = ?", userID, roleID, domain).
		Count(&count).Error
	return count > 0, err
}

// BatchAssign 批量分配角色
// 用于一次性给用户分配多个角色，提升性能
func (r *userRoleRepository) BatchAssign(ctx context.Context, userRoles []model.UserRole) error {
	if len(userRoles) == 0 {
		return nil
	}
	return r.db.DB.WithContext(ctx).Create(&userRoles).Error
}
