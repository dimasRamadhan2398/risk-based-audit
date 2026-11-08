# Shared helpers for the training scripts. Sourced, not executed.
# Every compose call goes through tcompose(), which pins the training project
# and file, so nothing here can ever address the prod project rb_audit_backend.

set -euo pipefail

TRAINING_DIR="${TRAINING_DIR:-/root/auditsphere-training}"
PROJECT="rb_audit_training"
NETWORK="rb_audit_training_network"
COMPOSE_FILE="$TRAINING_DIR/docker-compose.training.yml"
ENV_FILE="$TRAINING_DIR/training.env"
STORAGE="/var/data/auditsphere/storage/training"
PROD_CLAMAV="rb_audit_clamav"
# Prod DBs cloned into training (tenant DBs such as *_bai are deliberately excluded)
DBS="rb_audit rb_audit_auth_service rb_audit_audit_service rb_audit_master_service rb_audit_risk_service"
# Abort if MemAvailable would drop below this after the stack starts
MIN_AVAILABLE_AFTER_MB=1024
EXPECTED_TRAINING_MB=1600

log() { printf '[%s] %s\n' "$(date -u +%H:%M:%SZ)" "$*"; }
die() { log "ERROR: $*" >&2; exit 1; }

[ "$PROJECT" != "rb_audit_backend" ] || die "refusing to operate on the prod project"
[ -f "$COMPOSE_FILE" ] || die "missing $COMPOSE_FILE"
[ -f "$ENV_FILE" ] || die "missing $ENV_FILE"

tcompose() {
  docker compose -p "$PROJECT" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" --project-directory "$TRAINING_DIR" "$@"
}

ensure_network() {
  docker network inspect "$NETWORK" >/dev/null 2>&1 || docker network create --driver bridge "$NETWORK" >/dev/null
}

# Attach prod ClamAV to the training network as "clamav" (no restart, non-disruptive)
attach_clamav() {
  docker inspect "$PROD_CLAMAV" >/dev/null 2>&1 || { log "WARN: $PROD_CLAMAV not found; uploads will be refused (CLAMAV_MODE=required)"; return 0; }
  if docker inspect -f '{{json .NetworkSettings.Networks}}' "$PROD_CLAMAV" | grep -q "\"$NETWORK\""; then
    return 0
  fi
  docker network connect --alias clamav "$NETWORK" "$PROD_CLAMAV"
  log "attached $PROD_CLAMAV to $NETWORK as clamav"
}

detach_clamav() {
  if docker inspect -f '{{json .NetworkSettings.Networks}}' "$PROD_CLAMAV" 2>/dev/null | grep -q "\"$NETWORK\""; then
    docker network disconnect "$NETWORK" "$PROD_CLAMAV"
    log "detached $PROD_CLAMAV from $NETWORK"
  fi
}

# Available memory check. $1 = MB the step is about to add.
check_memory() {
  local add="${1:-$EXPECTED_TRAINING_MB}" avail
  avail=$(awk '/MemAvailable/{print int($2/1024)}' /proc/meminfo)
  log "MemAvailable=${avail}MB, about to add ~${add}MB"
  [ $((avail - add)) -ge "$MIN_AVAILABLE_AFTER_MB" ] || die "not enough memory: ${avail}MB available, need ${add}+${MIN_AVAILABLE_AFTER_MB}MB"
}

# MB the not-yet-running training containers will roughly use
pending_training_mb() {
  local running
  running=$(docker ps --filter "label=com.docker.compose.project=$PROJECT" -q | wc -l)
  # 12 containers ~1600MB total; scale linearly
  echo $(( EXPECTED_TRAINING_MB * (12 - running) / 12 ))
}

wait_healthy() { # $1 container, $2 timeout seconds
  local c="$1" t="${2:-180}" s
  for _ in $(seq 1 "$t"); do
    s=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$c" 2>/dev/null || echo missing)
    [ "$s" = "healthy" ] && return 0
    sleep 1
  done
  die "$c not healthy after ${t}s (state: $s)"
}
