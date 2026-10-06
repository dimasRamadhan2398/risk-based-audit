-- ============================================================================
-- Migration 002: TimescaleDB Extension & Immutable Bronze Zone
-- Converts bronze event tables to compressed hypertables & installs immutability guard
-- ============================================================================

-- 1. Enable TimescaleDB Extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- 2. Update Primary Key on Bronze Transactions for TimescaleDB Compatibility
DO $$
BEGIN
    -- Drop single-column PK if exists to allow composite PK with partition time
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_schema = 'bronze' AND table_name = 'cb_transactions' AND constraint_type = 'PRIMARY KEY'
    ) THEN
        ALTER TABLE bronze.cb_transactions DROP CONSTRAINT IF EXISTS cb_transactions_pkey;
        ALTER TABLE bronze.cb_transactions ADD CONSTRAINT cb_transactions_pkey PRIMARY KEY (_load_id, _loaded_at);
    END IF;
END $$;

-- 3. Add Delta Tracking & Immutability Metadata Columns to Bronze
ALTER TABLE bronze.cb_transactions ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE bronze.cb_transactions ADD COLUMN IF NOT EXISTS _op CHAR(1) DEFAULT 'I';  -- 'I'=Insert, 'U'=Update, 'D'=Delete tombstone
ALTER TABLE bronze.cb_transactions ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);

ALTER TABLE bronze.cb_gl_entries ADD COLUMN IF NOT EXISTS _batch_id VARCHAR(64);
ALTER TABLE bronze.cb_gl_entries ADD COLUMN IF NOT EXISTS _op CHAR(1) DEFAULT 'I';
ALTER TABLE bronze.cb_gl_entries ADD COLUMN IF NOT EXISTS _row_hash VARCHAR(64);

-- 4. Convert High-Volume Bronze Tables to Hypertables (Monthly Chunks)
SELECT create_hypertable(
    'bronze.cb_transactions',
    by_range('_loaded_at', INTERVAL '1 month'),
    if_not_exists => TRUE,
    migrate_data => TRUE
);

-- Enable Native Columnar Compression for Chunks Older Than 7 Days
DO $$
BEGIN
    BEGIN
        ALTER TABLE bronze.cb_transactions SET (
            timescaledb.compress,
            timescaledb.compress_segmentby = '_source',
            timescaledb.compress_orderby = '_loaded_at DESC'
        );
        PERFORM add_compression_policy('bronze.cb_transactions', INTERVAL '7 days', if_not_exists => TRUE);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Compression configuration notice: %', SQLERRM;
    END;
END $$;

-- 5. Immutability Guard Trigger (Append-Only Enforcement)
CREATE OR REPLACE FUNCTION ops.prevent_bronze_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'AuditSphere Compliance Violation: Bronze zone tables are an immutable audit trail. Modifying or deleting raw data is strictly prohibited.';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_bronze_cb_transactions_immutable ON bronze.cb_transactions;
CREATE TRIGGER trg_bronze_cb_transactions_immutable
BEFORE UPDATE OR DELETE ON bronze.cb_transactions
FOR EACH ROW EXECUTE FUNCTION ops.prevent_bronze_mutation();
