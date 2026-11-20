-- ============================================================================
-- Migration 012: Complete Delta Tracking Columns Across All CBS Bronze Tables
-- Ensures cb_accounts, cb_customers, cb_loans, and cb_branches carry _batch_id,
-- _op, and _row_hash metadata required by transform/silver.py set-based merges.
-- ============================================================================

ALTER TABLE bronze.cb_accounts  ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE bronze.cb_accounts  ADD COLUMN IF NOT EXISTS _op       CHAR(1) DEFAULT 'I';
ALTER TABLE bronze.cb_accounts  ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);

ALTER TABLE bronze.cb_customers ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE bronze.cb_customers ADD COLUMN IF NOT EXISTS _op       CHAR(1) DEFAULT 'I';
ALTER TABLE bronze.cb_customers ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);

ALTER TABLE bronze.cb_loans     ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE bronze.cb_loans     ADD COLUMN IF NOT EXISTS _op       CHAR(1) DEFAULT 'I';
ALTER TABLE bronze.cb_loans     ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);

ALTER TABLE bronze.cb_branches  ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE bronze.cb_branches  ADD COLUMN IF NOT EXISTS _op       CHAR(1) DEFAULT 'I';
ALTER TABLE bronze.cb_branches  ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);

