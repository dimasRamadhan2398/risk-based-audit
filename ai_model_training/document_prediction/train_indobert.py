import os
import re
import random
import numpy as np
import pandas as pd
import torch
from torch.utils.data import Dataset, DataLoader
from transformers import AutoTokenizer, AutoModelForSequenceClassification, AutoConfig
from torch.optim import AdamW
from sklearn.model_selection import StratifiedGroupKFold
from sklearn.preprocessing import LabelEncoder
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression, Ridge
from sklearn.svm import LinearSVC
from sklearn.metrics import confusion_matrix, classification_report, accuracy_score, f1_score, mean_absolute_error

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
BATCH_SIZE = 8
MAX_LEN = 128
LEARNING_RATE = 2e-5
QUICK_TRAIN = False  # Set to True to use a subset of 100 rows for fast testing

ACTIVE_TARGETS = [
    'TARGET: Kategori Risiko',
    'TARGET: Sentimen',
    'TARGET: Impact (1-5)',
    'TARGET: Likelihood (1-5)'
]

ROBUST_TARGETS = [
    'TARGET: Kategori Risiko',
    'TARGET: Sentimen',
    'TARGET: Impact (1-5)',
    'TARGET: Likelihood (1-5)'
]

RUN_INDOBERT_GROUP_CV = False

RUN_FINAL_TRAINING = True
VERIFY_SAVED_MODELS = True

FINAL_MODEL_DIRS = {
    'TARGET: Kategori Risiko': 'kategori_risiko',
    'TARGET: Sentimen': 'sentimen',
    'TARGET: Impact (1-5)': 'impact',
    'TARGET: Likelihood (1-5)': 'likelihood'
}

def build_indobert_classifier(num_classes, unique_classes):
    id2label = {i: str(label) for i, label in enumerate(unique_classes)}
    label2id = {str(label): i for i, label in enumerate(unique_classes)}

    config = AutoConfig.from_pretrained(MODEL_NAME)
    config.num_labels = num_classes
    config.id2label = id2label
    config.label2id = label2id
    config.problem_type = "single_label_classification"

    return AutoModelForSequenceClassification.from_pretrained(
        MODEL_NAME,
        config=config,
        ignore_mismatched_sizes=True
    )

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

text_duplicate_count = df.duplicated(subset=['text_clean']).sum()
unique_text_count = df['text_clean'].nunique()

print(f"Unique cleaned texts: {unique_text_count}")
print(f"Duplicate cleaned texts: {text_duplicate_count}")

def analyze_text_label_consistency(data, target_col):
    counts = data.groupby(['text_clean', target_col]).size().unstack(fill_value=0)
    totals = counts.sum(axis=1)
    majority_counts = counts.max(axis=1)
    label_counts_per_text = (counts > 0).sum(axis=1)

    ambiguous_texts = (label_counts_per_text > 1).sum()
    consistent_texts = (label_counts_per_text == 1).sum()
    weighted_majority_ratio = majority_counts.sum() / totals.sum()
    mean_majority_ratio = (majority_counts / totals).mean()

    print(f"\n--- Text/Label Consistency: {target_col} ---")
    print(f"Unique texts            : {len(counts)}")
    print(f"Consistent texts        : {consistent_texts}")
    print(f"Ambiguous texts         : {ambiguous_texts}")
    print(f"Mean majority agreement : {mean_majority_ratio:.4f}")
    print(f"Weighted agreement      : {weighted_majority_ratio:.4f}")

# Identify the 4 Target Columns
target_columns_map = {
    'TARGET: Kategori Risiko': [c for c in df.columns if 'Kategori Risiko' in c][0],
    'TARGET: Sentimen': [c for c in df.columns if 'Sentimen' in c][0],
    'TARGET: Impact (1-5)': [c for c in df.columns if 'Impact' in c][0],
    'TARGET: Likelihood (1-5)': [c for c in df.columns if 'Likelihood' in c][0]
}

for target_name in ['TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)']:
    analyze_text_label_consistency(df, target_columns_map[target_name])

for target_name, target_col in target_columns_map.items():
    conflict_count = df.groupby('text_clean')[target_col].nunique().gt(1).sum()
    print(f"{target_name} | Same text with multiple labels: {conflict_count}")

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

        encoding = self.tokenizer(
            text,
            add_special_tokens=True,
            max_length=self.max_len,
            padding='max_length',
            truncation=True,
            return_attention_mask=True,
            return_tensors='pt'
        )

        return {
            'input_ids': encoding['input_ids'].squeeze(0),
            'attention_mask': encoding['attention_mask'].squeeze(0),
            'label': torch.tensor(label, dtype=torch.long)
        }

def train_and_save_final_model(target_label, texts, labels, unique_classes, tokenizer, device):
    num_classes = len(unique_classes)
    save_root = os.path.join(current_dir, 'saved_models')
    save_dir = os.path.join(save_root, FINAL_MODEL_DIRS[target_label])
    os.makedirs(save_dir, exist_ok=True)

    print(f"\n--- FINAL TRAINING: {target_label} ---")
    print(f"Training samples : {len(texts)}")
    print(f"Classes          : {unique_classes}")

    dataset = AuditDataset(texts, labels, tokenizer, MAX_LEN)
    loader = DataLoader(dataset, batch_size=BATCH_SIZE, shuffle=True)

    set_seed(42)
    model = build_indobert_classifier(num_classes, unique_classes).to(device)
    optimizer = AdamW(model.parameters(), lr=LEARNING_RATE)

    for epoch in range(NUM_EPOCHS):
        model.train()
        total_loss = 0.0

        for batch in loader:
            input_ids = batch['input_ids'].to(device)
            attention_mask = batch['attention_mask'].to(device)
            labels_b = batch['label'].to(device)

            optimizer.zero_grad()
            outputs = model(input_ids=input_ids, attention_mask=attention_mask, labels=labels_b)
            loss = outputs.loss
            loss.backward()
            optimizer.step()

            total_loss += loss.item()

        avg_loss = total_loss / len(loader)
        print(f" Epoch {epoch+1}/{NUM_EPOCHS} | Train Loss: {avg_loss:.4f}", flush=True)

    model.save_pretrained(save_dir)
    tokenizer.save_pretrained(save_dir)

    print(f"Saved final model to: {save_dir}")

    del model

    if torch.cuda.is_available():
        torch.cuda.empty_cache()

    return save_dir

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
    if target_label not in ACTIVE_TARGETS:
        print(f"Skipping {target_label} (already validated).")
        continue
    
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
    
    groups = sub_df['text_clean'].values
    
    if RUN_FINAL_TRAINING:
        train_and_save_final_model(target_label, texts, labels, unique_classes, tokenizer, device)
        continue
    
    # Check class frequency for safe stratification
    # class_counts = pd.Series(labels).value_counts()
    # min_samples = class_counts.min()
    
    sgkf = StratifiedGroupKFold(n_splits=5, shuffle=True, random_state=42)

    train_idx, val_idx = next(sgkf.split(texts, labels, groups=groups))
    
    print("\n--- 5-Fold Grouped TF-IDF Robustness ---")
    
    group_cv_results = []

    for fold_no, (cv_train_idx, cv_val_idx) in enumerate(sgkf.split(texts, labels, groups=groups), start=1):
        X_cv_train = texts[cv_train_idx]
        X_cv_val = texts[cv_val_idx]
        y_cv_train = labels[cv_train_idx]
        y_cv_val = labels[cv_val_idx]

        tfidf_cv = TfidfVectorizer(ngram_range=(1, 2), max_features=3000, sublinear_tf=True)
        X_cv_train_tfidf = tfidf_cv.fit_transform(X_cv_train)
        X_cv_val_tfidf = tfidf_cv.transform(X_cv_val)

        lr_cv = LogisticRegression(max_iter=1000, random_state=42)
        lr_cv.fit(X_cv_train_tfidf, y_cv_train)
        lr_pred = lr_cv.predict(X_cv_val_tfidf)

        svc_cv = LinearSVC(C=1.0, random_state=42)
        svc_cv.fit(X_cv_train_tfidf, y_cv_train)
        svc_pred = svc_cv.predict(X_cv_val_tfidf)

        row = {
            'fold': fold_no,
            'lr_acc': accuracy_score(y_cv_val, lr_pred),
            'lr_f1': f1_score(y_cv_val, lr_pred, average='weighted', zero_division=0),
            'svc_acc': accuracy_score(y_cv_val, svc_pred),
            'svc_f1': f1_score(y_cv_val, svc_pred, average='weighted', zero_division=0)
        }

        if target_label in ['TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)']:
            ridge_cv = Ridge(alpha=1.0)
            ridge_cv.fit(X_cv_train_tfidf, y_cv_train.astype(float))
            ridge_raw = ridge_cv.predict(X_cv_val_tfidf)
            ridge_pred = np.clip(np.rint(ridge_raw), 0, num_classes - 1).astype(int)

            row['ridge_acc'] = accuracy_score(y_cv_val, ridge_pred)
            row['ridge_f1'] = f1_score(y_cv_val, ridge_pred, average='weighted', zero_division=0)

        group_cv_results.append(row)

        print(f"Fold {fold_no} | LR F1={row['lr_f1']:.4f} | SVC F1={row['svc_f1']:.4f}")

    X_train = texts[train_idx]
    X_val = texts[val_idx]
    y_train = labels[train_idx]
    y_val = labels[val_idx]

    train_groups = set(groups[train_idx])
    val_groups = set(groups[val_idx])

    group_overlap = train_groups.intersection(val_groups)

    print(f"Train samples: {len(X_train)} | Validation samples: {len(X_val)}")
    print(f"Train unique texts: {len(train_groups)} | Validation unique texts: {len(val_groups)}")
    print(f"Text overlap between train/validation: {len(group_overlap)}")
    
    group_cv_df = pd.DataFrame(group_cv_results)

    print("\n--- Grouped Robustness Summary ---")
    print(f"Logistic Accuracy : {group_cv_df['lr_acc'].mean():.4f} ± {group_cv_df['lr_acc'].std():.4f}")
    print(f"Logistic F1       : {group_cv_df['lr_f1'].mean():.4f} ± {group_cv_df['lr_f1'].std():.4f}")
    print(f"LinearSVC Accuracy: {group_cv_df['svc_acc'].mean():.4f} ± {group_cv_df['svc_acc'].std():.4f}")
    print(f"LinearSVC F1      : {group_cv_df['svc_f1'].mean():.4f} ± {group_cv_df['svc_f1'].std():.4f}")
    
    if target_label in ['TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)']:
        print(f"Ordinal Ridge Acc : {group_cv_df['ridge_acc'].mean():.4f} ± {group_cv_df['ridge_acc'].std():.4f}")
        print(f"Ordinal Ridge F1  : {group_cv_df['ridge_f1'].mean():.4f} ± {group_cv_df['ridge_f1'].std():.4f}")
    
    if RUN_INDOBERT_GROUP_CV and target_label in ROBUST_TARGETS:
        print("\n--- 5-Fold Grouped IndoBERT Robustness ---")

        indobert_cv_results = []

        for fold_no, (cv_train_idx, cv_val_idx) in enumerate(sgkf.split(texts, labels, groups=groups), start=1):
            set_seed(42 + fold_no)

            X_cv_train = texts[cv_train_idx]
            X_cv_val = texts[cv_val_idx]
            y_cv_train = labels[cv_train_idx]
            y_cv_val = labels[cv_val_idx]

            cv_train_dataset = AuditDataset(X_cv_train, y_cv_train, tokenizer, MAX_LEN)
            cv_val_dataset = AuditDataset(X_cv_val, y_cv_val, tokenizer, MAX_LEN)

            cv_train_loader = DataLoader(cv_train_dataset, batch_size=BATCH_SIZE, shuffle=True)
            cv_val_loader = DataLoader(cv_val_dataset, batch_size=BATCH_SIZE)

            cv_model = build_indobert_classifier(num_classes, unique_classes).to(device)

            cv_optimizer = AdamW(cv_model.parameters(), lr=LEARNING_RATE)

            for epoch in range(NUM_EPOCHS):
                cv_model.train()

                for batch in cv_train_loader:
                    input_ids = batch['input_ids'].to(device)
                    attention_mask = batch['attention_mask'].to(device)
                    labels_b = batch['label'].to(device)

                    cv_optimizer.zero_grad()
                    outputs = cv_model(input_ids=input_ids, attention_mask=attention_mask, labels=labels_b)
                    loss = outputs.loss
                    loss.backward()
                    cv_optimizer.step()

            cv_model.eval()
            cv_preds = []
            cv_targets = []

            with torch.no_grad():
                for batch in cv_val_loader:
                    input_ids = batch['input_ids'].to(device)
                    attention_mask = batch['attention_mask'].to(device)

                    outputs = cv_model(input_ids=input_ids, attention_mask=attention_mask)
                    preds = torch.argmax(outputs.logits, dim=1).cpu().numpy()

                    cv_preds.extend(preds)
                    cv_targets.extend(batch['label'].numpy())

            cv_acc = accuracy_score(cv_targets, cv_preds)
            cv_f1 = f1_score(cv_targets, cv_preds, average='weighted', zero_division=0)
            cv_macro = f1_score(cv_targets, cv_preds, average='macro', zero_division=0)

            indobert_cv_results.append({
                'fold': fold_no,
                'accuracy': cv_acc,
                'weighted_f1': cv_f1,
                'macro_f1': cv_macro
            })

            print(f"Fold {fold_no} | Accuracy={cv_acc:.4f} | Weighted F1={cv_f1:.4f} | Macro F1={cv_macro:.4f}")

            del cv_model

            if torch.cuda.is_available():
                torch.cuda.empty_cache()

        indobert_cv_df = pd.DataFrame(indobert_cv_results)

        print("\n--- IndoBERT Grouped Robustness Summary ---")
        print(f"Accuracy    : {indobert_cv_df['accuracy'].mean():.4f} ± {indobert_cv_df['accuracy'].std():.4f}")
        print(f"Weighted F1 : {indobert_cv_df['weighted_f1'].mean():.4f} ± {indobert_cv_df['weighted_f1'].std():.4f}")
        print(f"Macro F1    : {indobert_cv_df['macro_f1'].mean():.4f} ± {indobert_cv_df['macro_f1'].std():.4f}")
        
        overall_results[target_label] = {
            'num_classes': num_classes,
            'indobert_acc': indobert_cv_df['accuracy'].mean(),
            'indobert_acc_std': indobert_cv_df['accuracy'].std(),
            'indobert_f1': indobert_cv_df['weighted_f1'].mean(),
            'indobert_f1_std': indobert_cv_df['weighted_f1'].std(),
            'val_macro_f1': indobert_cv_df['macro_f1'].mean(),
            'val_macro_f1_std': indobert_cv_df['macro_f1'].std(),
            'baseline_acc': group_cv_df['svc_acc'].mean(),
            'baseline_f1': group_cv_df['svc_f1'].mean()
        }
    
    if RUN_INDOBERT_GROUP_CV and target_label in ROBUST_TARGETS:
        continue
    
    # ----------------------------------------------------------
    # Balanced class weights from TRAINING data only
    # ----------------------------------------------------------
    # train_class_counts = np.bincount(y_train, minlength=num_classes)

    # class_weights = (len(y_train)/(num_classes* np.maximum(train_class_counts, 1)))

    # class_weights_tensor = torch.tensor(class_weights, dtype=torch.float32, device=device)

    # criterion = torch.nn.CrossEntropyLoss(weight=class_weights_tensor)

    # print("Training class counts :", train_class_counts.tolist())
    # print("Class weights         :", [round(float(w), 4) for w in class_weights])
    
    # PyTorch DataLoaders
    train_dataset = AuditDataset(X_train, y_train, tokenizer, MAX_LEN)
    val_dataset = AuditDataset(X_val, y_val, tokenizer, MAX_LEN)
    
    train_loader = DataLoader(train_dataset, batch_size=BATCH_SIZE, shuffle=True)
    val_loader = DataLoader(val_dataset, batch_size=BATCH_SIZE)
    
    # Load IndoBERT Model
    model = build_indobert_classifier(num_classes, unique_classes).to(device)
    
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
    val_macro_f1 = f1_score(all_targets, all_preds, average='macro', zero_division=0)
    
    # Baseline Comparison: TF-IDF + Logistic Regression
    tfidf_vec = TfidfVectorizer(ngram_range=(1, 2), max_features=3000, sublinear_tf=True)
    X_tr_tfidf = tfidf_vec.fit_transform(X_train)
    X_val_tfidf = tfidf_vec.transform(X_val)
    clf_baseline = LogisticRegression(max_iter=1000, random_state=42)
    clf_baseline.fit(X_tr_tfidf, y_train)
    baseline_preds = clf_baseline.predict(X_val_tfidf)
    baseline_acc = accuracy_score(y_val, baseline_preds)
    baseline_f1 = f1_score(y_val, baseline_preds, average='weighted', zero_division=0)
    baseline_macro_f1 = f1_score(y_val, baseline_preds, average='macro', zero_division=0)
    
    # TF-IDF + LinearSVC Challenger
    svc_model = LinearSVC(C=1.0, random_state=42)

    svc_model.fit(X_tr_tfidf, y_train)
    svc_preds = svc_model.predict(X_val_tfidf)

    svc_acc = accuracy_score(y_val, svc_preds)
    svc_f1 = f1_score(y_val, svc_preds, average='weighted', zero_division=0)
    svc_macro_f1 = f1_score(y_val, svc_preds, average='macro', zero_division=0)
    
    # TF-IDF Ordinal Ridge Regression
    # Classes 0..4 are treated as ordered values
    if target_label in ['TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)']:
        ordinal_model = Ridge(alpha=1.0)
        ordinal_model.fit(X_tr_tfidf, y_train.astype(float))

        ordinal_raw = ordinal_model.predict(X_val_tfidf)
        ordinal_preds = np.clip(np.rint(ordinal_raw), 0, num_classes - 1).astype(int)

        ordinal_acc = accuracy_score(y_val, ordinal_preds)
        ordinal_f1 = f1_score(y_val, ordinal_preds, average='weighted', zero_division=0)
        ordinal_macro_f1 = f1_score(y_val, ordinal_preds, average='macro', zero_division=0)
        ordinal_mae = mean_absolute_error(y_val, ordinal_preds)
    
    report_str = classification_report(all_targets, all_preds, labels=list(range(num_classes)), target_names=unique_classes, zero_division=0)
    
    overall_results[target_label] = {
        'num_classes': num_classes,
        'indobert_acc': val_acc,
        'indobert_f1': val_f1,
        'val_macro_f1': val_macro_f1,
        'baseline_acc': baseline_acc,
        'baseline_f1': baseline_f1,
        'baseline_macro_f1': baseline_macro_f1,
        'confusion_matrix': confusion_matrix(all_targets, all_preds, labels=list(range(num_classes))),
        'classification_report': report_str
    }
    
    print(f"\n--- Validation Results ({target_label}) ---")
    print(f"IndoBERT Accuracy : {val_acc * 100:.2f} % (Weighted F1: {val_f1:.4f})")
    print(f"TF-IDF Baseline   : {baseline_acc * 100:.2f} % (Weighted F1: {baseline_f1:.4f})")
    print(f"Macro F1: {val_macro_f1:.4f} | Baseline Macro F1: {baseline_macro_f1:.4f}")
    print(f"TF-IDF LinearSVC :  {svc_acc*100:.2f} % | Weighted F1: {svc_f1:.4f} | Macro F1: {svc_macro_f1:.4f}")
    if target_label in ['TARGET: Impact (1-5)', 'TARGET: Likelihood (1-5)']:
        print(f"TF-IDF Ordinal Ridge: {ordinal_acc*100:.2f} % | Weighted F1: {ordinal_f1:.4f} | Macro F1: {ordinal_macro_f1:.4f} | MAE: {ordinal_mae:.4f}")
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
if not RUN_FINAL_TRAINING:
    print("\n======================================================================")
    print(" FINAL GROUPED CV SUMMARY REPORT FOR ACTIVE TARGETS ")
    print("======================================================================")
    print(f"{'Target Column':<30} | {'Acc Mean±Std':<18} | {'Weighted F1 Mean±Std':<22} | {'SVC F1 Mean':<12}")
    print("-" * 92)

    for t_label, res in overall_results.items():
        print(f"{t_label:<30} | {res['indobert_acc']:.4f} ± {res['indobert_acc_std']:.4f} | {res['indobert_f1']:.4f} ± {res['indobert_f1_std']:.4f}     | {res['baseline_f1']:.4f}")

    print("\nIndoBERT training and benchmark completed successfully!")


# ==========================================
# 4. VERIFY SAVED FINAL MODELS
# ==========================================
if RUN_FINAL_TRAINING and VERIFY_SAVED_MODELS:
    print("\n======================================================================")
    print(" VERIFYING SAVED FINAL MODELS ")
    print("======================================================================")

    for target_label, folder_name in FINAL_MODEL_DIRS.items():
        save_dir = os.path.join(current_dir, 'saved_models', folder_name)

        print(f"\nChecking: {target_label}")
        print(f"Path    : {save_dir}")

        test_model = AutoModelForSequenceClassification.from_pretrained(save_dir).to(device)
        test_tokenizer = AutoTokenizer.from_pretrained(save_dir)

        test_text = df['text_clean'].iloc[0]
        encoded = test_tokenizer(test_text, max_length=MAX_LEN, padding='max_length', truncation=True, return_tensors='pt')
        input_ids = encoded['input_ids'].to(device)
        attention_mask = encoded['attention_mask'].to(device)

        test_model.eval()

        with torch.no_grad():
            outputs = test_model(input_ids=input_ids, attention_mask=attention_mask)
            predicted_id = torch.argmax(outputs.logits, dim=1).item()

        predicted_label = test_model.config.id2label[predicted_id]

        print("Load test       : OK")
        print(f"Predicted ID    : {predicted_id}")
        print(f"Predicted label : {predicted_label}")

        del test_model

        if torch.cuda.is_available():
            torch.cuda.empty_cache()

    print("\nAll final IndoBERT models trained, saved, and verified successfully!")