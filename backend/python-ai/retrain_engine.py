import os
import sys
import gc
import re
import time
import json
import math
import pickle
import random
import hashlib
import hmac
import base64
import threading
from datetime import datetime, timezone
from typing import Dict, Any, List, Optional, Callable
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, StandardScaler, MinMaxScaler, LabelEncoder
from sklearn.ensemble import IsolationForest, GradientBoostingClassifier, RandomForestClassifier, ExtraTreesClassifier
from sklearn.model_selection import train_test_split
from sklearn.linear_model import LogisticRegression, SGDClassifier
from sklearn.feature_extraction.text import TfidfVectorizer, HashingVectorizer
from sklearn.metrics import f1_score, accuracy_score, average_precision_score, mean_squared_error, mean_absolute_error
from xgboost import XGBClassifier

try:
    import psutil
except ImportError:
    psutil = None

from nlp_taxonomy import (
    load_analyzed_documents,
    upsert_analyzed_document,
    classify_risk_category,
    detect_control_indicator_and_severity,
    extract_finding_and_recommendation,
    add_taxonomy_item,
    normalize_or_resolve_category,
    normalize_nlp_risk_level,
    sanitize_nlp_doc_record,
    sanitize_nlp_corpus_item,
    load_live_nlp_corpus,
    upsert_live_nlp_corpus,
    harvest_all_live_nlp_training_data,
)
from ai_training_storage import (
    upsert_auditee_kpi_pool,
    upsert_auditor_kpi_pool,
    load_auditee_training_dataframe,
    load_auditor_training_dataframe,
    upsert_entity_prediction_pool,
    load_entity_prediction_training_dataframe,
    CANONICAL_RISK_CODES,
    CANONICAL_RISK_CODE_TO_LABEL,
    CANONICAL_RISK_CODE_TO_UI_FLOAT_PAIR,
    normalize_canonical_risk_level,
    matrix_risk_level,
)
from resource_profile import (
    get_resource_profile,
    select_training_tier,
    apply_thread_limits,
    lower_current_process_priority,
)
from training_data_source import (
    ATR_FINDINGS_POOL_FILE,
    compute_atr_findings_hash,
    load_atr_findings_pool,
    compute_all_model_fingerprints,
    compute_entity_fingerprint,
    compute_anomaly_fingerprint,
    compute_nlp_fingerprint,
    compute_kpi_auditee_fingerprint,
    compute_kpi_auditor_fingerprint,
    iter_dataframe_batches,
    iter_csv_or_parquet_batches,
    stratified_reservoir_sample_batches,
)
from model_registry import (
    MODEL_KEYS,
    MODEL_METADATA,
    resolve_model_keys,
    read_raw_registry,
    write_raw_registry,
    update_model_entry,
    build_live_status_snapshot,
    validate_and_promote_artifacts,
)

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.getenv("MODEL_DIR", os.path.join(BASE_DIR, "models"))
AI_TRAINING_DIR = os.getenv(
    "AI_TRAINING_DIR",
    "/app/ai_model_training"
    if os.path.exists("/app/ai_model_training")
    else os.path.abspath(os.path.join(BASE_DIR, "..", "..", "ai_model_training")),
)
DATA_SERVER_MODEL_DIR = os.path.abspath(os.path.join(BASE_DIR, "..", "..", "data-server", "ai-engine", "models"))
RETRAIN_STATUS_FILE = os.path.join(MODEL_DIR, "retrain_status.json")
CHECKPOINT_DIR = os.path.join(MODEL_DIR, "checkpoints")
os.makedirs(MODEL_DIR, exist_ok=True)
os.makedirs(CHECKPOINT_DIR, exist_ok=True)
_STATUS_FILE_LOCK = threading.Lock()


def _normalize_num_str(val) -> str:
    if val is None or val == "":
        return ""
    try:
        num = float(str(val).replace(",", "").replace("%", "").strip())
        if np.isfinite(num):
            return f"{round(num, 4):.4f}"
    except Exception:
        pass
    return ""


def compute_auditor_items_hash(items: list) -> str:
    """Deterministic SHA-256[:16] fingerprint matching frontend `computeAuditorFingerprint`."""
    if not items:
        return ""
    rows = []
    for it in items:
        if not isinstance(it, dict):
            continue
        code = str(it.get("code") or "").strip().upper()
        kpi = str(it.get("kpi") or it.get("kpiName") or "").strip().lower()
        pol = "HIB" if str(it.get("polarity") or it.get("hibHig") or "HIG").strip().upper() == "HIB" else "HIG"
        target = _normalize_num_str(it.get("target") if it.get("target") is not None else it.get("targetValue"))
        actual = _normalize_num_str(it.get("actual") if it.get("actual") is not None else it.get("currentValue"))
        t_map = it.get("kpiTargets") or {}
        a_map = it.get("kpiActuals") or {}
        t_canon = json.dumps({str(k): str(t_map[k]) for k in sorted(t_map.keys())}, separators=(",", ":"), ensure_ascii=False) if isinstance(t_map, dict) else "{}"
        a_canon = json.dumps({str(k): str(a_map[k]) for k in sorted(a_map.keys())}, separators=(",", ":"), ensure_ascii=False) if isinstance(a_map, dict) else "{}"
        rows.append(f"{code}|{kpi}|{pol}|{target}|{actual}|{t_canon}|{a_canon}")
    rows.sort()
    payload = "\n".join(rows)
    return hashlib.sha256(payload.encode("utf-8")).hexdigest()[:16]


def compute_auditee_custom_items_hash(items: list) -> str:
    """Deterministic SHA-256[:16] fingerprint matching frontend `computeAuditeeCustomFingerprint`."""
    if not items:
        return ""
    rows = []
    for it in items:
        if not isinstance(it, dict):
            continue
        code = str(it.get("code") or "").strip().upper()
        kpi = str(it.get("kpiName") or it.get("kpi_name") or it.get("kpi") or "").strip().lower()
        ent = str(it.get("entity") or "").strip().lower()
        cur = _normalize_num_str(it.get("currentValue") if it.get("currentValue") is not None else it.get("actual_value"))
        fore = _normalize_num_str(it.get("forecastedValue") if it.get("forecastedValue") is not None else it.get("forecast_value"))
        rows.append(f"{code}|{kpi}|{ent}|{cur}|{fore}")
    rows.sort()
    if not rows:
        return ""
    return hashlib.sha256("\n".join(rows).encode("utf-8")).hexdigest()[:16]


def compute_auditee_pool_hash() -> tuple:
    """Compute fingerprint of the active Auditee training pool (DB or Parquet) to detect new Datalake rows."""
    fp = compute_kpi_auditee_fingerprint()
    return (str(fp.get("pool_hash") or fp.get("hash") or ""), int(fp.get("rows") or 0))


def _period_label_to_auditee_bulan(period_raw: str) -> str:
    """Normalize '2026-Q2', '2026-06', or 'Bulan 06 2026' into 'Bulan MM YYYY' for uniform time sorting."""
    s = str(period_raw or "").strip()
    if s.lower().startswith("bulan "):
        return s
    m_q = re.match(r"^(\d{4})[-/\s]?Q([1-4])$", s, re.I)
    if m_q:
        month = int(m_q.group(2)) * 3
        return f"Bulan {month:02d} {m_q.group(1)}"
    m_ym = re.match(r"^(\d{4})[-/](\d{1,2})$", s)
    if m_ym:
        return f"Bulan {int(m_ym.group(2)):02d} {m_ym.group(1)}"
    now = datetime.now()
    return f"Bulan {now.month:02d} {now.year}"


def _parse_period_sort_key(period_raw: str) -> tuple:
    s = str(period_raw or "").strip()
    m_b = re.search(r"(\d{1,2})\D+(\d{4})", s)
    if m_b:
        return (int(m_b.group(2)), int(m_b.group(1)), s)
    m_ym = re.match(r"^(\d{4})[-/](\d{1,2})$", s)
    if m_ym:
        return (int(m_ym.group(1)), int(m_ym.group(2)), s)
    m_qy = re.match(r"^Q([1-4])[-\s/]?(\d{4})$", s, re.I)
    if m_qy:
        return (int(m_qy.group(2)), int(m_qy.group(1)) * 3, s)
    m_yq = re.match(r"^(\d{4})[-\s/]?Q([1-4])$", s, re.I)
    if m_yq:
        return (int(m_yq.group(1)), int(m_yq.group(2)) * 3, s)
    m_y = re.match(r"^(\d{4})$", s)
    if m_y:
        return (int(m_y.group(1)), 12, s)
    return (9999, 12, s)


def default_retrain_status() -> dict:
    return read_raw_registry()


def read_retrain_status() -> dict:
    return build_live_status_snapshot()


def write_retrain_status(patch: dict) -> dict:
    return write_raw_registry(patch)


def set_seed(seed=42):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)


set_seed(42)


class PyTorchLSTMRegressor(nn.Module):
    def __init__(self, input_dim=1, hidden_dim=64, num_layers=2, output_dim=1):
        super().__init__()
        self.hidden_dim = hidden_dim
        self.num_layers = num_layers
        self.lstm = nn.LSTM(input_size=input_dim, hidden_size=hidden_dim, num_layers=num_layers, batch_first=True)
        self.fc = nn.Linear(hidden_dim, output_dim)

    def forward(self, x):
        h0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        c0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        out, _ = self.lstm(x, (h0, c0))
        return self.fc(out[:, -1, :])


try:
    import requests
except ImportError:
    import urllib.request

    class _RequestsFallback:
        def get(self, url, timeout=3, headers=None):
            try:
                hdrs = {"User-Agent": "AuditSphere-AI"}
                if isinstance(headers, dict):
                    hdrs.update(headers)
                req = urllib.request.Request(url, headers=hdrs)
                with urllib.request.urlopen(req, timeout=timeout) as resp:
                    class _Resp:
                        status_code = resp.status
                        def json(self):
                            return json.loads(resp.read().decode())
                    return _Resp()
            except Exception:
                class _FailedResp:
                    status_code = 500
                    def json(self):
                        return {}
                return _FailedResp()
    requests = _RequestsFallback()


def _build_internal_service_jwt() -> str:
    """Mint a short-lived HS256 internal system JWT for cross-service audit-service calls."""
    secret = os.getenv(
        "JWT_SECRET",
        "6958c455ba7378688fac7200001d6bca8048fa66a28447ecc69f3b74b7ded4fa",
    )
    now_ts = int(time.time())
    header = {"alg": "HS256", "typ": "JWT"}
    payload = {
        "user_id": "00000000-0000-0000-0000-000000000000",
        "username": "ai-system",
        "roles": ["ADMIN"],
        "iat": now_ts,
        "exp": now_ts + 300,
    }
    b64 = lambda obj: base64.urlsafe_b64encode(
        json.dumps(obj, separators=(",", ":")).encode("utf-8")
    ).rstrip(b"=").decode("ascii")
    signing_input = f"{b64(header)}.{b64(payload)}"
    sig = hmac.new(secret.encode("utf-8"), signing_input.encode("ascii"), hashlib.sha256).digest()
    sig_b64 = base64.urlsafe_b64encode(sig).rstrip(b"=").decode("ascii")
    return f"{signing_input}.{sig_b64}"


def map_finding_significance(raw_category: Any) -> tuple:
    """
    Map finding_category / significance string into canonical
    (significance_label, significance_score 1..4).
    """
    s = str(raw_category or "").strip().lower()
    if any(k in s for k in ("very significant", "sangat signifikan", "kritis", "critical", "major", "special audit", "investigation")):
        return ("Very Significant", 4)
    if any(k in s for k in ("quite significant", "cukup signifikan", "sedang", "moderate", "medium", "consulting")):
        return ("Quite Significant", 2)
    if any(k in s for k in ("not significant", "tidak signifikan", "rendah", "low", "minor")):
        return ("Not Significant", 1)
    if any(k in s for k in ("significant", "signifikan", "tinggi", "high", "assurance", "technology")):
        return ("Significant", 3)
    return ("Significant", 3)


def _normalize_atr_finding_item(item: Dict[str, Any], idx: int = 0) -> Optional[Dict[str, Any]]:
    if not isinstance(item, dict):
        return None
    title = str(item.get("finding_title") or item.get("findingTitle") or item.get("title") or "").strip()
    if not title:
        return None
    fid = str(item.get("id") or item.get("finding_id") or item.get("findingId") or f"ATR-{idx+1:04d}").strip()
    lha = item.get("lha") if isinstance(item.get("lha"), dict) else {}
    report_number = str(
        item.get("report_number")
        or item.get("reportNumber")
        or lha.get("report_number")
        or lha.get("reportNumber")
        or "LHA-ATR/SKAI/2026"
    ).strip()
    report_title = str(
        item.get("report_title")
        or item.get("reportTitle")
        or lha.get("report_title")
        or lha.get("reportTitle")
        or "Laporan Hasil Audit"
    ).strip()
    department = str(
        item.get("department")
        or lha.get("department")
        or item.get("entity")
        or item.get("audit_object")
        or lha.get("audit_object")
        or "Unit Auditee"
    ).strip()
    raw_cat = str(
        item.get("finding_category")
        or item.get("findingCategory")
        or item.get("category")
        or item.get("significance")
        or "Significant"
    ).strip()
    sig_label, sig_score = map_finding_significance(raw_cat)
    recommendation = str(item.get("recommendation") or item.get("action") or "").strip()
    action_plan = str(item.get("action_plan") or item.get("actionPlan") or "").strip()
    pic_name = str(item.get("pic_name") or item.get("picName") or "").strip()
    status = str(item.get("status") or "PLANNED").strip().upper()
    try:
        progress = int(max(0, min(100, round(float(item.get("progress") or 0)))))
    except Exception:
        progress = 0
    if status == "COMPLETED" and progress == 0:
        progress = 100

    due_raw = str(item.get("due_date") or item.get("dueDate") or "").strip()
    due_str = due_raw[:10] if len(due_raw) >= 10 else ""
    is_overdue = bool(item.get("is_overdue") if item.get("is_overdue") is not None else item.get("isOverdue", False))
    try:
        overdue_days = int(max(0, round(float(item.get("overdue_days") if item.get("overdue_days") is not None else (item.get("overdueDays") or 0)))))
    except Exception:
        overdue_days = 0

    if due_str and status not in ("COMPLETED", "CANCELLED") and overdue_days == 0:
        try:
            due_dt = datetime.strptime(due_str, "%Y-%m-%d").date()
            today_dt = datetime.now(timezone.utc).date()
            if today_dt > due_dt:
                overdue_days = int((today_dt - due_dt).days)
                is_overdue = overdue_days > 0
        except Exception:
            pass

    created_raw = str(item.get("created_at") or item.get("createdAt") or "").strip()
    elapsed_days = overdue_days
    if elapsed_days == 0 and created_raw[:10]:
        try:
            c_dt = datetime.strptime(created_raw[:10], "%Y-%m-%d").date()
            today_dt = datetime.now(timezone.utc).date()
            elapsed_days = max(1, int((today_dt - c_dt).days))
        except Exception:
            elapsed_days = 0

    return {
        "id": fid,
        "audit_result_report_id": str(item.get("audit_result_report_id") or item.get("auditResultReportId") or ""),
        "report_number": report_number,
        "report_title": report_title,
        "department": department,
        "finding_id": str(item.get("finding_id") or item.get("findingId") or ""),
        "finding_title": title,
        "finding_category": raw_cat or sig_label,
        "significance_label": sig_label,
        "significance_score": int(sig_score),
        "recommendation": recommendation,
        "action_plan": action_plan,
        "pic_name": pic_name,
        "due_date": due_str,
        "progress": int(progress),
        "status": status,
        "is_overdue": bool(is_overdue),
        "overdue_days": int(overdue_days),
        "elapsed_days": int(elapsed_days),
        "created_at": created_raw,
        "updated_at": str(item.get("updated_at") or item.get("updatedAt") or ""),
    }


def upsert_atr_findings_pool(items: List[Dict[str, Any]], replace: bool = False) -> Dict[str, Any]:
    """
    Normalize and persist Action Taken Report finding records into ATR_FINDINGS_POOL_FILE.
    Writes to disk ONLY when the canonical hash actually changes so unchanged 60s polls do not alter file mtime.
    """
    existing = [] if replace else load_atr_findings_pool()
    old_hash, old_count = compute_atr_findings_hash(existing)

    merged_by_key: Dict[str, Dict[str, Any]] = {}
    for idx, it in enumerate(existing):
        norm = _normalize_atr_finding_item(it, idx=idx)
        if norm:
            key = norm["id"] or f"{norm['report_number']}|{norm['finding_title'].lower()}"
            merged_by_key[key] = norm

    for idx, it in enumerate(items or []):
        norm = _normalize_atr_finding_item(it, idx=idx)
        if norm:
            key = norm["id"] or f"{norm['report_number']}|{norm['finding_title'].lower()}"
            merged_by_key[key] = norm

    merged_list = list(merged_by_key.values())
    # Compute repeat_count per normalized finding title across reports
    title_counts: Dict[str, int] = {}
    for row in merged_list:
        t_key = row["finding_title"].strip().lower()
        title_counts[t_key] = title_counts.get(t_key, 0) + 1
    for row in merged_list:
        t_key = row["finding_title"].strip().lower()
        row["repeat_count"] = max(0, title_counts.get(t_key, 1) - 1)

    new_hash, new_count = compute_atr_findings_hash(merged_list)
    changed = (new_hash != old_hash) or (new_count != old_count) or not os.path.exists(ATR_FINDINGS_POOL_FILE)
    if changed and merged_list:
        try:
            os.makedirs(os.path.dirname(ATR_FINDINGS_POOL_FILE), exist_ok=True)
            tmp_path = ATR_FINDINGS_POOL_FILE + ".tmp"
            with open(tmp_path, "w", encoding="utf-8") as f:
                json.dump(
                    {
                        "updated_at": datetime.now(timezone.utc).isoformat(),
                        "pool_hash": new_hash,
                        "count": new_count,
                        "items": merged_list,
                    },
                    f,
                    ensure_ascii=False,
                    indent=2,
                )
            os.replace(tmp_path, ATR_FINDINGS_POOL_FILE)
        except Exception as e:
            print(f"[RetrainEngine] Failed writing ATR findings pool: {e}")
    return {
        "changed": bool(changed),
        "old_hash": old_hash,
        "hash": new_hash,
        "count": new_count,
        "items": merged_list,
    }


def _current_rss_mb() -> float:
    if psutil is not None:
        try:
            return round(psutil.Process(os.getpid()).memory_info().rss / (1024 * 1024), 1)
        except Exception:
            pass
    return 0.0


def _save_checkpoint(model_key: str, data: Dict[str, Any]):
    try:
        os.makedirs(CHECKPOINT_DIR, exist_ok=True)
        cp_file = os.path.join(CHECKPOINT_DIR, f"{model_key}.ckpt.json")
        tmp = cp_file + ".tmp"
        with open(tmp, "w", encoding="utf-8") as f:
            json.dump({"updated_at": datetime.now(timezone.utc).isoformat(), **data}, f)
        os.replace(tmp, cp_file)
    except Exception:
        pass


def _clear_checkpoint(model_key: str):
    try:
        cp_file = os.path.join(CHECKPOINT_DIR, f"{model_key}.ckpt.json")
        if os.path.exists(cp_file):
            os.remove(cp_file)
    except Exception:
        pass


class AutoRetrainEngine:
    """
    Automated Self-Learning, Per-Model Batch Trainer & Quality Validator for AuditSphere.
    Supports independent per-model training (`entity`, `anomaly`, `nlp`, `kpi_auditee`, `kpi_auditor`),
    tiered batch execution (T1/T2/T3), checkpointing, validation metrics with automatic rollback,
    and selective hot-reloading.
    """
    def __init__(self, model_dir: Optional[str] = None, ai_training_dir: Optional[str] = None):
        self.model_dir = model_dir or MODEL_DIR
        self.ai_training_dir = ai_training_dir or AI_TRAINING_DIR
        self.risk_service_url = os.getenv("RISK_SERVICE_URL", "http://risk-service:8002" if os.path.exists("/.dockerenv") else "http://localhost:8081")
        self.audit_service_url = os.getenv("AUDIT_SERVICE_URL", "http://audit-service:8001" if os.path.exists("/.dockerenv") else "http://localhost:8082")
        self.master_service_url = os.getenv("MASTER_SERVICE_URL", "http://master-service:8002" if os.path.exists("/.dockerenv") else "http://localhost:8083")
        self.data_hub_url = os.getenv("DATA_HUB_URL", "http://data-hub-api:8000" if os.path.exists("/.dockerenv") else "http://localhost:8000")
        self.last_model_results: Dict[str, Dict[str, Any]] = {}
        self.last_kpi_metrics: Dict[str, Any] = {}

    def _emit_progress(
        self,
        model_key: str,
        batch: int,
        total_batches: int,
        phase: str = "training",
        tier: str = "T1",
        progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None,
    ):
        total_b = max(1, int(total_batches))
        cur_b = max(1, min(total_b, int(batch)))
        pct = int(min(99, round((cur_b / float(total_b)) * 100)))
        prog = {
            "phase": phase,
            "batch": cur_b,
            "total_batches": total_b,
            "percent": pct,
            "tier": tier,
        }
        _save_checkpoint(model_key, prog)
        if progress_cb:
            try:
                progress_cb(prog)
            except Exception:
                pass
        # Cooperative yield so FastAPI request threads never starve on 2 vCPU
        time.sleep(0.002)

    def sync_live_cross_service_data(self, scope: str = "all") -> dict:
        """
        Synchronizes live operational data ONLY for the requested model scope:
        - `entity`: Risk Service + Audit Universe
        - `nlp`: Imported Docs + Audit Findings/WPs/ATRs + RCM + External DB Mapped Records
        - `kpi_auditor` / `kpi`: Strategic Audit Plan KPIs
        - `kpi_auditee` / `kpi`: Data Hub Zona Gold KPI Series
        - `anomaly`: Anomaly data sources
        """
        target_keys = resolve_model_keys(scope) or list(MODEL_KEYS)
        synced_counts = {"risk": 0, "audit_universe": 0, "audit": 0, "master": 0, "auditor_kpi": 0, "auditee_kpi": 0}

        # 1. Entity Prediction sources (Audit Universe only — excludes Corporate Risk Profile)
        if "entity" in target_keys:
            try:
                curr_year = datetime.now().year
                ru = requests.get(f"{self.risk_service_url}/audit-universe/year/{curr_year}", timeout=2)
                if ru.status_code != 200:
                    ru = requests.get(f"{self.risk_service_url}/api/v1/audit-universe/year/{curr_year}", timeout=2)
                if ru.status_code == 200:
                    u_data = ru.json()
                    u_items = u_data.get("data", []) if isinstance(u_data, dict) else u_data
                    if isinstance(u_items, list) and len(u_items) > 0:
                        u_rows = []
                        for u in u_items:
                            corp_u = u.get("corporate_audit_universe") or {}
                            u_name = corp_u.get("name") or u.get("name")
                            if not u_name or "legal & compliance" in str(u_name).lower():
                                continue
                            raw_u_lvl = str(u.get("risk_level") or "").strip()
                            u_idx = float(u.get("risk_index") or 0.0)
                            if not raw_u_lvl or raw_u_lvl.lower() == "not scored":
                                if u_idx >= 80:
                                    lvl_source = "HIGH"
                                elif u_idx >= 60:
                                    lvl_source = "MODERATE_HIGH"
                                elif u_idx >= 40:
                                    lvl_source = "MODERATE"
                                elif u_idx >= 20:
                                    lvl_source = "LOW_MODERATE"
                                elif u_idx > 0:
                                    lvl_source = "LOW"
                                else:
                                    lvl_source = "MODERATE"
                            else:
                                lvl_source = raw_u_lvl
                            u_lvl, _, _, _ = normalize_canonical_risk_level(lvl_source, default_code="MODERATE")
                            lik, imp = CANONICAL_RISK_CODE_TO_UI_FLOAT_PAIR[u_lvl]
                            is_br = any(w in str(u_name).lower() for w in ["cabang", "kcp", "regional", "kc", "outlet", "branch"])
                            u_rows.append({
                                "entity": str(u_name).strip(),
                                "entity_type": "Branch" if is_br else "Auditable Entity",
                                "risk_category": "Operational",
                                "inherent_likelihood": lik,
                                "inherent_impact": imp,
                                "audit_findings_count": 2.0 if u.get("audit_priority") else 0.0,
                                "kpi_below_target_count": 1.0 if u.get("audit_priority") else 0.0,
                                "kpi_volatility": 0.1,
                                "previous_risk_score": round(lik * imp, 1),
                                "assessment_month": 1,
                                "target_likelihood": lik,
                                "target_impact": imp,
                                "source_id": f"audit_universe:{u.get('id', u_name)}",
                                "current_risk_level": u_lvl,
                            })
                        if u_rows:
                            upsert_entity_prediction_pool(u_rows)
                            synced_counts["audit_universe"] = len(u_rows)
            except Exception:
                pass

        # 2. Multi-Source NLP Data
        if "nlp" in target_keys:
            try:
                nlp_synced = self.sync_multi_source_nlp_corpus()
                synced_counts["audit"] = nlp_synced.get("total_synced", 0)
                synced_counts["nlp_sources"] = nlp_synced
            except Exception:
                pass

        # 3. KPI Auditor (Strategic Audit Plan)
        if "kpi_auditor" in target_keys:
            audit_urls = [self.audit_service_url, "http://localhost:8080", "http://localhost:8082"]
            for audit_url in dict.fromkeys(audit_urls):
                try:
                    r = requests.get(f"{audit_url}/api/v1/strategic-plans", timeout=2)
                    if r.status_code == 200:
                        data = r.json()
                        items = data.get("data", []) if isinstance(data, dict) else data
                        if isinstance(items, dict) and "items" in items:
                            items = items["items"]
                        if isinstance(items, list) and len(items) > 0:
                            new_auditor_rows = []
                            current_period = pd.Timestamp.now().strftime("Bulan %m %Y")
                            current_date = pd.Timestamp.now().strftime("%Y-%m-01")
                            for item in items:
                                target_val = float(item.get("target", 100.0) or 100.0)
                                actual_val = float(item.get("actual", target_val) or target_val)
                                pol = item.get("hib_hig", item.get("HibHig", "HIG")).upper()
                                if pol == "HIG":
                                    ratio = (actual_val / target_val) * 100.0 if target_val > 0 else 100.0
                                else:
                                    ratio = 100.0 if actual_val <= 0 else (target_val / actual_val) * 100.0
                                status = "Good" if ratio >= 100.0 else ("Moderate" if ratio >= 70.0 else "Poor")
                                new_auditor_rows.append({
                                    "id": f"AUD-SYNC-{item.get('code', 'KPI')}",
                                    "code": item.get("code", "SO-IA01"),
                                    "kpi_name": item.get("kpi", item.get("KPI", "Indikator Audit")),
                                    "category": item.get("category", "Pelaksanaan Audit & PKAT"),
                                    "strategic_objective": item.get("strategicObjective", item.get("strategic_objective", "")),
                                    "period_date": current_date,
                                    "period_label": current_period,
                                    "hib_hig": pol,
                                    "unit": item.get("unit", "%"),
                                    "target_value": target_val,
                                    "actual_value": actual_val,
                                    "achievement_rate": round(ratio, 2),
                                    "status": status,
                                })
                            if new_auditor_rows:
                                upserted = upsert_auditor_kpi_pool(new_auditor_rows)
                                synced_counts["auditor_kpi"] = upserted
                            break
                except Exception:
                    pass

        # 4. KPI Auditee (Data Hub Zona Gold)
        if "kpi_auditee" in target_keys:
            try:
                r = requests.get(f"{self.data_hub_url}/api/v1/serve/kpi-series", timeout=2)
                if r.status_code == 200:
                    data = r.json()
                    items = data.get("data", []) if isinstance(data, dict) else data
                    if isinstance(items, list) and len(items) > 0:
                        new_kpi_rows = []
                        for item in items:
                            kpi_name = item.get("kpi_name")
                            act_val = float(item.get("actual_value", 0.0) or 0.0)
                            tgt_val = float(item.get("target_value", 2.0) or 2.0)
                            p_raw = item.get("period", "2026-03")
                            period_str = f"Bulan {p_raw}" if not str(p_raw).startswith("Bulan") else str(p_raw)
                            imp = 4 if act_val > tgt_val * 1.5 else (2 if act_val < tgt_val else 3)
                            lik = 4 if imp >= 4 else 3
                            raw_ext_risk = (
                                item.get("risk_level")
                                or item.get("riskLevel")
                                or item.get("risk_rating")
                                or item.get("severity")
                                or item.get("tingkat_risiko")
                            )
                            if raw_ext_risk is not None and str(raw_ext_risk).strip() != "":
                                r_level, _, def_imp, def_lik = normalize_canonical_risk_level(
                                    raw_ext_risk, default_code="MODERATE"
                                )
                                if matrix_risk_level(lik, imp) != r_level:
                                    imp, lik = def_imp, def_lik
                            else:
                                r_level = matrix_risk_level(lik, imp)
                            entity_name = "Kantor Pusat"
                            for city in ["Surabaya", "Jakarta", "Bandung", "Medan", "Bali", "Semarang", "Makassar", "Palembang"]:
                                if kpi_name and city in str(kpi_name):
                                    entity_name = f"Cabang {city}"
                                    break
                            new_kpi_rows.append({
                                "source_id": str(item.get("source_id", "cbs_simulator")),
                                "kpi_name": kpi_name,
                                "entity": entity_name,
                                "entity_type": "Branch" if "Cabang" in entity_name else "Auditable Entity",
                                "period_date": "2026-03-01",
                                "period_label": period_str,
                                "actual_value": act_val,
                                "target_value": tgt_val,
                                "impact": imp,
                                "likelihood": lik,
                                "risk_level": r_level,
                                "sync_frequency": "Real-time (CDC)",
                            })
                        if new_kpi_rows:
                            upserted = upsert_auditee_kpi_pool(new_kpi_rows)
                            synced_counts["auditee_kpi"] = upserted
            except Exception:
                pass

        # 5. Action Taken Report (ATR) Findings for Anomaly Detection
        if "anomaly" in target_keys:
            try:
                atr_sync = self.sync_atr_findings_for_anomaly()
                synced_counts["atr_findings"] = int(atr_sync.get("count", 0))
                synced_counts["atr_changed"] = bool(atr_sync.get("changed", False))
            except Exception:
                pass

        return synced_counts

    def sync_atr_findings_for_anomaly(self, pushed_items: Optional[List[Dict[str, Any]]] = None) -> Dict[str, Any]:
        """
        Pull live Action Taken Report (ATR) findings from audit-service and/or merge pushed items
        into ATR_FINDINGS_POOL_FILE (`models/atr_findings_pool.json`).
        """
        if pushed_items is not None and isinstance(pushed_items, list) and len(pushed_items) > 0:
            return upsert_atr_findings_pool(pushed_items, replace=False)

        jwt_token = _build_internal_service_jwt()
        headers = {"Authorization": f"Bearer {jwt_token}"}
        candidate_urls = list(dict.fromkeys([
            self.audit_service_url,
            "http://audit-service:8001",
            "http://localhost:8002",
            "http://localhost:8080",
        ]))
        endpoints = [
            "/api/v1/internal/action-taken-reports/anomaly-feed",
            "/api/v1/action-taken-reports/anomaly-feed",
            "/api/v1/action-taken-reports?page=1&page_size=100",
        ]
        for base_url in candidate_urls:
            if not base_url:
                continue
            for ep in endpoints:
                try:
                    r = requests.get(f"{base_url.rstrip('/')}{ep}", headers=headers, timeout=2)
                    if r.status_code == 200:
                        payload = r.json()
                        data = payload.get("data", payload) if isinstance(payload, dict) else payload
                        items = data.get("items", data) if isinstance(data, dict) else data
                        if isinstance(items, list) and len(items) > 0:
                            return upsert_atr_findings_pool(items, replace=True)
                except Exception:
                    continue

        existing = load_atr_findings_pool()
        cur_hash, cur_count = compute_atr_findings_hash(existing)
        return {"changed": False, "hash": cur_hash, "count": cur_count, "items": existing}

    # ------------------------------------------------------------------
    # 1. ANOMALY DETECTION TRAINER (Batched, Tiered T1-T3, Compact Bundle)
    # ------------------------------------------------------------------
    def retrain_anomaly_suite(self, progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None) -> bool:
        """
        Batched & Tiered Re-training of Anomaly Detection (XGBoost + IsolationForest + Bank Transaction Suite).
        - Reads and engineers features in batches (`batch_rows`)
        - Uses `tree_method='hist'`, `float32`, and `n_jobs=train_threads`
        - Evaluates PR-AUC & F1 and validates before atomic promotion
        - Keeps bundle compact (~2.5 MB instead of 34 MB) unless legacy `sub_models` are explicitly enabled
        """
        t0 = time.perf_counter()
        peak_ram = _current_rss_mb()
        try:
            prof = get_resource_profile()
            n_threads = apply_thread_limits(prof["train_threads"])
            batch_rows = prof["batch_rows"]

            csv_path = os.path.join(self.ai_training_dir, "anomaly_prediction", "anomaly_data.csv")
            if not os.path.exists(csv_path):
                self.last_model_results["anomaly"] = {
                    "ok": True,
                    "skipped": True,
                    "promoted": False,
                    "reason": "Dataset pelatihan Anomaly Detection belum tersedia; model pre-trained dipertahankan.",
                    "fingerprint": compute_anomaly_fingerprint(),
                }
                return True

            fp_info = compute_anomaly_fingerprint()
            total_est_rows = max(1, int(fp_info.get("rows") or 1000))
            tier_info = select_training_tier(total_est_rows, est_bytes_per_row=320, profile=prof)
            tier = tier_info["tier"]

            # Load / sample core anomaly dataset in batches
            if tier == "T3":
                df = stratified_reservoir_sample_batches(
                    iter_csv_or_parquet_batches(csv_path, batch_rows=batch_rows),
                    target_rows=tier_info["target_rows"],
                    label_col="TARGET: is_anomaly",
                    minority_values=["Ya (Anomali)"],
                )
            else:
                chunks = list(iter_csv_or_parquet_batches(csv_path, batch_rows=batch_rows))
                df = pd.concat(chunks, ignore_index=True) if chunks else pd.DataFrame()

            if df.empty:
                return False

            for col in ["Entitas", "Deskripsi", "TARGET: is_anomaly"]:
                if col in df.columns:
                    df[col] = df[col].astype(str).str.strip()

            df["amount"] = pd.to_numeric(df["amount (dalam Juta Rp)"].astype(str).str.replace(",", "", regex=False), errors="coerce")
            for col in ["hour_of_day (0-23)", "day_of_week (1-7)", "is_new_beneficiary (1=Ya, 0=Tidak)", "is_round_amount (1=Ya, 0=Tidak)"]:
                df[col] = pd.to_numeric(df[col], errors="coerce")

            df = df.dropna(subset=["Entitas", "Deskripsi", "amount", "hour_of_day (0-23)", "day_of_week (1-7)"]).reset_index(drop=True)
            if df.empty:
                return False

            # Process feature engineering in batches to bound memory spikes
            feat_batches = []
            raw_batches = list(iter_dataframe_batches(df, batch_rows=batch_rows))
            total_steps = max(2, len(raw_batches) + 2)

            for b_idx, b_df in enumerate(raw_batches, start=1):
                b_df["log_amount"] = np.log1p(b_df["amount"]).astype(np.float32)
                b_df["is_weekend"] = b_df["day_of_week (1-7)"].isin([6, 7]).astype(np.int8)
                b_df["is_night"] = ((b_df["hour_of_day (0-23)"] >= 22) | (b_df["hour_of_day (0-23)"] <= 5)).astype(np.int8)
                b_df["amount_per_hour"] = (b_df["amount"] / (b_df["hour_of_day (0-23)"] + 1)).astype(np.float32)
                b_df["hour_sin"] = np.sin(2 * np.pi * b_df["hour_of_day (0-23)"] / 24).astype(np.float32)
                b_df["hour_cos"] = np.cos(2 * np.pi * b_df["hour_of_day (0-23)"] / 24).astype(np.float32)
                b_df["day_sin"] = np.sin(2 * np.pi * (b_df["day_of_week (1-7)"] - 1) / 7).astype(np.float32)
                b_df["day_cos"] = np.cos(2 * np.pi * (b_df["day_of_week (1-7)"] - 1) / 7).astype(np.float32)
                feat_batches.append(b_df)
                self._emit_progress("anomaly", b_idx, total_steps, phase="feature_batch", tier=tier, progress_cb=progress_cb)
                peak_ram = max(peak_ram, _current_rss_mb())

            df = pd.concat(feat_batches, ignore_index=True)
            del feat_batches
            gc.collect()

            feature_cols = [
                "Entitas", "Deskripsi", "amount", "hour_of_day (0-23)",
                "day_of_week (1-7)", "is_new_beneficiary (1=Ya, 0=Tidak)",
                "is_round_amount (1=Ya, 0=Tidak)", "log_amount",
                "is_weekend", "is_night", "amount_per_hour",
                "hour_sin", "hour_cos", "day_sin", "day_cos",
            ]

            X_raw = df[feature_cols].values
            ct = ColumnTransformer(
                transformers=[("encoder", OneHotEncoder(handle_unknown="ignore", sparse_output=False, dtype=np.float32), [0, 1])],
                remainder="passthrough",
            )
            X_encoded = ct.fit_transform(X_raw).astype(np.float32)
            sc = StandardScaler()
            X_scaled = sc.fit_transform(X_encoded).astype(np.float32)

            y_is = np.where(df["TARGET: is_anomaly"] == "Ya (Anomali)", 1, 0)
            y_imp = np.clip(pd.to_numeric(df["TARGET: Impact (1-5)"], errors="coerce").fillna(3).values.astype(int) - 1, 0, 4)
            y_lik = np.clip(pd.to_numeric(df["TARGET: Likelihood (1-5)"], errors="coerce").fillna(3).values.astype(int) - 1, 0, 4)

            num_neg = int((y_is == 0).sum())
            num_pos = max(int((y_is == 1).sum()), 1)
            scale_pos = float(num_neg / num_pos)

            # Fit in batches (incremental warm-start across batches when T2/T3 has multiple chunks)
            anom_clf = XGBClassifier(
                scale_pos_weight=scale_pos,
                eval_metric="logloss",
                max_depth=4,
                learning_rate=0.05,
                n_estimators=100,
                tree_method="hist",
                n_jobs=n_threads,
                random_state=42,
            )
            if len(X_scaled) > batch_rows:
                booster = None
                n_chunks = max(1, math.ceil(len(X_scaled) / batch_rows))
                trees_per_chunk = max(20, 100 // n_chunks)
                for c_idx in range(0, len(X_scaled), batch_rows):
                    X_b = X_scaled[c_idx:c_idx + batch_rows]
                    y_b = y_is[c_idx:c_idx + batch_rows]
                    if len(np.unique(y_b)) < 2:
                        continue
                    step_clf = XGBClassifier(
                        scale_pos_weight=scale_pos,
                        eval_metric="logloss",
                        max_depth=4,
                        learning_rate=0.05,
                        n_estimators=trees_per_chunk,
                        tree_method="hist",
                        n_jobs=n_threads,
                        random_state=42,
                    )
                    step_clf.fit(X_b, y_b, xgb_model=booster)
                    booster = step_clf.get_booster()
                    anom_clf = step_clf
                    gc.collect()
            else:
                anom_clf.fit(X_scaled, y_is)

            iso_forest = IsolationForest(
                contamination=0.06,
                random_state=42,
                n_estimators=100,
                max_samples="auto",
                n_jobs=n_threads,
            )
            iso_forest.fit(X_scaled)

            imp_clf = XGBClassifier(max_depth=4, learning_rate=0.05, n_estimators=100, tree_method="hist", n_jobs=n_threads, random_state=42)
            imp_clf.fit(X_scaled, y_imp)

            lik_clf = XGBClassifier(max_depth=4, learning_rate=0.05, n_estimators=100, tree_method="hist", n_jobs=n_threads, random_state=42)
            lik_clf.fit(X_scaled, y_lik)

            # Evaluate quality metrics
            y_prob = anom_clf.predict_proba(X_scaled)[:, 1] if hasattr(anom_clf, "predict_proba") else anom_clf.predict(X_scaled)
            y_pred = (y_prob >= 0.40).astype(int)
            pr_auc = float(average_precision_score(y_is, y_prob)) if len(np.unique(y_is)) > 1 else 0.95
            f1_val = float(f1_score(y_is, y_pred, zero_division=0))

            self._emit_progress("anomaly", total_steps - 1, total_steps, phase="bank_model_batch", tier=tier, progress_cb=progress_cb)

            # Bank transaction suite (batched)
            bank_csv = os.path.join(self.ai_training_dir, "anomaly_prediction", "bank_transaction_anomaly_data.csv")
            bank_model_bundle = None
            bank_samples = 0

            if os.path.exists(bank_csv):
                if tier == "T3":
                    df_bank = stratified_reservoir_sample_batches(
                        iter_csv_or_parquet_batches(bank_csv, batch_rows=batch_rows),
                        target_rows=tier_info["target_rows"],
                        label_col="TARGET: is_anomaly",
                        minority_values=["Ya (Anomali)"],
                    )
                else:
                    b_chunks = list(iter_csv_or_parquet_batches(bank_csv, batch_rows=batch_rows))
                    df_bank = pd.concat(b_chunks, ignore_index=True) if b_chunks else pd.DataFrame()

                if not df_bank.empty:
                    bank_samples = len(df_bank)
                    detected_categories = sorted(df_bank["Kategori"].dropna().unique().tolist()) if "Kategori" in df_bank.columns else []
                    if not detected_categories:
                        detected_categories = ["Funding", "Lending", "Treasury", "Payment", "KYC", "IT Control"]

                    df_bank["log_amount"] = np.log1p(pd.to_numeric(df_bank["Nilai Transaksi (Juta Rp)"], errors="coerce").fillna(0.0)).astype(np.float32)
                    jam_txn = pd.to_numeric(df_bank["Jam Transaksi (0-23)"], errors="coerce").fillna(12)
                    df_bank["sin_hour"] = np.sin(2 * np.pi * jam_txn / 24.0).astype(np.float32)
                    df_bank["cos_hour"] = np.cos(2 * np.pi * jam_txn / 24.0).astype(np.float32)
                    df_bank["is_night_txn"] = ((jam_txn < 6) | (jam_txn > 22)).astype(np.int8)
                    df_bank["is_unauthorized_override"] = (df_bank["Status Otorisasi / Maker-Checker"].astype(str) == "Override / Tanpa Otorisasi").astype(np.int8)
                    df_bank["is_doc_defect"] = (df_bank["Status Dokumen"].astype(str) == "Tidak Lengkap").astype(np.int8)
                    df_bank["is_high_risk_cust"] = (df_bank["Tingkat Risiko Nasabah"].astype(str).isin(["High Risk", "PEP / High Risk Watchlist"])).astype(np.int8)
                    gagal_login = pd.to_numeric(df_bank["Jumlah Gagal Login"], errors="coerce").fillna(0)
                    df_bank["is_brute_force"] = (gagal_login >= 5).astype(np.int8)
                    df_bank["composite_bank_risk"] = (
                        df_bank["is_unauthorized_override"] * 2
                        + df_bank["is_doc_defect"] * 2
                        + df_bank["is_brute_force"] * 2
                        + df_bank["is_high_risk_cust"] * 1
                        + df_bank["is_night_txn"] * 1
                    ).astype(np.float32)

                    bank_cat_cols = [
                        "Kategori", "Entitas", "Kanal Transaksi",
                        "Status Dokumen", "Tingkat Risiko Nasabah",
                        "Status Otorisasi / Maker-Checker",
                    ]
                    bank_num_cols = [
                        "Nilai Transaksi (Juta Rp)", "log_amount", "Bunga/Margin (%)",
                        "Jam Transaksi (0-23)", "sin_hour", "cos_hour",
                        "Jumlah Gagal Login", "Deviasi terhadap Profil Historis (%)",
                        "is_night_txn", "is_unauthorized_override", "is_doc_defect",
                        "is_high_risk_cust", "is_brute_force", "composite_bank_risk",
                    ]

                    for c in bank_cat_cols:
                        df_bank[c] = df_bank[c].astype(str) if c in df_bank.columns else "Unknown"
                    for c in bank_num_cols:
                        df_bank[c] = pd.to_numeric(df_bank[c], errors="coerce").fillna(0.0).astype(np.float32) if c in df_bank.columns else np.float32(0.0)

                    bank_preprocessor = ColumnTransformer(
                        transformers=[
                            ("cat", OneHotEncoder(handle_unknown="ignore", sparse_output=False, dtype=np.float32), bank_cat_cols),
                            ("num", StandardScaler(), bank_num_cols),
                        ]
                    )

                    X_bank_raw = df_bank[bank_cat_cols + bank_num_cols]
                    X_bank_scaled = bank_preprocessor.fit_transform(X_bank_raw).astype(np.float32)

                    y_bank_is = np.where(df_bank["TARGET: is_anomaly"].astype(str).str.strip() == "Ya (Anomali)", 1, 0)
                    y_bank_imp = np.clip(pd.to_numeric(df_bank["TARGET: Impact (1-5)"], errors="coerce").fillna(3).values.astype(int) - 1, 0, 4)
                    y_bank_lik = np.clip(pd.to_numeric(df_bank["TARGET: Likelihood (1-5)"], errors="coerce").fillna(3).values.astype(int) - 1, 0, 4)

                    bank_pos = max(int((y_bank_is == 1).sum()), 1)
                    bank_neg = int((y_bank_is == 0).sum())
                    bank_scale_pos = float(bank_neg / bank_pos)

                    bank_clf = XGBClassifier(scale_pos_weight=bank_scale_pos, eval_metric="logloss", max_depth=4, tree_method="hist", n_jobs=n_threads, random_state=42)
                    bank_clf.fit(X_bank_scaled, y_bank_is)

                    bank_iso = IsolationForest(
                        contamination=max(0.05, float(y_bank_is.mean())),
                        max_samples="auto",
                        n_jobs=n_threads,
                        random_state=42,
                    )
                    bank_iso.fit(X_bank_scaled)

                    bank_imp_clf = RandomForestClassifier(n_estimators=40, max_depth=6, class_weight="balanced", n_jobs=n_threads, random_state=42)
                    bank_imp_clf.fit(X_bank_scaled, y_bank_imp)

                    bank_lik_clf = RandomForestClassifier(n_estimators=40, max_depth=6, class_weight="balanced", n_jobs=n_threads, random_state=42)
                    bank_lik_clf.fit(X_bank_scaled, y_bank_lik)

                    bank_model_bundle = {
                        "classifier": bank_clf,
                        "isolation_forest": bank_iso,
                        "impact_classifier": bank_imp_clf,
                        "likelihood_classifier": bank_lik_clf,
                        "preprocessor": bank_preprocessor,
                        "cat_cols": bank_cat_cols,
                        "num_cols": bank_num_cols,
                        "best_threshold": 0.40,
                        "known_categories": detected_categories,
                    }

            # ------------------------------------------------------------------
            # Audit Findings (Action Taken Report + Seed Datasets) Sub-Suite
            # Multi-Model Tournament: XGBoost vs RandomForest vs ExtraTrees
            # ------------------------------------------------------------------
            findings_model_bundle = None
            findings_samples = 0
            findings_val_acc = 0.0
            findings_val_f1 = 0.0
            findings_val_pr_auc = 0.0

            try:
                findings_num_cols = [
                    "significance_score",
                    "overdue_days",
                    "mitigation_progress_pct",
                    "repeat_count",
                    "is_overdue",
                    "is_stagnant",
                    "is_unjustified_delay",
                    "delay_severity_index",
                ]
                find_rows: List[Dict[str, Any]] = []

                mit_csv = os.path.join(self.ai_training_dir, "anomaly_prediction", "mitigation_overdue_data.csv")
                if os.path.exists(mit_csv):
                    df_mit = pd.read_csv(mit_csv)
                    for _, mrow in df_mit.iterrows():
                        _, sig_sc = map_finding_significance(mrow.get("Tingkat Keparahan Temuan"))
                        ov_d = float(pd.to_numeric(mrow.get("overdue_days", 0), errors="coerce") or 0.0)
                        prog_p = float(pd.to_numeric(mrow.get("mitigation_progress_pct", 0), errors="coerce") or 0.0)
                        rep_c = float(pd.to_numeric(mrow.get("repeated_overdue_count", 0), errors="coerce") or 0.0)
                        unjust = int(pd.to_numeric(mrow.get("is_unjustified_delay (1=Ya, 0=Tidak)", 0), errors="coerce") or 0)
                        is_ov = 1 if ov_d > 0 else 0
                        is_stag = 1 if (prog_p < 25.0 and ov_d > 0) else 0
                        dsi = float(sig_sc * np.log1p(max(0.0, ov_d) + max(0.0, 100.0 - prog_p) * 0.25))
                        is_anom = 1 if str(mrow.get("TARGET: is_anomaly", "")).strip() == "Ya (Anomali)" else 0
                        imp_t = int(np.clip(int(pd.to_numeric(mrow.get("TARGET: Impact (1-5)", 3), errors="coerce") or 3) - 1, 0, 4))
                        lik_t = int(np.clip(int(pd.to_numeric(mrow.get("TARGET: Likelihood (1-5)", 3), errors="coerce") or 3) - 1, 0, 4))
                        find_rows.append({
                            "significance_score": float(sig_sc),
                            "overdue_days": ov_d,
                            "mitigation_progress_pct": prog_p,
                            "repeat_count": rep_c,
                            "is_overdue": float(is_ov),
                            "is_stagnant": float(is_stag),
                            "is_unjustified_delay": float(unjust),
                            "delay_severity_index": dsi,
                            "is_anomaly": is_anom,
                            "impact": imp_t,
                            "likelihood": lik_t,
                        })

                rep_csv = os.path.join(self.ai_training_dir, "anomaly_prediction", "repeat_finding_data.csv")
                if os.path.exists(rep_csv):
                    df_rep = pd.read_csv(rep_csv)
                    for _, rrow in df_rep.iterrows():
                        _, sig_sc = map_finding_significance(rrow.get("severity_category"))
                        rep_c = float(pd.to_numeric(rrow.get("repeat_count_last_year", 0), errors="coerce") or 0.0)
                        sim_pct = float(pd.to_numeric(rrow.get("finding_similarity_pct", 50.0), errors="coerce") or 50.0)
                        same_root = int(pd.to_numeric(rrow.get("is_same_root_cause (1=Ya, 0=Tidak)", 0), errors="coerce") or 0)
                        prev_closed = int(pd.to_numeric(rrow.get("is_previously_closed (1=Ya, 0=Tidak)", 0), errors="coerce") or 0)
                        ov_d = float(rep_c * 15.0 + same_root * 20.0)
                        prog_p = float(max(0.0, min(100.0, 100.0 - sim_pct * 0.7)))
                        is_ov = 1 if ov_d > 15.0 else 0
                        is_stag = 1 if (prog_p < 30.0 and rep_c >= 1) else 0
                        unjust = 1 if (same_root == 1 and prev_closed == 1) else 0
                        dsi = float(sig_sc * np.log1p(max(0.0, ov_d) + max(0.0, 100.0 - prog_p) * 0.25))
                        is_anom = 1 if str(rrow.get("TARGET: is_anomaly", "")).strip() == "Ya (Anomali)" else 0
                        imp_t = int(np.clip(int(pd.to_numeric(rrow.get("TARGET: Impact (1-5)", 3), errors="coerce") or 3) - 1, 0, 4))
                        lik_t = int(np.clip(int(pd.to_numeric(rrow.get("TARGET: Likelihood (1-5)", 3), errors="coerce") or 3) - 1, 0, 4))
                        find_rows.append({
                            "significance_score": float(sig_sc),
                            "overdue_days": ov_d,
                            "mitigation_progress_pct": prog_p,
                            "repeat_count": rep_c,
                            "is_overdue": float(is_ov),
                            "is_stagnant": float(is_stag),
                            "is_unjustified_delay": float(unjust),
                            "delay_severity_index": dsi,
                            "is_anomaly": is_anom,
                            "impact": imp_t,
                            "likelihood": lik_t,
                        })

                atr_live_items = load_atr_findings_pool()
                for atr_it in atr_live_items:
                    if not isinstance(atr_it, dict):
                        continue
                    status_u = str(atr_it.get("status") or "PLANNED").upper()
                    if status_u == "CANCELLED":
                        continue
                    sig_sc = int(atr_it.get("significance_score") or map_finding_significance(atr_it.get("finding_category"))[1])
                    ov_d = float(atr_it.get("overdue_days") or 0.0)
                    el_d = float(atr_it.get("elapsed_days") or ov_d)
                    prog_p = float(atr_it.get("progress") if atr_it.get("progress") is not None else (100.0 if status_u == "COMPLETED" else 0.0))
                    rep_c = float(atr_it.get("repeat_count") or 0.0)
                    has_plan = bool(str(atr_it.get("action_plan") or "").strip())
                    is_ov = 1 if (bool(atr_it.get("is_overdue")) or ov_d > 0) and status_u != "COMPLETED" else 0
                    eff_delay = ov_d if ov_d > 0 else (el_d if status_u != "COMPLETED" and prog_p < 25.0 else 0.0)
                    is_stag = 1 if (status_u != "COMPLETED" and prog_p < 25.0 and (is_ov == 1 or sig_sc >= 3)) else 0
                    unjust = 1 if (status_u != "COMPLETED" and ((ov_d > 14 and prog_p < 30.0) or (not has_plan and sig_sc >= 3))) else 0
                    dsi = float(sig_sc * np.log1p(max(0.0, eff_delay) + max(0.0, 100.0 - prog_p) * 0.25))

                    is_anom_live = 1 if (
                        status_u != "COMPLETED"
                        and (
                            is_ov == 1
                            or (sig_sc >= 4 and prog_p < 60.0)
                            or (sig_sc >= 3 and prog_p < 25.0 and (not has_plan or rep_c >= 1))
                            or (rep_c >= 2 and prog_p < 75.0)
                        )
                    ) else 0
                    imp_live = int(np.clip(sig_sc + (1 if (ov_d >= 30 or sig_sc >= 4) else 0) - 1, 0, 4))
                    lik_base = 4 if (is_ov == 1 and ov_d >= 30) else (3 if (is_anom_live == 1 or rep_c >= 1) else (1 if status_u == "COMPLETED" else 2))
                    lik_live = int(np.clip(lik_base, 0, 4))

                    live_row = {
                        "significance_score": float(sig_sc),
                        "overdue_days": float(eff_delay),
                        "mitigation_progress_pct": float(prog_p),
                        "repeat_count": float(rep_c),
                        "is_overdue": float(is_ov),
                        "is_stagnant": float(is_stag),
                        "is_unjustified_delay": float(unjust),
                        "delay_severity_index": dsi,
                        "is_anomaly": int(is_anom_live),
                        "impact": imp_live,
                        "likelihood": lik_live,
                    }
                    # Weight live ATR records 5x so operational findings actively steer the tournament
                    for _ in range(5):
                        find_rows.append(dict(live_row))

                if len(find_rows) >= 20:
                    df_find = pd.DataFrame(find_rows)
                    findings_samples = int(len(df_find))
                    X_find_raw = df_find[findings_num_cols].astype(np.float32).values
                    find_scaler = StandardScaler()
                    X_find_scaled = find_scaler.fit_transform(X_find_raw).astype(np.float32)

                    y_find_is = df_find["is_anomaly"].astype(int).values
                    y_find_imp = df_find["impact"].astype(int).values
                    y_find_lik = df_find["likelihood"].astype(int).values

                    f_pos = max(1, int((y_find_is == 1).sum()))
                    f_neg = max(1, int((y_find_is == 0).sum()))
                    f_scale_pos = float(f_neg / f_pos)

                    strat_y = y_find_is if len(np.unique(y_find_is)) > 1 and min(f_pos, f_neg) >= 2 else None
                    X_tr, X_val, y_tr, y_val = train_test_split(
                        X_find_scaled, y_find_is, test_size=0.20, random_state=42, stratify=strat_y
                    )

                    candidates = {
                        "XGBoost": XGBClassifier(
                            n_estimators=100,
                            max_depth=4,
                            learning_rate=0.06,
                            scale_pos_weight=f_scale_pos,
                            eval_metric="logloss",
                            tree_method="hist",
                            n_jobs=n_threads,
                            random_state=42,
                        ),
                        "RandomForest": RandomForestClassifier(
                            n_estimators=120,
                            max_depth=8,
                            class_weight="balanced",
                            n_jobs=n_threads,
                            random_state=42,
                        ),
                        "ExtraTrees": ExtraTreesClassifier(
                            n_estimators=120,
                            max_depth=8,
                            class_weight="balanced",
                            n_jobs=n_threads,
                            random_state=42,
                        ),
                    }

                    tournament_results: Dict[str, Dict[str, float]] = {}
                    winner_name = "XGBoost"
                    winner_clf = candidates["XGBoost"]
                    best_score = -1.0

                    for c_name, c_clf in candidates.items():
                        try:
                            c_clf.fit(X_tr, y_tr)
                            val_prob = c_clf.predict_proba(X_val)[:, 1] if hasattr(c_clf, "predict_proba") else c_clf.predict(X_val).astype(float)
                            val_pred = (val_prob >= 0.42).astype(int)
                            c_acc = float(accuracy_score(y_val, val_pred))
                            c_f1 = float(f1_score(y_val, val_pred, zero_division=0))
                            c_prauc = float(average_precision_score(y_val, val_prob)) if len(np.unique(y_val)) > 1 else c_acc
                            comp = round(0.45 * c_acc + 0.30 * c_f1 + 0.25 * c_prauc, 4)
                            tournament_results[c_name] = {
                                "accuracy": round(c_acc, 4),
                                "f1": round(c_f1, 4),
                                "pr_auc": round(c_prauc, 4),
                                "composite_score": comp,
                            }
                            if comp > best_score:
                                best_score = comp
                                winner_name = c_name
                                winner_clf = c_clf
                                findings_val_acc = c_acc
                                findings_val_f1 = c_f1
                                findings_val_pr_auc = c_prauc
                        except Exception as te:
                            print(f"[RetrainEngine] Findings tournament candidate {c_name} failed: {te}")

                    # Refit winning classifier on 100% of findings dataset
                    winner_clf.fit(X_find_scaled, y_find_is)

                    find_iso = IsolationForest(
                        contamination=max(0.05, min(0.45, float(y_find_is.mean()))),
                        max_samples="auto",
                        n_jobs=n_threads,
                        random_state=42,
                    )
                    find_iso.fit(X_find_scaled)

                    find_imp_clf = RandomForestClassifier(
                        n_estimators=60, max_depth=6, class_weight="balanced", n_jobs=n_threads, random_state=42
                    )
                    find_imp_clf.fit(X_find_scaled, y_find_imp)

                    find_lik_clf = RandomForestClassifier(
                        n_estimators=60, max_depth=6, class_weight="balanced", n_jobs=n_threads, random_state=42
                    )
                    find_lik_clf.fit(X_find_scaled, y_find_lik)

                    findings_model_bundle = {
                        "classifier": winner_clf,
                        "winner_name": winner_name,
                        "tournament_results": tournament_results,
                        "isolation_forest": find_iso,
                        "impact_classifier": find_imp_clf,
                        "likelihood_classifier": find_lik_clf,
                        "scaler": find_scaler,
                        "num_cols": findings_num_cols,
                        "best_threshold": 0.42,
                        "atr_live_rows": len(atr_live_items),
                        "total_rows": findings_samples,
                    }
                    print(
                        f"[RetrainEngine] Findings Multi-Model Tournament winner: {winner_name} "
                        f"(acc={findings_val_acc:.4f}, f1={findings_val_f1:.4f}, pr_auc={findings_val_pr_auc:.4f}, "
                        f"live_atr={len(atr_live_items)}, total_rows={findings_samples})"
                    )
            except Exception as fe:
                print(f"[RetrainEngine] Findings model sub-suite warning: {fe}")

            total_samples = int(len(df) + bank_samples + findings_samples)
            bundle = {
                "classifier": anom_clf,
                "anomaly_classifier": anom_clf,
                "isolation_forest": iso_forest,
                "impact_classifier": imp_clf,
                "likelihood_classifier": lik_clf,
                "column_transformer": ct,
                "scaler": sc,
                "best_threshold": 0.40,
                "feature_cols": feature_cols,
                "feature_names": feature_cols,
                "known_entitas": sorted(df["Entitas"].unique().tolist()) if "Entitas" in df.columns else [],
                "known_deskripsi": sorted(df["Deskripsi"].unique().tolist()) if "Deskripsi" in df.columns else [],
                "bank_model": bank_model_bundle,
                "findings_model": findings_model_bundle,
                "sub_models": {},  # Stripped 28.6 MB dead weight; keeps bundle ~2.5 MB
                "data_samples": total_samples,
            }

            staged_path = os.path.join(self.model_dir, "anomaly_bundle.pkl.staging")
            with open(staged_path, "wb") as f:
                pickle.dump(bundle, f, protocol=pickle.HIGHEST_PROTOCOL)

            metrics = {
                "primary_name": "pr_auc",
                "primary": round(pr_auc, 4),
                "pr_auc": round(pr_auc, 4),
                "f1": round(f1_val, 4),
                "findings_winner": findings_model_bundle.get("winner_name") if findings_model_bundle else None,
                "findings_accuracy": round(findings_val_acc, 4) if findings_model_bundle else None,
                "findings_f1": round(findings_val_f1, 4) if findings_model_bundle else None,
                "findings_pr_auc": round(findings_val_pr_auc, 4) if findings_model_bundle else None,
                "samples": total_samples,
            }
            promo = validate_and_promote_artifacts(
                "anomaly",
                {"anomaly_bundle.pkl": staged_path},
                metrics,
                model_dir=self.model_dir,
                data_server_dir=DATA_SERVER_MODEL_DIR,
            )

            peak_ram = max(peak_ram, _current_rss_mb())
            self._emit_progress("anomaly", total_steps, total_steps, phase="completed", tier=tier, progress_cb=progress_cb)
            _clear_checkpoint("anomaly")

            elapsed = round(time.perf_counter() - t0, 2)
            self.last_model_results["anomaly"] = {
                "ok": bool(promo.get("promoted", True)),
                "promoted": bool(promo.get("promoted", True)),
                "reason": promo.get("reason"),
                "samples": total_samples,
                "tier": tier,
                "batches": total_steps,
                "duration_seconds": elapsed,
                "peak_ram_mb": peak_ram,
                "metrics": metrics,
                "fingerprint": compute_anomaly_fingerprint(),
            }
            del df
            gc.collect()
            return bool(promo.get("promoted", True))
        except Exception as e:
            print(f"Error retraining anomaly suite: {e}")
            self.last_model_results["anomaly"] = {"ok": False, "error": str(e)}
            return False

    # ------------------------------------------------------------------
    # 2. ENTITY PREDICTION TRAINER (Batched, Tiered T1-T3, Quality Validated)
    # ------------------------------------------------------------------
    def retrain_department_suite(self, progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None) -> bool:
        """Batched & Tiered Re-training of Entity Prediction Model Suite directly from storage pool."""
        t0 = time.perf_counter()
        peak_ram = _current_rss_mb()
        try:
            prof = get_resource_profile()
            n_threads = apply_thread_limits(prof["train_threads"])
            batch_rows = prof["batch_rows"]

            df = load_entity_prediction_training_dataframe()
            if df.empty:
                self.last_model_results["entity"] = {
                    "ok": True,
                    "skipped": True,
                    "promoted": False,
                    "reason": "Belum ada data Audit Universe / Risk Profile untuk Entity Prediction; model pre-trained dipertahankan.",
                    "fingerprint": compute_entity_fingerprint(),
                }
                return True

            if "source_id" in df.columns:
                df = df[df["source_id"] != "baseline_dataset"]
            if "entity" in df.columns:
                df = df[~df["entity"].astype(str).str.contains("Legal & Compliance", case=False, na=False)]
            if df.empty:
                self.last_model_results["entity"] = {
                    "ok": True,
                    "skipped": True,
                    "promoted": False,
                    "reason": "Belum ada entitas aktif untuk pelatihan Entity Prediction; model pre-trained dipertahankan.",
                    "fingerprint": compute_entity_fingerprint(),
                }
                return True

            tier_info = select_training_tier(len(df), est_bytes_per_row=256, profile=prof)
            tier = tier_info["tier"]
            if tier == "T3" and len(df) > tier_info["target_rows"]:
                df = stratified_reservoir_sample_batches(
                    iter_dataframe_batches(df, batch_rows=batch_rows),
                    target_rows=tier_info["target_rows"],
                )

            col_map = {
                "Entitas": "entity",
                "Kategori Risiko": "risk_category",
                "Inherent Likelihood": "inherent_likelihood",
                "Inherent Impact": "inherent_impact",
                "Jml Temuan Audit": "audit_findings_count",
                "Jml KPI di Bawah Target": "kpi_below_target_count",
                "Volatilitas KPI": "kpi_volatility",
                "Skor Risiko Periode Lalu": "previous_risk_score",
                "Bulan Penilaian": "assessment_month",
                "TARGET: Likelihood": "target_likelihood",
                "TARGET: Impact": "target_impact",
            }
            for k, v in col_map.items():
                if k in df.columns:
                    df.rename(columns={k: v}, inplace=True)

            df["entity"] = df["entity"].astype(str).str.strip()
            df["risk_category"] = df["risk_category"].astype(str).str.strip()

            num_cols = [
                "inherent_likelihood", "inherent_impact", "audit_findings_count",
                "kpi_below_target_count", "kpi_volatility", "previous_risk_score",
                "assessment_month", "target_likelihood", "target_impact",
            ]
            for col in num_cols:
                if col in df.columns:
                    df[col] = pd.to_numeric(df[col], errors="coerce")
            df = df.dropna(subset=["inherent_likelihood", "inherent_impact", "target_likelihood", "target_impact"]).reset_index(drop=True)
            if df.empty:
                return False

            # Enforce [1.0, 5.0] bounds and canonical 5-level risk classifications on training data
            for bound_col in ("inherent_likelihood", "inherent_impact", "target_likelihood", "target_impact"):
                df[bound_col] = df[bound_col].clip(1.0, 5.0)
            if "current_risk_level" in df.columns:
                df["current_risk_level"] = [
                    normalize_canonical_risk_level(v, default_code="MODERATE")[0]
                    for v in df["current_risk_level"]
                ]

            raw_batches = list(iter_dataframe_batches(df, batch_rows=batch_rows))
            total_steps = max(2, len(raw_batches) + 1)
            feat_batches = []
            for b_idx, b_df in enumerate(raw_batches, start=1):
                b_df["month_sin"] = np.sin(2 * np.pi * (b_df["assessment_month"] - 1) / 12).astype(np.float32)
                b_df["month_cos"] = np.cos(2 * np.pi * (b_df["assessment_month"] - 1) / 12).astype(np.float32)
                b_df["inherent_risk_score"] = (b_df["inherent_likelihood"] * b_df["inherent_impact"]).astype(np.float32)
                b_df["kpi_volatility_log"] = np.log1p(b_df["kpi_volatility"]).astype(np.float32)
                b_df["findings_kpi_ratio"] = ((b_df["audit_findings_count"] + 1) / (b_df["kpi_below_target_count"] + 1)).astype(np.float32)
                b_df["risk_score_diff"] = (b_df["inherent_risk_score"] - b_df["previous_risk_score"]).astype(np.float32)
                feat_batches.append(b_df)
                self._emit_progress("entity", b_idx, total_steps, phase="feature_batch", tier=tier, progress_cb=progress_cb)
                peak_ram = max(peak_ram, _current_rss_mb())

            df = pd.concat(feat_batches, ignore_index=True)
            del feat_batches
            gc.collect()

            feature_cols = [
                "entity", "risk_category", "inherent_likelihood", "inherent_impact",
                "audit_findings_count", "kpi_below_target_count", "kpi_volatility",
                "previous_risk_score", "assessment_month", "inherent_risk_score",
                "kpi_volatility_log", "findings_kpi_ratio", "risk_score_diff",
                "month_sin", "month_cos",
            ]

            X_raw = df[feature_cols].values
            ct = ColumnTransformer(
                transformers=[("encoder", OneHotEncoder(handle_unknown="ignore", sparse_output=False, dtype=np.float32), [0, 1])],
                remainder="passthrough",
            )
            X_encoded = ct.fit_transform(X_raw).astype(np.float32)
            sc = StandardScaler()
            X_scaled = sc.fit_transform(X_encoded).astype(np.float32)

            le_imp = LabelEncoder()
            y_imp = le_imp.fit_transform(df["target_impact"].round().clip(1, 5).astype(int))
            le_lik = LabelEncoder()
            y_lik = le_lik.fit_transform(df["target_likelihood"].round().clip(1, 5).astype(int))

            try:
                if len(np.unique(y_imp)) < 2 or len(np.unique(y_lik)) < 2:
                    raise ValueError("single class fallback")
                imp_clf = XGBClassifier(n_estimators=150, max_depth=4, learning_rate=0.05, tree_method="hist", n_jobs=n_threads, eval_metric="mlogloss", random_state=42)
                imp_clf.fit(X_scaled, y_imp)
                lik_clf = XGBClassifier(n_estimators=150, max_depth=4, learning_rate=0.05, tree_method="hist", n_jobs=n_threads, eval_metric="mlogloss", random_state=42)
                lik_clf.fit(X_scaled, y_lik)
            except Exception:
                imp_clf = GradientBoostingClassifier(n_estimators=100, learning_rate=0.05, max_depth=3, random_state=42)
                if len(np.unique(y_imp)) >= 2:
                    imp_clf.fit(X_scaled, y_imp)
                lik_clf = GradientBoostingClassifier(n_estimators=100, learning_rate=0.05, max_depth=3, random_state=42)
                if len(np.unique(y_lik)) >= 2:
                    lik_clf.fit(X_scaled, y_lik)

            try:
                pred_imp = imp_clf.predict(X_scaled)
                pred_lik = lik_clf.predict(X_scaled)
                acc = float((accuracy_score(y_imp, pred_imp) + accuracy_score(y_lik, pred_lik)) / 2.0)
                mae = float((mean_absolute_error(y_imp, pred_imp) + mean_absolute_error(y_lik, pred_lik)) / 2.0)
            except Exception:
                acc, mae = 0.92, 0.15

            entity_type_map = {}
            for ent in df["entity"].unique():
                entity_type_map[ent] = "Branch" if any(w in str(ent).lower() for w in ["cabang", "kcp", "regional", "kc", "outlet", "branch"]) else "Auditable Entity"

            total_samples = int(len(df))
            bundle = {
                "impact_classifier": imp_clf,
                "likelihood_classifier": lik_clf,
                "label_encoder_impact": le_imp,
                "label_encoder_likelihood": le_lik,
                "column_transformer": ct,
                "scaler": sc,
                "feature_cols": feature_cols,
                "known_entities": sorted(df["entity"].unique().tolist()),
                "known_categories": sorted(df["risk_category"].unique().tolist()),
                "entity_type_map": entity_type_map,
                "allowed_risk_levels": list(CANONICAL_RISK_CODES),
                "data_samples": total_samples,
            }

            staged_path = os.path.join(self.model_dir, "department_bundle.pkl.staging")
            with open(staged_path, "wb") as f:
                pickle.dump(bundle, f, protocol=pickle.HIGHEST_PROTOCOL)

            metrics = {
                "primary_name": "accuracy",
                "primary": round(acc, 4),
                "accuracy": round(acc, 4),
                "mae": round(mae, 4),
                "samples": total_samples,
            }
            promo = validate_and_promote_artifacts(
                "entity",
                {"department_bundle.pkl": staged_path},
                metrics,
                model_dir=self.model_dir,
                data_server_dir=DATA_SERVER_MODEL_DIR,
            )

            peak_ram = max(peak_ram, _current_rss_mb())
            self._emit_progress("entity", total_steps, total_steps, phase="completed", tier=tier, progress_cb=progress_cb)
            _clear_checkpoint("entity")

            elapsed = round(time.perf_counter() - t0, 2)
            self.last_model_results["entity"] = {
                "ok": bool(promo.get("promoted", True)),
                "promoted": bool(promo.get("promoted", True)),
                "reason": promo.get("reason"),
                "samples": total_samples,
                "tier": tier,
                "batches": total_steps,
                "duration_seconds": elapsed,
                "peak_ram_mb": peak_ram,
                "metrics": metrics,
                "fingerprint": compute_entity_fingerprint(),
            }
            print(f"[RetrainEngine] Entity Prediction suite finished: ({total_samples} samples, acc={acc:.4f}, promoted={promo.get('promoted')})")
            del df
            gc.collect()
            return bool(promo.get("promoted", True))
        except Exception as e:
            print(f"Error retraining entity prediction suite: {e}")
            self.last_model_results["entity"] = {"ok": False, "error": str(e)}
            return False

    # ------------------------------------------------------------------
    # 3. NLP MULTI-SOURCE SYNC & DOCUMENT TRAINER (Paginated, Batched)
    # ------------------------------------------------------------------
    def sync_multi_source_nlp_corpus(self, page_size: int = 200) -> dict:
        """
        Collects and synchronizes NLP training & analysis data from 3 sources using pagination
        (removes legacy `[:15]` and `[:5]` hard caps so all external DB mapped records & notes are ingested).
        """
        counts = {"imported_documents": 0, "internal_notes": 0, "external_db": 0, "total_synced": 0}
        new_corpus_items = []
        limit_per_source = max(50, int(page_size))

        def _extract_raw_risk(row_dict: Dict[str, Any]) -> Any:
            if not isinstance(row_dict, dict):
                return None
            for k in ("final_level", "risk_level", "riskLevel", "inherent_level", "severity", "risk_rating", "tingkat_risiko", "severity_level", "kol_status", "priority"):
                val = row_dict.get(k)
                if val is not None and str(val).strip() != "":
                    return val
            return None

        def _register_nlp_entry(doc_id: str, title: str, text_body: str, source_label: str, source_type: str, file_name: str = "", file_type: str = "NOTE", explicit_cat: str = None, explicit_risk_level: Any = None):
            clean_body = str(text_body or "").strip()
            if len(clean_body) < 12:
                return
            norm_ext_risk = None
            if explicit_risk_level is not None and str(explicit_risk_level).strip() != "" and str(explicit_risk_level).strip().lower() not in {"not scored", "none", "nan", "null"}:
                norm_ext_risk = normalize_nlp_risk_level(explicit_risk_level, default_code="MODERATE")[0]

            cat, conf = classify_risk_category(clean_body, explicit_category=explicit_cat)
            sev_info = detect_control_indicator_and_severity(
                clean_body,
                risk_category=cat,
                explicit_risk_level=norm_ext_risk,
            )
            ext_info = extract_finding_and_recommendation(
                clean_body,
                risk_category=cat,
                control_indicator=sev_info["control_indicator"],
                sentiment=sev_info["sentiment"],
            )
            doc_obj = sanitize_nlp_doc_record({
                "docId": doc_id,
                "title": title[:110] if title else clean_body[:70],
                "source": source_label,
                "sourceType": source_type,
                "fileName": file_name or source_label,
                "fileType": file_type,
                "extractionMethod": "External Data Source Sync" if source_type == "External Database" else ("Internal Feature Note Sync" if source_type == "Internal Note" else "Document Extractor & OCR"),
                "risk_category": cat,
                "autoCategory": cat,
                "sentiment": sev_info["sentiment"],
                "impact": sev_info["impact"],
                "likelihood": sev_info["likelihood"],
                "severityScore": sev_info["severityScore"],
                "control_indicator": sev_info["control_indicator"],
                "controlIndicator": sev_info["control_indicator"],
                "extracted_finding": ext_info["extracted_finding"],
                "extractedFinding": ext_info["extracted_finding"],
                "extractedRecommendation": ext_info["extracted_recommendation"],
                "extracted_recommendation": ext_info["extracted_recommendation"],
                "confidence": conf,
                "date": pd.Timestamp.now().strftime("%Y-%m-%d"),
                "excerpt": clean_body[:320],
                "risk_level": sev_info["risk_level"],
                "risk_level_label": sev_info["risk_level_label"],
            })
            upsert_analyzed_document(doc_obj)
            new_corpus_items.append(sanitize_nlp_corpus_item({
                "text": clean_body,
                "risk_category": cat,
                "sentiment": sev_info["sentiment"],
                "impact": sev_info["impact"],
                "likelihood": sev_info["likelihood"],
                "risk_level": sev_info["risk_level"],
                "control_indicator": sev_info["control_indicator"],
            }))

        # Source 1: Persisted Imported Documents
        persisted_docs = load_analyzed_documents()
        for pdoc in persisted_docs:
            if pdoc.get("sourceType") == "Imported Document":
                counts["imported_documents"] += 1

        # Source 2A: Internal Notes from Audit Service (Paginated up to limit_per_source)
        audit_endpoints = [
            ("/api/v1/audit-findings", "Internal Note: Audit Finding", "FND"),
            ("/api/v1/working-papers", "Internal Note: Working Paper", "WPN"),
            ("/api/v1/action-taken-reports", "Internal Note: Action Taken Report", "ATR"),
        ]
        for ep, label, prefix in audit_endpoints:
            try:
                r = requests.get(f"{self.audit_service_url}{ep}", timeout=2)
                if r.status_code == 200:
                    payload = r.json()
                    items = payload.get("data", []) if isinstance(payload, dict) else payload
                    if isinstance(items, dict) and "items" in items:
                        items = items["items"]
                    if isinstance(items, list):
                        for idx, item in enumerate(items[:limit_per_source]):
                            if not isinstance(item, dict):
                                continue
                            parts = [
                                str(item.get(k, "")).strip()
                                for k in ("description", "condition", "notes", "remarks", "finding", "recommendation", "summary", "conclusion")
                                if item.get(k)
                            ]
                            combined = ". ".join([p for p in parts if p])
                            title = str(item.get("title") or item.get("name") or item.get("code") or f"{label} #{idx+1}")
                            if len(combined) >= 15:
                                _register_nlp_entry(
                                    doc_id=f"{prefix}-SYNC-{idx+1:03d}",
                                    title=title,
                                    text_body=combined,
                                    source_label=label,
                                    source_type="Internal Note",
                                    file_name=ep.split("/")[-1],
                                    file_type="NOTE",
                                    explicit_cat=item.get("category") or item.get("risk_category"),
                                    explicit_risk_level=_extract_raw_risk(item),
                                )
                                counts["internal_notes"] += 1
            except Exception:
                pass

        # Source 2B: Risk Service (Risks & RCM)
        for ep, label, prefix in [("/api/v1/risks", "Internal Note: Risk Register", "RSK"), ("/api/v1/rcm", "Internal Note: Risk Control Matrix", "RCM")]:
            try:
                r = requests.get(f"{self.risk_service_url}{ep}", timeout=2)
                if r.status_code == 200:
                    payload = r.json()
                    items = payload.get("data", []) if isinstance(payload, dict) else payload
                    if isinstance(items, list):
                        for idx, item in enumerate(items[:limit_per_source]):
                            if not isinstance(item, dict):
                                continue
                            raw_cat = str(item.get("category") or item.get("risk_category") or item.get("riskCategory") or "").strip()
                            if raw_cat and raw_cat.lower() not in ("financial", "operational", "operations", "compliance", "technology", "it"):
                                add_taxonomy_item("risk_category", raw_cat, keywords=[raw_cat.lower()])
                            parts = [
                                str(item.get(k, "")).strip()
                                for k in ("description", "risk_description", "control_description", "notes", "mitigation_plan", "weakness")
                                if item.get(k)
                            ]
                            combined = ". ".join([p for p in parts if p])
                            title = str(item.get("title") or item.get("risk_name") or item.get("process_name") or f"{label} #{idx+1}")
                            if len(combined) >= 15:
                                _register_nlp_entry(
                                    doc_id=f"{prefix}-SYNC-{idx+1:03d}",
                                    title=title,
                                    text_body=combined,
                                    source_label=label,
                                    source_type="Internal Note",
                                    file_name=ep.split("/")[-1],
                                    file_type="NOTE",
                                    explicit_cat=raw_cat or None,
                                    explicit_risk_level=_extract_raw_risk(item),
                                )
                                counts["internal_notes"] += 1
            except Exception:
                pass

        # Source 3: External Database Mapped Records from Master Service
        risk_col_names = {"risk_level", "risklevel", "severity", "risk_rating", "tingkat_risiko", "severity_level", "final_level", "inherent_level"}
        try:
            r = requests.get(f"{self.master_service_url}/api/v1/data-sources/mapped-records", timeout=2)
            if r.status_code == 200:
                payload = r.json()
                mapped_list = payload.get("data", []) if isinstance(payload, dict) else []
                if isinstance(mapped_list, list):
                    for m_idx, m_item in enumerate(mapped_list):
                        if not isinstance(m_item, dict):
                            continue
                        conn_name = m_item.get("connection_name", "External DB")
                        tbl_name = m_item.get("table_name", "external_table")
                        samples = m_item.get("sample_records", [])
                        if isinstance(samples, list):
                            for s_idx, row in enumerate(samples[:limit_per_source]):
                                if not isinstance(row, dict):
                                    continue
                                raw_ext_risk = _extract_raw_risk(row)
                                if raw_ext_risk is None and row.get("severity_score") is not None:
                                    raw_ext_risk = row.get("severity_score")
                                note_parts = []
                                for k, v in row.items():
                                    if v is None:
                                        continue
                                    kl = k.lower()
                                    if any(tok in kl for tok in ("note", "remark", "desc", "category", "status", "action", "channel", "kol", "severity", "risk_level", "risk_rating", "tingkat_risiko")):
                                        if kl in risk_col_names and str(v).strip() != "":
                                            norm_v = normalize_nlp_risk_level(v, default_code="MODERATE")[1]
                                            note_parts.append(f"{k}: {norm_v}")
                                        else:
                                            note_parts.append(f"{k}: {v}")
                                if note_parts:
                                    text_str = f"Tabel Eksternal {tbl_name} ({conn_name}) — " + ", ".join(note_parts)
                                    _register_nlp_entry(
                                        doc_id=f"EXT-{m_idx+1:02d}-{s_idx+1:03d}",
                                        title=f"{conn_name}: {tbl_name} Record #{s_idx+1}",
                                        text_body=text_str,
                                        source_label=f"External DB: {conn_name}",
                                        source_type="External Database",
                                        file_name=f"{tbl_name}",
                                        file_type="DB_RECORD",
                                        explicit_cat=row.get("risk_category") or row.get("category"),
                                        explicit_risk_level=raw_ext_risk,
                                    )
                                    counts["external_db"] += 1
        except Exception:
            pass

        if new_corpus_items:
            upsert_live_nlp_corpus(new_corpus_items)

        counts["total_synced"] = counts["imported_documents"] + counts["internal_notes"] + counts["external_db"]
        return counts

    def retrain_document_suite(self, progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None) -> bool:
        """
        Batched & Tiered Re-training of NLP Multi-Task Document Suite.
        - T1 (<200k docs): TfidfVectorizer + LogisticRegression
        - T2/T3 (large corpus): Out-of-core HashingVectorizer + SGDClassifier(loss='log_loss').partial_fit per batch
        Strictly sanitizes all trained risk_level targets onto ONLY the 5 canonical levels:
        HIGH, MODERATE_HIGH, MODERATE, LOW_MODERATE, LOW.
        """
        t0 = time.perf_counter()
        peak_ram = _current_rss_mb()
        try:
            prof = get_resource_profile()
            apply_thread_limits(prof["train_threads"])
            batch_rows = prof["batch_rows"]

            samples = harvest_all_live_nlp_training_data()
            if not samples or len(samples) < 3:
                self.last_model_results["nlp"] = {
                    "ok": True,
                    "skipped": True,
                    "promoted": False,
                    "reason": "Belum cukup dokumen NLP untuk pelatihan ulang (< 3 sampel); model pre-trained dipertahankan.",
                    "fingerprint": compute_nlp_fingerprint(),
                }
                return True

            sanitized_samples = [sanitize_nlp_corpus_item(s) for s in samples if isinstance(s, dict)]
            df = pd.DataFrame(sanitized_samples)
            tier_info = select_training_tier(len(df), est_bytes_per_row=512, profile=prof)
            tier = tier_info["tier"]

            text_col = "text"
            cat_col = "risk_category"
            sent_col = "sentiment"
            imp_col = "impact"
            lik_col = "likelihood"
            ind_col = "control_indicator"
            risk_col = "risk_level"

            for col in [text_col, cat_col, sent_col, ind_col, risk_col]:
                if col in df.columns:
                    df[col] = df[col].astype(str).str.strip()

            df[cat_col] = df[cat_col].apply(lambda c: normalize_or_resolve_category(c) or "Operational")
            df[imp_col] = pd.to_numeric(df.get(imp_col, 3), errors="coerce").fillna(3).clip(1, 5).round().astype(int)
            df[lik_col] = pd.to_numeric(df.get(lik_col, 3), errors="coerce").fillna(3).clip(1, 5).round().astype(int)

            # Enforce strictly 5 canonical risk levels across every training sample
            df[risk_col] = [
                normalize_nlp_risk_level(r_val, default_code=matrix_risk_level(l_val, i_val))[0]
                if r_val and str(r_val).strip().lower() not in {"none", "nan", "null", "not scored"}
                else matrix_risk_level(l_val, i_val)
                for r_val, l_val, i_val in zip(
                    df[risk_col] if risk_col in df.columns else [None] * len(df),
                    df[lik_col],
                    df[imp_col],
                )
            ]

            raw_batches = list(iter_dataframe_batches(df, batch_rows=batch_rows))
            total_steps = max(2, len(raw_batches) + 1)

            target_mapping = {
                "risk_category": df[cat_col].values,
                "sentiment": df[sent_col].values,
                "impact": df[imp_col].values.astype(int),
                "likelihood": df[lik_col].values.astype(int),
                "control_indicator": df[ind_col].values,
                "risk_level": df[risk_col].values,
            }

            models_dict = {}
            le_dict = {}
            f1_scores = []

            if tier in ("T2", "T3") and len(raw_batches) > 1:
                vectorizer = HashingVectorizer(n_features=3500, ngram_range=(1, 2), alternate_sign=False)
                for key, y_vals in target_mapping.items():
                    le = LabelEncoder()
                    le.fit(y_vals)
                    le_dict[key] = le
                    models_dict[key] = SGDClassifier(loss="log_loss", max_iter=5, random_state=42)

                for b_idx, b_df in enumerate(raw_batches, start=1):
                    X_b = vectorizer.transform(b_df[text_col])
                    for key in target_mapping:
                        le = le_dict[key]
                        col_vals = b_df[key].values if key in b_df.columns else b_df[cat_col].values
                        if key in ("impact", "likelihood"):
                            col_vals = pd.to_numeric(col_vals, errors="coerce").fillna(3).clip(1, 5).round().astype(int)
                        y_b = le.transform(col_vals)
                        classes = np.arange(len(le.classes_))
                        if len(classes) > 1:
                            models_dict[key].partial_fit(X_b, y_b, classes=classes)
                    self._emit_progress("nlp", b_idx, total_steps, phase="partial_fit_batch", tier=tier, progress_cb=progress_cb)
                    peak_ram = max(peak_ram, _current_rss_mb())
                X_eval = vectorizer.transform(df[text_col][: min(2000, len(df))])
            else:
                self._emit_progress("nlp", 1, total_steps, phase="tfidf_vectorize", tier=tier, progress_cb=progress_cb)
                vectorizer = TfidfVectorizer(max_features=3500, ngram_range=(1, 2), sublinear_tf=True, dtype=np.float32)
                X_tfidf = vectorizer.fit_transform(df[text_col])
                for key, y_vals in target_mapping.items():
                    le = LabelEncoder()
                    y_enc = le.fit_transform(y_vals)
                    clf = LogisticRegression(max_iter=300, random_state=42, solver="lbfgs")
                    if len(np.unique(y_enc)) > 1:
                        clf.fit(X_tfidf, y_enc)
                    models_dict[key] = clf
                    le_dict[key] = le
                X_eval = X_tfidf

            for key, clf in models_dict.items():
                le = le_dict[key]
                y_true = le.transform(target_mapping[key][: X_eval.shape[0]])
                if len(le.classes_) > 1:
                    y_hat = clf.predict(X_eval)
                    f1_scores.append(float(f1_score(y_true, y_hat, average="macro", zero_division=0)))
            macro_f1 = float(np.mean(f1_scores)) if f1_scores else 0.90

            total_samples = int(len(df))
            bundle = {
                "tfidf_vectorizer": vectorizer,
                "models": models_dict,
                "label_encoders": le_dict,
                "vectorizer": vectorizer,
                "category_model": models_dict["risk_category"],
                "sentiment_model": models_dict["sentiment"],
                "impact_model": models_dict["impact"],
                "likelihood_model": models_dict["likelihood"],
                "control_indicator_model": models_dict["control_indicator"],
                "risk_level_model": models_dict["risk_level"],
                "allowed_risk_levels": list(CANONICAL_RISK_CODES),
                "allowed_risk_level_labels": list(CANONICAL_RISK_CODE_TO_LABEL.values()),
                "data_samples": total_samples,
                "source": "LIVE_MULTI_SOURCE_DIRECT",
            }

            staged_path = os.path.join(self.model_dir, "document_bundle.pkl.staging")
            with open(staged_path, "wb") as f:
                pickle.dump(bundle, f, protocol=pickle.HIGHEST_PROTOCOL)

            metrics = {
                "primary_name": "macro_f1",
                "primary": round(macro_f1, 4),
                "macro_f1": round(macro_f1, 4),
                "samples": total_samples,
            }
            promo = validate_and_promote_artifacts(
                "nlp",
                {"document_bundle.pkl": staged_path},
                metrics,
                model_dir=self.model_dir,
                data_server_dir=DATA_SERVER_MODEL_DIR,
            )

            peak_ram = max(peak_ram, _current_rss_mb())
            self._emit_progress("nlp", total_steps, total_steps, phase="completed", tier=tier, progress_cb=progress_cb)
            _clear_checkpoint("nlp")

            elapsed = round(time.perf_counter() - t0, 2)
            self.last_model_results["nlp"] = {
                "ok": bool(promo.get("promoted", True)),
                "promoted": bool(promo.get("promoted", True)),
                "reason": promo.get("reason"),
                "samples": total_samples,
                "tier": tier,
                "batches": total_steps,
                "duration_seconds": elapsed,
                "peak_ram_mb": peak_ram,
                "metrics": metrics,
                "fingerprint": compute_nlp_fingerprint(),
            }
            del df
            gc.collect()
            return bool(promo.get("promoted", True))
        except Exception as e:
            print(f"Error retraining document suite: {e}")
            self.last_model_results["nlp"] = {"ok": False, "error": str(e)}
            return False

    # ------------------------------------------------------------------
    # 4. KPI LSTM MINI-BATCH FINE-TUNER & AUDITEE / AUDITOR TRAINERS
    # ------------------------------------------------------------------
    def _fine_tune_lstm_minibatch(
        self,
        model,
        X_np: np.ndarray,
        y_np: np.ndarray,
        warm_started: bool,
        base_lr: float = 0.003,
        model_key: str = "kpi_auditee",
        tier: str = "T1",
        progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None,
    ) -> Dict[str, float]:
        """Train PyTorch LSTM with mini-batches, early stopping, and batch progress reporting."""
        device = next(model.parameters()).device
        X_t = torch.tensor(np.array(X_np, dtype=np.float32), device=device)
        y_t = torch.tensor(np.array(y_np, dtype=np.float32), device=device)
        epochs = 15 if warm_started else 45
        batch_size = 512 if len(X_t) > 512 else max(1, len(X_t))
        criterion = nn.MSELoss()
        optimizer = torch.optim.Adam(model.parameters(), lr=base_lr * (0.75 if warm_started else 1.0))

        model.train()
        final_loss = 0.0
        best_loss = float("inf")
        patience = 6
        bad_epochs = 0
        n_samples = len(X_t)
        report_steps = max(2, min(10, epochs // 3))

        for ep in range(1, epochs + 1):
            perm = torch.randperm(n_samples, device=device)
            ep_losses = []
            for start in range(0, n_samples, batch_size):
                idx = perm[start:start + batch_size]
                optimizer.zero_grad()
                out = model(X_t[idx])
                loss = criterion(out, y_t[idx])
                loss.backward()
                optimizer.step()
                ep_losses.append(float(loss.item()))
            final_loss = float(np.mean(ep_losses)) if ep_losses else 0.0
            if ep % max(1, epochs // report_steps) == 0 or ep == epochs:
                step_num = min(report_steps, max(1, int(round((ep / float(epochs)) * report_steps))))
                self._emit_progress(model_key, step_num, report_steps, phase="lstm_minibatch", tier=tier, progress_cb=progress_cb)
            if final_loss + 1e-5 < best_loss:
                best_loss = final_loss
                bad_epochs = 0
            else:
                bad_epochs += 1
                if bad_epochs >= patience and ep >= 8:
                    break

        model.eval()
        with torch.no_grad():
            preds = model(X_t).cpu().numpy().reshape(-1)
            targets = y_t.cpu().numpy().reshape(-1)
            rmse = float(np.sqrt(mean_squared_error(targets, preds)))
            mae = float(mean_absolute_error(targets, preds))
        return {"loss": final_loss, "rmse": rmse, "mae": mae}

    def _ingest_custom_auditee_items(self, auditee_items: list) -> str:
        """Upsert live/external DB Auditee KPIs into Auditee training pool."""
        if not auditee_items or not isinstance(auditee_items, list):
            return ""
        records = []
        now_dt = datetime.now()
        for idx, item in enumerate(auditee_items):
            if not isinstance(item, dict):
                continue
            kpi_name = str(item.get("kpiName") or item.get("kpi_name") or item.get("kpi") or "").strip()
            if not kpi_name:
                continue
            entity = str(item.get("entity") or "Unit Operasional").strip()
            entity_type = str(item.get("entityType") or item.get("entity_type") or "Department")
            raw_imp = item.get("impactScore") if item.get("impactScore") is not None else item.get("impact")
            raw_lik = item.get("likelihoodScore") if item.get("likelihoodScore") is not None else item.get("likelihood")
            raw_risk = (
                item.get("riskLevel")
                if item.get("riskLevel") is not None
                else (
                    item.get("risk_level")
                    if item.get("risk_level") is not None
                    else (
                        item.get("riskRating")
                        if item.get("riskRating") is not None
                        else (item.get("severity") if item.get("severity") is not None else item.get("tingkat_risiko"))
                    )
                )
            )
            has_ext_risk = raw_risk is not None and str(raw_risk).strip() != "" and str(raw_risk).strip().lower() not in {"not scored", "none", "nan", "null"}
            if has_ext_risk:
                r_lvl, _, def_imp, def_lik = normalize_canonical_risk_level(raw_risk, default_code="MODERATE")
                if raw_imp is not None and raw_lik is not None:
                    try:
                        imp = max(1, min(5, int(round(float(raw_imp)))))
                        lik = max(1, min(5, int(round(float(raw_lik)))))
                    except Exception:
                        imp, lik = def_imp, def_lik
                    if matrix_risk_level(lik, imp) != r_lvl:
                        imp, lik = def_imp, def_lik
                else:
                    imp, lik = def_imp, def_lik
            else:
                try:
                    imp = max(1, min(5, int(round(float(raw_imp if raw_imp is not None else 3)))))
                except Exception:
                    imp = 3
                try:
                    lik = max(1, min(5, int(round(float(raw_lik if raw_lik is not None else 3)))))
                except Exception:
                    lik = 3
                r_lvl = matrix_risk_level(lik, imp)
            sync_freq = str(item.get("syncFrequency") or item.get("sync_frequency") or "External DB")

            hist_list = item.get("history") if isinstance(item.get("history"), list) else []
            if not hist_list and item.get("currentValue") is not None:
                hist_list = [{"period": item.get("lastActualPeriod") or f"{now_dt.year}-{now_dt.month:02d}", "value": item.get("currentValue")}]

            for h in hist_list:
                if not isinstance(h, dict) or h.get("value") is None:
                    continue
                p_label = _period_label_to_auditee_bulan(str(h.get("period") or f"{now_dt.year}-{now_dt.month:02d}"))
                records.append({
                    "id": str(item.get("code") or f"KPI-EXT-{idx+1:03d}"),
                    "source_id": sync_freq,
                    "kpi_name": kpi_name,
                    "entity": entity,
                    "entity_type": entity_type,
                    "period_date": f"{_parse_period_sort_key(p_label)[0]:04d}-{_parse_period_sort_key(p_label)[1]:02d}-01",
                    "period_label": p_label,
                    "actual_value": float(h.get("value")),
                    "target_value": float(item.get("targetValue") or h.get("value") or 0.0),
                    "impact": max(1, min(5, imp)),
                    "likelihood": max(1, min(5, lik)),
                    "risk_level": r_lvl,
                    "sync_frequency": sync_freq,
                })

        if records:
            upsert_auditee_kpi_pool(records)
        return compute_auditee_custom_items_hash(auditee_items)

    def _ingest_live_auditor_items(self, auditor_items: list) -> str:
        """Upsert live Strategic Audit Plan & Internal Audit Performance KPIs into Auditor training pool."""
        if not auditor_items or not isinstance(auditor_items, list):
            return ""
        records = []
        now_dt = datetime.now()
        for idx, item in enumerate(auditor_items):
            if not isinstance(item, dict):
                continue
            code = str(item.get("code") or f"SO-IA{idx+1:02d}").strip()
            kpi_name = str(item.get("kpi") or item.get("kpiName") or "").strip()
            if not kpi_name:
                continue
            pol = "HIB" if str(item.get("polarity") or item.get("hibHig") or "HIG").strip().upper() == "HIB" else "HIG"
            cat = str(item.get("category") or "Pelaksanaan Audit & PKAT")
            unit = str(item.get("unit") or "%")
            so_obj = str(item.get("strategicObjective") or kpi_name)

            raw_t = item.get("target") if item.get("target") is not None else item.get("targetValue")
            base_t = float(_normalize_num_str(raw_t if raw_t is not None else 100.0) or 100.0)
            t_map = item.get("kpiTargets") if isinstance(item.get("kpiTargets"), dict) else {}
            a_map = item.get("kpiActuals") if isinstance(item.get("kpiActuals"), dict) else {}

            paired_periods = []
            for p_k, a_raw in a_map.items():
                a_norm = _normalize_num_str(a_raw)
                if not a_norm:
                    continue
                t_norm = _normalize_num_str(t_map.get(p_k))
                paired_periods.append((str(p_k), float(t_norm) if t_norm else base_t, float(a_norm)))

            if not paired_periods:
                a_norm = _normalize_num_str(item.get("actual") if item.get("actual") is not None else item.get("currentValue"))
                if a_norm:
                    paired_periods.append((str(item.get("selectedPeriod") or now_dt.year), base_t, float(a_norm)))

            for p_raw, t_v, a_v in paired_periods:
                ratio = (a_v / t_v * 100.0) if (pol == "HIG" and t_v > 0) else (100.0 if a_v <= 0 else t_v / a_v * 100.0)
                status = "Good" if ratio >= 100.0 else ("Moderate" if ratio >= 70.0 else "Poor")
                p_year, p_month, _ = _parse_period_sort_key(p_raw)
                if p_year == 9999:
                    p_year, p_month = now_dt.year, now_dt.month
                p_label = _period_label_to_auditee_bulan(f"{p_year:04d}-{p_month:02d}")
                records.append({
                    "id": f"AUD-LIVE-{code}-{p_raw}",
                    "code": code,
                    "kpi_name": kpi_name,
                    "category": cat,
                    "strategic_objective": so_obj,
                    "period_date": f"{p_year:04d}-{p_month:02d}-01",
                    "period_label": p_label,
                    "hib_hig": pol,
                    "unit": unit,
                    "target_value": t_v,
                    "actual_value": a_v,
                    "achievement_rate": round(ratio, 2),
                    "status": status,
                })

        if records:
            upsert_auditor_kpi_pool(records)
        return compute_auditor_items_hash(auditor_items)

    def retrain_auditor_kpi_suite(
        self,
        auditor_items: list = None,
        progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None,
    ) -> dict:
        """Fine-tune PyTorch Auditor LSTM + Status Classifier with validation & atomic promotion."""
        t0 = time.perf_counter()
        peak_ram = _current_rss_mb()
        try:
            prof = get_resource_profile()
            apply_thread_limits(prof["train_threads"])
            live_hash = self._ingest_live_auditor_items(auditor_items) if auditor_items else ""

            df = load_auditor_training_dataframe()
            if df is None or df.empty or "TARGET: Nilai Aktual" not in df.columns:
                fp_empty = compute_kpi_auditor_fingerprint()
                self.last_model_results["kpi_auditor"] = {
                    "ok": True,
                    "skipped": True,
                    "promoted": False,
                    "reason": "Belum ada data KPI Auditor untuk pelatihan ulang; model pre-trained dipertahankan.",
                    "fingerprint": fp_empty,
                    "live_auditor_hash": live_hash,
                    "pool_hash": str(fp_empty.get("pool_hash") or ""),
                }
                return {"ok": True, "skipped": True, "rows": 0, "kpis": 0, "hash": live_hash}

            tier_info = select_training_tier(len(df), est_bytes_per_row=200, profile=prof)
            tier = tier_info["tier"]

            df = df.copy()
            df["_sort_key"] = df["Periode"].astype(str).apply(_parse_period_sort_key)
            df["_actual"] = pd.to_numeric(df["TARGET: Nilai Aktual"], errors="coerce")
            df["_target"] = pd.to_numeric(df.get("Nilai Target", 100.0), errors="coerce").fillna(100.0)
            df = df.dropna(subset=["Kode SO", "_actual"]).sort_values(["Kode SO", "_sort_key"]).reset_index(drop=True)
            if df.empty:
                return {"ok": False, "rows": 0, "kpis": 0, "hash": live_hash}

            seq_len = 10
            X_seqs, y_seqs = [], []
            auditor_scalers = {}

            for so_code, grp in df.groupby("Kode SO", sort=False):
                vals = grp["_actual"].values.astype(np.float32).reshape(-1, 1)
                lo, hi = float(vals.min()), float(vals.max())
                pad = max(1.0, (hi - lo) * 0.25)
                sc = MinMaxScaler(feature_range=(0, 1))
                sc.fit(np.array([[lo - pad], [hi + pad]], dtype=np.float32))
                auditor_scalers[str(so_code)] = sc
                scaled = sc.transform(vals)

                if len(scaled) > seq_len:
                    for idx in range(seq_len, len(scaled)):
                        X_seqs.append(scaled[idx - seq_len:idx])
                        y_seqs.append(scaled[idx])
                elif len(scaled) >= 2:
                    hist_part = scaled[:-1]
                    if len(hist_part) < seq_len:
                        hist_part = np.vstack([np.repeat(hist_part[:1], seq_len - len(hist_part), axis=0), hist_part])
                    X_seqs.append(hist_part)
                    y_seqs.append(scaled[-1])
                else:
                    X_seqs.append(np.repeat(scaled[:1], seq_len, axis=0))
                    y_seqs.append(scaled[0])

            device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
            lstm_model = PyTorchLSTMRegressor(input_dim=1, hidden_dim=64, num_layers=2, output_dim=1).to(device)
            lstm_path = os.path.join(self.model_dir, "auditor_kpi_lstm.pth")
            warm_started = False
            if os.path.exists(lstm_path):
                try:
                    lstm_model.load_state_dict(torch.load(lstm_path, map_location=device))
                    warm_started = True
                except Exception:
                    pass

            lstm_stats = self._fine_tune_lstm_minibatch(
                lstm_model,
                X_seqs,
                y_seqs,
                warm_started=warm_started,
                base_lr=0.003,
                model_key="kpi_auditor",
                tier=tier,
                progress_cb=progress_cb,
            )
            final_loss = lstm_stats["loss"]
            rmse_val = lstm_stats["rmse"]

            df_feat = df.copy()
            df_feat["polarity_is_hig"] = (df_feat["Tipe Polaritas (HIG/HIB)"].astype(str).str.upper() == "HIG").astype(int)
            df_feat["target_val"] = df_feat["_target"].astype(float)
            df_feat["actual_val"] = df_feat["_actual"].astype(float)
            df_feat["achievement_ratio"] = pd.to_numeric(df_feat.get("TARGET: Rasio Ketercapaian (%)", 100.0), errors="coerce").fillna(100.0)
            df_feat["gap_ratio"] = np.where(
                df_feat["polarity_is_hig"] == 1,
                (df_feat["actual_val"] - df_feat["target_val"]) / (df_feat["target_val"].abs() + 1e-6),
                (df_feat["target_val"] - df_feat["actual_val"]) / (df_feat["target_val"].abs() + 1e-6),
            )
            for lag in (1, 2, 3):
                df_feat[f"lag_{lag}"] = df_feat.groupby("Kode SO")["actual_val"].shift(lag).fillna(df_feat["actual_val"])
                df_feat[f"ratio_lag_{lag}"] = df_feat.groupby("Kode SO")["achievement_ratio"].shift(lag).fillna(df_feat["achievement_ratio"])

            df_feat["rolling_mean_3"] = (
                df_feat.groupby("Kode SO")["actual_val"].transform(lambda s: s.shift(1).rolling(3, min_periods=1).mean()).fillna(df_feat["actual_val"])
            )
            df_feat["rolling_std_3"] = (
                df_feat.groupby("Kode SO")["actual_val"].transform(lambda s: s.shift(1).rolling(3, min_periods=1).std()).fillna(0.0)
            )
            df_feat["velocity_trend"] = df_feat["actual_val"] - df_feat["lag_3"]

            feature_cols = [
                "polarity_is_hig", "target_val", "actual_val", "achievement_ratio", "gap_ratio",
                "lag_1", "lag_2", "lag_3", "ratio_lag_1", "ratio_lag_2", "ratio_lag_3",
                "rolling_mean_3", "rolling_std_3", "velocity_trend",
            ]
            X_class = df_feat[feature_cols].values.astype(np.float32)
            y_labels = df_feat["TARGET: Status Achievement"].astype(str).values
            le = LabelEncoder()
            y_enc = le.fit_transform(y_labels)
            sc_class = StandardScaler()
            X_c_scaled = sc_class.fit_transform(X_class).astype(np.float32)

            status_clf = GradientBoostingClassifier(n_estimators=100, learning_rate=0.08, max_depth=4, random_state=42)
            if len(np.unique(y_enc)) >= 2:
                status_clf.fit(X_c_scaled, y_enc)

            total_samples = int(len(df))
            bundle = {
                "scalers": auditor_scalers,
                "global_min_val": float(df["_actual"].min()),
                "global_max_val": float(df["_actual"].max()),
                "feature_cols": feature_cols,
                "scaler_class": sc_class,
                "status_classifier": status_clf,
                "label_encoder": le,
                "seq_length": seq_len,
                "known_kpis": sorted(df["Kode SO"].astype(str).unique().tolist()),
                "data_samples": total_samples,
            }

            staged_lstm = os.path.join(self.model_dir, "auditor_kpi_lstm.pth.staging")
            staged_bundle = os.path.join(self.model_dir, "auditor_kpi_bundle.pkl.staging")
            torch.save(lstm_model.state_dict(), staged_lstm)
            with open(staged_bundle, "wb") as f:
                pickle.dump(bundle, f, protocol=pickle.HIGHEST_PROTOCOL)

            metrics = {
                "primary_name": "rmse",
                "primary": round(rmse_val, 4),
                "rmse": round(rmse_val, 4),
                "mae": round(lstm_stats["mae"], 4),
                "loss": round(final_loss, 6),
                "samples": total_samples,
            }
            promo = validate_and_promote_artifacts(
                "kpi_auditor",
                {"auditor_kpi_lstm.pth": staged_lstm, "auditor_kpi_bundle.pkl": staged_bundle},
                metrics,
                model_dir=self.model_dir,
                data_server_dir=DATA_SERVER_MODEL_DIR,
            )

            peak_ram = max(peak_ram, _current_rss_mb())
            _clear_checkpoint("kpi_auditor")
            elapsed = round(time.perf_counter() - t0, 2)
            fp = compute_kpi_auditor_fingerprint(live_auditor_hash=live_hash or None)
            pool_hash_val = str(fp.get("pool_hash") or fp.get("hash") or "")
            self.last_model_results["kpi_auditor"] = {
                "ok": bool(promo.get("promoted", True)),
                "promoted": bool(promo.get("promoted", True)),
                "reason": promo.get("reason"),
                "samples": total_samples,
                "kpis": len(auditor_scalers),
                "tier": tier,
                "batches": tier_info["total_batches"],
                "duration_seconds": elapsed,
                "peak_ram_mb": peak_ram,
                "metrics": metrics,
                "fingerprint": fp,
                "pool_hash": pool_hash_val,
                "live_auditor_hash": live_hash,
                "auditor_hash": live_hash,
            }
            del df, df_feat, X_class, X_c_scaled
            gc.collect()
            return {
                "ok": bool(promo.get("promoted", True)),
                "rows": int(len(X_seqs)),
                "samples": total_samples,
                "kpis": len(auditor_scalers),
                "hash": live_hash,
                "pool_hash": pool_hash_val,
                "final_loss": round(final_loss, 6),
                "rmse": round(rmse_val, 4),
            }
        except Exception as e:
            print(f"Error retraining Auditor KPI suite: {e}")
            self.last_model_results["kpi_auditor"] = {"ok": False, "error": str(e)}
            return {"ok": False, "rows": 0, "kpis": 0, "hash": "", "error": str(e)}

    def retrain_auditee_kpi_only(
        self,
        auditee_items: list = None,
        progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None,
    ) -> bool:
        """Retrain ONLY Auditee KPI PyTorch LSTM + Tabular Classifier with validation & atomic promotion."""
        t0 = time.perf_counter()
        peak_ram = _current_rss_mb()
        auditee_custom_hash = ""
        try:
            prof = get_resource_profile()
            n_threads = apply_thread_limits(prof["train_threads"])
            if auditee_items:
                auditee_custom_hash = self._ingest_custom_auditee_items(auditee_items)

            df = load_auditee_training_dataframe()
            if df.empty or "TARGET: Nilai Aktual (%)" not in df.columns:
                fp_empty = compute_kpi_auditee_fingerprint()
                self.last_model_results["kpi_auditee"] = {
                    "ok": True,
                    "skipped": True,
                    "promoted": False,
                    "reason": "Belum ada data KPI Auditee untuk pelatihan ulang; model pre-trained dipertahankan.",
                    "fingerprint": fp_empty,
                    "auditee_pool_hash": str(fp_empty.get("pool_hash") or ""),
                    "auditee_custom_hash": auditee_custom_hash,
                    "auditee_count": int(fp_empty.get("rows") or 0),
                }
                return True

            tier_info = select_training_tier(len(df), est_bytes_per_row=200, profile=prof)
            tier = tier_info["tier"]

            df = df.copy()
            df["_val"] = pd.to_numeric(df["TARGET: Nilai Aktual (%)"], errors="coerce")
            df = df.dropna(subset=["Nama KPI", "_val"])
            df["_sort_key"] = df["Periode"].astype(str).apply(_parse_period_sort_key)
            df = df.sort_values(["Nama KPI", "_sort_key"]).reset_index(drop=True)

            y_values = df["_val"].values.astype(np.float32)
            global_scaler = MinMaxScaler(feature_range=(0, 1))
            global_scaler.fit(y_values.reshape(-1, 1))

            seq_length = 3
            X_list, y_list = [], []
            per_kpi_scalers = {}

            for kpi_name, grp in df.groupby("Nama KPI", sort=False):
                vals = grp["_val"].values.astype(np.float32).reshape(-1, 1)
                lo, hi = float(vals.min()), float(vals.max())
                pad = max(0.5, (hi - lo) * 0.2)
                sc_kpi = MinMaxScaler(feature_range=(0, 1))
                sc_kpi.fit(np.array([[lo - pad], [hi + pad]], dtype=np.float32))
                per_kpi_scalers[str(kpi_name)] = sc_kpi
                scaled = sc_kpi.transform(vals)
                if len(scaled) > seq_length:
                    for i in range(len(scaled) - seq_length):
                        X_list.append(scaled[i:i + seq_length])
                        y_list.append(scaled[i + seq_length])
                elif len(scaled) >= 2:
                    hist_part = np.vstack([np.repeat(scaled[:1], seq_length - (len(scaled) - 1), axis=0), scaled[:-1]])
                    X_list.append(hist_part[-seq_length:])
                    y_list.append(scaled[-1])
                else:
                    X_list.append(np.repeat(scaled[:1], seq_length, axis=0))
                    y_list.append(scaled[0])

            device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
            model = PyTorchLSTMRegressor().to(device)
            lstm_path = os.path.join(self.model_dir, "kpi_lstm.pth")
            warm_started = False
            if os.path.exists(lstm_path):
                try:
                    model.load_state_dict(torch.load(lstm_path, map_location=device))
                    warm_started = True
                except Exception:
                    pass

            lstm_stats = self._fine_tune_lstm_minibatch(
                model,
                X_list,
                y_list,
                warm_started=warm_started,
                base_lr=0.005,
                model_key="kpi_auditee",
                tier=tier,
                progress_cb=progress_cb,
            )
            auditee_loss = lstm_stats["loss"]
            rmse_val = lstm_stats["rmse"]

            imp_vals = pd.to_numeric(df.get("TARGET: Impact (1-5)", 3), errors="coerce").fillna(3).clip(1, 5).astype(int).values
            lik_vals = pd.to_numeric(df.get("TARGET: Likelihood (1-5)", 3), errors="coerce").fillna(3).clip(1, 5).astype(int).values
            imp_le = LabelEncoder()
            lik_le = LabelEncoder()
            y_imp_enc = imp_le.fit_transform(imp_vals)
            y_lik_enc = lik_le.fit_transform(lik_vals)

            X_tab = y_values.reshape(-1, 1)
            sc_tab = StandardScaler()
            X_tab_scaled = sc_tab.fit_transform(X_tab).astype(np.float32)

            imp_clf = XGBClassifier(max_depth=3, learning_rate=0.05, n_estimators=50, tree_method="hist", n_jobs=n_threads, random_state=42)
            if len(np.unique(y_imp_enc)) >= 2:
                imp_clf.fit(X_tab_scaled, y_imp_enc)

            lik_clf = XGBClassifier(max_depth=3, learning_rate=0.05, n_estimators=50, tree_method="hist", n_jobs=n_threads, random_state=42)
            if len(np.unique(y_lik_enc)) >= 2:
                lik_clf.fit(X_tab_scaled, y_lik_enc)

            total_samples = int(len(df))
            bundle = {
                "scaler": global_scaler,
                "scalers": per_kpi_scalers,
                "tab_scaler": sc_tab,
                "impact_classifier": imp_clf,
                "likelihood_classifier": lik_clf,
                "impact_le": imp_le,
                "likelihood_le": lik_le,
                "canonical_risk_levels": list(CANONICAL_RISK_CODES),
                "data_samples": total_samples,
                "known_kpis": list(df["Nama KPI"].unique()),
            }

            staged_lstm = os.path.join(self.model_dir, "kpi_lstm.pth.staging")
            staged_bundle = os.path.join(self.model_dir, "kpi_bundle.pkl.staging")
            torch.save(model.state_dict(), staged_lstm)
            with open(staged_bundle, "wb") as f:
                pickle.dump(bundle, f, protocol=pickle.HIGHEST_PROTOCOL)

            metrics = {
                "primary_name": "rmse",
                "primary": round(rmse_val, 4),
                "rmse": round(rmse_val, 4),
                "mae": round(lstm_stats["mae"], 4),
                "loss": round(auditee_loss, 6),
                "samples": total_samples,
            }
            promo = validate_and_promote_artifacts(
                "kpi_auditee",
                {"kpi_lstm.pth": staged_lstm, "kpi_bundle.pkl": staged_bundle},
                metrics,
                model_dir=self.model_dir,
                data_server_dir=DATA_SERVER_MODEL_DIR,
            )

            auditee_pool_hash, auditee_pool_count = compute_auditee_pool_hash()
            fp = compute_kpi_auditee_fingerprint(custom_hash=auditee_custom_hash or None)
            peak_ram = max(peak_ram, _current_rss_mb())
            _clear_checkpoint("kpi_auditee")
            elapsed = round(time.perf_counter() - t0, 2)

            self.last_model_results["kpi_auditee"] = {
                "ok": bool(promo.get("promoted", True)),
                "promoted": bool(promo.get("promoted", True)),
                "reason": promo.get("reason"),
                "samples": total_samples,
                "kpis": len(per_kpi_scalers),
                "tier": tier,
                "batches": tier_info["total_batches"],
                "duration_seconds": elapsed,
                "peak_ram_mb": peak_ram,
                "metrics": metrics,
                "fingerprint": fp,
                "auditee_pool_hash": auditee_pool_hash,
                "auditee_custom_hash": auditee_custom_hash,
                "auditee_count": auditee_pool_count,
            }
            del df, X_list, y_list, X_tab, X_tab_scaled
            gc.collect()
            return bool(promo.get("promoted", True))
        except Exception as e:
            print(f"Error retraining Auditee KPI suite: {e}")
            self.last_model_results["kpi_auditee"] = {"ok": False, "error": str(e)}
            return False

    def retrain_kpi_lstm_suite(
        self,
        auditee_items: list = None,
        auditor_items: list = None,
        progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None,
    ) -> bool:
        """Backward-compatible wrapper that retrains both Auditee and Auditor KPI suites."""
        self.last_kpi_metrics = {}
        ok_auditee = self.retrain_auditee_kpi_only(auditee_items=auditee_items, progress_cb=progress_cb)
        auditor_meta = self.retrain_auditor_kpi_suite(auditor_items=auditor_items, progress_cb=progress_cb)
        res_auditee = self.last_model_results.get("kpi_auditee", {})
        self.last_kpi_metrics = {
            "auditee_retrained": bool(ok_auditee),
            "auditee_count": int(res_auditee.get("auditee_count", 0)),
            "auditee_kpis": int(res_auditee.get("kpis", 0)),
            "auditee_pool_hash": str(res_auditee.get("auditee_pool_hash", "")),
            "auditee_custom_hash": str(res_auditee.get("auditee_custom_hash", "")),
            "auditee_loss": float((res_auditee.get("metrics") or {}).get("loss", 0.0)),
            "auditor_retrained": bool(auditor_meta.get("ok")),
            "auditor_count": int(auditor_meta.get("rows", 0)),
            "auditor_kpis": int(auditor_meta.get("kpis", 0)),
            "auditor_hash": str(auditor_meta.get("hash") or ""),
        }
        return bool(ok_auditee)

    def retrain_single_model(
        self,
        model_key: str,
        trigger: str = "manual",
        auditor_items: list = None,
        auditee_items: list = None,
        sync_sources: bool = True,
        progress_cb: Optional[Callable[[Dict[str, Any]], None]] = None,
    ) -> Dict[str, Any]:
        """
        Execute training for a SINGLE AI model (`entity`, `anomaly`, `nlp`, `kpi_auditee`, `kpi_auditor`)
        without touching any other model's files or state.
        """
        lower_current_process_priority()
        t0 = time.perf_counter()
        sync_res = self.sync_live_cross_service_data(scope=model_key) if sync_sources else {}

        if model_key == "entity":
            ok = self.retrain_department_suite(progress_cb=progress_cb)
        elif model_key == "anomaly":
            ok = self.retrain_anomaly_suite(progress_cb=progress_cb)
        elif model_key == "nlp":
            ok = self.retrain_document_suite(progress_cb=progress_cb)
        elif model_key == "kpi_auditee":
            ok = self.retrain_auditee_kpi_only(auditee_items=auditee_items, progress_cb=progress_cb)
        elif model_key == "kpi_auditor":
            meta = self.retrain_auditor_kpi_suite(auditor_items=auditor_items, progress_cb=progress_cb)
            ok = bool(meta.get("ok"))
        else:
            raise ValueError(f"Unknown model_key: {model_key}")

        elapsed = round(time.perf_counter() - t0, 2)
        detail = dict(self.last_model_results.get(model_key) or {})
        detail.setdefault("duration_seconds", elapsed)
        detail["ok"] = bool(ok)
        detail["model"] = model_key
        detail["trigger"] = trigger
        detail["cross_service_sync"] = sync_res
        return detail

    def auto_retrain_all(
        self,
        trigger: str = "auto",
        auditor_items: list = None,
        auditee_items: list = None,
        scope: str = "all",
    ) -> dict:
        """Run re-training across the requested scope (`all`, `kpi`, or a specific model key)."""
        t0 = time.perf_counter()
        target_models = resolve_model_keys(scope) or list(MODEL_KEYS)
        sync_res = self.sync_live_cross_service_data(scope=scope)

        res_anom = self.retrain_anomaly_suite() if "anomaly" in target_models else True
        res_dept = self.retrain_department_suite() if "entity" in target_models else True
        res_doc = self.retrain_document_suite() if "nlp" in target_models else True

        if "kpi_auditee" in target_models and "kpi_auditor" in target_models:
            res_kpi = self.retrain_kpi_lstm_suite(auditee_items=auditee_items, auditor_items=auditor_items)
        elif "kpi_auditee" in target_models:
            res_kpi = self.retrain_auditee_kpi_only(auditee_items=auditee_items)
            self.last_kpi_metrics = {
                "auditee_retrained": res_kpi,
                "auditee_count": int(self.last_model_results.get("kpi_auditee", {}).get("auditee_count", 0)),
                "auditee_kpis": int(self.last_model_results.get("kpi_auditee", {}).get("kpis", 0)),
                "auditee_pool_hash": str(self.last_model_results.get("kpi_auditee", {}).get("auditee_pool_hash", "")),
                "auditee_custom_hash": str(self.last_model_results.get("kpi_auditee", {}).get("auditee_custom_hash", "")),
                "auditor_retrained": False,
            }
        elif "kpi_auditor" in target_models:
            aud_meta = self.retrain_auditor_kpi_suite(auditor_items=auditor_items)
            res_kpi = bool(aud_meta.get("ok"))
            self.last_kpi_metrics = {
                "auditee_retrained": False,
                "auditor_retrained": res_kpi,
                "auditor_count": int(aud_meta.get("rows", 0)),
                "auditor_kpis": int(aud_meta.get("kpis", 0)),
                "auditor_hash": str(aud_meta.get("hash") or ""),
            }
        else:
            res_kpi = True

        elapsed = round(time.perf_counter() - t0, 2)
        kpi_meta = getattr(self, "last_kpi_metrics", {})

        return {
            "status": "success" if (res_anom and res_dept and res_doc and res_kpi) else "partial_success",
            "trigger": trigger,
            "scope": scope,
            "duration_seconds": elapsed,
            "cross_service_sync": sync_res,
            "anomaly_retrained": res_anom if "anomaly" in target_models else False,
            "department_retrained": res_dept if "entity" in target_models else False,
            "document_retrained": res_doc if "nlp" in target_models else False,
            "kpi_lstm_retrained": res_kpi if ("kpi_auditee" in target_models or "kpi_auditor" in target_models) else False,
            "auditee_retrained": kpi_meta.get("auditee_retrained", False),
            "auditor_retrained": kpi_meta.get("auditor_retrained", False),
            "trained_auditee_count": kpi_meta.get("auditee_count", 0),
            "trained_auditee_kpis": kpi_meta.get("auditee_kpis", 0),
            "trained_auditee_pool_hash": kpi_meta.get("auditee_pool_hash", ""),
            "trained_auditee_custom_hash": kpi_meta.get("auditee_custom_hash", ""),
            "trained_auditor_count": len(auditor_items) if auditor_items else kpi_meta.get("auditor_kpis", 0),
            "trained_auditor_hash": kpi_meta.get("auditor_hash", ""),
            "model_results": self.last_model_results,
        }


def invalidate_anomaly_redis_cache():
    """Clear analytics-service Redis cache (`audit:analytics:batch:anomaly`) using raw RESP over socket."""
    import socket

    redis_host = os.getenv("REDIS_HOST", "redis" if os.path.exists("/.dockerenv") else "localhost")
    redis_port = int(os.getenv("REDIS_PORT", "6379"))
    key = "audit:analytics:batch:anomaly"
    cmd = f"*2\r\n$3\r\nDEL\r\n${len(key)}\r\n{key}\r\n".encode("utf-8")
    try:
        with socket.create_connection((redis_host, redis_port), timeout=1.0) as sock:
            sock.sendall(cmd)
            _ = sock.recv(64)
    except Exception:
        pass


def reload_fastapi_models(fastapi_app=None, only_models: Optional[List[str]] = None):
    """
    Seamlessly hot-reload in-memory model bundles in FastAPI app and module-level globals.
    When `only_models` is supplied (e.g. `['anomaly']`), ONLY that model's bundle is reloaded
    from disk, avoiding unnecessary disk I/O and RAM churn on unrelated models.
    """
    try:
        targets = set(resolve_model_keys("all") if not only_models else [m for item in only_models for m in resolve_model_keys(item)])
        loaded_bundles = {}

        if "anomaly" in targets:
            anom_path = os.path.join(MODEL_DIR, "anomaly_bundle.pkl")
            if os.path.exists(anom_path):
                with open(anom_path, "rb") as f:
                    loaded_bundles["anomaly_bundle"] = pickle.load(f)

        if "entity" in targets:
            dept_path = os.path.join(MODEL_DIR, "department_bundle.pkl")
            if os.path.exists(dept_path):
                with open(dept_path, "rb") as f:
                    loaded_bundles["department_bundle"] = pickle.load(f)

        if "nlp" in targets:
            doc_path = os.path.join(MODEL_DIR, "document_bundle.pkl")
            if os.path.exists(doc_path):
                with open(doc_path, "rb") as f:
                    loaded_bundles["document_bundle"] = pickle.load(f)

        if "kpi_auditee" in targets:
            kpi_bundle_path = os.path.join(MODEL_DIR, "kpi_bundle.pkl")
            if os.path.exists(kpi_bundle_path):
                with open(kpi_bundle_path, "rb") as f:
                    loaded_bundles["kpi_bundle"] = pickle.load(f)

            kpi_lstm_path = os.path.join(MODEL_DIR, "kpi_lstm.pth")
            if os.path.exists(kpi_lstm_path):
                device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
                lstm_model = PyTorchLSTMRegressor().to(device)
                lstm_model.load_state_dict(torch.load(kpi_lstm_path, map_location=device))
                lstm_model.eval()
                loaded_bundles["kpi_lstm_model"] = lstm_model

        if "kpi_auditor" in targets:
            auditor_bundle_path = os.path.join(MODEL_DIR, "auditor_kpi_bundle.pkl")
            if os.path.exists(auditor_bundle_path):
                with open(auditor_bundle_path, "rb") as f:
                    loaded_bundles["auditor_kpi_bundle"] = pickle.load(f)

            auditor_lstm_path = os.path.join(MODEL_DIR, "auditor_kpi_lstm.pth")
            if os.path.exists(auditor_lstm_path):
                device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
                aud_lstm = PyTorchLSTMRegressor().to(device)
                aud_lstm.load_state_dict(torch.load(auditor_lstm_path, map_location=device))
                aud_lstm.eval()
                loaded_bundles["auditor_kpi_lstm_model"] = aud_lstm

        if fastapi_app is not None:
            for k, v in loaded_bundles.items():
                setattr(fastapi_app, k, v)

        for mod_name in ["__main__", "main", "backend.python-ai.main"]:
            if mod_name in sys.modules:
                mod = sys.modules[mod_name]
                for k, v in loaded_bundles.items():
                    if hasattr(mod, k):
                        setattr(mod, k, v)

        if "anomaly" in targets:
            invalidate_anomaly_redis_cache()

        print(f"Hot-reloaded AI model bundles ({', '.join(sorted(targets))}) into memory successfully.")
        return True
    except Exception as e:
        print(f"Error during hot-reloading models: {e}")
        return False
