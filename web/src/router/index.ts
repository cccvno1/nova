import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useTabsStore } from '@/stores/tabs'
import { ElMessage } from 'element-plus'

/**
 * 公共路由配置（无需权限验证）
 * 包括：首页重定向、登录页、错误页
 */
export const constantRoutes: RouteRecordRaw[] = [
    {
        path: '/',
        redirect: '/home'
    },
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/login/index.vue'),
        meta: { hidden: true, title: '登录' }
    },
    {
        path: '/404',
        name: 'NotFound',
        component: () => import('@/views/error/404.vue'),
        meta: { hidden: true, title: '404' }
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes: constantRoutes
})

/** 路由白名单（无需登录验证的页面） */
const whiteList = ['/login', '/404']

/**
 * 全局前置守卫
 * 功能：
 * 1. 验证登录状态
 * 2. 动态加载用户权限和路由
 * 3. 自动添加多标签页
 * 4. 处理路由重定向
 */
router.beforeEach(async (to, _from, next) => {
    const userStore = useUserStore()
    const tabsStore = useTabsStore()
    const token = localStorage.getItem('access_token')

    if (token) {
        // 已登录状态
        if (to.path === '/login') {
            // 已登录用户访问登录页，重定向到首页
            next({ path: '/' })
        } else {
            // 检查是否已加载权限（使用permissionsLoaded标记避免重复请求）
            if (!userStore.permissionsLoaded) {
                try {
                    // 从后端获取权限数据并生成动态路由
                    // API: GET /api/v1/user-roles/user/:userId/permissions
                    await userStore.getUserPermissions()

                    // 重新导航到目标页面，确保动态路由已加载
                    next({ ...to, replace: true })
                } catch (error) {
                    console.error('获取权限失败:', error)

                    // 获取权限失败，清除登录状态
                    userStore.logout()
                    ElMessage.error('获取用户权限失败，请重新登录')
                    next({ path: '/login', query: { redirect: to.fullPath } })
                }
            } else {
                // 权限已加载，自动添加标签
                if (!to.meta.hidden && to.meta.title && to.name) {
                    tabsStore.addTab({
                        path: to.path,
                        title: to.meta.title as string,
                        name: to.name as string,
                        closable: !to.meta.affix // affix=true 的标签不可关闭
                    })
                }

                // 直接放行
                next()
            }
        }
    } else {
        // 未登录状态
        if (whiteList.includes(to.path)) {
            // 白名单页面直接放行
            next()
        } else {
            // 重定向到登录页，并记录目标页面
            next({ path: '/login', query: { redirect: to.fullPath } })
        }
    }
})

/**
 * 全局后置守卫
 * 功能：设置页面标题
 */
router.afterEach((to) => {
    document.title = `${to.meta.title || 'Nova Admin'} - ${import.meta.env.VITE_APP_TITLE}`
})

export default router
