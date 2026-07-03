import os
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder, StandardScaler
from sklearn.model_selection import train_test_split, cross_val_score
from xgboost import XGBClassifier
from sklearn.metrics import confusion_matrix, accuracy_score

# 1. Importing the dataset
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'xgboost.csv')
print(f"Loading dataset from: {dataset_path}")
dataset = pd.read_csv(dataset_path)

X = dataset.iloc[:, :-1].values
y = dataset.iloc[:, -1].values

print("Original dataset shape:", dataset.shape)
print("Features (X) initial shape:", X.shape)
print("Target (y) initial shape:", y.shape)

# 2. Encoding categorical data
# Column index 0: 'Entitas' (e.g. 'Jakarta Branch', 'IT Dept')
# Column index 1: 'Kategori Risiko' (e.g. 'Financial', 'Technology')
print("\nApplying One-Hot Encoding on categorical columns (0: Entitas, 1: Kategori Risiko)...")
ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(), [0, 1])], remainder='passthrough')
X = np.array(ct.fit_transform(X))

print("Applying Label Encoding on target variable (mapping [3, 4, 5] to [0, 1, 2])...")
le = LabelEncoder()
y = le.fit_transform(y)

print("Encoded Features (X) shape:", X.shape)
print("Unique classes in encoded Target (y):", np.unique(y))

# 3. Splitting the dataset into the Training set and Test set
print("\nSplitting the dataset into Training and Test set (80/20 split)...")
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size = 0.2, random_state = 0)

# 4. Feature Scaling (Standardization)
print("Applying Standardization (StandardScaler) on features...")
sc = StandardScaler()
X_train = sc.fit_transform(X_train)
X_test = sc.transform(X_test)

# 5. Training XGBoost on the Training set
print("Training XGBoost Classifier on the training set...")
classifier = XGBClassifier(eval_metric='mlogloss')
classifier.fit(X_train, y_train)

# 6. Making the Confusion Matrix
print("\n--- Test Set Evaluation ---")
y_pred = classifier.predict(X_test)
cm = confusion_matrix(y_test, y_pred)
print("Confusion Matrix:")
print(cm)
test_accuracy = accuracy_score(y_test, y_pred)
print(f"Accuracy Score: {test_accuracy * 100:.2f} %")

# 7. Applying k-Fold Cross Validation
print("\n--- 10-Fold Cross Validation ---")
print("Applying 10-Fold Cross Validation on the training set...")
accuracies = cross_val_score(estimator = classifier, X = X_train, y = y_train, cv = 10)
print("Mean Accuracy: {:.2f} %".format(accuracies.mean()*100))
print("Standard Deviation: {:.2f} %".format(accuracies.std()*100))
