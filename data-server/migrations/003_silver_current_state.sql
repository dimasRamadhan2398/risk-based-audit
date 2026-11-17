-- ============================================================================
-- Migration 003: Silver Zone Current-State Models & Delete Tracking
-- Adds hash tracking, source_id multi-tenancy, and is_deleted state to Silver
-- ============================================================================

-- 1. Silver Transactions
ALTER TABLE silver.trx_cleaned ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
ALTER TABLE silver.trx_cleaned ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);
ALTER TABLE silver.trx_cleaned ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE silver.trx_cleaned ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE silver.trx_cleaned ADD COLUMN IF NOT EXISTS deleted_detected_at TIMESTAMPTZ;

-- 2. Silver Accounts
ALTER TABLE silver.acct_cleaned ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
ALTER TABLE silver.acct_cleaned ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);
ALTER TABLE silver.acct_cleaned ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE silver.acct_cleaned ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE silver.acct_cleaned ADD COLUMN IF NOT EXISTS deleted_detected_at TIMESTAMPTZ;

-- 3. Silver Customers
ALTER TABLE silver.cust_cleaned ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
ALTER TABLE silver.cust_cleaned ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);
ALTER TABLE silver.cust_cleaned ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE silver.cust_cleaned ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE silver.cust_cleaned ADD COLUMN IF NOT EXISTS deleted_detected_at TIMESTAMPTZ;

-- 4. Silver Loans
ALTER TABLE silver.loan_cleaned ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
ALTER TABLE silver.loan_cleaned ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);
ALTER TABLE silver.loan_cleaned ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE silver.loan_cleaned ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE silver.loan_cleaned ADD COLUMN IF NOT EXISTS deleted_detected_at TIMESTAMPTZ;

-- 5. Silver GL Entries
ALTER TABLE silver.gl_cleaned ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
ALTER TABLE silver.gl_cleaned ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);
ALTER TABLE silver.gl_cleaned ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE silver.gl_cleaned ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE silver.gl_cleaned ADD COLUMN IF NOT EXISTS deleted_detected_at TIMESTAMPTZ;

-- 6. External Data Sources & Silver External Source Data (ensure base tables exist on legacy DB volumes)
CREATE TABLE IF NOT EXISTS bronze.registered_sources (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id          UUID NOT NULL UNIQUE,
    name               VARCHAR(255) NOT NULL,
    source_type        VARCHAR(50) NOT NULL,
    host               VARCHAR(255) NOT NULL,
    port               INTEGER NOT NULL DEFAULT 5432,
    database_name      VARCHAR(255) NOT NULL,
    username           VARCHAR(255),
    password_encrypted TEXT,
    ssl_enabled        BOOLEAN DEFAULT TRUE,
    sync_schedule      VARCHAR(100) DEFAULT 'Manual Only',
    scopes             JSONB DEFAULT '[]',
    data_mappings      JSONB DEFAULT '[]',
    status             VARCHAR(50) DEFAULT 'registered',
    last_sync_at       TIMESTAMPTZ,
    records_synced     INTEGER DEFAULT 0,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS silver.external_source_data (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id      UUID NOT NULL,
    source_name    VARCHAR(255),
    table_name     VARCHAR(255),
    record_data    JSONB NOT NULL,
    dedup_key      VARCHAR(512),
    target_scope   VARCHAR(100),
    validated_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS gold.fact_external_audit_data (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id      UUID NOT NULL,
    source_name    VARCHAR(255),
    table_name     VARCHAR(255),
    target_scope   VARCHAR(100),
    target_module  VARCHAR(255),
    record_data    JSONB NOT NULL,
    anomaly_flags  JSONB DEFAULT '[]',
    processed_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS gold.ai_training_pool (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id      UUID NOT NULL,
    source_name    VARCHAR(255),
    table_name     VARCHAR(255),
    record_data    JSONB NOT NULL,
    feature_type   VARCHAR(100),
    labels         JSONB DEFAULT '{}',
    ingested_at    TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE silver.external_source_data ADD COLUMN IF NOT EXISTS record_hash VARCHAR(64);
ALTER TABLE silver.external_source_data ADD COLUMN IF NOT EXISTS batch_id VARCHAR(64);
ALTER TABLE silver.external_source_data ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE silver.external_source_data ADD COLUMN IF NOT EXISTS deleted_detected_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_bronze_registered_sources_sid ON bronze.registered_sources(source_id);
CREATE INDEX IF NOT EXISTS idx_silver_ext_data_src ON silver.external_source_data(source_id, table_name);
CREATE INDEX IF NOT EXISTS idx_gold_ext_audit_data_scope ON gold.fact_external_audit_data(target_scope, source_id);
CREATE INDEX IF NOT EXISTS idx_gold_ai_pool_source ON gold.ai_training_pool(source_id, feature_type);
CREATE INDEX IF NOT EXISTS idx_gold_ai_pool_ingested ON gold.ai_training_pool(ingested_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_silver_external_data
ON silver.external_source_data (source_id, table_name, dedup_key);

-- Partial index for active records
CREATE INDEX IF NOT EXISTS idx_silver_trx_active
ON silver.trx_cleaned (source_id, trx_date) WHERE NOT is_deleted;

CREATE INDEX IF NOT EXISTS idx_silver_loan_active
ON silver.loan_cleaned (source_id, status) WHERE NOT is_deleted;
