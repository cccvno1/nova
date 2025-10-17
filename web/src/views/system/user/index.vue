<template>
  <div class="user-manage-container">
    <el-card shadow="never">
      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="用户名">
          <el-input
            v-model="searchForm.username"
            placeholder="请输入用户名"
            clearable
            @clear="handleSearch"
          />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input
            v-model="searchForm.email"
            placeholder="请输入邮箱"
            clearable
            @clear="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="请选择状态"
            clearable
            @clear="handleSearch"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">
            搜索
          </el-button>
          <el-button :icon="Refresh" @click="handleReset">
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 操作按钮 -->
      <div class="table-toolbar">
        <el-button type="primary" :icon="Plus" @click="handleAdd">
          新增用户
        </el-button>
      </div>

      <!-- 用户表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="2"
              @change="handleStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              :icon="Edit"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-popconfirm
              title="确定删除该用户吗？"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="handleDelete(row.id)"
            >
              <template #reference>
                <el-button type="danger" link :icon="Delete">
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        class="pagination"
        @size-change="handleSearch"
        @current-change="handleSearch"
      />
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="formData.username"
            placeholder="请输入用户名"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="formData.email"
            placeholder="请输入邮箱"
            type="email"
          />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input
            v-model="formData.password"
            placeholder="请输入密码"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input
            v-model="formData.nickname"
            placeholder="请输入昵称"
          />
        </el-form-item>
        <el-form-item label="头像" prop="avatar">
          <el-input
            v-model="formData.avatar"
            placeholder="请输入头像URL"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Search, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { userApi } from '@/api/user'
import type { User, CreateUserDTO, UpdateUserDTO, UserListParams } from '@/types/api'

// 定义组件名称（用于 keep-alive）
defineOptions({
  name: 'UserManage'
})

// 搜索表单
const searchForm = reactive<UserListParams>({
  username: '',
  email: '',
  status: undefined,
  page: 1,
  page_size: 10
})

// 分页数据
const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

// 表格数据
const tableData = ref<User[]>([])
const loading = ref(false)

// 对话框相关
const dialogVisible = ref(false)
const dialogTitle = ref('新增用户')
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

// 表单数据
const formData = reactive<CreateUserDTO & UpdateUserDTO & { id?: number }>({
  username: '',
  email: '',
  password: '',
  nickname: '',
  avatar: ''
})

// 表单验证规则
const formRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度为 3-50 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 50, message: '密码长度为 6-50 个字符', trigger: 'blur' }
  ],
  nickname: [
    { max: 50, message: '昵称长度不能超过 50 个字符', trigger: 'blur' }
  ]
}

/**
 * 获取用户列表
 */
const getUserList = async () => {
  loading.value = true
  try {
    const params: UserListParams = {
      ...searchForm,
      page: pagination.page,
      page_size: pagination.page_size
    }
    
    const data = await userApi.list(params)
    tableData.value = data.items
    pagination.total = data.total
    pagination.page = data.page
    pagination.page_size = data.page_size
  } catch (error) {
    console.error('获取用户列表失败:', error)
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

/**
 * 搜索
 */
const handleSearch = () => {
  pagination.page = 1
  getUserList()
}

/**
 * 重置搜索
 */
const handleReset = () => {
  searchForm.username = ''
  searchForm.email = ''
  searchForm.status = undefined
  handleSearch()
}

/**
 * 新增用户
 */
const handleAdd = () => {
  isEdit.value = false
  dialogTitle.value = '新增用户'
  dialogVisible.value = true
}

/**
 * 编辑用户
 */
const handleEdit = (row: User) => {
  isEdit.value = true
  dialogTitle.value = '编辑用户'
  formData.id = row.id
  formData.username = row.username
  formData.email = row.email
  formData.nickname = row.nickname
  formData.avatar = row.avatar
  dialogVisible.value = true
}

/**
 * 删除用户
 */
const handleDelete = async (id: number) => {
  try {
    await userApi.delete(id)
    ElMessage.success('删除成功')
    getUserList()
  } catch (error) {
    console.error('删除用户失败:', error)
    ElMessage.error('删除用户失败')
  }
}

/**
 * 状态切换
 */
const handleStatusChange = async (row: User) => {
  try {
    await userApi.update(row.id, { status: row.status })
    ElMessage.success('状态修改成功')
  } catch (error) {
    console.error('修改状态失败:', error)
    ElMessage.error('修改状态失败')
    // 失败时恢复原状态
    row.status = row.status === 1 ? 2 : 1
  }
}

/**
 * 提交表单
 */
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      if (isEdit.value && formData.id) {
        // 编辑用户
        const updateData: UpdateUserDTO = {
          email: formData.email,
          nickname: formData.nickname,
          avatar: formData.avatar
        }
        await userApi.update(formData.id, updateData)
        ElMessage.success('编辑成功')
      } else {
        // 新增用户
        const createData: CreateUserDTO = {
          username: formData.username,
          email: formData.email,
          password: formData.password,
          nickname: formData.nickname,
          avatar: formData.avatar
        }
        await userApi.create(createData)
        ElMessage.success('新增成功')
      }
      dialogVisible.value = false
      getUserList()
    } catch (error) {
      console.error('提交失败:', error)
      ElMessage.error(isEdit.value ? '编辑失败' : '新增失败')
    } finally {
      submitLoading.value = false
    }
  })
}

/**
 * 对话框关闭
 */
const handleDialogClose = () => {
  formRef.value?.resetFields()
  formData.id = undefined
  formData.username = ''
  formData.email = ''
  formData.password = ''
  formData.nickname = ''
  formData.avatar = ''
}

/**
 * 格式化日期
 */
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 初始化
onMounted(() => {
  getUserList()
})
</script>

<style scoped lang="scss">
.user-manage-container {
  width: 100%;
  height: 100%;
  background: #f0f2f5;
  padding: 20px;

  :deep(.el-card) {
    border-radius: 8px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  }

  .search-form {
    margin-bottom: 16px;

    :deep(.el-form-item) {
      margin-bottom: 0;
    }
  }

  .table-toolbar {
    margin-bottom: 16px;
  }

  :deep(.el-table) {
    border-radius: 4px;

    th {
      background: #fafafa;
      color: #333;
      font-weight: 600;
    }

    .el-button {
      padding: 4px 8px;
      font-size: 13px;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}

// 暗色主题适配
html.dark {
  .user-manage-container {
    background: #141414;

    :deep(.el-card) {
      background: #1f1f1f;
      border-color: #333;
      box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.3);
    }

    :deep(.el-table) {
      background: #1f1f1f;
      color: #fff;

      th {
        background: #262626;
        color: #fff;
      }

      td {
        background: #1f1f1f;
        border-color: #333;
      }

      tr:hover > td {
        background: #262626 !important;
      }
    }

    :deep(.el-input__wrapper),
    :deep(.el-select .el-input__wrapper) {
      background: #1f1f1f;
      box-shadow: 0 0 0 1px #333 inset;
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
