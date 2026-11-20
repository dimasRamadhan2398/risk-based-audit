import importlib.util
import os
import sys
import tempfile
import unittest

import pandas as pd

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
if CURRENT_DIR not in sys.path:
    sys.path.insert(0, CURRENT_DIR)

import ai_training_storage as storage
import main as ai_main
import retrain_engine


CANONICAL_CODES = {"HIGH", "MODERATE_HIGH", "MODERATE", "LOW_MODERATE", "LOW"}
CANONICAL_LABELS = {"High", "Moderate to High", "Moderate", "Low to Moderate", "Low"}


class TestAuditeeKpi5TierRiskLevels(unittest.TestCase):
    def setUp(self):
        self.tmp_dir = tempfile.TemporaryDirectory()
        self.orig_parquet = storage.AUDITEE_PARQUET_FILE
        self.orig_pg_engine = storage._get_sqlalchemy_engine
        self.orig_kpi_model = ai_main.kpi_lstm_model
        self.orig_kpi_bundle = ai_main.kpi_bundle
        self.orig_retrain_model_dir = retrain_engine.MODEL_DIR
        self.orig_retrain_ds_dir = retrain_engine.DATA_SERVER_MODEL_DIR
        self.orig_load_df = ai_main.load_auditee_training_dataframe

        storage.AUDITEE_PARQUET_FILE = os.path.join(
            self.tmp_dir.name, "auditee_kpi_pool.parquet"
        )
        storage._get_sqlalchemy_engine = lambda: None
        retrain_engine.MODEL_DIR = os.path.join(self.tmp_dir.name, "models")
        retrain_engine.DATA_SERVER_MODEL_DIR = os.path.join(self.tmp_dir.name, "ds_models")
        os.makedirs(retrain_engine.MODEL_DIR, exist_ok=True)
        os.makedirs(retrain_engine.DATA_SERVER_MODEL_DIR, exist_ok=True)

    def tearDown(self):
        storage.AUDITEE_PARQUET_FILE = self.orig_parquet
        storage._get_sqlalchemy_engine = self.orig_pg_engine
        ai_main.kpi_lstm_model = self.orig_kpi_model
        ai_main.kpi_bundle = self.orig_kpi_bundle
        ai_main.load_auditee_training_dataframe = self.orig_load_df
        retrain_engine.MODEL_DIR = self.orig_retrain_model_dir
        retrain_engine.DATA_SERVER_MODEL_DIR = self.orig_retrain_ds_dir
        self.tmp_dir.cleanup()

    def test_upsert_and_load_auditee_pool_strictly_normalize_risk_levels(self):
        rows = [
            {
                "id": "KPI-EXT-001",
                "kpi_name": "NPL Kredit Eksternal A",
                "period_label": "Bulan 05 2026",
                "actual_value": 6.4,
                "impact": 3,
                "likelihood": 3,
                "risk_level": "CRITICAL",
            },
            {
                "id": "KPI-EXT-002",
                "kpi_name": "Yield Kredit Eksternal B",
                "period_label": "Bulan 05 2026",
                "actual_value": 8.9,
                "risk_level": "MEDIUM_TO_HIGH",
            },
            {
                "id": "KPI-EXT-003",
                "kpi_name": "Efisiensi Biaya Eksternal C",
                "period_label": "Bulan 05 2026",
                "actual_value": 72.1,
                "risk_level": "MEDIUM",
            },
            {
                "id": "KPI-EXT-004",
                "kpi_name": "Deviasi Cabang Eksternal D",
                "period_label": "Bulan 05 2026",
                "actual_value": 2.1,
                "risk_level": "Minor",
            },
            {
                "id": "KPI-EXT-005",
                "kpi_name": "Kesehatan Likuiditas Eksternal E",
                "period_label": "Bulan 05 2026",
                "actual_value": 152.0,
                "risk_level": "Very Low",
            },
        ]

        storage.upsert_auditee_kpi_pool(rows)
        df = storage.load_auditee_training_dataframe()

        self.assertIn("Risk Level", df.columns)
        sub = df[df["ID KPI"].astype(str).str.startswith("KPI-EXT-")]
        self.assertEqual(len(sub), 5)

        by_id = {row["ID KPI"]: row for _, row in sub.iterrows()}
        self.assertEqual(by_id["KPI-EXT-001"]["Risk Level"], "HIGH")
        self.assertEqual(by_id["KPI-EXT-002"]["Risk Level"], "MODERATE_HIGH")
        self.assertEqual(by_id["KPI-EXT-003"]["Risk Level"], "MODERATE")
        self.assertEqual(by_id["KPI-EXT-004"]["Risk Level"], "LOW_MODERATE")
        self.assertEqual(by_id["KPI-EXT-005"]["Risk Level"], "LOW")

        for _, row in df.iterrows():
            imp = int(row["TARGET: Impact (1-5)"])
            lik = int(row["TARGET: Likelihood (1-5)"])
            self.assertGreaterEqual(imp, 1)
            self.assertLessEqual(imp, 5)
            self.assertGreaterEqual(lik, 1)
            self.assertLessEqual(lik, 5)
            self.assertIn(row["Risk Level"], CANONICAL_CODES)
            self.assertEqual(storage.matrix_risk_level(lik, imp), row["Risk Level"])

    def test_ingest_custom_auditee_items_maps_non_canonical_external_values(self):
        engine = retrain_engine.AutoRetrainEngine(model_dir=retrain_engine.MODEL_DIR)
        custom_items = [
            {
                "code": "KPI-EXT-NPL",
                "kpiName": "Rasio NPL Kredit Korporasi",
                "entity": "Divisi Kredit",
                "riskLevel": "CRITICAL",
                "impactScore": 2,
                "likelihoodScore": 2,
                "history": [
                    {"period": "2026-01", "value": 4.9},
                    {"period": "2026-02", "value": 5.4},
                ],
            },
            {
                "code": "KPI-EXT-YLD",
                "kpiName": "Yield Suku Bunga Kredit",
                "entity": "Divisi Komersial",
                "riskLevel": "Major",
                "history": [
                    {"period": "2026-01", "value": 9.8},
                    {"period": "2026-02", "value": 9.4},
                ],
            },
            {
                "code": "KPI-EXT-OPS",
                "kpiName": "Deviasi Operasional Regional",
                "entity": "Operasional",
                "riskLevel": "KOL-2",
                "history": [{"period": "2026-02", "value": 3.1}],
            },
        ]

        h = engine._ingest_custom_auditee_items(custom_items)
        self.assertTrue(h)
        df = storage.load_auditee_training_dataframe()
        ext = df[df["Nama KPI"].isin([
            "Rasio NPL Kredit Korporasi",
            "Yield Suku Bunga Kredit",
            "Deviasi Operasional Regional",
        ])]
        self.assertGreaterEqual(len(ext), 5)
        self.assertTrue(set(ext["Risk Level"].unique()).issubset(CANONICAL_CODES))

        npl_rows = ext[ext["Nama KPI"] == "Rasio NPL Kredit Korporasi"]
        self.assertTrue(all(v == "HIGH" for v in npl_rows["Risk Level"]))
        # Mismatched 2x2 score was realigned so model training strictly learns HIGH
        self.assertTrue(all(int(v) == 5 for v in npl_rows["TARGET: Impact (1-5)"]))

        yld_rows = ext[ext["Nama KPI"] == "Yield Suku Bunga Kredit"]
        self.assertTrue(all(v == "MODERATE_HIGH" for v in yld_rows["Risk Level"]))

        ops_rows = ext[ext["Nama KPI"] == "Deviasi Operasional Regional"]
        self.assertTrue(all(v == "LOW_MODERATE" for v in ops_rows["Risk Level"]))

    def test_sanitize_heals_legacy_poisoned_parquet_and_retrains_canonically(self):
        poisoned = pd.DataFrame([
            {"ID KPI": "OLD-1", "Nama KPI": "Legacy Critical KPI", "Periode": "Bulan 01 2026", "TARGET: Nilai Aktual (%)": 9.1, "TARGET: Impact (1-5)": 9, "TARGET: Likelihood (1-5)": 9, "Risk Level": "CRITICAL"},
            {"ID KPI": "OLD-2", "Nama KPI": "Legacy Critical KPI", "Periode": "Bulan 02 2026", "TARGET: Nilai Aktual (%)": 9.5, "TARGET: Impact (1-5)": 5, "TARGET: Likelihood (1-5)": 4, "Risk Level": "Extreme"},
            {"ID KPI": "OLD-3", "Nama KPI": "Legacy Medium KPI", "Periode": "Bulan 01 2026", "TARGET: Nilai Aktual (%)": 40.0, "TARGET: Impact (1-5)": 3, "TARGET: Likelihood (1-5)": 3, "Risk Level": "MEDIUM"},
            {"ID KPI": "OLD-4", "Nama KPI": "Legacy Medium KPI", "Periode": "Bulan 02 2026", "TARGET: Nilai Aktual (%)": 42.0, "TARGET: Impact (1-5)": 1, "TARGET: Likelihood (1-5)": 1, "Risk Level": "Insignificant"},
            {"ID KPI": "OLD-5", "Nama KPI": "Legacy Elevated KPI", "Periode": "Bulan 01 2026", "TARGET: Nilai Aktual (%)": 14.0, "TARGET: Impact (1-5)": 4, "TARGET: Likelihood (1-5)": 3, "Risk Level": "MEDIUM_TO_HIGH"},
            {"ID KPI": "OLD-6", "Nama KPI": "Legacy Guarded KPI", "Periode": "Bulan 01 2026", "TARGET: Nilai Aktual (%)": 4.0, "TARGET: Impact (1-5)": 2, "TARGET: Likelihood (1-5)": 2, "Risk Level": "Guarded"},
        ])
        storage._write_pool_df(poisoned, storage.AUDITEE_PARQUET_FILE)

        loaded = storage.load_auditee_training_dataframe()
        self.assertTrue(set(loaded["Risk Level"].unique()).issubset(CANONICAL_CODES))
        self.assertNotIn("CRITICAL", set(loaded["Risk Level"].unique()))
        self.assertNotIn("MEDIUM", set(loaded["Risk Level"].unique()))
        self.assertNotIn("MEDIUM_TO_HIGH", set(loaded["Risk Level"].unique()))

        engine = retrain_engine.AutoRetrainEngine(model_dir=retrain_engine.MODEL_DIR)
        ok = engine.retrain_kpi_lstm_suite()
        self.assertTrue(ok)
        import pickle
        with open(os.path.join(retrain_engine.MODEL_DIR, "kpi_bundle.pkl"), "rb") as f:
            bundle = pickle.load(f)
        self.assertEqual(bundle.get("canonical_risk_levels"), list(storage.CANONICAL_RISK_CODES))

    def test_prediction_and_batch_endpoints_emit_only_5_canonical_levels(self):
        for raw_ext, exp_code in [
            ("CRITICAL", "HIGH"),
            ("Extreme", "HIGH"),
            ("MEDIUM_TO_HIGH", "MODERATE_HIGH"),
            ("MEDIUM", "MODERATE"),
            ("Minor", "LOW_MODERATE"),
            ("Insignificant", "LOW"),
        ]:
            pred = ai_main.predict_performance_trend(
                ai_main.PerformanceTrendRequest(
                    kpi_name="Test External KPI",
                    historical_data=[10.0, 11.0, 12.0],
                    risk_level=raw_ext,
                )
            )
            self.assertEqual(pred["risk_level"], exp_code)
            self.assertIn(pred["risk_level"], CANONICAL_CODES)
            self.assertIn(pred["risk_level_label"], CANONICAL_LABELS)

        batch_df = pd.DataFrame([
            {"ID KPI": "KPI-EXT-01", "Nama KPI": "Rasio NPL Eksternal", "Periode": "Bulan 04 2026", "TARGET: Nilai Aktual (%)": 6.2, "TARGET: Impact (1-5)": 5, "TARGET: Likelihood (1-5)": 5, "Risk Level": "CRITICAL"},
            {"ID KPI": "KPI-EXT-01", "Nama KPI": "Rasio NPL Eksternal", "Periode": "Bulan 05 2026", "TARGET: Nilai Aktual (%)": 6.8, "TARGET: Impact (1-5)": 5, "TARGET: Likelihood (1-5)": 5, "Risk Level": "CRITICAL"},
            {"ID KPI": "KPI-EXT-02", "Nama KPI": "Yield Kredit Komersial", "Periode": "Bulan 05 2026", "TARGET: Nilai Aktual (%)": 9.1, "TARGET: Impact (1-5)": 4, "TARGET: Likelihood (1-5)": 4, "Risk Level": "MEDIUM_TO_HIGH"},
            {"ID KPI": "KPI-EXT-03", "Nama KPI": "BOPO Cabang Pusat", "Periode": "Bulan 05 2026", "TARGET: Nilai Aktual (%)": 78.0, "TARGET: Impact (1-5)": 3, "TARGET: Likelihood (1-5)": 3, "Risk Level": "MEDIUM"},
        ])
        ai_main.load_auditee_training_dataframe = lambda: batch_df
        res = ai_main.get_performance_trend_batch()
        auditee_block = res.get("auditee_forecast", {})
        forecasts = auditee_block.get("kpi_forecasts", [])
        at_risk = auditee_block.get("at_risk_departments", [])

        self.assertEqual(len(forecasts), 3)
        for item in forecasts:
            self.assertIn(item["riskLevel"], CANONICAL_CODES)
            self.assertIn(item["riskLevelLabel"], CANONICAL_LABELS)
        for dep in at_risk:
            self.assertIn(dep["riskLevel"], CANONICAL_CODES)
            self.assertIn(dep["riskLevelLabel"], CANONICAL_LABELS)

    def test_data_server_mirror_storage_enforces_5_levels(self):
        ds_path = os.path.abspath(
            os.path.join(CURRENT_DIR, "..", "..", "data-server", "ai-engine", "ai_training_storage.py")
        )
        spec = importlib.util.spec_from_file_location("ds_ai_training_storage_auditee", ds_path)
        ds_storage = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(ds_storage)

        for raw, exp_code in [
            ("CRITICAL", "HIGH"),
            ("MEDIUM_TO_HIGH", "MODERATE_HIGH"),
            ("MEDIUM", "MODERATE"),
            ("Guarded", "LOW_MODERATE"),
            ("Very Low", "LOW"),
        ]:
            code, imp, lik = ds_storage._resolve_auditee_record_risk({"risk_level": raw})
            self.assertEqual(code, exp_code)
            self.assertIn(code, CANONICAL_CODES)
            self.assertGreaterEqual(imp, 1)
            self.assertGreaterEqual(lik, 1)


if __name__ == "__main__":
    unittest.main()
