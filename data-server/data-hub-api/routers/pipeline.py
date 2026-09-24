"""
Data Hub API — Pipeline Router
Endpoints for ETL pipeline management: trigger, status, health.
"""
from fastapi import APIRouter, HTTPException, BackgroundTasks
from sqlalchemy import text

from main import engine, cbs_engine
from services.etl_service import run_cbs_etl_pipeline
from services.transform_service import run_bronze_to_silver, run_silver_to_gold

router = APIRouter()

# Track pipeline state
pipeline_state = {"status": "idle", "last_run": None, "last_error": None, "records_processed": 0}


def _run_full_pipeline():
    """Background task: run the full ETL pipeline CBS → Bronze → Silver → Gold."""
    global pipeline_state
    pipeline_state["status"] = "running"

    try:
        # Step 1: CBS → Bronze
        cbs_result = run_cbs_etl_pipeline(cbs_engine, engine)
        pipeline_state["records_processed"] = cbs_result.get("total_records", 0)

        # Step 2: Bronze → Silver
        run_bronze_to_silver(engine)

        # Step 3: Silver → Gold
        run_silver_to_gold(engine)

        pipeline_state["status"] = "completed"
        pipeline_state["last_run"] = str(__import__("datetime").datetime.now())
        pipeline_state["last_error"] = None
        print(f"[Pipeline] Full ETL completed. Records: {pipeline_state['records_processed']}")

    except Exception as e:
        pipeline_state["status"] = "failed"
        pipeline_state["last_error"] = str(e)
        print(f"[Pipeline] ETL failed: {e}")


@router.post("/run")
def trigger_pipeline(background_tasks: BackgroundTasks):
    """Trigger the full ETL pipeline: CBS → Bronze → Silver → Gold."""
    if pipeline_state["status"] == "running":
        return {"status": "already_running", "message": "Pipeline is already running"}

    background_tasks.add_task(_run_full_pipeline)
    return {
        "status": "scheduled",
        "message": "Full ETL pipeline scheduled. Use GET /status to check progress.",
    }


@router.get("/status")
def get_pipeline_status():
    """Check current pipeline status."""
    return {
        "status": pipeline_state["status"],
        "last_run": pipeline_state["last_run"],
        "last_error": pipeline_state["last_error"],
        "records_processed": pipeline_state["records_processed"],
    }


@router.get("/health")
def pipeline_health():
    """Health check for all pipeline components."""
    health = {"datalake_db": False, "cbs_db": False, "bronze_tables": 0, "silver_tables": 0, "gold_tables": 0}

    try:
        with engine.connect() as conn:
            conn.execute(text("SELECT 1"))
            health["datalake_db"] = True

            for schema in ["bronze", "silver", "gold"]:
                r = conn.execute(text(
                    f"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '{schema}'"
                ))
                health[f"{schema}_tables"] = r.scalar() or 0
    except Exception:
        pass

    try:
        with cbs_engine.connect() as conn:
            conn.execute(text("SELECT 1"))
            health["cbs_db"] = True
    except Exception:
        pass

    return {"status": "ok" if health["datalake_db"] else "degraded", "health": health}
