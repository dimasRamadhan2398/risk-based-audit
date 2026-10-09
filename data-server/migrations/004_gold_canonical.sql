-- ============================================================================
-- Migration 004: Gold Canonical Models & Multi-Source Hypertables
-- Enables multi-source fact tables, TimescaleDB compression & unique keys
-- ============================================================================

-- 1. Multi-source & Delete Flag Support on Gold Fact Tables
ALTER TABLE gold.fact_transactions ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
ALTER TABLE gold.fact_transactions ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;

ALTER TABLE gold.fact_loans ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
ALTER TABLE gold.fact_loans ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;

ALTER TABLE gold.fact_gl_entries ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
ALTER TABLE gold.fact_gl_entries ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;

ALTER TABLE gold.dim_branches ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';

-- 2. TimescaleDB Hypertable Setup for Gold Transactions
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_schema = 'gold' AND table_name = 'fact_transactions' AND constraint_type = 'PRIMARY KEY'
    ) THEN
        ALTER TABLE gold.fact_transactions DROP CONSTRAINT IF EXISTS fact_transactions_pkey;
        ALTER TABLE gold.fact_transactions ADD CONSTRAINT fact_transactions_pkey PRIMARY KEY (trx_id, trx_date);
    END IF;
END $$;

SELECT create_hypertable(
    'gold.fact_transactions',
    by_range('trx_date', INTERVAL '1 month'),
    if_not_exists => TRUE,
    migrate_data => TRUE
);

-- Enable Native Columnar Compression for Gold Chunks Older Than 7 Days
DO $$
BEGIN
    BEGIN
        ALTER TABLE gold.fact_transactions SET (
            timescaledb.compress,
            timescaledb.compress_segmentby = 'source_id, category',
            timescaledb.compress_orderby = 'trx_date DESC'
        );
        PERFORM add_compression_policy('gold.fact_transactions', INTERVAL '7 days', if_not_exists => TRUE);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Gold compression setup notice: %', SQLERRM;
    END;
END $$;

-- 3. Unique Deduplication & Fast Lookup Indexes
CREATE UNIQUE INDEX IF NOT EXISTS uq_gold_fact_trx_ref
ON gold.fact_transactions (source_id, trx_ref_number, trx_date);

CREATE INDEX IF NOT EXISTS idx_gold_fact_trx_source_date
ON gold.fact_transactions (source_id, trx_date DESC) WHERE NOT is_deleted;

CREATE INDEX IF NOT EXISTS idx_gold_fact_loans_source
ON gold.fact_loans (source_id, status) WHERE NOT is_deleted;

CREATE INDEX IF NOT EXISTS idx_gold_fact_gl_source
ON gold.fact_gl_entries (source_id, gl_date DESC) WHERE NOT is_deleted;
