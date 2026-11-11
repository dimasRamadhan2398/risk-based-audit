import os
import numpy as np
import pandas as pd

np.random.seed(42)

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
PERIODS_PER_KPI = 60  # 60 monthly periods per KPI

# Auditee Business Health KPIs aligned with Gold Zone / CAATT indicators
KPI_SPECS = {
    # ── 1. Kredit Macet / NPL per Cabang ──
    'Rasio NPL Cabang Surabaya': {'base': 3.4, 'min': 1.1, 'max': 7.8, 'noise': 0.45, 'trend': 0.03},
    'Rasio NPL Cabang Jakarta': {'base': 2.5, 'min': 0.8, 'max': 6.2, 'noise': 0.35, 'trend': 0.02},
    'Rasio NPL Cabang Bandung': {'base': 2.8, 'min': 1.0, 'max': 6.5, 'noise': 0.40, 'trend': 0.02},
    'Rasio NPL Cabang Medan': {'base': 3.1, 'min': 1.2, 'max': 7.2, 'noise': 0.42, 'trend': 0.03},
    'Rasio NPL Cabang Bali': {'base': 2.2, 'min': 0.7, 'max': 5.5, 'noise': 0.30, 'trend': -0.01},
    'Rasio NPL Cabang Semarang': {'base': 2.4, 'min': 0.9, 'max': 5.8, 'noise': 0.32, 'trend': 0.01},
    'Rasio NPL Cabang Makassar': {'base': 3.3, 'min': 1.3, 'max': 7.5, 'noise': 0.44, 'trend': 0.02},
    'Rasio NPL Cabang Palembang': {'base': 2.9, 'min': 1.0, 'max': 6.8, 'noise': 0.38, 'trend': 0.02},
    
    # ── 2. Efisiensi & Rentabilitas Keuangan ──
    'Rasio BOPO (Efisiensi Operasional)': {'base': 78.5, 'min': 65.0, 'max': 96.0, 'noise': 2.5, 'trend': 0.08},
    'Gross Profit Margin': {'base': 35.0, 'min': 15.0, 'max': 48.0, 'noise': 2.2, 'trend': -0.05},
    'Net Interest Margin (NIM)': {'base': 4.8, 'min': 2.8, 'max': 6.5, 'noise': 0.25, 'trend': -0.02},
    'Return on Assets (ROA)': {'base': 2.1, 'min': 0.8, 'max': 3.5, 'noise': 0.20, 'trend': 0.01},
    'Return on Equity (ROE)': {'base': 14.2, 'min': 6.0, 'max': 22.0, 'noise': 1.1, 'trend': 0.04},
    'Cost of Funds (CoF)': {'base': 3.8, 'min': 2.0, 'max': 6.2, 'noise': 0.30, 'trend': 0.03},

    # ── 3. Anomali Transaksi Cabang (Continuous Auditing / CAATT) ──
    'Rasio Anomali Transaksi Cabang Surabaya': {'base': 2.2, 'min': 0.3, 'max': 6.5, 'noise': 0.40, 'trend': 0.04},
    'Rasio Anomali Transaksi Cabang Jakarta': {'base': 1.6, 'min': 0.2, 'max': 5.5, 'noise': 0.35, 'trend': 0.02},
    'Rasio Anomali Transaksi Cabang Bandung': {'base': 1.9, 'min': 0.3, 'max': 5.8, 'noise': 0.38, 'trend': 0.03},
    'Rasio Anomali Transaksi Cabang Medan': {'base': 2.4, 'min': 0.4, 'max': 7.0, 'noise': 0.45, 'trend': 0.05},
    'Rasio Anomali Transaksi Cabang Bali': {'base': 1.5, 'min': 0.2, 'max': 4.8, 'noise': 0.28, 'trend': -0.01},

    # ── 4. Likuiditas, Solvabilitas & Pendanaan ──
    'Liquidity Coverage Ratio (LCR)': {'base': 146.0, 'min': 95.0, 'max': 210.0, 'noise': 6.5, 'trend': -0.2},
    'Net Stable Funding Ratio (NSFR)': {'base': 118.0, 'min': 95.0, 'max': 145.0, 'noise': 4.2, 'trend': 0.1},
    'Loan to Deposit Ratio (LDR)': {'base': 84.5, 'min': 70.0, 'max': 96.0, 'noise': 2.8, 'trend': 0.1},
    'Capital Adequacy Ratio (CAR)': {'base': 22.5, 'min': 16.0, 'max': 28.0, 'noise': 1.2, 'trend': -0.05},
    'CASA Ratio (Dana Murah Tabungan & Giro)': {'base': 64.0, 'min': 48.0, 'max': 78.0, 'noise': 2.5, 'trend': 0.2},
    'Rasio Kredit Berisiko (LAR - Loan at Risk)': {'base': 9.2, 'min': 4.0, 'max': 18.0, 'noise': 1.1, 'trend': 0.06},

    # ── 5. Kepatuhan Transaksi & Deviasi Fraud ──
    'Rasio Pelanggaran Plafon Kredit': {'base': 1.4, 'min': 0.1, 'max': 4.5, 'noise': 0.30, 'trend': 0.02},
    'Rasio Pelanggaran BMPK / BMPD': {'base': 0.5, 'min': 0.0, 'max': 2.5, 'noise': 0.15, 'trend': 0.01},
    'Rasio Deviasi Profil KYC / High Risk EDD': {'base': 3.2, 'min': 0.5, 'max': 8.5, 'noise': 0.55, 'trend': 0.04},
    'Frekuensi Penarikan Tunai Di Luar Limit': {'base': 12.0, 'min': 1.0, 'max': 35.0, 'noise': 3.2, 'trend': 0.15},
    'Rasio Split-Bill Pembayaran Abnormal': {'base': 1.8, 'min': 0.2, 'max': 5.2, 'noise': 0.35, 'trend': 0.02},

    # ── 6. Bisnis, Layanan, SDM & IT ──
    'Net Promoter Score (NPS)': {'base': 62.0, 'min': 25.0, 'max': 88.0, 'noise': 4.2, 'trend': 0.15},
    'Customer Retention Rate': {'base': 88.0, 'min': 70.0, 'max': 98.0, 'noise': 2.2, 'trend': 0.08},
    'Loan Growth Rate': {'base': 10.2, 'min': 2.0, 'max': 18.5, 'noise': 1.5, 'trend': 0.05},
    'Employee Turnover Rate': {'base': 13.5, 'min': 4.0, 'max': 32.0, 'noise': 1.8, 'trend': 0.04},
    'IT Core Banking Transaction Latency': {'base': 180.0, 'min': 50.0, 'max': 600.0, 'noise': 25.0, 'trend': 1.2},
    'Mobile Banking Failure Rate': {'base': 0.8, 'min': 0.1, 'max': 3.5, 'noise': 0.22, 'trend': 0.02},
    'Cyber Security Incident Resolution Rate': {'base': 94.0, 'min': 75.0, 'max': 100.0, 'noise': 2.5, 'trend': 0.08}
}

rows = []
kpi_id_counter = 1
months = pd.date_range(start='2021-01-01', periods=PERIODS_PER_KPI, freq='MS')

for kpi_name, spec in KPI_SPECS.items():
    t = np.arange(PERIODS_PER_KPI)
    trend_comp = t * spec['trend']
    seasonal = np.sin(2 * np.pi * t / 12) * (spec['noise'] * 0.8)
    noise = np.random.normal(0, spec['noise'], size=PERIODS_PER_KPI)
    
    values = spec['base'] + trend_comp + seasonal + noise
    values = np.round(np.clip(values, spec['min'], spec['max']), 1)
    
    for i in range(PERIODS_PER_KPI):
        period_str = months[i].strftime('Bulan %m %Y')
        val = values[i]
        recent_trend = (val - values[max(0, i-3)]) if i >= 3 else 0.0
        recent_vol = np.std(values[max(0, i-6):i+1]) if i >= 1 else 1.0

        # Calculate impact and likelihood scores (1-5) strictly for Auditee
        if 'NPL' in kpi_name or 'LAR' in kpi_name or 'Anomali' in kpi_name:
            imp = int(np.clip(round((val / spec['base']) * 2.8), 1, 5))
            lik = int(np.clip(round(1.5 + recent_vol * 1.5 + (1 if recent_trend > 0 else 0)), 1, 5))
        elif 'BOPO' in kpi_name or 'Latency' in kpi_name or 'Turnover' in kpi_name:
            imp = int(np.clip(round((val / spec['base']) * 3.0), 1, 5))
            lik = int(np.clip(round(2.0 + recent_vol * 0.8), 1, 5))
        else:
            # Positive indicator: lower value means higher risk impact
            imp = int(np.clip(round((spec['base'] / max(val, 0.1)) * 2.5), 1, 5))
            lik = int(np.clip(round(1.8 + recent_vol * 0.9), 1, 5))

        rows.append({
            'ID KPI': f'KPI-{kpi_id_counter:05d}',
            'Nama KPI': kpi_name,
            'Periode': period_str,
            'TARGET: Nilai Aktual (%)': val,
            'TARGET: Impact (1-5)': imp,
            'TARGET: Likelihood (1-5)': lik
        })
        kpi_id_counter += 1

df = pd.DataFrame(rows)
output_path = os.path.join(CURRENT_DIR, 'kpi_data.csv')
df.to_csv(output_path, index=False)

print(f"[SUCCESS] Generated kpi_data.csv with {len(df)} rows across {len(KPI_SPECS)} Auditee Business KPIs!")
print("Unique KPIs count:", len(df['Nama KPI'].unique()))
