# RBAC 权限模块

## 组件概览
- 服务：`internal/service/rbac_service.go`
- 仓储：`internal/repository/{role_repository.go,permission_repository.go}`
- HTTP 接口：`internal/handler/{role_handler.go,permission_handler.go,user_role_handler.go}`
- 模型定义：`internal/model/{role.go,permission.go}`
- Casbin 封装：`pkg/casbin/enforcer.go`
- 模型配置：`configs/rbac_model.conf`

## 核心概念
- **域（Domain）**：多租户隔离维度，角色、权限、策略都绑定域。跨域操作会被拒绝。
- **角色（Role）**：用于描述职责，与用户关联后继承对应权限。
- **权限（Permission）**：与资源、操作一一对应，可分类型（API、菜单、按钮等）。
- **策略（Policy）**：Casbin 中的 `p` 规则，`sub` 为角色 ID，`obj/act` 分别对应资源与操作。
- **分配记录**：`model.UserRole` 表记录“谁给谁在什么域授予了哪个角色”，便于审计。

## 初始化流程
1. 在 `router.Setup` 中实例化角色、权限、用户角色仓储。
2. 使用 `service.NewRBACService` 将仓储与 `casbin.Enforcer` 组合成统一服务。
3. 向 Echo 注册角色、权限、用户角色相关的 RESTful API。
4. Casbin 模型由 `configs/rbac_model.conf` 定义，加载路径来自配置 `casbin.model_path`。

## 角色管理
- 新建角色：`RBACService.CreateRole`
  - 校验同域唯一性 (`ExistsByName`)
  - 持久化到 `roles` 表
  - 输出日志便于定位
- 更新角色：保持域不变，防止跨域污染。
- 删除角色：
  - 禁止删除 `is_system` 角色
  - 清理该角色的所有策略 (`RemoveAllPoliciesForRole`)
  - 删除 Casbin 中的用户-角色关系
  - 清理数据库记录
- 列表与搜索：依赖 `repository.RoleRepository.List/Search`，支持分页与关键词过滤。

### 角色接口示例
```http
POST /api/v1/roles
{
  "name":"editor",
  "display_name":"内容编辑",
  "domain":"default"
}
```
`role_handler.go` 会校验请求体并调用服务，失败时统一返回结构化错误。

## 权限管理
- 新建/更新权限：校验名称、域唯一性，更新时若 `resource` 或 `action` 变更，会动态重写 Casbin 策略。
- 删除权限：
  - 禁止删除系统权限
  - 遍历策略，移除命中的 `obj/act`
  - 删除数据库记录
- 查询：
  - `List` 支持分页与域过滤
  - `ListByType` 用于前端按类型筛选菜单/按钮
  - `ListTree` 基于父子关系构建树形结构（`permission_repository.go` 中的 `buildPermissionTree`）

### 权限接口示例
```http
POST /api/v1/permissions
{
  "name":"user:list",
  "display_name":"查看用户",
  "type":"api",
  "domain":"default",
  "resource":"/api/v1/users",
  "action":"GET"
}
```

## 角色与权限的绑定
- `AssignPermissionsToRole`：
  - 校验角色域一致性
  - 批量查询权限，跳过域不一致的项
  - 通过 `enforcer.AddPolicy` 将 `roleID`（字符串）与资源/操作建立关系
- `RevokePermissionsFromRole`：对应使用 `RemovePolicy`
- `GetRolePermissions`：
  - 读取 Casbin 策略后再回查权限表，确保返回完整的元数据。

## 用户与角色的绑定
- `AssignRolesToUser`
  - 使用 `AddRoleForUser` 将用户 ID 与角色 ID 绑定到指定域
  - 将分配情况写入 `user_roles` 表（包含 `assigned_by`）
- `RevokeRolesFromUser` 同时清理 Casbin 与数据库
- `GetUserRoles` 先从 Casbin 获取角色 ID，再批量查询详情
- `GetRoleUsers` 直接通过 `UserRoleRepository.FindByRole`

### 用户角色接口
```http
POST /api/v1/user-roles
{
  "role_ids": [1,2],
  "domain": "default"
}
```
处理器会从上下文读取当前操作人 `user_id` 作为 `assigned_by` 写入数据库。

## 权限校验
- `CheckPermission`：封装 `enforcer.Enforce`，用于 `/user-roles` 相关接口内做单个请求鉴权。处理器会从认证中间件写入的上下文读取当前用户 ID，因此无需在路由上额外携带 `:user_id`。
- `GetUserPermissions`：使用 `GetImplicitPermissionsForUser` 获取用户所有策略（包含继承角色），随后匹配权限表返回带文案的权限列表，避免直接暴露策略原始数据。
- 请求进入 `middleware.Auth` 后，可结合 RBAC 结果做细粒度控制（示例接口直接返回布尔值）。

## 策略维护
- `AddPolicy` / `RemovePolicy` / `ListPolicies` 提供给需要直接操控 Casbin 表的高级用户。
- `pkg/casbin/enforcer.go` 扩展方法：
  - `AddRoleInheritance` 与 `GetRoleInheritance` 支持角色树
  - `DeleteDomain` 一键清除域下所有策略与关系
- 配置中的 `auto_save`、`auto_load` 控制策略变更持久化及多实例同步（通过定时 `LoadPolicy`）。

## 常见扩展
1. **预置角色/权限**：在迁移或启动脚本中写入基础数据，再调用 `AddPoliciesForRole` 批量加载。
2. **数据权限**：梳理 `model.DataScope` / `RoleDataScope`，结合仓储扩展业务查询。
3. **审计追踪**：结合审计日志模块记录 `AssignRolesToUser` 等操作，便于追溯。
4. **界面构建**：`ListPermissionsTree` 输出树形结构，可直接驱动权限勾选组件。

掌握本模块，可实现基于角色的访问控制，支持多域隔离、角色继承和策略动态调整。