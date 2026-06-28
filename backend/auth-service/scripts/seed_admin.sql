-- Seed Admin User and Role for Risk-Based Audit System
-- Password: Admin@123 (bcrypt hash below)

-- Create admin role
INSERT INTO roles (id, name, description, created_at, updated_at)
VALUES (gen_random_uuid(), 'admin', 'System Administrator with full access', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Create auditor role
INSERT INTO roles (id, name, description, created_at, updated_at)
VALUES (gen_random_uuid(), 'auditor', 'Internal Auditor', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Create viewer role
INSERT INTO roles (id, name, description, created_at, updated_at)
VALUES (gen_random_uuid(), 'viewer', 'Read-only access', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Create admin user (password: Admin@123)
INSERT INTO users (id, employee_id, username, email, password_hash, full_name, department, is_active, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'EMP-001',
  'admin',
  'admin@rbaudit.local',
  '$2a$10$3r.aoE1pTojJMDuEVuSH0OQxZ5zLMPCwIIgz.v2K9JJOLPtuMk242',
  'System Administrator',
  'Internal Audit',
  true,
  NOW(),
  NOW()
)
ON CONFLICT (username) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.username = 'admin' AND r.name = 'admin'
ON CONFLICT DO NOTHING;
