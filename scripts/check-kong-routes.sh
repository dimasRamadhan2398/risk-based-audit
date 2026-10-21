#!/usr/bin/env bash
# =============================================================================
# Fails when a frontend API call has no matching route in the production Kong
# config. A missing route makes Kong answer 404 "no Route matched" for the
# feature, while the backend itself is fine.
#
# Frontend calls are matched by the base-URL variable they use:
#   ${...BaseUrl}/<seg>      -> /api/v1/<seg>          (audit/master/risk/auth)
#   ${...analyticsBase}/<seg> or ${...AnalyticsBase}/<seg> -> /api/analytics/<seg>
#   literal /api/v1/<seg> and /api/analytics/<seg> strings are checked as-is.
#
# Limitation: analytics calls written as safeApiFetch('/caatt/...') are relative
# to the analytics base and are not extracted; the /api/analytics/caatt route
# covers them today. Keep that route when editing the Kong config.
#
# Usage: scripts/check-kong-routes.sh [kong-config-file]
# Exit: 0 = every frontend segment is routed, 1 = gaps found, 2 = usage error.
# =============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KONG_FILE="${1:-$REPO_ROOT/backend/kong-gateway/kong/prod/kong.yml}"
FRONTEND_DIR="$REPO_ROOT/frontend"

[ -f "$KONG_FILE" ] || { echo "Kong config not found: $KONG_FILE" >&2; exit 2; }
[ -d "$FRONTEND_DIR" ] || { echo "Frontend dir not found: $FRONTEND_DIR" >&2; exit 2; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# Source files only (skip dependencies and build output), with comment lines
# removed: a TODO that mentions an endpoint is not a call.
scan_sources() {
    find "$FRONTEND_DIR" \( -path '*/node_modules' -o -path '*/.output' -o -path '*/.nuxt' -o -path '*/dist' \) -prune \
        -o -type f \( -name '*.vue' -o -name '*.ts' -o -name '*.js' \) -print0 \
        | xargs -0 cat 2>/dev/null \
        | sed -E '/<!--/,/-->/d; /^[[:space:]]*\/\*/,/\*\//d; /^[[:space:]]*(\/\/|\*)/d'
}

# 1) Base-URL calls: ${baseUrl}/seg, ${getAuditServiceBaseUrl()}/seg etc.
scan_sources | grep -hoE '\$\{[^}]*[Bb]ase[Uu]rl[^}]*\}/[a-zA-Z0-9_-]+' \
    | sed -E 's#.*\}/##' | sort -u > "$TMP/seg-v1.txt" || true
scan_sources | grep -hoE '\$\{[^}]*[Aa]nalytics[^}]*\}/[a-zA-Z0-9_-]+' \
    | sed -E 's#.*\}/##' | sort -u > "$TMP/seg-analytics.txt" || true

# 2) Literal paths written in the frontend.
scan_sources | grep -hoE '/api/v1/[a-zA-Z0-9_-]+' \
    | sed -E 's#/api/v1/##' | sort -u >> "$TMP/seg-v1.txt" || true
scan_sources | grep -hoE '/api/analytics/[a-zA-Z0-9_-]+' \
    | sed -E 's#/api/analytics/##' | sort -u >> "$TMP/seg-analytics.txt" || true
sort -u -o "$TMP/seg-v1.txt" "$TMP/seg-v1.txt"
sort -u -o "$TMP/seg-analytics.txt" "$TMP/seg-analytics.txt"

# Segments that the frontend calls but that are not API resources.
# Keep this list short and justified; anything here is intentionally unrouted.
IGNORE_RE='^(api|v1|analytics|uploads)$'

# 3) Routed prefixes from Kong. Each route like "- /api/v1/media" or
#    "- /api/v1/media/:id" covers the segment "media".
grep -oE '^\s*- /api/v1/[a-zA-Z0-9_-]+' "$KONG_FILE" | sed -E 's#.*/api/v1/##' | sort -u > "$TMP/kong-v1.txt"
grep -oE '^\s*- /api/analytics/[a-zA-Z0-9_-]+' "$KONG_FILE" | sed -E 's#.*/api/analytics/##' | sort -u > "$TMP/kong-analytics.txt"

MISSING=0
report_missing() {
    local label="$1" want="$2" have="$3"
    local gaps
    gaps="$(comm -23 "$want" "$have" | grep -vE "$IGNORE_RE" || true)"
    if [ -n "$gaps" ]; then
        MISSING=1
        echo "Frontend calls with no Kong route ($label):"
        sed 's/^/  - /' <<<"$gaps"
    fi
}

report_missing "/api/v1" "$TMP/seg-v1.txt" "$TMP/kong-v1.txt"
report_missing "/api/analytics" "$TMP/seg-analytics.txt" "$TMP/kong-analytics.txt"

echo "Checked $(wc -l < "$TMP/seg-v1.txt" | tr -d ' ') /api/v1 and $(wc -l < "$TMP/seg-analytics.txt" | tr -d ' ') /api/analytics frontend segments against $(basename "$KONG_FILE")."

if [ "$MISSING" -eq 1 ]; then
    echo "FAILED: add the missing route(s) to the Kong config, or remove the frontend call."
    exit 1
fi
echo "OK: every frontend segment has a Kong route."
