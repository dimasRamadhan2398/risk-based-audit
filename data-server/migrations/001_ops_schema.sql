-- ============================================================================
-- Migration 001: Operational Schema (ops)
-- Provides persistent job queue, watermark checkpoints, integrity manifests
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS ops;

-- 1. Schema Migration History Tracker
CREATE TABLE IF NOT EXISTS ops.schema_migrations (
    version         VARCHAR(100) PRIMARY KEY,
    description     TEXT,
    checksum        VARCHAR(64),
    applied_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Persistent Job Queue (PostgreSQL FOR UPDATE SKIP LOCKED)
CREATE TABLE IF NOT EXISTS ops.jobs (
    job_id          VARCHAR(64) PRIMARY KEY,
    job_type        VARCHAR(50) NOT NULL,          -- 'ingest', 'reconcile', 'full_resync', 'batch_score'
    source_id       VARCHAR(64),
    source_name     VARCHAR(255),
    priority        INT DEFAULT 10,                 -- lower = higher priority (e.g. 5=realtime, 10=incremental, 20=initial)
    status          VARCHAR(30) DEFAULT 'queued',   -- 'queued', 'running', 'completed', 'failed', 'cancelled'
    payload         JSONB DEFAULT '{}',
    progress        JSONB DEFAULT '{}',             -- e.g. {"records": 100000, "current_table": "...", "mb": 45.2}
    error_message   TEXT,
    worker_id       VARCHAR(100),
    heartbeat_at    TIMESTAMPTZ,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_jobs_queue ON ops.jobs (status, priority, created_at);
CREATE INDEX IF NOT EXISTS idx_ops_jobs_source ON ops.jobs (source_id, created_at DESC);

-- 3. Detailed Job Event Log
CREATE TABLE IF NOT EXISTS ops.job_events (
    event_id        BIGSERIAL PRIMARY KEY,
    job_id          VARCHAR(64) NOT NULL REFERENCES ops.jobs(job_id) ON DELETE CASCADE,
    level           VARCHAR(20) DEFAULT 'INFO',     -- 'INFO', 'WARN', 'ERROR'
    message         TEXT NOT NULL,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_job_events_job ON ops.job_events (job_id, created_at);

-- 4. Incremental Sync Watermarks
CREATE TABLE IF NOT EXISTS ops.sync_watermarks (
    source_id               VARCHAR(64) NOT NULL,
    table_name              VARCHAR(255) NOT NULL,
    watermark_column        VARCHAR(100),
    last_watermark_value    TEXT,
    last_pk_value           TEXT,
    last_sync_records       BIGINT DEFAULT 0,
    total_records_synced    BIGINT DEFAULT 0,
    last_sync_at            TIMESTAMPTZ,
    status                  VARCHAR(50) DEFAULT 'active',
    PRIMARY KEY (source_id, table_name)
);

-- 5. Batch Ingestion Manifests (Auditor Proof of Integrity)
CREATE TABLE IF NOT EXISTS ops.ingest_batches (
    batch_id        VARCHAR(64) PRIMARY KEY,
    source_id       VARCHAR(64) NOT NULL,
    table_name      VARCHAR(255) NOT NULL,
    target_zone     VARCHAR(30) DEFAULT 'bronze',
    rows_count      BIGINT NOT NULL,
    checksum_sha256 VARCHAR(64),
    watermark_start TEXT,
    watermark_end   TEXT,
    loaded_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_batches_source ON ops.ingest_batches (source_id, table_name, loaded_at DESC);

-- 6. Delete Reconciliation Runs
CREATE TABLE IF NOT EXISTS ops.reconcile_runs (
    run_id              VARCHAR(64) PRIMARY KEY,
    source_id           VARCHAR(64) NOT NULL,
    table_name          VARCHAR(255) NOT NULL,
    buckets_checked     INT DEFAULT 0,
    buckets_mismatched  INT DEFAULT 0,
    deletes_detected    BIGINT DEFAULT 0,
    started_at          TIMESTAMPTZ DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    status              VARCHAR(30) DEFAULT 'completed'
);

-- Grants
GRANT USAGE, CREATE ON SCHEMA ops TO etl_user, datahub_api;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA ops TO etl_user, datahub_api;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA ops TO etl_user, datahub_api;
ALTER DEFAULT PRIVILEGES IN SCHEMA ops GRANT ALL ON TABLES TO etl_user, datahub_api;
ALTER DEFAULT PRIVILEGES IN SCHEMA ops GRANT ALL ON SEQUENCES TO etl_user, datahub_api;
