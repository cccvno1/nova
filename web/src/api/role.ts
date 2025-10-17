/**
 * 角色管理 API
 * 对应后端路由：/api/v1/roles
 * 对应Handler：internal/handler/role_handler.go
 */

import request from '@/utils/request'
import type { Role, CreateRoleDTO, UpdateRoleDTO, RoleListParams, PageData, Permission, AssignPermissionsDTO } from '@/types/api'

/**
 * 用户角色关联信息
 */
export interface UserRole {
    id: number
    user_id: number
    role_id: number
    domain: string
    created_at: string
    user?: {
        id: number
        username: string
        email: string
        nickname: string
    }
}

/**
 * 角色API集合
 */
export const roleApi = {
    /**
     * 获取角色列表（分页）
     * @param params 查询参数（page, page_size, domain, keyword）
     * @returns 分页角色列表
     * @route GET /api/v1/roles
     */
    list(params: RoleListParams) {
        return request<PageData<Role>>({
            url: '/api/v1/roles',
            method: 'GET',
            params
        })
    },

    /**
     * 创建角色
     * @param data 角色信息
     * @returns 创建的角色
     * @route POST /api/v1/roles
     */
    create(data: CreateRoleDTO) {
        return request<Role>({
            url: '/api/v1/roles',
            method: 'POST',
            data
        })
    },

    /**
     * 获取角色详情
     * @param id 角色ID
     * @returns 角色详情
     * @route GET /api/v1/roles/:id
     */
    getById(id: number) {
        return request<Role>({
            url: `/api/v1/roles/${id}`,
            method: 'GET'
        })
    },

    /**
     * 更新角色
     * @param id 角色ID
     * @param data 更新的角色信息
     * @route PUT /api/v1/roles/:id
     */
    update(id: number, data: UpdateRoleDTO) {
        return request<void>({
            url: `/api/v1/roles/${id}`,
            method: 'PUT',
            data
        })
    },

    /**
     * 删除角色
     * @param id 角色ID
     * @route DELETE /api/v1/roles/:id
     */
    delete(id: number) {
        return request<void>({
            url: `/api/v1/roles/${id}`,
            method: 'DELETE'
        })
    },

    /**
     * 搜索角色
     * @param params 搜索参数
     * @returns 角色列表
     * @route GET /api/v1/roles/search
     */
    search(params: { keyword?: string; domain?: string; page?: number; page_size?: number }) {
        return request<PageData<Role>>({
            url: '/api/v1/roles/search',
            method: 'GET',
            params
        })
    },

    /**
     * 获取角色的权限列表
     * @param id 角色ID
     * @returns 权限列表
     * @route GET /api/v1/roles/:id/permissions
     */
    getPermissions(id: number) {
        return request<Permission[]>({
            url: `/api/v1/roles/${id}/permissions`,
            method: 'GET'
        })
    },

    /**
     * 更新角色权限（支持预览）
     * @param id 角色ID
     * @param data 权限更新数据
     * @route POST /api/v1/roles/:id/permissions/update
     */
    updatePermissions(id: number, data: { permission_ids: number[]; preview?: boolean }) {
        return request<{
            preview?: {
                added: Permission[]
                removed: Permission[]
                kept: Permission[]
            }
            result?: {
                added_count: number
                removed_count: number
            }
        }>({
            url: `/api/v1/roles/${id}/permissions/update`,
            method: 'POST',
            data
        })
    },

    /**
     * 给角色分配权限（已废弃，保留向后兼容）
     * @deprecated 请使用 updatePermissions 替代
     * @param id 角色ID
     * @param data 权限ID列表
     * @route POST /api/v1/roles/:id/permissions
     */
    assignPermissions(id: number, data: AssignPermissionsDTO) {
        return request<void>({
            url: `/api/v1/roles/${id}/permissions`,
            method: 'POST',
            data
        })
    },

    /**
     * 撤销角色的权限
     * @param id 角色ID
     * @param permissionId 权限ID
     * @route DELETE /api/v1/roles/:id/permissions/:permissionId
     */
    revokePermission(id: number, permissionId: number) {
        return request<void>({
            url: `/api/v1/roles/${id}/permissions/${permissionId}`,
            method: 'DELETE'
        })
    },

    /**
     * 获取拥有某个角色的用户列表
     * @param id 角色ID
     * @returns 用户角色关联列表
     * @route GET /api/v1/roles/:id/users
     */
    getUsers(id: number) {
        return request<UserRole[]>({
            url: `/api/v1/roles/${id}/users`,
            method: 'GET'
        })
    }
}

export default roleApi
