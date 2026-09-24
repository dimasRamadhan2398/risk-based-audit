# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 04 — CAATT: Stratification & Aging Analysis
# Stratify transactions by amount ranges and analyze loan aging.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd

dl = get_datalake_engine()

# %% [markdown]
# ## Transaction Stratification by Amount

# %%
strata = [
    ("0 - 10 Juta", 0, 10000000),
    ("10 - 50 Juta", 10000000, 50000000),
    ("50 - 100 Juta", 50000000, 100000000),
    ("100 - 500 Juta", 100000000, 500000000),
    ("500 Juta - 1 Miliar", 500000000, 1000000000),
    ("> 1 Miliar", 1000000000, 999999999999),
]

results = []
for label, min_v, max_v in strata:
    df = pd.read_sql(f"""
        SELECT COUNT(*) as cnt, COALESCE(SUM(amount), 0) as total
        FROM gold.fact_transactions
        WHERE amount >= {min_v} AND amount < {max_v}
    """, dl)
    results.append({
        "stratum_label": label, "min_value": min_v, "max_value": max_v,
        "trx_count": int(df['cnt'].iloc[0]), "total_amount": float(df['total'].iloc[0]),
    })

df_strata = pd.DataFrame(results)
total_all = df_strata['total_amount'].sum()
df_strata['pct_of_total'] = (df_strata['total_amount'] / total_all * 100).round(2) if total_all > 0 else 0
print(df_strata)

# %% [markdown]
# ## Loan Aging by Collectability

# %%
df_aging = pd.read_sql("""
    SELECT collectability, status,
           COUNT(*) as loan_count,
           SUM(outstanding) as total_outstanding,
           AVG(days_past_due) as avg_dpd,
           MAX(days_past_due) as max_dpd
    FROM gold.fact_loans
    GROUP BY collectability, status
    ORDER BY collectability
""", dl)
print(df_aging)

# %% [markdown]
# ## Save Results

# %%
df_strata.to_sql("caatt_stratification_results", dl, schema="gold", if_exists="append", index=False)
print("✅ Stratification results saved")
