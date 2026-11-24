"""
AuditSphere — Multi-Lane AI Retrain Queue, Resource-Guarded Worker & 60s Change Detector.

Implements:
1. Per-model lanes (`entity`, `anomaly`, `nlp`, `kpi_auditee`, `kpi_auditor`) with in-lane coalescing
   (multiple table mappings into the same feature merge into a single queued retrain) and
   at-most-one follow-up job when new data arrives while that model is actively training.
2. Cross-lane scheduling ordered by earliest arrival (`enqueued_at` FIFO across lanes — whichever
   feature receives new data first is trained first), with `manual` priority ahead of `auto`
   without preempting an actively running job.
3. Adaptive concurrency (`AI_TRAIN_MAX_CONCURRENT`, default 1 on 2 vCPU) and RAM/CPU Resource Guard.
4. Persistent queue (`models/retrain_queue.json`) that recovers interrupted `running` jobs as `queued`
   on startup.
5. Lightweight 60-second background detector (`RETRAIN_AUTO_INTERVAL=60`) with per-model debounce
   (`RETRAIN_DEBOUNCE_SECONDS=60`) and adaptive HTTP sync backoff.
"""
import os
import uuid
import time
import json
import threading
from datetime import datetime, timezone
from typing import Dict, Any, List, Optional

from resource_profile import get_resource_profile
from training_data_source import compute_all_model_fingerprints
from model_registry import (
    MODEL_KEYS,
    MODEL_METADATA,
    resolve_model_keys,
    read_raw_registry,
    write_raw_registry,
    update_model_entry,
    build_live_status_snapshot,
)
from retrain_engine import (
    AutoRetrainEngine,
    reload_fastapi_models,
    compute_auditor_items_hash,
    compute_auditee_custom_items_hash,
)

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.getenv("MODEL_DIR", os.path.join(BASE_DIR, "models"))
QUEUE_FILE = os.path.join(MODEL_DIR, "retrain_queue.json")
os.makedirs(MODEL_DIR, exist_ok=True)


class RetrainScheduler:
    def __init__(
        self,
        model_dir: Optional[str] = None,
        ai_training_dir: Optional[str] = None,
        fastapi_app: Any = None,
        auto_start_worker: bool = True,
        auto_start_detector: bool = True,
    ):
        self.model_dir = model_dir or MODEL_DIR
        self.ai_training_dir = ai_training_dir
        self.status_file = os.path.join(self.model_dir, "retrain_status.json")
        self.queue_file = os.path.join(self.model_dir, "retrain_queue.json")
        self.fastapi_app = fastapi_app

        self._lock = threading.RLock()
        self._wakeup = threading.Event()
        self._stop_event = threading.Event()

        self._jobs: List[Dict[str, Any]] = []
        self._active_workers: Dict[str, threading.Thread] = {}
        # Tracks first_seen timestamp of a new hash per model for debounce: {model: {"hash": str, "first_seen": float}}
        self._pending_changes: Dict[str, Dict[str, Any]] = {}
        self.last_detect_at: Optional[str] = None
        self._last_cross_sync_ts: float = 0.0
        self._cross_sync_backoff_sec: float = 60.0

        self._load_persisted_queue()

        self._dispatcher_thread: Optional[threading.Thread] = None
        self._detector_thread: Optional[threading.Thread] = None
        if auto_start_worker:
            self.start_worker()
        if auto_start_detector:
            self.start_detector()

    def set_fastapi_app(self, app: Any):
        self.fastapi_app = app

    # ------------------------------------------------------------------
    # Queue Persistence & Crash Recovery
    # ------------------------------------------------------------------
    def _load_persisted_queue(self):
        with self._lock:
            os.makedirs(self.model_dir, exist_ok=True)
            loaded: List[Dict[str, Any]] = []
            if os.path.exists(self.queue_file):
                try:
                    with open(self.queue_file, "r", encoding="utf-8") as f:
                        data = json.load(f)
                    if isinstance(data, list):
                        loaded = data
                except Exception:
                    loaded = []

            # Recover any interrupted 'running' jobs back to 'queued' at the front
            recovered: List[Dict[str, Any]] = []
            seen_queued_models = set()
            for job in loaded:
                if not isinstance(job, dict):
                    continue
                mk = job.get("model")
                if mk not in MODEL_KEYS:
                    continue
                st = job.get("state")
                if st in ("running", "queued"):
                    if mk in seen_queued_models:
                        continue
                    seen_queued_models.add(mk)
                    job["state"] = "queued"
                    job["recovered"] = True if st == "running" else False
                    recovered.append(job)

            self._jobs = recovered
            self._recompute_positions_and_persist()

    def _recompute_positions_and_persist(self):
        """Sort waiting jobs by (priority manual > auto, enqueued_epoch ASC) and assign 1-based queue positions."""
        with self._lock:
            running = [j for j in self._jobs if j.get("state") == "running"]
            waiting = [j for j in self._jobs if j.get("state") == "queued"]
            waiting.sort(
                key=lambda j: (
                    0 if j.get("priority") == "manual" else 1,
                    float(j.get("enqueued_epoch") or 0.0),
                )
            )
            for idx, j in enumerate(waiting, start=1):
                j["position"] = idx
            for j in running:
                j["position"] = 0
            self._jobs = running + waiting

            try:
                tmp = self.queue_file + ".tmp"
                with open(tmp, "w", encoding="utf-8") as f:
                    # Exclude bulky in-memory item lists if huge when persisting queue metadata
                    slim = []
                    for j in self._jobs:
                        copy_j = {k: v for k, v in j.items() if k not in ("auditor_items", "auditee_items")}
                        slim.append(copy_j)
                    json.dump(slim, f, ensure_ascii=False, indent=2)
                os.replace(tmp, self.queue_file)
            except Exception as e:
                print(f"[RetrainScheduler] Warning persisting queue: {e}")

            # Mirror queue summary into registry file
            write_raw_registry({"queue": self.get_queue_snapshot()}, status_file=self.status_file)

    def get_queue_snapshot(self) -> List[Dict[str, Any]]:
        with self._lock:
            return [
                {
                    "job_id": j.get("job_id"),
                    "model": j.get("model"),
                    "trigger": j.get("trigger"),
                    "priority": j.get("priority", "auto"),
                    "state": j.get("state"),
                    "enqueued_at": j.get("enqueued_at"),
                    "enqueued_epoch": j.get("enqueued_epoch"),
                    "position": j.get("position"),
                    "coalesced_count": j.get("coalesced_count", 1),
                }
                for j in self._jobs
            ]

    # ------------------------------------------------------------------
    # Enqueue / Coalesce / Follow-up Logic
    # ------------------------------------------------------------------
    def enqueue_model(
        self,
        model_key: str,
        trigger: str = "manual",
        priority: Optional[str] = None,
        force: bool = False,
        auditor_items: Optional[List[Dict[str, Any]]] = None,
        auditee_items: Optional[List[Dict[str, Any]]] = None,
        enqueued_epoch: Optional[float] = None,
    ) -> Dict[str, Any]:
        """
        Enqueue a single model (`entity`, `anomaly`, `nlp`, `kpi_auditee`, `kpi_auditor`) into its lane.
        - Ingests `auditor_items` or `auditee_items` immediately into the storage pool so fingerprints update.
        - If a `queued` job already exists for this model, coalesces into it (upgrades priority to `manual` if needed).
        - If a `running` job is currently training this model, creates/updates a single follow-up `queued` job.
        - If `force=False` and the model is already `trained` with no new data, returns `status="up_to_date"`.
        """
        if model_key not in MODEL_KEYS:
            raise ValueError(f"Unknown model key: {model_key}")

        eng = AutoRetrainEngine(model_dir=self.model_dir, ai_training_dir=self.ai_training_dir)
        auditor_hash = None
        auditee_custom_hash = None
        if model_key == "kpi_auditor" and auditor_items:
            auditor_hash = eng._ingest_live_auditor_items(auditor_items)
        elif model_key == "kpi_auditee" and auditee_items:
            auditee_custom_hash = eng._ingest_custom_auditee_items(auditee_items)

        prio = priority or ("manual" if "manual" in str(trigger).lower() or "force" in str(trigger).lower() or "rule" in str(trigger).lower() or "sync" in str(trigger).lower() else "auto")

        with self._lock:
            existing_running = next((j for j in self._jobs if j.get("model") == model_key and j.get("state") == "running"), None)
            existing_queued = next((j for j in self._jobs if j.get("model") == model_key and j.get("state") == "queued"), None)

            # Case 1: Already queued in this lane -> Coalesce!
            if existing_queued is not None:
                if prio == "manual" and existing_queued.get("priority") != "manual":
                    existing_queued["priority"] = "manual"
                    existing_queued["trigger"] = trigger
                existing_queued["coalesced_count"] = int(existing_queued.get("coalesced_count", 1)) + 1
                if auditor_items:
                    existing_queued["auditor_items"] = auditor_items
                if auditee_items:
                    existing_queued["auditee_items"] = auditee_items
                self._recompute_positions_and_persist()
                self._wakeup.set()
                status_kind = "followup_queued" if existing_running else "already_queued"
                label = MODEL_METADATA[model_key]["label"]
                return {
                    "status": status_kind,
                    "model": model_key,
                    "job_id": existing_queued["job_id"],
                    "queue_position": existing_queued.get("position", 1),
                    "message": f"Model {label} sudah berada dalam antrian pelatihan (posisi #{existing_queued.get('position', 1)}).",
                }

            # Check whether model actually has new data when force=False and not currently running
            if not force and existing_running is None:
                snap = build_live_status_snapshot(
                    auditor_hash=auditor_hash,
                    auditee_custom_hash=auditee_custom_hash,
                    queue_snapshot=self.get_queue_snapshot(),
                    status_file=self.status_file,
                )
                m_state = (snap.get("models") or {}).get(model_key, {})
                if m_state.get("state") == "trained" and not m_state.get("has_new_data"):
                    label = MODEL_METADATA[model_key]["label"]
                    return {
                        "status": "up_to_date",
                        "model": model_key,
                        "job_id": None,
                        "queue_position": None,
                        "message": f"Model {label} sudah up-to-date (tidak ada data baru yang belum dilatih).",
                    }

            now_ep = float(enqueued_epoch if enqueued_epoch is not None else time.time())
            now_iso = datetime.fromtimestamp(now_ep, tz=timezone.utc).isoformat()
            job_id = f"job-{model_key}-{uuid.uuid4().hex[:8]}"
            new_job = {
                "job_id": job_id,
                "model": model_key,
                "trigger": trigger,
                "priority": prio,
                "state": "queued",
                "enqueued_at": now_iso,
                "enqueued_epoch": now_ep,
                "position": 1,
                "coalesced_count": 1,
                "auditor_items": auditor_items,
                "auditee_items": auditee_items,
            }
            self._jobs.append(new_job)
            self._pending_changes.pop(model_key, None)
            self._recompute_positions_and_persist()

            if existing_running is None:
                update_model_entry(
                    model_key,
                    {
                        "state": "queued",
                        "job_id": job_id,
                        "trigger": trigger,
                        "queue_position": new_job["position"],
                        "last_error": None,
                        "message": f"Model {MODEL_METADATA[model_key]['label']} masuk antrian pelatihan (#{new_job['position']}).",
                    },
                    status_file=self.status_file,
                )

            self._wakeup.set()
            status_kind = "followup_queued" if existing_running is not None else "queued"
            label = MODEL_METADATA[model_key]["label"]
            return {
                "status": status_kind,
                "model": model_key,
                "job_id": job_id,
                "queue_position": new_job["position"],
                "message": (
                    f"Model {label} dijadwalkan untuk pelatihan lanjutan setelah proses saat ini selesai (posisi #{new_job['position']})."
                    if existing_running
                    else f"Model {label} masuk antrian pelatihan (posisi #{new_job['position']})."
                ),
            }

    def enqueue_scope(
        self,
        scope: str = "all",
        trigger: str = "auto_background",
        priority: Optional[str] = None,
        force: bool = False,
        auditor_items: Optional[List[Dict[str, Any]]] = None,
        auditee_items: Optional[List[Dict[str, Any]]] = None,
    ) -> Dict[str, Any]:
        """Enqueue one or multiple models resolved from `scope` (`entity`, `anomaly`, `nlp`, `kpi`, `all`)."""
        keys = resolve_model_keys(scope)
        if not keys:
            raise ValueError(f"Unknown model or scope: {scope}")

        results = []
        enqueued_models = []
        first_job_id = None
        first_pos = None

        for mk in keys:
            res = self.enqueue_model(
                model_key=mk,
                trigger=trigger,
                priority=priority,
                force=force,
                auditor_items=auditor_items if mk == "kpi_auditor" else None,
                auditee_items=auditee_items if mk == "kpi_auditee" else None,
            )
            results.append(res)
            if res["status"] in ("queued", "already_queued", "followup_queued"):
                enqueued_models.append(mk)
                if first_job_id is None:
                    first_job_id = res.get("job_id")
                    first_pos = res.get("queue_position")

        snap = self.get_status()
        if not enqueued_models:
            return {
                "status": "up_to_date",
                "is_running": bool(snap.get("is_running")),
                "model": scope,
                "models_enqueued": [],
                "job_id": None,
                "queue_position": None,
                "message": "Semua model pada cakupan ini sudah dilatih dengan data terbaru.",
                "results": results,
                "models": snap.get("models"),
                "queue": snap.get("queue"),
            }

        return {
            "status": results[0]["status"] if len(results) == 1 else "scheduled",
            "is_running": True,
            "model": scope,
            "models_enqueued": enqueued_models,
            "job_id": first_job_id,
            "queue_position": first_pos,
            "message": results[0]["message"] if len(results) == 1 else f"{len(enqueued_models)} model masuk antrian pelatihan: {', '.join(enqueued_models)}.",
            "results": results,
            "models": snap.get("models"),
            "queue": snap.get("queue"),
        }

    # ------------------------------------------------------------------
    # Worker Loop & Job Execution
    # ------------------------------------------------------------------
    def start_worker(self):
        with self._lock:
            if self._dispatcher_thread and self._dispatcher_thread.is_alive():
                return
            self._stop_event.clear()
            self._dispatcher_thread = threading.Thread(
                target=self._dispatcher_loop,
                name="AuditSphere-RetrainDispatcher",
                daemon=True,
            )
            self._dispatcher_thread.start()

    def start_detector(self):
        with self._lock:
            if self._detector_thread and self._detector_thread.is_alive():
                return
            self._detector_thread = threading.Thread(
                target=self._detector_loop,
                name="AuditSphere-RetrainDetector",
                daemon=True,
            )
            self._detector_thread.start()

    def stop(self, timeout: float = 3.0):
        self._stop_event.set()
        self._wakeup.set()
        if self._dispatcher_thread and self._dispatcher_thread.is_alive():
            self._dispatcher_thread.join(timeout=timeout)
        if self._detector_thread and self._detector_thread.is_alive():
            self._detector_thread.join(timeout=timeout)

    def _dispatcher_loop(self):
        while not self._stop_event.is_set():
            dispatched = self.step_once()
            if not dispatched:
                self._wakeup.wait(timeout=0.5)
                self._wakeup.clear()

    def step_once(self, run_synchronously: bool = False) -> bool:
        """
        Attempt to dispatch the next eligible queued job if concurrency and resource guard allow.
        Returns True if a job was started (or completed when `run_synchronously=True`).
        """
        prof = get_resource_profile()
        max_conc = int(prof.get("max_concurrent", 1))
        guard_ok = bool(prof.get("guard_ok", True))

        job_to_run = None
        with self._lock:
            # Clean up finished worker threads
            dead = [jid for jid, th in self._active_workers.items() if not th.is_alive()]
            for jid in dead:
                self._active_workers.pop(jid, None)

            running_jobs = [j for j in self._jobs if j.get("state") == "running"]
            if len(running_jobs) >= max_conc:
                return False

            running_models = {j.get("model") for j in running_jobs}
            waiting_jobs = [j for j in self._jobs if j.get("state") == "queued" and j.get("model") not in running_models]
            if not waiting_jobs:
                return False

            # Respect resource guard: if guard is not OK, defer auto jobs (and defer manual if RAM is critically low)
            next_candidate = waiting_jobs[0]
            if not guard_ok:
                if next_candidate.get("priority") == "auto":
                    return False
                if prof.get("ram_available_mb", 2048) < max(256, prof.get("min_free_ram_mb", 1500) // 2):
                    return False

            next_candidate["state"] = "running"
            next_candidate["started_at"] = datetime.now(timezone.utc).isoformat()
            job_to_run = next_candidate
            self._recompute_positions_and_persist()

        if job_to_run is None:
            return False

        if run_synchronously:
            self._execute_job(job_to_run)
        else:
            th = threading.Thread(
                target=self._execute_job,
                args=(job_to_run,),
                name=f"RetrainWorker-{job_to_run['model']}",
                daemon=True,
            )
            with self._lock:
                self._active_workers[job_to_run["job_id"]] = th
            th.start()
        return True

    def drain_queue_sync(self, max_jobs: int = 20):
        """Synchronously process all queued jobs in order (useful for deterministic testing & CLI)."""
        count = 0
        while count < max_jobs:
            with self._lock:
                has_waiting = any(j.get("state") == "queued" for j in self._jobs)
                has_running = any(j.get("state") == "running" for j in self._jobs)
            if not has_waiting and not has_running:
                break
            if has_running:
                time.sleep(0.05)
                continue
            ran = self.step_once(run_synchronously=True)
            if not ran:
                break
            count += 1

    def _execute_job(self, job: Dict[str, Any]):
        model_key = job["model"]
        job_id = job["job_id"]
        trigger = job.get("trigger", "auto")
        started_iso = job.get("started_at") or datetime.now(timezone.utc).isoformat()
        label = MODEL_METADATA[model_key]["label"]

        # Mark model as training in registry
        update_model_entry(
            model_key,
            {
                "state": "training",
                "job_id": job_id,
                "trigger": trigger,
                "queue_position": 0,
                "waiting_reason": None,
                "last_error": None,
                "progress": {"phase": "starting", "batch": 1, "total_batches": 2, "percent": 5, "tier": "T1"},
                "message": f"Sedang melatih {label} ({trigger})...",
            },
            top_patch={
                "status": "running",
                "is_running": True,
                "trigger": trigger,
                "started_at": started_iso,
                "last_error": None,
                "message": f"Sedang melatih model {label} ({trigger})...",
            },
            status_file=self.status_file,
        )

        def _on_progress(prog: Dict[str, Any]):
            update_model_entry(
                model_key,
                {
                    "state": "training",
                    "progress": prog,
                    "tier": prog.get("tier", "T1"),
                    "message": f"Sedang melatih {label} (Batch {prog.get('batch', 1)}/{prog.get('total_batches', 1)} — {prog.get('percent', 0)}%).",
                },
                status_file=self.status_file,
            )

        try:
            engine = AutoRetrainEngine(model_dir=self.model_dir, ai_training_dir=self.ai_training_dir)
            sync_sources = "manual" not in str(trigger).lower() or model_key in ("entity", "anomaly", "nlp", "kpi_auditor", "kpi_auditee")
            res = engine.retrain_single_model(
                model_key=model_key,
                trigger=trigger,
                auditor_items=job.get("auditor_items"),
                auditee_items=job.get("auditee_items"),
                sync_sources=sync_sources,
                progress_cb=_on_progress,
            )

            finished_iso = datetime.now(timezone.utc).isoformat()
            ok = bool(res.get("ok"))
            promoted = bool(res.get("promoted", ok))
            fp = res.get("fingerprint") or compute_all_model_fingerprints().get(model_key, {})
            new_hash = str(fp.get("hash") or "")
            new_rows = int(fp.get("rows") or res.get("samples") or 0)
            dur = float(res.get("duration_seconds") or 0.0)
            metrics = res.get("metrics") or {}
            tier = res.get("tier") or "T1"
            peak_ram = res.get("peak_ram_mb")

            reg_curr = read_raw_registry(status_file=self.status_file)
            prev_ver = int((reg_curr.get("models", {}).get(model_key) or {}).get("version") or 0)

            if ok and promoted:
                # Selectively hot-reload ONLY the newly trained model bundle
                reload_fastapi_models(self.fastapi_app, only_models=[model_key])
                m_patch = {
                    "state": "trained",
                    "has_new_data": False,
                    "new_rows": 0,
                    "trained_data_hash": new_hash,
                    "current_data_hash": new_hash,
                    "trained_rows": new_rows,
                    "current_rows": new_rows,
                    "trained_at": finished_iso,
                    "duration_seconds": dur,
                    "samples": int(res.get("samples") or new_rows),
                    "tier": tier,
                    "metrics": metrics,
                    "last_result": "promoted",
                    "progress": None,
                    "queue_position": None,
                    "waiting_reason": None,
                    "peak_ram_mb": peak_ram,
                    "version": prev_ver + 1,
                    "last_error": None,
                    "message": (
                        f"Model {label} selesai dilatih dalam {dur} dtk "
                        f"({int(res.get('samples') or new_rows)} sampel, {metrics.get('primary_name','score')}={metrics.get('primary','-')})."
                    ),
                }
                top_patch: Dict[str, Any] = {
                    "finished_at": finished_iso,
                    "duration_seconds": dur,
                    "model_updated_at": finished_iso,
                    "message": m_patch["message"],
                    "last_error": None,
                }
                if model_key == "kpi_auditee":
                    top_patch["auditee_retrained"] = True
                    top_patch["trained_auditee_pool_hash"] = str(res.get("auditee_pool_hash") or fp.get("pool_hash") or new_hash)
                    if res.get("auditee_custom_hash"):
                        top_patch["trained_auditee_custom_hash"] = str(res.get("auditee_custom_hash"))
                    top_patch["trained_auditee_count"] = new_rows
                elif model_key == "kpi_auditor":
                    top_patch["auditor_retrained"] = True
                    top_patch["trained_auditor_hash"] = str(res.get("auditor_hash") or new_hash)
                    top_patch["trained_auditor_count"] = int(res.get("kpis") or new_rows)
                update_model_entry(model_key, m_patch, top_patch=top_patch, status_file=self.status_file)
            else:
                err_reason = str(res.get("reason") or res.get("error") or f"Pelatihan {label} tidak memenuhi ambang validasi.")
                m_patch = {
                    "state": "failed",
                    "progress": None,
                    "queue_position": None,
                    "duration_seconds": dur,
                    "last_result": "rolled_back" if res.get("reason") else "failed",
                    "last_error": err_reason,
                    "message": err_reason,
                }
                update_model_entry(
                    model_key,
                    m_patch,
                    top_patch={
                        "status": "failed",
                        "finished_at": finished_iso,
                        "last_error": err_reason,
                        "message": err_reason,
                    },
                    status_file=self.status_file,
                )
        except Exception as e:
            finished_iso = datetime.now(timezone.utc).isoformat()
            err_msg = f"Pelatihan {label} gagal: {e}"
            update_model_entry(
                model_key,
                {
                    "state": "failed",
                    "progress": None,
                    "queue_position": None,
                    "last_result": "failed",
                    "last_error": str(e),
                    "message": err_msg,
                },
                top_patch={
                    "status": "failed",
                    "finished_at": finished_iso,
                    "last_error": str(e),
                    "message": err_msg,
                },
                status_file=self.status_file,
            )
        finally:
            with self._lock:
                self._jobs = [j for j in self._jobs if j.get("job_id") != job_id]
                self._recompute_positions_and_persist()
                any_still_running = any(j.get("state") == "running" for j in self._jobs)
                any_still_queued = any(j.get("state") == "queued" for j in self._jobs)
                if not any_still_running and not any_still_queued:
                    reg_now = read_raw_registry(status_file=self.status_file)
                    write_raw_registry(
                        {
                            "is_running": False,
                            "status": "failed" if reg_now.get("last_error") else "completed",
                        },
                        status_file=self.status_file,
                    )
            self._wakeup.set()

    # ------------------------------------------------------------------
    # 60-Second Background Auto-Detector with Debounce & Adaptive Backoff
    # ------------------------------------------------------------------
    def detect_and_enqueue_changes_once(self, now_ts: Optional[float] = None, sync_remote: bool = False) -> List[str]:
        """
        Inspect all 5 model fingerprints and enqueue any model that has new data
        which has remained stable for `debounce_seconds`, ordered by `first_seen` timestamp!
        """
        now = float(now_ts if now_ts is not None else time.time())
        prof = get_resource_profile()
        debounce = float(prof.get("debounce_seconds", 60))
        self.last_detect_at = datetime.fromtimestamp(now, tz=timezone.utc).isoformat()

        # Optional lightweight cross-service sync with adaptive backoff
        if sync_remote and (now - self._last_cross_sync_ts) >= self._cross_sync_backoff_sec:
            t_sync = time.perf_counter()
            try:
                eng = AutoRetrainEngine(model_dir=self.model_dir, ai_training_dir=self.ai_training_dir)
                eng.sync_live_cross_service_data(scope="all")
            except Exception:
                pass
            sync_elapsed = time.perf_counter() - t_sync
            self._last_cross_sync_ts = now
            # Adaptive backoff: if cross-service HTTP calls were slow (>2.5s), widen to 300s (5 min), else keep 60s
            self._cross_sync_backoff_sec = 300.0 if sync_elapsed > 2.5 else float(prof.get("auto_interval", 60))

        snap = build_live_status_snapshot(
            queue_snapshot=self.get_queue_snapshot(),
            status_file=self.status_file,
        )
        models_map = snap.get("models") or {}

        ready_to_enqueue: List[tuple] = []
        with self._lock:
            for mk in MODEL_KEYS:
                m_info = models_map.get(mk) or {}
                st = m_info.get("state")
                cur_hash = str(m_info.get("current_data_hash") or "")
                has_new = bool(m_info.get("has_new_data")) or (st == "never_trained" and int(m_info.get("current_rows") or 0) > 0)

                if not has_new or st in ("queued", "training"):
                    self._pending_changes.pop(mk, None)
                    continue

                prev_pend = self._pending_changes.get(mk)
                if prev_pend is None:
                    self._pending_changes[mk] = {
                        "hash": cur_hash,
                        "first_seen": now,
                        "last_changed": now,
                    }
                elif prev_pend.get("hash") != cur_hash:
                    # Data still actively changing -> keep original first_seen arrival order, reset stability timer
                    prev_pend["hash"] = cur_hash
                    prev_pend["last_changed"] = now

                pend = self._pending_changes[mk]
                if (now - float(pend["last_changed"])) >= debounce:
                    ready_to_enqueue.append((float(pend["first_seen"]), mk))

        # Sort by earliest first_seen so whichever model received new data first is enqueued & trained first
        ready_to_enqueue.sort(key=lambda pair: pair[0])
        enqueued: List[str] = []
        for first_seen_ep, mk in ready_to_enqueue:
            res = self.enqueue_model(
                model_key=mk,
                trigger="auto_detector",
                priority="auto",
                force=True,
                enqueued_epoch=first_seen_ep,
            )
            if res.get("status") in ("queued", "already_queued", "followup_queued"):
                enqueued.append(mk)
        return enqueued

    def _detector_loop(self):
        # Initial sleep so app startup finishes cleanly
        self._stop_event.wait(timeout=5.0)
        while not self._stop_event.is_set():
            prof = get_resource_profile()
            interval = max(10, int(prof.get("auto_interval", 60)))
            try:
                if prof.get("guard_ok", True):
                    self.detect_and_enqueue_changes_once(sync_remote=True)
            except Exception as e:
                print(f"[RetrainScheduler] Auto-detector warning: {e}")
            self._stop_event.wait(timeout=float(interval))

    def get_status(
        self,
        auditor_hash: Optional[str] = None,
        auditee_custom_hash: Optional[str] = None,
    ) -> Dict[str, Any]:
        return build_live_status_snapshot(
            auditor_hash=auditor_hash,
            auditee_custom_hash=auditee_custom_hash,
            queue_snapshot=self.get_queue_snapshot(),
            scheduler_info={
                "running": bool(self._dispatcher_thread and self._dispatcher_thread.is_alive()),
                "auto_enabled": bool(self._detector_thread and self._detector_thread.is_alive()),
                "last_detect_at": self.last_detect_at,
                "pending_debounce_models": list(self._pending_changes.keys()),
            },
            status_file=self.status_file,
        )


_GLOBAL_SCHEDULER: Optional[RetrainScheduler] = None
_GLOBAL_SCHEDULER_LOCK = threading.Lock()


def get_retrain_scheduler(fastapi_app: Any = None) -> RetrainScheduler:
    global _GLOBAL_SCHEDULER
    with _GLOBAL_SCHEDULER_LOCK:
        if _GLOBAL_SCHEDULER is None:
            auto_det = os.getenv("RETRAIN_AUTO_ENABLED", "true").strip().lower() not in ("0", "false", "no")
            _GLOBAL_SCHEDULER = RetrainScheduler(
                fastapi_app=fastapi_app,
                auto_start_worker=True,
                auto_start_detector=auto_det,
            )
        elif fastapi_app is not None:
            _GLOBAL_SCHEDULER.set_fastapi_app(fastapi_app)
        return _GLOBAL_SCHEDULER
