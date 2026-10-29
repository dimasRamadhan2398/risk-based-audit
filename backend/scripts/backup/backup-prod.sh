#!/usr/bin/env bash
# =============================================================================
# AuditSphere production nightly backup: all Postgres databases + uploads.
#
# Runs ON THE VPS. Read-only with respect to production data: it only runs
# pg_dumpall and tar inside containers, never writes to the database.
#
# Output (BACKUP_DIR, default /root/backups/auditsphere):
#   pg-YYYYmmdd-HHMMSS.sql.gz      full pg_dumpall, gzipped
#   uploads-YYYYmmdd-HHMMSS.tar.gz  contents of /root/uploads in the audit
#                                   service container (not in a volume: they
#                                   live in the container's writable layer)
#   *.sha256                        checksums for both
#
# If RCLONE_REMOTE is set (e.g. "b2:auditsphere-backups"), the files are also
# copied offsite with rclone. Requires rclone to be configured on the VPS.
#
# Usage (cron, nightly):
#   backend/scripts/backup/backup-prod.sh
#
# Env overrides:
#   PG_CONTAINER      default rb_audit_postgres
#   UPLOADS_CONTAINER default rb_audit_audit_service
#   UPLOADS_PATH      default /root/uploads (inside that container)
#   BACKUP_DIR        default /root/backups/auditsphere
#   RETENTION_DAYS    default 14 (local copies only; offsite retention is set
#                     in the remote)
#   RCLONE_REMOTE     optional offsite destination
# =============================================================================
set -euo pipefail

PG_CONTAINER="${PG_CONTAINER:-rb_audit_postgres}"
UPLOADS_CONTAINER="${UPLOADS_CONTAINER:-rb_audit_audit_service}"
UPLOADS_PATH="${UPLOADS_PATH:-/root/uploads}"
BACKUP_DIR="${BACKUP_DIR:-/root/backups/auditsphere}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
RCLONE_REMOTE="${RCLONE_REMOTE:-}"

STAMP="$(date +%Y%m%d-%H%M%S)"
PG_FILE="pg-${STAMP}.sql.gz"
UP_FILE="uploads-${STAMP}.tar.gz"

log() { echo "[backup $(date '+%F %T')] $*"; }

mkdir -p "$BACKUP_DIR"
cd "$BACKUP_DIR"

# 1. Database: every database, roles included. Fails if pg_dumpall fails.
log "dumping databases from ${PG_CONTAINER}"
docker exec "$PG_CONTAINER" pg_dumpall -U postgres --clean --if-exists \
  | gzip -9 > "$PG_FILE"
test -s "$PG_FILE" || { log "ERROR: database dump is empty"; exit 1; }
sha256sum "$PG_FILE" > "${PG_FILE}.sha256"

# 2. Uploads: archive the directory from inside the running container.
log "archiving ${UPLOADS_PATH} from container ${UPLOADS_CONTAINER}"
docker exec "$UPLOADS_CONTAINER" tar -C "$UPLOADS_PATH" -czf - . > "$UP_FILE"
test -s "$UP_FILE" || { log "ERROR: uploads archive is empty"; exit 1; }
sha256sum "$UP_FILE" > "${UP_FILE}.sha256"

# 3. Local retention.
find "$BACKUP_DIR" -maxdepth 1 -type f \( -name 'pg-*' -o -name 'uploads-*' \) \
  -mtime +"$RETENTION_DAYS" -print -delete

# 4. Offsite copy. A failed copy is a failed backup, so the script exits non-zero.
if [ -n "$RCLONE_REMOTE" ]; then
  log "copying to ${RCLONE_REMOTE}"
  rclone copy "$BACKUP_DIR" "$RCLONE_REMOTE" --include "${PG_FILE}*" --include "${UP_FILE}*" --checksum
  log "offsite copy done"
else
  log "WARNING: RCLONE_REMOTE not set, backup is local only"
fi

log "done: ${PG_FILE}, ${UP_FILE}"
