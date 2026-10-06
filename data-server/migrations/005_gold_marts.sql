-- ============================================================================
-- Migration 005: Continuous Aggregates & Feature Marts
-- Precomputed analytics for instant sub-300ms serving across 16 features
-- ============================================================================

-- 1. Continuous Aggregate: Benford's Law Daily Digits
CREATE MATERIALIZED VIEW IF NOT EXISTS gold.cagg_benford_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', trx_date) AS bucket_day,
    source_id,
    benford_first_digit,
    COUNT(*) AS digit_count
FROM gold.fact_transactions
WHERE benford_first_digit BETWEEN 1 AND 9 AND NOT is_deleted
GROUP BY bucket_day, source_id, benford_first_digit
WITH NO DATA;

DO $$
BEGIN
    PERFORM add_continuous_aggregate_policy('gold.cagg_benford_daily',
        start_offset => INTERVAL '3 months',
        end_offset => INTERVAL '1 hour',
        schedule_interval => INTERVAL '10 minutes',
        if_not_exists => TRUE);
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'cagg_benford policy notice: %', SQLERRM;
END $$;

-- 2. Continuous Aggregate: Stratification & Category Breakdown
CREATE MATERIALIZED VIEW IF NOT EXISTS gold.cagg_strat_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', trx_date) AS bucket_day,
    source_id,
    category,
    branch_name,
    CASE
        WHEN amount < 10000000 THEN '< Rp 10 Juta'
        WHEN amount < 50000000 THEN 'Rp 10 Juta - 50 Juta'
        WHEN amount < 100000000 THEN 'Rp 50 Juta - 100 Juta'
        WHEN amount < 500000000 THEN 'Rp 100 Juta - 500 Juta'
        ELSE '> Rp 500 Juta'
    END AS stratum_label,
    COUNT(*) AS trx_count,
    SUM(amount) AS total_amount
FROM gold.fact_transactions
WHERE NOT is_deleted
GROUP BY bucket_day, source_id, category, branch_name, stratum_label
WITH NO DATA;

DO $$
BEGIN
    PERFORM add_continuous_aggregate_policy('gold.cagg_strat_daily',
        start_offset => INTERVAL '3 months',
        end_offset => INTERVAL '1 hour',
        schedule_interval => INTERVAL '10 minutes',
        if_not_exists => TRUE);
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'cagg_strat policy notice: %', SQLERRM;
END $$;

-- 3. Feature Mart: Corporate Risk Profile
CREATE TABLE IF NOT EXISTS gold.mart_entity_risk_profile (
    source_id               VARCHAR(64) NOT NULL,
    entity                  VARCHAR(150) NOT NULL,
    entity_type             VARCHAR(30) DEFAULT 'Branch',
    total_trx_volume        NUMERIC(18,2) DEFAULT 0,
    total_anomalies         INT DEFAULT 0,
    policy_violations       INT DEFAULT 0,
    npl_loans               INT DEFAULT 0,
    ai_predicted_score      NUMERIC(5,2) DEFAULT 0,
    risk_level              VARCHAR(20) DEFAULT 'LOW',
    updated_at              TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source_id, entity)
);

-- 4. Feature Mart: Strategic Audit Plan Priority Ranking
CREATE TABLE IF NOT EXISTS gold.mart_audit_plan_priority (
    source_id               VARCHAR(64) NOT NULL,
    entity_code             VARCHAR(50) NOT NULL,
    entity_name             VARCHAR(200) NOT NULL,
    risk_score              NUMERIC(5,2) DEFAULT 0,
    risk_level              VARCHAR(20) DEFAULT 'LOW',
    last_audit_year         INT DEFAULT 2023,
    priority_rank           INT DEFAULT 999,
    recommended_cycle       VARCHAR(30) DEFAULT 'Annual',
    updated_at              TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source_id, entity_code)
);

-- 5. Feature Mart: Assignment Letter Scope
CREATE TABLE IF NOT EXISTS gold.mart_assignment_scope (
    source_id               VARCHAR(64) NOT NULL,
    entity_name             VARCHAR(150) NOT NULL,
    audit_period            VARCHAR(50) DEFAULT 'Current Year',
    total_population        BIGINT DEFAULT 0,
    high_risk_amount        NUMERIC(18,2) DEFAULT 0,
    top_findings            JSONB DEFAULT '[]',
    updated_at              TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source_id, entity_name)
);

-- 6. Feature Mart: Quality Assurance Review Metrics
CREATE TABLE IF NOT EXISTS gold.mart_qar_metrics (
    source_id               VARCHAR(64) NOT NULL,
    audit_year              INT NOT NULL,
    total_engagements       INT DEFAULT 0,
    population_coverage_pct NUMERIC(5,2) DEFAULT 100.0,
    resolved_exceptions_pct NUMERIC(5,2) DEFAULT 0.0,
    qa_score                NUMERIC(5,2) DEFAULT 95.0,
    updated_at              TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source_id, audit_year)
);

-- 7. Feature Mart: KPI Series
CREATE TABLE IF NOT EXISTS gold.mart_kpi_series (
    source_id               VARCHAR(64) NOT NULL,
    kpi_name                VARCHAR(100) NOT NULL,
    period                  VARCHAR(20) NOT NULL,
    actual_value            NUMERIC(10,2) DEFAULT 0,
    target_value            NUMERIC(10,2) DEFAULT 0,
    updated_at              TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source_id, kpi_name, period)
);

-- 8. Feature Mart: Loan Aging
CREATE TABLE IF NOT EXISTS gold.mart_loan_aging (
    source_id               VARCHAR(64) NOT NULL,
    collectability          INT NOT NULL,
    dpd_bucket              VARCHAR(30) NOT NULL,
    loan_count              INT DEFAULT 0,
    total_outstanding       NUMERIC(18,2) DEFAULT 0,
    updated_at              TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source_id, collectability, dpd_bucket)
);

-- 9. Feature Mart: Data Quality & Delete Tracking
CREATE TABLE IF NOT EXISTS gold.mart_data_quality (
    source_id               VARCHAR(64) NOT NULL,
    table_name              VARCHAR(100) NOT NULL,
    total_rows              BIGINT DEFAULT 0,
    null_count              BIGINT DEFAULT 0,
    duplicate_count         BIGINT DEFAULT 0,
    deletes_detected        BIGINT DEFAULT 0,
    completeness_pct        NUMERIC(5,2) DEFAULT 100.0,
    accuracy_pct            NUMERIC(5,2) DEFAULT 100.0,
    timeliness_days         NUMERIC(10,2) DEFAULT 0.0,
    updated_at              TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source_id, table_name)
);

-- 10. Mart Version Cache Counter (for ETag generation)
CREATE TABLE IF NOT EXISTS gold.mart_version (
    mart_name               VARCHAR(100) PRIMARY KEY,
    version                 BIGINT DEFAULT 1,
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO gold.mart_version (mart_name, version) VALUES
    ('risk_profile', 1),
    ('audit_plan', 1),
    ('assignment_scope', 1),
    ('qar_metrics', 1),
    ('kpi_series', 1),
    ('data_quality', 1),
    ('caatt_results', 1)
ON CONFLICT (mart_name) DO NOTHING;
