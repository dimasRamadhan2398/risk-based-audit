-- +migrate Up
-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Role-Permission many-to-many
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    phone VARCHAR(20),
    department VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- User-Role many-to-many
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- RBAC Matrix Features table
CREATE TABLE IF NOT EXISTS rbac_matrix_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_number INT NOT NULL,
    module VARCHAR(100) NOT NULL,
    submodule VARCHAR(150),
    feature_code VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    admin_access VARCHAR(50) NOT NULL DEFAULT 'FULL',
    audit_manager_access VARCHAR(50) NOT NULL DEFAULT 'READ',
    auditor_access VARCHAR(50) NOT NULL DEFAULT 'READ',
    auditee_access VARCHAR(50) NOT NULL DEFAULT 'NONE',
    viewer_access VARCHAR(50) NOT NULL DEFAULT 'READ',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS rbac_matrix_features CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
