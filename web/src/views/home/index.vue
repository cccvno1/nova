<template>
  <div class="home-container">
    <el-row :gutter="20">
      <!-- 欢迎卡片 -->
      <el-col :span="24">
        <el-card class="welcome-card">
          <div class="welcome-content">
            <div class="welcome-text">
              <h2>欢迎回来，{{ userStore.user?.username }}!</h2>
              <p>今天是 {{ currentDate }}，祝你工作愉快！</p>
            </div>
            <el-icon class="welcome-icon" :size="80">
              <component :is="'User'" />
            </el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 统计卡片 -->
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon user" :size="40">
              <component :is="'User'" />
            </el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.users }}</div>
              <div class="stat-label">用户总数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon role" :size="40">
              <component :is="'Avatar'" />
            </el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.roles }}</div>
              <div class="stat-label">角色总数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon permission" :size="40">
              <component :is="'Lock'" />
            </el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.permissions }}</div>
              <div class="stat-label">权限总数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon task" :size="40">
              <component :is="'Calendar'" />
            </el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.tasks }}</div>
              <div class="stat-label">任务总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 快捷操作 -->
      <el-col :span="24" :md="12">
        <el-card>
          <template #header>
            <span>快捷操作</span>
          </template>
          <div class="quick-actions">
            <el-button type="primary" :icon="'Plus'" @click="handleAction('user')">
              创建用户
            </el-button>
            <el-button type="success" :icon="'Plus'" @click="handleAction('role')">
              创建角色
            </el-button>
            <el-button type="warning" :icon="'Plus'" @click="handleAction('task')">
              创建任务
            </el-button>
            <el-button type="info" :icon="'Document'" @click="handleAction('logs')">
              查看日志
            </el-button>
          </div>
        </el-card>
      </el-col>

      <!-- 系统信息 -->
      <el-col :span="24" :md="12">
        <el-card>
          <template #header>
            <span>系统信息</span>
          </template>
          <div class="system-info">
            <div class="info-item">
              <span class="label">系统版本：</span>
              <span class="value">Nova v1.0.0</span>
            </div>
            <div class="info-item">
              <span class="label">后端框架：</span>
              <span class="value">Go + Echo</span>
            </div>
            <div class="info-item">
              <span class="label">前端框架：</span>
              <span class="value">Vue 3 + TypeScript</span>
            </div>
            <div class="info-item">
              <span class="label">数据库：</span>
              <span class="value">PostgreSQL 16</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

// 当前日期
const currentDate = ref(new Date().toLocaleDateString('zh-CN', {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
  weekday: 'long'
}))

// 统计数据
const stats = ref({
  users: 0,
  roles: 0,
  permissions: 0,
  tasks: 0
})

// 加载统计数据
const loadStats = async () => {
  try {
    // TODO: 调用实际的统计API
    stats.value = {
      users: 156,
      roles: 8,
      permissions: 45,
      tasks: 23
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// 快捷操作
const handleAction = (action: string) => {
  switch (action) {
    case 'user':
      router.push('/system/user')
      break
    case 'role':
      router.push('/system/role')
      break
    case 'task':
      ElMessage.info('任务管理功能开发中...')
      break
    case 'logs':
      ElMessage.info('日志管理功能开发中...')
      break
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped lang="scss">
.home-container {
  width: 100%;
  min-height: 100%;
  background: #f0f2f5;
  padding: 20px;

  :deep(.el-card) {
    border-radius: 8px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    margin-bottom: 20px;
  }

  .welcome-card {
    .welcome-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      .welcome-text {
        h2 {
          margin: 0 0 10px 0;
          color: #303133;
          font-size: 24px;
        }
        
        p {
          margin: 0;
          color: #909399;
          font-size: 14px;
        }
      }
      
      .welcome-icon {
        color: #409EFF;
        opacity: 0.3;
      }
    }
  }
  
  .stat-card {
    transition: transform 0.3s, box-shadow 0.3s;

    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.15);
    }

    .stat-content {
      display: flex;
      align-items: center;
      gap: 20px;
      
      .stat-icon {
        flex-shrink: 0;
        padding: 15px;
        border-radius: 8px;
        
        &.user {
          background-color: #ecf5ff;
          color: #409EFF;
        }
        
        &.role {
          background-color: #f0f9ff;
          color: #67C23A;
        }
        
        &.permission {
          background-color: #fef0f0;
          color: #F56C6C;
        }
        
        &.task {
          background-color: #fdf6ec;
          color: #E6A23C;
        }
      }
      
      .stat-info {
        flex: 1;
        
        .stat-value {
          font-size: 28px;
          font-weight: bold;
          color: #303133;
          line-height: 1;
          margin-bottom: 8px;
        }
        
        .stat-label {
          font-size: 14px;
          color: #909399;
        }
      }
    }
  }
  
  .quick-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
  
  .system-info {
    .info-item {
      padding: 12px 0;
      border-bottom: 1px solid #f0f0f0;
      
      &:last-child {
        border-bottom: none;
      }
      
      .label {
        color: #909399;
        margin-right: 10px;
      }
      
      .value {
        color: #303133;
        font-weight: 500;
      }
    }
  }
}

// 暗色主题适配
html.dark {
  .home-container {
    background: #141414;

    :deep(.el-card) {
      background: #1f1f1f;
      border-color: #333;
      box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.3);
    }

    .welcome-card {
      .welcome-content {
        .welcome-text {
          h2 {
            color: #fff;
          }

          p {
            color: #999;
          }
        }
      }
    }

    .stat-card {
      &:hover {
        box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.4);
      }

      .stat-content {
        .stat-icon {
          &.user {
            background-color: rgba(64, 158, 255, 0.2);
          }
          
          &.role {
            background-color: rgba(103, 194, 58, 0.2);
          }
          
          &.permission {
            background-color: rgba(245, 108, 108, 0.2);
          }
          
          &.task {
            background-color: rgba(230, 162, 60, 0.2);
          }
        }

        .stat-info {
          .stat-value {
            color: #fff;
          }

          .stat-label {
            color: #999;
          }
        }
      }
    }

    .system-info {
      .info-item {
        border-bottom-color: #333;

        .label {
          color: #999;
        }

        .value {
          color: #fff;
        }
      }
    }

    :deep(.el-button--default) {
      background: #262626;
      border-color: #333;
      color: #fff;

      &:hover {
        background: #333;
        border-color: #444;
      }
    }
  }
}
</style>
