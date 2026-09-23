#!/bin/bash
# =============================================================================
# AuditSphere Master Development Deploy Script (dev.auditsphere.app & api-dev.auditsphere.app)
# Deploys isolated development backend (Port 8081) and development frontend (Port 3001)
# =============================================================================

VPS_IP="202.10.34.166"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║        AuditSphere Full-Stack DEV VPS Deployment         ║"
echo "║        Target UI : dev.auditsphere.app (Port 3001)      ║"
echo "║        Target API: api-dev.auditsphere.app (Port 8081)  ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ── Step 1: Deploy Backend (DEV) ─────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " STEP 1 OF 2 — Development Backend (Port 8081)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/backend/deploy-dev.sh"

# ── Wait briefly for dev backend containers ─────────────────────────────────
echo ""
echo "⏳ Waiting 10s for dev backend containers to stabilise..."
sleep 10

# ── Step 2: Deploy Frontend (DEV) ────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " STEP 2 OF 2 — Development Frontend (Port 3001)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/frontend/deploy-dev.sh"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║            🎉  DEV Deployment Successful!   🎉            ║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║                                                          ║"
echo "║  Development UI (dev.auditsphere.app)                     ║"
echo "║    http://$VPS_IP:3001                          ║"
echo "║                                                          ║"
echo "║  Development API Gateway (api-dev.auditsphere.app)       ║"
echo "║    http://$VPS_IP:8081/api/v1/...               ║"
echo "║                                                          ║"
echo "║  Kong DEV Admin                                          ║"
echo "║    http://$VPS_IP:8010                          ║"
echo "║                                                          ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
