# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 10 — CAATT: Policy & Rule Engine
# Test transactions and loans against banking product policies.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd

dl = get_datalake_engine()

# %% [markdown]
# ## Test 1: Interest Rate Violations

# %%
df_rate = pd.read_sql("""
    SELECT loan_number, loan_type, customer_name, branch_name,
           interest_rate, product_name
    FROM gold.fact_loans
    WHERE rate_within_policy = FALSE
""", dl)
print(f"Interest rate violations: {len(df_rate)}")
df_rate.head(20)

# %% [markdown]
# ## Test 2: Plafond Exceeding Policy Maximum

# %%
df_plafond = pd.read_sql("""
    SELECT loan_number, loan_type, customer_name, branch_name,
           plafond, product_name
    FROM gold.fact_loans
    WHERE plafond_within_policy = FALSE
""", dl)
print(f"Plafond violations: {len(df_plafond)}")
df_plafond.head(20)

# %% [markdown]
# ## Test 3: Unauthorized Override Transactions

# %%
df_override = pd.read_sql("""
    SELECT trx_ref_number, trx_date, amount, branch_name,
           authorization_status, channel
    FROM gold.fact_transactions
    WHERE authorization_status = 'Override'
    ORDER BY amount DESC
""", dl)
print(f"Override transactions: {len(df_override)}")

# %% [markdown]
# ## Save All Policy Violations

# %%
violations = []

for _, row in df_rate.iterrows():
    violations.append({
        "violation_type": "RATE_VIOLATION",
        "rule_name": "Interest Rate Out of Policy Range",
        "reference_id": None,
        "reference_type": "LOAN",
        "actual_value": float(row["interest_rate"]),
        "severity": "High",
        "branch_name": row["branch_name"],
    })

for _, row in df_plafond.iterrows():
    violations.append({
        "violation_type": "PLAFOND_VIOLATION",
        "rule_name": "Plafond Exceeds Product Maximum",
        "reference_id": None,
        "reference_type": "LOAN",
        "actual_value": float(row["plafond"]),
        "severity": "Critical",
        "branch_name": row["branch_name"],
    })

for _, row in df_override.iterrows():
    violations.append({
        "violation_type": "AUTH_OVERRIDE",
        "rule_name": "Unauthorized Override Transaction",
        "reference_id": None,
        "reference_type": "TRANSACTION",
        "actual_value": float(row["amount"]),
        "severity": "High" if row["amount"] > 500000000 else "Medium",
        "branch_name": row["branch_name"],
    })

if violations:
    pd.DataFrame(violations).to_sql("caatt_policy_violations", dl, schema="gold", if_exists="append", index=False)
    print(f"✅ Saved {len(violations)} policy violations")
