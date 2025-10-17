<template>
  <div class="role-manage-container">
    <el-card shadow="never">
      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="角色名称">
          <el-input
            v-model="searchForm.keyword"
            placeholder="请输入角色名称"
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
        <el-button type="primary" :icon="Plus" @click="handleAdd">
          新增角色
        </el-button>
      </div>

      <!-- 角色表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="name" label="角色标识" min-width="120" />
        <el-table-column prop="display_name" label="角色名称" min-width="120" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              :icon="Setting"
              @click="handlePermission(row)"
            >
              权限配置
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
              title="确定删除该角色吗？"
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
        <el-form-item label="角色标识" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入角色标识，如：admin"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="角色名称" prop="display_name">
          <el-input
            v-model="formData.display_name"
            placeholder="请输入角色名称，如：管理员"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入角色描述"
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

    <!-- 权限配置对话框 -->
    <el-dialog
      v-model="permissionDialogVisible"
      title="权限配置"
      width="700px"
      @close="handlePermissionDialogClose"
    >
      <div v-loading="permissionLoading">
        <!-- 权限树 -->
        <div style="max-height: 400px; overflow-y: auto; border: 1px solid var(--el-border-color); border-radius: 4px; padding: 10px;">
          <el-tree
            ref="permissionTreeRef"
            :data="permissionTreeData"
            :props="treeProps"
            node-key="id"
            show-checkbox
            default-expand-all
            :default-checked-keys="checkedPermissions"
            @check="handlePermissionCheck"
          >
            <template #default="{ node, data }">
              <span class="tree-node">
                <el-icon v-if="data.icon" style="margin-right: 5px;">
                  <component :is="data.icon" />
                </el-icon>
                <span>{{ node.label }}</span>
                <el-tag v-if="data.type" size="small" style="margin-left: 8px;">
                  {{ getPermissionTypeLabel(data.type) }}
                </el-tag>
              </span>
            </template>
          </el-tree>
        </div>

        <!-- 变更预览 -->
        <el-card v-if="permissionDiff" shadow="never" style="margin-top: 15px; background-color: var(--el-fill-color-light);">
          <template #header>
            <div style="display: flex; align-items: center;">
              <el-icon style="margin-right: 5px;"><InfoFilled /></el-icon>
              <span>变更预览</span>
            </div>
          </template>
          
          <el-empty v-if="!permissionDiff.added?.length && !permissionDiff.removed?.length" 
                    description="没有任何变更" 
                    :image-size="60" />
          
          <div v-else>
            <!-- 新增的权限 -->
            <div v-if="permissionDiff.added && permissionDiff.added.length > 0" style="margin-bottom: 15px;">
              <div style="margin-bottom: 8px;">
                <el-tag type="success" size="small">
                  ➕ 将添加 {{ permissionDiff.added.length }} 个权限
                </el-tag>
              </div>
              <ul style="margin: 0; padding-left: 20px; line-height: 1.8;">
                <li v-for="p in permissionDiff.added" :key="p.id" style="color: var(--el-color-success);">
                  {{ p.display_name }} <el-tag size="small" style="margin-left: 5px;">{{ getPermissionTypeLabel(p.type) }}</el-tag>
                </li>
              </ul>
            </div>
            
            <!-- 删除的权限 -->
            <div v-if="permissionDiff.removed && permissionDiff.removed.length > 0">
              <div style="margin-bottom: 8px;">
                <el-tag type="danger" size="small">
                  ➖ 将删除 {{ permissionDiff.removed.length }} 个权限
                </el-tag>
              </div>
              <ul style="margin: 0; padding-left: 20px; line-height: 1.8;">
                <li v-for="p in permissionDiff.removed" :key="p.id" style="color: var(--el-color-danger);">
                  {{ p.display_name }} <el-tag size="small" style="margin-left: 5px;">{{ getPermissionTypeLabel(p.type) }}</el-tag>
                </li>
              </ul>
            </div>
          </div>
        </el-card>
      </div>
      
      <template #footer>
        <el-button @click="() => { console.log('预览按钮被点击'); handlePreviewPermissions(); }" :loading="previewLoading">
          <el-icon style="margin-right: 5px;"><View /></el-icon>
          预览变更
        </el-button>
        <el-button @click="() => { console.log('取消按钮被点击'); permissionDialogVisible = false; }">取消</el-button>
        <el-button 
          type="primary" 
          :loading="permissionSubmitLoading" 
          @click="() => { console.log('确定按钮被点击'); handlePermissionSubmit(); }"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type ElTree } from 'element-plus'
import { Search, Refresh, Plus, Edit, Delete, Setting, View, InfoFilled } from '@element-plus/icons-vue'
import { roleApi } from '@/api/role'
import { permissionApi } from '@/api/permission'
import type { Role, CreateRoleDTO, UpdateRoleDTO, RoleSearchParams, Permission } from '@/types/api'

// 定义组件名称（用于 keep-alive）
defineOptions({
  name: 'RoleManage'
})

// 搜索表单
const searchForm = reactive<RoleSearchParams>({
  keyword: '',
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
const tableData = ref<Role[]>([])
const loading = ref(false)

// 对话框相关
const dialogVisible = ref(false)
const dialogTitle = ref('新增角色')
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

// 表单数据
const formData = reactive<CreateRoleDTO & UpdateRoleDTO & { id?: number }>({
  name: '',
  display_name: '',
  description: ''
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入角色标识', trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: '角色标识只能包含字母、数字和下划线，且不能以数字开头', trigger: 'blur' }
  ],
  display_name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' },
    { max: 50, message: '角色名称长度不能超过 50 个字符', trigger: 'blur' }
  ]
}

// 权限配置对话框
const permissionDialogVisible = ref(false)
const permissionLoading = ref(false)
const permissionSubmitLoading = ref(false)
const previewLoading = ref(false)
const permissionTreeRef = ref<InstanceType<typeof ElTree>>()
const permissionTreeData = ref<Permission[]>([])
const checkedPermissions = ref<number[]>([])
const currentRoleId = ref<number>(0)
const permissionDiff = ref<{
  added: Permission[]
  removed: Permission[]
  kept: Permission[]
} | null>(null)

// 树形控件配置
const treeProps = {
  label: 'display_name',
  children: 'children'
}

/**
 * 获取角色列表
 */
const getRoleList = async () => {
  loading.value = true
  try {
    const params: RoleSearchParams = {
      ...searchForm,
      page: pagination.page,
      page_size: pagination.page_size
    }
    
    const data = await roleApi.search(params)
    tableData.value = data.items
    pagination.total = data.total
    pagination.page = data.page
    pagination.page_size = data.page_size
  } catch (error) {
    console.error('获取角色列表失败:', error)
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}

/**
 * 搜索
 */
const handleSearch = () => {
  pagination.page = 1
  getRoleList()
}

/**
 * 重置搜索
 */
const handleReset = () => {
  searchForm.keyword = ''
  handleSearch()
}

/**
 * 新增角色
 */
const handleAdd = () => {
  isEdit.value = false
  dialogTitle.value = '新增角色'
  dialogVisible.value = true
}

/**
 * 编辑角色
 */
const handleEdit = (row: Role) => {
  isEdit.value = true
  dialogTitle.value = '编辑角色'
  formData.id = row.id
  formData.name = row.name
  formData.display_name = row.display_name
  formData.description = row.description
  dialogVisible.value = true
}

/**
 * 删除角色
 */
const handleDelete = async (id: number) => {
  try {
    await roleApi.delete(id)
    ElMessage.success('删除成功')
    getRoleList()
  } catch (error) {
    console.error('删除角色失败:', error)
    ElMessage.error('删除角色失败')
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
        // 编辑角色
        const updateData: UpdateRoleDTO = {
          display_name: formData.display_name,
          description: formData.description
        }
        await roleApi.update(formData.id, updateData)
        ElMessage.success('编辑成功')
      } else {
        // 新增角色
        const createData: CreateRoleDTO = {
          name: formData.name,
          display_name: formData.display_name,
          description: formData.description
        }
        await roleApi.create(createData)
        ElMessage.success('新增成功')
      }
      dialogVisible.value = false
      getRoleList()
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
  formData.description = ''
}

/**
 * 权限配置
 */
const handlePermission = async (row: Role) => {
  currentRoleId.value = row.id
  permissionDialogVisible.value = true
  
  // 加载权限树和已分配的权限
  await loadPermissionTree()
  await loadRolePermissions(row.id)
}

/**
 * 加载权限树
 */
const loadPermissionTree = async () => {
  permissionLoading.value = true
  try {
    const data = await permissionApi.getTree()
    permissionTreeData.value = data
  } catch (error) {
    console.error('获取权限树失败:', error)
    ElMessage.error('获取权限树失败')
  } finally {
    permissionLoading.value = false
  }
}

/**
 * 加载角色已分配的权限
 */
const loadRolePermissions = async (roleId: number) => {
  try {
    const permissions = await roleApi.getPermissions(roleId)
    checkedPermissions.value = permissions.map(p => p.id)
  } catch (error) {
    console.error('获取角色权限失败:', error)
    ElMessage.error('获取角色权限失败')
  }
}

/**
 * 权限树复选框变化时清除预览
 */
const handlePermissionCheck = () => {
  permissionDiff.value = null
}

/**
 * 预览权限变更
 */
const handlePreviewPermissions = async () => {
  if (!permissionTreeRef.value) return

  previewLoading.value = true
  try {
    const checkedKeys = permissionTreeRef.value.getCheckedKeys() as number[]
    const halfCheckedKeys = permissionTreeRef.value.getHalfCheckedKeys() as number[]
    const allKeys = [...checkedKeys, ...halfCheckedKeys]

    console.log('发送预览请求，权限IDs:', allKeys)
    const response = await roleApi.updatePermissions(currentRoleId.value, {
      permission_ids: allKeys,
      preview: true
    })

    console.log('收到预览响应:', response)

    // 后端直接返回差异对象（added/removed/kept），不是嵌套在preview字段中
    if (response && (response.added || response.removed || response.kept)) {
      permissionDiff.value = response as any
      console.log('设置permissionDiff:', permissionDiff.value)
      
      // 如果有删除操作，额外提示
      if (response.removed && response.removed.length > 0) {
        ElMessage.warning({
          message: `⚠️ 将删除 ${response.removed.length} 个权限，请仔细确认！`,
          duration: 3000
        })
      }
      
      // 如果有添加操作，提示
      if (response.added && response.added.length > 0) {
        ElMessage.success({
          message: `✅ 将添加 ${response.added.length} 个权限`,
          duration: 3000
        })
      }
      
      // 如果没有任何变更
      if ((!response.added || response.added.length === 0) && 
          (!response.removed || response.removed.length === 0)) {
        ElMessage.info('没有任何变更')
      }
    } else {
      console.warn('响应格式异常:', response)
    }
  } catch (error) {
    console.error('预览权限变更失败:', error)
    ElMessage.error('预览权限变更失败')
  } finally {
    previewLoading.value = false
  }
}

/**
 * 提交权限配置
 */
const handlePermissionSubmit = async () => {
  if (!permissionTreeRef.value) return

  permissionSubmitLoading.value = true
  try {
    // 获取选中的权限ID（包括半选中的父节点）
    const checkedKeys = permissionTreeRef.value.getCheckedKeys() as number[]
    const halfCheckedKeys = permissionTreeRef.value.getHalfCheckedKeys() as number[]
    const allKeys = [...checkedKeys, ...halfCheckedKeys]

    // 如果没有预览过，先自动预览获取差异
    if (!permissionDiff.value) {
      const response = await roleApi.updatePermissions(currentRoleId.value, {
        permission_ids: allKeys,
        preview: true
      })
      // 后端直接返回差异对象
      permissionDiff.value = response || null
    }

    // 如果有删除操作，二次确认
    if (permissionDiff.value && permissionDiff.value.removed && permissionDiff.value.removed.length > 0) {
      try {
        await ElMessageBox.confirm(
          `将删除 ${permissionDiff.value.removed.length} 个权限，此操作不可撤销，确定继续吗？`,
          '删除权限确认',
          {
            type: 'warning',
            confirmButtonText: '确定删除',
            cancelButtonText: '取消'
          }
        )
      } catch {
        permissionSubmitLoading.value = false
        return // 用户取消
      }
    }

    // 执行更新
    const response = await roleApi.updatePermissions(currentRoleId.value, {
      permission_ids: allKeys,
      preview: false
    })

    // 后端直接返回结果对象（added_count/removed_count）
    if (response && (response.added_count !== undefined || response.removed_count !== undefined)) {
      const messages = []
      if (response.added_count > 0) {
        messages.push(`添加 ${response.added_count} 个权限`)
      }
      if (response.removed_count > 0) {
        messages.push(`删除 ${response.removed_count} 个权限`)
      }
      if (messages.length === 0) {
        messages.push('没有任何变更')
      }
      ElMessage.success(messages.join('，'))
    } else {
      ElMessage.success('权限配置成功')
    }
    
    permissionDialogVisible.value = false
  } catch (error) {
    console.error('配置权限失败:', error)
    ElMessage.error('配置权限失败')
  } finally {
    permissionSubmitLoading.value = false
  }
}

/**
 * 权限配置对话框关闭
 */
const handlePermissionDialogClose = () => {
  checkedPermissions.value = []
  permissionTreeData.value = []
  permissionDiff.value = null // 清除预览数据
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
  getRoleList()
})
</script>

<style scoped lang="scss">
.role-manage-container {
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

  .tree-node {
    display: flex;
    align-items: center;
  }
}

// 暗色主题适配
html.dark {
  .role-manage-container {
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

    :deep(.el-tree) {
      background: #1f1f1f;
      color: #fff;

      .el-tree-node__content {
        &:hover {
          background: #333;
        }
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
