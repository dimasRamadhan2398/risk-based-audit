#!/usr/bin/env bash
# Stop the training stack. Containers, volumes and data are kept; start
# again with start-training.sh. Prod is not touched apart from detaching
# rb_audit_clamav from the training network (no restart).
TRAINING_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$TRAINING_DIR/lib.sh"

tcompose stop
detach_clamav
log "training stack stopped"
