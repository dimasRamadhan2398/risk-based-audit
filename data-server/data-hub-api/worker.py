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

from core.db import engine, cbs_engine
from core.migrations import run_migrations
from core.hardware import get_hardware_profile
from jobs.queue import pop_next_job, update_job_progress, complete_job, fail_job
from ingestion.introspect import introspect_schema
from ingestion.bronze_ddl import ensure_bronze_table
from ingestion.extractors import stream_extract_table
from ingestion.loader import stream_load_to_bronze
from ingestion.watermark import get_table_watermark
from services.etl_service import run_cbs_etl_pipeline
from services.caatt_engine import run_all_caatt_tests
from services.transform_service import (
    run_bronze_to_silver,
    run_silver_to_gold,
    transform_source_bronze_to_silver,
    transform_source_silver_to_gold,
    feed_ai_training_pool,
)
from transform.gold import project_external_silver_to_canonical_gold

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
        # Handle built-in CBS Simulator pipeline cleanly
        if sid == "cbs_simulator":
            update_job_progress(job_id, {
                "current_source": "Core Banking Simulator",
                "source_id": sid,
                "status": "extracting_cbs"
            })
            etl_res = run_cbs_etl_pipeline(cbs_engine, engine)
            cbs_total = int(etl_res.get("total_records", 0) or 0)
            if run_transform:
                update_job_progress(job_id, {
                    "current_source": "Core Banking Simulator",
                    "source_id": sid,
                    "status": "transforming"
                })
                run_bronze_to_silver(engine)
                run_silver_to_gold(engine)
                try:
                    run_all_caatt_tests(engine, source_id="cbs_simulator")
                except Exception as caatt_err:
                    print(f"[Worker:CAATT] Notice on cbs_simulator CAATT execution: {caatt_err}")
            total_records_overall += cbs_total
            results_summary.append({
                "source_id": sid,
                "source_name": "Core Banking Simulator",
                "records": cbs_total,
                "tables": etl_res.get("tables", {}),
            })
            continue

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

        # Include explicitly mapped tables first, then automatically merge remaining
        # discovered tables (up to 25 total) so CAATT Engine receives full DB coverage
        # without requiring manual CAATT mappings in Data Source UI.
        tables_to_ingest = []
        queued_table_names = set()
        for m in active_mappings:
            tbl_name = m["tableName"]
            if tbl_name in queued_table_names:
                continue
            queued_table_names.add(tbl_name)
            table_meta = schema_tables.get(tbl_name, {})
            tables_to_ingest.append({
                "mapping": m,
                "tableName": tbl_name,
                "columns": [c["name"] for c in table_meta.get("columns", [])],
                "primaryKeys": table_meta.get("primaryKeys", []),
                "suggestedWm": table_meta.get("suggestedWatermarkColumn")
            })

        for tbl in intro.get("tables", []):
            if len(tables_to_ingest) >= 25:
                break
            tbl_name = tbl["tableName"]
            if tbl_name in queued_table_names:
                continue
            queued_table_names.add(tbl_name)
            tables_to_ingest.append({
                "mapping": {"tableName": tbl_name, "isActive": True, "targetScope": "audit_features"},
                "tableName": tbl_name,
                "columns": [c["name"] for c in tbl.get("columns", [])],
                "primaryKeys": tbl.get("primaryKeys", []),
                "suggestedWm": tbl.get("suggestedWatermarkColumn")
            })

        source_records = 0
        table_reports = {}

        # Check if Gold CAATT transactions already exist for this source (handles prior-synced Silver backfill)
        has_gold_for_source = False
        try:
            with engine.connect() as conn:
                has_gold_for_source = conn.execute(text(
                    "SELECT 1 FROM gold.fact_transactions WHERE source_id = :sid LIMIT 1"
                ), {"sid": sid}).fetchone() is not None
        except Exception:
            pass

        for item in tables_to_ingest:
            tbl_name = item["tableName"]
            cols = item["columns"]
            if not cols:
                continue

            mapping = item["mapping"]
            raw_wm_col = mapping.get("watermarkColumn")
            if raw_wm_col in (None, "", "__none__"):
                wm_col = None if raw_wm_col == "__none__" else item["suggestedWm"]
            else:
                wm_col = raw_wm_col
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

            # Run set-based transformations when new rows loaded, Silver is empty for this table,
            # or Gold needs canonical backfill (recovers cleanly from prior watermark advances)
            has_silver_for_table = False
            try:
                with engine.connect() as conn:
                    has_silver_for_table = conn.execute(text(
                        "SELECT 1 FROM silver.external_source_data WHERE source_id = :sid AND table_name = :tbl AND NOT is_deleted LIMIT 1"
                    ), {"sid": sid, "tbl": tbl_name}).fetchone() is not None
            except Exception:
                pass

            if run_transform and (records_loaded > 0 or not has_silver_for_table or not has_gold_for_source):
                update_job_progress(job_id, {
                    "current_source": sname,
                    "current_table": tbl_name,
                    "status": "transforming"
                })
                if records_loaded > 0 or not has_silver_for_table:
                    transform_source_bronze_to_silver(engine, sid, sname, tbl_name, mapping, None)
                    transform_source_silver_to_gold(engine, sid, sname, tbl_name, mapping)
                    feed_ai_training_pool(engine, sid, sname, tbl_name, mapping)
                else:
                    project_external_silver_to_canonical_gold(
                        engine, sid, sname, tbl_name,
                        canonical_entity=(mapping.get("canonicalEntity") or "")
                    )

        effective_records = source_records
        try:
            with engine.connect() as conn:
                silver_count = conn.execute(text(
                    "SELECT COUNT(*) FROM silver.external_source_data WHERE source_id = :sid AND NOT is_deleted"
                ), {"sid": sid}).scalar()
                wm_sum = conn.execute(text(
                    "SELECT COALESCE(SUM(total_records_synced), 0) FROM ops.sync_watermarks WHERE source_id = :sid"
                ), {"sid": sid}).scalar()
                if int(silver_count or 0) > 0:
                    effective_records = max(int(source_records or 0), int(silver_count or 0))
                else:
                    effective_records = max(int(source_records or 0), int(wm_sum or 0))
        except Exception:
            pass

        # Automatically execute CAATT Engine tests for this connected source
        caatt_error = None
        if run_transform:
            update_job_progress(job_id, {
                "current_source": sname,
                "source_id": sid,
                "status": "running_caatt",
                "records_loaded": total_records_overall + effective_records,
                "records_processed": total_records_overall + effective_records,
            })
            try:
                caatt_summary = run_all_caatt_tests(engine, source_id=sid)
                print(f"[Worker:CAATT] Automated CAATT execution completed for '{sname}' ({sid}): {caatt_summary}")
            except Exception as caatt_err:
                caatt_error = str(caatt_err)
                print(f"[Worker:CAATT] Error executing CAATT for {sid}: {caatt_err}")
                traceback.print_exc()

        total_records_overall += effective_records

        # Update source synced status
        with engine.connect() as conn:
            conn.execute(text("""
                UPDATE bronze.registered_sources 
                SET status = 'synced', last_sync_at = NOW(), records_synced = GREATEST(COALESCE(records_synced, 0) + :new_count, :eff_count), updated_at = NOW()
                WHERE source_id = :sid;
            """), {"new_count": source_records, "eff_count": effective_records, "sid": sid})
            conn.commit()

        results_summary.append({
            "source_id": sid,
            "source_name": sname,
            "records": effective_records,
            "new_records": source_records,
            "tables": table_reports,
            "caatt_error": caatt_error,
        })

    complete_job(job_id, {
        "status": "success",
        "total_records": total_records_overall,
        "records_processed": total_records_overall,
        "sources": results_summary,
        "caatt_errors": [r for r in results_summary if r.get("caatt_error")],
    })

def check_and_enqueue_scheduled_sources():
    """
    Checks registered sources and enqueues ingestion jobs based on sync_schedule:
    - 'Real-time (CDC)': interval >= 30 seconds
    - 'Every 15 mins': interval >= 15 minutes (900s)
    - 'Hourly': interval >= 1 hour (3600s)
    - 'Daily': interval >= 24 hours (86400s)
    - 'Manual Only': never auto-enqueued
    """
    try:
        with engine.connect() as conn:
            rows = conn.execute(text("""
                SELECT s.source_id, s.name, s.sync_schedule, s.last_sync_at,
                       EXTRACT(EPOCH FROM (NOW() - COALESCE(s.last_sync_at, '1970-01-01'::timestamptz))) AS elapsed_seconds
                FROM bronze.registered_sources s
                WHERE s.sync_schedule IS NOT NULL AND s.sync_schedule != 'Manual Only';
            """)).fetchall()

            for r in rows:
                sched = (r.sync_schedule or "").strip()
                elapsed = r.elapsed_seconds if r.elapsed_seconds is not None else 999999
                
                is_due = False
                if sched == 'Real-time (CDC)' and elapsed >= 30:
                    is_due = True
                elif sched == 'Every 15 mins' and elapsed >= 900:
                    is_due = True
                elif sched == 'Hourly' and elapsed >= 3600:
                    is_due = True
                elif sched == 'Daily' and elapsed >= 86400:
                    is_due = True

                if is_due:
                    active = conn.execute(text("""
                        SELECT 1 FROM ops.jobs 
                        WHERE (source_id = :sid OR payload->'source_ids' ? :sid) AND status IN ('queued', 'running')
                        LIMIT 1;
                    """), {"sid": r.source_id}).fetchone()

                    if not active:
                        from jobs.queue import enqueue_job
                        enqueue_job(
                            job_type="ingest",
                            source_id=r.source_id,
                            source_name=r.name,
                            payload={"source_ids": [r.source_id], "run_transform": True},
                            priority=5 if sched == 'Real-time (CDC)' else 10
                        )
                        print(f"[Worker:Scheduler] Auto-enqueued ingest job for source '{r.name}' (Schedule: {sched})")
    except Exception:
        pass

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

    last_sched_check = 0.0

    while RUNNING:
        try:
            # Check scheduled sources every 10 seconds
            now = time.time()
            if now - last_sched_check >= 10.0:
                check_and_enqueue_scheduled_sources()
                last_sched_check = now

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
