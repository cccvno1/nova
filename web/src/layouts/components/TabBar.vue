<template>
  <div class="tab-bar-container">
    <el-tabs
      v-model="tabsStore.activeTab"
      type="card"
      class="tab-bar-tabs"
      @tab-click="handleTabClick"
      @tab-remove="handleTabRemove"
    >
      <el-tab-pane
        v-for="tab in tabsStore.tabs"
        :key="tab.path"
        :label="tab.title"
        :name="tab.path"
        :closable="tab.closable"
      >
        <template #label>
          <span
            class="tab-label"
            @contextmenu.prevent="handleContextMenu($event, tab)"
          >
            {{ tab.title }}
          </span>
        </template>
      </el-tab-pane>
    </el-tabs>

    <!-- 右键菜单 -->
    <el-dropdown
      ref="contextMenuRef"
      trigger="contextmenu"
      :virtual-triggering="true"
      :virtual-ref="triggerRef"
      @command="handleCommand"
    >
      <span />
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="refresh">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-dropdown-item>
          <el-dropdown-item
            command="close"
            :disabled="!contextTab?.closable"
          >
            <el-icon><Close /></el-icon>
            关闭
          </el-dropdown-item>
          <el-dropdown-item command="closeOthers">
            <el-icon><CircleClose /></el-icon>
            关闭其他
          </el-dropdown-item>
          <el-dropdown-item command="closeLeft">
            <el-icon><Back /></el-icon>
            关闭左侧
          </el-dropdown-item>
          <el-dropdown-item command="closeRight">
            <el-icon><Right /></el-icon>
            关闭右侧
          </el-dropdown-item>
          <el-dropdown-item command="closeAll">
            <el-icon><CloseBold /></el-icon>
            关闭所有
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<script setup lang="ts">
import { ref, unref } from 'vue'
import { useRouter } from 'vue-router'
import { useTabsStore, type TabItem } from '@/stores/tabs'
import type { TabPaneName } from 'element-plus'

const router = useRouter()
const tabsStore = useTabsStore()

// 右键菜单相关
const contextMenuRef = ref()
const triggerRef = ref()
const contextTab = ref<TabItem>()

/**
 * 标签点击事件
 * @param tab 标签对象
 */
const handleTabClick = (tab: { paneName: TabPaneName }) => {
  const path = tab.paneName as string
  if (path && path !== router.currentRoute.value.path) {
    router.push(path)
  }
}

/**
 * 标签移除事件（点击×按钮）
 * @param tabName 标签名称（path）
 */
const handleTabRemove = (tabName: TabPaneName) => {
  tabsStore.closeTab(tabName as string)
}

/**
 * 右键菜单显示
 * @param event 鼠标事件
 * @param tab 标签对象
 */
const handleContextMenu = (event: MouseEvent, tab: TabItem) => {
  const dropdown = unref(contextMenuRef)
  if (!dropdown) return

  // 保存当前右键的标签
  contextTab.value = tab

  // 设置触发元素位置
  triggerRef.value = {
    getBoundingClientRect() {
      return {
        left: event.clientX,
        top: event.clientY,
        right: event.clientX,
        bottom: event.clientY,
        width: 0,
        height: 0
      }
    }
  }

  // 显示下拉菜单
  dropdown.handleOpen()
}

/**
 * 右键菜单命令处理
 * @param command 命令类型
 */
const handleCommand = (command: string) => {
  if (!contextTab.value) return

  const path = contextTab.value.path

  switch (command) {
    case 'refresh':
      // 刷新当前标签
      tabsStore.refreshTab(path)
      break
    case 'close':
      // 关闭当前标签
      tabsStore.closeTab(path)
      break
    case 'closeOthers':
      // 关闭其他标签
      tabsStore.closeOtherTabs(path)
      break
    case 'closeLeft':
      // 关闭左侧标签
      tabsStore.closeLeftTabs(path)
      break
    case 'closeRight':
      // 关闭右侧标签
      tabsStore.closeRightTabs(path)
      break
    case 'closeAll':
      // 关闭所有标签
      tabsStore.closeAllTabs()
      break
  }

  // 清空上下文标签
  contextTab.value = undefined
}
</script>

<style scoped lang="scss">
.tab-bar-container {
  position: relative;
  height: 40px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);

  .tab-bar-tabs {
    height: 100%;

    :deep(.el-tabs__header) {
      margin: 0;
      border-bottom: none;

      .el-tabs__nav {
        border: none;
      }

      .el-tabs__item {
        height: 40px;
        line-height: 40px;
        border: none;
        border-right: 1px solid #e8e8e8;
        background: #fff;
        color: #666;
        padding: 0 20px;
        transition: all 0.3s;

        &:hover {
          background: #f5f7fa;
          color: #333;
        }

        &.is-active {
          background: #1890ff;
          color: #fff;
          font-weight: 500;

          &:hover {
            background: #40a9ff;
          }
        }

        .el-icon-close {
          &:hover {
            background: rgba(0, 0, 0, 0.2);
            color: #fff;
          }
        }
      }

      .el-tabs__nav-wrap {
        &::after {
          display: none;
        }
      }
    }

    .tab-label {
      display: inline-block;
      user-select: none;
    }
  }
}

// 暗色主题适配
html.dark {
  .tab-bar-container {
    background: #1f1f1f;
    border-bottom-color: #333;

    .tab-bar-tabs {
      :deep(.el-tabs__item) {
        background: #1f1f1f;
        border-right-color: #333;
        color: #999;

        &:hover {
          background: #2a2a2a;
          color: #fff;
        }

        &.is-active {
          background: #1890ff;
          color: #fff;
        }
      }
    }
  }
}
</style>
