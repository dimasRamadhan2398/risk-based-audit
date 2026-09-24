# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 08 — CAATT: Continuous Monitoring Alerts
# Automated threshold-based monitoring of key banking metrics.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd
from datetime import datetime, timedelta

dl = get_datalake_engine()

# %% [markdown]
# ## Define Alert Thresholds

# %%
thresholds = {
    "daily_trx_count_max": 500,
    "daily_suspicious_max": 5,
    "single_trx_amount_max": 1000000000,  # 1 Miliar
    "after_hours_pct_max": 5.0,
    "npl_ratio_max": 5.0,
    "voucher_gap_pct_max": 2.0,
}

# %% [markdown]
# ## Check Daily Transaction Volume

# %%
df_daily = pd.read_sql("""
    SELECT trx_date, COUNT(*) as cnt,
           SUM(CASE WHEN is_suspicious THEN 1 ELSE 0 END) as suspicious_cnt,
           SUM(CASE WHEN is_after_hours THEN 1 ELSE 0 END) as after_hours_cnt
    FROM gold.fact_transactions
    GROUP BY trx_date ORDER BY trx_date DESC LIMIT 30
""", dl)

alerts = []
for _, row in df_daily.iterrows():
    if row['suspicious_cnt'] > thresholds['daily_suspicious_max']:
        alerts.append(f"⚠️ {row['trx_date']}: {row['suspicious_cnt']} suspicious transactions (threshold: {thresholds['daily_suspicious_max']})")
    after_pct = row['after_hours_cnt'] / row['cnt'] * 100 if row['cnt'] > 0 else 0
    if after_pct > thresholds['after_hours_pct_max']:
        alerts.append(f"⚠️ {row['trx_date']}: {after_pct:.1f}% after-hours (threshold: {thresholds['after_hours_pct_max']}%)")

print(f"Alerts generated: {len(alerts)}")
for a in alerts[:20]:
    print(f"  {a}")

# %% [markdown]
# ## NPL Ratio Monitoring

# %%
df_npl = pd.read_sql("""
    SELECT
        COUNT(*) as total_loans,
        SUM(CASE WHEN collectability >= 3 THEN 1 ELSE 0 END) as npl_loans,
        SUM(CASE WHEN collectability >= 3 THEN outstanding ELSE 0 END) as npl_outstanding,
        SUM(outstanding) as total_outstanding
    FROM gold.fact_loans
""", dl)

npl_ratio = (df_npl['npl_loans'].iloc[0] / df_npl['total_loans'].iloc[0] * 100) if df_npl['total_loans'].iloc[0] > 0 else 0
print(f"NPL Ratio: {npl_ratio:.2f}% (threshold: {thresholds['npl_ratio_max']}%)")
if npl_ratio > thresholds['npl_ratio_max']:
    alerts.append(f"🔴 NPL Ratio {npl_ratio:.2f}% exceeds threshold {thresholds['npl_ratio_max']}%!")

print(f"\nTotal alerts: {len(alerts)}")
print("✅ Continuous monitoring check complete")
