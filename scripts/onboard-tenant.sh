#!/bin/bash
# =============================================================================
# AuditSphere Automated Tenant Onboarding Script
# Usage: ./scripts/onboard-tenant.sh <tenant_slug> <client_name> [frontend_port] [kong_port] [--local] [--empty-data]
# Example: ./scripts/onboard-tenant.sh accenture "Accenture" 3014 8094 --local --empty-data
#
# Options:
#   --local            Target <slug>.localhost over http; skip Nginx + Certbot
#   --empty-data       Migrate only, no demo records (default)
#   --with-demo-data   Also run the demo seeders
#   --skip-start       Provision everything but leave the stack stopped
#
# Pipeline: generate the tenant stack -> create its databases -> build images
# -> migrate inside the containers -> allocate storage -> email key -> start
# -> expose.
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REGISTRY="$ROOT_DIR/backend/tenants/registry.tsv"

die() { echo "Error: $*" >&2; exit 1; }

SLUG=${1:?"Error: Tenant slug is required (e.g. accenture)"}
CLIENT_NAME=${2:?"Error: Client name is required (e.g. 'Accenture')"}

# Parse ports - support direct ports without Google Drive ID parameter
if [[ "${3:-}" =~ ^[0-9]+$ ]]; then
    FE_PORT=${3}
    KONG_PORT=${4:-}
else
    if [[ "${4:-}" =~ ^[0-9]+$ ]]; then
        FE_PORT=${4}
        KONG_PORT=${5:-}
    else
        FE_PORT=""
        KONG_PORT=""
    fi
fi
[[ "$KONG_PORT" =~ ^[0-9]+$ ]] || KONG_PORT=""

IS_LOCAL=false
EMPTY_DATA=true
SKIP_START=false

for arg in "$@"; do
    case $arg in
        --local) IS_LOCAL=true ;;
        --with-demo-data) EMPTY_DATA=false ;;
        --empty-data) EMPTY_DATA=true ;;
        --skip-start) SKIP_START=true ;;
    esac
done

if [ "$IS_LOCAL" = true ]; then
    PROTO="http"
    STORAGE_DIR="$ROOT_DIR/backend/audit-service/uploads/${SLUG}"
else
    PROTO="https"
    STORAGE_DIR="/var/data/auditsphere/storage/${SLUG}"
fi

TENANT_DIR="$ROOT_DIR/backend/tenants/${SLUG}"

# -----------------------------------------------------------------------------
# [1/8] Generate the tenant's Compose stack and Kong config
#
# This allocates the host ports and Redis index, so it runs before anything
# that needs to know them. --force makes re-onboarding an existing tenant safe:
# the generator reuses that tenant's ports and signing secrets.
# -----------------------------------------------------------------------------
echo ""
echo "==> [1/8] Generating isolated tenant stack for $CLIENT_NAME..."

GEN_ARGS=("$SLUG" "$CLIENT_NAME" --force)
[ "$IS_LOCAL" = true ] && GEN_ARGS+=(--local)
[ -n "$FE_PORT" ]   && GEN_ARGS+=(--fe-port "$FE_PORT")
[ -n "$KONG_PORT" ] && GEN_ARGS+=(--kong-port "$KONG_PORT")

bash "$SCRIPT_DIR/generate-tenant-stack.sh" "${GEN_ARGS[@]}"

# Read back what was actually allocated — the generator may have skipped past a
# requested port if another tenant already holds it.
registry_field() { awk -F'\t' -v s="$SLUG" -v c="$1" 'NR>1 && $1==s {print $c; exit}' "$REGISTRY"; }
DOMAIN_HOST=$(registry_field 3)
API_DOMAIN_HOST=$(registry_field 4)
FE_PORT=$(registry_field 5)
KONG_PORT=$(registry_field 6)
REDIS_DB=$(registry_field 8)
[ -n "$FE_PORT" ] || die "tenant '$SLUG' is missing from $REGISTRY after generation"

if [ "$IS_LOCAL" = true ]; then
    DOMAIN="${DOMAIN_HOST}:${FE_PORT}"
    API_DOMAIN="localhost:${KONG_PORT}"
else
    DOMAIN="$DOMAIN_HOST"
    API_DOMAIN="$API_DOMAIN_HOST"
fi

COMPOSE=(docker compose --project-directory "$TENANT_DIR" --env-file "$TENANT_DIR/.env")

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║          AuditSphere Client Onboarding: $CLIENT_NAME         "
echo "║          Mode      : $([ "$IS_LOCAL" = true ] && echo "Localhost Dev (RFC 6761)" || echo "Production Cloud")"
echo "║          Domain    : $PROTO://$DOMAIN                        "
echo "║          API Domain: $PROTO://$API_DOMAIN                    "
echo "║          Storage   : VPS Dedicated Silo Vault                "
echo "║          StorageDir: $STORAGE_DIR                            "
echo "║          GoogleDrv : Blocked / Disabled by Architecture      "
echo "║          FE Port   : $FE_PORT  |  Kong Port: $KONG_PORT      "
echo "║          Redis DB  : $REDIS_DB (isolated logical index)      "
echo "║          Data State: $([ "$EMPTY_DATA" = true ] && echo "Clean / Empty Data" || echo "Demo Pre-seeded")"
echo "║          Email Svc : Resend, shared domain + tenant-scoped key"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# -----------------------------------------------------------------------------
# [2/8] Create the isolated databases
#
# analytics-service is intentionally absent: cmd/main.go starts a gin server and
# opens no database handle, so an rb_audit_analytics_<slug> database would never
# be connected to.
# -----------------------------------------------------------------------------
echo "==> [2/8] Provisioning isolated PostgreSQL databases for $CLIENT_NAME..."

PG_CONTAINER=""
for candidate in rb_audit_postgres rb_audit_postgres_dev; do
    if docker ps --format '{{.Names}}' | grep -qx "$candidate"; then
        PG_CONTAINER="$candidate"
        break
    fi
done
[ -n "$PG_CONTAINER" ] \
    || die "no running Postgres container found (looked for rb_audit_postgres, rb_audit_postgres_dev). Start the control-plane stack first."

for svc in auth audit master risk; do
    DB="rb_audit_${svc}_${SLUG}"
    # CREATE DATABASE has no IF NOT EXISTS, and re-onboarding an existing tenant
    # is a normal operation (recovery, config change), so check first.
    if docker exec -i "$PG_CONTAINER" psql -U postgres -tAc \
            "SELECT 1 FROM pg_database WHERE datname='${DB}'" | grep -q 1; then
        echo "    -> ${DB} already exists, leaving it untouched"
    else
        docker exec -i "$PG_CONTAINER" psql -U postgres -c "CREATE DATABASE ${DB};" >/dev/null
        echo "    -> created ${DB}"
    fi
done

# -----------------------------------------------------------------------------
# [3/8] Build the tenant images
# -----------------------------------------------------------------------------
echo ""
echo "==> [3/8] Building tenant service images..."
"${COMPOSE[@]}" build

# -----------------------------------------------------------------------------
# [4/8] Migrate and seed, inside the containers
#
# These run through `compose run` rather than as host binaries. The services
# resolve Postgres at the hostname `postgres`, which is a Docker network alias —
# it does not resolve on the host, and backend/docker-compose.yml publishes no
# Postgres port, so host-run migrations cannot reach the database at all.
#
# Running in-container also means DATABASE_NAME comes from the generated compose
# file, so a tenant migration can no longer be aimed at the shared database.
#
# Each service's migrate verb differs — they are not interchangeable:
#     auth   -> migrate up      audit  -> migrate up
#     master -> migrate         risk   -> up
# -----------------------------------------------------------------------------
echo ""
echo "==> [4/8] Running schema migrations & role seeding..."

run_svc() {  # run_svc <compose-service> <command...>
    local svc=$1; shift
    "${COMPOSE[@]}" run --rm --no-deps "$svc" "$@"
}

run_svc "auth-service-${SLUG}"   ./auth migrate up
run_svc "auth-service-${SLUG}"   ./auth seed          # roles & permissions — required to log in
run_svc "audit-service-${SLUG}"  ./audit migrate up
run_svc "master-service-${SLUG}" ./master migrate
run_svc "risk-service-${SLUG}"   ./risk up

if [ "$EMPTY_DATA" = true ]; then
    echo "    -> empty data state: demo seeders skipped (0 business records)"
else
    echo "    -> seeding demo dataset..."
    run_svc "audit-service-${SLUG}"  ./audit seed
    run_svc "master-service-${SLUG}" ./master seed
    run_svc "risk-service-${SLUG}"   ./risk seed
fi

# -----------------------------------------------------------------------------
# [5/8] Provision the dedicated storage silo
# -----------------------------------------------------------------------------
echo ""
echo "==> [5/8] Provisioning dedicated evidence storage silo at $STORAGE_DIR..."
mkdir -p "$STORAGE_DIR"
chmod 755 "$STORAGE_DIR" 2>/dev/null || true
echo "    -> storage volume allocated, bind-mounted into audit-service at /root/uploads"
echo "    -> Google Drive API provider: BLOCKED / DISABLED by architecture policy"

# -----------------------------------------------------------------------------
# [6/8] Outbound email: a sending-only Resend key for this tenant
#
# Runs before the stack starts so auth-service boots with the credentials.
# Skipped, not fatal, when the VPS has no Resend config: the tenant then falls
# back to the Mailtrap sandbox in config.yaml and sends nothing.
# -----------------------------------------------------------------------------
echo ""
EMAIL_ENV=${AUDITSPHERE_EMAIL_ENV:-/etc/auditsphere/email.env}
EMAIL_STATUS="not provisioned"
if [ "$IS_LOCAL" = true ]; then
    echo "==> [6/8] Local environment: skipping email provisioning."
elif [ ! -f "$EMAIL_ENV" ]; then
    echo "==> [6/8] $EMAIL_ENV not found: skipping email provisioning."
    echo "    -> later: ./scripts/provision-tenant-email.sh tenant $SLUG \"$CLIENT_NAME\""
else
    echo "==> [6/8] Provisioning outbound email for $CLIENT_NAME..."
    if bash "$SCRIPT_DIR/provision-tenant-email.sh" tenant "$SLUG" "$CLIENT_NAME" --no-restart; then
        EMAIL_STATUS="Resend (shared sending domain, tenant-scoped key)"
    else
        echo "    ! email provisioning failed — the tenant is up without email."
        echo "      Re-run: ./scripts/provision-tenant-email.sh tenant $SLUG \"$CLIENT_NAME\""
    fi
fi

# -----------------------------------------------------------------------------
# [7/8] Start the tenant stack
# -----------------------------------------------------------------------------
echo ""
if [ "$SKIP_START" = true ]; then
    echo "==> [7/8] --skip-start given: leaving the stack stopped."
else
    echo "==> [7/8] Starting tenant stack (5 services + Kong + frontend)..."
    "${COMPOSE[@]}" up -d
    "${COMPOSE[@]}" ps
fi

# -----------------------------------------------------------------------------
# [8/8] Expose it: host Nginx vhost + SSL (production only)
# -----------------------------------------------------------------------------
echo ""
if [ "$IS_LOCAL" = true ]; then
    echo "==> [8/8] Local environment: skipping Nginx reverse proxy and SSL."
    echo "    -> reach the tenant directly at http://localhost:${FE_PORT}"
else
    echo "==> [8/8] Generating Nginx reverse proxy configuration for $DOMAIN..."

    # Nginx layout differs by distribution: Debian/Ubuntu splits vhosts into
    # sites-available + a sites-enabled symlink, while RHEL/CentOS (which this
    # VPS runs) includes /etc/nginx/conf.d/*.conf directly. Writing to the wrong
    # one either fails outright or silently produces a vhost nginx never loads.
    if [ -d /etc/nginx/sites-available ]; then
        NGINX_CONF="/etc/nginx/sites-available/${DOMAIN}.conf"
        NGINX_LINK="/etc/nginx/sites-enabled/${DOMAIN}.conf"
    elif [ -d /etc/nginx/conf.d ]; then
        NGINX_CONF="/etc/nginx/conf.d/${DOMAIN}.conf"
        NGINX_LINK=""
    else
        die "no recognised Nginx vhost directory (/etc/nginx/sites-available or /etc/nginx/conf.d)"
    fi
    echo "    -> writing $NGINX_CONF"

    cat <<NGINXCONF > "$NGINX_CONF"
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
    client_max_body_size 100m;
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

    [ -n "$NGINX_LINK" ] && ln -sf "$NGINX_CONF" "$NGINX_LINK"
    if command -v nginx > /dev/null 2>&1; then
        nginx -t && systemctl reload nginx
    fi

    # Certbot needs the domain to resolve to this host before it can complete
    # the HTTP-01 challenge, so a tenant provisioned ahead of its DNS record
    # will fail here. That is not fatal: the tenant serves plain HTTP until
    # someone re-runs certbot once the record propagates.
    echo "    -> provisioning SSL certificates via Certbot..."
    # getent rather than `host`: host ships in bind-utils, which a minimal RHEL
    # install lacks, and a missing binary would read as "does not resolve".
    # With the wildcard record *.auditsphere.app in place both names resolve
    # the moment the tenant exists, so this branch only fires without it.
    if ! getent hosts "$DOMAIN" > /dev/null 2>&1 || ! getent hosts "$API_DOMAIN" > /dev/null 2>&1; then
        echo "    ! $DOMAIN does not resolve yet — skipping Certbot."
        echo "      Add an A record for ${DOMAIN} and ${API_DOMAIN} -> this host, then run:"
        echo "      certbot --nginx -d ${DOMAIN} -d ${API_DOMAIN} --agree-tos -m admin@auditsphere.app --non-interactive"
    elif command -v certbot > /dev/null 2>&1; then
        certbot --nginx -d ${DOMAIN} -d ${API_DOMAIN} --agree-tos -m admin@auditsphere.app --non-interactive \
            || echo "    ! Certbot deferred — the tenant is serving plain HTTP until you re-run it."
    fi
fi

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
echo ""
echo "=============================================================="
echo "🎉 Client $CLIENT_NAME successfully onboarded!"
echo "   Frontend URL : $PROTO://${DOMAIN}"
if [ "$IS_LOCAL" = true ]; then
echo "   Direct Local : http://localhost:${FE_PORT}"
fi
echo "   API Endpoint : $PROTO://${API_DOMAIN}"
echo "   Stack Dir    : backend/tenants/${SLUG}"
echo "   Data State   : $([ "$EMPTY_DATA" = true ] && echo "Empty Data / Clean Slate (0 Business Records)" || echo "Demo Pre-seeded")"
echo "   Databases    : rb_audit_{auth,audit,master,risk}_${SLUG}"
echo "   Redis DB     : ${REDIS_DB} (isolated logical index)"
echo "   Storage Silo : ${STORAGE_DIR}"
echo "   Google Drive : Blocked / Disabled (Local VPS Storage Active)"
echo "   Email Svc    : ${EMAIL_STATUS}"
echo "=============================================================="
echo ""
echo "⚠️  Before handing this instance to the client:"
echo "   1. The auth seeder creates the default admin account (admin / password123)."
echo "      Rotate it — it is identical on every tenant."
echo "   2. Check email delivery:"
echo "      ./scripts/provision-tenant-email.sh test ${SLUG} you@example.com"
echo "   3. Back up backend/tenants/${SLUG}/.env — it holds this tenant's JWT"
echo "      signing secret and email key, and is not in git."
echo ""
