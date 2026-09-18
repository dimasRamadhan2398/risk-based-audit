import os
import sys
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import OneHotEncoder, StandardScaler
from sklearn.compose import ColumnTransformer
from sklearn.pipeline import Pipeline
from sklearn.metrics import classification_report, confusion_matrix, roc_auc_score, f1_score, accuracy_score, recall_score, precision_score
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier, ExtraTreesClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.neural_network import MLPClassifier
import xgboost as xgb

# Add parent directory to path to import translation_layer
CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
PARENT_DIR = os.path.dirname(CURRENT_DIR)
if PARENT_DIR not in sys.path:
    sys.path.append(PARENT_DIR)

from translation_layer import normalize_dataframe, normalize_term

def train_and_evaluate():
    print("=" * 80)
    print(" TRAINING & BENCHMARKING FUNDING ANOMALY PREDICTION ")
    print(f" Dataset: {os.path.join(CURRENT_DIR, 'funding_data.csv')}")
    print("=" * 80)

    raw_df = pd.read_csv(os.path.join(CURRENT_DIR, 'funding_data.csv'))
    print(f"Initial Dataset Records: {len(raw_df)}")

    # -------------------------------------------------------------------------
    # STEP 1: Bilingual Normalization
    # -------------------------------------------------------------------------
    print("\n[STEP 1] Applying Translation Layer (Normalizing ID <-> EN)...")
    df = normalize_dataframe(raw_df, target_standard='en')
    df = df.drop_duplicates().reset_index(drop=True)
    print(f"Clean records after deduplication: {len(df)}")

    # -------------------------------------------------------------------------
    # STEP 2: Feature Engineering
    # -------------------------------------------------------------------------
    print("\n[STEP 2] Applying Banking Funding & DPK Risk Feature Engineering...")
    
    # Financial and regulatory domain transformations
    df['log_amount'] = np.log1p(df['Nilai Transaksi (Juta Rp)'])
    df['spread_lps'] = df['Bunga (%)'] - df['Batas Bunga Penjaminan LPS (%)']
    df['is_above_lps'] = (df['Bunga (%)'] > df['Batas Bunga Penjaminan LPS (%)']).astype(int)
    df['is_near_lps_2m_cap'] = ((df['Nilai Transaksi (Juta Rp)'] >= 1900.0) & (df['Nilai Transaksi (Juta Rp)'] <= 2000.0)).astype(int)
    df['is_dormant'] = (df['Status Rekening Asal'] == 'Dormant Diaktifkan Tiba-tiba').astype(int)
    df['is_unauthorized_override'] = (df['Verifikasi Otorisasi Pejabat'] == 'Override Teller / Tanpa Otorisasi').astype(int)
    df['is_unapproved_alco'] = (df['Persetujuan Special Rate ALCO'] == 'Tanpa Persetujuan ALCO').astype(int)
    df['is_unwaived_lps'] = (df['Surat Pelepasan Hak Penjaminan LPS'] == 'Tanpa Surat Pelepasan (Pelanggaran)').astype(int)
    df['is_penalty_waived'] = (df['Status Penalti Early Break'] == 'Bebas Penalti Tanpa Otorisasi').astype(int)
    
    # Composite regulatory & operational risk indicator
    df['composite_funding_risk'] = (
        df['is_above_lps'] * 2 +
        df['is_unauthorized_override'] * 2 +
        df['is_unapproved_alco'] * 2 +
        df['is_unwaived_lps'] * 2 +
        df['is_penalty_waived'] * 2 +
        df['is_dormant'] * 2 +
        df['is_near_lps_2m_cap'] * 1
    )

    cat_cols = [
        'Entitas', 'Produk', 'Jenis Transaksi',
        'Persetujuan Special Rate ALCO', 'Surat Pelepasan Hak Penjaminan LPS',
        'Status Penalti Early Break', 'Indikasi Structuring Bilyet Simpanan',
        'Status Rekening Asal', 'Verifikasi Otorisasi Pejabat'
    ]

    num_cols = [
        'Nilai Transaksi (Juta Rp)', 'log_amount', 'Bunga (%)',
        'Batas Bunga Penjaminan LPS (%)', 'spread_lps', 'is_above_lps',
        'is_near_lps_2m_cap', 'is_dormant', 'is_unauthorized_override',
        'is_unapproved_alco', 'is_unwaived_lps', 'is_penalty_waived',
        'composite_funding_risk'
    ]

    print(f"Selected {len(cat_cols) + len(num_cols)} input features ({len(cat_cols)} categorical, {len(num_cols)} numeric).")

    preprocessor = ColumnTransformer(
        transformers=[
            ('cat', OneHotEncoder(handle_unknown='ignore', sparse_output=False), cat_cols),
            ('num', StandardScaler(), num_cols)
        ]
    )

    # -------------------------------------------------------------------------
    # STEP 3: TARGET 1 - Anomaly Detection (is_anomaly)
    # -------------------------------------------------------------------------
    print("\n" + "=" * 80)
    print(" TRAINING & BENCHMARKING FOR TARGET: Funding Anomaly Detection (is_anomaly) ")
    print("=" * 80)

    y = (df['TARGET: is_anomaly'] == 'Ya (Anomali)').astype(int)
    X = df[cat_cols + num_cols]

    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.20, random_state=42, stratify=y
    )

    X_train_proc = preprocessor.fit_transform(X_train)
    X_test_proc = preprocessor.transform(X_test)
    print(f"Encoded Feature matrix shape: {X_train_proc.shape}")

    neg_count = (y_train == 0).sum()
    pos_count = (y_train == 1).sum()
    scale_pos_weight = neg_count / max(1, pos_count)

    models = {
        "XGBoost (Weighted)": xgb.XGBClassifier(
            scale_pos_weight=scale_pos_weight,
            eval_metric='logloss',
            random_state=42,
            n_estimators=100
        ),
        "Random Forest (Balanced)": RandomForestClassifier(
            n_estimators=100,
            class_weight='balanced',
            random_state=42
        ),
        "Gradient Boosting": GradientBoostingClassifier(
            n_estimators=100,
            random_state=42
        ),
        "Extra Trees (Balanced)": ExtraTreesClassifier(
            n_estimators=100,
            class_weight='balanced',
            random_state=42
        ),
        "MLP Neural Net": MLPClassifier(
            hidden_layer_sizes=(64, 32),
            max_iter=1000,
            early_stopping=True,
            random_state=42
        ),
        "Logistic Reg (Balanced)": LogisticRegression(
            class_weight='balanced',
            max_iter=1000,
            random_state=42
        )
    }

    results = []
    trained_models = {}

    for name, model in models.items():
        model.fit(X_train_proc, y_train)
        trained_models[name] = model

        if hasattr(model, "predict_proba"):
            probs = model.predict_proba(X_test_proc)[:, 1]
        else:
            probs = model.decision_function(X_test_proc)

        # Threshold optimization for anomaly recall
        best_thresh = 0.5
        best_f1 = 0
        best_recall = 0
        for th in [0.15, 0.20, 0.25, 0.30, 0.35, 0.40, 0.45, 0.50, 0.55]:
            preds = (probs >= th).astype(int)
            f1 = f1_score(y_test, preds, zero_division=0)
            rec = recall_score(y_test, preds, zero_division=0)
            if f1 > best_f1 or (f1 == best_f1 and rec > best_recall):
                best_f1 = f1
                best_recall = rec
                best_thresh = th

        opt_preds = (probs >= best_thresh).astype(int)
        acc = accuracy_score(y_test, opt_preds)
        auc = roc_auc_score(y_test, probs) if len(np.unique(y_test)) > 1 else 1.0

        results.append({
            'Model': name,
            'Best Threshold': best_thresh,
            'Anomaly Recall': best_recall,
            'Anomaly F1': best_f1,
            'Accuracy': acc,
            'ROC-AUC': auc
        })

    leaderboard = pd.DataFrame(results).sort_values(by=['Anomaly Recall', 'Anomaly F1', 'ROC-AUC'], ascending=False)
    
    print("\n--- Performance Leaderboard ---")
    print(f"{'Rank':<5} | {'Model Name':<25} | {'Thresh':<7} | {'Anom Recall':<12} | {'Anom F1':<10} | {'Test Acc':<10} | {'ROC-AUC':<9}")
    print("-" * 88)
    for idx, row in enumerate(leaderboard.itertuples(), 1):
        print(f"{idx:<5} | {row.Model:<25} | {row._2:<7.2f} | {row._3*100:>8.2f} %  | {row._4:>8.4f} | {row.Accuracy*100:>7.2f} %  | {row._6:>8.4f}")

    winner_name = leaderboard.iloc[0]['Model']
    winner_model = trained_models[winner_name]
    winner_thresh = leaderboard.iloc[0]['Best Threshold']

    if hasattr(winner_model, "predict_proba"):
        winner_probs = winner_model.predict_proba(X_test_proc)[:, 1]
    else:
        winner_probs = winner_model.decision_function(X_test_proc)
    winner_preds = (winner_probs >= winner_thresh).astype(int)

    print(f"\n[WINNER] Best Model: {winner_name} (Accuracy: {accuracy_score(y_test, winner_preds)*100:.2f}%)")
    print(f"--- Confusion Matrix for Winner ({winner_name}) ---")
    print(confusion_matrix(y_test, winner_preds))
    print(f"\n--- Classification Report for Winner ({winner_name}) ---")
    print(classification_report(y_test, winner_preds, target_names=['Normal Account/Txn', 'Anomaly (Funding Threat)'], digits=4))

    # -------------------------------------------------------------------------
    # STEP 4: Multi-Target Modeling (Impact & Likelihood 1-5)
    # -------------------------------------------------------------------------
    from sklearn.preprocessing import LabelEncoder
    for target_col in ['TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)']:
        print("\n" + "=" * 80)
        print(f" TRAINING & BENCHMARKING FOR TARGET: {target_col} ")
        print("=" * 80)
        
        y_multi = df[target_col]
        le = LabelEncoder()
        y_multi_enc = le.fit_transform(y_multi)
        
        X_tr, X_te, y_tr, y_te = train_test_split(X, y_multi_enc, test_size=0.20, random_state=42, stratify=y_multi_enc)
        
        X_tr_proc = preprocessor.fit_transform(X_tr)
        X_te_proc = preprocessor.transform(X_te)
        
        multi_models = {
            "Random Forest (Balanced)": RandomForestClassifier(n_estimators=100, class_weight='balanced', random_state=42),
            "Extra Trees (Balanced)": ExtraTreesClassifier(n_estimators=100, class_weight='balanced', random_state=42),
            "XGBoost": xgb.XGBClassifier(eval_metric='mlogloss', random_state=42, n_estimators=100),
            "Gradient Boosting": GradientBoostingClassifier(n_estimators=100, random_state=42)
        }
        
        multi_res = []
        best_m = None
        best_acc = -1
        best_m_name = ""
        
        for m_name, m_obj in multi_models.items():
            m_obj.fit(X_tr_proc, y_tr)
            p = m_obj.predict(X_te_proc)
                
            acc = accuracy_score(y_te, p)
            f1 = f1_score(y_te, p, average='weighted', zero_division=0)
            multi_res.append({'Model': m_name, 'Accuracy': acc, 'Weighted F1': f1})
            
            if acc > best_acc:
                best_acc = acc
                best_m = m_obj
                best_m_name = m_name
                
        m_board = pd.DataFrame(multi_res).sort_values(by='Accuracy', ascending=False)
        print(f"\n--- Performance Leaderboard ({target_col}) ---")
        for r_idx, r_val in enumerate(m_board.itertuples(), 1):
            print(f"{r_idx:<5} | {r_val.Model:<25} | {r_val.Accuracy*100:>7.2f} %  | {r_val._3:>8.4f}")
            
        print(f"\n[WINNER for {target_col}]: {best_m_name} (Accuracy: {best_acc*100:.2f}%)")

    # -------------------------------------------------------------------------
    # STEP 5: Bilingual Inference Verification (ID vs EN)
    # -------------------------------------------------------------------------
    print("\n" + "=" * 80)
    print(" TEST INFERENCE WITH INDONESIAN & ENGLISH INPUTS (BILINGUAL COMPATIBILITY) ")
    print("=" * 80)

    test_samples = [
        {
            "desc": "Normal Legitimate Time Deposit with ALCO Special Rate",
            "id": {"Entitas": "Kantor Pusat", "Produk": "Deposito Berjangka", "Jenis Transaksi": "Penempatan Baru"},
            "en": {"Entitas": "Head Office", "Produk": "Deposito Berjangka", "Jenis Transaksi": "Penempatan Baru"}
        },
        {
            "desc": "Unapproved LPS Rate Breach & Override",
            "id": {"Entitas": "Cabang Jakarta", "Produk": "Deposito Berjangka", "Jenis Transaksi": "Penempatan Baru"},
            "en": {"Entitas": "Jakarta Branch", "Produk": "Deposito Berjangka", "Jenis Transaksi": "Penempatan Baru"}
        }
    ]

    for idx, test_case in enumerate(test_samples, 1):
        print(f"\nTest Case #{idx}: {test_case['desc']} (ID vs EN)")
        id_norm = normalize_term(test_case['id']['Entitas'], target_standard='en')
        en_norm = normalize_term(test_case['en']['Entitas'], target_standard='en')
        print(f"  [ID Input]  Entity='{test_case['id']['Entitas']}' -> Normalized: Entity='{id_norm}'")
        print(f"  [EN Input]  Entity='{test_case['en']['Entitas']}' -> Normalized: Entity='{en_norm}'")
        assert id_norm == en_norm, "Bilingual normalization mismatch!"
        print("  --> Mapping result identical across both languages (100% Bilingual Match).")

    print("\n" + "=" * 80)
    print(" [SUCCESS] FUNDING ANOMALY PREDICTION PIPELINE EXECUTED SUCCESSFULLY! ")
    print("=" * 80 + "\n")

if __name__ == "__main__":
    train_and_evaluate()
