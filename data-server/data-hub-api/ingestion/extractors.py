"""
AuditSphere Data Hub — Streaming Extractors
Extracts data from client databases in streaming chunks under strict Read-Only and timeout constraints.
"""
from typing import Generator, List, Dict, Any, Tuple, Optional
from core.db import get_read_only_client_conn
from core.identifiers import quote_identifier, quote_table_name

def stream_extract_table(
    src: dict,
    table_name: str,
    columns: List[str],
    wm_col: Optional[str] = None,
    last_wm_val: Optional[str] = None,
    last_pk_val: Optional[str] = None,
    pk_cols: Optional[List[str]] = None,
    chunk_size: int = 50000
) -> Generator[Tuple[List[Tuple], Optional[str], Optional[str]], None, None]:
    """
    Stream chunks of rows from client database.
    Yields: (rows_chunk, max_wm_val_in_chunk, max_pk_val_in_chunk)
    """
    quoted_cols = ", ".join(quote_identifier(c) for c in columns)
    quoted_tbl = quote_identifier(table_name)

    where_clauses = []
    params = []

    # Keyset watermark filter
    pk_col = pk_cols[0] if pk_cols else columns[0]
    if wm_col and last_wm_val:
        if pk_col and last_pk_val:
            where_clauses.append(
                f"({quote_identifier(wm_col)} > %s OR ({quote_identifier(wm_col)} = %s AND {quote_identifier(pk_col)} > %s))"
            )
            params.extend([last_wm_val, last_wm_val, last_pk_val])
        else:
            where_clauses.append(f"{quote_identifier(wm_col)} > %s")
            params.append(last_wm_val)

    where_sql = f"WHERE {' AND '.join(where_clauses)}" if where_clauses else ""
    order_sql = f"ORDER BY {quote_identifier(wm_col)} ASC, {quote_identifier(pk_col)} ASC" if wm_col else ""

    query = f"SELECT {quoted_cols} FROM {quoted_tbl} {where_sql} {order_sql};"

    wm_idx = columns.index(wm_col) if wm_col and wm_col in columns else None
    pk_idx = columns.index(pk_col) if pk_col and pk_col in columns else 0

    with get_read_only_client_conn(src) as conn:
        with conn.cursor(name="auditsphere_stream_cursor") as cur:
            cur.itersize = chunk_size
            cur.execute(query, params)

            while True:
                rows = cur.fetchmany(chunk_size)
                if not rows:
                    break

                max_wm = str(rows[-1][wm_idx]) if wm_idx is not None and rows[-1][wm_idx] is not None else None
                max_pk = str(rows[-1][pk_idx]) if pk_idx is not None and rows[-1][pk_idx] is not None else None

                yield rows, max_wm, max_pk
