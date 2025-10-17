package repository

import (
	"context"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/pkg/database"
)

// PermissionRepository 权限仓储接口
// 提供权限数据的持久化操作，包括：
// - 基础CRUD操作（Create, Update, Delete, FindByID）
// - 业务查询（按名称、类型、分类查询）
// - 树形结构查询
// - 分页和搜索功能
type PermissionRepository interface {
	// 基础CRUD方法
	Create(ctx context.Context, permission *model.Permission) error
	Update(ctx context.Context, permission *model.Permission) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Permission, error)

	// 业务查询方法
	FindByName(ctx context.Context, name, domain string) (*model.Permission, error)                                  // 按名称查询
	List(ctx context.Context, domain string, pagination *database.Pagination) ([]model.Permission, error)            // 分页查询
	ListByIDs(ctx context.Context, ids []uint) ([]model.Permission, error)                                           // 批量查询
	ListByType(ctx context.Context, permType model.PermissionType, domain string) ([]model.Permission, error)        // 按类型查询
	ListByCategory(ctx context.Context, category, domain string) ([]model.Permission, error)                         // 按分类查询
	Search(ctx context.Context, keyword, domain string, pagination *database.Pagination) ([]model.Permission, error) // 关键词搜索
	ListTree(ctx context.Context, domain string) ([]model.Permission, error)                                         // 树形结构查询
	ExistsByName(ctx context.Context, name, domain string, excludeID uint) (bool, error)                             // 检查名称是否存在
}

// permissionRepository 权限仓储实现
// 继承database.Repository提供的基础功能
type permissionRepository struct {
	*database.Repository[model.Permission]
}

// NewPermissionRepository 创建权限仓储实例
func NewPermissionRepository(db *database.Database) PermissionRepository {
	return &permissionRepository{
		Repository: database.NewRepository[model.Permission](db.DB),
	}
}

// FindByName 根据权限名称和域查询权限
func (r *permissionRepository) FindByName(ctx context.Context, name, domain string) (*model.Permission, error) {
	return r.Repository.FindOne(ctx, "name = ? AND domain = ?", name, domain)
}

// List 分页查询权限列表
// 支持按域过滤，domain为空则查询所有域
func (r *permissionRepository) List(ctx context.Context, domain string, pagination *database.Pagination) ([]model.Permission, error) {
	query := "domain = ?"
	if domain == "" {
		query = "1=1"
		return r.Repository.FindWithPagination(ctx, pagination, query)
	}
	return r.Repository.FindWithPagination(ctx, pagination, query, domain)
}

// ListByIDs 根据ID列表批量查询权限
// 用于Service层的批量权限匹配，避免N+1查询
func (r *permissionRepository) ListByIDs(ctx context.Context, ids []uint) ([]model.Permission, error) {
	if len(ids) == 0 {
		return []model.Permission{}, nil
	}
	return r.Repository.FindByCondition(ctx, "id IN ?", ids)
}

// ListByType 根据权限类型查询权限（menu/button/api/data/field）
func (r *permissionRepository) ListByType(ctx context.Context, permType model.PermissionType, domain string) ([]model.Permission, error) {
	if domain == "" {
		return r.Repository.FindByCondition(ctx, "type = ? AND status = ?", permType, 1)
	}
	return r.Repository.FindByCondition(ctx, "type = ? AND domain = ? AND status = ?", permType, domain, 1)
}

// ListByCategory 根据分类查询权限
func (r *permissionRepository) ListByCategory(ctx context.Context, category, domain string) ([]model.Permission, error) {
	if domain == "" {
		return r.Repository.FindByCondition(ctx, "category = ? AND status = ?", category, 1)
	}
	return r.Repository.FindByCondition(ctx, "category = ? AND domain = ? AND status = ?", category, domain, 1)
}

// Search 根据关键词搜索权限
// 支持按名称、显示名称、描述模糊查询
func (r *permissionRepository) Search(ctx context.Context, keyword, domain string, pagination *database.Pagination) ([]model.Permission, error) {
	var permissions []model.Permission

	db := r.Repository.DB().WithContext(ctx).Model(&model.Permission{})
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
		Find(&permissions).Error
	return permissions, err
}

// ListTree 查询树形权限结构
// 根据parent_id构建父子关系的树形结构
func (r *permissionRepository) ListTree(ctx context.Context, domain string) ([]model.Permission, error) {
	var permissions []model.Permission

	if domain == "" {
		permissions, err := r.Repository.FindByCondition(ctx, "status = ?", 1)
		if err != nil {
			return nil, err
		}
		return buildPermissionTree(permissions), nil
	}

	permissions, err := r.Repository.FindByCondition(ctx, "domain = ? AND status = ?", domain, 1)
	if err != nil {
		return nil, err
	}
	return buildPermissionTree(permissions), nil
}

// buildPermissionTree 构建权限树形结构
// 使用两次遍历算法：
// 1. 第一次遍历建立ID到权限的映射
// 2. 第二次遍历建立父子关系
func buildPermissionTree(permissions []model.Permission) []model.Permission {
	// 创建ID到权限的映射
	permMap := make(map[uint]*model.Permission)
	var roots []model.Permission

	// 第一遍遍历，建立映射
	for i := range permissions {
		permissions[i].Children = []model.Permission{}
		permMap[permissions[i].ID] = &permissions[i]
	}

	// 第二遍遍历，构建树形结构
	for i := range permissions {
		if permissions[i].ParentID == 0 {
			// 根节点
			roots = append(roots, permissions[i])
		} else if parent, ok := permMap[permissions[i].ParentID]; ok {
			// 添加到父节点
			parent.Children = append(parent.Children, permissions[i])
		}
	}

	return roots
}

// ExistsByName 检查权限名称是否已存在
// excludeID用于更新时排除自身
func (r *permissionRepository) ExistsByName(ctx context.Context, name, domain string, excludeID uint) (bool, error) {
	var count int64
	db := r.Repository.DB().WithContext(ctx).Model(&model.Permission{}).
		Where("name = ? AND domain = ?", name, domain)

	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}

	err := db.Count(&count).Error
	return count > 0, err
}
