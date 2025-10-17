package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/pkg/casbin"
	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/logger"
)

var configFile = flag.String("config", "../../configs/config.yaml", "config file path")

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 初始化日志
	loggerCfg := &logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
		Output: cfg.Logger.Output,
	}
	if err := logger.Init(loggerCfg); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// 初始化数据库
	if err := database.Init(&cfg.DB); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer database.Close()

	// 初始化 Casbin
	enforcer, err := casbin.NewEnforcer(
		database.GetDB(),
		casbin.Config{
			ModelPath:    cfg.Casbin.ModelPath,
			AutoSave:     cfg.Casbin.AutoSave,
			AutoLoad:     false, // 初始化时不需要自动加载
			AutoLoadTick: 0,
		},
		logger.Logger(),
	)
	if err != nil {
		log.Fatalf("failed to initialize casbin: %v", err)
	}
	defer enforcer.Close()

	ctx := context.Background()

	log.Println("开始初始化 RBAC 种子数据...")

	// 初始化默认域
	defaultDomain := "default"

	// 1. 创建角色
	if err := seedRoles(ctx, defaultDomain); err != nil {
		log.Fatalf("failed to seed roles: %v", err)
	}

	// 2. 创建权限
	if err := seedPermissions(ctx, defaultDomain); err != nil {
		log.Fatalf("failed to seed permissions: %v", err)
	}

	// 3. 分配权限给角色（通过 Casbin）
	if err := seedRolePermissions(ctx, enforcer, defaultDomain); err != nil {
		log.Fatalf("failed to seed role permissions: %v", err)
	}

	log.Println("RBAC 种子数据初始化完成！")
}

// seedRoles 初始化默认角色
func seedRoles(_ context.Context, domain string) error {
	db := database.GetDB()

	roles := []model.Role{
		{
			Name:        "super_admin",
			DisplayName: "超级管理员",
			Description: "系统最高权限，拥有所有操作权限",
			Domain:      domain,
			Category:    "system",
			IsSystem:    true,
			Sort:        100,
			Status:      1,
		},
		{
			Name:        "admin",
			DisplayName: "管理员",
			Description: "普通管理员，拥有大部分管理权限",
			Domain:      domain,
			Category:    "system",
			IsSystem:    true,
			Sort:        90,
			Status:      1,
		},
		{
			Name:        "editor",
			DisplayName: "编辑者",
			Description: "内容编辑权限",
			Domain:      domain,
			Category:    "business",
			IsSystem:    false,
			Sort:        50,
			Status:      1,
		},
		{
			Name:        "viewer",
			DisplayName: "查看者",
			Description: "只读权限",
			Domain:      domain,
			Category:    "business",
			IsSystem:    false,
			Sort:        10,
			Status:      1,
		},
	}

	for _, role := range roles {
		// 检查角色是否已存在
		var count int64
		if err := db.Model(&model.Role{}).
			Where("name = ? AND domain = ?", role.Name, domain).
			Count(&count).Error; err != nil {
			return fmt.Errorf("check role existence failed: %w", err)
		}

		if count == 0 {
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("create role %s failed: %w", role.Name, err)
			}
			log.Printf("创建角色: %s (%s)", role.Name, role.DisplayName)
		} else {
			log.Printf("角色已存在: %s", role.Name)
		}
	}

	return nil
}

// seedPermissions 初始化默认权限
func seedPermissions(_ context.Context, domain string) error {
	db := database.GetDB()

	permissions := []model.Permission{
		// 用户管理权限
		{
			Name:        "user:list",
			DisplayName: "查看用户列表",
			Description: "查看用户列表权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/users",
			Action:      "read",
			Category:    "用户管理",
			IsSystem:    true,
			Sort:        100,
			Status:      1,
		},
		{
			Name:        "user:create",
			DisplayName: "创建用户",
			Description: "创建新用户权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/users",
			Action:      "write",
			Category:    "用户管理",
			IsSystem:    true,
			Sort:        90,
			Status:      1,
		},
		{
			Name:        "user:update",
			DisplayName: "更新用户",
			Description: "更新用户信息权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/users/*",
			Action:      "write",
			Category:    "用户管理",
			IsSystem:    true,
			Sort:        80,
			Status:      1,
		},
		{
			Name:        "user:delete",
			DisplayName: "删除用户",
			Description: "删除用户权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/users/*",
			Action:      "delete",
			Category:    "用户管理",
			IsSystem:    true,
			Sort:        70,
			Status:      1,
		},

		// 角色管理权限
		{
			Name:        "role:list",
			DisplayName: "查看角色列表",
			Description: "查看角色列表权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/roles",
			Action:      "read",
			Category:    "角色管理",
			IsSystem:    true,
			Sort:        100,
			Status:      1,
		},
		{
			Name:        "role:create",
			DisplayName: "创建角色",
			Description: "创建新角色权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/roles",
			Action:      "write",
			Category:    "角色管理",
			IsSystem:    true,
			Sort:        90,
			Status:      1,
		},
		{
			Name:        "role:update",
			DisplayName: "更新角色",
			Description: "更新角色信息权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/roles/*",
			Action:      "write",
			Category:    "角色管理",
			IsSystem:    true,
			Sort:        80,
			Status:      1,
		},
		{
			Name:        "role:delete",
			DisplayName: "删除角色",
			Description: "删除角色权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/roles/*",
			Action:      "delete",
			Category:    "角色管理",
			IsSystem:    true,
			Sort:        70,
			Status:      1,
		},

		// 权限管理权限
		{
			Name:        "permission:list",
			DisplayName: "查看权限列表",
			Description: "查看权限列表权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/permissions",
			Action:      "read",
			Category:    "权限管理",
			IsSystem:    true,
			Sort:        100,
			Status:      1,
		},
		{
			Name:        "permission:create",
			DisplayName: "创建权限",
			Description: "创建新权限权限",
			Type:        model.PermissionTypeAPI,
			Domain:      domain,
			Resource:    "/api/v1/permissions",
			Action:      "write",
			Category:    "权限管理",
			IsSystem:    true,
			Sort:        90,
			Status:      1,
		},
	}

	for _, perm := range permissions {
		// 检查权限是否已存在
		var count int64
		if err := db.Model(&model.Permission{}).
			Where("name = ? AND domain = ?", perm.Name, domain).
			Count(&count).Error; err != nil {
			return fmt.Errorf("check permission existence failed: %w", err)
		}

		if count == 0 {
			if err := db.Create(&perm).Error; err != nil {
				return fmt.Errorf("create permission %s failed: %w", perm.Name, err)
			}
			log.Printf("创建权限: %s (%s)", perm.Name, perm.DisplayName)
		} else {
			log.Printf("权限已存在: %s", perm.Name)
		}
	}

	return nil
}

// seedRolePermissions 分配权限给角色
func seedRolePermissions(_ context.Context, enforcer *casbin.Enforcer, domain string) error {
	// super_admin 拥有所有权限
	superAdminPolicies := [][]string{
		{"super_admin", domain, "/api/v1/*", "*"},
	}

	// admin 拥有大部分权限
	adminPolicies := [][]string{
		{"admin", domain, "/api/v1/users", "read"},
		{"admin", domain, "/api/v1/users", "write"},
		{"admin", domain, "/api/v1/users/*", "write"},
		{"admin", domain, "/api/v1/roles", "read"},
		{"admin", domain, "/api/v1/roles", "write"},
		{"admin", domain, "/api/v1/permissions", "read"},
	}

	// editor 拥有编辑权限
	editorPolicies := [][]string{
		{"editor", domain, "/api/v1/users", "read"},
		{"editor", domain, "/api/v1/users/*", "write"},
	}

	// viewer 只有查看权限
	viewerPolicies := [][]string{
		{"viewer", domain, "/api/v1/users", "read"},
		{"viewer", domain, "/api/v1/roles", "read"},
		{"viewer", domain, "/api/v1/permissions", "read"},
	}

	allPolicies := append(superAdminPolicies, adminPolicies...)
	allPolicies = append(allPolicies, editorPolicies...)
	allPolicies = append(allPolicies, viewerPolicies...)

	for _, policy := range allPolicies {
		if _, err := enforcer.AddPolicy(policy[0], policy[1], policy[2], policy[3]); err != nil {
			log.Printf("警告: 添加策略失败 %v: %v", policy, err)
		} else {
			log.Printf("添加策略: %s -> %s:%s:%s", policy[0], policy[1], policy[2], policy[3])
		}
	}

	return nil
}
