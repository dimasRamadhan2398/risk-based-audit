#!/bin/bash
# =============================================================================
# AuditSphere Master Deploy Script (Alias to deploy-prod.sh)
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bash "$SCRIPT_DIR/deploy-prod.sh" "$@"
