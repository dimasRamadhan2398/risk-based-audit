#!/bin/bash
# =============================================================================
# AuditSphere Backend Deploy Script
# Packages backend source, uploads to VPS, and starts via docker-compose.prod.yml
# =============================================================================

VPS_IP="202.10.34.166"
VPS_USER="root"
TARGET_DIR="/app/rbia-backend"
ARCHIVE="backend.tar.gz"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

export COPYFILE_DISABLE=1

echo ""
echo "╔══════════════════════════════════════════════╗"
echo "║   AuditSphere Backend → VPS ($VPS_IP)   ║"
echo "╚══════════════════════════════════════════════╝"
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
echo "📦 [2/5] Packaging backend source and pre-built binaries..."
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
echo "🌐 [3/5] Preparing remote directory on VPS..."
ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$VPS_USER@$VPS_IP" "mkdir -p $TARGET_DIR"

# ── Step 4: Upload ────────────────────────────────────────────────────────────
echo "🚀 [4/5] Uploading archive to $VPS_USER@$VPS_IP:$TARGET_DIR ..."
scp -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$ARCHIVE" "$VPS_USER@$VPS_IP:$TARGET_DIR/"

# ── Step 5: Extract & Start on VPS ───────────────────────────────────────────
echo "🔧 [5/5] Extracting and starting backend on VPS..."
ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$VPS_USER@$VPS_IP" bash <<EOF
  set -e
  cd $TARGET_DIR

  echo "  → Extracting archive..."
  tar -xzf $ARCHIVE
  rm -f $ARCHIVE

  echo "  → Creating ai_model_training directory if missing..."
  mkdir -p /app/rbia-backend/ai_model_training

  echo "  → Starting containers with pre-built binaries..."
  docker compose -f docker-compose.prod.yml up -d --build

  echo "  → Checking container status..."
  sleep 5
  docker compose -f docker-compose.prod.yml ps
EOF

# ── Cleanup local archive ─────────────────────────────────────────────────────
rm -f "$ARCHIVE"

echo ""
echo "✅ Backend deployment complete!"
echo ""
echo "   Kong API Gateway : http://$VPS_IP:8080"
echo "   Kong Admin       : http://$VPS_IP:8009"
echo "   Auth Service     : http://$VPS_IP:8001"
echo "   Audit Service    : http://$VPS_IP:8002"
echo "   Master Service   : http://$VPS_IP:8003"
echo "   Risk Service     : http://$VPS_IP:8004"
echo "   Analytics Service: http://$VPS_IP:8084"
echo "   Python AI        : http://$VPS_IP:8000"
echo ""
