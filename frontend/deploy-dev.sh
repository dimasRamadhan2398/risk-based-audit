#!/bin/bash
# =============================================================================
# AuditSphere Frontend Development Deploy Script (dev.auditsphere.app)
# Builds Nuxt 4 locally, uploads .output archive, and starts rbia-frontend-dev.
# =============================================================================

VPS_IP="202.10.34.166"
VPS_USER="root"
TARGET_DIR="/app/rbia-frontend-dev"
ARCHIVE="frontend-dev.tar.gz"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

export COPYFILE_DISABLE=1

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║   AuditSphere Frontend (DEV) → VPS                         ║"
echo "║   Target Domain: dev.auditsphere.app (Port 3001)           ║"
echo "║   Target API   : api-dev.auditsphere.app (Port 8081)       ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# ── Step 1: Build locally ───────────────────────────────────────────────────
echo "⚡ [1/4] Building Nuxt 4 application locally..."
npm run build

# ── Step 2: Package .output and configuration ────────────────────────────────
echo "📦 [2/4] Packaging pre-built .output and Docker files..."
tar -czf "$ARCHIVE" .output Dockerfile docker-compose.dev.yml package.json

echo "   Archive size: $(du -sh $ARCHIVE | cut -f1)"

# ── Step 3: Prepare directory & Upload to VPS ───────────────────────────────
echo "🌐 [3/4] Uploading archive to $VPS_USER@$VPS_IP:$TARGET_DIR ..."
ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$VPS_USER@$VPS_IP" "mkdir -p $TARGET_DIR"
scp -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$ARCHIVE" "$VPS_USER@$VPS_IP:$TARGET_DIR/"

# ── Step 4: Extract & Run Container on VPS ──────────────────────────────────
echo "🚀 [4/4] Starting rbia-frontend-dev container on VPS..."
ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$VPS_USER@$VPS_IP" bash <<EOF
  set -e
  cd $TARGET_DIR

  echo "  → Extracting pre-built bundle..."
  tar -xzf $ARCHIVE
  rm -f $ARCHIVE

  echo "  → Creating development environment file (pointing to Kong Port 8081 / api-dev)..."
  cat <<ENVEOF > .env
API_BASE_URL=https://api-dev.auditsphere.app/api/v1
ANALYTICS_API_BASE_URL=https://api-dev.auditsphere.app/api/analytics
NUXT_PUBLIC_AUTH_SERVICE_BASE_URL=https://api-dev.auditsphere.app/api/v1
NUXT_PUBLIC_AUDIT_SERVICE_BASE_URL=https://api-dev.auditsphere.app/api/v1
NUXT_PUBLIC_RISK_SERVICE_BASE_URL=https://api-dev.auditsphere.app/api/v1
NUXT_PUBLIC_MASTER_SERVICE_BASE_URL=https://api-dev.auditsphere.app/api/v1
ENVEOF

  echo "  → Launching rbia-frontend-dev container..."
  docker rm -f rbia-frontend-dev || true
  docker compose -f docker-compose.dev.yml up -d --build --force-recreate

  echo "  → Checking status..."
  docker compose -f docker-compose.dev.yml ps
EOF

# ── Cleanup local archive ─────────────────────────────────────────────────────
rm -f "$ARCHIVE"

echo ""
echo "🎉 Development Frontend deployment completed successfully!"
echo "   Container: rbia-frontend-dev (Port 3001)"
echo "   Target UI Domain: dev.auditsphere.app"
echo "   Target API Domain: api-dev.auditsphere.app (Port 8081)"
echo ""
