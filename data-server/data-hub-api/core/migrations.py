"""
AuditSphere Data Hub — Database Migration Runner
Automatically checks and applies incremental SQL migrations from migrations directory.
"""
import os
import glob
import hashlib
from sqlalchemy import create_engine, text

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
