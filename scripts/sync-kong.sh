#!/usr/bin/env bash
# =============================================================================
# Keeps the running Kong gateway in step with the Kong config file in the repo.
#
# The Kong container bind-mounts the repo checkout, so a git sync updates the
# file on disk, but Kong only reads it on reload. Without this step a route
# added to the repo never reaches production (see the strategic plan import
# 404 of 2026-10-21).
#
# Steps:
#   1. Hash the config. If it equals the last applied hash, stop.
#   2. Validate it: start a throwaway DB-less Kong with the same image and
#      wait for "kong health". Invalid config is never applied.
#   3. Report the route-level diff against the last applied copy.
#   4. MODE=apply only: back up, reload Kong, check health. If the reload
#      fails, restore the last applied copy and reload again.
#
# Usage: sync-kong.sh <kong.yml> [report|apply]
#   Run on the VPS. Requires docker and the rb_audit_kong container.
# Exit: 0 = in sync or applied, 1 = validation/apply failed, 2 = usage error.
# =============================================================================

set -euo pipefail

CONFIG="${1:-}"
MODE="${2:-report}"
KONG_CONTAINER="${KONG_CONTAINER:-rb_audit_kong}"
STATE_DIR="${KONG_STATE_DIR:-/root/auditsphere-deploy/kong}"
APPLIED_COPY="$STATE_DIR/applied.yml"
APPLIED_HASH="$STATE_DIR/applied.sha256"
BACKUP_DIR="${KONG_BACKUP_DIR:-/root/kong-backups}"

log() { printf '[kong-sync %s] %s\n' "$(date -u +%H:%M:%S)" "$*"; }
die() { log "ERROR: $*"; exit 1; }

[ -n "$CONFIG" ] || { echo "usage: $0 <kong.yml> [report|apply]" >&2; exit 2; }
[ -f "$CONFIG" ] || { echo "config not found: $CONFIG" >&2; exit 2; }
case "$MODE" in report|apply) ;; *) echo "MODE must be report or apply" >&2; exit 2 ;; esac

CONFIG="$(cd "$(dirname "$CONFIG")" && pwd)/$(basename "$CONFIG")"
mkdir -p "$STATE_DIR" "$BACKUP_DIR"
NEW_HASH="$(sha256sum "$CONFIG" | cut -d' ' -f1)"
OLD_HASH="$(cat "$APPLIED_HASH" 2>/dev/null || echo none)"

if [ "$NEW_HASH" = "$OLD_HASH" ]; then
    log "Kong config unchanged since last apply ($NEW_HASH); nothing to do."
    exit 0
fi

# ---- 2. validate in a throwaway DB-less Kong (same image as the live gateway)
IMAGE="${KONG_IMAGE:-$(docker inspect "$KONG_CONTAINER" --format '{{.Config.Image}}')}"
VALIDATE="kong-validate-$$"
validate() {
    local cfg="$1" ok=0
    docker rm -f "$VALIDATE" >/dev/null 2>&1 || true
    docker run -d --name "$VALIDATE" \
        -e KONG_DATABASE=off \
        -e KONG_DECLARATIVE_CONFIG=/kong/kong.yml \
        -e KONG_PROXY_LISTEN=127.0.0.1:18000 \
        -e KONG_ADMIN_LISTEN=127.0.0.1:18001 \
        -v "$cfg:/kong/kong.yml:ro" \
        "$IMAGE" >/dev/null
    for _ in $(seq 1 30); do
        if docker exec "$VALIDATE" kong health >/dev/null 2>&1; then ok=1; break; fi
        if [ "$(docker inspect -f '{{.State.Running}}' "$VALIDATE" 2>/dev/null)" != "true" ]; then break; fi
        sleep 2
    done
    if [ "$ok" -ne 1 ]; then
        log "validation container logs:"
        docker logs --tail 30 "$VALIDATE" 2>&1 | sed 's/^/    /' || true
    fi
    docker rm -f "$VALIDATE" >/dev/null 2>&1 || true
    return $((1 - ok))
}

log "Validating candidate config with $IMAGE (DB-less)..."
validate "$CONFIG" || die "candidate config failed validation; nothing was changed."
log "Validation passed."

# ---- 3. route-level diff against the last applied copy
if [ -f "$APPLIED_COPY" ]; then
    ADDED="$(diff "$APPLIED_COPY" "$CONFIG" | grep -c '^>' || true)"
    REMOVED="$(diff "$APPLIED_COPY" "$CONFIG" | grep -c '^<' || true)"
    log "Changes vs last applied: +$ADDED lines, -$REMOVED lines"
    diff "$APPLIED_COPY" "$CONFIG" | sed 's/^/    /' | head -60 || true
else
    log "No previous applied copy recorded; treating the whole file as new."
fi

if [ "$MODE" = report ]; then
    log "MODE=report: not applied. Set the Kong sync mode to apply to reload."
    exit 0
fi

# ---- 4. apply: back up, reload, verify, roll back on failure
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
LIVE_BACKUP="$BACKUP_DIR/kong.yml.before-sync-$STAMP"
if [ -f "$APPLIED_COPY" ]; then cp -p "$APPLIED_COPY" "$LIVE_BACKUP"; else cp -p "$CONFIG" "$LIVE_BACKUP"; fi
log "Backup of previously applied config: $LIVE_BACKUP"

reload_ok() {
    docker exec "$KONG_CONTAINER" kong reload >/dev/null 2>&1 && \
    docker exec "$KONG_CONTAINER" kong health >/dev/null 2>&1
}

log "Reloading Kong..."
if reload_ok; then
    cp -p "$CONFIG" "$APPLIED_COPY"
    echo "$NEW_HASH" > "$APPLIED_HASH"
    log "Applied and healthy. Kong config hash $NEW_HASH."
    exit 0
fi

log "Reload or health check FAILED. Restoring the previously applied config."
if [ -f "$APPLIED_COPY" ]; then
    # The live file is the checkout copy; restore the last good content into it.
    cp -p "$APPLIED_COPY" "$CONFIG"
    if reload_ok; then
        die "reload failed; previous config restored and healthy. Investigate before retrying."
    fi
fi
die "reload failed and restore did not recover Kong. Backup: $LIVE_BACKUP. Check 'docker logs $KONG_CONTAINER'."
