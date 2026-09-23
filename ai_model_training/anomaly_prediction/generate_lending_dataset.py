import os
import numpy as np
import pandas as pd

def generate():
    np.random.seed(101)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1200
    N_ANOMALIES = 150
    N_NORMAL = N_SAMPLES - N_ANOMALIES

    entities = ['Kantor Pusat', 'Cabang Jakarta', 'Cabang Surabaya', 'Cabang Bandung', 'Cabang Medan', 'Cabang Bali']
    products = [
        'Kredit Pemilikan Rumah (KPR)',
        'Kredit Modal Kerja (KMK)',
        'Kredit Investasi (KI)',
        'Kredit Multiguna (KMG)',
        'Kredit Usaha Rakyat (KUR) / Mikro',
        'Kredit Sindikasi Korporasi'
    ]

    records = []

    # =========================================================================
    # 1. GENERATE NORMAL LENDING TRANSACTIONS (1050 records)
    # =========================================================================
    for i in range(N_NORMAL):
        txn_id = f"LEND-{i+1:04d}"
        entity = np.random.choice(entities, p=[0.25, 0.25, 0.15, 0.15, 0.10, 0.10])
        product = np.random.choice(products, p=[0.25, 0.25, 0.15, 0.15, 0.12, 0.08])

        # Plafon & Bunga berdasarkan produk
        if product == 'Kredit Usaha Rakyat (KUR) / Mikro':
            amount = round(np.random.uniform(50.0, 500.0), 2)
            interest = round(np.random.uniform(6.0, 7.5), 2)
            ltv = round(np.random.uniform(50.0, 75.0), 2)
        elif product == 'Kredit Pemilikan Rumah (KPR)':
            amount = round(np.random.uniform(300.0, 4500.0), 2)
            interest = round(np.random.uniform(7.5, 10.5), 2)
            ltv = round(np.random.uniform(60.0, 85.0), 2)
        elif product == 'Kredit Multiguna (KMG)':
            amount = round(np.random.uniform(150.0, 2500.0), 2)
            interest = round(np.random.uniform(10.0, 14.0), 2)
            ltv = round(np.random.uniform(55.0, 75.0), 2)
        elif product == 'Kredit Modal Kerja (KMK)':
            amount = round(np.random.uniform(500.0, 30000.0), 2)
            interest = round(np.random.uniform(8.5, 12.5), 2)
            ltv = round(np.random.uniform(60.0, 80.0), 2)
        elif product == 'Kredit Investasi (KI)':
            amount = round(np.random.uniform(1000.0, 45000.0), 2)
            interest = round(np.random.uniform(8.75, 12.0), 2)
            ltv = round(np.random.uniform(55.0, 75.0), 2)
        else:  # Kredit Sindikasi Korporasi
            amount = round(np.random.uniform(20000.0, 75000.0), 2)
            interest = round(np.random.uniform(8.0, 10.5), 2)
            ltv = round(np.random.uniform(50.0, 70.0), 2)

        # Nilai taksasi agunan dihitung dari LTV (Agunan > Plafon untuk data normal)
        collateral_value = round(amount / (ltv / 100.0), 2)

        doc_status = 'Lengkap & Terikat Sempurna'
        slik_status = np.random.choice(['Kol 1 (Lancar)', 'Kol 2 (Dalam Perhatian Khusus)'], p=[0.90, 0.10])
        bkmk_status = 'Sesuai Limit Wewenang Komite Kredit'
        bmpk_status = 'Sesuai Batas BMPK (<10% Modal)'
        insurance_status = 'Diasuransikan Penuh (All Risk)'
        purpose_status = 'Sesuai Proposal Usaha'

        # Target risiko normal
        impact = np.random.choice([1, 2], p=[0.75, 0.25])
        likelihood = np.random.choice([1, 2], p=[0.70, 0.30])

        records.append({
            'ID Transaksi': txn_id,
            'Entitas': entity,
            'Produk': product,
            'Nilai Plafon (Juta Rp)': amount,
            'Bunga (%)': interest,
            'Nilai Taksasi Agunan (Juta Rp)': collateral_value,
            'Rasio LTV (%)': ltv,
            'Status Dokumen': doc_status,
            'Kolektibilitas SLIK OJK (IDEB)': slik_status,
            'Otorisasi Batas Kewenangan (BKMK)': bkmk_status,
            'Kepatuhan Batas BMPK OJK': bmpk_status,
            'Status Penutupan Asuransi Agunan': insurance_status,
            'Verifikasi Tujuan Penggunaan (Side-Streaming)': purpose_status,
            'TARGET: is_anomaly': 'Tidak',
            'TARGET: Impact (1-5)': impact,
            'TARGET: Likelihood (1-5)': likelihood,
            'Alasan Anomali': '-'
        })

    # =========================================================================
    # 2. GENERATE REALISTIC ANOMALY LENDING TRANSACTIONS (150 records)
    # =========================================================================
    typologies = [
        'bkmk_limit_breach',        # Tipologi 1: Melampaui batas kewenangan cabang
        'under_collateralized_ltv', # Tipologi 2: LTV > 100%, agunan defisit
        'deficient_documentation',  # Tipologi 3: Dokumen syarat efektif / APHT tidak lengkap
        'npl_slik_approval',        # Tipologi 4: Persetujuan debitur macet SLIK tanpa mitigasi
        'affiliated_party_bmpk',    # Tipologi 5: Bunga murah ekstrem pihak terafiliasi melanggar BMPK
        'side_streaming_uninsured'  # Tipologi 6: Pengalihan dana kredit & agunan tanpa asuransi
    ]

    for j in range(N_ANOMALIES):
        txn_id = f"LEND-{N_NORMAL + j + 1:04d}"
        entity = np.random.choice(entities)
        typology = typologies[j % len(typologies)]

        if typology == 'bkmk_limit_breach':
            # Tipologi 1: Melampaui batas wewenang memutus kredit cabang
            product = np.random.choice(['Kredit Modal Kerja (KMK)', 'Kredit Investasi (KI)'])
            amount = round(np.random.uniform(35000.0, 75000.0), 2)
            interest = round(np.random.uniform(8.5, 11.5), 2)
            ltv = round(np.random.uniform(70.0, 85.0), 2)
            collateral_value = round(amount / (ltv / 100.0), 2)
            doc_status = 'Lengkap & Terikat Sempurna'
            slik_status = 'Kol 1 (Lancar)'
            bkmk_status = 'Melampaui Limit Wewenang Cabang (BKMK Breach)'
            bmpk_status = 'Sesuai Batas BMPK (<10% Modal)'
            insurance_status = 'Diasuransikan Penuh (All Risk)'
            purpose_status = 'Sesuai Proposal Usaha'
            reason = f"Pencairan kredit bernilai Rp {amount} Juta melampaui batas kewenangan memutus kredit (BKMK) komite cabang tanpa persetujuan Kantor Pusat"
            impact = 5
            likelihood = np.random.choice([4, 5], p=[0.4, 0.6])

        elif typology == 'under_collateralized_ltv':
            # Tipologi 2: LTV > 100% (agunan defisit, nilai agunan lebih kecil dari pinjaman)
            product = np.random.choice(['Kredit Pemilikan Rumah (KPR)', 'Kredit Multiguna (KMG)', 'Kredit Modal Kerja (KMK)'])
            amount = round(np.random.uniform(2000.0, 25000.0), 2)
            interest = round(np.random.uniform(9.0, 13.0), 2)
            ltv = round(np.random.uniform(115.0, 160.0), 2)
            collateral_value = round(amount / (ltv / 100.0), 2)  # agunan jauh lebih kecil dari plafon
            doc_status = 'Lengkap & Terikat Sempurna'
            slik_status = 'Kol 1 (Lancar)'
            bkmk_status = 'Sesuai Limit Wewenang Komite Kredit'
            bmpk_status = 'Sesuai Batas BMPK (<10% Modal)'
            insurance_status = 'Diasuransikan Penuh (All Risk)'
            purpose_status = 'Sesuai Proposal Usaha'
            reason = f"Rasio Loan to Value ({ltv}%) melampaui batas wajar (>100%), agunan defisit (Rp {collateral_value} Juta) tidak mencukupi plafon kredit"
            impact = 5
            likelihood = np.random.choice([4, 5], p=[0.3, 0.7])

        elif typology == 'deficient_documentation':
            # Tipologi 3: Dokumen syarat efektif / APHT tidak lengkap
            product = np.random.choice(['Kredit Pemilikan Rumah (KPR)', 'Kredit Investasi (KI)', 'Kredit Modal Kerja (KMK)'])
            amount = round(np.random.uniform(1500.0, 20000.0), 2)
            interest = round(np.random.uniform(8.5, 12.0), 2)
            ltv = round(np.random.uniform(65.0, 80.0), 2)
            collateral_value = round(amount / (ltv / 100.0), 2)
            doc_status = 'Pencairan Sebelum Syarat Efektif Lengkap (Pelanggaran)'
            slik_status = 'Kol 1 (Lancar)'
            bkmk_status = 'Sesuai Limit Wewenang Komite Kredit'
            bmpk_status = 'Sesuai Batas BMPK (<10% Modal)'
            insurance_status = 'Diasuransikan Penuh (All Risk)'
            purpose_status = 'Sesuai Proposal Usaha'
            reason = "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan (APHT/Fidusia)"
            impact = 4
            likelihood = np.random.choice([4, 5], p=[0.5, 0.5])

        elif typology == 'npl_slik_approval':
            # Tipologi 4: Persetujuan debitur berstatus macet di SLIK OJK
            product = np.random.choice(['Kredit Modal Kerja (KMK)', 'Kredit Multiguna (KMG)'])
            amount = round(np.random.uniform(800.0, 15000.0), 2)
            interest = round(np.random.uniform(9.0, 13.5), 2)
            ltv = round(np.random.uniform(60.0, 75.0), 2)
            collateral_value = round(amount / (ltv / 100.0), 2)
            doc_status = 'Lengkap & Terikat Sempurna'
            slik_status = 'Kol 3-5 (NPL / Macet)'
            bkmk_status = 'Sesuai Limit Wewenang Komite Kredit'
            bmpk_status = 'Sesuai Batas BMPK (<10% Modal)'
            insurance_status = 'Diasuransikan Penuh (All Risk)'
            purpose_status = 'Sesuai Proposal Usaha'
            reason = "Penyetujuan kredit kepada debitur dengan riwayat kolektibilitas macet (Kol 3-5) pada SLIK OJK tanpa analisis mitigasi risiko"
            impact = 5
            likelihood = np.random.choice([4, 5], p=[0.2, 0.8])

        elif typology == 'affiliated_party_bmpk':
            # Tipologi 5: Bunga murah ekstrem pihak terafiliasi melanggar BMPK
            product = np.random.choice(['Kredit Investasi (KI)', 'Kredit Sindikasi Korporasi'])
            amount = round(np.random.uniform(15000.0, 60000.0), 2)
            interest = round(np.random.uniform(3.5, 5.0), 2)  # bunga ekstrem rendah di bawah CoF
            ltv = round(np.random.uniform(50.0, 70.0), 2)
            collateral_value = round(amount / (ltv / 100.0), 2)
            doc_status = 'Lengkap & Terikat Sempurna'
            slik_status = 'Kol 1 (Lancar)'
            bkmk_status = 'Sesuai Limit Wewenang Komite Kredit'
            bmpk_status = 'Pelanggaran BMPK Pihak Terafiliasi'
            insurance_status = 'Diasuransikan Penuh (All Risk)'
            purpose_status = 'Sesuai Proposal Usaha'
            reason = f"Pemberian bunga kredit murah ekstrem ({interest}%) kepada pihak terafiliasi bank melanggar batas ketentuan BMPK OJK"
            impact = 5
            likelihood = np.random.choice([4, 5], p=[0.4, 0.6])

        else:  # side_streaming_uninsured
            # Tipologi 6: Pengalihan dana kredit & agunan tanpa asuransi
            product = 'Kredit Modal Kerja (KMK)'
            amount = round(np.random.uniform(5000.0, 25000.0), 2)
            interest = round(np.random.uniform(9.5, 12.5), 2)
            ltv = round(np.random.uniform(70.0, 85.0), 2)
            collateral_value = round(amount / (ltv / 100.0), 2)
            doc_status = 'Lengkap & Terikat Sempurna'
            slik_status = 'Kol 1 (Lancar)'
            bkmk_status = 'Sesuai Limit Wewenang Komite Kredit'
            bmpk_status = 'Sesuai Batas BMPK (<10% Modal)'
            insurance_status = 'Tanpa Asuransi Agunan (Risiko Tinggi)'
            purpose_status = 'Indikasi Pengalihan Dana / Side-Streaming'
            reason = "Indikasi penyimpangan penggunaan fasilitas kredit modal kerja (Side-Streaming) dan agunan tidak ditutup asuransi"
            impact = 4
            likelihood = np.random.choice([4, 5], p=[0.5, 0.5])

        records.append({
            'ID Transaksi': txn_id,
            'Entitas': entity,
            'Produk': product,
            'Nilai Plafon (Juta Rp)': amount,
            'Bunga (%)': interest,
            'Nilai Taksasi Agunan (Juta Rp)': collateral_value,
            'Rasio LTV (%)': ltv,
            'Status Dokumen': doc_status,
            'Kolektibilitas SLIK OJK (IDEB)': slik_status,
            'Otorisasi Batas Kewenangan (BKMK)': bkmk_status,
            'Kepatuhan Batas BMPK OJK': bmpk_status,
            'Status Penutupan Asuransi Agunan': insurance_status,
            'Verifikasi Tujuan Penggunaan (Side-Streaming)': purpose_status,
            'TARGET: is_anomaly': 'Ya (Anomali)',
            'TARGET: Impact (1-5)': impact,
            'TARGET: Likelihood (1-5)': likelihood,
            'Alasan Anomali': reason
        })

    df = pd.DataFrame(records)

    # Shuffle dataset reproducibility
    df = df.sample(frac=1.0, random_state=101).reset_index(drop=True)
    df['ID Transaksi'] = [f"LEND-{i+1:04d}" for i in range(len(df))]

    output_path = os.path.join(CURRENT_DIR, 'lending_data.csv')
    df.to_csv(output_path, index=False)

    n_anom = (df['TARGET: is_anomaly'] == 'Ya (Anomali)').sum()
    n_norm = (df['TARGET: is_anomaly'] == 'Tidak').sum()
    print(f"[SUCCESS] Generated {len(df)} audit-ready Lending records: {output_path}")
    print(f"Total Normal: {n_norm} | Total Anomaly: {n_anom} ({n_anom/len(df)*100:.2f}%)")

if __name__ == "__main__":
    generate()
