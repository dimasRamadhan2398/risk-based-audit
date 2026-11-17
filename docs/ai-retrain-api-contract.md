# Retrain API Contract (python-ai  <->  analytics-service (Go)  <->  frontend)

Model keys: `entity` (Entity Prediction / risk-scoring page), `anomaly` (Anomaly Detection + Data Source rules),
`nlp` (NLP Analysis), `kpi_auditee` (KPI Forecast - Auditee), `kpi_auditor` (KPI Forecast - Auditor / Strategic Audit Plan).
Alias `kpi` = both kpi_auditee + kpi_auditor.

## python-ai endpoints

### POST /retrain/model/{model}
Internal-token protected exactly like `/retrain/auto` (header `X-Internal-Token`, same `internal_call_allowed`).
Body (optional JSON):
```json
{ "trigger": "manual", "force": false, "auditor_items": [..], "auditee_items": [..] }
```
- Non-blocking. Ingests auditor/auditee items into the pools first (kpi only), then enqueues.
- `force=false` and no new data -> status `up_to_date` (nothing queued).
Response 200:
```json
{
  "status": "queued" | "already_queued" | "followup_queued" | "up_to_date" | "running",
  "model": "anomaly",
  "models_enqueued": ["anomaly"],
  "job_id": "job-...",
  "queue_position": 1,
  "message": "Model Anomaly Detection masuk antrian pelatihan (posisi #1).",
  "models": { ...same as GET /retrain/status .models... },
  "queue": [ ... ]
}
```
400 for unknown model: `{"status":"error","message":"Unknown model ..."}`.

### GET /retrain/status  (existing, extended – legacy fields unchanged)
Legacy fields kept: status, is_running, trigger, started_at, finished_at, duration_seconds, trained_auditor_hash,
trained_auditee_pool_hash, trained_auditee_custom_hash, trained_auditor_count, trained_auditee_count, auditor_retrained,
auditee_retrained, model_updated_at, message, last_error, current_auditee_pool_hash, current_auditee_pool_count,
has_untrained_auditee_pool_data, has_untrained_auditor_data, has_untrained_auditee_custom_data.
(`is_running` = true when ANY model is training.)

New fields:
```json
"models": {
  "anomaly": {
    "key": "anomaly",
    "label": "Anomaly Detection",
    "state": "never_trained" | "trained" | "untrained_data" | "queued" | "training" | "failed",
    "has_new_data": true,
    "new_rows": 120,                 // current_rows - trained_rows (>=0), may be 0 even if has_new_data (edited rows)
    "current_rows": 5120,
    "trained_rows": 5000,
    "trained_at": "2026-11-17T09:00:00+00:00" | null,
    "duration_seconds": 12.4 | null,
    "samples": 5000,                  // rows actually used for training
    "tier": "T1" | "T2" | "T3" | null,
    "metrics": { "primary_name": "pr_auc", "primary": 0.91, "previous": 0.89, ... },
    "last_result": "promoted" | "rolled_back" | "failed" | null,
    "message": "Human readable Indonesian status message",
    "last_error": null,
    "progress": { "phase": "training", "batch": 3, "total_batches": 10, "percent": 45 } | null,
    "queue_position": 2 | null,       // 1-based position among waiting jobs, null if not queued
    "job_id": "job-..." | null,
    "trigger": "manual" | "auto" | ... | null,
    "waiting_reason": "Menunggu resource (RAM bebas 900 MB < 1500 MB)" | null,
    "peak_ram_mb": 410.5 | null,
    "version": 3
  },
  "entity": {...}, "nlp": {...}, "kpi_auditee": {...}, "kpi_auditor": {...}
},
"queue": [ { "job_id": "...", "model": "nlp", "trigger": "auto", "priority": "auto"|"manual", "state": "queued"|"running", "enqueued_at": "...", "position": 1 } ],
"resource": { "vcpu": 2, "ram_total_mb": 8192, "ram_available_mb": 5000, "cpu_percent": 12.0, "max_concurrent": 1,
              "train_threads": 1, "batch_rows": 5000, "auto_interval": 60, "guard_ok": true, "guard_reason": null },
"scheduler": { "running": true, "auto_enabled": true, "last_detect_at": "..." }
```

### POST /retrain/auto  (existing, kept compatible)
Now always non-blocking: ingests items (if any), enqueues every model with new data (scope `all`), or kpi models when
`scope=kpi`, or a single model when `scope` is a model key. `force=true` enqueues even without new data.
Response: `{ "status": "scheduled" | "up_to_date" | "already_running", "is_running": bool, "message": str,
"models_enqueued": [...], "models": {...}, "queue": [...] }`.

### Existing data endpoints – behaviour change (no more synchronous training)
- `POST /predict/entity-prediction/sync` (+ `/predict/risk-score/sync-profile`): upserts pool, enqueues `entity` (manual)
  when `trigger_retrain`, returns batch predictions + `"retrain": {status, job_id, queue_position, model}`.
- `POST /retrain/anomaly-rules` (+ `/predict/anomaly/train-rules`): writes rule rows, enqueues `anomaly` (manual), returns
  previous fields + `"retrain": {...}`; `anomaly_model_retrained` is now false (queued) – message says it's queued.
- `POST /predict/text-analysis/sync-sources`: pulls sources, enqueues `nlp` (manual), returns + `"retrain": {...}`.
- `POST /taxonomy/nlp`: enqueues `nlp`.

## analytics-service (Go) – frontend-facing
- NEW `POST /api/analytics/retrain/model/:model` -> python `POST /retrain/model/{model}` (body passthrough, X-Internal-Token).
  Must be blocked for viewer role exactly like `/retrain/auto`. After a successful enqueue, do NOT invalidate cache yet
  (training is async). Validate model in {entity, anomaly, nlp, kpi, kpi_auditee, kpi_auditor} -> 400 otherwise.
- `GET /api/analytics/retrain/status` passthrough (already exists) – must NOT be cached for long (≤2s or uncached) since UI polls.
  When the status shows a model whose `trained_at`/`version` changed since the last seen status, invalidate the batch cache
  for that model (anomaly->anomaly batch, entity->risk-score batch, nlp->text-analysis batch, kpi_*->performance-trend batch).
- `POST /api/analytics/retrain/auto`: the global 6h throttle stays only for non-force auto calls; manual per-model has no
  6h throttle (python dedups itself).
