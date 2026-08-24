import os
import sys
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder, StandardScaler
from sklearn.model_selection import train_test_split, StratifiedKFold, cross_validate
from sklearn.pipeline import Pipeline
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier, ExtraTreesClassifier
from sklearn.svm import SVC
from sklearn.linear_model import LogisticRegression
from sklearn.neural_network import MLPClassifier
from xgboost import XGBClassifier
from sklearn.metrics import confusion_matrix, accuracy_score, classification_report, f1_score

# Add root directory to sys.path to enable importing preprocess_xgboost
current_dir = os.path.dirname(os.path.abspath(__file__))
root_dir = os.path.abspath(os.path.join(current_dir, '..', '..'))
if root_dir not in sys.path:
    sys.path.insert(0, root_dir)

try:
    from ai_model_training.preprocess_anomaly_dept.preprocess_xgboost import preprocess_xgboost_df
    HAS_PREPROCESS_MODULE = True
except ImportError:
    HAS_PREPROCESS_MODULE = False

# 1. Importing the dataset
dataset_path = os.path.join(current_dir, 'department_dataset.csv')
print(f"Loading dataset from: {dataset_path}")
df_raw = pd.read_csv(dataset_path)
initial_rows = len(df_raw)

# ======================================================================
# DATA PREPROCESSING PIPELINE (from preprocess_xgboost.py)
# ======================================================================
print("\n======================================================================")
print(" 1. DATA PREPROCESSING & CLEANING PIPELINE (via preprocess_xgboost) ")
print("======================================================================")

if HAS_PREPROCESS_MODULE:
    print("Applying Preprocessing Pipeline from preprocess_xgboost.py...")
    df = preprocess_xgboost_df(df_raw)
else:
    # Inline fallback implementation of preprocess_xgboost.py
    df = df_raw.copy().rename(
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
    categorical_columns = ["entity", "risk_category"]
    for col in categorical_columns:
        if col in df.columns:
            df[col] = df[col].astype("string").str.strip()

    numeric_columns = [
        "inherent_likelihood", "inherent_impact", "audit_findings_count",
        "kpi_below_target_count", "kpi_volatility", "previous_risk_score",
        "assessment_month", "target_likelihood", "target_impact"
    ]
    df[numeric_columns] = df[numeric_columns].apply(pd.to_numeric, errors="coerce")
    df.loc[~df["assessment_month"].between(1, 12), "assessment_month"] = np.nan
    df = df.drop_duplicates().dropna().reset_index(drop=True)
    df["month_sin"] = np.sin(2 * np.pi * (df["assessment_month"] - 1) / 12)
    df["month_cos"] = np.cos(2 * np.pi * (df["assessment_month"] - 1) / 12)

clean_rows = len(df)
duplicate_rows = initial_rows - clean_rows
print(f"Initial rows: {initial_rows} | Clean rows after preprocess_xgboost: {clean_rows}")

# Step 2: Domain Feature Engineering
print("\nApplying Feature Engineering...")
df['inherent_risk_score'] = df['inherent_likelihood'] * df['inherent_impact']
df['kpi_volatility_log'] = np.log1p(df['kpi_volatility'])
df['findings_kpi_ratio'] = (df['audit_findings_count'] + 1) / (df['kpi_below_target_count'] + 1)
df['risk_score_diff'] = df['inherent_risk_score'] - df['previous_risk_score']

# Explicit feature columns (Target Leakage Removed)
feature_cols = [
    'entity', 'risk_category', 'inherent_likelihood', 'inherent_impact',
    'audit_findings_count', 'kpi_below_target_count', 'kpi_volatility',
    'previous_risk_score', 'assessment_month', 'inherent_risk_score',
    'kpi_volatility_log', 'findings_kpi_ratio', 'risk_score_diff',
    'month_sin', 'month_cos'
]

print(f"Selected {len(feature_cols)} clean input features (Target Leakage Removed).")

# Step 3: Prepare RAW features.
# Encoding and scaling will be performed inside the sklearn Pipeline
# so each CV fold only learns preprocessing from its own training data.
X = df[feature_cols].copy()

categorical_features = [
    'entity',
    'risk_category'
]

numeric_features = [
    col for col in feature_cols
    if col not in categorical_features
]

print(f"Raw feature matrix shape: {X.shape}")
print(f"Categorical features: {categorical_features}")
print(f"Numeric features: {len(numeric_features)}")

# Machine Learning candidate models
candidate_models = {
    'Gradient Boosting': GradientBoostingClassifier(n_estimators=300, learning_rate=0.05, max_depth=5, random_state=42),
    'XGBoost': XGBClassifier(n_estimators=300, max_depth=5, learning_rate=0.05, eval_metric='mlogloss', random_state=42),
    'Extra Trees': ExtraTreesClassifier(n_estimators=300, max_depth=12, random_state=42),
    'Random Forest': RandomForestClassifier(n_estimators=300, max_depth=12, random_state=42),
    'MLP Neural Net': MLPClassifier(hidden_layer_sizes=(128, 64), max_iter=800, random_state=42),
    'SVM (RBF)': SVC(kernel='rbf', C=2.0, random_state=42),
    'Logistic Regression': LogisticRegression(max_iter=1000, C=2.0, random_state=42)
}

# Build a fresh preprocessing pipeline.
# A new preprocessor is created for each model so OneHotEncoder
# and StandardScaler are fitted only on the training fold during CV.
def build_preprocessor():
    return ColumnTransformer(
        transformers=[
            (
                'categorical',
                OneHotEncoder(
                    handle_unknown='ignore',
                    sparse_output=False
                ),
                categorical_features
            ),
            (
                'numeric',
                'passthrough',
                numeric_features
            )
        ],
        remainder='drop'
    )
    
# Complete ML pipeline: raw dataframe -> encoding -> scaling -> classifier
def build_model_pipeline(model):
    return Pipeline(
        steps=[
            ('preprocessor', build_preprocessor()),
            ('scaler', StandardScaler()),
            ('classifier', model)
        ]
    )

# Multi-Model Evaluation Loop
for target_name, config in targets_config.items():
    print(f"\n======================================================================")
    print(f" TRAINING & EVALUATING MODELS FOR: {target_name} ")
    print(f"======================================================================")
    
    le = LabelEncoder()
    y = le.fit_transform(config['series'])
    
    # 80/20 Stratified Train-Test Split
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42, stratify=y
    )
    
    # # Feature Scaling (StandardScaler)
    # sc = StandardScaler()
    # X_train_scaled = sc.fit_transform(X_train)
    # X_test_scaled = sc.transform(X_test)
    
    # Stratified 5-Fold Cross Validation
    skf = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)
    
    results = []
    trained_models = {}
    
    for m_name, model in candidate_models.items():
        print(f"\nEvaluating: {m_name}")

        # Build a fresh pipeline for every model.
        # Preprocessing is performed INSIDE each CV fold.
        pipeline = build_model_pipeline(model)

        scoring = {
            'accuracy': 'accuracy',
            'weighted_f1': 'f1_weighted'
        }

        cv_results = cross_validate(
            pipeline,
            X_train,
            y_train,
            cv=skf,
            scoring=scoring,
            n_jobs=-1,
            return_train_score=False
        )

        # Fit complete pipeline using training set only.
        pipeline.fit(X_train, y_train)

        # Test data remains untouched until this point.
        y_pred = pipeline.predict(X_test)

        test_acc = accuracy_score(
            y_test,
            y_pred
        )

        weighted_f1 = f1_score(
            y_test,
            y_pred,
            average='weighted'
        )

        results.append({
            'model_name': m_name,
            'test_acc': test_acc,
            'weighted_f1': weighted_f1,
            'cv_mean': cv_results['test_accuracy'].mean(),
            'cv_std': cv_results['test_accuracy'].std(),
            'cv_f1_mean': cv_results['test_weighted_f1'].mean(),
            'cv_f1_std': cv_results['test_weighted_f1'].std(),
            'y_pred': y_pred
        })

        trained_models[m_name] = pipeline
        
    # Sort results by CV Mean and Test Accuracy
    results_sorted = sorted(results, key=lambda r: (r['cv_f1_mean'], r['cv_mean']), reverse=True)
    winner = results_sorted[0]
    
    print("\n--- Performance Leaderboard ---")
    print(f"{'Rank':<5} | {'Model Name':<20} | {'Test Accuracy':<10} | {'Test F1':<10} | {'CV Accuracy':<18} | {'CV Weighted F1':<18}")
    print("-" * 100)
    for idx, r in enumerate(results_sorted):
        print(f"{idx + 1:<5} | {r['model_name']:<20} | {r['test_acc'] * 100:>6.2f}%   | {r['weighted_f1']:>8.4f} | {r['cv_mean'] * 100:>6.2f}% ± {r['cv_std'] * 100:.2f}% | {r['cv_f1_mean']:>7.4f} ± {r['cv_f1_std']:.4f}")
        
    print(f"\n[WINNER] Winner Model: {winner['model_name']}")
    print(f"CV Weighted F1: {winner['cv_f1_mean']:.4f} | CV Accuracy: {winner['cv_mean']:.4f} | Test Accuracy: {winner['test_acc']:.4f}")
    
    # Detailed Evaluation for Winner Model
    print(f"\n--- Detailed Classification Report for Winner ({winner['model_name']}) ---")
    print("Confusion Matrix:")
    print(confusion_matrix(y_test, winner['y_pred']))
    print("\nClassification Report:")
    print(classification_report(y_test, winner['y_pred'], target_names=config['class_names'], zero_division=0))
