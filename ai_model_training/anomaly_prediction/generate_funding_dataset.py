import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(42)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1000

    entities = ['Jakarta Branch', 'Surabaya Branch', 'Bandung Branch', 'Medan Branch', 'Head Office']
    products = ['Deposito', 'Tabungan', 'Giro']

    data = {
        'ID Transaksi': [f'FUND-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas': np.random.choice(entities, N_SAMPLES),
        'Produk': np.random.choice(products, N_SAMPLES),
        'Nilai Transaksi (Juta Rp)': np.zeros(N_SAMPLES),
        'Bunga (%)': np.zeros(N_SAMPLES),
        'TARGET: is_anomaly': ['Tidak'] * N_SAMPLES,
        'TARGET: Impact (1-5)': np.ones(N_SAMPLES, dtype=int),
        'TARGET: Likelihood (1-5)': np.ones(N_SAMPLES, dtype=int),
        'Alasan Anomali': ['-'] * N_SAMPLES
    }

    for i in range(N_SAMPLES):
        is_anom = False
        alasan = []
        impact = np.random.randint(1, 3)
        likelihood = np.random.randint(1, 3)

        # Funding: Normal bunga 2% - 6%
        bunga = round(np.random.uniform(2.0, 6.0), 2)
        amount = round(np.random.exponential(scale=500) + 10, 2)
        
        if np.random.rand() < 0.1:  # 10% chance of anomaly
            if np.random.rand() < 0.5:
                bunga = round(np.random.uniform(7.0, 12.0), 2)
                alasan.append(f"Bunga Funding tidak wajar ({bunga}%)")
            else:
                amount = round(np.random.uniform(10000, 50000), 2)
                alasan.append("Penempatan/Penarikan sangat besar")
            is_anom = True
            impact = np.random.randint(3, 6)
            likelihood = np.random.randint(3, 6)
        
        data['Bunga (%)'][i] = bunga
        data['Nilai Transaksi (Juta Rp)'][i] = amount

        if is_anom:
            data['TARGET: is_anomaly'][i] = 'Ya (Anomali)'
            data['TARGET: Impact (1-5)'][i] = impact
            data['TARGET: Likelihood (1-5)'][i] = likelihood
            data['Alasan Anomali'][i] = " & ".join(alasan)
        else:
            data['TARGET: Impact (1-5)'][i] = impact
            data['TARGET: Likelihood (1-5)'][i] = likelihood

    df = pd.DataFrame(data)
    output_path = os.path.join(CURRENT_DIR, 'funding_data.csv')
    df.to_csv(output_path, index=False)
    print(f"Generated {N_SAMPLES} records for Funding: {output_path}")

if __name__ == "__main__":
    generate()
