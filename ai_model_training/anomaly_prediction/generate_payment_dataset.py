import os
import numpy as np
import pandas as pd

def generate():
    """
    Generates realistic, audit-ready, high-quality Banking Payment Anomaly Dataset
    compliant with Bank Indonesia BI-FAST/RTGS/SKNBI Regulations, OJK Fraud Risk Management,
    and PPATK Suspicious Transaction Reporting.
    """
    np.random.seed(103)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1200

    ENTITIES = [
        'Head Office', 'Jakarta Branch', 'Surabaya Branch', 
        'Bandung Branch', 'Medan Branch', 'Digital Banking Unit'
    ]
    PRODUCTS = ['BI-FAST', 'RTGS', 'SKN / Kliring', 'Transfer Internal', 'SWIFT / Remittance Valas']
    CHANNELS = ['Mobile Banking', 'Internet Banking Corporate', 'Teller Cabang', 'ATM', 'API Host-to-Host']
    CUSTOMER_TYPES = ['Individu / Retail', 'Korporasi / Komersial']

    records = []

    for i in range(N_SAMPLES):
        log_id = f"PAYM-{i+1:04d}"
        entity = np.random.choice(ENTITIES)
        
        # Link channel and customer type logically
        cust_type = np.random.choice(CUSTOMER_TYPES, p=[0.70, 0.30])
        if cust_type == 'Korporasi / Komersial':
            product = np.random.choice(['RTGS', 'Transfer Internal', 'SWIFT / Remittance Valas', 'SKN / Kliring'], p=[0.45, 0.25, 0.20, 0.10])
            channel = np.random.choice(['Internet Banking Corporate', 'API Host-to-Host', 'Teller Cabang'], p=[0.60, 0.25, 0.15])
        else:
            product = np.random.choice(['BI-FAST', 'Transfer Internal', 'SKN / Kliring'], p=[0.60, 0.30, 0.10])
            channel = np.random.choice(['Mobile Banking', 'ATM', 'Teller Cabang'], p=[0.75, 0.20, 0.05])

        is_anom_scenario = np.random.rand() < 0.125
        scenario_type = None

        if is_anom_scenario:
            scenario_type = np.random.choice([
                'midnight_large_new_beneficiary',
                'smurfing_structuring',
                'velocity_rapid_burst',
                'dormant_account_drain',
                'high_risk_offshore_transfer',
                'channel_limit_breach'
            ], p=[0.20, 0.20, 0.20, 0.15, 0.15, 0.10])

        alasan = []
        is_anom = False

        if scenario_type == 'midnight_large_new_beneficiary':
            # Scenario 1: Suspicious Large Midnight Transfer to New Beneficiary on Weekend
            amount = round(float(np.random.uniform(800.0, 4800.0)), 2)
            jam = int(np.random.choice([23, 0, 1, 2, 3, 4]))
            day = int(np.random.choice([6, 7])) # Weekend night
            new_beneficiary = 'Ya'
            acct_status = 'Aktif Normal'
            freq_24h = int(np.random.randint(1, 4))
            jurisdiction = 'Domestik Indonesia'
            auth_method = np.random.choice(['SMS OTP Standar', 'Override Tanpa 2FA'])

            is_anom = True
            alasan.append(f"Transfer Jumlah Sangat Besar di Dini Hari (Jam {jam}) Akhir Pekan ke Penerima Baru")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'smurfing_structuring':
            # Scenario 2: Smurfing / Structuring just below the PPATK Rp 500 Juta threshold
            # e.g., Rp 480.00 - Rp 498.00 Juta repeated
            amount = round(float(np.random.uniform(480.0, 498.5)), 2)
            jam = int(np.random.randint(8, 18))
            day = int(np.random.randint(1, 6))
            new_beneficiary = np.random.choice(['Ya', 'Tidak'])
            acct_status = 'Aktif Normal'
            freq_24h = int(np.random.randint(5, 12)) # Multiple transfers in 24h
            jurisdiction = 'Domestik Indonesia'
            auth_method = 'MFA Biometrik / Token Valid'

            is_anom = True
            alasan.append(f"Indikasi Smurfing / Structuring: Transaksi Terpecah Berulang Tepat di Bawah Batas Lapor PPATK Rp 500 Juta ({amount} Juta)")
            impact = 5
            likelihood = 5

        elif scenario_type == 'velocity_rapid_burst':
            # Scenario 3: Velocity Risk / Rapid burst of multiple transactions (Account Takeover)
            amount = round(float(np.random.uniform(50.0, 350.0)), 2)
            jam = int(np.random.randint(0, 24))
            day = int(np.random.randint(1, 8))
            new_beneficiary = 'Ya'
            acct_status = 'Aktif Normal'
            freq_24h = int(np.random.randint(15, 30)) # High burst
            jurisdiction = 'Domestik Indonesia'
            auth_method = np.random.choice(['SMS OTP Standar', 'Override Tanpa 2FA'])

            is_anom = True
            alasan.append(f"Lonjakan Frekuensi Transaksi Abnormal (Velocity Risk: {freq_24h}x dalam 24 Jam) ke Penerima Baru")
            impact = int(np.random.randint(4, 6))
            likelihood = 5

        elif scenario_type == 'dormant_account_drain':
            # Scenario 4: Dormant Account Sudden Drain
            amount = round(float(np.random.uniform(300.0, 2500.0)), 2)
            jam = int(np.random.randint(6, 23))
            day = int(np.random.randint(1, 8))
            new_beneficiary = 'Ya'
            acct_status = 'Dormant / Pasif >180 Hari'
            freq_24h = int(np.random.randint(1, 3))
            jurisdiction = 'Domestik Indonesia'
            auth_method = 'Override Tanpa 2FA'

            is_anom = True
            alasan.append("Pengurasan Rekening Pasif (Dormant Account Drain): Penarikan Dana Besar Tanpa Otentikasi Memadai")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'high_risk_offshore_transfer':
            # Scenario 5: High-Risk Cross-Border Wire Transfer / Tax Haven
            product = 'SWIFT / Remittance Valas'
            channel = np.random.choice(['Internet Banking Corporate', 'Teller Cabang'])
            cust_type = 'Korporasi / Komersial'
            amount = round(float(np.random.uniform(1500.0, 8500.0)), 2)
            jam = int(np.random.randint(8, 20))
            day = int(np.random.randint(1, 6))
            new_beneficiary = 'Ya'
            acct_status = 'Aktif Normal'
            freq_24h = int(np.random.randint(1, 4))
            jurisdiction = 'High-Risk Offshore Jurisdiction'
            auth_method = np.random.choice(['MFA Biometrik / Token Valid', 'SMS OTP Standar'])

            is_anom = True
            alasan.append("Transfer Valas Lintas Negara ke Yurisdiksi Berisiko Tinggi (High-Risk Offshore Tax Haven)")
            impact = 5
            likelihood = 4

        elif scenario_type == 'channel_limit_breach':
            # Scenario 6: Channel Limit Breach (e.g. BI-FAST > 250 Juta limit set by Bank Indonesia)
            product = 'BI-FAST'
            channel = 'Mobile Banking'
            amount = round(float(np.random.uniform(280.0, 650.0)), 2) # BI-FAST max is 250 Juta
            jam = int(np.random.randint(8, 18))
            day = int(np.random.randint(1, 6))
            new_beneficiary = 'Tidak'
            acct_status = 'Aktif Normal'
            freq_24h = int(np.random.randint(1, 3))
            jurisdiction = 'Domestik Indonesia'
            auth_method = 'MFA Biometrik / Token Valid'

            is_anom = True
            alasan.append(f"Pelanggaran Limit Kanal Pembayaran: Transaksi BI-FAST Melebihi Batas Maksimal Regulasi Bank Indonesia (>Rp 250 Juta)")
            impact = 4
            likelihood = int(np.random.randint(3, 5))

        else:
            # NORMAL COMPLIANT TRANSACTIONS
            # Realistic amounts depending on customer type
            if cust_type == 'Korporasi / Komersial':
                amount = round(float(np.clip(np.random.exponential(scale=350.0) + 20.0, 5.0, 2500.0)), 2)
                # Keep corporate below the strict smurfing pattern (not in 480-498 range with high freq)
                if 480.0 <= amount <= 498.5:
                    amount = round(amount + 40.0, 2)
            else: # Retail
                if product == 'BI-FAST':
                    amount = round(float(np.clip(np.random.exponential(scale=15.0) + 0.5, 0.1, 240.0)), 2)
                else:
                    amount = round(float(np.clip(np.random.exponential(scale=25.0) + 1.0, 0.5, 200.0)), 2)

            # Normal banking hours: 85% daytime (07:00 - 21:00), 15% other hours
            if np.random.rand() < 0.85:
                jam = int(np.random.randint(7, 22))
                day = int(np.random.randint(1, 6)) # Monday-Friday
            else:
                jam = int(np.random.randint(6, 23))
                day = int(np.random.randint(1, 8))

            new_beneficiary = np.random.choice(['Ya', 'Tidak'], p=[0.15, 0.85])
            acct_status = np.random.choice(['Aktif Normal', 'Akun Baru <7 Hari'], p=[0.92, 0.08])
            freq_24h = int(np.random.choice([1, 2, 3, 4], p=[0.60, 0.25, 0.10, 0.05]))
            
            if product == 'SWIFT / Remittance Valas':
                jurisdiction = 'Standard International'
            else:
                jurisdiction = 'Domestik Indonesia'

            auth_method = np.random.choice(['MFA Biometrik / Token Valid', 'SMS OTP Standar'], p=[0.75, 0.25])

            is_anom = False
            impact = int(np.random.choice([1, 2], p=[0.65, 0.35]))
            likelihood = int(np.random.choice([1, 2], p=[0.70, 0.30]))

        target_is_anom = 'Ya (Anomali)' if is_anom else 'Tidak'
        target_reason = " & ".join(alasan) if is_anom else '-'

        records.append({
            'ID Transaksi': log_id,
            'Entitas': entity,
            'Produk': product,
            'Kanal Pembayaran': channel,
            'Tipe Nasabah': cust_type,
            'Nilai Transaksi (Juta Rp)': amount,
            'Jam Transaksi (0-23)': jam,
            'Hari Transaksi (1-7)': day,
            'Penerima Baru': new_beneficiary,
            'Status Rekening Pengirim': acct_status,
            'Frekuensi Transaksi 24 Jam': freq_24h,
            'Yurisdiksi Tujuan': jurisdiction,
            'Metode Otentikasi': auth_method,
            'TARGET: is_anomaly': target_is_anom,
            'TARGET: Impact (1-5)': impact,
            'TARGET: Likelihood (1-5)': likelihood,
            'Alasan Anomali': target_reason
        })

    df = pd.DataFrame(records)

    # Sanity deduplication on features to prevent identical samples
    feature_cols = [c for c in df.columns if c not in ['ID Transaksi', 'TARGET: is_anomaly', 'TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)', 'Alasan Anomali']]
    df = df.drop_duplicates(subset=feature_cols).reset_index(drop=True)

    # Re-index ID Transaksi
    df['ID Transaksi'] = [f"PAYM-{i+1:04d}" for i in range(len(df))]

    output_path = os.path.join(CURRENT_DIR, 'payment_data.csv')
    df.to_csv(output_path, index=False)

    anom_count = (df['TARGET: is_anomaly'] == 'Ya (Anomali)').sum()
    print(f"[SUCCESS] Generated {len(df)} audit-ready Payment records: {output_path}")
    print(f"Total Normal: {len(df) - anom_count} | Total Anomaly: {anom_count} ({anom_count/len(df)*100:.2f}%)")

if __name__ == "__main__":
    generate()
