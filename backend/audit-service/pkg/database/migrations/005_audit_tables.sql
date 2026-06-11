-- +migrate Up
-- Audit Charters table
CREATE TABLE IF NOT EXISTS audit_charters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version VARCHAR(20) UNIQUE NOT NULL,
    filename VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS audit_charters CASCADE;
