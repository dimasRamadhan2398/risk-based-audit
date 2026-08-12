-- Create cloned databases
CREATE DATABASE rb_audit_clone;
CREATE DATABASE core_banking;
CREATE DATABASE rb_audit_master_service;
CREATE DATABASE rb_audit_audit_service;
CREATE DATABASE rb_audit_risk_service;
CREATE DATABASE rb_audit_auth_service;

-- Connect to core_banking and seed sample audit evidence tables
\c core_banking;

CREATE TABLE IF NOT EXISTS customer_accounts (
    id SERIAL PRIMARY KEY,
    account_number VARCHAR(50) UNIQUE NOT NULL,
    customer_name VARCHAR(100) NOT NULL,
    balance NUMERIC(15,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_evidence_logs (
    id SERIAL PRIMARY KEY,
    log_code VARCHAR(50) UNIQUE NOT NULL,
    module_name VARCHAR(100) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO customer_accounts (account_number, customer_name, balance, status) VALUES
('ACC-1001-9821', 'PT Bank Audit Sejahtera', 250000000.00, 'ACTIVE'),
('ACC-1002-4412', 'Budi Santoso', 15400000.50, 'ACTIVE'),
('ACC-1003-8819', 'Siti Aminah', 8920000.00, 'ACTIVE')
ON CONFLICT DO NOTHING;

INSERT INTO audit_evidence_logs (log_code, module_name, action_type, details) VALUES
('LOG-2026-001', 'Core Banking System', 'AUTHORIZATION', 'User Admin authorized daily transaction limit override'),
('LOG-2026-002', 'Risk Management', 'CDC_INGESTION', 'Real-time CDC evidence record ingested for audit compliance check')
ON CONFLICT DO NOTHING;
