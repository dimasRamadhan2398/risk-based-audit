import os
import re
import random
import pickle
import json
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder, StandardScaler, MinMaxScaler
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier, ExtraTreesClassifier, IsolationForest
from sklearn.linear_model import LogisticRegression
from sklearn.feature_extraction.text import TfidfVectorizer
from xgboost import XGBClassifier

# Set seeds for reproducibility
def set_seed(seed=42):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    if torch.cuda.is_available():
        torch.cuda.manual_seed_all(seed)

set_seed(42)

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.abspath(os.path.join(CURRENT_DIR, '..'))
OUTPUT_MODEL_DIR = os.path.join(ROOT_DIR, 'backend', 'python-ai', 'models')
os.makedirs(OUTPUT_MODEL_DIR, exist_ok=True)

print(f"Target Output Model Directory: {OUTPUT_MODEL_DIR}\n")

# ======================================================================
# 1. ANOMALY PREDICTION MODEL SUITE (Primary + 8 Domain Sub-Models)
# ======================================================================
print("--- 1. Training & Saving Anomaly Models ---")
anomaly_csv = os.path.join(CURRENT_DIR, 'anomaly_prediction', 'anomaly_data.csv')
df_anom = pd.read_csv(anomaly_csv)

# Clean text columns
for col in ['Entitas', 'Deskripsi', 'TARGET: is_anomaly']:
    if col in df_anom.columns:
        df_anom[col] = df_anom[col].astype(str).str.strip()

# Amount parsing
df_anom['amount'] = pd.to_numeric(
    df_anom['amount (dalam Juta Rp)'].astype(str).str.replace(',', '', regex=False),
    errors='coerce'
)

# Numeric range check
for col in ['hour_of_day (0-23)', 'day_of_week (1-7)', 'is_new_beneficiary (1=Ya, 0=Tidak)', 'is_round_amount (1=Ya, 0=Tidak)']:
    df_anom[col] = pd.to_numeric(df_anom[col], errors='coerce')

df_anom = df_anom.dropna(subset=['Entitas', 'Deskripsi', 'amount', 'hour_of_day (0-23)', 'day_of_week (1-7)']).reset_index(drop=True)

# Feature engineering
df_anom['log_amount'] = np.log1p(df_anom['amount'])
df_anom['is_weekend'] = df_anom['day_of_week (1-7)'].isin([6, 7]).astype(int)
df_anom['is_night'] = ((df_anom['hour_of_day (0-23)'] >= 22) | (df_anom['hour_of_day (0-23)'] <= 5)).astype(int)
df_anom['amount_per_hour'] = df_anom['amount'] / (df_anom['hour_of_day (0-23)'] + 1)
df_anom['hour_sin'] = np.sin(2 * np.pi * df_anom['hour_of_day (0-23)'] / 24)
df_anom['hour_cos'] = np.cos(2 * np.pi * df_anom['hour_of_day (0-23)'] / 24)
df_anom['day_sin'] = np.sin(2 * np.pi * (df_anom['day_of_week (1-7)'] - 1) / 7)
df_anom['day_cos'] = np.cos(2 * np.pi * (df_anom['day_of_week (1-7)'] - 1) / 7)

feature_cols_anom = [
    'Entitas', 'Deskripsi', 'amount', 'hour_of_day (0-23)', 
    'day_of_week (1-7)', 'is_new_beneficiary (1=Ya, 0=Tidak)', 
    'is_round_amount (1=Ya, 0=Tidak)', 'log_amount', 
    'is_weekend', 'is_night', 'amount_per_hour',
    'hour_sin', 'hour_cos', 'day_sin', 'day_cos'
]

X_raw_anom = df_anom[feature_cols_anom].values
ct_anom = ColumnTransformer(transformers=[('encoder', OneHotEncoder(handle_unknown='ignore', sparse_output=False), [0, 1])], remainder='passthrough')
X_anom = ct_anom.fit_transform(X_raw_anom)

sc_anom = StandardScaler()
X_anom_scaled = sc_anom.fit_transform(X_anom)

# Targets
y_anom_is = np.where(df_anom['TARGET: is_anomaly'] == 'Ya (Anomali)', 1, 0)
y_anom_imp = df_anom['TARGET: Impact (1-5)'].values.astype(int) - 1
y_anom_lik = df_anom['TARGET: Likelihood (1-5)'].values.astype(int) - 1

# Train Models
num_neg = (y_anom_is == 0).sum()
num_pos = max((y_anom_is == 1).sum(), 1)
scale_pos = num_neg / num_pos

# 1a. Supervised XGBoost Classifier with tuned hyperparameters
anom_clf = XGBClassifier(
    scale_pos_weight=scale_pos,
    eval_metric='logloss',
    max_depth=4,
    learning_rate=0.05,
    n_estimators=100,
    random_state=42
)
anom_clf.fit(X_anom_scaled, y_anom_is)

# 1b. Unsupervised Isolation Forest Anomaly Detector
iso_forest_anom = IsolationForest(contamination=0.20, random_state=42)
iso_forest_anom.fit(X_anom_scaled)

# 1c. Impact & Likelihood Classifiers
anom_imp_clf = RandomForestClassifier(n_estimators=200, class_weight='balanced', random_state=42)
anom_imp_clf.fit(X_anom_scaled, y_anom_imp)

anom_lik_clf = RandomForestClassifier(n_estimators=200, class_weight='balanced', random_state=42)
anom_lik_clf.fit(X_anom_scaled, y_anom_lik)

# 1d. Train 8 Domain-Specific Sub-Models
sub_domain_files = {
    'access_pattern': 'access_pattern_data.csv',
    'audit_budget': 'audit_budget_data.csv',
    'expense_report': 'expense_report_data.csv',
    'fieldwork': 'fieldwork_data.csv',
    'inventory': 'inventory_data.csv',
    'mitigation_overdue': 'mitigation_overdue_data.csv',
    'repeat_finding': 'repeat_finding_data.csv',
    'risk_score_spike': 'risk_score_spike_data.csv'
}

sub_models_bundle = {}

for domain_name, csv_filename in sub_domain_files.items():
    sub_csv_path = os.path.join(CURRENT_DIR, 'anomaly_prediction', csv_filename)
    if os.path.exists(sub_csv_path):
        df_sub = pd.read_csv(sub_csv_path)
        feat_cols = [c for c in df_sub.columns if not c.startswith('TARGET') and not c.startswith('ID') and c != 'User ID']
        
        cat_cols = [c for c in feat_cols if (df_sub[c].dtype == object or df_sub[c].dtype == 'string') and df_sub[c].nunique() <= 50]
        num_cols = [c for c in feat_cols if df_sub[c].dtype in ['int64', 'float64']]
        
        sub_feat_cols = cat_cols + num_cols
        
        sub_ct = ColumnTransformer(
            transformers=[('cat', OneHotEncoder(handle_unknown='ignore', sparse_output=False), cat_cols)],
            remainder='passthrough'
        )
        X_sub = sub_ct.fit_transform(df_sub[sub_feat_cols])
        
        sub_sc = StandardScaler()
        X_sub_scaled = sub_sc.fit_transform(X_sub)
        
        y_sub_is = np.where(df_sub['TARGET: is_anomaly'] == 'Ya (Anomali)', 1, 0)
        y_sub_imp = df_sub['TARGET: Impact (1-5)'].values.astype(int) - 1
        y_sub_lik = df_sub['TARGET: Likelihood (1-5)'].values.astype(int) - 1
        
        sub_pos = max((y_sub_is == 1).sum(), 1)
        sub_neg = (y_sub_is == 0).sum()
        sub_scale_pos = sub_neg / sub_pos
        
        sub_clf = XGBClassifier(scale_pos_weight=sub_scale_pos, eval_metric='logloss', max_depth=4, random_state=42)
        sub_clf.fit(X_sub_scaled, y_sub_is)
        
        sub_iso = IsolationForest(contamination=max(0.05, float(y_sub_is.mean())), random_state=42)
        sub_iso.fit(X_sub_scaled)
        
        sub_imp_clf = RandomForestClassifier(n_estimators=100, class_weight='balanced', random_state=42)
        sub_imp_clf.fit(X_sub_scaled, y_sub_imp)
        
        sub_lik_clf = RandomForestClassifier(n_estimators=100, class_weight='balanced', random_state=42)
        sub_lik_clf.fit(X_sub_scaled, y_sub_lik)
        
        sub_models_bundle[domain_name] = {
            'classifier': sub_clf,
            'isolation_forest': sub_iso,
            'impact_classifier': sub_imp_clf,
            'likelihood_classifier': sub_lik_clf,
            'column_transformer': sub_ct,
            'scaler': sub_sc,
            'feature_cols': sub_feat_cols
        }
        print(f" -> Trained domain sub-model [{domain_name}]")

# Save complete Anomaly Model Suite
anomaly_bundle = {
    'classifier': anom_clf,
    'isolation_forest': iso_forest_anom,
    'impact_classifier': anom_imp_clf,
    'likelihood_classifier': anom_lik_clf,
    'column_transformer': ct_anom,
    'scaler': sc_anom,
    'best_threshold': 0.40,
    'feature_cols': feature_cols_anom,
    'known_entitas': sorted(df_anom['Entitas'].unique().tolist()),
    'known_deskripsi': sorted(df_anom['Deskripsi'].unique().tolist()),
    'sub_models': sub_models_bundle
}

with open(os.path.join(OUTPUT_MODEL_DIR, 'anomaly_bundle.pkl'), 'wb') as f:
    pickle.dump(anomaly_bundle, f)

print(" -> Saved enhanced anomaly_bundle.pkl with IsolationForest & 8 Domain Sub-Models successfully.")

# ======================================================================
# 2. DEPARTMENT RISK PREDICTION MODEL SUITE
# ======================================================================
print("\n--- 2. Training & Saving Department Risk Models ---")
dept_csv = os.path.join(CURRENT_DIR, 'department_prediction', 'department_dataset.csv')
df_dept = pd.read_csv(dept_csv)

df_dept = df_dept.rename(
    columns={
        "Entitas": "entity",
        "Kategori Risiko": "risk_category",
        "Inherent Likelihood": "inherent_likelihood",
        "Inherent Impact": "inherent_impact",
        "Jml Temuan Audit": "audit_findings_count",
        "Jml KPI di Bawah Target": "kpi_below_target_count",
        "Volatilitas KPI": "kpi_volatility",
        "Skor Risiko Periode Lalu": "previous_risk_score",
        "Bulan Penilaian": "assessment_month",
        "TARGET: Likelihood": "target_likelihood",
        "TARGET: Impact": "target_impact",
    }
)

df_dept['entity'] = df_dept['entity'].astype(str).str.strip()
df_dept['risk_category'] = df_dept['risk_category'].astype(str).str.strip()

num_cols_dept = [
    "inherent_likelihood", "inherent_impact", "audit_findings_count",
    "kpi_below_target_count", "kpi_volatility", "previous_risk_score",
    "assessment_month", "target_likelihood", "target_impact"
]
df_dept[num_cols_dept] = df_dept[num_cols_dept].apply(pd.to_numeric, errors="coerce")
df_dept = df_dept.dropna().reset_index(drop=True)

df_dept["month_sin"] = np.sin(2 * np.pi * (df_dept["assessment_month"] - 1) / 12)
df_dept["month_cos"] = np.cos(2 * np.pi * (df_dept["assessment_month"] - 1) / 12)

df_dept['inherent_risk_score'] = df_dept['inherent_likelihood'] * df_dept['inherent_impact']
df_dept['kpi_volatility_log'] = np.log1p(df_dept['kpi_volatility'])
df_dept['findings_kpi_ratio'] = (df_dept['audit_findings_count'] + 1) / (df_dept['kpi_below_target_count'] + 1)
df_dept['risk_score_diff'] = df_dept['inherent_risk_score'] - df_dept['previous_risk_score']

feature_cols_dept = [
    'entity', 'risk_category', 'inherent_likelihood', 'inherent_impact',
    'audit_findings_count', 'kpi_below_target_count', 'kpi_volatility',
    'previous_risk_score', 'assessment_month', 'inherent_risk_score',
    'kpi_volatility_log', 'findings_kpi_ratio', 'risk_score_diff',
    'month_sin', 'month_cos'
]

X_raw_dept = df_dept[feature_cols_dept].values
ct_dept = ColumnTransformer(transformers=[('encoder', OneHotEncoder(handle_unknown='ignore'), [0, 1])], remainder='passthrough')
X_dept = ct_dept.fit_transform(X_raw_dept)

sc_dept = StandardScaler()
X_dept_scaled = sc_dept.fit_transform(X_dept)

le_dept_imp = LabelEncoder()
y_dept_imp = le_dept_imp.fit_transform(df_dept['target_impact'])

le_dept_lik = LabelEncoder()
y_dept_lik = le_dept_lik.fit_transform(df_dept['target_likelihood'])

dept_imp_clf = GradientBoostingClassifier(random_state=42)
dept_imp_clf.fit(X_dept_scaled, y_dept_imp)

dept_lik_clf = GradientBoostingClassifier(random_state=42)
dept_lik_clf.fit(X_dept_scaled, y_dept_lik)

dept_bundle = {
    'impact_classifier': dept_imp_clf,
    'likelihood_classifier': dept_lik_clf,
    'label_encoder_impact': le_dept_imp,
    'label_encoder_likelihood': le_dept_lik,
    'column_transformer': ct_dept,
    'scaler': sc_dept,
    'feature_cols': feature_cols_dept,
    'known_entities': sorted(df_dept['entity'].unique().tolist()),
    'known_categories': sorted(df_dept['risk_category'].unique().tolist())
}

with open(os.path.join(OUTPUT_MODEL_DIR, 'department_bundle.pkl'), 'wb') as f:
    pickle.dump(dept_bundle, f)

print(" -> Saved department_bundle.pkl successfully.")

# ======================================================================
# 3. DOCUMENT NLP (INDOBERT / TF-IDF BUNDLE) MODEL SUITE
# ======================================================================
print("\n--- 3. Training & Saving Document NLP Models ---")
doc_csv = os.path.join(CURRENT_DIR, 'document_prediction', 'document_data.csv')
df_doc = pd.read_csv(doc_csv).drop_duplicates().reset_index(drop=True)

def clean_audit_text(text):
    text = str(text)
    text = re.sub(r'^\s*\d+[\.\)]\s*', '', text)
    return text.strip()

df_doc['text_clean'] = df_doc['Teks Input (Kutipan dari Laporan Audit)'].apply(clean_audit_text)

target_map_doc = {
    'risk_category': [c for c in df_doc.columns if 'Kategori Risiko' in c][0],
    'sentiment': [c for c in df_doc.columns if 'Sentimen' in c][0],
    'impact': [c for c in df_doc.columns if 'Impact' in c][0],
    'likelihood': [c for c in df_doc.columns if 'Likelihood' in c][0]
}

tfidf_vec = TfidfVectorizer(ngram_range=(1, 2), max_features=5000, sublinear_tf=True)
X_doc_tfidf = tfidf_vec.fit_transform(df_doc['text_clean'])

doc_models = {}
doc_label_encoders = {}

for key, col_name in target_map_doc.items():
    sub_df = df_doc.dropna(subset=['text_clean', col_name]).copy()
    le = LabelEncoder()
    y_sub = le.fit_transform(sub_df[col_name].astype(str))
    
    X_sub_tfidf = tfidf_vec.transform(sub_df['text_clean'])
    
    clf = LogisticRegression(max_iter=1000, random_state=42)
    clf.fit(X_sub_tfidf, y_sub)
    
    doc_models[key] = clf
    doc_label_encoders[key] = le

doc_bundle = {
    'tfidf_vectorizer': tfidf_vec,
    'models': doc_models,
    'label_encoders': doc_label_encoders
}

with open(os.path.join(OUTPUT_MODEL_DIR, 'document_bundle.pkl'), 'wb') as f:
    pickle.dump(doc_bundle, f)

print(" -> Saved document_bundle.pkl successfully.")

# ======================================================================
# 4. KPI PYTORCH LSTM & TABULAR MODEL SUITE
# ======================================================================
print("\n--- 4. Training & Saving KPI PyTorch LSTM & Tabular Models ---")
kpi_csv = os.path.join(CURRENT_DIR, 'kpi_prediction', 'kpi_data.csv')
df_kpi = pd.read_csv(kpi_csv).drop_duplicates().reset_index(drop=True)

reg_target = 'TARGET: Nilai Aktual (%)'

SEQ_LENGTH = 10
X_reg_list, y_reg_list = [], []
kpi_scalers = {}

for kpi_name, group in df_kpi.groupby('Nama KPI'):
    vals = pd.to_numeric(group[reg_target], errors='coerce').dropna().values.reshape(-1, 1)
    if len(vals) <= SEQ_LENGTH:
        continue
    
    scaler = MinMaxScaler(feature_range=(0, 1))
    scaled_vals = scaler.fit_transform(vals)
    kpi_scalers[kpi_name] = scaler
    
    for i in range(len(scaled_vals) - SEQ_LENGTH):
        X_reg_list.append(scaled_vals[i:i + SEQ_LENGTH])
        y_reg_list.append(scaled_vals[i + SEQ_LENGTH])

X_reg = np.array(X_reg_list, dtype=np.float32)
y_reg = np.array(y_reg_list, dtype=np.float32)

class PyTorchLSTMRegressor(nn.Module):
    def __init__(self, input_dim=1, hidden_dim=64, num_layers=2, output_dim=1):
        super().__init__()
        self.hidden_dim = hidden_dim
        self.num_layers = num_layers
        self.lstm = nn.LSTM(input_size=input_dim, hidden_size=hidden_dim, num_layers=num_layers, batch_first=True)
        self.fc = nn.Linear(hidden_dim, output_dim)
        
    def forward(self, x):
        h0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        c0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        out, _ = self.lstm(x, (h0, c0))
        return self.fc(out[:, -1, :])

lstm_model = PyTorchLSTMRegressor()
criterion = nn.MSELoss()
optimizer = torch.optim.Adam(lstm_model.parameters(), lr=0.001)

inputs = torch.tensor(X_reg)
targets = torch.tensor(y_reg)

lstm_model.train()
for epoch in range(100):
    optimizer.zero_grad()
    outputs = lstm_model(inputs)
    loss = criterion(outputs, targets)
    loss.backward()
    optimizer.step()

# Global min/max of reg_target for unscaling fallback
all_actual_vals = pd.to_numeric(df_kpi[reg_target], errors='coerce').dropna().values
global_min_val = float(all_actual_vals.min())
global_max_val = float(all_actual_vals.max())

# KPI Classification models for Impact & Likelihood
df_feat = df_kpi.copy()
for lag in [1, 2, 3]:
    df_feat[f'lag_{lag}'] = df_feat.groupby('Nama KPI')[reg_target].shift(lag)
df_feat['rolling_mean_3'] = df_feat.groupby('Nama KPI')[reg_target].shift(1).rolling(3).mean()
df_feat = df_feat.dropna().reset_index(drop=True)

ohe_kpi = OneHotEncoder(sparse_output=False, handle_unknown='ignore')
kpi_names_encoded = ohe_kpi.fit_transform(df_feat[['Nama KPI']])

num_feats = df_feat[[reg_target, 'lag_1', 'lag_2', 'lag_3', 'rolling_mean_3']].values
X_tab = np.hstack([kpi_names_encoded, num_feats])

sc_kpi_tab = StandardScaler()
X_tab_scaled = sc_kpi_tab.fit_transform(X_tab)

y_kpi_imp = df_feat['TARGET: Impact (1-5)'].astype(int).values - 1
y_kpi_lik = df_feat['TARGET: Likelihood (1-5)'].astype(int).values - 1

kpi_imp_clf = XGBClassifier(eval_metric='mlogloss', random_state=42)
kpi_imp_clf.fit(X_tab_scaled, y_kpi_imp)

kpi_lik_clf = XGBClassifier(eval_metric='mlogloss', random_state=42)
kpi_lik_clf.fit(X_tab_scaled, y_kpi_lik)

# Save LSTM state dict
torch.save(lstm_model.state_dict(), os.path.join(OUTPUT_MODEL_DIR, 'kpi_lstm.pth'))

# Save KPI bundle
kpi_bundle = {
    'scalers': kpi_scalers,
    'global_min_val': global_min_val,
    'global_max_val': global_max_val,
    'ohe_kpi': ohe_kpi,
    'scaler_tab': sc_kpi_tab,
    'impact_classifier': kpi_imp_clf,
    'likelihood_classifier': kpi_lik_clf,
    'seq_length': SEQ_LENGTH,
    'known_kpis': sorted(df_kpi['Nama KPI'].unique().tolist())
}

with open(os.path.join(OUTPUT_MODEL_DIR, 'kpi_bundle.pkl'), 'wb') as f:
    pickle.dump(kpi_bundle, f)

print(" -> Saved kpi_lstm.pth and kpi_bundle.pkl successfully.")
print("\n[SUCCESS] ALL 4 MODEL SUITES TRAINED AND SAVED SUCCESSFULLY!")
