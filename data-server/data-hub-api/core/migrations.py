"""
AuditSphere Data Hub — Database Migration Runner
Applies incremental SQL migrations from the migrations directory, tracked per
file in ops.schema_migrations.

Safety properties:
  * A session-level Postgres advisory lock serializes concurrent runners
    (e.g. several uvicorn workers starting at once). The loser waits, then
    re-reads the applied set and finds nothing to do.
  * Databases created before per-file tracking existed (migrations 001-008
    applied "the old way") are detected by sentinel objects and backfilled as
    applied instead of being re-run.
  * Each file runs in its own transaction and is recorded in the same
    transaction.
"""
import os
import glob
import hashlib
from sqlalchemy import create_engine, text

# Arbitrary constant key for pg_advisory_lock ("ASDH" + "MIGR").
MIGRATION_LOCK_KEY = 0x41534448_4D494752 & 0x7FFFFFFFFFFFFFFF

# Legacy (pre-tracking) migrations -> SQL returning a row if the LAST effect of
# that file is present in the database. Only used to backfill; never to skip a
# file on a database where the sentinel is absent.
LEGACY_SENTINELS = {
    "001_ops_schema.sql": "SELECT 1 FROM pg_tables WHERE schemaname='ops' AND tablename='reconcile_runs'",
    "002_timescale_bronze.sql": "SELECT 1 FROM pg_trigger WHERE tgname='trg_bronze_cb_transactions_immutable'",
    "003_silver_current_state.sql": "SELECT 1 FROM pg_indexes WHERE schemaname='silver' AND indexname='idx_silver_loan_active'",
    "004_gold_canonical.sql": "SELECT 1 FROM pg_indexes WHERE schemaname='gold' AND indexname='uq_gold_fact_trx_ref'",
    "005_gold_marts.sql": "SELECT 1 FROM pg_tables WHERE schemaname='gold' AND tablename='mart_version'",
    "006_indexes.sql": "SELECT 1 FROM pg_indexes WHERE schemaname='gold' AND indexname='idx_gold_fact_gl_cursor'",
    "007_silver_gold_unique_keys.sql": "SELECT 1 FROM pg_indexes WHERE schemaname='gold' AND indexname='uq_gold_dim_branches_source_code'",
    "008_caatt_multi_source.sql": "SELECT 1 FROM pg_indexes WHERE schemaname='gold' AND indexname='idx_caatt_dq_src'",
}


def _backfill_legacy(conn, applied):
    """Mark pre-tracking migrations as applied when their objects already exist."""
    for version, probe in LEGACY_SENTINELS.items():
        if version in applied:
            continue
        if conn.execute(text(probe)).first() is not None:
            conn.execute(text("""
                INSERT INTO ops.schema_migrations (version, description, checksum, applied_at)
                VALUES (:ver, :desc, NULL, NOW())
                ON CONFLICT (version) DO NOTHING
            """), {"ver": version, "desc": f"Backfilled: {version} already present in schema (pre-tracking)"})
            applied.add(version)
            print(f"[Migrations] Backfilled as applied (not re-run): {version}")
    conn.commit()


def run_migrations(db_engine, migrations_dir="/app/migrations"):
    """Run any pending migrations found in migrations_dir."""
    print(f"[Migrations] Scanning migrations in: {migrations_dir}")
    if not os.path.isdir(migrations_dir):
        # Fallback to relative path if running locally
        alt_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "migrations"))
        if os.path.isdir(alt_dir):
            migrations_dir = alt_dir
        else:
            print(f"[Migrations] Notice: directory {migrations_dir} not found. Skipping.")
            return

    # Serialize concurrent runners: hold a session-level advisory lock on a
    # dedicated connection for the whole run (taken BEFORE any DDL, so the
    # CREATE SCHEMA race between workers cannot happen either).
    lock_conn = db_engine.connect()
    try:
        lock_conn.execute(text("SELECT pg_advisory_lock(:k)"), {"k": MIGRATION_LOCK_KEY})
        lock_conn.commit()
        _run_locked(db_engine, migrations_dir)
    finally:
        try:
            lock_conn.rollback()
            lock_conn.execute(text("SELECT pg_advisory_unlock(:k)"), {"k": MIGRATION_LOCK_KEY})
            lock_conn.commit()
        finally:
            lock_conn.close()


def _run_locked(db_engine, migrations_dir):
    # 1. Ensure ops schema and tracking table exist
    with db_engine.connect() as conn:
        conn.execute(text("CREATE SCHEMA IF NOT EXISTS ops;"))
        conn.execute(text("""
            CREATE TABLE IF NOT EXISTS ops.schema_migrations (
                version         VARCHAR(100) PRIMARY KEY,
                description     TEXT,
                checksum        VARCHAR(64),
                applied_at      TIMESTAMPTZ DEFAULT NOW()
            );
        """))
        conn.commit()

        # 2. Get list of already applied migrations
        result = conn.execute(text("SELECT version FROM ops.schema_migrations"))
        applied = {row[0] for row in result.fetchall()}
        conn.commit()

        # 2b. Databases migrated before per-file tracking: backfill, don't re-run
        _backfill_legacy(conn, applied)

    # 3. Discover migration files
    sql_files = sorted(glob.glob(os.path.join(migrations_dir, "*.sql")))
    print(f"[Migrations] Found {len(sql_files)} migration scripts ({len(applied)} already applied).")

    for filepath in sql_files:
        version = os.path.basename(filepath)
        if version in applied:
            continue

        print(f"[Migrations] Applying migration: {version} ...")
        with open(filepath, "r", encoding="utf-8") as f:
            sql_content = f.read()

        checksum = hashlib.sha256(sql_content.encode("utf-8")).hexdigest()

        with db_engine.connect() as conn:
            # Execute raw SQL script
            try:
                conn.execute(text(sql_content))
                conn.execute(text("""
                    INSERT INTO ops.schema_migrations (version, description, checksum, applied_at)
                    VALUES (:ver, :desc, :chk, NOW())
                """), {"ver": version, "desc": f"Applied {version}", "chk": checksum})
                conn.commit()
                print(f"[Migrations] ✓ Successfully applied: {version}")
            except Exception as e:
                conn.rollback()
                print(f"[Migrations] ✗ Migration failed for {version}: {e}")
                raise e

    print("[Migrations] All migrations are up to date.")

if __name__ == "__main__":
    db_url = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/auditsphere_datalake")
    eng = create_engine(db_url)
    run_migrations(eng)
    eng.dispose()
