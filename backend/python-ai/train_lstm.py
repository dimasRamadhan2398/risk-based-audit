import os
import numpy as np

# Use the CPU only to prevent CUDA segfaults
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"

import tensorflow as tf
from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import LSTM, Dense

MODEL_DIR = "models"

def train_lstm():
    print("Training LSTM...")
    np.random.seed(42)
    tf.random.set_seed(42)

    X = np.random.rand(100, 5, 1)
    y = np.random.rand(100, 1)

    model = Sequential()
    model.add(LSTM(10, activation='relu', input_shape=(5, 1)))
    model.add(Dense(1))
    model.compile(optimizer='adam', loss='mse')

    model.fit(X, y, epochs=5, verbose=0)

    model.save(os.path.join(MODEL_DIR, "lstm_model.keras"))
    print("LSTM trained and saved.")

if __name__ == "__main__":
    train_lstm()
