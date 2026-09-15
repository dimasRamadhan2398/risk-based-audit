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
    # 1. ANOMALY_DATA.CSV (Transaction Anomalies)
    # ======================================================================
    print("Generating realistic anomaly_data.csv...")
    descriptions = [
        'Pembayaran Vendor', 'Reimbursement Operasional', 'Transfer Kas Internal', 
        'Pengadaan Perangkat IT', 'Biaya Perjalanan Dinas', 'Jasa Konsultan Audit', 
        'Pembayaran Tagihan Listrik', 'Biaya Jamuan Klien', 'Pembelian Material Gudang', 
        'Pembayaran Gaji Lembur'
    ]

    amounts = np.round(np.random.exponential(scale=50, size=N_SAMPLES) + 1.0, 2)
    hours = np.random.randint(0, 24, size=N_SAMPLES)
    days = np.random.randint(1, 8, size=N_SAMPLES)
    new_ben = np.random.binomial(1, 0.25, size=N_SAMPLES)
    round_amt = np.random.binomial(1, 0.35, size=N_SAMPLES)
    selected_entities = np.random.choice(ENTITIES, N_SAMPLES)
    selected_descs = np.random.choice(descriptions, N_SAMPLES)

    is_night = ((hours >= 22) | (hours <= 5)).astype(int)
    high_amount = (amounts > 120).astype(int)

    anom_score = (
        0.35 * high_amount +
        0.30 * is_night +
        0.25 * new_ben +
        0.20 * round_amt +
        0.15 * (pd.Series(selected_entities).isin(['Finance Dept', 'Operations Dept']).values & is_night) +
        np.random.normal(0, 0.1, size=N_SAMPLES)
    )

    is_anom = (anom_score > 0.48).astype(int)
    target_anom = np.where(is_anom == 1, 'Ya (Anomali)', 'Tidak')

    # Impact (1-5): Driven by amount magnitude
    impact = np.clip(np.ceil(np.log1p(amounts) / 1.1) + np.random.choice([-1, 0, 1], size=N_SAMPLES, p=[0.15, 0.7, 0.15]), 1, 5).astype(int)

    # Likelihood (1-5): Driven by anomaly severity & violation count
    likelihood = np.clip(np.ceil(anom_score * 5.0) + np.random.choice([-1, 0, 1], size=N_SAMPLES, p=[0.15, 0.7, 0.15]), 1, 5).astype(int)

    df_anom = pd.DataFrame({
        'ID Transaksi': [f'TRX-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas': selected_entities,
        'Deskripsi': selected_descs,
        'amount (dalam Juta Rp)': amounts,
        'hour_of_day (0-23)': hours,
        'day_of_week (1-7)': days,
        'is_new_beneficiary (1=Ya, 0=Tidak)': new_ben,
        'is_round_amount (1=Ya, 0=Tidak)': round_amt,
        'TARGET: is_anomaly': target_anom,
        'TARGET: Impact (1-5)': impact,
        'TARGET: Likelihood (1-5)': likelihood
    })
    df_anom.to_csv(os.path.join(CURRENT_DIR, 'anomaly_data.csv'), index=False)

if __name__ == "__main__":
    generate()
