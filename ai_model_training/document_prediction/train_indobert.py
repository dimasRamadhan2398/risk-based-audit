import os
import random
import numpy as np
import pandas as pd
import torch
from torch.utils.data import Dataset, DataLoader
from transformers import AutoTokenizer, AutoModelForSequenceClassification
from torch.optim import AdamW
from sklearn.model_selection import train_test_split
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score

def set_seed(seed=42):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    if torch.cuda.is_available():
        torch.cuda.manual_seed_all(seed)
        torch.backends.cudnn.deterministic = True
        torch.backends.cudnn.benchmark = False

set_seed(42)

# ==========================================
# CONFIGURATION
# ==========================================
TARGET_COLUMN = 'TARGET: Kategori Risiko'  # Options: 'TARGET: Risk Level', 'TARGET: Sentimen', 'TARGET: Kategori Risiko'
NUM_EPOCHS = 3
BATCH_SIZE = 8
MAX_LEN = 128
LEARNING_RATE = 2e-5
QUICK_TRAIN = False  # Set to True to use a subset of 100 rows for fast testing, False for full dataset

# ==========================================
# 1. LOAD AND PREPROCESS DATA
# ==========================================
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'document_data.csv')
print(f"Loading dataset from: {dataset_path}")
df = pd.read_csv(dataset_path)

print("Original dataset shape:", df.shape)
print("Target column selected:", TARGET_COLUMN)

if QUICK_TRAIN:
    print(f"QUICK_TRAIN is active. Slicing dataset to the first 100 rows...")
    df = df.iloc[:100]

# Check columns
text_col = 'Teks Input (Kutipan dari Laporan Audit)'
if text_col not in df.columns or TARGET_COLUMN not in df.columns:
    raise ValueError(f"Required columns not found in dataset. Check column names.")

# Drop null values in target or text
df = df.dropna(subset=[text_col, TARGET_COLUMN])

# Get unique labels and map them to integers
unique_labels = sorted(df[TARGET_COLUMN].unique())
num_classes = len(unique_labels)
label_to_id = {label: idx for idx, label in enumerate(unique_labels)}
id_to_label = {idx: label for label, idx in label_to_id.items()}

print(f"Labels found ({num_classes} classes): {label_to_id}")
df['label_id'] = df[TARGET_COLUMN].map(label_to_id)

texts = df[text_col].values
labels = df['label_id'].values

# ==========================================
# 2. SPLITTING DATA
# ==========================================
print("\nSplitting dataset into train and validation sets (80/20)...")
X_train, X_val, y_train, y_val = train_test_split(
    texts, labels, test_size=0.2, random_state=0, stratify=labels
)
print(f"Train size: {len(X_train)} | Validation size: {len(X_val)}")

# ==========================================
# 3. TOKENIZER & DATASET DEFINITION
# ==========================================
MODEL_NAME = 'indobenchmark/indobert-base-p1'
print(f"\nLoading IndoBERT tokenizer: {MODEL_NAME}...")
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)

class AuditDataset(Dataset):
    def __init__(self, texts, labels, tokenizer, max_len):
        self.texts = texts
        self.labels = labels
        self.tokenizer = tokenizer
        self.max_len = max_len

    def __len__(self):
        return len(self.texts)

    def __getitem__(self, idx):
        text = str(self.texts[idx])
        label = self.labels[idx]
        
        encoding = self.tokenizer.encode_plus(
            text,
            add_special_tokens=True,
            max_length=self.max_len,
            padding='max_length',
            truncation=True,
            return_attention_mask=True,
            return_tensors='pt'
        )
        
        return {
            'input_ids': encoding['input_ids'].flatten(),
            'attention_mask': encoding['attention_mask'].flatten(),
            'label': torch.tensor(label, dtype=torch.long)
        }

train_dataset = AuditDataset(X_train, y_train, tokenizer, MAX_LEN)
val_dataset = AuditDataset(X_val, y_val, tokenizer, MAX_LEN)

train_loader = DataLoader(train_dataset, batch_size=BATCH_SIZE, shuffle=True)
val_loader = DataLoader(val_dataset, batch_size=BATCH_SIZE)

# ==========================================
# 4. INITIALIZE MODEL & DEVICE
# ==========================================
device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
print(f"\nUsing device: {device}")

print(f"Loading IndoBERT classification model (classes={num_classes})...")
model = AutoModelForSequenceClassification.from_pretrained(MODEL_NAME, num_labels=num_classes)
model = model.to(device)

optimizer = AdamW(model.parameters(), lr=LEARNING_RATE)

# ==========================================
# 5. TRAINING LOOP
# ==========================================
print("\nStarting model training...")
for epoch in range(NUM_EPOCHS):
    model.train()
    total_train_loss = 0
    
    for step, batch in enumerate(train_loader):
        input_ids = batch['input_ids'].to(device)
        attention_mask = batch['attention_mask'].to(device)
        labels_batch = batch['label'].to(device)
        
        model.zero_grad()
        
        outputs = model(
            input_ids=input_ids,
            attention_mask=attention_mask,
            labels=labels_batch
        )
        
        loss = outputs.loss
        total_train_loss += loss.item()
        
        loss.backward()
        optimizer.step()
        
        if (step + 1) % 5 == 0 or (step + 1) == len(train_loader):
            print(f"Epoch {epoch+1}/{NUM_EPOCHS} | Batch {step+1}/{len(train_loader)} | Loss: {loss.item():.4f}")
            
    avg_loss = total_train_loss / len(train_loader)
    print(f"Epoch {epoch+1} complete. Average Loss: {avg_loss:.4f}")

# ==========================================
# 6. EVALUATION
# ==========================================
print("\nEvaluating model on validation set...")
model.eval()
all_preds = []
all_targets = []

with torch.no_grad():
    for batch in val_loader:
        input_ids = batch['input_ids'].to(device)
        attention_mask = batch['attention_mask'].to(device)
        labels_batch = batch['label']
        
        outputs = model(input_ids=input_ids, attention_mask=attention_mask)
        logits = outputs.logits
        preds = torch.argmax(logits, dim=1).cpu().numpy()
        
        all_preds.extend(preds)
        all_targets.extend(labels_batch.numpy())

accuracy = accuracy_score(all_targets, all_preds)
print("\n--- Validation Set Evaluation ---")
print(f"Accuracy Score: {accuracy * 100:.2f} %")

print("\nConfusion Matrix:")
print(confusion_matrix(all_targets, all_preds))

target_names = [id_to_label[i] for i in range(num_classes)]
print("\nClassification Report:")
print(classification_report(all_targets, all_preds, target_names=target_names))

# ==========================================
# 7. SAVE MODEL
# ==========================================
output_dir = os.path.join(current_dir, 'model')
if not os.path.exists(output_dir):
    os.makedirs(output_dir)

print(f"\nSaving model and tokenizer to: {output_dir}")
model.save_pretrained(output_dir)
tokenizer.save_pretrained(output_dir)

print("\nIndoBERT Training script executed successfully!")
