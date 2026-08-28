import os
import numpy as np
import pandas as pd
from sklearn.pipeline import Pipeline
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder, StandardScaler
from sklearn.model_selection import RandomizedSearchCV, cross_val_predict, train_test_split, StratifiedKFold, cross_val_score, RepeatedStratifiedKFold
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier, ExtraTreesClassifier
from sklearn.svm import SVC
from sklearn.linear_model import LogisticRegression
from sklearn.neural_network import MLPClassifier
from xgboost import XGBClassifier
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score, f1_score, recall_score, precision_score, roc_auc_score, average_precision_score

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
            'XGBoost (Weighted)': XGBClassifier(scale_pos_weight=scale_pos, eval_metric='logloss', random_state=42, n_jobs=1),
            'SVM (Balanced)': SVC(class_weight='balanced', probability=True, random_state=42),
            'Gradient Boosting': GradientBoostingClassifier(random_state=42),
            'MLP Neural Net': MLPClassifier(hidden_layer_sizes=(64, 32), max_iter=500, random_state=42)
        }
    else:
        candidate_models = {
            'XGBoost': XGBClassifier(eval_metric='mlogloss', random_state=42, n_jobs=1),
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
    
# ======================================================================
# IMPROVED ANOMALY BENCHMARK
# Leakage-safe preprocessing + out-of-fold threshold tuning
# ======================================================================

print("\n" + "=" * 78)
print(" IMPROVED ANOMALY DETECTION BENCHMARK ")
print("=" * 78)

# Raw features
X_anomaly = df[feature_cols].copy()

anomaly_column = (
    'TARGET: is_anomaly'
    if 'TARGET: is_anomaly' in df.columns
    else 'is_anomaly (Ground Truth)'
)

clean_mask = ~df[anomaly_column].isna()
X_anomaly = X_anomaly.loc[clean_mask].reset_index(drop=True)
y_anomaly_raw = df.loc[clean_mask, anomaly_column].values
y_anomaly = np.where(y_anomaly_raw == 'Ya (Anomali)', 1, 0)

# Train / test split
X_train_anom, X_test_anom, y_train_anom, y_test_anom = train_test_split(
    X_anomaly, y_anomaly,
    test_size=0.20,
    random_state=42,
    stratify=y_anomaly
)

categorical_features = ['Entitas', 'Deskripsi']
numeric_features = [col for col in feature_cols if col not in categorical_features]


def build_anomaly_pipeline(model):
    preprocessor = ColumnTransformer(
        transformers=[
            ('categorical', OneHotEncoder(handle_unknown='ignore', sparse_output=False), categorical_features),
            ('numeric', 'passthrough', numeric_features)
        ],
        remainder='drop'
    )

    return Pipeline([
        ('preprocessor', preprocessor),
        ('scaler', StandardScaler()),
        ('model', model)
    ])


# Class imbalance
num_neg = (y_train_anom == 0).sum()
num_pos = max((y_train_anom == 1).sum(), 1)
scale_pos = num_neg / num_pos

improved_models = {
    'Gradient Boosting': GradientBoostingClassifier(random_state=42),

    'Random Forest (Balanced)': RandomForestClassifier(
        class_weight='balanced',
        n_estimators=300,
        random_state=42,
        n_jobs=1
    ),

    'Extra Trees (Balanced)': ExtraTreesClassifier(
        class_weight='balanced',
        n_estimators=300,
        random_state=42,
        n_jobs=1
    ),

    'XGBoost (Weighted)': XGBClassifier(
        scale_pos_weight=scale_pos,
        eval_metric='logloss',
        random_state=42,
        n_jobs=1
    ),

    'Logistic Reg (Balanced)': LogisticRegression(
        class_weight='balanced',
        max_iter=1000,
        random_state=42
    )
}

skf_anom = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)
improved_results = []


for model_name, model in improved_models.items():
    print(f"\nEvaluating improved model: {model_name}")

    pipeline = build_anomaly_pipeline(model)

    # Out-of-fold probability predictions
    oof_probs = cross_val_predict(
        pipeline,
        X_train_anom,
        y_train_anom,
        cv=skf_anom,
        method='predict_proba',
        n_jobs=1
    )[:, 1]

    # Select threshold only from OOF predictions
    best_threshold = 0.50
    best_oof_f1 = -1

    for threshold in np.arange(0.10, 0.81, 0.01):
        oof_pred = (oof_probs >= threshold).astype(int)

        current_f1 = f1_score(
            y_train_anom,
            oof_pred,
            pos_label=1,
            zero_division=0
        )

        if current_f1 > best_oof_f1:
            best_oof_f1 = current_f1
            best_threshold = threshold

    # Final model fit
    pipeline.fit(X_train_anom, y_train_anom)

    test_probs = pipeline.predict_proba(X_test_anom)[:, 1]
    test_pred = (test_probs >= best_threshold).astype(int)

    # Metrics
    test_accuracy = accuracy_score(y_test_anom, test_pred)

    anomaly_precision = precision_score(
        y_test_anom,
        test_pred,
        pos_label=1,
        zero_division=0
    )

    anomaly_recall = recall_score(
        y_test_anom,
        test_pred,
        pos_label=1,
        zero_division=0
    )

    anomaly_f1 = f1_score(
        y_test_anom,
        test_pred,
        pos_label=1,
        zero_division=0
    )

    roc_auc = roc_auc_score(y_test_anom, test_probs)
    pr_auc = average_precision_score(y_test_anom, test_probs)

    improved_results.append({
        'model': model_name,
        'threshold': best_threshold,
        'oof_f1': best_oof_f1,
        'precision': anomaly_precision,
        'recall': anomaly_recall,
        'f1': anomaly_f1,
        'accuracy': test_accuracy,
        'roc_auc': roc_auc,
        'pr_auc': pr_auc,
        'prediction': test_pred
    })


# Leaderboard
improved_results = sorted(
    improved_results,
    key=lambda r: (r['oof_f1'], r['pr_auc']),
    reverse=True
)

print("\n--- Improved Anomaly Leaderboard ---")

print(
    f"{'Rank':<5} | {'Model':<27} | {'Threshold':<9} | {'OOF F1':<8} | "
    f"{'Test F1':<8} | {'Recall':<8} | {'Precision':<9} | "
    f"{'PR-AUC':<8} | {'ROC-AUC':<8}"
)

print("-" * 115)

for rank, result in enumerate(improved_results, start=1):
    print(
        f"{rank:<5} | {result['model']:<27} | "
        f"{result['threshold']:<9.2f} | "
        f"{result['oof_f1']:<8.4f} | "
        f"{result['f1']:<8.4f} | "
        f"{result['recall']:<8.4f} | "
        f"{result['precision']:<9.4f} | "
        f"{result['pr_auc']:<8.4f} | "
        f"{result['roc_auc']:<8.4f}"
    )


winner_anomaly = improved_results[0]

print("\n[IMPROVED ANOMALY WINNER]")
print(f"Model          : {winner_anomaly['model']}")
print(f"OOF Threshold  : {winner_anomaly['threshold']:.2f}")
print(f"OOF F1         : {winner_anomaly['oof_f1']:.4f}")
print(f"Test F1        : {winner_anomaly['f1']:.4f}")
print(f"Test Recall    : {winner_anomaly['recall']:.4f}")
print(f"Test Precision : {winner_anomaly['precision']:.4f}")
print(f"Test Accuracy  : {winner_anomaly['accuracy'] * 100:.2f}%")
print(f"PR-AUC         : {winner_anomaly['pr_auc']:.4f}")
print(f"ROC-AUC        : {winner_anomaly['roc_auc']:.4f}")

print("\nConfusion Matrix:")
print(confusion_matrix(y_test_anom, winner_anomaly['prediction']))

# ======================================================================
# FINE TUNING XGBOOST ANOMALY
# Hyperparameter selection by CV PR-AUC + OOF threshold optimization
# ======================================================================

print("\n" + "=" * 78)
print(" FINE TUNING: XGBOOST WEIGHTED - ANOMALY ")
print("=" * 78)

xgb_anomaly_params = {
    'model__n_estimators': [150, 200, 300, 400, 500, 700],
    'model__max_depth': [2, 3, 4, 5, 6, 7],
    'model__learning_rate': [0.01, 0.03, 0.05, 0.08, 0.10],
    'model__min_child_weight': [1, 2, 3, 5, 8],
    'model__subsample': [0.70, 0.80, 0.90, 1.00],
    'model__colsample_bytree': [0.70, 0.80, 0.90, 1.00],
    'model__gamma': [0.0, 0.05, 0.10, 0.20, 0.30],
    'model__reg_alpha': [0.0, 0.01, 0.05, 0.10, 0.20],
    'model__reg_lambda': [0.5, 1.0, 2.0, 3.0, 5.0],
    'model__scale_pos_weight': [
        scale_pos * 0.75,
        scale_pos,
        scale_pos * 1.25,
        scale_pos * 1.50
    ]
}

xgb_tuning_pipeline = build_anomaly_pipeline(
    XGBClassifier(
        objective='binary:logistic',
        eval_metric='logloss',
        random_state=42,
        n_jobs=1
    )
)

xgb_search = RandomizedSearchCV(
    estimator=xgb_tuning_pipeline,
    param_distributions=xgb_anomaly_params,
    n_iter=40,
    scoring='average_precision',
    cv=skf_anom,
    random_state=42,
    n_jobs=1,
    verbose=1,
    refit=True
)

xgb_search.fit(X_train_anom, y_train_anom)

print(f"\nBest CV PR-AUC: {xgb_search.best_score_:.4f}")
print("\nBest Parameters:")
for key, value in xgb_search.best_params_.items():
    print(f"  {key}: {value}")
    
# ----------------------------------------------------------------------
# OOF threshold optimization for tuned XGBoost
# ----------------------------------------------------------------------

best_xgb_pipeline = xgb_search.best_estimator_

xgb_oof_probs = cross_val_predict(
    best_xgb_pipeline, X_train_anom, y_train_anom,
    cv=skf_anom, method='predict_proba', n_jobs=1
)[:, 1]

best_xgb_threshold = 0.50
best_xgb_oof_f1 = -1

for threshold in np.arange(0.10, 0.81, 0.01):
    xgb_oof_pred = (xgb_oof_probs >= threshold).astype(int)
    current_f1 = f1_score(y_train_anom, xgb_oof_pred, pos_label=1, zero_division=0)

    if current_f1 > best_xgb_oof_f1:
        best_xgb_oof_f1 = current_f1
        best_xgb_threshold = threshold

print(f"\nBest XGB OOF Threshold : {best_xgb_threshold:.2f}")
print(f"Best XGB OOF F1        : {best_xgb_oof_f1:.4f}")

# ----------------------------------------------------------------------
# Held-out test evaluation - XGBoost
# ----------------------------------------------------------------------

best_xgb_pipeline.fit(X_train_anom, y_train_anom)

xgb_test_probs = best_xgb_pipeline.predict_proba(X_test_anom)[:, 1]
xgb_test_pred = (xgb_test_probs >= best_xgb_threshold).astype(int)

xgb_tuned_accuracy = accuracy_score(y_test_anom, xgb_test_pred)
xgb_tuned_precision = precision_score(y_test_anom, xgb_test_pred, zero_division=0)
xgb_tuned_recall = recall_score(y_test_anom, xgb_test_pred, zero_division=0)
xgb_tuned_f1 = f1_score(y_test_anom, xgb_test_pred, zero_division=0)
xgb_tuned_pr_auc = average_precision_score(y_test_anom, xgb_test_probs)
xgb_tuned_roc_auc = roc_auc_score(y_test_anom, xgb_test_probs)

print("\n" + "=" * 78)
print(" TUNED XGBOOST TEST PERFORMANCE ")
print("=" * 78)

print(f"Threshold      : {best_xgb_threshold:.2f}")
print(f"Accuracy       : {xgb_tuned_accuracy * 100:.2f}%")
print(f"Anomaly F1     : {xgb_tuned_f1:.4f}")
print(f"Anomaly Recall : {xgb_tuned_recall:.4f}")
print(f"Precision      : {xgb_tuned_precision:.4f}")
print(f"PR-AUC         : {xgb_tuned_pr_auc:.4f}")
print(f"ROC-AUC        : {xgb_tuned_roc_auc:.4f}")

print("\nConfusion Matrix:")
print(confusion_matrix(y_test_anom, xgb_test_pred))
    
# ======================================================================
# FINE TUNING GRADIENT BOOSTING - ANOMALY
# ======================================================================

print("\n" + "=" * 78)
print(" FINE TUNING: GRADIENT BOOSTING - ANOMALY ")
print("=" * 78)

gb_anomaly_params = {
    'model__n_estimators': [100, 150, 200, 300, 400, 500],
    'model__learning_rate': [0.01, 0.03, 0.05, 0.08, 0.10],
    'model__max_depth': [2, 3, 4, 5],
    'model__min_samples_split': [2, 3, 5, 8, 10],
    'model__min_samples_leaf': [1, 2, 3, 4, 5],
    'model__subsample': [0.70, 0.80, 0.90, 1.00],
    'model__max_features': [None, 'sqrt', 'log2']
}

gb_tuning_pipeline = build_anomaly_pipeline(
    GradientBoostingClassifier(random_state=42)
)

gb_search = RandomizedSearchCV(
    estimator=gb_tuning_pipeline,
    param_distributions=gb_anomaly_params,
    n_iter=40,
    scoring='average_precision',
    cv=skf_anom,
    random_state=42,
    n_jobs=1,
    verbose=1,
    refit=True
)

gb_search.fit(X_train_anom, y_train_anom)

print(f"\nBest CV PR-AUC: {gb_search.best_score_:.4f}")
print("\nBest Parameters:")
for key, value in gb_search.best_params_.items():
    print(f"  {key}: {value}")

# ----------------------------------------------------------------------
# OOF threshold optimization for tuned Gradient Boosting
# ----------------------------------------------------------------------

best_gb_pipeline = gb_search.best_estimator_

gb_oof_probs = cross_val_predict(
    best_gb_pipeline, X_train_anom, y_train_anom,
    cv=skf_anom, method='predict_proba', n_jobs=1
)[:, 1]

best_gb_threshold = 0.50
best_gb_oof_f1 = -1

for threshold in np.arange(0.10, 0.81, 0.01):
    oof_pred = (gb_oof_probs >= threshold).astype(int)
    current_f1 = f1_score(y_train_anom, oof_pred, pos_label=1, zero_division=0)

    if current_f1 > best_gb_oof_f1:
        best_gb_oof_f1 = current_f1
        best_gb_threshold = threshold

print(f"\nBest GB OOF Threshold : {best_gb_threshold:.2f}")
print(f"Best GB OOF F1        : {best_gb_oof_f1:.4f}")

# ----------------------------------------------------------------------
# Held-out test evaluation - Gradient Boosting
# ----------------------------------------------------------------------

best_gb_pipeline.fit(X_train_anom, y_train_anom)

gb_test_probs = best_gb_pipeline.predict_proba(X_test_anom)[:, 1]
gb_test_pred = (gb_test_probs >= best_gb_threshold).astype(int)

gb_tuned_accuracy = accuracy_score(y_test_anom, gb_test_pred)
gb_tuned_precision = precision_score(y_test_anom, gb_test_pred, zero_division=0)
gb_tuned_recall = recall_score(y_test_anom, gb_test_pred, zero_division=0)
gb_tuned_f1 = f1_score(y_test_anom, gb_test_pred, zero_division=0)
gb_tuned_pr_auc = average_precision_score(y_test_anom, gb_test_probs)
gb_tuned_roc_auc = roc_auc_score(y_test_anom, gb_test_probs)

print("\n" + "=" * 78)
print(" TUNED GRADIENT BOOSTING TEST PERFORMANCE ")
print("=" * 78)

print(f"Threshold      : {best_gb_threshold:.2f}")
print(f"Accuracy       : {gb_tuned_accuracy * 100:.2f}%")
print(f"Anomaly F1     : {gb_tuned_f1:.4f}")
print(f"Anomaly Recall : {gb_tuned_recall:.4f}")
print(f"Precision      : {gb_tuned_precision:.4f}")
print(f"PR-AUC         : {gb_tuned_pr_auc:.4f}")
print(f"ROC-AUC        : {gb_tuned_roc_auc:.4f}")

print("\nConfusion Matrix:")
print(confusion_matrix(y_test_anom, gb_test_pred))

# ----------------------------------------------------------------------
# Baseline improved XGBoost vs tuned XGBoost
# ----------------------------------------------------------------------

baseline_xgb = next(r for r in improved_results if r['model'] == 'XGBoost (Weighted)')

print("\n" + "=" * 78)
print(" XGBOOST BEFORE VS AFTER FINE TUNING ")
print("=" * 78)

print(f"{'Metric':<20} | {'Before':<12} | {'Tuned':<12} | {'Delta':<10}")
print("-" * 62)

print(f"{'OOF F1':<20} | {baseline_xgb['oof_f1']:<12.4f} | {best_xgb_oof_f1:<12.4f} | {best_xgb_oof_f1 - baseline_xgb['oof_f1']:+.4f}")
print(f"{'Test F1':<20} | {baseline_xgb['f1']:<12.4f} | {xgb_tuned_f1:<12.4f} | {xgb_tuned_f1 - baseline_xgb['f1']:+.4f}")
print(f"{'Recall':<20} | {baseline_xgb['recall']:<12.4f} | {xgb_tuned_recall:<12.4f} | {xgb_tuned_recall - baseline_xgb['recall']:+.4f}")
print(f"{'Precision':<20} | {baseline_xgb['precision']:<12.4f} | {xgb_tuned_precision:<12.4f} | {xgb_tuned_precision - baseline_xgb['precision']:+.4f}")
print(f"{'PR-AUC':<20} | {baseline_xgb['pr_auc']:<12.4f} | {xgb_tuned_pr_auc:<12.4f} | {xgb_tuned_pr_auc - baseline_xgb['pr_auc']:+.4f}")
print(f"{'ROC-AUC':<20} | {baseline_xgb['roc_auc']:<12.4f} | {xgb_tuned_roc_auc:<12.4f} | {xgb_tuned_roc_auc - baseline_xgb['roc_auc']:+.4f}")

# 6. FINAL comparison: tuned XGB vs tuned GB
print("\n" + "=" * 78)
print(" FINAL TUNED MODEL COMPARISON ")
print("=" * 78)

print(f"{'Metric':<20} | {'XGBoost':<12} | {'Grad Boost':<12}")
print("-" * 52)

print(f"{'CV PR-AUC':<20} | {xgb_search.best_score_:<12.4f} | {gb_search.best_score_:<12.4f}")
print(f"{'OOF F1':<20} | {best_xgb_oof_f1:<12.4f} | {best_gb_oof_f1:<12.4f}")
print(f"{'Test F1':<20} | {xgb_tuned_f1:<12.4f} | {gb_tuned_f1:<12.4f}")
print(f"{'Recall':<20} | {xgb_tuned_recall:<12.4f} | {gb_tuned_recall:<12.4f}")
print(f"{'Precision':<20} | {xgb_tuned_precision:<12.4f} | {gb_tuned_precision:<12.4f}")
print(f"{'PR-AUC':<20} | {xgb_tuned_pr_auc:<12.4f} | {gb_tuned_pr_auc:<12.4f}")
print(f"{'ROC-AUC':<20} | {xgb_tuned_roc_auc:<12.4f} | {gb_tuned_roc_auc:<12.4f}")

# ======================================================================
# FINAL ROBUSTNESS VALIDATION - LOCKED XGBOOST
# Repeated outer CV + inner OOF threshold selection
# ======================================================================

print("\n" + "=" * 78)
print(" FINAL ROBUSTNESS VALIDATION: LOCKED XGBOOST ")
print("=" * 78)

# Locked parameters from previous fine-tuning
locked_xgb_params = {
    'objective': 'binary:logistic',
    'eval_metric': 'logloss',
    'n_estimators': 500,
    'max_depth': 4,
    'learning_rate': 0.10,
    'min_child_weight': 2,
    'subsample': 0.80,
    'colsample_bytree': 0.70,
    'gamma': 0.20,
    'reg_alpha': 0.0,
    'reg_lambda': 1.0,
    'scale_pos_weight': 4.780150753768845,
    'random_state': 42,
    'n_jobs': 1
}

outer_cv = RepeatedStratifiedKFold(n_splits=5, n_repeats=3, random_state=2026)
robustness_results = []

for fold_no, (train_idx, val_idx) in enumerate(outer_cv.split(X_anomaly, y_anomaly), start=1):
    X_outer_train = X_anomaly.iloc[train_idx]
    X_outer_val = X_anomaly.iloc[val_idx]
    y_outer_train = y_anomaly[train_idx]
    y_outer_val = y_anomaly[val_idx]

    # Inner CV is used only to determine threshold.
    inner_cv = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)

    inner_pipeline = build_anomaly_pipeline(XGBClassifier(**locked_xgb_params))

    inner_oof_probs = cross_val_predict(
        inner_pipeline, X_outer_train, y_outer_train,
        cv=inner_cv, method='predict_proba', n_jobs=1
    )[:, 1]

    # Find best threshold using inner OOF predictions only
    fold_threshold = 0.50
    fold_best_f1 = -1

    for threshold in np.arange(0.10, 0.81, 0.01):
        inner_pred = (inner_oof_probs >= threshold).astype(int)
        current_f1 = f1_score(y_outer_train, inner_pred, pos_label=1, zero_division=0)

        if current_f1 > fold_best_f1:
            fold_best_f1 = current_f1
            fold_threshold = threshold

    # Fit locked model on the complete outer training fold
    final_pipeline = build_anomaly_pipeline(XGBClassifier(**locked_xgb_params))
    final_pipeline.fit(X_outer_train, y_outer_train)

    val_probs = final_pipeline.predict_proba(X_outer_val)[:, 1]
    val_pred = (val_probs >= fold_threshold).astype(int)

    fold_result = {
        'fold': fold_no,
        'threshold': fold_threshold,
        'accuracy': accuracy_score(y_outer_val, val_pred),
        'precision': precision_score(y_outer_val, val_pred, zero_division=0),
        'recall': recall_score(y_outer_val, val_pred, zero_division=0),
        'f1': f1_score(y_outer_val, val_pred, zero_division=0),
        'pr_auc': average_precision_score(y_outer_val, val_probs),
        'roc_auc': roc_auc_score(y_outer_val, val_probs)
    }

    robustness_results.append(fold_result)

    print(
        f"Fold {fold_no:>2} | Th={fold_threshold:.2f} | "
        f"F1={fold_result['f1']:.4f} | Recall={fold_result['recall']:.4f} | "
        f"Precision={fold_result['precision']:.4f} | PR-AUC={fold_result['pr_auc']:.4f}"
    )


# ======================================================================
# ROBUSTNESS SUMMARY
# ======================================================================

robust_df = pd.DataFrame(robustness_results)

print("\n" + "=" * 78)
print(" FINAL ROBUSTNESS SUMMARY ")
print("=" * 78)

print(f"{'Metric':<15} | {'Mean':<10} | {'Std':<10} | {'Min':<10} | {'Max':<10}")
print("-" * 65)

for metric in ['accuracy', 'precision', 'recall', 'f1', 'pr_auc', 'roc_auc']:
    values = robust_df[metric]

    print(
        f"{metric.upper():<15} | "
        f"{values.mean():<10.4f} | "
        f"{values.std():<10.4f} | "
        f"{values.min():<10.4f} | "
        f"{values.max():<10.4f}"
    )

print("\nThreshold Stability:")
print(f"Mean Threshold : {robust_df['threshold'].mean():.3f}")
print(f"Std Threshold  : {robust_df['threshold'].std():.3f}")
print(f"Min Threshold  : {robust_df['threshold'].min():.2f}")
print(f"Max Threshold  : {robust_df['threshold'].max():.2f}")