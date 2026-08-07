import os
import pickle
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder, StandardScaler
from xgboost import XGBClassifier

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR = os.path.abspath(os.path.join(CURRENT_DIR, '..', '..'))
OUTPUT_MODEL_DIR = os.path.join(ROOT_DIR, 'backend', 'python-ai', 'models')
os.makedirs(OUTPUT_MODEL_DIR, exist_ok=True)

dept_csv = os.path.join(CURRENT_DIR, 'department_dataset.csv')
print(f"Loading dataset from: {dept_csv}")
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
ct_dept = ColumnTransformer(transformers=[('encoder', OneHotEncoder(handle_unknown='ignore', sparse_output=False), [0, 1])], remainder='passthrough')
X_dept = ct_dept.fit_transform(X_raw_dept)

sc_dept = StandardScaler()
X_dept_scaled = sc_dept.fit_transform(X_dept)

le_dept_imp = LabelEncoder()
y_dept_imp = le_dept_imp.fit_transform(df_dept['target_impact'])

le_dept_lik = LabelEncoder()
y_dept_lik = le_dept_lik.fit_transform(df_dept['target_likelihood'])

print("Training high-performance XGBoost models for Impact & Likelihood...")
dept_imp_clf = XGBClassifier(n_estimators=300, max_depth=5, learning_rate=0.05, eval_metric='mlogloss', random_state=42)
dept_imp_clf.fit(X_dept_scaled, y_dept_imp)

dept_lik_clf = XGBClassifier(n_estimators=300, max_depth=5, learning_rate=0.05, eval_metric='mlogloss', random_state=42)
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

output_bundle_path = os.path.join(OUTPUT_MODEL_DIR, 'department_bundle.pkl')
with open(output_bundle_path, 'wb') as f:
    pickle.dump(dept_bundle, f)

print(f" -> Successfully saved ONLY department_bundle.pkl to: {output_bundle_path}")
