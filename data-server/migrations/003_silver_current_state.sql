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

-- 6. Silver External Source Data (Generic schema)
ALTER TABLE silver.external_source_data ADD COLUMN IF NOT EXISTS record_hash VARCHAR(64);
ALTER TABLE silver.external_source_data ADD COLUMN IF NOT EXISTS batch_id VARCHAR(64);
ALTER TABLE silver.external_source_data ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE silver.external_source_data ADD COLUMN IF NOT EXISTS deleted_detected_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS uq_silver_external_data
ON silver.external_source_data (source_id, table_name, dedup_key);

-- Partial index for active records
CREATE INDEX IF NOT EXISTS idx_silver_trx_active
ON silver.trx_cleaned (source_id, trx_date) WHERE NOT is_deleted;

CREATE INDEX IF NOT EXISTS idx_silver_loan_active
ON silver.loan_cleaned (source_id, status) WHERE NOT is_deleted;
