"""Register ack echoes the stored ssl_mode; /health advertises the capability;
migration runner tracking/backfill/locking logic (DB mocked)."""
import importlib
import sys
import types
from unittest.mock import MagicMock

import pytest

pytest.importorskip("fastapi")


@pytest.fixture
def real_main(monkeypatch):
    """Import the real `main` app, undoing any `main` stub other tests installed."""
    for name in [n for n in sys.modules if n == "main" or n.startswith("routers")]:
        monkeypatch.delitem(sys.modules, name, raising=False)
    mod = importlib.import_module("main")
    yield mod
    for name in [n for n in sys.modules if n == "main" or n.startswith("routers")]:
        sys.modules.pop(name, None)


def _fake_engine(returned_ssl_mode):
    row = types.SimpleNamespace(ssl_mode=returned_ssl_mode, ssl_enabled=returned_ssl_mode != "disable",
                                has_root_cert=True)
    conn = MagicMock()
    conn.execute.return_value.one.return_value = row
    eng = MagicMock()
    eng.connect.return_value.__enter__.return_value = conn
    return eng, conn


def _payload(**kw):
    base = dict(source_id="s1", name="CBS", host="db.example", database_name="cbs")
    base.update(kw)
    return base


def test_register_echoes_stored_ssl_mode(real_main, monkeypatch):
    from routers import sources
    eng, conn = _fake_engine("verify-full")
    monkeypatch.setattr(sources, "engine", eng)
    res = sources.register_source(sources.SourceRegistration(**_payload(ssl_mode="verify-full")))
    assert res["status"] == "success"          # existing fields preserved
    assert res["ssl_mode"] == "verify-full"
    assert res["ssl_enabled"] is True and res["has_ssl_root_cert"] is True
    assert "ssl_root_cert" not in res           # never echo the CA / secrets
    params = conn.execute.call_args[0][1]
    assert params["ssl_mode"] == "verify-full"
    conn.commit.assert_called_once()


def test_register_ack_reflects_db_value_not_request(real_main, monkeypatch):
    """The ack must come from what the DB stored (RETURNING), so a mismatch is detectable."""
    from routers import sources
    eng, _ = _fake_engine("require")
    monkeypatch.setattr(sources, "engine", eng)
    res = sources.register_source(sources.SourceRegistration(**_payload(ssl_mode="disable")))
    assert res["ssl_mode"] == "require"


def test_register_legacy_flag_acks_effective_mode(real_main, monkeypatch):
    from routers import sources
    eng, conn = _fake_engine("disable")
    monkeypatch.setattr(sources, "engine", eng)
    res = sources.register_source(sources.SourceRegistration(**_payload(ssl_enabled=False)))
    assert conn.execute.call_args[0][1]["ssl_mode"] == "disable"
    assert res["ssl_mode"] == "disable"


def test_health_advertises_ssl_mode_capability(real_main, monkeypatch):
    eng = MagicMock()
    monkeypatch.setattr(real_main, "engine", eng)
    monkeypatch.setattr(real_main, "cbs_engine", eng)
    body = real_main.health()
    assert "ssl_mode" in body["features"]
    assert body["ssl_modes"] == ["disable", "require", "verify-ca", "verify-full"]
    assert {"status", "service", "database", "schemas"} <= set(body)  # additive only


# ─── Migration runner ────────────────────────────────────────────────────────
from core import migrations  # noqa: E402


class FakeConn:
    def __init__(self, log, sentinels_present=(), applied=()):
        self.log, self.present, self.applied = log, set(sentinels_present), set(applied)

    def __enter__(self): return self
    def __exit__(self, *a): return False
    def commit(self): pass
    def rollback(self): pass
    def close(self): self.log.append("close")

    def execute(self, stmt, params=None):
        sql = str(stmt)
        res = MagicMock()
        if "pg_advisory_lock" in sql:
            self.log.append("lock")
        elif "pg_advisory_unlock" in sql:
            self.log.append("unlock")
        elif "SELECT version FROM ops.schema_migrations" in sql:
            res.fetchall.return_value = [(v,) for v in self.applied]
        elif sql.strip().startswith("SELECT 1 FROM pg_"):
            hit = any(s in sql for s in self.present)
            res.first.return_value = (1,) if hit else None
        elif "INSERT INTO ops.schema_migrations" in sql:
            self.log.append(("record", params["ver"], params.get("chk") is None))
        elif "CREATE SCHEMA" in sql:
            self.log.append("ddl-init")
        elif "-- " in sql and "CREATE" not in sql:   # a migration script body
            self.log.append(("run", sql))
        return res


def _engine(conn):
    eng = MagicMock()
    eng.connect.side_effect = lambda: conn
    return eng


def test_runner_takes_lock_before_any_ddl_and_releases(tmp_path):
    (tmp_path / "001_a.sql").write_text("SELECT 1;")
    log = []
    migrations.run_migrations(_engine(FakeConn(log)), str(tmp_path))
    assert log.index("lock") < log.index("ddl-init")
    assert log.index("unlock") > log.index(("record", "001_a.sql", False))


def test_runner_releases_lock_when_migration_fails(tmp_path):
    (tmp_path / "001_a.sql").write_text("SELECT 1;")
    log = []

    class Boom(FakeConn):
        def execute(self, stmt, params=None):
            if str(stmt).strip() == "SELECT 1;":
                raise RuntimeError("boom")
            return super().execute(stmt, params)

    with pytest.raises(RuntimeError):
        migrations.run_migrations(_engine(Boom(log)), str(tmp_path))
    assert "unlock" in log


def test_legacy_db_is_backfilled_not_rerun(tmp_path):
    for v in migrations.LEGACY_SENTINELS:
        (tmp_path / v).write_text("SELECT 1; -- would be destructive")
    (tmp_path / "009_new.sql").write_text("SELECT 1; -- new")
    log = []
    # every legacy sentinel present => schema was migrated the old way
    conn = FakeConn(log, sentinels_present=["pg_"])
    migrations.run_migrations(_engine(conn), str(tmp_path))
    backfilled = [e[1] for e in log if isinstance(e, tuple) and e[0] == "record" and e[2]]
    assert sorted(backfilled) == sorted(migrations.LEGACY_SENTINELS)
    ran = [e for e in log if isinstance(e, tuple) and e[0] == "run"]
    assert len(ran) == 1                                    # only 009 executed


def test_fresh_db_runs_everything(tmp_path):
    for v in ["001_a.sql", "002_b.sql"]:
        (tmp_path / v).write_text("SELECT 1; -- x")
    log = []
    migrations.run_migrations(_engine(FakeConn(log, sentinels_present=[])), str(tmp_path))
    assert len([e for e in log if isinstance(e, tuple) and e[0] == "run"]) == 2
    assert not [e for e in log if isinstance(e, tuple) and e[0] == "record" and e[2]]  # nothing backfilled


def test_sentinel_names_match_migration_files():
    import glob, os
    mdir = os.path.join(os.path.dirname(__file__), "..", "..", "migrations")
    files = {os.path.basename(f) for f in glob.glob(os.path.join(mdir, "*.sql"))}
    assert set(migrations.LEGACY_SENTINELS) <= files
    assert "009_registered_sources_ssl_mode.sql" not in migrations.LEGACY_SENTINELS


def test_004_creates_unique_index_before_enabling_compression():
    import os
    sql = open(os.path.join(os.path.dirname(__file__), "..", "..", "migrations", "004_gold_canonical.sql")).read()
    assert sql.index("uq_gold_fact_trx_ref") < sql.index("timescaledb.compress,")
