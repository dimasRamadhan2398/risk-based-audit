"""
AuditSphere — Per-Model AI Registry, Status State Machine & Artifact Rollback Manager.

Tracks each AI model suite independently:
- `entity`       : Entity Prediction (Risk Scoring & Audit Universe)
- `anomaly`      : Anomaly Detection (XGBoost + IsolationForest + Bank Transaction & Data Source Rules)
- `nlp`          : NLP Document & Multi-Source Text Analysis
- `kpi_auditee`  : KPI Forecast — Auditee (Datalake / External DB KPI Series)
- `kpi_auditor`  : KPI Forecast — Auditor (Strategic Audit Plan & Internal Audit KPI)

States per model:
- `never_trained`  : Model artifact does not exist or has never been trained
- `trained`        : Model is trained and up-to-date with current data hash
- `untrained_data` : New or modified data detected (`current_data_hash != trained_data_hash`)
- `queued`         : Waiting in the scheduler queue (`queue_position >= 1`)
- `training`       : Actively training (`progress` shows batch x/y and %)
- `failed`         : Last training or validation rollback failed (`last_error` populated)

Maintains 100% backward compatibility with legacy top-level `retrain_status.json` fields.
"""
import os
import shutil
import json
import threading
from datetime import datetime, timezone
from typing import Dict, Any, List, Optional

from resource_profile import get_resource_profile
from training_data_source import compute_all_model_fingerprints

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.getenv("MODEL_DIR", os.path.join(BASE_DIR, "models"))
RETRAIN_STATUS_FILE = os.path.join(MODEL_DIR, "retrain_status.json")
os.makedirs(MODEL_DIR, exist_ok=True)

_REGISTRY_LOCK = threading.RLock()

MODEL_KEYS = ("entity", "anomaly", "nlp", "kpi_auditee", "kpi_auditor")

MODEL_METADATA: Dict[str, Dict[str, Any]] = {
    "entity": {
        "label": "Entity Prediction",
        "artifacts": ["department_bundle.pkl"],
        "primary_metric": "accuracy",
        "higher_is_better": True,
    },
    "anomaly": {
        "label": "Anomaly Detection",
        "artifacts": ["anomaly_bundle.pkl"],
        "primary_metric": "pr_auc",
        "higher_is_better": True,
    },
    "nlp": {
        "label": "NLP Analysis",
        "artifacts": ["document_bundle.pkl"],
        "primary_metric": "macro_f1",
        "higher_is_better": True,
    },
    "kpi_auditee": {
        "label": "KPI Forecast (Auditee)",
        "artifacts": ["kpi_bundle.pkl", "kpi_lstm.pth"],
        "primary_metric": "rmse",
        "higher_is_better": False,
    },
    "kpi_auditor": {
        "label": "KPI Forecast (Auditor)",
        "artifacts": ["auditor_kpi_bundle.pkl", "auditor_kpi_lstm.pth"],
        "primary_metric": "rmse",
        "higher_is_better": False,
    },
}


def resolve_model_keys(scope_or_model: str) -> List[str]:
    """Resolve a model key or scope alias (`kpi`, `all`) into canonical model keys."""
    s = str(scope_or_model or "").strip().lower()
    if s in ("all", "*"):
        return list(MODEL_KEYS)
    if s in ("kpi", "kpi_forecast", "performance-trend"):
        return ["kpi_auditee", "kpi_auditor"]
    if s in ("department", "risk-score", "entity-prediction"):
        return ["entity"]
    if s in ("document", "text-analysis"):
        return ["nlp"]
    if s in MODEL_KEYS:
        return [s]
    return []


def _artifact_exists(model_key: str, model_dir: Optional[str] = None) -> bool:
    mdir = model_dir or MODEL_DIR
    meta = MODEL_METADATA.get(model_key, {})
    arts = meta.get("artifacts", [])
    return bool(arts) and all(os.path.exists(os.path.join(mdir, a)) for a in arts)


def _artifact_mtime_iso(model_key: str, model_dir: Optional[str] = None) -> Optional[str]:
    mdir = model_dir or MODEL_DIR
    meta = MODEL_METADATA.get(model_key, {})
    mtimes = []
    for a in meta.get("artifacts", []):
        p = os.path.join(mdir, a)
        if os.path.exists(p):
            try:
                mtimes.append(os.path.getmtime(p))
            except Exception:
                pass
    if not mtimes:
        return None
    return datetime.fromtimestamp(max(mtimes), tz=timezone.utc).isoformat()


def default_model_entry(model_key: str, model_dir: Optional[str] = None) -> Dict[str, Any]:
    meta = MODEL_METADATA.get(model_key, {"label": model_key, "primary_metric": "score"})
    exists = _artifact_exists(model_key, model_dir=model_dir)
    mtime_iso = _artifact_mtime_iso(model_key, model_dir=model_dir) if exists else None
    return {
        "key": model_key,
        "label": meta["label"],
        "state": "trained" if exists else "never_trained",
        "has_new_data": False,
        "new_rows": 0,
        "current_rows": 0,
        "trained_rows": 0,
        "trained_data_hash": "",
        "current_data_hash": "",
        "trained_at": mtime_iso,
        "duration_seconds": None,
        "samples": 0,
        "tier": "T1" if exists else None,
        "metrics": None,
        "last_result": "promoted" if exists else None,
        "message": f"Model {meta['label']} siap digunakan." if exists else f"Model {meta['label']} belum pernah dilatih.",
        "last_error": None,
        "progress": None,
        "queue_position": None,
        "job_id": None,
        "trigger": None,
        "waiting_reason": None,
        "peak_ram_mb": None,
        "version": 1 if exists else 0,
    }


def default_registry_payload(model_dir: Optional[str] = None) -> Dict[str, Any]:
    models_dict = {k: default_model_entry(k, model_dir=model_dir) for k in MODEL_KEYS}
    return {
        # Legacy top-level fields
        "status": "idle",
        "is_running": False,
        "trigger": "none",
        "started_at": None,
        "finished_at": None,
        "duration_seconds": None,
        "trained_auditor_hash": "",
        "trained_auditee_pool_hash": "",
        "trained_auditee_custom_hash": "",
        "trained_auditor_count": 0,
        "trained_auditee_count": 0,
        "auditor_retrained": False,
        "auditee_retrained": False,
        "model_updated_at": None,
        "message": "Model AI siap digunakan.",
        "last_error": None,
        # New per-model & queue fields
        "models": models_dict,
        "queue": [],
    }


def read_raw_registry(status_file: Optional[str] = None) -> Dict[str, Any]:
    sfile = status_file or RETRAIN_STATUS_FILE
    mdir = os.path.dirname(sfile)
    base = default_registry_payload(model_dir=mdir)
    with _REGISTRY_LOCK:
        if os.path.exists(sfile):
            try:
                with open(sfile, "r", encoding="utf-8") as f:
                    saved = json.load(f)
                if isinstance(saved, dict):
                    saved_models = saved.get("models")
                    for k, v in saved.items():
                        if k != "models":
                            base[k] = v
                    if isinstance(saved_models, dict):
                        for mk in MODEL_KEYS:
                            if isinstance(saved_models.get(mk), dict):
                                base["models"][mk].update(saved_models[mk])
            except Exception:
                pass
    return base


def write_raw_registry(patch: Dict[str, Any], status_file: Optional[str] = None) -> Dict[str, Any]:
    sfile = status_file or RETRAIN_STATUS_FILE
    os.makedirs(os.path.dirname(sfile), exist_ok=True)
    with _REGISTRY_LOCK:
        curr = read_raw_registry(status_file=sfile)
        models_patch = patch.get("models")
        for k, v in patch.items():
            if k != "models":
                curr[k] = v
        if isinstance(models_patch, dict):
            for mk, mv in models_patch.items():
                if mk in curr["models"] and isinstance(mv, dict):
                    curr["models"][mk].update(mv)
        try:
            tmp = sfile + ".tmp"
            with open(tmp, "w", encoding="utf-8") as f:
                json.dump(curr, f, ensure_ascii=False, indent=2)
            os.replace(tmp, sfile)
        except Exception as e:
            print(f"[ModelRegistry] Warning: could not persist {sfile}: {e}")
        return curr


def update_model_entry(model_key: str, patch: Dict[str, Any], top_patch: Optional[Dict[str, Any]] = None, status_file: Optional[str] = None) -> Dict[str, Any]:
    payload: Dict[str, Any] = {"models": {model_key: patch}}
    if top_patch:
        payload.update(top_patch)
    return write_raw_registry(payload, status_file=status_file)


def bootstrap_trained_hashes_if_empty(status_file: Optional[str] = None) -> Dict[str, Any]:
    """
    On first boot, if a model artifact `.pkl/.pth` already exists on disk and `trained_data_hash`
    has not been recorded yet, seed `trained_data_hash` and `trained_rows` from the current baseline
    so pre-trained models start in `trained` state until new data arrives.
    """
    sfile = status_file or RETRAIN_STATUS_FILE
    mdir = os.path.dirname(sfile)
    with _REGISTRY_LOCK:
        reg = read_raw_registry(status_file=sfile)
        fps = compute_all_model_fingerprints()
        models_patch: Dict[str, Dict[str, Any]] = {}
        top_patch: Dict[str, Any] = {}
        changed = False

        for mk in MODEL_KEYS:
            entry = reg["models"].get(mk, {})
            fp = fps.get(mk, {})
            cur_hash = fp.get("hash", "")
            cur_rows = int(fp.get("rows", 0))
            exists = _artifact_exists(mk, model_dir=mdir)

            if exists and not entry.get("trained_data_hash") and cur_hash:
                mtime_iso = entry.get("trained_at") or _artifact_mtime_iso(mk, model_dir=mdir) or datetime.now(timezone.utc).isoformat()
                models_patch[mk] = {
                    "state": "trained",
                    "trained_data_hash": cur_hash,
                    "current_data_hash": cur_hash,
                    "trained_rows": cur_rows,
                    "current_rows": cur_rows,
                    "samples": entry.get("samples") or cur_rows,
                    "trained_at": mtime_iso,
                    "has_new_data": False,
                    "new_rows": 0,
                    "version": max(1, int(entry.get("version") or 1)),
                    "message": f"Model {MODEL_METADATA[mk]['label']} siap digunakan ({cur_rows} sampel).",
                }
                changed = True
                if mk == "kpi_auditee" and not reg.get("trained_auditee_pool_hash"):
                    top_patch["trained_auditee_pool_hash"] = fp.get("pool_hash", cur_hash)
                    top_patch["trained_auditee_count"] = cur_rows
                elif mk == "kpi_auditor" and not reg.get("trained_auditor_hash"):
                    top_patch["trained_auditor_hash"] = cur_hash
                    top_patch["trained_auditor_count"] = cur_rows

        if changed:
            return write_raw_registry({"models": models_patch, **top_patch}, status_file=sfile)
        return reg


def build_live_status_snapshot(
    auditor_hash: Optional[str] = None,
    auditee_custom_hash: Optional[str] = None,
    queue_snapshot: Optional[List[Dict[str, Any]]] = None,
    scheduler_info: Optional[Dict[str, Any]] = None,
    status_file: Optional[str] = None,
) -> Dict[str, Any]:
    """
    Assemble the complete `/retrain/status` response combining persisted registry state,
    live incremental fingerprints, active queue positions, and resource guard metrics.
    """
    sfile = status_file or RETRAIN_STATUS_FILE
    mdir = os.path.dirname(sfile)
    reg = bootstrap_trained_hashes_if_empty(status_file=sfile)
    fps = compute_all_model_fingerprints(auditor_hash=auditor_hash, auditee_custom_hash=auditee_custom_hash)
    prof = get_resource_profile()
    q_list = queue_snapshot if queue_snapshot is not None else reg.get("queue", [])

    # Index active / queued jobs by model
    queued_by_model: Dict[str, Dict[str, Any]] = {}
    running_models: List[str] = []
    for job in q_list:
        mk = job.get("model")
        if not mk:
            continue
        if job.get("state") == "running":
            running_models.append(mk)
            queued_by_model[mk] = job
        elif job.get("state") == "queued" and mk not in queued_by_model:
            queued_by_model[mk] = job

    models_out: Dict[str, Dict[str, Any]] = {}
    for mk in MODEL_KEYS:
        entry = dict(reg["models"].get(mk) or default_model_entry(mk, model_dir=mdir))
        fp = fps.get(mk, {})
        cur_hash = str(fp.get("hash") or "")
        cur_rows = int(fp.get("rows") or 0)
        trained_hash = str(entry.get("trained_data_hash") or "")
        trained_rows = int(entry.get("trained_rows") or 0)
        exists = _artifact_exists(mk, model_dir=mdir)

        has_new = bool(cur_hash and cur_hash != trained_hash)
        if mk == "kpi_auditee" and auditee_custom_hash:
            if auditee_custom_hash != str(reg.get("trained_auditee_custom_hash") or ""):
                has_new = True
        if mk == "kpi_auditor" and auditor_hash:
            if auditor_hash != str(reg.get("trained_auditor_hash") or ""):
                has_new = True

        new_rows = max(0, cur_rows - trained_rows) if has_new else 0
        if has_new and new_rows == 0 and cur_rows > 0:
            # Rows were modified in-place or custom items updated
            new_rows = max(1, cur_rows - trained_rows) if cur_rows != trained_rows else 1

        q_job = queued_by_model.get(mk)
        if q_job and q_job.get("state") == "running":
            state = "training"
            q_pos = 0
            wait_reason = None
        elif q_job and q_job.get("state") == "queued":
            state = "queued"
            q_pos = int(q_job.get("position") or 1)
            wait_reason = prof.get("guard_reason") if not prof.get("guard_ok") else (
                f"Menunggu giliran antrian #{q_pos}"
            )
        elif entry.get("state") == "failed" and entry.get("last_error"):
            state = "failed"
            q_pos = None
            wait_reason = None
        elif not exists and not trained_hash:
            state = "never_trained"
            q_pos = None
            wait_reason = None
        elif has_new:
            state = "untrained_data"
            q_pos = None
            wait_reason = None
        else:
            state = "trained"
            q_pos = None
            wait_reason = None

        label = MODEL_METADATA[mk]["label"]
        if state == "training":
            prog = entry.get("progress") or {}
            pct = prog.get("percent", 0)
            b_cur = prog.get("batch", 1)
            b_tot = prog.get("total_batches", 1)
            msg = f"Sedang melatih {label} (Batch {b_cur}/{b_tot} — {pct}%)."
        elif state == "queued":
            msg = wait_reason or f"{label} dalam antrian pelatihan (#{q_pos})."
        elif state == "untrained_data":
            msg = f"Ada data baru ({new_rows} pembaruan) pada {label} yang belum dilatih."
        elif state == "never_trained":
            msg = f"Model {label} belum pernah dilatih."
        elif state == "failed":
            msg = entry.get("message") or f"Pelatihan {label} gagal: {entry.get('last_error')}"
        else:
            msg = entry.get("message") or f"Model {label} sudah dilatih ({trained_rows or cur_rows} sampel)."

        entry.update({
            "state": state,
            "has_new_data": has_new,
            "new_rows": int(new_rows),
            "current_rows": cur_rows,
            "trained_rows": trained_rows,
            "current_data_hash": cur_hash,
            "queue_position": q_pos,
            "job_id": q_job.get("job_id") if q_job else entry.get("job_id"),
            "waiting_reason": wait_reason,
            "message": msg,
        })
        models_out[mk] = entry

    any_running = len(running_models) > 0
    any_queued = any(m["state"] == "queued" for m in models_out.values())

    auditee_fp = fps.get("kpi_auditee", {})
    pool_hash = str(auditee_fp.get("pool_hash") or auditee_fp.get("hash") or "")
    pool_count = int(auditee_fp.get("rows") or 0)
    trained_pool_hash = str(reg.get("trained_auditee_pool_hash") or models_out["kpi_auditee"].get("trained_data_hash") or "")
    has_untrained_pool = bool(pool_hash and pool_hash != trained_pool_hash)
    has_untrained_auditor = bool(auditor_hash and auditor_hash != str(reg.get("trained_auditor_hash") or ""))
    has_untrained_custom = bool(auditee_custom_hash and auditee_custom_hash != str(reg.get("trained_auditee_custom_hash") or ""))

    top_status = "running" if any_running else ("queued" if any_queued else str(reg.get("status") or "idle"))
    if top_status == "running" and not any_running:
        top_status = "completed" if reg.get("finished_at") else "idle"

    out = {
        **reg,
        "status": top_status,
        "is_running": bool(any_running),
        "current_auditee_pool_hash": pool_hash,
        "current_auditee_pool_count": pool_count,
        "has_untrained_auditee_pool_data": has_untrained_pool,
        "has_untrained_auditor_data": has_untrained_auditor,
        "has_untrained_auditee_custom_data": has_untrained_custom,
        "models": models_out,
        "queue": q_list,
        "resource": prof,
        "scheduler": scheduler_info or {"running": True, "auto_enabled": True, "last_detect_at": None},
    }
    return out


def validate_and_promote_artifacts(
    model_key: str,
    staged_files: Dict[str, str],
    new_metrics: Dict[str, Any],
    model_dir: Optional[str] = None,
    data_server_dir: Optional[str] = None,
    status_file: Optional[str] = None,
) -> Dict[str, Any]:
    """
    Validate newly trained model metrics against the previously promoted model.
    - If valid (within tolerance), atomically promotes staged files into `model_dir`
      and keeps at most 2 backup versions (`.v1.bak`, `.v2.bak`) to bound disk usage.
    - If worse than previous model beyond tolerance, deletes staged files, preserves
      existing production artifacts, and returns `{"promoted": False, "reason": ...}`.
    """
    mdir = model_dir or MODEL_DIR
    sfile = status_file or os.path.join(mdir, "retrain_status.json")
    prof = get_resource_profile()
    tol = float(prof.get("metric_tolerance", 0.03))

    reg = read_raw_registry(status_file=sfile)
    prev_entry = reg["models"].get(model_key, {})
    prev_metrics = prev_entry.get("metrics") or {}
    meta = MODEL_METADATA.get(model_key, {})
    metric_name = meta.get("primary_metric", "primary")
    higher_is_better = bool(meta.get("higher_is_better", True))

    new_val = new_metrics.get("primary", new_metrics.get(metric_name))
    prev_val = prev_metrics.get("primary", prev_metrics.get(metric_name))

    # Check quality regression if previous metric exists
    if new_val is not None and prev_val is not None:
        try:
            nv = float(new_val)
            pv = float(prev_val)
            if higher_is_better:
                if nv + tol < pv:
                    for st_path in staged_files.values():
                        if os.path.exists(st_path):
                            try:
                                os.remove(st_path)
                            except Exception:
                                pass
                    return {
                        "promoted": False,
                        "reason": f"Rollback otomatis: metrik {metric_name} turun dari {pv:.4f} ke {nv:.4f} (toleransi {tol}). Model lama dipertahankan.",
                        "previous_metric": pv,
                        "new_metric": nv,
                    }
            else:
                # Lower is better (e.g. RMSE): allow up to (1 + tol*2) relative or +tol absolute
                max_allowed = pv * (1.0 + max(tol * 2.0, 0.10)) + tol
                if nv > max_allowed:
                    for st_path in staged_files.values():
                        if os.path.exists(st_path):
                            try:
                                os.remove(st_path)
                            except Exception:
                                pass
                    return {
                        "promoted": False,
                        "reason": f"Rollback otomatis: error {metric_name} naik dari {pv:.4f} ke {nv:.4f}. Model lama dipertahankan.",
                        "previous_metric": pv,
                        "new_metric": nv,
                    }
        except Exception:
            pass

    # Promote staged files atomically and rotate at most 2 backup versions
    os.makedirs(mdir, exist_ok=True)
    for final_name, staged_path in staged_files.items():
        if not os.path.exists(staged_path):
            continue
        dest_path = os.path.join(mdir, final_name)
        bak1 = dest_path + ".v1.bak"
        bak2 = dest_path + ".v2.bak"
        try:
            if os.path.exists(bak1):
                shutil.copy2(bak1, bak2)
            if os.path.exists(dest_path):
                shutil.copy2(dest_path, bak1)
        except Exception:
            pass
        os.replace(staged_path, dest_path)

        # Mirror to data-server if directory exists
        if data_server_dir and os.path.isdir(os.path.dirname(data_server_dir)):
            try:
                os.makedirs(data_server_dir, exist_ok=True)
                shutil.copy2(dest_path, os.path.join(data_server_dir, final_name))
            except Exception:
                pass

    # Clean up any leftover .tmp files in model_dir
    try:
        for fn in os.listdir(mdir):
            if fn.endswith(".tmp") or fn.endswith(".staging"):
                try:
                    os.remove(os.path.join(mdir, fn))
                except Exception:
                    pass
    except Exception:
        pass

    return {
        "promoted": True,
        "previous_metric": prev_val,
        "new_metric": new_val,
    }
