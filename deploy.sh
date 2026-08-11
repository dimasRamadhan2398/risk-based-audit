#!/bin/bash
# =============================================================================
# AuditSphere Master Deploy Script
# Deploys both backend (microservices + Kong) and frontend (Nuxt.js) to VPS
# =============================================================================

VPS_IP="202.10.34.166"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║        AuditSphere Full-Stack VPS Deployment             ║"
echo "║              Target: $VPS_IP                    ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ── Deploy Backend ────────────────────────────────────────────────────────────
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " STEP 1 OF 2 — Backend (microservices + Kong)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/backend/deploy-backend.sh"

# ── Wait briefly so backend containers stabilise before frontend goes up ──────
echo ""
echo "⏳ Waiting 10s for backend containers to stabilise..."
sleep 10

# ── Deploy Frontend ───────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " STEP 2 OF 2 — Frontend (Nuxt.js)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/frontend/deploy.sh"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║            🎉  Deployment Successful!  🎉                ║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║                                                          ║"
echo "║  Frontend (AuditSphere UI)                               ║"
echo "║    http://$VPS_IP:3000                          ║"
echo "║                                                          ║"
echo "║  Backend API Gateway (Kong)                              ║"
echo "║    http://$VPS_IP:8080/api/v1/...               ║"
echo "║                                                          ║"
echo "║  Kong Admin                                              ║"
echo "║    http://$VPS_IP:8009                          ║"
echo "║                                                          ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
