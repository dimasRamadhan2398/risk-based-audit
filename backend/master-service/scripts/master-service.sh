#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

usage() {
  cat <<'EOF'
Usage: ./scripts/master-service.sh <command> [args]

Commands:
  migrate   Run database migrations
  seed      Seed database (runs migrations first)
  root      Drop all tables, migrate, and seed
  serve     Start the HTTP server

Extra arguments are passed through to the underlying command (e.g. --config path).
EOF
}

command="${1:-}"
if [[ -z "$command" ]]; then
  usage
  exit 1
fi
shift || true

case "$command" in
  migrate|seed|serve)
    go run ./cmd "$command" "$@"
    ;;
  root)
    go run ./cmd fresh "$@"
    ;;
  *)
    usage
    exit 1
    ;;
esac
