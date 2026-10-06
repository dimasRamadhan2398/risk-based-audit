-- ============================================================================
-- Migration 006: BRIN & Partial Indexes for Terabyte-Scale Performance
-- Massive storage savings and ultra-fast targeted scans
-- ============================================================================

-- 1. BRIN (Block Range Index) for naturally ordered time-series columns
-- Takes only tens of kilobytes instead of gigabytes on 1TB+ datasets
CREATE INDEX IF NOT EXISTS idx_brin_bronze_cb_trx_loaded
ON bronze.cb_transactions USING BRIN (_loaded_at);

CREATE INDEX IF NOT EXISTS idx_brin_silver_trx_date
ON silver.trx_cleaned USING BRIN (trx_date);

CREATE INDEX IF NOT EXISTS idx_brin_gold_fact_trx_date
ON gold.fact_transactions USING BRIN (trx_date);

CREATE INDEX IF NOT EXISTS idx_brin_gold_fact_gl_date
ON gold.fact_gl_entries USING BRIN (gl_date);

-- 2. Partial Indexes for targeted audit queries (ignores non-relevant bulk)
CREATE INDEX IF NOT EXISTS idx_gold_fact_trx_suspicious
ON gold.fact_transactions (source_id, trx_date DESC)
WHERE is_suspicious AND NOT is_deleted;

CREATE INDEX IF NOT EXISTS idx_gold_fact_trx_after_hours
ON gold.fact_transactions (source_id, trx_date DESC)
WHERE is_after_hours AND NOT is_deleted;

CREATE INDEX IF NOT EXISTS idx_gold_fact_loans_npl
ON gold.fact_loans (source_id, collectability, days_past_due)
WHERE collectability >= 3 AND NOT is_deleted;

CREATE INDEX IF NOT EXISTS idx_gold_caatt_exc_unconfirmed
ON gold.caatt_exceptions (source_test, severity, exception_date DESC)
WHERE NOT is_confirmed;

-- 3. Composite Lookup Indexes for Keyset Pagination & Multi-Source
CREATE INDEX IF NOT EXISTS idx_gold_fact_trx_cursor
ON gold.fact_transactions (source_id, trx_date DESC, trx_id DESC);

CREATE INDEX IF NOT EXISTS idx_gold_fact_loans_cursor
ON gold.fact_loans (source_id, outstanding DESC, loan_id DESC);

CREATE INDEX IF NOT EXISTS idx_gold_fact_gl_cursor
ON gold.fact_gl_entries (source_id, gl_date DESC, gl_id DESC);
