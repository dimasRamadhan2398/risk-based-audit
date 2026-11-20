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


class TestEntityPrediction5TierRiskLevels(unittest.TestCase):
    def setUp(self):
        self.tmp_dir = tempfile.TemporaryDirectory()
        self.orig_parquet = storage.ENTITY_PREDICTION_PARQUET_FILE
        self.orig_bundle = ai_main.department_bundle
        self.orig_retrain_model_dir = retrain_engine.MODEL_DIR
        self.orig_retrain_ds_dir = retrain_engine.DATA_SERVER_MODEL_DIR
        storage.ENTITY_PREDICTION_PARQUET_FILE = os.path.join(
            self.tmp_dir.name, "entity_prediction_pool.parquet"
        )
        self.orig_pg_engine = storage._get_sqlalchemy_engine
        storage._get_sqlalchemy_engine = lambda: None
        retrain_engine.MODEL_DIR = os.path.join(self.tmp_dir.name, "models")
        retrain_engine.DATA_SERVER_MODEL_DIR = os.path.join(self.tmp_dir.name, "ds_models")
        os.makedirs(retrain_engine.MODEL_DIR, exist_ok=True)
        os.makedirs(retrain_engine.DATA_SERVER_MODEL_DIR, exist_ok=True)

    def tearDown(self):
        storage.ENTITY_PREDICTION_PARQUET_FILE = self.orig_parquet
        storage._get_sqlalchemy_engine = self.orig_pg_engine
        ai_main.department_bundle = self.orig_bundle
        retrain_engine.MODEL_DIR = self.orig_retrain_model_dir
        retrain_engine.DATA_SERVER_MODEL_DIR = self.orig_retrain_ds_dir
        self.tmp_dir.cleanup()

    def test_canonical_normalization_eliminates_non_five_levels(self):
        cases = {
            # Canonical pass-through
            "HIGH": ("HIGH", "High"),
            "High": ("HIGH", "High"),
            "MODERATE_HIGH": ("MODERATE_HIGH", "Moderate to High"),
            "Moderate to High": ("MODERATE_HIGH", "Moderate to High"),
            "MODERATE": ("MODERATE", "Moderate"),
            "Moderate": ("MODERATE", "Moderate"),
            "LOW_MODERATE": ("LOW_MODERATE", "Low to Moderate"),
            "Low to Moderate": ("LOW_MODERATE", "Low to Moderate"),
            "LOW": ("LOW", "Low"),
            "Low": ("LOW", "Low"),
            # Legacy / External non-5-tier labels
            "Critical": ("HIGH", "High"),
            "CRITICAL": ("HIGH", "High"),
            "Extreme": ("HIGH", "High"),
            "Very High": ("HIGH", "High"),
            "Catastrophic": ("HIGH", "High"),
            "Severe": ("HIGH", "High"),
            "Urgent": ("HIGH", "High"),
            "Sangat Tinggi": ("HIGH", "High"),
            "Kritis": ("HIGH", "High"),
            "Tier 1": ("HIGH", "High"),
            "Level 5": ("HIGH", "High"),
            "KOL-5": ("HIGH", "High"),
            "KOL-4": ("HIGH", "High"),
            "Major": ("MODERATE_HIGH", "Moderate to High"),
            "Significant": ("MODERATE_HIGH", "Moderate to High"),
            "Elevated": ("MODERATE_HIGH", "Moderate to High"),
            "Medium to High": ("MODERATE_HIGH", "Moderate to High"),
            "MEDIUM_TO_HIGH": ("MODERATE_HIGH", "Moderate to High"),
            "KOL-3": ("MODERATE_HIGH", "Moderate to High"),
            "Medium": ("MODERATE", "Moderate"),
            "MEDIUM": ("MODERATE", "Moderate"),
            "Watch": ("MODERATE", "Moderate"),
            "Warning": ("MODERATE", "Moderate"),
            "Not Scored": ("MODERATE", "Moderate"),
            "Unknown": ("MODERATE", "Moderate"),
            "Low to Medium": ("LOW_MODERATE", "Low to Moderate"),
            "LOW_TO_MEDIUM": ("LOW_MODERATE", "Low to Moderate"),
            "Minor": ("LOW_MODERATE", "Low to Moderate"),
            "Guarded": ("LOW_MODERATE", "Low to Moderate"),
            "KOL-2": ("LOW_MODERATE", "Low to Moderate"),
            "Insignificant": ("LOW", "Low"),
            "Negligible": ("LOW", "Low"),
            "Minimal": ("LOW", "Low"),
            "Trivial": ("LOW", "Low"),
            "Safe": ("LOW", "Low"),
            "Very Low": ("LOW", "Low"),
            "KOL-1": ("LOW", "Low"),
        }

        for raw, (expected_code, expected_label) in cases.items():
            code, label, imp, lik = ai_main.normalize_entity_risk_level(raw)
            self.assertEqual(code, expected_code, f"Failed code for raw={raw!r}")
            self.assertEqual(label, expected_label, f"Failed label for raw={raw!r}")
            self.assertIn(code, CANONICAL_CODES)
            self.assertIn(label, CANONICAL_LABELS)
            self.assertGreaterEqual(imp, 1)
            self.assertLessEqual(imp, 5)
            self.assertGreaterEqual(lik, 1)
            self.assertLessEqual(lik, 5)

        # Numeric & matrix alignment check (verifies Python 3 half-to-even rounding stays on diagonal)
        for code, (lik, imp) in storage.CANONICAL_RISK_CODE_TO_UI_FLOAT_PAIR.items():
            self.assertEqual(storage.matrix_risk_level(lik, imp), code)

    def test_training_pool_upsert_and_load_sanitizes_all_risk_levels(self):
        # 1. Even if raw poisoned records were written to Parquet directly, loading sanitizes them
        poisoned_df = pd.DataFrame([
            {
                "entity": "Unit Poisoned Critical",
                "entity_type": "Department",
                "target_timeline": "Q4 2026",
                "risk_category": "Operational",
                "inherent_likelihood": 9.5,
                "inherent_impact": 0.0,
                "audit_findings_count": 5.0,
                "kpi_below_target_count": 2.0,
                "kpi_volatility": 0.25,
                "previous_risk_score": 25.0,
                "assessment_month": 6,
                "target_likelihood": 7.0,
                "target_impact": 0.0,
                "current_risk_level": "Critical",
                "source_id": "external_db:legacy",
            },
            {
                "entity": "Unit Poisoned Insignificant",
                "entity_type": "Auditable Entity",
                "target_timeline": "FY 2026",
                "risk_category": "Financial",
                "inherent_likelihood": 1.0,
                "inherent_impact": 1.0,
                "audit_findings_count": 0.0,
                "kpi_below_target_count": 0.0,
                "kpi_volatility": 0.05,
                "previous_risk_score": 1.0,
                "assessment_month": 1,
                "target_likelihood": 1.0,
                "target_impact": 1.0,
                "current_risk_level": "Insignificant",
                "source_id": "external_db:legacy",
            },
        ])
        storage._write_pool_df(poisoned_df, storage.ENTITY_PREDICTION_PARQUET_FILE)

        loaded_poisoned = storage.load_entity_prediction_training_dataframe()
        self.assertEqual(len(loaded_poisoned), 2)
        for _, row in loaded_poisoned.iterrows():
            self.assertIn(row["current_risk_level"], CANONICAL_CODES)
            self.assertGreaterEqual(row["inherent_likelihood"], 1.0)
            self.assertLessEqual(row["inherent_likelihood"], 5.0)
            self.assertGreaterEqual(row["inherent_impact"], 1.0)
            self.assertLessEqual(row["inherent_impact"], 5.0)
            self.assertNotEqual(row["entity_type"], "Department")

        # 2. Upsert messy external records and verify pool contains strictly 5 levels
        messy_items = [
            {
                "entity": "Cabang Extreme Ext",
                "entity_type": "Branch",
                "risk_category": "Financial",
                "inherent_likelihood": 5.0,
                "inherent_impact": 5.0,
                "current_risk_level": "Extreme",
                "source_id": "external_db:core:loans",
            },
            {
                "entity": "Unit MedHigh Legacy",
                "entity_type": "Department",
                "risk_category": "Technology",
                "inherent_likelihood": 3.8,
                "inherent_impact": 4.0,
                "current_risk_level": "MEDIUM_TO_HIGH",
                "source_id": "audit_universe:it",
            },
            {
                "entity": "Unit Unknown Rating",
                "entity_type": "Auditable Entity",
                "risk_category": "Compliance",
                "inherent_likelihood": 3.0,
                "inherent_impact": 3.2,
                "current_risk_level": "Not Scored",
                "source_id": "external_db:compliance",
            },
        ]
        count = storage.upsert_entity_prediction_pool(messy_items)
        self.assertEqual(count, 3)
        df_pool = storage.load_entity_prediction_training_dataframe()
        self.assertEqual(len(df_pool), 5)
        levels = set(df_pool["current_risk_level"].unique())
        self.assertTrue(levels.issubset(CANONICAL_CODES), f"Unexpected levels found in training pool: {levels}")
        by_ent = {r["entity"]: r for _, r in df_pool.iterrows()}
        self.assertEqual(by_ent["Cabang Extreme Ext"]["current_risk_level"], "HIGH")
        self.assertEqual(by_ent["Unit MedHigh Legacy"]["current_risk_level"], "MODERATE_HIGH")
        self.assertEqual(by_ent["Unit MedHigh Legacy"]["entity_type"], "Auditable Entity")
        self.assertEqual(by_ent["Unit Unknown Rating"]["current_risk_level"], "MODERATE")

    def test_sync_and_predict_endpoints_enforce_five_levels(self):
        ai_main.department_bundle = None  # Deterministic 5x5 matrix path without mutating global bundle
        req = ai_main.EntityRiskSyncRequest(
            trigger_retrain=False,
            items=[
                ai_main.EntityRiskSyncItem(entity="Cabang Critical", current_risk_level="Critical", source_id="external_db:a"),
                ai_main.EntityRiskSyncItem(entity="Cabang Extreme", current_risk_level="Extreme", source_id="external_db:b"),
                ai_main.EntityRiskSyncItem(entity="Unit Major", current_risk_level="Major", source_id="external_db:c"),
                ai_main.EntityRiskSyncItem(entity="Unit Minor", current_risk_level="Minor", source_id="external_db:d"),
                ai_main.EntityRiskSyncItem(entity="Unit Insignificant", current_risk_level="Insignificant", source_id="external_db:e"),
                ai_main.EntityRiskSyncItem(entity="Cabang KOL-3", current_risk_level="KOL-3", source_id="external_db:f"),
                ai_main.EntityRiskSyncItem(entity="Unit Not Scored", current_risk_level="Not Scored", source_id="external_db:g"),
            ],
        )
        res = ai_main.sync_entity_prediction_with_sources(req)
        self.assertEqual(res["status"], "success")
        preds = res["predictions"]
        self.assertEqual(len(preds), 7)

        expected_actual = {
            "Cabang Critical": "HIGH",
            "Cabang Extreme": "HIGH",
            "Unit Major": "MODERATE_HIGH",
            "Unit Minor": "LOW_MODERATE",
            "Unit Insignificant": "LOW",
            "Cabang KOL-3": "MODERATE_HIGH",
            "Unit Not Scored": "MODERATE",
        }
        for p in preds:
            self.assertIn(p["risk_level"], CANONICAL_CODES)
            self.assertIn(p["actual_risk_level"], CANONICAL_CODES)
            self.assertNotIn(p["risk_level"], {"CRITICAL", "EXTREME", "INSIGNIFICANT", "MEDIUM", "MEDIUM_TO_HIGH"})
            self.assertNotIn(p["actual_risk_level"], {"CRITICAL", "EXTREME", "INSIGNIFICANT", "MEDIUM", "MEDIUM_TO_HIGH"})
            self.assertEqual(p["actual_risk_level"], expected_actual[p["entity"]])

    def test_retrain_department_suite_trains_strictly_on_five_levels(self):
        # Seed 30 rows spanning all 5 canonical tiers + non-canonical aliases
        raw_levels = ["Critical", "Extreme", "Major", "MEDIUM_TO_HIGH", "Moderate", "Minor", "Insignificant", "Safe"]
        records = []
        for idx in range(32):
            raw_lvl = raw_levels[idx % len(raw_levels)]
            code, _, imp, lik = storage.normalize_canonical_risk_level(raw_lvl)
            records.append({
                "entity": f"Unit Test {idx}",
                "entity_type": "Branch" if idx % 2 == 0 else "Auditable Entity",
                "target_timeline": "FY 2026",
                "risk_category": "Financial",
                "inherent_likelihood": float(lik),
                "inherent_impact": float(imp),
                "audit_findings_count": float((idx % 4) + 1),
                "kpi_below_target_count": float(idx % 3),
                "kpi_volatility": 0.15,
                "previous_risk_score": float(lik * imp),
                "assessment_month": (idx % 12) + 1,
                "target_likelihood": float(lik),
                "target_impact": float(imp),
                "current_risk_level": raw_lvl,
                "source_id": "external_db:test",
            })
        storage.upsert_entity_prediction_pool(records)

        engine = retrain_engine.AutoRetrainEngine(model_dir=retrain_engine.MODEL_DIR)
        ok = engine.retrain_department_suite()
        self.assertTrue(ok)
        bundle_path = os.path.join(retrain_engine.MODEL_DIR, "department_bundle.pkl")
        self.assertTrue(os.path.exists(bundle_path))

        import pickle
        with open(bundle_path, "rb") as f:
            bundle = pickle.load(f)
        self.assertEqual(bundle.get("allowed_risk_levels"), list(storage.CANONICAL_RISK_CODES))
        self.assertTrue(set(bundle["label_encoder_impact"].classes_).issubset({1, 2, 3, 4, 5}))
        self.assertTrue(set(bundle["label_encoder_likelihood"].classes_).issubset({1, 2, 3, 4, 5}))

    def test_data_server_ai_engine_normalization_parity(self):
        ds_dir = os.path.abspath(os.path.join(CURRENT_DIR, "..", "..", "data-server", "ai-engine"))
        spec = importlib.util.spec_from_file_location(
            "ds_ai_training_storage",
            os.path.join(ds_dir, "ai_training_storage.py"),
        )
        ds_storage = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(ds_storage)
        self.assertEqual(tuple(ds_storage.CANONICAL_RISK_CODES), tuple(storage.CANONICAL_RISK_CODES))
        for raw, expected in [
            ("Critical", "HIGH"),
            ("Extreme", "HIGH"),
            ("MEDIUM_TO_HIGH", "MODERATE_HIGH"),
            ("LOW_TO_MEDIUM", "LOW_MODERATE"),
            ("Insignificant", "LOW"),
            ("Not Scored", "MODERATE"),
        ]:
            code, label, _, _ = ds_storage.normalize_canonical_risk_level(raw)
            self.assertEqual(code, expected)
            self.assertIn(label, CANONICAL_LABELS)


if __name__ == "__main__":
    unittest.main()
