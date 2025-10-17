package casbin

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// Enforcer 是 Casbin enforcer 的企业级封装
type Enforcer struct {
	enforcer     *casbin.Enforcer
	adapter      *gormadapter.Adapter
	mu           sync.RWMutex
	autoSave     bool
	autoLoad     bool
	autoLoadTick time.Duration
	stopAutoLoad chan struct{}
	logger       *slog.Logger
}

// Config Casbin 配置
type Config struct {
	ModelPath    string        // 模型文件路径
	AutoSave     bool          // 是否自动保存策略
	AutoLoad     bool          // 是否自动加载策略（多实例同步）
	AutoLoadTick time.Duration // 自动加载间隔
}

// NewEnforcer 创建新的 Casbin enforcer
func NewEnforcer(db *gorm.DB, cfg Config, logger *slog.Logger) (*Enforcer, error) {
	// 初始化 GORM 适配器
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// 创建 enforcer
	e, err := casbin.NewEnforcer(cfg.ModelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// 设置自动保存
	e.EnableAutoSave(cfg.AutoSave)

	// 加载策略
	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	enforcer := &Enforcer{
		enforcer:     e,
		adapter:      adapter,
		autoSave:     cfg.AutoSave,
		autoLoad:     cfg.AutoLoad,
		autoLoadTick: cfg.AutoLoadTick,
		stopAutoLoad: make(chan struct{}),
		logger:       logger,
	}

	// 启动自动加载（用于多实例部署时的策略同步）
	if cfg.AutoLoad && cfg.AutoLoadTick > 0 {
		go enforcer.startAutoLoad()
	}

	logger.Info("casbin enforcer initialized successfully",
		"model", cfg.ModelPath,
		"autoSave", cfg.AutoSave,
		"autoLoad", cfg.AutoLoad,
	)

	return enforcer, nil
}

// startAutoLoad 定期从数据库重新加载策略（多实例同步）
func (e *Enforcer) startAutoLoad() {
	ticker := time.NewTicker(e.autoLoadTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := e.LoadPolicy(); err != nil {
				e.logger.Error("failed to auto-load policy", "error", err)
			} else {
				e.logger.Debug("policy auto-loaded successfully")
			}
		case <-e.stopAutoLoad:
			e.logger.Info("auto-load stopped")
			return
		}
	}
}

// Close 关闭 enforcer
func (e *Enforcer) Close() error {
	if e.autoLoad {
		close(e.stopAutoLoad)
	}
	return nil
}

// LoadPolicy 从数据库重新加载所有策略
func (e *Enforcer) LoadPolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.LoadPolicy()
}

// SavePolicy 保存所有策略到数据库
func (e *Enforcer) SavePolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.SavePolicy()
}

// ============================
// 权限验证方法
// ============================

// Enforce 验证权限
// sub: 用户ID, dom: 域/租户, obj: 资源, act: 操作
func (e *Enforcer) Enforce(sub, dom, obj, act string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.Enforce(sub, dom, obj, act)
}

// EnforceWithContext 带上下文的权限验证
func (e *Enforcer) EnforceWithContext(ctx context.Context, sub, dom, obj, act string) (bool, error) {
	// TODO: 可以在这里添加上下文超时控制
	return e.Enforce(sub, dom, obj, act)
}

// BatchEnforce 批量验证权限
func (e *Enforcer) BatchEnforce(requests [][]interface{}) ([]bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.BatchEnforce(requests)
}

// ============================
// 策略管理方法（p 表）
// ============================

// AddPolicy 添加权限策略
func (e *Enforcer) AddPolicy(sub, dom, obj, act string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.AddPolicy(sub, dom, obj, act)
}

// AddPolicies 批量添加权限策略
func (e *Enforcer) AddPolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.AddPolicies(rules)
}

// RemovePolicy 删除权限策略
func (e *Enforcer) RemovePolicy(sub, dom, obj, act string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.RemovePolicy(sub, dom, obj, act)
}

// RemovePolicies 批量删除权限策略
func (e *Enforcer) RemovePolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.RemovePolicies(rules)
}

// RemoveFilteredPolicy 根据过滤条件删除策略
// fieldIndex: 0=sub, 1=dom, 2=obj, 3=act
func (e *Enforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.RemoveFilteredPolicy(fieldIndex, fieldValues...)
}

// GetPolicy 获取所有权限策略
func (e *Enforcer) GetPolicy() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetPolicy()
}

// GetFilteredPolicy 获取过滤后的权限策略
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetFilteredPolicy(fieldIndex, fieldValues...)
}

// HasPolicy 检查策略是否存在
func (e *Enforcer) HasPolicy(params ...interface{}) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.HasPolicy(params...)
}

// ============================
// 角色管理方法（g 表）
// ============================

// AddRoleForUser 给用户分配角色
func (e *Enforcer) AddRoleForUser(user, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.AddRoleForUser(user, role, domain)
}

// AddRolesForUser 给用户批量分配角色
func (e *Enforcer) AddRolesForUser(user string, roles []string, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var rules [][]string
	for _, role := range roles {
		rules = append(rules, []string{user, role, domain})
	}
	return e.enforcer.AddGroupingPolicies(rules)
}

// DeleteRoleForUser 删除用户的角色
func (e *Enforcer) DeleteRoleForUser(user, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.DeleteRoleForUser(user, role, domain)
}

// DeleteRolesForUser 删除用户的所有角色
func (e *Enforcer) DeleteRolesForUser(user, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.DeleteRolesForUser(user, domain)
}

// GetRolesForUser 获取用户的所有角色
func (e *Enforcer) GetRolesForUser(user, domain string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetRolesForUser(user, domain)
}

// GetUsersForRole 获取拥有某个角色的所有用户
func (e *Enforcer) GetUsersForRole(role, domain string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetUsersForRole(role, domain)
}

// HasRoleForUser 检查用户是否有某个角色
func (e *Enforcer) HasRoleForUser(user, role, domain string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.HasRoleForUser(user, role, domain)
}

// ============================
// 角色继承管理（g2 表）
// ============================

// AddRoleInheritance 添加角色继承关系（role1 继承 role2 的权限）
func (e *Enforcer) AddRoleInheritance(role1, role2, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.AddNamedGroupingPolicy("g2", role1, role2, domain)
}

// DeleteRoleInheritance 删除角色继承关系
func (e *Enforcer) DeleteRoleInheritance(role1, role2, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.RemoveNamedGroupingPolicy("g2", role1, role2, domain)
}

// GetRoleInheritance 获取角色继承的所有父角色
func (e *Enforcer) GetRoleInheritance(role, domain string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies, err := e.enforcer.GetNamedGroupingPolicy("g2")
	if err != nil {
		return nil, err
	}

	var parentRoles []string
	for _, policy := range policies {
		if len(policy) >= 3 && policy[0] == role && policy[2] == domain {
			parentRoles = append(parentRoles, policy[1])
		}
	}
	return parentRoles, nil
}

// ============================
// 高级查询方法
// ============================

// GetPermissionsForUser 获取用户的所有权限（包括通过角色继承的）
func (e *Enforcer) GetPermissionsForUser(user, domain string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetPermissionsForUser(user, domain)
}

// GetImplicitPermissionsForUser 获取用户的隐式权限（包括角色继承）
func (e *Enforcer) GetImplicitPermissionsForUser(user, domain string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetImplicitPermissionsForUser(user, domain)
}

// GetImplicitRolesForUser 获取用户的隐式角色（包括角色继承链）
func (e *Enforcer) GetImplicitRolesForUser(user, domain string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetImplicitRolesForUser(user, domain)
}

// GetAllSubjects 获取所有主体（用户）
func (e *Enforcer) GetAllSubjects() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllSubjects()
}

// GetAllObjects 获取所有对象（资源）
func (e *Enforcer) GetAllObjects() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllObjects()
}

// GetAllActions 获取所有操作
func (e *Enforcer) GetAllActions() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllActions()
}

// GetAllRoles 获取所有角色
func (e *Enforcer) GetAllRoles() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllRoles()
}

// ============================
// 域管理方法
// ============================

// GetAllDomains 获取所有域
func (e *Enforcer) GetAllDomains() ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetAllDomains()
}

// DeleteDomain 删除整个域的所有策略和角色
func (e *Enforcer) DeleteDomain(domain string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 删除该域的所有权限策略
	if _, err := e.enforcer.RemoveFilteredPolicy(1, domain); err != nil {
		return fmt.Errorf("failed to remove policies for domain: %w", err)
	}

	// 删除该域的所有角色分配
	if _, err := e.enforcer.RemoveFilteredGroupingPolicy(2, domain); err != nil {
		return fmt.Errorf("failed to remove role assignments for domain: %w", err)
	}

	// 删除该域的所有角色继承
	if _, err := e.enforcer.RemoveFilteredNamedGroupingPolicy("g2", 2, domain); err != nil {
		return fmt.Errorf("failed to remove role inheritance for domain: %w", err)
	}

	return nil
}

// ============================
// 权限检查辅助方法
// ============================

// GetPoliciesForRole 获取角色的所有权限策略
func (e *Enforcer) GetPoliciesForRole(role, domain string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.GetFilteredPolicy(0, role, domain)
}

// AddPoliciesForRole 批量给角色添加权限
func (e *Enforcer) AddPoliciesForRole(role, domain string, permissions [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var rules [][]string
	for _, perm := range permissions {
		if len(perm) >= 2 {
			rules = append(rules, []string{role, domain, perm[0], perm[1]})
		}
	}
	return e.enforcer.AddPolicies(rules)
}

// RemoveAllPoliciesForRole 删除角色的所有权限
func (e *Enforcer) RemoveAllPoliciesForRole(role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.RemoveFilteredPolicy(0, role, domain)
}
