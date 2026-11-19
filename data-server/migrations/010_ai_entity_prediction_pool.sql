-- ============================================================================
-- Migration 010: Dedicated AI Entity Prediction Training Pool (Outside CSV)
-- High-throughput, indexed tables for Entity Prediction XGBoost AI model retraining
-- Prevents file I/O bottlenecks and memory freezes on large scale datasets
-- ============================================================================

DO $$
BEGIN
    -- 1. Entity Prediction Training Pool (Risk Profile & External DB direct source)
    CREATE TABLE IF NOT EXISTS gold.ai_entity_prediction_training_pool (
        id                      BIGSERIAL PRIMARY KEY,
        source_id               VARCHAR(64) DEFAULT 'risk_profile',
        entity                  VARCHAR(150) NOT NULL,
        entity_type             VARCHAR(50) DEFAULT 'Branch',
        risk_category           VARCHAR(150) NOT NULL,
        inherent_likelihood     NUMERIC(10,2) NOT NULL DEFAULT 3.0,
        inherent_impact         NUMERIC(10,2) NOT NULL DEFAULT 3.0,
        audit_findings_count    NUMERIC(10,2) NOT NULL DEFAULT 0.0,
        kpi_below_target_count  NUMERIC(10,2) NOT NULL DEFAULT 0.0,
        kpi_volatility          NUMERIC(10,4) NOT NULL DEFAULT 0.1,
        previous_risk_score     NUMERIC(10,2) NOT NULL DEFAULT 9.0,
        assessment_month        INT NOT NULL DEFAULT 1,
        target_likelihood       NUMERIC(10,2) NOT NULL DEFAULT 3.0,
        target_impact           NUMERIC(10,2) NOT NULL DEFAULT 3.0,
        current_risk_level      VARCHAR(30) DEFAULT 'MODERATE',
        created_at              TIMESTAMPTZ DEFAULT NOW(),
        updated_at              TIMESTAMPTZ DEFAULT NOW(),
        CONSTRAINT uq_entity_prediction_pool UNIQUE (entity, risk_category, assessment_month, source_id)
    );

    CREATE INDEX IF NOT EXISTS idx_entity_pool_entity ON gold.ai_entity_prediction_training_pool (entity);
    CREATE INDEX IF NOT EXISTS idx_entity_pool_cat ON gold.ai_entity_prediction_training_pool (risk_category);
    CREATE INDEX IF NOT EXISTS idx_entity_pool_type ON gold.ai_entity_prediction_training_pool (entity_type);
    CREATE INDEX IF NOT EXISTS idx_entity_pool_month ON gold.ai_entity_prediction_training_pool (assessment_month);
    CREATE INDEX IF NOT EXISTS idx_entity_pool_risk ON gold.ai_entity_prediction_training_pool (current_risk_level);

    -- 2. Register to Mart Version Tracking
    INSERT INTO gold.mart_version (mart_name, version) VALUES
        ('ai_entity_prediction_training_pool', 1)
    ON CONFLICT (mart_name) DO NOTHING;

EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Migration 010 notices: %', SQLERRM;
END $$;

