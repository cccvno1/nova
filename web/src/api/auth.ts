import request from '@/utils/request'
import type { LoginRequest, TokenPair } from '@/types/api'

/**
 * 用户登录
 * API: POST /api/v1/auth/login
 * 后端: internal/handler/auth_handler.go -> Login
 * 
 * @param data 登录表单数据
 * @returns 访问令牌和刷新令牌
 */
export function login(data: LoginRequest) {
    return request<TokenPair>({
        url: '/api/v1/auth/login',
        method: 'post',
        data
    })
}

/**
 * 用户登出
 * API: POST /api/v1/auth/logout
 * 后端: internal/handler/auth_handler.go -> Logout
 * 
 * @returns void
 */
export function logout() {
    return request({
        url: '/api/v1/auth/logout',
        method: 'post'
    })
}

/**
 * 刷新访问令牌
 * API: POST /api/v1/auth/refresh
 * 后端: internal/handler/auth_handler.go -> RefreshToken
 * 
 * @param refreshToken 刷新令牌
 * @returns 新的访问令牌
 */
export function refreshToken(refreshToken: string) {
    return request<{ access_token: string }>({
        url: '/api/v1/auth/refresh',
        method: 'post',
        data: { refresh_token: refreshToken }
    })
}
