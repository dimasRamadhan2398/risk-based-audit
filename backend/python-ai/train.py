import os
import numpy as np
import pandas as pd
import xgboost as xgb
from sklearn.ensemble import IsolationForest
import pickle
from transformers import AutoTokenizer, AutoModelForSequenceClassification

MODEL_DIR = "models"

def ensure_model_dir():
    if not os.path.exists(MODEL_DIR):
        os.makedirs(MODEL_DIR)

def train_xgboost():
    print("Training XGBoost...")
    np.random.seed(42)
    X = np.random.rand(100, 3)
    y = np.random.rand(100)

    model = xgb.XGBRegressor(objective='reg:squarederror')
    model.fit(X, y)

    model.save_model(os.path.join(MODEL_DIR, "xgboost_model.json"))
    print("XGBoost trained and saved.")

def train_isolation_forest():
    print("Training Isolation Forest...")
    np.random.seed(42)
    X = np.random.randn(100, 2)
    X_outliers = np.random.uniform(low=-4, high=4, size=(20, 2))
    X_train = np.vstack([X, X_outliers])

    model = IsolationForest(contamination=0.1, random_state=42)
    model.fit(X_train)

    with open(os.path.join(MODEL_DIR, "isolation_forest.pkl"), "wb") as f:
        pickle.dump(model, f)
    print("Isolation Forest trained and saved.")

def train_indobert():
    print("Training/Loading IndoBERT...")
    try:
        model_name = "indobenchmark/indobert-base-p1"
        tokenizer = AutoTokenizer.from_pretrained(model_name)
        model = AutoModelForSequenceClassification.from_pretrained(model_name, num_labels=3)

        tokenizer.save_pretrained(os.path.join(MODEL_DIR, "indobert_tokenizer"))
        model.save_pretrained(os.path.join(MODEL_DIR, "indobert_model"))
        print("IndoBERT downloaded and saved.")
    except Exception as e:
        print(f"Error downloading IndoBERT (likely network issue). Fallback will be used during inference. Error: {e}")

if __name__ == "__main__":
    ensure_model_dir()
    train_xgboost()
    train_isolation_forest()
    train_indobert()

    # Run TF in a separate process to avoid segmentation faults
    print("Training LSTM in separate process...")
    os.system("python train_lstm.py")

    print("All training completed.")
