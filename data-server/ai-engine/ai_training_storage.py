"""
AuditSphere — High-Performance AI Training Storage Manager
Decouples AI training from legacy CSV files using:
1. PostgreSQL / TimescaleDB Datalake Training Pool (gold.ai_*_training_pool)
2. Columnar Binary Parquet Cache (Snappy compressed, memory-mapped, zero-copy pyarrow)
3. Non-blocking out-of-core streaming batch extraction
"""
import os
import gc
import json
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

def upsert_auditee_kpi_pool(new_records: list) -> int:
    """
    Upserts new Auditee KPI records directly into:
    1. PostgreSQL gold.ai_auditee_kpi_training_pool (if DB online)
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

        # Format incoming records to pool schema
        converted = []
        for idx, r in enumerate(new_records):
            p_label = str(r.get('period_label') or r.get('period') or 'Bulan 01 2026')
            if not p_label.startswith('Bulan'):
                p_label = f"Bulan {p_label}"
            converted.append({
                'ID KPI': str(r.get('id') or f"KPI-GOLD-{len(df_current) + idx + 1:04d}"),
                'Nama KPI': str(r.get('kpi_name') or r.get('Nama KPI') or 'KPI Bisnis'),
                'Periode': p_label,
                'TARGET: Nilai Aktual (%)': float(r.get('actual_value') or r.get('TARGET: Nilai Aktual (%)') or 0.0),
                'TARGET: Impact (1-5)': int(r.get('impact') or r.get('TARGET: Impact (1-5)') or 3),
                'TARGET: Likelihood (1-5)': int(r.get('likelihood') or r.get('TARGET: Likelihood (1-5)') or 3)
            })

        df_new = pd.DataFrame(converted)
        if not df_current.empty:
            df_updated = pd.concat([df_current, df_new], ignore_index=True).drop_duplicates(
                subset=['Nama KPI', 'Periode'], keep='last'
            )
        else:
            df_updated = df_new

        _write_pool_df(df_updated, AUDITEE_PARQUET_FILE)
        return len(new_records)
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
                    likelihood AS "TARGET: Likelihood (1-5)"
                FROM gold.ai_auditee_kpi_training_pool
                ORDER BY kpi_name ASC, period_date ASC;
            """
            df = pd.read_sql(query, con=engine)
            if not df.empty:
                return df
        except Exception:
            pass

    # High-speed Binary Pool Fallback
    df_pool = _read_pool_df(AUDITEE_PARQUET_FILE)
    if not df_pool.empty:
        return df_pool

    # Legacy CSV bootstrap fallback
    legacy_csv = os.path.join(resolve_ai_training_dir(), 'kpi_prediction', 'kpi_data.csv')
    if os.path.exists(legacy_csv):
        return pd.read_csv(legacy_csv)

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

def upsert_entity_prediction_pool(new_records: list) -> int:
    """
    Upserts new Entity Prediction records (from Risk Profile or External DB) directly into:
    1. PostgreSQL gold.ai_entity_prediction_training_pool (if DB online)
    2. Atomic memory-mapped Parquet columnar cache (never touches CSV during runtime)
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

        converted = []
        for r in new_records:
            ent = str(r.get('entity') or r.get('Entitas') or 'Kantor Pusat')
            etype = str(r.get('entity_type') or ('Branch' if any(w in ent.lower() for w in ['cabang', 'kcp', 'regional', 'kc', 'outlet']) else 'Department'))
            cat = str(r.get('risk_category') or r.get('category') or r.get('Kategori Risiko') or 'Operational')
            lik = float(r.get('inherent_likelihood') or r.get('likelihood') or 3.0)
            imp = float(r.get('inherent_impact') or r.get('impact') or 3.0)
            t_lik = float(r.get('target_likelihood') or lik)
            t_imp = float(r.get('target_impact') or imp)
            m = int(r.get('assessment_month') or r.get('month') or 1)
            src = str(r.get('source_id') or 'risk_profile')
            lvl = str(r.get('current_risk_level') or r.get('risk_level') or 'MODERATE')
            
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

        df_new = pd.DataFrame(converted)
        if not df_current.empty:
            df_updated = pd.concat([df_current, df_new], ignore_index=True).drop_duplicates(
                subset=['entity', 'risk_category', 'assessment_month', 'source_id'], keep='last'
            )
        else:
            df_updated = df_new

        _write_pool_df(df_updated, ENTITY_PREDICTION_PARQUET_FILE)

        return len(new_records)
    except Exception as e:
        print(f"[AI-Storage] Entity Prediction pool buffer upsert error: {e}")
        return 0

def load_entity_prediction_training_dataframe() -> pd.DataFrame:
    """
    Loads Entity Prediction training dataframe from PostgreSQL or binary Parquet.
    Zero CSV text parsing during runtime — sub-50ms latency with low RAM footprint.
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
                return df
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
            return df
    except Exception as e:
        print(f"[AI-Storage] Entity pool read error: {e}")

    return pd.DataFrame()


