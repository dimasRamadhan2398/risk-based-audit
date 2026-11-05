#!/usr/bin/env bash
# Start the training stack (https://training.auditsphere.app).
# Only touches compose project rb_audit_training; the one prod-side action is
# attaching rb_audit_clamav to the training network (no restart).
TRAINING_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$TRAINING_DIR/lib.sh"

check_memory "$(pending_training_mb)"
ensure_network
mkdir -p "$STORAGE/audit-uploads" "$STORAGE/master-uploads" "$STORAGE/python-ai-models"
attach_clamav
tcompose up -d --no-build --remove-orphans
wait_healthy rb_audit_kong_training 180
log "training stack up"
tcompose ps --format 'table {{.Name}}\t{{.Status}}'
