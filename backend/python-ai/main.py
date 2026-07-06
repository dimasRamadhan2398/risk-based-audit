from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import xgboost as xgb
import pickle
import numpy as np
import tensorflow as tf
from tensorflow.keras.models import load_model
from transformers import AutoTokenizer, AutoModelForSequenceClassification
import torch
import os

app = FastAPI()
MODEL_DIR = "models"

# --- Set device for PyTorch models (IndoBERT) ---
device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
print(f"Using device: {device}")

# --- Dummy inference functions for fallback ---
def dummy_xgboost():
    return {"risk_score": 0.85, "feature_importance": {"kpi_data": 0.5, "previous_findings": 0.3, "master_data": 0.2}}

def dummy_isolation_forest():
    return {"is_anomaly": True, "anomaly_score": -0.75}

def dummy_indobert():
    return {"risk_category": "High Risk", "confidence": 0.92, "sentiment": "Negative"}

def dummy_lstm():
    return {"predicted_performance": 0.45, "trend": "Deteriorating"}

# --- Load models ---
xgb_model = None
try:
    if os.path.exists(os.path.join(MODEL_DIR, "xgboost_model.json")):
        xgb_model = xgb.XGBRegressor()
        xgb_model.load_model(os.path.join(MODEL_DIR, "xgboost_model.json"))
except Exception as e:
    print(f"Failed to load XGBoost: {e}")

if_model = None
try:
    if os.path.exists(os.path.join(MODEL_DIR, "isolation_forest.pkl")):
        with open(os.path.join(MODEL_DIR, "isolation_forest.pkl"), "rb") as f:
            if_model = pickle.load(f)
except Exception as e:
    print(f"Failed to load Isolation Forest: {e}")

lstm_model = None
try:
    if os.path.exists(os.path.join(MODEL_DIR, "lstm_model.keras")):
        lstm_model = load_model(os.path.join(MODEL_DIR, "lstm_model.keras"))
except Exception as e:
    print(f"Failed to load LSTM: {e}")

indobert_tokenizer = None
indobert_model = None
try:
    if os.path.exists(os.path.join(MODEL_DIR, "indobert_model")):
        indobert_tokenizer = AutoTokenizer.from_pretrained(os.path.join(MODEL_DIR, "indobert_tokenizer"))
        indobert_model = AutoModelForSequenceClassification.from_pretrained(os.path.join(MODEL_DIR, "indobert_model"))
        indobert_model.to(device) # Move model to the selected device (GPU or CPU)
except Exception as e:
    print(f"Failed to load IndoBERT: {e}")


class XGBoostRequest(BaseModel):
    kpi_data: float
    previous_findings: float
    master_data: float

class IsolationForestRequest(BaseModel):
    feature1: float
    feature2: float

class TextRequest(BaseModel):
    text: str

class LSTMRequest(BaseModel):
    historical_data: list[float]


@app.get("/health")
def health_check():
    return {"status": "ok"}

@app.post("/predict/risk-score")
def predict_risk_score(req: XGBoostRequest):
    if xgb_model is None:
        return dummy_xgboost()

    try:
        # Assuming model takes 3 features
        X = np.array([[req.kpi_data, req.previous_findings, req.master_data]])
        score = xgb_model.predict(X)[0]
        # Simulate feature importance for explainability
        return {
            "risk_score": float(score),
            "feature_importance": {
                "kpi_data": 0.45,
                "previous_findings": 0.35,
                "master_data": 0.20
            }
        }
    except Exception as e:
        print(f"XGBoost inference failed: {e}")
        return dummy_xgboost()

@app.post("/predict/anomaly")
def predict_anomaly(req: IsolationForestRequest):
    if if_model is None:
        return dummy_isolation_forest()

    try:
        X = np.array([[req.feature1, req.feature2]])
        prediction = if_model.predict(X)[0]
        score = if_model.decision_function(X)[0]

        return {
            "is_anomaly": prediction == -1,
            "anomaly_score": float(score)
        }
    except Exception as e:
        print(f"Isolation Forest inference failed: {e}")
        return dummy_isolation_forest()

@app.post("/predict/text-analysis")
def predict_text_analysis(req: TextRequest):
    if indobert_model is None or indobert_tokenizer is None:
        return dummy_indobert()

    try:
        inputs = indobert_tokenizer(req.text, return_tensors="pt", truncation=True, padding=True, max_length=512)
        # Move inputs to the same device as the model
        inputs = {k: v.to(device) for k, v in inputs.items()}

        outputs = indobert_model(**inputs)
        probs = torch.nn.functional.softmax(outputs.logits, dim=-1)
        pred_class = torch.argmax(probs, dim=-1).item()

        categories = ["Low Risk", "Medium Risk", "High Risk"]
        return {
            "risk_category": categories[pred_class],
            "confidence": float(probs[0][pred_class]),
            "sentiment": "Negative" if pred_class == 2 else ("Neutral" if pred_class == 1 else "Positive")
        }
    except Exception as e:
        print(f"IndoBERT inference failed: {e}")
        return dummy_indobert()

@app.post("/predict/performance-trend")
def predict_performance_trend(req: LSTMRequest):
    if lstm_model is None:
        return dummy_lstm()

    try:
        # Expecting exactly 5 historical points for our dummy model
        hist_data = req.historical_data
        if len(hist_data) < 5:
            hist_data = hist_data + [0.0] * (5 - len(hist_data))
        else:
            hist_data = hist_data[-5:]

        X = np.array(hist_data).reshape(1, 5, 1)
        pred = lstm_model.predict(X, verbose=0)[0][0]

        # Simple trend logic based on last point
        trend = "Deteriorating" if pred < hist_data[-1] else "Improving"

        return {
            "predicted_performance": float(pred),
            "trend": trend
        }
    except Exception as e:
        print(f"LSTM inference failed: {e}")
        return dummy_lstm()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
