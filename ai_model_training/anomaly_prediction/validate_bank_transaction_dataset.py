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
        'Kategori', 'Entitas', 'Kanal Transaksi', 'Nilai Transaksi (Juta Rp)',
        'Bunga/Margin (%)', 'Jam Transaksi (0-23)', 'Status Dokumen',
        'Jumlah Gagal Login', 'Tingkat Risiko Nasabah',
        'Status Otorisasi / Maker-Checker', 'Deviasi terhadap Profil Historis (%)'
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

    if (df['Nilai Transaksi (Juta Rp)'] < 0).any():
        print("  --> [FAIL] Negative transaction amount found.")
        range_errors += 1
    if (df['Bunga/Margin (%)'] < 0).any() or (df['Bunga/Margin (%)'] > 35.0).any():
        print("  --> [FAIL] Interest/margin outside realistic bounds (0% - 35%).")
        range_errors += 1
    if (df['Jam Transaksi (0-23)'] < 0).any() or (df['Jam Transaksi (0-23)'] > 23).any():
        print("  --> [FAIL] Transaction hour outside 0-23.")
        range_errors += 1
    if (df['Jumlah Gagal Login'] < 0).any():
        print("  --> [FAIL] Negative failed login count.")
        range_errors += 1
    if (df['Deviasi terhadap Profil Historis (%)'] < 0).any():
        print("  --> [FAIL] Negative historical deviation.")
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
        print("  --> [PASS] All values conform strictly to defined bank transaction ranges.")
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
    # [CHECK 5] Checking Class Distribution & Category Balance
    # -------------------------------------------------------------------------
    print("\n[CHECK 5] Checking Class Distribution & Category Balance...")
    anom_rate = len(anomalies) / len(df)
    print(f"  Normal Records : {len(normals)} ({len(normals)/len(df)*100:.2f} %)")
    print(f"  Anomaly Records: {len(anomalies)} ({anom_rate*100:.2f} %)")

    cat_counts = df['Kategori'].value_counts()
    print("\n  Records per Banking Category:")
    for c_name, c_cnt in cat_counts.items():
        c_anom = len(df[(df['Kategori'] == c_name) & (df['TARGET: is_anomaly'] == 'Ya (Anomali)')])
        print(f"    - {c_name:<12}: {c_cnt} records (Anomalies: {c_anom} = {c_anom/c_cnt*100:.1f}%)")

    cat_balanced = (cat_counts == 200).all()
    if cat_balanced and 0.08 <= anom_rate <= 0.20:
        print(f"\n  --> [PASS] Perfect category balance (200 records each) with optimal anomaly rate ({anom_rate*100:.2f}%).")
    else:
        print(f"\n  --> [WARNING/FAIL] Category distribution unbalanced or anomaly rate outside bounds.")
        all_passed = False

    # -------------------------------------------------------------------------
    # [CHECK 6] Verifying Cross-Category Banking Domain Ground Truth Rules
    # -------------------------------------------------------------------------
    print("\n[CHECK 6] Verifying Cross-Category Banking Domain Ground Truth Rules...")
    rule_fails = 0

    # Rule 6a: Incomplete documents MUST NOT exist on normal records
    norm_doc_defects = normals[normals['Status Dokumen'] != 'Lengkap']
    if len(norm_doc_defects) > 0:
        print(f"  --> [FAIL] Found {len(norm_doc_defects)} normal records with incomplete documents!")
        rule_fails += 1
    else:
        print("  --> [PASS] Exactly 0 normal transactions have incomplete documents (No false normal documents).")

    # Rule 6b: Normal Treasury MUST NOT occur outside business hours (07:00 - 18:00)
    norm_treasury_night = normals[(normals['Kategori'] == 'Treasury') & 
                                  ((normals['Jam Transaksi (0-23)'] < 7) | (normals['Jam Transaksi (0-23)'] > 18))]
    if len(norm_treasury_night) > 0:
        print(f"  --> [FAIL] Found {len(norm_treasury_night)} normal Treasury transactions outside business hours!")
        rule_fails += 1
    else:
        print("  --> [PASS] Exactly 0 normal Treasury transactions occur outside business hours (No false night trades).")

    # Rule 6c: Brute Force login attempts (>= 5) MUST be anomaly
    brute_force = df[df['Jumlah Gagal Login'] >= 5]
    if not (brute_force['TARGET: is_anomaly'] == 'Ya (Anomali)').all():
        print("  --> [FAIL] Brute force login attempts found labeled as normal!")
        rule_fails += 1
    else:
        print(f"  --> [PASS] 100% of Brute Force login attempts ({len(brute_force)} rows) flagged as Anomaly.")

    # Rule 6d: Override without authorization MUST be anomaly
    overrides = df[df['Status Otorisasi / Maker-Checker'] == 'Override / Tanpa Otorisasi']
    if not (overrides['TARGET: is_anomaly'] == 'Ya (Anomali)').all():
        print("  --> [FAIL] Unauthorized overrides found labeled as normal!")
        rule_fails += 1
    else:
        print(f"  --> [PASS] 100% of unauthorized overrides ({len(overrides)} rows) flagged as Anomaly.")

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
        print(" The Bank Transaction dataset is ROBUST, TRUSTWORTHY, HIGH-QUALITY, and RELIABLE.")
        print(" Ready for AI Model Training and Benchmarking.")
    else:
        print(" [AUDIT FAILED] One or more data quality checks did not pass.")
    print("=" * 80 + "\n")

    return all_passed

if __name__ == "__main__":
    current_dir = os.path.dirname(os.path.abspath(__file__))
    target_csv = os.path.join(current_dir, "bank_transaction_anomaly_data.csv")
    success = run_qa_audit(target_csv)
    sys.exit(0 if success else 1)
