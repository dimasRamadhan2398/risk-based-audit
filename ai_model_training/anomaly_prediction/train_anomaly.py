import os
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder, StandardScaler
from sklearn.model_selection import train_test_split, StratifiedKFold, cross_val_score
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier, ExtraTreesClassifier
from sklearn.svm import SVC
from sklearn.linear_model import LogisticRegression
from sklearn.neural_network import MLPClassifier
from xgboost import XGBClassifier
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score, f1_score, recall_score, precision_score, roc_auc_score

# 1. Load the dataset
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'anomaly_data.csv')
print(f"Loading dataset from: {dataset_path}")
df = pd.read_csv(dataset_path)
initial_rows = len(df)

# ======================================================================
# 1. ADVANCED PREPROCESSING & FEATURE ENGINEERING PIPELINE
# ======================================================================
print("\n======================================================================")
print(" 1. ADVANCED DATA PREPROCESSING & FEATURE ENGINEERING ")
print("======================================================================")

# Step 1: Strip whitespace from string/categorical columns
text_columns = [col for col in ['Entitas', 'Deskripsi', 'TARGET: is_anomaly'] if col in df.columns]
df[text_columns] = df[text_columns].apply(lambda col: col.astype("string").str.strip())
print(f"Stripped surrounding spaces from text columns: {text_columns}")

# Step 2: Convert transaction amount to numeric float
df['amount'] = pd.to_numeric(
    df['amount (dalam Juta Rp)'].astype(str).str.replace(',', '', regex=False),
    errors='coerce'
)

# Step 3: Range Validation & Out-of-Bounds Filtering
numeric_columns = ['hour_of_day (0-23)', 'day_of_week (1-7)', 'is_new_beneficiary (1=Ya, 0=Tidak)', 'is_round_amount (1=Ya, 0=Tidak)']
df[numeric_columns] = df[numeric_columns].apply(pd.to_numeric, errors='coerce')

df.loc[~df['hour_of_day (0-23)'].between(0, 23), 'hour_of_day (0-23)'] = np.nan
df.loc[~df['day_of_week (1-7)'].between(1, 7), 'day_of_week (1-7)'] = np.nan
df.loc[~df['is_new_beneficiary (1=Ya, 0=Tidak)'].isin([0, 1]), 'is_new_beneficiary (1=Ya, 0=Tidak)'] = np.nan
df.loc[~df['is_round_amount (1=Ya, 0=Tidak)'].isin([0, 1]), 'is_round_amount (1=Ya, 0=Tidak)'] = np.nan
df.loc[df['amount'] < 0, 'amount'] = np.nan

# Step 4: Deduplication (Ignore ID Transaksi)
duplicate_columns = [col for col in df.columns if col != 'ID Transaksi']
duplicate_rows = int(df.duplicated(subset=duplicate_columns).sum())
df = df.drop_duplicates(subset=duplicate_columns).reset_index(drop=True)
print(f"Initial dataset rows: {initial_rows} | Duplicates removed: {duplicate_rows} | Clean rows: {len(df)}")

# Step 5: Advanced Feature Engineering
print("Applying Domain-Specific Feature Engineering...")
df['log_amount'] = np.log1p(df['amount'])
df['is_weekend'] = df['day_of_week (1-7)'].isin([6, 7]).astype(int)
df['is_night'] = ((df['hour_of_day (0-23)'] >= 22) | (df['hour_of_day (0-23)'] <= 5)).astype(int)
df['amount_per_hour'] = df['amount'] / (df['hour_of_day (0-23)'] + 1)
df['hour_sin'] = np.sin(2 * np.pi * df['hour_of_day (0-23)'] / 24)
df['hour_cos'] = np.cos(2 * np.pi * df['hour_of_day (0-23)'] / 24)
df['day_sin'] = np.sin(2 * np.pi * (df['day_of_week (1-7)'] - 1) / 7)
df['day_cos'] = np.cos(2 * np.pi * (df['day_of_week (1-7)'] - 1) / 7)

# Features list (Target Leakage Removed)
feature_cols = [
    'Entitas', 'Deskripsi', 'amount', 'hour_of_day (0-23)', 
    'day_of_week (1-7)', 'is_new_beneficiary (1=Ya, 0=Tidak)', 
    'is_round_amount (1=Ya, 0=Tidak)', 'log_amount', 
    'is_weekend', 'is_night', 'amount_per_hour',
    'hour_sin', 'hour_cos', 'day_sin', 'day_cos'
]

X_raw = df[feature_cols].values
print(f"Selected {len(feature_cols)} clean input features.")

# One-Hot Encoding on categorical columns (0: Entitas, 1: Deskripsi)
ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(), [0, 1])], remainder='passthrough')
X = np.array(ct.fit_transform(X_raw))
print(f"Encoded Feature matrix shape: {X.shape}")

# Targets configuration
targets_config = {
    'Anomaly Detection (is_anomaly)': {
        'column': 'TARGET: is_anomaly' if 'TARGET: is_anomaly' in df.columns else 'is_anomaly (Ground Truth)',
        'transform': lambda y: np.where(y == 'Ya (Anomali)', 1, 0),
        'class_names': ['Normal', 'Anomaly'],
        'is_binary': True
    },
    'Impact Prediction (1-5)': {
        'column': 'TARGET: Impact (1-5)',
        'transform': lambda y: y.astype(int) - 1,
        'class_names': ['1', '2', '3', '4', '5'],
        'is_binary': False
    },
    'Likelihood Prediction (1-5)': {
        'column': 'TARGET: Likelihood (1-5)',
        'transform': lambda y: y.astype(int) - 1,
        'class_names': ['1', '2', '3', '4', '5'],
        'is_binary': False
    }
}

# ======================================================================
# 2. MULTI-MODEL BENCHMARK & CLASS IMBALANCE TUNING LOOP
# ======================================================================
for target_name, config in targets_config.items():
    if config['column'] not in df.columns:
        continue

    print(f"\n======================================================================")
    print(f" TRAINING & BENCHMARKING FOR TARGET: {target_name} ")
    print(f"======================================================================")
    
    clean_mask = ~df[config['column']].isna()
    X_clean = X[clean_mask]
    y_raw = df[config['column']].values[clean_mask]
    y = config['transform'](y_raw)
    
    # 80/20 Stratified Train-Test Split
    X_train, X_test, y_train, y_test = train_test_split(
        X_clean, y, test_size=0.2, random_state=42, stratify=y
    )
    
    # Feature Scaling (StandardScaler)
    sc = StandardScaler()
    X_train_scaled = sc.fit_transform(X_train)
    X_test_scaled = sc.transform(X_test)
    
    # Define Candidate Machine Learning Models (with Class Weighting to handle imbalance)
    if config['is_binary']:
        # Calculate ratio of negative to positive for XGBoost scale_pos_weight
        num_neg = (y_train == 0).sum()
        num_pos = max((y_train == 1).sum(), 1)
        scale_pos = num_neg / num_pos
        
        candidate_models = {
            'Logistic Reg (Balanced)': LogisticRegression(class_weight='balanced', max_iter=1000, random_state=42),
            'Extra Trees (Balanced)': ExtraTreesClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'Random Forest (Balanced)': RandomForestClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'XGBoost (Weighted)': XGBClassifier(scale_pos_weight=scale_pos, eval_metric='logloss', random_state=42),
            'SVM (Balanced)': SVC(class_weight='balanced', probability=True, random_state=42),
            'Gradient Boosting': GradientBoostingClassifier(random_state=42),
            'MLP Neural Net': MLPClassifier(hidden_layer_sizes=(64, 32), max_iter=500, random_state=42)
        }
    else:
        candidate_models = {
            'XGBoost': XGBClassifier(eval_metric='mlogloss', random_state=42),
            'Random Forest (Balanced)': RandomForestClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'Extra Trees (Balanced)': ExtraTreesClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'Gradient Boosting': GradientBoostingClassifier(random_state=42),
            'MLP Neural Net': MLPClassifier(hidden_layer_sizes=(64, 32), max_iter=500, random_state=42),
            'SVM (Balanced)': SVC(class_weight='balanced', probability=True, random_state=42),
            'Logistic Reg (Balanced)': LogisticRegression(class_weight='balanced', max_iter=1000, random_state=42)
        }
        
    results = []
    
    for m_name, model in candidate_models.items():
        try:
            model.fit(X_train_scaled, y_train)
            
            if config['is_binary'] and hasattr(model, "predict_proba"):
                # DECISION THRESHOLD TUNING for Anomaly Detection
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
                'y_pred': y_pred
            })
        except Exception as e:
            print(f"Error training {m_name}: {e}")
            
    # Rank models using Primary Evaluation Metric: Anomaly F1-Score (for binary) / Weighted F1 (for multi-class)
    if config['is_binary']:
        results_sorted = sorted(results, key=lambda r: (r['anomaly_f1'], r['anomaly_recall'], r['test_acc']), reverse=True)
    else:
        results_sorted = sorted(results, key=lambda r: (r['weighted_f1'], r['test_acc']), reverse=True)
        
    winner = results_sorted[0]
    
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
            
    print(f"\n[WINNER] Winner Model: {winner['model_name']} (Primary Metric Score: {winner['anomaly_f1']:.4f}, Test Acc: {winner['test_acc']*100:.2f}%)")
    
    # Detailed Classification Report for Winner Model
    print(f"\n--- Detailed Classification Report for Winner ({winner['model_name']}) ---")
    print("Confusion Matrix:")
    print(confusion_matrix(y_test, winner['y_pred']))
    print("\nClassification Report:")
    print(classification_report(y_test, winner['y_pred'], target_names=config['class_names'], zero_division=0))
