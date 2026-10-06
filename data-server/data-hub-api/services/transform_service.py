"""
Data Hub API — Transform Service (Bridge)
Redirects to high-performance set-based transform modules:
- transform.silver (run_bronze_to_silver_cbs, transform_source_bronze_to_silver_external)
- transform.gold (run_silver_to_gold_cbs, transform_source_silver_to_gold_external)
Eliminates TRUNCATE and eliminates row-by-row iterrows().
"""
from transform.silver import run_bronze_to_silver_cbs, transform_source_bronze_to_silver_external
from transform.gold import (
    run_silver_to_gold_cbs,
    transform_source_silver_to_gold_external,
    refresh_gold_marts
)

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

def feed_ai_training_pool(engine, source_id, source_name, table_name, mapping):
    """AI training pool is populated automatically in transform_source_silver_to_gold_external."""
    pass
