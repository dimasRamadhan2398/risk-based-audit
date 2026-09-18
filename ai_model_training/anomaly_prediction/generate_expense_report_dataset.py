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
    # 4. EXPENSE_REPORT_DATA.CSV (Expense Claims & Reimbursement)
    # ======================================================================
    print("Generating expense_report_data.csv...")
    expense_cats = ['Travel & Entertainment', 'Kertas & Alat Tulis', 'Konsumsi Rapat', 'Sewa Kendaraan Operasional', 'Vendor Pihak Ketiga']
    claim_amount = np.round(np.random.exponential(scale=15, size=N_SAMPLES) + 0.5, 2)
    seq_invoice = np.random.binomial(1, 0.12, size=N_SAMPLES)
    exact_round = np.random.binomial(1, 0.25, size=N_SAMPLES)
    dup_claims = np.random.binomial(3, 0.1, size=N_SAMPLES)
    days_since_last = np.random.randint(0, 45, size=N_SAMPLES)
    over_limit = ((claim_amount > 25.0) & np.random.binomial(1, 0.8, size=N_SAMPLES)).astype(int)

    exp_anom_score = (
        0.35 * seq_invoice +
        0.30 * (dup_claims >= 1).astype(int) +
        0.25 * over_limit +
        0.15 * ((days_since_last <= 1) & (claim_amount > 10)).astype(int) +
        np.random.normal(0, 0.08, size=N_SAMPLES)
    )
    is_exp_anom = (exp_anom_score > 0.38).astype(int)
    exp_target_anom = np.where(is_exp_anom == 1, 'Ya (Anomali)', 'Tidak')

    exp_impact = np.clip(np.ceil(claim_amount / 8.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
    exp_likelihood = np.clip(np.ceil(dup_claims * 1.2 + seq_invoice * 1.8 + over_limit * 1.5 + (days_since_last <= 2) * 1.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

    df_expense = pd.DataFrame({
        'ID Klaim': [f'CLM-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas': np.random.choice(ENTITIES, N_SAMPLES),
        'Kategori Pengeluaran': np.random.choice(expense_cats, N_SAMPLES),
        'Deskripsi Pengeluaran': [f'Klaim Pengeluaran Operasional #{i+1}' for i in range(N_SAMPLES)],
        'claim_amount (Juta Rp)': claim_amount,
        'is_sequential_invoice (1=Ya, 0=Tidak)': seq_invoice,
        'is_exact_round_amount (1=Ya, 0=Tidak)': exact_round,
        'duplicate_claim_count': dup_claims,
        'days_since_last_claim': days_since_last,
        'is_over_limit_threshold (1=Ya, 0=Tidak)': over_limit,
        'TARGET: is_anomaly': exp_target_anom,
        'TARGET: Impact (1-5)': exp_impact,
        'TARGET: Likelihood (1-5)': exp_likelihood
    })
    df_expense.to_csv(os.path.join(CURRENT_DIR, 'expense_report_data.csv'), index=False)

if __name__ == "__main__":
    generate()
