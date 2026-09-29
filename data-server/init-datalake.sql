-- ============================================================================
-- AuditSphere Data Hub — Data Lake DDL
-- Architecture: 3-Zone (Bronze / Silver / Gold)
-- Roles: etl_user (FULL all zones), auditor_ai (SELECT gold), datahub_api
-- ============================================================================

-- ============================================================================
-- SCHEMAS
-- ============================================================================
CREATE SCHEMA IF NOT EXISTS bronze;
CREATE SCHEMA IF NOT EXISTS silver;
CREATE SCHEMA IF NOT EXISTS gold;

-- ============================================================================
-- ROLES
-- ============================================================================

\getenv etl_password ETL_USER_PASSWORD
\getenv auditor_password AUDITOR_AI_PASSWORD
\getenv datahub_password DATAHUB_API_PASSWORD

SELECT set_config('app.etl_password', COALESCE(NULLIF(:'etl_password', ''), 'etl_password_dev123'), false);
SELECT set_config('app.auditor_password', COALESCE(NULLIF(:'auditor_password', ''), 'auditor_password_dev123'), false);
SELECT set_config('app.datahub_password', COALESCE(NULLIF(:'datahub_password', ''), 'datahub_password_dev123'), false);

-- etl_user: FULL access to all zones (Bronze, Silver, Gold)
DO $$
DECLARE
    v_pass text := COALESCE(NULLIF(current_setting('app.etl_password', true), ''), 'etl_password_dev123');
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'etl_user') THEN
        EXECUTE format('CREATE ROLE etl_user LOGIN PASSWORD %L', v_pass);
    ELSE
        EXECUTE format('ALTER ROLE etl_user WITH PASSWORD %L', v_pass);
    END IF;
END $$;

-- auditor_ai: Read-only access to Gold zone
DO $$
DECLARE
    v_pass text := COALESCE(NULLIF(current_setting('app.auditor_password', true), ''), 'auditor_password_dev123');
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'auditor_ai') THEN
        EXECUTE format('CREATE ROLE auditor_ai LOGIN PASSWORD %L', v_pass);
    ELSE
        EXECUTE format('ALTER ROLE auditor_ai WITH PASSWORD %L', v_pass);
    END IF;
END $$;

-- datahub_api: INSERT to bronze, SELECT on silver/gold
DO $$
DECLARE
    v_pass text := COALESCE(NULLIF(current_setting('app.datahub_password', true), ''), 'datahub_password_dev123');
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'datahub_api') THEN
        EXECUTE format('CREATE ROLE datahub_api LOGIN PASSWORD %L', v_pass);
    ELSE
        EXECUTE format('ALTER ROLE datahub_api WITH PASSWORD %L', v_pass);
    END IF;
END $$;

-- ============================================================================
-- GRANTS — etl_user: FULL on all zones
-- ============================================================================
GRANT USAGE, CREATE ON SCHEMA bronze, silver, gold TO etl_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA bronze TO etl_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA silver TO etl_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA gold   TO etl_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA bronze TO etl_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA silver TO etl_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA gold   TO etl_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA bronze GRANT ALL ON TABLES TO etl_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA silver GRANT ALL ON TABLES TO etl_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA gold   GRANT ALL ON TABLES TO etl_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA bronze GRANT ALL ON SEQUENCES TO etl_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA silver GRANT ALL ON SEQUENCES TO etl_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA gold   GRANT ALL ON SEQUENCES TO etl_user;

-- ============================================================================
-- GRANTS — auditor_ai: SELECT only on Gold
-- ============================================================================
GRANT USAGE ON SCHEMA gold TO auditor_ai;
GRANT SELECT ON ALL TABLES IN SCHEMA gold TO auditor_ai;
ALTER DEFAULT PRIVILEGES IN SCHEMA gold GRANT SELECT ON TABLES TO auditor_ai;

-- ============================================================================
-- GRANTS — datahub_api: INSERT/SELECT on bronze, SELECT on silver/gold
-- ============================================================================
GRANT USAGE ON SCHEMA bronze, silver, gold TO datahub_api;
GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA bronze TO datahub_api;
GRANT SELECT ON ALL TABLES IN SCHEMA silver TO datahub_api;
GRANT SELECT ON ALL TABLES IN SCHEMA gold TO datahub_api;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA bronze TO datahub_api;
ALTER DEFAULT PRIVILEGES IN SCHEMA bronze GRANT SELECT, INSERT ON TABLES TO datahub_api;
ALTER DEFAULT PRIVILEGES IN SCHEMA silver GRANT SELECT ON TABLES TO datahub_api;
ALTER DEFAULT PRIVILEGES IN SCHEMA gold   GRANT SELECT ON TABLES TO datahub_api;

-- ============================================================================
-- BRONZE ZONE — Raw data (as-is from sources)
-- ============================================================================

-- CBS raw data (mirrored from cbs_simulator or real CBS)
CREATE TABLE bronze.cb_transactions (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    transaction_id  INT,
    trx_ref_number  VARCHAR(30),
    trx_date        DATE,
    trx_time        TIME,
    trx_type        VARCHAR(30),
    category        VARCHAR(30),
    channel         VARCHAR(30),
    amount          NUMERIC(18,2),
    currency        VARCHAR(3),
    description     TEXT,
    account_id      INT,
    counterpart_account VARCHAR(20),
    branch_id       INT,
    teller_id       VARCHAR(20),
    authorization_status VARCHAR(30),
    is_suspicious   BOOLEAN
);

CREATE TABLE bronze.cb_accounts (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    account_id      INT,
    account_number  VARCHAR(20),
    account_type    VARCHAR(30),
    balance         NUMERIC(18,2),
    interest_rate   NUMERIC(5,2),
    status          VARCHAR(20),
    customer_id     INT,
    branch_id       INT,
    opened_date     DATE
);

CREATE TABLE bronze.cb_customers (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    customer_id     INT,
    cif_number      VARCHAR(20),
    customer_name   VARCHAR(150),
    customer_type   VARCHAR(20),
    risk_profile    VARCHAR(20),
    branch_id       INT,
    registration_date DATE,
    is_active       BOOLEAN
);

CREATE TABLE bronze.cb_loans (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    loan_id         INT,
    loan_number     VARCHAR(20),
    loan_type       VARCHAR(30),
    customer_id     INT,
    account_id      INT,
    product_id      INT,
    branch_id       INT,
    plafond         NUMERIC(18,2),
    outstanding     NUMERIC(18,2),
    interest_rate   NUMERIC(5,2),
    tenor_months    INT,
    disbursement_date DATE,
    maturity_date   DATE,
    collateral_value NUMERIC(18,2),
    collectability  INT,
    days_past_due   INT,
    status          VARCHAR(20)
);

CREATE TABLE bronze.cb_gl_entries (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    gl_id           INT,
    gl_date         DATE,
    gl_account_code VARCHAR(20),
    gl_account_name VARCHAR(100),
    voucher_number  VARCHAR(30),
    debit_amount    NUMERIC(18,2),
    credit_amount   NUMERIC(18,2),
    description     TEXT,
    branch_id       INT,
    posted_by       VARCHAR(50)
);

CREATE TABLE bronze.cb_products (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    product_id      INT,
    product_code    VARCHAR(20),
    product_name    VARCHAR(100),
    product_type    VARCHAR(30),
    min_rate        NUMERIC(5,2),
    max_rate        NUMERIC(5,2),
    max_plafond     NUMERIC(18,2),
    is_active       BOOLEAN
);

CREATE TABLE bronze.cb_product_policies (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    policy_id       INT,
    product_id      INT,
    rule_name       VARCHAR(100),
    rule_type       VARCHAR(30),
    min_value       NUMERIC(18,4),
    max_value       NUMERIC(18,4),
    description     TEXT
);

CREATE TABLE bronze.cb_documents (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    document_id     INT,
    document_number VARCHAR(30),
    document_type   VARCHAR(30),
    reference_id    INT,
    reference_type  VARCHAR(30),
    document_date   DATE,
    status          VARCHAR(20),
    expiry_date     DATE,
    branch_id       INT
);

CREATE TABLE bronze.cb_branches (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    branch_id       INT,
    branch_code     VARCHAR(10),
    branch_name     VARCHAR(100),
    branch_type     VARCHAR(20),
    region          VARCHAR(50),
    city            VARCHAR(50)
);

CREATE TABLE bronze.cb_failed_logins (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'cbs_simulator',
    log_id          INT,
    user_id         VARCHAR(50),
    ip_address      VARCHAR(45),
    attempt_time    TIMESTAMP,
    reason          VARCHAR(50),
    branch_id       INT
);

-- Master data sync from AuditSphere (Audit Server DB)
CREATE TABLE bronze.sync_companies (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'audit_server',
    id              UUID,
    company_code    VARCHAR(20),
    company_name    VARCHAR(150),
    legal_name      VARCHAR(200),
    company_type    VARCHAR(20),
    is_active       BOOLEAN
);

CREATE TABLE bronze.sync_departments (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'audit_server',
    id              UUID,
    department_code VARCHAR(50),
    department_name VARCHAR(100),
    business_unit_id UUID,
    is_active       BOOLEAN
);

CREATE TABLE bronze.sync_employees (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'audit_server',
    id              UUID,
    employee_id_number VARCHAR(50),
    full_name       VARCHAR(100),
    email           VARCHAR(100),
    department_id   UUID,
    job_role_id     UUID,
    is_active       BOOLEAN
);

CREATE TABLE bronze.sync_risk_categories (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'audit_server',
    id              UUID,
    name            VARCHAR(100),
    description     TEXT
);

CREATE TABLE bronze.sync_risk_registers (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'audit_server',
    id              UUID,
    code            VARCHAR(50),
    name            VARCHAR(200),
    department_id   UUID,
    risk_category_id UUID,
    inherent_score  INT,
    residual_score  INT,
    is_active       BOOLEAN
);

CREATE TABLE bronze.sync_audit_findings (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'audit_server',
    id              UUID,
    finding_code    VARCHAR(50),
    title           TEXT,
    description     TEXT,
    severity        VARCHAR(20),
    status          VARCHAR(20),
    department_id   UUID,
    created_at      TIMESTAMP
);

CREATE TABLE bronze.sync_audit_universe (
    _load_id        SERIAL PRIMARY KEY,
    _loaded_at      TIMESTAMP DEFAULT NOW(),
    _source         VARCHAR(50) DEFAULT 'audit_server',
    id              UUID,
    entity_code     VARCHAR(50),
    entity_name     VARCHAR(200),
    entity_type     VARCHAR(30),
    risk_rating     VARCHAR(20),
    last_audit_year INT,
    department_id   UUID,
    status          VARCHAR(20)
);

-- Data freshness tracking
CREATE TABLE bronze.data_freshness_log (
    log_id          SERIAL PRIMARY KEY,
    source_name     VARCHAR(50) NOT NULL,
    table_name      VARCHAR(100) NOT NULL,
    records_loaded  INT DEFAULT 0,
    loaded_at       TIMESTAMP DEFAULT NOW(),
    is_new          BOOLEAN DEFAULT TRUE
);

-- ============================================================================
-- SILVER ZONE — Cleaned, normalized, deduplicated
-- ============================================================================

CREATE TABLE silver.trx_cleaned (
    trx_id          SERIAL PRIMARY KEY,
    trx_ref_number  VARCHAR(30) UNIQUE,
    trx_date        DATE NOT NULL,
    trx_hour        INT,
    trx_day_of_week INT,
    trx_type        VARCHAR(30),
    category        VARCHAR(30),
    channel         VARCHAR(30),
    amount          NUMERIC(18,2),
    amount_millions NUMERIC(18,4),
    description     TEXT,
    account_id      INT,
    branch_id       INT,
    teller_id       VARCHAR(20),
    authorization_status VARCHAR(30),
    is_suspicious   BOOLEAN DEFAULT FALSE,
    is_round_amount BOOLEAN DEFAULT FALSE,
    is_after_hours  BOOLEAN DEFAULT FALSE,
    cleaned_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE silver.acct_cleaned (
    account_id      INT PRIMARY KEY,
    account_number  VARCHAR(20) UNIQUE,
    account_type    VARCHAR(30),
    balance         NUMERIC(18,2),
    interest_rate   NUMERIC(5,2),
    status          VARCHAR(20),
    customer_id     INT,
    branch_id       INT,
    opened_date     DATE,
    cleaned_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE silver.cust_cleaned (
    customer_id     INT PRIMARY KEY,
    cif_number      VARCHAR(20) UNIQUE,
    customer_name   VARCHAR(150),
    customer_type   VARCHAR(20),
    risk_profile    VARCHAR(20),
    branch_id       INT,
    is_active       BOOLEAN,
    cleaned_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE silver.loan_cleaned (
    loan_id         INT PRIMARY KEY,
    loan_number     VARCHAR(20) UNIQUE,
    loan_type       VARCHAR(30),
    customer_id     INT,
    branch_id       INT,
    product_id      INT,
    plafond         NUMERIC(18,2),
    outstanding     NUMERIC(18,2),
    interest_rate   NUMERIC(5,2),
    tenor_months    INT,
    disbursement_date DATE,
    maturity_date   DATE,
    collateral_value NUMERIC(18,2),
    ltv_ratio       NUMERIC(5,2),
    collectability  INT,
    days_past_due   INT,
    status          VARCHAR(20),
    cleaned_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE silver.gl_cleaned (
    gl_id           INT PRIMARY KEY,
    gl_date         DATE NOT NULL,
    gl_account_code VARCHAR(20),
    gl_account_name VARCHAR(100),
    voucher_number  VARCHAR(30),
    debit_amount    NUMERIC(18,2),
    credit_amount   NUMERIC(18,2),
    net_amount      NUMERIC(18,2),
    description     TEXT,
    branch_id       INT,
    posted_by       VARCHAR(50),
    has_voucher     BOOLEAN DEFAULT TRUE,
    cleaned_at      TIMESTAMP DEFAULT NOW()
);

-- Reference tables from master data
CREATE TABLE silver.ref_companies (
    id              UUID PRIMARY KEY,
    company_code    VARCHAR(20) UNIQUE,
    company_name    VARCHAR(150),
    company_type    VARCHAR(20),
    is_active       BOOLEAN,
    synced_at       TIMESTAMP DEFAULT NOW()
);

CREATE TABLE silver.ref_departments (
    id              UUID PRIMARY KEY,
    department_code VARCHAR(50),
    department_name VARCHAR(100),
    business_unit_id UUID,
    is_active       BOOLEAN,
    synced_at       TIMESTAMP DEFAULT NOW()
);

CREATE TABLE silver.ref_employees (
    id              UUID PRIMARY KEY,
    employee_id_number VARCHAR(50),
    full_name       VARCHAR(100),
    email           VARCHAR(100),
    department_id   UUID,
    is_active       BOOLEAN,
    synced_at       TIMESTAMP DEFAULT NOW()
);

CREATE TABLE silver.ref_risk_categories (
    id              UUID PRIMARY KEY,
    name            VARCHAR(100),
    description     TEXT,
    synced_at       TIMESTAMP DEFAULT NOW()
);

CREATE TABLE silver.ref_branches (
    branch_id       INT PRIMARY KEY,
    branch_code     VARCHAR(10) UNIQUE,
    branch_name     VARCHAR(100),
    branch_type     VARCHAR(20),
    region          VARCHAR(50),
    city            VARCHAR(50),
    cleaned_at      TIMESTAMP DEFAULT NOW()
);

-- ============================================================================
-- GOLD ZONE — Enriched, audit-ready views and tables
-- ============================================================================

-- ─── Dimension Tables (enriched from master data + CBS) ─────────────────────

CREATE TABLE gold.dim_branches (
    branch_id       INT PRIMARY KEY,
    branch_code     VARCHAR(10),
    branch_name     VARCHAR(100),
    branch_type     VARCHAR(20),
    region          VARCHAR(50),
    city            VARCHAR(50),
    total_accounts  INT DEFAULT 0,
    total_customers INT DEFAULT 0,
    total_trx_volume NUMERIC(18,2) DEFAULT 0,
    anomaly_count   INT DEFAULT 0,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE gold.dim_products (
    product_id      INT PRIMARY KEY,
    product_code    VARCHAR(20),
    product_name    VARCHAR(100),
    product_type    VARCHAR(30),
    min_rate        NUMERIC(5,2),
    max_rate        NUMERIC(5,2),
    max_plafond     NUMERIC(18,2),
    policy_count    INT DEFAULT 0,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

-- ─── Fact Tables (enriched transactions) ────────────────────────────────────

CREATE TABLE gold.fact_transactions (
    trx_id          SERIAL PRIMARY KEY,
    trx_ref_number  VARCHAR(30),
    trx_date        DATE,
    trx_hour        INT,
    trx_day_of_week INT,
    trx_type        VARCHAR(30),
    category        VARCHAR(30),
    channel         VARCHAR(30),
    amount          NUMERIC(18,2),
    amount_millions NUMERIC(18,4),
    branch_id       INT,
    branch_name     VARCHAR(100),
    region          VARCHAR(50),
    customer_name   VARCHAR(150),
    customer_type   VARCHAR(20),
    customer_risk   VARCHAR(20),
    account_type    VARCHAR(30),
    authorization_status VARCHAR(30),
    is_suspicious   BOOLEAN,
    is_round_amount BOOLEAN,
    is_after_hours  BOOLEAN,
    -- Enrichment flags
    benford_first_digit INT,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE gold.fact_loans (
    loan_id         INT PRIMARY KEY,
    loan_number     VARCHAR(20),
    loan_type       VARCHAR(30),
    customer_name   VARCHAR(150),
    customer_type   VARCHAR(20),
    branch_name     VARCHAR(100),
    region          VARCHAR(50),
    product_name    VARCHAR(100),
    plafond         NUMERIC(18,2),
    outstanding     NUMERIC(18,2),
    interest_rate   NUMERIC(5,2),
    tenor_months    INT,
    collateral_value NUMERIC(18,2),
    ltv_ratio       NUMERIC(5,2),
    collectability  INT,
    days_past_due   INT,
    status          VARCHAR(20),
    -- Policy compliance flags
    rate_within_policy BOOLEAN DEFAULT TRUE,
    plafond_within_policy BOOLEAN DEFAULT TRUE,
    tenor_within_policy BOOLEAN DEFAULT TRUE,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE gold.fact_gl_entries (
    gl_id           INT PRIMARY KEY,
    gl_date         DATE,
    gl_account_code VARCHAR(20),
    gl_account_name VARCHAR(100),
    voucher_number  VARCHAR(30),
    debit_amount    NUMERIC(18,2),
    credit_amount   NUMERIC(18,2),
    net_amount      NUMERIC(18,2),
    branch_name     VARCHAR(100),
    has_voucher     BOOLEAN,
    -- Sequence analysis
    expected_voucher VARCHAR(30),
    has_gap         BOOLEAN DEFAULT FALSE,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

-- ─── AI Training Tables ─────────────────────────────────────────────────────

CREATE TABLE gold.anomaly_training_data (
    id              SERIAL PRIMARY KEY,
    entity          VARCHAR(150),
    description     TEXT,
    amount_millions NUMERIC(18,4),
    hour_of_day     INT,
    day_of_week     INT,
    category        VARCHAR(30),
    channel         VARCHAR(30),
    is_round_amount INT DEFAULT 0,
    is_new_beneficiary INT DEFAULT 0,
    is_anomaly      BOOLEAN,
    impact          INT,
    likelihood      INT,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE gold.department_training_data (
    id              SERIAL PRIMARY KEY,
    entity          VARCHAR(150),
    risk_category   VARCHAR(50),
    inherent_likelihood INT,
    inherent_impact INT,
    findings_count  INT DEFAULT 0,
    kpi_below_target INT DEFAULT 0,
    kpi_volatility  NUMERIC(5,2) DEFAULT 0,
    previous_risk_score INT DEFAULT 0,
    assessment_month INT,
    target_likelihood INT,
    target_impact   INT,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE gold.document_training_data (
    id              SERIAL PRIMARY KEY,
    text_input      TEXT,
    risk_category   VARCHAR(50),
    sentiment       VARCHAR(20),
    impact          INT,
    likelihood      INT,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE gold.kpi_training_data (
    id              SERIAL PRIMARY KEY,
    kpi_name        VARCHAR(100),
    period          VARCHAR(20),
    actual_value    NUMERIC(10,2),
    impact          INT,
    likelihood      INT,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

-- ─── CAATT Results Tables ───────────────────────────────────────────────────

CREATE TABLE gold.caatt_full_population_results (
    id              SERIAL PRIMARY KEY,
    test_name       VARCHAR(100),
    test_date       TIMESTAMP DEFAULT NOW(),
    total_population INT,
    violations_found INT,
    violation_rate  NUMERIC(5,4),
    criteria        TEXT,
    details         JSONB,
    branch_name     VARCHAR(100),
    category        VARCHAR(30)
);

CREATE TABLE gold.caatt_duplicate_gap_results (
    id              SERIAL PRIMARY KEY,
    test_date       TIMESTAMP DEFAULT NOW(),
    result_type     VARCHAR(20), -- DUPLICATE or GAP
    reference_field VARCHAR(50),
    reference_value VARCHAR(100),
    duplicate_count INT,
    gap_start       VARCHAR(50),
    gap_end         VARCHAR(50),
    branch_name     VARCHAR(100),
    details         JSONB
);

CREATE TABLE gold.caatt_benford_results (
    id              SERIAL PRIMARY KEY,
    test_date       TIMESTAMP DEFAULT NOW(),
    digit           INT,
    expected_pct    NUMERIC(5,2),
    actual_pct      NUMERIC(5,2),
    deviation_pct   NUMERIC(5,2),
    chi_square      NUMERIC(10,4),
    is_significant  BOOLEAN,
    field_tested    VARCHAR(50),
    sample_size     INT
);

CREATE TABLE gold.caatt_stratification_results (
    id              SERIAL PRIMARY KEY,
    test_date       TIMESTAMP DEFAULT NOW(),
    stratum_label   VARCHAR(50),
    min_value       NUMERIC(18,2),
    max_value       NUMERIC(18,2),
    trx_count       INT,
    total_amount    NUMERIC(18,2),
    pct_of_total    NUMERIC(5,2),
    category        VARCHAR(30),
    branch_name     VARCHAR(100)
);

CREATE TABLE gold.caatt_reconciliation_results (
    id              SERIAL PRIMARY KEY,
    test_date       TIMESTAMP DEFAULT NOW(),
    source_a        VARCHAR(50),
    source_b        VARCHAR(50),
    matched_count   INT,
    unmatched_a     INT,
    unmatched_b     INT,
    match_rate_pct  NUMERIC(5,2),
    total_difference NUMERIC(18,2),
    details         JSONB
);

CREATE TABLE gold.caatt_policy_violations (
    id              SERIAL PRIMARY KEY,
    test_date       TIMESTAMP DEFAULT NOW(),
    violation_type  VARCHAR(50),
    rule_name       VARCHAR(100),
    reference_id    INT,
    reference_type  VARCHAR(30),
    actual_value    NUMERIC(18,4),
    policy_min      NUMERIC(18,4),
    policy_max      NUMERIC(18,4),
    severity        VARCHAR(20),
    branch_name     VARCHAR(100),
    details         JSONB
);

CREATE TABLE gold.caatt_data_quality_metrics (
    id              SERIAL PRIMARY KEY,
    test_date       TIMESTAMP DEFAULT NOW(),
    table_name      VARCHAR(100),
    total_rows      INT,
    null_count      INT,
    duplicate_count INT,
    completeness_pct NUMERIC(5,2),
    accuracy_pct    NUMERIC(5,2),
    timeliness_days NUMERIC(10,2),
    details         JSONB
);

-- CAATT feedback for AI retraining
CREATE TABLE gold.caatt_exceptions (
    id              SERIAL PRIMARY KEY,
    exception_date  TIMESTAMP DEFAULT NOW(),
    source_test     VARCHAR(50), -- which CAATT technique found it
    reference_id    INT,
    reference_type  VARCHAR(30),
    description     TEXT,
    severity        VARCHAR(20),
    is_confirmed    BOOLEAN DEFAULT FALSE, -- auditor confirms it's a real issue
    confirmed_by    VARCHAR(50),
    confirmed_at    TIMESTAMP
);

CREATE TABLE gold.caatt_labeled_data (
    id              SERIAL PRIMARY KEY,
    labeled_at      TIMESTAMP DEFAULT NOW(),
    data_type       VARCHAR(30), -- anomaly, department, document, kpi
    features        JSONB,
    label           VARCHAR(50),
    source          VARCHAR(50) DEFAULT 'caatt_feedback'
);

-- ─── Analytical Views (for AuditSphere features) ───────────────────────────

-- AI prediction results (written by AI Engine)
CREATE TABLE gold.ai_risk_score_predictions (
    id              SERIAL PRIMARY KEY,
    predicted_at    TIMESTAMP DEFAULT NOW(),
    entity          VARCHAR(150),
    entity_type     VARCHAR(30),
    risk_category   VARCHAR(50),
    predicted_impact INT,
    predicted_likelihood INT,
    predicted_score NUMERIC(5,2),
    risk_level      VARCHAR(20),
    model_version   VARCHAR(20)
);

CREATE TABLE gold.ai_anomaly_predictions (
    id              SERIAL PRIMARY KEY,
    predicted_at    TIMESTAMP DEFAULT NOW(),
    trx_ref_number  VARCHAR(30),
    entity          VARCHAR(150),
    category        VARCHAR(30),
    anomaly_score   NUMERIC(5,4),
    is_anomaly      BOOLEAN,
    severity        VARCHAR(20),
    amount          NUMERIC(18,2),
    model_version   VARCHAR(20)
);

CREATE TABLE gold.ai_document_analysis (
    id              SERIAL PRIMARY KEY,
    analyzed_at     TIMESTAMP DEFAULT NOW(),
    doc_id          VARCHAR(30),
    title           TEXT,
    risk_category   VARCHAR(50),
    sentiment       VARCHAR(20),
    impact          INT,
    likelihood      INT,
    severity_score  INT,
    confidence      NUMERIC(5,4),
    model_version   VARCHAR(20)
);

CREATE TABLE gold.ai_kpi_forecasts (
    id              SERIAL PRIMARY KEY,
    forecasted_at   TIMESTAMP DEFAULT NOW(),
    kpi_name        VARCHAR(100),
    current_value   NUMERIC(10,2),
    forecasted_value NUMERIC(10,2),
    trend           VARCHAR(20),
    alert_level     VARCHAR(20),
    risk_level      VARCHAR(20),
    model_version   VARCHAR(20)
);

-- ─── Dashboard / Report Views ───────────────────────────────────────────────

-- Summary statistics (refreshed by ETL)
CREATE TABLE gold.dashboard_statistics (
    id              SERIAL PRIMARY KEY,
    stat_date       DATE DEFAULT CURRENT_DATE,
    total_transactions INT DEFAULT 0,
    total_anomalies INT DEFAULT 0,
    total_policy_violations INT DEFAULT 0,
    total_duplicate_trx INT DEFAULT 0,
    total_gl_gaps   INT DEFAULT 0,
    total_npl_loans INT DEFAULT 0,
    data_quality_score NUMERIC(5,2) DEFAULT 0,
    refreshed_at    TIMESTAMP DEFAULT NOW()
);

-- Data freshness tracking (Gold zone level)
CREATE TABLE gold.data_freshness_log (
    log_id          SERIAL PRIMARY KEY,
    source_name     VARCHAR(50) NOT NULL,
    table_name      VARCHAR(100) NOT NULL,
    records_count   INT DEFAULT 0,
    last_refreshed  TIMESTAMP DEFAULT NOW(),
    is_new          BOOLEAN DEFAULT TRUE
);

-- Config cache from Audit Server
CREATE TABLE gold.cached_data_source_connections (
    id              UUID PRIMARY KEY,
    name            VARCHAR(255),
    type            VARCHAR(50),
    host            VARCHAR(255),
    port            INT,
    database_name   VARCHAR(255),
    environment     VARCHAR(50),
    status          VARCHAR(50),
    sync_schedule   VARCHAR(100),
    last_sync       VARCHAR(100),
    scopes          JSONB,
    data_mappings   JSONB,
    cached_at       TIMESTAMP DEFAULT NOW()
);

-- ============================================================================
-- EXTERNAL DATA SOURCES & AI POOL (Multi-source integration)
-- ============================================================================

-- Bronze zone: Registered external sources from Audit Server
CREATE TABLE IF NOT EXISTS bronze.registered_sources (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id         UUID NOT NULL UNIQUE,     -- ID from data_source_connections (Audit Server)
    name              VARCHAR(255) NOT NULL,
    source_type       VARCHAR(50) NOT NULL,      -- postgres, mysql, oracle, mssql, etc.
    host              VARCHAR(255) NOT NULL,
    port              INTEGER NOT NULL DEFAULT 5432,
    database_name     VARCHAR(255) NOT NULL,
    username          VARCHAR(255),
    password_encrypted TEXT,
    ssl_enabled       BOOLEAN DEFAULT TRUE,
    sync_schedule     VARCHAR(100) DEFAULT 'Manual Only',
    scopes            JSONB DEFAULT '[]',        -- ["risk_management","audit_features","qar_features","data_analytics"]
    data_mappings     JSONB DEFAULT '[]',
    status            VARCHAR(50) DEFAULT 'registered',
    last_sync_at      TIMESTAMPTZ,
    records_synced    INTEGER DEFAULT 0,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

-- Silver zone: Generic cleaned external source data
CREATE TABLE IF NOT EXISTS silver.external_source_data (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id      UUID NOT NULL,
    source_name    VARCHAR(255),
    table_name     VARCHAR(255),
    record_data    JSONB NOT NULL,
    dedup_key      VARCHAR(512),
    target_scope   VARCHAR(100),
    validated_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Gold zone: Enriched external audit data
CREATE TABLE IF NOT EXISTS gold.fact_external_audit_data (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id      UUID NOT NULL,
    source_name    VARCHAR(255),
    table_name     VARCHAR(255),
    target_scope   VARCHAR(100),
    target_module  VARCHAR(255),
    record_data    JSONB NOT NULL,
    anomaly_flags  JSONB DEFAULT '[]',
    processed_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Gold zone: AI Training Pool (populated only if scope includes "data_analytics")
CREATE TABLE IF NOT EXISTS gold.ai_training_pool (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id      UUID NOT NULL,
    source_name    VARCHAR(255),
    table_name     VARCHAR(255),
    record_data    JSONB NOT NULL,
    feature_type   VARCHAR(100),              -- anomaly, risk_score, text, performance
    labels         JSONB DEFAULT '{}',
    ingested_at    TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================
CREATE INDEX idx_bronze_trx_loaded ON bronze.cb_transactions(_loaded_at);
CREATE INDEX idx_bronze_registered_sources_sid ON bronze.registered_sources(source_id);
CREATE INDEX idx_silver_trx_date ON silver.trx_cleaned(trx_date);
CREATE INDEX idx_silver_trx_category ON silver.trx_cleaned(category);
CREATE INDEX idx_silver_ext_data_src ON silver.external_source_data(source_id, table_name);
CREATE INDEX idx_gold_fact_trx_date ON gold.fact_transactions(trx_date);
CREATE INDEX idx_gold_fact_trx_cat ON gold.fact_transactions(category);
CREATE INDEX idx_gold_fact_trx_branch ON gold.fact_transactions(branch_id);
CREATE INDEX idx_gold_fact_loan_coll ON gold.fact_loans(collectability);
CREATE INDEX idx_gold_fact_gl_date ON gold.fact_gl_entries(gl_date);
CREATE INDEX idx_gold_anomaly_train ON gold.anomaly_training_data(is_anomaly);
CREATE INDEX idx_gold_caatt_exc_src ON gold.caatt_exceptions(source_test);
CREATE INDEX idx_gold_freshness ON gold.data_freshness_log(source_name, is_new);
CREATE INDEX idx_gold_ext_audit_data_scope ON gold.fact_external_audit_data(target_scope, source_id);
CREATE INDEX idx_gold_ai_pool_source ON gold.ai_training_pool(source_id, feature_type);
CREATE INDEX idx_gold_ai_pool_ingested ON gold.ai_training_pool(ingested_at DESC);

-- Privileges for roles on all tables
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA bronze TO etl_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA silver TO etl_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA gold   TO etl_user;
GRANT SELECT ON ALL TABLES IN SCHEMA gold TO auditor_ai;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA bronze TO datahub_api;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA silver TO datahub_api;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA gold TO datahub_api;

-- Done
SELECT 'Data Lake initialized successfully' AS status,
       (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'bronze') AS bronze_tables,
       (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'silver') AS silver_tables,
       (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'gold') AS gold_tables;
