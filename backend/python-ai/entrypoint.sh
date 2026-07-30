#!/bin/sh
set -e

PRETRAINED_DIR="/app/pretrained_models"
MODELS_DIR="/app/models"

echo "=================================================="
echo "Checking AI Model Directory Status..."
echo "=================================================="

# Ensure target directory exists
mkdir -p "$MODELS_DIR"

# Check if target directory has the critical xgboost model
if [ ! -f "$MODELS_DIR/xgboost_model.json" ]; then
    if [ -d "$PRETRAINED_DIR" ] && [ "$(ls -A $PRETRAINED_DIR 2>/dev/null)" ]; then
        echo "Detected empty or incomplete mounted volume at $MODELS_DIR."
        echo "Copying pre-baked ML models from $PRETRAINED_DIR..."
        cp -R "$PRETRAINED_DIR"/* "$MODELS_DIR/"
        echo "Models successfully restored to volume."
    else
        echo "No pre-baked models found in $PRETRAINED_DIR."
        echo "Attempting to train/download models locally..."
        python train.py || echo "Warning: Local training failed. Fallbacks will be used during inference."
    fi
else
    echo "Existing models found in $MODELS_DIR. Skipping restoration."
fi

echo "=================================================="
echo "Starting FastAPI Application..."
echo "=================================================="

# Execute the main application command (passed from Docker CMD)
exec "$@"
