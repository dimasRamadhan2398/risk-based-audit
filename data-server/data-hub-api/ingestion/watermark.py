"""
AuditSphere Data Hub — Keyset Watermark Tracker
Tracks incremental ingestion progress to ensure only new/updated records are extracted.
"""
from typing import Optional, Tuple
from sqlalchemy import text
from core.db import engine

def get_table_watermark(source_id: str, table_name: str) -> Tuple[Optional[str], Optional[str]]:
    """Return (last_watermark_value, last_pk_value) for given source and table."""
    with engine.connect() as conn:
        row = conn.execute(text("""
            SELECT last_watermark_value, last_pk_value 
            FROM ops.sync_watermarks 
            WHERE source_id = :sid AND table_name = :tbl
        """), {"sid": source_id, "tbl": table_name}).fetchone()
        
        if row:
            return row[0], row[1]
        return None, None

def commit_table_watermark(source_id: str, table_name: str, wm_column: str,
                           new_wm_val: Optional[str], new_pk_val: Optional[str],
                           synced_count: int):
    """Atomically commit watermark checkpoint and update synced record count."""
    with engine.connect() as conn:
        conn.execute(text("""
            INSERT INTO ops.sync_watermarks 
                (source_id, table_name, watermark_column, last_watermark_value, last_pk_value,
                 last_sync_records, total_records_synced, last_sync_at, status)
            VALUES 
                (:sid, :tbl, :wm_col, :wm_val, :pk_val, :count, :count, NOW(), 'active')
            ON CONFLICT (source_id, table_name) DO UPDATE SET
                watermark_column = EXCLUDED.watermark_column,
                last_watermark_value = COALESCE(EXCLUDED.last_watermark_value, ops.sync_watermarks.last_watermark_value),
                last_pk_value = COALESCE(EXCLUDED.last_pk_value, ops.sync_watermarks.last_pk_value),
                last_sync_records = EXCLUDED.last_sync_records,
                total_records_synced = ops.sync_watermarks.total_records_synced + EXCLUDED.last_sync_records,
                last_sync_at = NOW(),
                status = 'active';
        """), {
            "sid": source_id,
            "tbl": table_name,
            "wm_col": wm_column,
            "wm_val": str(new_wm_val) if new_wm_val is not None else None,
            "pk_val": str(new_pk_val) if new_pk_val is not None else None,
            "count": synced_count
        })
        conn.commit()
