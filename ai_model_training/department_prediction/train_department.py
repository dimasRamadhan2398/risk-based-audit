import os
import sys
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder, StandardScaler
from sklearn.model_selection import train_test_split, StratifiedKFold, cross_val_score, RandomizedSearchCV
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier, ExtraTreesClassifier, ExtraTreesRegressor, GradientBoostingRegressor
from sklearn.model_selection import cross_val_score, StratifiedKFold, train_test_split
from sklearn.svm import SVC
from sklearn.linear_model import LogisticRegression
from sklearn.neural_network import MLPClassifier
from xgboost import XGBClassifier, XGBRegressor
from sklearn.metrics import confusion_matrix, accuracy_score, classification_report, f1_score, make_scorer, mean_absolute_error

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

# Step 3: Categorical One-Hot Encoding (Columns 0: entity, 1: risk_category)
X_raw = df[feature_cols].values
ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(), [0, 1])], remainder='passthrough')
X = np.array(ct.fit_transform(X_raw))
print(f"Encoded Feature matrix shape: {X.shape}")

# Targets configuration
targets_config = {
    'Impact Prediction (target_impact)': {
        'series': df['target_impact'],
        'class_names': [str(c) for c in sorted(df['target_impact'].unique())]
    },
    'Likelihood Prediction (target_likelihood)': {
        'series': df['target_likelihood'],
        'class_names': [str(c) for c in sorted(df['target_likelihood'].unique())]
    }
}

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
    
    # Feature Scaling (StandardScaler)
    sc = StandardScaler()
    X_train_scaled = sc.fit_transform(X_train)
    X_test_scaled = sc.transform(X_test)
    
    # Stratified 5-Fold Cross Validation
    skf = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)
    
    results = []
    trained_models = {}
    
    for m_name, model in candidate_models.items():
        # Cross-validation on training set
        cv_scores = cross_val_score(model, X_train_scaled, y_train, cv=skf, scoring='accuracy')
        
        # Fit on training set and predict on test set
        model.fit(X_train_scaled, y_train)
        y_pred = model.predict(X_test_scaled)
        
        test_acc = accuracy_score(y_test, y_pred)
        weighted_f1 = f1_score(y_test, y_pred, average='weighted')
        
        results.append({
            'model_name': m_name,
            'test_acc': test_acc,
            'weighted_f1': weighted_f1,
            'cv_mean': cv_scores.mean(),
            'cv_std': cv_scores.std(),
            'y_pred': y_pred
        })
        trained_models[m_name] = model
        
    # Sort results by CV Mean and Test Accuracy
    results_sorted = sorted(results, key=lambda r: (r['cv_mean'], r['test_acc']), reverse=True)
    winner = results_sorted[0]
    
    print("\n--- Performance Leaderboard ---")
    print(f"{'Rank':<5} | {'Model Name':<20} | {'Test Accuracy':<15} | {'Weighted F1':<13} | {'5-Fold CV Mean ± Std':<22}")
    print("-" * 82)
    for idx, r in enumerate(results_sorted):
        print(f"{idx+1:<5} | {r['model_name']:<20} | {r['test_acc']*100:>6.2f} %        | {r['weighted_f1']:>8.4f}    | {r['cv_mean']*100:>6.2f}% ± {r['cv_std']*100:.2f}%")
        
    print(f"\n[WINNER] Winner Model: {winner['model_name']} (CV Mean: {winner['cv_mean']*100:.2f}%, Test Acc: {winner['test_acc']*100:.2f}%)")
    
    # Detailed Evaluation for Winner Model
    print(f"\n--- Detailed Classification Report for Winner ({winner['model_name']}) ---")
    print("Confusion Matrix:")
    print(confusion_matrix(y_test, winner['y_pred']))
    print("\nClassification Report:")
    print(classification_report(y_test, winner['y_pred'], target_names=config['class_names'], zero_division=0))

# ========================================================================

print("\n" + "=" * 78)
print(" ORDINAL DEPARTMENT MODEL EXPERIMENT ")
print("=" * 78)

# target-specific feature sets
impact_features = ['inherent_impact', 'audit_findings_count', 'kpi_below_target_count', 'kpi_volatility', 'previous_risk_score']
likelihood_features = ['inherent_likelihood', 'audit_findings_count', 'kpi_below_target_count', 'kpi_volatility', 'previous_risk_score']

optimized_targets = {
    'Impact': {
        'features': impact_features,
        'target': 'target_impact',
    },
    'Likelihood': {
        'features': likelihood_features,
        'target': 'target_likelihood',
    }
}

# scorer
def ordinal_weighted_f1(y_true, y_pred):
    y_pred_class = np.clip(np.rint(y_pred), 1, 5).astype(int)
    return f1_score(y_true.astype(int), y_pred_class, average='weighted')

ordinal_f1_scorer = make_scorer(ordinal_weighted_f1, greater_is_better=True)

# candidate models
ordinal_models = {
    'XGBoost Regressor': XGBRegressor(n_estimators=300, max_depth=5, learning_rate=0.05, subsample=1.0, colsample_bytree=1.0, random_state=42, n_jobs=1),
    'Extra Trees Regressor': ExtraTreesRegressor(n_estimators=300, max_depth=12, random_state=42, n_jobs=1),
    'Gradient Boosting Regressor': GradientBoostingRegressor(n_estimators=300, learning_rate=0.05, max_depth=3, random_state=42),
}

# run the experiment
for target_name, config in optimized_targets.items():
    print("\n" + "=" * 78)
    print(f" ORDINAL MODEL EXPERIMENT: {target_name}")
    print("=" * 78)

    X_opt = df[config['features']].copy()
    y_opt = df[config['target']].astype(int).values

    X_train_opt, X_test_opt, y_train_opt, y_test_opt = (
        train_test_split(X_opt, y_opt, test_size=0.20, random_state=42, stratify=y_opt))

    skf_opt = StratifiedKFold(n_splits=5, shuffle=True,random_state=42)

    ordinal_results = []

    for model_name, model in ordinal_models.items():
        print(f"\nEvaluating ordinal model: {model_name}")
        
        #cross-validation
        cv_scores = cross_val_score(model, X_train_opt, y_train_opt, cv=skf_opt, scoring=ordinal_f1_scorer, n_jobs=-1)
        
        # final fit
        model.fit( X_train_opt, y_train_opt)

        # Regression output, example:
        # 1.23, 2.79, 4.12
        raw_pred = model.predict(X_test_opt)

        # Convert back to ordinal classes 1-5
        y_pred_opt = np.clip( np.rint(raw_pred), 1, 5 ).astype(int)

        test_acc = accuracy_score(y_test_opt, y_pred_opt)

        test_f1 = f1_score(y_test_opt, y_pred_opt, average='weighted')

        test_mae = mean_absolute_error(y_test_opt, y_pred_opt)

        ordinal_results.append({'model': model_name, 'cv_f1': cv_scores.mean(), 'cv_std': cv_scores.std(), 'test_accuracy': test_acc, 'test_f1': test_f1, 'mae': test_mae})

    # Sort by CV Weighted F1
    ordinal_results = sorted(ordinal_results, key=lambda x: x['cv_f1'], reverse=True)

    print("\n--- Ordinal Model Leaderboard ---")
    # print(f"{'Rank':<5} | {'Model':<30} | {'CV F1':<16} | {'Test Acc':<10} | {'Test F1':<10} | {'MAE':<8}")
    # print("-" * 95)

    for rank, result in enumerate(ordinal_results, start=1):
        print(
            f"{rank}. {result['model']} | "
            f"CV F1: {result['cv_f1']:.4f} ± "
            f"{result['cv_std']:.4f} | "
            f"Test Acc: "
            f"{result['test_accuracy'] * 100:.2f}% | "
            f"Test F1: {result['test_f1']:.4f} | "
            f"MAE: {result['mae']:.4f}"
        )
        
    winner_opt = ordinal_results[0]

    print(f"\n[ORDINAL WINNER]\n {winner_opt['model']}")

    # print( f"CV Weighted F1 : {winner_opt['cv_f1']:.4f}")
    # print(f"Test Accuracy  : {winner_opt['test_accuracy'] * 100:.2f}%")
    # print(f"Test F1        : {winner_opt['test_f1']:.4f}")
    # print(f"Ordinal MAE    : {winner_opt['mae']:.4f}")
    
    # ========================================================================

    if target_name == 'Impact':   
        print("\n" + "=" * 78)
        print(" FINE TUNING: EXTRA TREES - IMPACT ")
        print("=" * 78)
        impact_extra_trees_params = {
            'n_estimators': [300, 400, 500, 700, 900],
            'max_depth': [None, 8, 10, 12, 14, 16, 20],
            'min_samples_split': [2, 3, 4, 5, 8],
            'min_samples_leaf': [1, 2, 3, 4],
            'max_features': [0.6, 0.7, 0.8, 0.9, 1.0, 'sqrt']
        }

        impact_search = RandomizedSearchCV(
            estimator=ExtraTreesRegressor(random_state=42,n_jobs=1),
            param_distributions=impact_extra_trees_params,
            n_iter=40,
            scoring=ordinal_f1_scorer,
            cv=skf_opt,
            random_state=42,
            n_jobs=-1,
            verbose=1,
            refit=True
        )

        impact_search.fit(X_train_opt, y_train_opt)

        print(f"\nBest CV F1: {impact_search.best_score_:.4f}")

        print("\nBest Parameters:")
        for key, value in impact_search.best_params_.items():
            print(f"  {key}: {value}")
            
        best_impact_model = (impact_search.best_estimator_)
        impact_raw_pred = (best_impact_model.predict(X_test_opt))
        impact_pred = np.clip(np.rint(impact_raw_pred), 1, 5).astype(int)

        impact_tuned_acc = accuracy_score(y_test_opt,impact_pred)
        impact_tuned_f1 = f1_score(y_test_opt,impact_pred,average='weighted')
        impact_tuned_mae = mean_absolute_error(y_test_opt, impact_pred)

        print(f"\nTuned Test Accuracy : {impact_tuned_acc * 100:.2f}%")
        print(f"Tuned Test F1       : {impact_tuned_f1:.4f}")
        print(f"Tuned Ordinal MAE   : {impact_tuned_mae:.4f}")

    # tune likelihood model for gradient boost and extra trees likelihood
    elif target_name == 'Likelihood':
        print("\n" + "=" * 78)
        print(" FINE TUNING: LIKELIHOOD ")
        print("=" * 78)
        
        likelihood_gb_params = {
            'n_estimators': [150, 200, 250, 300, 400, 500],
            'learning_rate': [0.01, 0.02, 0.03, 0.05, 0.08, 0.10],
            'max_depth': [2, 3, 4, 5],
            'min_samples_split': [2, 3, 4, 5, 8],
            'min_samples_leaf': [1, 2, 3, 4],
            'subsample': [ 0.7, 0.8, 0.9, 1.0],
            'max_features': [None, 'sqrt', 'log2']
        }
        likelihood_gb_search = RandomizedSearchCV(
            estimator=GradientBoostingRegressor(random_state=42),
            param_distributions=likelihood_gb_params,
            n_iter=40,
            scoring=ordinal_f1_scorer,
            cv=skf_opt,
            random_state=42,
            n_jobs=-1,
            verbose=1,
            refit=True
        )

        likelihood_et_search = RandomizedSearchCV(
            estimator=ExtraTreesRegressor(random_state=42, n_jobs=1),
            param_distributions=impact_extra_trees_params,
            n_iter=40,
            scoring=ordinal_f1_scorer,
            cv=skf_opt,
            random_state=42,
            n_jobs=-1,
            verbose=1,
            refit=True
        )
        
        likelihood_gb_search.fit(X_train_opt, y_train_opt)
        likelihood_et_search.fit(X_train_opt, y_train_opt)

        print("\nGradient Boosting " + "Best CV F1: " + f"{likelihood_gb_search.best_score_:.4f}")
        print("Extra Trees " + "Best CV F1: " + f"{likelihood_et_search.best_score_:.4f}")
    
# ========================================================================

# final test evaluation for tuned likelihood model
print("\n" + "=" * 78)
print(" FINAL TEST: TUNED LIKELIHOOD MODELS ")
print("=" * 78)

# gradient boosting
best_likelihood_gb = (likelihood_gb_search.best_estimator_)

gb_raw_pred = best_likelihood_gb.predict(X_test_opt)
gb_pred = np.clip(np.rint(gb_raw_pred), 1, 5).astype(int)
gb_test_acc = accuracy_score(y_test_opt, gb_pred)
gb_test_f1 = f1_score(y_test_opt, gb_pred, average='weighted')
gb_test_mae = mean_absolute_error(y_test_opt, gb_pred)

# extra trees
best_likelihood_et = (likelihood_et_search.best_estimator_)
et_raw_pred = best_likelihood_et.predict(X_test_opt)
et_pred = np.clip(np.rint(et_raw_pred), 1, 5).astype(int)
et_test_acc = accuracy_score(y_test_opt, et_pred)
et_test_f1 = f1_score(y_test_opt, et_pred, average='weighted')
et_test_mae = mean_absolute_error(y_test_opt, et_pred)

# results
print("\nGradient Boosting Tuned")
print(f"Test Accuracy : {gb_test_acc * 100:.2f}%")
print(f"Test F1       : {gb_test_f1:.4f}")
print(f"Ordinal MAE   : {gb_test_mae:.4f}")

print("\nExtra Trees Tuned")
print(f"Test Accuracy : {et_test_acc * 100:.2f}%")
print(f"Test F1       : {et_test_f1:.4f}")
print(f"Ordinal MAE   : {et_test_mae:.4f}")

print("\nBest Gradient Boosting Parameters:")
for key, value in (likelihood_gb_search.best_params_.items()):
    print(f"  {key}: {value}")

print("\nBest Extra Trees Parameters:")
for key, value in (likelihood_et_search.best_params_.items()):
    print(f"  {key}: {value}")