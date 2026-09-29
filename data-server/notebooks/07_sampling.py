# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 07 — CAATT: Statistical Sampling
# Generate audit samples using various sampling methods.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd
import numpy as np

dl = get_datalake_engine()

# %% [markdown]
# ## Load Transaction Data

# %%
df = pd.read_sql("SELECT * FROM gold.fact_transactions ORDER BY trx_date", dl)
print(f"Total transactions: {len(df)}")

# %% [markdown]
# ## Method 1: Simple Random Sampling

# %%
sample_size = min(100, len(df))
df_random = df.sample(n=sample_size, random_state=42)
print(f"Random sample: {len(df_random)} transactions")
print(f"  Amount range: {df_random['amount'].min():,.0f} - {df_random['amount'].max():,.0f}")

# %% [markdown]
# ## Method 2: Stratified Sampling (by branch)

# %%
df_stratified = df.groupby('branch_name', group_keys=False).apply(
    lambda x: x.sample(min(10, len(x)), random_state=42)
)
print(f"Stratified sample: {len(df_stratified)} transactions across {df_stratified['branch_name'].nunique()} branches")

# %% [markdown]
# ## Method 3: Monetary Unit Sampling (MUS)

# %%
df_sorted = df.sort_values('amount', ascending=False)
cumsum = df_sorted['amount'].cumsum()
total = df_sorted['amount'].sum()
interval = total / sample_size

mus_indices = []
next_threshold = np.random.uniform(0, interval)
for i, val in enumerate(cumsum):
    while val >= next_threshold and len(mus_indices) < sample_size:
        mus_indices.append(i)
        next_threshold += interval

df_mus = df_sorted.iloc[mus_indices[:sample_size]]
print(f"MUS sample: {len(df_mus)} transactions")
print(f"  Covers {df_mus['amount'].sum() / total * 100:.1f}% of total monetary value")

# %% [markdown]
# ## Method 4: High-Value Selection (top N)

# %%
df_highvalue = df.nlargest(50, 'amount')
print(f"Top 50 by amount: total = {df_highvalue['amount'].sum():,.0f}")

print("\n✅ Sampling complete. Export samples for fieldwork.")
