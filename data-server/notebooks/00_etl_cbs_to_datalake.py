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
# # Notebook 00 — ETL: CBS → Data Lake (Bronze → Silver → Gold)
# 
# Runs the full ETL pipeline from CBS Simulator into the Data Lake.
# This is the starting point before running any CAATT technique.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_cbs_engine, get_datalake_engine, get_datahub_url
import pandas as pd
import requests
from sqlalchemy import text

cbs = get_cbs_engine()
dl = get_datalake_engine()
hub_url = get_datahub_url()

# %% [markdown]
# ## Step 1: Verify CBS data

# %%
with cbs.connect() as conn:
    for table in ['branches', 'customers', 'accounts', 'transactions', 'loans', 'gl_entries', 'products', 'documents']:
        r = conn.execute(text(f"SELECT COUNT(*) FROM cbs.{table}"))
        print(f"  cbs.{table}: {r.scalar()} rows")

# %% [markdown]
# ## Step 2: Trigger ETL Pipeline via Data Hub API

# %%
api_key = "YOUR_DATA_HUB_API_KEY"  # Replace or load from env
response = requests.post(f"{hub_url}/api/v1/pipeline/run", headers={"X-API-Key": api_key})
print(response.json())

# %% [markdown]
# ## Step 3: Check Pipeline Status

# %%
import time
for i in range(30):
    r = requests.get(f"{hub_url}/api/v1/pipeline/status", headers={"X-API-Key": api_key})
    status = r.json()
    print(f"  [{i*10}s] Status: {status['status']}")
    if status['status'] in ('completed', 'failed'):
        break
    time.sleep(10)
print(f"\nFinal: {status}")

# %% [markdown]
# ## Step 4: Verify Gold Zone

# %%
with dl.connect() as conn:
    for table in ['dim_branches', 'fact_transactions', 'fact_loans', 'fact_gl_entries']:
        r = conn.execute(text(f"SELECT COUNT(*) FROM gold.{table}"))
        print(f"  gold.{table}: {r.scalar()} rows")

print("\n✅ ETL Pipeline complete. Gold zone is ready for CAATT analysis.")
