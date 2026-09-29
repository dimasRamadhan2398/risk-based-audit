# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 05 — CAATT: Cross-System Reconciliation
# Reconcile CBS transactions vs GL entries to find discrepancies.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd

dl = get_datalake_engine()

# %% [markdown]
# ## CBS Transactions vs GL Entries — Daily Totals

# %%
df_trx = pd.read_sql("""
    SELECT trx_date, SUM(amount) as trx_total, COUNT(*) as trx_count
    FROM gold.fact_transactions
    GROUP BY trx_date ORDER BY trx_date
""", dl)

df_gl = pd.read_sql("""
    SELECT gl_date as trx_date,
           SUM(debit_amount) as gl_debit_total,
           SUM(credit_amount) as gl_credit_total,
           COUNT(*) as gl_count
    FROM gold.fact_gl_entries
    GROUP BY gl_date ORDER BY gl_date
""", dl)

df_recon = pd.merge(df_trx, df_gl, on="trx_date", how="outer")
df_recon['difference'] = (df_recon['trx_total'].fillna(0) - df_recon['gl_debit_total'].fillna(0)).abs()
df_recon['matched'] = df_recon['difference'] < 1000  # tolerance Rp 1.000

matched_count = df_recon['matched'].sum()
total_days = len(df_recon)
match_rate = round(matched_count / total_days * 100, 2) if total_days > 0 else 0

print(f"Days analyzed: {total_days}")
print(f"Days matched: {matched_count}")
print(f"Match rate: {match_rate}%")
print(f"\nUnmatched days:")
print(df_recon[~df_recon['matched']].head(20))

# %% [markdown]
# ## Save Reconciliation Results

# %%
results = [{
    "source_a": "CBS Transactions",
    "source_b": "General Ledger",
    "matched_count": int(matched_count),
    "unmatched_a": int(df_recon['trx_total'].isna().sum()),
    "unmatched_b": int(df_recon['gl_debit_total'].isna().sum()),
    "match_rate_pct": match_rate,
    "total_difference": float(df_recon['difference'].sum()),
}]

pd.DataFrame(results).to_sql("caatt_reconciliation_results", dl, schema="gold", if_exists="append", index=False)
print("✅ Reconciliation results saved")
