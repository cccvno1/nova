<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="collapse ? '64px' : '200px'" class="layout-aside">
      <div class="logo" :class="{ 'logo-collapse': collapse }">
        <span v-if="!collapse" class="logo-title">Nova Admin</span>
        <span v-else class="logo-icon">N</span>
      </div>

      <el-scrollbar class="menu-scrollbar">
        <el-menu
          :default-active="activeMenu"
          :collapse="collapse"
          :unique-opened="true"
          router
        >
          <menu-item
            v-for="route in menuRoutes"
            :key="route.path"
            :route="route"
          />
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="layout-header">
        <div class="header-left">
          <el-icon class="collapse-icon" @click="collapse = !collapse">
            <component :is="collapse ? 'Expand' : 'Fold'" />
          </el-icon>
        </div>

        <div class="header-right">
          <!-- 主题切换按钮 -->
          <el-tooltip :content="isDark ? '切换到亮色模式' : '切换到暗色模式'" placement="bottom">
            <div class="theme-switch" @click="toggleTheme">
              <el-icon :size="18">
                <Sunny v-if="isDark" />
                <Moon v-else />
              </el-icon>
            </div>
          </el-tooltip>

          <!-- 用户信息 -->
          <el-dropdown @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" icon="User" />
              <span class="username">{{ userStore.user?.username || '用户' }}</span>
              <el-icon><CaretBottom /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 标签栏 -->
      <tab-bar />

      <!-- 内容区 -->
      <el-main class="layout-main">
        <router-view v-slot="{ Component }">
          <keep-alive :include="tabsStore.cachedViews">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </keep-alive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useTabsStore } from '@/stores/tabs'
import { useTheme } from '@/composables/useTheme'
import TabBar from './components/TabBar.vue'
import MenuItem from './MenuItem.vue'
import type { Permission } from '@/types/api'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const tabsStore = useTabsStore()

/** 主题切换 */
const { isDark, toggleTheme } = useTheme()

/** 侧边栏折叠状态 */
const collapse = ref(false)

/** 当前激活的菜单路径 */
const activeMenu = computed(() => route.path)

/**
 * 菜单路由列表
 * 功能：从userStore.permissions构建菜单树形结构
 * 注意：直接从权限生成，而不是从router.getRoutes()获取，确保结构一致
 */
const menuRoutes = computed(() => {
  const menuPermissions = userStore.permissions.filter(p => p.type === 'menu')
  
  if (menuPermissions.length === 0) {
    return []
  }
  
  /**
   * 构建权限树形结构
   * @param permissions 权限列表
   * @param parentId 父节点ID
   * @returns 树形结构数组
   */
  const buildTree = (permissions: Permission[], parentId: number | null = null): Permission[] => {
    const tree: Permission[] = []
    
    for (const perm of permissions) {
      const isMatch = (parentId === null)
        ? (perm.parent_id === null || perm.parent_id === 0)
        : (perm.parent_id === parentId)
      
      if (isMatch) {
        const node = { ...perm }
        const children = buildTree(permissions, perm.id)
        if (children.length > 0) {
          node.children = children
        }
        tree.push(node)
      }
    }
    
    return tree.sort((a, b) => a.sort - b.sort)
  }
  
  const tree = buildTree(menuPermissions)
  
  /**
   * 转换为菜单路由格式
   * @param node 权限节点
   * @returns 菜单路由对象
   */
  const toRoute = (node: Permission): any => {
    const menuRoute = {
      path: node.path,
      name: node.name,
      meta: {
        title: node.display_name,
        icon: node.icon || 'Menu'
      },
      children: undefined as any[] | undefined
    }
    
    if (node.children && node.children.length > 0) {
      menuRoute.children = node.children.map(toRoute)
    }
    
    return menuRoute
  }
  
  return tree.map(toRoute)
})

/**
 * 下拉菜单命令处理
 * @param command 命令类型
 */
const handleCommand = async (command: string) => {
  if (command === 'logout') {
    await userStore.logout()
  }
}
</script>

<style scoped lang="scss">
.layout-container {
  width: 100%;
  height: 100vh;
}

.layout-aside {
  background: #001529;
  transition: width 0.3s;

  .logo {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 60px;
    color: #fff;
    font-size: 20px;
    font-weight: bold;
    border-bottom: 1px solid #002140;

    &.logo-collapse {
      .logo-icon {
        font-size: 24px;
      }
    }
  }

  .menu-scrollbar {
    height: calc(100vh - 60px);

    :deep(.el-menu) {
      border-right: none;
      background: #001529;
    }

    :deep(.el-menu-item),
    :deep(.el-sub-menu__title) {
      color: rgba(255, 255, 255, 0.65);

      &:hover {
        background: #002140 !important;
        color: #fff;
      }

      &.is-active {
        background: #1890ff !important;
        color: #fff;
      }
    }
  }
}

.layout-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);

  .header-left {
    display: flex;
    align-items: center;

    .collapse-icon {
      font-size: 20px;
      cursor: pointer;
      transition: color 0.3s;

      &:hover {
        color: #1890ff;
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;

    .theme-switch {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 36px;
      height: 36px;
      cursor: pointer;
      border-radius: 4px;
      transition: all 0.3s;

      &:hover {
        background: #f5f7fa;
        color: #1890ff;
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 0 12px;
      cursor: pointer;
      transition: background 0.3s;
      border-radius: 4px;

      &:hover {
        background: #f5f7fa;
      }

      .username {
        font-size: 14px;
        color: #333;
      }
    }
  }
}

.layout-main {
  background: #f0f2f5;
  padding: 0;
  overflow-y: auto;
  min-height: calc(100vh - 60px - 40px); // 100vh - header高度 - tabbar高度
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s, transform 0.3s;
}

.fade-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

/* 暗色主题适配 */
html.dark {
  .layout-header {
    background: #1f1f1f;
    border-bottom-color: #333;

    .header-right {
      .theme-switch:hover {
        background: #333;
      }

      .user-info {
        &:hover {
          background: #333;
        }

        .username {
          color: #fff;
        }
      }
    }
  }

  .layout-main {
    background: #141414;
  }
}
</style>
