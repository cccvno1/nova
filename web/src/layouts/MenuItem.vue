<template>
  <!-- 子菜单（有children） -->
  <el-sub-menu v-if="route.children && route.children.length > 0" :index="route.path">
    <template #title>
      <el-icon v-if="route.meta?.icon">
        <component :is="route.meta.icon" />
      </el-icon>
      <span>{{ route.meta?.title }}</span>
    </template>
    <menu-item
      v-for="child in route.children"
      :key="child.path"
      :route="child"
    />
  </el-sub-menu>

  <!-- 菜单项（无children） -->
  <el-menu-item v-else :index="route.path">
    <el-icon v-if="route.meta?.icon">
      <component :is="route.meta.icon" />
    </el-icon>
    <template #title>
      <span>{{ route.meta?.title }}</span>
    </template>
  </el-menu-item>
</template>

<script setup lang="ts">
/**
 * 递归菜单项组件
 * 功能：根据路由配置渲染菜单项或子菜单
 */
import type { RouteRecordRaw } from 'vue-router'

defineProps<{
  /** 路由配置对象 */
  route: RouteRecordRaw
}>()
</script>
