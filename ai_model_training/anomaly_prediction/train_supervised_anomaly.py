import os
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, StandardScaler
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from xgboost import XGBClassifier
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score

# 1. Importing the dataset
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'isolation_forest.csv')
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
y = df['is_anomaly (Ground Truth)'].values

print("Features (X) initial shape:", X.shape)
print("Target (y) initial shape:", y.shape)

# 3. Encoding categorical data
# Column index 0: 'Entitas'
# Column index 1: 'Deskripsi'
print("\nApplying One-Hot Encoding on categorical columns (0: Entitas, 1: Deskripsi)...")
ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(), [0, 1])], remainder='passthrough')
X = np.array(ct.fit_transform(X))
print("Encoded Features (X) shape:", X.shape)

# Convert target y labels to binary integers: "Tidak" -> 0 (normal), "Ya (Anomali)" -> 1 (anomaly)
print("Converting target variable labels to binary integers (0 = Normal, 1 = Anomaly)...")
y_binary = np.where(y == 'Ya (Anomali)', 1, 0)

# 4. Splitting the dataset into the Training set and Test set
print("\nSplitting the dataset into Training and Test set (80/20 split, stratified)...")
X_train, X_test, y_train, y_test = train_test_split(X, y_binary, test_size = 0.2, random_state = 0, stratify=y_binary)

# 5. Feature Scaling (Standardization)
print("Applying Standardization (StandardScaler) on features...")
sc = StandardScaler()
X_train = sc.fit_transform(X_train)
X_test = sc.transform(X_test)

# 6. Model Training & Evaluation

# --- MODEL 1: Random Forest Classifier ---
print("\n=== Model 1: Random Forest Classifier ===")
rf_classifier = RandomForestClassifier(n_estimators=100, random_state=0)
rf_classifier.fit(X_train, y_train)

y_pred_rf = rf_classifier.predict(X_test)
rf_accuracy = accuracy_score(y_test, y_pred_rf)
print("Confusion Matrix:")
print(confusion_matrix(y_test, y_pred_rf))
print(f"Accuracy Score: {rf_accuracy * 100:.2f} %")
print("Classification Report:")
print(classification_report(y_test, y_pred_rf, target_names=['Normal', 'Anomaly']))

# --- MODEL 2: XGBoost Classifier ---
print("\n=== Model 2: XGBoost Classifier ===")
xgb_classifier = XGBClassifier(eval_metric='logloss', random_state=0)
xgb_classifier.fit(X_train, y_train)

y_pred_xgb = xgb_classifier.predict(X_test)
xgb_accuracy = accuracy_score(y_test, y_pred_xgb)
print("Confusion Matrix:")
print(confusion_matrix(y_test, y_pred_xgb))
print(f"Accuracy Score: {xgb_accuracy * 100:.2f} %")
print("Classification Report:")
print(classification_report(y_test, y_pred_xgb, target_names=['Normal', 'Anomaly']))

# 7. Model Comparison Summary
print("\n=== Performance Comparison Summary ===")
print(f"Random Forest Accuracy : {rf_accuracy * 100:.2f} %")
print(f"XGBoost Accuracy       : {xgb_accuracy * 100:.2f} %")
if rf_accuracy > xgb_accuracy:
    print("Winner: Random Forest Classifier!")
elif xgb_accuracy > rf_accuracy:
    print("Winner: XGBoost Classifier!")
else:
    print("Result: Tie!")
