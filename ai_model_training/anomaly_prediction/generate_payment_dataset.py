import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(103)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1000

    entities = ['Jakarta Branch', 'Surabaya Branch', 'Bandung Branch', 'Medan Branch']
    products = ['RTGS', 'SKN', 'Transfer Internal', 'BI-FAST']

    data = {
        'ID Transaksi': [f'PAYM-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas': np.random.choice(entities, N_SAMPLES),
        'Produk': np.random.choice(products, N_SAMPLES),
        'Nilai Transaksi (Juta Rp)': np.zeros(N_SAMPLES),
        'Jam Transaksi (0-23)': np.random.randint(0, 24, size=N_SAMPLES),
        'Penerima Baru': np.random.choice(['Ya', 'Tidak'], size=N_SAMPLES, p=[0.2, 0.8]),
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

        amount = round(np.random.exponential(scale=50) + 1, 2)
        
        if np.random.rand() < 0.1:
            jam = data['Jam Transaksi (0-23)'][i]
            if jam >= 23 or jam <= 4:
                amount = round(np.random.uniform(500, 5000), 2)
                alasan.append(f"Transfer jumlah besar di dini hari (Jam {jam})")
                if data['Penerima Baru'][i] == 'Ya':
                    alasan.append("Penerima Dana Baru (Mencurigakan)")
                    impact = 5
                is_anom = True
                impact = max(impact, np.random.randint(3, 6))
                likelihood = np.random.randint(3, 6)

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
    output_path = os.path.join(CURRENT_DIR, 'payment_data.csv')
    df.to_csv(output_path, index=False)
    print(f"Generated {N_SAMPLES} records for Payment: {output_path}")

if __name__ == "__main__":
    generate()
