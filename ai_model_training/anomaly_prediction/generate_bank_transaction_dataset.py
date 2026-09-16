import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(42)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1000

    categories = ['Funding', 'Lending', 'Treasury', 'Payment', 'KYC', 'IT Control']
    entities = ['Jakarta Branch', 'Surabaya Branch', 'Bandung Branch', 'Medan Branch', 'Head Office']

    # Initialize lists to store data
    data = {
        'ID Transaksi': [f'BANK-TRX-{i+1:04d}' for i in range(N_SAMPLES)],
        'Kategori': np.random.choice(categories, N_SAMPLES),
        'Entitas': np.random.choice(entities, N_SAMPLES),
        'Nilai Transaksi (Juta Rp)': np.zeros(N_SAMPLES),
        'Bunga/Margin (%)': np.zeros(N_SAMPLES),
        'Jam Transaksi (0-23)': np.random.randint(0, 24, size=N_SAMPLES),
        'Status Dokumen': np.random.choice(['Lengkap', 'Tidak Lengkap'], size=N_SAMPLES, p=[0.9, 0.1]),
        'Jumlah Gagal Login': np.zeros(N_SAMPLES, dtype=int),
        'TARGET: is_anomaly': ['Tidak'] * N_SAMPLES,
        'TARGET: Impact (1-5)': np.ones(N_SAMPLES, dtype=int),
        'TARGET: Likelihood (1-5)': np.ones(N_SAMPLES, dtype=int),
        'Alasan Anomali': ['-'] * N_SAMPLES
    }

    # Generate specific features and determine anomalies based on category rules
    for i in range(N_SAMPLES):
        cat = data['Kategori'][i]
        is_anom = False
        alasan = []
        
        # Default Impact and Likelihood
        impact = np.random.randint(1, 3)
        likelihood = np.random.randint(1, 3)

        if cat == 'Funding':
            # Funding: Normal bunga 2% - 6%
            bunga = round(np.random.uniform(2.0, 6.0), 2)
            amount = round(np.random.exponential(scale=500), 2)
            
            # Inject anomalies
            if np.random.rand() < 0.1:  # 10% chance of anomaly
                if np.random.rand() < 0.5:
                    bunga = round(np.random.uniform(7.0, 12.0), 2) # Bunga deposito terlalu tinggi
                    alasan.append(f"Bunga Funding tidak wajar ({bunga}%)")
                else:
                    amount = round(np.random.uniform(10000, 50000), 2) # Nominal sangat besar
                    alasan.append("Penempatan Funding/Penarikan sangat besar")
                is_anom = True
                impact = np.random.randint(3, 6)
                likelihood = np.random.randint(3, 6)
            
            data['Bunga/Margin (%)'][i] = bunga
            data['Nilai Transaksi (Juta Rp)'][i] = amount

        elif cat == 'Lending':
            # Lending: Asumsi bunga normal 12% - 17% (sesuai request)
            bunga = round(np.random.uniform(12.0, 17.0), 2)
            amount = round(np.random.exponential(scale=1000) + 100, 2)
            
            # Inject anomalies
            if np.random.rand() < 0.12:
                # Bunga di luar range 12-17%
                if np.random.rand() < 0.5:
                    bunga = round(np.random.uniform(5.0, 11.9), 2) # Bunga terlalu rendah
                    alasan.append(f"Bunga Kredit terlalu rendah ({bunga}%)")
                else:
                    bunga = round(np.random.uniform(17.1, 25.0), 2) # Bunga terlalu tinggi
                    alasan.append(f"Bunga Kredit terlalu tinggi ({bunga}%)")
                
                if data['Status Dokumen'][i] == 'Tidak Lengkap':
                    alasan.append("Dokumen Kredit Tidak Lengkap")
                    impact = 5
                is_anom = True
                impact = max(impact, np.random.randint(4, 6))
                likelihood = np.random.randint(3, 5)

            data['Bunga/Margin (%)'][i] = bunga
            data['Nilai Transaksi (Juta Rp)'][i] = amount

        elif cat == 'Treasury':
            # Treasury: Transaksi FX/Bonds. Normal di jam kerja (8-17)
            amount = round(np.random.exponential(scale=2000) + 500, 2)
            data['Bunga/Margin (%)'][i] = 0.0 # N/A
            
            # Inject anomalies
            if np.random.rand() < 0.1:
                jam = data['Jam Transaksi (0-23)'][i]
                if jam < 7 or jam > 18:
                    amount = round(np.random.uniform(5000, 20000), 2)
                    alasan.append(f"Transaksi Treasury besar di luar jam kerja (Jam {jam})")
                    is_anom = True
                    impact = np.random.randint(4, 6)
                    likelihood = np.random.randint(2, 5)
                else:
                    # just high amount
                    amount = round(np.random.uniform(15000, 30000), 2)
                    alasan.append("Nilai Transaksi Treasury melebihi limit wajar")
                    is_anom = True
                    impact = 5
                    likelihood = 3

            data['Nilai Transaksi (Juta Rp)'][i] = amount

        elif cat == 'Payment':
            # Payment: Transfer nasabah, RTGS.
            amount = round(np.random.exponential(scale=50) + 1, 2)
            data['Bunga/Margin (%)'][i] = 0.0
            
            if np.random.rand() < 0.1:
                jam = data['Jam Transaksi (0-23)'][i]
                if jam >= 23 or jam <= 4:
                    amount = round(np.random.uniform(500, 5000), 2)
                    alasan.append(f"Transfer Payment jumlah besar di dini hari (Jam {jam})")
                    is_anom = True
                    impact = np.random.randint(3, 5)
                    likelihood = np.random.randint(3, 6)

            data['Nilai Transaksi (Juta Rp)'][i] = amount

        elif cat == 'KYC':
            # KYC: Customer onboarding / update data
            data['Nilai Transaksi (Juta Rp)'][i] = 0.0
            data['Bunga/Margin (%)'][i] = 0.0
            
            if np.random.rand() < 0.15:
                data['Status Dokumen'][i] = 'Tidak Lengkap'
                alasan.append("Proses KYC disetujui padahal Dokumen Tidak Lengkap")
                is_anom = True
                impact = np.random.randint(4, 6)
                likelihood = np.random.randint(4, 6)

        elif cat == 'IT Control':
            # IT Control: Akses sistem, user login
            data['Nilai Transaksi (Juta Rp)'][i] = 0.0
            data['Bunga/Margin (%)'][i] = 0.0
            gagal_login = np.random.choice([0, 1, 2], p=[0.8, 0.15, 0.05])
            
            if np.random.rand() < 0.1:
                gagal_login = np.random.randint(4, 15)
                alasan.append(f"Indikasi Brute Force / Gagal Login Tinggi ({gagal_login} kali)")
                is_anom = True
                impact = np.random.randint(4, 6)
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
    
    # Save dataset
    output_path = os.path.join(CURRENT_DIR, 'bank_transaction_anomaly_data.csv')
    df.to_csv(output_path, index=False)
    print(f"Dataset generated with 1000 records: {output_path}")

if __name__ == "__main__":
    generate()
