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
    # 8. REPEAT_FINDING_DATA.CSV (Recurring Audit Findings)
    # ======================================================================
    print("Generating repeat_finding_data.csv...")
    risk_cats = ['Operasional', 'Finansial', 'Legal & Kepatuhan', 'Reputasi', 'Keamanan TI']
    severities = ['Rendah', 'Sedang', 'Tinggi', 'Kritis']
    repeat_count = np.random.negative_binomial(1, 0.4, size=N_SAMPLES)
    similarity_pct = np.round(np.random.uniform(20.0, 99.0, size=N_SAMPLES), 1)
    same_root = np.random.binomial(1, 0.35, size=N_SAMPLES)
    prev_closed = np.random.binomial(1, 0.60, size=N_SAMPLES)

    rep_anom_score = (
        0.35 * (repeat_count >= 2).astype(int) +
        0.30 * (similarity_pct > 80.0).astype(int) * same_root +
        0.25 * (prev_closed & (similarity_pct > 75.0)).astype(int) +
        np.random.normal(0, 0.08, size=N_SAMPLES)
    )
    is_rep_anom = (rep_anom_score > 0.40).astype(int)
    rep_target_anom = np.where(is_rep_anom == 1, 'Ya (Anomali)', 'Tidak')

    rep_impact = np.clip(np.ceil(repeat_count * 1.2 + same_root * 1.5) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
    rep_likelihood = np.clip(np.ceil((similarity_pct / 25.0) + prev_closed * 1.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

    df_repeat = pd.DataFrame({
        'ID Temuan': [f'FND-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
        'Kategori Risiko': np.random.choice(risk_cats, N_SAMPLES),
        'Deskripsi Temuan': [f'Deskripsi Temuan Audit Berulang #{i+1}' for i in range(N_SAMPLES)],
        'repeat_count_last_year': repeat_count,
        'finding_similarity_pct': similarity_pct,
        'is_same_root_cause (1=Ya, 0=Tidak)': same_root,
        'is_previously_closed (1=Ya, 0=Tidak)': prev_closed,
        'severity_category': np.random.choice(severities, N_SAMPLES),
        'TARGET: is_anomaly': rep_target_anom,
        'TARGET: Impact (1-5)': rep_impact,
        'TARGET: Likelihood (1-5)': rep_likelihood
    })
    df_repeat.to_csv(os.path.join(CURRENT_DIR, 'repeat_finding_data.csv'), index=False)

if __name__ == "__main__":
    generate()
