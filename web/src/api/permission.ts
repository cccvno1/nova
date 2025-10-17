/**
 * 权限管理 API
 * 对应后端路由：/api/v1/permissions
 * 对应Handler：internal/handler/permission_handler.go
 */

import request from '@/utils/request'
import type { Permission, CreatePermissionDTO, UpdatePermissionDTO, PermissionListParams, PageData, PermissionType } from '@/types/api'

/**
 * 获取用户权限列表
 * @param userId 用户ID
 * @returns 用户的权限列表
 * @route GET /api/v1/user-roles/user/:userId/permissions
 */
export function getUserPermissions(userId: number) {
    return request<Permission[]>({
        url: `/api/v1/user-roles/user/${userId}/permissions`,
        method: 'GET'
    })
}

/**
 * 权限API集合
 */
export const permissionApi = {
    /**
     * 获取权限列表（分页）
     * @param params 查询参数（page, page_size, domain, type）
     * @returns 分页权限列表
     * @route GET /api/v1/permissions
     */
    list(params: PermissionListParams) {
        return request<PageData<Permission>>({
            url: '/api/v1/permissions',
            method: 'GET',
            params
        })
    },

    /**
     * 创建权限
     * @param data 权限信息
     * @returns 创建的权限
     * @route POST /api/v1/permissions
     */
    create(data: CreatePermissionDTO) {
        return request<Permission>({
            url: '/api/v1/permissions',
            method: 'POST',
            data
        })
    },

    /**
     * 获取权限详情
     * @param id 权限ID
     * @returns 权限详情
     * @route GET /api/v1/permissions/:id
     */
    getById(id: number) {
        return request<Permission>({
            url: `/api/v1/permissions/${id}`,
            method: 'GET'
        })
    },

    /**
     * 更新权限
     * @param id 权限ID
     * @param data 更新的权限信息
     * @route PUT /api/v1/permissions/:id
     */
    update(id: number, data: UpdatePermissionDTO) {
        return request<void>({
            url: `/api/v1/permissions/${id}`,
            method: 'PUT',
            data
        })
    },

    /**
     * 删除权限
     * @param id 权限ID
     * @route DELETE /api/v1/permissions/:id
     */
    delete(id: number) {
        return request<void>({
            url: `/api/v1/permissions/${id}`,
            method: 'DELETE'
        })
    },

    /**
     * 获取树形权限结构
     * @param params 查询参数（domain）
     * @returns 树形权限列表
     * @route GET /api/v1/permissions/tree
     */
    getTree(params?: { domain?: string }) {
        return request<Permission[]>({
            url: '/api/v1/permissions/tree',
            method: 'GET',
            params
        })
    },

    /**
     * 根据类型获取权限列表
     * @param type 权限类型
     * @param domain 域
     * @returns 权限列表
     * @route GET /api/v1/permissions/type/:type
     */
    listByType(type: PermissionType, domain?: string) {
        return request<Permission[]>({
            url: `/api/v1/permissions/type/${type}`,
            method: 'GET',
            params: { domain }
        })
    },

    /**
     * 搜索权限
     * @param params 搜索参数
     * @returns 权限列表
     * @route GET /api/v1/permissions/search
     */
    search(params: { keyword?: string; domain?: string; page?: number; page_size?: number }) {
        return request<PageData<Permission>>({
            url: '/api/v1/permissions/search',
            method: 'GET',
            params
        })
    }
}

export default permissionApi
