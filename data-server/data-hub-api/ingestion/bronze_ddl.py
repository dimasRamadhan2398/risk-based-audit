"""
AuditSphere Data Hub — Bronze Schema Provisioner
Ensures bronze raw tables exist, applies additive schema evolution,
converts to TimescaleDB hypertables, and attaches immutable append-only triggers.
"""
from typing import List, Dict, Any
from core.identifiers import validate_identifier, quote_identifier

# Mapping standard cross-database types to PostgreSQL Bronze storage types
TYPE_MAP = {
    "integer": "BIGINT",
    "int": "BIGINT",
    "bigint": "BIGINT",
    "smallint": "INT",
    "numeric": "NUMERIC(18,4)",
    "decimal": "NUMERIC(18,4)",
    "float": "DOUBLE PRECISION",
    "double": "DOUBLE PRECISION",
    "real": "REAL",
    "boolean": "BOOLEAN",
    "bool": "BOOLEAN",
    "date": "DATE",
    "timestamp": "TIMESTAMPTZ",
    "timestamptz": "TIMESTAMPTZ",
    "datetime": "TIMESTAMPTZ",
    "time": "TIME",
    "json": "JSONB",
    "jsonb": "JSONB",
    "text": "TEXT",
    "varchar": "TEXT",
    "character varying": "TEXT",
    "char": "TEXT",
    "uuid": "UUID"
}

def get_bronze_table_name(source_name: str, table_name: str) -> str:
    """Generate clean bronze table name: bronze.<source_name>_<table_name>."""
    clean_src = "".join(c if c.isalnum() else "_" for c in source_name.lower())
    clean_tbl = "".join(c if c.isalnum() else "_" for c in table_name.lower())
    return f"{clean_src}_{clean_tbl}"

def ensure_bronze_table(source_name: str, table_name: str, columns: List[Dict[str, Any]]):
    """
    Ensure Bronze raw table exists with typed columns, metadata columns,
    hypertable conversion, and immutability guard trigger.
    """
    target_table = get_bronze_table_name(source_name, table_name)
    validate_identifier(target_table)

    col_defs = []
    for col in columns:
        col_name = col["name"]
        raw_type = col.get("dataType", "text").lower().split("(")[0].strip()
        pg_type = TYPE_MAP.get(raw_type, "TEXT")
        col_defs.append(f"{quote_identifier(col_name)} {pg_type}")

    col_defs_sql = ",\n    ".join(col_defs) if col_defs else "raw_payload JSONB"

    create_table_sql = f"""
    CREATE TABLE IF NOT EXISTS bronze.{quote_identifier(target_table)} (
        _load_id        BIGSERIAL,
        _loaded_at      TIMESTAMPTZ DEFAULT NOW(),
        _source_id      VARCHAR(64),
        _batch_id       VARCHAR(64),
        _op             CHAR(1) DEFAULT 'I',
        _row_hash       VARCHAR(64),
        {col_defs_sql},
        PRIMARY KEY (_load_id, _loaded_at)
    );
    """

    from sqlalchemy import text
    from core.db import engine

    with engine.connect() as conn:
        conn.execute(text(create_table_sql))
        conn.commit()

        # Additive schema evolution: add any newly discovered columns
        for col in columns:
            col_name = col["name"]
            raw_type = col.get("dataType", "text").lower().split("(")[0].strip()
            pg_type = TYPE_MAP.get(raw_type, "TEXT")
            try:
                conn.execute(text(f"""
                    ALTER TABLE bronze.{quote_identifier(target_table)}
                    ADD COLUMN IF NOT EXISTS {quote_identifier(col_name)} {pg_type};
                """))
                conn.commit()
            except Exception:
                pass

        # Convert to TimescaleDB Hypertable
        try:
            conn.execute(text(f"""
                SELECT create_hypertable(
                    'bronze.{quote_identifier(target_table)}',
                    by_range('_loaded_at', INTERVAL '1 month'),
                    if_not_exists => TRUE,
                    migrate_data => TRUE
                );
            """))
            conn.commit()
        except Exception:
            pass

        # Attach Immutability Guard Trigger (Append-Only Enforcement)
        try:
            trg_name = f"trg_bronze_{target_table}_immutable"
            conn.execute(text(f"""
                DROP TRIGGER IF EXISTS {quote_identifier(trg_name)} ON bronze.{quote_identifier(target_table)};
                CREATE TRIGGER {quote_identifier(trg_name)}
                BEFORE UPDATE OR DELETE ON bronze.{quote_identifier(target_table)}
                FOR EACH ROW EXECUTE FUNCTION ops.prevent_bronze_mutation();
            """))
            conn.commit()
        except Exception:
            pass

    return target_table
