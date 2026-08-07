import re
import pandas as pd

# ======================================================================
# MASTER BILINGUAL DICTIONARIES (Indonesian <-> English Mapping)
# ======================================================================

ENTITIES_MAP = {
    # Indonesian -> English & English -> Indonesian
    'kantor pusat': 'Head Office',
    'head office': 'Head Office',
    'cabang jakarta': 'Jakarta Branch',
    'jakarta branch': 'Jakarta Branch',
    'cabang surabaya': 'Surabaya Branch',
    'surabaya branch': 'Surabaya Branch',
    'cabang bandung': 'Bandung Branch',
    'bandung branch': 'Bandung Branch',
    'cabang bali': 'Bali Branch',
    'bali branch': 'Bali Branch',
    'departemen keuangan': 'Finance Dept',
    'finance dept': 'Finance Dept',
    'departemen ti': 'IT Dept',
    'it dept': 'IT Dept',
    'departemen sdm': 'HR Dept',
    'hr dept': 'HR Dept',
    'departemen operasional': 'Operations Dept',
    'operations dept': 'Operations Dept',
    'hukum & kepatuhan': 'Legal & Compliance',
    'legal & compliance': 'Legal & Compliance'
}

RISK_CATEGORY_MAP = {
    'pengendalian internal': 'Internal Control',
    'internal control': 'Internal Control',
    'kepatuhan regulator': 'Compliance & Legal',
    'compliance & legal': 'Compliance & Legal',
    'risiko keuangan': 'Financial Risk',
    'financial risk': 'Financial Risk',
    'risiko operasional': 'Operational Risk',
    'operational risk': 'Operational Risk',
    'keamanan ti & siber': 'IT & Cybersecurity',
    'it & cybersecurity': 'IT & Cybersecurity',
    'kecurangan & fraud': 'Fraud & Misconduct',
    'fraud & misconduct': 'Fraud & Misconduct',
    'risiko reputasi': 'Reputation Risk',
    'reputation risk': 'Reputation Risk'
}

USER_ROLE_MAP = {
    'auditor': 'Auditor',
    'manajer audit': 'Audit Manager',
    'audit manager': 'Audit Manager',
    'staf keuangan': 'Staff Finance',
    'staff finance': 'Staff Finance',
    'administrator sistem': 'System Admin',
    'system admin': 'System Admin',
    'dukungan it': 'IT Support',
    'it support': 'IT Support',
    'mantan pegawai': 'Ex-Employee',
    'ex-employee': 'Ex-Employee'
}

KPI_NAME_MAP = {
    'biaya akuisisi pelanggan': 'Customer Acquisition Cost',
    'customer acquisition cost': 'Customer Acquisition Cost',
    'nilai seumur hidup pelanggan': 'Customer Lifetime Value',
    'customer lifetime value': 'Customer Lifetime Value',
    'skor kepuasan pelanggan': 'Net Promoter Score',
    'net promoter score': 'Net Promoter Score',
    'rasio konversi penjualan': 'Sales Conversion Rate',
    'sales conversion rate': 'Sales Conversion Rate',
    'tingkat perputaran karyawan': 'Employee Turnover Rate',
    'employee turnover rate': 'Employee Turnover Rate',
    'pendapatan berulang bulanan': 'Monthly Recurring Revenue',
    'monthly recurring revenue': 'Monthly Recurring Revenue',
    'tingkat retensi pelanggan': 'Customer Retention Rate',
    'customer retention rate': 'Customer Retention Rate',
    'nilai rata-rata pesanan': 'Average Order Value',
    'average order value': 'Average Order Value',
    'pengembalian investasi': 'Return on Investment',
    'return on investment': 'Return on Investment',
    'margin laba kotor': 'Gross Profit Margin',
    'gross profit margin': 'Gross Profit Margin'
}

SEVERITY_MAP = {
    'sangat signifikan': 'Very Significant',
    'very significant': 'Very Significant',
    'signifikan': 'Significant',
    'significant': 'Significant',
    'cukup signifikan': 'Quite Significant',
    'quite significant': 'Quite Significant',
    'tidak signifikan': 'Not Significant',
    'not significant': 'Not Significant'
}

MASTER_BILINGUAL_MAP = {}
for d in [ENTITIES_MAP, RISK_CATEGORY_MAP, USER_ROLE_MAP, KPI_NAME_MAP, SEVERITY_MAP]:
    MASTER_BILINGUAL_MAP.update(d)

# Reverse mapping for English -> Indonesian standard if needed
MASTER_INDONESIAN_MAP = {
    'Head Office': 'Kantor Pusat',
    'Jakarta Branch': 'Cabang Jakarta',
    'Surabaya Branch': 'Cabang Surabaya',
    'Bandung Branch': 'Cabang Bandung',
    'Bali Branch': 'Cabang Bali',
    'Finance Dept': 'Departemen Keuangan',
    'IT Dept': 'Departemen IT',
    'HR Dept': 'Departemen SDM',
    'Operations Dept': 'Departemen Operasional',
    'Legal & Compliance': 'Hukum & Kepatuhan',
    'Internal Control': 'Pengendalian Internal',
    'Compliance & Legal': 'Kepatuhan Regulator',
    'Financial Risk': 'Risiko Keuangan',
    'Operational Risk': 'Risiko Operasional',
    'IT & Cybersecurity': 'Keamanan TI & Siber',
    'Fraud & Misconduct': 'Kecurangan & Fraud',
    'Customer Acquisition Cost': 'Biaya Akuisisi Pelanggan',
    'Customer Lifetime Value': 'Nilai Seumur Hidup Pelanggan',
    'Net Promoter Score': 'Skor Kepuasan Pelanggan',
    'Sales Conversion Rate': 'Rasio Konversi Penjualan',
    'Employee Turnover Rate': 'Tingkat Perputaran Karyawan',
    'Monthly Recurring Revenue': 'Pendapatan Berulang Bulanan',
    'Customer Retention Rate': 'Tingkat Retensi Pelanggan',
    'Average Order Value': 'Nilai Rata-rata Pesanan',
    'Return on Investment': 'Pengembalian Investasi',
    'Gross Profit Margin': 'Margin Laba Kotor',
    'Very Significant': 'Sangat Signifikan',
    'Significant': 'Signifikan',
    'Quite Significant': 'Cukup Signifikan',
    'Not Significant': 'Tidak Signifikan'
}


def normalize_term(term: str, target_standard: str = 'en') -> str:
    """
    Translates/Normalizes a single text string/term to canonical form.
    
    Args:
        term (str): Input text (Indonesian or English).
        target_standard (str): 'en' for English canonical, 'id' for Indonesian.
        
    Returns:
        str: Normalized canonical string term.
    """
    if not isinstance(term, str):
        return term
    
    clean_term = term.strip().lower()
    
    if target_standard == 'en':
        return MASTER_BILINGUAL_MAP.get(clean_term, term.strip())
    elif target_standard == 'id':
        canonical_en = MASTER_BILINGUAL_MAP.get(clean_term, term.strip())
        return MASTER_INDONESIAN_MAP.get(canonical_en, canonical_en)
    return term.strip()


def normalize_dataframe(df: pd.DataFrame, target_standard: str = 'en') -> pd.DataFrame:
    """
    Translates and normalizes all string/categorical columns in a DataFrame.
    
    Args:
        df (pd.DataFrame): Input dataframe.
        target_standard (str): 'en' or 'id'.
        
    Returns:
        pd.DataFrame: Clean normalized DataFrame.
    """
    df_clean = df.copy()
    string_cols = df_clean.select_dtypes(include=['object', 'string']).columns
    
    for col in string_cols:
        if col.startswith('TARGET:'):
            continue  # Leave targets intact or handle separately
        df_clean[col] = df_clean[col].apply(lambda val: normalize_term(val, target_standard=target_standard))
        
    return df_clean


if __name__ == '__main__':
    print("======================================================================")
    print(" TESTING BILINGUAL TRANSLATION & MAPPING LAYER ")
    print("======================================================================")
    
    test_terms = [
        ("Biaya Akuisisi Pelanggan", "Customer Acquisition Cost"),
        ("departemen keuangan", "Finance Dept"),
        ("Sangat Signifikan", "Very Significant"),
        ("Head Office", "Head Office"),
        ("Kepatuhan Regulator", "Compliance & Legal")
    ]
    
    print(f"{'Input Term':<30} | {'Normalized (EN)':<25} | {'Normalized (ID)':<25}")
    print("-" * 85)
    for ind, eng in test_terms:
        norm_en = normalize_term(ind, target_standard='en')
        norm_id = normalize_term(eng, target_standard='id')
        print(f"{ind:<30} | {norm_en:<25} | {norm_id:<25}")
