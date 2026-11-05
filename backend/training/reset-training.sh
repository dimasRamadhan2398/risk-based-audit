#!/usr/bin/env bash
# Re-clone prod data into the training environment.
#
#   reset-training.sh --yes                 # DB + uploads + AI models from prod
#   reset-training.sh --yes --with-images   # also re-tag the images prod runs now,
#                                           # and re-copy audit config + Kong config
#
# Prod is only READ: pg_dump of each DB, docker cp of uploads out of the
# running containers, rsync of the python-ai model volume. Only training
# containers are stopped/started. Training data written since the last reset
# is DISCARDED. No seeders run (migrate-only start commands).
TRAINING_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$TRAINING_DIR/lib.sh"

YES=0; WITH_IMAGES=0
for a in "$@"; do
  case "$a" in
    --yes) YES=1 ;;
    --with-images) WITH_IMAGES=1 ;;
    *) die "unknown argument $a" ;;
  esac
done
[ "$YES" = 1 ] || die "this discards all training data; re-run with --yes"

exec 9>"$TRAINING_DIR/.reset.lock"
flock -n 9 || die "another reset is running"

TS=$(date -u +%Y%m%dT%H%M%SZ)
DUMP_DIR="$TRAINING_DIR/dumps/$TS"
STAGE="$TRAINING_DIR/stage-$TS"
mkdir -p "$DUMP_DIR" "$STAGE" "$STORAGE"/{audit-uploads,master-uploads,python-ai-models}
chmod 700 "$TRAINING_DIR/dumps"
trap 'rm -rf "$STAGE"' EXIT

PROD_PG=rb_audit_postgres
TRAIN_PG=rb_audit_postgres_training

# ---- 1. code/config from prod (optional) ------------------------------------
sync_code_from_prod() {
  local tag; tag=$(date -u +%Y%m%d)
  : > "$TRAINING_DIR/images.txt.new"
  for pair in auth-service:rb_audit_auth_service audit-service:rb_audit_audit_service \
              master-service:rb_audit_master_service risk-service:rb_audit_risk_service \
              analytics-service:rb_audit_analytics_service python-ai:rb_audit_python_ai \
              frontend:rbia-frontend-prod; do
    local name="${pair%%:*}" ctr="${pair#*:}" img ref
    img=$(docker inspect -f '{{.Image}}' "$ctr") || die "prod container $ctr not found"
    ref=$(docker inspect -f '{{.Config.Image}}' "$ctr")
    docker tag "$img" "auditsphere-training/$name:$tag"
    echo "auditsphere-training/$name:$tag <- $ctr ($ref, ${img:7:12})" >> "$TRAINING_DIR/images.txt.new"
  done
  mv "$TRAINING_DIR/images.txt.new" "$TRAINING_DIR/images.txt"
  sed -i -E "s|^TRAINING_TAG=.*|TRAINING_TAG=$tag|; s|^FRONTEND_IMAGE=.*|FRONTEND_IMAGE=auditsphere-training/frontend:$tag|" "$ENV_FILE"

  # audit-service config dir prod bind-mounts (config.yaml; JWT comes from env)
  local src; src=$(docker inspect -f '{{range .Mounts}}{{if eq .Destination "/root/pkg/config"}}{{.Source}}{{end}}{{end}}' rb_audit_audit_service)
  [ -n "$src" ] || die "cannot find prod audit-service config mount"
  mkdir -p "$TRAINING_DIR/config"
  rsync -a --delete "$src/" "$TRAINING_DIR/config/audit-service/"

  # Kong: the file prod Kong has mounted, plus the training origin for CORS
  src=$(docker inspect -f '{{range .Mounts}}{{if eq .Destination "/usr/local/kong/declarative"}}{{.Source}}{{end}}{{end}}' rb_audit_kong)
  mkdir -p "$TRAINING_DIR/kong"
  cp "$src/kong.yml" "$TRAINING_DIR/kong/kong.yml"
  grep -q 'https://training.auditsphere.app' "$TRAINING_DIR/kong/kong.yml" || \
    sed -i 's|^\(\s*\)- "https://auditsphere.app"|&\n\1- "https://training.auditsphere.app"|' "$TRAINING_DIR/kong/kong.yml"
  grep -q 'https://training.auditsphere.app' "$TRAINING_DIR/kong/kong.yml" || die "could not add training CORS origin"

  # frontend env: prod's (relative /api/v1, proxies to kong/python-ai DNS names)
  grep -vE '^\s*(#|$)' /root/rbia-frontend-prod.env > "$TRAINING_DIR/frontend.env"
  if grep -qiE 'https?://[^ ]*auditsphere\.app' "$TRAINING_DIR/frontend.env"; then
    die "frontend.env contains an absolute auditsphere.app URL; fix before continuing"
  fi
  log "images re-tagged ($tag):"; cat "$TRAINING_DIR/images.txt"
}
if [ "$WITH_IMAGES" = 1 ]; then sync_code_from_prod; fi

# ---- 2. read prod (dump + uploads + models) ---------------------------------
log "dumping prod DBs to $DUMP_DIR"
for db in $DBS; do
  docker exec "$PROD_PG" pg_dump -U postgres -Fc "$db" > "$DUMP_DIR/$db.dump"
  log "  $db $(du -h "$DUMP_DIR/$db.dump" | cut -f1)"
done
chmod 600 "$DUMP_DIR"/*.dump

log "copying prod uploads"
docker cp rb_audit_audit_service:/root/uploads/. "$STAGE/audit-uploads/" 2>/dev/null || mkdir -p "$STAGE/audit-uploads"
if docker exec rb_audit_master_service test -d /root/uploads; then
  docker cp rb_audit_master_service:/root/uploads/. "$STAGE/master-uploads/"
else
  mkdir -p "$STAGE/master-uploads"
fi
PROD_MODELS=$(docker inspect -f '{{range .Mounts}}{{if eq .Destination "/app/models"}}{{.Source}}{{end}}{{end}}' rb_audit_python_ai)
PROD_AITRAIN=$(docker inspect -f '{{range .Mounts}}{{if eq .Destination "/app/ai_model_training"}}{{.Source}}{{end}}{{end}}' rb_audit_python_ai)

# ---- 3. stop training app containers -----------------------------------------
log "stopping training app containers"
tcompose stop frontend kong analytics-service python-ai auth-service audit-service master-service risk-service 2>/dev/null || true

# ---- 4. restore into training postgres ---------------------------------------
check_memory "$(pending_training_mb)"
ensure_network
tcompose up -d --no-build postgres redis kafka
wait_healthy "$TRAIN_PG" 120
wait_healthy rb_audit_redis_training 60
wait_healthy rb_audit_kafka_training 180

for db in $DBS; do
  log "restoring $db"
  docker exec "$TRAIN_PG" dropdb -U postgres --if-exists --force "$db"
  docker exec "$TRAIN_PG" createdb -U postgres "$db"
  docker exec -i "$TRAIN_PG" pg_restore -U postgres -d "$db" --no-owner --exit-on-error < "$DUMP_DIR/$db.dump"
done

# Prod MFA secrets are encrypted with prod's MFA_ENCRYPTION_KEY, which training
# must not have (tenant isolation). Drop every MFA enrolment and trusted device
# so training users sign in with password only and can re-enrol MFA with
# training's own key. The row-count comparison below skips these two tables.
MFA_RESET_TABLES="public.mfa_setups public.trusted_devices"
log "clearing MFA enrolments and trusted devices in rb_audit_auth_service"
docker exec "$TRAIN_PG" psql -U postgres -d rb_audit_auth_service -v ON_ERROR_STOP=1 -qAtc \
  "delete from trusted_devices; delete from mfa_setups;"

# ---- 5. files ----------------------------------------------------------------
rsync -a --delete "$STAGE/audit-uploads/" "$STORAGE/audit-uploads/"
rsync -a --delete "$STAGE/master-uploads/" "$STORAGE/master-uploads/"
[ -n "$PROD_MODELS" ] && rsync -a --delete "$PROD_MODELS/" "$STORAGE/python-ai-models/"
[ -n "$PROD_AITRAIN" ] && rsync -a --delete "$PROD_AITRAIN/" "$TRAINING_DIR/ai_model_training/"
log "uploads: audit=$(find "$STORAGE/audit-uploads" -type f | wc -l) files, master=$(find "$STORAGE/master-uploads" -type f | wc -l) files"

# ---- 6. clean caches/queues --------------------------------------------------
docker exec rb_audit_redis_training redis-cli FLUSHALL >/dev/null
for t in audit-universe.events department.events activity-logs service-logs; do
  docker exec rb_audit_kafka_training kafka-topics --bootstrap-server kafka:29092 --create --if-not-exists --topic "$t" --partitions 3 --replication-factor 1 >/dev/null
done

# ---- 7. start everything (migrate-only) -------------------------------------
"$TRAINING_DIR/start-training.sh"

# ---- 8. row-count comparison prod vs training --------------------------------
counts() { # $1 container $2 db
  local q
  q=$(docker exec "$1" psql -U postgres -d "$2" -Atc "select string_agg(format('select %L, count(*) from %I.%I', schemaname||'.'||tablename, schemaname, tablename), ' union all ' order by schemaname, tablename) from pg_tables where schemaname not in ('pg_catalog','information_schema')")
  [ -n "$q" ] && docker exec "$1" psql -U postgres -d "$2" -Atc "$q" |
    grep -vE "^($(echo "$MFA_RESET_TABLES" | tr ' ' '|'))\|"
}
log "row counts (prod vs training; differences only, writes to prod since the dump show up here)"
diffs=0
for db in $DBS; do
  if ! diff <(counts "$PROD_PG" "$db" | sort) <(counts "$TRAIN_PG" "$db" | sort) > "$STAGE/diff.txt"; then
    diffs=1; echo "== $db"; cat "$STAGE/diff.txt"
  fi
done
[ "$diffs" = 0 ] && log "row counts identical for all tables in: $DBS"

# The dumps are unencrypted copies of prod and are only needed for the restore
# above, so delete them by default. Prod has its own encrypted nightly backups.
# KEEP_DUMP_SETS=N keeps the newest N sets (for debugging a failed reset).
KEEP_DUMP_SETS="${KEEP_DUMP_SETS:-0}"
ls -1dt "$TRAINING_DIR"/dumps/*/ | tail -n +"$((KEEP_DUMP_SETS + 1))" | xargs -r rm -rf
log "reset done ($TS)"
