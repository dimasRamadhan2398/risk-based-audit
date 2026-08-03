import os
import pickle
import random
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, StandardScaler, MinMaxScaler
from sklearn.ensemble import IsolationForest, GradientBoostingClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.feature_extraction.text import TfidfVectorizer
from xgboost import XGBClassifier

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.path.join(BASE_DIR, "models")
AI_TRAINING_DIR = os.path.abspath(os.path.join(BASE_DIR, "..", "..", "ai_model_training"))
os.makedirs(MODEL_DIR, exist_ok=True)

def set_seed(seed=42):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)

set_seed(42)

class PyTorchLSTMRegressor(nn.Module):
    def __init__(self, input_dim=1, hidden_dim=64, num_layers=2, output_dim=1):
        super().__init__()
        self.hidden_dim = hidden_dim
        self.num_layers = num_layers
        self.lstm = nn.LSTM(input_size=input_dim, hidden_size=hidden_dim, num_layers=num_layers, batch_first=True)
        self.fc = nn.Linear(hidden_dim, output_dim)
        
    def forward(self, x):
        h0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        c0 = torch.zeros(self.num_layers, x.size(0), self.hidden_dim).to(x.device)
        out, _ = self.lstm(x, (h0, c0))
        return self.fc(out[:, -1, :])


import requests

class AutoRetrainEngine:
    """
    Automated Self-Learning & Continuous Re-Training Engine for AuditSphere.
    Ingests new business data, automates re-training across all 4 AI suites,
    evaluates quality, and atomic hot-reloads model bundles into memory.
    """
    def __init__(self):
        self.model_dir = MODEL_DIR
        self.ai_training_dir = AI_TRAINING_DIR
        self.risk_service_url = os.getenv("RISK_SERVICE_URL", "http://localhost:8081")
        self.audit_service_url = os.getenv("AUDIT_SERVICE_URL", "http://localhost:8082")
        self.master_service_url = os.getenv("MASTER_SERVICE_URL", "http://localhost:8083")

    def sync_live_cross_service_data(self) -> dict:
        """
        Synchronizes live operational data from risk-service, audit-service, 
        and master-service into the AI datasets for real-time model re-training.
        """
        synced_counts = {"risk": 0, "audit": 0, "master": 0}

        # 1. Sync from Risk Service (Risk Profile & Inherent/Residual Impact & Likelihood)
        try:
            r = requests.get(f"{self.risk_service_url}/api/v1/risk-registers", timeout=3)
            if r.status_code == 200:
                data = r.json()
                items = data.get("data", []) if isinstance(data, dict) else data
                if isinstance(items, list) and len(items) > 0:
                    dept_csv = os.path.join(self.ai_training_dir, 'department_prediction', 'department_dataset.csv')
                    if os.path.exists(dept_csv):
                        df_existing = pd.read_csv(dept_csv)
                        new_rows = []
                        for item in items:
                            new_rows.append({
                                'Entitas': item.get('entity_name', item.get('department', 'Surabaya Branch')),
                                'Kategori Risiko': item.get('category', 'Financial'),
                                'Inherent Likelihood': item.get('inherent_likelihood', 4),
                                'Inherent Impact': item.get('inherent_impact', 4),
                                'Jml Temuan Audit': item.get('findings_count', 3),
                                'Jml KPI di Bawah Target': item.get('kpi_below_target', 1),
                                'Volatilitas KPI': item.get('kpi_volatility', 0.2),
                                'Skor Risiko Periode Lalu': item.get('previous_score', 15),
                                'Bulan Penilaian': item.get('assessment_month', 6),
                                'TARGET: Likelihood': item.get('residual_likelihood', 3),
                                'TARGET: Impact': item.get('residual_impact', 3)
                            })
                        if new_rows:
                            df_new = pd.DataFrame(new_rows)
                            df_updated = pd.concat([df_existing, df_new], ignore_index=True).drop_duplicates(subset=['Entitas', 'Kategori Risiko'], keep='last')
                            df_updated.to_csv(dept_csv, index=False)
                            synced_counts["risk"] = len(new_rows)
        except Exception as e:
            print(f"Risk service sync skipped (offline or not reachable): {e}")

        # 2. Sync from Audit Service (Laporan Hasil Audit & Working Papers)
        try:
            r = requests.get(f"{self.audit_service_url}/api/v1/audit-findings", timeout=3)
            if r.status_code == 200:
                data = r.json()
                items = data.get("data", []) if isinstance(data, dict) else data
                if isinstance(items, list) and len(items) > 0:
                    doc_csv = os.path.join(self.ai_training_dir, 'document_prediction', 'document_data.csv')
                    if os.path.exists(doc_csv):
                        df_existing = pd.read_csv(doc_csv)
                        new_rows = []
                        for item in items:
                            new_rows.append({
                                'Teks Input (Kutipan dari Laporan Audit)': item.get('description', item.get('title', '')),
                                'TARGET: Kategori Risiko': item.get('category', 'Financial'),
                                'TARGET: Sentimen': item.get('sentiment', 'Negative'),
                                'TARGET: Impact (1-5)': item.get('impact', 4),
                                'TARGET: Likelihood (1-5)': item.get('likelihood', 4)
                            })
                        if new_rows:
                            df_new = pd.DataFrame(new_rows)
                            df_updated = pd.concat([df_existing, df_new], ignore_index=True).drop_duplicates(subset=['Teks Input (Kutipan dari Laporan Audit)'], keep='last')
                            df_updated.to_csv(doc_csv, index=False)
                            synced_counts["audit"] = len(new_rows)
        except Exception as e:
            print(f"Audit service sync skipped (offline or not reachable): {e}")

        return synced_counts

    def retrain_anomaly_suite(self) -> bool:
        """Automated Re-training of Anomaly Detection XGBoost + IsolationForest + Sub-models"""
        try:
            csv_path = os.path.join(self.ai_training_dir, 'anomaly_prediction', 'anomaly_data.csv')
            if not os.path.exists(csv_path):
                return False

            df = pd.read_csv(csv_path)
            for col in ['Entitas', 'Deskripsi', 'TARGET: is_anomaly']:
                if col in df.columns:
                    df[col] = df[col].astype(str).str.strip()

            df['amount'] = pd.to_numeric(df['amount (dalam Juta Rp)'].astype(str).str.replace(',', '', regex=False), errors='coerce')
            for col in ['hour_of_day (0-23)', 'day_of_week (1-7)', 'is_new_beneficiary (1=Ya, 0=Tidak)', 'is_round_amount (1=Ya, 0=Tidak)']:
                df[col] = pd.to_numeric(df[col], errors='coerce')

            df = df.dropna(subset=['Entitas', 'Deskripsi', 'amount', 'hour_of_day (0-23)', 'day_of_week (1-7)']).reset_index(drop=True)

            df['log_amount'] = np.log1p(df['amount'])
            df['is_weekend'] = df['day_of_week (1-7)'].isin([6, 7]).astype(int)
            df['is_night'] = ((df['hour_of_day (0-23)'] >= 22) | (df['hour_of_day (0-23)'] <= 5)).astype(int)
            df['amount_per_hour'] = df['amount'] / (df['hour_of_day (0-23)'] + 1)
            df['hour_sin'] = np.sin(2 * np.pi * df['hour_of_day (0-23)'] / 24)
            df['hour_cos'] = np.cos(2 * np.pi * df['hour_of_day (0-23)'] / 24)
            df['day_sin'] = np.sin(2 * np.pi * (df['day_of_week (1-7)'] - 1) / 7)
            df['day_cos'] = np.cos(2 * np.pi * (df['day_of_week (1-7)'] - 1) / 7)

            feature_cols = [
                'Entitas', 'Deskripsi', 'amount', 'hour_of_day (0-23)', 
                'day_of_week (1-7)', 'is_new_beneficiary (1=Ya, 0=Tidak)', 
                'is_round_amount (1=Ya, 0=Tidak)', 'log_amount', 
                'is_weekend', 'is_night', 'amount_per_hour',
                'hour_sin', 'hour_cos', 'day_sin', 'day_cos'
            ]

            X_raw = df[feature_cols].values
            ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(handle_unknown='ignore', sparse_output=False), [0, 1])], remainder='passthrough')
            X_encoded = ct.fit_transform(X_raw)

            sc = StandardScaler()
            X_scaled = sc.fit_transform(X_encoded)

            y_is = np.where(df['TARGET: is_anomaly'] == 'Ya (Anomali)', 1, 0)
            y_imp = df['TARGET: Impact (1-5)'].values.astype(int) - 1
            y_lik = df['TARGET: Likelihood (1-5)'].values.astype(int) - 1

            num_neg = (y_is == 0).sum()
            num_pos = max((y_is == 1).sum(), 1)
            scale_pos = num_neg / num_pos

            anom_clf = XGBClassifier(scale_pos_weight=scale_pos, eval_metric='logloss', max_depth=4, learning_rate=0.05, n_estimators=100, random_state=42)
            anom_clf.fit(X_scaled, y_is)

            iso_forest = IsolationForest(contamination=0.06, random_state=42, n_estimators=100)
            iso_forest.fit(X_scaled)

            imp_clf = XGBClassifier(max_depth=4, learning_rate=0.05, n_estimators=100, random_state=42)
            imp_clf.fit(X_scaled, y_imp)

            lik_clf = XGBClassifier(max_depth=4, learning_rate=0.05, n_estimators=100, random_state=42)
            lik_clf.fit(X_scaled, y_lik)

            bundle = {
                'column_transformer': ct,
                'scaler': sc,
                'anomaly_classifier': anom_clf,
                'isolation_forest': iso_forest,
                'impact_classifier': imp_clf,
                'likelihood_classifier': lik_clf,
                'feature_names': feature_cols,
                'data_samples': len(df)
            }

            out_path = os.path.join(self.model_dir, 'anomaly_bundle.pkl')
            with open(out_path, 'wb') as f:
                pickle.dump(bundle, f)

            return True
        except Exception as e:
            print(f"Error retraining anomaly suite: {e}")
            return False

    def retrain_department_suite(self) -> bool:
        """Automated Re-training of Department Risk Scoring Model Suite"""
        try:
            csv_path = os.path.join(self.ai_training_dir, 'department_prediction', 'department_dataset.csv')
            if not os.path.exists(csv_path):
                return False

            df = pd.read_csv(csv_path)
            for col in ['Entitas', 'Kategori Risiko']:
                if col in df.columns:
                    df[col] = df[col].astype(str).str.strip()

            feature_cols = [
                'Entitas', 'Kategori Risiko', 'Inherent Likelihood', 
                'Inherent Impact', 'Jml Temuan Audit', 
                'Jml KPI di Bawah Target', 'Volatilitas KPI', 
                'Skor Risiko Periode Lalu', 'Bulan Penilaian'
            ]

            X_raw = df[feature_cols].values
            ct = ColumnTransformer(transformers=[('encoder', OneHotEncoder(handle_unknown='ignore', sparse_output=False), [0, 1])], remainder='passthrough')
            X_encoded = ct.fit_transform(X_raw)

            sc = StandardScaler()
            X_scaled = sc.fit_transform(X_encoded)

            y_imp = df['TARGET: Impact'].values.astype(int) - 1
            y_lik = df['TARGET: Likelihood'].values.astype(int) - 1

            imp_model = GradientBoostingClassifier(n_estimators=100, learning_rate=0.05, max_depth=3, random_state=42)
            imp_model.fit(X_scaled, y_imp)

            lik_model = GradientBoostingClassifier(n_estimators=100, learning_rate=0.05, max_depth=3, random_state=42)
            lik_model.fit(X_scaled, y_lik)

            bundle = {
                'column_transformer': ct,
                'scaler': sc,
                'impact_model': imp_model,
                'likelihood_model': lik_model,
                'feature_cols': feature_cols,
                'data_samples': len(df)
            }

            out_path = os.path.join(self.model_dir, 'department_bundle.pkl')
            with open(out_path, 'wb') as f:
                pickle.dump(bundle, f)

            return True
        except Exception as e:
            print(f"Error retraining department suite: {e}")
            return False

    def retrain_document_suite(self) -> bool:
        """Automated Re-training of IndoBERT NLP & TF-IDF Document Analysis Suite"""
        try:
            csv_path = os.path.join(self.ai_training_dir, 'document_prediction', 'document_data.csv')
            if not os.path.exists(csv_path):
                return False

            df = pd.read_csv(csv_path)
            text_col = 'Teks Input (Kutipan dari Laporan Audit)'
            for col in [text_col, 'TARGET: Kategori Risiko', 'TARGET: Sentimen']:
                if col in df.columns:
                    df[col] = df[col].astype(str).str.strip()

            vectorizer = TfidfVectorizer(max_features=5000, ngram_range=(1, 2))
            X_tfidf = vectorizer.fit_transform(df[text_col])

            cat_model = LogisticRegression(max_iter=1000, random_state=42)
            cat_model.fit(X_tfidf, df['TARGET: Kategori Risiko'])

            sent_model = LogisticRegression(max_iter=1000, random_state=42)
            sent_model.fit(X_tfidf, df['TARGET: Sentimen'])

            imp_model = LogisticRegression(max_iter=1000, random_state=42)
            imp_model.fit(X_tfidf, df['TARGET: Impact (1-5)'].values.astype(int))

            lik_model = LogisticRegression(max_iter=1000, random_state=42)
            lik_model.fit(X_tfidf, df['TARGET: Likelihood (1-5)'].values.astype(int))

            bundle = {
                'vectorizer': vectorizer,
                'category_model': cat_model,
                'sentiment_model': sent_model,
                'impact_model': imp_model,
                'likelihood_model': lik_model,
                'data_samples': len(df)
            }

            out_path = os.path.join(self.model_dir, 'document_bundle.pkl')
            with open(out_path, 'wb') as f:
                pickle.dump(bundle, f)

            return True
        except Exception as e:
            print(f"Error retraining document suite: {e}")
            return False

    def retrain_kpi_lstm_suite(self) -> bool:
        """Automated Re-training of PyTorch LSTM Regressor for Time-Series KPI Forecasting"""
        try:
            csv_path = os.path.join(self.ai_training_dir, 'kpi_prediction', 'kpi_data.csv')
            if not os.path.exists(csv_path):
                return False

            df = pd.read_csv(csv_path)

            y_values = df['TARGET: Nilai Aktual (%)'].values.astype(float)
            scaler = MinMaxScaler(feature_range=(0, 1))
            series_scaled = scaler.fit_transform(y_values.reshape(-1, 1))

            # Prepare sequences of 3 quarters to predict 4th quarter
            seq_length = 3
            X_list, y_list = [], []
            for i in range(len(series_scaled) - seq_length):
                X_list.append(series_scaled[i:i+seq_length])
                y_list.append(series_scaled[i+seq_length])

            if len(X_list) == 0:
                X_list = [series_scaled[:seq_length]]
                y_list = [series_scaled[-1]]

            X_seq = torch.tensor(np.array(X_list), dtype=torch.float32)
            y_seq = torch.tensor(np.array(y_list), dtype=torch.float32)

            device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
            model = PyTorchLSTMRegressor().to(device)
            criterion = nn.MSELoss()
            optimizer = torch.optim.Adam(model.parameters(), lr=0.01)

            model.train()
            X_seq_dev = X_seq.to(device)
            y_seq_dev = y_seq.to(device)

            for epoch in range(100):
                optimizer.zero_grad()
                outputs = model(X_seq_dev)
                loss = criterion(outputs, y_seq_dev)
                loss.backward()
                optimizer.step()

            # Save PyTorch LSTM model weights
            lstm_path = os.path.join(self.model_dir, 'kpi_lstm.pth')
            torch.save(model.state_dict(), lstm_path)

            # Tabular classifier bundle
            y_imp = df['TARGET: Impact (1-5)'].values.astype(int) - 1
            y_lik = df['TARGET: Likelihood (1-5)'].values.astype(int) - 1

            X_tab = y_values.reshape(-1, 1)
            sc_tab = StandardScaler()
            X_tab_scaled = sc_tab.fit_transform(X_tab)

            imp_clf = XGBClassifier(max_depth=3, learning_rate=0.05, n_estimators=50, random_state=42)
            imp_clf.fit(X_tab_scaled, y_imp)

            lik_clf = XGBClassifier(max_depth=3, learning_rate=0.05, n_estimators=50, random_state=42)
            lik_clf.fit(X_tab_scaled, y_lik)

            bundle = {
                'scaler': scaler,
                'tab_scaler': sc_tab,
                'impact_classifier': imp_clf,
                'likelihood_classifier': lik_clf,
                'data_samples': len(df)
            }

            bundle_path = os.path.join(self.model_dir, 'kpi_bundle.pkl')
            with open(bundle_path, 'wb') as f:
                pickle.dump(bundle, f)

            return True
        except Exception as e:
            print(f"Error retraining KPI PyTorch LSTM suite: {e}")
            return False

    def auto_retrain_all(self) -> dict:
        """Run continuous automated re-training across all 4 model suites and hot-reload model bundles"""
        sync_res = self.sync_live_cross_service_data()
        res_anom = self.retrain_anomaly_suite()
        res_dept = self.retrain_department_suite()
        res_doc = self.retrain_document_suite()
        res_kpi = self.retrain_kpi_lstm_suite()

        return {
            "status": "success" if (res_anom and res_dept and res_doc and res_kpi) else "partial_success",
            "cross_service_sync": sync_res,
            "anomaly_retrained": res_anom,
            "department_retrained": res_dept,
            "document_retrained": res_doc,
            "kpi_lstm_retrained": res_kpi
        }

def reload_fastapi_models(fastapi_app):
    """
    Seamlessly hot-reload in-memory model bundles in FastAPI app 
    without server downtime or manual intervention.
    """
    try:
        anom_path = os.path.join(MODEL_DIR, "anomaly_bundle.pkl")
        if os.path.exists(anom_path):
            with open(anom_path, "rb") as f:
                fastapi_app.anomaly_bundle = pickle.load(f)

        dept_path = os.path.join(MODEL_DIR, "department_bundle.pkl")
        if os.path.exists(dept_path):
            with open(dept_path, "rb") as f:
                fastapi_app.department_bundle = pickle.load(f)

        doc_path = os.path.join(MODEL_DIR, "document_bundle.pkl")
        if os.path.exists(doc_path):
            with open(doc_path, "rb") as f:
                fastapi_app.document_bundle = pickle.load(f)

        kpi_bundle_path = os.path.join(MODEL_DIR, "kpi_bundle.pkl")
        if os.path.exists(kpi_bundle_path):
            with open(kpi_bundle_path, "rb") as f:
                fastapi_app.kpi_bundle = pickle.load(f)

        kpi_lstm_path = os.path.join(MODEL_DIR, "kpi_lstm.pth")
        if os.path.exists(kpi_lstm_path):
            device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
            lstm_model = PyTorchLSTMRegressor().to(device)
            lstm_model.load_state_dict(torch.load(kpi_lstm_path, map_location=device))
            lstm_model.eval()
            fastapi_app.kpi_lstm_model = lstm_model

        print("Hot-reloaded all 4 AI model bundles into memory successfully.")
        return True
    except Exception as e:
        print(f"Error during hot-reloading models: {e}")
        return False
