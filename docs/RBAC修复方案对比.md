# RBAC 权限分配修复方案对比

## 🎯 核心问题

**现状：** 权限配置使用 `Association.Replace()` 全量替换，导致未选中的权限被误删

---

## 📋 三种修复方案对比

| 方案 | 前端改动 | 后端改动 | API数量 | 开发时间 | 推荐度 |
|------|----------|----------|---------|----------|--------|
| **A. 增强现有API** | 添加确认框 | 添加日志+警告 | 0个新增 | 2小时 | ⭐⭐⭐ |
| **B. 增量API（推荐）** | 支持双模式 | 新增add/remove | 2个新增 | 4小时 | ⭐⭐⭐⭐⭐ |
| **C. Diff API（最佳）** | 预览+确认 | 智能Diff计算 | 1个新增 | 3小时 | ⭐⭐⭐⭐⭐ |

---

## 方案A：增强现有API（最小改动）

### 后端改动

```go
// internal/service/rbac_service.go

func (s *rbacService) AssignPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint) error {
    // 1. 获取变更前的权限
    var oldPermissions []Permission
    s.db.Model(&Role{ID: roleID}).Association("Permissions").Find(&oldPermissions)
    
    oldIDs := make(map[uint]bool)
    for _, p := range oldPermissions {
        oldIDs[p.ID] = true
    }
    
    // 2. 计算将被删除的权限
    var willRemove []Permission
    for _, p := range oldPermissions {
        found := false
        for _, newID := range permissionIDs {
            if p.ID == newID {
                found = true
                break
            }
        }
        if !found {
            willRemove = append(willRemove, p)
        }
    }
    
    // 3. 记录警告日志
    if len(willRemove) > 0 {
        s.logger.Warn("permissions will be removed",
            "role_id", roleID,
            "remove_count", len(willRemove),
            "removed_permissions", willRemove,
        )
    }
    
    // 4. 执行替换
    permissions, _ := s.permRepo.ListByIDs(ctx, permissionIDs)
    if err := s.db.Model(&Role{ID: roleID}).Association("Permissions").Replace(permissions); err != nil {
        return err
    }
    
    // 5. 记录审计日志
    s.auditService.Log(ctx, &AuditLog{
        Action:    "assign_permissions",
        TargetType: "role",
        TargetID:  roleID,
        OldValue:  oldIDs,
        NewValue:  permissionIDs,
    })
    
    return nil
}
```

### 前端改动

```vue
<!-- web/src/views/system/role/index.vue -->

<script setup>
const handlePermissionSubmit = async () => {
  const allKeys = [...checkedKeys, ...halfCheckedKeys]
  
  // ⚠️ 二次确认
  const { data: current } = await roleApi.getRolePermissions(currentRoleId.value)
  const willRemove = current.permissions.filter(p => !allKeys.includes(p.id))
  
  if (willRemove.length > 0) {
    await ElMessageBox.confirm(
      `⚠️ 将删除以下 ${willRemove.length} 个权限：\n` +
      willRemove.map(p => `- ${p.display_name}`).join('\n') +
      `\n\n确定继续吗？`,
      '权限变更确认',
      { type: 'warning' }
    )
  }
  
  await roleApi.assignPermissions(currentRoleId.value, { permission_ids: allKeys })
}
</script>
```

**优点：** 改动最小，立即可用  
**缺点：** 仍是全量替换，只是加了保护

---

## 方案B：增量API（业界常用）

### 后端API设计（参考Django REST Framework）

```go
// 新增：批量添加权限
// POST /api/v1/roles/:id/permissions/add
type AddPermissionsRequest struct {
    PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

func (h *RoleHandler) AddPermissions(c echo.Context) error {
    roleID := c.Param("id")
    var req AddPermissionsRequest
    c.Bind(&req)
    
    // 使用 Append 而不是 Replace
    role := &Role{ID: roleID}
    permissions := s.permRepo.ListByIDs(req.PermissionIDs)
    
    s.db.Model(role).Association("Permissions").Append(permissions)
    
    return c.JSON(200, response.Success("成功添加权限"))
}

// 新增：批量删除权限
// POST /api/v1/roles/:id/permissions/remove
type RemovePermissionsRequest struct {
    PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

func (h *RoleHandler) RemovePermissions(c echo.Context) error {
    roleID := c.Param("id")
    var req RemovePermissionsRequest
    c.Bind(&req)
    
    // 使用 Delete
    role := &Role{ID: roleID}
    permissions := s.permRepo.ListByIDs(req.PermissionIDs)
    
    s.db.Model(role).Association("Permissions").Delete(permissions)
    
    return c.JSON(200, response.Success("成功删除权限"))
}

// 路由注册
roleGroup.POST("/:id/permissions/add", handler.AddPermissions)
roleGroup.POST("/:id/permissions/remove", handler.RemovePermissions)
```

### 前端改动（支持两种模式）

```vue
<!-- web/src/views/system/role/index.vue -->

<template>
  <el-dialog v-model="permissionDialogVisible" title="权限配置">
    <!-- 模式选择 -->
    <el-radio-group v-model="mode" style="margin-bottom: 20px">
      <el-radio-button value="incremental">增量修改</el-radio-button>
      <el-radio-button value="full">全量替换</el-radio-button>
    </el-radio-group>
    
    <el-tree
      ref="permissionTreeRef"
      :data="permissionTree"
      show-checkbox
      node-key="id"
      :default-checked-keys="originalKeys"
    />
    
    <template #footer>
      <el-button @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
const mode = ref('incremental') // 默认增量模式
const originalKeys = ref([]) // 原始选中的权限

const handleSubmit = async () => {
  const currentKeys = permissionTreeRef.value.getCheckedKeys()
  
  if (mode.value === 'incremental') {
    // 增量模式：计算差异
    const toAdd = currentKeys.filter(id => !originalKeys.value.includes(id))
    const toRemove = originalKeys.value.filter(id => !currentKeys.includes(id))
    
    if (toAdd.length > 0) {
      await roleApi.addPermissions(roleId, { permission_ids: toAdd })
    }
    if (toRemove.length > 0) {
      await roleApi.removePermissions(roleId, { permission_ids: toRemove })
    }
    
  } else {
    // 全量模式：直接替换
    await roleApi.assignPermissions(roleId, { permission_ids: currentKeys })
  }
}
</script>
```

### 前端API文件

```typescript
// web/src/api/role.ts

export const roleApi = {
  // 新增
  addPermissions(roleId: number, data: { permission_ids: number[] }) {
    return request.post(`/roles/${roleId}/permissions/add`, data)
  },
  
  // 新增
  removePermissions(roleId: number, data: { permission_ids: number[] }) {
    return request.post(`/roles/${roleId}/permissions/remove`, data)
  },
  
  // 保留（用于全量模式）
  assignPermissions(roleId: number, data: { permission_ids: number[] }) {
    return request.post(`/roles/${roleId}/permissions`, data)
  }
}
```

**优点：** API语义清晰，支持精确操作  
**缺点：** 需要2个新API

---

## 方案C：Diff API（最优雅）

### 后端API设计（参考GitHub PR Review）

```go
// POST /api/v1/roles/:id/permissions/update
type UpdatePermissionsRequest struct {
    PermissionIDs []uint  `json:"permission_ids" binding:"required"`
    AutoDiff      bool    `json:"auto_diff"`      // 自动计算差异
    Preview       bool    `json:"preview"`        // 仅预览，不执行
}

type UpdatePermissionsResponse struct {
    Preview *PermissionDiff `json:"preview,omitempty"`
    Result  *ChangeResult   `json:"result,omitempty"`
}

type PermissionDiff struct {
    Added   []Permission `json:"added"`
    Removed []Permission `json:"removed"`
    Kept    []Permission `json:"kept"`
}

func (h *RoleHandler) UpdatePermissions(c echo.Context) error {
    roleID := c.Param("id")
    var req UpdatePermissionsRequest
    c.Bind(&req)
    
    // 1. 获取当前权限
    current, _ := h.rbacService.GetRolePermissions(roleID)
    currentIDs := make(map[uint]bool)
    for _, p := range current {
        currentIDs[p.ID] = true
    }
    
    // 2. 计算差异
    newIDs := make(map[uint]bool)
    for _, id := range req.PermissionIDs {
        newIDs[id] = true
    }
    
    var toAdd, toRemove []uint
    for _, p := range current {
        if !newIDs[p.ID] {
            toRemove = append(toRemove, p.ID)
        }
    }
    for id := range newIDs {
        if !currentIDs[id] {
            toAdd = append(toAdd, id)
        }
    }
    
    // 3. 如果是预览模式，返回差异
    if req.Preview {
        added := h.permRepo.ListByIDs(toAdd)
        removed := h.permRepo.ListByIDs(toRemove)
        
        return c.JSON(200, response.Success(&UpdatePermissionsResponse{
            Preview: &PermissionDiff{
                Added:   added,
                Removed: removed,
            },
        }))
    }
    
    // 4. 执行变更（使用事务）
    err := h.db.Transaction(func(tx *gorm.DB) error {
        role := &Role{ID: roleID}
        
        if len(toRemove) > 0 {
            removePerms := h.permRepo.ListByIDs(toRemove)
            tx.Model(role).Association("Permissions").Delete(removePerms)
        }
        
        if len(toAdd) > 0 {
            addPerms := h.permRepo.ListByIDs(toAdd)
            tx.Model(role).Association("Permissions").Append(addPerms)
        }
        
        return nil
    })
    
    return c.JSON(200, response.Success(&UpdatePermissionsResponse{
        Result: &ChangeResult{
            AddedCount:   len(toAdd),
            RemovedCount: len(toRemove),
        },
    }))
}
```

### 前端改动（两步操作）

```vue
<!-- web/src/views/system/role/index.vue -->

<template>
  <el-dialog v-model="permissionDialogVisible" title="权限配置">
    <el-tree
      ref="permissionTreeRef"
      :data="permissionTree"
      show-checkbox
      node-key="id"
    />
    
    <!-- 变更预览 -->
    <el-card v-if="previewData" style="margin-top: 20px">
      <template #header>
        <span>📋 变更预览</span>
      </template>
      
      <div v-if="previewData.added?.length > 0">
        <el-tag type="success">➕ 新增 {{ previewData.added.length }} 个</el-tag>
        <ul>
          <li v-for="p in previewData.added" :key="p.id">
            {{ p.display_name }}
          </li>
        </ul>
      </div>
      
      <div v-if="previewData.removed?.length > 0">
        <el-tag type="danger">➖ 删除 {{ previewData.removed.length }} 个</el-tag>
        <ul>
          <li v-for="p in previewData.removed" :key="p.id">
            {{ p.display_name }}
          </li>
        </ul>
      </div>
      
      <el-empty v-if="!previewData.added?.length && !previewData.removed?.length" 
                description="没有任何变更" />
    </el-card>
    
    <template #footer>
      <el-button @click="handlePreview">🔍 预览变更</el-button>
      <el-button type="primary" :disabled="!previewData" @click="handleSubmit">
        ✅ 确认提交
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
const previewData = ref(null)

// 步骤1：预览变更
const handlePreview = async () => {
  const checkedKeys = permissionTreeRef.value.getCheckedKeys()
  const halfCheckedKeys = permissionTreeRef.value.getHalfCheckedKeys()
  const allKeys = [...checkedKeys, ...halfCheckedKeys]
  
  const { data } = await roleApi.updatePermissions(roleId, {
    permission_ids: allKeys,
    preview: true  // 预览模式
  })
  
  previewData.value = data.preview
}

// 步骤2：确认提交
const handleSubmit = async () => {
  const checkedKeys = permissionTreeRef.value.getCheckedKeys()
  const halfCheckedKeys = permissionTreeRef.value.getHalfCheckedKeys()
  const allKeys = [...checkedKeys, ...halfCheckedKeys]
  
  // 二次确认
  if (previewData.value.removed?.length > 0) {
    await ElMessageBox.confirm(
      `将删除 ${previewData.value.removed.length} 个权限，确定吗？`,
      '确认变更'
    )
  }
  
  await roleApi.updatePermissions(roleId, {
    permission_ids: allKeys,
    preview: false  // 执行模式
  })
  
  ElMessage.success('权限更新成功')
  permissionDialogVisible.value = false
}
</script>
```

### 前端API

```typescript
// web/src/api/role.ts

export const roleApi = {
  // 统一的更新接口
  updatePermissions(
    roleId: number, 
    data: { 
      permission_ids: number[]
      preview?: boolean 
    }
  ) {
    return request.post(`/roles/${roleId}/permissions/update`, data)
  }
}
```

**优点：** 
- ✅ 只需1个新API
- ✅ 自动计算差异
- ✅ 支持预览功能
- ✅ 用户体验最好

**缺点：** 需要前后端配合改动

---

## 🎯 推荐选择

### 如果你赶时间 → 选方案A（2小时）
- 改动最小
- 立即可用

### 如果你要规范 → 选方案B（4小时）
- API语义清晰
- 业界标准

### 如果你要完美 → 选方案C（3小时）✨ **最推荐**
- 用户体验最好
- 代码最优雅
- 只需1个新API

---

## 📝 实施步骤（以方案C为例）

### Step 1: 后端（30分钟）
1. 在 `internal/handler/role_handler.go` 添加 `UpdatePermissions` 方法
2. 在 `internal/service/rbac_service.go` 添加差异计算逻辑
3. 在路由中注册 `POST /roles/:id/permissions/update`

### Step 2: 前端（30分钟）
1. 在 `web/src/api/role.ts` 添加 `updatePermissions` 方法
2. 在 `web/src/views/system/role/index.vue` 添加预览卡片
3. 改造 `handlePermissionSubmit` 为两步操作

### Step 3: 测试（1小时）
1. 测试预览功能
2. 测试增量添加
3. 测试增量删除
4. 测试二次确认

---

**选哪个？告诉我，我立即开始写代码！** 🚀
