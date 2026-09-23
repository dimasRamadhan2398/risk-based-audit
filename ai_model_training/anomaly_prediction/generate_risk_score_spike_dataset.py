import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(42)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1200

    # Common entities
    ENTITIES = ['Jakarta Branch', 'Surabaya Branch', 'Bandung Branch', 'Medan Branch', 'Finance Dept', 'IT Dept', 'Operations Dept']
    risk_cats = ['Operasional', 'Finansial', 'Legal & Kepatuhan', 'Reputasi', 'Keamanan TI']

    # ======================================================================
    # 9. RISK_SCORE_SPIKE_DATA.CSV (Sudden Volatility in Risk Score)
    # ======================================================================
    print("Generating risk_score_spike_data.csv...")
    prev_score = np.round(np.random.uniform(3.0, 18.0, size=N_SAMPLES), 1)
    diff = np.round(np.random.normal(1.5, 4.0, size=N_SAMPLES), 1)
    curr_score = np.round(np.clip(prev_score + diff, 1.0, 25.0), 1)
    new_findings = np.random.negative_binomial(1, 0.3, size=N_SAMPLES)
    volatility = np.round(np.random.exponential(scale=0.15, size=N_SAMPLES), 2)
    unexplained_spike = ((diff > 6.0) & (new_findings <= 1) & np.random.binomial(1, 0.8, size=N_SAMPLES)).astype(int)

    spike_anom_score = (
        0.45 * unexplained_spike +
        0.30 * (diff > 8.0).astype(int) +
        0.20 * (volatility > 0.35).astype(int) +
        np.random.normal(0, 0.08, size=N_SAMPLES)
    )
    is_spike_anom = (spike_anom_score > 0.42).astype(int)
    spike_target_anom = np.where(is_spike_anom == 1, 'Ya (Anomali)', 'Tidak')

    spike_impact = np.clip(np.ceil(curr_score / 5.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
    spike_likelihood = np.clip(np.ceil(unexplained_spike * 2.5 + volatility * 8.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

    df_spike = pd.DataFrame({
        'ID Penilaian Risiko': [f'RSK-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
        'Kategori Risiko': np.random.choice(risk_cats, N_SAMPLES),
        'Deskripsi Evaluasi Risiko': [f'Evaluasi Lonjakan Skor Risiko #{i+1}' for i in range(N_SAMPLES)],
        'current_risk_score': curr_score,
        'previous_risk_score': prev_score,
        'risk_score_diff': diff,
        'new_findings_count': new_findings,
        'risk_score_volatility': volatility,
        'is_unexplained_spike (1=Ya, 0=Tidak)': unexplained_spike,
        'TARGET: is_anomaly': spike_target_anom,
        'TARGET: Impact (1-5)': spike_impact,
        'TARGET: Likelihood (1-5)': spike_likelihood
    })
    df_spike.to_csv(os.path.join(CURRENT_DIR, 'risk_score_spike_data.csv'), index=False)

if __name__ == "__main__":
    generate()
