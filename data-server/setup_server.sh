#!/usr/bin/env bash
# ============================================================================
# AuditSphere Data Hub — Server Setup Script
# Target: AlmaLinux 10 (KVM VPS, 2 CPU / 8 GB RAM / 100 GB Disk)
# Run as root: chmod +x setup_server.sh && ./setup_server.sh
# ============================================================================
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()   { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
err()   { echo -e "${RED}[✗]${NC} $1"; exit 1; }

echo "============================================================"
echo "  AuditSphere Data Hub — Server Setup"
echo "  Target: AlmaLinux 10 KVM"
echo "============================================================"

# --- 1. OS Update ---
log "Updating OS packages..."
dnf update -y -q

# --- 2. Remove old Docker packages ---
log "Removing old Docker packages (if any)..."
dnf remove -y docker docker-client docker-client-latest docker-common \
  docker-latest docker-latest-logrotate docker-logrotate docker-engine 2>/dev/null || true

# --- 3. Install Docker CE + Compose Plugin ---
log "Installing Docker CE repository..."
dnf install -y dnf-plugins-core
dnf config-manager --add-repo https://download.docker.com/linux/rhel/docker-ce.repo

log "Installing Docker CE, CLI, containerd, and Compose plugin..."
dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# --- 4. Enable & Start Docker ---
log "Enabling and starting Docker service..."
systemctl enable --now docker
systemctl is-active --quiet docker && log "Docker is running." || err "Docker failed to start."

# --- 5. Install WireGuard ---
log "Installing WireGuard..."
dnf install -y wireguard-tools

# --- 6. Create directory structure ---
DATA_ROOT="/opt/auditsphere-data"
log "Creating directory structure at ${DATA_ROOT}..."
mkdir -p "${DATA_ROOT}"/{nginx,wireguard,data-hub-api,ai-engine,caatt-engine,notebooks/helpers}
mkdir -p "${DATA_ROOT}"/volumes/{cbs-data,datalake-data,ai-models,jupyter-data,hub-data}
chown -R root:root "${DATA_ROOT}"
chmod -R 750 "${DATA_ROOT}/volumes"

# --- 7. Install httpd-tools for htpasswd ---
log "Installing httpd-tools (for htpasswd)..."
dnf install -y httpd-tools

# --- 8. Generate htpasswd file ---
HTPASSWD_FILE="${DATA_ROOT}/nginx/htpasswd"
if [ ! -f "${HTPASSWD_FILE}" ]; then
    log "Generating htpasswd file..."
    # Read passwords from .env or generate defaults
    ETL_PASS="${ETL_USER_PASSWORD:-$(openssl rand -base64 16)}"
    AUDITOR_PASS="${AUDITOR_AI_PASSWORD:-$(openssl rand -base64 16)}"

    htpasswd -bc "${HTPASSWD_FILE}" etl_user "${ETL_PASS}"
    htpasswd -b  "${HTPASSWD_FILE}" auditor_ai "${AUDITOR_PASS}"
    chmod 640 "${HTPASSWD_FILE}"
    log "htpasswd created. Users: etl_user, auditor_ai"
    warn "ETL_USER password: ${ETL_PASS}"
    warn "AUDITOR_AI password: ${AUDITOR_PASS}"
    warn "Save these passwords securely! They will not be shown again."
else
    log "htpasswd already exists. Skipping."
fi

# --- 9. Generate WireGuard keys ---
WG_DIR="${DATA_ROOT}/wireguard"
if [ ! -f "${WG_DIR}/server_private.key" ]; then
    log "Generating WireGuard keys..."
    wg genkey | tee "${WG_DIR}/server_private.key" | wg pubkey > "${WG_DIR}/server_public.key"
    chmod 600 "${WG_DIR}/server_private.key"
    log "WireGuard keys generated."
    echo ""
    warn "=== WireGuard Server Public Key (share with audit server) ==="
    cat "${WG_DIR}/server_public.key"
    echo ""
else
    log "WireGuard keys already exist. Skipping."
fi

# --- 10. Configure Firewall ---
log "Configuring firewall..."
if command -v firewall-cmd &>/dev/null; then
    # Basic ports
    firewall-cmd --permanent --add-port=22/tcp    2>/dev/null || true  # SSH
    firewall-cmd --permanent --add-port=51820/udp  2>/dev/null || true  # WireGuard

    # Create trusted zone for VPN traffic
    firewall-cmd --permanent --new-zone=vpn 2>/dev/null || true
    firewall-cmd --permanent --zone=vpn --add-source=10.0.0.0/24 2>/dev/null || true
    firewall-cmd --permanent --zone=vpn --add-port=80/tcp    2>/dev/null || true  # Nginx
    firewall-cmd --permanent --zone=vpn --add-port=8000/tcp  2>/dev/null || true  # AI Engine
    firewall-cmd --permanent --zone=vpn --add-port=8100/tcp  2>/dev/null || true  # Data Hub API

    firewall-cmd --reload
    log "Firewall configured. Ports 8000/8100 only accessible via VPN (10.0.0.0/24)."
else
    warn "firewall-cmd not found. Please configure iptables manually."
fi

# --- 11. Verify Installation ---
echo ""
echo "============================================================"
echo "  Setup Complete — Verification"
echo "============================================================"
docker --version
docker compose version
wg --version 2>/dev/null || echo "WireGuard CLI: installed"
echo ""
log "All components installed successfully."
log "Next steps:"
echo "  1. Copy project files to ${DATA_ROOT}/"
echo "  2. cp .env.example .env && nano .env"
echo "  3. Configure WireGuard: cp wireguard/wg0-server.conf /etc/wireguard/wg0.conf"
echo "  4. Start WireGuard: systemctl enable --now wg-quick@wg0"
echo "  5. docker compose up -d --build"
echo "============================================================"
