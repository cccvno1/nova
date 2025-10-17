/**
 * 主题切换 Composable
 * 功能：切换亮色/暗色主题，并持久化到 localStorage
 * 集成：Element Plus 暗色模式
 */

import { watch } from 'vue'
import { useDark, useToggle } from '@vueuse/core'

/**
 * 主题切换 Hook
 */
export function useTheme() {
    // 使用 VueUse 的 useDark（自动同步到 html.dark class）
    const isDark = useDark({
        selector: 'html',
        attribute: 'class',
        valueDark: 'dark',
        valueLight: '',
        storageKey: 'nova-theme',
        storage: localStorage
    })

    // 切换主题
    const toggleTheme = useToggle(isDark)

    // 监听主题变化，同步到 Element Plus
    watch(
        isDark,
        (dark) => {
            // Element Plus 暗色模式通过 html class 控制
            // 已通过 useDark 自动添加/移除 'dark' class
            if (dark) {
                document.documentElement.classList.add('dark')
            } else {
                document.documentElement.classList.remove('dark')
            }
        },
        { immediate: true }
    )

    return {
        /** 是否暗色主题 */
        isDark,
        /** 切换主题 */
        toggleTheme
    }
}

export default useTheme
