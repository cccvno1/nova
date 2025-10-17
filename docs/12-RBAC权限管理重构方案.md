# RBAC 权限管理系统重构方案

## 📚 业界开源方案调研

### 1. Casbin (Go) - 权限引擎
**GitHub:** https://github.com/casbin/casbin (15k+ stars)

**核心理念：**
- 基于PERM模型（Policy, Effect, Request, Matchers）
- 支持RBAC、ABAC、RESTful等多种模型
- 策略存储与执行分离

**优点：**
- ✅ 高性能（纯内存匹配）
- ✅ 模型灵活（支持自定义规则）
- ✅ 社区活跃（多语言SDK）

**缺点：**
- ❌ 学习曲线陡峭
- ❌ UI管理复杂（需要自己实现）
- ❌ 策略语法不直观

---

### 2. Django Admin (Python) - 后台管理
**GitHub:** https://github.com/django/django (75k+ stars)

**核心理念：**
- Permission = Model + Action（如：blog.add_post）
- 基于装饰器的权限检查
- 自动生成CRUD权限

**优点：**
- ✅ 开箱即用（零配置权限管理）
- ✅ 用户体验好（管理后台完善）
- ✅ 自动化程度高

**缺点：**
- ❌ 仅限Python生态
- ❌ 定制化困难
- ❌ 不适合微服务架构

---

### 3. Spring Security (Java) - 企业级安全框架
**GitHub:** https://github.com/spring-projects/spring-security (8k+ stars)

**核心理念：**
- 注解式权限控制（@PreAuthorize）
- 表达式语言（SpEL）
- Filter Chain安全过滤

**优点：**
- ✅ 企业级成熟度
- ✅ 集成OAuth2/SAML
- ✅ 细粒度控制

**缺点：**
- ❌ 配置复杂
- ❌ Java特有
- ❌ 性能开销大

---

### 4. vue-element-admin (Vue) - 前端权限最佳实践
**GitHub:** https://github.com/PanJiaChen/vue-element-admin (85k+ stars)

**核心理念：**
```typescript
// 1. 路由级权限
{
  path: '/permission',
  meta: { roles: ['admin', 'editor'] }
}

// 2. 按钮级权限
<el-button v-permission="['admin']">删除</el-button>

// 3. 指令式权限
v-if="checkPermission(['admin'])"
```

**优点：**
- ✅ 最佳实践（大量企业采用）
- ✅ 开箱即用（完整示例）
- ✅ 动态路由（基于权限生成菜单）

**缺点：**
- ❌ 仅前端方案（需配合后端）
- ❌ 权限粒度固定

---

### 5. Keycloak (Java) - 身份认证与授权
**GitHub:** https://github.com/keycloak/keycloak (20k+ stars)

**核心理念：**
- 独立的认证授权服务
- 支持SSO、OIDC、SAML
- 细粒度资源权限

**优点：**
- ✅ 企业级标准
- ✅ 完整的IAM解决方案
- ✅ 支持多租户

**缺点：**
- ❌ 过于重量级
- ❌ 学习成本高
- ❌ 需要独立部署

---

### 6. Ant Design Pro (React) - 企业级权限方案
**GitHub:** https://github.com/ant-design/ant-design-pro (35k+ stars)

**核心理念：**
```typescript
// access.ts - 权限定义
export default function access(initialState) {
  const { currentUser } = initialState;
  return {
    canAdmin: currentUser?.role === 'admin',
    canEditPost: (post) => post.author === currentUser?.id,
  };
}

// 使用
<Access accessible={access.canAdmin}>
  <Button>删除</Button>
</Access>
```

**优点：**
- ✅ 函数式权限判断（灵活）
- ✅ 与业务逻辑解耦
- ✅ 支持动态权限

**缺点：**
- ❌ React生态限定

---

## 🎯 最佳实践总结

### 权限模型对比

| 模型 | 适用场景 | 复杂度 | 灵活性 | 代表项目 |
|------|----------|--------|--------|----------|
| **RBAC** | 企业内部系统 | ⭐⭐ | ⭐⭐⭐ | Django |
| **ABAC** | 复杂业务规则 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Casbin |
| **ACL** | 简单权限控制 | ⭐ | ⭐⭐ | WordPress |
| **ReBAC** | 社交网络 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Zanzibar |

---

## 🚀 Nova 权限系统重构方案

### 设计原则

1. **简单优于复杂**：80%的场景用RBAC就够了
2. **渐进式增强**：基础功能稳定，高级功能可选
3. **前后端分离**：权限判断逻辑在后端，UI在前端
4. **可审计性**：所有权限变更都有日志

---

### 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                     前端应用层                            │
├─────────────────────────────────────────────────────────┤
│ ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│ │路由权限  │  │按钮权限   │  │数据权限   │  │字段权限   │ │
│ └─────────┘  └──────────┘  └──────────┘  └──────────┘ │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                     API网关层                             │
├─────────────────────────────────────────────────────────┤
│ ┌─────────┐  ┌──────────┐  ┌──────────┐               │
│ │JWT认证   │  │权限拦截   │  │审计日志   │               │
│ └─────────┘  └──────────┘  └──────────┘               │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                   权限服务层                              │
├─────────────────────────────────────────────────────────┤
│ ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│ │权限引擎       │  │权限管理       │  │权限查询       │  │
│ │(Casbin)      │  │(CRUD)        │  │(Cache)       │  │
│ └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                   数据存储层                              │
├─────────────────────────────────────────────────────────┤
│ ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│ │PostgreSQL │  │Redis     │  │审计日志   │              │
│ │(权限数据)  │  │(缓存)     │  │(ES/PG)   │              │
│ └──────────┘  └──────────┘  └──────────┘              │
└─────────────────────────────────────────────────────────┘
```

---

### 数据模型设计（参考Django + Casbin）

#### 1. 核心表结构

```sql
-- 用户表（已存在）
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255),
    -- ...
);

-- 角色表（改进版）
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,        -- 角色代码（如：admin）
    name VARCHAR(100) NOT NULL,              -- 角色名称（如：系统管理员）
    description TEXT,
    category VARCHAR(50),                     -- 分类（system/business）
    is_system BOOLEAN DEFAULT false,         -- 系统角色（不可删除）
    status SMALLINT DEFAULT 1,
    sort INTEGER DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 权限表（扁平化设计，参考Django）
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,       -- 权限代码（blog.add_post）
    name VARCHAR(100) NOT NULL,              -- 权限名称（创建文章）
    type VARCHAR(20) NOT NULL,               -- 类型：menu/api/button/data/field
    resource VARCHAR(100),                    -- 资源（API路径、菜单路径）
    action VARCHAR(50),                      -- 动作（read/write/delete/execute）
    description TEXT,
    category VARCHAR(50),                     -- 分类
    
    -- 仅用于菜单类型
    parent_id INTEGER REFERENCES permissions(id),
    path VARCHAR(255),
    component VARCHAR(255),
    icon VARCHAR(50),
    
    -- 元数据
    metadata JSONB,                          -- 扩展字段（条件规则、数据过滤等）
    is_system BOOLEAN DEFAULT false,
    status SMALLINT DEFAULT 1,
    sort INTEGER DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 角色权限关联表（保留，但改进）
CREATE TABLE role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    
    -- 新增：权限约束条件（ABAC扩展）
    conditions JSONB,                        -- 条件规则（如：{"department": "IT"}）
    
    granted_by INTEGER REFERENCES users(id), -- 授权人
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,                    -- 过期时间（临时权限）
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(role_id, permission_id)
);

-- 用户角色关联表（改进）
CREATE TABLE user_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    
    -- 新增：角色范围限定
    scope VARCHAR(50),                       -- 范围（如：department/project）
    scope_value VARCHAR(100),                -- 范围值（如：IT部门/项目A）
    
    assigned_by INTEGER REFERENCES users(id),
    assigned_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,                    -- 临时角色
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, role_id, scope, scope_value)
);

-- 权限变更历史表（审计）
CREATE TABLE permission_changes (
    id SERIAL PRIMARY KEY,
    change_type VARCHAR(20) NOT NULL,        -- assign/revoke/modify
    target_type VARCHAR(20) NOT NULL,        -- role/user
    target_id INTEGER NOT NULL,
    permission_ids INTEGER[],                -- 涉及的权限ID列表
    old_value JSONB,                         -- 变更前
    new_value JSONB,                         -- 变更后
    reason TEXT,                             -- 变更原因
    operator_id INTEGER REFERENCES users(id),
    operator_ip VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 权限模板表（可选）
CREATE TABLE permission_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permission_ids INTEGER[],                -- 权限ID列表
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### 2. 索引优化

```sql
-- 高频查询索引
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_perm ON role_permissions(permission_id);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_permissions_code ON permissions(code);
CREATE INDEX idx_permissions_type ON permissions(type);
CREATE INDEX idx_permissions_parent ON permissions(parent_id);

-- GIN索引（用于JSONB查询）
CREATE INDEX idx_role_permissions_conditions ON role_permissions USING GIN(conditions);
CREATE INDEX idx_permissions_metadata ON permissions USING GIN(metadata);
```

---

### API设计（参考RESTful + Django）

#### 1. 权限操作API（改进版）

```go
// =============== 角色权限管理 ===============

// 1. 获取角色的所有权限（只读）
GET /api/v1/roles/:id/permissions
Response: {
  "code": 0,
  "data": {
    "role": { "id": 1, "code": "admin", "name": "管理员" },
    "permissions": [
      {
        "id": 1,
        "code": "menu.home",
        "name": "首页",
        "type": "menu",
        "granted_at": "2025-10-17T10:00:00Z"
      }
    ],
    "total": 15
  }
}

// 2. 批量添加权限（增量）
POST /api/v1/roles/:id/permissions/batch-add
Request: {
  "permission_ids": [4, 5, 6],
  "reason": "添加用户管理权限"
}
Response: {
  "code": 0,
  "message": "成功添加3个权限",
  "data": {
    "added_count": 3,
    "added_permissions": [
      { "id": 4, "name": "用户管理" }
    ]
  }
}

// 3. 批量移除权限（增量）
POST /api/v1/roles/:id/permissions/batch-remove
Request: {
  "permission_ids": [7, 8],
  "reason": "移除临时权限"
}
Response: {
  "code": 0,
  "message": "成功移除2个权限",
  "data": {
    "removed_count": 2,
    "removed_permissions": [
      { "id": 7, "name": "系统配置" }
    ]
  }
}

// 4. 同步权限（全量替换，高危操作）
POST /api/v1/roles/:id/permissions/sync
Request: {
  "permission_ids": [1, 2, 3, 4, 5],
  "reason": "重置为默认权限",
  "confirm": true  // 必须明确确认
}
Response: {
  "code": 0,
  "message": "权限同步成功",
  "data": {
    "previous_count": 15,
    "current_count": 5,
    "added_count": 0,
    "removed_count": 10,
    "changes": {
      "added": [],
      "removed": [
        { "id": 6, "name": "权限管理" },
        // ...
      ]
    }
  }
}

// 5. 从模板应用权限（推荐）
POST /api/v1/roles/:id/permissions/apply-template
Request: {
  "template_id": 3,
  "mode": "merge"  // merge/replace
}

// 6. 对比权限差异（预览变更）
POST /api/v1/roles/:id/permissions/diff
Request: {
  "permission_ids": [1, 2, 3]
}
Response: {
  "code": 0,
  "data": {
    "will_add": [
      { "id": 2, "name": "角色管理" }
    ],
    "will_remove": [
      { "id": 5, "name": "权限管理" }
    ],
    "will_keep": [
      { "id": 1, "name": "首页" }
    ]
  }
}

// =============== 用户角色管理 ===============

// 7. 获取用户的所有角色
GET /api/v1/users/:id/roles

// 8. 给用户分配角色（增量）
POST /api/v1/users/:id/roles/assign
Request: {
  "role_ids": [2, 3],
  "scope": "department",       // 可选：角色范围
  "scope_value": "IT",
  "expires_at": "2025-12-31"   // 可选：临时角色
}

// 9. 移除用户角色
POST /api/v1/users/:id/roles/revoke
Request: {
  "role_ids": [3]
}

// =============== 权限查询 ===============

// 10. 获取用户的所有权限（合并角色权限）
GET /api/v1/users/:id/permissions
Response: {
  "code": 0,
  "data": {
    "user": { "id": 1, "username": "admin" },
    "roles": [
      { "id": 1, "code": "admin", "name": "管理员" }
    ],
    "permissions": [
      {
        "id": 1,
        "code": "menu.home",
        "type": "menu",
        "source": "role:admin"  // 权限来源
      }
    ],
    "grouped": {
      "menu": [...],
      "api": [...],
      "button": [...]
    }
  }
}

// 11. 权限检查（单个）
POST /api/v1/permissions/check
Request: {
  "user_id": 1,
  "permission": "user.delete",
  "context": {                 // 可选：上下文（ABAC）
    "resource_owner": 2,
    "department": "IT"
  }
}
Response: {
  "code": 0,
  "data": {
    "allowed": true,
    "reason": "role:admin grants user.delete"
  }
}

// 12. 批量权限检查
POST /api/v1/permissions/check-batch
Request: {
  "user_id": 1,
  "permissions": ["user.read", "user.write", "user.delete"]
}
Response: {
  "code": 0,
  "data": {
    "user.read": true,
    "user.write": true,
    "user.delete": false
  }
}

// =============== 权限管理（CRUD） ===============

// 13. 权限列表（支持分组）
GET /api/v1/permissions?type=menu&group_by=category

// 14. 权限树（菜单类型）
GET /api/v1/permissions/tree

// 15. 创建权限
POST /api/v1/permissions
Request: {
  "code": "blog.add_post",     // Django风格
  "name": "创建文章",
  "type": "api",
  "resource": "/api/v1/posts",
  "action": "create",
  "description": "允许创建新文章"
}

// 16. 批量创建权限（自动生成CRUD）
POST /api/v1/permissions/auto-generate
Request: {
  "resource": "article",
  "actions": ["read", "create", "update", "delete"]
}
Response: {
  "code": 0,
  "data": {
    "created": [
      { "code": "article.read", "name": "查看文章" },
      { "code": "article.create", "name": "创建文章" },
      { "code": "article.update", "name": "更新文章" },
      { "code": "article.delete", "name": "删除文章" }
    ]
  }
}

// =============== 审计日志 ===============

// 17. 权限变更历史
GET /api/v1/permission-changes?target_type=role&target_id=1

// 18. 用户操作历史
GET /api/v1/users/:id/permission-history
```

---

### 后端实现方案

#### 1. Service层架构（分层设计）

```go
// =============== 权限服务接口 ===============

type PermissionService interface {
    // CRUD
    CreatePermission(ctx context.Context, perm *Permission) error
    UpdatePermission(ctx context.Context, perm *Permission) error
    DeletePermission(ctx context.Context, id uint) error
    GetPermission(ctx context.Context, id uint) (*Permission, error)
    ListPermissions(ctx context.Context, filter PermissionFilter) ([]Permission, error)
    GetPermissionTree(ctx context.Context, types []string) ([]Permission, error)
    
    // 自动生成
    AutoGeneratePermissions(ctx context.Context, resource string, actions []string) error
}

type RolePermissionService interface {
    // 查询
    GetRolePermissions(ctx context.Context, roleID uint) ([]Permission, error)
    
    // 增量操作（推荐）
    AddPermissionsToRole(ctx context.Context, req AddPermissionsRequest) (*PermissionChangeResult, error)
    RemovePermissionsFromRole(ctx context.Context, req RemovePermissionsRequest) (*PermissionChangeResult, error)
    
    // 全量操作（高危）
    SyncRolePermissions(ctx context.Context, req SyncPermissionsRequest) (*PermissionChangeResult, error)
    
    // 模板
    ApplyTemplate(ctx context.Context, roleID uint, templateID uint, mode string) error
    
    // 预览
    PreviewChanges(ctx context.Context, roleID uint, permissionIDs []uint) (*PermissionDiff, error)
}

type UserRoleService interface {
    GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
    AssignRolesToUser(ctx context.Context, req AssignRolesRequest) error
    RevokeRolesFromUser(ctx context.Context, req RevokeRolesRequest) error
}

type PermissionCheckService interface {
    // 单个检查
    CheckPermission(ctx context.Context, req CheckPermissionRequest) (*CheckResult, error)
    
    // 批量检查
    CheckPermissions(ctx context.Context, userID uint, permissions []string) (map[string]bool, error)
    
    // 获取用户所有权限（用于前端）
    GetUserAllPermissions(ctx context.Context, userID uint) (*UserPermissions, error)
}

type PermissionAuditService interface {
    // 记录变更
    LogPermissionChange(ctx context.Context, change *PermissionChange) error
    
    // 查询历史
    GetChangeHistory(ctx context.Context, filter ChangeFilter) ([]PermissionChange, error)
}
```

#### 2. 核心实现（增量操作）

```go
// =============== 增量添加权限 ===============

type AddPermissionsRequest struct {
    RoleID        uint
    PermissionIDs []uint
    Reason        string
    OperatorID    uint
}

type PermissionChangeResult struct {
    AddedCount      int
    AddedPermissions []Permission
    FailedIDs       []uint
    FailedReasons   map[uint]string
}

func (s *rolePermissionService) AddPermissionsToRole(
    ctx context.Context, 
    req AddPermissionsRequest,
) (*PermissionChangeResult, error) {
    // 1. 验证角色
    role, err := s.roleRepo.FindByID(ctx, req.RoleID)
    if err != nil {
        return nil, fmt.Errorf("role not found: %w", err)
    }
    
    // 2. 获取现有权限（用于去重）
    existingPerms, err := s.GetRolePermissions(ctx, req.RoleID)
    if err != nil {
        return nil, err
    }
    
    existingIDs := make(map[uint]bool)
    for _, p := range existingPerms {
        existingIDs[p.ID] = true
    }
    
    // 3. 过滤已存在的权限
    var toAdd []uint
    var skipped []uint
    for _, id := range req.PermissionIDs {
        if existingIDs[id] {
            skipped = append(skipped, id)
        } else {
            toAdd = append(toAdd, id)
        }
    }
    
    if len(toAdd) == 0 {
        return &PermissionChangeResult{
            AddedCount: 0,
            FailedReasons: map[uint]string{},
        }, nil
    }
    
    // 4. 验证权限是否存在
    permissions, err := s.permRepo.ListByIDs(ctx, toAdd)
    if err != nil {
        return nil, err
    }
    
    if len(permissions) != len(toAdd) {
        // 找出不存在的权限ID
        foundIDs := make(map[uint]bool)
        for _, p := range permissions {
            foundIDs[p.ID] = true
        }
        
        var notFound []uint
        for _, id := range toAdd {
            if !foundIDs[id] {
                notFound = append(notFound, id)
            }
        }
        
        return nil, fmt.Errorf("permissions not found: %v", notFound)
    }
    
    // 5. 使用Append添加（不影响现有权限）
    err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 添加权限
        if err := tx.Model(role).Association("Permissions").Append(permissions); err != nil {
            return err
        }
        
        // 6. 记录审计日志
        change := &PermissionChange{
            ChangeType:    "add_permissions",
            TargetType:    "role",
            TargetID:      req.RoleID,
            PermissionIDs: toAdd,
            OldValue:      existingIDs,
            NewValue:      append(existingIDs, toAdd...),
            Reason:        req.Reason,
            OperatorID:    req.OperatorID,
            CreatedAt:     time.Now(),
        }
        
        return tx.Create(change).Error
    })
    
    if err != nil {
        return nil, err
    }
    
    // 7. 清理缓存
    s.clearRolePermissionsCache(ctx, req.RoleID)
    s.clearUserPermissionsCacheByRole(ctx, req.RoleID)
    
    // 8. 返回结果
    result := &PermissionChangeResult{
        AddedCount:       len(permissions),
        AddedPermissions: permissions,
    }
    
    s.logger.Info("permissions added to role",
        "role_id", req.RoleID,
        "role_name", role.Name,
        "added_count", len(permissions),
        "skipped_count", len(skipped),
        "operator_id", req.OperatorID,
    )
    
    return result, nil
}

// =============== 增量移除权限 ===============

func (s *rolePermissionService) RemovePermissionsFromRole(
    ctx context.Context,
    req RemovePermissionsRequest,
) (*PermissionChangeResult, error) {
    // 实现类似，使用Association.Delete
    // ...
}

// =============== 预览变更（在提交前） ===============

type PermissionDiff struct {
    WillAdd    []Permission
    WillRemove []Permission
    WillKeep   []Permission
}

func (s *rolePermissionService) PreviewChanges(
    ctx context.Context,
    roleID uint,
    newPermissionIDs []uint,
) (*PermissionDiff, error) {
    // 1. 获取当前权限
    currentPerms, err := s.GetRolePermissions(ctx, roleID)
    if err != nil {
        return nil, err
    }
    
    currentIDs := make(map[uint]bool)
    for _, p := range currentPerms {
        currentIDs[p.ID] = true
    }
    
    newIDs := make(map[uint]bool)
    for _, id := range newPermissionIDs {
        newIDs[id] = true
    }
    
    // 2. 计算差异
    var willAdd, willRemove, willKeep []Permission
    
    // 找出要删除的
    for _, p := range currentPerms {
        if newIDs[p.ID] {
            willKeep = append(willKeep, p)
        } else {
            willRemove = append(willRemove, p)
        }
    }
    
    // 找出要添加的
    var toAddIDs []uint
    for id := range newIDs {
        if !currentIDs[id] {
            toAddIDs = append(toAddIDs, id)
        }
    }
    
    if len(toAddIDs) > 0 {
        willAdd, err = s.permRepo.ListByIDs(ctx, toAddIDs)
        if err != nil {
            return nil, err
        }
    }
    
    return &PermissionDiff{
        WillAdd:    willAdd,
        WillRemove: willRemove,
        WillKeep:   willKeep,
    }, nil
}
```

---

### 前端实现方案（参考vue-element-admin）

#### 1. 权限配置页面改进

```vue
<template>
  <el-dialog 
    v-model="visible" 
    title="权限配置" 
    width="800px"
  >
    <!-- 操作模式选择 -->
    <el-alert 
      :title="modeDescription" 
      :type="mode === 'incremental' ? 'info' : 'warning'"
      style="margin-bottom: 20px"
    />
    
    <el-radio-group v-model="mode" @change="handleModeChange">
      <el-radio-button label="incremental">
        <el-icon><Plus /></el-icon>
        增量修改（推荐）
      </el-radio-button>
      <el-radio-button label="sync">
        <el-icon><Warning /></el-icon>
        全量同步（危险）
      </el-radio-button>
    </el-radio-group>
    
    <!-- 权限树 -->
    <el-tree
      ref="treeRef"
      :data="permissionTree"
      :props="{ label: 'display_name', children: 'children' }"
      show-checkbox
      node-key="id"
      :default-checked-keys="checkedKeys"
      @check="handleCheck"
    />
    
    <!-- 变更预览 -->
    <el-collapse v-if="diff" style="margin-top: 20px">
      <el-collapse-item title="变更预览" name="1">
        <div v-if="diff.will_add.length > 0">
          <el-tag type="success">新增 {{ diff.will_add.length }} 个</el-tag>
          <ul>
            <li v-for="p in diff.will_add" :key="p.id">
              {{ p.display_name }}
            </li>
          </ul>
        </div>
        
        <div v-if="diff.will_remove.length > 0">
          <el-tag type="danger">删除 {{ diff.will_remove.length }} 个</el-tag>
          <ul>
            <li v-for="p in diff.will_remove" :key="p.id">
              {{ p.display_name }}
            </li>
          </ul>
        </div>
        
        <div v-if="diff.will_keep.length > 0">
          <el-tag type="info">保留 {{ diff.will_keep.length }} 个</el-tag>
        </div>
      </el-collapse-item>
    </el-collapse>
    
    <!-- 操作原因（全量同步时必填） -->
    <el-input
      v-if="mode === 'sync'"
      v-model="reason"
      type="textarea"
      placeholder="请说明全量同步的原因（必填）"
      :rows="3"
      style="margin-top: 20px"
    />
    
    <template #footer>
      <el-button @click="handlePreview">
        <el-icon><View /></el-icon>
        预览变更
      </el-button>
      <el-button @click="visible = false">取消</el-button>
      <el-button 
        type="primary" 
        :loading="loading"
        @click="handleSubmit"
      >
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { roleApi } from '@/api/role'

const visible = ref(false)
const mode = ref<'incremental' | 'sync'>('incremental')
const permissionTree = ref([])
const checkedKeys = ref<number[]>([])
const originalKeys = ref<number[]>([])
const diff = ref(null)
const reason = ref('')
const loading = ref(false)

const modeDescription = computed(() => {
  return mode.value === 'incremental'
    ? '📝 增量模式：只会添加或删除您选择的权限，不影响其他权限'
    : '⚠️ 全量模式：将用当前选择的权限完全替换原有权限，请谨慎操作！'
})

// 预览变更
const handlePreview = async () => {
  const treeRef = treeRef.value
  const checkedKeys = treeRef.getCheckedKeys()
  const halfCheckedKeys = treeRef.getHalfCheckedKeys()
  const allKeys = [...checkedKeys, ...halfCheckedKeys]
  
  diff.value = await roleApi.previewPermissionChanges(roleId, allKeys)
}

// 提交变更
const handleSubmit = async () => {
  const treeRef = treeRef.value
  const checkedKeys = treeRef.getCheckedKeys()
  const halfCheckedKeys = treeRef.getHalfCheckedKeys()
  const allKeys = [...checkedKeys, ...halfCheckedKeys]
  
  if (mode.value === 'incremental') {
    // 增量模式：计算要添加和删除的
    const toAdd = allKeys.filter(id => !originalKeys.value.includes(id))
    const toRemove = originalKeys.value.filter(id => !allKeys.includes(id))
    
    // 二次确认
    const message = [
      toAdd.length > 0 ? `✅ 将添加 ${toAdd.length} 个权限` : '',
      toRemove.length > 0 ? `⚠️ 将删除 ${toRemove.length} 个权限` : '',
      toAdd.length === 0 && toRemove.length === 0 ? '没有任何变更' : ''
    ].filter(Boolean).join('\n')
    
    await ElMessageBox.confirm(message + '\n\n确定继续吗？', '确认变更', {
      type: 'info'
    })
    
    // 分别调用添加和删除接口
    if (toAdd.length > 0) {
      await roleApi.addPermissions(roleId, {
        permission_ids: toAdd,
        reason: reason.value
      })
    }
    
    if (toRemove.length > 0) {
      await roleApi.removePermissions(roleId, {
        permission_ids: toRemove,
        reason: reason.value
      })
    }
    
    ElMessage.success(`权限更新成功：添加${toAdd.length}个，删除${toRemove.length}个`)
    
  } else {
    // 全量模式：需要强确认
    if (!reason.value) {
      ElMessage.error('全量同步必须填写原因')
      return
    }
    
    if (!diff.value) {
      await handlePreview()
    }
    
    const message = `
      ⚠️ 危险操作确认 ⚠️
      
      您即将执行全量权限同步，这将：
      - 删除 ${diff.value.will_remove.length} 个现有权限
      - 添加 ${diff.value.will_add.length} 个新权限
      - 保留 ${diff.value.will_keep.length} 个权限
      
      此操作不可撤销！请确认原因：
      "${reason.value}"
      
      确定要继续吗？
    `
    
    await ElMessageBox.confirm(message, '全量同步确认', {
      type: 'warning',
      confirmButtonText: '我已了解风险，继续',
      cancelButtonText: '取消',
      dangerouslyUseHTMLString: true
    })
    
    await roleApi.syncPermissions(roleId, {
      permission_ids: allKeys,
      reason: reason.value,
      confirm: true
    })
    
    ElMessage.success('权限同步成功')
  }
  
  visible.value = false
}
</script>
```

#### 2. 权限指令（参考vue-element-admin）

```typescript
// directives/permission.ts

import { DirectiveBinding } from 'vue'
import { useUserStore } from '@/stores/user'

/**
 * 权限指令
 * 用法：v-permission="['admin']"
 */
export const permission = {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    const { value } = binding
    const userStore = useUserStore()
    
    if (value && value instanceof Array && value.length > 0) {
      const permissions = value
      const hasPermission = userStore.hasPermission(permissions)
      
      if (!hasPermission) {
        // 移除元素
        el.parentNode?.removeChild(el)
      }
    } else {
      throw new Error('使用示例: v-permission="[\'admin\']"')
    }
  }
}

// 注册全局指令
app.directive('permission', permission)
```

#### 3. 权限检查函数

```typescript
// composables/usePermission.ts

import { useUserStore } from '@/stores/user'

export function usePermission() {
  const userStore = useUserStore()
  
  /**
   * 检查是否有权限
   * @param permissions 权限代码列表
   * @returns 是否有权限
   */
  const hasPermission = (permissions: string[]): boolean => {
    return userStore.hasAnyPermission(permissions)
  }
  
  /**
   * 检查是否有所有权限
   */
  const hasAllPermissions = (permissions: string[]): boolean => {
    return userStore.hasAllPermissions(permissions)
  }
  
  /**
   * 检查是否有角色
   */
  const hasRole = (roles: string[]): boolean => {
    return userStore.hasRole(roles)
  }
  
  return {
    hasPermission,
    hasAllPermissions,
    hasRole
  }
}

// 使用示例
const { hasPermission } = usePermission()

if (hasPermission(['user.delete'])) {
  // 显示删除按钮
}
```

---

## 🎯 实施计划

### Phase 1: 基础重构（1周）

1. ✅ 数据库表结构改造
2. ✅ 新增API接口（增量操作）
3. ✅ 改造Service层
4. ✅ 添加审计日志

### Phase 2: 前端改造（3天）

1. ✅ 权限配置页面重构
2. ✅ 添加预览功能
3. ✅ 添加二次确认
4. ✅ 权限指令优化

### Phase 3: 测试与优化（2天）

1. ✅ 单元测试
2. ✅ 集成测试
3. ✅ 性能测试
4. ✅ 安全测试

### Phase 4: 高级功能（可选）

1. 权限模板
2. 临时权限
3. 条件权限（ABAC）
4. 审批工作流

---

## 📚 参考资料

- Casbin官方文档: https://casbin.org/
- Django Permissions: https://docs.djangoproject.com/en/5.0/topics/auth/
- Spring Security: https://spring.io/projects/spring-security
- vue-element-admin: https://github.com/PanJiaChen/vue-element-admin
- Ant Design Pro: https://pro.ant.design/

---

## 💬 总结

这套方案结合了：
- ✅ Django的简洁性（扁平化权限设计）
- ✅ vue-element-admin的最佳实践（前端权限控制）
- ✅ Casbin的灵活性（支持扩展ABAC）
- ✅ Spring Security的企业级特性（审计、安全）

**核心改进：**
1. 明确的操作语义（增量 vs 全量）
2. 完善的变更确认机制
3. 详细的审计日志
4. 渐进式的功能增强

**下一步行动：**
告诉我你想从哪个Phase开始实施，我会提供具体的代码实现！
