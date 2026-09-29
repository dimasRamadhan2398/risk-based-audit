# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 11 — CAATT + AI: Anomaly Detection with Gold Zone Data
# Uses Isolation Forest on enriched Gold zone data to detect anomalies.
# Results feed back into the Gold zone for AI retraining.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd
import numpy as np
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler

dl = get_datalake_engine()

# %% [markdown]
# ## Load Gold Zone Transactions

# %%
df = pd.read_sql("""
    SELECT trx_ref_number, amount_millions, trx_hour, trx_day_of_week,
           is_round_amount, is_after_hours, category, branch_name, is_suspicious
    FROM gold.fact_transactions
    WHERE amount_millions IS NOT NULL
""", dl)
print(f"Total transactions: {len(df)}")

# %% [markdown]
# ## Feature Engineering

# %%
features = ['amount_millions', 'trx_hour', 'trx_day_of_week']
df['is_round_int'] = df['is_round_amount'].astype(int)
df['is_after_int'] = df['is_after_hours'].astype(int)
features += ['is_round_int', 'is_after_int']

X = df[features].fillna(0)
scaler = StandardScaler()
X_scaled = scaler.fit_transform(X)

# %% [markdown]
# ## Isolation Forest

# %%
model = IsolationForest(contamination=0.05, random_state=42, n_estimators=200)
df['anomaly_prediction'] = model.fit_predict(X_scaled)
df['anomaly_score'] = -model.decision_function(X_scaled)  # Higher = more anomalous
df['is_anomaly_pred'] = df['anomaly_prediction'] == -1

true_anomalies = df['is_anomaly_pred'].sum()
print(f"Anomalies detected: {true_anomalies} ({true_anomalies/len(df)*100:.2f}%)")

# %% [markdown]
# ## Compare with Known Suspicious

# %%
comparison = pd.crosstab(df['is_suspicious'], df['is_anomaly_pred'], margins=True)
print(comparison)

# %% [markdown]
# ## Write Anomaly Training Data to Gold Zone (Feedback Loop)

# %%
training_data = df[df['is_anomaly_pred']].head(200)
for _, row in training_data.iterrows():
    pd.DataFrame([{
        "entity": row.get("branch_name", "Unknown"),
        "description": f"CAATT anomaly: {row.get('category', 'Unknown')}",
        "amount_millions": float(row["amount_millions"]),
        "hour_of_day": int(row["trx_hour"]),
        "day_of_week": int(row["trx_day_of_week"]),
        "is_round_amount": int(row["is_round_int"]),
        "is_anomaly": True,
    }]).to_sql("anomaly_training_data", dl, schema="gold", if_exists="append", index=False)

print(f"✅ Wrote {len(training_data)} anomaly training records to gold.anomaly_training_data")
print("   → AI Engine will auto-retrain on next scheduled cycle")
