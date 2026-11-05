#!/usr/bin/env bash
# =============================================================================
# AuditSphere production nightly backup: all Postgres databases + uploads,
# encrypted with age before anything is written to disk.
#
# Runs ON THE VPS. Read-only with respect to production data: it only runs
# pg_dumpall and tar inside containers, never writes to the database.
#
# Encryption: the stream is piped through `age -R <recipients file>`, so the
# plaintext dump never touches the disk. The recipients file holds PUBLIC keys
# only (lines starting with "age1"). The matching private keys live OFF the
# VPS (password manager + offline copy). A compromised VPS or a leaked B2 key
# therefore cannot read the backups. See decrypt-backup.sh for the restore.
#
# Output (BACKUP_DIR, default /root/backups/auditsphere):
#   pg-YYYYmmdd-HHMMSS.sql.gz.age       pg_dumpall | gzip | age
#   uploads-YYYYmmdd-HHMMSS.tar.gz.age  tar of /root/uploads | age
#   *.sha256                            checksums of the ENCRYPTED files
#
# If RCLONE_REMOTE is set (e.g. "b2:auditsphere-prod-backups"), the files are
# also copied offsite with rclone.
#
# Usage (cron, nightly):
#   backend/scripts/backup/backup-prod.sh
#
# Env overrides:
#   PG_CONTAINER         default rb_audit_postgres
#   UPLOADS_CONTAINER    default rb_audit_audit_service
#   UPLOADS_PATH         default /root/uploads (inside that container)
#   BACKUP_DIR           default /root/backups/auditsphere
#   RETENTION_DAYS       default 14 (local copies only; offsite retention is
#                        set by the bucket lifecycle)
#   RCLONE_REMOTE        optional offsite destination
#   AGE_RECIPIENTS_FILE  default /root/auditsphere-deploy/backup/age-recipients.txt
#   ALLOW_PLAINTEXT      default 0. With 1 and no age/recipients, falls back to
#                        an unencrypted backup (emergency only). Otherwise a
#                        missing key or binary fails the backup (fail closed).
# =============================================================================
set -euo pipefail
umask 077

PG_CONTAINER="${PG_CONTAINER:-rb_audit_postgres}"
UPLOADS_CONTAINER="${UPLOADS_CONTAINER:-rb_audit_audit_service}"
UPLOADS_PATH="${UPLOADS_PATH:-/root/uploads}"
BACKUP_DIR="${BACKUP_DIR:-/root/backups/auditsphere}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
RCLONE_REMOTE="${RCLONE_REMOTE:-}"
AGE_RECIPIENTS_FILE="${AGE_RECIPIENTS_FILE:-/root/auditsphere-deploy/backup/age-recipients.txt}"
ALLOW_PLAINTEXT="${ALLOW_PLAINTEXT:-0}"

STAMP="$(date +%Y%m%d-%H%M%S)"

log() { echo "[backup $(date '+%F %T')] $*"; }
die() { log "ERROR: $*"; exit 1; }

# ---- encryption pre-flight ---------------------------------------------------
ENCRYPT=1
if ! command -v age >/dev/null 2>&1 || [ ! -s "$AGE_RECIPIENTS_FILE" ]; then
  [ "$ALLOW_PLAINTEXT" = 1 ] || die "age or ${AGE_RECIPIENTS_FILE} missing; refusing to write an unencrypted backup (set ALLOW_PLAINTEXT=1 to override)"
  ENCRYPT=0
  log "WARNING: ALLOW_PLAINTEXT=1, writing an UNENCRYPTED backup"
else
  # The private key must never be on this host.
  if grep -q 'AGE-SECRET-KEY' "$AGE_RECIPIENTS_FILE"; then
    die "${AGE_RECIPIENTS_FILE} contains a private key; it must hold public keys (age1...) only"
  fi
  grep -qE '^age1[0-9a-z]+' "$AGE_RECIPIENTS_FILE" || die "no age1... public key in ${AGE_RECIPIENTS_FILE}"
fi

if [ "$ENCRYPT" = 1 ]; then
  PG_FILE="pg-${STAMP}.sql.gz.age"
  UP_FILE="uploads-${STAMP}.tar.gz.age"
  seal() { age -R "$AGE_RECIPIENTS_FILE"; }
else
  PG_FILE="pg-${STAMP}.sql.gz"
  UP_FILE="uploads-${STAMP}.tar.gz"
  seal() { cat; }
fi

mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"
cd "$BACKUP_DIR"

# Write to .partial and rename only on success, so a failed run never leaves a
# truncated file that looks like a good backup (and never gets copied offsite).
cleanup_partial() { rm -f -- "$BACKUP_DIR"/*.partial; }
trap cleanup_partial EXIT

# Encrypted files start with this header; checks that age really ran.
check_sealed() {
  test -s "$1" || die "$1 is empty"
  if [ "$ENCRYPT" = 1 ]; then
    head -c 21 "$1" | grep -q '^age-encryption.org/v1' || die "$1 has no age header"
  fi
}

# 1. Database: every database, roles included. pipefail fails the run if
#    pg_dumpall, gzip or age fails.
log "dumping databases from ${PG_CONTAINER} (encrypt=${ENCRYPT})"
docker exec "$PG_CONTAINER" pg_dumpall -U postgres --clean --if-exists \
  | gzip -9 | seal > "${PG_FILE}.partial"
check_sealed "${PG_FILE}.partial"
mv -- "${PG_FILE}.partial" "$PG_FILE"
sha256sum "$PG_FILE" > "${PG_FILE}.sha256"

# 2. Uploads: archive the directory from inside the running container.
log "archiving ${UPLOADS_PATH} from container ${UPLOADS_CONTAINER}"
docker exec "$UPLOADS_CONTAINER" tar -C "$UPLOADS_PATH" -czf - . | seal > "${UP_FILE}.partial"
check_sealed "${UP_FILE}.partial"
mv -- "${UP_FILE}.partial" "$UP_FILE"
sha256sum "$UP_FILE" > "${UP_FILE}.sha256"

# 3. Local retention (encrypted and any leftover plaintext files alike).
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
