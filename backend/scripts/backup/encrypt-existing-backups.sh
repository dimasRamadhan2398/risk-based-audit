#!/usr/bin/env bash
# =============================================================================
# One-off: encrypt the plaintext backups/dumps that already sit on the VPS,
# then (in a separate, explicit step) delete the plaintext originals.
#
# Runs ON THE VPS with the PUBLIC key only. It cannot decrypt, so before the
# delete step, prove the round trip on the admin machine (see step 2 below).
#
# Usage:
#   encrypt-existing-backups.sh [--encrypt | --delete-plaintext] PATH...
#
#   (no mode)           dry run: list what would be encrypted / deleted
#   --encrypt           write <file>.age + <file>.age.sha256 next to each
#                       plaintext file; originals are left untouched
#   --delete-plaintext  delete each original (and its old .sha256) ONLY if its
#                       .age exists, is non-empty, has the age header and
#                       matches its .age.sha256
#
# PATH may be a file or a directory (files directly inside it, not recursive).
# Skips *.age, *.sha256, *.log, *.partial and files still being written
# (modified < 10 minutes ago).
#
# Recommended order:
#   1. --encrypt on all paths
#   2. admin machine: scp one pg-*.age + uploads-*.age, run restore-test.sh with
#      AGE_IDENTITY, confirm PASS
#   3. --delete-plaintext on the same paths
#
# Note: on a journaling filesystem on a virtual disk, rm (or shred) cannot
# guarantee the old blocks are unrecoverable. Deleting still removes the
# files from anything that reads the filesystem (backup tools, an attacker
# with file access), which is the realistic threat here.
#
# Env: AGE_RECIPIENTS_FILE (default /root/auditsphere-deploy/backup/age-recipients.txt)
# =============================================================================
set -euo pipefail
umask 077

AGE_RECIPIENTS_FILE="${AGE_RECIPIENTS_FILE:-/root/auditsphere-deploy/backup/age-recipients.txt}"
MODE=dry
case "${1:-}" in
  --encrypt) MODE=encrypt; shift ;;
  --delete-plaintext) MODE=delete; shift ;;
  -h|--help|"") sed -n '2,36p' "$0"; exit 0 ;;
esac
[ $# -gt 0 ] || { echo "no paths given" >&2; exit 1; }

log() { echo "[encrypt-existing] $*"; }

if [ "$MODE" = encrypt ]; then
  command -v age >/dev/null || { log "age is not installed"; exit 1; }
  [ -s "$AGE_RECIPIENTS_FILE" ] || { log "missing $AGE_RECIPIENTS_FILE"; exit 1; }
  if grep -q 'AGE-SECRET-KEY' "$AGE_RECIPIENTS_FILE"; then
    log "$AGE_RECIPIENTS_FILE contains a private key; refusing"; exit 1
  fi
fi

candidates() {
  local p
  for p in "$@"; do
    if [ -d "$p" ]; then
      find "$p" -maxdepth 1 -type f -mmin +10
    elif [ -f "$p" ]; then
      echo "$p"
    else
      log "skip (not found): $p" >&2
    fi
  done | grep -vE '\.(age|sha256|log|partial)$' | sort -u
}

has_good_age() { # $1 = plaintext path
  local a="$1.age"
  [ -s "$a" ] || return 1
  head -c 21 "$a" | grep -q '^age-encryption.org/v1' || return 1
  [ -f "$a.sha256" ] || return 1
  (cd "$(dirname "$a")" && sha256sum -c --status "$(basename "$a").sha256")
}

n=0; bytes=0; failed=0
while IFS= read -r f; do
  [ -n "$f" ] || continue
  size=$(wc -c < "$f" | tr -d " ")
  case "$MODE" in
    dry)
      if has_good_age "$f"; then state="encrypted copy exists -> would delete plaintext"
      else state="would encrypt"; fi
      log "$state: $f ($size bytes)" ;;
    encrypt)
      if has_good_age "$f"; then log "already encrypted: $f"; continue; fi
      age -R "$AGE_RECIPIENTS_FILE" -o "$f.age.partial" "$f"
      head -c 21 "$f.age.partial" | grep -q '^age-encryption.org/v1' || { log "FAILED (no header): $f"; rm -f "$f.age.partial"; failed=$((failed+1)); continue; }
      mv "$f.age.partial" "$f.age"
      (cd "$(dirname "$f")" && sha256sum "$(basename "$f").age" > "$(basename "$f").age.sha256")
      chmod 600 "$f.age" "$f.age.sha256"
      log "encrypted: $f -> $f.age" ;;
    delete)
      if has_good_age "$f"; then
        rm -f -- "$f" "$f.sha256"
        log "deleted plaintext: $f"
      else
        log "KEPT (no verified .age): $f"; failed=$((failed+1)); continue
      fi ;;
  esac
  n=$((n+1)); bytes=$((bytes+size))
done < <(candidates "$@")

log "mode=$MODE files=$n bytes=$bytes problems=$failed"
[ "$failed" = 0 ]
