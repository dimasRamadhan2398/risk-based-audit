#!/bin/bash
# =============================================================================
# AuditSphere Automated Tenant Onboarding Script
# Usage: ./scripts/onboard-tenant.sh <tenant_slug> <client_name> <gdrive_folder_id> [frontend_port] [kong_port]
# Example: ./scripts/onboard-tenant.sh accenture "Accenture" "1bX7yZ9kL0mN8pQ2rS4tU6vW8xYz1234A" 3010 8090
# =============================================================================

set -e

SLUG=${1:?"Error: Tenant slug is required (e.g. accenture)"}
CLIENT_NAME=${2:?"Error: Client name is required (e.g. 'Accenture')"}
GDRIVE_ID=${3:?"Error: Google Drive Folder ID is required"}
FE_PORT=${4:-3010}
KONG_PORT=${5:-8090}

DOMAIN="${SLUG}.auditsphere.id"
API_DOMAIN="api-${SLUG}.auditsphere.id"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║          AuditSphere Client Onboarding: $CLIENT_NAME         "
echo "║          Domain    : https://$DOMAIN                         "
echo "║          API Domain: https://$API_DOMAIN                     "
echo "║          GDrive ID : $GDRIVE_ID                              "
echo "║          FE Port   : $FE_PORT  |  Kong Port: $KONG_PORT      "
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
echo "==> [2/5] Running schema migrations & initial seeds..."
(cd backend/auth-service && DB_NAME=rb_audit_auth_${SLUG} ./auth migrate up && DB_NAME=rb_audit_auth_${SLUG} ./auth seed)
(cd backend/audit-service && DB_NAME=rb_audit_audit_${SLUG} ./audit migrate up)
(cd backend/master-service && DB_NAME=rb_audit_master_${SLUG} ./master migrate up)
(cd backend/risk-service && DB_NAME=rb_audit_risk_${SLUG} ./risk migrate up && DB_NAME=rb_audit_risk_${SLUG} ./risk seed)
(cd backend/analytics-service && DB_NAME=rb_audit_analytics_${SLUG} ./analytics migrate up)

# 3. Generate Host Nginx Configuration
echo "==> [3/5] Generating Nginx reverse proxy configuration for $DOMAIN..."
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

# 4. Issue SSL Certificates via Certbot
echo "==> [4/5] Provisioning SSL certificates via Certbot..."
if command -v certbot > /dev/null 2>&1; then
    certbot --nginx -d ${DOMAIN} -d ${API_DOMAIN} --agree-tos -m admin@auditsphere.id --non-interactive || echo "Notice: Certbot step deferred until DNS propagates."
fi

# 5. Summary Output
echo ""
echo "=============================================================="
echo "🎉 Client $CLIENT_NAME successfully onboarded!"
echo "   Frontend URL : https://${DOMAIN}"
echo "   API Endpoint : https://${API_DOMAIN}"
echo "   Databases    : rb_audit_*_${SLUG}"
echo "   GDrive Root  : $GDRIVE_ID"
echo "=============================================================="
echo ""
