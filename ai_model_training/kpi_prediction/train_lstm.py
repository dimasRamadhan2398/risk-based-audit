import os
import random
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from torch.utils.data import Dataset, DataLoader
from sklearn.preprocessing import MinMaxScaler, OneHotEncoder, LabelEncoder, StandardScaler
from sklearn.model_selection import train_test_split, StratifiedKFold, cross_val_score
from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score, accuracy_score, f1_score, classification_report, confusion_matrix
from xgboost import XGBClassifier
from sklearn.ensemble import GradientBoostingClassifier, RandomForestClassifier, ExtraTreesClassifier
import matplotlib.pyplot as plt

# ==========================================
# REPRODUCIBILITY & SEED
# ==========================================
def set_seed(seed=42):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    if torch.cuda.is_available():
        torch.cuda.manual_seed_all(seed)

set_seed(42)
device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
print(f"Using compute device: {device}")

# ==========================================
# 1. LOAD & PREPROCESS DATASET
# ==========================================
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'kpi_data.csv')

if not os.path.exists(dataset_path):
    fallback_path = os.path.join(current_dir, 'lstm.csv')
    if os.path.exists(fallback_path):
        dataset_path = fallback_path

print(f"Loading KPI dataset from: {dataset_path}")
df_raw = pd.read_csv(dataset_path)

# Clean dataset
df = df_raw.drop_duplicates().reset_index(drop=True)
print(f"Original rows: {len(df_raw)} | Clean rows: {len(df)} | Unique KPIs: {df['Nama KPI'].nunique()}")

# Identify Target Columns
reg_target = 'TARGET: Nilai Aktual (%)'
class_targets = [col for col in df.columns if col.startswith('TARGET:') and col != reg_target]

# ======================================================================
# PART 1: TIME-SERIES LSTM REGRESSION (TARGET: Nilai Aktual (%))
# GROUPED BY NAMA KPI + FEATURE SCALING
# ======================================================================
print("\n======================================================================")
print(" PART 1: TIME-SERIES LSTM REGRESSION (TARGET: Nilai Aktual (%)) ")
print(" (Grouping History Per Nama KPI + MinMaxScaler) ")
print("======================================================================")

SEQ_LENGTH = 10
X_reg_list, y_reg_list, unscaled_y_list, kpi_name_list = [], [], [], []
kpi_scalers = {}

for kpi_name, group in df.groupby('Nama KPI'):
    vals = pd.to_numeric(group[reg_target], errors='coerce').dropna().values.reshape(-1, 1)
    if len(vals) <= SEQ_LENGTH:
        continue
    
    scaler = MinMaxScaler(feature_range=(0, 1))
    scaled_vals = scaler.fit_transform(vals)
    kpi_scalers[kpi_name] = scaler
    
    for i in range(len(scaled_vals) - SEQ_LENGTH):
        X_reg_list.append(scaled_vals[i:i + SEQ_LENGTH])
        y_reg_list.append(scaled_vals[i + SEQ_LENGTH])
        unscaled_y_list.append(vals[i + SEQ_LENGTH][0])
        kpi_name_list.append(kpi_name)

X_reg = np.array(X_reg_list)
y_reg = np.array(y_reg_list)
unscaled_y = np.array(unscaled_y_list)
kpi_names = np.array(kpi_name_list)

print(f"Created {len(X_reg)} sliding window sequences across {len(kpi_scalers)} KPI groups.")

# 80/20 Train-Test Split
split_idx = int(len(X_reg) * 0.8)
X_train_r, X_val_r = X_reg[:split_idx], X_reg[split_idx:]
y_train_r, y_val_r = y_reg[:split_idx], y_reg[split_idx:]
unscaled_y_val = unscaled_y[split_idx:]
val_kpi_names = kpi_names[split_idx:]

class RegressionDataset(Dataset):
    def __init__(self, X, y):
        self.X = torch.tensor(X, dtype=torch.float32)
        self.y = torch.tensor(y, dtype=torch.float32)
    def __len__(self):
        return len(self.X)
    def __getitem__(self, idx):
        return self.X[idx], self.y[idx]

class LSTMRegressor(nn.Module):
    def __init__(self, input_dim=1, hidden_dim=64, num_layers=2, output_dim=1, dropout=0.1):
        super().__init__()
        self.hidden_dim = hidden_dim
        self.num_layers = num_layers
        self.lstm = nn.LSTM(input_size=input_dim, hidden_size=hidden_dim, num_layers=num_layers, batch_first=True, dropout=dropout if num_layers > 1 else 0)
        self.fc = nn.Linear(hidden_dim, output_dim)
    def forward(self, x):
        h0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        c0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        out, _ = self.lstm(x, (h0, c0))
        return self.fc(out[:, -1, :])

train_loader_r = DataLoader(RegressionDataset(X_train_r, y_train_r), batch_size=32, shuffle=True)
val_loader_r = DataLoader(RegressionDataset(X_val_r, y_val_r), batch_size=32, shuffle=False)

model_r = LSTMRegressor().to(device)
criterion_r = nn.MSELoss()
optimizer_r = torch.optim.Adam(model_r.parameters(), lr=0.001)

EPOCHS = 100
print(f"Training LSTM Regressor for {EPOCHS} epochs on {device}...")
for epoch in range(EPOCHS):
    model_r.train()
    for b_x, b_y in train_loader_r:
        b_x, b_y = b_x.to(device), b_y.to(device)
        optimizer_r.zero_grad()
        loss = criterion_r(model_r(b_x), b_y)
        loss.backward()
        optimizer_r.step()

model_r.eval()
val_preds_scaled = []
with torch.no_grad():
    for b_x, _ in val_loader_r:
        b_x = b_x.to(device)
        val_preds_scaled.extend(model_r(b_x).cpu().numpy())

val_preds_scaled = np.array(val_preds_scaled).reshape(-1, 1)

# Inverse transform predictions per KPI group
val_preds_unscaled = []
for idx, p_scaled in enumerate(val_preds_scaled):
    kname = val_kpi_names[idx]
    un_val = kpi_scalers[kname].inverse_transform([[p_scaled[0]]])[0, 0]
    val_preds_unscaled.append(un_val)

val_preds_unscaled = np.array(val_preds_unscaled)

mse_r = mean_squared_error(unscaled_y_val, val_preds_unscaled)
rmse_r = np.sqrt(mse_r)
mae_r = mean_absolute_error(unscaled_y_val, val_preds_unscaled)
mape_r = np.mean(np.abs((unscaled_y_val - val_preds_unscaled) / unscaled_y_val)) * 100
acc_r = max(0.0, 100.0 - mape_r)
r2_r = r2_score(unscaled_y_val, val_preds_unscaled)

print(f"\n--- Validation Performance: {reg_target} ---")
print(f"LSTM Prediction Accuracy : {acc_r:.2f} %")
print(f"Mean Absolute Error (MAE): {mae_r:.4f}")
print(f"Root Mean Sq Error (RMSE): {rmse_r:.4f}")
print(f"R2 Score                 : {r2_r:.4f}")

# Plot & Save Evaluation Curve
fig, ax = plt.subplots(figsize=(10, 5))
ax.plot(unscaled_y_val[:50], label='Actual Target (Ground Truth)', color='green', marker='o')
ax.plot(val_preds_unscaled[:50], label='LSTM Forecast Prediction', color='red', linestyle='--', marker='x')
ax.set_title(f'LSTM Time-Series Forecast ({reg_target})\nAccuracy: {acc_r:.2f}% | MAE: {mae_r:.4f}')
ax.set_xlabel('Time Index (Validation Sample)')
ax.set_ylabel('KPI Actual Value (%)')
ax.legend()
ax.grid(True)
plt.tight_layout()
plot_path_r = os.path.join(current_dir, "lstm_forecast_Nilai_Aktual.png")
plt.savefig(plot_path_r)
plt.close()
print(f"Saved forecast evaluation plot to: {plot_path_r}")

del model_r
if torch.cuda.is_available():
    torch.cuda.empty_cache()

# ======================================================================
# PART 2: DISCRETE TARGET CLASSIFICATION FOR Impact & Likelihood
# MULTI-MODEL BENCHMARK (XGBoost, Gradient Boosting, Random Forest, Extra Trees)
# ======================================================================
print("\n======================================================================")
print(" PART 2: DISCRETE TARGET CLASSIFICATION (Impact & Likelihood) ")
print("======================================================================")

classification_summary = {}

for target_col in class_targets:
    print(f"\n" + "-" * 70)
    print(f" TRAINING & EVALUATING CLASSIFICATION MODELS FOR: {target_col} ")
    print("-" * 70)
    
    df_feat = df.copy()
    df_feat['target_class'] = pd.to_numeric(df_feat[target_col], errors='coerce').astype(int) - 1
    
    # Feature Engineering: Lag features per KPI group
    for lag in [1, 2, 3]:
        df_feat[f'lag_{lag}'] = df_feat.groupby('Nama KPI')[reg_target].shift(lag)
    df_feat['rolling_mean_3'] = df_feat.groupby('Nama KPI')[reg_target].shift(1).rolling(3).mean()
    df_feat = df_feat.dropna().reset_index(drop=True)
    
    # One-Hot Encoding for KPI Name
    ohe = OneHotEncoder(sparse_output=False)
    kpi_encoded = ohe.fit_transform(df_feat[['Nama KPI']])
    
    num_feats = df_feat[[reg_target, 'lag_1', 'lag_2', 'lag_3', 'rolling_mean_3']].values
    X_tab = np.hstack([kpi_encoded, num_feats])
    y_tab = df_feat['target_class'].values
    
    # Safe 80/20 Train-Test Split
    class_counts = pd.Series(y_tab).value_counts()
    stratify_param = y_tab if class_counts.min() >= 2 else None
    X_tr_t, X_val_t, y_tr_t, y_val_t = train_test_split(X_tab, y_tab, test_size=0.2, random_state=42, stratify=stratify_param)
    
    # Feature Scaling
    sc = StandardScaler()
    X_tr_scaled = sc.fit_transform(X_tr_t)
    X_val_scaled = sc.transform(X_val_t)
    
    candidate_tab_models = {
        'XGBoost': XGBClassifier(eval_metric='mlogloss', random_state=42),
        'Gradient Boosting': GradientBoostingClassifier(random_state=42),
        'Random Forest': RandomForestClassifier(n_estimators=200, random_state=42),
        'Extra Trees': ExtraTreesClassifier(n_estimators=200, random_state=42)
    }
    
    skf = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)
    results = []
    
    for m_name, model in candidate_tab_models.items():
        cv_scores = cross_val_score(model, X_tr_scaled, y_tr_t, cv=skf, scoring='accuracy')
        model.fit(X_tr_scaled, y_tr_t)
        y_pred = model.predict(X_val_scaled)
        
        acc = accuracy_score(y_val_t, y_pred)
        f1 = f1_score(y_val_t, y_pred, average='weighted')
        
        results.append({
            'model_name': m_name,
            'test_acc': acc,
            'weighted_f1': f1,
            'cv_mean': cv_scores.mean(),
            'cv_std': cv_scores.std(),
            'y_pred': y_pred
        })
        
    results_sorted = sorted(results, key=lambda r: (r['test_acc'], r['cv_mean']), reverse=True)
    winner = results_sorted[0]
    
    print("\n--- Performance Leaderboard ---")
    print(f"{'Rank':<5} | {'Model Name':<20} | {'Test Accuracy':<15} | {'Weighted F1':<13} | {'5-Fold CV Mean ± Std':<22}")
    print("-" * 82)
    for idx, r in enumerate(results_sorted):
        print(f"{idx+1:<5} | {r['model_name']:<20} | {r['test_acc']*100:>6.2f} %        | {r['weighted_f1']:>8.4f}    | {r['cv_mean']*100:>6.2f}% ± {r['cv_std']*100:.2f}%")
        
    print(f"\n[WINNER] Best Model for {target_col}: {winner['model_name']} (Accuracy: {winner['test_acc']*100:.2f}%)")
    
    class_names = [str(c + 1) for c in sorted(np.unique(y_tab))]
    print(f"\n--- Detailed Classification Report ({winner['model_name']}) ---")
    print(classification_report(y_val_t, winner['y_pred'], target_names=class_names, zero_division=0))
    
    classification_summary[target_col] = winner

# ======================================================================
# OVERALL ACCURACY SUMMARY REPORT
# ======================================================================
print("\n======================================================================")
print(" FINAL IMPROVED ACCURACY SUMMARY REPORT FOR ALL TARGETS ")
print("======================================================================")
print(f" 1. TARGET: Nilai Aktual (%)   | Model: PyTorch LSTM Regressor | Acc: {acc_r:.2f} % | MAE: {mae_r:.4f}")
for target_name, win in classification_summary.items():
    print(f" 2. {target_name:<27} | Model: {win['model_name']:<23} | Acc: {win['test_acc']*100:.2f} % | Weighted F1: {win['weighted_f1']:.4f}")

print("\nKPI Multi-Model Improvement Pipeline completed successfully!")
