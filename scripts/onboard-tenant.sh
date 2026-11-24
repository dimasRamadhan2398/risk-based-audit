#!/bin/bash
# =============================================================================
# AuditSphere Automated Tenant Onboarding Script
#
# Usage:
#   ./scripts/onboard-tenant.sh [options] [--] <tenant_slug> <client_name> [frontend_port] [kong_port]
#
# Examples:
#   ./scripts/onboard-tenant.sh accenture "Accenture" 3014 8094 --local --empty-data
#   ./scripts/onboard-tenant.sh --progress --domain bai.auditsphere.app \
#       --admin-name "Siti CAE" --admin-email cae@bai.co.id -- bai "PT Bank Arta Indonesia"
#   ./scripts/onboard-tenant.sh --dry-run --progress --local -- demo "Demo Client"
#
# Options:
#   --local              Target <slug>.localhost over http; skip email, Nginx and Certbot
#   --domain <fqdn>      Tenant frontend domain. Must be <slug>.<base-domain>; the API
#                        is served at api-<slug>.<base-domain>. Production only.
#   --base-domain <fqdn> Parent domain (default: $BASE_DOMAIN or auditsphere.app)
#   --fe-port <n>        Frontend host port (same as the 3rd positional argument)
#   --api-port <n>       Kong (API) host port (alias --kong-port; 4th positional)
#                        Ports are optional. A requested port that another tenant
#                        holds or that is in use on this host is REJECTED.
#   --admin-name <name>  Full name of the tenant's first admin (username stays "admin")
#   --admin-email <addr> Email of the tenant's first admin
#                        The initial admin password is generated here (random, 20
#                        chars) and printed ONLY in the final @@RESULT line.
#   --empty-data         Migrate only, no demo records (default)
#   --with-demo-data     Also run the demo seeders
#   --no-email           Do not create a Resend sending key (step 6 is skipped).
#                        Alias: --skip-email
#   --skip-start         Provision everything but leave the stack stopped
#   --health-timeout <s> Seconds step 8 waits for Kong + auth-service (default 120)
#   --allow-no-tls       Production: report a DNS/Certbot problem as `skipped` and
#                        finish on plain HTTP instead of failing step 9
#   --progress           Also print machine-readable @@PROGRESS lines (see
#                        docs/provisioning-job-contract.md)
#   --dry-run            Validate, then emit the full step sequence with short
#                        sleeps and a fake @@RESULT. No side effects at all.
#   --dry-run-fail <key> With --dry-run: simulate a failure at that step (UI testing)
#   --ic-lha-due-days <n>
#                        Internal control: audit report (LHA) due days, 1-365.
#                        Default: the tenant's current value, else 30.
#   --ic-mgmt-response-due-days <n>
#                        Internal control: management response due days, 1-365.
#                        Default: the tenant's current value, else 14.
#   Both are written to the tenant .env and seeded into audit-service's
#   settings table by `audit migrate up` only if the setting is missing, so
#   re-onboarding with a new value does not change an existing tenant's setting.
#   --                   End of options; everything after it is positional.
#
# Steps (fixed, see docs/provisioning-job-contract.md):
#   1 generate   2 databases   3 build   4 migrate   5 storage
#   6 email      7 start       8 health  9 ingress
#
# Output protocol (with --progress), one JSON object per line on stdout:
#   @@PROGRESS {"key":"build","index":3,"total":9,"status":"running","message":"..."}
#   status: running | done | skipped | failed. On error the current step is
#   reported `failed` and the script exits non-zero.
# The final @@RESULT line is printed with or without --progress, because it is
# the only place the generated admin password appears. Never log it elsewhere.
# =============================================================================

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REGISTRY="$ROOT_DIR/backend/tenants/registry.tsv"

# ── Steps and the progress protocol ──────────────────────────────────────────

STEP_KEYS=(generate databases build migrate storage email start health ingress)
STEP_TITLES=(
    "Generate tenant stack"
    "Create databases"
    "Build images"
    "Migrate and seed"
    "Create storage"
    "Provision email domain"
    "Start containers"
    "Wait for services to be healthy"
    "Configure nginx and TLS"
)
TOTAL_STEPS=${#STEP_KEYS[@]}

# Known before full argument parsing, so even an argument error is reported
# in the protocol and labelled as a dry run.
PROGRESS=false
DRY_RUN=false
for arg in "$@"; do
    case $arg in
        --progress) PROGRESS=true ;;
        --dry-run)  DRY_RUN=true ;;
        --) break ;;
    esac
done
DRY_LABEL=""
[ "$DRY_RUN" = true ] && DRY_LABEL="[dry run] "

CUR_KEY=""
CUR_INDEX=0
CUR_T0=0
LAST_INDEX=0
FAILED_EMITTED=false
TMP_DIR=""

# JSON string literal: strip control characters other than \t \n \r, then
# escape backslash, quote and those three.
json_str() {
    local s
    s=$(printf '%s' "$1" | tr -d '\000-\010\013\014\016-\037')
    s=${s//\\/\\\\}
    s=${s//\"/\\\"}
    s=${s//$'\t'/\\t}
    s=${s//$'\n'/\\n}
    s=${s//$'\r'/\\r}
    printf '"%s"' "$s"
}

json_or_null() {
    if [ -n "$1" ]; then json_str "$1"; else printf 'null'; fi
}

step_no() {  # step_no <key> -> 1-based index
    local i
    for i in "${!STEP_KEYS[@]}"; do
        if [ "${STEP_KEYS[$i]}" = "$1" ]; then echo $((i + 1)); return 0; fi
    done
    echo 0
}

emit_progress() {  # emit_progress <key> <index> <status> <message>
    if [ "$PROGRESS" = true ]; then
        printf '@@PROGRESS {"key":"%s","index":%d,"total":%d,"status":"%s","message":%s}\n' \
            "$1" "$2" "$TOTAL_STEPS" "$3" "$(json_str "${DRY_LABEL}$4")"
    fi
}

step_header() {  # step_header <index> <text>
    echo ""
    echo "==> ${DRY_LABEL}[$1/$TOTAL_STEPS] ${STEP_TITLES[$1 - 1]}: $2"
}

step_begin() {  # step_begin <key> <message>
    CUR_KEY=$1
    CUR_INDEX=$(step_no "$1")
    CUR_T0=$SECONDS
    LAST_INDEX=$CUR_INDEX
    step_header "$CUR_INDEX" "$2"
    emit_progress "$CUR_KEY" "$CUR_INDEX" running "$2"
}

step_done() {  # step_done <message>
    echo "    -> done: $1"
    emit_progress "$CUR_KEY" "$CUR_INDEX" done "$1"
    CUR_KEY=""
}

step_skip() {  # step_skip <key> <reason>   (the step may or may not have begun)
    local idx
    idx=$(step_no "$1")
    if [ "$CUR_KEY" != "$1" ]; then
        step_header "$idx" "skipped"
    fi
    LAST_INDEX=$idx
    echo "    -> skipped: $2"
    emit_progress "$1" "$idx" skipped "$2"
    CUR_KEY=""
}

step_fail() {  # step_fail <message> — the current step, else the next one due
    local key=$CUR_KEY idx=$CUR_INDEX
    if [ -z "$key" ]; then
        idx=$((LAST_INDEX + 1))
        [ "$idx" -le "$TOTAL_STEPS" ] || idx=$TOTAL_STEPS
        key=${STEP_KEYS[$idx - 1]}
    fi
    echo "    !! ${DRY_LABEL}step $idx/$TOTAL_STEPS ($key) FAILED: $1" >&2
    emit_progress "$key" "$idx" failed "$1"
    FAILED_EMITTED=true
    CUR_KEY=""
}

# die from the main shell only: inside $(...) the parent's ERR trap reports.
die() {
    echo "Error: $*" >&2
    if [ "$BASH_SUBSHELL" -eq 0 ] && [ "$FAILED_EMITTED" != true ]; then
        step_fail "$*"
    fi
    exit 1
}

on_err() {  # on_err <rc> <line> <command>
    local rc=$1 line=$2 cmd=$3
    if [ "$BASH_SUBSHELL" -ne 0 ] || [ "$FAILED_EMITTED" = true ]; then
        return 0
    fi
    # BASH_COMMAND is the unexpanded source text, so it never carries the
    # generated password.
    cmd=${cmd:0:200}
    echo "Error: command failed (exit $rc) at line $line: $cmd" >&2
    step_fail "command failed (exit $rc): $cmd"
}

on_exit() {
    local rc=$?
    if [ "$rc" -ne 0 ] && [ "$FAILED_EMITTED" != true ]; then
        step_fail "onboarding aborted (exit $rc)"
    fi
    if [ -n "$TMP_DIR" ]; then rm -rf "$TMP_DIR"; fi
}

trap 'on_err $? $LINENO "$BASH_COMMAND"' ERR
trap on_exit EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# ── Arguments ────────────────────────────────────────────────────────────────

IS_LOCAL=false
EMPTY_DATA=true
SKIP_START=false
PROVISION_EMAIL=true
ALLOW_NO_TLS=false
FE_PORT_OPT=""
API_PORT_OPT=""
DOMAIN_OPT=""
BASE_DOMAIN_OPT=""
ADMIN_NAME=""
ADMIN_EMAIL=""
HEALTH_TIMEOUT=120
DRY_RUN_FAIL=""
DRY_RUN_DELAY=${DRY_RUN_DELAY:-0.5}
IC_LHA_DUE_DAYS=""
IC_MGMT_RESPONSE_DUE_DAYS=""

set_opt() {  # set_opt <flag> <value>
    case $1 in
        --fe-port)                   FE_PORT_OPT=$2 ;;
        --api-port|--kong-port)      API_PORT_OPT=$2 ;;
        --domain)                    DOMAIN_OPT=$2 ;;
        --base-domain)               BASE_DOMAIN_OPT=$2 ;;
        --admin-name)                ADMIN_NAME=$2 ;;
        --admin-email)               ADMIN_EMAIL=$2 ;;
        --health-timeout)            HEALTH_TIMEOUT=$2 ;;
        --dry-run-fail)              DRY_RUN_FAIL=$2 ;;
        --ic-lha-due-days)           IC_LHA_DUE_DAYS=$2 ;;
        --ic-mgmt-response-due-days) IC_MGMT_RESPONSE_DUE_DAYS=$2 ;;
        *) die "unknown option: $1" ;;
    esac
}

POSITIONAL=()
while [ $# -gt 0 ]; do
    case $1 in
        --local)          IS_LOCAL=true ;;
        --with-demo-data) EMPTY_DATA=false ;;
        --empty-data)     EMPTY_DATA=true ;;
        --skip-start)     SKIP_START=true ;;
        --no-email|--skip-email) PROVISION_EMAIL=false ;;
        --allow-no-tls)   ALLOW_NO_TLS=true ;;
        --progress|--dry-run) ;;  # read above
        --fe-port|--api-port|--kong-port|--domain|--base-domain|--admin-name|--admin-email|--health-timeout|--dry-run-fail|--ic-lha-due-days|--ic-mgmt-response-due-days)
            [ $# -ge 2 ] || die "$1 needs a value"
            set_opt "$1" "$2"
            shift ;;
        --*=*)            set_opt "${1%%=*}" "${1#*=}" ;;
        --)               shift; POSITIONAL+=("$@"); break ;;
        -*)               die "unknown option: $1" ;;
        *)                POSITIONAL+=("$1") ;;
    esac
    shift
done
set -- ${POSITIONAL[@]+"${POSITIONAL[@]}"}

[ $# -ge 2 ] || die "usage: onboard-tenant.sh [options] [--] <tenant_slug> <client_name> [frontend_port] [kong_port]"
SLUG=$1
CLIENT_NAME=$2

# Legacy positional ports: <slug> <name> <fe> <kong>, or an ignored third
# argument (the retired Google Drive folder id) followed by the ports.
POS_FE="" POS_KONG=""
if [[ "${3:-}" =~ ^[0-9]+$ ]]; then
    POS_FE=$3; POS_KONG=${4:-}
elif [[ "${4:-}" =~ ^[0-9]+$ ]]; then
    POS_FE=$4; POS_KONG=${5:-}
fi
[[ "$POS_KONG" =~ ^[0-9]+$ ]] || POS_KONG=""
FE_PORT=${FE_PORT_OPT:-$POS_FE}
KONG_PORT=${API_PORT_OPT:-$POS_KONG}

# ── [1/9] generate, part 1: validate (no side effects yet) ───────────────────

step_begin generate "Validating input and checking prerequisites"

[[ "$SLUG" =~ ^[a-z][a-z0-9]{1,30}$ ]] \
    || die "slug '$SLUG' is invalid — use 2-31 chars, lowercase letters and digits, starting with a letter"
[ ${#CLIENT_NAME} -ge 2 ] && [ ${#CLIENT_NAME} -le 120 ] \
    || die "client name must be 2-120 characters"
[[ "$CLIENT_NAME" != *[[:cntrl:]]* ]] || die "client name must not contain control characters"
for p in "$FE_PORT" "$KONG_PORT"; do
    if [ -n "$p" ] && { ! [[ "$p" =~ ^[1-9][0-9]{3,4}$ ]] || [ "$p" -gt 65535 ]; }; then
        die "port '$p' must be a number between 1024 and 65535"
    fi
done
if [ -n "$ADMIN_EMAIL" ]; then
    [[ "$ADMIN_EMAIL" =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}$ ]] && [ ${#ADMIN_EMAIL} -le 254 ] \
        || die "--admin-email '$ADMIN_EMAIL' is not a valid email address"
fi
if [ -n "$ADMIN_NAME" ]; then
    [ ${#ADMIN_NAME} -le 120 ] || die "--admin-name must be at most 120 characters"
    [[ "$ADMIN_NAME" != *[[:cntrl:]]* ]] || die "--admin-name must not contain control characters"
fi
[[ "$HEALTH_TIMEOUT" =~ ^[1-9][0-9]{0,3}$ ]] && [ "$HEALTH_TIMEOUT" -le 1800 ] \
    || die "--health-timeout must be 1-1800 seconds (got '$HEALTH_TIMEOUT')"
valid_due_days() { [[ "$1" =~ ^[1-9][0-9]{0,2}$ ]] && [ "$1" -le 365 ]; }
[ -z "$IC_LHA_DUE_DAYS" ] || valid_due_days "$IC_LHA_DUE_DAYS" \
    || die "--ic-lha-due-days must be a whole number of days between 1 and 365 (got '$IC_LHA_DUE_DAYS')"
[ -z "$IC_MGMT_RESPONSE_DUE_DAYS" ] || valid_due_days "$IC_MGMT_RESPONSE_DUE_DAYS" \
    || die "--ic-mgmt-response-due-days must be a whole number of days between 1 and 365 (got '$IC_MGMT_RESPONSE_DUE_DAYS')"
if [ -n "$DRY_RUN_FAIL" ]; then
    [ "$DRY_RUN" = true ] || die "--dry-run-fail only works together with --dry-run"
    [ "$(step_no "$DRY_RUN_FAIL")" -gt 0 ] || die "--dry-run-fail: unknown step '$DRY_RUN_FAIL' (one of: ${STEP_KEYS[*]})"
fi

# Domain: --domain is <slug>.<base>; the API always lives at api-<slug>.<base>,
# which is what the generator renders and the wildcard DNS record covers.
FQDN_RE='^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)+$'
BASE_DOMAIN_EFF=${BASE_DOMAIN_OPT:-${BASE_DOMAIN:-auditsphere.app}}
if [ -n "$DOMAIN_OPT" ]; then
    [ "$IS_LOCAL" = false ] || die "--domain cannot be combined with --local (local tenants are <slug>.localhost)"
    DOMAIN_OPT=$(printf '%s' "$DOMAIN_OPT" | tr '[:upper:]' '[:lower:]')
    [[ "$DOMAIN_OPT" == "$SLUG".* ]] \
        || die "--domain '$DOMAIN_OPT' must be <slug>.<base-domain>, i.e. start with '$SLUG.'"
    DERIVED_BASE=${DOMAIN_OPT#"$SLUG".}
    if [ -n "$BASE_DOMAIN_OPT" ] && [ "$BASE_DOMAIN_OPT" != "$DERIVED_BASE" ]; then
        die "--domain '$DOMAIN_OPT' does not match --base-domain '$BASE_DOMAIN_OPT'"
    fi
    BASE_DOMAIN_EFF=$DERIVED_BASE
fi
[[ "$BASE_DOMAIN_EFF" =~ $FQDN_RE ]] || die "base domain '$BASE_DOMAIN_EFF' is not a valid domain name"

if [ "$IS_LOCAL" = true ]; then
    PROTO="http"
    STORAGE_DIR="$ROOT_DIR/backend/audit-service/uploads/${SLUG}"
else
    PROTO="https"
    STORAGE_DIR="/var/data/auditsphere/storage/${SLUG}"
fi

ADMIN_USERNAME="admin"   # the auth seeder's fixed username

emit_result() {  # emit_result <frontend_url> <api_url> <admin_email|""> <password|"">
    printf '@@RESULT {"slug":%s,"frontend_url":%s,"api_url":%s,"admin_username":%s,"admin_email":%s,"admin_initial_password":%s}\n' \
        "$(json_str "$SLUG")" "$(json_str "$1")" "$(json_str "$2")" \
        "$(json_str "$ADMIN_USERNAME")" "$(json_or_null "$3")" "$(json_or_null "$4")"
}

# Random initial admin password: 20 chars from the same look-alike-free
# alphabet as auth-service's temporary passwords, always with an upper, a
# lower and a digit (~115 bits).
gen_password() {
    local pool pw
    while :; do
        pool=$(openssl rand -base64 96 | tr -dc 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789')
        pw=${pool:0:20}
        if [ ${#pw} -eq 20 ] && [[ $pw =~ [[:upper:]] ]] && [[ $pw =~ [[:lower:]] ]] && [[ $pw =~ [0-9] ]]; then
            printf '%s' "$pw"
            return 0
        fi
    done
}

# ── Dry run ──────────────────────────────────────────────────────────────────
# Full event sequence, no side effects: nothing is written, no container,
# database, DNS or Resend call is touched. --next-ports only reads.

if [ "$DRY_RUN" = true ]; then
    if [ -z "$FE_PORT" ] || [ -z "$KONG_PORT" ]; then
        NEXT_PORTS=$(bash "$SCRIPT_DIR/generate-tenant-stack.sh" --next-ports)
        [ -n "$FE_PORT" ]   || FE_PORT=$(sed -E 's/.*"fe_port":([0-9]+).*/\1/' <<<"$NEXT_PORTS")
        [ -n "$KONG_PORT" ] || KONG_PORT=$(sed -E 's/.*"api_port":([0-9]+).*/\1/' <<<"$NEXT_PORTS")
    fi
    if [ "$IS_LOCAL" = true ]; then
        DOMAIN="${SLUG}.localhost:${FE_PORT}"
        API_DOMAIN="localhost:${KONG_PORT}"
    else
        DOMAIN="${SLUG}.${BASE_DOMAIN_EFF}"
        API_DOMAIN="api-${SLUG}.${BASE_DOMAIN_EFF}"
    fi
    DRY_TAG="$(date -u +%Y%m%d%H%M%S)-dryrun"

    dry_step() {  # dry_step <key> <running message> <done message>
        if [ "$CUR_KEY" != "$1" ]; then step_begin "$1" "$2"; fi
        sleep "$DRY_RUN_DELAY"
        if [ "$DRY_RUN_FAIL" = "$1" ]; then
            die "simulated failure at step '$1' (--dry-run-fail)"
        fi
        step_done "$3"
    }
    dry_skip() { sleep "$DRY_RUN_DELAY"; step_skip "$1" "$2"; }

    dry_step generate "Validating input and checking prerequisites" \
        "would write backend/tenants/${SLUG} (frontend port ${FE_PORT}, API port ${KONG_PORT})"
    dry_step databases "Creating rb_audit_{auth,audit,master,risk}_${SLUG}" "would create 4 databases"
    dry_step build "Building 6 images (tag ${DRY_TAG})" "would build 6 images"
    dry_step migrate "Running migrations and seeders" \
        "would migrate 4 databases and create admin '${ADMIN_USERNAME}'$([ -n "$ADMIN_EMAIL" ] && echo " <${ADMIN_EMAIL}>")$([ "$EMPTY_DATA" = false ] && echo " plus demo data")"
    dry_step storage "Creating ${STORAGE_DIR}" "would create the evidence storage silo"
    if [ "$IS_LOCAL" = true ]; then
        dry_skip email "local mode: email is not provisioned"
    elif [ "$PROVISION_EMAIL" = false ]; then
        dry_skip email "email provisioning disabled for this job"
    else
        dry_step email "Creating a sending-only Resend key" "would write the tenant's Resend key to its .env"
    fi
    if [ "$SKIP_START" = true ]; then
        dry_skip start "--skip-start: stack left stopped"
        dry_skip health "stack not started"
    else
        dry_step start "Starting 7 containers" "would start 7 containers"
        dry_step health "Waiting up to ${HEALTH_TIMEOUT}s for Kong and auth-service" "would wait for Kong and auth-service to report healthy"
    fi
    if [ "$IS_LOCAL" = true ]; then
        dry_skip ingress "local mode: no nginx vhost or TLS"
    else
        dry_step ingress "Writing nginx vhosts for ${DOMAIN} and ${API_DOMAIN}" \
            "would write nginx vhosts and issue a TLS certificate for ${DOMAIN} and ${API_DOMAIN}"
    fi

    FAKE_PW="dryrun-$(openssl rand -hex 6 2>/dev/null || printf '%06d%06d' "$RANDOM" "$RANDOM")"
    echo ""
    echo "==> [dry run] complete — nothing was created, written or started."
    emit_result "${PROTO}://${DOMAIN}" "${PROTO}://${API_DOMAIN}" "$ADMIN_EMAIL" "$FAKE_PW"
    exit 0
fi

# ── [1/9] generate, part 2: preflight (still no side effects) ────────────────
# Everything a later step needs is checked here, so a missing tool or artifact
# fails the job before any database, file or container exists.

TMP_DIR=$(mktemp -d)
echo "    -> preflight"
for bin in docker openssl awk sed; do
    command -v "$bin" >/dev/null 2>&1 || die "'$bin' is required on this host"
done
docker info >/dev/null 2>&1 || die "cannot reach the Docker daemon (is it running, and may this user use it?)"
docker compose version >/dev/null 2>&1 || die "the docker compose v2 plugin is required"

MISSING=()
for svc in auth audit master risk analytics; do
    [ -f "$ROOT_DIR/backend/${svc}-service/${svc}" ] || MISSING+=("backend/${svc}-service/${svc}")
done
[ -d "$ROOT_DIR/frontend/.output" ] || MISSING+=("frontend/.output")
[ ${#MISSING[@]} -eq 0 ] \
    || die "prebuilt artifacts missing: ${MISSING[*]}. Build the Go binaries for linux/amd64 and run 'npm run build' in frontend/ first."

docker network inspect rb_audit_network >/dev/null 2>&1 \
    || die "the rb_audit_network bridge does not exist — start the control-plane stack first"

PG_CONTAINER=""
for candidate in rb_audit_postgres rb_audit_postgres_dev; do
    if docker ps --format '{{.Names}}' | grep -qx "$candidate"; then
        PG_CONTAINER="$candidate"
        break
    fi
done
[ -n "$PG_CONTAINER" ] \
    || die "no running Postgres container found (looked for rb_audit_postgres, rb_audit_postgres_dev). Start the control-plane stack first."

if [ "$SKIP_START" = false ]; then
    command -v curl >/dev/null 2>&1 || die "'curl' is required for the health check (step 8)"
fi
if [ "$IS_LOCAL" = false ]; then
    command -v nginx >/dev/null 2>&1 || die "nginx is not installed on this host; step 9 cannot expose the tenant"
    if ! command -v certbot >/dev/null 2>&1 && [ "$ALLOW_NO_TLS" = false ]; then
        die "certbot is not installed, so step 9 cannot issue TLS. Install it, or pass --allow-no-tls to finish on plain HTTP."
    fi
fi
echo "    -> preflight passed"

# ── [1/9] generate, part 3: render the stack ─────────────────────────────────
# Allocates the host ports and Redis index. --force makes re-onboarding an
# existing tenant safe: the generator reuses its ports and signing secrets.
# A fresh image tag per run means this build cannot replace images that any
# other tenant (or this tenant's previous release) runs.

# generate-tenant-stack.sh refuses to run without TENANT_STORAGE_ROOT and derives
# the silo as <TENANT_STORAGE_ROOT>/<slug>. Pass it the parent of STORAGE_DIR so
# the generator and this script use one path.
TENANT_STORAGE_ROOT="$(dirname "$STORAGE_DIR")"
export TENANT_STORAGE_ROOT

registry_field() { [ -f "$REGISTRY" ] && awk -F'\t' -v s="$SLUG" -v c="$1" 'NR>1 && $1==s {print $c; exit}' "$REGISTRY"; }
IS_NEW_TENANT=true
[ -z "$(registry_field 1 || true)" ] || IS_NEW_TENANT=false
[ "$IS_NEW_TENANT" = true ] || echo "    -> '$SLUG' is already registered: re-onboarding (ports, secrets and data are kept)"

IMAGE_TAG="$(date -u +%Y%m%d%H%M%S)-$(git -C "$ROOT_DIR" rev-parse --short=7 HEAD 2>/dev/null || echo nogit)"

GEN_ARGS=("$SLUG" "$CLIENT_NAME" --force --image-tag "$IMAGE_TAG")
[ "$IS_LOCAL" = true ] && GEN_ARGS+=(--local)
[ "$IS_LOCAL" = true ] || GEN_ARGS+=(--base-domain "$BASE_DOMAIN_EFF")
[ -n "$FE_PORT" ]   && GEN_ARGS+=(--fe-port "$FE_PORT")
[ -n "$KONG_PORT" ] && GEN_ARGS+=(--kong-port "$KONG_PORT")
[ -n "$IC_LHA_DUE_DAYS" ] && GEN_ARGS+=(--ic-lha-due-days "$IC_LHA_DUE_DAYS")
[ -n "$IC_MGMT_RESPONSE_DUE_DAYS" ] && GEN_ARGS+=(--ic-mgmt-response-due-days "$IC_MGMT_RESPONSE_DUE_DAYS")

GEN_ERR="$TMP_DIR/generate.err"
if ! bash "$SCRIPT_DIR/generate-tenant-stack.sh" "${GEN_ARGS[@]}" 2>"$GEN_ERR"; then
    cat "$GEN_ERR" >&2
    die "stack generation failed: $(grep '^Error: ' "$GEN_ERR" | tail -n1 | sed 's/^Error: //')"
fi
cat "$GEN_ERR" >&2

DOMAIN_HOST=$(registry_field 3)
API_DOMAIN_HOST=$(registry_field 4)
FE_PORT=$(registry_field 5)
KONG_PORT=$(registry_field 6)
REDIS_DB=$(registry_field 8)
[ -n "$FE_PORT" ] || die "tenant '$SLUG' is missing from $REGISTRY after generation"

TENANT_DIR="$ROOT_DIR/backend/tenants/${SLUG}"

# The thresholds the generator resolved (flag > existing .env > default). Plain
# day counts; nothing else is read from the secrets file.
tenant_env_value() { awk -F= -v k="$1" '$1==k {sub(/^[^=]*=/,""); print; exit}' "$TENANT_DIR/.env"; }
IC_LHA_DUE_DAYS=$(tenant_env_value TENANT_IC_LHA_DUE_DAYS)
IC_MGMT_RESPONSE_DUE_DAYS=$(tenant_env_value TENANT_IC_MGMT_RESPONSE_DUE_DAYS)

if [ "$IS_LOCAL" = true ]; then
    DOMAIN="${DOMAIN_HOST}:${FE_PORT}"
    API_DOMAIN="localhost:${KONG_PORT}"
else
    DOMAIN="$DOMAIN_HOST"
    API_DOMAIN="$API_DOMAIN_HOST"
fi

COMPOSE=(docker compose --project-directory "$TENANT_DIR" --env-file "$TENANT_DIR/.env")

echo ""
echo "  AuditSphere client onboarding: $CLIENT_NAME ($SLUG)"
echo "    Mode       : $([ "$IS_LOCAL" = true ] && echo "Localhost dev (RFC 6761)" || echo "Production")"
echo "    Domain     : $PROTO://$DOMAIN"
echo "    API domain : $PROTO://$API_DOMAIN"
echo "    Ports      : frontend $FE_PORT | Kong $KONG_PORT | Redis DB $REDIS_DB"
echo "    Storage    : $STORAGE_DIR"
echo "    Data state : $([ "$EMPTY_DATA" = true ] && echo "clean / empty" || echo "demo pre-seeded")"
echo "    IC due     : LHA ${IC_LHA_DUE_DAYS}d | management response ${IC_MGMT_RESPONSE_DUE_DAYS}d"
echo "    Images     : rb_audit_tenant_${SLUG}_*:${IMAGE_TAG}"

step_done "Stack generated in backend/tenants/${SLUG} (frontend port ${FE_PORT}, API port ${KONG_PORT}, Redis DB ${REDIS_DB})"

# ── [2/9] Create the isolated databases ──────────────────────────────────────
# analytics-service is intentionally absent: cmd/main.go starts a gin server and
# opens no database handle, so an rb_audit_analytics_<slug> database would never
# be connected to.

step_begin databases "Creating rb_audit_{auth,audit,master,risk}_${SLUG} in ${PG_CONTAINER}"

DB_CREATED=0
DB_EXISTING=0
AUTH_DB_EXISTED=false
for svc in auth audit master risk; do
    DB="rb_audit_${svc}_${SLUG}"
    # CREATE DATABASE has no IF NOT EXISTS, and re-onboarding an existing tenant
    # is a normal operation (recovery, config change), so check first.
    if docker exec -i "$PG_CONTAINER" psql -U postgres -tAc \
            "SELECT 1 FROM pg_database WHERE datname='${DB}'" | grep -q 1; then
        echo "    -> ${DB} already exists, leaving it untouched"
        DB_EXISTING=$((DB_EXISTING + 1))
        [ "$svc" != auth ] || AUTH_DB_EXISTED=true
    else
        docker exec -i "$PG_CONTAINER" psql -U postgres -c "CREATE DATABASE ${DB};" >/dev/null
        echo "    -> created ${DB}"
        DB_CREATED=$((DB_CREATED + 1))
    fi
done
step_done "${DB_CREATED} created, ${DB_EXISTING} already existed"

# ── [3/9] Build the tenant images ────────────────────────────────────────────

step_begin build "Building 6 images (tag ${IMAGE_TAG})"
# Plain BuildKit output: one readable line per event in the job log.
BUILDKIT_PROGRESS=${BUILDKIT_PROGRESS:-plain} "${COMPOSE[@]}" build
step_done "6 images built in $((SECONDS - CUR_T0))s (tag ${IMAGE_TAG})"

# ── [4/9] Migrate and seed, inside the containers ────────────────────────────
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

step_begin migrate "Running schema migrations and seeding roles and the first admin"

run_svc() {  # run_svc <compose-service> <command...>
    local svc=$1; shift
    "${COMPOSE[@]}" run --rm --no-deps -T "$svc" "$@"
}

# Belt and braces for the job log: mask the generated password and any
# "Password: x" the seeder prints. The password is read from the environment,
# not argv, so it never shows up in `ps`.
redact_secrets() {
    REDACT_PW="${ADMIN_PASSWORD:-}" awk '{
        pw = ENVIRON["REDACT_PW"]
        if (pw != "") {
            while ((i = index($0, pw)) > 0) $0 = substr($0, 1, i - 1) "[redacted]" substr($0, i + length(pw))
        }
        gsub(/[Pp]assword:[ ]*[^ |",]+/, "Password: [redacted]")
        print
        fflush()
    }'
}

run_svc "auth-service-${SLUG}" ./auth migrate up

# The first admin. The auth seeder creates `admin` only when it is missing and
# reads ADMIN_NAME / ADMIN_EMAIL / ADMIN_PASSWORD from its environment (see
# docs/provisioning-job-contract.md, "Auth seeder interface"). An image that
# predates that would silently fall back to the shared default password, so
# its `seed --help` must mention ADMIN_PASSWORD or this step fails.
ADMIN_EXISTED=false
if [ "$AUTH_DB_EXISTED" = true ] && docker exec -i "$PG_CONTAINER" psql -U postgres -d "rb_audit_auth_${SLUG}" -tAc \
        "SELECT 1 FROM users WHERE username='${ADMIN_USERNAME}' LIMIT 1" 2>/dev/null | grep -q 1; then
    ADMIN_EXISTED=true
fi
AUTH_SEED_HELP=$(run_svc "auth-service-${SLUG}" ./auth seed --help 2>&1 || true)
if ! grep -q 'ADMIN_PASSWORD' <<<"$AUTH_SEED_HELP"; then
    if [ "$ADMIN_EXISTED" = true ]; then
        echo "    ! this auth-service image does not support ADMIN_PASSWORD; the admin already exists, continuing"
    else
        die "this auth-service image does not read ADMIN_PASSWORD, so the new admin would get the shared default password. Deploy an auth-service build whose 'auth seed --help' documents ADMIN_PASSWORD, then re-run."
    fi
fi

ADMIN_PASSWORD=""
if [ "$ADMIN_EXISTED" = false ]; then
    ADMIN_PASSWORD=$(gen_password)
fi
SEED_ENV=()
[ -z "$ADMIN_PASSWORD" ] || SEED_ENV+=(-e ADMIN_PASSWORD)
[ -z "$ADMIN_NAME" ]     || SEED_ENV+=(-e ADMIN_NAME)
[ -z "$ADMIN_EMAIL" ]    || SEED_ENV+=(-e ADMIN_EMAIL)
# Sample users (auditor, dept_head, ...) share a default password, so they are
# created only for demo tenants. Always set explicitly: when it is unset and
# ADMIN_PASSWORD is empty (admin already existed), the seeder would create them.
if [ "$EMPTY_DATA" = false ]; then SEED_DEMO_USERS=true; else SEED_DEMO_USERS=false; fi
SEED_ENV+=(-e SEED_DEMO_USERS)
# `-e NAME` without a value: compose copies it from this environment, so the
# values never appear on a command line.
SEED_DEMO_USERS="$SEED_DEMO_USERS" \
ADMIN_NAME="$ADMIN_NAME" ADMIN_EMAIL="$ADMIN_EMAIL" ADMIN_PASSWORD="$ADMIN_PASSWORD" \
    "${COMPOSE[@]}" run --rm --no-deps -T ${SEED_ENV[@]+"${SEED_ENV[@]}"} "auth-service-${SLUG}" ./auth seed 2>&1 \
    | redact_secrets

run_svc "audit-service-${SLUG}"  ./audit migrate up
run_svc "master-service-${SLUG}" ./master migrate
run_svc "risk-service-${SLUG}"   ./risk up

# Reference catalogs (standard risk factors, standard audit universe), for every
# tenant. These are product content, not demo data: without them a clean-slate
# tenant cannot build its Audit Universe.
#
# `./risk seed --catalog-only` must seed the standard catalogs only and be
# idempotent. Older images lack the flag; probe for it instead of letting
# `set -e` abort onboarding on "unknown flag". Captured first: piping straight
# into `grep -q` under pipefail can SIGPIPE compose and read as "flag missing".
CATALOG_STATUS="seeded"
RISK_SEED_HELP=$(run_svc "risk-service-${SLUG}" ./risk seed --help 2>/dev/null || true)
if grep -q -- '--catalog-only' <<<"$RISK_SEED_HELP"; then
    echo "    -> seeding reference catalogs (standard risk factors, standard audit universe)..."
    run_svc "risk-service-${SLUG}" ./risk seed --catalog-only
else
    CATALOG_STATUS="MISSING (risk image has no 'seed --catalog-only')"
    echo "    ! this risk-service image has no 'seed --catalog-only': the tenant has NO"
    echo "      standard risk-factor / audit-universe catalogs. After deploying a risk"
    echo "      image that has it, run:"
    echo "      docker compose --project-directory backend/tenants/${SLUG} --env-file backend/tenants/${SLUG}/.env \\"
    echo "        run --rm --no-deps risk-service-${SLUG} ./risk seed --catalog-only"
fi

if [ "$EMPTY_DATA" = true ]; then
    echo "    -> empty data state: demo seeders skipped (0 business records)"
else
    echo "    -> seeding demo dataset..."
    run_svc "audit-service-${SLUG}"  ./audit seed
    run_svc "master-service-${SLUG}" ./master seed
    run_svc "risk-service-${SLUG}"   ./risk seed
fi

# Kafka topic for this tenant's Audit Universe events. The broker is shared, so
# the compose file sets KAFKA_TOPIC_PREFIX=${SLUG} on risk-service and
# audit-service. Not fatal — a missing Kafka container only delays the topic
# until risk-service's relay creates it on start.
KAFKA_TOPIC="${SLUG}.audit-universe.events"
KAFKA_CONTAINER=""
for candidate in rb_audit_kafka rb_audit_kafka_dev; do
    if docker ps --format '{{.Names}}' | grep -qx "$candidate"; then
        KAFKA_CONTAINER="$candidate"
        break
    fi
done
if [ -n "$KAFKA_CONTAINER" ] && docker exec "$KAFKA_CONTAINER" kafka-topics \
        --bootstrap-server localhost:29092 --create --if-not-exists --topic "$KAFKA_TOPIC" >/dev/null 2>&1; then
    echo "    -> Kafka topic ${KAFKA_TOPIC} ready"
else
    echo "    ! could not pre-create Kafka topic ${KAFKA_TOPIC}; risk-service will create it on start"
fi

MIGRATE_MSG="Migrations applied"
if [ "$ADMIN_EXISTED" = true ]; then
    MIGRATE_MSG+="; existing admin kept, password unchanged"
else
    MIGRATE_MSG+="; admin created with a generated initial password"
fi
[ "$EMPTY_DATA" = true ] || MIGRATE_MSG+="; demo data seeded"
[ "$CATALOG_STATUS" = seeded ] || MIGRATE_MSG+="; WARNING: risk catalogs missing"
step_done "$MIGRATE_MSG"

# ── [5/9] Provision the dedicated storage silo ───────────────────────────────

step_begin storage "Creating the evidence storage silo at ${STORAGE_DIR}"
mkdir -p "$STORAGE_DIR"
chmod 755 "$STORAGE_DIR" 2>/dev/null || true
echo "    -> bind-mounted into audit-service at /root/uploads; Google Drive is disabled by policy"
step_done "Storage silo ready at ${STORAGE_DIR}"

# ── [6/9] Outbound email: a sending-only Resend key for this tenant ──────────
# Runs before the stack starts so auth-service boots with the credentials.
# Never fatal: without it the tenant falls back to the Mailtrap sandbox in
# config.yaml and sends nothing, and the step says so.

EMAIL_ENV=${AUDITSPHERE_EMAIL_ENV:-/etc/auditsphere/email.env}
EMAIL_STATUS="not provisioned"
EMAIL_RETRY="./scripts/provision-tenant-email.sh tenant ${SLUG} \"${CLIENT_NAME}\""
if [ "$IS_LOCAL" = true ]; then
    step_skip email "Local mode: email is not provisioned"
elif [ "$PROVISION_EMAIL" = false ]; then
    step_skip email "Email provisioning disabled for this job. Later: ${EMAIL_RETRY}"
elif [ ! -f "$EMAIL_ENV" ]; then
    step_skip email "${EMAIL_ENV} not found on this host, so the tenant cannot send email. Later: ${EMAIL_RETRY}"
else
    step_begin email "Creating a sending-only Resend key for ${CLIENT_NAME}"
    if bash "$SCRIPT_DIR/provision-tenant-email.sh" tenant "$SLUG" "$CLIENT_NAME" --no-restart; then
        EMAIL_STATUS="Resend (shared sending domain, tenant-scoped key)"
        step_done "Sending-only Resend key written to the tenant .env"
    else
        EMAIL_STATUS="FAILED"
        step_skip email "Email provisioning FAILED (see log); the tenant runs without email. Re-run: ${EMAIL_RETRY}"
    fi
fi

# ── [7/9] Start the tenant stack ─────────────────────────────────────────────

if [ "$SKIP_START" = true ]; then
    step_skip start "--skip-start given: the stack is left stopped"
else
    step_begin start "Starting 7 containers (5 services, Kong, frontend)"
    "${COMPOSE[@]}" up -d
    "${COMPOSE[@]}" ps
    step_done "7 containers started"
fi

# ── [8/9] Wait until Kong and auth-service are healthy ───────────────────────
# Healthy means all three of:
#   * Kong's container healthcheck (`kong health`) reports healthy
#   * auth-service answers GET /health from inside its container (busybox wget)
#   * a request through this tenant's Kong reaches auth-service: any 2xx-4xx
#     except 429 (a 401/404 is auth-service answering; 502/503 is Kong without
#     an upstream)
# plus the frontend answering on its host port, so the URL handed out works.

if [ "$SKIP_START" = true ]; then
    step_skip health "Stack not started (--skip-start)"
else
    step_begin health "Waiting up to ${HEALTH_TIMEOUT}s for Kong and auth-service"
    KONG_CONTAINER="rb_audit_${SLUG}_kong"
    AUTH_CONTAINER="rb_audit_${SLUG}_auth_service"
    HC_UA="Mozilla/5.0 (compatible; AuditSphere-onboarding-healthcheck)"
    DEADLINE=$((SECONDS + HEALTH_TIMEOUT))
    kong_state="" auth_state="" auth_health="" route_code="" fe_code=""
    while :; do
        kong_state=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$KONG_CONTAINER" 2>/dev/null || echo missing)
        auth_state=$(docker inspect -f '{{.State.Status}}' "$AUTH_CONTAINER" 2>/dev/null || echo missing)
        auth_health=fail
        if [ "$auth_state" = running ] \
                && docker exec "$AUTH_CONTAINER" wget -qO- -T 3 http://127.0.0.1:8001/health 2>/dev/null | grep -q healthy; then
            auth_health=ok
        fi
        route_code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 -A "$HC_UA" \
            "http://127.0.0.1:${KONG_PORT}/api/v1/auth/me" 2>/dev/null || true)
        fe_code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 -A "$HC_UA" \
            "http://127.0.0.1:${FE_PORT}/" 2>/dev/null || true)
        STATE_LINE="kong=${kong_state} auth-service=${auth_state}/${auth_health} kong->auth=HTTP ${route_code:-000} frontend=HTTP ${fe_code:-000}"
        echo "    .. $((SECONDS - CUR_T0))s: ${STATE_LINE}"

        if [ "$kong_state" = healthy ] && [ "$auth_health" = ok ] \
                && [[ "$route_code" =~ ^[234][0-9][0-9]$ ]] && [ "$route_code" != 429 ] \
                && [[ "$fe_code" =~ ^[234][0-9][0-9]$ ]]; then
            break
        fi
        # A container that exited will not recover by waiting.
        for c in "$kong_state" "$auth_state"; do
            if [ "$c" = exited ] || [ "$c" = dead ] || [ "$c" = missing ]; then
                docker logs --tail 30 "$AUTH_CONTAINER" 2>&1 | redact_secrets || true
                die "a tenant container is not running (${STATE_LINE}). See the auth-service log above."
            fi
        done
        if [ "$SECONDS" -ge "$DEADLINE" ]; then
            "${COMPOSE[@]}" ps || true
            echo "    -- last auth-service log lines:"
            docker logs --tail 30 "$AUTH_CONTAINER" 2>&1 | redact_secrets || true
            echo "    -- last Kong log lines:"
            docker logs --tail 15 "$KONG_CONTAINER" 2>&1 || true
            die "services not healthy after ${HEALTH_TIMEOUT}s: ${STATE_LINE}"
        fi
        sleep 3
    done
    step_done "Kong, auth-service and frontend healthy after $((SECONDS - CUR_T0))s"
fi

# ── [9/9] Expose it: host Nginx vhost + TLS (production only) ────────────────

CERTBOT_CMD="certbot --nginx -d ${DOMAIN} -d ${API_DOMAIN} --agree-tos -m admin@auditsphere.app --non-interactive --keep-until-expiring"
if [ "$IS_LOCAL" = true ]; then
    step_skip ingress "Local mode: no nginx vhost or TLS; reach the tenant at http://localhost:${FE_PORT}"
else
    step_begin ingress "Writing nginx vhosts for ${DOMAIN} and ${API_DOMAIN}"

    # Nginx layout differs by distribution: Debian/Ubuntu splits vhosts into
    # sites-available + a sites-enabled symlink, while RHEL/CentOS (which this
    # VPS runs) includes /etc/nginx/conf.d/*.conf directly.
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

    # Keep the previous vhost: a config nginx rejects must not stay in place,
    # or every later reload of every site on this host fails.
    NGINX_BACKUP=""
    if [ -f "$NGINX_CONF" ]; then
        NGINX_BACKUP="$TMP_DIR/vhost.bak"
        cp -p "$NGINX_CONF" "$NGINX_BACKUP"
    fi

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

    [ -z "$NGINX_LINK" ] || ln -sf "$NGINX_CONF" "$NGINX_LINK"
    if ! NGINX_TEST=$(nginx -t 2>&1); then
        printf '%s\n' "$NGINX_TEST"
        if [ -n "$NGINX_BACKUP" ]; then
            cp -p "$NGINX_BACKUP" "$NGINX_CONF"
        else
            rm -f "$NGINX_CONF"
            [ -z "$NGINX_LINK" ] || rm -f "$NGINX_LINK"
        fi
        die "nginx -t rejected the vhost (previous config restored): $(tail -n1 <<<"$NGINX_TEST")"
    fi
    systemctl reload nginx || die "systemctl reload nginx failed"

    # TLS. Certbot needs both names to resolve to this host for the HTTP-01
    # challenge. With the wildcard *.auditsphere.app record they resolve the
    # moment the tenant exists. getent rather than `host`: host ships in
    # bind-utils, which a minimal RHEL install lacks.
    TLS_PROBLEM=""
    if ! getent hosts "$DOMAIN" >/dev/null 2>&1 || ! getent hosts "$API_DOMAIN" >/dev/null 2>&1; then
        TLS_PROBLEM="DNS for ${DOMAIN} / ${API_DOMAIN} does not resolve yet"
    elif ! command -v certbot >/dev/null 2>&1; then
        TLS_PROBLEM="certbot is not installed on this host"
    else
        echo "    -> requesting TLS certificates via Certbot..."
        # shellcheck disable=SC2086
        if ! $CERTBOT_CMD; then
            TLS_PROBLEM="certbot could not issue the certificate (see log)"
        fi
    fi

    if [ -z "$TLS_PROBLEM" ]; then
        step_done "nginx vhosts live and TLS certificate installed for ${DOMAIN} and ${API_DOMAIN}"
    elif [ "$ALLOW_NO_TLS" = true ]; then
        PROTO="http"
        echo ""
        echo "    !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
        echo "    !! TLS NOT CONFIGURED: ${TLS_PROBLEM}."
        echo "    !! The tenant serves PLAIN HTTP until you run:"
        echo "    !!   ${CERTBOT_CMD}"
        echo "    !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
        step_skip ingress "TLS NOT CONFIGURED, plain HTTP only: ${TLS_PROBLEM}. Fix it, then run: ${CERTBOT_CMD}"
    else
        die "${TLS_PROBLEM}. nginx serves the tenant on plain HTTP only. Fix it, then run: ${CERTBOT_CMD} (or re-run with --allow-no-tls to accept HTTP)"
    fi
fi

# ── Summary ──────────────────────────────────────────────────────────────────
# No credentials here: the initial admin password is only in @@RESULT below.

echo ""
echo "=============================================================="
echo "Client $CLIENT_NAME onboarded."
echo "   Frontend URL : $PROTO://${DOMAIN}"
[ "$IS_LOCAL" = false ] || echo "   Direct Local : http://localhost:${FE_PORT}"
echo "   API Endpoint : $PROTO://${API_DOMAIN}"
echo "   Stack Dir    : backend/tenants/${SLUG}"
echo "   Images       : rb_audit_tenant_${SLUG}_*:${IMAGE_TAG}"
echo "   Data State   : $([ "$EMPTY_DATA" = true ] && echo "Empty Data / Clean Slate (0 Business Records)" || echo "Demo Pre-seeded")"
echo "   Databases    : rb_audit_{auth,audit,master,risk}_${SLUG}"
echo "   Redis DB     : ${REDIS_DB} (isolated logical index)"
echo "   Kafka        : topic ${KAFKA_TOPIC}, group audit-service-annual-plan-sync-${SLUG}"
echo "   Risk catalogs: ${CATALOG_STATUS}"
echo "   IC thresholds: LHA ${IC_LHA_DUE_DAYS} days, management response ${IC_MGMT_RESPONSE_DUE_DAYS} days"
echo "   Storage Silo : ${STORAGE_DIR}"
echo "   Email Svc    : ${EMAIL_STATUS}"
echo "   Admin        : username '${ADMIN_USERNAME}'$([ -n "$ADMIN_EMAIL" ] && echo ", ${ADMIN_EMAIL}")"
if [ "$ADMIN_EXISTED" = true ]; then
echo "                  existing account kept; its password was not changed"
else
echo "                  initial password: in the @@RESULT line below, shown once"
fi
echo "=============================================================="
echo ""
echo "Before handing this instance to the client:"
echo "   1. Give the admin the initial password over a separate channel and have"
echo "      them change it at first login."
echo "   2. Check email delivery:"
echo "      ./scripts/provision-tenant-email.sh test ${SLUG} you@example.com"
echo "   3. Back up backend/tenants/${SLUG}/.env — it holds this tenant's JWT"
echo "      signing secret and email key, and is not in git."
echo ""

emit_result "$PROTO://${DOMAIN}" "$PROTO://${API_DOMAIN}" "$ADMIN_EMAIL" "$ADMIN_PASSWORD"
