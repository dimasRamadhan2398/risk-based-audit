#!/usr/bin/env bash
# =============================================================================
# Decrypt an age-encrypted AuditSphere backup, after verifying its checksum.
#
# Run on a trusted ADMIN machine that holds the private key, never on the VPS.
#
# Usage:
#   AGE_IDENTITY=~/.age/auditsphere-backup.key \
#     backend/scripts/backup/decrypt-backup.sh <file.age> [output-file|-]
#
# Output defaults to the name without ".age", written with mode 600. Use "-"
# to stream to stdout (e.g. straight into psql) so no plaintext hits the disk.
#
# -----------------------------------------------------------------------------
# Key management (one time, on the admin machine)
#   age-keygen -o ~/.age/auditsphere-backup.key     # prints "Public key: age1..."
#   chmod 600 ~/.age/auditsphere-backup.key
#   Store the key file in the password manager AND an offline copy (USB/paper).
#   Losing every copy of the private key = losing every backup.
#   Optional second recipient (e.g. a second admin): put both age1... lines in
#   the VPS recipients file; either private key can decrypt.
#   Only the PUBLIC key goes to the VPS:
#     /root/auditsphere-deploy/backup/age-recipients.txt
#
# -----------------------------------------------------------------------------
# Restore procedure
#   1. Fetch from offsite (admin machine, rclone with a read-capable B2 key):
#        rclone copy b2:auditsphere-prod-backups ./restore \
#          --include 'pg-YYYYmmdd-HHMMSS.sql.gz.age*' --include 'uploads-YYYYmmdd-HHMMSS.tar.gz.age*'
#      or from the VPS: scp root@<vps>:/root/backups/auditsphere/pg-...age* ./restore/
#   2. Prove it restores (throwaway container, nothing touches prod):
#        AGE_IDENTITY=... backend/scripts/backup/restore-test.sh restore/pg-...age restore/uploads-...age
#   3. Real restore into prod Postgres (DESTRUCTIVE: pg_dumpall --clean drops
#      and recreates every database; take a fresh dump first and stop the app
#      services so nothing writes during the restore):
#        ssh root@<vps> 'docker exec rb_audit_postgres pg_dumpall -U postgres | gzip > /root/db-backups/pre-restore-$(date +%s).sql.gz'
#        ssh root@<vps> 'docker stop rb_audit_auth_service rb_audit_audit_service rb_audit_master_service rb_audit_risk_service rb_audit_analytics_service'
#        AGE_IDENTITY=... decrypt-backup.sh restore/pg-...sql.gz.age - | gunzip -c \
#          | grep -vE '^(DROP|CREATE) ROLE (IF EXISTS )?postgres;' \
#          | ssh root@<vps> 'docker exec -i rb_audit_postgres psql -U postgres -v ON_ERROR_STOP=1 -q -d postgres'
#        ssh root@<vps> 'docker start rb_audit_auth_service rb_audit_audit_service rb_audit_master_service rb_audit_risk_service rb_audit_analytics_service'
#      The stream is decrypted on the admin machine and travels over SSH, so
#      the private key never reaches the VPS.
#   4. Uploads:
#        AGE_IDENTITY=... decrypt-backup.sh restore/uploads-...tar.gz.age - \
#          | ssh root@<vps> 'docker exec -i rb_audit_audit_service tar -C /root/uploads -xzf -'
# =============================================================================
set -euo pipefail
umask 077

IN="${1:-}"; OUT="${2:-}"
AGE_IDENTITY="${AGE_IDENTITY:-}"
die() { echo "decrypt-backup: $*" >&2; exit 1; }

[ -n "$IN" ] && [ -f "$IN" ] || die "usage: AGE_IDENTITY=<key> $0 <file.age> [output|-]"
case "$IN" in *.age) ;; *) die "$IN does not end in .age" ;; esac
[ -n "$AGE_IDENTITY" ] && [ -f "$AGE_IDENTITY" ] || die "set AGE_IDENTITY to the private key file"
command -v age >/dev/null || die "age is not installed (brew install age / dnf install age)"

if [ -f "${IN}.sha256" ]; then
  # The .sha256 file names the file without its directory.
  (cd "$(dirname "$IN")" && sha256sum -c --status "$(basename "$IN").sha256" 2>/dev/null \
     || shasum -a 256 -c --status "$(basename "$IN").sha256") || die "checksum mismatch for $IN"
  echo "decrypt-backup: checksum OK" >&2
else
  echo "decrypt-backup: WARNING no ${IN}.sha256, checksum not verified" >&2
fi

if [ "$OUT" = "-" ]; then
  exec age -d -i "$AGE_IDENTITY" "$IN"
fi
OUT="${OUT:-${IN%.age}}"
[ -e "$OUT" ] && die "$OUT already exists"
age -d -i "$AGE_IDENTITY" -o "$OUT" "$IN"
echo "decrypt-backup: wrote $OUT (mode 600); delete it when done" >&2
