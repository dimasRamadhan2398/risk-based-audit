import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(42)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1200
    N_ANOMALIES = 152
    N_NORMAL = N_SAMPLES - N_ANOMALIES

    entities = ['Kantor Pusat', 'Cabang Jakarta', 'Cabang Surabaya', 'Cabang Bandung', 'Cabang Medan', 'Cabang Bali']
    products = [
        'Deposito Berjangka', 'Deposito On Call', 
        'Tabungan Reguler', 'Tabungan Bisnis', 
        'Giro Korporasi', 'Giro Individu'
    ]
    
    LPS_RATE_CAP = 4.25  # Batas Suku Bunga Penjaminan LPS acuan

    records = []

    # =========================================================================
    # 1. GENERATE NORMAL FUNDING TRANSACTIONS (~1048 records)
    # =========================================================================
    for i in range(N_NORMAL):
        txn_id = f"FUND-{i+1:04d}"
        entity = np.random.choice(entities, p=[0.25, 0.25, 0.15, 0.15, 0.10, 0.10])
        product = np.random.choice(products, p=[0.30, 0.10, 0.25, 0.15, 0.12, 0.08])
        
        # Jenis transaksi
        if 'Deposito' in product:
            txn_type = np.random.choice([
                'Penempatan Baru', 'Pembaruan Otomatis (ARO)', 
                'Penarikan / Pencairan', 'Pencairan Sebelum Jatuh Tempo (Early Break)'
            ], p=[0.45, 0.35, 0.15, 0.05])
        else:
            txn_type = np.random.choice([
                'Penempatan Baru', 'Penarikan / Pencairan', 'Setoran Tambahan'
            ], p=[0.35, 0.40, 0.25])

        # Nominal berdasarkan produk
        if product == 'Tabungan Reguler':
            amount = round(np.random.uniform(5.0, 350.0), 2)
            interest = round(np.random.uniform(0.75, 2.0), 2)
        elif product == 'Tabungan Bisnis':
            amount = round(np.random.uniform(50.0, 2500.0), 2)
            interest = round(np.random.uniform(1.25, 2.75), 2)
        elif product in ['Giro Korporasi', 'Giro Individu']:
            amount = round(np.random.uniform(100.0, 8000.0), 2)
            interest = round(np.random.uniform(1.0, 2.25), 2)
        elif product == 'Deposito On Call':
            amount = round(np.random.uniform(500.0, 15000.0), 2)
            interest = round(np.random.uniform(2.5, 3.75), 2)
        else:  # Deposito Berjangka
            amount = round(np.random.uniform(50.0, 18000.0), 2)
            # Normal rate: mostly under LPS cap, 15% legitimate ALCO special rate with signed waiver
            is_legit_special_rate = (np.random.rand() < 0.15)
            if is_legit_special_rate:
                interest = round(np.random.uniform(4.35, 5.75), 2)
            else:
                interest = round(np.random.uniform(2.75, 4.25), 2)

        # Status suku bunga & kepatuhan LPS
        if interest > LPS_RATE_CAP:
            alco_approval = 'Sesuai Ketentuan ALCO'
            lps_waiver = 'LPS Ditandatangani'
        else:
            alco_approval = 'Bukan Special Rate'
            lps_waiver = 'Tidak Diperlukan (< Batas LPS)'

        # Early Break status
        if txn_type == 'Pencairan Sebelum Jatuh Tempo (Early Break)':
            penalty_status = 'Penalti Dikenakan Sesuai Ketentuan'
        else:
            penalty_status = 'Bukan Early Break'

        structuring_flag = 'Tidak Ada Indikasi'
        account_status = 'Aktif Normal'
        authorization_status = 'Disetujui Branch Manager / Operasional'

        # Impact & Likelihood for normal operations
        impact = np.random.choice([1, 2], p=[0.70, 0.30])
        likelihood = np.random.choice([1, 2], p=[0.75, 0.25])

        records.append({
            'ID Transaksi': txn_id,
            'Entitas': entity,
            'Produk': product,
            'Jenis Transaksi': txn_type,
            'Nilai Transaksi (Juta Rp)': amount,
            'Bunga (%)': interest,
            'Batas Bunga Penjaminan LPS (%)': LPS_RATE_CAP,
            'Persetujuan Special Rate ALCO': alco_approval,
            'Surat Pelepasan Hak Penjaminan LPS': lps_waiver,
            'Status Penalti Early Break': penalty_status,
            'Indikasi Structuring Bilyet Simpanan': structuring_flag,
            'Status Rekening Asal': account_status,
            'Verifikasi Otorisasi Pejabat': authorization_status,
            'TARGET: is_anomaly': 'Tidak',
            'TARGET: Impact (1-5)': impact,
            'TARGET: Likelihood (1-5)': likelihood,
            'Alasan Anomali': '-'
        })

    # =========================================================================
    # 2. GENERATE REALISTIC ANOMALY FUNDING TRANSACTIONS (152 records)
    # =========================================================================
    typologies = [
        'lps_rate_violation',        # Tipologi 1: Bunga > LPS tanpa waiver & tanpa izin ALCO
        'early_break_penalty_bypass',# Tipologi 2: Early break tanpa penalti & override teller
        'deposit_structuring',       # Tipologi 3: Structuring mendekati batas LPS Rp 2 Miliar
        'dormant_account_takeover',  # Tipologi 4: Rekening dormant aktif tiba-tiba dikuras
        'affiliated_party_high_rate',# Tipologi 5: Bunga ekstrem tidak wajar pihak terkait
        'unverified_massive_cash_wd' # Tipologi 6: Penarikan masif tanpa verifikasi CTR
    ]

    for j in range(N_ANOMALIES):
        txn_id = f"FUND-{N_NORMAL + j + 1:04d}"
        entity = np.random.choice(entities)
        typology = typologies[j % len(typologies)]

        if typology == 'lps_rate_violation':
            # Tipologi 1: Bunga > LPS tanpa izin ALCO & tanpa surat pelepasan LPS
            product = np.random.choice(['Deposito Berjangka', 'Deposito On Call'])
            txn_type = np.random.choice(['Penempatan Baru', 'Pembaruan Otomatis (ARO)'])
            amount = round(np.random.uniform(500.0, 15000.0), 2)
            interest = round(np.random.uniform(6.5, 9.75), 2)
            alco_approval = 'Tanpa Persetujuan ALCO'
            lps_waiver = 'Tanpa Surat Pelepasan (Pelanggaran)'
            penalty_status = 'Bukan Early Break'
            structuring_flag = 'Tidak Ada Indikasi'
            account_status = 'Aktif Normal'
            authorization_status = 'Override Teller / Tanpa Otorisasi'
            reason = f"Bunga Deposito ({interest}%) melampaui batas LPS ({LPS_RATE_CAP}%) tanpa Surat Pelepasan LPS & tanpa persetujuan ALCO"
            impact = 5
            likelihood = np.random.choice([4, 5], p=[0.4, 0.6])

        elif typology == 'early_break_penalty_bypass':
            # Tipologi 2: Early break tanpa penalti & override teller
            product = 'Deposito Berjangka'
            txn_type = 'Pencairan Sebelum Jatuh Tempo (Early Break)'
            amount = round(np.random.uniform(1000.0, 20000.0), 2)
            interest = round(np.random.uniform(3.5, 4.25), 2)
            alco_approval = 'Bukan Special Rate'
            lps_waiver = 'Tidak Diperlukan (< Batas LPS)'
            penalty_status = 'Bebas Penalti Tanpa Otorisasi'
            structuring_flag = 'Tidak Ada Indikasi'
            account_status = 'Aktif Normal'
            authorization_status = 'Override Teller / Tanpa Otorisasi'
            reason = "Pencairan Deposito sebelum jatuh tempo (Early Break) tanpa pengenaan penalti dan tanpa otorisasi Branch Manager"
            impact = 4
            likelihood = np.random.choice([4, 5], p=[0.5, 0.5])

        elif typology == 'deposit_structuring':
            # Tipologi 3: Structuring mendekati batas LPS Rp 2 Miliar (Rp 1.950 - 1.995 Juta)
            product = 'Deposito Berjangka'
            txn_type = 'Penempatan Baru'
            amount = round(np.random.uniform(1950.0, 1998.0), 2)
            interest = round(np.random.uniform(3.75, 4.25), 2)
            alco_approval = 'Bukan Special Rate'
            lps_waiver = 'Tidak Diperlukan (< Batas LPS)'
            penalty_status = 'Bukan Early Break'
            structuring_flag = 'Potensi Structuring Bilyet Deposito'
            account_status = 'Aktif Normal'
            authorization_status = 'Disetujui Branch Manager / Operasional'
            reason = f"Indikasi Structuring bilyet deposito berulang nominal Rp {amount} Juta mendekati batas maksimal penjaminan LPS (Rp 2 Miliar)"
            impact = np.random.choice([4, 5], p=[0.6, 0.4])
            likelihood = 4

        elif typology == 'dormant_account_takeover':
            # Tipologi 4: Rekening dormant aktif tiba-tiba dikuras
            product = np.random.choice(['Tabungan Reguler', 'Tabungan Bisnis', 'Giro Individu'])
            txn_type = 'Penarikan / Pencairan'
            amount = round(np.random.uniform(800.0, 6500.0), 2)
            interest = round(np.random.uniform(1.0, 2.0), 2)
            alco_approval = 'Bukan Special Rate'
            lps_waiver = 'Tidak Diperlukan (< Batas LPS)'
            penalty_status = 'Bukan Early Break'
            structuring_flag = 'Tidak Ada Indikasi'
            account_status = 'Dormant Diaktifkan Tiba-tiba'
            authorization_status = 'Override Teller / Tanpa Otorisasi'
            reason = "Rekening pasif (Dormant) diaktifkan tiba-tiba diikuti penarikan dana masif tanpa otorisasi Pejabat Cabang"
            impact = 5
            likelihood = np.random.choice([4, 5], p=[0.3, 0.7])

        elif typology == 'affiliated_party_high_rate':
            # Tipologi 5: Bunga ekstrem tidak wajar pihak terkait
            product = np.random.choice(['Deposito Berjangka', 'Giro Korporasi'])
            txn_type = 'Penempatan Baru'
            amount = round(np.random.uniform(2000.0, 30000.0), 2)
            interest = round(np.random.uniform(8.5, 11.5), 2)
            alco_approval = 'Tanpa Persetujuan ALCO'
            lps_waiver = 'LPS Ditandatangani'  # Disiasati tanda tangan tapi tanpa izin ALCO
            penalty_status = 'Bukan Early Break'
            structuring_flag = 'Tidak Ada Indikasi'
            account_status = 'Aktif Normal'
            authorization_status = 'Override Teller / Tanpa Otorisasi'
            reason = f"Pemberian suku bunga pendanaan ekstrem ({interest}%) di luar kewajaran pasar tanpa persetujuan ALCO (potensi conflict of interest)"
            impact = 5
            likelihood = np.random.choice([4, 5], p=[0.4, 0.6])

        else:  # unverified_massive_cash_wd
            # Tipologi 6: Penarikan masif tanpa verifikasi CTR
            product = np.random.choice(['Giro Korporasi', 'Tabungan Bisnis'])
            txn_type = 'Penarikan / Pencairan'
            amount = round(np.random.uniform(18000.0, 48000.0), 2)
            interest = round(np.random.uniform(1.0, 2.0), 2)
            alco_approval = 'Bukan Special Rate'
            lps_waiver = 'Tidak Diperlukan (< Batas LPS)'
            penalty_status = 'Bukan Early Break'
            structuring_flag = 'Tidak Ada Indikasi'
            account_status = 'Aktif Normal'
            authorization_status = 'Override Teller / Tanpa Otorisasi'
            reason = f"Penarikan dana masif sebesar Rp {amount} Juta tanpa verifikasi Cash Transaction Report (CTR) dan otorisasi Pejabat"
            impact = np.random.choice([4, 5], p=[0.5, 0.5])
            likelihood = 4

        records.append({
            'ID Transaksi': txn_id,
            'Entitas': entity,
            'Produk': product,
            'Jenis Transaksi': txn_type,
            'Nilai Transaksi (Juta Rp)': amount,
            'Bunga (%)': interest,
            'Batas Bunga Penjaminan LPS (%)': LPS_RATE_CAP,
            'Persetujuan Special Rate ALCO': alco_approval,
            'Surat Pelepasan Hak Penjaminan LPS': lps_waiver,
            'Status Penalti Early Break': penalty_status,
            'Indikasi Structuring Bilyet Simpanan': structuring_flag,
            'Status Rekening Asal': account_status,
            'Verifikasi Otorisasi Pejabat': authorization_status,
            'TARGET: is_anomaly': 'Ya (Anomali)',
            'TARGET: Impact (1-5)': impact,
            'TARGET: Likelihood (1-5)': likelihood,
            'Alasan Anomali': reason
        })

    # Convert to DataFrame
    df = pd.DataFrame(records)

    # Shuffle dataset reproducibility
    df = df.sample(frac=1.0, random_state=42).reset_index(drop=True)
    df['ID Transaksi'] = [f"FUND-{i+1:04d}" for i in range(len(df))]

    output_path = os.path.join(CURRENT_DIR, 'funding_data.csv')
    df.to_csv(output_path, index=False)
    
    n_anom = (df['TARGET: is_anomaly'] == 'Ya (Anomali)').sum()
    n_norm = (df['TARGET: is_anomaly'] == 'Tidak').sum()
    print(f"[SUCCESS] Generated {len(df)} audit-ready Funding records: {output_path}")
    print(f"Total Normal: {n_norm} | Total Anomaly: {n_anom} ({n_anom/len(df)*100:.2f}%)")

if __name__ == "__main__":
    generate()
