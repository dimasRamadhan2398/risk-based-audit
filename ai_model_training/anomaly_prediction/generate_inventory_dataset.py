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
    # 6. INVENTORY_DATA.CSV (Inventory Adjustments)
    # ======================================================================
    print("Generating inventory_data.csv...")
    item_cats = ['Suku Cadang Utama', 'Bahan Baku Industri', 'Barang Jadi', 'Peralatan TI', 'Komoditas Umum']
    adj_amount = np.round(np.random.exponential(scale=30, size=N_SAMPLES) + 1.0, 2)
    stock_var_pct = np.round(np.random.uniform(0.5, 35.0, size=N_SAMPLES), 1)
    adj_freq = np.random.randint(1, 15, size=N_SAMPLES)
    has_doc = np.random.binomial(1, 0.75, size=N_SAMPLES)
    role_level = np.random.randint(1, 5, size=N_SAMPLES)
    manual_override = np.random.binomial(1, 0.20, size=N_SAMPLES)

    inv_anom_score = (
        0.35 * manual_override +
        0.30 * (1 - has_doc) +
        0.25 * (stock_var_pct > 20.0).astype(int) +
        0.20 * (adj_freq > 8).astype(int) +
        np.random.normal(0, 0.08, size=N_SAMPLES)
    )
    is_inv_anom = (inv_anom_score > 0.45).astype(int)
    inv_target_anom = np.where(is_inv_anom == 1, 'Ya (Anomali)', 'Tidak')

    inv_impact = np.clip(np.ceil(adj_amount / 15.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
    inv_likelihood = np.clip(np.ceil((1 - has_doc) * 2.0 + manual_override * 1.5 + (adj_freq / 3.0)) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

    df_inventory = pd.DataFrame({
        'ID Penyesuaian': [f'ADJ-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas Gudang': np.random.choice(ENTITIES, N_SAMPLES),
        'Kategori Barang': np.random.choice(item_cats, N_SAMPLES),
        'Deskripsi Penyesuaian': [f'Penyesuaian Stok Gudang #{i+1}' for i in range(N_SAMPLES)],
        'adjustment_amount (Juta Rp)': adj_amount,
        'stock_variance_pct': stock_var_pct,
        'adjustment_frequency_month': adj_freq,
        'has_supporting_doc (1=Ya, 0=Tidak)': has_doc,
        'approver_role_level (1-4)': role_level,
        'is_manual_override (1=Ya, 0=Tidak)': manual_override,
        'TARGET: is_anomaly': inv_target_anom,
        'TARGET: Impact (1-5)': inv_impact,
        'TARGET: Likelihood (1-5)': inv_likelihood
    })
    df_inventory.to_csv(os.path.join(CURRENT_DIR, 'inventory_data.csv'), index=False)

if __name__ == "__main__":
    generate()
