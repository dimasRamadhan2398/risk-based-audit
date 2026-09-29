"""
Data Hub API — Ingest Router
Endpoints for ingesting data from external sources into Bronze zone.
"""
from typing import Optional, List
from fastapi import APIRouter, UploadFile, File, Form, HTTPException
from pydantic import BaseModel
from sqlalchemy import text
import pandas as pd
import io
import json

from main import engine

router = APIRouter()


class DatabaseIngestRequest(BaseModel):
    """Request to ingest from an external database via connection string."""
    connection_string: str
    source_name: str
    tables: List[str]
    query: Optional[str] = None


class BatchRecord(BaseModel):
    """A single record for batch ingest."""
    data: dict


class BatchIngestRequest(BaseModel):
    """Request to ingest a batch of records."""
    source_name: str
    target_table: str
    records: List[dict]


@router.post("/batch")
def ingest_batch(req: BatchIngestRequest):
    """
    Ingest a batch of JSON records into a Bronze zone table.
    Creates the table dynamically if it doesn't exist.
    """
    if not req.records:
        raise HTTPException(status_code=400, detail="No records provided")

    table_name = f"bronze.{req.target_table}"
    df = pd.DataFrame(req.records)

    # Add metadata columns
    df["_source"] = req.source_name
    df["_loaded_at"] = pd.Timestamp.now()

    try:
        df.to_sql(
            req.target_table,
            engine,
            schema="bronze",
            if_exists="append",
            index=False,
            method="multi",
        )

        # Log data freshness
        with engine.connect() as conn:
            conn.execute(
                text("""
                    INSERT INTO bronze.data_freshness_log (source_name, table_name, records_loaded, is_new)
                    VALUES (:source, :table, :count, TRUE)
                """),
                {"source": req.source_name, "table": table_name, "count": len(req.records)},
            )
            conn.commit()

        return {
            "status": "success",
            "message": f"Ingested {len(req.records)} records into {table_name}",
            "source": req.source_name,
            "records_loaded": len(req.records),
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Ingest failed: {str(e)}")


@router.post("/csv")
async def ingest_csv(
    file: UploadFile = File(...),
    source_name: str = Form("csv_upload"),
    target_table: str = Form(...),
):
    """
    Upload a CSV file and ingest into Bronze zone.
    """
    if not file.filename.endswith(".csv"):
        raise HTTPException(status_code=400, detail="Only CSV files are accepted")

    try:
        contents = await file.read()
        df = pd.read_csv(io.BytesIO(contents))
        df["_source"] = source_name
        df["_loaded_at"] = pd.Timestamp.now()

        df.to_sql(
            target_table,
            engine,
            schema="bronze",
            if_exists="append",
            index=False,
            method="multi",
        )

        # Log freshness
        with engine.connect() as conn:
            conn.execute(
                text("""
                    INSERT INTO bronze.data_freshness_log (source_name, table_name, records_loaded, is_new)
                    VALUES (:source, :table, :count, TRUE)
                """),
                {"source": source_name, "table": f"bronze.{target_table}", "count": len(df)},
            )
            conn.commit()

        return {
            "status": "success",
            "filename": file.filename,
            "records_loaded": len(df),
            "target": f"bronze.{target_table}",
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"CSV ingest failed: {str(e)}")


@router.post("/database")
def ingest_from_database(req: DatabaseIngestRequest):
    """
    Connect to an external database and pull data into Bronze zone.
    Uses connection string from the request (or from cached DataSourceConnection config).
    """
    try:
        from sqlalchemy import create_engine as ce

        ext_engine = ce(req.connection_string, pool_pre_ping=True)
        total = 0

        for table in req.tables:
            query = req.query or f"SELECT * FROM {table}"
            df = pd.read_sql(query, ext_engine)
            df["_source"] = req.source_name
            df["_loaded_at"] = pd.Timestamp.now()

            target = table.replace(".", "_")
            df.to_sql(target, engine, schema="bronze", if_exists="append", index=False, method="multi")
            total += len(df)

        # Log
        with engine.connect() as conn:
            conn.execute(
                text("""
                    INSERT INTO bronze.data_freshness_log (source_name, table_name, records_loaded, is_new)
                    VALUES (:source, :table, :count, TRUE)
                """),
                {"source": req.source_name, "table": ",".join(req.tables), "count": total},
            )
            conn.commit()

        ext_engine.dispose()

        return {
            "status": "success",
            "source": req.source_name,
            "tables_ingested": req.tables,
            "total_records": total,
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Database ingest failed: {str(e)}")
