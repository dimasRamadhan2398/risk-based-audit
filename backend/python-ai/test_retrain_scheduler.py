"""
Unit & Integration Tests for AuditSphere Per-Model AI Retraining,
Multi-Lane Scheduler, Resource Profile, Tier Selection, Validation Rollback, and Model Isolation.
"""
import os
import time
import json
import shutil
import tempfile
import unittest
import pandas as pd

from resource_profile import get_resource_profile, select_training_tier
from training_data_source import stratified_reservoir_sample_batches, iter_dataframe_batches
from model_registry import (
    MODEL_KEYS,
    read_raw_registry,
    write_raw_registry,
    validate_and_promote_artifacts,
)
from retrain_scheduler import RetrainScheduler
from retrain_engine import AutoRetrainEngine


class ResourceProfileAndTierTests(unittest.TestCase):
    def test_auto_detect_2vcpu_8gb_vs_8vcpu_vs_16vcpu(self):
        # Clear any env overrides
        for k in ("AI_TRAIN_MAX_CONCURRENT", "AI_TRAIN_THREADS", "AI_TRAIN_MAX_RAM_MB", "BATCH_ROWS"):
            os.environ.pop(k, None)

        p2 = get_resource_profile(override_vcpu=2, override_ram_mb=8192)
        self.assertEqual(p2["max_concurrent"], 1)
        self.assertEqual(p2["train_threads"], 1)
        self.assertGreaterEqual(p2["max_ram_mb"], 2500)
        self.assertEqual(p2["auto_interval"], 60)

        p8 = get_resource_profile(override_vcpu=8, override_ram_mb=32768)
        self.assertEqual(p8["max_concurrent"], 3)
        self.assertGreaterEqual(p8["train_threads"], 2)
        self.assertGreater(p8["batch_rows"], p2["batch_rows"])

        p16 = get_resource_profile(override_vcpu=16, override_ram_mb=65536)
        self.assertEqual(p16["max_concurrent"], 7)

    def test_tier_selection_t1_t2_t3(self):
        p2 = get_resource_profile(override_vcpu=2, override_ram_mb=8192)
        t1 = select_training_tier(50_000, est_bytes_per_row=256, profile=p2)
        self.assertEqual(t1["tier"], "T1")

        t2 = select_training_tier(1_000_000, est_bytes_per_row=256, profile=p2)
        self.assertEqual(t2["tier"], "T2")

        t3 = select_training_tier(80_000_000, est_bytes_per_row=256, profile=p2)
        self.assertEqual(t3["tier"], "T3")
        self.assertLessEqual(t3["target_rows"], p2["t3_sample_rows"])

    def test_stratified_reservoir_preserves_minority_anomalies(self):
        # 10,000 rows with 150 anomalies ("Ya (Anomali)"): sample down to 1,000 rows
        df = pd.DataFrame({
            "id": range(10000),
            "TARGET: is_anomaly": ["Ya (Anomali)" if i < 150 else "Tidak" for i in range(10000)],
        })
        sampled = stratified_reservoir_sample_batches(
            iter_dataframe_batches(df, batch_rows=2000),
            target_rows=1000,
            label_col="TARGET: is_anomaly",
            minority_values=["Ya (Anomali)"],
        )
        self.assertLessEqual(len(sampled), 1000)
        anom_kept = (sampled["TARGET: is_anomaly"] == "Ya (Anomali)").sum()
        self.assertEqual(anom_kept, 150, "All 150 minority anomaly rows must be preserved!")


class ValidationRollbackTests(unittest.TestCase):
    def test_rollback_when_metric_degrades_and_promote_when_improved(self):
        tmp_dir = tempfile.mkdtemp()
        try:
            sfile = os.path.join(tmp_dir, "retrain_status.json")
            prod_file = os.path.join(tmp_dir, "anomaly_bundle.pkl")
            with open(prod_file, "wb") as f:
                f.write(b"v1-good-model")

            write_raw_registry(
                {
                    "models": {
                        "anomaly": {
                            "state": "trained",
                            "metrics": {"primary_name": "pr_auc", "primary": 0.90, "pr_auc": 0.90},
                            "version": 1,
                        }
                    }
                },
                status_file=sfile,
            )

            # 1. Stage a degraded model (pr_auc=0.75 < 0.90 - 0.03) -> must rollback!
            staged_bad = os.path.join(tmp_dir, "anomaly_bundle.pkl.staging")
            with open(staged_bad, "wb") as f:
                f.write(b"v2-bad-model")

            res_bad = validate_and_promote_artifacts(
                "anomaly",
                {"anomaly_bundle.pkl": staged_bad},
                {"primary_name": "pr_auc", "primary": 0.75, "pr_auc": 0.75},
                model_dir=tmp_dir,
                status_file=sfile,
            )
            self.assertFalse(res_bad["promoted"])
            self.assertIn("Rollback otomatis", res_bad["reason"])
            with open(prod_file, "rb") as f:
                self.assertEqual(f.read(), b"v1-good-model")

            # 2. Stage an improved model (pr_auc=0.93 >= 0.90) -> must promote & backup v1!
            staged_good = os.path.join(tmp_dir, "anomaly_bundle.pkl.staging")
            with open(staged_good, "wb") as f:
                f.write(b"v3-better-model")

            res_good = validate_and_promote_artifacts(
                "anomaly",
                {"anomaly_bundle.pkl": staged_good},
                {"primary_name": "pr_auc", "primary": 0.93, "pr_auc": 0.93},
                model_dir=tmp_dir,
                status_file=sfile,
            )
            self.assertTrue(res_good["promoted"])
            with open(prod_file, "rb") as f:
                self.assertEqual(f.read(), b"v3-better-model")
            self.assertTrue(os.path.exists(prod_file + ".v1.bak"))
        finally:
            shutil.rmtree(tmp_dir, ignore_errors=True)


class MultiLaneSchedulerTests(unittest.TestCase):
    def test_fifo_across_lanes_coalescing_manual_priority_and_crash_recovery(self):
        tmp_dir = tempfile.mkdtemp()
        try:
            sched = RetrainScheduler(
                model_dir=tmp_dir,
                auto_start_worker=False,
                auto_start_detector=False,
            )
            # 1. Enqueue auto jobs: Anomaly arrives first (t=100), Entity second (t=105), NLP third (t=110)
            r1 = sched.enqueue_model("anomaly", trigger="auto_detector", priority="auto", force=True, enqueued_epoch=100.0)
            r2 = sched.enqueue_model("entity", trigger="auto_detector", priority="auto", force=True, enqueued_epoch=105.0)
            r3 = sched.enqueue_model("nlp", trigger="auto_detector", priority="auto", force=True, enqueued_epoch=110.0)

            self.assertEqual(r1["status"], "queued")
            self.assertEqual(r1["queue_position"], 1)
            self.assertEqual(r2["queue_position"], 2)
            self.assertEqual(r3["queue_position"], 3)

            # 2. Coalesce: second anomaly table mapped while anomaly is still queued -> same job_id, no duplicate!
            r1_dup = sched.enqueue_model("anomaly", trigger="auto_detector", priority="auto", force=True, enqueued_epoch=112.0)
            self.assertEqual(r1_dup["status"], "already_queued")
            self.assertEqual(r1_dup["job_id"], r1["job_id"])
            self.assertEqual(len(sched.get_queue_snapshot()), 3)

            # 3. Manual priority: user manually triggers NLP -> NLP moves to position #1 ahead of auto jobs
            r3_manual = sched.enqueue_model("nlp", trigger="manual_nlp", priority="manual", force=True)
            self.assertEqual(r3_manual["status"], "already_queued")
            q = sched.get_queue_snapshot()
            self.assertEqual(q[0]["model"], "nlp")
            self.assertEqual(q[0]["priority"], "manual")
            self.assertEqual(q[1]["model"], "anomaly")
            self.assertEqual(q[2]["model"], "entity")

            # 4. Crash recovery: simulate an interrupted 'running' job persisted on disk
            sched._jobs[0]["state"] = "running"
            sched._recompute_positions_and_persist()

            sched_restarted = RetrainScheduler(
                model_dir=tmp_dir,
                auto_start_worker=False,
                auto_start_detector=False,
            )
            q_rec = sched_restarted.get_queue_snapshot()
            self.assertEqual(len(q_rec), 3)
            self.assertEqual(q_rec[0]["state"], "queued")
            self.assertEqual(q_rec[0]["model"], "nlp")
        finally:
            shutil.rmtree(tmp_dir, ignore_errors=True)

    def test_single_model_retrain_isolates_other_model_files(self):
        """Retraining `nlp` or `anomaly` must NOT touch the mtime of other models' artifacts."""
        engine = AutoRetrainEngine()
        dept_path = os.path.join(engine.model_dir, "department_bundle.pkl")
        kpi_path = os.path.join(engine.model_dir, "kpi_bundle.pkl")
        anom_path = os.path.join(engine.model_dir, "anomaly_bundle.pkl")

        dept_mtime_before = os.path.getmtime(dept_path) if os.path.exists(dept_path) else 0
        kpi_mtime_before = os.path.getmtime(kpi_path) if os.path.exists(kpi_path) else 0

        time.sleep(0.05)
        res = engine.retrain_single_model("anomaly", trigger="unit_test", sync_sources=False)
        self.assertTrue(res.get("ok"), f"Anomaly retrain failed: {res}")

        if dept_mtime_before > 0:
            self.assertEqual(os.path.getmtime(dept_path), dept_mtime_before, "department_bundle.pkl must not be modified when retraining anomaly!")
        if kpi_mtime_before > 0:
            self.assertEqual(os.path.getmtime(kpi_path), kpi_mtime_before, "kpi_bundle.pkl must not be modified when retraining anomaly!")

        # Also verify anomaly_bundle.pkl was compacted (< 5 MB instead of 34 MB)
        anom_size_mb = os.path.getsize(anom_path) / (1024 * 1024)
        self.assertLess(anom_size_mb, 6.0, f"Expected compacted anomaly_bundle.pkl < 6 MB, got {anom_size_mb:.2f} MB")


if __name__ == "__main__":
    unittest.main()
