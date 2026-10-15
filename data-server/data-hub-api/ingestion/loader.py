"""
AuditSphere Data Hub — Streaming Bronze Loader
Streams batches directly into Bronze tables using PostgreSQL COPY protocol.
Calculates row hashes, manifests, and commits watermark progress.
"""
import uuid
import hashlib
from datetime import datetime, timezone
from typing import List, Generator, Tuple, Optional
from core.db import get_datalake_raw_conn, engine
from core.identifiers import quote_identifier
from ingestion.watermark import commit_table_watermark
from sqlalchemy import text

def stream_load_to_bronze(
    source_id: str,
    source_name: str,
    target_bronze_table: str,
    column_names: List[str],
    extractor_stream: Generator[Tuple[List[Tuple], Optional[str], Optional[str]], None, None],
    wm_col: Optional[str] = None,
    op: str = "I"
) -> Tuple[int, List[str]]:
    """
    Consume chunks from extractor stream and load into Bronze via COPY protocol.
    Returns: (total_records_loaded, list_of_batch_ids_created)
    """
    total_records = 0
    batch_ids = []

    # Columns: (_source_id, _batch_id, _loaded_at, _op, _row_hash, col1, col2, ...)
    dest_cols = ["_source_id", "_batch_id", "_loaded_at", "_op", "_row_hash"] + column_names
    dest_cols_sql = ", ".join(quote_identifier(c) for c in dest_cols)
    copy_sql = f"COPY bronze.{quote_identifier(target_bronze_table)} ({dest_cols_sql}) FROM STDIN"

    with get_datalake_raw_conn() as conn:
        for chunk_rows, chunk_wm, chunk_pk in extractor_stream:
            if not chunk_rows:
                continue

            batch_id = str(uuid.uuid4())
            batch_ids.append(batch_id)
            now_dt = datetime.now(timezone.utc)
            now_str = now_dt.isoformat()

            hasher = hashlib.sha256()

            with conn.cursor() as cur:
                with cur.copy(copy_sql) as copy:
                    for row in chunk_rows:
                        # Compute fast MD5 row hash for silver change-detection
                        row_bytes = str(row).encode("utf-8")
                        row_hash = hashlib.md5(row_bytes).hexdigest()
                        hasher.update(row_bytes)

                        full_row = (source_id, batch_id, now_str, op, row_hash) + tuple(row)
                        copy.write_row(full_row)

            conn.commit()

            batch_rows_count = len(chunk_rows)
            total_records += batch_rows_count
            checksum = hasher.hexdigest()

            # Record batch manifest
            with engine.connect() as sql_conn:
                sql_conn.execute(text("""
                    INSERT INTO ops.ingest_batches 
                        (batch_id, source_id, table_name, target_zone, rows_count, checksum_sha256, watermark_end, loaded_at)
                    VALUES 
                        (:bid, :sid, :tbl, 'bronze', :rows, :chk, :wm, NOW());
                """), {
                    "bid": batch_id,
                    "sid": source_id,
                    "tbl": target_bronze_table,
                    "rows": batch_rows_count,
                    "chk": checksum,
                    "wm": chunk_wm
                })
                sql_conn.commit()

            # Advance watermark
            if wm_col and chunk_wm:
                commit_table_watermark(source_id, target_bronze_table, wm_col, chunk_wm, chunk_pk, batch_rows_count)

    return total_records, batch_ids
