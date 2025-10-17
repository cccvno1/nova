<template>
  <div class="permission-manage-container">
    <el-card shadow="never">
      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="权限类型">
          <el-select
            v-model="searchForm.type"
            placeholder="请选择权限类型"
            clearable
            @change="handleSearch"
          >
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
            <el-option label="接口" value="api" />
          </el-select>
        </el-form-item>
        <el-form-item label="域">
          <el-input
            v-model="searchForm.domain"
            placeholder="请输入域"
            clearable
            @clear="handleSearch"
          />
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
        <el-button type="primary" :icon="Plus" @click="handleAdd(null)">
          新增根权限
        </el-button>
        <el-button :icon="Expand" @click="expandAll = true">
          展开全部
        </el-button>
        <el-button :icon="Fold" @click="expandAll = false">
          折叠全部
        </el-button>
      </div>

      <!-- 权限表格（树形） -->
      <el-table
        v-loading="loading"
        :data="tableData"
        row-key="id"
        :tree-props="{ children: 'children' }"
        :default-expand-all="expandAll"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="display_name" label="权限名称" min-width="200">
          <template #default="{ row }">
            <el-icon v-if="row.icon" style="margin-right: 5px;">
              <component :is="row.icon" />
            </el-icon>
            <span>{{ row.display_name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="权限标识" min-width="180" />
        <el-table-column prop="type" label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">
              {{ getPermissionTypeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由/路径" min-width="150" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="domain" label="域" width="100" />
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              type="success"
              link
              :icon="Plus"
              @click="handleAdd(row)"
            >
              添加子权限
            </el-button>
            <el-button
              type="primary"
              link
              :icon="Edit"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-popconfirm
              title="确定删除该权限吗？"
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
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="700px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item label="父级权限">
          <el-input
            :value="parentPermissionName"
            disabled
            placeholder="无（根权限）"
          />
        </el-form-item>
        <el-form-item label="权限类型" prop="type">
          <el-radio-group v-model="formData.type">
            <el-radio value="menu">菜单</el-radio>
            <el-radio value="button">按钮</el-radio>
            <el-radio value="api">接口</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="权限标识" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="如：menu:system:user"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="权限名称" prop="display_name">
          <el-input
            v-model="formData.display_name"
            placeholder="如：用户管理"
          />
        </el-form-item>
        <el-form-item v-if="formData.type === 'menu'" label="路由路径" prop="path">
          <el-input
            v-model="formData.path"
            placeholder="如：/system/user"
          />
        </el-form-item>
        <el-form-item v-if="formData.type === 'api'" label="API路径" prop="path">
          <el-input
            v-model="formData.path"
            placeholder="如：/api/v1/users"
          />
        </el-form-item>
        <el-form-item v-if="formData.type === 'menu'" label="组件路径" prop="component">
          <el-input
            v-model="formData.component"
            placeholder="如：views/system/user/index 或 Layout"
          />
        </el-form-item>
        <el-form-item v-if="formData.type === 'api'" label="请求方法" prop="method">
          <el-select v-model="formData.method" placeholder="请选择请求方法">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="formData.type === 'menu'" label="图标" prop="icon">
          <el-input
            v-model="formData.icon"
            placeholder="如：User"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number
            v-model="formData.sort"
            :min="0"
            :max="9999"
          />
        </el-form-item>
        <el-form-item label="域" prop="domain">
          <el-input
            v-model="formData.domain"
            placeholder="如：system"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入权限描述"
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Search, Refresh, Plus, Edit, Delete, Expand, Fold } from '@element-plus/icons-vue'
import { permissionApi } from '@/api/permission'
import type { Permission, CreatePermissionDTO, UpdatePermissionDTO } from '@/types/api'

// 定义组件名称（用于 keep-alive）
defineOptions({
  name: 'PermissionManage'
})

// 搜索表单
const searchForm = reactive({
  type: undefined as string | undefined,
  domain: ''
})

// 表格数据
const tableData = ref<Permission[]>([])
const loading = ref(false)
const expandAll = ref(false)

// 对话框相关
const dialogVisible = ref(false)
const dialogTitle = ref('新增权限')
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const parentPermissionId = ref<number | null>(null)
const parentPermissionName = ref('')

// 表单数据
const formData = reactive<CreatePermissionDTO & UpdatePermissionDTO & { id?: number }>({
  name: '',
  display_name: '',
  type: 'menu',
  path: '',
  component: '',
  method: '',
  icon: '',
  sort: 0,
  domain: '',
  description: '',
  parent_id: null
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入权限标识', trigger: 'blur' },
    { pattern: /^[a-zA-Z_:][a-zA-Z0-9_:]*$/, message: '权限标识只能包含字母、数字、下划线和冒号', trigger: 'blur' }
  ],
  display_name: [
    { required: true, message: '请输入权限名称', trigger: 'blur' },
    { max: 50, message: '权限名称长度不能超过 50 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择权限类型', trigger: 'change' }
  ],
  path: [
    { required: true, message: '请输入路径', trigger: 'blur' }
  ]
}

/**
 * 获取权限树
 */
const getPermissionTree = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (searchForm.domain) {
      params.domain = searchForm.domain
    }
    
    let data: Permission[]
    if (searchForm.type) {
      data = await permissionApi.listByType(searchForm.type as any, searchForm.domain || undefined)
    } else {
      data = await permissionApi.getTree(params)
    }
    
    tableData.value = data
  } catch (error) {
    console.error('获取权限树失败:', error)
    ElMessage.error('获取权限树失败')
  } finally {
    loading.value = false
  }
}

/**
 * 搜索
 */
const handleSearch = () => {
  getPermissionTree()
}

/**
 * 重置搜索
 */
const handleReset = () => {
  searchForm.type = undefined
  searchForm.domain = ''
  handleSearch()
}

/**
 * 新增权限
 */
const handleAdd = (parent: Permission | null) => {
  isEdit.value = false
  dialogTitle.value = parent ? '新增子权限' : '新增根权限'
  parentPermissionId.value = parent ? parent.id : null
  parentPermissionName.value = parent ? parent.display_name : ''
  formData.parent_id = parent ? parent.id : null
  dialogVisible.value = true
}

/**
 * 编辑权限
 */
const handleEdit = (row: Permission) => {
  isEdit.value = true
  dialogTitle.value = '编辑权限'
  formData.id = row.id
  formData.name = row.name
  formData.display_name = row.display_name
  formData.type = row.type
  formData.path = row.path || ''
  formData.component = row.component || ''
  formData.method = row.method || ''
  formData.icon = row.icon || ''
  formData.sort = row.sort
  formData.domain = row.domain || ''
  formData.description = row.description || ''
  formData.parent_id = row.parent_id
  
  // 查找父权限名称
  if (row.parent_id) {
    const findParent = (permissions: Permission[]): string => {
      for (const perm of permissions) {
        if (perm.id === row.parent_id) {
          return perm.display_name
        }
        if (perm.children) {
          const found = findParent(perm.children)
          if (found) return found
        }
      }
      return ''
    }
    parentPermissionName.value = findParent(tableData.value)
  } else {
    parentPermissionName.value = ''
  }
  
  dialogVisible.value = true
}

/**
 * 删除权限
 */
const handleDelete = async (id: number) => {
  try {
    await permissionApi.delete(id)
    ElMessage.success('删除成功')
    getPermissionTree()
  } catch (error) {
    console.error('删除权限失败:', error)
    ElMessage.error('删除权限失败')
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
        // 编辑权限
        const updateData: UpdatePermissionDTO = {
          display_name: formData.display_name,
          path: formData.path || undefined,
          component: formData.component || undefined,
          method: formData.method || undefined,
          icon: formData.icon || undefined,
          sort: formData.sort,
          domain: formData.domain || undefined,
          description: formData.description || undefined
        }
        await permissionApi.update(formData.id, updateData)
        ElMessage.success('编辑成功')
      } else {
        // 新增权限
        const createData: CreatePermissionDTO = {
          name: formData.name,
          display_name: formData.display_name,
          type: formData.type,
          path: formData.path || undefined,
          component: formData.component || undefined,
          method: formData.method || undefined,
          icon: formData.icon || undefined,
          sort: formData.sort,
          domain: formData.domain || undefined,
          description: formData.description || undefined,
          parent_id: formData.parent_id
        }
        await permissionApi.create(createData)
        ElMessage.success('新增成功')
      }
      dialogVisible.value = false
      getPermissionTree()
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
  formData.name = ''
  formData.display_name = ''
  formData.type = 'menu'
  formData.path = ''
  formData.component = ''
  formData.method = ''
  formData.icon = ''
  formData.sort = 0
  formData.domain = ''
  formData.description = ''
  formData.parent_id = null
  parentPermissionId.value = null
  parentPermissionName.value = ''
}

/**
 * 获取权限类型标签
 */
const getPermissionTypeLabel = (type: string) => {
  const typeMap: Record<string, string> = {
    menu: '菜单',
    button: '按钮',
    api: '接口'
  }
  return typeMap[type] || type
}

/**
 * 获取类型标签颜色
 */
const getTypeTagType = (type: string) => {
  const typeMap: Record<string, any> = {
    menu: 'primary',
    button: 'success',
    api: 'warning'
  }
  return typeMap[type] || 'info'
}

// 初始化
onMounted(() => {
  getPermissionTree()
})
</script>

<style scoped lang="scss">
.permission-manage-container {
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
}

// 暗色主题适配
html.dark {
  .permission-manage-container {
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
