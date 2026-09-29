# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 02 — CAATT: Duplicate & Gap Detection
# Find duplicate transactions and missing document/voucher numbers.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd
from sqlalchemy import text

dl = get_datalake_engine()

# %% [markdown]
# ## Duplicate Transaction Detection

# %%
df_dupes = pd.read_sql("""
    SELECT trx_ref_number, trx_date, amount, branch_name, channel, COUNT(*) as occurrence
    FROM gold.fact_transactions
    GROUP BY trx_ref_number, trx_date, amount, branch_name, channel
    HAVING COUNT(*) > 1
    ORDER BY occurrence DESC
""", dl)
print(f"Duplicate transaction groups: {len(df_dupes)}")
df_dupes.head(20)

# %% [markdown]
# ## GL Voucher Number Gaps

# %%
df_gaps = pd.read_sql("""
    SELECT gl_id, gl_date, voucher_number, gl_account_name, branch_name, has_voucher
    FROM gold.fact_gl_entries
    WHERE has_gap = TRUE
    ORDER BY gl_date
""", dl)
print(f"GL entries with missing vouchers: {len(df_gaps)}")
df_gaps.head(20)

# %% [markdown]
# ## Save Results

# %%
results = []
for _, row in df_dupes.iterrows():
    results.append({
        "result_type": "DUPLICATE",
        "reference_field": "trx_ref_number",
        "reference_value": row["trx_ref_number"],
        "duplicate_count": int(row["occurrence"]),
        "branch_name": row["branch_name"],
    })
for _, row in df_gaps.iterrows():
    results.append({
        "result_type": "GAP",
        "reference_field": "voucher_number",
        "reference_value": str(row.get("voucher_number", "MISSING")),
        "gap_start": str(row["gl_id"]),
        "branch_name": row["branch_name"],
    })

if results:
    pd.DataFrame(results).to_sql("caatt_duplicate_gap_results", dl, schema="gold", if_exists="append", index=False)
    print(f"✅ Saved {len(results)} results to gold.caatt_duplicate_gap_results")
