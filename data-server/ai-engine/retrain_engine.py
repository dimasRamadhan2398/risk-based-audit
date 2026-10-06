"""
AuditSphere AI Engine — Auto-Retrain Engine (Data Hub Version)
Reads training data from Gold zone with hardware-safe stratified sampling.
Groups KPI forecasting by metric name to preserve sequence coherence.
"""
import os
import pickle
import random
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from datetime import datetime
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, StandardScaler, MinMaxScaler
from sklearn.ensemble import IsolationForest, GradientBoostingClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.multioutput import MultiOutputClassifier
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score, r2_score


def set_seed(seed=42):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)


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


class AutoRetrainEngine:
    """
    Automated retraining engine that reads from Gold zone Data Lake.
    Trains 4 model suites:
    1. Department Risk Scoring (GradientBoosting)
    2. Anomaly Detection (Isolation Forest with stratified sampling)
    3. Document NLP (TF-IDF + LogReg)
    4. KPI Forecasting (PyTorch LSTM per KPI series)
    """

    def __init__(self, db_engine, model_dir="/app/models", gold_schema="gold"):
        self.engine = db_engine
        self.model_dir = model_dir
        self.gold_schema = gold_schema
        self.device = torch.device("cpu")
        os.makedirs(model_dir, exist_ok=True)
        set_seed(42)

    def retrain_all(self) -> dict:
        """Retrain all 4 model suites from Gold zone data."""
        results = {}
        results["anomaly"] = self._retrain_anomaly()
        results["department"] = self._retrain_department()
        results["document"] = self._retrain_document()
        results["kpi"] = self._retrain_kpi()

        # Write retrain timestamp
        timestamp = datetime.now().isoformat()
        with open(os.path.join(self.model_dir, "last_retrain.txt"), "w") as f:
            f.write(timestamp)

        print(f"[Retrain] All models retrained at {timestamp}")
        return results

    def _retrain_anomaly(self) -> dict:
        """Retrain Isolation Forest from gold.anomaly_training_data with bounded memory sampling."""
        try:
            # Stratified/random sampling at SQL level to cap at 50,000 rows (safe for 512MB RAM)
            df = pd.read_sql(
                f"""
                SELECT amount_millions, hour_of_day, day_of_week, is_round_amount, is_new_beneficiary, is_anomaly
                FROM {self.gold_schema}.anomaly_training_data
                ORDER BY RANDOM()
                LIMIT 50000;
                """,
                self.engine,
            )
            if len(df) < 50:
                print("[Retrain] Anomaly: Not enough data. Skipping.")
                return {"status": "skipped", "reason": "insufficient data", "rows": len(df)}

            features = ["amount_millions", "hour_of_day", "day_of_week", "is_round_amount", "is_new_beneficiary"]
            X = df[features].fillna(0)
            scaler = StandardScaler()
            X_scaled = scaler.fit_transform(X)

            model = IsolationForest(contamination=0.1, random_state=42, n_estimators=100)
            model.fit(X_scaled)

            bundle = {
                "model": model,
                "preprocessor": scaler,
                "features": features,
                "version": datetime.now().strftime("%Y%m%d_%H%M"),
                "metrics": {"n_samples": len(df)},
            }
            with open(os.path.join(self.model_dir, "anomaly_bundle.pkl"), "wb") as f:
                pickle.dump(bundle, f)

            return {"status": "success", "rows_used": len(df)}
        except Exception as e:
            print(f"[Retrain] Anomaly failed: {e}")
            return {"status": "error", "error": str(e)}

    def _retrain_department(self) -> dict:
        """Retrain department risk model from gold.department_training_data."""
        try:
            df = pd.read_sql(
                f"""
                SELECT risk_category, inherent_likelihood, inherent_impact, findings_count,
                       kpi_below_target, kpi_volatility, previous_risk_score, assessment_month,
                       target_likelihood, target_impact
                FROM {self.gold_schema}.department_training_data
                LIMIT 25000;
                """,
                self.engine,
            )
            if len(df) < 30:
                return {"status": "skipped", "reason": "insufficient data"}

            feature_cols = ["risk_category", "inherent_likelihood", "inherent_impact", "findings_count", "kpi_below_target", "kpi_volatility", "previous_risk_score", "assessment_month"]
            target_cols = ["target_likelihood", "target_impact"]

            cat_features = ["risk_category"]
            num_features = [c for c in feature_cols if c not in cat_features]

            preprocessor = ColumnTransformer([
                ("cat", OneHotEncoder(handle_unknown="ignore"), cat_features),
                ("num", StandardScaler(), num_features),
            ])

            X = preprocessor.fit_transform(df[feature_cols])
            y = df[target_cols].values

            X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)
            model = MultiOutputClassifier(GradientBoostingClassifier(n_estimators=100, random_state=42))
            model.fit(X_train, y_train)

            y_pred = model.predict(X_test)
            r2 = r2_score(y_test, y_pred)

            bundle = {
                "model": model,
                "preprocessor": preprocessor,
                "version": datetime.now().strftime("%Y%m%d_%H%M"),
                "metrics": {"r2_score": round(r2, 4), "n_samples": len(df)},
            }
            with open(os.path.join(self.model_dir, "department_bundle.pkl"), "wb") as f:
                pickle.dump(bundle, f)

            return {"status": "success", "r2_score": round(r2, 4), "rows_used": len(df)}
        except Exception as e:
            print(f"[Retrain] Department failed: {e}")
            return {"status": "error", "error": str(e)}

    def _retrain_document(self) -> dict:
        """Retrain document NLP model from gold.document_training_data."""
        try:
            df = pd.read_sql(
                f"""
                SELECT text_input, risk_category
                FROM {self.gold_schema}.document_training_data
                WHERE text_input IS NOT NULL
                LIMIT 25000;
                """,
                self.engine,
            )
            if len(df) < 20:
                return {"status": "skipped", "reason": "insufficient data"}

            vectorizer = TfidfVectorizer(max_features=5000, ngram_range=(1, 2))
            X = vectorizer.fit_transform(df["text_input"])
            y = df["risk_category"]

            X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)
            model = LogisticRegression(max_iter=500, random_state=42)
            model.fit(X_train, y_train)
            acc = accuracy_score(y_test, model.predict(X_test))

            bundle = {
                "vectorizer": vectorizer,
                "category_model": model,
                "version": datetime.now().strftime("%Y%m%d_%H%M"),
                "metrics": {"accuracy": round(acc, 4), "n_samples": len(df)},
            }
            with open(os.path.join(self.model_dir, "document_bundle.pkl"), "wb") as f:
                pickle.dump(bundle, f)

            return {"status": "success", "accuracy": round(acc, 4), "rows_used": len(df)}
        except Exception as e:
            print(f"[Retrain] Document failed: {e}")
            return {"status": "error", "error": str(e)}

    def _retrain_kpi(self) -> dict:
        """Retrain KPI LSTM from gold.mart_kpi_series or kpi_training_data, grouped by KPI name."""
        try:
            # Priority: gold.mart_kpi_series, fallback to gold.kpi_training_data
            try:
                df = pd.read_sql(
                    f"SELECT kpi_name, actual_value, period FROM {self.gold_schema}.mart_kpi_series ORDER BY kpi_name, period",
                    self.engine,
                )
            except Exception:
                df = pd.DataFrame()

            if df.empty or len(df) < 10:
                try:
                    df = pd.read_sql(
                        f"SELECT 'DEFAULT' AS kpi_name, actual_value, period FROM {self.gold_schema}.kpi_training_data ORDER BY period",
                        self.engine,
                    )
                except Exception:
                    df = pd.DataFrame()

            if len(df) < 10:
                return {"status": "skipped", "reason": "insufficient data", "rows": len(df)}

            # Process sequences partitioned per KPI to avoid cross-metric contamination
            seq_len = 3
            all_X, all_y = [], []
            scalers = {}

            for kpi_name, group in df.groupby("kpi_name"):
                if len(group) <= seq_len:
                    continue
                vals = group["actual_value"].values.reshape(-1, 1).astype(np.float32)
                scaler = MinMaxScaler()
                scaled = scaler.fit_transform(vals)
                scalers[kpi_name] = scaler

                for i in range(len(scaled) - seq_len):
                    all_X.append(scaled[i:i + seq_len])
                    all_y.append(scaled[i + seq_len])

            if len(all_X) < 5:
                return {"status": "skipped", "reason": "not enough sequences created", "n_seq": len(all_X)}

            X_t = torch.tensor(np.array(all_X), dtype=torch.float32)
            y_t = torch.tensor(np.array(all_y), dtype=torch.float32)

            model = PyTorchLSTMRegressor(input_dim=1, hidden_dim=64, num_layers=2, output_dim=1)
            optimizer = torch.optim.Adam(model.parameters(), lr=0.001)
            criterion = nn.MSELoss()

            model.train()
            final_loss = 0.0
            for epoch in range(100):
                optimizer.zero_grad()
                output = model(X_t)
                loss = criterion(output, y_t)
                loss.backward()
                optimizer.step()
                final_loss = loss.item()

            # Save model and artifacts
            torch.save(model.state_dict(), os.path.join(self.model_dir, "kpi_lstm.pth"))
            bundle = {
                "scalers": scalers,
                "seq_len": seq_len,
                "version": datetime.now().strftime("%Y%m%d_%H%M"),
                "metrics": {"final_loss": round(final_loss, 6), "n_samples": len(all_X)},
            }
            with open(os.path.join(self.model_dir, "kpi_bundle.pkl"), "wb") as f:
                pickle.dump(bundle, f)

            return {"status": "success", "loss": round(final_loss, 6), "sequences_trained": len(all_X)}
        except Exception as e:
            print(f"[Retrain] KPI failed: {e}")
            return {"status": "error", "error": str(e)}
