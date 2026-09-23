import os
import numpy as np
import pandas as pd

np.random.seed(42)

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
PERIODS_PER_KPI = 132  # 132 monthly periods (11 years of time-series data per KPI)

KPI_SPECS = {
    'Net Promoter Score': {
        'base': 58.0, 'min_val': 15.0, 'max_val': 85.0, 'noise': 4.5,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((65.0 - v) / 12.0) + (1 if tr < -2 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 0.8 + (1 if v < 40 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Customer Lifetime Value': {
        'base': 680.0, 'min_val': 200.0, 'max_val': 1250.0, 'noise': 35.0,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((750.0 - v) / 110.0) + (1 if tr < -30 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 0.08 + (1 if v < 450 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Sales Conversion Rate': {
        'base': 8.5, 'min_val': 1.8, 'max_val': 18.5, 'noise': 1.2,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((11.0 - v) / 2.2) + (1 if tr < -1.0 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 2.2 + (1 if v < 5.0 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Customer Acquisition Cost': {
        'base': 42.0, 'min_val': 15.0, 'max_val': 95.0, 'noise': 3.8,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((v - 35.0) / 12.0) + (1 if tr > 5.0 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 0.9 + (1 if v > 65.0 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Gross Profit Margin': {
        'base': 34.0, 'min_val': 14.0, 'max_val': 48.0, 'noise': 2.2,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((38.0 - v) / 5.2) + (1 if tr < -2.0 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 1.4 + (1 if v < 25.0 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Customer Retention Rate': {
        'base': 86.0, 'min_val': 68.0, 'max_val': 98.0, 'noise': 2.5,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((90.0 - v) / 5.0) + (1 if tr < -2.0 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 1.2 + (1 if v < 78.0 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Average Order Value': {
        'base': 135.0, 'min_val': 40.0, 'max_val': 260.0, 'noise': 10.0,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((160.0 - v) / 25.0) + (1 if tr < -10 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 0.35 + (1 if v < 90 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Return on Investment': {
        'base': 25.0, 'min_val': 6.0, 'max_val': 46.0, 'noise': 2.8,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((30.0 - v) / 5.5) + (1 if tr < -3.0 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 1.3 + (1 if v < 14.0 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Monthly Recurring Revenue': {
        'base': 180.0, 'min_val': 45.0, 'max_val': 350.0, 'noise': 12.0,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((220.0 - v) / 38.0) + (1 if tr < -15 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 0.28 + (1 if v < 110 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    },
    'Employee Turnover Rate': {
        'base': 14.0, 'min_val': 4.0, 'max_val': 34.0, 'noise': 1.8,
        'imp_fn': lambda v, tr: int(np.clip(np.ceil((v - 10.0) / 4.8) + (1 if tr > 2.0 else 0) + np.random.choice([0, 1], p=[0.7, 0.3]), 1, 5)),
        'lik_fn': lambda v, vol: int(np.clip(np.ceil(vol * 1.5 + (1 if v > 22.0 else 0) + np.random.choice([-1, 0, 1], p=[0.2, 0.6, 0.2])), 1, 5))
    }
}

rows = []
kpi_id_counter = 1
months = pd.date_range(start='2015-01-01', periods=PERIODS_PER_KPI, freq='MS')

for kpi_name, spec in KPI_SPECS.items():
    t = np.arange(PERIODS_PER_KPI)
    trend_component = np.sin(2 * np.pi * t / 12) * (spec['noise'] * 1.5) + (t * 0.04 * (1 if np.random.rand() > 0.4 else -1))
    noise = np.random.normal(0, spec['noise'], size=PERIODS_PER_KPI)
    
    values = spec['base'] + trend_component + noise
    values = np.round(np.clip(values, spec['min_val'], spec['max_val']), 1)
    
    for i in range(PERIODS_PER_KPI):
        period_str = months[i].strftime('Bulan %m %Y')
        val = values[i]
        recent_trend = (val - values[max(0, i-3)]) if i >= 3 else 0.0
        recent_vol = np.std(values[max(0, i-6):i+1]) if i >= 1 else 1.0
        
        impact = spec['imp_fn'](val, recent_trend)
        likelihood = spec['lik_fn'](val, recent_vol)
        
        rows.append({
            'ID KPI': f'KPI-{kpi_id_counter:04d}',
            'Nama KPI': kpi_name,
            'Periode': period_str,
            'TARGET: Nilai Aktual (%)': val,
            'TARGET: Impact (1-5)': impact,
            'TARGET: Likelihood (1-5)': likelihood
        })
        kpi_id_counter += 1

df_kpi = pd.DataFrame(rows)
df_kpi.to_csv(os.path.join(CURRENT_DIR, 'kpi_data.csv'), index=False)

print(f"[SUCCESS] Generated realistic kpi_data.csv with {len(df_kpi)} rows across {len(KPI_SPECS)} unique KPIs!")
print("\n--- KPI Summary Table ---")
for kpi_name, grp in df_kpi.groupby('Nama KPI'):
    vals = grp['TARGET: Nilai Aktual (%)']
    print(f"{kpi_name:<30} | Min: {vals.min():<6} | Max: {vals.max():<6} | Mean: {vals.mean():<6.1f} | Std: {vals.std():<5.1f}")

print("\n--- Target Impact (1-5) Value Counts ---")
print(df_kpi['TARGET: Impact (1-5)'].value_counts().sort_index())

print("\n--- Target Likelihood (1-5) Value Counts ---")
print(df_kpi['TARGET: Likelihood (1-5)'].value_counts().sort_index())

eq = (df_kpi['TARGET: Impact (1-5)'] == df_kpi['TARGET: Likelihood (1-5)']).mean()
print(f"\nImpact vs Likelihood Target Equality: {eq*100:.2f}%")
