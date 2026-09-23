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
    # 7. MITIGATION_OVERDUE_DATA.CSV (Overdue Remediation)
    # ======================================================================
    print("Generating mitigation_overdue_data.csv...")
    finding_cats = ['Kepatuhan SOP', 'Pengendalian Internal', 'Sistem Informasi', 'Manajemen Risiko']
    severities = ['Rendah', 'Sedang', 'Tinggi', 'Kritis']
    overdue_days = np.random.randint(1, 180, size=N_SAMPLES)
    repeated_overdue = np.random.randint(0, 6, size=N_SAMPLES)
    progress_pct = np.round(np.random.uniform(0.0, 95.0, size=N_SAMPLES), 1)
    unjustified_delay = ((overdue_days > 45) & (progress_pct < 25.0) & np.random.binomial(1, 0.8, size=N_SAMPLES)).astype(int)

    mit_anom_score = (
        0.40 * unjustified_delay +
        0.30 * (repeated_overdue >= 2).astype(int) +
        0.20 * (overdue_days > 90).astype(int) +
        0.15 * (progress_pct < 10.0).astype(int) +
        np.random.normal(0, 0.08, size=N_SAMPLES)
    )
    is_mit_anom = (mit_anom_score > 0.42).astype(int)
    mit_target_anom = np.where(is_mit_anom == 1, 'Ya (Anomali)', 'Tidak')

    mit_impact = np.clip(np.ceil((overdue_days / 40.0) + repeated_overdue * 0.5) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
    mit_likelihood = np.clip(np.ceil(repeated_overdue * 1.2 + unjustified_delay * 2.0 + (100 - progress_pct) / 30.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

    df_mitigation = pd.DataFrame({
        'ID Mitigasi': [f'MIT-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
        'Kategori Temuan': np.random.choice(finding_cats, N_SAMPLES),
        'Tingkat Keparahan Temuan': np.random.choice(severities, N_SAMPLES),
        'Deskripsi Rencana Aksi': [f'Rencana Aksi Mitigasi Temuan #{i+1}' for i in range(N_SAMPLES)],
        'overdue_days': overdue_days,
        'repeated_overdue_count': repeated_overdue,
        'mitigation_progress_pct': progress_pct,
        'is_unjustified_delay (1=Ya, 0=Tidak)': unjustified_delay,
        'TARGET: is_anomaly': mit_target_anom,
        'TARGET: Impact (1-5)': mit_impact,
        'TARGET: Likelihood (1-5)': mit_likelihood
    })
    df_mitigation.to_csv(os.path.join(CURRENT_DIR, 'mitigation_overdue_data.csv'), index=False)

if __name__ == "__main__":
    generate()
