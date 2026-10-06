-- ============================================================================
-- Migration 008: Multi-Source Support on CAATT Gold Analytics Tables
-- Adds source_id column and composite indexes to all 8 CAATT result tables
-- ============================================================================

DO $$
BEGIN
    ALTER TABLE gold.caatt_full_population_results ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
    ALTER TABLE gold.caatt_duplicate_gap_results ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
    ALTER TABLE gold.caatt_benford_results ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
    ALTER TABLE gold.caatt_stratification_results ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
    ALTER TABLE gold.caatt_reconciliation_results ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
    ALTER TABLE gold.caatt_policy_violations ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
    ALTER TABLE gold.caatt_data_quality_metrics ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';
    ALTER TABLE gold.caatt_exceptions ADD COLUMN IF NOT EXISTS source_id VARCHAR(64) DEFAULT 'cbs_simulator';

    CREATE INDEX IF NOT EXISTS idx_caatt_full_pop_src ON gold.caatt_full_population_results (source_id, test_date DESC);
    CREATE INDEX IF NOT EXISTS idx_caatt_dup_gap_src ON gold.caatt_duplicate_gap_results (source_id, test_date DESC);
    CREATE INDEX IF NOT EXISTS idx_caatt_benford_src ON gold.caatt_benford_results (source_id, test_date DESC);
    CREATE INDEX IF NOT EXISTS idx_caatt_strat_src ON gold.caatt_stratification_results (source_id, test_date DESC);
    CREATE INDEX IF NOT EXISTS idx_caatt_policy_src ON gold.caatt_policy_violations (source_id, test_date DESC);
    CREATE INDEX IF NOT EXISTS idx_caatt_dq_src ON gold.caatt_data_quality_metrics (source_id, test_date DESC);
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Migration 008 notices: %', SQLERRM;
END $$;
