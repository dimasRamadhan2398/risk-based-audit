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

# 1. Load Dataset
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'repeat_finding_data.csv')
print(f"Loading Repeat Finding dataset from: {dataset_path}")
df = pd.read_csv(dataset_path)
initial_rows = len(df)

# ======================================================================
# 1. ADVANCED PREPROCESSING & FEATURE ENGINEERING PIPELINE
# ======================================================================
print("\n======================================================================")
print(" 1. REPEAT FINDING DATA PREPROCESSING & FEATURE ENGINEERING ")
print("======================================================================")

# Step 1: Strip surrounding spaces
text_columns = [col for col in ['Entitas Auditee', 'Kategori Risiko', 'severity_category', 'Deskripsi Temuan', 'TARGET: is_anomaly'] if col in df.columns]
df[text_columns] = df[text_columns].apply(lambda col: col.astype("string").str.strip())

# Step 2: Convert numeric columns
numeric_cols = [
    'repeat_count_last_year', 'finding_similarity_pct', 
    'is_same_root_cause (1=Ya, 0=Tidak)', 'is_previously_closed (1=Ya, 0=Tidak)'
]
df[numeric_cols] = df[numeric_cols].apply(pd.to_numeric, errors='coerce')

# Step 3: Deduplication
duplicate_cols = [col for col in df.columns if col != 'ID Temuan']
df = df.drop_duplicates(subset=duplicate_cols).reset_index(drop=True)
print(f"Initial rows: {initial_rows} | Clean rows after deduplication: {len(df)}")

# Step 4: Advanced Repeat Finding Domain Feature Engineering
print("Applying Domain-Specific Audit Finding Feature Engineering...")
severity_map = {
    'Very Significant': 4,
    'Significant': 3,
    'Quite Significant': 2,
    'Not Significant': 1
}
df['severity_score'] = df['severity_category'].map(severity_map).fillna(1)
df['log_repeat_count'] = np.log1p(df['repeat_count_last_year'])
df['repeat_similarity_composite'] = df['repeat_count_last_year'] * df['finding_similarity_pct']
df['risk_composite_score'] = (df['is_same_root_cause (1=Ya, 0=Tidak)'] * 3.0) + \
                             (df['is_previously_closed (1=Ya, 0=Tidak)'] * 2.5) + \
                             (df['repeat_count_last_year'] * 1.2) + \
                             df['severity_score']

# Input Feature Columns List
feature_cols = [
    'Entitas Auditee', 'Kategori Risiko', 'severity_category', 'Deskripsi Temuan',
    'repeat_count_last_year', 'finding_similarity_pct', 
    'is_same_root_cause (1=Ya, 0=Tidak)', 'is_previously_closed (1=Ya, 0=Tidak)', 
    'log_repeat_count', 'repeat_similarity_composite', 'risk_composite_score'
]

X_raw = df[feature_cols].values
print(f"Selected {len(feature_cols)} clean repeat finding input features.")

# One-Hot Encoding for categorical columns (0: Entitas Auditee, 1: Kategori Risiko, 2: severity_category, 3: Deskripsi Temuan)
ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(sparse_output=False, handle_unknown='ignore'), [0, 1, 2, 3])], remainder='passthrough')
X = np.array(ct.fit_transform(X_raw))
print(f"Encoded Feature matrix shape: {X.shape}")

# Targets configuration
targets_config = {
    'Repeat Finding Anomaly Detection (is_anomaly)': {
        'column': 'TARGET: is_anomaly',
        'transform': lambda y: np.where(y == 'Ya (Anomali)', 1, 0),
        'class_names': ['Normal Finding', 'Repeat Finding Anomaly'],
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
# 2. MULTI-MODEL BENCHMARK & THRESHOLD TUNING LOOP
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
    
    if config['is_binary']:
        y = config['transform'](y_raw)
        target_classes = [0, 1]
    else:
        le = LabelEncoder()
        y = le.fit_transform(config['transform'](y_raw))
        target_classes = list(range(len(le.classes_)))
        target_class_names = [str(c) for c in le.classes_]
    
    # 80/20 Stratified Train-Test Split
    X_train, X_test, y_train, y_test = train_test_split(
        X_clean, y, test_size=0.2, random_state=42, stratify=y
    )
    
    # Feature Scaling
    sc = StandardScaler()
    X_train_scaled = sc.fit_transform(X_train)
    X_test_scaled = sc.transform(X_test)
    
    if config['is_binary']:
        num_neg = (y_train == 0).sum()
        num_pos = max((y_train == 1).sum(), 1)
        scale_pos = num_neg / num_pos
        
        candidate_models = {
            'XGBoost (Weighted)': XGBClassifier(scale_pos_weight=scale_pos, eval_metric='logloss', random_state=42),
            'Gradient Boosting': GradientBoostingClassifier(random_state=42),
            'Random Forest (Balanced)': RandomForestClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'Extra Trees (Balanced)': ExtraTreesClassifier(class_weight='balanced', n_estimators=200, random_state=42),
            'MLP Neural Net': MLPClassifier(hidden_layer_sizes=(64, 32), max_iter=500, random_state=42),
            'Logistic Reg (Balanced)': LogisticRegression(class_weight='balanced', max_iter=1000, random_state=42)
        }
    else:
        candidate_models = {
            'XGBoost': XGBClassifier(eval_metric='mlogloss', random_state=42),
            'Gradient Boosting': GradientBoostingClassifier(random_state=42),
            'Random Forest (Balanced)': RandomForestClassifier(class_weight='balanced', n_estimators=200, random_state=42),
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
                'y_pred': y_pred
            })
        except Exception as e:
            print(f"Error training {m_name}: {e}")
            
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
            
    print(f"\n[WINNER] Best Model: {winner['model_name']} (Accuracy: {winner['test_acc']*100:.2f}%)")
    
    print(f"\n--- Classification Report for Winner ({winner['model_name']}) ---")
    print(confusion_matrix(y_test, winner['y_pred']))
    names = config['class_names'] if config['is_binary'] else target_class_names
    print(classification_report(y_test, winner['y_pred'], labels=target_classes, target_names=names, zero_division=0))
