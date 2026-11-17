"""
Unit tests for CAATT Benford forensic telemetry, canonical projection field detection,
Data Hub empty-state safety (no dummy fallback), and Sources SSL echo contract.
"""
import importlib
import os
import sys
import types
import unittest
from contextlib import contextmanager

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))


class _FakeResult:
    def __init__(self, rows=None, rowcount=0):
        self._rows = rows or []
        self.rowcount = rowcount

    def fetchone(self):
        return self._rows[0] if self._rows else None

    def fetchall(self):
        return list(self._rows)


class _FakeRow:
    def __init__(self, mapping: dict):
        self._mapping = dict(mapping)
        self._tuple = tuple(mapping.values())

    def __getitem__(self, idx):
        return self._tuple[idx]

    def __len__(self):
        return len(self._tuple)


class _FakeConnection:
    def __init__(self, responder=None):
        self.queries = []
        self._responder = responder or (lambda sql, params: _FakeResult())

    def execute(self, stmt, params=None):
        sql_str = str(stmt)
        self.queries.append((sql_str, params or {}))
        return self._responder(sql_str, params or {})

    def commit(self):
        pass


class _FakeEngine:
    def __init__(self, responder=None):
        self.conn = _FakeConnection(responder)

    @contextmanager
    def connect(self):
        yield self.conn

    @contextmanager
    def begin(self):
        yield self.conn


_CREATED_STUBS = {}


def _install_missing_stubs():
    try:
        import sqlalchemy  # noqa: F401
    except ImportError:
        stub = types.ModuleType("sqlalchemy")
        stub.text = lambda s: s
        stub.create_engine = lambda *args, **kwargs: _FakeEngine()
        sys.modules["sqlalchemy"] = stub
        _CREATED_STUBS["sqlalchemy"] = stub

    try:
        import psycopg_pool  # noqa: F401
    except ImportError:
        pool_stub = types.ModuleType("psycopg_pool")
        pool_stub.ConnectionPool = object
        sys.modules["psycopg_pool"] = pool_stub
        _CREATED_STUBS["psycopg_pool"] = pool_stub

    try:
        import psycopg  # noqa: F401
    except ImportError:
        psycopg_stub = types.ModuleType("psycopg")
        psycopg_stub.__path__ = []
        conninfo_stub = types.ModuleType("psycopg.conninfo")
        conninfo_stub.make_conninfo = lambda *args, **kwargs: ""
        psycopg_stub.conninfo = conninfo_stub
        sys.modules["psycopg"] = psycopg_stub
        sys.modules["psycopg.conninfo"] = conninfo_stub
        _CREATED_STUBS["psycopg"] = psycopg_stub
        _CREATED_STUBS["psycopg.conninfo"] = conninfo_stub


_install_missing_stubs()


class TestCAATTBenfordAndSources(unittest.TestCase):
    def test_detect_external_field_mapping_indonesian_transactions(self):
        from transform.gold import _detect_external_field_mapping

        sample_rows = [
            {
                "trx_ref": "TRX-2026-001",
                "tgl_trx": "2026-07-01 10:15:00",
                "nominal": "12500000",
                "nama_cabang": "Cabang Jakarta",
                "nama_nasabah": "PT Contoh Nyata",
                "jenis_transaksi": "Transfer",
                "status": "Authorized",
            }
        ]
        fmap = _detect_external_field_mapping(sample_rows, "trx_nasabah")
        self.assertEqual(fmap["ref"], "trx_ref")
        self.assertEqual(fmap["date"], "tgl_trx")
        self.assertEqual(fmap["amount"], "nominal")
        self.assertEqual(fmap["branch"], "nama_cabang")
        self.assertEqual(fmap["customer"], "nama_nasabah")
        self.assertEqual(fmap["category"], "jenis_transaksi")

    def test_pick_jsonb_key_case_and_fuzzy_match(self):
        from transform.gold import _pick_jsonb_key

        keys = ["TransactionID", "TrxAmountUSD", "Branch_Name"]
        lower_to_actual = {k.lower(): k for k in keys}
        sample_rows = [{"TransactionID": 1, "TrxAmountUSD": 55000, "Branch_Name": "Surabaya"}]
        self.assertEqual(
            _pick_jsonb_key(lower_to_actual, ("amount", "trxamount", "nominal"), sample_rows=sample_rows, require_numeric=True),
            "TrxAmountUSD",
        )
        self.assertEqual(_pick_jsonb_key(lower_to_actual, ("branch_name", "branch")), "Branch_Name")
        self.assertEqual(_pick_jsonb_key(lower_to_actual, ("collectability",)), "")

    def test_benford_constants_and_all_seven_routines_execute(self):
        from services.caatt_engine import (
            BENFORD_EXPECTED_PROBS,
            CHI_SQUARE_CRIT_DF8_95,
            CHI_SQUARE_CRIT_DF1_95,
            run_all_caatt_tests,
        )

        self.assertAlmostEqual(sum(BENFORD_EXPECTED_PROBS.values()), 1.0, places=4)
        self.assertEqual(CHI_SQUARE_CRIT_DF8_95, 15.507)
        self.assertEqual(CHI_SQUARE_CRIT_DF1_95, 3.841)

        engine = _FakeEngine(lambda sql, params: _FakeResult(rows=[], rowcount=3))
        res = run_all_caatt_tests(engine, source_id="ext-src-1")
        self.assertEqual(set(res.keys()), {
            "full_population",
            "duplicate_gap",
            "benford",
            "stratification",
            "reconciliation",
            "policy_violations",
            "data_quality",
        })
        sql_blob = "\n".join(q for q, _ in engine.conn.queries)
        self.assertIn("gold.caatt_benford_results", sql_blob)
        self.assertIn("gold.caatt_reconciliation_results", sql_blob)
        self.assertIn("gold.caatt_data_quality_metrics", sql_blob)

    def test_caatt_router_empty_and_real_benford_payload(self):
        fake_main = types.ModuleType("main")
        fake_main.engine = _FakeEngine()
        prev_main = sys.modules.get("main")
        sys.modules["main"] = fake_main
        try:
            import routers.caatt as caatt_router
            caatt_router = importlib.reload(caatt_router)

            # Empty DB must NEVER return simulated/dummy numbers
            empty_res = caatt_router.get_benford_results(source_id=None)
            self.assertEqual(empty_res["data"], [])
            self.assertEqual(empty_res["findings"], [])
            self.assertEqual(empty_res["summary"]["conclusion"], "NO_DATA")
            self.assertEqual(empty_res["summary"]["total_digits_analyzed"], 0)

            # Real rows + real findings pass-through
            def real_responder(sql, params):
                if "FROM bronze.registered_sources WHERE source_id" in sql:
                    return _FakeResult([_FakeRow({"name": "Bank Core Production"})])
                if "details IS NOT NULL AND details != '{}'::jsonb" in sql:
                    return _FakeResult([_FakeRow({"cnt": 9, "valid_cnt": 9})])
                if "FROM gold.caatt_benford_results" in sql:
                    return _FakeResult([_FakeRow({
                        "digit": 5,
                        "actual_count": 42,
                        "expected_count": 15.84,
                        "actual_pct": 21.0,
                        "expected_pct": 7.92,
                        "deviation_pct": 13.08,
                        "chi_square": 43.25,
                        "chi_square_share_pct": 64.5,
                        "is_significant": True,
                        "sample_size": 200,
                        "details": {
                            "count_diff": 26.16,
                            "dominant_branch": "Cabang Surabaya",
                            "dominant_branch_count": 19,
                            "dominant_category": "Transfer",
                            "dominant_category_count": 25,
                            "total_amount": 2100000000,
                            "avg_amount": 50000000,
                            "round_amount_count": 14,
                            "after_hours_count": 5,
                            "suspicious_count": 3,
                            "direction": "over-represented",
                            "finding_origin": "Digit 5 over-represented at Cabang Surabaya (19 trx).",
                            "low_sample_warning": False,
                        },
                    })])
                if "SELECT COUNT(*) FROM gold.fact_transactions" in sql:
                    return _FakeResult([_FakeRow({"cnt": 200})])
                if "SELECT COUNT(*) FROM silver.external_source_data" in sql:
                    return _FakeResult([_FakeRow({"cnt": 200})])
                if "FROM gold.fact_transactions" in sql:
                    return _FakeResult([_FakeRow({
                        "trx_id": 901,
                        "trx_ref_number": "REAL-55001",
                        "trx_date": "2026-07-01",
                        "digit": 5,
                        "amount": 500000000.0,
                        "branch_name": "Cabang Surabaya",
                        "category": "Transfer",
                        "customer_name": "PT Nusantara Real",
                        "channel": "RTGS",
                        "is_round_amount": True,
                        "is_after_hours": False,
                        "is_suspicious": True,
                        "authorization_status": "Pending",
                    })])
                return _FakeResult()

            fake_main.engine = _FakeEngine(real_responder)
            caatt_router.engine = fake_main.engine
            real_res = caatt_router.get_benford_results(source_id="src-real-1")
            self.assertEqual(len(real_res["data"]), 1)
            self.assertEqual(real_res["data"][0]["dominant_branch"], "Cabang Surabaya")
            self.assertEqual(real_res["data"][0]["chi_square_share_pct"], 64.5)
            self.assertEqual(len(real_res["findings"]), 1)
            self.assertEqual(real_res["findings"][0]["trx_ref_number"], "REAL-55001")
            self.assertEqual(real_res["summary"]["source_name"], "Bank Core Production")
            self.assertEqual(real_res["summary"]["significant_deviations"], 1)
            self.assertEqual(real_res["summary"]["gold_transactions_count"], 200)
            self.assertEqual(real_res["summary"]["silver_records_count"], 200)

            # Verify legacy 0 expected_count rows trigger automatic Benford recomputation
            def legacy_responder(sql, params):
                if "details IS NOT NULL AND details != '{}'::jsonb" in sql:
                    return _FakeResult([_FakeRow({"cnt": 9, "valid_cnt": 0})])
                if "FROM gold.fact_transactions" in sql and "LIMIT 1" in sql:
                    return _FakeResult([_FakeRow({"exists": 1})])
                return _FakeResult()

            legacy_engine = _FakeEngine(legacy_responder)
            caatt_router.engine = legacy_engine
            called_with = []
            orig_run_benford = caatt_router.run_benford_tests
            caatt_router.run_benford_tests = lambda eng, sid: called_with.append(sid) or 9
            try:
                caatt_router._ensure_benford_for_source("src-legacy-0")
            finally:
                caatt_router.run_benford_tests = orig_run_benford
            self.assertEqual(called_with, ["src-legacy-0"])
        finally:
            if prev_main is None:
                sys.modules.pop("main", None)
            else:
                sys.modules["main"] = prev_main

    def test_sources_register_echoes_ssl_mode_and_source_id(self):
        fake_main = types.ModuleType("main")
        fake_main.engine = _FakeEngine()
        prev_main = sys.modules.get("main")
        sys.modules["main"] = fake_main
        try:
            import routers.sources as sources_router
            sources_router = importlib.reload(sources_router)
            req = sources_router.SourceRegistration(
                source_id="src-ssl-1",
                name="External PG",
                host="db.example.internal",
                port=5432,
                database_name="core_bank",
                ssl_mode="require",
            )
            resp = sources_router.register_source(req)
            self.assertEqual(resp["status"], "success")
            self.assertEqual(resp["source_id"], "src-ssl-1")
            self.assertEqual(resp["ssl_mode"], "require")
        finally:
            if prev_main is None:
                sys.modules.pop("main", None)
            else:
                sys.modules["main"] = prev_main

    def test_silver_external_resolves_canonical_bronze_table_without_src_prefix(self):
        from transform.silver import transform_source_bronze_to_silver_external

        executed_sql = []

        def silver_responder(sql, params):
            executed_sql.append((sql, params))
            if "information_schema.columns" in sql:
                if params.get("tname") == "bank_prod_transactions":
                    return _FakeResult([
                        _FakeRow({"column_name": "_load_id"}),
                        _FakeRow({"column_name": "_source_id"}),
                        _FakeRow({"column_name": "trx_id"}),
                        _FakeRow({"column_name": "amount"}),
                    ])
                return _FakeResult([])
            if "INSERT INTO silver.external_source_data" in sql:
                return _FakeResult(rowcount=42)
            return _FakeResult()

        eng = _FakeEngine(silver_responder)
        rows = transform_source_bronze_to_silver_external(
            eng,
            source_id="src-ext-1",
            source_name="Bank Prod",
            table_name="transactions",
            mapping={"primaryKey": "trx_id"},
        )
        self.assertEqual(rows, 42)
        insert_sqls = [sql for sql, _ in executed_sql if "INSERT INTO silver.external_source_data" in sql]
        self.assertEqual(len(insert_sqls), 1)
        self.assertIn('bronze."bank_prod_transactions"', insert_sqls[0])
        self.assertNotIn("src_bank_prod_transactions", insert_sqls[0])
        self.assertIn("_source_id = :sid", insert_sqls[0])
        self.assertIn("- '_source_id'", insert_sqls[0])

    def test_caatt_router_closes_read_conn_before_computing_benford(self):
        fake_main = types.ModuleType("main")
        active_conns = {"count": 0}
        overlap_detected = {"flag": False}

        class _TrackingEngine(_FakeEngine):
            @contextmanager
            def connect(self):
                active_conns["count"] += 1
                try:
                    yield self.conn
                finally:
                    active_conns["count"] -= 1

        def benford_responder(sql, params):
            if "FROM bronze.registered_sources" in sql:
                return _FakeResult([_FakeRow({"name": "External Core"})])
            if "FROM gold.caatt_benford_results" in sql:
                return _FakeResult([])
            if "FROM gold.fact_transactions" in sql and "LIMIT 1" in sql:
                return _FakeResult([_FakeRow({"ok": 1})])
            return _FakeResult([])

        fake_main.engine = _TrackingEngine(benford_responder)
        prev_main = sys.modules.get("main")
        sys.modules["main"] = fake_main
        try:
            import routers.caatt as caatt_router
            caatt_router = importlib.reload(caatt_router)
            orig_run = caatt_router.run_benford_tests

            def checked_run(eng, sid):
                if active_conns["count"] > 0:
                    overlap_detected["flag"] = True
                return 9

            caatt_router.run_benford_tests = checked_run
            try:
                caatt_router.get_benford_results(source_id="src-real-lock-check")
            finally:
                caatt_router.run_benford_tests = orig_run

            self.assertFalse(overlap_detected["flag"])
        finally:
            if prev_main is None:
                sys.modules.pop("main", None)
            else:
                sys.modules["main"] = prev_main


def tearDownModule():
    for name, mod in list(_CREATED_STUBS.items()):
        if sys.modules.get(name) is mod:
            sys.modules.pop(name, None)


if __name__ == "__main__":
    unittest.main()
