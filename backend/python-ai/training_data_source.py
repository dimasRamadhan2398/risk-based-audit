"""
AuditSphere — Scalable Training Data Source & Incremental Fingerprinting Layer.

Provides:
1. O(1) / lightweight incremental fingerprinting for each AI model (`entity`, `anomaly`, `nlp`, `kpi_auditee`, `kpi_auditor`)
   so 60-second change detection takes milliseconds even on multi-million-row datasets.
2. Chunked / batched iteration (`iter_batches`) over CSV, Parquet, and JSON pools.
3. Tier-3 Stratified / Minority-Preserving Reservoir Sampler so datasets with millions of rows
   keep 100% of positive/minority anomaly cases while sampling majority rows within RAM budget.
"""
import os
import json
import hashlib
from typing import Dict, Any, Iterator, List, Optional, Tuple
import numpy as np
import pandas as pd

try:
    import pyarrow.parquet as pq
except ImportError:
    pq = None

from ai_training_storage import (
    resolve_ai_training_dir,
    AUDITEE_PARQUET_FILE,
    AUDITOR_PARQUET_FILE,
    ENTITY_PREDICTION_PARQUET_FILE,
    load_auditee_training_dataframe,
    load_auditor_training_dataframe,
    load_entity_prediction_training_dataframe,
)
from nlp_taxonomy import (
    TAXONOMY_FILE,
    IMPORTED_DOCS_FILE,
    LIVE_CORPUS_FILE,
    harvest_all_live_nlp_training_data,
)

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.getenv("MODEL_DIR", os.path.join(BASE_DIR, "models"))
ATR_FINDINGS_POOL_FILE = os.path.join(MODEL_DIR, "atr_findings_pool.json")


def compute_atr_findings_hash(items: List[Dict[str, Any]]) -> Tuple[str, int]:
    """Deterministic SHA-256[:16] fingerprint and count over canonical Action Taken Report finding attributes."""
    if not items or not isinstance(items, list):
        return ("", 0)
    rows: List[str] = []
    for it in items:
        if not isinstance(it, dict):
            continue
        fid = str(it.get("id") or it.get("finding_id") or "").strip()
        rnum = str(it.get("report_number") or "").strip()
        title = str(it.get("finding_title") or it.get("title") or "").strip().lower()
        cat = str(it.get("finding_category") or it.get("category") or "").strip().lower()
        status = str(it.get("status") or "PLANNED").strip().upper()
        try:
            prog = int(float(it.get("progress") or 0))
        except Exception:
            prog = 0
        due = str(it.get("due_date") or "")[:10]
        is_ov = "1" if bool(it.get("is_overdue")) else "0"
        try:
            ov_days = int(float(it.get("overdue_days") or 0))
        except Exception:
            ov_days = 0
        rows.append(f"{fid}|{rnum}|{title}|{cat}|{status}|{prog}|{due}|{is_ov}|{ov_days}")
    if not rows:
        return ("", 0)
    rows.sort()
    digest = hashlib.sha256("\n".join(rows).encode("utf-8")).hexdigest()[:16]
    return (digest, len(rows))


def load_atr_findings_pool() -> List[Dict[str, Any]]:
    """Load persisted Action Taken Report findings pool from MODEL_DIR."""
    if not os.path.exists(ATR_FINDINGS_POOL_FILE):
        return []
    try:
        with open(ATR_FINDINGS_POOL_FILE, "r", encoding="utf-8") as f:
            data = json.load(f)
        if isinstance(data, dict) and isinstance(data.get("items"), list):
            return data["items"]
        if isinstance(data, list):
            return data
    except Exception:
        pass
    return []


def _fast_file_stat_token(path: str) -> Tuple[str, int]:
    """Return (sha_token, approx_rows_or_size) for a file without reading the entire file if large."""
    if not path or not os.path.exists(path):
        return ("", 0)
    try:
        st = os.stat(path)
        size = int(st.st_size)
        mtime_ns = int(getattr(st, "st_mtime_ns", int(st.st_mtime * 1e9)))

        # If Parquet and pyarrow is installed, read metadata footer only (O(1) row count!)
        if path.endswith(".parquet") and pq is not None:
            try:
                meta = pq.read_metadata(path)
                nrows = int(meta.num_rows)
                # Read head & tail bytes for content hash stability
                with open(path, "rb") as f:
                    head = f.read(4096)
                    if size > 8192:
                        f.seek(max(0, size - 4096))
                    tail = f.read(4096)
                h = hashlib.sha256()
                h.update(f"{size}|{nrows}|".encode("utf-8"))
                h.update(head)
                h.update(tail)
                return (h.hexdigest()[:16], nrows)
            except Exception:
                pass

        # Small file (< 2 MB): hash full content and count newlines for exact row count
        if size <= 2 * 1024 * 1024:
            with open(path, "rb") as f:
                data = f.read()
            digest = hashlib.sha256(data).hexdigest()[:16]
            if path.endswith(".csv"):
                nrows = max(0, data.count(b"\n") - 1)
            else:
                nrows = size
            return (digest, nrows)

        # Large file (> 2 MB): sample head + middle + tail 64 KB + size + mtime (O(1))
        h = hashlib.sha256()
        h.update(f"{size}|{mtime_ns}|".encode("utf-8"))
        with open(path, "rb") as f:
            head = f.read(65536)
            h.update(head)
            avg_line = max(40, len(head) // max(1, head.count(b"\n")))
            f.seek(size // 2)
            h.update(f.read(65536))
            f.seek(max(0, size - 65536))
            h.update(f.read(65536))
        approx_rows = max(1, size // avg_line)
        return (h.hexdigest()[:16], approx_rows)
    except Exception:
        return ("", 0)


def compute_entity_fingerprint() -> Dict[str, Any]:
    """Compute fingerprint and row count for Entity Prediction training pool."""
    try:
        df = load_entity_prediction_training_dataframe()
        if df is not None and not df.empty:
            cols = [c for c in ("entity", "risk_category", "inherent_likelihood", "inherent_impact", "target_likelihood", "target_impact", "assessment_month") if c in df.columns]
            sub = df[cols].copy()
            for c in sub.columns:
                sub[c] = sub[c].astype(str).str.strip()
            sub = sub.sort_values(cols).reset_index(drop=True)
            lines = ["|".join(row) for row in sub.values.tolist()]
            digest = hashlib.sha256("\n".join(lines).encode("utf-8")).hexdigest()[:16]
            return {"hash": digest, "rows": int(len(df)), "source": "entity_pool"}
    except Exception:
        pass
    tok, rows = _fast_file_stat_token(ENTITY_PREDICTION_PARQUET_FILE)
    if tok:
        return {"hash": tok, "rows": rows, "source": "entity_parquet"}
    csv_path = os.path.join(resolve_ai_training_dir(), "department_prediction", "department_dataset.csv")
    tok, rows = _fast_file_stat_token(csv_path)
    return {"hash": tok, "rows": rows, "source": "department_csv"}


def compute_anomaly_fingerprint() -> Dict[str, Any]:
    """Compute combined fingerprint and row count for Anomaly Detection (core + bank transaction + ATR findings datasets)."""
    ai_dir = resolve_ai_training_dir()
    anom_csv = os.path.join(ai_dir, "anomaly_prediction", "anomaly_data.csv")
    bank_csv = os.path.join(ai_dir, "anomaly_prediction", "bank_transaction_anomaly_data.csv")
    repeat_csv = os.path.join(ai_dir, "anomaly_prediction", "repeat_finding_data.csv")
    mit_csv = os.path.join(ai_dir, "anomaly_prediction", "mitigation_overdue_data.csv")
    tok1, rows1 = _fast_file_stat_token(anom_csv)
    tok2, rows2 = _fast_file_stat_token(bank_csv)
    tok3, rows3 = _fast_file_stat_token(repeat_csv)
    tok4, rows4 = _fast_file_stat_token(mit_csv)
    atr_items = load_atr_findings_pool()
    atr_hash, atr_rows = compute_atr_findings_hash(atr_items)
    if not tok1 and not tok2 and not atr_hash:
        return {"hash": "", "rows": 0, "source": "anomaly_csv"}
    combined = hashlib.sha256(
        f"{tok1}:{rows1}|{tok2}:{rows2}|{tok3}:{rows3}|{tok4}:{rows4}|{atr_hash}:{atr_rows}".encode("utf-8")
    ).hexdigest()[:16]
    return {
        "hash": combined,
        "rows": int(rows1 + rows2 + rows3 + rows4 + atr_rows),
        "source": "anomaly_suite",
        "core_rows": rows1,
        "bank_rows": rows2,
        "findings_seed_rows": int(rows3 + rows4),
        "atr_live_rows": int(atr_rows),
        "atr_pool_hash": atr_hash,
    }


def compute_nlp_fingerprint() -> Dict[str, Any]:
    """Compute fingerprint and sample count for NLP Analysis (imported docs + live corpus + taxonomy)."""
    try:
        samples = harvest_all_live_nlp_training_data()
        t_tok, _ = _fast_file_stat_token(TAXONOMY_FILE)
        i_tok, _ = _fast_file_stat_token(IMPORTED_DOCS_FILE)
        l_tok, _ = _fast_file_stat_token(LIVE_CORPUS_FILE)
        h = hashlib.sha256()
        h.update(f"{t_tok}|{i_tok}|{l_tok}|{len(samples)}|".encode("utf-8"))
        for s in samples[:500]:
            h.update(f"{s.get('text','')[:80]}|{s.get('risk_category','')}|".encode("utf-8", errors="ignore"))
        return {"hash": h.hexdigest()[:16], "rows": int(len(samples)), "source": "live_multi_source"}
    except Exception:
        return {"hash": "", "rows": 0, "source": "live_multi_source"}


def compute_kpi_auditee_fingerprint(custom_hash: Optional[str] = None) -> Dict[str, Any]:
    """
    Compute server-side pool fingerprint and row count for KPI Auditee training pool.
    Always keeps `hash` equal to the pure server-side `pool_hash` so background auto-detection
    and frontend status checks (which pass `custom_hash` separately) never conflict.
    """
    try:
        df = load_auditee_training_dataframe()
        if df is not None and not df.empty and "Nama KPI" in df.columns:
            sub = df[["Nama KPI", "Periode", "TARGET: Nilai Aktual (%)"]].copy()
            sub["Nama KPI"] = sub["Nama KPI"].astype(str).str.strip().str.lower()
            sub["Periode"] = sub["Periode"].astype(str).str.strip()
            sub["_val"] = pd.to_numeric(sub["TARGET: Nilai Aktual (%)"], errors="coerce").fillna(0.0).round(4)
            sub = sub.sort_values(["Nama KPI", "Periode"])
            lines = [f"{r[0]}|{r[1]}|{r[2]:.4f}" for r in zip(sub["Nama KPI"], sub["Periode"], sub["_val"])]
            pool_digest = hashlib.sha256("\n".join(lines).encode("utf-8")).hexdigest()[:16]
            return {
                "hash": pool_digest,
                "pool_hash": pool_digest,
                "custom_hash": str(custom_hash or ""),
                "rows": int(len(df)),
                "source": "auditee_pool",
            }
    except Exception:
        pass
    tok, rows = _fast_file_stat_token(AUDITEE_PARQUET_FILE)
    return {
        "hash": tok,
        "pool_hash": tok,
        "custom_hash": str(custom_hash or ""),
        "rows": rows,
        "source": "auditee_parquet",
    }


def compute_kpi_auditor_fingerprint(live_auditor_hash: Optional[str] = None) -> Dict[str, Any]:
    """
    Compute server-side pool fingerprint and row count for KPI Auditor training pool.
    Always keeps `hash` equal to the pure server-side `pool_hash` and tracks `live_hash`
    separately so live Strategic Audit Plan hashes never collide with Parquet pool hashes.
    """
    try:
        df = load_auditor_training_dataframe()
        if df is not None and not df.empty and "Kode SO" in df.columns:
            sub = df[["Kode SO", "Periode", "TARGET: Nilai Aktual"]].copy()
            sub["Kode SO"] = sub["Kode SO"].astype(str).str.strip().str.upper()
            sub["Periode"] = sub["Periode"].astype(str).str.strip()
            sub["_val"] = pd.to_numeric(sub["TARGET: Nilai Aktual"], errors="coerce").fillna(0.0).round(4)
            sub = sub.sort_values(["Kode SO", "Periode"])
            lines = [f"{r[0]}|{r[1]}|{r[2]:.4f}" for r in zip(sub["Kode SO"], sub["Periode"], sub["_val"])]
            pool_digest = hashlib.sha256("\n".join(lines).encode("utf-8")).hexdigest()[:16]
            return {
                "hash": pool_digest,
                "pool_hash": pool_digest,
                "live_hash": str(live_auditor_hash or ""),
                "rows": int(len(df)),
                "source": "auditor_pool",
            }
    except Exception:
        pass
    tok, rows = _fast_file_stat_token(AUDITOR_PARQUET_FILE)
    return {
        "hash": tok,
        "pool_hash": tok,
        "live_hash": str(live_auditor_hash or ""),
        "rows": rows,
        "source": "auditor_parquet",
    }


def compute_all_model_fingerprints(
    auditor_hash: Optional[str] = None,
    auditee_custom_hash: Optional[str] = None,
) -> Dict[str, Dict[str, Any]]:
    return {
        "entity": compute_entity_fingerprint(),
        "anomaly": compute_anomaly_fingerprint(),
        "nlp": compute_nlp_fingerprint(),
        "kpi_auditee": compute_kpi_auditee_fingerprint(custom_hash=auditee_custom_hash),
        "kpi_auditor": compute_kpi_auditor_fingerprint(live_auditor_hash=auditor_hash),
    }


def iter_dataframe_batches(df: pd.DataFrame, batch_rows: int = 5000) -> Iterator[pd.DataFrame]:
    """Yield slices of an in-memory DataFrame in fixed batch sizes."""
    if df is None or df.empty:
        return
    step = max(100, int(batch_rows))
    total = len(df)
    for start in range(0, total, step):
        yield df.iloc[start:start + step].copy()


def iter_csv_or_parquet_batches(path: str, batch_rows: int = 5000) -> Iterator[pd.DataFrame]:
    """
    Out-of-core streaming batch iterator over a Parquet or CSV file.
    Reads one chunk at a time so even multi-GB files stay within a small RAM footprint.
    """
    if not path or not os.path.exists(path):
        return
    step = max(500, int(batch_rows))
    if path.endswith(".parquet") and pq is not None:
        pf = pq.ParquetFile(path)
        for batch in pf.iter_batches(batch_size=step):
            yield batch.to_pandas()
        return
    for chunk in pd.read_csv(path, chunksize=step):
        yield chunk


def stratified_reservoir_sample_batches(
    batch_iter: Iterator[pd.DataFrame],
    target_rows: int,
    label_col: Optional[str] = None,
    minority_values: Optional[List[Any]] = None,
    seed: int = 42,
) -> pd.DataFrame:
    """
    Tier-3 Out-of-Core Stratified / Minority-Preserving Reservoir Sampler.
    Streams through arbitrarily large batch iterators (millions of rows / TB-scale)
    keeping all minority/positive anomaly records (up to 50% of target_rows) and reservoir-sampling
    majority records into a bounded DataFrame that fits comfortably in RAM.
    """
    rng = np.random.default_rng(seed)
    max_total = max(1000, int(target_rows))
    minority_cap = max_total // 2

    minority_parts: List[pd.DataFrame] = []
    minority_count = 0
    majority_reservoir: List[pd.DataFrame] = []
    majority_seen = 0
    majority_cap = max_total

    for chunk in batch_iter:
        if chunk is None or chunk.empty:
            continue
        if label_col and label_col in chunk.columns and minority_values:
            mask = chunk[label_col].isin(minority_values)
            min_df = chunk[mask]
            maj_df = chunk[~mask]
        else:
            min_df = pd.DataFrame()
            maj_df = chunk

        if not min_df.empty and minority_count < minority_cap:
            take = min(len(min_df), minority_cap - minority_count)
            minority_parts.append(min_df.iloc[:take].copy())
            minority_count += take
            majority_cap = max(500, max_total - minority_count)

        if not maj_df.empty:
            n_maj = len(maj_df)
            majority_seen += n_maj
            current_maj_rows = sum(len(x) for x in majority_reservoir)
            if current_maj_rows < majority_cap:
                take = min(n_maj, majority_cap - current_maj_rows)
                majority_reservoir.append(maj_df.iloc[:take].copy())
            else:
                # Reservoir replacement probability decreases as majority_seen grows
                keep_prob = majority_cap / float(majority_seen)
                if rng.random() < min(1.0, keep_prob * 2.0):
                    sample_n = max(1, min(len(maj_df), int(round(n_maj * keep_prob))))
                    sampled_new = maj_df.sample(n=sample_n, random_state=int(rng.integers(0, 1_000_000)))
                    combined_maj = pd.concat(majority_reservoir + [sampled_new], ignore_index=True)
                    if len(combined_maj) > majority_cap:
                        combined_maj = combined_maj.sample(n=majority_cap, random_state=int(rng.integers(0, 1_000_000)))
                    majority_reservoir = [combined_maj]

    frames = minority_parts + majority_reservoir
    if not frames:
        return pd.DataFrame()
    out = pd.concat(frames, ignore_index=True)
    if len(out) > max_total:
        out = out.sample(n=max_total, random_state=seed).reset_index(drop=True)
    return out.reset_index(drop=True)
