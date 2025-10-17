/**
 * 多标签页Store
 * 功能：管理页面标签状态，支持多标签切换、缓存、关闭等
 * 依赖：Pinia、Vue Router
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

/**
 * 标签项接口
 */
export interface TabItem {
    /** 路由路径（唯一标识） */
    path: string
    /** 标签显示标题 */
    title: string
    /** 组件名称（keep-alive缓存用，必须与组件定义的name一致） */
    name: string
    /** 是否可关闭（首页固定不可关闭） */
    closable: boolean
    /** 查询参数（可选） */
    query?: Record<string, any>
}

export const useTabsStore = defineStore('tabs', () => {
    const router = useRouter()

    // ========== 状态 ==========

    /** 标签列表 */
    const tabs = ref<TabItem[]>([
        {
            path: '/home',
            title: '首页',
            name: 'Home',
            closable: false // 首页固定不可关闭
        }
    ])

    /** 当前激活的标签路径 */
    const activeTab = ref<string>('/home')

    /** keep-alive缓存的组件名列表 */
    const cachedViews = computed(() => {
        return tabs.value.map(tab => tab.name).filter(name => name)
    })

    // ========== 方法 ==========

    /**
     * 添加标签
     * @param tab 标签信息
     */
    function addTab(tab: TabItem) {
        // 检查标签是否已存在
        const existIndex = tabs.value.findIndex(t => t.path === tab.path)

        if (existIndex > -1) {
            // 已存在，只切换激活状态
            activeTab.value = tab.path
        } else {
            // 不存在，添加新标签
            tabs.value.push({
                ...tab,
                closable: tab.closable !== false // 默认可关闭
            })
            activeTab.value = tab.path
        }
    }

    /**
     * 关闭标签
     * @param path 标签路径
     */
    function closeTab(path: string) {
        const index = tabs.value.findIndex(t => t.path === path)

        if (index === -1) return

        const tab = tabs.value[index]

        if (!tab) return

        // 检查是否可关闭
        if (tab.closable === false) {
            console.warn(`标签 "${tab.title}" 不可关闭`)
            return
        }

        // 删除标签
        tabs.value.splice(index, 1)

        // 如果关闭的是当前激活标签，需要切换到其他标签
        if (activeTab.value === path) {
            // 优先切换到右侧标签，否则切换到左侧标签
            const nextTab = tabs.value[index] || tabs.value[index - 1] || tabs.value[0]
            if (nextTab) {
                activeTab.value = nextTab.path
                router.push(nextTab.path)
            }
        }
    }

    /**
     * 关闭其他标签（保留固定标签和当前标签）
     * @param path 要保留的标签路径
     */
    function closeOtherTabs(path: string) {
        tabs.value = tabs.value.filter(tab => {
            return tab.path === path || tab.closable === false
        })

        activeTab.value = path
        router.push(path)
    }

    /**
     * 关闭左侧标签
     * @param path 参考标签路径
     */
    function closeLeftTabs(path: string) {
        const index = tabs.value.findIndex(t => t.path === path)
        if (index === -1) return

        // 保留左侧固定标签
        tabs.value = tabs.value.filter((tab, i) => {
            return i >= index || tab.closable === false
        })
    }

    /**
     * 关闭右侧标签
     * @param path 参考标签路径
     */
    function closeRightTabs(path: string) {
        const index = tabs.value.findIndex(t => t.path === path)
        if (index === -1) return

        // 保留右侧固定标签
        tabs.value = tabs.value.filter((tab, i) => {
            return i <= index || tab.closable === false
        })

        // 如果当前激活标签被关闭，切换到参考标签
        const exists = tabs.value.find(t => t.path === activeTab.value)
        if (!exists) {
            activeTab.value = path
            router.push(path)
        }
    }

    /**
     * 关闭所有标签（仅保留固定标签）
     */
    function closeAllTabs() {
        tabs.value = tabs.value.filter(tab => tab.closable === false)

        // 切换到第一个固定标签（通常是首页）
        const firstTab = tabs.value[0]
        if (firstTab) {
            activeTab.value = firstTab.path
            router.push(firstTab.path)
        }
    }

    /**
     * 刷新标签（重新加载页面）
     * @param path 标签路径
     */
    function refreshTab(path: string) {
        const tab = tabs.value.find(t => t.path === path)
        if (!tab) return

        // 临时从缓存中移除，触发组件重新加载
        const index = tabs.value.findIndex(t => t.path === path)
        if (index > -1) {
            tabs.value.splice(index, 1)

            // 下一帧重新添加
            setTimeout(() => {
                tabs.value.splice(index, 0, tab)
            }, 0)
        }
    }

    /**
     * 设置当前激活标签
     * @param path 标签路径
     */
    function setActiveTab(path: string) {
        activeTab.value = path
        router.push(path)
    }

    /**
     * 获取标签信息
     * @param path 标签路径
     * @returns 标签信息
     */
    function getTab(path: string): TabItem | undefined {
        return tabs.value.find(t => t.path === path)
    }

    /**
     * 检查标签是否存在
     * @param path 标签路径
     * @returns 是否存在
     */
    function hasTab(path: string): boolean {
        return tabs.value.some(t => t.path === path)
    }

    return {
        // 状态
        tabs,
        activeTab,
        cachedViews,

        // 方法
        addTab,
        closeTab,
        closeOtherTabs,
        closeLeftTabs,
        closeRightTabs,
        closeAllTabs,
        refreshTab,
        setActiveTab,
        getTab,
        hasTab
    }
})
