import os
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, StandardScaler
import torch
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier, AdaBoostClassifier
from sklearn.svm import SVC
from sklearn.linear_model import LogisticRegression
from xgboost import XGBClassifier
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score

# 1. Importing the dataset
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'anomaly_data.csv')
print(f"Loading dataset from: {dataset_path}")
df = pd.read_csv(dataset_path)

# Cleaning formatting from 'amount (dalam Juta Rp)' (e.g. converting "1,500.00" to 1500.0)
print("Cleaning 'amount (dalam Juta Rp)' column formatting...")
df['amount'] = df['amount (dalam Juta Rp)'].astype(str).str.replace(',', '').astype(float)

# 2. Feature Engineering
print("\nApplying Feature Engineering...")
# Check if transaction is made on a weekend (Saturday = 6, Sunday = 7)
df['is_weekend_tx'] = df['day_of_week (1-7)'].isin([6, 7]).astype(int)

# Check if transaction is made during night/early hours (22:00 to 05:00)
df['is_night_tx'] = ((df['hour_of_day (0-23)'] >= 22) | (df['hour_of_day (0-23)'] <= 5)).astype(int)

# Transaction amount to hour ratio
df['amount_per_hour'] = df['amount'] / (df['hour_of_day (0-23)'] + 1)

# Features list (excluding ID Transaksi, original amount, and target)
feature_cols = [
    'Entitas', 'Deskripsi', 'amount', 'hour_of_day (0-23)', 
    'day_of_week (1-7)', 'is_new_beneficiary (1=Ya, 0=Tidak)', 
    'is_round_amount (1=Ya, 0=Tidak)', 'is_weekend_tx', 
    'is_night_tx', 'amount_per_hour'
]

X = df[feature_cols].values

print("Features (X) initial shape:", X.shape)

# 3. Encoding categorical data
# Column index 0: 'Entitas'
# Column index 1: 'Deskripsi'
print("\nApplying One-Hot Encoding on categorical columns (0: Entitas, 1: Deskripsi)...")
ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(), [0, 1])], remainder='passthrough')
X = np.array(ct.fit_transform(X))
print("Encoded Features (X) shape:", X.shape)

# 4. Targets to train and test
targets_config = {
    'Anomaly Detection (is_anomaly)': {
        'column': 'TARGET: is_anomaly' if 'TARGET: is_anomaly' in df.columns else 'is_anomaly (Ground Truth)',
        'transform': lambda y: np.where(y == 'Ya (Anomali)', 1, 0),
        'class_names': ['Normal', 'Anomaly']
    },
    'Impact Prediction (1-5)': {
        'column': 'TARGET: Impact (1-5)',
        'transform': lambda y: y.astype(int) - 1,  # Convert 1-5 to 0-4 for XGBoost
        'class_names': ['1', '2', '3', '4', '5']
    },
    'Likelihood Prediction (1-5)': {
        'column': 'TARGET: Likelihood (1-5)',
        'transform': lambda y: y.astype(int) - 1,  # Convert 1-5 to 0-4 for XGBoost
        'class_names': ['1', '2', '3', '4', '5']
    }
}

# Filter targets_config to only include columns that exist in df
available_targets = {}
for name, config in targets_config.items():
    if config['column'] in df.columns:
        available_targets[name] = config
    else:
        print(f"Warning: Target column '{config['column']}' not found in dataset. Skipping {name}.")

# 5. Model Training & Evaluation Loop for each target
for target_name, config in available_targets.items():
    print(f"\n======================================================================")
    print(f" TRAINING & EVALUATING FOR TARGET: {target_name} ")
    print(f"======================================================================")
    
    # Drop rows with null values in target for safety
    clean_mask = ~df[config['column']].isna()
    X_clean = X[clean_mask]
    y_raw = df[config['column']].values[clean_mask]
    y = config['transform'](y_raw)
    
    # Splitting the dataset into the Training set and Test set
    print("Splitting the dataset into Training and Test set (80/20 split, stratified)...")
    X_train, X_test, y_train, y_test = train_test_split(
        X_clean, y, test_size=0.2, random_state=0, stratify=y
    )
    
    # Feature Scaling (Standardization)
    print("Applying Standardization (StandardScaler) on features...")
    sc = StandardScaler()
    X_train_scaled = sc.fit_transform(X_train)
    X_test_scaled = sc.transform(X_test)
    
    # Determine devices / parameters
    device = 'cuda' if torch.cuda.is_available() else 'cpu'
    eval_metric = 'mlogloss' if len(config['class_names']) > 2 else 'logloss'
    
    # Define models to compare (with optimal parameters for specific targets)
    models = {
        'Random Forest': RandomForestClassifier(n_estimators=100, random_state=0),
        'XGBoost': XGBClassifier(eval_metric=eval_metric, random_state=0, device=device)
    }
    
    # Configure tuned/alternative models for specific targets to improve performance
    if 'is_anomaly' in target_name:
        models['AdaBoost (Tuned)'] = AdaBoostClassifier(learning_rate=0.01, n_estimators=50, random_state=0)
        models['Logistic Regression (Tuned)'] = LogisticRegression(C=0.001, max_iter=1000, random_state=0)
        models['SVM (RBF) (Tuned)'] = SVC(kernel='rbf', C=0.1, random_state=0)
    elif 'Likelihood' in target_name:
        models['AdaBoost (Tuned)'] = AdaBoostClassifier(learning_rate=0.1, n_estimators=200, random_state=0)
        models['Logistic Regression (Tuned)'] = LogisticRegression(C=0.5, max_iter=1000, random_state=0)
        models['SVM (RBF) (Tuned)'] = SVC(kernel='rbf', C=5, gamma=0.01, random_state=0)
    else: # e.g., Impact Prediction (baseline models)
        models['Logistic Regression'] = LogisticRegression(max_iter=1000, random_state=0)
        models['AdaBoost'] = AdaBoostClassifier(random_state=0)
        models['SVM (RBF)'] = SVC(kernel='rbf', random_state=0)
        
    # Dictionary to store results
    results = {}
    
    # Train and evaluate each model
    for model_name, model in models.items():
        print(f"\n--- Model: {model_name} ---")
        try:
            model.fit(X_train_scaled, y_train)
            y_pred = model.predict(X_test_scaled)
            accuracy = accuracy_score(y_test, y_pred)
            results[model_name] = accuracy
            
            print("Confusion Matrix:")
            print(confusion_matrix(y_test, y_pred))
            print(f"Accuracy Score: {accuracy * 100:.2f} %")
            print("Classification Report:")
            print(classification_report(y_test, y_pred, target_names=config['class_names'], zero_division=0))
        except Exception as e:
            print(f"Error training {model_name}: {e}")
            
    # Performance Comparison Summary
    print(f"\n=== Performance Comparison Summary ({target_name}) ===")
    sorted_results = sorted(results.items(), key=lambda item: item[1], reverse=True)
    for idx, (m_name, m_acc) in enumerate(sorted_results):
        print(f"{idx+1}. {m_name:<30} : {m_acc * 100:.2f} %")
    
    if sorted_results:
        print(f"Winner: {sorted_results[0][0]}!")
