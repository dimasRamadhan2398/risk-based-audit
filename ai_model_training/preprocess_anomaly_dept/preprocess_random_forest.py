from pathlib import Path

import numpy as np
import pandas as pd


# configure file paths
base_dir = Path(__file__).resolve().parent
input_path = base_dir / "data" / "isolation_forest.csv"
output_dir = base_dir / "outputs"
output_path = output_dir / "random_forest_cleaned.csv"


# load the raw dataset
df = pd.read_csv(input_path)
initial_rows = len(df)


# rename columns into consistent snake case
df = df.rename(
    columns={
        "ID Transaksi": "transaction_id",
        "Entitas": "entity",
        "Deskripsi": "description",
        "amount (dalam Juta Rp)": "amount",
        "hour_of_day (0-23)": "hour",
        "day_of_week (1-7)": "day_of_week",
        "is_new_beneficiary (1=Ya, 0=Tidak)": "is_new_beneficiary",
        "is_round_amount (1=Ya, 0=Tidak)": "is_round_amount",
        "is_anomaly (Ground Truth)": "is_anomaly",
    }
)


# remove surrounding spaces from text columns
text_columns = ["entity", "description", "is_anomaly"]
df[text_columns] = df[text_columns].apply(
    lambda column: column.astype("string").str.strip()
)


# convert the transaction amount into numeric values
df["amount"] = pd.to_numeric(
    df["amount"].astype(str).str.replace(",", "", regex=False),
    errors="coerce",
)


# convert the remaining numeric columns safely
numeric_columns = [
    "hour",
    "day_of_week",
    "is_new_beneficiary",
    "is_round_amount",
]
df[numeric_columns] = df[numeric_columns].apply(
    pd.to_numeric,
    errors="coerce",
)


# replace values outside the valid ranges with missing values
df.loc[~df["hour"].between(0, 23), "hour"] = np.nan
df.loc[~df["day_of_week"].between(1, 7), "day_of_week"] = np.nan
df.loc[~df["is_new_beneficiary"].isin([0, 1]), "is_new_beneficiary"] = np.nan
df.loc[~df["is_round_amount"].isin([0, 1]), "is_round_amount"] = np.nan
df.loc[df["amount"] < 0, "amount"] = np.nan


# remove duplicated transactions while ignoring the unique transaction id
duplicate_columns = [
    column for column in df.columns if column != "transaction_id"
]
duplicate_rows = int(df.duplicated(subset=duplicate_columns).sum())
df = df.drop_duplicates(subset=duplicate_columns).reset_index(drop=True)


# encode the binary target as zero and one
target_mapping = {
    "Tidak": 0,
    "Ya (Anomali)": 1,
}
df["is_anomaly"] = df["is_anomaly"].map(target_mapping)


# stop the process when required values are missing or invalid
required_columns = [
    "entity",
    "description",
    "amount",
    "hour",
    "day_of_week",
    "is_new_beneficiary",
    "is_round_amount",
    "is_anomaly",
]
if df[required_columns].isna().any().any():
    missing_counts = df[required_columns].isna().sum()
    raise ValueError(
        "missing or invalid values found:\n"
        f"{missing_counts[missing_counts > 0]}"
    )


# create additional transaction features
df["log_amount"] = np.log1p(df["amount"])
df["is_weekend"] = df["day_of_week"].isin([6, 7]).astype(int)
df["is_night"] = ((df["hour"] >= 22) | (df["hour"] <= 5)).astype(int)
df["hour_sin"] = np.sin(2 * np.pi * df["hour"] / 24)
df["hour_cos"] = np.cos(2 * np.pi * df["hour"] / 24)
df["day_sin"] = np.sin(2 * np.pi * (df["day_of_week"] - 1) / 7)
df["day_cos"] = np.cos(2 * np.pi * (df["day_of_week"] - 1) / 7)


# remove the identifier and arrange the final columns
final_columns = [
    "entity",
    "description",
    "amount",
    "hour",
    "day_of_week",
    "is_new_beneficiary",
    "is_round_amount",
    "log_amount",
    "is_weekend",
    "is_night",
    "hour_sin",
    "hour_cos",
    "day_sin",
    "day_cos",
    "is_anomaly",
]
df = df[final_columns]


# save one clean dataset for random forest training
output_dir.mkdir(parents=True, exist_ok=True)
df.to_csv(output_path, index=False)


# display a concise preprocessing summary
print("random forest preprocessing completed")
print(f"input rows       : {initial_rows}")
print(f"duplicates removed: {duplicate_rows}")
print(f"output rows      : {len(df)}")
print(f"output file      : {output_path}")
