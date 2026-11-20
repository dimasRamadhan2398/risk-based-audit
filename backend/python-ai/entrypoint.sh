#!/bin/sh
set -e

PRETRAINED_DIR="/app/pretrained_models"
MODELS_DIR="/app/models"

echo "=================================================="
echo "Checking AI Model Directory Status..."
echo "=================================================="

# Ensure target directory exists
mkdir -p "$MODELS_DIR"

# Sync pre-baked models from image to mounted volume (update newer files or copy if missing)
if [ -d "$PRETRAINED_DIR" ] && [ "$(ls -A $PRETRAINED_DIR 2>/dev/null)" ]; then
    echo "Synchronizing pre-baked ML models from $PRETRAINED_DIR to $MODELS_DIR..."
    # cp -u updates existing destination files only if source is newer or destination is missing
    cp -u -R "$PRETRAINED_DIR"/* "$MODELS_DIR/" 2>/dev/null || cp -R "$PRETRAINED_DIR"/* "$MODELS_DIR/"
    echo "Models synchronized successfully to $MODELS_DIR."
elif [ ! -f "$MODELS_DIR/xgboost_model.json" ] && [ ! -f "$MODELS_DIR/anomaly_bundle.pkl" ]; then
    echo "No pre-baked models found. Attempting to train/download models locally..."
    python train.py || echo "Warning: Local training failed. Fallbacks will be used during inference."
fi

echo "=================================================="
echo "Starting FastAPI Application..."
echo "=================================================="

# Execute the main application command (passed from Docker CMD)
exec "$@"
