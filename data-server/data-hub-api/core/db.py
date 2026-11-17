"""
AuditSphere Data Hub — Database Connection & Pool Management
Provides SQLAlchemy engines and psycopg3 connection pools with strict read-only enforcement for client databases.
"""
import os
import ssl
import tempfile
from urllib.parse import quote
from contextlib import contextmanager
from sqlalchemy import create_engine
from psycopg_pool import ConnectionPool
import psycopg
from psycopg.conninfo import make_conninfo
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

# ─── Client DB SSL/TLS (libpq semantics, mirrors master-service pkg/datasource/ssl.go) ───
SSL_MODES = ("disable", "require", "verify-ca", "verify-full")
SSL_VERIFY_MODES = ("verify-ca", "verify-full")
MAX_ROOT_CERT_BYTES = 64 * 1024


class SSLConfigError(ValueError):
    """Invalid or unusable SSL configuration for a client data source."""


def resolve_ssl_mode(ssl_mode, ssl_enabled=None) -> str:
    """Effective sslmode. A non-empty ssl_mode is authoritative; otherwise the legacy boolean
    decides (True/None -> require, False -> disable). Unknown modes raise SSLConfigError."""
    m = (ssl_mode or "").strip().lower() if isinstance(ssl_mode, str) else ""
    if not m:
        return "disable" if ssl_enabled is False else "require"
    if m not in SSL_MODES:
        raise SSLConfigError(f"unknown ssl_mode {ssl_mode!r}: allowed values are {', '.join(SSL_MODES)}")
    return m


def validate_ssl_config(mode: str, root_cert_pem) -> None:
    """Check mode + optional CA PEM together (same rules as the Go master-service)."""
    if mode not in SSL_MODES:
        raise SSLConfigError(f"unknown ssl_mode {mode!r}")
    pem = (root_cert_pem or "").strip()
    if not pem:
        return
    if mode not in SSL_VERIFY_MODES:
        raise SSLConfigError(f"ssl_root_cert only applies to verify-ca and verify-full, not {mode}")
    if len(pem.encode("utf-8")) > MAX_ROOT_CERT_BYTES:
        raise SSLConfigError(f"ssl_root_cert exceeds {MAX_ROOT_CERT_BYTES} bytes")
    try:
        ssl.create_default_context(cadata=pem)
    except (ssl.SSLError, ValueError) as e:
        raise SSLConfigError("ssl_root_cert is not a valid PEM CA certificate") from e


def _system_root_ref(mode: str) -> str:
    """sslrootcert value for 'verify with the system CA pool' (empty custom CA)."""
    # libpq >= 16 understands sslrootcert=system, but only together with verify-full.
    if mode == "verify-full" and psycopg.pq.version() >= 160000:
        return "system"
    cands = [ssl.get_default_verify_paths().cafile, ssl.get_default_verify_paths().openssl_cafile]
    try:
        import certifi
        cands.append(certifi.where())
    except ImportError:
        pass
    for c in cands:
        if c and os.path.isfile(c):
            return c
    raise SSLConfigError(f"ssl_mode={mode} without ssl_root_cert needs a system CA bundle, none found")


def _client_ssl_params(src: dict, root_cert_path=None) -> dict:
    mode = resolve_ssl_mode(src.get("ssl_mode"), src.get("ssl_enabled"))
    pem = (src.get("ssl_root_cert") or "").strip()
    validate_ssl_config(mode, pem)
    params = {"sslmode": mode}
    if mode in SSL_VERIFY_MODES:
        if pem:
            if not root_cert_path:
                raise SSLConfigError("ssl_root_cert provided but no CA file path was supplied")
            params["sslrootcert"] = root_cert_path
        else:
            params["sslrootcert"] = _system_root_ref(mode)
    return params


def _postgres_conninfo(src: dict, root_cert_path=None) -> str:
    host = src.get("host") or ""
    if not host:
        raise ValueError("data source host is empty")
    kw = dict(
        host=host,
        port=int(src.get("port") or 5432),
        dbname=src.get("database_name") or "",
        **_client_ssl_params(src, root_cert_path),
    )
    if src.get("username"):
        kw["user"] = src["username"]
    if src.get("password_encrypted"):
        kw["password"] = src["password_encrypted"]
    # make_conninfo quotes/escapes every value (spaces, quotes, backslashes in passwords).
    return make_conninfo(**kw)


@contextmanager
def client_postgres_conninfo(src: dict):
    """Yield a libpq conninfo for a client PostgreSQL source honoring its SSL settings.
    A custom CA PEM is written to a 0600 temp file that is removed on exit (libpq reads it at
    connect time, so connect inside the with-block)."""
    pem = (src.get("ssl_root_cert") or "").strip()
    path = None
    try:
        if pem:
            # validate first so we never write junk to disk
            validate_ssl_config(resolve_ssl_mode(src.get("ssl_mode"), src.get("ssl_enabled")), pem)
            fd, path = tempfile.mkstemp(prefix="dh-ca-", suffix=".pem")  # mkstemp creates 0600
            with os.fdopen(fd, "w", encoding="utf-8") as f:
                f.write(pem + "\n")
        yield _postgres_conninfo(src, path)
    finally:
        if path:
            try:
                os.unlink(path)
            except FileNotFoundError:
                pass


def build_client_conn_string(src: dict, root_cert_path=None) -> str:
    """Build connection string for an external client database.
    Postgres: libpq conninfo (use client_postgres_conninfo() when a custom CA PEM is set).
    No host-name based sslmode override: the source's own setting always wins."""
    stype = (src.get("source_type") or "postgres").lower()
    host = src.get("host", "localhost")
    port = src.get("port", 5432)
    db = src.get("database_name", "")
    user = src.get("username") or ""
    pwd = src.get("password_encrypted", "") or ""

    if "mysql" in stype or "mssql" in stype or "oracle" in stype:
        auth = f"{quote(user, safe='')}:{quote(pwd, safe='')}@" if user or pwd else ""
        if "mysql" in stype:
            return f"mysql+pymysql://{auth}{host}:{port}/{db}"
        if "mssql" in stype:
            return f"mssql+pyodbc://{auth}{host}:{port}/{db}?driver=ODBC+Driver+17+for+SQL+Server"
        return f"oracle+cx_oracle://{auth}{host}:{port}/?service_name={db}"
    return _postgres_conninfo(src, root_cert_path)

def _docker_fallback_sources(src: dict) -> list:
    """When running inside a Docker container, resolve loopback hosts (localhost/127.0.0.1)
    to host.docker.internal or the local Docker PostgreSQL service if local loopback is refused."""
    if not os.path.exists("/.dockerenv"):
        return []
    host = (src.get("host") or "").strip().lower()
    port = int(src.get("port") or 5432)
    fallbacks = []
    if host in ("localhost", "127.0.0.1", "::1"):
        fallbacks.append({**src, "host": "host.docker.internal", "port": port})
        if port == 5432:
            fallbacks.append({**src, "host": "postgres-cbs-sim", "port": 5432})
            fallbacks.append({**src, "host": "host.docker.internal", "port": 5434})
    elif host == "host.docker.internal" and port == 5432:
        fallbacks.append({**src, "host": "postgres-cbs-sim", "port": 5432})
        fallbacks.append({**src, "host": "host.docker.internal", "port": 5434})
    return fallbacks


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
        # Connect with psycopg3 (fail-fast within 3s if external host is down)
        conn = None
        last_err = None
        candidates = [src] + _docker_fallback_sources(src)
        for cand in candidates:
            try:
                with client_postgres_conninfo(cand) as conn_str:
                    conn = psycopg.connect(conn_str, autocommit=True, connect_timeout=3)
                break
            except psycopg.OperationalError as err:
                if last_err is None:
                    last_err = err
        if conn is None:
            raise last_err
        # CA temp file is already removed here; the TLS session is established.
        with conn:
            with conn.cursor() as cur:
                cur.execute("SET search_path TO public, cbs;")
                cur.execute("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY;")
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
