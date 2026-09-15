import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(101)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1000

    entities = ['Jakarta Branch', 'Surabaya Branch', 'Bandung Branch', 'Medan Branch']
    products = ['Kredit Multiguna', 'Kredit Modal Kerja', 'Kredit KPR']

    data = {
        'ID Transaksi': [f'LEND-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas': np.random.choice(entities, N_SAMPLES),
        'Produk': np.random.choice(products, N_SAMPLES),
        'Nilai Plafon (Juta Rp)': np.zeros(N_SAMPLES),
        'Bunga (%)': np.zeros(N_SAMPLES),
        'Status Dokumen': np.random.choice(['Lengkap', 'Tidak Lengkap'], size=N_SAMPLES, p=[0.85, 0.15]),
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

        bunga = round(np.random.uniform(12.0, 17.0), 2)
        amount = round(np.random.exponential(scale=1000) + 100, 2)
        
        if np.random.rand() < 0.15:
            if np.random.rand() < 0.5:
                bunga = round(np.random.uniform(5.0, 11.9), 2)
                alasan.append(f"Bunga Kredit terlalu rendah ({bunga}%)")
            else:
                bunga = round(np.random.uniform(17.1, 25.0), 2)
                alasan.append(f"Bunga Kredit terlalu tinggi ({bunga}%)")
            
            if data['Status Dokumen'][i] == 'Tidak Lengkap':
                alasan.append("Pencairan padahal Dokumen Tidak Lengkap")
                impact = 5
            is_anom = True
            impact = max(impact, np.random.randint(4, 6))
            likelihood = np.random.randint(3, 5)

        data['Bunga (%)'][i] = bunga
        data['Nilai Plafon (Juta Rp)'][i] = amount

        if is_anom:
            data['TARGET: is_anomaly'][i] = 'Ya (Anomali)'
            data['TARGET: Impact (1-5)'][i] = impact
            data['TARGET: Likelihood (1-5)'][i] = likelihood
            data['Alasan Anomali'][i] = " & ".join(alasan)
        else:
            data['TARGET: Impact (1-5)'][i] = impact
            data['TARGET: Likelihood (1-5)'][i] = likelihood

    df = pd.DataFrame(data)
    output_path = os.path.join(CURRENT_DIR, 'lending_data.csv')
    df.to_csv(output_path, index=False)
    print(f"Generated {N_SAMPLES} records for Lending: {output_path}")

if __name__ == "__main__":
    generate()
