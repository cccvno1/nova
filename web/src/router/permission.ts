import type { RouteRecordRaw } from 'vue-router'
import type { Permission } from '@/types/api'

/**
 * 权限路由生成模块
 * 
 * 功能：将后端返回的权限列表转换为Vue Router路由配置
 * 
 * 路由生成策略：
 * 1. 顶级菜单（parent_id = null）
 *    - component = "Layout": 作为父路由容器，包含子路由
 *    - component = "views/xxx": 独立页面，使用Layout包裹
 * 2. 子菜单（parent_id != null）
 *    - 直接使用组件路径，作为父路由的children
 * 
 * 示例：
 * - menu:home (parent_id=null, component=views/home/index) → 独立页面
 * - menu:system (parent_id=null, component=Layout) → 父菜单容器
 *   ├─ menu:system:user (parent_id=2, component=views/system/user/index)
 */

/** 权限节点（包含children字段） */
interface PermissionNode extends Permission {
    children?: PermissionNode[]
}

/** Layout组件加载器 */
const LayoutComponent = () => import('@/layouts/index.vue')

/** 使用Vite的glob import预加载所有views组件，支持动态路由 */
const viewModules = import.meta.glob('@/views/**/*.vue')

/**
 * 从权限列表生成路由配置
 * @param permissions 后端返回的权限数组
 * @returns Vue Router路由配置数组
 */
export function generateRoutesFromPermissions(permissions: Permission[]): RouteRecordRaw[] {
    const menuPermissions = permissions.filter(p => p.type === 'menu')

    if (menuPermissions.length === 0) {
        console.warn('没有菜单类型的权限')
        return []
    }

    // 构建权限树
    const tree = buildPermissionTree(menuPermissions)

    // 转换为路由
    const routes: RouteRecordRaw[] = []
    tree.forEach(node => {
        const route = transformToRoute(node)
        if (route) {
            routes.push(route)
        }
    })

    if (import.meta.env.DEV) {
        console.log('生成路由数量:', routes.length)
    }

    return routes
}

/**
 * 构建权限树形结构
 * @param permissions 权限列表
 * @param parentId 父节点ID
 * @returns 树形权限节点数组
 */
function buildPermissionTree(permissions: Permission[], parentId: number | null = null): PermissionNode[] {
    const tree: PermissionNode[] = []

    for (const permission of permissions) {
        // 判断是否为目标节点的子节点
        const isMatch = (parentId === null)
            ? (permission.parent_id === null || permission.parent_id === 0)
            : (permission.parent_id === parentId)

        if (isMatch) {
            const node: PermissionNode = { ...permission }
            const children = buildPermissionTree(permissions, permission.id)
            if (children.length > 0) {
                node.children = children
            }
            tree.push(node)
        }
    }

    // 按sort字段排序
    return tree.sort((a, b) => a.sort - b.sort)
}

/**
 * 将权限节点转换为路由配置
 * @param node 权限节点
 * @returns 路由配置对象或null
 */
function transformToRoute(node: PermissionNode): RouteRecordRaw | null {
    const hasChildren = node.children && node.children.length > 0

    // 情况1: 父级菜单容器（component = "Layout"）
    if (node.component === 'Layout') {
        const route: Partial<RouteRecordRaw> = {
            path: node.path,
            name: node.name,
            component: LayoutComponent,
            meta: {
                title: node.display_name,
                icon: node.icon || 'Menu',
                hidden: false,
                affix: node.path === '/home' // 首页标签固定不可关闭
            }
        }

        if (hasChildren && node.children && node.children.length > 0) {
            const firstChild = node.children[0]
            if (firstChild) {
                route.redirect = firstChild.path
            }
            route.children = node.children
                .map(child => transformToRoute(child))
                .filter((r): r is RouteRecordRaw => r !== null)
        }

        return route as RouteRecordRaw
    }

    // 情况2: 独立页面（顶级菜单，但不是Layout）
    if (node.parent_id === null || node.parent_id === 0) {
        return {
            path: node.path,
            component: LayoutComponent,
            meta: { hidden: true },
            children: [{
                path: '',
                name: node.name,
                component: loadViewComponent(node.component),
                meta: {
                    title: node.display_name,
                    icon: node.icon || 'Menu',
                    hidden: false,
                    affix: node.path === '/home' // 首页标签固定不可关闭
                }
            }]
        } as RouteRecordRaw
    }

    // 情况3: 子菜单页面
    const route: Partial<RouteRecordRaw> = {
        path: node.path,
        name: node.name,
        component: loadViewComponent(node.component),
        meta: {
            title: node.display_name,
            icon: node.icon || 'Menu',
            hidden: false,
            affix: false // 子菜单默认可关闭
        }
    }

    // 递归处理子节点
    if (hasChildren && node.children) {
        route.children = node.children
            .map(child => transformToRoute(child))
            .filter((r): r is RouteRecordRaw => r !== null)
    }

    return route as RouteRecordRaw
}

/**
 * 加载视图组件
 * @param componentPath 组件路径（如: "views/home/index" 或 "home/index"）
 * @returns 组件加载函数
 */
function loadViewComponent(componentPath: string) {
    if (!componentPath) {
        console.error('组件路径为空')
        return LayoutComponent
    }

    // 规范化路径：移除可能的 "views/" 前缀和 ".vue" 后缀
    const cleanPath = componentPath
        .replace(/^views\//, '')
        .replace(/\.vue$/, '')

    // 构建完整的模块路径
    const modulePath = `/src/views/${cleanPath}.vue`

    // 从预加载的模块中查找
    const matchedModule = viewModules[modulePath]

    if (!matchedModule) {
        console.error(`找不到组件: ${modulePath}`)
        if (import.meta.env.DEV) {
            console.log('可用的组件:', Object.keys(viewModules))
        }
        return LayoutComponent
    }

    return matchedModule
}

/**
 * 检查用户是否有权限访问指定路由
 * @param routeName 路由名称
 * @param userPermissions 用户权限列表
 * @returns 是否有权限
 */
export function hasRoutePermission(routeName: string, userPermissions: Permission[]): boolean {
    return userPermissions.some(p => p.name === routeName)
}
