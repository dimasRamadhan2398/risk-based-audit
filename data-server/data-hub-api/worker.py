"""
AuditSphere Data Hub — Background Pipeline Worker Daemon
Executes schema migrations on startup, continuously processes queued jobs,
streams data from client databases, runs set-based transforms, and handles health heartbeats.
"""
import os
import sys
import time
import signal
import socket
import json
import traceback
from sqlalchemy import text

from core.db import engine
from core.migrations import run_migrations
from core.hardware import get_hardware_profile
from jobs.queue import pop_next_job, update_job_progress, complete_job, fail_job
from ingestion.introspect import introspect_schema
from ingestion.bronze_ddl import ensure_bronze_table
from ingestion.extractors import stream_extract_table
from ingestion.loader import stream_load_to_bronze
from ingestion.watermark import get_table_watermark
from services.transform_service import (
    transform_source_bronze_to_silver,
    transform_source_silver_to_gold,
    feed_ai_training_pool
)

RUNNING = True

def handle_shutdown(signum, frame):
    global RUNNING
    print(f"\n[Worker] Received shutdown signal ({signum}). Exiting gracefully...")
    RUNNING = False

signal.signal(signal.SIGINT, handle_shutdown)
signal.signal(signal.SIGTERM, handle_shutdown)

def execute_ingest_job(job: dict):
    """Execute streaming ingestion for one or multiple registered sources."""
    job_id = job["job_id"]
    payload = job.get("payload", {})
    source_ids = payload.get("source_ids", [])
    run_transform = payload.get("run_transform", True)

    hw = get_hardware_profile()
    chunk_size = hw["chunk_size"]

    total_records_overall = 0
    results_summary = []

    for sid in source_ids:
        # Load source configuration
        with engine.connect() as conn:
            row = conn.execute(text(
                "SELECT * FROM bronze.registered_sources WHERE source_id = :sid"
            ), {"sid": sid}).fetchone()

        if not row:
            results_summary.append({"source_id": sid, "status": "error", "error": "Source not found"})
            continue

        src = dict(row._mapping)
        sname = src["name"]
        raw_mappings = src.get("data_mappings") or []
        mappings = json.loads(raw_mappings) if isinstance(raw_mappings, str) else raw_mappings
        active_mappings = [m for m in mappings if m.get("isActive")]

        update_job_progress(job_id, {
            "current_source": sname,
            "source_id": sid,
            "status": "introspecting"
        })

        # Schema introspection
        intro = introspect_schema(src)
        schema_tables = {t["tableName"]: t for t in intro.get("tables", [])}

        tables_to_ingest = []
        if active_mappings:
            for m in active_mappings:
                tbl_name = m["tableName"]
                table_meta = schema_tables.get(tbl_name, {})
                tables_to_ingest.append({
                    "mapping": m,
                    "tableName": tbl_name,
                    "columns": [c["name"] for c in table_meta.get("columns", [])],
                    "primaryKeys": table_meta.get("primaryKeys", []),
                    "suggestedWm": table_meta.get("suggestedWatermarkColumn")
                })
        else:
            # Fallback auto-ingest first 5 tables
            for tbl in intro.get("tables", [])[:5]:
                tables_to_ingest.append({
                    "mapping": {"tableName": tbl["tableName"], "isActive": True, "targetScope": "audit_features"},
                    "tableName": tbl["tableName"],
                    "columns": [c["name"] for c in tbl.get("columns", [])],
                    "primaryKeys": tbl.get("primaryKeys", []),
                    "suggestedWm": tbl.get("suggestedWatermarkColumn")
                })

        source_records = 0
        table_reports = {}

        for item in tables_to_ingest:
            tbl_name = item["tableName"]
            cols = item["columns"]
            if not cols:
                continue

            mapping = item["mapping"]
            wm_col = mapping.get("watermarkColumn") or item["suggestedWm"]
            pk_cols = item["primaryKeys"]

            update_job_progress(job_id, {
                "current_source": sname,
                "current_table": tbl_name,
                "records_loaded": total_records_overall + source_records
            })

            # Ensure Bronze table with hypertable and immutability trigger
            target_bronze_table = ensure_bronze_table(sname, tbl_name, schema_tables.get(tbl_name, {}).get("columns", []))

            # Retrieve last watermark
            last_wm, last_pk = get_table_watermark(sid, target_bronze_table)

            # Stream extract & load
            extractor = stream_extract_table(
                src=src,
                table_name=tbl_name,
                columns=cols,
                wm_col=wm_col,
                last_wm_val=last_wm,
                last_pk_val=last_pk,
                pk_cols=pk_cols,
                chunk_size=chunk_size
            )

            records_loaded, batch_ids = stream_load_to_bronze(
                source_id=sid,
                source_name=sname,
                target_bronze_table=target_bronze_table,
                column_names=cols,
                extractor_stream=extractor,
                wm_col=wm_col
            )

            source_records += records_loaded
            table_reports[tbl_name] = records_loaded

            # Run set-based transformations if requested
            if run_transform and records_loaded > 0:
                update_job_progress(job_id, {
                    "current_source": sname,
                    "current_table": tbl_name,
                    "status": "transforming"
                })
                # Silver and Gold set-based transforms
                transform_source_bronze_to_silver(engine, sid, sname, tbl_name, mapping, None)
                transform_source_silver_to_gold(engine, sid, sname, tbl_name, mapping)
                feed_ai_training_pool(engine, sid, sname, tbl_name, mapping)

        total_records_overall += source_records

        # Update source synced status
        with engine.connect() as conn:
            conn.execute(text("""
                UPDATE bronze.registered_sources 
                SET status = 'synced', last_sync_at = NOW(), records_synced = records_synced + :count, updated_at = NOW()
                WHERE source_id = :sid;
            """), {"count": source_records, "sid": sid})
            conn.commit()

        results_summary.append({
            "source_id": sid,
            "source_name": sname,
            "records": source_records,
            "tables": table_reports
        })

    complete_job(job_id, {
        "status": "success",
        "total_records": total_records_overall,
        "sources": results_summary
    })

def main_loop():
    worker_id = f"{socket.gethostname()}_{os.getpid()}"
    print(f"[Worker] Starting Data Hub Worker (ID: {worker_id})...")

    # Step 1: Run pending database migrations
    try:
        run_migrations(engine)
    except Exception as e:
        print(f"[Worker] Warning: Startup migrations failed: {e}")

    hw = get_hardware_profile()
    print(f"[Worker] Active Hardware Profile: {hw['profile']} (Chunk Size: {hw['chunk_size']:,} rows)")

    while RUNNING:
        try:
            job = pop_next_job(worker_id)
            if not job:
                time.sleep(1)
                continue

            print(f"[Worker] Processing job {job['job_id']} (Type: {job['job_type']}, Priority: {job['priority']})...")

            if job["job_type"] == "ingest":
                execute_ingest_job(job)
            else:
                print(f"[Worker] Unknown job type: {job['job_type']}")
                complete_job(job["job_id"], {"status": "unsupported_job_type"})

            print(f"[Worker] Completed job {job['job_id']}.")

        except Exception as e:
            tb = traceback.format_exc()
            print(f"[Worker] Error processing job: {e}\n{tb}")
            if 'job' in locals() and job:
                fail_job(job["job_id"], str(e))
            time.sleep(2)

    print("[Worker] Daemon stopped.")

if __name__ == "__main__":
    main_loop()
