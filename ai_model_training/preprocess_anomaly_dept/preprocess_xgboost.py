from pathlib import Path

import numpy as np
import pandas as pd


# configure file paths
base_dir = Path(__file__).resolve().parent
input_path = base_dir / "data" / "xgboost.csv"
output_dir = base_dir / "outputs"
output_path = output_dir / "xgboost_cleaned_unique.csv"


# load the raw dataset
df = pd.read_csv(input_path)
initial_rows = len(df)


# rename columns into consistent snake case
df = df.rename(
    columns={
        "Entitas": "entity",
        "Kategori Risiko": "risk_category",
        "Inherent Likelihood": "inherent_likelihood",
        "Inherent Impact": "inherent_impact",
        "Jml Temuan Audit": "audit_findings_count",
        "Jml KPI di Bawah Target": "kpi_below_target_count",
        "Volatilitas KPI": "kpi_volatility",
        "Skor Risiko Periode Lalu": "previous_risk_score",
        "Bulan Penilaian": "assessment_month",
        "TARGET: Likelihood": "target_likelihood",
        "TARGET: Impact": "target_impact",
    }
)


# remove surrounding spaces from categorical columns
categorical_columns = ["entity", "risk_category"]
df[categorical_columns] = df[categorical_columns].apply(
    lambda column: column.astype("string").str.strip()
)


# convert all measurement and target columns into numeric values
numeric_columns = [
    "inherent_likelihood",
    "inherent_impact",
    "audit_findings_count",
    "kpi_below_target_count",
    "kpi_volatility",
    "previous_risk_score",
    "assessment_month",
    "target_likelihood",
    "target_impact",
]
df[numeric_columns] = df[numeric_columns].apply(
    pd.to_numeric,
    errors="coerce",
)


# replace values outside the valid assessment month range
df.loc[~df["assessment_month"].between(1, 12), "assessment_month"] = np.nan


# remove fully duplicated observations
duplicate_rows = int(df.duplicated().sum())
df = df.drop_duplicates().reset_index(drop=True)


# stop the process when required values are missing or invalid
required_columns = categorical_columns + numeric_columns
if df[required_columns].isna().any().any():
    missing_counts = df[required_columns].isna().sum()
    raise ValueError(
        "missing or invalid values found:\n"
        f"{missing_counts[missing_counts > 0]}"
    )


# create cyclical month features
df["month_sin"] = np.sin(
    2 * np.pi * (df["assessment_month"] - 1) / 12
)
df["month_cos"] = np.cos(
    2 * np.pi * (df["assessment_month"] - 1) / 12
)


# arrange one shared dataset for both xgboost targets
final_columns = [
    "entity",
    "risk_category",
    "inherent_likelihood",
    "inherent_impact",
    "audit_findings_count",
    "kpi_below_target_count",
    "kpi_volatility",
    "previous_risk_score",
    "assessment_month",
    "month_sin",
    "month_cos",
    "target_likelihood",
    "target_impact",
]
df = df[final_columns].sort_values(
    by=["assessment_month", "entity", "risk_category"]
).reset_index(drop=True)


# save one clean dataset for likelihood and impact training
output_dir.mkdir(parents=True, exist_ok=True)
df.to_csv(output_path, index=False)


# display a concise preprocessing summary
print("xgboost preprocessing completed")
print(f"input rows       : {initial_rows}")
print(f"duplicates removed: {duplicate_rows}")
print(f"output rows      : {len(df)}")
print(f"output file      : {output_path}")
