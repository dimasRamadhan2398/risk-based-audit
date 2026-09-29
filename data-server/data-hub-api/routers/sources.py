"""
Data Hub API — Sources Router
Manages registered external data sources, multi-source ingestion, and real schema introspection.
"""
from typing import Optional, List, Any, Dict
from fastapi import APIRouter, HTTPException, BackgroundTasks
from pydantic import BaseModel
from sqlalchemy import text, create_engine
import pandas as pd
import uuid
import json
import os
import urllib.request

from main import engine
from services.transform_service import (
    transform_source_bronze_to_silver,
    transform_source_silver_to_gold,
    feed_ai_training_pool
)

router = APIRouter()

# ─── Models ──────────────────────────────────────────────────────────────────
class SourceRegistration(BaseModel):
    source_id: str                      # UUID from Audit Server
    name: str
    source_type: str = "postgres"       # postgres, mysql, oracle, mssql
    host: str
    port: int = 5432
    database_name: str
    username: Optional[str] = None
    password: Optional[str] = None
    ssl_enabled: bool = True
    sync_schedule: str = "Manual Only"
    scopes: List[str] = []              # ["risk_management","audit_features","qar_features","data_analytics"]
    data_mappings: list = []

class IngestRequest(BaseModel):
    source_ids: Optional[List[str]] = [] # Can be empty if mode is "all"
    mode: str = "selected"               # "selected" | "all"
    run_transform: bool = True           # Auto Bronze -> Silver -> Gold?

# ─── Track multi-source pipeline progress ────────────────────────────────────
pipeline_progress: Dict[str, Any] = {}

def _build_connection_string(src: dict) -> str:
    stype = (src.get("source_type") or "postgres").lower()
    host = src.get("host", "localhost")
    port = src.get("port", 5432)
    db = src.get("database_name", "")
    user = src.get("username", "")
    pwd = src.get("password_encrypted", "") or ""

    auth = f"{user}:{pwd}@" if user or pwd else ""

    if "postgres" in stype:
        ssl_mode = "?sslmode=require" if src.get("ssl_enabled", True) else "?sslmode=disable"
        # If localhost or 127.0.0.1 or docker internal, disable ssl requirement by default for stability
        if host in ("localhost", "127.0.0.1", "cbs-simulator", "datalake-db", "host.docker.internal"):
            ssl_mode = "?sslmode=disable"
        return f"postgresql://{auth}{host}:{port}/{db}{ssl_mode}"
    elif "mysql" in stype:
        return f"mysql+pymysql://{auth}{host}:{port}/{db}"
    elif "mssql" in stype:
        return f"mssql+pyodbc://{auth}{host}:{port}/{db}?driver=ODBC+Driver+17+for+SQL+Server"
    elif "oracle" in stype:
        return f"oracle+cx_oracle://{auth}{host}:{port}/?service_name={db}"
    else:
        return f"postgresql://{auth}{host}:{port}/{db}"


# ─── Source Registration Endpoints ───────────────────────────────────────────
@router.post("/register")
def register_source(req: SourceRegistration):
    """Register or update an external data source from Audit Server."""
    with engine.connect() as conn:
        conn.execute(text("""
            INSERT INTO bronze.registered_sources
                (source_id, name, source_type, host, port, database_name,
                 username, password_encrypted, ssl_enabled, sync_schedule, scopes, data_mappings, updated_at)
            VALUES (:sid, :name, :stype, :host, :port, :db,
                    :user, :pass, :ssl, :schedule, :scopes::jsonb, :mappings::jsonb, NOW())
            ON CONFLICT (source_id) DO UPDATE SET
                name = EXCLUDED.name,
                source_type = EXCLUDED.source_type,
                host = EXCLUDED.host,
                port = EXCLUDED.port,
                database_name = EXCLUDED.database_name,
                username = EXCLUDED.username,
                password_encrypted = EXCLUDED.password_encrypted,
                ssl_enabled = EXCLUDED.ssl_enabled,
                sync_schedule = EXCLUDED.sync_schedule,
                scopes = EXCLUDED.scopes,
                data_mappings = EXCLUDED.data_mappings,
                updated_at = NOW()
        """), {
            "sid": req.source_id,
            "name": req.name,
            "stype": req.source_type,
            "host": req.host,
            "port": req.port,
            "db": req.database_name,
            "user": req.username,
            "pass": req.password,
            "ssl": req.ssl_enabled,
            "schedule": req.sync_schedule,
            "scopes": json.dumps(req.scopes),
            "mappings": json.dumps(req.data_mappings)
        })
        conn.commit()
    return {"status": "success", "message": f"Source '{req.name}' successfully registered"}


@router.get("")
def list_sources():
    """List all registered external sources."""
    with engine.connect() as conn:
        result = conn.execute(text("SELECT * FROM bronze.registered_sources ORDER BY created_at DESC"))
        sources = [dict(r._mapping) for r in result]
    return {"status": "success", "data": sources}


@router.delete("/{source_id}")
def deregister_source(source_id: str):
    """Remove a registered source."""
    with engine.connect() as conn:
        conn.execute(text("DELETE FROM bronze.registered_sources WHERE source_id = :sid"), {"sid": source_id})
        conn.commit()
    return {"status": "success", "message": f"Source {source_id} deregistered"}


# ─── Multi-Source Ingestion Endpoints ─────────────────────────────────────────
@router.post("/ingest")
def trigger_multi_ingest(req: IngestRequest, background_tasks: BackgroundTasks):
    """
    Trigger ingestion for 1, selected, or ALL registered sources.
    Executes sequentially in background with progress tracking.
    """
    job_id = str(uuid.uuid4())[:8]

    if req.mode == "all":
        with engine.connect() as conn:
            result = conn.execute(text("SELECT source_id FROM bronze.registered_sources"))
            source_ids = [str(r[0]) for r in result]
    else:
        source_ids = req.source_ids or []

    if not source_ids:
        raise HTTPException(status_code=400, detail="No sources specified or registered for ingestion")

    pipeline_progress[job_id] = {
        "job_id": job_id,
        "total": len(source_ids),
        "completed": 0,
        "current_source": None,
        "status": "running",
        "results": []
    }

    background_tasks.add_task(_run_multi_source_pipeline, job_id, source_ids, req.run_transform)
    return {"status": "scheduled", "job_id": job_id, "sources_count": len(source_ids)}


@router.get("/ingest/{job_id}/status")
def get_ingest_status(job_id: str):
    """Check progress of a multi-source ingest job."""
    if job_id not in pipeline_progress:
        raise HTTPException(status_code=404, detail="Job not found")
    return pipeline_progress[job_id]


def _run_multi_source_pipeline(job_id: str, source_ids: list, run_transform: bool):
    """Background task: sequential ingest for each source."""
    for sid in source_ids:
        pipeline_progress[job_id]["current_source"] = sid
        try:
            result = _ingest_single_source(sid, run_transform)
            pipeline_progress[job_id]["results"].append({"source_id": sid, **result})
        except Exception as e:
            pipeline_progress[job_id]["results"].append({
                "source_id": sid,
                "status": "failed",
                "error": str(e)
            })
        pipeline_progress[job_id]["completed"] += 1

    pipeline_progress[job_id]["status"] = "completed"
    pipeline_progress[job_id]["current_source"] = None


def _ingest_single_source(source_id: str, run_transform: bool) -> dict:
    """Ingest data from a single registered source into Bronze zone, then Silver and Gold."""
    # 1. Read source config
    with engine.connect() as conn:
        row = conn.execute(text(
            "SELECT * FROM bronze.registered_sources WHERE source_id = :sid"
        ), {"sid": source_id}).fetchone()

    if not row:
        return {"status": "error", "error": "Source not registered in Data Hub"}

    src = dict(row._mapping)
    raw_mappings = src.get("data_mappings") or []
    mappings = json.loads(raw_mappings) if isinstance(raw_mappings, str) else raw_mappings

    raw_scopes = src.get("scopes") or []
    scopes = json.loads(raw_scopes) if isinstance(raw_scopes, str) else raw_scopes

    active_mappings = [m for m in mappings if m.get("isActive")]

    # If no specific active mappings, discover tables or fallback
    conn_str = _build_connection_string(src)
    src_engine = create_engine(conn_str, pool_pre_ping=True)

    total_records = 0
    table_results = {}
    has_analytics_scope = "data_analytics" in scopes

    try:
        tables_to_ingest = []
        if active_mappings:
            tables_to_ingest = active_mappings
        else:
            # Auto-ingest first 5 public base tables if no mappings defined
            with src_engine.connect() as s_conn:
                res = s_conn.execute(text("""
                    SELECT table_name FROM information_schema.tables 
                    WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
                    LIMIT 5
                """))
                tables_to_ingest = [{"tableName": r[0], "isActive": True, "targetScope": "audit_features"} for r in res]

        for mapping in tables_to_ingest:
            table = mapping["tableName"]
            try:
                df = pd.read_sql(f"SELECT * FROM {table}", src_engine)
                df["_source"] = src["name"]
                df["_source_id"] = source_id
                df["_loaded_at"] = pd.Timestamp.now()

                clean_src_name = "".join(c if c.isalnum() else "_" for c in src["name"].lower())
                target_table = f"{clean_src_name}_{table}".lower()

                df.to_sql(target_table, engine, schema="bronze", if_exists="replace", index=False, method="multi", chunksize=1000)

                rec_count = len(df)
                total_records += rec_count
                table_results[table] = rec_count

                if run_transform:
                    # Silver transformation
                    transform_source_bronze_to_silver(engine, source_id, src["name"], table, mapping, df)
                    # Gold transformation
                    transform_source_silver_to_gold(engine, source_id, src["name"], table, mapping)

                    # Check table-level or connection-level data_analytics scope
                    table_scope = mapping.get("targetScope", "")
                    if table_scope == "data_analytics" or has_analytics_scope:
                        feed_ai_training_pool(engine, source_id, src["name"], table, mapping)

            except Exception as ex:
                table_results[table] = f"ERROR: {str(ex)}"

        # If data_analytics scope present, notify AI Engine
        if has_analytics_scope or any(m.get("targetScope") == "data_analytics" for m in tables_to_ingest):
            _notify_ai_engine(source_id, src["name"])

    finally:
        src_engine.dispose()

    # Update bronze.registered_sources record
    with engine.connect() as conn:
        conn.execute(text("""
            UPDATE bronze.registered_sources 
            SET status = 'synced', last_sync_at = NOW(), records_synced = :count, updated_at = NOW()
            WHERE source_id = :sid
        """), {"count": total_records, "sid": source_id})
        conn.commit()

    return {"status": "success", "total_records": total_records, "tables": table_results}


def _notify_ai_engine(source_id: str, source_name: str):
    """Notify AI engine that new data is available in gold.ai_training_pool."""
    ai_url = os.getenv("AI_ENGINE_URL", "http://ai-engine:8200")
    endpoint = f"{ai_url}/retrain/auto"
    try:
        req = urllib.request.Request(
            endpoint,
            data=json.dumps({"source_id": source_id, "source_name": source_name}).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST"
        )
        with urllib.request.urlopen(req, timeout=3) as resp:
            pass
    except Exception as e:
        print(f"[Sources] Notice: AI Engine auto-retrain trigger: {e}")


# ─── Schema Introspection (Task 4.1) ──────────────────────────────────────────
@router.get("/{source_id}/schema")
def introspect_source_schema(source_id: str):
    """Connect to a registered source and return real database catalog schema."""
    with engine.connect() as conn:
        row = conn.execute(text(
            "SELECT * FROM bronze.registered_sources WHERE source_id = :sid"
        ), {"sid": source_id}).fetchone()

    if not row:
        raise HTTPException(status_code=404, detail="Source connection not registered in Data Hub")

    src = dict(row._mapping)
    conn_str = _build_connection_string(src)
    src_engine = create_engine(conn_str, pool_pre_ping=True)

    tables = []
    try:
        with src_engine.connect() as src_conn:
            # Query base tables
            tbl_res = src_conn.execute(text("""
                SELECT table_name FROM information_schema.tables 
                WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
                ORDER BY table_name
            """))

            for tbl_row in tbl_res:
                tbl_name = tbl_row[0]
                # Columns metadata
                cols_res = src_conn.execute(text("""
                    SELECT c.column_name, c.data_type, c.is_nullable,
                           CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN TRUE ELSE FALSE END as is_primary
                    FROM information_schema.columns c
                    LEFT JOIN information_schema.key_column_usage kcu 
                        ON c.column_name = kcu.column_name AND c.table_name = kcu.table_name
                    LEFT JOIN information_schema.table_constraints tc 
                        ON kcu.constraint_name = tc.constraint_name AND tc.constraint_type = 'PRIMARY KEY'
                    WHERE c.table_name = :tbl AND c.table_schema = 'public'
                    ORDER BY c.ordinal_position
                """), {"tbl": tbl_name})

                columns = [
                    {
                        "name": c[0],
                        "dataType": c[1],
                        "isNullable": c[2] == "YES",
                        "isPrimary": bool(c[3])
                    }
                    for c in cols_res
                ]

                # Estimated or exact row count
                try:
                    cnt_res = src_conn.execute(text(f"SELECT COUNT(*) FROM {tbl_name}"))
                    row_count = cnt_res.scalar() or 0
                except Exception:
                    row_count = 0

                tables.append({
                    "tableName": tbl_name,
                    "rowCount": row_count,
                    "columnCount": len(columns),
                    "description": f"Real catalog table '{tbl_name}' from {src['database_name']}",
                    "columns": columns
                })
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to introspect source schema: {str(e)}")
    finally:
        src_engine.dispose()

    return {
        "success": True,
        "data": {
            "tables": tables,
            "database": src["database_name"],
            "source_name": src["name"]
        }
    }


@router.get("/{source_id}/preview/{table_name}")
def preview_source_table(source_id: str, table_name: str, limit: int = 10):
    """Preview live records from a specific table in the registered source."""
    with engine.connect() as conn:
        row = conn.execute(text(
            "SELECT * FROM bronze.registered_sources WHERE source_id = :sid"
        ), {"sid": source_id}).fetchone()

    if not row:
        raise HTTPException(status_code=404, detail="Source not registered")

    src = dict(row._mapping)
    conn_str = _build_connection_string(src)
    src_engine = create_engine(conn_str, pool_pre_ping=True)

    rows = []
    try:
        with src_engine.connect() as src_conn:
            res = src_conn.execute(text(f"SELECT * FROM {table_name} LIMIT :lim"), {"lim": limit})
            for r in res:
                rows.append(dict(r._mapping))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to preview table {table_name}: {str(e)}")
    finally:
        src_engine.dispose()

    return {"success": True, "data": {"tableName": table_name, "rows": rows}}
