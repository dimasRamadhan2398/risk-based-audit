"""
AuditSphere Data Hub API — Main Application
FastAPI REST gateway for Data Lake ingest, serve, sync, and pipeline management.
"""
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, Depends, HTTPException, Security
from fastapi.security import APIKeyHeader
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy import create_engine, text
from sqlalchemy.orm import sessionmaker

# ─── Database ────────────────────────────────────────────────────────────────
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/auditsphere_datalake")
CBS_DATABASE_URL = os.getenv("CBS_DATABASE_URL", "postgresql://postgres:postgres@localhost:5434/cbs_simulator")
API_KEY = os.getenv("API_KEY", "dev-api-key")

engine = create_engine(DATABASE_URL, pool_size=5, max_overflow=10, pool_pre_ping=True)
cbs_engine = create_engine(CBS_DATABASE_URL, pool_size=3, max_overflow=5, pool_pre_ping=True)
SessionLocal = sessionmaker(bind=engine)
CBSSessionLocal = sessionmaker(bind=cbs_engine)


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def get_cbs_db():
    db = CBSSessionLocal()
    try:
        yield db
    finally:
        db.close()


# ─── API Key Security ────────────────────────────────────────────────────────
api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)


async def verify_api_key(api_key: str = Security(api_key_header)):
    """Verify API key for inter-service authentication."""
    if api_key is None or api_key != API_KEY:
        raise HTTPException(status_code=401, detail="Invalid or missing API key")
    return api_key


# ─── Lifespan ────────────────────────────────────────────────────────────────
@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup: verify database connections
    try:
        with engine.connect() as conn:
            conn.execute(text("SELECT 1"))
        print("[DataHub] ✓ Data Lake DB connected")
    except Exception as e:
        print(f"[DataHub] ✗ Data Lake DB connection failed: {e}")

    try:
        with cbs_engine.connect() as conn:
            conn.execute(text("SELECT 1"))
        print("[DataHub] ✓ CBS Simulator DB connected")
    except Exception as e:
        print(f"[DataHub] ✗ CBS Simulator DB connection failed: {e}")

    yield
    # Shutdown
    engine.dispose()
    cbs_engine.dispose()


# ─── App ─────────────────────────────────────────────────────────────────────
app = FastAPI(
    title="AuditSphere Data Hub API",
    description="REST gateway for Data Lake ingest, serve, sync, and pipeline management",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ─── Import Routers ─────────────────────────────────────────────────────────
from routers import ingest, serve, sync, pipeline, caatt

app.include_router(ingest.router, prefix="/api/v1/ingest", tags=["Ingest"], dependencies=[Depends(verify_api_key)])
app.include_router(serve.router, prefix="/api/v1/gold", tags=["Serve Gold Zone"], dependencies=[Depends(verify_api_key)])
app.include_router(sync.router, prefix="/api/v1/sync", tags=["Master Data Sync"], dependencies=[Depends(verify_api_key)])
app.include_router(pipeline.router, prefix="/api/v1/pipeline", tags=["ETL Pipeline"], dependencies=[Depends(verify_api_key)])
app.include_router(caatt.router, prefix="/api/v1/caatt", tags=["CAATT Analytics"], dependencies=[Depends(verify_api_key)])


# ─── Health (no auth) ───────────────────────────────────────────────────────
@app.get("/health")
def health():
    """Health check — no API key required."""
    db_ok = False
    cbs_ok = False
    try:
        with engine.connect() as conn:
            conn.execute(text("SELECT 1"))
            db_ok = True
    except Exception:
        pass
    try:
        with cbs_engine.connect() as conn:
            conn.execute(text("SELECT 1"))
            cbs_ok = True
    except Exception:
        pass

    schemas = []
    if db_ok:
        try:
            with engine.connect() as conn:
                result = conn.execute(text("SELECT schema_name FROM information_schema.schemata WHERE schema_name IN ('bronze','silver','gold')"))
                schemas = [r[0] for r in result]
        except Exception:
            pass

    return {
        "status": "ok" if db_ok else "degraded",
        "service": "auditsphere-data-hub-api",
        "database": {"datalake": db_ok, "cbs_simulator": cbs_ok},
        "schemas": schemas,
    }
