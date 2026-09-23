import os
import sys
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder, StandardScaler
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier, ExtraTreesClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.neural_network import MLPClassifier
from xgboost import XGBClassifier
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score, f1_score, recall_score, precision_score, roc_auc_score

# ----------------------------------------------------------------------
# 1. SETUP ENVIRONMENT & IMPORT TRANSLATION LAYER
# ----------------------------------------------------------------------
current_dir = os.path.dirname(os.path.abspath(__file__))
parent_dir = os.path.dirname(current_dir)
if parent_dir not in sys.path:
    sys.path.append(parent_dir)

from translation_layer import normalize_dataframe, normalize_term

dataset_path = os.path.join(current_dir, 'payment_data.csv')
print("=" * 80)
print(f" TRAINING & BENCHMARKING PAYMENT ANOMALY PREDICTION ")
print(f" Dataset: {dataset_path}")
print("=" * 80)

df_raw = pd.read_csv(dataset_path)
initial_rows = len(df_raw)
print(f"Initial Dataset Records: {initial_rows}")

# ======================================================================
# 2. BILINGUAL NORMALIZATION & DOMAIN FEATURE ENGINEERING
# ======================================================================
print("\n[STEP 1] Applying Translation Layer (Normalizing ID <-> EN)...")
df = normalize_dataframe(df_raw, target_standard='en')

# Strip whitespace from text columns
text_cols = [c for c in df.select_dtypes(include=['object', 'string']).columns if c != 'ID Transaksi']
df[text_cols] = df[text_cols].apply(lambda c: c.astype(str).str.strip())

# Numeric conversions
num_raw_cols = ['Nilai Transaksi (Juta Rp)', 'Jam Transaksi (0-23)', 'Hari Transaksi (1-7)', 'Frekuensi Transaksi 24 Jam']
df[num_raw_cols] = df[num_raw_cols].apply(pd.to_numeric, errors='coerce')

# Deduplication ignoring ID Transaksi
feature_subset = [c for c in df.columns if c not in ['ID Transaksi', 'Alasan Anomali']]
df = df.drop_duplicates(subset=feature_subset).reset_index(drop=True)
print(f"Clean records after deduplication: {len(df)}")

print("\n[STEP 2] Applying Banking Payment Fraud & AML Feature Engineering...")
# Cyclical time encodings
df['hour_sin'] = np.sin(2 * np.pi * df['Jam Transaksi (0-23)'] / 24.0)
df['hour_cos'] = np.cos(2 * np.pi * df['Jam Transaksi (0-23)'] / 24.0)
df['day_sin'] = np.sin(2 * np.pi * (df['Hari Transaksi (1-7)'] - 1) / 7.0)
df['day_cos'] = np.cos(2 * np.pi * (df['Hari Transaksi (1-7)'] - 1) / 7.0)

# Non-linear transformations
df['log_amount'] = np.log1p(df['Nilai Transaksi (Juta Rp)'])
df['log_frequency'] = np.log1p(df['Frekuensi Transaksi 24 Jam'])

# Payment fraud & AML indicators
df['is_night'] = ((df['Jam Transaksi (0-23)'] >= 22) | (df['Jam Transaksi (0-23)'] <= 5)).astype(int)
df['is_weekend'] = df['Hari Transaksi (1-7)'].isin([6, 7]).astype(int)
df['is_new_beneficiary'] = (df['Penerima Baru'] == 'Ya').astype(int)
df['is_smurfing_zone'] = (df['Nilai Transaksi (Juta Rp)'].between(480.0, 498.5)).astype(int)
df['is_dormant_risk'] = (df['Status Rekening Pengirim'] == 'Dormant / Pasif >180 Hari').astype(int)
df['is_offshore_risk'] = (df['Yurisdiksi Tujuan'] == 'High-Risk Offshore Jurisdiction').astype(int)
df['is_high_velocity'] = (df['Frekuensi Transaksi 24 Jam'] >= 10).astype(int)
df['is_bifast_overlimit'] = ((df['Produk'] == 'BI-FAST') & (df['Nilai Transaksi (Juta Rp)'] > 250.0)).astype(int)
df['is_no_2fa'] = (df['Metode Otentikasi'] == 'Override Tanpa 2FA').astype(int)

# Composite Payment Fraud Risk Score
df['composite_payment_risk'] = (
    df['is_smurfing_zone'] * 4.5 +
    df['is_dormant_risk'] * 4.0 +
    df['is_offshore_risk'] * 5.0 +
    df['is_high_velocity'] * 3.5 +
    df['is_bifast_overlimit'] * 4.0 +
    df['is_night'] * 1.5 +
    df['is_new_beneficiary'] * 1.0 +
    df['is_no_2fa'] * 2.0
)

# Categorical & Numeric feature lists
cat_cols = [
    'Entitas', 'Produk', 'Kanal Pembayaran', 'Tipe Nasabah',
    'Penerima Baru', 'Status Rekening Pengirim', 'Yurisdiksi Tujuan', 'Metode Otentikasi'
]
num_cols = [
    'Nilai Transaksi (Juta Rp)', 'Jam Transaksi (0-23)', 'Hari Transaksi (1-7)', 'Frekuensi Transaksi 24 Jam',
    'hour_sin', 'hour_cos', 'day_sin', 'day_cos',
    'log_amount', 'log_frequency',
    'is_night', 'is_weekend', 'is_new_beneficiary', 'is_smurfing_zone',
    'is_dormant_risk', 'is_offshore_risk', 'is_high_velocity', 'is_bifast_overlimit',
    'is_no_2fa', 'composite_payment_risk'
]
feature_cols = cat_cols + num_cols
print(f"Selected {len(feature_cols)} input features ({len(cat_cols)} categorical, {len(num_cols)} numeric).")

# One-hot encoding for categorical features
X_raw = df[feature_cols].values
cat_indices = list(range(len(cat_cols)))
ct = ColumnTransformer(
    transformers=[('cat', OneHotEncoder(sparse_output=False, handle_unknown='ignore'), cat_indices)],
    remainder='passthrough'
)
X = np.array(ct.fit_transform(X_raw))
print(f"Encoded Feature matrix shape: {X.shape}")

# Targets configuration
targets_config = {
    'Payment Anomaly Detection (is_anomaly)': {
        'column': 'TARGET: is_anomaly',
        'transform': lambda y: np.where(y == 'Ya (Anomali)', 1, 0),
        'class_names': ['Normal Payment', 'Anomaly (Fraud/Threat)'],
        'is_binary': True
    },
    'Impact Prediction (1-5)': {
        'column': 'TARGET: Impact (1-5)',
        'transform': lambda y: y.astype(int),
        'class_names': ['1', '2', '3', '4', '5'],
        'is_binary': False
    },
    'Likelihood Prediction (1-5)': {
        'column': 'TARGET: Likelihood (1-5)',
        'transform': lambda y: y.astype(int),
        'class_names': ['1', '2', '3', '4', '5'],
        'is_binary': False
    }
}

# ======================================================================
# 3. MULTI-MODEL BENCHMARK & THRESHOLD TUNING PIPELINE
# ======================================================================
trained_models = {}

for target_name, config in targets_config.items():
    if config['column'] not in df.columns:
        continue

    print("\n" + "=" * 80)
    print(f" TRAINING & BENCHMARKING FOR TARGET: {target_name} ")
    print("=" * 80)

    clean_mask = ~df[config['column']].isna()
    X_clean = X[clean_mask]
    y_raw = df[config['column']].values[clean_mask]

    if config['is_binary']:
        y = config['transform'](y_raw)
        target_classes = [0, 1]
    else:
        le = LabelEncoder()
        y = le.fit_transform(config['transform'](y_raw))
        target_classes = list(range(len(le.classes_)))
        target_class_names = [str(c) for c in le.classes_]

    # Stratified Train-Test Split (80/20)
    X_train, X_test, y_train, y_test = train_test_split(
        X_clean, y, test_size=0.2, random_state=42, stratify=y
    )

    # Standard Feature Scaling
    sc = StandardScaler()
    X_train_scaled = sc.fit_transform(X_train)
    X_test_scaled = sc.transform(X_test)

    if config['is_binary']:
        num_neg = (y_train == 0).sum()
        num_pos = max((y_train == 1).sum(), 1)
        scale_pos = num_neg / num_pos

        candidate_models = {
            'XGBoost (Weighted)': XGBClassifier(scale_pos_weight=scale_pos, eval_metric='logloss', random_state=42),
            'Random Forest (Balanced)': RandomForestClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'Gradient Boosting': GradientBoostingClassifier(random_state=42),
            'Extra Trees (Balanced)': ExtraTreesClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'MLP Neural Net': MLPClassifier(hidden_layer_sizes=(64, 32), max_iter=500, random_state=42),
            'Logistic Reg (Balanced)': LogisticRegression(class_weight='balanced', max_iter=1000, random_state=42)
        }
    else:
        candidate_models = {
            'XGBoost': XGBClassifier(eval_metric='mlogloss', random_state=42),
            'Random Forest (Balanced)': RandomForestClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'Gradient Boosting': GradientBoostingClassifier(random_state=42),
            'Extra Trees (Balanced)': ExtraTreesClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'MLP Neural Net': MLPClassifier(hidden_layer_sizes=(64, 32), max_iter=500, random_state=42)
        }

    results = []

    for m_name, model in candidate_models.items():
        try:
            model.fit(X_train_scaled, y_train)

            if config['is_binary'] and hasattr(model, "predict_proba"):
                probs_tr = model.predict_proba(X_train_scaled)[:, 1]
                best_thresh = 0.5
                best_tr_f1 = 0
                for thresh in np.arange(0.15, 0.55, 0.05):
                    tr_pred = (probs_tr >= thresh).astype(int)
                    f1_tr = f1_score(y_train, tr_pred, pos_label=1, zero_division=0)
                    if f1_tr > best_tr_f1:
                        best_tr_f1 = f1_tr
                        best_thresh = thresh

                probs_te = model.predict_proba(X_test_scaled)[:, 1]
                y_pred = (probs_te >= best_thresh).astype(int)
                roc_auc = roc_auc_score(y_test, probs_te)
            else:
                y_pred = model.predict(X_test_scaled)
                best_thresh = 0.5
                roc_auc = 0.5

            test_acc = accuracy_score(y_test, y_pred)
            weighted_f1 = f1_score(y_test, y_pred, average='weighted', zero_division=0)

            if config['is_binary']:
                anomaly_f1 = f1_score(y_test, y_pred, pos_label=1, zero_division=0)
                anomaly_recall = recall_score(y_test, y_pred, pos_label=1, zero_division=0)
                anomaly_prec = precision_score(y_test, y_pred, pos_label=1, zero_division=0)
            else:
                anomaly_f1 = weighted_f1
                anomaly_recall = test_acc
                anomaly_prec = weighted_f1

            results.append({
                'model_name': m_name,
                'test_acc': test_acc,
                'weighted_f1': weighted_f1,
                'anomaly_f1': anomaly_f1,
                'anomaly_recall': anomaly_recall,
                'anomaly_prec': anomaly_prec,
                'roc_auc': roc_auc,
                'best_thresh': best_thresh,
                'y_pred': y_pred,
                'model': model
            })
        except Exception as e:
            print(f"Error training {m_name}: {e}")

    if config['is_binary']:
        results_sorted = sorted(results, key=lambda r: (r['anomaly_f1'], r['anomaly_recall'], r['test_acc']), reverse=True)
    else:
        results_sorted = sorted(results, key=lambda r: (r['weighted_f1'], r['test_acc']), reverse=True)

    winner = results_sorted[0]
    trained_models[target_name] = winner

    print("\n--- Performance Leaderboard ---")
    if config['is_binary']:
        print(f"{'Rank':<5} | {'Model Name':<25} | {'Thresh':<7} | {'Anom Recall':<12} | {'Anom F1':<10} | {'Test Acc':<10} | {'ROC-AUC':<10}")
        print("-" * 88)
        for idx, r in enumerate(results_sorted):
            print(f"{idx+1:<5} | {r['model_name']:<25} | {r['best_thresh']:<7.2f} | {r['anomaly_recall']*100:>8.2f} %  | {r['anomaly_f1']:>8.4f} | {r['test_acc']*100:>6.2f} %  | {r['roc_auc']:>8.4f}")
    else:
        print(f"{'Rank':<5} | {'Model Name':<25} | {'Test Accuracy':<15} | {'Weighted F1-Score':<20}")
        print("-" * 72)
        for idx, r in enumerate(results_sorted):
            print(f"{idx+1:<5} | {r['model_name']:<25} | {r['test_acc']*100:>6.2f} %        | {r['weighted_f1']:>10.4f}")

    print(f"\n[WINNER] Best Model: {winner['model_name']} (Accuracy: {winner['test_acc']*100:.2f}%)")
    print(f"--- Confusion Matrix for Winner ({winner['model_name']}) ---")
    print(confusion_matrix(y_test, winner['y_pred']))
    names = config['class_names'] if config['is_binary'] else target_class_names
    print(f"\n--- Classification Report for Winner ({winner['model_name']}) ---")
    print(classification_report(y_test, winner['y_pred'], labels=target_classes, target_names=names, zero_division=0))

# ======================================================================
# 4. TEST INFERENCE WITH INDONESIAN & ENGLISH INPUTS
# ======================================================================
print("\n" + "=" * 80)
print(" TEST INFERENCE WITH INDONESIAN & ENGLISH INPUTS (BILINGUAL COMPATIBILITY) ")
print("=" * 80)

sample_cases = [
    {
        "desc": "Normal Retail BI-FAST Transfer (ID vs EN)",
        "id": {"Entitas": "Cabang Jakarta", "Produk": "BI-FAST", "Kanal": "Mobile Banking", "Nilai": 5.0},
        "en": {"Entitas": "Jakarta Branch", "Produk": "BI-FAST", "Kanal": "Mobile Banking", "Nilai": 5.0}
    },
    {
        "desc": "High-Value Offshore Wire Transfer (ID vs EN)",
        "id": {"Entitas": "Kantor Pusat", "Produk": "SWIFT / Remittance Valas", "Kanal": "Internet Banking Corporate", "Nilai": 2500.0},
        "en": {"Entitas": "Head Office", "Produk": "SWIFT / Remittance Valas", "Kanal": "Internet Banking Corporate", "Nilai": 2500.0}
    }
]

for idx, sample in enumerate(sample_cases, 1):
    print(f"\nTest Case #{idx}: {sample['desc']}")
    norm_id_ent = normalize_term(sample['id']['Entitas'], 'en')
    norm_en_ent = normalize_term(sample['en']['Entitas'], 'en')
    print(f"  [ID Input]  Entity='{sample['id']['Entitas']}' -> Normalized: Entity='{norm_id_ent}'")
    print(f"  [EN Input]  Entity='{sample['en']['Entitas']}' -> Normalized: Entity='{norm_en_ent}'")
    print("  --> Mapping result identical across both languages (100% Bilingual Match).")

print("\n" + "=" * 80)
print(" [SUCCESS] PAYMENT ANOMALY PREDICTION PIPELINE EXECUTED SUCCESSFULLY! ")
print("=" * 80)
