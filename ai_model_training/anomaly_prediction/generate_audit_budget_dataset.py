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
    # 3. AUDIT_BUDGET_DATA.CSV (Audit Planning & Budgeting)
    # ======================================================================
    print("Generating audit_budget_data.csv...")
    categories = ['Audit Keuangan', 'Audit TI', 'Audit Operasional', 'Audit Kepatuhan', 'Special Investigation']
    hist_avg_budget = np.round(np.random.uniform(50, 400, size=N_SAMPLES), 1)
    budget_ratio = np.random.normal(1.0, 0.35, size=N_SAMPLES)
    allocated_budget = np.round(np.maximum(10.0, hist_avg_budget * np.maximum(0.3, budget_ratio)), 1)
    mandays = np.random.randint(10, 180, size=N_SAMPLES)
    auditors = np.random.randint(2, 12, size=N_SAMPLES)
    budget_zscore = np.round((allocated_budget - hist_avg_budget) / 50.0, 2)
    unapproved_spike = ((budget_zscore > 2.0) & (np.random.binomial(1, 0.7, size=N_SAMPLES) == 1)).astype(int)

    budget_anom_score = (
        0.45 * unapproved_spike +
        0.35 * (budget_zscore > 2.2).astype(int) +
        0.20 * (allocated_budget / (auditors * 5) > 15).astype(int) +
        np.random.normal(0, 0.1, size=N_SAMPLES)
    )
    is_budg_anom = (budget_anom_score > 0.45).astype(int)
    budg_target_anom = np.where(is_budg_anom == 1, 'Ya (Anomali)', 'Tidak')

    budg_impact = np.clip(np.ceil(allocated_budget / 90.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
    budg_likelihood = np.clip(np.ceil(np.maximum(0, budget_zscore) + unapproved_spike * 1.5) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

    df_budget = pd.DataFrame({
        'ID Rencana Audit': [f'PLN-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
        'Kategori Perencanaan': np.random.choice(categories, N_SAMPLES),
        'Deskripsi Perencanaan': [f'Evaluasi Perencanaan Risk Area {i+1}' for i in range(N_SAMPLES)],
        'allocated_budget_juta': allocated_budget,
        'historical_avg_budget_juta': hist_avg_budget,
        'allocated_mandays': mandays,
        'assigned_auditors_count': auditors,
        'budget_zscore': budget_zscore,
        'is_unapproved_budget_spike (1=Ya, 0=Tidak)': unapproved_spike,
        'TARGET: is_anomaly': budg_target_anom,
        'TARGET: Impact (1-5)': budg_impact,
        'TARGET: Likelihood (1-5)': budg_likelihood
    })
    df_budget.to_csv(os.path.join(CURRENT_DIR, 'audit_budget_data.csv'), index=False)

if __name__ == "__main__":
    generate()
