import os
import numpy as np
import pandas as pd

def generate():
    """
    Generates realistic, audit-ready, high-quality IT Control Anomaly Dataset
    compliant with Banking IT Governance (OJK POJK No. 11/POJK.03/2022 & ISO 27001).
    """
    np.random.seed(105)
    CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
    N_SAMPLES = 1200

    # 1. Enterprise Banking Entities & Roles (aligned with translation_layer.py)
    ENTITIES = [
        'Head Office', 'Jakarta Branch', 'Surabaya Branch', 
        'Bandung Branch', 'Finance Dept', 'IT Dept', 'Operations Dept'
    ]
    USER_ROLES = [
        'Core Banking Operator', 'Database Administrator', 'IT Support', 
        'Staff Finance', 'Senior Auditor', 'System Admin', 'Ex-Employee'
    ]
    SYSTEMS = [
        'Core Banking', 'Database Server', 'Email Server', 'Payment Gateway / SWIFT'
    ]
    ACCESS_TYPES = ['Read', 'Write', 'Admin']
    MFA_STATUSES = ['Valid', 'Bypassed', 'Failed']
    NETWORK_LOCATIONS = ['Internal Secure LAN', 'Corporate VPN', 'Untrusted Public IP']
    CR_STATUSES = ['Ada Tiket CR', 'Tanpa Tiket CR']

    records = []

    for i in range(N_SAMPLES):
        log_id = f"IT-{i+1:04d}"
        
        # Determine whether this sample will be an injected anomaly (~12% target rate)
        # We define distinct, realistic banking cyber & IT audit threat scenarios.
        is_anom_scenario = np.random.rand() < 0.125
        scenario_type = None

        if is_anom_scenario:
            scenario_type = np.random.choice([
                'brute_force',
                'ex_employee_access',
                'data_exfiltration',
                'sod_violation_unauthorized_cr',
                'night_admin_untrusted_ip',
                'mfa_bypass_foreign_ip'
            ], p=[0.25, 0.15, 0.20, 0.15, 0.15, 0.10])

        alasan = []
        is_anom = False

        if scenario_type == 'brute_force':
            # Scenario 1: Password Spraying / Brute Force Attack
            entity = np.random.choice(ENTITIES)
            role = np.random.choice(['Core Banking Operator', 'Staff Finance', 'System Admin'])
            system = np.random.choice(['Core Banking', 'Database Server', 'Email Server'])
            jam = np.random.randint(0, 24)
            day = np.random.randint(1, 8)
            gagal_login = int(np.random.randint(5, 25))
            tipe_akses = np.random.choice(['Read', 'Write', 'Admin'])
            mfa = np.random.choice(['Failed', 'Bypassed'])
            network = np.random.choice(['Untrusted Public IP', 'Corporate VPN'])
            export_mb = round(float(np.random.exponential(scale=5.0)), 1)
            cr_status = 'Tanpa Tiket CR'
            
            is_anom = True
            alasan.append(f"Indikasi Serangan Brute Force (Gagal Login {gagal_login}x)")
            impact = int(np.random.randint(4, 6))
            likelihood = 5

        elif scenario_type == 'ex_employee_access':
            # Scenario 2: Terminated / Ex-Employee Account Not Revoked
            entity = np.random.choice(['Finance Dept', 'Operations Dept', 'Jakarta Branch'])
            role = 'Ex-Employee'
            system = np.random.choice(['Core Banking', 'Database Server', 'Payment Gateway / SWIFT'])
            jam = np.random.randint(8, 22)
            day = np.random.randint(1, 8)
            gagal_login = int(np.random.choice([0, 1, 2]))
            tipe_akses = np.random.choice(['Read', 'Write'])
            mfa = np.random.choice(['Valid', 'Bypassed'])
            network = np.random.choice(['Untrusted Public IP', 'Corporate VPN'])
            export_mb = round(float(np.random.uniform(50.0, 500.0)), 1)
            cr_status = 'Tanpa Tiket CR'

            is_anom = True
            alasan.append("Akses Tidak Sah Menggunakan Akun Mantan Pegawai (Ex-Employee)")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'data_exfiltration':
            # Scenario 3: Massive Bulk Data Dumping / Data Exfiltration Risk
            entity = np.random.choice(['IT Dept', 'Operations Dept', 'Finance Dept'])
            role = np.random.choice(['Database Administrator', 'Core Banking Operator', 'Staff Finance'])
            system = np.random.choice(['Core Banking', 'Database Server'])
            jam = int(np.random.choice([19, 20, 21, 22, 23, 0, 1, 2, 3]))
            day = np.random.randint(1, 8)
            gagal_login = int(np.random.choice([0, 1]))
            tipe_akses = np.random.choice(['Read', 'Write'])
            mfa = 'Valid'
            network = np.random.choice(['Untrusted Public IP', 'Corporate VPN'])
            export_mb = round(float(np.random.uniform(800.0, 4500.0)), 1)
            cr_status = 'Tanpa Tiket CR'

            is_anom = True
            alasan.append(f"Indikasi Eksfiltrasi Data Volume Ekspor Sangat Besar ({export_mb} MB)")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'sod_violation_unauthorized_cr':
            # Scenario 4: Segregation of Duties (SoD) & Unauthorized Production Modification
            entity = np.random.choice(['Finance Dept', 'Operations Dept'])
            role = np.random.choice(['Staff Finance', 'Core Banking Operator'])
            system = 'Database Server'
            jam = np.random.randint(9, 18)
            day = np.random.randint(1, 6)
            gagal_login = 0
            tipe_akses = np.random.choice(['Write', 'Admin'])
            mfa = 'Valid'
            network = 'Internal Secure LAN'
            export_mb = round(float(np.random.exponential(scale=10.0)), 1)
            cr_status = 'Tanpa Tiket CR'

            is_anom = True
            alasan.append("Pelanggaran SoD: Modifikasi Database Server Tanpa Tiket Perubahan (CR)")
            impact = int(np.random.randint(4, 6))
            likelihood = int(np.random.randint(3, 5))

        elif scenario_type == 'night_admin_untrusted_ip':
            # Scenario 5: Off-Hours Admin Access via Untrusted Network
            entity = np.random.choice(ENTITIES)
            role = np.random.choice(['System Admin', 'IT Support'])
            system = np.random.choice(['Core Banking', 'Payment Gateway / SWIFT'])
            jam = int(np.random.choice([0, 1, 2, 3, 4, 23]))
            day = int(np.random.choice([6, 7])) # Weekend night
            gagal_login = int(np.random.choice([1, 2]))
            tipe_akses = 'Admin'
            mfa = np.random.choice(['Valid', 'Bypassed'])
            network = 'Untrusted Public IP'
            export_mb = round(float(np.random.exponential(scale=20.0)), 1)
            cr_status = 'Tanpa Tiket CR'

            is_anom = True
            alasan.append(f"Akses Admin Sistem Kritis Dini Hari Akhir Pekan (Jam {jam}) dari IP Publik")
            impact = 5
            likelihood = int(np.random.randint(4, 6))

        elif scenario_type == 'mfa_bypass_foreign_ip':
            # Scenario 6: MFA Bypass & Suspicious Network Route
            entity = np.random.choice(ENTITIES)
            role = np.random.choice(['Core Banking Operator', 'System Admin'])
            system = np.random.choice(['Payment Gateway / SWIFT', 'Core Banking'])
            jam = np.random.randint(0, 24)
            day = np.random.randint(1, 8)
            gagal_login = int(np.random.choice([0, 1]))
            tipe_akses = np.random.choice(['Write', 'Admin'])
            mfa = 'Bypassed'
            network = 'Untrusted Public IP'
            export_mb = round(float(np.random.exponential(scale=15.0)), 1)
            cr_status = 'Tanpa Tiket CR'

            is_anom = True
            alasan.append("Bypass Autentikasi MFA Tidak Sah Melalui Jaringan Publik Tak Dikenal")
            impact = 5
            likelihood = 5

        else:
            # NORMAL BASELINE: Legitimate, well-governed banking IT operations
            role = np.random.choice([
                'Core Banking Operator', 'Database Administrator', 'IT Support', 
                'Staff Finance', 'Senior Auditor', 'System Admin'
            ], p=[0.30, 0.15, 0.15, 0.20, 0.10, 0.10])

            # Entitas and System align with Role (No SoD violation)
            if role in ['Staff Finance', 'Core Banking Operator']:
                entity = np.random.choice(['Finance Dept', 'Operations Dept', 'Jakarta Branch', 'Surabaya Branch', 'Bandung Branch', 'Head Office'])
                system = np.random.choice(['Core Banking', 'Email Server'], p=[0.7, 0.3])
                tipe_akses = np.random.choice(['Read', 'Write'], p=[0.75, 0.25])
            elif role == 'Database Administrator':
                entity = np.random.choice(['IT Dept', 'Head Office'])
                system = 'Database Server'
                tipe_akses = np.random.choice(['Read', 'Write', 'Admin'], p=[0.5, 0.35, 0.15])
            elif role == 'System Admin':
                entity = np.random.choice(['IT Dept', 'Head Office'])
                system = np.random.choice(SYSTEMS)
                tipe_akses = np.random.choice(['Read', 'Write', 'Admin'], p=[0.4, 0.3, 0.3])
            elif role == 'Senior Auditor':
                entity = np.random.choice(['Head Office', 'Jakarta Branch', 'Surabaya Branch'])
                system = np.random.choice(SYSTEMS)
                tipe_akses = 'Read'
            else: # IT Support
                entity = 'IT Dept'
                system = np.random.choice(['Email Server', 'Core Banking'])
                tipe_akses = np.random.choice(['Read', 'Write'], p=[0.8, 0.2])

            # Banking Hours: 85% normal daytime (08:00 - 18:00), 15% other hours
            if np.random.rand() < 0.85:
                jam = int(np.random.randint(8, 19))
                day = int(np.random.randint(1, 6)) # Monday - Friday
            else:
                jam = int(np.random.randint(7, 22))
                day = int(np.random.randint(1, 8))

            gagal_login = int(np.random.choice([0, 1, 2], p=[0.90, 0.08, 0.02]))
            mfa = 'Valid'
            network = np.random.choice(['Internal Secure LAN', 'Corporate VPN'], p=[0.85, 0.15])
            export_mb = round(float(np.clip(np.random.exponential(scale=12.0), 0.5, 80.0)), 1)
            
            if tipe_akses in ['Write', 'Admin'] and system in ['Database Server', 'Core Banking']:
                cr_status = 'Ada Tiket CR'
            else:
                cr_status = 'Tanpa Tiket CR'

            is_anom = False
            impact = int(np.random.choice([1, 2], p=[0.6, 0.4]))
            likelihood = int(np.random.choice([1, 2], p=[0.7, 0.3]))

        target_is_anom = 'Ya (Anomali)' if is_anom else 'Tidak'
        target_reason = " & ".join(alasan) if is_anom else '-'

        records.append({
            'ID Log': log_id,
            'Entitas': entity,
            'Peran User': role,
            'Sistem': system,
            'Jam Akses (0-23)': jam,
            'Hari Akses (1-7)': day,
            'Jumlah Gagal Login': gagal_login,
            'Tipe Akses': tipe_akses,
            'Status MFA': mfa,
            'Lokasi Akses': network,
            'Volume Ekspor Data (MB)': export_mb,
            'Status Tiket CR': cr_status,
            'TARGET: is_anomaly': target_is_anom,
            'TARGET: Impact (1-5)': impact,
            'TARGET: Likelihood (1-5)': likelihood,
            'Alasan Anomali': target_reason
        })

    df = pd.DataFrame(records)

    # Sanity deduplication on features to prevent exact accidental duplicates
    feature_cols = [c for c in df.columns if c not in ['ID Log', 'TARGET: is_anomaly', 'TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)', 'Alasan Anomali']]
    df = df.drop_duplicates(subset=feature_cols).reset_index(drop=True)

    # Re-index ID Log
    df['ID Log'] = [f"IT-{i+1:04d}" for i in range(len(df))]

    output_path = os.path.join(CURRENT_DIR, 'it_control_data.csv')
    df.to_csv(output_path, index=False)
    
    anom_count = (df['TARGET: is_anomaly'] == 'Ya (Anomali)').sum()
    print(f"[SUCCESS] Generated {len(df)} audit-ready IT Control records: {output_path}")
    print(f"Total Normal: {len(df) - anom_count} | Total Anomaly: {anom_count} ({anom_count/len(df)*100:.2f}%)")

if __name__ == "__main__":
    generate()
