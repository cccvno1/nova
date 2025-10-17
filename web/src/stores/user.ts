import { defineStore } from 'pinia'
import { login as loginApi, logout as logoutApi } from '@/api/auth'
import { getUserPermissions } from '@/api/permission'
import type { LoginRequest, User, Permission, Role } from '@/types/api'
import router from '@/router'
import { generateRoutesFromPermissions } from '@/router/permission'

/**
 * 用户状态接口
 */
interface UserState {
    /** 用户信息 */
    user: User | null
    /** 访问令牌 */
    token: string
    /** 用户权限列表（从后端获取） */
    permissions: Permission[]
    /** 用户角色列表 */
    roles: Role[]
    /** 权限是否已加载（防止重复请求） */
    permissionsLoaded: boolean
}

export const useUserStore = defineStore('user', {
    state: (): UserState => ({
        user: null,
        token: localStorage.getItem('access_token') || '',
        permissions: [],
        roles: [],
        permissionsLoaded: false
    }),

    getters: {
        /**
         * 是否已登录
         */
        isLoggedIn: (state) => !!state.token,

        /**
         * 获取菜单权限（用于生成侧边栏菜单）
         */
        menuPermissions: (state) => {
            return state.permissions.filter(p => p.type === 'menu' && p.status === 1)
        },

        /**
         * 获取按钮权限（用于控制按钮显示）
         */
        buttonPermissions: (state) => {
            return state.permissions
                .filter(p => p.type === 'button')
                .map(p => p.name)
        }
    },

    actions: {
        /**
         * 用户登录
         * 流程：
         * 1. 调用后端登录接口（POST /api/v1/auth/login）
         * 2. 保存访问令牌到localStorage
         * 3. 获取用户权限并生成动态路由
         * 
         * @param loginForm 登录表单数据
         * @returns 令牌数据
         */
        async login(loginForm: LoginRequest) {
            // request.ts拦截器已处理响应，直接返回TokenPair类型
            const tokenData = await loginApi(loginForm) as any

            // 保存令牌
            this.token = tokenData.access_token
            localStorage.setItem('access_token', tokenData.access_token)
            localStorage.setItem('refresh_token', tokenData.refresh_token)

            // 获取用户权限
            await this.getUserPermissions()

            return tokenData
        },

        /**
         * 获取用户权限并生成动态路由
         * API: GET /api/v1/user-roles/user/:userId/permissions
         * 
         * 流程：
         * 1. 检查权限是否已加载（防止重复请求）
         * 2. 从JWT中解析用户ID
         * 3. 调用后端获取权限列表
         * 4. 根据权限生成动态路由并注册
         * 5. 设置权限加载标记
         */
        async getUserPermissions() {
            // 防止重复获取
            if (this.permissionsLoaded) {
                if (import.meta.env.DEV) {
                    console.log('权限已加载，跳过重复请求')
                }
                return
            }

            try {
                // 从JWT解析用户ID
                const userId = this.getUserIdFromToken()

                // 从后端获取权限列表
                // @ts-ignore - request拦截器已处理响应类型
                const permissionsData: Permission[] = await getUserPermissions(userId)

                // 处理后端可能返回null的情况
                this.permissions = permissionsData || []

                if (import.meta.env.DEV) {
                    console.log(`获取到 ${this.permissions.length} 个权限`)
                }

                // 根据权限生成动态路由
                if (this.permissions.length > 0) {
                    const dynamicRoutes = generateRoutesFromPermissions(this.permissions)

                    // 注意：Vue Router 4中addRoute重复添加相同name会覆盖
                    dynamicRoutes.forEach((route) => {
                        router.addRoute(route)
                    })
                }

                // 添加404路由（必须放在最后）
                router.addRoute({
                    path: '/:pathMatch(.*)*',
                    name: 'NotFoundCatch',
                    redirect: '/404'
                })

                // 标记权限已加载
                this.permissionsLoaded = true
            } catch (error) {
                console.error('获取用户权限失败:', error)
                // 失败时重置标记，允许重试
                this.permissionsLoaded = false
                throw error
            }
        },

        /**
         * 从JWT令牌中解析用户ID
         * JWT格式: header.payload.signature
         * payload中包含user_id字段
         * 
         * @returns 用户ID，解析失败返回0
         */
        getUserIdFromToken(): number {
            try {
                const token = this.token || localStorage.getItem('access_token')
                if (!token) return 0

                // 解析JWT payload（Base64编码）
                const parts = token.split('.')
                if (parts.length !== 3) return 0

                const payload = parts[1]
                if (!payload) return 0

                const decoded = JSON.parse(atob(payload))
                return decoded.user_id || 0
            } catch (error) {
                console.error('解析JWT失败:', error)
                return 0
            }
        },

        /**
         * 检查用户是否拥有指定权限
         * @param permissionName 权限名称
         * @returns 是否拥有权限
         */
        hasPermission(permissionName: string): boolean {
            return this.permissions.some(p => p.name === permissionName)
        },

        /**
         * 用户登出
         * 流程：
         * 1. 调用后端登出接口（POST /api/v1/auth/logout）
         * 2. 清空本地状态和localStorage
         * 3. 跳转到登录页
         */
        async logout() {
            try {
                await logoutApi()
            } catch (error) {
                console.error('登出接口调用失败:', error)
            } finally {
                // 清空状态
                this.token = ''
                this.user = null
                this.permissions = []
                this.roles = []
                this.permissionsLoaded = false

                // 清空本地存储
                localStorage.removeItem('access_token')
                localStorage.removeItem('refresh_token')

                // 跳转到登录页
                router.push('/login')
            }
        }
    }
})
