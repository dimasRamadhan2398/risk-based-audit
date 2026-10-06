"""
AuditSphere Data Hub — Persistent Job Queue
PostgreSQL FOR UPDATE SKIP LOCKED job queue for fair, resilient multi-source pipeline orchestration.
"""
import uuid
import json
from datetime import datetime, timezone
from typing import Optional, Dict, Any, List
from sqlalchemy import text
from core.db import engine

def enqueue_job(job_type: str, source_id: Optional[str] = None, source_name: Optional[str] = None,
                payload: Optional[Dict[str, Any]] = None, priority: int = 10) -> str:
    """Add a new job to the ops.jobs queue."""
    job_id = str(uuid.uuid4())[:12]
    with engine.connect() as conn:
        conn.execute(text("""
            INSERT INTO ops.jobs 
                (job_id, job_type, source_id, source_name, priority, status, payload, progress, created_at, updated_at)
            VALUES 
                (:jid, :jtype, :sid, :sname, :prio, 'queued', :payload::jsonb, '{}'::jsonb, NOW(), NOW());
        """), {
            "jid": job_id,
            "jtype": job_type,
            "sid": source_id,
            "sname": source_name,
            "prio": priority,
            "payload": json.dumps(payload or {})
        })
        conn.commit()
    return job_id

def pop_next_job(worker_id: str) -> Optional[Dict[str, Any]]:
    """Atomically claim the next queued job using FOR UPDATE SKIP LOCKED."""
    with engine.connect() as conn:
        # Atomic lock & claim
        row = conn.execute(text("""
            SELECT job_id, job_type, source_id, source_name, payload, priority
            FROM ops.jobs
            WHERE status = 'queued'
            ORDER BY priority ASC, created_at ASC
            FOR UPDATE SKIP LOCKED
            LIMIT 1;
        """)).fetchone()

        if not row:
            return None

        job_dict = dict(row._mapping)
        job_id = job_dict["job_id"]

        conn.execute(text("""
            UPDATE ops.jobs
            SET status = 'running',
                worker_id = :wid,
                started_at = NOW(),
                heartbeat_at = NOW(),
                updated_at = NOW()
            WHERE job_id = :jid;
        """), {"wid": worker_id, "jid": job_id})
        conn.commit()

        # Parse payload
        if isinstance(job_dict.get("payload"), str):
            job_dict["payload"] = json.loads(job_dict["payload"])

        return job_dict

def update_job_progress(job_id: str, progress: Dict[str, Any]):
    """Update running job progress and heartbeat."""
    with engine.connect() as conn:
        conn.execute(text("""
            UPDATE ops.jobs
            SET progress = :prog::jsonb,
                heartbeat_at = NOW(),
                updated_at = NOW()
            WHERE job_id = :jid;
        """), {"jid": job_id, "prog": json.dumps(progress)})
        conn.commit()

def complete_job(job_id: str, result_summary: Optional[Dict[str, Any]] = None):
    """Mark job as successfully completed."""
    with engine.connect() as conn:
        conn.execute(text("""
            UPDATE ops.jobs
            SET status = 'completed',
                progress = COALESCE(:res::jsonb, progress),
                completed_at = NOW(),
                updated_at = NOW()
            WHERE job_id = :jid;
        """), {"jid": job_id, "res": json.dumps(result_summary or {})})
        conn.commit()

def fail_job(job_id: str, error_message: str):
    """Mark job as failed with error trace."""
    with engine.connect() as conn:
        conn.execute(text("""
            UPDATE ops.jobs
            SET status = 'failed',
                error_message = :err,
                completed_at = NOW(),
                updated_at = NOW()
            WHERE job_id = :jid;
        """), {"jid": job_id, "err": error_message})
        conn.commit()

def get_job_status(job_id: str) -> Optional[Dict[str, Any]]:
    """Retrieve full status of a job."""
    with engine.connect() as conn:
        row = conn.execute(text("""
            SELECT job_id, job_type, source_id, source_name, priority, status, progress, 
                   error_message, started_at, completed_at, created_at
            FROM ops.jobs
            WHERE job_id = :jid;
        """), {"jid": job_id}).fetchone()

        if not row:
            return None
        res = dict(row._mapping)
        if isinstance(res.get("progress"), str):
            res["progress"] = json.loads(res["progress"])
        return res
