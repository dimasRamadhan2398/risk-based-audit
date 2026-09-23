import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(42)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1200

    # Common entities
    ENTITIES = ['Jakarta Branch', 'Surabaya Branch', 'Bandung Branch', 'Medan Branch', 'Finance Dept', 'IT Dept', 'Operations Dept']

    # ======================================================================
    # 5. FIELDWORK_DATA.CSV (Fieldwork & Audit Execution)
    # ======================================================================
    print("Generating fieldwork_data.csv...")
    audit_types = ['Full Scope Audit', 'Limited Review', 'Surprise Audit', 'Compliance Audit']
    comp_duration = np.random.randint(2, 60, size=N_SAMPLES)
    hist_avg_duration = np.random.randint(15, 45, size=N_SAMPLES)
    sampling_cov = np.round(np.random.uniform(5.0, 95.0, size=N_SAMPLES), 1)
    findings_count = np.random.negative_binomial(2, 0.4, size=N_SAMPLES)
    dur_zscore = np.round((comp_duration - hist_avg_duration) / 10.0, 2)
    instant_approval = ((comp_duration <= 3) & (sampling_cov < 20.0) & np.random.binomial(1, 0.85, size=N_SAMPLES)).astype(int)

    field_anom_score = (
        0.45 * instant_approval +
        0.30 * (sampling_cov < 10.0).astype(int) +
        0.20 * (dur_zscore < -1.8).astype(int) +
        0.15 * (findings_count == 0).astype(int) * (comp_duration > 30) +
        np.random.normal(0, 0.08, size=N_SAMPLES)
    )
    is_field_anom = (field_anom_score > 0.40).astype(int)
    field_target_anom = np.where(is_field_anom == 1, 'Ya (Anomali)', 'Tidak')

    field_impact = np.clip(np.ceil(findings_count * 0.8 + (100 - sampling_cov) / 25.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
    field_likelihood = np.clip(np.ceil(instant_approval * 3.0 + np.maximum(0, -dur_zscore) * 1.2) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

    df_fieldwork = pd.DataFrame({
        'ID Penugasan': [f'ASG-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
        'Jenis Audit': np.random.choice(audit_types, N_SAMPLES),
        'Deskripsi Fieldwork': [f'Penugasan Audit Lapangan #{i+1}' for i in range(N_SAMPLES)],
        'completion_duration_days': comp_duration,
        'historical_avg_duration_days': hist_avg_duration,
        'sampling_coverage_pct': sampling_cov,
        'audit_findings_count': findings_count,
        'duration_zscore': dur_zscore,
        'is_instant_approval (1=Ya, 0=Tidak)': instant_approval,
        'TARGET: is_anomaly': field_target_anom,
        'TARGET: Impact (1-5)': field_impact,
        'TARGET: Likelihood (1-5)': field_likelihood
    })
    df_fieldwork.to_csv(os.path.join(CURRENT_DIR, 'fieldwork_data.csv'), index=False)

if __name__ == "__main__":
    generate()
