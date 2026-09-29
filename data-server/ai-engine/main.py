"""
AuditSphere AI Engine — Data Hub Version
Migrated from backend/python-ai/main.py to run on Data Hub server.
Reads training data from Gold zone instead of CSV files.
Supports auto-retrain when new Gold data arrives.
"""
import os
import pickle
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from datetime import datetime
from contextlib import asynccontextmanager
from typing import List, Optional

from fastapi import FastAPI, HTTPException, BackgroundTasks, Depends, Security
from fastapi.security import APIKeyHeader
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from sqlalchemy import create_engine, text
from sqlalchemy.orm import sessionmaker

# ─── Config ──────────────────────────────────────────────────────────────────
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://etl_user:password@localhost:5432/auditsphere_datalake")
GOLD_SCHEMA = os.getenv("GOLD_SCHEMA", "gold")
API_KEY = os.getenv("API_KEY", "dev-ai-api-key")
MODEL_DIR = os.getenv("MODEL_DIR", "/app/models")

engine = create_engine(DATABASE_URL, pool_size=3, max_overflow=5, pool_pre_ping=True)
device = torch.device("cpu")

# ─── API Key Auth ────────────────────────────────────────────────────────────
api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)


async def verify_api_key(api_key: str = Security(api_key_header)):
    if api_key is None or api_key != API_KEY:
        raise HTTPException(status_code=401, detail="Invalid or missing API key")
    return api_key


# ─── LSTM Model Class ────────────────────────────────────────────────────────
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


# ─── Model Loading ───────────────────────────────────────────────────────────
anomaly_bundle = None
department_bundle = None
document_bundle = None
kpi_bundle = None
kpi_lstm_model = None


def load_models():
    global anomaly_bundle, department_bundle, document_bundle, kpi_bundle, kpi_lstm_model
    os.makedirs(MODEL_DIR, exist_ok=True)

    bundles = [
        ("anomaly_bundle.pkl", "anomaly_bundle"),
        ("department_bundle.pkl", "department_bundle"),
        ("document_bundle.pkl", "document_bundle"),
        ("kpi_bundle.pkl", "kpi_bundle"),
    ]
    g = globals()
    for fname, varname in bundles:
        path = os.path.join(MODEL_DIR, fname)
        if os.path.exists(path):
            try:
                with open(path, "rb") as f:
                    g[varname] = pickle.load(f)
                print(f"[AI] ✓ Loaded {fname}")
            except Exception as e:
                print(f"[AI] ✗ Failed to load {fname}: {e}")

    lstm_path = os.path.join(MODEL_DIR, "kpi_lstm.pth")
    if os.path.exists(lstm_path):
        try:
            model = PyTorchLSTMRegressor().to(device)
            model.load_state_dict(torch.load(lstm_path, map_location=device))
            model.eval()
            kpi_lstm_model = model
            print("[AI] ✓ Loaded kpi_lstm.pth")
        except Exception as e:
            print(f"[AI] ✗ Failed to load LSTM: {e}")


# ─── Helpers ─────────────────────────────────────────────────────────────────
def score_to_risk_level(score: float) -> str:
    if score >= 20: return "HIGH"
    elif score >= 13: return "MODERATE_HIGH"
    elif score >= 8: return "MODERATE"
    elif score >= 4: return "LOW_MODERATE"
    return "LOW"


def get_current_target_timeline() -> str:
    now = datetime.now()
    quarter = (now.month - 1) // 3 + 1
    return f"Q{quarter} {now.year}"


# ─── Request Models ──────────────────────────────────────────────────────────
class DepartmentRiskRequest(BaseModel):
    entity: str = "Jakarta Branch"
    risk_category: str = "Financial"
    inherent_likelihood: float = Field(3.0, ge=1.0, le=5.0)
    inherent_impact: float = Field(4.0, ge=1.0, le=5.0)
    audit_findings_count: float = Field(5.0, ge=0.0)
    kpi_below_target_count: float = Field(2.0, ge=0.0)
    kpi_volatility: float = Field(0.15, ge=0.0)
    previous_risk_score: float = Field(12.0, ge=0.0)
    assessment_month: int = Field(6, ge=1, le=12)


class AnomalyRequest(BaseModel):
    entity: str = "Jakarta Branch"
    description: Optional[str] = "Pembayaran vendor"
    amount: Optional[float] = Field(15.5, ge=0.0)
    hour_of_day: Optional[int] = Field(14, ge=0, le=23)
    day_of_week: Optional[int] = Field(3, ge=1, le=7)
    is_new_beneficiary: Optional[int] = Field(0, ge=0, le=1)
    is_round_amount: Optional[int] = Field(0, ge=0, le=1)
    category: Optional[str] = None
    channel: Optional[str] = "Teller"


class TextAnalysisRequest(BaseModel):
    text: str = "Ditemukan indikasi ketidaksesuaian prosedur dalam otorisasi transaksi kas besar."


class PerformanceTrendRequest(BaseModel):
    kpi_name: Optional[str] = "NPL Ratio"
    historical_data: List[float] = Field(default_factory=lambda: [80.0, 82.0, 85.0, 81.0, 79.0])


class BatchPredictRequest(BaseModel):
    items: List[DepartmentRiskRequest]


class ReturnRetrainResult(BaseModel):
    status: str
    models_retrained: List[str] = []
    accuracy_comparison: dict = {}


# ─── Lifespan ────────────────────────────────────────────────────────────────
@asynccontextmanager
async def lifespan(app: FastAPI):
    load_models()

    # Start scheduler for auto-retrain
    try:
        from apscheduler.schedulers.background import BackgroundScheduler
        scheduler = BackgroundScheduler()
        scheduler.add_job(check_and_retrain, 'interval', hours=6, id='auto_retrain')
        scheduler.start()
        print("[AI] ✓ Auto-retrain scheduler started (every 6 hours)")
    except Exception as e:
        print(f"[AI] ! Scheduler not started: {e}")

    yield
    engine.dispose()


# ─── Auto Retrain ────────────────────────────────────────────────────────────
def check_and_retrain():
    """Check for new Gold data and retrain if available."""
    try:
        with engine.connect() as conn:
            result = conn.execute(text(
                f"SELECT COUNT(*) FROM {GOLD_SCHEMA}.data_freshness_log WHERE is_new = TRUE"
            ))
            new_count = result.scalar() or 0

            if new_count > 0:
                print(f"[AI] New data detected ({new_count} updates). Starting retrain...")
                from retrain_engine import AutoRetrainEngine
                retrain = AutoRetrainEngine(engine, MODEL_DIR, GOLD_SCHEMA)
                retrain.retrain_all()

                # Mark as processed
                conn.execute(text(
                    f"UPDATE {GOLD_SCHEMA}.data_freshness_log SET is_new = FALSE WHERE is_new = TRUE"
                ))
                conn.commit()
                print("[AI] ✓ Retrain completed.")
                load_models()  # Hot-reload
            else:
                print("[AI] No new data. Skipping retrain.")
    except Exception as e:
        print(f"[AI] ✗ Retrain check failed: {e}")


# ─── App ─────────────────────────────────────────────────────────────────────
app = FastAPI(
    title="AuditSphere AI Engine",
    description="AI prediction engine for risk scoring, anomaly detection, NLP analysis, and KPI forecasting. Reads from Gold zone.",
    version="2.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ─── Health (no auth) ───────────────────────────────────────────────────────
@app.get("/health")
def health_check():
    return {
        "status": "ok",
        "models": {
            "anomaly": anomaly_bundle is not None,
            "department": department_bundle is not None,
            "document": document_bundle is not None,
            "kpi_lstm": kpi_lstm_model is not None and kpi_bundle is not None,
        },
    }


# ─── 1. Department Risk Scoring ─────────────────────────────────────────────
@app.post("/predict/risk-score", dependencies=[Depends(verify_api_key)])
def predict_department_risk(req: DepartmentRiskRequest):
    if department_bundle is None:
        imp = int(round(req.inherent_impact))
        lik = int(round(req.inherent_likelihood))
        score = float(imp * lik)
        return {
            "entity": req.entity,
            "type": "Department" if "Dept" in req.entity else "Branch",
            "risk_category": req.risk_category,
            "predicted_impact": imp,
            "predicted_likelihood": lik,
            "predicted_score": score,
            "actual_score": int(req.inherent_impact * req.inherent_likelihood),
            "risk_level": score_to_risk_level(score),
            "actual_risk_level": score_to_risk_level(req.inherent_impact * req.inherent_likelihood),
            "confidence": 0.88,
            "delta": round(score - (req.inherent_impact * req.inherent_likelihood), 1),
            "trend": "up" if score > 12 else "stable",
            "target_timeline": get_current_target_timeline(),
            "model_version": "fallback",
        }

    try:
        model = department_bundle["model"]
        preprocessor = department_bundle["preprocessor"]
        X_input = pd.DataFrame([{
            "Kategori Risiko": req.risk_category,
            "Inherent Likelihood": req.inherent_likelihood,
            "Inherent Impact": req.inherent_impact,
            "Jml Temuan Audit": req.audit_findings_count,
            "Jml KPI di Bawah Target": req.kpi_below_target_count,
            "Volatilitas KPI": req.kpi_volatility,
            "Skor Risiko Periode Lalu": req.previous_risk_score,
            "Bulan Penilaian": req.assessment_month,
        }])
        X_processed = preprocessor.transform(X_input)
        prediction = model.predict(X_processed)[0]
        pred_lik = int(round(prediction[0]))
        pred_imp = int(round(prediction[1]))
        pred_lik = max(1, min(5, pred_lik))
        pred_imp = max(1, min(5, pred_imp))
        pred_score = float(pred_lik * pred_imp)
        actual_score = int(req.inherent_impact * req.inherent_likelihood)

        return {
            "entity": req.entity,
            "type": "Department" if "Dept" in req.entity else "Branch",
            "risk_category": req.risk_category,
            "predicted_impact": pred_imp,
            "predicted_likelihood": pred_lik,
            "predicted_score": pred_score,
            "actual_score": actual_score,
            "risk_level": score_to_risk_level(pred_score),
            "actual_risk_level": score_to_risk_level(actual_score),
            "confidence": float(department_bundle.get("metrics", {}).get("r2_score", 0.85)),
            "delta": round(pred_score - actual_score, 1),
            "trend": "up" if pred_score > actual_score else "down" if pred_score < actual_score else "stable",
            "target_timeline": get_current_target_timeline(),
            "model_version": department_bundle.get("version", "1.0"),
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Prediction failed: {str(e)}")


# ─── 2. Anomaly Detection ───────────────────────────────────────────────────
@app.post("/predict/anomaly", dependencies=[Depends(verify_api_key)])
def predict_anomaly(req: AnomalyRequest):
    if anomaly_bundle is None:
        score = 0.3
        if req.amount and req.amount > 100: score += 0.2
        if req.hour_of_day and (req.hour_of_day < 7 or req.hour_of_day >= 20): score += 0.2
        if req.is_round_amount: score += 0.1
        if req.is_new_beneficiary: score += 0.1
        score = min(score, 0.99)
        is_anomaly = score > 0.5

        return {
            "entity": req.entity,
            "anomaly_score": round(score, 4),
            "is_anomaly": is_anomaly,
            "severity": "Critical" if score > 0.8 else "High" if score > 0.6 else "Medium" if score > 0.4 else "Low",
            "category": req.category or "Transaction",
            "amount": req.amount,
            "description": req.description,
            "risk_level": "HIGH" if is_anomaly else "MODERATE",
            "model_version": "fallback",
        }

    try:
        model = anomaly_bundle["model"]
        preprocessor = anomaly_bundle.get("preprocessor")
        X_input = pd.DataFrame([{
            "amount_millions": req.amount or 0,
            "hour_of_day": req.hour_of_day or 12,
            "day_of_week": req.day_of_week or 3,
            "is_round_amount": req.is_round_amount or 0,
            "is_new_beneficiary": req.is_new_beneficiary or 0,
        }])
        if preprocessor:
            X_input = preprocessor.transform(X_input)
        prediction = model.predict(X_input)[0]
        score_raw = model.decision_function(X_input)[0] if hasattr(model, 'decision_function') else 0.5
        anomaly_score = round(1 / (1 + np.exp(-score_raw * -1)), 4)
        is_anomaly = prediction == -1

        return {
            "entity": req.entity,
            "anomaly_score": anomaly_score,
            "is_anomaly": bool(is_anomaly),
            "severity": "Critical" if anomaly_score > 0.8 else "High" if anomaly_score > 0.6 else "Medium" if anomaly_score > 0.4 else "Low",
            "category": req.category or "Transaction",
            "amount": req.amount,
            "description": req.description,
            "risk_level": "HIGH" if is_anomaly else "MODERATE",
            "model_version": anomaly_bundle.get("version", "1.0"),
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Anomaly prediction failed: {str(e)}")


# ─── 3. Text/NLP Analysis ───────────────────────────────────────────────────
@app.post("/predict/text-analysis", dependencies=[Depends(verify_api_key)])
def predict_text_analysis(req: TextAnalysisRequest):
    if document_bundle is None:
        text_lower = req.text.lower()
        if any(w in text_lower for w in ["fraud", "korupsi", "manipulasi", "ilegal"]):
            cat, severity, sentiment = "Critical Finding", 5, "Negative"
        elif any(w in text_lower for w in ["ketidaksesuaian", "pelanggaran", "tidak sesuai"]):
            cat, severity, sentiment = "Compliance Issue", 4, "Negative"
        elif any(w in text_lower for w in ["kelemahan", "kurang", "belum"]):
            cat, severity, sentiment = "Weakness", 3, "Neutral"
        else:
            cat, severity, sentiment = "Observation", 2, "Positive"

        return {
            "text_input": req.text[:200],
            "predicted_category": cat,
            "sentiment": sentiment,
            "severity_score": severity,
            "confidence": 0.75,
            "impact": min(5, severity),
            "likelihood": min(5, max(1, severity - 1)),
            "risk_level": score_to_risk_level(severity * max(1, severity - 1)),
            "model_version": "fallback",
        }

    try:
        vectorizer = document_bundle["vectorizer"]
        cat_model = document_bundle["category_model"]
        X = vectorizer.transform([req.text])
        category = cat_model.predict(X)[0]
        proba = cat_model.predict_proba(X)[0] if hasattr(cat_model, 'predict_proba') else [0.8]
        confidence = float(max(proba))

        return {
            "text_input": req.text[:200],
            "predicted_category": category,
            "sentiment": "Negative" if confidence > 0.7 else "Neutral",
            "severity_score": 4 if confidence > 0.8 else 3,
            "confidence": round(confidence, 4),
            "impact": 4 if confidence > 0.8 else 3,
            "likelihood": 3,
            "risk_level": score_to_risk_level(12 if confidence > 0.8 else 9),
            "model_version": document_bundle.get("version", "1.0"),
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Text analysis failed: {str(e)}")


# ─── 4. KPI Forecasting ─────────────────────────────────────────────────────
@app.post("/predict/performance-trend", dependencies=[Depends(verify_api_key)])
def predict_performance_trend(req: PerformanceTrendRequest):
    if kpi_lstm_model is None or kpi_bundle is None:
        data = req.historical_data
        avg = sum(data) / len(data) if data else 0
        trend = "Improving" if len(data) >= 2 and data[-1] > data[-2] else "Declining" if len(data) >= 2 and data[-1] < data[-2] else "Stable"
        forecast = round(avg * 1.02, 2) if trend == "Improving" else round(avg * 0.98, 2) if trend == "Declining" else round(avg, 2)

        return {
            "kpi_name": req.kpi_name,
            "current_value": data[-1] if data else 0,
            "forecasted_value": forecast,
            "trend": trend,
            "alert_level": "Warning" if trend == "Declining" else "Normal",
            "risk_level": "MODERATE" if trend == "Declining" else "LOW",
            "historical_data": data,
            "model_version": "fallback",
        }

    try:
        scaler = kpi_bundle.get("scaler")
        data = np.array(req.historical_data).reshape(-1, 1)
        if scaler:
            data = scaler.transform(data)
        tensor = torch.tensor(data, dtype=torch.float32).unsqueeze(0).to(device)
        with torch.no_grad():
            forecast_scaled = kpi_lstm_model(tensor).item()
        forecast = scaler.inverse_transform([[forecast_scaled]])[0][0] if scaler else forecast_scaled
        current = req.historical_data[-1] if req.historical_data else 0
        trend = "Improving" if forecast > current else "Declining" if forecast < current else "Stable"

        return {
            "kpi_name": req.kpi_name,
            "current_value": current,
            "forecasted_value": round(forecast, 2),
            "trend": trend,
            "alert_level": "Warning" if trend == "Declining" else "Normal",
            "risk_level": "MODERATE" if trend == "Declining" else "LOW",
            "historical_data": req.historical_data,
            "model_version": kpi_bundle.get("version", "1.0"),
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"KPI forecast failed: {str(e)}")


# ─── Batch Prediction ───────────────────────────────────────────────────────
@app.post("/predict/batch-risk-score", dependencies=[Depends(verify_api_key)])
def batch_predict(req: BatchPredictRequest):
    results = []
    for item in req.items:
        try:
            r = predict_department_risk(item)
            results.append(r)
        except Exception as e:
            results.append({"entity": item.entity, "error": str(e)})
    return {"predictions": results, "total": len(results)}


# ─── Manual Retrain Trigger ──────────────────────────────────────────────────
@app.post("/retrain/auto", dependencies=[Depends(verify_api_key)])
def trigger_retrain(background_tasks: BackgroundTasks):
    background_tasks.add_task(check_and_retrain)
    return {"status": "scheduled", "message": "Retrain task scheduled. Models will hot-reload on completion."}


# ─── Model Info ──────────────────────────────────────────────────────────────
@app.get("/models/info", dependencies=[Depends(verify_api_key)])
def model_info():
    models = []
    for fname in os.listdir(MODEL_DIR):
        path = os.path.join(MODEL_DIR, fname)
        if os.path.isfile(path):
            models.append({
                "filename": fname,
                "size_mb": round(os.path.getsize(path) / (1024 * 1024), 2),
                "modified": datetime.fromtimestamp(os.path.getmtime(path)).isoformat(),
            })
    return {"model_dir": MODEL_DIR, "models": models}
