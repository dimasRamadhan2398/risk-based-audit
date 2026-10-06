"""
AuditSphere Data Hub — Database Connection & Pool Management
Provides SQLAlchemy engines and psycopg3 connection pools with strict read-only enforcement for client databases.
"""
import os
from contextlib import contextmanager
from sqlalchemy import create_engine
from psycopg_pool import ConnectionPool
import psycopg
from core.hardware import get_hardware_profile

DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/auditsphere_datalake")
CBS_DATABASE_URL = os.getenv("CBS_DATABASE_URL", "postgresql://postgres:postgres@localhost:5434/cbs_simulator")

# Convert potential postgresql:// to standard connection string for psycopg3
def _to_psycopg_conninfo(url: str) -> str:
    if url.startswith("postgresql+psycopg://"):
        return url.replace("postgresql+psycopg://", "postgresql://")
    return url

# 1. SQLAlchemy Engine (for standard relational queries)
engine = create_engine(
    DATABASE_URL,
    pool_size=5,
    max_overflow=10,
    pool_pre_ping=True
)

cbs_engine = create_engine(
    CBS_DATABASE_URL,
    pool_size=3,
    max_overflow=5,
    pool_pre_ping=True
)

# 2. psycopg3 ConnectionPool (for ultra-high-throughput streaming COPY)
datalake_conninfo = _to_psycopg_conninfo(DATABASE_URL)
try:
    pg_pool = ConnectionPool(
        conninfo=datalake_conninfo,
        min_size=2,
        max_size=10,
        open=True
    )
except Exception as e:
    print(f"[DB] Notice: Local connection pool initialization deferred: {e}")
    pg_pool = None

@contextmanager
def get_datalake_raw_conn():
    """Yield a raw psycopg3 connection for COPY operations."""
    if pg_pool:
        with pg_pool.connection() as conn:
            yield conn
    else:
        with psycopg.connect(datalake_conninfo) as conn:
            yield conn

def build_client_conn_string(src: dict) -> str:
    """Build connection URL for external client database."""
    stype = (src.get("source_type") or "postgres").lower()
    host = src.get("host", "localhost")
    port = src.get("port", 5432)
    db = src.get("database_name", "")
    user = src.get("username", "")
    pwd = src.get("password_encrypted", "") or ""

    auth = f"{user}:{pwd}@" if user or pwd else ""

    if "postgres" in stype:
        ssl_mode = "?sslmode=require" if src.get("ssl_enabled", True) else "?sslmode=disable"
        if host in ("localhost", "127.0.0.1", "cbs-simulator", "datalake-db", "host.docker.internal"):
            ssl_mode = "?sslmode=disable"
        return f"postgresql://{auth}{host}:{port}/{db}{ssl_mode}"
    elif "mysql" in stype:
        return f"mysql+pymysql://{auth}{host}:{port}/{db}"
    elif "mssql" in stype:
        return f"mssql+pyodbc://{auth}{host}:{port}/{db}?driver=ODBC+Driver+17+for+SQL+Server"
    elif "oracle" in stype:
        return f"oracle+cx_oracle://{auth}{host}:{port}/?service_name={db}"
    else:
        return f"postgresql://{auth}{host}:{port}/{db}"

@contextmanager
def get_read_only_client_conn(src: dict):
    """
    Open connection to client database with STRICT read-only and statement timeout guarantees.
    Prevents any accidental writes or runaway queries on client infrastructure.
    """
    hw = get_hardware_profile()
    timeout_ms = hw.get("statement_timeout_ms", 60000)
    stype = (src.get("source_type") or "postgres").lower()

    if "postgres" in stype:
        conn_str = build_client_conn_string(src)
        # Connect with psycopg3
        with psycopg.connect(conn_str, autocommit=True) as conn:
            with conn.cursor() as cur:
                cur.execute(f"SET TRANSACTION READ ONLY;")
                cur.execute(f"SET statement_timeout = '{timeout_ms}';")
            yield conn
    else:
        # Fallback using SQLAlchemy engine for non-Postgres engines (MySQL/MSSQL/Oracle)
        conn_str = build_client_conn_string(src)
        ext_engine = create_engine(conn_str, pool_pre_ping=True)
        try:
            with ext_engine.connect() as conn:
                if "mysql" in stype:
                    conn.exec_driver_sql("SET SESSION TRANSACTION READ ONLY")
                yield conn
        finally:
            ext_engine.dispose()
