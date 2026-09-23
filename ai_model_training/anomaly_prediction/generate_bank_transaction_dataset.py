import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(42)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    
    categories = ['Funding', 'Lending', 'Treasury', 'Payment', 'KYC', 'IT Control']
    entities = ['Kantor Pusat', 'Cabang Jakarta', 'Cabang Surabaya', 'Cabang Bandung', 'Cabang Medan', 'Cabang Bali']
    
    # 200 records per category = 1200 total records
    N_PER_CAT = 200
    N_ANOM_PER_CAT = 25
    N_NORM_PER_CAT = N_PER_CAT - N_ANOM_PER_CAT  # 175 normal

    records = []

    for cat in categories:
        # ---------------------------------------------------------------------
        # 1. NORMAL RECORDS FOR CATEGORY (175 records)
        # ---------------------------------------------------------------------
        for _ in range(N_NORM_PER_CAT):
            entity = np.random.choice(entities)
            doc_status = 'Lengkap'  # 100% clean ground truth for normal
            failed_login = 0
            risk_profile = np.random.choice(['Low Risk', 'Medium Risk', 'High Risk'], p=[0.70, 0.25, 0.05])
            auth_status = np.random.choice([
                'Terverifikasi Dual Control (Maker-Checker)',
                'Otorisasi Mandiri Sesuai Limit'
            ], p=[0.65, 0.35])
            hist_dev = round(np.random.uniform(0.0, 25.0), 2)
            impact = np.random.choice([1, 2], p=[0.75, 0.25])
            likelihood = np.random.choice([1, 2], p=[0.70, 0.30])

            if cat == 'Funding':
                channel = np.random.choice(['Cabang / Teller', 'Mobile / Internet Banking', 'ATM / CRM'])
                amount = round(np.random.exponential(scale=350) + 15.0, 2)
                interest = round(np.random.uniform(1.75, 4.25), 2)  # Dalam batas wajar LPS <= 4.25%
                hour = np.random.randint(7, 21)

            elif cat == 'Lending':
                channel = 'Cabang / Teller'
                amount = round(np.random.exponential(scale=1200) + 150.0, 2)
                interest = round(np.random.uniform(8.0, 14.5), 2)  # Bunga kredit komersial wajar
                hour = np.random.randint(8, 17)

            elif cat == 'Treasury':
                channel = np.random.choice(['Dealing Room / SWIFT', 'Host-to-Host Core Banking'])
                amount = round(np.random.exponential(scale=2500) + 800.0, 2)
                interest = 0.0
                hour = np.random.randint(8, 17)  # Jam pasar resmi (08:00 - 16:30)

            elif cat == 'Payment':
                channel = np.random.choice(['Mobile / Internet Banking', 'ATM / CRM', 'Cabang / Teller'])
                amount = round(np.random.exponential(scale=35) + 1.0, 2)
                interest = 0.0
                hour = np.random.randint(6, 23)

            elif cat == 'KYC':
                channel = np.random.choice(['Cabang / Teller', 'Mobile / Internet Banking'])
                amount = 0.0
                interest = 0.0
                hour = np.random.randint(8, 17)

            else:  # IT Control
                channel = np.random.choice(['Host-to-Host Core Banking', 'Mobile / Internet Banking'])
                amount = 0.0
                interest = 0.0
                hour = np.random.randint(6, 22)
                failed_login = np.random.choice([0, 1, 2], p=[0.85, 0.12, 0.03])

            records.append({
                'Kategori': cat,
                'Entitas': entity,
                'Kanal Transaksi': channel,
                'Nilai Transaksi (Juta Rp)': amount,
                'Bunga/Margin (%)': interest,
                'Jam Transaksi (0-23)': hour,
                'Status Dokumen': doc_status,
                'Jumlah Gagal Login': failed_login,
                'Tingkat Risiko Nasabah': risk_profile,
                'Status Otorisasi / Maker-Checker': auth_status,
                'Deviasi terhadap Profil Historis (%)': hist_dev,
                'TARGET: is_anomaly': 'Tidak',
                'TARGET: Impact (1-5)': impact,
                'TARGET: Likelihood (1-5)': likelihood,
                'Alasan Anomali': '-'
            })

        # ---------------------------------------------------------------------
        # 2. ANOMALY RECORDS FOR CATEGORY (25 records)
        # ---------------------------------------------------------------------
        for a_idx in range(N_ANOM_PER_CAT):
            entity = np.random.choice(entities)
            risk_profile = np.random.choice(['High Risk', 'PEP / High Risk Watchlist'], p=[0.60, 0.40])
            auth_status = 'Override / Tanpa Otorisasi'
            hist_dev = round(np.random.uniform(90.0, 350.0), 2)
            impact = np.random.choice([4, 5], p=[0.45, 0.55])
            likelihood = np.random.choice([4, 5], p=[0.40, 0.60])
            failed_login = 0

            if cat == 'Funding':
                channel = 'Cabang / Teller'
                hour = np.random.randint(8, 17)
                doc_status = 'Lengkap'
                sub_type = a_idx % 2
                if sub_type == 0:
                    amount = round(np.random.uniform(2000.0, 15000.0), 2)
                    interest = round(np.random.uniform(7.5, 11.5), 2)  # Bunga simpanan di atas batas LPS tanpa izin ALCO
                    reason = f"Bunga simpanan deposito ({interest}%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO"
                else:
                    amount = round(np.random.uniform(15000.0, 45000.0), 2)
                    interest = round(np.random.uniform(1.5, 3.0), 2)
                    reason = f"Penarikan dana masif (Rp {amount} Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager"

            elif cat == 'Lending':
                channel = 'Cabang / Teller'
                hour = np.random.randint(8, 17)
                sub_type = a_idx % 2
                if sub_type == 0:
                    amount = round(np.random.uniform(2500.0, 25000.0), 2)
                    interest = round(np.random.uniform(8.5, 12.0), 2)
                    doc_status = 'Tidak Lengkap'
                    reason = "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan"
                else:
                    amount = round(np.random.uniform(35000.0, 75000.0), 2)
                    interest = round(np.random.uniform(3.5, 5.0), 2)  # Bunga kredit murah ekstrem pihak terkait
                    doc_status = 'Lengkap'
                    reason = f"Pencairan kredit bernilai Rp {amount} Juta dengan bunga murah ekstrem ({interest}%) pihak terafiliasi melanggar BMPK"

            elif cat == 'Treasury':
                channel = 'Dealing Room / SWIFT'
                interest = 0.0
                doc_status = 'Lengkap'
                sub_type = a_idx % 2
                if sub_type == 0:
                    hour = np.random.choice([0, 1, 2, 3, 22, 23])  # Transaksi dini hari di luar jam pasar
                    amount = round(np.random.uniform(15000.0, 45000.0), 2)
                    reason = f"Transaksi pasar uang/valas bernilai Rp {amount} Juta dieksekusi di luar jam pasar resmi (Jam {hour:02d}:00)"
                else:
                    hour = np.random.randint(8, 16)
                    amount = round(np.random.uniform(30000.0, 65000.0), 2)
                    reason = f"Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office"

            elif cat == 'Payment':
                channel = np.random.choice(['Mobile / Internet Banking', 'Host-to-Host Core Banking'])
                interest = 0.0
                doc_status = 'Lengkap'
                sub_type = a_idx % 2
                if sub_type == 0:
                    hour = np.random.choice([0, 1, 2, 3, 4])  # Transfer jumbo dini hari
                    amount = round(np.random.uniform(500.0, 4500.0), 2)
                    reason = f"Transfer pembayaran jumlah masif (Rp {amount} Juta) pada dini hari (Jam {hour:02d}:00) tanpa maker-checker"
                else:
                    hour = np.random.randint(9, 21)
                    amount = round(np.random.uniform(1000.0, 5000.0), 2)
                    reason = f"Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)"

            elif cat == 'KYC':
                channel = 'Cabang / Teller'
                amount = 0.0
                interest = 0.0
                hour = np.random.randint(8, 17)
                sub_type = a_idx % 2
                if sub_type == 0:
                    doc_status = 'Tidak Lengkap'
                    reason = "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap"
                else:
                    doc_status = 'Lengkap'
                    reason = "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)"

            else:  # IT Control
                channel = 'Host-to-Host Core Banking'
                amount = 0.0
                interest = 0.0
                doc_status = 'Lengkap'
                sub_type = a_idx % 2
                if sub_type == 0:
                    hour = np.random.choice([1, 2, 3, 4])
                    failed_login = np.random.randint(6, 15)  # Brute force login
                    reason = f"Percobaan login gagal berulang kali ({failed_login} kali) pada jam tidak wajar mengindikasikan serangan Brute Force"
                else:
                    hour = np.random.choice([0, 1, 2, 23])
                    failed_login = 0
                    reason = f"Aktivitas akses administratif basis data core banking pada dini hari (Jam {hour:02d}:00) tanpa tiket Change Request"

            records.append({
                'Kategori': cat,
                'Entitas': entity,
                'Kanal Transaksi': channel,
                'Nilai Transaksi (Juta Rp)': amount,
                'Bunga/Margin (%)': interest,
                'Jam Transaksi (0-23)': hour,
                'Status Dokumen': doc_status,
                'Jumlah Gagal Login': failed_login,
                'Tingkat Risiko Nasabah': risk_profile,
                'Status Otorisasi / Maker-Checker': auth_status,
                'Deviasi terhadap Profil Historis (%)': hist_dev,
                'TARGET: is_anomaly': 'Ya (Anomali)',
                'TARGET: Impact (1-5)': impact,
                'TARGET: Likelihood (1-5)': likelihood,
                'Alasan Anomali': reason
            })

    df = pd.DataFrame(records)

    # Shuffle dataset for true random distribution
    df = df.sample(frac=1.0, random_state=42).reset_index(drop=True)
    df['ID Transaksi'] = [f"BANK-TRX-{i+1:04d}" for i in range(len(df))]

    # Order columns preserving existing columns
    col_order = [
        'ID Transaksi', 'Kategori', 'Entitas', 'Kanal Transaksi',
        'Nilai Transaksi (Juta Rp)', 'Bunga/Margin (%)', 'Jam Transaksi (0-23)',
        'Status Dokumen', 'Jumlah Gagal Login', 'Tingkat Risiko Nasabah',
        'Status Otorisasi / Maker-Checker', 'Deviasi terhadap Profil Historis (%)',
        'TARGET: is_anomaly', 'TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)',
        'Alasan Anomali'
    ]
    df = df[col_order]

    output_path = os.path.join(CURRENT_DIR, 'bank_transaction_anomaly_data.csv')
    df.to_csv(output_path, index=False)

    n_anom = (df['TARGET: is_anomaly'] == 'Ya (Anomali)').sum()
    n_norm = (df['TARGET: is_anomaly'] == 'Tidak').sum()
    print(f"[SUCCESS] Generated {len(df)} audit-ready Bank Transaction records: {output_path}")
    print(f"Total Normal: {n_norm} | Total Anomaly: {n_anom} ({n_anom/len(df)*100:.2f}%)")
    print("Breakdown per Category:")
    print(pd.crosstab(df['Kategori'], df['TARGET: is_anomaly']))

if __name__ == "__main__":
    generate()
