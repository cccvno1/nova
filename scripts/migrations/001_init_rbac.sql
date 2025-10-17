-- ============================================
-- RBAC 系统表结构初始化脚本
-- 注意：此脚本仅供参考，实际建议使用 GORM AutoMigrate
-- ============================================

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    domain VARCHAR(100) NOT NULL,
    category VARCHAR(50),
    is_system BOOLEAN DEFAULT FALSE,
    sort INTEGER DEFAULT 0,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT idx_role_domain UNIQUE (name, domain)
);

CREATE INDEX idx_roles_domain ON roles(domain);
CREATE INDEX idx_roles_is_system ON roles(is_system);
CREATE INDEX idx_roles_status ON roles(status);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    type VARCHAR(20) NOT NULL,
    domain VARCHAR(100) NOT NULL,
    resource VARCHAR(200) NOT NULL,
    action VARCHAR(50) NOT NULL,
    category VARCHAR(50),
    parent_id BIGINT DEFAULT 0,
    path VARCHAR(200),
    component VARCHAR(200),
    icon VARCHAR(50),
    is_system BOOLEAN DEFAULT FALSE,
    sort INTEGER DEFAULT 0,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT idx_permission_domain UNIQUE (name, domain)
);

CREATE INDEX idx_permissions_domain ON permissions(domain);
CREATE INDEX idx_permissions_type ON permissions(type);
CREATE INDEX idx_permissions_category ON permissions(category);
CREATE INDEX idx_permissions_parent_id ON permissions(parent_id);
CREATE INDEX idx_permissions_is_system ON permissions(is_system);
CREATE INDEX idx_permissions_status ON permissions(status);
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at);

-- 角色-权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_role_permissions_deleted_at ON role_permissions(deleted_at);

-- 用户-角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    domain VARCHAR(100) NOT NULL,
    assigned_by BIGINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_domain ON user_roles(domain);
CREATE INDEX idx_user_roles_deleted_at ON user_roles(deleted_at);

-- Casbin 策略表 (由 gorm-adapter 自动创建，此处仅作参考)
CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100)
);

CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype);
CREATE INDEX idx_casbin_rule_v0 ON casbin_rule(v0);
CREATE INDEX idx_casbin_rule_v1 ON casbin_rule(v1);
CREATE INDEX idx_casbin_rule_v2 ON casbin_rule(v2);

-- 添加注释
COMMENT ON TABLE roles IS '角色表';
COMMENT ON TABLE permissions IS '权限表';
COMMENT ON TABLE role_permissions IS '角色-权限关联表';
COMMENT ON TABLE user_roles IS '用户-角色关联表';
COMMENT ON TABLE casbin_rule IS 'Casbin策略规则表';

COMMENT ON COLUMN roles.name IS '角色标识';
COMMENT ON COLUMN roles.display_name IS '角色显示名称';
COMMENT ON COLUMN roles.domain IS '所属域/租户';
COMMENT ON COLUMN roles.is_system IS '是否系统角色（不可删除）';
COMMENT ON COLUMN roles.status IS '状态：1=启用，0=禁用';

COMMENT ON COLUMN permissions.name IS '权限标识';
COMMENT ON COLUMN permissions.display_name IS '权限显示名称';
COMMENT ON COLUMN permissions.type IS '权限类型：api, menu, button, data, field';
COMMENT ON COLUMN permissions.domain IS '所属域/租户';
COMMENT ON COLUMN permissions.resource IS '资源路径，对应Casbin的obj';
COMMENT ON COLUMN permissions.action IS '操作，对应Casbin的act';
COMMENT ON COLUMN permissions.parent_id IS '父权限ID（用于树形结构）';
