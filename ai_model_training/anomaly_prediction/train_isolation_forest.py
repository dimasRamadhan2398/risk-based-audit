import os
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, StandardScaler
from sklearn.model_selection import train_test_split
from sklearn.ensemble import IsolationForest
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score

# 1. Importing the dataset
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'isolation_forest.csv')
print(f"Loading dataset from: {dataset_path}")
df = pd.read_csv(dataset_path)

# Cleaning formatting from 'amount (dalam Juta Rp)' (e.g. converting "1,500.00" to 1500.0)
print("Cleaning 'amount (dalam Juta Rp)' column formatting...")
df['amount (dalam Juta Rp)'] = df['amount (dalam Juta Rp)'].astype(str).str.replace(',', '').astype(float)

# Select features (columns 1 to 7) and target (column 8)
# Features: Entitas, Deskripsi, amount, hour_of_day, day_of_week, is_new_beneficiary, is_round_amount
X = df.iloc[:, 1:-1].values
y = df.iloc[:, -1].values

print("Original dataset shape:", df.shape)
print("Features (X) initial shape:", X.shape)
print("Target (y) initial shape:", y.shape)

# 2. Encoding categorical data
# Column index 0 of X is 'Entitas'
# Column index 1 of X is 'Deskripsi'
print("\nApplying One-Hot Encoding on categorical columns (0: Entitas, 1: Deskripsi)...")
ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(), [0, 1])], remainder='passthrough')
X = np.array(ct.fit_transform(X))
print("Encoded Features (X) shape:", X.shape)

# Convert target y labels to binary integers: "Tidak" -> 0 (normal), "Ya (Anomali)" -> 1 (anomaly)
print("Converting target variable labels to binary integers (0 = Normal, 1 = Anomaly)...")
y_binary = np.where(y == 'Ya (Anomali)', 1, 0)

# Calculate overall contamination rate
overall_contamination = np.mean(y_binary)
print(f"Overall dataset anomaly rate (contamination): {overall_contamination * 100:.2f} %")

# 3. Splitting the dataset into the Training set and Test set
print("\nSplitting the dataset into Training and Test set (80/20 split, stratified)...")
X_train, X_test, y_train, y_test = train_test_split(X, y_binary, test_size = 0.2, random_state = 0, stratify=y_binary)

# 4. Feature Scaling (Standardization)
print("Applying Standardization (StandardScaler) on features...")
sc = StandardScaler()
X_train = sc.fit_transform(X_train)
X_test = sc.transform(X_test)

# 5. Training Isolation Forest on the Training set
# We use the training set contamination rate
train_contamination = np.mean(y_train)
print(f"\nTraining anomaly rate (contamination): {train_contamination * 100:.2f} %")
print("Training Isolation Forest model on the training set...")
classifier = IsolationForest(contamination=train_contamination, random_state=0)
classifier.fit(X_train)

# 6. Evaluation
# Isolation Forest predicts: 1 for inliers (normal), -1 for outliers (anomaly)
# We map: 1 -> 0 (normal), -1 -> 1 (anomaly)
print("\n--- Test Set Evaluation ---")
y_pred_raw = classifier.predict(X_test)
y_pred = np.where(y_pred_raw == -1, 1, 0)

cm = confusion_matrix(y_test, y_pred)
print("Confusion Matrix:")
print(cm)
test_accuracy = accuracy_score(y_test, y_pred)
print(f"Accuracy Score: {test_accuracy * 100:.2f} %")

print("\nClassification Report:")
print(classification_report(y_test, y_pred, target_names=['Normal', 'Anomaly']))
