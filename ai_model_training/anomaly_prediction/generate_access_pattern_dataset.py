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
    # 2. ACCESS_PATTERN_DATA.CSV (System & Access Security)
    # ======================================================================
    print("Generating access_pattern_data.csv...")
    roles = ['Audit Staff', 'Senior Auditor', 'Audit Manager', 'System Admin', 'External Guest']
    failed_logins = np.random.negative_binomial(1, 0.3, size=N_SAMPLES)
    export_mb = np.round(np.random.exponential(scale=250, size=N_SAMPLES), 1)
    is_terminated = np.random.binomial(1, 0.08, size=N_SAMPLES)
    ip_risk = np.round(np.clip(np.random.beta(0.5, 2.0, size=N_SAMPLES), 0.0, 1.0), 2)

    access_anom_score = (
        0.40 * is_terminated +
        0.30 * (ip_risk > 0.6).astype(int) +
        0.20 * (failed_logins >= 4).astype(int) +
        0.20 * (export_mb > 500).astype(int) +
        np.random.normal(0, 0.08, size=N_SAMPLES)
    )
    is_acc_anom = (access_anom_score > 0.42).astype(int)
    acc_target_anom = np.where(is_acc_anom == 1, 'Ya (Anomali)', 'Tidak')

    acc_impact = np.clip(np.ceil((export_mb / 250) + is_terminated * 2) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
    acc_likelihood = np.clip(np.ceil(ip_risk * 4.0 + failed_logins * 0.4) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

    df_access = pd.DataFrame({
        'ID Akses': [f'ACC-{i+1:04d}' for i in range(N_SAMPLES)],
        'User ID': [f'USR-{np.random.randint(100, 999)}' for _ in range(N_SAMPLES)],
        'Entitas': np.random.choice(ENTITIES, N_SAMPLES),
        'Peran User': np.random.choice(roles, N_SAMPLES),
        'hour_of_day (0-23)': np.random.randint(0, 24, size=N_SAMPLES),
        'day_of_week (1-7)': np.random.randint(1, 8, size=N_SAMPLES),
        'failed_login_attempts': failed_logins,
        'data_export_volume_mb': export_mb,
        'is_terminated_user (1=Ya, 0=Tidak)': is_terminated,
        'ip_risk_score (0.0-1.0)': ip_risk,
        'TARGET: is_anomaly': acc_target_anom,
        'TARGET: Impact (1-5)': acc_impact,
        'TARGET: Likelihood (1-5)': acc_likelihood
    })
    df_access.to_csv(os.path.join(CURRENT_DIR, 'access_pattern_data.csv'), index=False)

if __name__ == "__main__":
    generate()
