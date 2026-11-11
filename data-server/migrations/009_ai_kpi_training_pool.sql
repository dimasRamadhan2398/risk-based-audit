-- ============================================================================
-- Migration 009: Dedicated AI KPI Training Pool (Outside CSV)
-- High-throughput, indexed tables for Auditee & Auditor AI model retraining
-- Prevents file I/O bottlenecks and memory freezes on large scale datasets
-- ============================================================================

DO $$
BEGIN
    -- 1. Auditee KPI Training Pool (Business Health & CAATT Aggregates)
    CREATE TABLE IF NOT EXISTS gold.ai_auditee_kpi_training_pool (
        id                  BIGSERIAL PRIMARY KEY,
        source_id           VARCHAR(64) DEFAULT 'cbs_simulator',
        kpi_name            VARCHAR(150) NOT NULL,
        entity              VARCHAR(150) NOT NULL,
        entity_type         VARCHAR(50) DEFAULT 'Branch',
        period_date         DATE NOT NULL,
        period_label        VARCHAR(30) NOT NULL,
        actual_value        NUMERIC(14,4) NOT NULL,
        target_value        NUMERIC(14,4) NOT NULL,
        impact              INT DEFAULT 3,
        likelihood          INT DEFAULT 3,
        risk_level          VARCHAR(20) DEFAULT 'LOW',
        sync_frequency      VARCHAR(50) DEFAULT 'Real-time (CDC)',
        created_at          TIMESTAMPTZ DEFAULT NOW(),
        updated_at          TIMESTAMPTZ DEFAULT NOW(),
        CONSTRAINT uq_auditee_kpi_pool UNIQUE (kpi_name, period_date, entity)
    );

    CREATE INDEX IF NOT EXISTS idx_auditee_pool_kpi_date ON gold.ai_auditee_kpi_training_pool (kpi_name, period_date DESC);
    CREATE INDEX IF NOT EXISTS idx_auditee_pool_entity_date ON gold.ai_auditee_kpi_training_pool (entity, period_date DESC);
    CREATE INDEX IF NOT EXISTS idx_auditee_pool_risk ON gold.ai_auditee_kpi_training_pool (risk_level);

    -- 2. Auditor KPI Training Pool (Internal Audit Effectiveness & Strategic Plans)
    CREATE TABLE IF NOT EXISTS gold.ai_auditor_kpi_training_pool (
        id                  BIGSERIAL PRIMARY KEY,
        code                VARCHAR(50) NOT NULL,
        kpi_name            VARCHAR(255) NOT NULL,
        category            VARCHAR(150) NOT NULL,
        strategic_objective TEXT,
        period_date         DATE NOT NULL,
        period_label        VARCHAR(50) NOT NULL,
        target_value        NUMERIC(10,2) NOT NULL,
        actual_value        NUMERIC(10,2) NOT NULL,
        achievement_rate    NUMERIC(10,2) NOT NULL,
        hib_hig             VARCHAR(5) DEFAULT 'HIG',
        unit                VARCHAR(30) DEFAULT '%',
        status              VARCHAR(20) DEFAULT 'Good',
        created_at          TIMESTAMPTZ DEFAULT NOW(),
        updated_at          TIMESTAMPTZ DEFAULT NOW(),
        CONSTRAINT uq_auditor_kpi_pool UNIQUE (code, period_date)
    );

    CREATE INDEX IF NOT EXISTS idx_auditor_pool_code_date ON gold.ai_auditor_kpi_training_pool (code, period_date DESC);
    CREATE INDEX IF NOT EXISTS idx_auditor_pool_cat_date ON gold.ai_auditor_kpi_training_pool (category, period_date DESC);
    CREATE INDEX IF NOT EXISTS idx_auditor_pool_status ON gold.ai_auditor_kpi_training_pool (status);

    -- 3. Register to Mart Version Tracking
    INSERT INTO gold.mart_version (mart_name, version) VALUES
        ('ai_auditee_kpi_training_pool', 1),
        ('ai_auditor_kpi_training_pool', 1)
    ON CONFLICT (mart_name) DO NOTHING;

EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Migration 009 notices: %', SQLERRM;
END $$;

