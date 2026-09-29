"""
AuditSphere CAATT Engine — Database Connection Helper
Shared by all CAATT notebooks for consistent database access.
"""
import os
from sqlalchemy import create_engine


def get_datalake_engine():
    """Get SQLAlchemy engine for the Data Lake (Bronze/Silver/Gold zones)."""
    host = os.getenv("DATALAKE_HOST", "postgres-datalake")
    port = os.getenv("DATALAKE_PORT", "5432")
    db = os.getenv("DATALAKE_DB", "auditsphere_datalake")
    user = os.getenv("DATALAKE_USER", "etl_user")
    password = os.getenv("DATALAKE_PASSWORD", "password")
    url = f"postgresql://{user}:{password}@{host}:{port}/{db}"
    return create_engine(url, pool_pre_ping=True)


def get_cbs_engine():
    """Get SQLAlchemy engine for the CBS Simulator."""
    host = os.getenv("CBS_HOST", "postgres-cbs-sim")
    port = os.getenv("CBS_PORT", "5432")
    db = os.getenv("CBS_DB", "cbs_simulator")
    user = os.getenv("CBS_USER", "postgres")
    password = os.getenv("CBS_PASSWORD", "password")
    url = f"postgresql://{user}:{password}@{host}:{port}/{db}"
    return create_engine(url, pool_pre_ping=True)


def get_datahub_url():
    """Get Data Hub API base URL."""
    return os.getenv("DATA_HUB_URL", "http://data-hub-api:8100")


def get_ai_engine_url():
    """Get AI Engine base URL."""
    return os.getenv("AI_ENGINE_URL", "http://ai-engine:8000")
