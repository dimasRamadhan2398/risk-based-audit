#!/usr/bin/env bash
# One-off isolation/smoke test for the training environment. Training-only
# writes; prod is only read (psql SELECT, anonymous/foreign-token GETs).
# Temporarily replaces the password hash of one non-MFA training user and
# always restores it on exit. No secrets are printed.
set -euo pipefail
T=https://training.auditsphere.app
P=https://auditsphere.app
U="${1:-auditor}"
PGT() { docker exec rb_audit_postgres_training psql -U postgres -v ON_ERROR_STOP=1 -At "$@"; }
PGP() { docker exec rb_audit_postgres psql -U postgres -At "$@"; }
code() { curl -s -o /dev/null -w '%{http_code}' "$@"; }
ok() { printf '  %-58s %s\n' "$1" "$2"; }
WORK=$(mktemp -d); chmod 700 "$WORK"

ORIG=$(PGT -d rb_audit_auth_service -c "select password_hash from users where username='$U'")
[ -n "$ORIG" ] || { echo "user $U not found"; exit 1; }
printf '%s' "$ORIG" > "$WORK/orig"
restore() {
  PGT -d rb_audit_auth_service -c "update users set password_hash='$(cat "$WORK/orig")' where username='$U'" >/dev/null
  PGT -d postgres -c "drop database if exists rb_audit_scratch" >/dev/null 2>&1 || true
  rm -rf "$WORK"
}
trap restore EXIT

echo "== routing / anonymous"
ok "training GET /api/v1/audit-sops (no token) -> 401" "$(code $T/api/v1/audit-sops)"
ok "training POST /api/v1/auth/login {} -> 400" "$(code -X POST -H 'Content-Type: application/json' -d '{}' $T/api/v1/auth/login)"
ok "prod     GET /api/v1/audit-sops (no token) -> 401" "$(code $P/api/v1/audit-sops)"
ok "prod     POST /api/v1/auth/login {} -> 400" "$(code -X POST -H 'Content-Type: application/json' -d '{}' $P/api/v1/auth/login)"

echo "== login with cloned prod account '$U' (temporary password, training DB only)"
PW="Tr$(openssl rand -hex 12)"
PGT -d postgres -c "drop database if exists rb_audit_scratch" -c "create database rb_audit_scratch" >/dev/null
PGT -d rb_audit_scratch -c "create extension pgcrypto" >/dev/null
HASH=$(PGT -d rb_audit_scratch -c "select crypt('$PW', gen_salt('bf', 10))")
PGT -d rb_audit_auth_service -c "update users set password_hash='$HASH' where username='$U'" >/dev/null
curl -s -X POST -H 'Content-Type: application/json' -d "{\"username\":\"$U\",\"password\":\"$PW\"}" $T/api/v1/auth/login > "$WORK/login.json"
TOK=$(jq -r '.data.token // .token // .data.access_token // empty' "$WORK/login.json")
restore_now() { PGT -d rb_audit_auth_service -c "update users set password_hash='$(cat "$WORK/orig")' where username='$U'" >/dev/null; }
restore_now
[ -n "$TOK" ] || { echo "login failed: $(jq -c '{success,message}' "$WORK/login.json" 2>/dev/null)"; exit 1; }
ok "training login -> token" "yes (len ${#TOK}), password hash restored"
ok "training GET /api/v1/auth/me with token -> 200" "$(code -H "Authorization: Bearer $TOK" $T/api/v1/auth/me)"
ok "training GET /api/v1/audit-sops with token -> 200" "$(code -H "Authorization: Bearer $TOK" $T/api/v1/audit-sops)"
ok "prod GET /api/v1/audit-sops with TRAINING token -> 401" "$(code -H "Authorization: Bearer $TOK" $P/api/v1/audit-sops)"

echo "== token signed with a different secret (stand-in for a prod token)"
H=$(cut -d. -f1 <<<"$TOK"); B=$(cut -d. -f2 <<<"$TOK")
FOREIGN=$(openssl rand -hex 48)
SIG=$(printf '%s' "$H.$B" | openssl dgst -sha256 -hmac "$FOREIGN" -binary | openssl base64 -A | tr '+/' '-_' | tr -d '=')
ok "training GET /api/v1/audit-sops, same claims, foreign sig -> 401" "$(code -H "Authorization: Bearer $H.$B.$SIG" $T/api/v1/audit-sops)"

echo "== write isolation (audit-sops)"
GID=$(PGT -d rb_audit_audit_service -c "select id from audit_guidelines where deleted_at is null limit 1")
NAME="TRAINING-SMOKE-$(date -u +%H%M%S)"
curl -s -X POST -H "Authorization: Bearer $TOK" -H 'Content-Type: application/json' \
  -d "{\"name\":\"$NAME\",\"guideline_id\":\"$GID\",\"effective_date\":\"2026-11\",\"status\":\"Aktif\"}" \
  $T/api/v1/audit-sops > "$WORK/sop.json"
SOP=$(jq -r '.data.id // .id // empty' "$WORK/sop.json")
if [ -n "$SOP" ]; then
  ok "training create audit_sop" "id=$SOP"
  ok "rows with that id/name in TRAINING db" "$(PGT -d rb_audit_audit_service -c "select count(*) from audit_sops where id='$SOP'")"
  ok "rows with that id/name in PROD db (expect 0)" "$(PGP -d rb_audit_audit_service -c "select count(*) from audit_sops where id='$SOP' or name='$NAME'")"
  ok "training DELETE audit_sop -> 200" "$(code -X DELETE -H "Authorization: Bearer $TOK" $T/api/v1/audit-sops/$SOP)"
  PGT -d rb_audit_audit_service -c "delete from audit_sops where id='$SOP'" >/dev/null
  ok "hard-deleted soft-deleted row in training" "done"
else
  ok "training create audit_sop FAILED" "$(jq -c '{success,message}' "$WORK/sop.json" 2>/dev/null || head -c 200 "$WORK/sop.json")"
fi

echo "== upload through ClamAV"
printf '%%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[]/Count 0>>endobj\ntrailer<</Root 1 0 R>>\n%%%%EOF\n' > "$WORK/training-smoke.pdf"
curl -s -X POST -H "Authorization: Bearer $TOK" -F "file=@$WORK/training-smoke.pdf;type=application/pdf" \
  -F title=TRAINING-SMOKE-UPLOAD -F fileName=training-smoke.pdf -F fileType=application/pdf \
  $T/api/v1/uploaded-plan-documents > "$WORK/up.json"
UP=$(jq -r '.data.id // empty' "$WORK/up.json")
if [ -n "$UP" ]; then
  FP=$(jq -r '.data.file_path // .data.filePath // empty' "$WORK/up.json")
  ok "training upload accepted (scanned)" "id=$UP path=$FP"
  ok "clamd log lines for the scan" "$(docker logs --since 2m rb_audit_audit_service_training 2>&1 | grep -ci 'avscan\|clam' || true)"
  ok "file in training storage" "$(find /var/data/auditsphere/storage/training/audit-uploads -name "$(basename "$FP")" | wc -l)"
  ok "file in PROD audit container (expect 0)" "$(docker exec rb_audit_audit_service sh -c "find /root/uploads -name '$(basename "$FP")' | wc -l")"
  ok "row in PROD db (expect 0)" "$(PGP -d rb_audit_audit_service -c "select count(*) from uploaded_plan_documents where id='$UP'")"
  ok "training DELETE upload -> 200" "$(code -X DELETE -H "Authorization: Bearer $TOK" $T/api/v1/uploaded-plan-documents/$UP)"
  PGT -d rb_audit_audit_service -c "delete from uploaded_plan_documents where id='$UP'" >/dev/null
  find /var/data/auditsphere/storage/training/audit-uploads -name "$(basename "$FP")" -delete
else
  ok "training upload FAILED" "$(jq -c '{success,message}' "$WORK/up.json" 2>/dev/null || head -c 200 "$WORK/up.json")"
fi

echo "== AI"
ok "python-ai /health (inside training network)" "$(docker exec rb_audit_analytics_service_training wget -qO- http://python-ai:8000/health 2>&1 | head -c 160)"
ok "training /api/python-ai/health via frontend" "$(code $T/api/python-ai/health)"
ok "training GET /api/analytics/risk-score/batch -> 200" "$(code -H "Authorization: Bearer $TOK" "$T/api/analytics/risk-score/batch")"
ok "training GET /api/analytics/report (Data Hub read) -> 200" "$(code -H "Authorization: Bearer $TOK" "$T/api/analytics/report")"
