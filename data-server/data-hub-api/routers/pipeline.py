"""
Data Hub API — Pipeline Router
Endpoints for ETL pipeline management: trigger, status, health.
Uses persistent ops.jobs queue for background orchestration.
"""
from typing import Optional, List
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from sqlalchemy import text

from core.db import engine, cbs_engine
from jobs.queue import enqueue_job, get_job_status

router = APIRouter()

class PipelineRequest(BaseModel):
    source_ids: Optional[List[str]] = None
    mode: str = "cbs"  # "cbs" | "selected" | "all"
    run_transform: bool = True
    priority: int = 10

@router.post("/run")
def trigger_pipeline(req: Optional[PipelineRequest] = None):
    """Trigger the ETL pipeline asynchronously via ops.jobs queue."""
    mode = req.mode if req else "cbs"
    source_ids = req.source_ids if req else []
    run_transform = req.run_transform if req else True
    priority = req.priority if req else 10

    if mode in ("selected", "all") or (source_ids and len(source_ids) > 0):
        from routers.sources import trigger_multi_ingest, IngestRequest
        ingest_req = IngestRequest(
            source_ids=source_ids or [],
            mode=mode,
            run_transform=run_transform,
            priority=priority
        )
        return trigger_multi_ingest(ingest_req)

    # CBS pipeline queued
    job_id = enqueue_job(
        job_type="ingest",
        source_id="cbs_simulator",
        source_name="Core Banking Simulator",
        payload={
            "source_ids": ["cbs_simulator"],
            "mode": "cbs",
            "run_transform": run_transform
        },
        priority=priority
    )

    return {
        "status": "scheduled",
        "job_id": job_id,
        "message": "Full ETL pipeline scheduled. Use GET /api/v1/pipeline/status?job_id=... to check progress.",
    }


@router.get("/status")
def get_pipeline_status(job_id: Optional[str] = None):
    """Check current or latest pipeline status."""
    if job_id:
        st = get_job_status(job_id)
        if st:
            return st

    # Fetch latest job from ops.jobs
    with engine.connect() as conn:
        row = conn.execute(text("""
            SELECT job_id, job_type, status, progress, error_message, started_at, completed_at, created_at
            FROM ops.jobs
            ORDER BY created_at DESC
            LIMIT 1;
        """)).fetchone()

        if row:
            res = dict(row._mapping)
            if isinstance(res.get("progress"), str):
                import json
                res["progress"] = json.loads(res["progress"])
            return res

    return {"status": "idle", "message": "No jobs executed yet"}


@router.get("/health")
def pipeline_health():
    """Health check for all pipeline components."""
    health = {"datalake_db": False, "cbs_db": False, "bronze_tables": 0, "silver_tables": 0, "gold_tables": 0, "ops_jobs": 0}

    try:
        with engine.connect() as conn:
            conn.execute(text("SELECT 1"))
            health["datalake_db"] = True

            for schema in ["bronze", "silver", "gold"]:
                r = conn.execute(text(
                    f"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '{schema}'"
                ))
                health[f"{schema}_tables"] = r.scalar() or 0

            r_jobs = conn.execute(text("SELECT COUNT(*) FROM ops.jobs;"))
            health["ops_jobs"] = r_jobs.scalar() or 0
    except Exception:
        pass

    try:
        with cbs_engine.connect() as conn:
            conn.execute(text("SELECT 1"))
            health["cbs_db"] = True
    except Exception:
        pass

    return {"status": "ok" if health["datalake_db"] else "degraded", "health": health}
