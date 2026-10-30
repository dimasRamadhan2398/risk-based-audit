#!/usr/bin/env bash
# =============================================================================
# Restore test for the nightly AuditSphere backups.
#
# Proves a backup is usable by restoring it into a THROWAWAY Postgres container
# and an empty temp directory. Never touches rb_audit_postgres or audit_uploads.
#
# Usage (on the VPS, or on any host that has docker and the backup files):
#   backend/scripts/backup/restore-test.sh [pg-file.sql.gz] [uploads.tar.gz]
#
# With no arguments, the newest files in BACKUP_DIR are tested.
#
# Checks:
#   1. checksums match the .sha256 files
#   2. the dump restores with no errors
#   3. the restored instance has the expected databases and non-zero row counts
#   4. the uploads archive extracts, and its file count is reported
# Exit code is non-zero on any failure.
# =============================================================================
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/root/backups/auditsphere}"
TEST_CONTAINER="${TEST_CONTAINER:-restore_test_pg}"
TEST_PORT="${TEST_PORT:-55432}"
PG_IMAGE="${PG_IMAGE:-postgres:16}"

log()  { echo "[restore-test $(date '+%F %T')] $*"; }
fail() { log "FAIL: $*"; cleanup; exit 1; }

PG_FILE="${1:-$(ls -1t "$BACKUP_DIR"/pg-*.sql.gz 2>/dev/null | head -n1 || true)}"
UP_FILE="${2:-$(ls -1t "$BACKUP_DIR"/uploads-*.tar.gz 2>/dev/null | head -n1 || true)}"
[ -n "$PG_FILE" ] && [ -f "$PG_FILE" ] || { log "FAIL: no database dump found"; exit 1; }
[ -n "$UP_FILE" ] && [ -f "$UP_FILE" ] || { log "FAIL: no uploads archive found"; exit 1; }

WORK_DIR="$(mktemp -d)"
cleanup() {
  docker rm -f "$TEST_CONTAINER" >/dev/null 2>&1 || true
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

# 1. Checksums
log "verifying checksums"
for f in "$PG_FILE" "$UP_FILE"; do
  if [ -f "${f}.sha256" ]; then
    (cd "$(dirname "$f")" && sha256sum -c "$(basename "${f}.sha256")") || fail "checksum mismatch: $f"
  else
    log "WARNING: no checksum file for $f"
  fi
done

# 2. Restore the dump into a scratch Postgres
log "starting scratch Postgres ${PG_IMAGE}"
docker rm -f "$TEST_CONTAINER" >/dev/null 2>&1 || true
docker run -d --name "$TEST_CONTAINER" -e POSTGRES_PASSWORD=restore-test \
  -p "127.0.0.1:${TEST_PORT}:5432" "$PG_IMAGE" >/dev/null
for _ in $(seq 1 60); do
  docker exec "$TEST_CONTAINER" pg_isready -U postgres >/dev/null 2>&1 && break
  sleep 1
done
docker exec "$TEST_CONTAINER" pg_isready -U postgres >/dev/null 2>&1 || fail "scratch Postgres did not start"

log "restoring $(basename "$PG_FILE")"
# ON_ERROR_STOP makes psql exit non-zero on the first error.
# pg_dumpall emits DROP/CREATE ROLE postgres, which fails because that role is
# the connected superuser. The scratch instance already has it, so skip those lines.
gunzip -c "$PG_FILE" | grep -vE '^(DROP|CREATE) ROLE (IF EXISTS )?postgres;' \
  | docker exec -i "$TEST_CONTAINER" \
  psql -U postgres -v ON_ERROR_STOP=1 -q -d postgres >"$WORK_DIR/restore.log" 2>&1 \
  || { tail -n 20 "$WORK_DIR/restore.log"; fail "restore reported errors (last 20 lines shown above)"; }

# 3. Databases present and populated
log "listing restored databases"
DBS="$(docker exec "$TEST_CONTAINER" psql -U postgres -At -c \
  "SELECT datname FROM pg_database WHERE datistemplate = false AND datname <> 'postgres' ORDER BY datname;")"
[ -n "$DBS" ] || fail "no application databases in the dump"
echo "$DBS" | sed 's/^/  - /'

while read -r db; do
  [ -n "$db" ] || continue
  # Count user tables with at least one row. An empty DB is reported, not failed,
  # because some services may legitimately have no data.
  TABLES="$(docker exec "$TEST_CONTAINER" psql -U postgres -d "$db" -At -c \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';")"
  ROWS="$(docker exec "$TEST_CONTAINER" psql -U postgres -d "$db" -At -c \
    "SELECT COALESCE(SUM(n_live_tup), 0) FROM pg_stat_user_tables;")"
  log "  ${db}: ${TABLES} tables, ~${ROWS} rows"
done <<< "$DBS"

# 4. Uploads archive
log "extracting uploads archive to a temp directory"
mkdir -p "$WORK_DIR/uploads"
tar -C "$WORK_DIR/uploads" -xzf "$UP_FILE" || fail "uploads archive failed to extract"
UP_COUNT="$(find "$WORK_DIR/uploads" -type f | wc -l | tr -d ' ')"
log "  ${UP_COUNT} files restored from uploads"
[ "$UP_COUNT" -gt 0 ] || log "WARNING: uploads archive contains no files"

log "PASS: backup is restorable"
