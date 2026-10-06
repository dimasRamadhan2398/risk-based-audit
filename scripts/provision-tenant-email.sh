#!/bin/bash
# =============================================================================
# AuditSphere Tenant Email Provisioning (Resend)
#
# Every tenant sends through ONE shared, verified Resend domain
# (default mail.auditsphere.app). Isolation comes from the API key, not the
# domain: each tenant gets its own sending-only key, so one tenant's key can be
# revoked or rate-limited without touching the others, and no tenant ever holds
# the full-access master key.
#
# Usage (run on the VPS, from the repo root):
#   ./scripts/provision-tenant-email.sh setup-domain
#       One-time. Registers the sending domain in Resend and prints the DNS
#       records to add at the registrar (Rumahweb).
#   ./scripts/provision-tenant-email.sh verify-domain
#       Asks Resend to re-check DNS and prints the domain and record status.
#   ./scripts/provision-tenant-email.sh tenant <slug> "<Client Name>" [--rotate] [--no-restart]
#       Creates the tenant's sending-only key, writes TENANT_SMTP_* into
#       backend/tenants/<slug>/.env and recreates its auth-service. Idempotent:
#       an existing key is kept unless --rotate is given.
#   ./scripts/provision-tenant-email.sh test <slug> <recipient@example.com>
#       Sends a test email with the tenant's own key.
#
# Configuration: /etc/auditsphere/email.env (override with AUDITSPHERE_EMAIL_ENV),
# root-only (chmod 600), never committed:
#   RESEND_MASTER_API_KEY=re_xxx              # full-access key, required
#   EMAIL_SENDING_DOMAIN=mail.auditsphere.app # optional
#   EMAIL_FROM_LOCALPART=no-reply             # optional
#   RESEND_REGION=ap-northeast-1              # optional, used by setup-domain
#
# Requires: curl, jq.
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
RESEND_API="https://api.resend.com"
CONFIG_FILE=${AUDITSPHERE_EMAIL_ENV:-/etc/auditsphere/email.env}

die()  { echo "Error: $*" >&2; exit 1; }
info() { echo "==> $*"; }

for bin in curl jq; do
    command -v "$bin" >/dev/null 2>&1 || die "$bin is required (RHEL: dnf install -y $bin)"
done

[ -f "$CONFIG_FILE" ] || die "$CONFIG_FILE not found. Create it (chmod 600) with RESEND_MASTER_API_KEY=re_..."
# shellcheck disable=SC1090
source "$CONFIG_FILE"

: "${RESEND_MASTER_API_KEY:?RESEND_MASTER_API_KEY is not set in $CONFIG_FILE}"
SENDING_DOMAIN=${EMAIL_SENDING_DOMAIN:-mail.auditsphere.app}
FROM_LOCALPART=${EMAIL_FROM_LOCALPART:-no-reply}
REGION=${RESEND_REGION:-ap-northeast-1}

# Resend SMTP relay. Port 587 is STARTTLS, which Go's net/smtp.SendMail
# negotiates on its own; 465 (implicit TLS) is not supported by SendMail.
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_USERNAME=resend

# resend <method> <path> [json-body] — prints the response body, dies on non-2xx.
resend() {
    local method=$1 path=$2 data=${3:-} out status
    out=$(mktemp)
    local args=(-sS -o "$out" -w '%{http_code}' -X "$method"
        -H "Authorization: Bearer ${RESEND_MASTER_API_KEY}"
        -H "Content-Type: application/json")
    [ -n "$data" ] && args+=(-d "$data")
    status=$(curl "${args[@]}" "${RESEND_API}${path}") || { rm -f "$out"; die "request to Resend failed: $method $path"; }
    if [[ "$status" != 2* ]]; then
        echo "Resend API $method $path -> HTTP $status" >&2
        cat "$out" >&2; echo >&2
        rm -f "$out"
        exit 1
    fi
    cat "$out"; rm -f "$out"
}

domain_id() {
    resend GET /domains | jq -r --arg n "$SENDING_DOMAIN" '.data[] | select(.name == $n) | .id' | head -1
}

print_records() {  # print_records <domain-json>
    echo ""
    echo "    Add these records in the DNS zone of ${SENDING_DOMAIN#*.} (Rumahweb)."
    echo "    HOST is relative to that zone, which is the form Rumahweb's editor expects."
    echo ""
    printf '    %-6s %-40s %-9s %s\n' TYPE HOST PRIORITY VALUE
    jq -r '.records[] | [.type, .name, (.priority // "-" | tostring), .value, .status] | @tsv' <<<"$1" \
        | while IFS=$'\t' read -r type name prio value status; do
            printf '    %-6s %-40s %-9s %s   [%s]\n' "$type" "$name" "$prio" "$value" "$status"
        done
    echo ""
}

cmd_setup_domain() {
    local id json
    id=$(domain_id)
    if [ -n "$id" ]; then
        info "$SENDING_DOMAIN is already registered in Resend (id $id)."
        json=$(resend GET "/domains/$id")
    else
        info "Registering $SENDING_DOMAIN in Resend (region $REGION)..."
        json=$(resend POST /domains "$(jq -n --arg n "$SENDING_DOMAIN" --arg r "$REGION" '{name: $n, region: $r}')")
    fi
    echo "    status: $(jq -r '.status' <<<"$json")"
    print_records "$json"
    echo "    Once the records are in, run: $0 verify-domain"
}

cmd_verify_domain() {
    local id json
    id=$(domain_id)
    [ -n "$id" ] || die "$SENDING_DOMAIN is not registered yet — run: $0 setup-domain"
    resend POST "/domains/$id/verify" >/dev/null
    json=$(resend GET "/domains/$id")
    info "$SENDING_DOMAIN status: $(jq -r '.status' <<<"$json")"
    print_records "$json"
    [ "$(jq -r '.status' <<<"$json")" = "verified" ] \
        || echo "    DNS can take a while to propagate. Verification runs again on Resend's side; re-run this later."
}

tenant_paths() {  # sets TENANT_DIR, ENV_FILE
    TENANT_DIR="$ROOT_DIR/backend/tenants/$1"
    ENV_FILE="$TENANT_DIR/.env"
    [ -f "$ENV_FILE" ] || die "$ENV_FILE not found — generate the tenant stack first (scripts/generate-tenant-stack.sh)"
}

env_value() { awk -F= -v k="$1" '$1==k {sub(/^[^=]*=/,""); print; exit}' "$ENV_FILE"; }

cmd_tenant() {
    local slug=${1:?"usage: $0 tenant <slug> \"<Client Name>\" [--rotate] [--no-restart]"}
    local client=${2:?"usage: $0 tenant <slug> \"<Client Name>\" [--rotate] [--no-restart]"}
    shift 2
    local rotate=false restart=true
    for arg in "$@"; do
        case $arg in
            --rotate) rotate=true ;;
            --no-restart) restart=false ;;
            *) die "unknown option: $arg" ;;
        esac
    done

    tenant_paths "$slug"
    local key_name="auditsphere-tenant-${slug}"

    local id status
    id=$(domain_id)
    [ -n "$id" ] || die "$SENDING_DOMAIN is not registered yet — run: $0 setup-domain"
    status=$(resend GET "/domains/$id" | jq -r '.status')
    [ "$status" = "verified" ] \
        || echo "    ! $SENDING_DOMAIN is '$status', not verified — the key is created now but mail will bounce until DNS verifies."

    local token
    token=$(env_value TENANT_SMTP_PASSWORD)
    if [ -n "$token" ] && [ "$rotate" = false ]; then
        info "Tenant '$slug' already has a sending key — keeping it (use --rotate to replace)."
    else
        # Resend never returns a token again after creation, so a rotation must
        # create a new key; revoke the old ones by name so they stop working.
        local old_ids
        old_ids=$(resend GET /api-keys | jq -r --arg n "$key_name" '.data[] | select(.name == $n) | .id')
        info "Creating sending-only key '$key_name' scoped to $SENDING_DOMAIN..."
        token=$(resend POST /api-keys "$(jq -n --arg n "$key_name" --arg d "$id" \
            '{name: $n, permission: "sending_access", domain_id: $d}')" | jq -r '.token')
        [ -n "$token" ] && [ "$token" != "null" ] || die "Resend returned no token for $key_name"
        for old in $old_ids; do
            resend DELETE "/api-keys/$old" >/dev/null && echo "    -> revoked previous key $old"
        done
    fi

    # Compose env-files take the rest of the line as the value, and Go's
    # mail.ParseAddress rejects a bare comma and silently drops "(...)" as a
    # comment. Strip those so "Bank X, Tbk (Persero)" stays a valid From.
    local display
    display=$(printf '%s' "$client" | tr -d '"<>#(),;:@[]\\\r\n' | tr -s ' ')
    local from="${display} via AuditSphere <${FROM_LOCALPART}@${SENDING_DOMAIN}>"

    local tmp
    tmp=$(mktemp)
    grep -v '^TENANT_SMTP_' "$ENV_FILE" > "$tmp" || true
    cat >> "$tmp" <<EOF
TENANT_SMTP_HOST=$SMTP_HOST
TENANT_SMTP_PORT=$SMTP_PORT
TENANT_SMTP_USERNAME=$SMTP_USERNAME
TENANT_SMTP_PASSWORD=$token
TENANT_SMTP_FROM=$from
EOF
    chmod 600 "$tmp"
    mv "$tmp" "$ENV_FILE"
    info "Wrote TENANT_SMTP_* to backend/tenants/$slug/.env (From: $from)"

    local container="rb_audit_${slug}_auth_service"
    if [ "$restart" = true ] && docker ps --format '{{.Names}}' | grep -qx "$container"; then
        info "Recreating auth-service-${slug} to load the new credentials..."
        docker compose --project-directory "$TENANT_DIR" --env-file "$ENV_FILE" \
            up -d --no-deps "auth-service-${slug}"
    else
        echo "    -> $container is not running; the credentials apply on its next start."
    fi
}

cmd_test() {
    local slug=${1:?"usage: $0 test <slug> <recipient>"} to=${2:?"usage: $0 test <slug> <recipient>"}
    tenant_paths "$slug"
    local token from
    token=$(env_value TENANT_SMTP_PASSWORD)
    from=$(env_value TENANT_SMTP_FROM)
    [ -n "$token" ] || die "tenant '$slug' has no sending key — run: $0 tenant $slug \"<Client Name>\""

    # Uses the tenant's own key on purpose: this proves the scoped key works,
    # not the master key.
    RESEND_MASTER_API_KEY=$token resend POST /emails "$(jq -n --arg f "$from" --arg t "$to" --arg s "$slug" \
        '{from: $f, to: [$t], subject: ("AuditSphere email test — " + $s),
          text: ("This is a test email from the AuditSphere tenant \"" + $s + "\". If you received it, outbound email is working.")}')" \
        | jq -r '"==> Sent, Resend id " + .id'
}

case ${1:-} in
    setup-domain)  shift; cmd_setup_domain "$@" ;;
    verify-domain) shift; cmd_verify_domain "$@" ;;
    tenant)        shift; cmd_tenant "$@" ;;
    test)          shift; cmd_test "$@" ;;
    *) sed -n '2,34p' "$0" | sed 's/^# \{0,1\}//'; exit 1 ;;
esac
