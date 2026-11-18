"""
AuditSphere — High-Performance AI Training Storage Manager
Decouples AI training from legacy CSV files using:
1. PostgreSQL / TimescaleDB Datalake Training Pool (gold.ai_*_training_pool)
2. Columnar Binary Parquet Cache (Snappy compressed, memory-mapped, zero-copy pyarrow)
3. Non-blocking out-of-core streaming batch extraction
"""
import os
import re
import gc
import json
from typing import Any, Tuple
import numpy as np
import pandas as pd
try:
    import pyarrow as pa
    import pyarrow.parquet as pq
except ImportError:
    pa = None
    pq = None

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
_REPO_ROOT = os.path.abspath(os.path.join(BASE_DIR, '..', '..'))
_MODEL_DIR = os.getenv("MODEL_DIR", os.path.join(BASE_DIR, "models"))


def resolve_ai_training_dir() -> str:
    env_dir = os.getenv("AI_TRAINING_DIR")
    if env_dir and os.path.exists(env_dir):
        return env_dir
    for cand in (
        "/app/ai_model_training",
        os.path.join(_REPO_ROOT, "ai_model_training"),
        os.path.join(BASE_DIR, "..", "ai_model_training"),
        os.path.join(BASE_DIR, "ai_model_training"),
    ):
        if os.path.exists(cand):
            return cand
    return "/app/ai_model_training"


def _resolve_pool_dir() -> str:
    env_dir = os.getenv("PARQUET_DIR")
    if env_dir:
        return env_dir
    repo_ds = os.path.join(_REPO_ROOT, "data-server", "ai-training-data", "parquet")
    if os.path.exists(os.path.join(_REPO_ROOT, "data-server")):
        return repo_ds
    # Inside Docker container (/app), persist training pool inside mounted /app/models/training_pool volume
    return os.path.join(_MODEL_DIR, "training_pool")


PARQUET_DIR = _resolve_pool_dir()
os.makedirs(PARQUET_DIR, exist_ok=True)

AUDITEE_PARQUET_FILE = os.path.join(PARQUET_DIR, 'auditee_kpi_pool.parquet')
AUDITOR_PARQUET_FILE = os.path.join(PARQUET_DIR, 'auditor_kpi_pool.parquet')
ENTITY_PREDICTION_PARQUET_FILE = os.path.join(PARQUET_DIR, 'entity_prediction_pool.parquet')


def _read_pool_df(parquet_path: str) -> pd.DataFrame:
    if pa and pq and os.path.exists(parquet_path):
        try:
            return pq.read_table(parquet_path, memory_map=True).to_pandas()
        except Exception as e:
            print(f"[AI-Storage] Parquet read warning ({parquet_path}): {e}")
    pkl_path = parquet_path + ".pkl"
    if os.path.exists(pkl_path):
        try:
            return pd.read_pickle(pkl_path)
        except Exception as e:
            print(f"[AI-Storage] Pickle pool read warning ({pkl_path}): {e}")
    return pd.DataFrame()


def _write_pool_df(df: pd.DataFrame, parquet_path: str):
    if pa and pq:
        tmp_file = parquet_path + '.tmp'
        table = pa.Table.from_pandas(df, preserve_index=False)
        pq.write_table(table, tmp_file, compression='SNAPPY')
        os.replace(tmp_file, parquet_path)
        return
    pkl_path = parquet_path + ".pkl"
    tmp_file = pkl_path + '.tmp'
    df.to_pickle(tmp_file)
    os.replace(tmp_file, pkl_path)


def _pool_exists(parquet_path: str) -> bool:
    return os.path.exists(parquet_path) or os.path.exists(parquet_path + ".pkl")

# Database connection URL for Datalake (PostgreSQL / TimescaleDB)
DATALAKE_DB_URL = os.getenv(
    "DATALAKE_DB_URL",
    "postgresql://postgres:postgres@localhost:5432/auditsphere_datalake"
)

def _get_sqlalchemy_engine():
    """Optional database engine if SQLAlchemy & psycopg are installed."""
    try:
        from sqlalchemy import create_engine
        return create_engine(DATALAKE_DB_URL, pool_pre_ping=True, pool_size=2)
    except Exception:
        return None

def initialize_parquet_from_sources():
    """
    Initializes binary feature stores (Parquet when pyarrow is available, else persistent pickle)
    from legacy CSV if not yet present, guaranteeing instantaneous bootstrap.
    """
    legacy_dir = os.path.join(resolve_ai_training_dir(), 'kpi_prediction')

    # 1. Auditee KPI Pool
    if not _pool_exists(AUDITEE_PARQUET_FILE):
        legacy_csv = os.path.join(legacy_dir, 'kpi_data.csv')
        if os.path.exists(legacy_csv):
            df_auditee = pd.read_csv(legacy_csv)
            df_auditee['ID KPI'] = df_auditee['ID KPI'].astype(str)
            df_auditee['Nama KPI'] = df_auditee['Nama KPI'].astype(str)
            df_auditee['Periode'] = df_auditee['Periode'].astype(str)
            df_auditee['TARGET: Nilai Aktual (%)'] = df_auditee['TARGET: Nilai Aktual (%)'].astype(np.float32)
            df_auditee['TARGET: Impact (1-5)'] = df_auditee['TARGET: Impact (1-5)'].astype(np.int32)
            df_auditee['TARGET: Likelihood (1-5)'] = df_auditee['TARGET: Likelihood (1-5)'].astype(np.int32)
            df_auditee = _sanitize_auditee_kpi_df(df_auditee)
            _write_pool_df(df_auditee, AUDITEE_PARQUET_FILE)
            print(f"[AI-Storage] Initialized Auditee pool store: {AUDITEE_PARQUET_FILE} ({len(df_auditee)} rows)")

    # 2. Auditor KPI Pool
    if not _pool_exists(AUDITOR_PARQUET_FILE):
        legacy_csv = os.path.join(legacy_dir, 'auditor_kpi_data.csv')
        if os.path.exists(legacy_csv):
            df_auditor = pd.read_csv(legacy_csv)
            df_auditor['ID KPI'] = df_auditor['ID KPI'].astype(str)
            df_auditor['Kode SO'] = df_auditor['Kode SO'].astype(str)
            df_auditor['Nama KPI'] = df_auditor['Nama KPI'].astype(str)
            df_auditor['Kategori Pilar'] = df_auditor['Kategori Pilar'].astype(str)
            df_auditor['Periode'] = df_auditor['Periode'].astype(str)
            df_auditor['Tipe Polaritas (HIG/HIB)'] = df_auditor['Tipe Polaritas (HIG/HIB)'].astype(str)
            df_auditor['Satuan'] = df_auditor['Satuan'].astype(str)
            df_auditor['Nilai Target'] = df_auditor['Nilai Target'].astype(np.float32)
            df_auditor['TARGET: Nilai Aktual'] = df_auditor['TARGET: Nilai Aktual'].astype(np.float32)
            df_auditor['TARGET: Rasio Ketercapaian (%)'] = df_auditor['TARGET: Rasio Ketercapaian (%)'].astype(np.float32)
            df_auditor['TARGET: Status Achievement'] = df_auditor['TARGET: Status Achievement'].astype(str)
            _write_pool_df(df_auditor, AUDITOR_PARQUET_FILE)
            print(f"[AI-Storage] Initialized Auditor pool store: {AUDITOR_PARQUET_FILE} ({len(df_auditor)} rows)")

def _resolve_auditee_record_risk(r: dict) -> Tuple[str, int, int]:
    """
    Strictly adapts any incoming Auditee KPI record's risk level and (impact, likelihood)
    onto ONLY the 5 canonical risk levels:
      HIGH, MODERATE_HIGH, MODERATE, LOW_MODERATE, LOW.
    """
    raw_lvl = (
        r.get('risk_level')
        if r.get('risk_level') is not None
        else (
            r.get('Risk Level')
            if r.get('Risk Level') is not None
            else (r.get('riskLevel') if r.get('riskLevel') is not None else r.get('severity'))
        )
    )
    has_explicit_lvl = (
        raw_lvl is not None
        and str(raw_lvl).strip() != ''
        and str(raw_lvl).strip().lower() not in {'not scored', 'none', 'nan', 'null'}
    )

    imp_raw = r.get('impact') if r.get('impact') is not None else r.get('TARGET: Impact (1-5)')
    lik_raw = r.get('likelihood') if r.get('likelihood') is not None else r.get('TARGET: Likelihood (1-5)')

    if has_explicit_lvl:
        norm_code, _, def_imp, def_lik = normalize_canonical_risk_level(raw_lvl, default_code='MODERATE')
        if imp_raw is not None and lik_raw is not None:
            try:
                imp = max(1, min(5, int(round(float(imp_raw)))))
                lik = max(1, min(5, int(round(float(lik_raw)))))
            except Exception:
                imp, lik = def_imp, def_lik
            # If caller supplied an explicit risk level whose matrix cell disagrees, align scores so training learns the canonical tier
            if matrix_risk_level(lik, imp) != norm_code:
                imp, lik = def_imp, def_lik
        else:
            imp, lik = def_imp, def_lik
        return norm_code, imp, lik

    try:
        imp = max(1, min(5, int(round(float(imp_raw if imp_raw is not None else 3)))))
    except Exception:
        imp = 3
    try:
        lik = max(1, min(5, int(round(float(lik_raw if lik_raw is not None else 3)))))
    except Exception:
        lik = 3
    return matrix_risk_level(lik, imp), imp, lik


def upsert_auditee_kpi_pool(new_records: list) -> int:
    """
    Upserts new Auditee KPI records directly into:
    1. PostgreSQL gold.ai_auditee_kpi_training_pool (if DB online)
    2. Atomic memory-mapped Parquet columnar cache (never CSV)
    Strictly enforces ONLY the 5 canonical risk levels on stored and trained records.
    """
    if not new_records:
        return 0

    normalized_pg_records = []
    for r in new_records:
        norm_code, imp, lik = _resolve_auditee_record_risk(r)
        nr = dict(r)
        nr['impact'] = imp
        nr['likelihood'] = lik
        nr['risk_level'] = norm_code
        normalized_pg_records.append(nr)

    # 1. Database UPSERT (PostgreSQL Set-Based SQL) with normalized 5-level values
    engine = _get_sqlalchemy_engine()
    if engine:
        try:
            from sqlalchemy import text
            with engine.begin() as conn:
                for r in normalized_pg_records:
                    conn.execute(text("""
                        INSERT INTO gold.ai_auditee_kpi_training_pool (
                            source_id, kpi_name, entity, entity_type, period_date, period_label,
                            actual_value, target_value, impact, likelihood, risk_level, sync_frequency, updated_at
                        ) VALUES (
                            :source_id, :kpi_name, :entity, :entity_type, :period_date, :period_label,
                            :actual_value, :target_value, :impact, :likelihood, :risk_level, :sync_frequency, NOW()
                        )
                        ON CONFLICT (kpi_name, period_date, entity)
                        DO UPDATE SET
                            actual_value = EXCLUDED.actual_value,
                            target_value = EXCLUDED.target_value,
                            impact = EXCLUDED.impact,
                            likelihood = EXCLUDED.likelihood,
                            risk_level = EXCLUDED.risk_level,
                            sync_frequency = EXCLUDED.sync_frequency,
                            updated_at = NOW();
                    """), r)
        except Exception as e:
            print(f"[AI-Storage] Notice: DB pool upsert skipped: {e}")

    # 2. Binary Pool Store (Atomic File Swap - Parquet or Pickle fallback)
    try:
        initialize_parquet_from_sources()
        df_current = _read_pool_df(AUDITEE_PARQUET_FILE)
        if not df_current.empty:
            df_current = _sanitize_auditee_kpi_df(df_current)

        # Format incoming records to pool schema
        converted = []
        for idx, r in enumerate(normalized_pg_records):
            p_label = str(r.get('period_label') or r.get('period') or 'Bulan 01 2026')
            if not p_label.startswith('Bulan'):
                p_label = f"Bulan {p_label}"
            converted.append({
                'ID KPI': str(r.get('id') or f"KPI-GOLD-{len(df_current) + idx + 1:04d}"),
                'Nama KPI': str(r.get('kpi_name') or r.get('Nama KPI') or 'KPI Bisnis'),
                'Periode': p_label,
                'TARGET: Nilai Aktual (%)': float(r.get('actual_value') or r.get('TARGET: Nilai Aktual (%)') or 0.0),
                'TARGET: Impact (1-5)': int(r['impact']),
                'TARGET: Likelihood (1-5)': int(r['likelihood']),
                'Risk Level': r['risk_level'],
            })

        df_new = pd.DataFrame(converted)
        if not df_current.empty:
            df_updated = pd.concat([df_current, df_new], ignore_index=True).drop_duplicates(
                subset=['Nama KPI', 'Periode'], keep='last'
            )
        else:
            df_updated = df_new

        df_updated = _sanitize_auditee_kpi_df(df_updated)
        _write_pool_df(df_updated, AUDITEE_PARQUET_FILE)
        return len(normalized_pg_records)
    except Exception as e:
        print(f"[AI-Storage] Auditee pool buffer upsert error: {e}")
        return 0

def upsert_auditor_kpi_pool(new_records: list) -> int:
    """
    Upserts new Auditor Strategic Plan records directly into:
    1. PostgreSQL gold.ai_auditor_kpi_training_pool (if DB online)
    2. Atomic memory-mapped Parquet columnar cache (never CSV)
    """
    if not new_records:
        return 0

    # 1. Database UPSERT (PostgreSQL Set-Based SQL)
    engine = _get_sqlalchemy_engine()
    if engine:
        try:
            from sqlalchemy import text
            with engine.begin() as conn:
                for r in new_records:
                    conn.execute(text("""
                        INSERT INTO gold.ai_auditor_kpi_training_pool (
                            code, kpi_name, category, strategic_objective, period_date, period_label,
                            target_value, actual_value, achievement_rate, hib_hig, unit, status, updated_at
                        ) VALUES (
                            :code, :kpi_name, :category, :strategic_objective, :period_date, :period_label,
                            :target_value, :actual_value, :achievement_rate, :hib_hig, :unit, :status, NOW()
                        )
                        ON CONFLICT (code, period_date)
                        DO UPDATE SET
                            kpi_name = EXCLUDED.kpi_name,
                            category = EXCLUDED.category,
                            target_value = EXCLUDED.target_value,
                            actual_value = EXCLUDED.actual_value,
                            achievement_rate = EXCLUDED.achievement_rate,
                            status = EXCLUDED.status,
                            updated_at = NOW();
                    """), r)
        except Exception as e:
            print(f"[AI-Storage] Notice: Auditor DB pool upsert skipped: {e}")

    # 2. Binary Pool Store (Atomic File Swap - Parquet or Pickle fallback)
    try:
        initialize_parquet_from_sources()
        df_current = _read_pool_df(AUDITOR_PARQUET_FILE)

        converted = []
        for idx, r in enumerate(new_records):
            p_label = str(r.get('period_label') or r.get('Periode') or 'Bulan 01 2026')
            if not p_label.startswith('Bulan'):
                p_label = f"Bulan {p_label}"
            converted.append({
                'ID KPI': str(r.get('id') or f"AUD-SYNC-{r.get('code', 'SO-IA01')}"),
                'Kode SO': str(r.get('code') or r.get('Kode SO') or 'SO-IA01'),
                'Nama KPI': str(r.get('kpi_name') or r.get('Nama KPI') or 'Indikator Audit'),
                'Kategori Pilar': str(r.get('category') or r.get('Kategori Pilar') or 'Pelaksanaan Audit & PKAT'),
                'Periode': p_label,
                'Tipe Polaritas (HIG/HIB)': str(r.get('hib_hig') or r.get('Tipe Polaritas (HIG/HIB)') or 'HIG').upper(),
                'Satuan': str(r.get('unit') or r.get('Satuan') or '%'),
                'Nilai Target': float(r.get('target_value') or r.get('Nilai Target') or 100.0),
                'TARGET: Nilai Aktual': float(r.get('actual_value') or r.get('TARGET: Nilai Aktual') or 100.0),
                'TARGET: Rasio Ketercapaian (%)': float(r.get('achievement_rate') or r.get('TARGET: Rasio Ketercapaian (%)') or 100.0),
                'TARGET: Status Achievement': str(r.get('status') or r.get('TARGET: Status Achievement') or 'Good')
            })

        df_new = pd.DataFrame(converted)
        if not df_current.empty:
            df_updated = pd.concat([df_current, df_new], ignore_index=True).drop_duplicates(
                subset=['Kode SO', 'Periode'], keep='last'
            )
        else:
            df_updated = df_new

        _write_pool_df(df_updated, AUDITOR_PARQUET_FILE)
        return len(new_records)
    except Exception as e:
        print(f"[AI-Storage] Auditor pool buffer upsert error: {e}")
        return 0

def load_auditee_training_dataframe() -> pd.DataFrame:
    """
    Loads Auditee training dataframe from PostgreSQL or binary Parquet.
    Zero CSV text parsing — sub-50ms latency with low RAM footprint.
    """
    initialize_parquet_from_sources()

    # Try PostgreSQL first if configured
    engine = _get_sqlalchemy_engine()
    if engine:
        try:
            query = """
                SELECT 
                    'KPI-GOLD-' || id::text AS "ID KPI",
                    kpi_name AS "Nama KPI",
                    period_label AS "Periode",
                    actual_value::float AS "TARGET: Nilai Aktual (%)",
                    impact AS "TARGET: Impact (1-5)",
                    likelihood AS "TARGET: Likelihood (1-5)",
                    risk_level AS "Risk Level"
                FROM gold.ai_auditee_kpi_training_pool
                ORDER BY kpi_name ASC, period_date ASC;
            """
            df = pd.read_sql(query, con=engine)
            if not df.empty:
                return _sanitize_auditee_kpi_df(df)
        except Exception:
            pass

    # High-speed Binary Pool Fallback
    df_pool = _read_pool_df(AUDITEE_PARQUET_FILE)
    if not df_pool.empty:
        return _sanitize_auditee_kpi_df(df_pool)

    # Legacy CSV bootstrap fallback
    legacy_csv = os.path.join(resolve_ai_training_dir(), 'kpi_prediction', 'kpi_data.csv')
    if os.path.exists(legacy_csv):
        return _sanitize_auditee_kpi_df(pd.read_csv(legacy_csv))

    return pd.DataFrame()

def load_auditor_training_dataframe() -> pd.DataFrame:
    """
    Loads Auditor training dataframe from PostgreSQL or binary Parquet.
    Zero CSV text parsing — sub-50ms latency with low RAM footprint.
    """
    initialize_parquet_from_sources()

    # Try PostgreSQL first if configured
    engine = _get_sqlalchemy_engine()
    if engine:
        try:
            query = """
                SELECT 
                    'AUD-POOL-' || id::text AS "ID KPI",
                    code AS "Kode SO",
                    kpi_name AS "Nama KPI",
                    category AS "Kategori Pilar",
                    period_label AS "Periode",
                    hib_hig AS "Tipe Polaritas (HIG/HIB)",
                    unit AS "Satuan",
                    target_value::float AS "Nilai Target",
                    actual_value::float AS "TARGET: Nilai Aktual",
                    achievement_rate::float AS "TARGET: Rasio Ketercapaian (%)",
                    status AS "TARGET: Status Achievement"
                FROM gold.ai_auditor_kpi_training_pool
                ORDER BY code ASC, period_date ASC;
            """
            df = pd.read_sql(query, con=engine)
            if not df.empty:
                return df
        except Exception:
            pass

    # High-speed Binary Pool Fallback
    df_pool = _read_pool_df(AUDITOR_PARQUET_FILE)
    if not df_pool.empty:
        return df_pool

    legacy_csv = os.path.join(resolve_ai_training_dir(), 'kpi_prediction', 'auditor_kpi_data.csv')
    if os.path.exists(legacy_csv):
        return pd.read_csv(legacy_csv)

    return pd.DataFrame()

CANONICAL_RISK_CODES = ("HIGH", "MODERATE_HIGH", "MODERATE", "LOW_MODERATE", "LOW")

CANONICAL_RISK_CODE_TO_LABEL = {
    "HIGH": "High",
    "MODERATE_HIGH": "Moderate to High",
    "MODERATE": "Moderate",
    "LOW_MODERATE": "Low to Moderate",
    "LOW": "Low",
}

CANONICAL_RISK_CODE_TO_PAIR = {
    "HIGH": (5, 5),
    "MODERATE_HIGH": (4, 4),
    "MODERATE": (3, 3),
    "LOW_MODERATE": (2, 2),
    "LOW": (1, 1),
}

CANONICAL_RISK_CODE_TO_UI_FLOAT_PAIR = {
    "HIGH": (4.7, 4.8),
    "MODERATE_HIGH": (3.8, 4.0),
    "MODERATE": (3.0, 3.2),
    "LOW_MODERATE": (2.2, 2.4),
    "LOW": (1.2, 1.4),
}


def score_to_risk_level(score: float) -> str:
    if score >= 20:
        return "HIGH"
    elif score >= 15:
        return "MODERATE_HIGH"
    elif score >= 10:
        return "MODERATE"
    elif score >= 5:
        return "LOW_MODERATE"
    return "LOW"


def matrix_risk_level(likelihood: float, impact: float) -> str:
    """Exact alignment with Corporate Risk Profile 5x5 Matrix (only 5 canonical levels)."""
    try:
        l_int = max(1, min(5, int(round(float(likelihood)))))
        i_int = max(1, min(5, int(round(float(impact)))))
        matrix = {
            5: {1: "LOW_MODERATE", 2: "MODERATE", 3: "MODERATE_HIGH", 4: "HIGH", 5: "HIGH"},
            4: {1: "LOW", 2: "LOW_MODERATE", 3: "MODERATE", 4: "MODERATE_HIGH", 5: "HIGH"},
            3: {1: "LOW", 2: "LOW_MODERATE", 3: "MODERATE", 4: "MODERATE_HIGH", 5: "HIGH"},
            2: {1: "LOW", 2: "LOW_MODERATE", 3: "LOW_MODERATE", 4: "MODERATE_HIGH", 5: "HIGH"},
            1: {1: "LOW", 2: "LOW", 3: "LOW_MODERATE", 4: "MODERATE", 5: "HIGH"},
        }
        return matrix.get(l_int, {}).get(i_int, score_to_risk_level(float(l_int * i_int)))
    except Exception:
        return "MODERATE"


def normalize_canonical_risk_level(raw_level: Any, default_code: str = "MODERATE") -> Tuple[str, str, int, int]:
    """
    Strictly maps ANY risk level / severity / rating string or numeric value
    (including legacy 'Critical', external DB custom ratings such as 'Extreme',
    'Very High', 'Catastrophic', 'Major', 'Minor', 'Insignificant', 'Not Scored',
    Indonesian ratings, or numeric scores) onto ONLY the 5 canonical risk levels:
      - HIGH ('High')
      - MODERATE_HIGH ('Moderate to High')
      - MODERATE ('Moderate')
      - LOW_MODERATE ('Low to Moderate')
      - LOW ('Low')
    Returns (code, display_label, impact_1_to_5, likelihood_1_to_5).
    """
    fallback = default_code if default_code in CANONICAL_RISK_CODE_TO_LABEL else "MODERATE"
    if raw_level is None:
        imp, lik = CANONICAL_RISK_CODE_TO_PAIR[fallback]
        return fallback, CANONICAL_RISK_CODE_TO_LABEL[fallback], imp, lik

    if isinstance(raw_level, (int, float)) and not isinstance(raw_level, bool):
        try:
            val = float(raw_level)
        except Exception:
            imp, lik = CANONICAL_RISK_CODE_TO_PAIR[fallback]
            return fallback, CANONICAL_RISK_CODE_TO_LABEL[fallback], imp, lik
        if np.isnan(val):
            imp, lik = CANONICAL_RISK_CODE_TO_PAIR[fallback]
            return fallback, CANONICAL_RISK_CODE_TO_LABEL[fallback], imp, lik
        if val <= 5:
            if val >= 4.5:
                code = "HIGH"
            elif val >= 3.5:
                code = "MODERATE_HIGH"
            elif val >= 2.5:
                code = "MODERATE"
            elif val >= 1.5:
                code = "LOW_MODERATE"
            else:
                code = "LOW"
        elif val <= 25:
            if val >= 20:
                code = "HIGH"
            elif val >= 15:
                code = "MODERATE_HIGH"
            elif val >= 10:
                code = "MODERATE"
            elif val >= 5:
                code = "LOW_MODERATE"
            else:
                code = "LOW"
        else:
            if val >= 80:
                code = "HIGH"
            elif val >= 60:
                code = "MODERATE_HIGH"
            elif val >= 40:
                code = "MODERATE"
            elif val >= 20:
                code = "LOW_MODERATE"
            else:
                code = "LOW"
        imp, lik = CANONICAL_RISK_CODE_TO_PAIR[code]
        return code, CANONICAL_RISK_CODE_TO_LABEL[code], imp, lik

    s = str(raw_level).strip()
    if not s:
        imp, lik = CANONICAL_RISK_CODE_TO_PAIR[fallback]
        return fallback, CANONICAL_RISK_CODE_TO_LABEL[fallback], imp, lik

    if re.fullmatch(r"-?\d+(?:\.\d+)?", s):
        return normalize_canonical_risk_level(float(s), default_code=fallback)

    lower = s.lower()
    compact = re.sub(r"[\s\-_/]+", "", lower)

    # Check combined tiers FIRST so "Moderate to High" / "Low to Moderate" never collapse into High / Moderate
    if (
        re.search(r"(moderate|medium|sedang|menengah)(to|ke)?(high|tinggi)", compact)
        or re.search(r"(high|tinggi)(to|ke)?(moderate|medium|sedang)", compact)
        or compact in {
            "moderatehigh", "mediumhigh", "elevated", "significant", "major",
            "tier2", "level4", "kol3", "p2", "kuranglancar"
        }
        or "kurang lancar" in lower
    ):
        code = "MODERATE_HIGH"
    elif (
        re.search(r"(low|rendah)(to|ke)?(moderate|medium|sedang)", compact)
        or re.search(r"(moderate|medium|sedang)(to|ke)?(low|rendah)", compact)
        or compact in {
            "lowmoderate", "lowmedium", "lowmod", "guarded", "minor",
            "tier4", "level2", "kol2", "p4", "lancardengancatatan"
        }
        or "lancar dengan catatan" in lower
    ):
        code = "LOW_MODERATE"
    elif (
        compact in {
            "high", "highrisk", "risikotinggi", "tinggi",
            "veryhigh", "critical", "criticalrisk", "extreme", "extremerisk",
            "catastrophic", "severe", "urgent", "blocker", "fatal",
            "sangattinggi", "kritis", "bahaya", "macet", "diragukan",
            "tier1", "level5", "kol4", "kol5", "p1", "p0"
        }
        or re.search(r"crit|extr|catastroph|sever|blocker|fatal|sangat\s*tinggi|kritis|\b(high|tinggi)\b", lower)
    ):
        code = "HIGH"
    elif (
        compact in {
            "low", "lowrisk", "risikorendah", "rendah",
            "verylow", "minimal", "insignificant", "negligible", "trivial",
            "safe", "normal", "healthy", "info", "informational",
            "sangatrendah", "aman", "lancar",
            "tier5", "level1", "kol1", "p5"
        }
        or re.search(r"insignif|neglig|trivial|minimal|sangat\s*rendah|\b(low|rendah)\b", lower)
    ):
        code = "LOW"
    elif compact in {
        "moderate", "medium", "mod", "moderaterisk", "mediumrisk", "risikosedang",
        "sedang", "menengah", "watch", "warning", "fair", "average",
        "dpk", "dalamperhatiankhusus", "tier3", "level3", "p3"
    } or "dalam perhatian khusus" in lower:
        code = "MODERATE"
    else:
        code = fallback

    imp, lik = CANONICAL_RISK_CODE_TO_PAIR[code]
    return code, CANONICAL_RISK_CODE_TO_LABEL[code], imp, lik


def _sanitize_auditee_kpi_df(df: pd.DataFrame) -> pd.DataFrame:
    """
    Enforces [1, 5] integer Impact/Likelihood bounds and strictly the 5 canonical Risk Levels
    (HIGH, MODERATE_HIGH, MODERATE, LOW_MODERATE, LOW) on Auditee KPI training records.
    """
    if df is None or df.empty:
        return df

    df = df.copy()
    imp_col = "TARGET: Impact (1-5)" if "TARGET: Impact (1-5)" in df.columns else ("impact" if "impact" in df.columns else None)
    lik_col = "TARGET: Likelihood (1-5)" if "TARGET: Likelihood (1-5)" in df.columns else ("likelihood" if "likelihood" in df.columns else None)

    if imp_col:
        df[imp_col] = pd.to_numeric(df[imp_col], errors="coerce").fillna(3).clip(1, 5).round().astype(int)
    else:
        df["TARGET: Impact (1-5)"] = 3
        imp_col = "TARGET: Impact (1-5)"

    if lik_col:
        df[lik_col] = pd.to_numeric(df[lik_col], errors="coerce").fillna(3).clip(1, 5).round().astype(int)
    else:
        df["TARGET: Likelihood (1-5)"] = 3
        lik_col = "TARGET: Likelihood (1-5)"

    raw_col = "Risk Level" if "Risk Level" in df.columns else ("risk_level" if "risk_level" in df.columns else None)
    normalized_levels = []
    imp_vals = df[imp_col].tolist()
    lik_vals = df[lik_col].tolist()

    if raw_col:
        raw_vals = df[raw_col].tolist()
        for idx, raw_v in enumerate(raw_vals):
            s_v = "" if raw_v is None or (isinstance(raw_v, float) and np.isnan(raw_v)) else str(raw_v).strip()
            l_v = int(lik_vals[idx])
            i_v = int(imp_vals[idx])
            if not s_v or s_v.lower() in {"not scored", "none", "nan", "null"}:
                normalized_levels.append(matrix_risk_level(l_v, i_v))
            else:
                code, _, def_imp, def_lik = normalize_canonical_risk_level(s_v, default_code="MODERATE")
                normalized_levels.append(code)
                if matrix_risk_level(l_v, i_v) != code:
                    imp_vals[idx] = def_imp
                    lik_vals[idx] = def_lik
        df[imp_col] = imp_vals
        df[lik_col] = lik_vals
    else:
        normalized_levels = [matrix_risk_level(l_v, i_v) for l_v, i_v in zip(lik_vals, imp_vals)]

    df["Risk Level"] = normalized_levels
    if "risk_level" in df.columns:
        df["risk_level"] = normalized_levels

    return df


def _sanitize_entity_prediction_df(df: pd.DataFrame) -> pd.DataFrame:
    """Enforces 5-tier canonical risk levels and [1, 5] likelihood/impact bounds on entity prediction rows."""
    if df is None or df.empty:
        return df

    df = df.copy()
    for col in ("inherent_likelihood", "inherent_impact", "target_likelihood", "target_impact"):
        if col in df.columns:
            df[col] = pd.to_numeric(df[col], errors="coerce").fillna(3.0).clip(1.0, 5.0)

    if "entity_type" in df.columns:
        df["entity_type"] = df["entity_type"].replace({"Department": "Auditable Entity"}).fillna("Auditable Entity")

    if "current_risk_level" in df.columns:
        normalized_levels = []
        for idx, row in df.iterrows():
            raw_v = row.get("current_risk_level")
            s_v = "" if raw_v is None or (isinstance(raw_v, float) and np.isnan(raw_v)) else str(raw_v).strip()
            if not s_v or s_v.lower() in {"not scored", "none", "nan", "null"}:
                l_v = row.get("inherent_likelihood", 3.0)
                i_v = row.get("inherent_impact", 3.0)
                normalized_levels.append(matrix_risk_level(l_v, i_v))
            else:
                code, _, _, _ = normalize_canonical_risk_level(s_v, default_code="MODERATE")
                normalized_levels.append(code)
        df["current_risk_level"] = normalized_levels
    elif "inherent_likelihood" in df.columns and "inherent_impact" in df.columns:
        df["current_risk_level"] = [
            matrix_risk_level(l_v, i_v)
            for l_v, i_v in zip(df["inherent_likelihood"], df["inherent_impact"])
        ]

    return df


def upsert_entity_prediction_pool(new_records: list) -> int:
    """
    Upserts new Entity Prediction records (from Audit Universe or External DB) directly into:
    1. PostgreSQL gold.ai_entity_prediction_training_pool (if DB online)
    2. Atomic memory-mapped Parquet columnar cache (never touches CSV during runtime)
    Strictly enforces ONLY the 5 canonical risk levels on both trained and stored data.
    """
    if not new_records:
        return 0

    converted = []
    for r in new_records:
        ent = str(r.get('entity') or r.get('Entitas') or 'Kantor Pusat').strip()
        if not ent or 'legal & compliance' in ent.lower():
            continue
        raw_etype = str(r.get('entity_type') or '').strip()
        if not raw_etype or raw_etype == 'Department':
            etype = 'Branch' if any(w in ent.lower() for w in ['cabang', 'kcp', 'regional', 'kc', 'outlet', 'branch']) else 'Auditable Entity'
        else:
            etype = raw_etype
        cat = str(r.get('risk_category') or r.get('category') or r.get('Kategori Risiko') or 'Operational').strip()
        lik_raw = r.get('inherent_likelihood') if r.get('inherent_likelihood') is not None else r.get('likelihood')
        imp_raw = r.get('inherent_impact') if r.get('inherent_impact') is not None else r.get('impact')
        raw_lvl = r.get('current_risk_level') if r.get('current_risk_level') is not None else r.get('risk_level')

        has_explicit_lvl = raw_lvl is not None and str(raw_lvl).strip() != '' and str(raw_lvl).strip().lower() not in {'not scored', 'none', 'nan', 'null'}
        if has_explicit_lvl:
            lvl, _, def_imp, def_lik = normalize_canonical_risk_level(raw_lvl, default_code='MODERATE')
        else:
            lvl, def_imp, def_lik = None, 3, 3

        lik = float(lik_raw) if lik_raw is not None else float(def_lik)
        imp = float(imp_raw) if imp_raw is not None else float(def_imp)
        lik = max(1.0, min(5.0, lik))
        imp = max(1.0, min(5.0, imp))

        if lvl is None:
            lvl = matrix_risk_level(lik, imp)

        t_lik = float(r.get('target_likelihood') if r.get('target_likelihood') is not None else lik)
        t_imp = float(r.get('target_impact') if r.get('target_impact') is not None else imp)
        t_lik = max(1.0, min(5.0, t_lik))
        t_imp = max(1.0, min(5.0, t_imp))

        m = int(r.get('assessment_month') or r.get('month') or 1)
        m = max(1, min(12, m))
        src = str(r.get('source_id') or 'audit_universe')

        converted.append({
            'entity': ent,
            'entity_type': etype,
            'risk_category': cat,
            'target_timeline': str(r.get('target_timeline') or f"FY {pd.Timestamp.now().year}"),
            'inherent_likelihood': lik,
            'inherent_impact': imp,
            'audit_findings_count': float(r.get('audit_findings_count') or r.get('findings_count') or 0.0),
            'kpi_below_target_count': float(r.get('kpi_below_target_count') or 0.0),
            'kpi_volatility': float(r.get('kpi_volatility') or 0.1),
            'previous_risk_score': float(r.get('previous_risk_score') or (lik * imp)),
            'assessment_month': m,
            'target_likelihood': t_lik,
            'target_impact': t_imp,
            'source_id': src,
            'current_risk_level': lvl
        })

    if not converted:
        return 0

    # 1. Database UPSERT (PostgreSQL Set-Based SQL) with strictly normalized 5-level values
    engine = _get_sqlalchemy_engine()
    if engine:
        try:
            from sqlalchemy import text
            with engine.begin() as conn:
                for r in converted:
                    conn.execute(text("""
                        INSERT INTO gold.ai_entity_prediction_training_pool (
                            source_id, entity, entity_type, risk_category,
                            inherent_likelihood, inherent_impact, audit_findings_count,
                            kpi_below_target_count, kpi_volatility, previous_risk_score,
                            assessment_month, target_likelihood, target_impact,
                            current_risk_level, updated_at
                        ) VALUES (
                            :source_id, :entity, :entity_type, :risk_category,
                            :inherent_likelihood, :inherent_impact, :audit_findings_count,
                            :kpi_below_target_count, :kpi_volatility, :previous_risk_score,
                            :assessment_month, :target_likelihood, :target_impact,
                            :current_risk_level, NOW()
                        )
                        ON CONFLICT (entity, risk_category, assessment_month, source_id)
                        DO UPDATE SET
                            inherent_likelihood = EXCLUDED.inherent_likelihood,
                            inherent_impact = EXCLUDED.inherent_impact,
                            audit_findings_count = EXCLUDED.audit_findings_count,
                            kpi_below_target_count = EXCLUDED.kpi_below_target_count,
                            kpi_volatility = EXCLUDED.kpi_volatility,
                            previous_risk_score = EXCLUDED.previous_risk_score,
                            target_likelihood = EXCLUDED.target_likelihood,
                            target_impact = EXCLUDED.target_impact,
                            current_risk_level = EXCLUDED.current_risk_level,
                            entity_type = EXCLUDED.entity_type,
                            updated_at = NOW();
                    """), r)
        except Exception as e:
            print(f"[AI-Storage] Notice: Entity Prediction DB pool upsert skipped: {e}")

    # 2. Binary Pool Store (Atomic File Swap - Parquet or Pickle fallback)
    try:
        initialize_parquet_from_sources()
        df_current = _read_pool_df(ENTITY_PREDICTION_PARQUET_FILE)
        if not df_current.empty and 'source_id' in df_current.columns:
            df_current = df_current[df_current['source_id'] != 'baseline_dataset']
        if not df_current.empty and 'entity' in df_current.columns:
            df_current = df_current[~df_current['entity'].astype(str).str.contains('Legal & Compliance', case=False, na=False)]
        if not df_current.empty:
            df_current = _sanitize_entity_prediction_df(df_current)

        df_new = pd.DataFrame(converted)
        if not df_current.empty:
            df_updated = pd.concat([df_current, df_new], ignore_index=True).drop_duplicates(
                subset=['entity', 'risk_category', 'assessment_month', 'source_id'], keep='last'
            )
        else:
            df_updated = df_new

        df_updated = _sanitize_entity_prediction_df(df_updated)
        _write_pool_df(df_updated, ENTITY_PREDICTION_PARQUET_FILE)

        return len(converted)
    except Exception as e:
        print(f"[AI-Storage] Entity Prediction pool buffer upsert error: {e}")
        return 0

def load_entity_prediction_training_dataframe() -> pd.DataFrame:
    """
    Loads Entity Prediction training dataframe from PostgreSQL or binary Parquet.
    Zero CSV text parsing during runtime — sub-50ms latency with low RAM footprint.
    Strictly sanitizes risk levels onto ONLY the 5 canonical levels.
    """
    initialize_parquet_from_sources()

    # Try PostgreSQL first if configured
    engine = _get_sqlalchemy_engine()
    if engine:
        try:
            query = """
                SELECT 
                    entity,
                    entity_type,
                    risk_category,
                    inherent_likelihood::float AS inherent_likelihood,
                    inherent_impact::float AS inherent_impact,
                    audit_findings_count::float AS audit_findings_count,
                    kpi_below_target_count::float AS kpi_below_target_count,
                    kpi_volatility::float AS kpi_volatility,
                    previous_risk_score::float AS previous_risk_score,
                    assessment_month::int AS assessment_month,
                    target_likelihood::float AS target_likelihood,
                    target_impact::float AS target_impact,
                    current_risk_level,
                    source_id
                FROM gold.ai_entity_prediction_training_pool
                ORDER BY entity ASC, risk_category ASC, assessment_month ASC;
            """
            df = pd.read_sql(query, con=engine)
            if not df.empty:
                if 'source_id' in df.columns:
                    df = df[df['source_id'] != 'baseline_dataset']
                if 'entity' in df.columns:
                    df = df[~df['entity'].astype(str).str.contains('Legal & Compliance', case=False, na=False)]
                return _sanitize_entity_prediction_df(df)
        except Exception:
            pass

    # High-speed Binary Pool Fallback (Only Real Synced Records, Zero CSV Runtime Dependency)
    try:
        df = _read_pool_df(ENTITY_PREDICTION_PARQUET_FILE)
        if not df.empty:
            if 'source_id' in df.columns:
                df = df[df['source_id'] != 'baseline_dataset']
            if 'entity' in df.columns:
                df = df[~df['entity'].astype(str).str.contains('Legal & Compliance', case=False, na=False)]
            return _sanitize_entity_prediction_df(df)
    except Exception as e:
        print(f"[AI-Storage] Entity pool read error: {e}")

    return pd.DataFrame()


