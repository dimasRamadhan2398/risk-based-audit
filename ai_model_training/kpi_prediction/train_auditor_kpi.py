import os
import random
import pickle
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from torch.utils.data import Dataset, DataLoader
from sklearn.preprocessing import MinMaxScaler, StandardScaler, LabelEncoder
from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score, accuracy_score, f1_score, classification_report
from sklearn.utils.class_weight import compute_sample_weight
from sklearn.ensemble import GradientBoostingClassifier, RandomForestClassifier
from xgboost import XGBClassifier

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

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.abspath(os.path.join(CURRENT_DIR, '..', '..'))
BACKEND_MODEL_DIR = os.path.join(PROJECT_ROOT, 'backend', 'python-ai', 'models')
DATA_SERVER_MODEL_DIR = os.path.join(PROJECT_ROOT, 'data-server', 'ai-engine', 'models')

os.makedirs(BACKEND_MODEL_DIR, exist_ok=True)
os.makedirs(DATA_SERVER_MODEL_DIR, exist_ok=True)

# Try loading from high-performance Parquet / DB pool first (zero CSV parsing overhead)
import sys
sys.path.append(os.path.join(PROJECT_ROOT, 'backend', 'python-ai'))
try:
    from ai_training_storage import load_auditor_training_dataframe
    df = load_auditor_training_dataframe()
    print(f"Loaded Auditor KPI dataset via AI Training Storage (Parquet/DB Pool): {len(df)} rows")
except Exception as e_load:
    dataset_path = os.path.join(CURRENT_DIR, 'auditor_kpi_data.csv')
    if not os.path.exists(dataset_path):
        raise FileNotFoundError(f"Dataset {dataset_path} not found. Run generate_auditor_kpi_dataset.py first.")
    print(f"Loading Auditor KPI dataset from CSV fallback: {dataset_path}")
    df = pd.read_csv(dataset_path)

# Parse explicit monthly time index
period_parts = df['Periode'].astype(str).str.extract(r'Bulan\s+(\d{1,2})\s+(\d{4})')
df['periode_dt'] = pd.to_datetime(
    period_parts[1] + '-' + period_parts[0].str.zfill(2) + '-01',
    errors='coerce'
)
df = df.sort_values(['Kode SO', 'periode_dt']).reset_index(drop=True)

print(f"Loaded {len(df)} rows across {df['Kode SO'].nunique()} Strategic Audit Plan KPIs.")
print(f"Period range: {df['periode_dt'].min().strftime('%Y-%m')} to {df['periode_dt'].max().strftime('%Y-%m')}")

# ======================================================================
# PART 1: AUDITOR PYTORCH LSTM REGRESSOR (Nilai Aktual & Rasio Ketercapaian)
# ======================================================================
print("\n======================================================================")
print(" PART 1: AUDITOR PYTORCH LSTM REGRESSOR FOR STRATEGIC AUDIT PLAN ")
print("======================================================================")

SEQ_LENGTH = 10
reg_target = 'TARGET: Nilai Aktual'

X_train_list, y_train_list = [], []
X_val_list, y_val_list = [], []
unscaled_y_val_list, val_so_list = [], []
auditor_scalers = {}

for so_code, group in df.groupby('Kode SO', sort=False):
    group = group.sort_values('periode_dt')
    vals = pd.to_numeric(group[reg_target], errors='coerce').dropna().values.reshape(-1, 1)
    
    if len(vals) <= SEQ_LENGTH + 1:
        continue
        
    split_point = int(len(vals) * 0.80)
    train_vals = vals[:split_point]
    
    scaler = MinMaxScaler(feature_range=(0, 1))
    scaler.fit(train_vals)
    auditor_scalers[so_code] = scaler
    
    scaled_vals = scaler.transform(vals)
    
    for idx in range(SEQ_LENGTH, split_point):
        X_train_list.append(scaled_vals[idx - SEQ_LENGTH:idx])
        y_train_list.append(scaled_vals[idx])
        
    for idx in range(split_point, len(scaled_vals)):
        X_val_list.append(scaled_vals[idx - SEQ_LENGTH:idx])
        y_val_list.append(scaled_vals[idx])
        unscaled_y_val_list.append(vals[idx][0])
        val_so_list.append(so_code)

X_train_r = np.array(X_train_list, dtype=np.float32)
y_train_r = np.array(y_train_list, dtype=np.float32)
X_val_r = np.array(X_val_list, dtype=np.float32)
y_val_r = np.array(y_val_list, dtype=np.float32)

print(f"Training sequences: {len(X_train_r)} | Validation sequences: {len(X_val_r)}")

class AuditorLSTMRegressor(nn.Module):
    def __init__(self, input_dim=1, hidden_dim=64, num_layers=2, output_dim=1, dropout=0.1):
        super().__init__()
        self.hidden_dim = hidden_dim
        self.num_layers = num_layers
        self.lstm = nn.LSTM(
            input_size=input_dim,
            hidden_size=hidden_dim,
            num_layers=num_layers,
            batch_first=True,
            dropout=dropout if num_layers > 1 else 0.0
        )
        self.fc = nn.Linear(hidden_dim, output_dim)
        
    def forward(self, x):
        h0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        c0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        out, _ = self.lstm(x, (h0, c0))
        return self.fc(out[:, -1, :])

class TimeSeriesDataset(Dataset):
    def __init__(self, X, y):
        self.X = torch.tensor(X, dtype=torch.float32)
        self.y = torch.tensor(y, dtype=torch.float32)
    def __len__(self):
        return len(self.X)
    def __getitem__(self, idx):
        return self.X[idx], self.y[idx]

train_dataset = TimeSeriesDataset(X_train_r, y_train_r)
val_dataset = TimeSeriesDataset(X_val_r, y_val_r)

train_loader = DataLoader(train_dataset, batch_size=32, shuffle=True)
val_loader = DataLoader(val_dataset, batch_size=32, shuffle=False)

lstm_model = AuditorLSTMRegressor(input_dim=1, hidden_dim=64, num_layers=2, output_dim=1, dropout=0.1).to(device)
criterion = nn.MSELoss()
optimizer = torch.optim.Adam(lstm_model.parameters(), lr=0.002, weight_decay=1e-5)
scheduler = torch.optim.lr_scheduler.ReduceLROnPlateau(optimizer, mode='min', factor=0.5, patience=5)

best_val_loss = float('inf')
best_state = None

print("Training Auditor LSTM model...")
for epoch in range(1, 41):
    lstm_model.train()
    total_train_loss = 0.0
    for X_b, y_b in train_loader:
        X_b, y_b = X_b.to(device), y_b.to(device)
        optimizer.zero_grad()
        preds = lstm_model(X_b)
        loss = criterion(preds, y_b)
        loss.backward()
        optimizer.step()
        total_train_loss += loss.item() * len(X_b)
        
    train_loss = total_train_loss / len(train_dataset)
    
    lstm_model.eval()
    total_val_loss = 0.0
    with torch.no_grad():
        for X_b, y_b in val_loader:
            X_b, y_b = X_b.to(device), y_b.to(device)
            preds = lstm_model(X_b)
            loss = criterion(preds, y_b)
            total_val_loss += loss.item() * len(X_b)
            
    val_loss = total_val_loss / len(val_dataset)
    scheduler.step(val_loss)
    
    if val_loss < best_val_loss:
        best_val_loss = val_loss
        best_state = lstm_model.state_dict()
        
    if epoch % 10 == 0:
        print(f"Epoch {epoch:02d}/40 | Train Loss: {train_loss:.5f} | Val Loss: {val_loss:.5f}")

if best_state is not None:
    lstm_model.load_state_dict(best_state)

# Evaluate on Validation Set
lstm_model.eval()
with torch.no_grad():
    val_preds_scaled = lstm_model(torch.tensor(X_val_r).to(device)).cpu().numpy()

# Invert scaling
val_preds_unscaled = []
for i in range(len(val_preds_scaled)):
    so = val_so_list[i]
    pred_s = val_preds_scaled[i][0]
    sc = auditor_scalers[so]
    pred_u = sc.inverse_transform([[pred_s]])[0][0]
    val_preds_unscaled.append(pred_u)

val_preds_unscaled = np.array(val_preds_unscaled)
unscaled_y_val = np.array(unscaled_y_val_list)

rmse = np.sqrt(mean_squared_error(unscaled_y_val, val_preds_unscaled))
mae = mean_absolute_error(unscaled_y_val, val_preds_unscaled)
r2 = r2_score(unscaled_y_val, val_preds_unscaled)
print(f"\n[LSTM EVALUATION] Validation RMSE: {rmse:.3f} | MAE: {mae:.3f} | R2 Score: {r2:.3f}")

# ======================================================================
# PART 2: FINE-TUNED AUDITOR STATUS CLASSIFIER (Good, Moderate, Poor)
# ======================================================================
print("\n======================================================================")
print(" PART 2: FINE-TUNED STATUS CLASSIFIER (Good / Moderate / Poor) ")
print("======================================================================")

df_feat = df.copy()

# Feature engineering based on HIG / HIB polarity
df_feat['polarity_is_hig'] = (df_feat['Tipe Polaritas (HIG/HIB)'] == 'HIG').astype(int)
df_feat['target_val'] = df_feat['Nilai Target'].astype(float)
df_feat['actual_val'] = df_feat['TARGET: Nilai Aktual'].astype(float)
df_feat['achievement_ratio'] = df_feat['TARGET: Rasio Ketercapaian (%)'].astype(float)

# Normalized gap adjusted for polarity
df_feat['gap_ratio'] = np.where(
    df_feat['polarity_is_hig'] == 1,
    (df_feat['actual_val'] - df_feat['target_val']) / (df_feat['target_val'] + 1e-6),
    (df_feat['target_val'] - df_feat['actual_val']) / (df_feat['target_val'] + 1e-6)
)

# Temporal Lag & Rolling Features
for lag in [1, 2, 3]:
    df_feat[f'lag_{lag}'] = df_feat.groupby('Kode SO')['TARGET: Nilai Aktual'].shift(lag)
    df_feat[f'ratio_lag_{lag}'] = df_feat.groupby('Kode SO')['TARGET: Rasio Ketercapaian (%)'].shift(lag)

df_feat['rolling_mean_3'] = df_feat.groupby('Kode SO')['TARGET: Nilai Aktual'].shift(1).rolling(3).mean()
df_feat['rolling_std_3'] = df_feat.groupby('Kode SO')['TARGET: Nilai Aktual'].shift(1).rolling(3).std().fillna(0)
df_feat['velocity_trend'] = df_feat['actual_val'] - df_feat['lag_3']

# Clean NaN from shifts
df_feat = df_feat.dropna().reset_index(drop=True)

feature_cols = [
    'polarity_is_hig', 'target_val', 'actual_val', 'achievement_ratio', 'gap_ratio',
    'lag_1', 'lag_2', 'lag_3', 'ratio_lag_1', 'ratio_lag_2', 'ratio_lag_3',
    'rolling_mean_3', 'rolling_std_3', 'velocity_trend'
]

X_class = df_feat[feature_cols].values
status_labels = df_feat['TARGET: Status Achievement'].values

le = LabelEncoder()
y_class = le.fit_transform(status_labels)
print(f"Status classes: {le.classes_}")

# Split 80/20 chronologically per KPI
split_mask = []
for so, grp in df_feat.groupby('Kode SO'):
    n = len(grp)
    s = int(n * 0.80)
    split_mask.extend([True] * s + [False] * (n - s))
split_mask = np.array(split_mask)

X_train_c, y_train_c = X_class[split_mask], y_class[split_mask]
X_val_c, y_val_c = X_class[~split_mask], y_class[~split_mask]

scaler_class = StandardScaler()
X_train_c_scaled = scaler_class.fit_transform(X_train_c)
X_val_c_scaled = scaler_class.transform(X_val_c)

sample_weights = compute_sample_weight('balanced', y_train_c)

# Fine-tuned Gradient Boosting Classifier
status_clf = GradientBoostingClassifier(
    n_estimators=120,
    learning_rate=0.08,
    max_depth=4,
    min_samples_split=5,
    random_state=42
)
status_clf.fit(X_train_c_scaled, y_train_c, sample_weight=sample_weights)

val_preds_c = status_clf.predict(X_val_c_scaled)
acc = accuracy_score(y_val_c, val_preds_c)
f1_macro = f1_score(y_val_c, val_preds_c, average='macro')

print(f"\n[STATUS CLASSIFIER EVALUATION]")
print(f"Validation Accuracy: {acc * 100:.2f}% | Macro F1-Score: {f1_macro:.3f}")
print("\nClassification Report:")
print(classification_report(y_val_c, val_preds_c, labels=list(range(len(le.classes_))), target_names=le.classes_, zero_division=0))

# ======================================================================
# PART 3: SAVE ARTIFACTS (auditor_kpi_lstm.pth & auditor_kpi_bundle.pkl)
# ======================================================================
all_actual_vals = pd.to_numeric(df[reg_target], errors='coerce').dropna().values
global_min_val = float(all_actual_vals.min())
global_max_val = float(all_actual_vals.max())

# Build KPI metadata map
kpi_meta = {}
for _, row in df[['Kode SO', 'Nama KPI', 'Kategori Pilar', 'Tipe Polaritas (HIG/HIB)', 'Satuan', 'Nilai Target']].drop_duplicates().iterrows():
    kpi_meta[row['Kode SO']] = {
        'code': row['Kode SO'],
        'name': row['Nama KPI'],
        'pillar': row['Kategori Pilar'],
        'polarity': row['Tipe Polaritas (HIG/HIB)'],
        'unit': row['Satuan'],
        'target': float(row['Nilai Target'])
    }

auditor_kpi_bundle = {
    'scalers': auditor_scalers,
    'global_min_val': global_min_val,
    'global_max_val': global_max_val,
    'feature_cols': feature_cols,
    'scaler_class': scaler_class,
    'status_classifier': status_clf,
    'label_encoder': le,
    'kpi_meta': kpi_meta,
    'seq_length': SEQ_LENGTH,
    'known_kpis': sorted(df['Kode SO'].unique().tolist())
}

for target_dir in [BACKEND_MODEL_DIR, DATA_SERVER_MODEL_DIR]:
    torch.save(lstm_model.state_dict(), os.path.join(target_dir, 'auditor_kpi_lstm.pth'))
    with open(os.path.join(target_dir, 'auditor_kpi_bundle.pkl'), 'wb') as f:
        pickle.dump(auditor_kpi_bundle, f)
    print(f"[SAVED] auditor_kpi_lstm.pth and auditor_kpi_bundle.pkl in {target_dir}")

print("\n[SUCCESS] AUDITOR KPI SUITE TRAINED AND SAVED SUCCESSFULLY!")

