#!/bin/bash
# =============================================================================
# AuditSphere Master Production Deploy Script (auditsphere.app)
# Deploys backend (microservices + Kong) and production frontend
# =============================================================================

VPS_IP="202.10.34.166"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║        AuditSphere Full-Stack PROD VPS Deployment        ║"
echo "║        Target: $VPS_IP (auditsphere.app)                ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ── Deploy Backend ────────────────────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " STEP 1 OF 2 — Backend (microservices + Kong)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/backend/deploy-backend.sh"

# ── Wait briefly so backend containers stabilise ──────────────────────────────
echo ""
echo "⏳ Waiting 10s for backend containers to stabilise..."
sleep 10

# ── Deploy Production Frontend ────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " STEP 2 OF 2 — Production Frontend (auditsphere.app)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/frontend/deploy-prod.sh"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║            🎉  PROD Deployment Successful!  🎉            ║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║                                                          ║"
echo "║  Frontend (auditsphere.app)                              ║"
echo "║    http://$VPS_IP:3000                          ║"
echo "║                                                          ║"
echo "║  Backend API Gateway (api.auditsphere.app)               ║"
echo "║    http://$VPS_IP:8080/api/v1/...               ║"
echo "║                                                          ║"
echo "║  Kong Admin                                              ║"
echo "║    http://$VPS_IP:8009                          ║"
echo "║                                                          ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
