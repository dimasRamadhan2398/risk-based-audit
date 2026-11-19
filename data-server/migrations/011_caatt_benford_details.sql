-- ============================================================================
-- Migration 011: Benford Chi-Square Origin Details & External Canonical Sequences
-- Enhances gold.caatt_benford_results with observed/expected counts, chi-square
-- contribution share, and forensic origin JSONB details. Adds safe sequences
-- for projecting external Data Source records into gold.fact_loans / fact_gl_entries.
-- ============================================================================

DO $$
BEGIN
    ALTER TABLE gold.caatt_benford_results ADD COLUMN IF NOT EXISTS actual_count INT DEFAULT 0;
    ALTER TABLE gold.caatt_benford_results ADD COLUMN IF NOT EXISTS expected_count NUMERIC(12,2) DEFAULT 0;
    ALTER TABLE gold.caatt_benford_results ADD COLUMN IF NOT EXISTS chi_square_share_pct NUMERIC(6,2) DEFAULT 0;
    ALTER TABLE gold.caatt_benford_results ADD COLUMN IF NOT EXISTS details JSONB DEFAULT '{}'::jsonb;

    CREATE INDEX IF NOT EXISTS idx_caatt_reconcil_src
        ON gold.caatt_reconciliation_results (source_id, test_date DESC);

    CREATE SEQUENCE IF NOT EXISTS gold.external_loan_id_seq START WITH 1000000 INCREMENT BY 1;
    CREATE SEQUENCE IF NOT EXISTS gold.external_gl_id_seq START WITH 1000000 INCREMENT BY 1;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Migration 011 notices: %', SQLERRM;
END $$;

