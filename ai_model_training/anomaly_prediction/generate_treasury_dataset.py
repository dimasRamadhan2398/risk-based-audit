import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(102)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1000

    entities = ['Head Office', 'Singapore Branch']
    products = ['FX Spot', 'FX Forward', 'Government Bonds']

    data = {
        'ID Transaksi': [f'TRSY-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas': np.random.choice(entities, N_SAMPLES),
        'Produk': np.random.choice(products, N_SAMPLES),
        'Nilai Transaksi (Juta Rp)': np.zeros(N_SAMPLES),
        'Jam Transaksi (0-23)': np.random.randint(0, 24, size=N_SAMPLES),
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

        amount = round(np.random.exponential(scale=2000) + 500, 2)
        
        if np.random.rand() < 0.12:
            jam = data['Jam Transaksi (0-23)'][i]
            if jam < 7 or jam > 18:
                amount = round(np.random.uniform(5000, 20000), 2)
                alasan.append(f"Transaksi besar di luar jam kerja (Jam {jam})")
                is_anom = True
                impact = np.random.randint(4, 6)
                likelihood = np.random.randint(3, 5)
            else:
                amount = round(np.random.uniform(20000, 50000), 2)
                alasan.append("Nilai Transaksi Treasury melebihi limit wajar")
                is_anom = True
                impact = 5
                likelihood = 3

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
    output_path = os.path.join(CURRENT_DIR, 'treasury_data.csv')
    df.to_csv(output_path, index=False)
    print(f"Generated {N_SAMPLES} records for Treasury: {output_path}")

if __name__ == "__main__":
    generate()
