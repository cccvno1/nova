package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/logger"
)

var menuConfigFile = flag.String("menu-config", "../../configs/config.yaml", "config file path")

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*menuConfigFile)
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

	ctx := context.Background()

	log.Println("开始初始化菜单权限数据...")

	// 初始化默认域
	defaultDomain := "default"

	// 创建菜单权限
	if err := seedMenuPermissions(ctx, defaultDomain); err != nil {
		log.Fatalf("failed to seed menu permissions: %v", err)
	}

	// 分配菜单权限给admin角色
	if err := assignMenuPermissionsToAdmin(ctx, defaultDomain); err != nil {
		log.Fatalf("failed to assign menu permissions to admin: %v", err)
	}

	log.Println("菜单权限数据初始化完成！")
}

// seedMenuPermissions 初始化菜单权限
func seedMenuPermissions(_ context.Context, domain string) error {
	db := database.GetDB()

	// 定义菜单权限（树形结构）
	permissions := []model.Permission{
		// 1. 首页（顶级菜单）
		{
			Name:        "menu:home",
			DisplayName: "首页",
			Description: "系统首页",
			Type:        model.PermissionTypeMenu,
			Domain:      domain,
			Resource:    "/home",
			Action:      "view",
			ParentID:    0,
			Path:        "/home",
			Component:   "views/home/index",
			Icon:        "HomeFilled",
			IsSystem:    true,
			Sort:        100,
			Status:      1,
		},
		// 2. 系统管理（顶级菜单）
		{
			Name:        "menu:system",
			DisplayName: "系统管理",
			Description: "系统管理模块",
			Type:        model.PermissionTypeMenu,
			Domain:      domain,
			Resource:    "/system",
			Action:      "view",
			ParentID:    0,
			Path:        "/system",
			Component:   "Layout",
			Icon:        "Setting",
			IsSystem:    true,
			Sort:        90,
			Status:      1,
		},
	}

	// 先创建顶级菜单
	for _, perm := range permissions {
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
			log.Printf("创建菜单权限: %s (%s)", perm.Name, perm.DisplayName)
		} else {
			log.Printf("菜单权限已存在: %s", perm.Name)
		}
	}

	// 获取系统管理的ID作为父级ID
	var systemMenu model.Permission
	if err := db.Where("name = ? AND domain = ?", "menu:system", domain).First(&systemMenu).Error; err != nil {
		return fmt.Errorf("find system menu failed: %w", err)
	}

	// 创建系统管理的子菜单
	subMenus := []model.Permission{
		// 2.1 用户管理
		{
			Name:        "menu:system:user",
			DisplayName: "用户管理",
			Description: "用户管理页面",
			Type:        model.PermissionTypeMenu,
			Domain:      domain,
			Resource:    "/system/user",
			Action:      "view",
			ParentID:    systemMenu.ID,
			Path:        "/system/user",
			Component:   "views/system/user/index",
			Icon:        "User",
			IsSystem:    true,
			Sort:        100,
			Status:      1,
		},
		// 2.2 角色管理
		{
			Name:        "menu:system:role",
			DisplayName: "角色管理",
			Description: "角色管理页面",
			Type:        model.PermissionTypeMenu,
			Domain:      domain,
			Resource:    "/system/role",
			Action:      "view",
			ParentID:    systemMenu.ID,
			Path:        "/system/role",
			Component:   "views/system/role/index",
			Icon:        "UserFilled",
			IsSystem:    true,
			Sort:        90,
			Status:      1,
		},
		// 2.3 权限管理
		{
			Name:        "menu:system:permission",
			DisplayName: "权限管理",
			Description: "权限管理页面",
			Type:        model.PermissionTypeMenu,
			Domain:      domain,
			Resource:    "/system/permission",
			Action:      "view",
			ParentID:    systemMenu.ID,
			Path:        "/system/permission",
			Component:   "views/system/permission/index",
			Icon:        "Lock",
			IsSystem:    true,
			Sort:        80,
			Status:      1,
		},
	}

	for _, perm := range subMenus {
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
			log.Printf("创建子菜单权限: %s (%s)", perm.Name, perm.DisplayName)
		} else {
			log.Printf("子菜单权限已存在: %s", perm.Name)
		}
	}

	return nil
}

// assignMenuPermissionsToAdmin 分配菜单权限给admin用户
func assignMenuPermissionsToAdmin(_ context.Context, domain string) error {
	db := database.GetDB()

	// 获取admin用户
	var adminUser model.User
	if err := db.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		return fmt.Errorf("find admin user failed: %w", err)
	}

	// 获取或创建admin角色
	var adminRole model.Role
	err := db.Where("name = ? AND domain = ?", "admin", domain).First(&adminRole).Error
	if err != nil {
		// 如果admin角色不存在，创建它
		adminRole = model.Role{
			Name:        "admin",
			DisplayName: "管理员",
			Description: "系统管理员，拥有所有菜单权限",
			Domain:      domain,
			Category:    "system",
			IsSystem:    true,
			Sort:        90,
			Status:      1,
		}
		if err := db.Create(&adminRole).Error; err != nil {
			return fmt.Errorf("create admin role failed: %w", err)
		}
		log.Printf("创建admin角色")
	}

	// 检查用户是否已有admin角色
	var userRoleCount int64
	if err := db.Model(&model.UserRole{}).
		Where("user_id = ? AND role_id = ? AND domain = ?", adminUser.ID, adminRole.ID, domain).
		Count(&userRoleCount).Error; err != nil {
		return fmt.Errorf("check user role failed: %w", err)
	}

	if userRoleCount == 0 {
		// 分配admin角色给admin用户
		userRole := model.UserRole{
			UserID: adminUser.ID,
			RoleID: adminRole.ID,
			Domain: domain,
		}
		if err := db.Create(&userRole).Error; err != nil {
			return fmt.Errorf("assign admin role to admin user failed: %w", err)
		}
		log.Printf("分配admin角色给admin用户")
	} else {
		log.Printf("admin用户已有admin角色")
	}

	// 获取所有菜单权限
	var menuPermissions []model.Permission
	if err := db.Where("type = ? AND domain = ?", model.PermissionTypeMenu, domain).Find(&menuPermissions).Error; err != nil {
		return fmt.Errorf("find menu permissions failed: %w", err)
	}

	// 分配所有菜单权限给admin角色
	for _, perm := range menuPermissions {
		var count int64
		if err := db.Model(&model.RolePermission{}).
			Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("check role permission failed: %w", err)
		}

		if count == 0 {
			rolePermission := model.RolePermission{
				RoleID:       adminRole.ID,
				PermissionID: perm.ID,
			}
			if err := db.Create(&rolePermission).Error; err != nil {
				return fmt.Errorf("assign permission %s to admin role failed: %w", perm.Name, err)
			}
			log.Printf("分配菜单权限 %s 给admin角色", perm.Name)
		}
	}

	log.Println("已将所有菜单权限分配给admin角色")
	return nil
}
