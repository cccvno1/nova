/**
 * 用户管理 API
 * 对应后端路由：/api/v1/users
 * 对应Handler：internal/handler/user_handler.go
 */

import request from '@/utils/request'
import type { User, CreateUserDTO, UpdateUserDTO, UserListParams, PageData } from '@/types/api'

/**
 * 用户API集合
 */
export const userApi = {
    /**
     * 获取用户列表（分页）
     * @param params 查询参数（page, page_size, keyword, status）
     * @returns 分页用户列表
     * @route GET /api/v1/users
     */
    list(params: UserListParams) {
        return request<PageData<User>>({
            url: '/api/v1/users',
            method: 'GET',
            params
        })
    },

    /**
     * 创建用户
     * @param data 用户信息
     * @returns 创建的用户
     * @route POST /api/v1/users
     */
    create(data: CreateUserDTO) {
        return request<User>({
            url: '/api/v1/users',
            method: 'POST',
            data
        })
    },

    /**
     * 获取用户详情
     * @param id 用户ID
     * @returns 用户详情
     * @route GET /api/v1/users/:id
     */
    getById(id: number) {
        return request<User>({
            url: `/api/v1/users/${id}`,
            method: 'GET'
        })
    },

    /**
     * 更新用户
     * @param id 用户ID
     * @param data 更新的用户信息
     * @route PUT /api/v1/users/:id
     */
    update(id: number, data: UpdateUserDTO) {
        return request<void>({
            url: `/api/v1/users/${id}`,
            method: 'PUT',
            data
        })
    },

    /**
     * 删除用户
     * @param id 用户ID
     * @route DELETE /api/v1/users/:id
     */
    delete(id: number) {
        return request<void>({
            url: `/api/v1/users/${id}`,
            method: 'DELETE'
        })
    }
}

export default userApi
