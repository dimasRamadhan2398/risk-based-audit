import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(105)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1000

    entities = ['IT Dept', 'Operations Dept', 'Finance Dept']
    systems = ['Core Banking', 'Database Server', 'Email Server']

    data = {
        'ID Log': [f'IT-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas': np.random.choice(entities, N_SAMPLES),
        'Sistem': np.random.choice(systems, N_SAMPLES),
        'Jam Akses (0-23)': np.random.randint(0, 24, size=N_SAMPLES),
        'Jumlah Gagal Login': np.zeros(N_SAMPLES, dtype=int),
        'Tipe Akses': np.random.choice(['Read', 'Write', 'Admin'], size=N_SAMPLES, p=[0.7, 0.25, 0.05]),
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

        gagal_login = np.random.choice([0, 1, 2], p=[0.8, 0.15, 0.05])
        jam = data['Jam Akses (0-23)'][i]
        
        if np.random.rand() < 0.12:
            if np.random.rand() < 0.6:
                gagal_login = np.random.randint(5, 20)
                alasan.append(f"Indikasi Brute Force (Gagal Login {gagal_login}x)")
            else:
                if jam < 6 or jam > 22:
                    alasan.append(f"Akses sistem di luar jam kerja (Jam {jam})")
                if data['Tipe Akses'][i] == 'Admin':
                    alasan.append("Penggunaan Akses Admin mencurigakan")
                    impact = 5
            is_anom = True
            impact = max(impact, np.random.randint(3, 6))
            likelihood = np.random.randint(4, 6)
            
        data['Jumlah Gagal Login'][i] = gagal_login

        if is_anom:
            data['TARGET: is_anomaly'][i] = 'Ya (Anomali)'
            data['TARGET: Impact (1-5)'][i] = impact
            data['TARGET: Likelihood (1-5)'][i] = likelihood
            data['Alasan Anomali'][i] = " & ".join(alasan)
        else:
            data['TARGET: Impact (1-5)'][i] = impact
            data['TARGET: Likelihood (1-5)'][i] = likelihood

    df = pd.DataFrame(data)
    output_path = os.path.join(CURRENT_DIR, 'it_control_data.csv')
    df.to_csv(output_path, index=False)
    print(f"Generated {N_SAMPLES} records for IT Control: {output_path}")

if __name__ == "__main__":
    generate()
