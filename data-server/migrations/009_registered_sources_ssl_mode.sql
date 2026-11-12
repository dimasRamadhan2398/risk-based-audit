-- ============================================================================
-- Migration 009: Per-source SSL mode and custom CA on bronze.registered_sources
-- ssl_mode: disable | require | verify-ca | verify-full (libpq semantics)
-- ssl_root_cert: PEM CA bundle for verify-ca/verify-full; NULL/empty = system CA pool
-- Backward compatible: legacy ssl_enabled true -> require, false -> disable.
-- ============================================================================

ALTER TABLE bronze.registered_sources ADD COLUMN IF NOT EXISTS ssl_mode VARCHAR(16);
ALTER TABLE bronze.registered_sources ADD COLUMN IF NOT EXISTS ssl_root_cert TEXT;

UPDATE bronze.registered_sources
   SET ssl_mode = CASE WHEN COALESCE(ssl_enabled, TRUE) THEN 'require' ELSE 'disable' END
 WHERE ssl_mode IS NULL OR ssl_mode = '';

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_registered_sources_ssl_mode') THEN
        ALTER TABLE bronze.registered_sources
            ADD CONSTRAINT chk_registered_sources_ssl_mode
            CHECK (ssl_mode IS NULL OR ssl_mode IN ('disable', 'require', 'verify-ca', 'verify-full'));
    END IF;
END $$;
