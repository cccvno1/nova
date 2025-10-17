#!/bin/bash

# ============================================
# Nova 测试数据初始化脚本
# ============================================
# 功能：
# 1. 创建测试用户（admin, editor, viewer）
# 2. 创建角色（super_admin, admin, editor, viewer）
# 3. 创建权限（菜单权限 + API权限）
# 4. 关联角色-权限
# 5. 关联用户-角色
# ============================================

set -e

API_URL="http://localhost:8080/api/v1"
DB_CONTAINER="nova-postgres-test"

echo "🚀 Nova 测试数据初始化开始..."
echo "============================================"
echo ""

# ============================================
# 方法1: 通过 API 创建用户
# ============================================
echo "📝 1. 创建测试用户..."
echo "--------------------------------------------"

# 创建超级管理员
echo "创建超级管理员 (admin)..."
ADMIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123",
    "nickname": "超级管理员"
  }')

ADMIN_TOKEN=$(echo $ADMIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ 超级管理员创建失败"
    echo "响应: $ADMIN_RESPONSE"
    exit 1
fi

echo "✅ 超级管理员创建成功"
echo ""

# 创建编辑者
echo "创建编辑者 (editor)..."
curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "editor",
    "email": "editor@example.com",
    "password": "editor123",
    "nickname": "编辑者"
  }' > /dev/null

echo "✅ 编辑者创建成功"
echo ""

# 创建查看者
echo "创建查看者 (viewer)..."
curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "viewer",
    "email": "viewer@example.com",
    "password": "viewer123",
    "nickname": "查看者"
  }' > /dev/null

echo "✅ 查看者创建成功"
echo ""

# ============================================
# 方法2: 通过 SQL 创建 RBAC 数据
# ============================================
echo "📦 2. 初始化 RBAC 数据（角色、权限、关联）..."
echo "--------------------------------------------"

sudo docker exec -i $DB_CONTAINER psql -U postgres -d nova <<'SQL'

-- ============================================
-- 插入角色
-- ============================================
INSERT INTO roles (name, display_name, description, domain, category, is_system, sort, status, created_at, updated_at)
VALUES 
    ('super_admin', '超级管理员', '拥有所有权限', 'default', 'system', true, 1, 1, NOW(), NOW()),
    ('admin', '管理员', '系统管理员', 'default', 'system', true, 2, 1, NOW(), NOW()),
    ('editor', '编辑者', '可以编辑内容', 'default', 'business', false, 3, 1, NOW(), NOW()),
    ('viewer', '查看者', '只读权限', 'default', 'business', false, 4, 1, NOW(), NOW())
ON CONFLICT (name, domain) DO NOTHING;

-- ============================================
-- 插入菜单权限（前端路由）
-- ============================================

-- 第一步：插入顶级菜单（parent_id 为 NULL）
INSERT INTO permissions (name, display_name, description, type, domain, resource, action, category, parent_id, path, component, icon, is_system, sort, status, created_at, updated_at)
VALUES 
    -- 首页
    ('menu:home', '首页', '系统首页', 'menu', 'default', '/home', 'read', 'system', NULL, '/home', 'views/home/index', 'HomeFilled', true, 1, 1, NOW(), NOW()),
    
    -- 系统管理（父菜单）
    ('menu:system', '系统管理', '系统管理模块', 'menu', 'default', '/system', 'read', 'system', NULL, '/system', 'Layout', 'Setting', true, 10, 1, NOW(), NOW()),
    
    -- 内容管理（父菜单）
    ('menu:content', '内容管理', '内容管理模块', 'menu', 'default', '/content', 'read', 'business', NULL, '/content', 'Layout', 'Document', false, 20, 1, NOW(), NOW())
ON CONFLICT (name, domain) DO NOTHING;

-- 第二步：插入子菜单（使用父菜单的 ID）
INSERT INTO permissions (name, display_name, description, type, domain, resource, action, category, parent_id, path, component, icon, is_system, sort, status, created_at, updated_at)
VALUES 
    ('menu:system:user', '用户管理', '用户管理页面', 'menu', 'default', '/system/user', 'read', 'system', (SELECT id FROM permissions WHERE name = 'menu:system' LIMIT 1), '/system/user', 'views/system/user/index', 'User', true, 1, 1, NOW(), NOW()),
    ('menu:system:role', '角色管理', '角色管理页面', 'menu', 'default', '/system/role', 'read', 'system', (SELECT id FROM permissions WHERE name = 'menu:system' LIMIT 1), '/system/role', 'views/system/role/index', 'Avatar', true, 2, 1, NOW(), NOW()),
    ('menu:system:permission', '权限管理', '权限管理页面', 'menu', 'default', '/system/permission', 'read', 'system', (SELECT id FROM permissions WHERE name = 'menu:system' LIMIT 1), '/system/permission', 'views/system/permission/index', 'Lock', true, 3, 1, NOW(), NOW()),
    ('menu:content:article', '文章管理', '文章管理页面', 'menu', 'default', '/content/article', 'read', 'business', (SELECT id FROM permissions WHERE name = 'menu:content' LIMIT 1), '/content/article', 'views/content/article/index', 'Document', false, 1, 1, NOW(), NOW())
ON CONFLICT (name, domain) DO NOTHING;

-- 第三步：插入 API 权限（不需要 parent_id）
INSERT INTO permissions (name, display_name, description, type, domain, resource, action, category, parent_id, path, component, icon, is_system, sort, status, created_at, updated_at)
VALUES 
    ('api:user:list', '用户列表', '查看用户列表', 'api', 'default', '/api/v1/users', 'GET', 'system', NULL, '', '', '', true, 1, 1, NOW(), NOW()),
    ('api:user:create', '创建用户', '创建新用户', 'api', 'default', '/api/v1/users', 'POST', 'system', NULL, '', '', '', true, 2, 1, NOW(), NOW()),
    ('api:user:update', '更新用户', '更新用户信息', 'api', 'default', '/api/v1/users/:id', 'PUT', 'system', NULL, '', '', '', true, 3, 1, NOW(), NOW()),
    ('api:user:delete', '删除用户', '删除用户', 'api', 'default', '/api/v1/users/:id', 'DELETE', 'system', NULL, '', '', '', true, 4, 1, NOW(), NOW()),
    
    ('api:role:list', '角色列表', '查看角色列表', 'api', 'default', '/api/v1/roles', 'GET', 'system', NULL, '', '', '', true, 1, 1, NOW(), NOW()),
    ('api:role:create', '创建角色', '创建新角色', 'api', 'default', '/api/v1/roles', 'POST', 'system', NULL, '', '', '', true, 2, 1, NOW(), NOW()),
    ('api:role:update', '更新角色', '更新角色信息', 'api', 'default', '/api/v1/roles/:id', 'PUT', 'system', NULL, '', '', '', true, 3, 1, NOW(), NOW()),
    ('api:role:delete', '删除角色', '删除角色', 'api', 'default', '/api/v1/roles/:id', 'DELETE', 'system', NULL, '', '', '', true, 4, 1, NOW(), NOW())
ON CONFLICT (name, domain) DO NOTHING;

-- ============================================
-- 角色-权限关联
-- ============================================

-- 超级管理员：所有权限
INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
SELECT 
    (SELECT id FROM roles WHERE name = 'super_admin' LIMIT 1),
    p.id,
    NOW(),
    NOW()
FROM permissions p
ON CONFLICT DO NOTHING;

-- 管理员：系统管理权限
INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
SELECT 
    (SELECT id FROM roles WHERE name = 'admin' LIMIT 1),
    p.id,
    NOW(),
    NOW()
FROM permissions p
WHERE p.name LIKE 'menu:home' 
   OR p.name LIKE 'menu:system%'
   OR (p.name LIKE 'api:%' AND p.name NOT LIKE '%delete')
ON CONFLICT DO NOTHING;

-- 编辑者：内容管理 + 部分API
INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
SELECT 
    (SELECT id FROM roles WHERE name = 'editor' LIMIT 1),
    p.id,
    NOW(),
    NOW()
FROM permissions p
WHERE p.name LIKE 'menu:home' 
   OR p.name LIKE 'menu:content%'
   OR (p.type = 'api' AND p.action IN ('GET', 'POST', 'PUT'))
ON CONFLICT DO NOTHING;

-- 查看者：只有查看权限
INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
SELECT 
    (SELECT id FROM roles WHERE name = 'viewer' LIMIT 1),
    p.id,
    NOW(),
    NOW()
FROM permissions p
WHERE p.type = 'menu' 
   OR (p.type = 'api' AND p.action = 'GET')
ON CONFLICT DO NOTHING;

-- ============================================
-- 用户-角色关联
-- ============================================
INSERT INTO user_roles (user_id, role_id, domain, assigned_by, created_at, updated_at)
VALUES 
    ((SELECT id FROM users WHERE username = 'admin' LIMIT 1), (SELECT id FROM roles WHERE name = 'super_admin' LIMIT 1), 'default', 0, NOW(), NOW()),
    ((SELECT id FROM users WHERE username = 'editor' LIMIT 1), (SELECT id FROM roles WHERE name = 'editor' LIMIT 1), 'default', 0, NOW(), NOW()),
    ((SELECT id FROM users WHERE username = 'viewer' LIMIT 1), (SELECT id FROM roles WHERE name = 'viewer' LIMIT 1), 'default', 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

SQL

if [ $? -eq 0 ]; then
    echo "✅ RBAC 数据初始化成功"
else
    echo "❌ RBAC 数据初始化失败"
    exit 1
fi

echo ""

# ============================================
# 方案A: 不再需要手动同步 Casbin 策略表
# ============================================
echo "✅ RBAC数据初始化完成！（方案A：Casbin将从RBAC表自动加载策略）"
echo "============================================"
echo ""
echo "📋 测试账号信息："
echo "--------------------------------------------"
echo "超级管理员:"
echo "  用户名: admin"
echo "  密码:   admin123"
echo "  权限:   所有权限"
echo ""
echo "编辑者:"
echo "  用户名: editor"
echo "  密码:   editor123"
echo "  权限:   内容管理 + 部分API"
echo ""
echo "查看者:"
echo "  用户名: viewer"
echo "  密码:   viewer123"
echo "  权限:   只读权限"
echo "--------------------------------------------"
echo ""
echo "🌐 访问前端: http://localhost:5173"
echo "📚 访问文档: http://localhost:8080/swagger/index.html"
echo ""
