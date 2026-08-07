import os
import numpy as np
import pandas as pd

np.random.seed(42)

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
N_SAMPLES = 1200

# Common entities
ENTITIES = ['Jakarta Branch', 'Surabaya Branch', 'Bandung Branch', 'Medan Branch', 'Finance Dept', 'IT Dept', 'Operations Dept']

# ======================================================================
# 1. ANOMALY_DATA.CSV (Transaction Anomalies)
# ======================================================================
print("Generating realistic anomaly_data.csv...")
descriptions = [
    'Pembayaran Vendor', 'Reimbursement Operasional', 'Transfer Kas Internal', 
    'Pengadaan Perangkat IT', 'Biaya Perjalanan Dinas', 'Jasa Konsultan Audit', 
    'Pembayaran Tagihan Listrik', 'Biaya Jamuan Klien', 'Pembelian Material Gudang', 
    'Pembayaran Gaji Lembur'
]

amounts = np.round(np.random.exponential(scale=50, size=N_SAMPLES) + 1.0, 2)
hours = np.random.randint(0, 24, size=N_SAMPLES)
days = np.random.randint(1, 8, size=N_SAMPLES)
new_ben = np.random.binomial(1, 0.25, size=N_SAMPLES)
round_amt = np.random.binomial(1, 0.35, size=N_SAMPLES)
selected_entities = np.random.choice(ENTITIES, N_SAMPLES)
selected_descs = np.random.choice(descriptions, N_SAMPLES)

is_night = ((hours >= 22) | (hours <= 5)).astype(int)
high_amount = (amounts > 120).astype(int)

anom_score = (
    0.35 * high_amount +
    0.30 * is_night +
    0.25 * new_ben +
    0.20 * round_amt +
    0.15 * (pd.Series(selected_entities).isin(['Finance Dept', 'Operations Dept']).values & is_night) +
    np.random.normal(0, 0.1, size=N_SAMPLES)
)

is_anom = (anom_score > 0.48).astype(int)
target_anom = np.where(is_anom == 1, 'Ya (Anomali)', 'Tidak')

# Impact (1-5): Driven by amount magnitude
impact = np.clip(np.ceil(np.log1p(amounts) / 1.1) + np.random.choice([-1, 0, 1], size=N_SAMPLES, p=[0.15, 0.7, 0.15]), 1, 5).astype(int)

# Likelihood (1-5): Driven by anomaly severity & violation count
likelihood = np.clip(np.ceil(anom_score * 5.0) + np.random.choice([-1, 0, 1], size=N_SAMPLES, p=[0.15, 0.7, 0.15]), 1, 5).astype(int)

df_anom = pd.DataFrame({
    'ID Transaksi': [f'TRX-{i+1:04d}' for i in range(N_SAMPLES)],
    'Entitas': selected_entities,
    'Deskripsi': selected_descs,
    'amount (dalam Juta Rp)': amounts,
    'hour_of_day (0-23)': hours,
    'day_of_week (1-7)': days,
    'is_new_beneficiary (1=Ya, 0=Tidak)': new_ben,
    'is_round_amount (1=Ya, 0=Tidak)': round_amt,
    'TARGET: is_anomaly': target_anom,
    'TARGET: Impact (1-5)': impact,
    'TARGET: Likelihood (1-5)': likelihood
})
df_anom.to_csv(os.path.join(CURRENT_DIR, 'anomaly_data.csv'), index=False)


# ======================================================================
# 2. ACCESS_PATTERN_DATA.CSV (System & Access Security)
# ======================================================================
print("Generating access_pattern_data.csv...")
roles = ['Audit Staff', 'Senior Auditor', 'Audit Manager', 'System Admin', 'External Guest']
failed_logins = np.random.negative_binomial(1, 0.3, size=N_SAMPLES)
export_mb = np.round(np.random.exponential(scale=250, size=N_SAMPLES), 1)
is_terminated = np.random.binomial(1, 0.08, size=N_SAMPLES)
ip_risk = np.round(np.clip(np.random.beta(0.5, 2.0, size=N_SAMPLES), 0.0, 1.0), 2)

access_anom_score = (
    0.40 * is_terminated +
    0.30 * (ip_risk > 0.6).astype(int) +
    0.20 * (failed_logins >= 4).astype(int) +
    0.20 * (export_mb > 500).astype(int) +
    np.random.normal(0, 0.08, size=N_SAMPLES)
)
is_acc_anom = (access_anom_score > 0.42).astype(int)
acc_target_anom = np.where(is_acc_anom == 1, 'Ya (Anomali)', 'Tidak')

acc_impact = np.clip(np.ceil((export_mb / 250) + is_terminated * 2) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
acc_likelihood = np.clip(np.ceil(ip_risk * 4.0 + failed_logins * 0.4) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

df_access = pd.DataFrame({
    'ID Akses': [f'ACC-{i+1:04d}' for i in range(N_SAMPLES)],
    'User ID': [f'USR-{np.random.randint(100, 999)}' for _ in range(N_SAMPLES)],
    'Entitas': np.random.choice(ENTITIES, N_SAMPLES),
    'Peran User': np.random.choice(roles, N_SAMPLES),
    'hour_of_day (0-23)': np.random.randint(0, 24, size=N_SAMPLES),
    'day_of_week (1-7)': np.random.randint(1, 8, size=N_SAMPLES),
    'failed_login_attempts': failed_logins,
    'data_export_volume_mb': export_mb,
    'is_terminated_user (1=Ya, 0=Tidak)': is_terminated,
    'ip_risk_score (0.0-1.0)': ip_risk,
    'TARGET: is_anomaly': acc_target_anom,
    'TARGET: Impact (1-5)': acc_impact,
    'TARGET: Likelihood (1-5)': acc_likelihood
})
df_access.to_csv(os.path.join(CURRENT_DIR, 'access_pattern_data.csv'), index=False)


# ======================================================================
# 3. AUDIT_BUDGET_DATA.CSV (Audit Planning & Budgeting)
# ======================================================================
print("Generating audit_budget_data.csv...")
categories = ['Audit Keuangan', 'Audit TI', 'Audit Operasional', 'Audit Kepatuhan', 'Special Investigation']
hist_avg_budget = np.round(np.random.uniform(50, 400, size=N_SAMPLES), 1)
budget_ratio = np.random.normal(1.0, 0.35, size=N_SAMPLES)
allocated_budget = np.round(np.maximum(10.0, hist_avg_budget * np.maximum(0.3, budget_ratio)), 1)
mandays = np.random.randint(10, 180, size=N_SAMPLES)
auditors = np.random.randint(2, 12, size=N_SAMPLES)
budget_zscore = np.round((allocated_budget - hist_avg_budget) / 50.0, 2)
unapproved_spike = ((budget_zscore > 2.0) & (np.random.binomial(1, 0.7, size=N_SAMPLES) == 1)).astype(int)

budget_anom_score = (
    0.45 * unapproved_spike +
    0.35 * (budget_zscore > 2.2).astype(int) +
    0.20 * (allocated_budget / (auditors * 5) > 15).astype(int) +
    np.random.normal(0, 0.1, size=N_SAMPLES)
)
is_budg_anom = (budget_anom_score > 0.45).astype(int)
budg_target_anom = np.where(is_budg_anom == 1, 'Ya (Anomali)', 'Tidak')

budg_impact = np.clip(np.ceil(allocated_budget / 90.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
budg_likelihood = np.clip(np.ceil(np.maximum(0, budget_zscore) + unapproved_spike * 1.5) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

df_budget = pd.DataFrame({
    'ID Rencana Audit': [f'PLN-{i+1:04d}' for i in range(N_SAMPLES)],
    'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
    'Kategori Perencanaan': np.random.choice(categories, N_SAMPLES),
    'Deskripsi Perencanaan': [f'Evaluasi Perencanaan Risk Area {i+1}' for i in range(N_SAMPLES)],
    'allocated_budget_juta': allocated_budget,
    'historical_avg_budget_juta': hist_avg_budget,
    'allocated_mandays': mandays,
    'assigned_auditors_count': auditors,
    'budget_zscore': budget_zscore,
    'is_unapproved_budget_spike (1=Ya, 0=Tidak)': unapproved_spike,
    'TARGET: is_anomaly': budg_target_anom,
    'TARGET: Impact (1-5)': budg_impact,
    'TARGET: Likelihood (1-5)': budg_likelihood
})
df_budget.to_csv(os.path.join(CURRENT_DIR, 'audit_budget_data.csv'), index=False)


# ======================================================================
# 4. EXPENSE_REPORT_DATA.CSV (Expense Claims & Reimbursement)
# ======================================================================
print("Generating expense_report_data.csv...")
expense_cats = ['Travel & Entertainment', 'Kertas & Alat Tulis', 'Konsumsi Rapat', 'Sewa Kendaraan Operasional', 'Vendor Pihak Ketiga']
claim_amount = np.round(np.random.exponential(scale=15, size=N_SAMPLES) + 0.5, 2)
seq_invoice = np.random.binomial(1, 0.12, size=N_SAMPLES)
exact_round = np.random.binomial(1, 0.25, size=N_SAMPLES)
dup_claims = np.random.binomial(3, 0.1, size=N_SAMPLES)
days_since_last = np.random.randint(0, 45, size=N_SAMPLES)
over_limit = ((claim_amount > 25.0) & np.random.binomial(1, 0.8, size=N_SAMPLES)).astype(int)

exp_anom_score = (
    0.35 * seq_invoice +
    0.30 * (dup_claims >= 1).astype(int) +
    0.25 * over_limit +
    0.15 * ((days_since_last <= 1) & (claim_amount > 10)).astype(int) +
    np.random.normal(0, 0.08, size=N_SAMPLES)
)
is_exp_anom = (exp_anom_score > 0.38).astype(int)
exp_target_anom = np.where(is_exp_anom == 1, 'Ya (Anomali)', 'Tidak')

exp_impact = np.clip(np.ceil(claim_amount / 8.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
exp_likelihood = np.clip(np.ceil(dup_claims + seq_invoice * 2.0 + (days_since_last == 0) * 1.5) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

df_expense = pd.DataFrame({
    'ID Klaim': [f'CLM-{i+1:04d}' for i in range(N_SAMPLES)],
    'Entitas': np.random.choice(ENTITIES, N_SAMPLES),
    'Kategori Pengeluaran': np.random.choice(expense_cats, N_SAMPLES),
    'Deskripsi Pengeluaran': [f'Klaim Pengeluaran Operasional #{i+1}' for i in range(N_SAMPLES)],
    'claim_amount (Juta Rp)': claim_amount,
    'is_sequential_invoice (1=Ya, 0=Tidak)': seq_invoice,
    'is_exact_round_amount (1=Ya, 0=Tidak)': exact_round,
    'duplicate_claim_count': dup_claims,
    'days_since_last_claim': days_since_last,
    'is_over_limit_threshold (1=Ya, 0=Tidak)': over_limit,
    'TARGET: is_anomaly': exp_target_anom,
    'TARGET: Impact (1-5)': exp_impact,
    'TARGET: Likelihood (1-5)': exp_likelihood
})
df_expense.to_csv(os.path.join(CURRENT_DIR, 'expense_report_data.csv'), index=False)


# ======================================================================
# 5. FIELDWORK_DATA.CSV (Fieldwork & Audit Execution)
# ======================================================================
print("Generating fieldwork_data.csv...")
audit_types = ['Full Scope Audit', 'Limited Review', 'Surprise Audit', 'Compliance Audit']
comp_duration = np.random.randint(2, 60, size=N_SAMPLES)
hist_avg_duration = np.random.randint(15, 45, size=N_SAMPLES)
sampling_cov = np.round(np.random.uniform(5.0, 95.0, size=N_SAMPLES), 1)
findings_count = np.random.negative_binomial(2, 0.4, size=N_SAMPLES)
dur_zscore = np.round((comp_duration - hist_avg_duration) / 10.0, 2)
instant_approval = ((comp_duration <= 3) & (sampling_cov < 20.0) & np.random.binomial(1, 0.85, size=N_SAMPLES)).astype(int)

field_anom_score = (
    0.45 * instant_approval +
    0.30 * (sampling_cov < 10.0).astype(int) +
    0.20 * (dur_zscore < -1.8).astype(int) +
    0.15 * (findings_count == 0).astype(int) * (comp_duration > 30) +
    np.random.normal(0, 0.08, size=N_SAMPLES)
)
is_field_anom = (field_anom_score > 0.40).astype(int)
field_target_anom = np.where(is_field_anom == 1, 'Ya (Anomali)', 'Tidak')

field_impact = np.clip(np.ceil(findings_count * 0.8 + (100 - sampling_cov) / 25.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
field_likelihood = np.clip(np.ceil(instant_approval * 3.0 + np.maximum(0, -dur_zscore) * 1.2) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

df_fieldwork = pd.DataFrame({
    'ID Penugasan': [f'ASG-{i+1:04d}' for i in range(N_SAMPLES)],
    'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
    'Jenis Audit': np.random.choice(audit_types, N_SAMPLES),
    'Deskripsi Fieldwork': [f'Penugasan Audit Lapangan #{i+1}' for i in range(N_SAMPLES)],
    'completion_duration_days': comp_duration,
    'historical_avg_duration_days': hist_avg_duration,
    'sampling_coverage_pct': sampling_cov,
    'audit_findings_count': findings_count,
    'duration_zscore': dur_zscore,
    'is_instant_approval (1=Ya, 0=Tidak)': instant_approval,
    'TARGET: is_anomaly': field_target_anom,
    'TARGET: Impact (1-5)': field_impact,
    'TARGET: Likelihood (1-5)': field_likelihood
})
df_fieldwork.to_csv(os.path.join(CURRENT_DIR, 'fieldwork_data.csv'), index=False)


# ======================================================================
# 6. INVENTORY_DATA.CSV (Inventory Adjustments)
# ======================================================================
print("Generating inventory_data.csv...")
item_cats = ['Suku Cadang Utama', 'Bahan Baku Industri', 'Barang Jadi', 'Peralatan TI', 'Komoditas Umum']
adj_amount = np.round(np.random.exponential(scale=30, size=N_SAMPLES) + 1.0, 2)
stock_var_pct = np.round(np.random.uniform(0.5, 35.0, size=N_SAMPLES), 1)
adj_freq = np.random.randint(1, 15, size=N_SAMPLES)
has_doc = np.random.binomial(1, 0.75, size=N_SAMPLES)
role_level = np.random.randint(1, 5, size=N_SAMPLES)
manual_override = np.random.binomial(1, 0.20, size=N_SAMPLES)

inv_anom_score = (
    0.35 * manual_override +
    0.30 * (1 - has_doc) +
    0.25 * (stock_var_pct > 20.0).astype(int) +
    0.20 * (adj_freq > 8).astype(int) +
    np.random.normal(0, 0.08, size=N_SAMPLES)
)
is_inv_anom = (inv_anom_score > 0.45).astype(int)
inv_target_anom = np.where(is_inv_anom == 1, 'Ya (Anomali)', 'Tidak')

inv_impact = np.clip(np.ceil(adj_amount / 15.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
inv_likelihood = np.clip(np.ceil((1 - has_doc) * 2.0 + manual_override * 1.5 + (adj_freq / 3.0)) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

df_inventory = pd.DataFrame({
    'ID Penyesuaian': [f'ADJ-{i+1:04d}' for i in range(N_SAMPLES)],
    'Entitas Gudang': np.random.choice(ENTITIES, N_SAMPLES),
    'Kategori Barang': np.random.choice(item_cats, N_SAMPLES),
    'Deskripsi Penyesuaian': [f'Penyesuaian Stok Gudang #{i+1}' for i in range(N_SAMPLES)],
    'adjustment_amount (Juta Rp)': adj_amount,
    'stock_variance_pct': stock_var_pct,
    'adjustment_frequency_month': adj_freq,
    'has_supporting_doc (1=Ya, 0=Tidak)': has_doc,
    'approver_role_level (1-4)': role_level,
    'is_manual_override (1=Ya, 0=Tidak)': manual_override,
    'TARGET: is_anomaly': inv_target_anom,
    'TARGET: Impact (1-5)': inv_impact,
    'TARGET: Likelihood (1-5)': inv_likelihood
})
df_inventory.to_csv(os.path.join(CURRENT_DIR, 'inventory_data.csv'), index=False)


# ======================================================================
# 7. MITIGATION_OVERDUE_DATA.CSV (Overdue Remediation)
# ======================================================================
print("Generating mitigation_overdue_data.csv...")
finding_cats = ['Kepatuhan SOP', 'Pengendalian Internal', 'Sistem Informasi', 'Manajemen Risiko']
severities = ['Rendah', 'Sedang', 'Tinggi', 'Kritis']
overdue_days = np.random.randint(1, 180, size=N_SAMPLES)
repeated_overdue = np.random.randint(0, 6, size=N_SAMPLES)
progress_pct = np.round(np.random.uniform(0.0, 95.0, size=N_SAMPLES), 1)
unjustified_delay = ((overdue_days > 45) & (progress_pct < 25.0) & np.random.binomial(1, 0.8, size=N_SAMPLES)).astype(int)

mit_anom_score = (
    0.40 * unjustified_delay +
    0.30 * (repeated_overdue >= 2).astype(int) +
    0.20 * (overdue_days > 90).astype(int) +
    0.15 * (progress_pct < 10.0).astype(int) +
    np.random.normal(0, 0.08, size=N_SAMPLES)
)
is_mit_anom = (mit_anom_score > 0.42).astype(int)
mit_target_anom = np.where(is_mit_anom == 1, 'Ya (Anomali)', 'Tidak')

mit_impact = np.clip(np.ceil((overdue_days / 40.0) + repeated_overdue * 0.5) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
mit_likelihood = np.clip(np.ceil(repeated_overdue * 1.2 + unjustified_delay * 2.0 + (100 - progress_pct) / 30.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

df_mitigation = pd.DataFrame({
    'ID Mitigasi': [f'MIT-{i+1:04d}' for i in range(N_SAMPLES)],
    'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
    'Kategori Temuan': np.random.choice(finding_cats, N_SAMPLES),
    'Tingkat Keparahan Temuan': np.random.choice(severities, N_SAMPLES),
    'Deskripsi Rencana Aksi': [f'Rencana Aksi Mitigasi Temuan #{i+1}' for i in range(N_SAMPLES)],
    'overdue_days': overdue_days,
    'repeated_overdue_count': repeated_overdue,
    'mitigation_progress_pct': progress_pct,
    'is_unjustified_delay (1=Ya, 0=Tidak)': unjustified_delay,
    'TARGET: is_anomaly': mit_target_anom,
    'TARGET: Impact (1-5)': mit_impact,
    'TARGET: Likelihood (1-5)': mit_likelihood
})
df_mitigation.to_csv(os.path.join(CURRENT_DIR, 'mitigation_overdue_data.csv'), index=False)


# ======================================================================
# 8. REPEAT_FINDING_DATA.CSV (Recurring Audit Findings)
# ======================================================================
print("Generating repeat_finding_data.csv...")
risk_cats = ['Operasional', 'Finansial', 'Legal & Kepatuhan', 'Reputasi', 'Keamanan TI']
repeat_count = np.random.negative_binomial(1, 0.4, size=N_SAMPLES)
similarity_pct = np.round(np.random.uniform(20.0, 99.0, size=N_SAMPLES), 1)
same_root = np.random.binomial(1, 0.35, size=N_SAMPLES)
prev_closed = np.random.binomial(1, 0.60, size=N_SAMPLES)

rep_anom_score = (
    0.35 * (repeat_count >= 2).astype(int) +
    0.30 * (similarity_pct > 80.0).astype(int) * same_root +
    0.25 * (prev_closed & (similarity_pct > 75.0)).astype(int) +
    np.random.normal(0, 0.08, size=N_SAMPLES)
)
is_rep_anom = (rep_anom_score > 0.40).astype(int)
rep_target_anom = np.where(is_rep_anom == 1, 'Ya (Anomali)', 'Tidak')

rep_impact = np.clip(np.ceil(repeat_count * 1.2 + same_root * 1.5) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
rep_likelihood = np.clip(np.ceil((similarity_pct / 25.0) + prev_closed * 1.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

df_repeat = pd.DataFrame({
    'ID Temuan': [f'FND-{i+1:04d}' for i in range(N_SAMPLES)],
    'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
    'Kategori Risiko': np.random.choice(risk_cats, N_SAMPLES),
    'Deskripsi Temuan': [f'Deskripsi Temuan Audit Berulang #{i+1}' for i in range(N_SAMPLES)],
    'repeat_count_last_year': repeat_count,
    'finding_similarity_pct': similarity_pct,
    'is_same_root_cause (1=Ya, 0=Tidak)': same_root,
    'is_previously_closed (1=Ya, 0=Tidak)': prev_closed,
    'severity_category': np.random.choice(severities, N_SAMPLES),
    'TARGET: is_anomaly': rep_target_anom,
    'TARGET: Impact (1-5)': rep_impact,
    'TARGET: Likelihood (1-5)': rep_likelihood
})
df_repeat.to_csv(os.path.join(CURRENT_DIR, 'repeat_finding_data.csv'), index=False)


# ======================================================================
# 9. RISK_SCORE_SPIKE_DATA.CSV (Sudden Volatility in Risk Score)
# ======================================================================
print("Generating risk_score_spike_data.csv...")
prev_score = np.round(np.random.uniform(3.0, 18.0, size=N_SAMPLES), 1)
diff = np.round(np.random.normal(1.5, 4.0, size=N_SAMPLES), 1)
curr_score = np.round(np.clip(prev_score + diff, 1.0, 25.0), 1)
new_findings = np.random.negative_binomial(1, 0.3, size=N_SAMPLES)
volatility = np.round(np.random.exponential(scale=0.15, size=N_SAMPLES), 2)
unexplained_spike = ((diff > 6.0) & (new_findings <= 1) & np.random.binomial(1, 0.8, size=N_SAMPLES)).astype(int)

spike_anom_score = (
    0.45 * unexplained_spike +
    0.30 * (diff > 8.0).astype(int) +
    0.20 * (volatility > 0.35).astype(int) +
    np.random.normal(0, 0.08, size=N_SAMPLES)
)
is_spike_anom = (spike_anom_score > 0.42).astype(int)
spike_target_anom = np.where(is_spike_anom == 1, 'Ya (Anomali)', 'Tidak')

spike_impact = np.clip(np.ceil(curr_score / 5.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)
spike_likelihood = np.clip(np.ceil(unexplained_spike * 2.5 + volatility * 8.0) + np.random.choice([0, 1], size=N_SAMPLES, p=[0.7, 0.3]), 1, 5).astype(int)

df_spike = pd.DataFrame({
    'ID Penilaian Risiko': [f'RSK-{i+1:04d}' for i in range(N_SAMPLES)],
    'Entitas Auditee': np.random.choice(ENTITIES, N_SAMPLES),
    'Kategori Risiko': np.random.choice(risk_cats, N_SAMPLES),
    'Deskripsi Evaluasi Risiko': [f'Evaluasi Lonjakan Skor Risiko #{i+1}' for i in range(N_SAMPLES)],
    'current_risk_score': curr_score,
    'previous_risk_score': prev_score,
    'risk_score_diff': diff,
    'new_findings_count': new_findings,
    'risk_score_volatility': volatility,
    'is_unexplained_spike (1=Ya, 0=Tidak)': unexplained_spike,
    'TARGET: is_anomaly': spike_target_anom,
    'TARGET: Impact (1-5)': spike_impact,
    'TARGET: Likelihood (1-5)': spike_likelihood
})
df_spike.to_csv(os.path.join(CURRENT_DIR, 'risk_score_spike_data.csv'), index=False)

print("\n[SUCCESS] ALL 9 SYNTHETIC DATASETS GENERATED SUCCESSFULLY WITH REALISTIC DOMAIN LOGIC!")
