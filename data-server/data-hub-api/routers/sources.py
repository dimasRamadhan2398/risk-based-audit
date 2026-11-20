"""
Data Hub API — Sources Router
Manages registered external data sources, multi-source ingestion queue, and real schema introspection.
"""
from typing import Optional, List, Any, Dict
from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel, model_validator
from sqlalchemy import text, create_engine
import json
import os
import urllib.request

from main import engine
from core.identifiers import validate_identifier, quote_identifier
from core.db import build_client_conn_string, get_read_only_client_conn, resolve_ssl_mode, validate_ssl_config, SSLConfigError
from ingestion.introspect import introspect_schema
from ingestion.profile import profile_client_table
from jobs.queue import enqueue_job, get_job_status

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
    ssl_enabled: Optional[bool] = None  # legacy flag: kept for compatibility (None/True -> require, False -> disable)
    ssl_mode: Optional[str] = None      # disable | require | verify-ca | verify-full (authoritative when set)
    ssl_root_cert: Optional[str] = None # PEM CA for verify-ca/verify-full; empty = system CA pool
    sync_schedule: str = "Manual Only"
    scopes: Optional[List[str]] = None  # ["risk_management","audit_features","qar_features","data_analytics"]
    data_mappings: Optional[list] = None

    @model_validator(mode="after")
    def _check_ssl(self):
        try:
            mode = resolve_ssl_mode(self.ssl_mode, self.ssl_enabled)
            validate_ssl_config(mode, self.ssl_root_cert)
        except SSLConfigError as e:
            raise ValueError(str(e))  # surfaces as HTTP 422
        self.ssl_mode = mode
        self.ssl_enabled = mode != "disable"
        self.ssl_root_cert = (self.ssl_root_cert or "").strip() or None
        if self.scopes is None:
            self.scopes = []
        if self.data_mappings is None:
            self.data_mappings = []
        return self

class IngestRequest(BaseModel):
    source_ids: Optional[List[str]] = [] # Can be empty if mode is "all"
    mode: str = "selected"               # "selected" | "all"
    run_transform: bool = True           # Auto Bronze -> Silver -> Gold?
    priority: int = 10                   # 5=high, 10=normal, 20=low

# ─── Source Registration Endpoints ───────────────────────────────────────────
@router.post("/register")
def register_source(req: SourceRegistration):
    """Register or update an external data source from Audit Server."""
    with engine.connect() as conn:
        stored = conn.execute(text("""
            INSERT INTO bronze.registered_sources
                (source_id, name, source_type, host, port, database_name,
                 username, password_encrypted, ssl_enabled, ssl_mode, ssl_root_cert, sync_schedule, scopes, data_mappings, updated_at)
            VALUES (:sid, :name, :stype, :host, :port, :db,
                    :user, :pass, :ssl, :ssl_mode, :ssl_root_cert, :schedule, CAST(:scopes AS jsonb), CAST(:mappings AS jsonb), NOW())
            ON CONFLICT (source_id) DO UPDATE SET
                name = EXCLUDED.name,
                source_type = EXCLUDED.source_type,
                host = EXCLUDED.host,
                port = EXCLUDED.port,
                database_name = EXCLUDED.database_name,
                username = EXCLUDED.username,
                password_encrypted = EXCLUDED.password_encrypted,
                ssl_enabled = EXCLUDED.ssl_enabled,
                ssl_mode = EXCLUDED.ssl_mode,
                ssl_root_cert = EXCLUDED.ssl_root_cert,
                sync_schedule = EXCLUDED.sync_schedule,
                scopes = EXCLUDED.scopes,
                data_mappings = EXCLUDED.data_mappings,
                updated_at = NOW()
            RETURNING ssl_mode, ssl_enabled, (ssl_root_cert IS NOT NULL) AS has_root_cert;
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
            "ssl_mode": req.ssl_mode,
            "ssl_root_cert": req.ssl_root_cert,
            "schedule": req.sync_schedule,
            "scopes": json.dumps(req.scopes),
            "mappings": json.dumps(req.data_mappings)
        })
        row = stored.one()
        conn.commit()
    # Echo the value actually persisted (read back via RETURNING, not the request) so the
    # caller (master-service) can detect an older Data Hub that ignores ssl_mode.
    return {
        "status": "success",
        "message": f"Source '{req.name}' successfully registered",
        "source_id": req.source_id,
        "ssl_mode": row.ssl_mode,
        "ssl_enabled": row.ssl_enabled,
        "has_ssl_root_cert": row.has_root_cert,
    }

@router.get("")
def list_sources():
    """List all registered external sources."""
    with engine.connect() as conn:
        result = conn.execute(text("SELECT * FROM bronze.registered_sources ORDER BY created_at DESC;"))
        sources = [dict(r._mapping) for r in result]
    for s in sources:  # never return secrets (DB password, CA PEM) to callers
        s.pop("password_encrypted", None)
        s["has_ssl_root_cert"] = bool(s.pop("ssl_root_cert", None))
    return {"status": "success", "data": sources}

@router.delete("/{source_id}")
def deregister_source(source_id: str):
    """Remove a registered source."""
    with engine.connect() as conn:
        conn.execute(text("DELETE FROM bronze.registered_sources WHERE source_id = :sid;"), {"sid": source_id})
        conn.commit()
    return {"status": "success", "message": f"Source {source_id} deregistered"}

# ─── Multi-Source Ingestion Endpoints (Persistent Job Queue) ─────────────────
@router.post("/ingest")
def trigger_multi_ingest(req: IngestRequest):
    """
    Enqueue multi-source ingestion job to ops.jobs queue.
    Executed asynchronously by datahub-worker daemon with full crash recovery.
    """
    if req.mode == "all":
        with engine.connect() as conn:
            result = conn.execute(text("SELECT source_id FROM bronze.registered_sources;"))
            source_ids = [str(r[0]) for r in result]
    else:
        source_ids = req.source_ids or []

    if not source_ids:
        raise HTTPException(status_code=400, detail="No sources specified or registered for ingestion")

    job_id = enqueue_job(
        job_type="ingest",
        source_id=source_ids[0] if len(source_ids) == 1 else "multiple",
        source_name="Multi-Source Batch",
        payload={
            "source_ids": source_ids,
            "run_transform": req.run_transform
        },
        priority=req.priority
    )

    return {
        "status": "scheduled",
        "job_id": job_id,
        "sources_count": len(source_ids),
        "message": "Ingestion job queued for background streaming worker"
    }

@router.get("/ingest/{job_id}/status")
def get_ingest_job_status(job_id: str):
    """Check persistent progress and status of an ingest job."""
    status_info = get_job_status(job_id)
    if not status_info:
        raise HTTPException(status_code=404, detail="Job not found")
    result_data = status_info.get("result") or {}
    progress = status_info.get("progress") or {}
    rec_val = 0
    if isinstance(result_data, dict):
        rec_val = result_data.get("records_processed", result_data.get("total_records", 0)) or 0
    if not rec_val and isinstance(progress, dict):
        rec_val = progress.get("records_processed", progress.get("total_records", progress.get("records_loaded", 0))) or 0
    status_info["records_processed"] = int(rec_val)
    return status_info

@router.post("/{source_id}/full-resync")
def reset_source_watermark(source_id: str):
    """Reset watermark for a source to trigger complete re-extraction."""
    with engine.connect() as conn:
        conn.execute(text("""
            DELETE FROM ops.sync_watermarks WHERE source_id = :sid;
        """), {"sid": source_id})
        conn.commit()
    return {"status": "success", "message": f"Watermark reset for source {source_id}. Next sync will be full."}

@router.get("/{source_id}/batches")
def list_source_batch_manifests(source_id: str, limit: int = Query(20, le=100)):
    """Retrieve audit trail batch manifests (checksums, records loaded)."""
    with engine.connect() as conn:
        res = conn.execute(text("""
            SELECT batch_id, table_name, rows_count, checksum_sha256, watermark_end, loaded_at
            FROM ops.ingest_batches
            WHERE source_id = :sid
            ORDER BY loaded_at DESC
            LIMIT :lim;
        """), {"sid": source_id, "lim": limit})
        batches = [dict(r._mapping) for r in res]
    return {"status": "success", "data": batches}

@router.post("/{source_id}/reconcile")
def run_source_reconciliation(source_id: str):
    """Trigger read-only delete detection reconciliation against registered client database."""
    from reconcile.delete_detection import reconcile_source_deletes
    return reconcile_source_deletes(source_id)

@router.get("/{source_id}/reconcile-runs")
def list_source_reconcile_runs(source_id: str, limit: int = Query(20, le=100)):
    """List historical delete detection and reconciliation runs."""
    with engine.connect() as conn:
        res = conn.execute(text("""
            SELECT run_id, table_name, buckets_checked, buckets_mismatched, deletes_detected, started_at, completed_at, status
            FROM ops.reconcile_runs
            WHERE source_id = :sid
            ORDER BY started_at DESC
            LIMIT :lim;
        """), {"sid": source_id, "lim": limit})
        runs = [dict(r._mapping) for r in res]
    return {"status": "success", "data": runs}

# ─── Fast Schema Introspection ────────────────────────────────────────────────
@router.get("/{source_id}/schema")
def introspect_source_schema(source_id: str):
    """Connect to registered source (Read-Only) and return table catalog in sub-second."""
    with engine.connect() as conn:
        row = conn.execute(text(
            "SELECT * FROM bronze.registered_sources WHERE source_id = :sid;"
        ), {"sid": source_id}).fetchone()

    if not row:
        raise HTTPException(status_code=404, detail="Source connection not registered in Data Hub")

    src = dict(row._mapping)
    try:
        catalog = introspect_schema(src)
        return {"success": True, "data": catalog}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to introspect schema: {str(e)}")

@router.get("/{source_id}/preview/{table_name}")
def preview_source_table(source_id: str, table_name: str, limit: int = 50):
    """Preview stratified sample records and real table profile from client database safely."""
    validate_identifier(table_name)

    with engine.connect() as conn:
        row = conn.execute(text(
            "SELECT * FROM bronze.registered_sources WHERE source_id = :sid;"
        ), {"sid": source_id}).fetchone()

    if not row:
        raise HTTPException(status_code=404, detail="Source not registered")

    src = dict(row._mapping)
    try:
        with get_read_only_client_conn(src) as conn:
            data = profile_client_table(conn, table_name, sample_limit=limit)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to preview table {table_name}: {str(e)}")

    return {"success": True, "data": data}


@router.get("/{source_id}/profile/{table_name}")
def get_source_table_profile(source_id: str, table_name: str, limit: int = 50):
    """Return exact row count, real metric aggregates, dimension breakdown, and stratified sample rows."""
    return preview_source_table(source_id=source_id, table_name=table_name, limit=limit)

def notify_ai_engine(source_id: str, source_name: str):
    """Notify AI engine that new data is available in gold.ai_training_pool with authentication."""
    ai_url = os.getenv("AI_ENGINE_URL", "http://ai-engine:8000")
    api_key = os.getenv("AI_ENGINE_API_KEY", "dev-ai-api-key")
    endpoint = f"{ai_url}/retrain/auto"
    try:
        req = urllib.request.Request(
            endpoint,
            data=json.dumps({"source_id": source_id, "source_name": source_name}).encode("utf-8"),
            headers={
                "Content-Type": "application/json",
                "X-API-Key": api_key
            },
            method="POST"
        )
        with urllib.request.urlopen(req, timeout=3) as resp:
            pass
    except Exception as e:
        print(f"[Sources] Notice: AI Engine auto-retrain trigger: {e}")
