# ---
# jupyter:
#   jupytext:
#     text_representation:
#       extension: .py
#       format_name: percent
# ---

# %% [markdown]
# # Notebook 03 — CAATT: Benford's Law Analysis
# Detect anomalous first-digit distribution in transaction amounts.

# %%
import sys
sys.path.insert(0, '/home/jovyan/notebooks/helpers')
from db_connection import get_datalake_engine
import pandas as pd
import numpy as np
from scipy import stats
import matplotlib.pyplot as plt

dl = get_datalake_engine()

# %% [markdown]
# ## Expected Benford Distribution

# %%
benford_expected = {d: np.log10(1 + 1/d) * 100 for d in range(1, 10)}
print("Benford's Expected Distribution:")
for d, pct in benford_expected.items():
    print(f"  Digit {d}: {pct:.2f}%")

# %% [markdown]
# ## Actual First-Digit Distribution

# %%
df = pd.read_sql("""
    SELECT benford_first_digit, COUNT(*) as count
    FROM gold.fact_transactions
    WHERE benford_first_digit BETWEEN 1 AND 9
    GROUP BY benford_first_digit
    ORDER BY benford_first_digit
""", dl)

total = df['count'].sum()
df['actual_pct'] = (df['count'] / total * 100).round(2)
df['expected_pct'] = df['benford_first_digit'].map(benford_expected).round(2)
df['deviation_pct'] = (df['actual_pct'] - df['expected_pct']).round(2)
print(df)

# %% [markdown]
# ## Chi-Square Test

# %%
observed = df['count'].values
expected = np.array([benford_expected[d] / 100 * total for d in range(1, 10)])

chi2, p_value = stats.chisquare(observed, expected)
print(f"Chi-Square: {chi2:.4f}")
print(f"P-Value: {p_value:.6f}")
print(f"Significant at 5%? {'YES — Anomaly detected!' if p_value < 0.05 else 'No — Distribution is normal'}")

# %% [markdown]
# ## Visualization

# %%
fig, ax = plt.subplots(figsize=(10, 6))
x = range(1, 10)
ax.bar([d - 0.15 for d in x], df['expected_pct'], 0.3, label='Benford Expected', color='steelblue', alpha=0.7)
ax.bar([d + 0.15 for d in x], df['actual_pct'], 0.3, label='Actual', color='coral', alpha=0.7)
ax.set_xlabel('First Digit')
ax.set_ylabel('Percentage (%)')
ax.set_title("Benford's Law Analysis — Transaction Amounts")
ax.legend()
ax.set_xticks(x)
plt.tight_layout()
plt.savefig('/home/jovyan/work/benford_analysis.png', dpi=150)
plt.show()

# %% [markdown]
# ## Save to Gold Zone

# %%
results = []
for _, row in df.iterrows():
    results.append({
        "digit": int(row["benford_first_digit"]),
        "expected_pct": float(row["expected_pct"]),
        "actual_pct": float(row["actual_pct"]),
        "deviation_pct": float(row["deviation_pct"]),
        "chi_square": round(chi2, 4),
        "is_significant": p_value < 0.05,
        "field_tested": "amount",
        "sample_size": int(total),
    })

pd.DataFrame(results).to_sql("caatt_benford_results", dl, schema="gold", if_exists="append", index=False)
print("✅ Benford results saved to gold.caatt_benford_results")
