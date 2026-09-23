import os
import random
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from torch.utils.data import Dataset, DataLoader
from sklearn.preprocessing import MinMaxScaler, OneHotEncoder, LabelEncoder, StandardScaler
from sklearn.model_selection import train_test_split, StratifiedKFold, cross_val_score, ParameterSampler
from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score, accuracy_score, f1_score, classification_report, confusion_matrix
from sklearn.utils.class_weight import compute_sample_weight
from xgboost import XGBClassifier, XGBRegressor
from sklearn.ensemble import GradientBoostingClassifier, RandomForestClassifier, ExtraTreesClassifier, GradientBoostingRegressor, RandomForestRegressor, ExtraTreesRegressor
from sklearn.pipeline import Pipeline
from sklearn.compose import ColumnTransformer
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

# Parse explicit monthly time index
period_parts = df['Periode'].astype(str).str.extract(r'Bulan\s+(\d{1,2})\s+(\d{4})')

df['periode_dt'] = pd.to_datetime(
    period_parts[1] + '-' + period_parts[0].str.zfill(2) + '-01',
    errors='coerce'
)

if df['periode_dt'].isna().any():
    raise ValueError("Some Periode values could not be parsed.")

df = df.sort_values(['Nama KPI', 'periode_dt']).reset_index(drop=True)

print(
    f"Period range: {df['periode_dt'].min().strftime('%Y-%m')} "
    f"to {df['periode_dt'].max().strftime('%Y-%m')}"
)

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

X_train_list, y_train_list = [], []
X_val_list, y_val_list = [], []
unscaled_y_val_list, val_kpi_name_list = [], []

kpi_scalers = {}

for kpi_name, group in df.groupby('Nama KPI', sort=False):
    # Preserve original CSV order within each KPI
    group = group.sort_values('periode_dt')

    vals = pd.to_numeric(
        group[reg_target], errors='coerce'
    ).dropna().values.reshape(-1, 1)

    if len(vals) <= SEQ_LENGTH + 1:
        continue

    # Chronological 80/20 split PER KPI
    split_point = int(len(vals) * 0.80)

    if split_point <= SEQ_LENGTH or split_point >= len(vals):
        continue

    train_vals = vals[:split_point]

    # Fit scaler ONLY on training history
    scaler = MinMaxScaler(feature_range=(0, 1))
    scaler.fit(train_vals)
    kpi_scalers[kpi_name] = scaler

    # Transform full series using train-fitted scaler
    scaled_vals = scaler.transform(vals)

    # Training sequences
    for target_idx in range(SEQ_LENGTH, split_point):
        X_train_list.append(
            scaled_vals[target_idx - SEQ_LENGTH:target_idx]
        )
        y_train_list.append(scaled_vals[target_idx])

    # Validation sequences:
    # targets are strictly after the chronological split,
    # but may use the last training observations as historical context.
    for target_idx in range(split_point, len(scaled_vals)):
        X_val_list.append(
            scaled_vals[target_idx - SEQ_LENGTH:target_idx]
        )
        y_val_list.append(scaled_vals[target_idx])

        unscaled_y_val_list.append(vals[target_idx][0])
        val_kpi_name_list.append(kpi_name)


X_train_r = np.array(X_train_list)
y_train_r = np.array(y_train_list)

X_val_r = np.array(X_val_list)
y_val_r = np.array(y_val_list)

unscaled_y_val = np.array(unscaled_y_val_list)
val_kpi_names = np.array(val_kpi_name_list)

print(
    f"Leakage-safe sequences created across {len(kpi_scalers)} KPI groups."
)
print(f"Training sequences   : {len(X_train_r)}")
print(f"Validation sequences : {len(X_val_r)}")

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
    
# ======================================================================
# LSTM INNER TEMPORAL TUNING
# Final 20% validation remains outside hyperparameter selection
# ======================================================================

print("\n--- LSTM Inner Temporal Hyperparameter Tuning ---")

def create_inner_lstm_data():
    X_inner_train, y_inner_train = [], []
    X_inner_val, y_inner_val = [], []

    for kpi_name, group in df.groupby('Nama KPI', sort=False):
        group = group.sort_values('periode_dt')

        vals = pd.to_numeric(
            group[reg_target], errors='coerce'
        ).dropna().values.reshape(-1, 1)

        outer_end = int(len(vals) * 0.80)
        outer_train_vals = vals[:outer_end]

        inner_end = int(len(outer_train_vals) * 0.80)

        if inner_end <= SEQ_LENGTH:
            continue

        # scaler only sees inner training history
        scaler = MinMaxScaler(feature_range=(0, 1))
        scaler.fit(outer_train_vals[:inner_end])

        scaled_vals = scaler.transform(outer_train_vals)

        for target_idx in range(SEQ_LENGTH, inner_end):
            X_inner_train.append(
                scaled_vals[target_idx - SEQ_LENGTH:target_idx]
            )
            y_inner_train.append(scaled_vals[target_idx])

        for target_idx in range(inner_end, outer_end):
            X_inner_val.append(
                scaled_vals[target_idx - SEQ_LENGTH:target_idx]
            )
            y_inner_val.append(scaled_vals[target_idx])

    return (
        np.array(X_inner_train),
        np.array(y_inner_train),
        np.array(X_inner_val),
        np.array(y_inner_val)
    )


X_inner_train, y_inner_train, X_inner_val, y_inner_val = create_inner_lstm_data()

print(f"Inner training sequences  : {len(X_inner_train)}")
print(f"Inner validation sequences: {len(X_inner_val)}")

lstm_configs = [
    {
        'hidden_dim': 32,
        'num_layers': 1,
        'dropout': 0.0,
        'lr': 0.001,
        'batch_size': 32
    },
    {
        'hidden_dim': 64,
        'num_layers': 1,
        'dropout': 0.0,
        'lr': 0.001,
        'batch_size': 32
    },
    {
        'hidden_dim': 64,
        'num_layers': 2,
        'dropout': 0.1,
        'lr': 0.001,
        'batch_size': 32
    },
    {
        'hidden_dim': 64,
        'num_layers': 2,
        'dropout': 0.2,
        'lr': 0.0005,
        'batch_size': 32
    },
    {
        'hidden_dim': 128,
        'num_layers': 1,
        'dropout': 0.0,
        'lr': 0.0005,
        'batch_size': 32
    },
    {
        'hidden_dim': 64,
        'num_layers': 2,
        'dropout': 0.1,
        'lr': 0.0005,
        'batch_size': 16
    }
]

lstm_tuning_results = []
TUNING_EPOCHS = 60

for config_no, config in enumerate(lstm_configs, start=1):
    set_seed(42)

    tuning_train_loader = DataLoader(
        RegressionDataset(X_inner_train, y_inner_train),
        batch_size=config['batch_size'],
        shuffle=True
    )

    tuning_val_loader = DataLoader(
        RegressionDataset(X_inner_val, y_inner_val),
        batch_size=config['batch_size'],
        shuffle=False
    )

    tuning_model = LSTMRegressor(
        hidden_dim=config['hidden_dim'],
        num_layers=config['num_layers'],
        dropout=config['dropout']
    ).to(device)

    tuning_optimizer = torch.optim.Adam(
        tuning_model.parameters(),
        lr=config['lr']
    )

    tuning_criterion = nn.MSELoss()

    for epoch in range(TUNING_EPOCHS):
        tuning_model.train()

        for b_x, b_y in tuning_train_loader:
            b_x, b_y = b_x.to(device), b_y.to(device)

            tuning_optimizer.zero_grad()
            loss = tuning_criterion(tuning_model(b_x), b_y)
            loss.backward()
            tuning_optimizer.step()

    tuning_model.eval()
    inner_predictions = []

    with torch.no_grad():
        for b_x, _ in tuning_val_loader:
            b_x = b_x.to(device)
            inner_predictions.extend(
                tuning_model(b_x).cpu().numpy()
            )

    inner_predictions = np.array(inner_predictions).reshape(-1)

    inner_mae = mean_absolute_error(
        y_inner_val.reshape(-1),
        inner_predictions
    )

    lstm_tuning_results.append({
        'config': config,
        'mae': inner_mae
    })

    print(
        f"Config {config_no} | "
        f"Hidden={config['hidden_dim']} | "
        f"Layers={config['num_layers']} | "
        f"Dropout={config['dropout']} | "
        f"LR={config['lr']} | "
        f"Batch={config['batch_size']} | "
        f"Inner MAE={inner_mae:.6f}"
    )

    del tuning_model

    if torch.cuda.is_available():
        torch.cuda.empty_cache()


lstm_tuning_results = sorted(
    lstm_tuning_results,
    key=lambda r: r['mae']
)

best_lstm_result = lstm_tuning_results[0]
best_lstm_config = best_lstm_result['config']

print("\nBest LSTM Configuration:")
for key, value in best_lstm_config.items():
    print(f"  {key}: {value}")

print(
    f"Best Inner Validation MAE: "
    f"{best_lstm_result['mae']:.6f}"
)

train_loader_r = DataLoader(
    RegressionDataset(X_train_r, y_train_r),
    batch_size=best_lstm_config['batch_size'],
    shuffle=True
)

val_loader_r = DataLoader(
    RegressionDataset(X_val_r, y_val_r),
    batch_size=best_lstm_config['batch_size'],
    shuffle=False
)

set_seed(42)

model_r = LSTMRegressor(
    hidden_dim=best_lstm_config['hidden_dim'],
    num_layers=best_lstm_config['num_layers'],
    dropout=best_lstm_config['dropout']
).to(device)

criterion_r = nn.MSELoss()

optimizer_r = torch.optim.Adam(
    model_r.parameters(),
    lr=best_lstm_config['lr']
)

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

# ----------------------------------------------------------------------
# Naive persistence baseline
# Prediction = last observed value in each input sequence
# ----------------------------------------------------------------------

naive_preds_unscaled = []

for idx, sequence in enumerate(X_val_r):
    kname = val_kpi_names[idx]
    last_scaled_value = sequence[-1, 0]
    last_unscaled_value = kpi_scalers[kname].inverse_transform(
        [[last_scaled_value]]
    )[0, 0]
    naive_preds_unscaled.append(last_unscaled_value)

naive_preds_unscaled = np.array(naive_preds_unscaled)

naive_mae = mean_absolute_error(unscaled_y_val, naive_preds_unscaled)
naive_rmse = np.sqrt(mean_squared_error(unscaled_y_val, naive_preds_unscaled))
naive_r2 = r2_score(unscaled_y_val, naive_preds_unscaled)

safe_denominator = np.where(
    np.abs(unscaled_y_val) < 1e-8,
    1e-8,
    np.abs(unscaled_y_val)
)

naive_mape = np.mean(
    np.abs(unscaled_y_val - naive_preds_unscaled) / safe_denominator
) * 100

naive_acc = max(0.0, 100.0 - naive_mape)

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

print("\n--- Naive Persistence Baseline ---")
print(f"Naive Prediction Accuracy : {naive_acc:.2f} %")
print(f"Mean Absolute Error (MAE) : {naive_mae:.4f}")
print(f"Root Mean Sq Error (RMSE) : {naive_rmse:.4f}")
print(f"R2 Score                  : {naive_r2:.4f}")

print("\n--- LSTM vs Naive Baseline ---")
print(f"{'Metric':<15} | {'Naive':<12} | {'LSTM':<12} | {'Delta':<12}")
print("-" * 58)
print(f"{'MAE':<15} | {naive_mae:<12.4f} | {mae_r:<12.4f} | {mae_r - naive_mae:+.4f}")
print(f"{'RMSE':<15} | {naive_rmse:<12.4f} | {rmse_r:<12.4f} | {rmse_r - naive_rmse:+.4f}")
print(f"{'R2':<15} | {naive_r2:<12.4f} | {r2_r:<12.4f} | {r2_r - naive_r2:+.4f}")

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
# FINAL LSTM WALK-FORWARD ROBUSTNESS
# Locked LSTM configuration vs Naive Persistence
# ======================================================================

print("\n" + "=" * 78)
print(" FINAL LSTM WALK-FORWARD ROBUSTNESS ")
print("=" * 78)

all_periods = np.array(sorted(df['periode_dt'].unique()))
initial_train_size = int(len(all_periods) * 0.60)

future_period_blocks = np.array_split(
    all_periods[initial_train_size:],
    4
)

lstm_robustness_results = []

for fold_no, val_periods_fold in enumerate(future_period_blocks, start=1):
    val_periods_fold = pd.DatetimeIndex(
        pd.to_datetime(val_periods_fold)
    )

    val_start = val_periods_fold.min()

    train_periods_fold = pd.DatetimeIndex(
        all_periods[all_periods < val_start]
    )

    X_fold_train_list, y_fold_train_list = [], []
    X_fold_val_list, y_fold_val_list = [], []
    fold_actual_list, fold_kpi_names = [], []

    fold_scalers = {}

    print(
        f"\nFold {fold_no} periods | "
        f"Train: {train_periods_fold.min().strftime('%Y-%m')} → "
        f"{train_periods_fold.max().strftime('%Y-%m')} | "
        f"Validation: {val_periods_fold.min().strftime('%Y-%m')} → "
        f"{val_periods_fold.max().strftime('%Y-%m')}"
    )

    for kpi_name, group in df.groupby('Nama KPI', sort=False):
        group = group.sort_values('periode_dt').copy()

        group[reg_target] = pd.to_numeric(
            group[reg_target],
            errors='coerce'
        )

        group = group.dropna(
            subset=[reg_target, 'periode_dt']
        ).reset_index(drop=True)

        vals = group[reg_target].values.reshape(-1, 1)

        # Gunakan pandas isin untuk datetime
        train_mask = group['periode_dt'].isin(
            train_periods_fold
        ).to_numpy()

        val_mask = group['periode_dt'].isin(
            val_periods_fold
        ).to_numpy()

        train_indices = np.flatnonzero(train_mask)
        val_indices = np.flatnonzero(val_mask)

        if len(train_indices) <= SEQ_LENGTH:
            print(
                f"Skipping {kpi_name}: "
                f"only {len(train_indices)} training observations."
            )
            continue

        if len(val_indices) == 0:
            print(
                f"Skipping {kpi_name}: "
                f"no validation observations."
            )
            continue

        last_train_idx = train_indices[-1]

        scaler = MinMaxScaler(
            feature_range=(0, 1)
        )

        # scaler hanya melihat historical training data
        scaler.fit(vals[train_indices])

        fold_scalers[kpi_name] = scaler
        scaled_vals = scaler.transform(vals)

        # Training sequences
        for target_idx in train_indices:
            if target_idx < SEQ_LENGTH:
                continue

            X_fold_train_list.append(scaled_vals[target_idx - SEQ_LENGTH:target_idx])
            y_fold_train_list.append(scaled_vals[target_idx])

        # Walk-forward validation sequences
        for target_idx in val_indices:
            if target_idx < SEQ_LENGTH:
                continue

            X_fold_val_list.append(scaled_vals[target_idx - SEQ_LENGTH:target_idx])
            y_fold_val_list.append(scaled_vals[target_idx])
            fold_actual_list.append(vals[target_idx][0])
            fold_kpi_names.append(kpi_name)

    X_fold_train = np.array(X_fold_train_list)
    y_fold_train = np.array(y_fold_train_list)

    X_fold_val = np.array(X_fold_val_list)
    y_fold_val = np.array(y_fold_val_list)

    fold_actual = np.array(fold_actual_list)
    fold_kpi_names = np.array(fold_kpi_names)

    print(
        f"Fold {fold_no} sequences | "
        f"Train={len(X_fold_train)} | "
        f"Validation={len(X_fold_val)}"
    )

    if len(X_fold_train) == 0:
        print(
            f"WARNING: Fold {fold_no} has no training sequences. "
            f"Skipping fold."
        )
        continue

    if len(X_fold_val) == 0:
        print(
            f"WARNING: Fold {fold_no} has no validation sequences. "
            f"Skipping fold."
        )
        continue

    set_seed(42 + fold_no)

    fold_train_loader = DataLoader(
        RegressionDataset(X_fold_train, y_fold_train),
        batch_size=best_lstm_config['batch_size'],
        shuffle=True
    )

    fold_val_loader = DataLoader(
        RegressionDataset(X_fold_val, y_fold_val),
        batch_size=best_lstm_config['batch_size'],
        shuffle=False
    )

    fold_model = LSTMRegressor(
        hidden_dim=best_lstm_config['hidden_dim'],
        num_layers=best_lstm_config['num_layers'],
        dropout=best_lstm_config['dropout']
    ).to(device)

    fold_optimizer = torch.optim.Adam(
        fold_model.parameters(),
        lr=best_lstm_config['lr']
    )

    fold_criterion = nn.MSELoss()

    for epoch in range(EPOCHS):
        fold_model.train()

        for b_x, b_y in fold_train_loader:
            b_x, b_y = b_x.to(device), b_y.to(device)

            fold_optimizer.zero_grad()
            loss = fold_criterion(fold_model(b_x), b_y)
            loss.backward()
            fold_optimizer.step()

    # LSTM predictions
    fold_model.eval()
    fold_preds_scaled = []

    with torch.no_grad():
        for b_x, _ in fold_val_loader:
            b_x = b_x.to(device)
            fold_preds_scaled.extend(
                fold_model(b_x).cpu().numpy()
            )

    fold_preds_scaled = np.array(
        fold_preds_scaled
    ).reshape(-1, 1)

    fold_preds_unscaled = []
    fold_naive_unscaled = []

    for idx, pred_scaled in enumerate(fold_preds_scaled):
        kname = fold_kpi_names[idx]
        scaler = fold_scalers[kname]

        pred_value = scaler.inverse_transform(
            [[pred_scaled[0]]]
        )[0, 0]

        last_scaled = X_fold_val[idx][-1, 0]

        naive_value = scaler.inverse_transform(
            [[last_scaled]]
        )[0, 0]

        fold_preds_unscaled.append(pred_value)
        fold_naive_unscaled.append(naive_value)

    fold_preds_unscaled = np.array(fold_preds_unscaled)
    fold_naive_unscaled = np.array(fold_naive_unscaled)

    lstm_mae_fold = mean_absolute_error(
        fold_actual,
        fold_preds_unscaled
    )

    lstm_rmse_fold = np.sqrt(
        mean_squared_error(
            fold_actual,
            fold_preds_unscaled
        )
    )

    lstm_r2_fold = r2_score(
        fold_actual,
        fold_preds_unscaled
    )

    naive_mae_fold = mean_absolute_error(
        fold_actual,
        fold_naive_unscaled
    )

    naive_rmse_fold = np.sqrt(
        mean_squared_error(
            fold_actual,
            fold_naive_unscaled
        )
    )

    improvement_mae = (
        (naive_mae_fold - lstm_mae_fold)
        / naive_mae_fold
    ) * 100

    lstm_robustness_results.append({
        'fold': fold_no,
        'lstm_mae': lstm_mae_fold,
        'lstm_rmse': lstm_rmse_fold,
        'lstm_r2': lstm_r2_fold,
        'naive_mae': naive_mae_fold,
        'naive_rmse': naive_rmse_fold,
        'mae_improvement_pct': improvement_mae
    })

    print(
        f"Fold {fold_no} | "
        f"LSTM MAE={lstm_mae_fold:.4f} | "
        f"Naive MAE={naive_mae_fold:.4f} | "
        f"Improvement={improvement_mae:+.2f}% | "
        f"R2={lstm_r2_fold:.4f}"
    )

    del fold_model

    if torch.cuda.is_available():
        torch.cuda.empty_cache()
        
lstm_robust_df = pd.DataFrame(lstm_robustness_results)

print("\n" + "=" * 78)
print(" FINAL LSTM ROBUSTNESS SUMMARY ")
print("=" * 78)

print(
    f"{'Metric':<22} | {'Mean':<10} | "
    f"{'Std':<10} | {'Min':<10} | {'Max':<10}"
)
print("-" * 72)

for metric in [
    'lstm_mae',
    'lstm_rmse',
    'lstm_r2',
    'naive_mae',
    'naive_rmse',
    'mae_improvement_pct'
]:
    values = lstm_robust_df[metric]

    print(
        f"{metric.upper():<22} | "
        f"{values.mean():<10.4f} | "
        f"{values.std():<10.4f} | "
        f"{values.min():<10.4f} | "
        f"{values.max():<10.4f}"
    )

lstm_wins = (
    lstm_robust_df['lstm_mae']
    < lstm_robust_df['naive_mae']
).sum()

print(
    f"\nLSTM beats Naive in "
    f"{lstm_wins}/{len(lstm_robust_df)} folds."
)

# ======================================================================
# PART 2: DISCRETE TARGET CLASSIFICATION FOR Impact & Likelihood
# MULTI-MODEL BENCHMARK (XGBoost, Gradient Boosting, Random Forest, Extra Trees)
# ======================================================================
print("\n======================================================================")
print(" PART 2: DISCRETE TARGET CLASSIFICATION (Impact & Likelihood) ")
print("======================================================================")

classification_summary = {}

def make_period_walk_forward_folds(data, n_splits=4):
    periods = np.array(sorted(data['periode_dt'].unique()))
    fold_size = max(1, len(periods) // (n_splits + 1))
    folds = []

    for i in range(1, n_splits + 1):
        train_end = fold_size * i
        val_end = min(train_end + fold_size, len(periods))

        train_periods = periods[:train_end]
        val_periods = periods[train_end:val_end]

        if len(val_periods) == 0:
            continue

        train_idx = np.where(data['periode_dt'].isin(train_periods))[0]
        val_idx = np.where(data['periode_dt'].isin(val_periods))[0]
        folds.append((train_idx, val_idx))

    return folds

for target_col in class_targets:
    print(f"\n" + "-" * 70)
    print(f" TRAINING & EVALUATING CLASSIFICATION MODELS FOR: {target_col} ")
    print("-" * 70)
    
    df_feat = df.sort_values(['Nama KPI', 'periode_dt']).copy()

    df_feat['target_class'] = (
        pd.to_numeric(df_feat[target_col], errors='coerce').astype(int) - 1
    )

    # Numeric actual KPI value
    df_feat[reg_target] = pd.to_numeric(df_feat[reg_target], errors='coerce')

    # Leakage-safe lag features within each KPI
    for lag in [1, 2, 3]:
        df_feat[f'lag_{lag}'] = df_feat.groupby('Nama KPI')[reg_target].shift(lag)

    df_feat['rolling_mean_3'] = df_feat.groupby('Nama KPI')[reg_target].transform(
        lambda s: s.shift(1).rolling(3).mean()
    )

    df_feat = df_feat.dropna(
        subset=[
            target_col, reg_target,
            'lag_1', 'lag_2', 'lag_3',
            'rolling_mean_3', 'periode_dt'
        ]
    ).reset_index(drop=True)

    # ----------------------------------------------------------------------
    # Chronological 80/20 split by PERIOD
    # ----------------------------------------------------------------------

    unique_periods = np.array(sorted(df_feat['periode_dt'].unique()))
    period_split_idx = int(len(unique_periods) * 0.80)

    train_periods = unique_periods[:period_split_idx]
    val_periods = unique_periods[period_split_idx:]

    train_df = df_feat[df_feat['periode_dt'].isin(train_periods)].copy()
    val_df = df_feat[df_feat['periode_dt'].isin(val_periods)].copy()

    tab_feature_cols = [
        'Nama KPI',
        reg_target,
        'lag_1',
        'lag_2',
        'lag_3',
        'rolling_mean_3'
    ]

    X_tr_t = train_df[tab_feature_cols]
    X_val_t = val_df[tab_feature_cols]

    y_tr_t = train_df['target_class'].values
    y_val_t = val_df['target_class'].values

    print(
        f"Chronological split: "
        f"Train {train_df['periode_dt'].min().strftime('%Y-%m')} → "
        f"{train_df['periode_dt'].max().strftime('%Y-%m')} | "
        f"Validation {val_df['periode_dt'].min().strftime('%Y-%m')} → "
        f"{val_df['periode_dt'].max().strftime('%Y-%m')}"
    )

    print(f"Train rows: {len(train_df)} | Validation rows: {len(val_df)}")

    categorical_tab = ['Nama KPI']
    numeric_tab = [
        reg_target,
        'lag_1',
        'lag_2',
        'lag_3',
        'rolling_mean_3'
    ]

    def build_kpi_classifier(model):
        preprocessor = ColumnTransformer([
            (
                'categorical',
                OneHotEncoder(handle_unknown='ignore', sparse_output=False),
                categorical_tab
            ),
            (
                'numeric',
                StandardScaler(),
                numeric_tab
            )
        ])

        return Pipeline([
            ('preprocessor', preprocessor),
            ('model', model)
        ])

    candidate_tab_models = {
        'XGBoost': XGBClassifier(
            eval_metric='mlogloss',
            random_state=42,
            n_jobs=1
        ),
        'Gradient Boosting': GradientBoostingClassifier(random_state=42),
        'Random Forest': RandomForestClassifier(
            n_estimators=200,
            random_state=42,
            n_jobs=1
        ),
        'Extra Trees': ExtraTreesClassifier(
            n_estimators=200,
            random_state=42,
            n_jobs=1
        )
    }

    results = []

    for m_name, model in candidate_tab_models.items():
        pipeline = build_kpi_classifier(model)

        pipeline.fit(X_tr_t, y_tr_t)
        y_pred = pipeline.predict(X_val_t)

        acc = accuracy_score(y_val_t, y_pred)
        f1 = f1_score(
            y_val_t,
            y_pred,
            average='weighted',
            zero_division=0
        )

        results.append({
            'model_name': m_name,
            'test_acc': acc,
            'weighted_f1': f1,
            'y_pred': y_pred
        })
    
    results_sorted = sorted(
        results,
        key=lambda r: (r['weighted_f1'], r['test_acc']),
        reverse=True
    )
    winner = results_sorted[0]
    
    print("\n--- Leakage-Safe Chronological Leaderboard ---")
    print(f"{'Rank':<5} | {'Model Name':<20} | {'Validation Accuracy':<20} | {'Weighted F1':<13}")
    print("-" * 68)

    for idx, r in enumerate(results_sorted):
        print(
            f"{idx+1:<5} | {r['model_name']:<20} | "
            f"{r['test_acc']*100:>6.2f} %             | "
            f"{r['weighted_f1']:>8.4f}"
        )
    
    # ======================================================================
    # ORDINAL REGRESSION BENCHMARK
    # Treat Impact / Likelihood as ordered classes
    # ======================================================================

    print("\n--- Ordinal Regression Benchmark ---")

    ordinal_models = {
        'XGB Regressor': XGBRegressor(
            n_estimators=300,
            max_depth=4,
            learning_rate=0.05,
            objective='reg:squarederror',
            random_state=42,
            n_jobs=1
        ),
        'Gradient Boosting Regressor': GradientBoostingRegressor(random_state=42),
        'Random Forest Regressor': RandomForestRegressor(
            n_estimators=300,
            random_state=42,
            n_jobs=1
        ),
        'Extra Trees Regressor': ExtraTreesRegressor(
            n_estimators=300,
            random_state=42,
            n_jobs=1
        )
    }

    ordinal_results = []

    for model_name, model in ordinal_models.items():
        pipeline = build_kpi_classifier(model)
        pipeline.fit(X_tr_t, y_tr_t)

        raw_pred = pipeline.predict(X_val_t)

        # target_class is 0..4
        ordinal_pred = np.clip(
            np.rint(raw_pred),
            0,
            4
        ).astype(int)

        ordinal_acc = accuracy_score(y_val_t, ordinal_pred)
        ordinal_f1 = f1_score(
            y_val_t,
            ordinal_pred,
            average='weighted',
            zero_division=0
        )

        ordinal_mae = mean_absolute_error(
            y_val_t,
            ordinal_pred
        )

        ordinal_results.append({
            'model_name': model_name,
            'accuracy': ordinal_acc,
            'weighted_f1': ordinal_f1,
            'mae': ordinal_mae,
            'y_pred': ordinal_pred
        })

    ordinal_results = sorted(
        ordinal_results,
        key=lambda r: (r['weighted_f1'], -r['mae']),
        reverse=True
    )

    print(
        f"{'Rank':<5} | {'Model':<28} | "
        f"{'Accuracy':<12} | {'Weighted F1':<12} | {'MAE':<8}"
    )
    print("-" * 78)

    for idx, r in enumerate(ordinal_results, start=1):
        print(
            f"{idx:<5} | {r['model_name']:<28} | "
            f"{r['accuracy']*100:>6.2f}%      | "
            f"{r['weighted_f1']:<12.4f} | "
            f"{r['mae']:<8.4f}"
        )

    ordinal_winner = ordinal_results[0]

    print(
        f"\n[ORDINAL WINNER] {ordinal_winner['model_name']} | "
        f"Accuracy: {ordinal_winner['accuracy']*100:.2f}% | "
        f"Weighted F1: {ordinal_winner['weighted_f1']:.4f} | "
        f"MAE: {ordinal_winner['mae']:.4f}"
    )
    
    print(f"\n--- Ordinal Classification Report ({ordinal_winner['model_name']}) ---")
    
    all_classes = np.array(sorted(np.unique(np.concatenate([y_tr_t, y_val_t]))))
    class_names = [str(c + 1) for c in all_classes]

    print(classification_report(
        y_val_t,
        ordinal_winner['y_pred'],
        labels=all_classes,
        target_names=class_names,
        zero_division=0
    ))
    
    if target_col == 'TARGET: Impact (1-5)':
        print("\n--- Fine Tuning XGB Ordinal Regressor: Impact ---")

        impact_param_space = {
            'n_estimators': [200, 300, 400, 500, 700],
            'max_depth': [2, 3, 4, 5, 6],
            'learning_rate': [0.01, 0.03, 0.05, 0.08, 0.10],
            'min_child_weight': [1, 2, 3, 5],
            'subsample': [0.7, 0.8, 0.9, 1.0],
            'colsample_bytree': [0.7, 0.8, 0.9, 1.0],
            'gamma': [0, 0.05, 0.1, 0.2],
            'reg_alpha': [0, 0.01, 0.05, 0.1],
            'reg_lambda': [1, 2, 3, 5]
        }

        tuning_df = train_df.reset_index(drop=True)
        tuning_folds = make_period_walk_forward_folds(tuning_df, n_splits=4)
        sampled_params = list(ParameterSampler(impact_param_space, n_iter=30, random_state=42))

        tuning_results = []

        for params in sampled_params:
            fold_f1 = []
            fold_mae = []

            for tr_idx, va_idx in tuning_folds:
                fold_train = tuning_df.iloc[tr_idx]
                fold_val = tuning_df.iloc[va_idx]

                X_fold_train = fold_train[tab_feature_cols]
                X_fold_val = fold_val[tab_feature_cols]
                y_fold_train = fold_train['target_class'].values
                y_fold_val = fold_val['target_class'].values

                model = XGBRegressor(
                    objective='reg:squarederror',
                    random_state=42,
                    n_jobs=1,
                    **params
                )

                pipeline = build_kpi_classifier(model)
                pipeline.fit(X_fold_train, y_fold_train)

                raw_pred = pipeline.predict(X_fold_val)
                pred = np.clip(np.rint(raw_pred), 0, 4).astype(int)

                fold_f1.append(f1_score(
                    y_fold_val, pred, average='weighted', zero_division=0
                ))
                fold_mae.append(mean_absolute_error(y_fold_val, pred))

            tuning_results.append({
                'params': params,
                'f1': np.mean(fold_f1),
                'mae': np.mean(fold_mae)
            })

        tuning_results = sorted(
            tuning_results,
            key=lambda r: (r['f1'], -r['mae']),
            reverse=True
        )

        best_impact = tuning_results[0]

        print(f"Best Walk-Forward F1 : {best_impact['f1']:.4f}")
        print(f"Best Walk-Forward MAE: {best_impact['mae']:.4f}")
        print("Best Parameters:")
        for k, v in best_impact['params'].items():
            print(f"  {k}: {v}")
            
        final_impact_model = XGBRegressor(
            objective='reg:squarederror',
            random_state=42,
            n_jobs=1,
            **best_impact['params']
        )

        final_impact_pipeline = build_kpi_classifier(final_impact_model)
        final_impact_pipeline.fit(X_tr_t, y_tr_t)

        impact_raw_pred = final_impact_pipeline.predict(X_val_t)
        impact_tuned_pred = np.clip(np.rint(impact_raw_pred), 0, 4).astype(int)

        impact_tuned_acc = accuracy_score(y_val_t, impact_tuned_pred)
        impact_tuned_f1 = f1_score(
            y_val_t, impact_tuned_pred,
            average='weighted', zero_division=0
        )
        impact_tuned_mae = mean_absolute_error(y_val_t, impact_tuned_pred)

        print("\n--- Tuned Impact Final Validation ---")
        print(f"Accuracy    : {impact_tuned_acc*100:.2f}%")
        print(f"Weighted F1 : {impact_tuned_f1:.4f}")
        print(f"Ordinal MAE : {impact_tuned_mae:.4f}")
    
    if target_col == 'TARGET: Likelihood (1-5)':
        print("\n--- Fine Tuning Balanced Gradient Boosting: Likelihood ---")

        likelihood_param_space = {
            'n_estimators': [100, 150, 200, 300, 400],
            'learning_rate': [0.01, 0.03, 0.05, 0.08, 0.10],
            'max_depth': [2, 3, 4, 5],
            'min_samples_split': [2, 4, 6, 8, 10],
            'min_samples_leaf': [1, 2, 3, 4, 5],
            'subsample': [0.7, 0.8, 0.9, 1.0],
            'max_features': [None, 'sqrt', 'log2']
        }

        tuning_df = train_df.reset_index(drop=True)
        tuning_folds = make_period_walk_forward_folds(tuning_df, n_splits=4)
        sampled_params = list(
            ParameterSampler(
                likelihood_param_space,
                n_iter=30,
                random_state=42
            )
        )

        likelihood_tuning_results = []

        for params in sampled_params:
            fold_weighted_f1 = []
            fold_macro_f1 = []

            for tr_idx, va_idx in tuning_folds:
                fold_train = tuning_df.iloc[tr_idx]
                fold_val = tuning_df.iloc[va_idx]

                X_fold_train = fold_train[tab_feature_cols]
                X_fold_val = fold_val[tab_feature_cols]

                y_fold_train = fold_train['target_class'].values
                y_fold_val = fold_val['target_class'].values

                sample_weights = compute_sample_weight(
                    class_weight='balanced',
                    y=y_fold_train
                )

                model = GradientBoostingClassifier(
                    random_state=42,
                    **params
                )

                pipeline = build_kpi_classifier(model)

                pipeline.fit(
                    X_fold_train,
                    y_fold_train,
                    model__sample_weight=sample_weights
                )

                pred = pipeline.predict(X_fold_val)

                fold_weighted_f1.append(
                    f1_score(
                        y_fold_val,
                        pred,
                        average='weighted',
                        zero_division=0
                    )
                )

                fold_macro_f1.append(
                    f1_score(
                        y_fold_val,
                        pred,
                        average='macro',
                        zero_division=0
                    )
                )

            likelihood_tuning_results.append({
                'params': params,
                'weighted_f1': np.mean(fold_weighted_f1),
                'macro_f1': np.mean(fold_macro_f1)
            })

        likelihood_tuning_results = sorted(
            likelihood_tuning_results,
            key=lambda r: (r['weighted_f1'], r['macro_f1']),
            reverse=True
        )

        best_likelihood = likelihood_tuning_results[0]

        print(
            f"Best Walk-Forward Weighted F1 : "
            f"{best_likelihood['weighted_f1']:.4f}"
        )
        print(
            f"Best Walk-Forward Macro F1    : "
            f"{best_likelihood['macro_f1']:.4f}"
        )

        print("Best Parameters:")
        for k, v in best_likelihood['params'].items():
            print(f"  {k}: {v}")

        # Final fit using all chronological training data
        final_likelihood_model = GradientBoostingClassifier(
            random_state=42,
            **best_likelihood['params']
        )

        final_likelihood_pipeline = build_kpi_classifier(
            final_likelihood_model
        )

        final_sample_weights = compute_sample_weight(
            class_weight='balanced',
            y=y_tr_t
        )

        final_likelihood_pipeline.fit(
            X_tr_t,
            y_tr_t,
            model__sample_weight=final_sample_weights
        )

        likelihood_tuned_pred = final_likelihood_pipeline.predict(X_val_t)

        likelihood_tuned_acc = accuracy_score(
            y_val_t,
            likelihood_tuned_pred
        )

        likelihood_tuned_f1 = f1_score(
            y_val_t,
            likelihood_tuned_pred,
            average='weighted',
            zero_division=0
        )

        likelihood_tuned_macro_f1 = f1_score(
            y_val_t,
            likelihood_tuned_pred,
            average='macro',
            zero_division=0
        )

        print("\n--- Tuned Likelihood Final Validation ---")
        print(f"Accuracy    : {likelihood_tuned_acc*100:.2f}%")
        print(f"Weighted F1 : {likelihood_tuned_f1:.4f}")
        print(f"Macro F1    : {likelihood_tuned_macro_f1:.4f}")

        print("\n--- Tuned Likelihood Classification Report ---")
        print(classification_report(
            y_val_t,
            likelihood_tuned_pred,
            labels=all_classes,
            target_names=class_names,
            zero_division=0
        ))
        
    # ======================================================================
    # LIKELIHOOD ROBUSTNESS COMPARISON
    # Baseline Gradient Boosting vs Balanced Tuned Gradient Boosting
    # ======================================================================

    if target_col == 'TARGET: Likelihood (1-5)':
        print("\n--- Likelihood Walk-Forward Robustness Comparison ---")

        robustness_df = train_df.reset_index(drop=True)
        robustness_folds = make_period_walk_forward_folds(
            robustness_df,
            n_splits=4
        )

        baseline_wf = []
        tuned_wf = []

        for fold_no, (tr_idx, va_idx) in enumerate(robustness_folds, start=1):
            fold_train = robustness_df.iloc[tr_idx]
            fold_val = robustness_df.iloc[va_idx]

            X_fold_train = fold_train[tab_feature_cols]
            X_fold_val = fold_val[tab_feature_cols]
            y_fold_train = fold_train['target_class'].values
            y_fold_val = fold_val['target_class'].values

            # --------------------------------------------------------------
            # Baseline Gradient Boosting
            # --------------------------------------------------------------
            baseline_model = GradientBoostingClassifier(random_state=42)
            baseline_pipeline = build_kpi_classifier(baseline_model)

            baseline_pipeline.fit(X_fold_train, y_fold_train)
            baseline_pred = baseline_pipeline.predict(X_fold_val)

            baseline_acc = accuracy_score(y_fold_val, baseline_pred)
            baseline_weighted_f1 = f1_score(
                y_fold_val,
                baseline_pred,
                average='weighted',
                zero_division=0
            )
            baseline_macro_f1 = f1_score(
                y_fold_val,
                baseline_pred,
                average='macro',
                zero_division=0
            )

            baseline_wf.append({
                'accuracy': baseline_acc,
                'weighted_f1': baseline_weighted_f1,
                'macro_f1': baseline_macro_f1
            })

            # --------------------------------------------------------------
            # Balanced Tuned Gradient Boosting
            # --------------------------------------------------------------
            tuned_model = GradientBoostingClassifier(
                random_state=42,
                **best_likelihood['params']
            )

            tuned_pipeline = build_kpi_classifier(tuned_model)

            fold_weights = compute_sample_weight(
                class_weight='balanced',
                y=y_fold_train
            )

            tuned_pipeline.fit(
                X_fold_train,
                y_fold_train,
                model__sample_weight=fold_weights
            )

            tuned_pred = tuned_pipeline.predict(X_fold_val)

            tuned_acc = accuracy_score(y_fold_val, tuned_pred)
            tuned_weighted_f1 = f1_score(
                y_fold_val,
                tuned_pred,
                average='weighted',
                zero_division=0
            )
            tuned_macro_f1 = f1_score(
                y_fold_val,
                tuned_pred,
                average='macro',
                zero_division=0
            )

            tuned_wf.append({
                'accuracy': tuned_acc,
                'weighted_f1': tuned_weighted_f1,
                'macro_f1': tuned_macro_f1
            })

            print(
                f"Fold {fold_no} | "
                f"Baseline F1={baseline_weighted_f1:.4f}, Macro={baseline_macro_f1:.4f} | "
                f"Tuned F1={tuned_weighted_f1:.4f}, Macro={tuned_macro_f1:.4f}"
            )

        baseline_wf_df = pd.DataFrame(baseline_wf)
        tuned_wf_df = pd.DataFrame(tuned_wf)

        print("\n--- Likelihood Robustness Summary ---")
        print(
            f"{'Metric':<15} | {'Baseline':<12} | "
            f"{'Balanced Tuned':<15} | {'Delta':<10}"
        )
        print("-" * 62)

        for metric in ['accuracy', 'weighted_f1', 'macro_f1']:
            baseline_mean = baseline_wf_df[metric].mean()
            tuned_mean = tuned_wf_df[metric].mean()

            print(
                f"{metric.upper():<15} | "
                f"{baseline_mean:<12.4f} | "
                f"{tuned_mean:<15.4f} | "
                f"{tuned_mean - baseline_mean:+.4f}"
            )

        baseline_wf_weighted = baseline_wf_df['weighted_f1'].mean()
        baseline_wf_macro = baseline_wf_df['macro_f1'].mean()

        tuned_wf_weighted = tuned_wf_df['weighted_f1'].mean()
        tuned_wf_macro = tuned_wf_df['macro_f1'].mean()

        # Select using walk-forward performance, not final validation
        likelihood_use_tuned = (
            tuned_wf_weighted > baseline_wf_weighted
            or (
                np.isclose(tuned_wf_weighted, baseline_wf_weighted, atol=0.005)
                and tuned_wf_macro > baseline_wf_macro
            )
        )

        print(
            "\nRobustness Winner: "
            + (
                "Balanced Tuned Gradient Boosting"
                if likelihood_use_tuned
                else "Baseline Gradient Boosting"
            )
        )
    
    # ----------------------------------------------------------------------
    # Final target model selection
    # ----------------------------------------------------------------------

    if target_col == 'TARGET: Impact (1-5)':
        if impact_tuned_f1 >= ordinal_winner['weighted_f1']:
            final_name = 'Tuned XGB Regressor'
            final_acc = impact_tuned_acc
            final_f1 = impact_tuned_f1
            final_pred = impact_tuned_pred
        else:
            final_name = ordinal_winner['model_name']
            final_acc = ordinal_winner['accuracy']
            final_f1 = ordinal_winner['weighted_f1']
            final_pred = ordinal_winner['y_pred']

    elif target_col == 'TARGET: Likelihood (1-5)':
        if likelihood_use_tuned:
            final_name = 'Balanced Tuned Gradient Boosting'
            final_acc = likelihood_tuned_acc
            final_f1 = likelihood_tuned_f1
            final_pred = likelihood_tuned_pred
        else:
            final_name = winner['model_name']
            final_acc = winner['test_acc']
            final_f1 = winner['weighted_f1']
            final_pred = winner['y_pred']

    print(
        f"\n[FINAL WINNER] Best Model for {target_col}: "
        f"{final_name} (Accuracy: {final_acc*100:.2f}% | "
        f"Weighted F1: {final_f1:.4f})"
    )

    print(f"\n--- Detailed Classification Report ({final_name}) ---")
    print(classification_report(
        y_val_t,
        final_pred,
        labels=all_classes,
        target_names=class_names,
        zero_division=0
    ))

    classification_summary[target_col] = {
        'model_name': final_name,
        'test_acc': final_acc,
        'weighted_f1': final_f1,
        'y_pred': final_pred
    }

# ======================================================================
# OVERALL ACCURACY SUMMARY REPORT
# ======================================================================

print("\n======================================================================")
print(" FINAL IMPROVED ACCURACY SUMMARY REPORT FOR ALL TARGETS ")
print("======================================================================")
print(f" 1. TARGET: Nilai Aktual (%)   | Model: PyTorch LSTM Regressor | Acc: {acc_r:.2f} % | MAE: {mae_r:.4f}")

summary_no = 2
for target_name, win in classification_summary.items():
    print(f" {summary_no}. {target_name:<27} | Model: {win['model_name']:<35} | Acc: {win['test_acc']*100:.2f} % | Weighted F1: {win['weighted_f1']:.4f}")
    summary_no += 1

print("\nKPI Multi-Model Improvement Pipeline completed successfully!")
