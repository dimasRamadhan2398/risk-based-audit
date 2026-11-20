import importlib.util
import json
import os
import pickle
import sys
import tempfile
import unittest
from unittest.mock import patch

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
if CURRENT_DIR not in sys.path:
    sys.path.insert(0, CURRENT_DIR)

import ai_training_storage as storage
import main as ai_main
import nlp_taxonomy as nt
import retrain_engine


CANONICAL_CODES = {"HIGH", "MODERATE_HIGH", "MODERATE", "LOW_MODERATE", "LOW"}
CANONICAL_LABELS = {"High", "Moderate to High", "Moderate", "Low to Moderate", "Low"}


class _DummyResponse:
    def __init__(self, status_code, payload):
        self.status_code = status_code
        self._payload = payload

    def json(self):
        return self._payload


class TestNLPRiskLevels5TierEnforcement(unittest.TestCase):
    def setUp(self):
        self.tmp_dir = tempfile.TemporaryDirectory()
        self.orig_docs_file = nt.IMPORTED_DOCS_FILE
        self.orig_corpus_file = nt.LIVE_CORPUS_FILE
        self.orig_tax_file = nt.TAXONOMY_FILE
        self.orig_model_dir = retrain_engine.MODEL_DIR
        self.orig_ds_dir = retrain_engine.DATA_SERVER_MODEL_DIR
        self.orig_bundle = ai_main.document_bundle

        nt.IMPORTED_DOCS_FILE = os.path.join(self.tmp_dir.name, "imported_nlp_documents.json")
        nt.LIVE_CORPUS_FILE = os.path.join(self.tmp_dir.name, "live_nlp_corpus.json")
        nt.TAXONOMY_FILE = os.path.join(self.tmp_dir.name, "nlp_taxonomy.json")

        retrain_engine.MODEL_DIR = os.path.join(self.tmp_dir.name, "models")
        retrain_engine.DATA_SERVER_MODEL_DIR = os.path.join(self.tmp_dir.name, "ds_models")
        os.makedirs(retrain_engine.MODEL_DIR, exist_ok=True)
        os.makedirs(retrain_engine.DATA_SERVER_MODEL_DIR, exist_ok=True)

    def tearDown(self):
        nt.IMPORTED_DOCS_FILE = self.orig_docs_file
        nt.LIVE_CORPUS_FILE = self.orig_corpus_file
        nt.TAXONOMY_FILE = self.orig_tax_file
        retrain_engine.MODEL_DIR = self.orig_model_dir
        retrain_engine.DATA_SERVER_MODEL_DIR = self.orig_ds_dir
        ai_main.document_bundle = self.orig_bundle
        self.tmp_dir.cleanup()

    def test_normalization_maps_external_and_legacy_levels_strictly_onto_five_tiers(self):
        cases = {
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
            "Critical": ("HIGH", "High"),
            "CRITICAL": ("HIGH", "High"),
            "Extreme": ("HIGH", "High"),
            "Very High": ("HIGH", "High"),
            "Catastrophic": ("HIGH", "High"),
            "Sangat Tinggi": ("HIGH", "High"),
            "Kritis": ("HIGH", "High"),
            "Major": ("MODERATE_HIGH", "Moderate to High"),
            "Significant": ("MODERATE_HIGH", "Moderate to High"),
            "Elevated": ("MODERATE_HIGH", "Moderate to High"),
            "Medium to High": ("MODERATE_HIGH", "Moderate to High"),
            "KOL-3": ("MODERATE_HIGH", "Moderate to High"),
            "Medium": ("MODERATE", "Moderate"),
            "MEDIUM": ("MODERATE", "Moderate"),
            "Watch": ("MODERATE", "Moderate"),
            "Not Scored": ("MODERATE", "Moderate"),
            "Minor": ("LOW_MODERATE", "Low to Moderate"),
            "Guarded": ("LOW_MODERATE", "Low to Moderate"),
            "Low to Medium": ("LOW_MODERATE", "Low to Moderate"),
            "KOL-2": ("LOW_MODERATE", "Low to Moderate"),
            "Insignificant": ("LOW", "Low"),
            "Negligible": ("LOW", "Low"),
            "Minimal": ("LOW", "Low"),
            "Safe": ("LOW", "Low"),
            "KOL-1": ("LOW", "Low"),
        }
        for raw, (exp_code, exp_label) in cases.items():
            code, label, imp, lik = nt.normalize_nlp_risk_level(raw)
            self.assertEqual(code, exp_code, f"Failed code for raw={raw!r}")
            self.assertEqual(label, exp_label, f"Failed label for raw={raw!r}")
            self.assertIn(code, CANONICAL_CODES)
            self.assertIn(label, CANONICAL_LABELS)
            self.assertGreaterEqual(imp, 1)
            self.assertLessEqual(imp, 5)
            self.assertGreaterEqual(lik, 1)
            self.assertLessEqual(lik, 5)

    def test_load_and_harvest_sanitize_poisoned_training_data(self):
        # Write poisoned persistent JSONs with non-5-tier values (CRITICAL, Extreme, Medium, Minor, Insignificant)
        poisoned_docs = [
            {
                "docId": "EXT-POISON-1",
                "title": "External Poison Critical",
                "source": "External DB",
                "sourceType": "External Database",
                "risk_category": "Compliance",
                "sentiment": "Negative",
                "impact": 5,
                "likelihood": 5,
                "severityScore": 95,
                "excerpt": "Ditemukan transaksi rekening dormant terindikasi fraud tanpa pelaporan PPATK.",
                "risk_level": "CRITICAL",
            },
            {
                "docId": "EXT-POISON-2",
                "title": "External Poison Insignificant",
                "source": "External DB",
                "sourceType": "External Database",
                "risk_category": "Operational",
                "sentiment": "Positive",
                "impact": 1,
                "likelihood": 1,
                "severityScore": 15,
                "excerpt": "Prosedur administrasi cabang berjalan aman tanpa temuan material.",
                "risk_level": "Insignificant",
            },
        ]
        with open(nt.IMPORTED_DOCS_FILE, "w", encoding="utf-8") as f:
            json.dump(poisoned_docs, f)

        poisoned_corpus = [
            {
                "text": "Rekaman eksternal menandai pelanggaran otorisasi kas dengan rating Extreme Risk.",
                "risk_category": "Financial",
                "sentiment": "Negative",
                "impact": 5,
                "likelihood": 4,
                "risk_level": "Extreme",
                "control_indicator": "Fraud Indication",
            },
            {
                "text": "Catatan audit lapangan menunjukkan temuan minor pengarsipan dokumen operasional.",
                "risk_category": "Operational",
                "sentiment": "Neutral",
                "impact": 2,
                "likelihood": 2,
                "risk_level": "Minor",
                "control_indicator": "Internal Control Weakness",
            },
            {
                "text": "Tinjauan kontrol umum berada pada tingkat Medium tanpa selisih signifikan.",
                "risk_category": "Compliance",
                "sentiment": "Neutral",
                "impact": 3,
                "likelihood": 3,
                "risk_level": "MEDIUM",
                "control_indicator": "Compliance Deviation",
            },
        ]
        with open(nt.LIVE_CORPUS_FILE, "w", encoding="utf-8") as f:
            json.dump(poisoned_corpus, f)

        loaded_docs = nt.load_analyzed_documents()
        self.assertEqual([d["risk_level"] for d in loaded_docs], ["HIGH", "LOW"])
        for d in loaded_docs:
            self.assertIn(d["risk_level"], CANONICAL_CODES)
            self.assertIn(d["risk_level_label"], CANONICAL_LABELS)

        loaded_corpus = nt.load_live_nlp_corpus()
        self.assertEqual([c["risk_level"] for c in loaded_corpus], ["HIGH", "LOW_MODERATE", "MODERATE"])

        harvested = nt.harvest_all_live_nlp_training_data()
        harvested_levels = {s.get("risk_level") for s in harvested}
        self.assertTrue(harvested_levels.issubset(CANONICAL_CODES), f"Unexpected training risk levels: {harvested_levels}")
        self.assertFalse(harvested_levels.intersection({"CRITICAL", "EXTREME", "MEDIUM", "MINOR", "INSIGNIFICANT"}))

    def test_retrain_document_suite_learns_only_five_canonical_risk_levels(self):
        # Retrain using clean multi-source corpus (with some non-canonical entries pre-seeded in file)
        nt.upsert_live_nlp_corpus([
            {
                "text": "Insiden keamanan akses core banking dengan eskalasi rating Critical dari SIEM.",
                "risk_category": "Technology",
                "sentiment": "Negative",
                "impact": 5,
                "likelihood": 5,
                "risk_level": "Critical",
                "control_indicator": "IT & Access Control Weakness",
            },
            {
                "text": "Keterlambatan dokumentasi minor pencatatan logistik gudang tanpa selisih fisik.",
                "risk_category": "Operational",
                "sentiment": "Neutral",
                "impact": 2,
                "likelihood": 2,
                "risk_level": "Minor",
                "control_indicator": "Internal Control Weakness",
            },
        ])

        engine = retrain_engine.AutoRetrainEngine(model_dir=retrain_engine.MODEL_DIR)
        ok = engine.retrain_document_suite()
        self.assertTrue(ok)

        bundle_path = os.path.join(retrain_engine.MODEL_DIR, "document_bundle.pkl")
        self.assertTrue(os.path.exists(bundle_path))
        with open(bundle_path, "rb") as f:
            bundle = pickle.load(f)

        self.assertEqual(bundle.get("allowed_risk_levels"), list(storage.CANONICAL_RISK_CODES))
        self.assertEqual(bundle.get("allowed_risk_level_labels"), list(storage.CANONICAL_RISK_CODE_TO_LABEL.values()))
        trained_classes = set(bundle["label_encoders"]["risk_level"].classes_)
        self.assertTrue(trained_classes.issubset(CANONICAL_CODES), f"Non-5-tier classes found in NLP model: {trained_classes}")

    def test_predict_and_sync_endpoints_adapt_external_non_five_levels(self):
        ai_main.document_bundle = None  # Pure deterministic path

        ext_inputs = [
            ("Critical", "HIGH", "High"),
            ("Extreme", "HIGH", "High"),
            ("Major", "MODERATE_HIGH", "Moderate to High"),
            ("KOL-3", "MODERATE_HIGH", "Moderate to High"),
            ("Medium", "MODERATE", "Moderate"),
            ("Minor", "LOW_MODERATE", "Low to Moderate"),
            ("Insignificant", "LOW", "Low"),
        ]
        for raw_ext, exp_code, exp_label in ext_inputs:
            res = ai_main.predict_text_analysis(
                ai_main.TextAnalysisRequest(
                    text=f"Temuan database eksternal atas anomali transaksi dengan peringkat risiko {raw_ext}.",
                    title=f"Ext Doc {raw_ext}",
                    source="External DB: Core Audit",
                    sourceType="External Database",
                    fileType="DB_RECORD",
                    riskLevel=raw_ext,
                    persist=True,
                )
            )
            self.assertEqual(res["risk_level"], exp_code, f"Failed predict code for {raw_ext!r}")
            self.assertEqual(res["risk_level_label"], exp_label, f"Failed predict label for {raw_ext!r}")
            self.assertIn(res["risk_level"], CANONICAL_CODES)
            self.assertIn(res["risk_level_label"], CANONICAL_LABELS)

        # Verify sync_multi_source_nlp_corpus adapts external mapped-records with Critical/Medium/Insignificant
        def mock_get(url, timeout=2):
            if "mapped-records" in url:
                return _DummyResponse(200, {
                    "data": [
                        {
                            "connection_name": "Core Banking DB",
                            "table_name": "audit_risk_incidents",
                            "sample_records": [
                                {"incident_title": "Override Otorisasi Cabang", "notes": "Terjadi bypass dual control kas", "severity": "Critical"},
                                {"incident_title": "Delay Pelaporan Rutin", "notes": "Laporan telat 1 hari tanpa sanksi", "risk_rating": "Medium"},
                                {"incident_title": "Verifikasi Arsip Harian", "notes": "Arsip terverifikasi lengkap dan aman", "risk_level": "Insignificant"},
                            ],
                        }
                    ]
                })
            return _DummyResponse(200, {"data": []})

        engine = retrain_engine.AutoRetrainEngine(model_dir=retrain_engine.MODEL_DIR)
        with patch.object(retrain_engine.requests, "get", side_effect=mock_get):
            sync_counts = engine.sync_multi_source_nlp_corpus()
        self.assertGreaterEqual(sync_counts["external_db"], 3)

        batch = ai_main.get_text_analysis_batch()
        self.assertEqual(batch.get("allowed_risk_levels"), list(storage.CANONICAL_RISK_CODES))
        for doc in batch["documents"]:
            self.assertIn(doc["risk_level"], CANONICAL_CODES)
            self.assertIn(doc["risk_level_label"], CANONICAL_LABELS)
            self.assertNotIn(doc["risk_level"], {"CRITICAL", "EXTREME", "MEDIUM", "MINOR", "INSIGNIFICANT"})

    def test_data_server_ai_engine_nlp_taxonomy_parity(self):
        ds_dir = os.path.abspath(os.path.join(CURRENT_DIR, "..", "..", "data-server", "ai-engine"))
        spec = importlib.util.spec_from_file_location(
            "ds_nlp_taxonomy",
            os.path.join(ds_dir, "nlp_taxonomy.py"),
        )
        ds_nt = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(ds_nt)
        for raw, expected in [
            ("Critical", "HIGH"),
            ("Extreme", "HIGH"),
            ("Major", "MODERATE_HIGH"),
            ("Medium", "MODERATE"),
            ("Minor", "LOW_MODERATE"),
            ("Insignificant", "LOW"),
        ]:
            code, label, _, _ = ds_nt.normalize_nlp_risk_level(raw)
            self.assertEqual(code, expected)
            self.assertIn(label, CANONICAL_LABELS)


if __name__ == "__main__":
    unittest.main()
