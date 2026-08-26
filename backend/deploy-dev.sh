#!/bin/bash
# =============================================================================
# AuditSphere Backend Development Deploy Script (api-dev.auditsphere.app)
# Packages backend source, uploads to VPS, and starts via docker-compose.dev.yml
# =============================================================================

VPS_IP="202.10.34.166"
VPS_USER="root"
TARGET_DIR="/app/rbia-backend-dev"
ARCHIVE="backend-dev.tar.gz"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

export COPYFILE_DISABLE=1

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║   AuditSphere Backend (DEV) → VPS ($VPS_IP)             ║"
echo "║   Target API: api-dev.auditsphere.app (Kong Port 8081)   ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ── Step 1: Pre-compile Go Services Locally ──────────────────────────────────
echo "⚡ [1/5] Pre-compiling Go microservices locally (target: linux/amd64)..."

build_go_service() {
  local svc_dir="$1"
  local cmd_dir="$2"
  local bin_name="$3"
  echo "  → Compiling $bin_name ($svc_dir)..."
  (cd "$svc_dir" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "./$bin_name" "./$cmd_dir")
}

build_go_service "auth-service" "cmd/auth" "auth"
build_go_service "audit-service" "cmd" "audit"
build_go_service "master-service" "cmd" "master"
build_go_service "risk-service" "cmd" "risk"
build_go_service "analytics-service" "cmd" "analytics"

echo "   All Go microservices compiled successfully!"

# ── Step 2: Package ──────────────────────────────────────────────────────────
echo "📦 [2/5] Packaging backend source and pre-built binaries for dev..."
tar \
  --exclude='.git' \
  --exclude='*/vendor' \
  --exclude='*.tar.gz' \
  --exclude='*.exe' \
  --exclude='*/audit-service/audit-service-test' \
  --exclude='*/audit-service/tmp_*' \
  --exclude='*/risk-service/tmp_*' \
  --exclude='__pycache__' \
  --exclude='.pytest_cache' \
  --exclude='.venv' \
  --exclude='.DS_Store' \
  -czf "$ARCHIVE" .

echo "   Archive size: $(du -sh $ARCHIVE | cut -f1)"

# ── Step 3: Prepare remote directory ─────────────────────────────────────────
echo "🌐 [3/5] Preparing remote directory $TARGET_DIR on VPS..."
ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$VPS_USER@$VPS_IP" "mkdir -p $TARGET_DIR"

# ── Step 4: Upload ────────────────────────────────────────────────────────────
echo "🚀 [4/5] Uploading archive to $VPS_USER@$VPS_IP:$TARGET_DIR ..."
scp -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$ARCHIVE" "$VPS_USER@$VPS_IP:$TARGET_DIR/"

# ── Step 5: Extract & Start on VPS ───────────────────────────────────────────
echo "🔧 [5/5] Extracting and starting backend (DEV) on VPS..."
ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$VPS_USER@$VPS_IP" bash <<EOF
  set -e
  cd $TARGET_DIR

  echo "  → Extracting archive..."
  tar -xzf $ARCHIVE
  rm -f $ARCHIVE

  echo "  → Creating ai_model_training directory if missing..."
  mkdir -p $TARGET_DIR/ai_model_training

  echo "  → Starting development containers with pre-built binaries (Kong Port 8081)..."
  docker compose -f docker-compose.dev.yml up -d --build

  echo "  → Checking dev container status..."
  sleep 5
  docker compose -f docker-compose.dev.yml ps
EOF

# ── Cleanup local archive ─────────────────────────────────────────────────────
rm -f "$ARCHIVE"

echo ""
echo "✅ Development Backend deployment complete!"
echo ""
echo "   Kong DEV API Gateway : http://$VPS_IP:8081 (api-dev.auditsphere.app)"
echo "   Kong DEV Admin       : http://$VPS_IP:8010"
echo "   Auth Service DEV     : http://$VPS_IP:8101"
echo "   Audit Service DEV    : http://$VPS_IP:8102"
echo "   Master Service DEV   : http://$VPS_IP:8103"
echo "   Risk Service DEV     : http://$VPS_IP:8104"
echo "   Analytics DEV        : http://$VPS_IP:8184"
echo "   Python AI DEV        : http://$VPS_IP:8100"
echo "   PostgreSQL DEV       : http://$VPS_IP:5433"
echo ""
