#!/bin/bash
# =============================================================================
# AuditSphere Frontend Production Deploy Script (Alias to deploy-prod.sh)
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bash "$SCRIPT_DIR/deploy-prod.sh" "$@"
