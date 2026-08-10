import os
import numpy as np
import pandas as pd

def generate_dataset(num_samples=1000, seed=42):
    np.random.seed(seed)
    
    entities = [
        "HR Dept", "Finance Dept", "IT Dept", "Operational Dept", 
        "Supply Chain Dept", "Marketing Dept", "Legal & Compliance Dept", 
        "Risk Management Dept", "Jakarta Branch", "Surabaya Branch", 
        "Bandung Branch", "Medan Branch", "Semarang Branch", 
        "Makassar Branch", "Bali Branch", "Palembang Branch", 
        "Balikpapan Branch", "Manado Branch"
    ]
    
    risk_categories = [
        "Financial", "Technology", "Operational", "Compliance", 
        "Strategic", "Reputational", "Legal", "Fraud & Security"
    ]
    
    # 1. Random sample entities and risk categories
    entitas_col = np.random.choice(entities, size=num_samples)
    kategori_col = np.random.choice(risk_categories, size=num_samples)
    
    # 2. Inherent Likelihood & Inherent Impact (1 to 5)
    inherent_lik = np.random.choice([1, 2, 3, 4, 5], size=num_samples, p=[0.15, 0.25, 0.30, 0.20, 0.10])
    inherent_imp = np.random.choice([1, 2, 3, 4, 5], size=num_samples, p=[0.10, 0.20, 0.35, 0.23, 0.12])
    
    # 3. Operational indicators correlated with inherent risk
    base_risk = (inherent_lik + inherent_imp) / 2.0
    
    audit_findings = np.round(np.clip(
        np.random.poisson(lam=base_risk * 1.2, size=num_samples) + np.random.normal(0, 0.5, size=num_samples),
        0, 15
    )).astype(int)
    
    kpi_below_target = np.round(np.clip(
        np.random.poisson(lam=base_risk * 0.8, size=num_samples) + np.random.normal(0, 0.4, size=num_samples),
        0, 10
    )).astype(int)
    
    kpi_volatility = np.round(np.clip(
        base_risk * np.random.uniform(1.0, 3.5, size=num_samples) + np.random.exponential(scale=1.5, size=num_samples),
        0.1, 25.0
    ), 1)
    
    # Previous risk score (inherent risk from last period + minor trend)
    prev_risk_score = np.round(np.clip(
        (inherent_lik * inherent_imp) + np.random.normal(0, 2.5, size=num_samples),
        1.0, 25.0
    ), 1)
    
    # Assessment month (1 to 12)
    assessment_month = np.random.randint(1, 13, size=num_samples)
    
    # 4. Generate Target Likelihood & Impact using domain risk scoring formula
    raw_target_lik = (
        0.48 * inherent_lik +
        0.15 * np.log1p(audit_findings) * 2.0 +
        0.12 * np.log1p(kpi_below_target) * 2.0 +
        0.10 * (kpi_volatility / 5.0) +
        0.15 * (prev_risk_score / 5.0) +
        np.random.normal(0, 0.08, size=num_samples)
    )
    
    raw_target_imp = (
        0.52 * inherent_imp +
        0.14 * np.log1p(audit_findings) * 2.0 +
        0.14 * np.log1p(kpi_below_target) * 2.0 +
        0.08 * (kpi_volatility / 5.0) +
        0.12 * (prev_risk_score / 5.0) +
        np.random.normal(0, 0.08, size=num_samples)
    )
    
    # Map raw targets to discrete 1-5 scale using quantiles to ensure adequate representation across all risk levels (1-5)
    def map_to_1_to_5_balanced(arr):
        # Calculate percentiles (20%, 40%, 60%, 80%) for balanced class distribution
        q20, q40, q60, q80 = np.percentile(arr, [18, 40, 62, 84])
        labels = np.zeros_like(arr, dtype=int)
        labels[arr < q20] = 1
        labels[(arr >= q20) & (arr < q40)] = 2
        labels[(arr >= q40) & (arr < q60)] = 3
        labels[(arr >= q60) & (arr < q80)] = 4
        labels[arr >= q80] = 5
        return labels

    target_lik = map_to_1_to_5_balanced(raw_target_lik)
    target_imp = map_to_1_to_5_balanced(raw_target_imp)
    
    df = pd.DataFrame({
        "Entitas": entitas_col,
        "Kategori Risiko": kategori_col,
        "Inherent Likelihood": inherent_lik,
        "Inherent Impact": inherent_imp,
        "Jml Temuan Audit": audit_findings,
        "Jml KPI di Bawah Target": kpi_below_target,
        "Volatilitas KPI": kpi_volatility,
        "Skor Risiko Periode Lalu": prev_risk_score,
        "Bulan Penilaian": assessment_month,
        "TARGET: Likelihood": target_lik,
        "TARGET: Impact": target_imp
    })
    
    df = df.drop_duplicates().reset_index(drop=True)
    if len(df) < num_samples:
        return generate_dataset(num_samples=num_samples, seed=seed + 1)
        
    return df.iloc[:num_samples]

if __name__ == "__main__":
    current_dir = os.path.dirname(os.path.abspath(__file__))
    output_path = os.path.join(current_dir, "department_dataset.csv")
    
    df_1000 = generate_dataset(1000, seed=42)
    df_1000.to_csv(output_path, index=False)
    print(f"Successfully generated {len(df_1000)} high quality department risk data rows at: {output_path}")
