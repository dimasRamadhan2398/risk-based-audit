#!/bin/bash
set -e

echo "[AI Engine] Checking pre-trained models..."
MODEL_DIR="/app/models"

if [ -f "${MODEL_DIR}/xgboost_risk_model.pkl" ]; then
    echo "[AI Engine] ✓ Pre-trained models found. Loading..."
else
    echo "[AI Engine] ! No pre-trained models found. Will train on first request or Gold data arrival."
fi

echo "[AI Engine] Starting FastAPI server..."
exec uvicorn main:app --host 0.0.0.0 --port 8000 --workers 1
