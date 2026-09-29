"""
Data Hub API — ETL Service
Pulls data from CBS Simulator into Bronze zone.
"""
import pandas as pd
from sqlalchemy import text


CBS_TABLES = [
    ("cbs.transactions", "cb_transactions"),
    ("cbs.accounts", "cb_accounts"),
    ("cbs.customers", "cb_customers"),
    ("cbs.loans", "cb_loans"),
    ("cbs.gl_entries", "cb_gl_entries"),
    ("cbs.products", "cb_products"),
    ("cbs.product_policies", "cb_product_policies"),
    ("cbs.documents", "cb_documents"),
    ("cbs.branches", "cb_branches"),
    ("cbs.failed_logins", "cb_failed_logins"),
]


def run_cbs_etl_pipeline(cbs_engine, datalake_engine) -> dict:
    """
    Extract data from CBS Simulator and load into Bronze zone.
    Full reload strategy (truncate + insert) for simplicity.
    """
    total_records = 0
    table_counts = {}

    for source_table, target_table in CBS_TABLES:
        try:
            # Extract from CBS
            df = pd.read_sql(f"SELECT * FROM {source_table}", cbs_engine)

            if df.empty:
                table_counts[target_table] = 0
                continue

            # Add metadata columns
            df["_source"] = "cbs_simulator"
            df["_loaded_at"] = pd.Timestamp.now()

            # Drop the _load_id column if exists in target (it's auto-generated)
            # Truncate existing data in bronze target
            with datalake_engine.connect() as conn:
                conn.execute(text(f"TRUNCATE TABLE bronze.{target_table}"))
                conn.commit()

            # Load into Bronze
            df.to_sql(
                target_table,
                datalake_engine,
                schema="bronze",
                if_exists="append",
                index=False,
                method="multi",
                chunksize=1000,
            )

            table_counts[target_table] = len(df)
            total_records += len(df)

            print(f"  [ETL] {source_table} → bronze.{target_table}: {len(df)} records")

        except Exception as e:
            print(f"  [ETL] ERROR {source_table}: {e}")
            table_counts[target_table] = f"ERROR: {str(e)}"

    # Log freshness
    try:
        with datalake_engine.connect() as conn:
            conn.execute(text("""
                INSERT INTO bronze.data_freshness_log (source_name, table_name, records_loaded, is_new)
                VALUES ('cbs_simulator', 'all_cbs_tables', :count, TRUE)
            """), {"count": total_records})
            conn.commit()
    except Exception as e:
        print(f"  [ETL] Freshness log failed: {e}")

    return {
        "status": "success",
        "total_records": total_records,
        "tables": table_counts,
    }
