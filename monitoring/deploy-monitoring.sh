#!/bin/bash
# =============================================================================
# AuditSphere Monitoring Stack Deploy Script
# Deploys Prometheus + Grafana + Node Exporter + Loki + Promtail to VPS
# =============================================================================

VPS_IP="202.10.34.166"
VPS_USER="root"
TARGET_DIR="/app/rbia-monitoring"
ARCHIVE="monitoring.tar.gz"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

export COPYFILE_DISABLE=1

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║   AuditSphere Monitoring Stack → VPS ($VPS_IP)         ║"
echo "║   Prometheus · Grafana · Loki · Promtail · Node Exp    ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ── Step 1: Package monitoring configs ────────────────────────────
echo "📦 [1/4] Packaging monitoring configuration..."
tar \
  --exclude='.DS_Store' \
  --exclude='*.tar.gz' \
  -czf "$ARCHIVE" \
  docker-compose.monitoring.yml \
  prometheus/ \
  grafana/ \
  loki/ \
  promtail/

echo "   Archive size: $(du -sh $ARCHIVE | cut -f1)"

# ── Step 2: Prepare remote directory ──────────────────────────────
echo "🌐 [2/4] Preparing remote directory $TARGET_DIR on VPS..."
ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$VPS_USER@$VPS_IP" "mkdir -p $TARGET_DIR"

# ── Step 3: Upload ────────────────────────────────────────────────
echo "🚀 [3/4] Uploading monitoring stack to VPS..."
scp -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$ARCHIVE" "$VPS_USER@$VPS_IP:$TARGET_DIR/"

# ── Step 4: Extract & Start on VPS ────────────────────────────────
echo "🔧 [4/4] Extracting and starting monitoring stack on VPS..."
ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=6 -o StrictHostKeyChecking=no "$VPS_USER@$VPS_IP" bash <<EOF
  set -e
  cd $TARGET_DIR

  echo "  → Extracting archive..."
  tar -xzf $ARCHIVE
  rm -f $ARCHIVE

  echo "  → Pulling latest monitoring images..."
  docker compose -f docker-compose.monitoring.yml pull

  echo "  → Starting monitoring stack..."
  docker compose -f docker-compose.monitoring.yml up -d

  echo "  → Waiting for containers to stabilise..."
  sleep 8

  echo "  → Checking monitoring container status..."
  docker compose -f docker-compose.monitoring.yml ps

  echo ""
  echo "  → Verifying service health..."
  echo "    Prometheus: \$(curl -s -o /dev/null -w '%{http_code}' http://localhost:9090/-/healthy)"
  echo "    Grafana:    \$(curl -s -o /dev/null -w '%{http_code}' http://localhost:9300/api/health)"
  echo "    Loki:       \$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3100/ready)"
  echo "    Node Exp:   \$(curl -s -o /dev/null -w '%{http_code}' http://localhost:9100/metrics)"
EOF

# ── Cleanup local archive ─────────────────────────────────────────
rm -f "$ARCHIVE"

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║         🎉  Monitoring Stack Deployed!   🎉              ║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║                                                          ║"
echo "║  Grafana Dashboard                                       ║"
echo "║    http://$VPS_IP:9300                            ║"
echo "║    Login: admin / admin                                  ║"
echo "║                                                          ║"
echo "║  Prometheus                                              ║"
echo "║    http://$VPS_IP:9090                            ║"
echo "║    Targets: http://$VPS_IP:9090/targets           ║"
echo "║                                                          ║"
echo "║  Loki (Log Aggregation)                                  ║"
echo "║    http://$VPS_IP:3100                            ║"
echo "║                                                          ║"
echo "║  Node Exporter (Host Metrics)                            ║"
echo "║    http://$VPS_IP:9100                            ║"
echo "║                                                          ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
