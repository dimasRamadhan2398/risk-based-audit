#!/bin/bash
# =============================================================================
# AuditSphere Tenant Stack Generator
#
# Renders a tenant's Docker Compose stack and Kong declarative config from
# scripts/templates/, allocating host ports, a Redis logical DB, and
# per-tenant signing secrets.
#
# Output: backend/tenants/<slug>/
#     docker-compose.yml   generated, safe to commit
#     kong/kong.yml        generated, safe to commit
#     .env                 secrets, mode 600, gitignored
#
# Usage:
#   ./scripts/generate-tenant-stack.sh <slug> <client_name> [options]
#
# Options:
#   --local                 Target <slug>.localhost, http, no API subdomain
#   --fe-port <n>           Pin the frontend host port   (default: auto)
#   --kong-port <n>         Pin the Kong proxy host port (default: auto)
#   --kong-admin-port <n>   Pin the Kong admin host port (default: auto)
#   --storage-dir <path>    Evidence silo host path      (default: per env)
#   --force                 Overwrite an existing tenant's generated files
#
# Example:
#   ./scripts/generate-tenant-stack.sh accenture "Accenture" --fe-port 3010
#
# Regenerating an existing tenant reuses its allocated ports, Redis index and
# secrets from the registry. Rotating JWT_SECRET would invalidate every live
# session for that client, so it never happens implicitly.
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TEMPLATE_DIR="$SCRIPT_DIR/templates"
TENANTS_DIR="$ROOT_DIR/backend/tenants"
REGISTRY="$TENANTS_DIR/registry.tsv"

# Allocation bases, matching the worked example in
# docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md section 2.
BASE_FE_PORT=3010
BASE_KONG_PORT=8090
BASE_KONG_ADMIN_PORT=8019
BASE_REDIS_DB=1   # index 0 belongs to the control-plane stack

die() { echo "Error: $*" >&2; exit 1; }

# ── Arguments ────────────────────────────────────────────────────────────────

SLUG=${1:-}
CLIENT_NAME=${2:-}
[ -n "$SLUG" ]        || die "tenant slug is required (e.g. accenture)"
[ -n "$CLIENT_NAME" ] || die "client name is required (e.g. 'Accenture')"
shift 2

IS_LOCAL=false
FORCE=false
FE_PORT=""
KONG_PORT=""
KONG_ADMIN_PORT=""
STORAGE_DIR=""

while [ $# -gt 0 ]; do
    case $1 in
        --local)            IS_LOCAL=true ;;
        --force)            FORCE=true ;;
        --fe-port)          FE_PORT=${2:?"--fe-port needs a value"}; shift ;;
        --kong-port)        KONG_PORT=${2:?"--kong-port needs a value"}; shift ;;
        --kong-admin-port)  KONG_ADMIN_PORT=${2:?"--kong-admin-port needs a value"}; shift ;;
        --storage-dir)      STORAGE_DIR=${2:?"--storage-dir needs a value"}; shift ;;
        *) die "unknown option: $1" ;;
    esac
    shift
done

# The slug becomes a DNS label, a Postgres database name and a Docker network
# alias. Restricting it to [a-z0-9] keeps it legal in all three.
[[ "$SLUG" =~ ^[a-z][a-z0-9]{1,30}$ ]] \
    || die "slug '$SLUG' is invalid — use 2-31 chars, lowercase letters and digits, starting with a letter"

[ -d "$TEMPLATE_DIR" ] || die "template directory not found: $TEMPLATE_DIR"

if [ "$IS_LOCAL" = true ]; then
    PROTO="http"
    DOMAIN_HOST="${SLUG}.localhost"
    API_DOMAIN_HOST="localhost"
    DEFAULT_STORAGE_DIR="$ROOT_DIR/backend/audit-service/uploads/${SLUG}"
else
    PROTO="https"
    DOMAIN_HOST="${SLUG}.auditsphere.id"
    API_DOMAIN_HOST="api-${SLUG}.auditsphere.id"
    DEFAULT_STORAGE_DIR="/var/data/auditsphere/storage/${SLUG}"
fi
STORAGE_DIR=${STORAGE_DIR:-$DEFAULT_STORAGE_DIR}

# ── Registry ─────────────────────────────────────────────────────────────────
# A TSV file is the interim registry. The tenant table in Postgres (step 3 of
# the provisioning plan) replaces this; until then it is what stops two tenants
# being handed the same host port or Redis index.

mkdir -p "$TENANTS_DIR"
if [ ! -f "$REGISTRY" ]; then
    printf 'slug\tclient_name\tdomain\tapi_domain\tfe_port\tkong_port\tkong_admin_port\tredis_db\tcreated_at\n' > "$REGISTRY"
fi

registry_field() {  # registry_field <slug> <1-based column>
    awk -F'\t' -v s="$1" -v c="$2" 'NR>1 && $1==s {print $c; exit}' "$REGISTRY"
}

port_taken() {      # port_taken <port> <column> <excluding-slug>
    awk -F'\t' -v p="$1" -v c="$2" -v s="$3" 'NR>1 && $1!=s && $c==p {found=1} END{exit !found}' "$REGISTRY"
}

next_free() {       # next_free <base> <column> <excluding-slug>
    local candidate=$1
    while port_taken "$candidate" "$2" "$3"; do
        candidate=$((candidate + 1))
    done
    echo "$candidate"
}

EXISTING=$(registry_field "$SLUG" 1)
TENANT_DIR="$TENANTS_DIR/$SLUG"

if [ -n "$EXISTING" ]; then
    # Reuse every allocation. Changing a live tenant's ports would break the
    # host Nginx vhost that points at them.
    FE_PORT=${FE_PORT:-$(registry_field "$SLUG" 5)}
    KONG_PORT=${KONG_PORT:-$(registry_field "$SLUG" 6)}
    KONG_ADMIN_PORT=${KONG_ADMIN_PORT:-$(registry_field "$SLUG" 7)}
    REDIS_DB=$(registry_field "$SLUG" 8)
    CREATED_AT=$(registry_field "$SLUG" 9)
    echo "==> Tenant '$SLUG' already registered — reusing ports ${FE_PORT}/${KONG_PORT} and Redis DB ${REDIS_DB}."
    if [ -d "$TENANT_DIR" ] && [ "$FORCE" != true ]; then
        die "$TENANT_DIR already exists. Re-run with --force to regenerate its files."
    fi
else
    FE_PORT=${FE_PORT:-$(next_free "$BASE_FE_PORT" 5 "$SLUG")}
    KONG_PORT=${KONG_PORT:-$(next_free "$BASE_KONG_PORT" 6 "$SLUG")}
    KONG_ADMIN_PORT=${KONG_ADMIN_PORT:-$(next_free "$BASE_KONG_ADMIN_PORT" 7 "$SLUG")}
    REDIS_DB=$(next_free "$BASE_REDIS_DB" 8 "$SLUG")
    CREATED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)

    # Redis ships with `databases 16`. Past index 15 the SELECT fails at
    # runtime, which would surface as tenants silently sharing index 0.
    [ "$REDIS_DB" -le 15 ] \
        || die "no free Redis logical DB (allocated up to $REDIS_DB, limit 15). Raise \`databases\` in the Redis config or move to key-prefixed isolation."
fi

for p in "$FE_PORT" "$KONG_PORT" "$KONG_ADMIN_PORT"; do
    [[ "$p" =~ ^[0-9]+$ ]] || die "port '$p' is not numeric"
done

# ── Secrets ──────────────────────────────────────────────────────────────────
# Reused if the tenant's .env already holds them, so regeneration does not log
# every auditor out.

mkdir -p "$TENANT_DIR/kong"
ENV_FILE="$TENANT_DIR/.env"

read_env_value() {  # read_env_value <key>
    [ -f "$ENV_FILE" ] || return 0
    awk -F= -v k="$1" '$1==k {sub(/^[^=]*=/,""); print; exit}' "$ENV_FILE"
}

JWT_SECRET=$(read_env_value TENANT_JWT_SECRET)
SIGNATURE_KEY=$(read_env_value TENANT_SIGNATURE_KEY)

if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -hex 32) || die "openssl is required to generate tenant secrets"
    echo "==> Generated a new JWT signing secret for '$SLUG'."
else
    echo "==> Reusing the existing JWT signing secret for '$SLUG' (rotating it would end all active sessions)."
fi
[ -n "$SIGNATURE_KEY" ] || SIGNATURE_KEY=$(openssl rand -hex 32)

umask 077
cat > "$ENV_FILE" <<ENVEOF
# =============================================================================
# Tenant secrets — $CLIENT_NAME ($SLUG)
#
# GENERATED. Not committed (see .gitignore). Loaded by docker compose via
# --env-file. Back this up with the tenant's database: losing JWT_SECRET ends
# every active session; losing SIGNATURE_KEY invalidates anything signed with it.
# =============================================================================
TENANT_JWT_SECRET=$JWT_SECRET
TENANT_SIGNATURE_KEY=$SIGNATURE_KEY
ENVEOF
chmod 600 "$ENV_FILE"
umask 022

# ── Render ───────────────────────────────────────────────────────────────────
# Secrets are rendered as compose variable references, not literals, so the
# generated compose file stays safe to commit. `:?` makes compose fail loudly
# if the stack is started without --env-file rather than booting with an empty
# secret.

render() {  # render <template> <destination>
    sed \
        -e "s|__SLUG__|${SLUG}|g" \
        -e "s|__CLIENT_NAME__|${CLIENT_NAME}|g" \
        -e "s|__DOMAIN__|${PROTO}://${DOMAIN_HOST}|g" \
        -e "s|__API_DOMAIN__|${API_DOMAIN_HOST}|g" \
        -e "s|__PROTO__|${PROTO}|g" \
        -e "s|__FE_PORT__|${FE_PORT}|g" \
        -e "s|__KONG_PORT__|${KONG_PORT}|g" \
        -e "s|__KONG_ADMIN_PORT__|${KONG_ADMIN_PORT}|g" \
        -e "s|__REDIS_DB__|${REDIS_DB}|g" \
        -e "s|__STORAGE_DIR__|${STORAGE_DIR}|g" \
        -e "s|__JWT_SECRET__|\${TENANT_JWT_SECRET:?TENANT_JWT_SECRET missing — start this stack with --env-file $(basename "$ENV_FILE")}|g" \
        -e "s|__SIGNATURE_KEY__|\${TENANT_SIGNATURE_KEY:?TENANT_SIGNATURE_KEY missing — start this stack with --env-file $(basename "$ENV_FILE")}|g" \
        "$1" > "$2"
}

# __DOMAIN__ expands to a full origin (scheme included) for the Kong CORS
# plugin, so strip the duplicated scheme the template writes in front of it.
render "$TEMPLATE_DIR/docker-compose.tenant.yml.tpl" "$TENANT_DIR/docker-compose.yml"
render "$TEMPLATE_DIR/kong-tenant.yml.tpl" "$TENANT_DIR/kong/kong.yml"
sed -i.bak "s|${PROTO}://${PROTO}://|${PROTO}://|g" "$TENANT_DIR/kong/kong.yml"
rm -f "$TENANT_DIR/kong/kong.yml.bak"

# ── Record ───────────────────────────────────────────────────────────────────

TMP_REGISTRY=$(mktemp)
awk -F'\t' -v s="$SLUG" 'NR==1 || $1!=s' "$REGISTRY" > "$TMP_REGISTRY"
printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
    "$SLUG" "$CLIENT_NAME" "$DOMAIN_HOST" "$API_DOMAIN_HOST" \
    "$FE_PORT" "$KONG_PORT" "$KONG_ADMIN_PORT" "$REDIS_DB" "$CREATED_AT" \
    >> "$TMP_REGISTRY"
mv "$TMP_REGISTRY" "$REGISTRY"

# ── Preflight ────────────────────────────────────────────────────────────────
# Both Dockerfiles copy pre-built artefacts rather than compiling, so the build
# succeeds with a stale or missing binary and only fails at container start.

WARNINGS=0
warn() { echo "    ! $*"; WARNINGS=$((WARNINGS + 1)); }

echo ""
echo "==> Preflight"
for svc in auth audit master risk analytics; do
    [ -f "$ROOT_DIR/backend/${svc}-service/${svc}" ] \
        || warn "backend/${svc}-service/${svc} is missing — build it for linux/amd64 before starting the stack"
done
[ -d "$ROOT_DIR/frontend/.output" ] \
    || warn "frontend/.output is missing — run 'npm run build' in frontend/ before starting the stack"
docker network inspect rb_audit_network >/dev/null 2>&1 \
    || warn "the rb_audit_network bridge does not exist — start the control-plane stack (backend/docker-compose.yml) first"
[ "$WARNINGS" -eq 0 ] && echo "    all checks passed"

# ── Summary ──────────────────────────────────────────────────────────────────

cat <<SUMMARY

==============================================================
 Tenant stack generated: $CLIENT_NAME ($SLUG)
==============================================================
 Directory     : backend/tenants/$SLUG
 Frontend      : $PROTO://$DOMAIN_HOST          -> 127.0.0.1:$FE_PORT
 API Gateway   : $PROTO://$API_DOMAIN_HOST      -> 127.0.0.1:$KONG_PORT
 Kong Admin    : 127.0.0.1:$KONG_ADMIN_PORT (loopback only)
 Redis DB      : $REDIS_DB
 Storage silo  : $STORAGE_DIR
 Databases     : rb_audit_{auth,audit,master,risk}_$SLUG

 Start it:
   docker compose --project-directory backend/tenants/$SLUG \\
     --env-file backend/tenants/$SLUG/.env up -d --build
==============================================================

SUMMARY
