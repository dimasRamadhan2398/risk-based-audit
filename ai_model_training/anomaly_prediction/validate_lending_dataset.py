import os
import sys
import pandas as pd
import numpy as np

def run_qa_audit(csv_path: str) -> bool:
    print("=" * 80)
    print(" AUTOMATED QUALITY ASSURANCE & AUDIT VERIFICATION ")
    print(f" File: {csv_path}")
    print("=" * 80)

    if not os.path.exists(csv_path):
        print(f"[FAIL] Dataset file not found: {csv_path}")
        return False

    df = pd.read_csv(csv_path)
    print(f"Total Records: {len(df)} | Total Columns: {len(df.columns)}")
    print(f"Columns: {df.columns.tolist()}\n")

    all_passed = True

    # -------------------------------------------------------------------------
    # [CHECK 1] Scanning for Missing / NaN Values
    # -------------------------------------------------------------------------
    print("[CHECK 1] Scanning for Missing / NaN Values...")
    null_counts = df.isnull().sum()
    total_nulls = null_counts.sum()
    if total_nulls > 0:
        print(f"  --> [FAIL] Found {total_nulls} null/NaN values across columns:")
        print(null_counts[null_counts > 0])
        all_passed = False
    else:
        print("  --> [PASS] Zero NaN / Null values detected in all columns.")

    # -------------------------------------------------------------------------
    # [CHECK 2] Scanning for Label Contradictions (Ground Truth Integrity)
    # -------------------------------------------------------------------------
    print("\n[CHECK 2] Scanning for Label Contradictions (Ground Truth Integrity)...")
    feature_cols = [
        'Entitas', 'Produk', 'Nilai Plafon (Juta Rp)', 'Bunga (%)',
        'Nilai Taksasi Agunan (Juta Rp)', 'Rasio LTV (%)', 'Status Dokumen',
        'Kolektibilitas SLIK OJK (IDEB)', 'Otorisasi Batas Kewenangan (BKMK)',
        'Kepatuhan Batas BMPK OJK', 'Status Penutupan Asuransi Agunan',
        'Verifikasi Tujuan Penggunaan (Side-Streaming)'
    ]
    grouped = df.groupby(feature_cols)['TARGET: is_anomaly'].nunique()
    contradictions = grouped[grouped > 1]
    if len(contradictions) > 0:
        print(f"  --> [FAIL] Detected {len(contradictions)} contradictory feature rows!")
        all_passed = False
    else:
        print(f"  --> [PASS] Zero contradictory rows found across all {len(feature_cols)} feature dimensions.")

    # -------------------------------------------------------------------------
    # [CHECK 3] Validating Feature & Target Value Ranges
    # -------------------------------------------------------------------------
    print("\n[CHECK 3] Validating Feature & Target Value Ranges...")
    range_errors = 0

    if (df['Nilai Plafon (Juta Rp)'] <= 0).any():
        print("  --> [FAIL] Negative or zero loan ceiling found.")
        range_errors += 1
    if (df['Bunga (%)'] < 0).any() or (df['Bunga (%)'] > 35.0).any():
        print("  --> [FAIL] Interest rate outside realistic bounds (0% - 35%).")
        range_errors += 1
    if (df['Nilai Taksasi Agunan (Juta Rp)'] <= 0).any():
        print("  --> [FAIL] Non-positive collateral appraisal found.")
        range_errors += 1
    if (df['Rasio LTV (%)'] <= 0).any():
        print("  --> [FAIL] Non-positive LTV ratio found.")
        range_errors += 1
    if not set(df['TARGET: Impact (1-5)'].unique()).issubset({1, 2, 3, 4, 5}):
        print("  --> [FAIL] Target Impact values outside 1-5.")
        range_errors += 1
    if not set(df['TARGET: Likelihood (1-5)'].unique()).issubset({1, 2, 3, 4, 5}):
        print("  --> [FAIL] Target Likelihood values outside 1-5.")
        range_errors += 1
    if not set(df['TARGET: is_anomaly'].unique()).issubset({'Tidak', 'Ya (Anomali)'}):
        print("  --> [FAIL] Unexpected labels in TARGET: is_anomaly.")
        range_errors += 1

    if range_errors == 0:
        print("  --> [PASS] All values conform strictly to defined banking lending ranges.")
    else:
        all_passed = False

    # -------------------------------------------------------------------------
    # [CHECK 4] Checking Anomaly Explainability & Reason Consistency
    # -------------------------------------------------------------------------
    print("\n[CHECK 4] Checking Anomaly Explainability & Reason Consistency...")
    anomalies = df[df['TARGET: is_anomaly'] == 'Ya (Anomali)']
    normals = df[df['TARGET: is_anomaly'] == 'Tidak']

    unexplained_anom = anomalies[anomalies['Alasan Anomali'].isin(['-', '', np.nan])]
    invalid_normal = normals[normals['Alasan Anomali'] != '-']

    if len(unexplained_anom) > 0:
        print(f"  --> [FAIL] Found {len(unexplained_anom)} anomalies without explanation.")
        all_passed = False
    else:
        print(f"  --> [PASS] 100% of anomalies ({len(anomalies)} rows) have clear, explainable reasons.")

    if len(invalid_normal) > 0:
        print(f"  --> [FAIL] Found {len(invalid_normal)} normal rows with an anomaly reason.")
        all_passed = False
    else:
        print("  --> [PASS] 100% of normal records have clean '-' placeholder reasons.")

    # -------------------------------------------------------------------------
    # [CHECK 5] Checking Class Distribution & Proportions
    # -------------------------------------------------------------------------
    print("\n[CHECK 5] Checking Class Distribution & Proportions...")
    anom_rate = len(anomalies) / len(df)
    print(f"  Normal Records : {len(normals)} ({len(normals)/len(df)*100:.2f} %)")
    print(f"  Anomaly Records: {len(anomalies)} ({anom_rate*100:.2f} %)")

    if 0.08 <= anom_rate <= 0.20:
        print(f"  --> [PASS] Anomaly proportion ({anom_rate*100:.2f}%) is within optimal audit range (8% - 20%).")
    else:
        print(f"  --> [WARNING/FAIL] Anomaly proportion ({anom_rate*100:.2f}%) is outside optimal range.")
        all_passed = False

    # -------------------------------------------------------------------------
    # [CHECK 6] Verifying Banking Lending Domain Ground Truth Rules
    # -------------------------------------------------------------------------
    print("\n[CHECK 6] Verifying Banking Lending Domain Ground Truth Rules...")
    rule_fails = 0

    # Rule 6a: BKMK limit breach MUST be anomaly
    bkmk_breaches = df[df['Otorisasi Batas Kewenangan (BKMK)'] == 'Melampaui Limit Wewenang Cabang (BKMK Breach)']
    if not (bkmk_breaches['TARGET: is_anomaly'] == 'Ya (Anomali)').all():
        print("  --> [FAIL] BKMK limit breaches found labeled as normal!")
        rule_fails += 1
    else:
        print(f"  --> [PASS] 100% of BKMK limit breaches ({len(bkmk_breaches)} rows) flagged as Anomaly.")

    # Rule 6b: Under-collateralized loans (LTV > 100%) MUST be anomaly
    under_collat = df[df['Rasio LTV (%)'] > 100.0]
    if not (under_collat['TARGET: is_anomaly'] == 'Ya (Anomali)').all():
        print("  --> [FAIL] Under-collateralized loans (LTV > 100%) found labeled as normal!")
        rule_fails += 1
    else:
        print(f"  --> [PASS] 100% of under-collateralized loans ({len(under_collat)} rows) flagged as Anomaly.")

    # Rule 6c: Incomplete documentation MUST be anomaly
    doc_defect = df[df['Status Dokumen'] == 'Pencairan Sebelum Syarat Efektif Lengkap (Pelanggaran)']
    if not (doc_defect['TARGET: is_anomaly'] == 'Ya (Anomali)').all():
        print("  --> [FAIL] Loans with deficient documentation found labeled as normal!")
        rule_fails += 1
    else:
        print(f"  --> [PASS] 100% of deficient documentation loans ({len(doc_defect)} rows) flagged as Anomaly.")

    # Rule 6d: Normal loans MUST NOT have incomplete documentation (flaw resolution check)
    normal_doc_defects = normals[normals['Status Dokumen'] != 'Lengkap & Terikat Sempurna']
    if len(normal_doc_defects) > 0:
        print(f"  --> [FAIL] Found {len(normal_doc_defects)} normal loans with deficient documentation!")
        rule_fails += 1
    else:
        print("  --> [PASS] Exactly 0 normal loans have deficient documentation (100% Clean Ground Truth).")

    # Rule 6e: NPL SLIK approvals MUST be anomaly
    slik_npl = df[df['Kolektibilitas SLIK OJK (IDEB)'] == 'Kol 3-5 (NPL / Macet)']
    if not (slik_npl['TARGET: is_anomaly'] == 'Ya (Anomali)').all():
        print("  --> [FAIL] NPL SLIK approvals found labeled as normal!")
        rule_fails += 1
    else:
        print(f"  --> [PASS] 100% of NPL SLIK loans ({len(slik_npl)} rows) flagged as Anomaly.")

    # Rule 6f: BMPK rule breach MUST be anomaly
    bmpk_breaches = df[df['Kepatuhan Batas BMPK OJK'] == 'Pelanggaran BMPK Pihak Terafiliasi']
    if not (bmpk_breaches['TARGET: is_anomaly'] == 'Ya (Anomali)').all():
        print("  --> [FAIL] BMPK rule breaches found labeled as normal!")
        rule_fails += 1
    else:
        print(f"  --> [PASS] 100% of BMPK rule breaches ({len(bmpk_breaches)} rows) flagged as Anomaly.")

    if rule_fails > 0:
        all_passed = False

    # -------------------------------------------------------------------------
    # [CHECK 7] Verifying Inherent Risk Scoring (Impact vs Likelihood)
    # -------------------------------------------------------------------------
    print("\n[CHECK 7] Verifying Inherent Risk Scoring (Impact vs Likelihood)...")
    norm_impact_max = normals['TARGET: Impact (1-5)'].max()
    anom_impact_min = anomalies['TARGET: Impact (1-5)'].min()
    print(f"  Normal Impact range   : {normals['TARGET: Impact (1-5)'].min()} - {norm_impact_max}")
    print(f"  Anomaly Impact range  : {anom_impact_min} - {anomalies['TARGET: Impact (1-5)'].max()}")
    print(f"  Normal Likelihood range : {normals['TARGET: Likelihood (1-5)'].min()} - {normals['TARGET: Likelihood (1-5)'].max()}")
    print(f"  Anomaly Likelihood range: {anomalies['TARGET: Likelihood (1-5)'].min()} - {anomalies['TARGET: Likelihood (1-5)'].max()}")

    if norm_impact_max <= 2 and anom_impact_min >= 4:
        print("  --> [PASS] Risk severity scaling strictly reflects standard 5x5 RBA matrix.")
    else:
        print("  --> [WARNING/FAIL] Inconsistent risk scoring between normal and anomaly classes.")
        all_passed = False

    # -------------------------------------------------------------------------
    # FINAL CERTIFICATE
    # -------------------------------------------------------------------------
    print("\n" + "=" * 80)
    if all_passed:
        print(" [AUDIT-READY CERTIFICATE] ALL 7 DATA QUALITY & INTEGRITY CHECKS PASSED!")
        print(" The Lending dataset is ROBUST, TRUSTWORTHY, HIGH-QUALITY, and RELIABLE.")
        print(" Ready for AI Model Training and Benchmarking.")
    else:
        print(" [AUDIT FAILED] One or more data quality checks did not pass.")
    print("=" * 80 + "\n")

    return all_passed

if __name__ == "__main__":
    current_dir = os.path.dirname(os.path.abspath(__file__))
    target_csv = os.path.join(current_dir, "lending_data.csv")
    success = run_qa_audit(target_csv)
    sys.exit(0 if success else 1)
