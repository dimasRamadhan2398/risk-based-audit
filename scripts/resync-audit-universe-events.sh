#!/bin/bash
# =============================================================================
# Re-emit one stack's Audit Universe state to Kafka.
#
# Usage:
#   ./scripts/resync-audit-universe-events.sh <api-base-url> <token|->
#
#   <api-base-url>  The stack's API origin, with or without /api/v1, e.g.
#                   https://api.auditsphere.app or https://api-bai.auditsphere.app
#   <token|->       A bearer token for an admin user OF THAT STACK (each stack
#                   has its own JWT secret, so a token from another stack 401s).
#                   Pass "-" to read it from stdin instead of the command line:
#                     pbpaste | ./scripts/resync-audit-universe-events.sh https://api-bai.auditsphere.app -
#
# Calls POST /api/v1/audit-universe/events/resync on risk-service. It queues an
# entity.updated event for every corporate Audit Universe node and a
# year.scored event for every yearly score; risk-service's relay publishes them
# and audit-service re-applies them to annual plan (PKAT) activities. Consumers
# are idempotent, so running it twice is harmless.
#
# When to run: once per stack (control plane and every tenant) after deploying
# the Kafka topic-isolation change, and after any long broker outage. Before
# that change every audit-service shared one consumer group, so each stack
# missed the share of events that landed on another stack's consumer.
#
# The token never goes on curl's command line: it is handed to curl on stdin as
# a config line, so it does not show up in `ps` or shell history (unless you
# typed it as the second argument — prefer "-").
# =============================================================================

set -euo pipefail

die() { echo "Error: $*" >&2; exit 1; }

BASE=${1:-}
TOKEN=${2:-}
[ -n "$BASE" ] && [ -n "$TOKEN" ] || die "usage: $0 <api-base-url> <token|->"
command -v curl >/dev/null 2>&1 || die "curl is required"

if [ "$TOKEN" = "-" ]; then
    IFS= read -r TOKEN || true
    [ -n "$TOKEN" ] || die "no token on stdin"
fi
TOKEN=${TOKEN#Bearer }
case "$TOKEN" in
    *[[:space:]\"\\]*) die "token contains whitespace, quotes or backslashes — pass the raw JWT" ;;
esac

BASE=${BASE%/}
case "$BASE" in
    http://*|https://*) ;;
    *) die "api-base-url must start with http:// or https://" ;;
esac
case "$BASE" in
    */api/v1) URL="$BASE/audit-universe/events/resync" ;;
    *)        URL="$BASE/api/v1/audit-universe/events/resync" ;;
esac

echo "==> POST $URL"
BODY=$(mktemp)
trap 'rm -f "$BODY"' EXIT

STATUS=$(printf 'header = "Authorization: Bearer %s"\n' "$TOKEN" | curl --config - \
    --silent --show-error --max-time 60 \
    --request POST --header 'Content-Type: application/json' --data '{}' \
    --output "$BODY" --write-out '%{http_code}' "$URL") || die "request failed"

echo "    HTTP $STATUS"
cat "$BODY"; echo

case "$STATUS" in
    2??) echo "==> Resync queued. audit-service applies the events within seconds; check its logs for 'annual plan sync'." ;;
    401) die "401: token rejected — it must be minted by THIS stack's auth-service" ;;
    403) die "403: the user lacks permission to resync" ;;
    503) die "503: risk-service has Kafka publishing disabled (KAFKA_ENABLED)" ;;
    *)   die "unexpected status $STATUS" ;;
esac
