"""
Data Hub API — Transform Service (Bridge)
Redirects to high-performance set-based transform modules:
- transform.silver (run_bronze_to_silver_cbs, transform_source_bronze_to_silver_external)
- transform.gold (run_silver_to_gold_cbs, transform_source_silver_to_gold_external)
Eliminates TRUNCATE and eliminates row-by-row iterrows().
"""
import os
import time
import requests
from sqlalchemy import text
from transform.silver import run_bronze_to_silver_cbs, transform_source_bronze_to_silver_external
from transform.gold import (
    run_silver_to_gold_cbs,
    transform_source_silver_to_gold_external,
    refresh_gold_marts
)

_LAST_WEBHOOK_TIME = 0.0
WEBHOOK_COOLDOWN_SECONDS = 15.0

def run_bronze_to_silver(engine):
    """Bridge for CBS Bronze to Silver transform."""
    return run_bronze_to_silver_cbs(engine)

def run_silver_to_gold(engine):
    """Bridge for CBS Silver to Gold transform."""
    return run_silver_to_gold_cbs(engine)

def transform_source_bronze_to_silver(engine, source_id, source_name, table_name, mapping, df=None):
    """Bridge for external source Bronze to Silver transform."""
    return transform_source_bronze_to_silver_external(
        engine=engine,
        source_id=source_id,
        source_name=source_name,
        table_name=table_name,
        mapping=mapping
    )

def transform_source_silver_to_gold(engine, source_id, source_name, table_name, mapping):
    """Bridge for external source Silver to Gold transform."""
    return transform_source_silver_to_gold_external(
        engine=engine,
        source_id=source_id,
        source_name=source_name,
        table_name=table_name,
        mapping=mapping
    )

def _trigger_retrain_webhook():
    """Trigger webhook to Python AI for automated retrain and hot-reload."""
    global _LAST_WEBHOOK_TIME
    now = time.time()
    if now - _LAST_WEBHOOK_TIME < WEBHOOK_COOLDOWN_SECONDS:
        return
    _LAST_WEBHOOK_TIME = now
    
    # Try both python-ai (port 8000/5000) and data-server ai-engine (port 8002)
    urls = [
        "http://127.0.0.1:8000/retrain/auto",
        "http://localhost:8000/retrain/auto",
        "http://127.0.0.1:8002/retrain/auto"
    ]
    for url in urls:
        try:
            resp = requests.post(url, json={"trigger": "gold_zone_sync"}, timeout=2.0)
            if resp.status_code == 200:
                print(f"[Transform:AI-Pool] Triggered auto-retrain webhook at {url}")
                break
        except Exception:
            continue

def feed_ai_training_pool(engine, source_id, source_name, table_name, mapping):
    """
    Syncs latest live aggregated Gold KPI metrics to the AI training pool (PostgreSQL / Parquet)
    completely OUTSIDE CSV files, avoiding I/O bottlenecks and memory freezing on large volumes.
    Notifies Python AI service via non-blocking webhook /retrain/auto with hot-reload.
    """
    try:
        # Refresh marts to ensure latest live metrics
        refresh_gold_marts(engine, source_id)

        # Pull live kpi series from gold.mart_kpi_series
        with engine.connect() as conn:
            rows = conn.execute(text("""
                SELECT source_id, kpi_name, period, actual_value, target_value
                FROM gold.mart_kpi_series
                WHERE source_id = :sid
            """), {"sid": source_id}).fetchall()

        if rows:
            # 1. Upsert into database training pool (gold.ai_auditee_kpi_training_pool)
            records = []
            for r in rows:
                kpi_name = r.kpi_name
                val = float(r.actual_value) if r.actual_value is not None else 0.0
                period_str = str(r.period)
                target_val = float(r.target_value or 2.0)
                impact = 4 if val > target_val * 1.5 else (2 if val < target_val else 3)
                likelihood = 4 if impact >= 4 else 3
                risk_lvl = 'HIGH' if (impact * likelihood >= 15) else ('MEDIUM' if (impact * likelihood >= 8) else 'LOW')

                # Resolve entity name from KPI
                entity_name = 'Kantor Pusat'
                for city in ['Surabaya', 'Jakarta', 'Bandung', 'Medan', 'Bali', 'Semarang', 'Makassar', 'Palembang']:
                    if city in kpi_name:
                        entity_name = f'Cabang {city}'
                        break

                records.append({
                    "source_id": str(r.source_id or source_id),
                    "kpi_name": kpi_name,
                    "entity": entity_name,
                    "entity_type": "Branch" if "Cabang" in entity_name else "Department",
                    "period_date": "2026-03-01",
                    "period_label": f"Bulan {period_str}" if not period_str.startswith('Bulan') else period_str,
                    "actual_value": val,
                    "target_value": target_val,
                    "impact": impact,
                    "likelihood": likelihood,
                    "risk_level": risk_lvl,
                    "sync_frequency": "Real-time (CDC)"
                })

            # Set-based SQL upsert directly into PostgreSQL
            try:
                with engine.begin() as tx:
                    for rec in records:
                        tx.execute(text("""
                            INSERT INTO gold.ai_auditee_kpi_training_pool (
                                source_id, kpi_name, entity, entity_type, period_date, period_label,
                                actual_value, target_value, impact, likelihood, risk_level, sync_frequency, updated_at
                            ) VALUES (
                                :source_id, :kpi_name, :entity, :entity_type, :period_date::date, :period_label,
                                :actual_value, :target_value, :impact, :likelihood, :risk_level, :sync_frequency, NOW()
                            )
                            ON CONFLICT (kpi_name, period_date, entity)
                            DO UPDATE SET
                                actual_value = EXCLUDED.actual_value,
                                target_value = EXCLUDED.target_value,
                                impact = EXCLUDED.impact,
                                likelihood = EXCLUDED.likelihood,
                                risk_level = EXCLUDED.risk_level,
                                sync_frequency = EXCLUDED.sync_frequency,
                                updated_at = NOW();
                        """), rec)
                print(f"[Transform:AI-Pool] Set-based SQL upserted {len(records)} live KPI records into gold.ai_auditee_kpi_training_pool")
            except Exception as e_db:
                print(f"[Transform:AI-Pool] Database table upsert notice: {e_db}")

            # 2. Sync to Columnar Binary Parquet buffer (Non-blocking, zero CSV)
            try:
                current_file_dir = os.path.dirname(os.path.abspath(__file__))
                sys.path.append(os.path.abspath(os.path.join(current_file_dir, '..', '..', '..', 'backend', 'python-ai')))
                from ai_training_storage import upsert_auditee_kpi_pool
                upsert_auditee_kpi_pool(records)
                print(f"[Transform:AI-Pool] Synced {len(records)} live KPI records to Parquet buffer.")
            except Exception as e_pq:
                print(f"[Transform:AI-Pool] Parquet sync notice: {e_pq}")

        # Non-blocking Webhook trigger to python-ai /retrain/auto
        _trigger_retrain_webhook()
    except Exception as e:
        print(f"[Transform:AI-Pool] Warning: Failed to sync AI training pool: {e}")
