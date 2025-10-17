/**
 * API类型定义
 * 与后端模型保持一致，确保前后端类型同步
 * 参考：数据库表结构、Go Model定义、Handler请求/响应
 */

// ============ 基础类型（对应后端 pkg/response/response.go） ============

/** API统一响应格式 */
export interface ApiResponse<T = any> {
    /** 状态码：0表示成功 */
    code: number
    /** 响应消息 */
    message: string
    /** 响应数据 */
    data: T
}

/** 分页数据格式（后端返回） */
export interface PageData<T = any> {
    /** 数据列表 */
    items: T[]
    /** 总数 */
    total: number
    /** 当前页码 */
    page: number
    /** 每页数量 */
    page_size: number
}

/** 分页查询参数（前端发送） */
export interface PageParams {
    /** 页码（从1开始） */
    page?: number
    /** 每页数量 */
    page_size?: number
    /** 关键词搜索（可选） */
    keyword?: string
}

// ============ 认证相关（对应 internal/handler/auth_handler.go） ============

/** 登录请求参数 */
export interface LoginRequest {
    /** 用户名 */
    username: string
    /** 密码 */
    password: string
}

/** 注册请求参数 */
export interface RegisterRequest {
    /** 用户名 */
    username: string
    /** 邮箱 */
    email: string
    /** 密码 */
    password: string
    /** 昵称（可选） */
    nickname?: string
}

/** JWT令牌对 */
export interface TokenPair {
    /** 访问令牌 */
    access_token: string
    /** 刷新令牌 */
    refresh_token: string
    /** 过期时间（秒） */
    expires_in: number
}

// ============ 用户相关（对应 internal/model/user.go 和数据库表 users） ============

/**
 * 用户模型
 * 对应数据库表：users
 * 对应Go Model：internal/model/user.go User
 */
export interface User {
    /** 用户ID */
    id: number
    /** 用户名（唯一） */
    username: string
    /** 邮箱（唯一） */
    email: string
    /** 昵称 */
    nickname: string
    /** 头像URL */
    avatar: string
    /** 状态：1-启用，2-禁用（注意：数据库是bigint，Go是int，前端是number） */
    status: number
    /** 创建时间 */
    created_at: string
    /** 更新时间 */
    updated_at: string
}

/**
 * 创建用户请求（对应 internal/service/user_service.go CreateUserRequest）
 */
export interface CreateUserDTO {
    /** 用户名（必填，3-50字符） */
    username: string
    /** 邮箱（必填，邮箱格式） */
    email: string
    /** 密码（必填，6-32字符） */
    password: string
    /** 昵称（可选，最多50字符） */
    nickname?: string
}

/**
 * 更新用户请求（对应 internal/service/user_service.go UpdateUserRequest）
 */
export interface UpdateUserDTO {
    /** 昵称（可选） */
    nickname?: string
    /** 头像URL（可选） */
    avatar?: string
}

/**
 * 用户列表查询参数
 */
export interface UserListParams extends PageParams {
    /** 状态筛选：1-启用，2-禁用 */
    status?: number
}

// ============ 角色相关（对应 internal/model/role.go 和数据库表 roles） ============

/**
 * 角色模型
 * 对应数据库表：roles
 * 对应Go Model：internal/model/role.go Role
 */
export interface Role {
    /** 角色ID */
    id: number
    /** 角色标识（如：admin, editor） */
    name: string
    /** 显示名称（如：系统管理员） */
    display_name: string
    /** 描述 */
    description: string
    /** 域（默认：default） */
    domain: string
    /** 分类（如：system, business） */
    category: string
    /** 是否系统角色（不可删除） */
    is_system: boolean
    /** 排序权重 */
    sort: number
    /** 状态：1-启用，0-禁用（注意：数据库是smallint，Go是int8，前端是number） */
    status: number
    /** 创建时间 */
    created_at: string
    /** 更新时间 */
    updated_at: string
}

/**
 * 创建角色请求（对应 internal/handler/role_handler.go CreateRoleRequest）
 */
export interface CreateRoleDTO {
    /** 角色标识（必填，2-100字符） */
    name: string
    /** 显示名称（必填，2-100字符） */
    display_name: string
    /** 描述（可选，最多500字符） */
    description?: string
    /** 域（必填，1-100字符） */
    domain: string
    /** 分类（可选，最多50字符） */
    category?: string
    /** 排序权重（可选，默认0） */
    sort?: number
}

/**
 * 更新角色请求（对应 internal/handler/role_handler.go UpdateRoleRequest）
 */
export interface UpdateRoleDTO {
    /** 显示名称（必填） */
    display_name: string
    /** 描述（可选） */
    description?: string
    /** 分类（可选） */
    category?: string
    /** 排序权重（可选） */
    sort?: number
    /** 状态：1-启用，0-禁用（可选） */
    status?: number
}

/**
 * 角色列表查询参数
 */
export interface RoleListParams extends PageParams {
    /** 域筛选 */
    domain?: string
}

/**
 * 分配权限请求（对应 internal/handler/role_handler.go AssignPermissionsRequest）
 */
export interface AssignPermissionsDTO {
    /** 权限ID列表（必填，至少1个） */
    permission_ids: number[]
    /** 域（可选，默认default） */
    domain?: string
}

// ============ 权限相关（对应 internal/model/permission.go 和数据库表 permissions） ============

/** 权限类型 */
export type PermissionType = 'api' | 'menu' | 'button' | 'data' | 'field'

/**
 * 权限模型
 * 对应数据库表：permissions
 * 对应Go Model：internal/model/permission.go Permission
 */
export interface Permission {
    /** 权限ID */
    id: number
    /** 权限标识（如：user:read） */
    name: string
    /** 显示名称（如：查看用户） */
    display_name: string
    /** 描述 */
    description: string
    /** 权限类型：api、menu、button、data、field */
    type: PermissionType
    /** 域（默认：default） */
    domain: string
    /** 资源路径（如：/api/v1/users） */
    resource: string
    /** 操作（如：GET, POST, read, write） */
    action: string
    /** 分类（如：用户管理、订单管理） */
    category: string
    /** 父权限ID（用于树形结构，0表示根节点） */
    parent_id: number
    /** 前端路由路径（菜单权限用） */
    path: string
    /** 前端组件路径（菜单权限用） */
    component: string
    /** 图标（菜单权限用） */
    icon: string
    /** 是否系统权限（不可删除） */
    is_system: boolean
    /** 排序权重 */
    sort: number
    /** 状态：1-启用，0-禁用（注意：数据库是smallint，Go是int8，前端是number） */
    status: number
    /** 创建时间 */
    created_at: string
    /** 更新时间 */
    updated_at: string
    /** 子权限列表（树形结构用） */
    children?: Permission[]
}

/**
 * 创建权限请求（对应 internal/handler/permission_handler.go CreatePermissionRequest）
 */
export interface CreatePermissionDTO {
    /** 权限标识（必填，2-100字符） */
    name: string
    /** 显示名称（必填，2-100字符） */
    display_name: string
    /** 描述（可选，最多500字符） */
    description?: string
    /** 权限类型（必填） */
    type: PermissionType
    /** 域（必填，1-100字符） */
    domain: string
    /** 资源路径（必填，最多200字符） */
    resource: string
    /** 操作（必填，最多50字符） */
    action: string
    /** 分类（可选，最多50字符） */
    category?: string
    /** 父权限ID（可选，0表示根节点） */
    parent_id?: number
    /** 前端路由路径（可选） */
    path?: string
    /** 前端组件路径（可选） */
    component?: string
    /** 图标（可选） */
    icon?: string
    /** 排序权重（可选，默认0） */
    sort?: number
}

/**
 * 更新权限请求（对应 internal/handler/permission_handler.go UpdatePermissionRequest）
 */
export interface UpdatePermissionDTO {
    /** 显示名称（必填） */
    display_name: string
    /** 描述（可选） */
    description?: string
    /** 权限类型（必填） */
    type: PermissionType
    /** 资源路径（必填） */
    resource: string
    /** 操作（必填） */
    action: string
    /** 分类（可选） */
    category?: string
    /** 父权限ID（可选） */
    parent_id?: number
    /** 前端路由路径（可选） */
    path?: string
    /** 前端组件路径（可选） */
    component?: string
    /** 图标（可选） */
    icon?: string
    /** 排序权重（可选） */
    sort?: number
    /** 状态：1-启用，0-禁用（可选） */
    status?: number
}

/**
 * 权限列表查询参数
 */
export interface PermissionListParams extends PageParams {
    /** 域筛选 */
    domain?: string
    /** 类型筛选 */
    type?: PermissionType
}
