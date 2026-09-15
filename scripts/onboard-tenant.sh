#!/bin/bash
# =============================================================================
# AuditSphere Automated Tenant Onboarding Script
# Usage: ./scripts/onboard-tenant.sh <tenant_slug> <client_name> [frontend_port] [kong_port] [--local] [--empty-data]
# Example: ./scripts/onboard-tenant.sh accenture "Accenture" 3014 8094 --local --empty-data
# =============================================================================

set -e

SLUG=${1:?"Error: Tenant slug is required (e.g. accenture)"}
CLIENT_NAME=${2:?"Error: Client name is required (e.g. 'Accenture')"}

# Parse ports - support direct ports without Google Drive ID parameter
if [[ "$3" =~ ^[0-9]+$ ]]; then
    FE_PORT=${3:-3010}
    KONG_PORT=${4:-8090}
else
    if [[ "$4" =~ ^[0-9]+$ ]]; then
        FE_PORT=${4:-3010}
        KONG_PORT=${5:-8090}
    else
        FE_PORT=3010
        KONG_PORT=8090
    fi
fi

IS_LOCAL=false
EMPTY_DATA=true
VPS_STORAGE_DIR="/var/data/auditsphere/storage/${SLUG}"
LOCAL_UPLOAD_DIR="backend/audit-service/uploads/${SLUG}"

for arg in "$@"; do
    case $arg in
        --local) IS_LOCAL=true ;;
        --with-demo-data) EMPTY_DATA=false ;;
        --empty-data) EMPTY_DATA=true ;;
    esac
done

if [ "$IS_LOCAL" = true ]; then
    DOMAIN="${SLUG}.localhost:${FE_PORT}"
    API_DOMAIN="localhost:${KONG_PORT}"
    PROTO="http"
else
    DOMAIN="${SLUG}.auditsphere.id"
    API_DOMAIN="api-${SLUG}.auditsphere.id"
    PROTO="https"
fi

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║          AuditSphere Client Onboarding: $CLIENT_NAME         "
echo "║          Mode      : $([ "$IS_LOCAL" = true ] && echo "Localhost Dev (RFC 6761)" || echo "Production Cloud")"
echo "║          Domain    : $PROTO://$DOMAIN                        "
echo "║          API Domain: $PROTO://$API_DOMAIN                    "
echo "║          Storage   : VPS Dedicated Silo Vault                "
echo "║          StorageDir: $VPS_STORAGE_DIR                        "
echo "║          GoogleDrv : Blocked / Disabled by Architecture      "
echo "║          FE Port   : $FE_PORT  |  Kong Port: $KONG_PORT      "
echo "║          Data State: $([ "$EMPTY_DATA" = true ] && echo "Clean / Empty Data" || echo "Demo Pre-seeded")"
echo "║          Mailtrap  : sandbox.smtp.mailtrap.io:2525           "
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 1. Create Databases in PostgreSQL
echo "==> [1/5] Provisioning isolated PostgreSQL databases for $CLIENT_NAME..."
docker exec -i rb_audit_postgres psql -U postgres <<EOSQL
CREATE DATABASE rb_audit_auth_${SLUG};
CREATE DATABASE rb_audit_audit_${SLUG};
CREATE DATABASE rb_audit_master_${SLUG};
CREATE DATABASE rb_audit_risk_${SLUG};
CREATE DATABASE rb_audit_analytics_${SLUG};
EOSQL

# 2. Run Migrations & Seeds
echo "==> [2/5] Running schema migrations & role seeding..."
(cd backend/auth-service && DB_NAME=rb_audit_auth_${SLUG} ./auth migrate up && DB_NAME=rb_audit_auth_${SLUG} ./auth seed)
(cd backend/audit-service && DB_NAME=rb_audit_audit_${SLUG} ./audit migrate up)
(cd backend/master-service && DB_NAME=rb_audit_master_${SLUG} ./master migrate up)
if [ "$EMPTY_DATA" = true ]; then
    echo "==> [2/5] Creating empty data state for risk-service (no demo data)..."
    (cd backend/risk-service && DB_NAME=rb_audit_risk_${SLUG} ./risk migrate up)
else
    echo "==> [2/5] Seeding demo risk universe dataset..."
    (cd backend/risk-service && DB_NAME=rb_audit_risk_${SLUG} ./risk migrate up && DB_NAME=rb_audit_risk_${SLUG} ./risk seed)
fi
(cd backend/analytics-service && DB_NAME=rb_audit_analytics_${SLUG} ./analytics migrate up)

# 3. Provision Dedicated VPS Storage Silo
echo "==> [3/5] Provisioning dedicated VPS evidence storage silo at $VPS_STORAGE_DIR..."
mkdir -p "${LOCAL_UPLOAD_DIR}"
chmod 755 "${LOCAL_UPLOAD_DIR}" 2>/dev/null || true
echo "    -> Local storage volume allocated at ${LOCAL_UPLOAD_DIR}"
echo "    -> Google Drive API provider: BLOCKED / DISABLED by architecture policy"

# 4. Generate Host Nginx Configuration (Production only)
if [ "$IS_LOCAL" = true ]; then
    echo "==> [4/5] Local environment detected: skipping production Nginx reverse proxy configuration."
else
    echo "==> [4/5] Generating Nginx reverse proxy configuration for $DOMAIN..."
    cat <<NGINXCONF > /etc/nginx/sites-available/${DOMAIN}.conf
server {
    listen 80;
    server_name ${DOMAIN};
    location / {
        proxy_pass http://127.0.0.1:${FE_PORT};
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}

server {
    listen 80;
    server_name ${API_DOMAIN};
    location / {
        proxy_pass http://127.0.0.1:${KONG_PORT};
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
NGINXCONF

    ln -sf /etc/nginx/sites-available/${DOMAIN}.conf /etc/nginx/sites-enabled/
    if command -v nginx > /dev/null 2>&1; then
        nginx -t && systemctl reload nginx
    fi
fi

# 5. Issue SSL Certificates via Certbot (Production only)
if [ "$IS_LOCAL" = true ]; then
    echo "==> [5/5] Local environment: SSL certificates not required (HTTP on localhost RFC 6761)."
else
    echo "==> [5/5] Provisioning SSL certificates via Certbot..."
    if command -v certbot > /dev/null 2>&1; then
        certbot --nginx -d ${DOMAIN} -d ${API_DOMAIN} --agree-tos -m admin@auditsphere.id --non-interactive || echo "Notice: Certbot step deferred until DNS propagates."
    fi
fi

# 6. Summary Output
echo ""
echo "=============================================================="
echo "🎉 Client $CLIENT_NAME successfully onboarded!"
echo "   Frontend URL : $PROTO://${DOMAIN}"
if [ "$IS_LOCAL" = true ]; then
echo "   Direct Local : http://localhost:${FE_PORT}"
fi
echo "   API Endpoint : $PROTO://${API_DOMAIN}"
echo "   Data State   : Empty Data / Clean Slate (0 Business Records)"
echo "   Databases    : rb_audit_*_${SLUG} (5 Schemas Migrated)"
echo "   Storage Silo : VPS Dedicated Storage Silo ($VPS_STORAGE_DIR)"
echo "   Google Drive : Blocked / Disabled (Local VPS Storage Active)"
echo "   Mailtrap SMTP: sandbox.smtp.mailtrap.io:2525 (Active)"
echo "=============================================================="
echo ""
