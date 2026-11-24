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
#   TENANT_STORAGE_ROOT=/var/data/auditsphere/storage \
#     ./scripts/generate-tenant-stack.sh <slug> <client_name> [options]
#
#   ./scripts/generate-tenant-stack.sh --next-ports
#       Read-only. Prints the next free frontend and Kong (API) ports as one
#       line of JSON, e.g. {"fe_port":3011,"api_port":8091}, skipping ports
#       held in the registry and ports already listening on this host.
#       Needs no other argument or env var and writes nothing. The
#       provisioner uses it for GET /ports/next.
#
# The evidence silo is always <TENANT_STORAGE_ROOT>/<slug>. TENANT_STORAGE_ROOT
# must be set to an absolute path; the generator refuses to run without it.
#
# Options:
#   --local                 Target <slug>.localhost, http, no API subdomain
#   --fe-port <n>           Pin the frontend host port   (default: auto)
#   --kong-port <n>         Pin the Kong proxy host port (default: auto)
#   --kong-admin-port <n>   Pin the Kong admin host port (default: auto)
#                           A pinned port is REJECTED (exit 1) when another
#                           tenant holds it in the registry or something on
#                           this host already listens on it. A tenant's own
#                           registered ports are accepted even while its
#                           stack is running.
#   --base-domain <fqdn>    Parent domain for the tenant (default: auditsphere.app)
#   --image-tag <tag>       Tag for this tenant's images, written to .env as
#                           TENANT_IMAGE_TAG. Default: keep the tag already in
#                           the tenant's .env, else <UTC timestamp>-<git sha>.
#                           onboard-tenant.sh passes a fresh tag on every run.
#   --no-preflight          Render even when the prebuilt binaries or
#                           frontend/.output are missing (config-only)
#   --ic-lha-due-days <n>   Internal control: audit report (LHA) due days, 1-365
#                           (default: the tenant's current .env value, else 30)
#   --ic-mgmt-response-due-days <n>
#                           Internal control: management response due days,
#                           1-365 (default: current .env value, else 14)
#   --force                 Overwrite an existing tenant's generated files
#
# Example:
#   ./scripts/generate-tenant-stack.sh accenture "Accenture" --fe-port 3010
#
# Regenerating an existing tenant reuses its allocated ports, Redis index,
# secrets and image tag from the registry and its .env. Rotating JWT_SECRET
# would invalidate every live session for that client, so it never happens
# implicitly.
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

# The registered, delegated domain. Tenants are subdomains of it. Overridable
# with --base-domain or BASE_DOMAIN for staging or a future second domain.
# NOTE: auditsphere.id is NOT registered — it has no NS or SOA records and the
# .id registry returns DOMAIN NOT FOUND. Earlier drafts of the provisioning
# docs assumed it; anything provisioned under it can never resolve.
BASE_DOMAIN=${BASE_DOMAIN:-auditsphere.app}

die() { echo "Error: $*" >&2; exit 1; }

# ── Registry helpers ─────────────────────────────────────────────────────────
# A TSV file is the interim registry. The tenant table in Postgres (step 3 of
# the provisioning plan) replaces this; until then it is what stops two tenants
# being handed the same host port or Redis index. A missing registry reads as
# empty, so the read-only --next-ports mode never has to create it.

registry_field() {  # registry_field <slug> <1-based column>
    [ -f "$REGISTRY" ] || return 0
    awk -F'\t' -v s="$1" -v c="$2" 'NR>1 && $1==s {print $c; exit}' "$REGISTRY"
}

registry_holder() { # registry_holder <port> <column> <excluding-slug>  -> slug or ""
    [ -f "$REGISTRY" ] || return 0
    awk -F'\t' -v p="$1" -v c="$2" -v s="$3" 'NR>1 && $1!=s && $c==p {print $1; exit}' "$REGISTRY"
}

port_taken() {      # port_taken <port> <column> <excluding-slug>
    [ -n "$(registry_holder "$1" "$2" "$3")" ]
}

# Is anything listening on this TCP port on the host? Docker-published ports
# show up here too (docker-proxy, or Docker Desktop's backend). Tries ss
# (Linux), then netstat (also Git Bash on Windows), then a loopback connect.
port_in_use_os() {  # port_in_use_os <port>
    local p=$1
    if command -v ss >/dev/null 2>&1; then
        [ -n "$(ss -Htln "sport = :$p" 2>/dev/null)" ]
    elif command -v netstat >/dev/null 2>&1; then
        netstat -an 2>/dev/null \
            | grep -Eq "[:.]${p}[[:space:]]+[^[:space:]]+[[:space:]]+LISTEN(ING)?([[:space:]]|\$)"
    else
        (exec 3<>"/dev/tcp/127.0.0.1/$p") 2>/dev/null
    fi
}

next_free() {       # next_free <base> <column> <excluding-slug>
    local candidate=$1
    while port_taken "$candidate" "$2" "$3" || port_in_use_os "$candidate"; do
        candidate=$((candidate + 1))
    done
    echo "$candidate"
}

# ── Read-only mode: --next-ports ─────────────────────────────────────────────
# No slug, no storage root, no writes. The excluded slug is empty, so every
# registered tenant counts. Not a reservation: the provisioner runs one job at
# a time, and the real run checks the ports again.

if [ "${1:-}" = "--next-ports" ]; then
    [ $# -eq 1 ] || die "--next-ports takes no other arguments"
    NEXT_FE=$(next_free "$BASE_FE_PORT" 5 "")
    NEXT_API=$(next_free "$BASE_KONG_PORT" 6 "")
    printf '{"fe_port":%d,"api_port":%d}\n' "$NEXT_FE" "$NEXT_API"
    exit 0
fi

# ── Arguments ────────────────────────────────────────────────────────────────

SLUG=${1:-}
CLIENT_NAME=${2:-}
[ -n "$SLUG" ]        || die "tenant slug is required (e.g. accenture)"
[ -n "$CLIENT_NAME" ] || die "client name is required (e.g. 'Accenture')"
shift 2

IS_LOCAL=false
FORCE=false
PREFLIGHT=true
FE_PORT=""
KONG_PORT=""
KONG_ADMIN_PORT=""
IMAGE_TAG=""
IC_LHA_DUE_DAYS=""
IC_MGMT_RESPONSE_DUE_DAYS=""

DEFAULT_IC_LHA_DUE_DAYS=30
DEFAULT_IC_MGMT_RESPONSE_DUE_DAYS=14

while [ $# -gt 0 ]; do
    case $1 in
        --local)            IS_LOCAL=true ;;
        --force)            FORCE=true ;;
        --no-preflight)     PREFLIGHT=false ;;
        --fe-port)          FE_PORT=${2:?"--fe-port needs a value"}; shift ;;
        --kong-port)        KONG_PORT=${2:?"--kong-port needs a value"}; shift ;;
        --kong-admin-port)  KONG_ADMIN_PORT=${2:?"--kong-admin-port needs a value"}; shift ;;
        --base-domain)      BASE_DOMAIN=${2:?"--base-domain needs a value"}; shift ;;
        --image-tag)        IMAGE_TAG=${2:?"--image-tag needs a value"}; shift ;;
        --ic-lha-due-days)  IC_LHA_DUE_DAYS=${2:?"--ic-lha-due-days needs a value"}; shift ;;
        --ic-lha-due-days=*) IC_LHA_DUE_DAYS=${1#*=} ;;
        --ic-mgmt-response-due-days) IC_MGMT_RESPONSE_DUE_DAYS=${2:?"--ic-mgmt-response-due-days needs a value"}; shift ;;
        --ic-mgmt-response-due-days=*) IC_MGMT_RESPONSE_DUE_DAYS=${1#*=} ;;
        *) die "unknown option: $1" ;;
    esac
    shift
done

# The slug becomes a DNS label, a Postgres database name and a Docker network
# alias. Restricting it to [a-z0-9] keeps it legal in all three.
[[ "$SLUG" =~ ^[a-z][a-z0-9]{1,30}$ ]] \
    || die "slug '$SLUG' is invalid — use 2-31 chars, lowercase letters and digits, starting with a letter"

# Internal control thresholds are whole days, 1-365. No leading zeros: bash
# arithmetic would read "030" as octal.
valid_due_days() { [[ "$1" =~ ^[1-9][0-9]{0,2}$ ]] && [ "$1" -le 365 ]; }
[ -z "$IC_LHA_DUE_DAYS" ] || valid_due_days "$IC_LHA_DUE_DAYS" \
    || die "--ic-lha-due-days must be a whole number of days between 1 and 365 (got '$IC_LHA_DUE_DAYS')"
[ -z "$IC_MGMT_RESPONSE_DUE_DAYS" ] || valid_due_days "$IC_MGMT_RESPONSE_DUE_DAYS" \
    || die "--ic-mgmt-response-due-days must be a whole number of days between 1 and 365 (got '$IC_MGMT_RESPONSE_DUE_DAYS')"

# Docker tag grammar.
[ -z "$IMAGE_TAG" ] || [[ "$IMAGE_TAG" =~ ^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$ ]] \
    || die "--image-tag '$IMAGE_TAG' is not a valid Docker tag"

[ -d "$TEMPLATE_DIR" ] || die "template directory not found: $TEMPLATE_DIR"

# Evidence silo root. Required: there is no default, so a tenant can never
# silently land in a developer checkout or a path that a prod deploy does not
# back up. Each tenant gets <root>/<slug> under it.
[ -n "${TENANT_STORAGE_ROOT:-}" ] \
    || die "TENANT_STORAGE_ROOT is not set. Set it to the absolute storage root, e.g. /var/data/auditsphere/storage"
case "$TENANT_STORAGE_ROOT" in
    /*) ;;
    *) die "TENANT_STORAGE_ROOT must be an absolute path (got '$TENANT_STORAGE_ROOT')" ;;
esac
STORAGE_DIR="${TENANT_STORAGE_ROOT%/}/${SLUG}"

if [ "$IS_LOCAL" = true ]; then
    PROTO="http"
    DOMAIN_HOST="${SLUG}.localhost"
    API_DOMAIN_HOST="localhost"
else
    PROTO="https"
    DOMAIN_HOST="${SLUG}.${BASE_DOMAIN}"
    API_DOMAIN_HOST="api-${SLUG}.${BASE_DOMAIN}"
fi

# ── Preflight ────────────────────────────────────────────────────────────────
# Both Dockerfiles copy pre-built artefacts rather than compiling, so a build
# with a missing binary "succeeds" and the tenant only fails at container
# start. Checked before anything is written, and fatal unless --no-preflight.

echo "==> Preflight"
MISSING=()
for svc in auth audit master risk analytics; do
    [ -f "$ROOT_DIR/backend/${svc}-service/${svc}" ] \
        || MISSING+=("backend/${svc}-service/${svc}")
done
[ -d "$ROOT_DIR/frontend/.output" ] || MISSING+=("frontend/.output")
if [ ${#MISSING[@]} -gt 0 ]; then
    for m in "${MISSING[@]}"; do echo "    ! missing prebuilt artifact: $m" >&2; done
    [ "$PREFLIGHT" = false ] \
        || die "prebuilt artifacts missing: ${MISSING[*]} — build the Go binaries for linux/amd64 and run 'npm run build' in frontend/ (or pass --no-preflight to render config only)"
    echo "    --no-preflight: rendering anyway; the stack cannot start until they exist"
else
    echo "    prebuilt artifacts present"
fi
docker network inspect rb_audit_network >/dev/null 2>&1 \
    || echo "    ! the rb_audit_network bridge does not exist — start the control-plane stack (backend/docker-compose.yml) first"

# ── Registry ─────────────────────────────────────────────────────────────────

mkdir -p "$TENANTS_DIR"
if [ ! -f "$REGISTRY" ]; then
    printf 'slug\tclient_name\tdomain\tapi_domain\tfe_port\tkong_port\tkong_admin_port\tredis_db\tcreated_at\n' > "$REGISTRY"
fi

EXISTING=$(registry_field "$SLUG" 1)
TENANT_DIR="$TENANTS_DIR/$SLUG"

# Explicitly requested ports are validated before defaults are filled in, so
# only a request can be rejected. A tenant's own registered port never counts
# as "in use on the OS" against itself: its stack may well be running.
check_requested_port() {  # check_requested_port <flag> <port> <registry-column>
    local flag=$1 p=$2 col=$3 holder
    [[ "$p" =~ ^[1-9][0-9]{3,4}$ ]] && [ "$p" -ge 1024 ] && [ "$p" -le 65535 ] \
        || die "$flag '$p' must be a port number between 1024 and 65535"
    holder=$(registry_holder "$p" "$col" "$SLUG")
    [ -z "$holder" ] \
        || die "$flag $p is already allocated to tenant '$holder' in backend/tenants/registry.tsv"
    if [ "$p" != "$(registry_field "$SLUG" "$col")" ] && port_in_use_os "$p"; then
        die "$flag $p is already in use on this host (something is listening on it)"
    fi
}
[ -z "$FE_PORT" ]         || check_requested_port --fe-port "$FE_PORT" 5
[ -z "$KONG_PORT" ]       || check_requested_port --kong-port "$KONG_PORT" 6
[ -z "$KONG_ADMIN_PORT" ] || check_requested_port --kong-admin-port "$KONG_ADMIN_PORT" 7

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
[ "$FE_PORT" != "$KONG_PORT" ] && [ "$FE_PORT" != "$KONG_ADMIN_PORT" ] && [ "$KONG_PORT" != "$KONG_ADMIN_PORT" ] \
    || die "frontend, Kong and Kong admin ports must all differ (got $FE_PORT/$KONG_PORT/$KONG_ADMIN_PORT)"

# The CORS origin must match what the browser sends in the Origin header,
# which includes the port on localhost (there is no Nginx vhost in front of a
# local tenant) but not in production, where 443 is implicit.
if [ "$IS_LOCAL" = true ]; then
    TENANT_ORIGIN="${PROTO}://${DOMAIN_HOST}:${FE_PORT}"
else
    TENANT_ORIGIN="${PROTO}://${DOMAIN_HOST}"
fi

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

# Image tag: explicit flag > tag already in .env > a new versioned tag. Images
# are named per tenant (rb_audit_tenant_<slug>_*), so building one tenant can
# never replace the images another tenant runs, and the previous tag stays on
# disk for a rollback (put the old TENANT_IMAGE_TAG back in .env, `up -d`).
IMAGE_TAG=${IMAGE_TAG:-$(read_env_value TENANT_IMAGE_TAG)}
if [ -z "$IMAGE_TAG" ]; then
    IMAGE_TAG="$(date -u +%Y%m%d%H%M%S)-$(git -C "$ROOT_DIR" rev-parse --short=7 HEAD 2>/dev/null || echo nogit)"
fi

# Internal control thresholds: explicit flag > value already in .env > default.
# Keeping the .env value makes a plain --force regeneration a no-op for them.
# audit-service copies these into its settings table at migrate time only when
# the row is missing, so on an existing tenant a new value here does not change
# the live setting — ADMIN edits it in the app.
IC_LHA_DUE_DAYS=${IC_LHA_DUE_DAYS:-$(read_env_value TENANT_IC_LHA_DUE_DAYS)}
IC_LHA_DUE_DAYS=${IC_LHA_DUE_DAYS:-$DEFAULT_IC_LHA_DUE_DAYS}
IC_MGMT_RESPONSE_DUE_DAYS=${IC_MGMT_RESPONSE_DUE_DAYS:-$(read_env_value TENANT_IC_MGMT_RESPONSE_DUE_DAYS)}
IC_MGMT_RESPONSE_DUE_DAYS=${IC_MGMT_RESPONSE_DUE_DAYS:-$DEFAULT_IC_MGMT_RESPONSE_DUE_DAYS}
valid_due_days "$IC_LHA_DUE_DAYS" \
    || die "TENANT_IC_LHA_DUE_DAYS in $ENV_FILE is not 1-365 ('$IC_LHA_DUE_DAYS'); fix it or pass --ic-lha-due-days"
valid_due_days "$IC_MGMT_RESPONSE_DUE_DAYS" \
    || die "TENANT_IC_MGMT_RESPONSE_DUE_DAYS in $ENV_FILE is not 1-365 ('$IC_MGMT_RESPONSE_DUE_DAYS'); fix it or pass --ic-mgmt-response-due-days"

# SMTP credentials are owned by provision-tenant-email.sh. Carry them over
# verbatim, otherwise regenerating the stack would silently stop the tenant's
# email.
SMTP_LINES=""
[ -f "$ENV_FILE" ] && SMTP_LINES=$(grep '^TENANT_SMTP_' "$ENV_FILE" || true)

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

# Internal control assessment thresholds (days). Not secrets. Passed to
# audit-service as IC_LHA_DUE_DAYS / IC_MGMT_RESPONSE_DUE_DAYS and seeded into
# its settings table at migrate time if missing; ADMIN edits them in the app.
TENANT_IC_LHA_DUE_DAYS=$IC_LHA_DUE_DAYS
TENANT_IC_MGMT_RESPONSE_DUE_DAYS=$IC_MGMT_RESPONSE_DUE_DAYS

# Tag of this tenant's own images (rb_audit_tenant_${SLUG}_*). Not a secret.
# onboard-tenant.sh writes a new one on every run; set it back to roll back.
TENANT_IMAGE_TAG=$IMAGE_TAG
ENVEOF
[ -n "$SMTP_LINES" ] && printf '%s\n' "$SMTP_LINES" >> "$ENV_FILE"
chmod 600 "$ENV_FILE"
umask 022

# ── Render ───────────────────────────────────────────────────────────────────
# Secrets are rendered as compose variable references, not literals, so the
# generated compose file stays safe to commit. `:?` makes compose fail loudly
# if the stack is started without --env-file rather than booting with an empty
# secret.

# Values interpolated into the sed replacement side must be escaped: a bare `&`
# expands to the whole match, so "Ernst & Young" would render as
# "Ernst __CLIENT_NAME__ Young". `|` is the delimiter and `\` starts an escape.
sed_escape() { printf '%s' "$1" | sed -e 's/[\\&|]/\\&/g'; }

CLIENT_NAME_ESC=$(sed_escape "$CLIENT_NAME")
STORAGE_DIR_ESC=$(sed_escape "$STORAGE_DIR")
ENV_HINT="start this stack with --env-file $(basename "$ENV_FILE")"

render() {  # render <template> <destination>
    sed \
        -e "s|__SLUG__|${SLUG}|g" \
        -e "s|__CLIENT_NAME__|${CLIENT_NAME_ESC}|g" \
        -e "s|__TENANT_ORIGIN__|${TENANT_ORIGIN}|g" \
        -e "s|__DOMAIN__|${DOMAIN_HOST}|g" \
        -e "s|__API_DOMAIN__|${API_DOMAIN_HOST}|g" \
        -e "s|__PROTO__|${PROTO}|g" \
        -e "s|__FE_PORT__|${FE_PORT}|g" \
        -e "s|__KONG_PORT__|${KONG_PORT}|g" \
        -e "s|__KONG_ADMIN_PORT__|${KONG_ADMIN_PORT}|g" \
        -e "s|__REDIS_DB__|${REDIS_DB}|g" \
        -e "s|__STORAGE_DIR__|${STORAGE_DIR_ESC}|g" \
        -e "s|__IMAGE_TAG__|\${TENANT_IMAGE_TAG:?TENANT_IMAGE_TAG missing — ${ENV_HINT}}|g" \
        -e "s|__JWT_SECRET__|\${TENANT_JWT_SECRET:?TENANT_JWT_SECRET missing — ${ENV_HINT}}|g" \
        -e "s|__SIGNATURE_KEY__|\${TENANT_SIGNATURE_KEY:?TENANT_SIGNATURE_KEY missing — ${ENV_HINT}}|g" \
        "$1" > "$2"
}

render "$TEMPLATE_DIR/docker-compose.tenant.yml.tpl" "$TENANT_DIR/docker-compose.yml"
render "$TEMPLATE_DIR/kong-tenant.yml.tpl" "$TENANT_DIR/kong/kong.yml"

# ── Record ───────────────────────────────────────────────────────────────────

TMP_REGISTRY=$(mktemp)
awk -F'\t' -v s="$SLUG" 'NR==1 || $1!=s' "$REGISTRY" > "$TMP_REGISTRY"
printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
    "$SLUG" "$CLIENT_NAME" "$DOMAIN_HOST" "$API_DOMAIN_HOST" \
    "$FE_PORT" "$KONG_PORT" "$KONG_ADMIN_PORT" "$REDIS_DB" "$CREATED_AT" \
    >> "$TMP_REGISTRY"
mv "$TMP_REGISTRY" "$REGISTRY"

# ── Summary ──────────────────────────────────────────────────────────────────

cat <<SUMMARY

==============================================================
 Tenant stack generated: $CLIENT_NAME ($SLUG)
==============================================================
 Directory     : backend/tenants/$SLUG
 Frontend      : $TENANT_ORIGIN  -> 127.0.0.1:$FE_PORT
 API Gateway   : $PROTO://$API_DOMAIN_HOST  -> 127.0.0.1:$KONG_PORT
 Kong Admin    : 127.0.0.1:$KONG_ADMIN_PORT (loopback only)
 Redis DB      : $REDIS_DB
 Storage silo  : $STORAGE_DIR
 Databases     : rb_audit_{auth,audit,master,risk}_$SLUG
 IC thresholds : LHA $IC_LHA_DUE_DAYS days, management response $IC_MGMT_RESPONSE_DUE_DAYS days
 Images        : rb_audit_tenant_${SLUG}_*:$IMAGE_TAG

 Start it:
   docker compose --project-directory backend/tenants/$SLUG \\
     --env-file backend/tenants/$SLUG/.env up -d --build
==============================================================

SUMMARY
