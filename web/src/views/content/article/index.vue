<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文章管理</span>
          <el-button type="primary" :icon="'Plus'">新增文章</el-button>
        </div>
      </template>
      
      <el-alert
        title="功能开发中"
        type="info"
        description="文章管理功能正在开发中，敬请期待！"
        :closable="false"
        show-icon
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
// 文章管理页面占位
</script>

<style scoped lang="scss">
.page-container {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
