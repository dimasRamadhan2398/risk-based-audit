import os
import re
import random
import numpy as np
import pandas as pd
import torch
from torch.utils.data import Dataset, DataLoader
from transformers import AutoTokenizer, AutoModelForSequenceClassification
from torch.optim import AdamW
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.svm import LinearSVC
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score, f1_score

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
MODEL_NAME = 'indobenchmark/indobert-base-p1'
NUM_EPOCHS = 3
BATCH_SIZE = 32
MAX_LEN = 128
LEARNING_RATE = 2e-5
QUICK_TRAIN = False  # Set to True to use a subset of 100 rows for fast testing

# Text Preprocessing Function
def clean_audit_text(text):
    text = str(text)
    # Remove leading numbering patterns like "1. ", "10. ", "100. "
    text = re.sub(r'^\s*\d+[\.\)]\s*', '', text)
    # Strip surrounding whitespace
    return text.strip()

# ==========================================
# 1. LOAD AND PREPROCESS DATA
# ==========================================
current_dir = os.path.dirname(os.path.abspath(__file__))
dataset_path = os.path.join(current_dir, 'document_data.csv')
print(f"Loading dataset from: {dataset_path}")
df = pd.read_csv(dataset_path)
initial_len = len(df)

# Deduplicate identical records
df = df.drop_duplicates().reset_index(drop=True)
print(f"Original rows: {initial_len} | Deduplicated clean rows: {len(df)}")

text_col = 'Teks Input (Kutipan dari Laporan Audit)'
if text_col not in df.columns:
    raise ValueError(f"Required text column '{text_col}' not found in dataset.")

# Apply text cleaning
df['text_clean'] = df[text_col].apply(clean_audit_text)

# Identify the 4 Target Columns
target_columns_map = {
    'TARGET: Kategori Risiko': [c for c in df.columns if 'Kategori Risiko' in c][0],
    'TARGET: Sentimen': [c for c in df.columns if 'Sentimen' in c][0],
    'TARGET: Impact (1-5)': [c for c in df.columns if 'Impact' in c][0],
    'TARGET: Likelihood (1-5)': [c for c in df.columns if 'Likelihood' in c][0]
}

print("\n--- Identified 4 Target Columns ---")
for key, col in target_columns_map.items():
    print(f" - {key}: '{col}' (Unique classes: {df[col].nunique()})")

if QUICK_TRAIN:
    print(f"\nQUICK_TRAIN active. Using first 100 rows for testing...")
    df = df.iloc[:100].reset_index(drop=True)

# Dataset Class for PyTorch IndoBERT
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

# Device selection
device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
print(f"\nUsing compute device: {device}")

print(f"Loading IndoBERT Tokenizer ({MODEL_NAME})...")
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)

overall_results = {}

# ==========================================
# 2. TRAINING & EVALUATION LOOP FOR ALL 4 TARGETS
# ==========================================
for target_label, target_col in target_columns_map.items():
    print(f"\n======================================================================")
    print(f" TRAINING & EVALUATING INDOBERT FOR TARGET: {target_label} ")
    print(f"======================================================================")
    
    # Clean null values in target
    sub_df = df.dropna(subset=['text_clean', target_col]).copy()
    
    le = LabelEncoder()
    sub_df['label_id'] = le.fit_transform(sub_df[target_col].astype(str))
    
    unique_classes = list(le.classes_)
    num_classes = len(unique_classes)
    
    print(f"Target Column: '{target_col}' | Classes count: {num_classes}")
    print(f"Classes list: {unique_classes}")
    
    texts = sub_df['text_clean'].values
    labels = sub_df['label_id'].values
    
    # Check class frequency for safe stratification
    class_counts = pd.Series(labels).value_counts()
    min_samples = class_counts.min()
    
    if min_samples < 2:
        rare_classes = [unique_classes[idx] for idx in class_counts[class_counts < 2].index]
        print(f"Warning: Class(es) {rare_classes} have fewer than 2 samples (min={min_samples}). Disabling stratify for train_test_split.")
        stratify_param = None
    else:
        stratify_param = labels

    # 80/20 Train-Test Split
    X_train, X_val, y_train, y_val = train_test_split(
        texts, labels, test_size=0.2, random_state=42, stratify=stratify_param
    )
    print(f"Train samples: {len(X_train)} | Validation samples: {len(X_val)}")
    
    # PyTorch DataLoaders
    train_dataset = AuditDataset(X_train, y_train, tokenizer, MAX_LEN)
    val_dataset = AuditDataset(X_val, y_val, tokenizer, MAX_LEN)
    
    train_loader = DataLoader(train_dataset, batch_size=BATCH_SIZE, shuffle=True)
    val_loader = DataLoader(val_dataset, batch_size=BATCH_SIZE)
    
    # Load IndoBERT Model
    model = AutoModelForSequenceClassification.from_pretrained(MODEL_NAME, num_labels=num_classes)
    model = model.to(device)
    
    optimizer = AdamW(model.parameters(), lr=LEARNING_RATE)
    
    # Training Loop
    print(f"Fine-tuning IndoBERT for {NUM_EPOCHS} epochs...")
    for epoch in range(NUM_EPOCHS):
        model.train()
        total_loss = 0
        for step, batch in enumerate(train_loader):
            input_ids = batch['input_ids'].to(device)
            attention_mask = batch['attention_mask'].to(device)
            labels_b = batch['label'].to(device)
            
            optimizer.zero_grad()
            outputs = model(input_ids=input_ids, attention_mask=attention_mask, labels=labels_b)
            loss = outputs.loss
            total_loss += loss.item()
            loss.backward()
            optimizer.step()
            
        avg_loss = total_loss / len(train_loader)
        print(f" Epoch {epoch+1}/{NUM_EPOCHS} | Train Loss: {avg_loss:.4f}", flush=True)
        
    # Validation Loop
    model.eval()
    all_preds = []
    all_targets = []
    
    with torch.no_grad():
        for batch in val_loader:
            input_ids = batch['input_ids'].to(device)
            attention_mask = batch['attention_mask'].to(device)
            labels_b = batch['label']
            
            outputs = model(input_ids=input_ids, attention_mask=attention_mask)
            preds = torch.argmax(outputs.logits, dim=1).cpu().numpy()
            
            all_preds.extend(preds)
            all_targets.extend(labels_b.numpy())
            
    val_acc = accuracy_score(all_targets, all_preds)
    val_f1 = f1_score(all_targets, all_preds, average='weighted', zero_division=0)
    
    # Baseline Comparison: TF-IDF + Logistic Regression
    tfidf_vec = TfidfVectorizer(ngram_range=(1, 2), max_features=3000, sublinear_tf=True)
    X_tr_tfidf = tfidf_vec.fit_transform(X_train)
    X_val_tfidf = tfidf_vec.transform(X_val)
    clf_baseline = LogisticRegression(max_iter=1000, random_state=42)
    clf_baseline.fit(X_tr_tfidf, y_train)
    baseline_preds = clf_baseline.predict(X_val_tfidf)
    baseline_acc = accuracy_score(y_val, baseline_preds)
    baseline_f1 = f1_score(y_val, baseline_preds, average='weighted', zero_division=0)
    
    report_str = classification_report(all_targets, all_preds, labels=list(range(num_classes)), target_names=unique_classes, zero_division=0)
    
    overall_results[target_label] = {
        'num_classes': num_classes,
        'indobert_acc': val_acc,
        'indobert_f1': val_f1,
        'baseline_acc': baseline_acc,
        'baseline_f1': baseline_f1,
        'confusion_matrix': confusion_matrix(all_targets, all_preds, labels=list(range(num_classes))),
        'classification_report': report_str
    }
    
    print(f"\n--- Validation Results ({target_label}) ---")
    print(f"IndoBERT Accuracy : {val_acc * 100:.2f} % (Weighted F1: {val_f1:.4f})")
    print(f"TF-IDF Baseline   : {baseline_acc * 100:.2f} % (Weighted F1: {baseline_f1:.4f})")
    print("\nConfusion Matrix (IndoBERT):")
    print(confusion_matrix(all_targets, all_preds, labels=list(range(num_classes))))
    print("\nClassification Report (IndoBERT):")
    print(report_str)
    
    # Cleanup memory for next target
    del model
    if torch.cuda.is_available():
        torch.cuda.empty_cache()

# ==========================================
# 3. OVERALL ACCURACY SUMMARY REPORT
# ==========================================
print("\n======================================================================")
print(" FINAL ACCURACY SUMMARY REPORT FOR ALL 4 TARGET COLUMNS ")
print("======================================================================")
print(f"{'Target Column':<30} | {'Classes':<8} | {'IndoBERT Acc':<15} | {'TF-IDF Baseline':<15}")
print("-" * 75)
for t_label, res in overall_results.items():
    print(f"{t_label:<30} | {res['num_classes']:<8} | {res['indobert_acc']*100:>6.2f} %        | {res['baseline_acc']*100:>6.2f} %")

print("\nIndoBERT 4-Target Training and Benchmark completed successfully!")
