-- 添加角色等级字段
-- 等级说明：
-- 100: 超级管理员（系统最高权限）
-- 80:  系统管理员（可管理大部分系统配置）
-- 50:  部门管理员（可管理部门内用户和资源）
-- 30:  项目管理员（可管理项目相关资源）
-- 10:  普通用户（默认等级）

-- 添加 level 字段
ALTER TABLE roles ADD COLUMN IF NOT EXISTS level INT NOT NULL DEFAULT 10;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_roles_level ON roles(level);

-- 更新现有角色的等级
-- 根据角色名称设置合适的等级
UPDATE roles SET level = 100 WHERE name = 'super_admin' OR name = '超级管理员' OR display_name = '超级管理员';
UPDATE roles SET level = 80 WHERE name = 'admin' OR name = '系统管理员' OR display_name = '系统管理员';
UPDATE roles SET level = 50 WHERE name = 'manager' OR name = '部门管理员' OR display_name LIKE '%管理员%' AND level = 10;
UPDATE roles SET level = 30 WHERE name LIKE '%lead%' OR display_name LIKE '%负责人%' OR display_name LIKE '%主管%';
UPDATE roles SET level = 10 WHERE name = 'user' OR name = 'guest' OR display_name LIKE '%普通%' OR display_name LIKE '%访客%';

-- 添加注释
COMMENT ON COLUMN roles.level IS '角色等级(1-100)，数字越大权限越高，用于防止低等级角色修改高等级角色';
