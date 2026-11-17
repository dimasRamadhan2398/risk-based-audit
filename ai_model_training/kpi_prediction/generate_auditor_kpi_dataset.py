import os
import json
import numpy as np
import pandas as pd

np.random.seed(42)

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
SCRATCH_JSON = os.path.join(CURRENT_DIR, '..', '..', 'scratch', 'extracted_strategic_plans.json')
ALT_JSON = 'C:/Users/Dimas/.gemini/antigravity/brain/cc489e7b-523d-4160-a61c-425bbdd2ef9a/scratch/extracted_strategic_plans.json'
PERIODS_PER_KPI = 60  # 60 monthly periods (5 years of historical time-series per KPI for 100 KPIs)

def get_pillar_for_code(code_str: str) -> str:
    try:
        num = int(code_str.replace('SO-IA', ''))
        if num <= 10: return "Tata Kelola Perusahaan & GCG"
        elif num <= 20: return "Manajemen Risiko & Pengendalian Internal"
        elif num <= 30: return "Pelaksanaan Audit & PKAT"
        elif num <= 40: return "Kualitas LHA & Pelaporan"
        elif num <= 50: return "Monitoring Tindak Lanjut Temuan"
        elif num <= 60: return "Investigasi Khusus & Fraud"
        elif num <= 70: return "Teknologi & Continuous Auditing"
        elif num <= 80: return "Pengembangan SDM & Sertifikasi"
        elif num <= 90: return "Efisiensi & Produktivitas Audit"
        else: return "Kepuasan Stakeholder & Komite Audit"
    except Exception:
        return "Pelaksanaan Audit Internal"

def calculate_achievement_and_status(target: float, actual: float, polarity: str):
    target = float(target)
    actual = float(actual)
    if polarity == 'HIG':
        ratio = (actual / target) * 100.0 if target > 0 else 100.0
    else:
        ratio = 100.0 if actual <= 0 else (target / actual) * 100.0
        
    ratio = round(ratio, 2)
    if ratio >= 100.0:
        status = 'Good'
    elif ratio >= 70.0:
        status = 'Moderate'
    else:
        status = 'Poor'
        
    return ratio, status

def generate_auditor_dataset():
    # Load 100 Strategic Audit Plans
    json_path = SCRATCH_JSON if os.path.exists(SCRATCH_JSON) else (ALT_JSON if os.path.exists(ALT_JSON) else None)
    plans = []
    if json_path and os.path.exists(json_path):
        with open(json_path, 'r', encoding='utf-8') as f:
            plans = json.load(f)
        print(f"Loaded {len(plans)} plans from {json_path}")
    
    if not plans:
        print("Warning: JSON not found, generating standard 100 SO-IA plans...")
        for i in range(1, 101):
            code = f"SO-IA{i:02d}"
            plans.append({
                'code': code,
                'kpi': f'Indikator Sasaran Strategis {code}',
                'unit': '%' if i % 3 != 0 else ('Score' if i % 2 == 0 else 'Hari'),
                'hibHig': 'HIB' if i in [5, 8, 18, 25, 31, 42, 55, 68, 75, 88] else 'HIG',
                'target': '100.0' if i % 3 == 0 else ('85.0' if i % 2 == 0 else '15.0'),
                'actual': '92.0' if i % 3 == 0 else ('82.5' if i % 2 == 0 else '16.5'),
                'strategicObjective': f'Pilar Strategis {code}'
            })

    rows = []
    kpi_id_counter = 1
    months = pd.date_range(start='2021-01-01', periods=PERIODS_PER_KPI, freq='MS')

    for p in plans:
        code = p.get('code', f'SO-IA{kpi_id_counter:02d}')
        kpi_name = p.get('kpi', 'Indikator Audit')
        pillar = get_pillar_for_code(code)
        pol = p.get('hibHig', 'HIG').upper()
        unit = p.get('unit', '%')
        
        try:
            target_val = float(p.get('target', 100.0) or 100.0)
        except Exception:
            target_val = 100.0
            
        try:
            actual_base = float(p.get('actual', target_val) or target_val)
        except Exception:
            actual_base = target_val

        noise_std = max(0.5, actual_base * 0.04)
        t = np.arange(PERIODS_PER_KPI)
        seasonal = np.sin(2 * np.pi * t / 12) * (noise_std * 0.6)
        trend = (t - PERIODS_PER_KPI // 2) * (noise_std * 0.02) * (1 if pol == 'HIG' else -1)
        noise = np.random.normal(0, noise_std, size=PERIODS_PER_KPI)

        raw_values = actual_base + seasonal + trend + noise
        
        # Enforce reasonable positive bounds
        if pol == 'HIG':
            values = np.clip(raw_values, actual_base * 0.6, actual_base * 1.3)
        else:
            values = np.clip(raw_values, max(0.1, actual_base * 0.5), actual_base * 1.6)

        values = np.round(values, 1)

        for i in range(PERIODS_PER_KPI):
            period_str = months[i].strftime('Bulan %m %Y')
            val = values[i]
            ratio, status = calculate_achievement_and_status(target_val, val, pol)

            rows.append({
                'ID KPI': f'AUD-KPI-{kpi_id_counter:05d}',
                'Kode SO': code,
                'Nama KPI': kpi_name,
                'Kategori Pilar': pillar,
                'Periode': period_str,
                'Tipe Polaritas (HIG/HIB)': pol,
                'Satuan': unit,
                'Nilai Target': target_val,
                'TARGET: Nilai Aktual': val,
                'TARGET: Rasio Ketercapaian (%)': ratio,
                'TARGET: Status Achievement': status
            })
            kpi_id_counter += 1

    df_auditor = pd.DataFrame(rows)
    output_path = os.path.join(CURRENT_DIR, 'auditor_kpi_data.csv')
    df_auditor.to_csv(output_path, index=False)

    print(f"[SUCCESS] Generated auditor_kpi_data.csv with {len(df_auditor)} rows across {len(plans)} Strategic Audit Plan KPIs!")
    print("\n--- Distribution of TARGET: Status Achievement ---")
    print(df_auditor['TARGET: Status Achievement'].value_counts())
    print("\n--- Unique SO Codes Count ---")
    print(len(df_auditor['Kode SO'].unique()))

if __name__ == '__main__':
    generate_auditor_dataset()
