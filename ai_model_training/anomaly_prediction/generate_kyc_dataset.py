import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(104)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1000

    entities = ['Customer Service JKT', 'Customer Service SBY', 'Onboarding Dept']
    products = ['Pembukaan Rekening', 'Update Data Nasabah']

    data = {
        'ID Aktivitas': [f'KYC-{i+1:04d}' for i in range(N_SAMPLES)],
        'Entitas': np.random.choice(entities, N_SAMPLES),
        'Aktivitas': np.random.choice(products, N_SAMPLES),
        'Profil Risiko': np.random.choice(['Low', 'Medium', 'High'], size=N_SAMPLES, p=[0.6, 0.3, 0.1]),
        'Status Dokumen': np.random.choice(['Lengkap', 'Tidak Lengkap'], size=N_SAMPLES, p=[0.8, 0.2]),
        'Status Approval': np.random.choice(['Approved', 'Rejected', 'Pending'], size=N_SAMPLES, p=[0.7, 0.1, 0.2]),
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
        
        if data['Status Dokumen'][i] == 'Tidak Lengkap' and data['Status Approval'][i] == 'Approved':
            alasan.append("Aktivitas disetujui padahal Dokumen Tidak Lengkap")
            is_anom = True
            impact = np.random.randint(4, 6)
            likelihood = np.random.randint(4, 6)
            
        if data['Profil Risiko'][i] == 'High' and data['Status Approval'][i] == 'Approved' and np.random.rand() < 0.2:
            alasan.append("Nasabah High Risk disetujui tanpa eskalasi")
            is_anom = True
            impact = 5
            likelihood = 3

        if is_anom:
            data['TARGET: is_anomaly'][i] = 'Ya (Anomali)'
            data['TARGET: Impact (1-5)'][i] = impact
            data['TARGET: Likelihood (1-5)'][i] = likelihood
            data['Alasan Anomali'][i] = " & ".join(alasan)
        else:
            data['TARGET: Impact (1-5)'][i] = impact
            data['TARGET: Likelihood (1-5)'][i] = likelihood

    df = pd.DataFrame(data)
    output_path = os.path.join(CURRENT_DIR, 'kyc_data.csv')
    df.to_csv(output_path, index=False)
    print(f"Generated {N_SAMPLES} records for KYC: {output_path}")

if __name__ == "__main__":
    generate()
