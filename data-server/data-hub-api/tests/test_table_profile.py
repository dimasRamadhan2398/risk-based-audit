import os
import sys
import unittest

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from ingestion.profile import profile_client_table


class FakeCursor:
    def __init__(self, columns, total_rows, flagged_rows, metric_vals, dim_rows, flagged_sample, normal_sample):
        self.columns = columns
        self.total_rows = total_rows
        self.flagged_rows = flagged_rows
        self.metric_vals = metric_vals
        self.dim_rows = dim_rows
        self.flagged_sample = flagged_sample
        self.normal_sample = normal_sample
        self.description = []
        self._last_rows = []
        self._last_one = None

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        return False

    def execute(self, sql, params=None):
        s = " ".join(sql.split())
        if "SELECT * FROM" in s and "LIMIT 1;" in s:
            self.description = [(c,) for c in self.columns]
            self._last_rows = []
        elif "COUNT(*) AS total_rows" in s:
            self._last_one = (self.total_rows, self.flagged_rows, 5000000000.0, 250000000.0, 8500000000.0)
        elif "GROUP BY" in s:
            self._last_rows = self.dim_rows
        elif "COUNT(*) FILTER" in s:
            self._last_one = self.metric_vals
        elif "AND NOT (" in s:
            self.description = [(c,) for c in self.columns]
            self._last_rows = self.normal_sample
        else:
            self.description = [(c,) for c in self.columns]
            self._last_rows = self.flagged_sample

    def fetchone(self):
        return self._last_one

    def fetchall(self):
        return self._last_rows


class FakeConn:
    def __init__(self, cursor_factory):
        self._factory = cursor_factory

    def cursor(self):
        return self._factory()


class TestTableProfile(unittest.TestCase):
    def test_profile_corporate_transactions_stratified_and_metrics(self):
        cols = [
            "trx_id", "trx_ref_number", "trx_date", "branch_name", "customer_name",
            "category", "channel", "account_type", "amount", "authorization_status",
            "is_suspicious", "notes", "updated_at"
        ]
        flagged_row = (
            2801, "NSA-AFT8-00001", "2026-08-15 22:15:00", "KCU Sudirman Plaza",
            "PT Samudra Niaga Sejahtera", "Treasury Placement", "BRANCH_OVERRIDE",
            "ESCROW_PROJECT", 850000000.0, "PENDING_OVERRIDE", True,
            "Otorisasi override di luar jam operasional", "2026-11-16 09:10:00"
        )
        normal_row = (
            1, "NSA-2026-000001", "2026-07-02 09:07:00", "KCP Menteng Raya",
            "CV Timur Jaya Abadi", "Commercial Transfer", "SKNBI",
            "GIRO_KORPORASI", 15000000.0, "AUTHORIZED", False,
            "Transaksi operasional rutin korporasi", "2026-11-16 09:00:00"
        )
        conn = FakeConn(lambda: FakeCursor(
            columns=cols,
            total_rows=3244,
            flagged_rows=318,
            metric_vals=(245, 210),
            dim_rows=[
                ("KCU Sudirman Plaza", 390, 95, 45000000000.0, 8500000000.0),
                ("Kantor Pusat Operasional", 410, 88, 62000000000.0, 8500000000.0),
            ],
            flagged_sample=[flagged_row],
            normal_sample=[normal_row],
        ))

        res = profile_client_table(conn, "corporate_transactions", sample_limit=20)
        self.assertEqual(res["total_rows"], 3244)
        self.assertTrue(res["profile"]["has_numeric_measure"])
        self.assertEqual(res["profile"]["dimension_column"], "branch_name")
        self.assertEqual(res["profile"]["amount_column"], "amount")
        self.assertEqual(res["profile"]["flagged_rows"], 318)
        self.assertEqual(res["profile"]["metrics"]["suspicious_count"], 245)
        self.assertEqual(res["profile"]["metrics"]["unauthorized_count"], 210)
        self.assertEqual(len(res["profile"]["dimension_breakdown"]), 2)
        self.assertEqual(res["rows"][0]["trx_ref_number"], "NSA-AFT8-00001")
        self.assertTrue(res["rows"][0]["is_suspicious"])
        self.assertEqual(res["rows"][1]["trx_ref_number"], "NSA-2026-000001")

    def test_profile_branches_reference_table_has_no_numeric_measure(self):
        cols = ["cabang_id", "branch_code", "branch_name", "region", "city", "risk_tier", "updated_at"]
        sample_row = (1, "BR-001", "KCU Sudirman Plaza", "DKI Jakarta", "Jakarta Pusat", "High Risk", "2026-11-16")
        conn = FakeConn(lambda: FakeCursor(
            columns=cols,
            total_rows=8,
            flagged_rows=0,
            metric_vals=(),
            dim_rows=[("KCU Sudirman Plaza", 1, 0, 0.0, 0.0)],
            flagged_sample=[sample_row],
            normal_sample=[],
        ))

        res = profile_client_table(conn, "branches", sample_limit=10)
        self.assertEqual(res["total_rows"], 8)
        self.assertFalse(res["profile"]["has_numeric_measure"])
        self.assertIsNone(res["profile"]["amount_column"])

    def test_profile_rejects_sql_injection_identifier(self):
        conn = FakeConn(lambda: None)
        with self.assertRaises(ValueError):
            profile_client_table(conn, "corporate_transactions; DROP TABLE branches;--")


if __name__ == "__main__":
    unittest.main()
