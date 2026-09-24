# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
#   kernelspec:
#     display_name: Python 3
#     language: python
#     name: python3
# ---

# %% [markdown]
# # Notebook 01 — CAATT: Full Population Testing
# 
# Test 100% of transactions against business rules and policy limits.
# No sampling = no sampling error.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd
from sqlalchemy import text

dl = get_datalake_engine()

# %% [markdown]
# ## Test 1: Transactions exceeding authorization limits

# %%
df_trx = pd.read_sql("""
    SELECT t.*, l.plafond, l.outstanding
    FROM gold.fact_transactions t
    LEFT JOIN gold.fact_loans l ON l.branch_name = t.branch_name
    WHERE t.amount > 500000000  -- > Rp 500 juta
    ORDER BY t.amount DESC
""", dl)
print(f"Transactions > Rp 500 juta: {len(df_trx)}")
df_trx.head(20)

# %% [markdown]
# ## Test 2: Loans exceeding product plafond policy

# %%
df_loans = pd.read_sql("""
    SELECT loan_number, loan_type, customer_name, branch_name,
           plafond, outstanding, interest_rate,
           rate_within_policy, plafond_within_policy, tenor_within_policy
    FROM gold.fact_loans
    WHERE plafond_within_policy = FALSE
       OR rate_within_policy = FALSE
       OR tenor_within_policy = FALSE
    ORDER BY plafond DESC
""", dl)
print(f"Loans violating policy: {len(df_loans)}")
df_loans.head(20)

# %% [markdown]
# ## Test 3: After-hours transactions (all)

# %%
df_after = pd.read_sql("""
    SELECT trx_ref_number, trx_date, trx_hour, amount, branch_name,
           customer_name, channel, authorization_status
    FROM gold.fact_transactions
    WHERE is_after_hours = TRUE
    ORDER BY amount DESC
""", dl)
print(f"After-hours transactions: {len(df_after)}")
df_after.head(20)

# %% [markdown]
# ## Save Results to Gold Zone

# %%
results = [
    {"test_name": "Trx > 500 Juta", "total_population": 10000, "violations_found": len(df_trx),
     "violation_rate": round(len(df_trx)/10000, 4), "criteria": "amount > 500000000"},
    {"test_name": "Loan Policy Violation", "total_population": 150, "violations_found": len(df_loans),
     "violation_rate": round(len(df_loans)/150, 4) if len(df_loans) > 0 else 0, "criteria": "rate/plafond/tenor out of policy"},
    {"test_name": "After Hours Trx", "total_population": 10000, "violations_found": len(df_after),
     "violation_rate": round(len(df_after)/10000, 4), "criteria": "trx_hour < 7 OR trx_hour >= 20"},
]

df_results = pd.DataFrame(results)
df_results.to_sql("caatt_full_population_results", dl, schema="gold", if_exists="append", index=False)
print("✅ Full population results saved to gold.caatt_full_population_results")
