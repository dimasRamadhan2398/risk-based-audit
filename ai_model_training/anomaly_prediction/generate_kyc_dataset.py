import os
import numpy as np
import pandas as pd

def generate():
    """
    Generates realistic, audit-ready, high-quality Banking KYC Anomaly Dataset
    compliant with OJK APU-PPT (POJK No. 12/POJK.01/2017 & POJK No. 23/POJK.01/2019)
    and FATF CDD/EDD Standards.
    """
    np.random.seed(104)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1200

    ENTITIES = [
        'Head Office', 'Jakarta Branch', 'Surabaya Branch', 
        'Bandung Branch', 'Medan Branch', 'Customer Onboarding Unit'
    ]
    ACTIVITIES = [
        'Pembukaan Rekening Perorangan', 'Pembukaan Rekening Korporasi',
        'Update Data Nasabah', 'Pengkinian Data Berkala'
    ]
    CUSTOMER_TYPES = ['Perorangan WNI', 'Korporasi / Badan Usaha', 'WNA / Non-Resident']
    RISK_PROFILES = ['Low', 'Medium', 'High']

    records = []

    for i in range(N_SAMPLES):
        log_id = f"KYC-{i+1:04d}"
        entity = np.random.choice(ENTITIES)
        activity = np.random.choice(ACTIVITIES)
        
        # Link customer type logically to activity
        if 'Korporasi' in activity:
            cust_type = 'Korporasi / Badan Usaha'
        elif np.random.rand() < 0.15:
            cust_type = 'WNA / Non-Resident'
        else:
            cust_type = 'Perorangan WNI'

        is_anom_scenario = np.random.rand() < 0.125
        scenario_type = None

        if is_anom_scenario:
            scenario_type = np.random.choice([
                'dttot_sanctions_match',
                'dukcapil_biometric_mismatch',
                'pep_high_risk_no_edd',
                'incomplete_expired_docs',
                'corporate_undisclosed_bo',
                'unauthorized_senior_override'
            ], p=[0.15, 0.20, 0.20, 0.20, 0.15, 0.10])

        alasan = []
        is_anom = False

        if scenario_type == 'dttot_sanctions_match':
            # Scenario 1: Approved despite DTTOT / Sanctions List Match (Severe AML/CFT Violation)
            risk_profile = 'High'
            aml_score = int(np.random.randint(85, 99))
            pep = np.random.choice(['PEP Asing', 'Bukan PEP'])
            sanctions = 'Match Terindikasi'
            dukcapil = 'Match Terverifikasi'
            bo_status = 'Terverifikasi Sah' if cust_type == 'Korporasi / Badan Usaha' else 'Tidak Berlaku'
            doc_status = 'Lengkap & Valid'
            edd_status = 'Dilakukan & Memadai'
            approval = 'Approved'
            senior_auth = 'Tanpa Otorisasi Senior'

            is_anom = True
            alasan.append("Pelanggaran Berat APU-PPT: Nasabah Terindikasi DTTOT / Daftar Sanksi Tetap Disetujui")
            impact = 5
            likelihood = 5

        elif scenario_type == 'dukcapil_biometric_mismatch':
            # Scenario 2: Identity Fraud / Forged Identity Approved
            risk_profile = np.random.choice(['Medium', 'High'])
            aml_score = int(np.random.randint(60, 90))
            pep = 'Bukan PEP'
            sanctions = 'Clear / Bersih'
            dukcapil = np.random.choice(['Mismatch Data', 'Gagal Biometrik'])
            bo_status = 'Tidak Berlaku' if cust_type != 'Korporasi / Badan Usaha' else 'Terverifikasi Sah'
            doc_status = np.random.choice(['Lengkap & Valid', 'Tidak Lengkap'])
            edd_status = 'Tidak Diperlukan'
            approval = 'Approved'
            senior_auth = 'Tanpa Otorisasi Senior'

            is_anom = True
            alasan.append(f"Indikasi Pemalsuan Identitas: Disetujui Padahal Verifikasi Kependudukan / Biometrik {dukcapil}")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'pep_high_risk_no_edd':
            # Scenario 3: High-Risk / PEP Customer Approved without Mandatory EDD
            risk_profile = 'High'
            aml_score = int(np.random.randint(70, 95))
            pep = np.random.choice(['PEP Domestik', 'PEP Asing'])
            sanctions = 'Clear / Bersih'
            dukcapil = 'Match Terverifikasi'
            bo_status = 'Terverifikasi Sah' if cust_type == 'Korporasi / Badan Usaha' else 'Tidak Berlaku'
            doc_status = 'Lengkap & Valid'
            edd_status = 'Tanpa EDD / Dilewati'
            approval = 'Approved'
            senior_auth = 'Tanpa Otorisasi Senior'

            is_anom = True
            alasan.append("Pelanggaran Regulasi: Nasabah PEP / High Risk Disetujui Tanpa Pelaksanaan EDD yang Memadai")
            impact = int(np.random.randint(4, 6))
            likelihood = 4

        elif scenario_type == 'incomplete_expired_docs':
            # Scenario 4: Approval with Incomplete or Expired KYC Documents
            risk_profile = np.random.choice(['Low', 'Medium'])
            aml_score = int(np.random.randint(25, 60))
            pep = 'Bukan PEP'
            sanctions = 'Clear / Bersih'
            dukcapil = 'Match Terverifikasi'
            bo_status = 'Terverifikasi Sah' if cust_type == 'Korporasi / Badan Usaha' else 'Tidak Berlaku'
            doc_status = np.random.choice(['Tidak Lengkap', 'Dokumen Kadaluarsa'])
            edd_status = 'Tidak Diperlukan'
            approval = 'Approved'
            senior_auth = 'Tanpa Otorisasi Senior'

            is_anom = True
            alasan.append(f"Pelanggaran Prosedur: Aktivitas Disetujui Padahal Status Dokumen {doc_status}")
            impact = 4
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'corporate_undisclosed_bo':
            # Scenario 5: Corporate Customer Onboarding without Beneficial Owner (Shell Company Risk)
            cust_type = 'Korporasi / Badan Usaha'
            activity = 'Pembukaan Rekening Korporasi'
            risk_profile = 'High'
            aml_score = int(np.random.randint(75, 95))
            pep = 'Bukan PEP'
            sanctions = 'Clear / Bersih'
            dukcapil = 'Match Terverifikasi'
            bo_status = 'Tidak Diungkap / Fiktif'
            doc_status = 'Lengkap & Valid'
            edd_status = 'Tanpa EDD / Dilewati'
            approval = 'Approved'
            senior_auth = 'Tanpa Otorisasi Senior'

            is_anom = True
            alasan.append("Risiko Shell Company: Pembukaan Rekening Korporasi Disetujui Tanpa Pengungkapan Beneficial Owner (BO)")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'unauthorized_senior_override':
            # Scenario 6: High Risk Onboarding Approved without Senior AML / Compliance Authorization
            risk_profile = 'High'
            aml_score = int(np.random.randint(68, 85))
            pep = np.random.choice(['Bukan PEP', 'PEP Domestik'])
            sanctions = 'Clear / Bersih'
            dukcapil = 'Match Terverifikasi'
            bo_status = 'Terverifikasi Sah' if cust_type == 'Korporasi / Badan Usaha' else 'Tidak Berlaku'
            doc_status = 'Lengkap & Valid'
            edd_status = 'Dilakukan & Memadai'
            approval = 'Approved'
            senior_auth = 'Tanpa Otorisasi Senior'

            is_anom = True
            alasan.append("Pelanggaran Wewenang: Nasabah High Risk Disetujui Tanpa Otorisasi Pejabat Senior AML")
            impact = 4
            likelihood = int(np.random.randint(3, 5))

        else:
            # NORMAL COMPLIANT ONBOARDING (Well-governed banking audit trail)
            risk_profile = np.random.choice(RISK_PROFILES, p=[0.60, 0.30, 0.10])
            pep = 'Bukan PEP' if risk_profile != 'High' else np.random.choice(['Bukan PEP', 'PEP Domestik'], p=[0.7, 0.3])
            sanctions = 'Clear / Bersih'
            dukcapil = 'Match Terverifikasi'
            
            if cust_type == 'Korporasi / Badan Usaha':
                bo_status = 'Terverifikasi Sah'
            else:
                bo_status = 'Tidak Berlaku'

            if risk_profile == 'Low':
                aml_score = int(np.random.randint(10, 36))
                edd_status = 'Tidak Diperlukan'
                doc_status = np.random.choice(['Lengkap & Valid', 'Tidak Lengkap'], p=[0.90, 0.10])
                if doc_status == 'Tidak Lengkap':
                    approval = np.random.choice(['Rejected', 'Pending Review'], p=[0.6, 0.4])
                else:
                    approval = np.random.choice(['Approved', 'Pending Review'], p=[0.85, 0.15])
                senior_auth = 'Tidak Berlaku'
            elif risk_profile == 'Medium':
                aml_score = int(np.random.randint(36, 66))
                edd_status = 'Tidak Diperlukan'
                doc_status = np.random.choice(['Lengkap & Valid', 'Tidak Lengkap'], p=[0.88, 0.12])
                if doc_status == 'Tidak Lengkap':
                    approval = np.random.choice(['Rejected', 'Pending Review'], p=[0.7, 0.3])
                else:
                    approval = np.random.choice(['Approved', 'Pending Review'], p=[0.80, 0.20])
                senior_auth = 'Tidak Berlaku'
            else: # High Risk Normal Compliant
                aml_score = int(np.random.randint(66, 85))
                doc_status = 'Lengkap & Valid'
                edd_status = 'Dilakukan & Memadai'
                senior_auth = 'Disetujui Pejabat Berwenang'
                approval = np.random.choice(['Approved', 'Pending Review', 'Rejected'], p=[0.5, 0.3, 0.2])

            is_anom = False
            impact = int(np.random.choice([1, 2], p=[0.65, 0.35]))
            likelihood = int(np.random.choice([1, 2], p=[0.70, 0.30]))

        target_is_anom = 'Ya (Anomali)' if is_anom else 'Tidak'
        target_reason = " & ".join(alasan) if is_anom else '-'

        records.append({
            'ID Aktivitas': log_id,
            'Entitas': entity,
            'Aktivitas': activity,
            'Tipe Nasabah': cust_type,
            'Profil Risiko': risk_profile,
            'Skor Risiko AML (1-100)': aml_score,
            'Status PEP': pep,
            'Skrining Sanksi & DTTOT': sanctions,
            'Verifikasi Dukcapil & Biometrik': dukcapil,
            'Status Beneficial Owner (BO)': bo_status,
            'Status Dokumen': doc_status,
            'Pelaksanaan EDD': edd_status,
            'Status Approval': approval,
            'Persetujuan Pejabat Senior AML': senior_auth,
            'TARGET: is_anomaly': target_is_anom,
            'TARGET: Impact (1-5)': impact,
            'TARGET: Likelihood (1-5)': likelihood,
            'Alasan Anomali': target_reason
        })

    df = pd.DataFrame(records)

    # Sanity deduplication on features to prevent accidental identical samples
    feature_cols = [c for c in df.columns if c not in ['ID Aktivitas', 'TARGET: is_anomaly', 'TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)', 'Alasan Anomali']]
    df = df.drop_duplicates(subset=feature_cols).reset_index(drop=True)

    # Re-index ID Aktivitas
    df['ID Aktivitas'] = [f"KYC-{i+1:04d}" for i in range(len(df))]

    output_path = os.path.join(CURRENT_DIR, 'kyc_data.csv')
    df.to_csv(output_path, index=False)

    anom_count = (df['TARGET: is_anomaly'] == 'Ya (Anomali)').sum()
    print(f"[SUCCESS] Generated {len(df)} audit-ready KYC records: {output_path}")
    print(f"Total Normal: {len(df) - anom_count} | Total Anomaly: {anom_count} ({anom_count/len(df)*100:.2f}%)")

if __name__ == "__main__":
    generate()
