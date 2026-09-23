import os
import numpy as np
import pandas as pd

def generate():
    """
    Generates realistic, audit-ready, high-quality Banking Treasury Anomaly Dataset
    compliant with Bank Indonesia Regulations (PBI Pasar Uang & Devisa, Ketentuan PDN)
    and OJK Market Risk Management Standards.
    """
    np.random.seed(102)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1200

    ENTITIES = [
        'Head Office', 'Singapore Branch', 'Treasury Jakarta', 
        'Treasury Surabaya', 'Treasury Medan'
    ]
    PRODUCTS = [
        'FX Spot', 'FX Forward / Swap', 'SBN / Government Bonds', 
        'SRBI / Sekuritas Rupiah BI', 'PUAB / Call Money', 'Repo / Reverse Repo'
    ]
    TX_TYPES = ['Buy / Purchase', 'Sell / Placement']

    records = []

    for i in range(N_SAMPLES):
        log_id = f"TRSY-{i+1:04d}"
        entity = np.random.choice(ENTITIES)
        product = np.random.choice(PRODUCTS)
        tx_type = np.random.choice(TX_TYPES)

        is_anom_scenario = np.random.rand() < 0.125
        scenario_type = None

        if is_anom_scenario:
            scenario_type = np.random.choice([
                'midnight_large_unauthorized',
                'off_market_rate_dealing',
                'dealer_trading_limit_breach',
                'counterparty_credit_limit_breach',
                'unconfirmed_mismatched_deal_slip',
                'middle_office_bypass_pdn_risk'
            ], p=[0.20, 0.20, 0.20, 0.15, 0.15, 0.10])

        alasan = []
        is_anom = False

        if scenario_type == 'midnight_large_unauthorized':
            # Scenario 1: Unauthorized large off-hours / weekend transaction
            amount = round(float(np.random.uniform(25000.0, 95000.0)), 2)
            jam = int(np.random.choice([22, 23, 0, 1, 2, 3, 4, 5]))
            day = int(np.random.choice([6, 7])) # Weekend night
            dev_rate = round(float(np.random.uniform(0.10, 0.45)), 2)
            dealer_limit = 'Dalam Limit Dealer'
            counterparty_limit = 'Limit Tersedia'
            confirmation = 'Unconfirmed Deal Slip'
            mid_office = 'Tanpa Verifikasi / Override Dealer'
            pdn_status = 'Sesuai Ketentuan BI <20%'

            is_anom = True
            alasan.append(f"Transaksi Nilai Sangat Besar Dini Hari di Luar Jam Pasar Tanpa Otorisasi (Jam {jam})")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'off_market_rate_dealing':
            # Scenario 2: Off-Market Rate Trading (Rate deviation > 1.5% from benchmark)
            amount = round(float(np.random.uniform(8000.0, 50000.0)), 2)
            jam = int(np.random.randint(9, 16))
            day = int(np.random.randint(1, 6))
            dev_rate = round(float(np.random.uniform(1.60, 6.50)), 2)
            dealer_limit = 'Dalam Limit Dealer'
            counterparty_limit = 'Limit Tersedia'
            confirmation = np.random.choice(['Konfirmasi Match', 'Unconfirmed Deal Slip'])
            mid_office = 'Tanpa Verifikasi / Override Dealer'
            pdn_status = 'Sesuai Ketentuan BI <20%'

            is_anom = True
            alasan.append(f"Transaksi dengan Deviasi Kurs / Yield Tidak Wajar (Off-Market Rate: {dev_rate:.2f}%)")
            impact = 5
            likelihood = 5

        elif scenario_type == 'dealer_trading_limit_breach':
            # Scenario 3: Dealer Trading Limit Breach
            amount = round(float(np.random.uniform(45000.0, 120000.0)), 2)
            jam = int(np.random.randint(8, 17))
            day = int(np.random.randint(1, 6))
            dev_rate = round(float(np.random.uniform(0.05, 0.30)), 2)
            dealer_limit = 'Melampaui Limit Dealer'
            counterparty_limit = 'Limit Tersedia'
            confirmation = 'Konfirmasi Match'
            mid_office = 'Tanpa Verifikasi / Override Dealer'
            pdn_status = 'Sesuai Ketentuan BI <20%'

            is_anom = True
            alasan.append(f"Pelanggaran Limit Wewenang Dealer (Dealer Trading Limit Breach: Rp {amount:.2f} Juta)")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'counterparty_credit_limit_breach':
            # Scenario 4: Counterparty Exposure Limit Exceeded
            amount = round(float(np.random.uniform(30000.0, 85000.0)), 2)
            jam = int(np.random.randint(8, 17))
            day = int(np.random.randint(1, 6))
            dev_rate = round(float(np.random.uniform(0.02, 0.25)), 2)
            dealer_limit = 'Dalam Limit Dealer'
            counterparty_limit = 'Melampaui Limit Counterparty'
            confirmation = 'Konfirmasi Match'
            mid_office = 'Tanpa Verifikasi / Override Dealer'
            pdn_status = 'Sesuai Ketentuan BI <20%'

            is_anom = True
            alasan.append("Pelanggaran Limit Kredit Counterparty Bank Lawan (Exceeded Counterparty Exposure Limit)")
            impact = int(np.random.randint(4, 6))
            likelihood = 4

        elif scenario_type == 'unconfirmed_mismatched_deal_slip':
            # Scenario 5: Settlement & Operational Risk - Mismatched or Unconfirmed Deal
            amount = round(float(np.random.uniform(10000.0, 60000.0)), 2)
            jam = int(np.random.randint(8, 17))
            day = int(np.random.randint(1, 6))
            dev_rate = round(float(np.random.uniform(0.05, 0.40)), 2)
            dealer_limit = 'Dalam Limit Dealer'
            counterparty_limit = 'Limit Tersedia'
            confirmation = np.random.choice(['Mismatched Deal', 'Unconfirmed Deal Slip'])
            mid_office = 'Tanpa Verifikasi / Override Dealer'
            pdn_status = 'Sesuai Ketentuan BI <20%'

            is_anom = True
            alasan.append(f"Risiko Operasional & Settlement: Konfirmasi Transaksi {confirmation}")
            impact = 4
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'middle_office_bypass_pdn_risk':
            # Scenario 6: Segregation of Duties (SoD) violation with Net Open Position (PDN) Risk
            product = np.random.choice(['FX Spot', 'FX Forward / Swap'])
            amount = round(float(np.random.uniform(35000.0, 100000.0)), 2)
            jam = int(np.random.randint(8, 17))
            day = int(np.random.randint(1, 6))
            dev_rate = round(float(np.random.uniform(0.10, 0.50)), 2)
            dealer_limit = 'Dalam Limit Dealer'
            counterparty_limit = 'Limit Tersedia'
            confirmation = 'Konfirmasi Match'
            mid_office = 'Tanpa Verifikasi / Override Dealer'
            pdn_status = 'Potensi Melanggar Batas PDN'

            is_anom = True
            alasan.append("Pelanggaran SoD: Dealer Melakukan Eksekusi Mandiri Tanpa Verifikasi Middle Office & Berisiko Melanggar Batas PDN BI")
            impact = 5
            likelihood = 5

        else:
            # NORMAL COMPLIANT TREASURY TRADES
            # Realistic transaction size based on product
            if 'SBN' in product or 'SRBI' in product:
                amount = round(float(np.clip(np.random.exponential(scale=15000.0) + 5000.0, 2000.0, 40000.0)), 2)
            elif 'PUAB' in product or 'Repo' in product:
                amount = round(float(np.clip(np.random.exponential(scale=10000.0) + 2000.0, 1000.0, 35000.0)), 2)
            else: # FX Spot / Forward
                amount = round(float(np.clip(np.random.exponential(scale=8000.0) + 1000.0, 500.0, 25000.0)), 2)

            # Normal interbank trading hours: 90% in official trading window (08:00 - 16:30), 10% late settlement (16:30 - 18:00)
            if np.random.rand() < 0.90:
                jam = int(np.random.randint(8, 17))
                day = int(np.random.randint(1, 6)) # Monday-Friday
            else:
                jam = int(np.random.choice([7, 17, 18]))
                day = int(np.random.randint(1, 6))

            dev_rate = round(float(np.clip(np.random.exponential(scale=0.06) + 0.01, 0.01, 0.35)), 2)
            dealer_limit = 'Dalam Limit Dealer'
            counterparty_limit = 'Limit Tersedia'
            confirmation = 'Konfirmasi Match'
            mid_office = 'Terverifikasi Middle Office'
            pdn_status = 'Sesuai Ketentuan BI <20%'

            is_anom = False
            impact = int(np.random.choice([1, 2], p=[0.65, 0.35]))
            likelihood = int(np.random.choice([1, 2], p=[0.70, 0.30]))

        target_is_anom = 'Ya (Anomali)' if is_anom else 'Tidak'
        target_reason = " & ".join(alasan) if is_anom else '-'

        records.append({
            'ID Transaksi': log_id,
            'Entitas': entity,
            'Produk': product,
            'Tipe Transaksi': tx_type,
            'Nilai Transaksi (Juta Rp)': amount,
            'Jam Transaksi (0-23)': jam,
            'Hari Transaksi (1-7)': day,
            'Deviasi Kurs / Yield (%)': dev_rate,
            'Status Limit Dealer': dealer_limit,
            'Status Limit Counterparty': counterparty_limit,
            'Status Konfirmasi Deal': confirmation,
            'Verifikasi Middle Office': mid_office,
            'Indikasi Batas PDN': pdn_status,
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
    df['ID Transaksi'] = [f"TRSY-{i+1:04d}" for i in range(len(df))]

    output_path = os.path.join(CURRENT_DIR, 'treasury_data.csv')
    df.to_csv(output_path, index=False)

    anom_count = (df['TARGET: is_anomaly'] == 'Ya (Anomali)').sum()
    print(f"[SUCCESS] Generated {len(df)} audit-ready Treasury records: {output_path}")
    print(f"Total Normal: {len(df) - anom_count} | Total Anomaly: {anom_count} ({anom_count/len(df)*100:.2f}%)")

if __name__ == "__main__":
    generate()
