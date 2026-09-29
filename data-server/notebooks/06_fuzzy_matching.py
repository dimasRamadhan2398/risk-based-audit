# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 06 — CAATT: Fuzzy Matching
# Detect suspiciously similar vendor/customer names (possible duplicates or shell entities).

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd
from fuzzywuzzy import fuzz, process

dl = get_datalake_engine()

# %% [markdown]
# ## Load Customer Names

# %%
df = pd.read_sql("SELECT DISTINCT customer_name FROM gold.fact_transactions WHERE customer_name IS NOT NULL", dl)
names = df['customer_name'].tolist()
print(f"Unique customer names: {len(names)}")

# %% [markdown]
# ## Find Similar Name Pairs (Threshold: 85%)

# %%
similar_pairs = []
for i, name1 in enumerate(names):
    for name2 in names[i+1:]:
        ratio = fuzz.ratio(name1, name2)
        if 85 <= ratio < 100:  # Similar but not identical
            similar_pairs.append({"name_1": name1, "name_2": name2, "similarity_pct": ratio})

df_similar = pd.DataFrame(similar_pairs).sort_values("similarity_pct", ascending=False)
print(f"Similar name pairs found: {len(df_similar)}")
df_similar.head(20)

# %% [markdown]
# ## Save for Review
# %%
if len(df_similar) > 0:
    # Save as CAATT exceptions for auditor review
    for _, row in df_similar.iterrows():
        pd.DataFrame([{
            "result_type": "DUPLICATE",
            "reference_field": "customer_name",
            "reference_value": f"{row['name_1']} ≈ {row['name_2']}",
            "duplicate_count": int(row['similarity_pct']),
        }]).to_sql("caatt_duplicate_gap_results", dl, schema="gold", if_exists="append", index=False)
    print(f"✅ Saved {len(df_similar)} fuzzy match results")
