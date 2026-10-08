#!/usr/bin/env bash
# =============================================================================
# AuditSphere Backend Production Deploy Script (auditsphere.app)
#
# Manually deploys selected Go backend services to the production VPS without
# going through .github/workflows/deploy.yml, and without touching data:
#
#   * builds from `git archive <ref>` (never the local working tree)
#   * ships only the selected <svc>-service dirs (committed source + binary)
#   * pg_dumpall backup (verified) + exact row counts / content hashes before
#   * rollback tarball of each current service dir + image tag
#   * audit-service /root/uploads copied out and restored after recreate
#   * containers are started with MIGRATE-ONLY commands via a compose override
#     kept outside the repo (/root/auditsphere-deploy/docker-compose.prod.noseed.yml)
#     because the seeders in docker-compose.prod.yml overwrite production rows
#   * one service at a time: `up -d --no-deps --build <svc>-service`
#     (never whole-stack up, never --remove-orphans, never down, never prune)
#   * health check, automatic rollback of a failing service, stop on failure
#   * recount rows afterwards; any table that lost rows => non-zero exit
#
# Usage:
#   backend/deploy-prod.sh [--ref <git-ref>] [--dry-run] [--yes] <service...>
#   services: auth audit master risk analytics | all
#
# Examples:
#   backend/deploy-prod.sh --dry-run auth audit master
#   backend/deploy-prod.sh --ref origin/staging --yes auth
#
# Never touches: postgres, redis, kafka, zookeeper, kong, python-ai, frontend,
# *_dev containers, or tenant stacks (backend/tenants/*).
# Local side is bash 3.2 compatible (macOS).
# =============================================================================

set -euo pipefail

VPS_HOST="${VPS_HOST:-root@202.10.34.166}"
REMOTE_REPO="/app/rbia-repo"
REMOTE_DEPLOY_ROOT="/root/auditsphere-deploy"
ALL_SERVICES="auth audit master risk analytics"

REF="origin/staging"
DRY_RUN=0
ASSUME_YES=0
REQUESTED=""

usage() {
    sed -n '/^# Usage:/,/^# Examples:/p' "$0" | sed '$d; s/^# \{0,1\}//'
    exit "${1:-0}"
}

die() { echo "ERROR: $*" >&2; exit 1; }

while [ $# -gt 0 ]; do
    case "$1" in
        --ref)      [ $# -ge 2 ] || die "--ref needs a value"; REF="$2"; shift 2 ;;
        --ref=*)    REF="${1#--ref=}"; shift ;;
        --dry-run)  DRY_RUN=1; shift ;;
        --yes|-y)   ASSUME_YES=1; shift ;;
        -h|--help)  usage 0 ;;
        all)        REQUESTED="$REQUESTED $ALL_SERVICES"; shift ;;
        auth|audit|master|risk|analytics) REQUESTED="$REQUESTED $1"; shift ;;
        *)          echo "Unknown argument: $1" >&2; usage 2 ;;
    esac
done

# Canonical order (auth first), de-duplicated.
SERVICES=""
for s in $ALL_SERVICES; do
    case " $REQUESTED " in *" $s "*) SERVICES="$SERVICES $s" ;; esac
done
SERVICES="${SERVICES# }"
[ -n "$SERVICES" ] || { echo "No services selected." >&2; usage 2; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
export COPYFILE_DISABLE=1

TS="$(date -u +%Y%m%dT%H%M%SZ)"
CTL_PATH="/tmp/auditsphere-ssh-%C"
SSH_OPTS=(-o BatchMode=yes -o ControlMaster=auto -o "ControlPath=$CTL_PATH" -o ControlPersist=120
          -o ServerAliveInterval=15 -o ServerAliveCountMax=8 -o ConnectTimeout=20)
TMP_DIR="$(mktemp -d /tmp/auditsphere-deploy.XXXXXX)"

cleanup() {
    rm -rf "$TMP_DIR"
    ssh "${SSH_OPTS[@]}" -O exit "$VPS_HOST" >/dev/null 2>&1 || true
}
trap cleanup EXIT

rssh() { ssh "${SSH_OPTS[@]}" "$VPS_HOST" "$@"; }

section() {
    echo ""
    echo "=== $* ==="
}

# -----------------------------------------------------------------------------
# Remote part. Sent verbatim (quoted heredoc). Args: SHA TS MODE(dry|real) SVC...
# Wrapped in main() and run with stdin from /dev/null so nothing it runs can
# swallow the rest of the script when it is piped through `bash -s`.
# -----------------------------------------------------------------------------
remote_script() {
cat <<'REMOTE'
REPO=/app/rbia-repo
BACKEND=$REPO/backend
COMPOSE_FILE=docker-compose.prod.yml
COMPOSE_PROJECT=rb_audit_backend
DEPLOY_ROOT=/root/auditsphere-deploy
OVERRIDE=$DEPLOY_ROOT/docker-compose.prod.noseed.yml
BACKUP_DIR=/root/db-backups
LOCK_FILE=/tmp/auditsphere_deploy.lock
PG=rb_audit_postgres

log()  { printf '[%s] %s\n' "$(date -u +%H:%M:%S)" "$*"; }
warn() { log "WARNING: $*"; }
die()  { log "ERROR: $*"; exit 1; }

container_of() { echo "rb_audit_${1}_service"; }
port_of() {
    case "$1" in
        auth) echo 8001 ;; audit) echo 8002 ;; master) echo 8003 ;;
        risk) echo 8004 ;; analytics) echo 8084 ;;
    esac
}

write_override() {
cat > "$1" <<'YAML'
# Managed by backend/deploy-prod.sh -- DO NOT put this file in the repo.
# Migrate-only start commands. The default commands in docker-compose.prod.yml
# run seeders on every container start, and those overwrite production rows.
# Use: docker compose -f docker-compose.prod.yml -f <this file> up -d --no-deps --build <svc>
services:
  auth-service:
    command: ["sh", "-c", "./auth --config ./pkg/config/config.yaml migrate up && ./auth --config ./pkg/config/config.yaml serve"]
  audit-service:
    command: ["sh", "-c", "./audit migrate up && ./audit serve"]
  master-service:
    command: ["sh", "-c", "./master migrate && ./master serve"]
YAML
}

rows_sql() {
cat <<'SQL'
SELECT s.table_schema || '.' || s.table_name,
       (xpath('/row/c/text()', s.x))[1]::text,
       (xpath('/row/h/text()', s.x))[1]::text
FROM (
  SELECT table_schema, table_name,
         query_to_xml(
           format('SELECT count(*) AS c, md5(coalesce(string_agg(md5(t::text), '','' ORDER BY md5(t::text)), '''')) AS h FROM %I.%I t',
                  table_schema, table_name),
           false, true, '') AS x
  FROM information_schema.tables
  WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
) s
ORDER BY 1;
SQL
}

list_dbs() {
    docker exec "$PG" psql -U postgres -X -At -c \
        "SELECT datname FROM pg_database WHERE NOT datistemplate ORDER BY 1"
}

# Output lines: db/schema.table|rows|md5-of-content
snapshot_rows() {
    local out="$1" db sql
    sql="$(rows_sql)"
    : > "$out.tmp"
    for db in $(list_dbs); do
        docker exec "$PG" psql -U postgres -X -At -F '|' -v ON_ERROR_STOP=1 -d "$db" -c "$sql" \
            | sed "s|^|$db/|" >> "$out.tmp"
    done
    sort "$out.tmp" > "$out"
    rm -f "$out.tmp"
}

# Output lines: name|id|startedAt|status
snapshot_containers() {
    local n
    docker ps -a --format '{{.Names}}' | sort | while read -r n; do
        docker inspect -f '{{.Name}}|{{.Id}}|{{.State.StartedAt}}|{{.State.Status}}' "$n"
    done | sed 's|^/||' > "$1"
}

compose() {
    (cd "$BACKEND" && docker compose -f "$COMPOSE_FILE" -f "$OVERRIDE_IN_USE" "$@")
}

disk_pct() {
    df --output=pcent / /var/lib/docker /root 2>/dev/null | tail -n +2 | tr -dc '0-9\n' | sort -n | tail -1
}

health_ok() {
    # analytics has no /health route: any HTTP answer means it is serving.
    case "$1" in
        analytics) [ "$2" != "000" ] ;;
        *)         [ "$2" = "200" ] ;;
    esac
}

wait_healthy() {
    local svc="$1" c port st code ok=0 deadline
    c="$(container_of "$svc")"; port="$(port_of "$svc")"
    deadline=$((SECONDS + 150))
    while [ "$SECONDS" -lt "$deadline" ]; do
        st="$(docker inspect -f '{{.State.Status}} {{.State.Restarting}} {{.RestartCount}}' "$c" 2>/dev/null || echo missing)"
        code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "http://127.0.0.1:$port/health" 2>/dev/null || true)"
        [ -n "$code" ] || code=000
        if [ "$st" = "running false 0" ] && health_ok "$svc" "$code"; then
            ok=$((ok + 1))
            log "  $c: state='$st' http=$code (ok $ok/2)"
            [ "$ok" -ge 2 ] && return 0
        else
            [ "$ok" -gt 0 ] && log "  $c: lost health (state='$st' http=$code)"
            ok=0
        fi
        sleep 3
    done
    log "  $c: NOT healthy after 150s (state='$st' http=$code)"
    return 1
}

cmd_ok() {
    # Container must not run seeders; auth/audit/master must use migrate.
    local svc="$1" cmd
    cmd="$(docker inspect -f '{{json .Config.Cmd}}' "$(container_of "$svc")")"
    log "  Cmd: $cmd"
    case "$cmd" in *" seed"*|*'"seed'*|*"/seed"*) log "  Cmd still runs a seeder"; return 1 ;; esac
    case "$svc" in
        auth|audit|master) case "$cmd" in *migrate*) ;; *) log "  Cmd has no migrate step"; return 1 ;; esac ;;
    esac
    return 0
}

uploads_count() {
    docker exec "$(container_of audit)" sh -c 'if [ -d /root/uploads ]; then find /root/uploads -type f | wc -l; else echo none; fi' 2>/dev/null \
        | tr -d ' ' || echo "?"
}

backup_uploads() {
    local c; c="$(container_of audit)"
    rm -rf "$DIR/audit-uploads"
    if docker exec "$c" test -d /root/uploads; then
        docker cp "$c:/root/uploads" "$DIR/audit-uploads" || return 1
        UPLOADS_BEFORE="$(find "$DIR/audit-uploads" -type f | wc -l | tr -d ' ')"
        log "  audit uploads copied out: $UPLOADS_BEFORE files -> $DIR/audit-uploads"
    else
        UPLOADS_BEFORE=none
        log "  audit container has no /root/uploads; nothing to copy"
    fi
}

restore_uploads() {
    local c after; c="$(container_of audit)"
    [ -d "$DIR/audit-uploads" ] || { log "  no uploads backup to restore"; return 0; }
    docker exec "$c" mkdir -p /root/uploads || return 1
    docker cp "$DIR/audit-uploads/." "$c:/root/uploads/" || return 1
    after="$(uploads_count)"
    log "  audit uploads restored: before=$UPLOADS_BEFORE now=$after"
    [ "$after" != "?" ] && [ "$after" != none ] && [ "$after" -ge "$UPLOADS_BEFORE" ] || return 1
}

rollback_service() {
    local svc="$1"
    log "  ROLLBACK $svc-service: restoring previous source/binary from $DIR/rollback-$svc-service.tgz"
    docker logs --tail 80 "$(container_of "$svc")" 2>&1 | sed 's/^/    | /' || true
    if [ -d "$BACKEND/$svc-service" ]; then
        mv "$BACKEND/$svc-service" "$DIR/failed-$svc-service" || return 1
    fi
    tar -xzf "$DIR/rollback-$svc-service.tgz" -C "$BACKEND" || return 1
    compose up -d --no-deps --build "$svc-service" || return 1
    if wait_healthy "$svc"; then log "  rollback of $svc-service healthy"; else log "  ROLLBACK OF $svc-service IS ALSO UNHEALTHY"; fi
    if [ "$svc" = audit ]; then restore_uploads || log "  UPLOADS RESTORE AFTER ROLLBACK FAILED - copy manually from $DIR/audit-uploads"; fi
}

deploy_one() {
    local svc="$1"
    log "=== Deploying $svc-service ($(container_of "$svc")) ==="
    tar -xzf "$PAYLOAD" --no-same-owner -C "$BACKEND" "$svc-service" || { log "  extract failed"; return 1; }
    if [ "$svc" = audit ]; then backup_uploads || { log "  uploads copy-out failed; not recreating audit"; rollback_service "$svc"; return 1; }; fi
    if ! compose up -d --no-deps --build "$svc-service"; then
        log "  compose up failed"; rollback_service "$svc"; return 1
    fi
    if ! wait_healthy "$svc"; then rollback_service "$svc"; return 1; fi
    if ! cmd_ok "$svc"; then log "  unexpected container command"; return 1; fi
    local want got
    want="$(tar -xzOf "$PAYLOAD" BINARY_SHA256 | awk -v s="$svc" '$2 == s { print $1 }')"
    got="$(docker exec "$(container_of "$svc")" sha256sum "/root/$svc" 2>/dev/null | awk '{ print $1 }')"
    log "  binary sha256 built=$want running=$got"
    if [ -z "$want" ] || [ "$want" != "$got" ]; then log "  running binary is NOT the freshly built one"; return 1; fi
    if [ "$svc" = audit ]; then
        restore_uploads || { log "  UPLOADS RESTORE FAILED - files are in $DIR/audit-uploads"; return 1; }
    fi
    log "  $svc-service deployed and healthy"
    return 0
}

preflight() {
    local svc c pct dbs
    log "--- Pre-flight ($MODE) ---"

    if [ "$MODE" = real ]; then
        exec 9>>"$LOCK_FILE"
        flock -n 9 || die "Another AuditSphere deployment holds $LOCK_FILE"
        log "Lock $LOCK_FILE acquired"
    elif [ -e "$LOCK_FILE" ]; then
        exec 9<"$LOCK_FILE"
        if flock -n 9; then log "Lock $LOCK_FILE: free"; flock -u 9; else warn "lock $LOCK_FILE is HELD by another deploy"; fi
        exec 9<&-
    else
        log "Lock $LOCK_FILE: free (file absent)"
    fi

    for t in docker flock curl jq gzip tar setsid; do
        command -v "$t" >/dev/null 2>&1 || die "missing tool on server: $t"
    done

    pct="$(disk_pct)"
    log "Disk usage (max of /, /var/lib/docker, /root): ${pct}%"
    [ "$pct" -lt 90 ] || die "disk usage ${pct}% >= 90%"

    [ "$(docker inspect -f '{{.State.Running}}' "$PG" 2>/dev/null)" = true ] || die "$PG is not running"
    log "Postgres $PG running"

    log "Server repo HEAD: $(git -C "$REPO" rev-parse HEAD) ($(git -C "$REPO" rev-parse --abbrev-ref HEAD))"
    for svc in $SERVICES; do
        if [ -n "$(git -C "$REPO" status --porcelain -- "backend/$svc-service")" ]; then
            warn "server working tree has local changes in backend/$svc-service (they are captured in the rollback tarball and will be overwritten):"
            git -C "$REPO" status --porcelain -- "backend/$svc-service" | head -20 | sed 's/^/    /'
        fi
    done

    log "Target containers:"
    for svc in $SERVICES; do
        c="$(container_of "$svc")"
        docker inspect "$c" >/dev/null 2>&1 || die "container $c not found"
        docker inspect -f '  {{.Name}} status={{.State.Status}} restarts={{.RestartCount}} started={{.State.StartedAt}} image={{.Config.Image}}' "$c"
        docker inspect -f '    cmd={{json .Config.Cmd}}' "$c"
        local proj wd
        proj="$(docker inspect -f '{{index .Config.Labels "com.docker.compose.project"}}' "$c")"
        wd="$(docker inspect -f '{{index .Config.Labels "com.docker.compose.project.working_dir"}}' "$c")"
        log "    compose project=$proj working_dir=$wd"
        [ "$proj" = "$COMPOSE_PROJECT" ] || die "$c belongs to compose project '$proj', expected $COMPOSE_PROJECT"
        [ "$wd" = "$BACKEND" ] || die "$c was created from '$wd', expected $BACKEND"
        [ -d "$BACKEND/$svc-service" ] || die "missing $BACKEND/$svc-service"
        log "    service dir size: $(du -sh "$BACKEND/$svc-service" | cut -f1)"
    done

    case " $SERVICES " in *" audit "*) log "audit /root/uploads files: $(uploads_count)" ;; esac

    dbs="$(list_dbs | tr '\n' ' ')"
    log "Databases (all are counted before/after): $dbs"
    log "Latest existing backup: $(ls -1t "$BACKUP_DIR"/*.gz 2>/dev/null | head -1 || echo none)"

    # Compose override: validate and show effective commands.
    if [ "$MODE" = real ]; then
        mkdir -p "$DEPLOY_ROOT"
        write_override "$OVERRIDE.new"
        if [ -f "$OVERRIDE" ] && ! cmp -s "$OVERRIDE" "$OVERRIDE.new"; then cp "$OVERRIDE" "$DIR/previous-noseed-override.yml"; fi
        mv "$OVERRIDE.new" "$OVERRIDE"
        OVERRIDE_IN_USE="$OVERRIDE"
    else
        OVERRIDE_IN_USE="$(mktemp --suffix=.yml /tmp/auditsphere-noseed.XXXXXX)"
        write_override "$OVERRIDE_IN_USE"
    fi
    compose config --quiet || die "compose config with override is invalid"
    log "Compose override valid ($OVERRIDE_IN_USE). Effective commands after deploy:"
    local cfg; cfg="$(compose config --format json 2>/dev/null || true)"
    for svc in $SERVICES; do
        if [ -n "$cfg" ]; then
            printf '  %s-service: %s\n' "$svc" "$(printf '%s' "$cfg" | jq -r --arg s "$svc-service" '.services[$s].command | if type=="array" then join(" ") else tostring end')"
        fi
    done
    if [ "$MODE" != real ]; then rm -f "$OVERRIDE_IN_USE"; fi
}

main() {
    set -euo pipefail
    SHA="$1"; TS="$2"; MODE="$3"; shift 3
    SERVICES="$*"
    DIR="$DEPLOY_ROOT/$TS"
    PAYLOAD="$DIR/payload.tgz"
    UPLOADS_BEFORE=none
    OVERRIDE_IN_USE=""

    if [ "$MODE" = real ]; then
        [ -d "$DIR" ] || die "$DIR missing"
        echo $$ > "$DIR/pid"
        trap 'echo $? > "$DIR/exit-code"' EXIT
        log "AuditSphere backend deploy $TS  sha=$SHA  services: $SERVICES  host=$(hostname)"
    fi

    preflight

    if [ "$MODE" != real ]; then
        log "--- Dry run: nothing was changed on the server ---"
        return 0
    fi

    # ---- payload validation ----
    log "--- Validating payload ---"
    [ -f "$PAYLOAD" ] || die "payload $PAYLOAD missing"
    [ "$(tar -xzOf "$PAYLOAD" DEPLOY_SHA)" = "$SHA" ] || die "payload DEPLOY_SHA does not match $SHA"
    local re svc bad list
    list="$(tar -tzf "$PAYLOAD")" || die "cannot list payload"
    re="^(DEPLOY_SHA$|BINARY_SHA256$"
    for svc in $SERVICES; do re="$re|$svc-service(/|$)"; done
    re="$re)"
    bad="$(grep -vE "$re" <<<"$list" || true)"
    [ -z "$bad" ] || die "payload has unexpected entries: $(echo "$bad" | head -5 | tr '\n' ' ')"
    bad="$(grep -E '(^/|(^|/)\.\.(/|$))' <<<"$list" || true)"
    [ -z "$bad" ] || die "payload has absolute or .. paths"
    for svc in $SERVICES; do
        grep -qx "$svc-service/$svc" <<<"$list" || die "payload lacks binary $svc-service/$svc"
    done
    log "Payload OK ($(du -h "$PAYLOAD" | cut -f1))"

    # ---- database backup ----
    log "--- Backing up all databases (pg_dumpall) ---"
    mkdir -p "$BACKUP_DIR"
    BACKUP="$BACKUP_DIR/pg_dumpall-$TS-pre-backend-deploy.sql.gz"
    if ! docker exec "$PG" pg_dumpall -U postgres | gzip > "$BACKUP.partial"; then
        die "pg_dumpall failed; nothing was changed"
    fi
    gzip -t "$BACKUP.partial" || die "backup gzip is corrupt; nothing was changed"
    local trailer
    trailer="$(gzip -dc "$BACKUP.partial" | tail -n 5)"
    case "$trailer" in
        *"cluster dump complete"*) ;;
        *) die "backup does not end with 'cluster dump complete'; nothing was changed" ;;
    esac
    mv "$BACKUP.partial" "$BACKUP"
    chmod 600 "$BACKUP"
    log "Backup OK: $BACKUP ($(du -h "$BACKUP" | cut -f1))"

    log "--- Snapshotting row counts ---"
    snapshot_rows "$DIR/rows-before.txt"
    log "Counted $(wc -l < "$DIR/rows-before.txt" | tr -d ' ') tables -> $DIR/rows-before.txt"
    snapshot_containers "$DIR/containers-before.txt"

    # ---- rollback copies ----
    log "--- Rollback copies ---"
    local img
    for svc in $SERVICES; do
        tar -czf "$DIR/rollback-$svc-service.tgz" -C "$BACKEND" "$svc-service"
        img="$(docker inspect -f '{{.Image}}' "$(container_of "$svc")")"
        docker tag "$img" "auditsphere-rollback/$svc-service:$TS"
        log "  $svc-service -> $DIR/rollback-$svc-service.tgz ($(du -h "$DIR/rollback-$svc-service.tgz" | cut -f1)), image tagged auditsphere-rollback/$svc-service:$TS"
    done

    # ---- deploy, one service at a time ----
    local deployed="" failed=""
    for svc in $SERVICES; do
        if deploy_one "$svc"; then
            deployed="$deployed $svc"
        else
            failed="$svc"
            log "Deploy of $svc-service FAILED - not deploying further services"
            break
        fi
    done

    # ---- verification ----
    log "--- Row count comparison ---"
    snapshot_rows "$DIR/rows-after.txt"
    local rc=0
    if awk -F'|' '
        NR == FNR { bc[$1] = $2; bh[$1] = $3; next }
        { ac[$1] = $2; ah[$1] = $3 }
        END {
            bad = 0; same = 0; n = 0
            for (k in bc) {
                n++
                if (!(k in ac))              { print "MISSING  " k " (had " bc[k] " rows)"; bad++ }
                else if (ac[k] + 0 < bc[k] + 0) { print "LOST     " k " " bc[k] " -> " ac[k]; bad++ }
                else if (ac[k] + 0 > bc[k] + 0) { print "GREW     " k " " bc[k] " -> " ac[k] }
                else if (ah[k] != bh[k])     { print "CHANGED  " k " (same count " ac[k] ", content differs)" }
                else same++
            }
            for (k in ac) if (!(k in bc)) print "NEW      " k " " ac[k] " rows"
            printf "SUMMARY  tables_before=%d identical=%d problems=%d\n", n, same, bad > "/dev/stderr"
            exit (bad > 0)
        }' "$DIR/rows-before.txt" "$DIR/rows-after.txt" > "$DIR/rows-compare.txt" 2> "$DIR/rows-summary.txt"; then
        :
    else
        rc=1
    fi
    sort "$DIR/rows-compare.txt" | sed 's/^/  /'
    sed 's/^/  /' "$DIR/rows-summary.txt"
    [ "$rc" -eq 0 ] && log "No table lost rows." || log "ROW LOSS DETECTED - see above; backup: $BACKUP"
    log "(GREW/CHANGED/NEW can be live user activity or additive AutoMigrate columns; LOST/MISSING are problems.)"

    log "--- Other containers (must be untouched) ---"
    snapshot_containers "$DIR/containers-after.txt"
    local targets=" "
    for svc in $SERVICES; do targets="$targets$(container_of "$svc") "; done
    local cdiff
    cdiff="$(awk -F'|' -v t="$targets" '
        NR == FNR { b[$1] = $2 "|" $3 "|" $4; next }
        index(t, " " $1 " ") { next }
        { seen[$1] = 1
          if (!($1 in b)) print "NEW      " $1
          else if (b[$1] != $2 "|" $3 "|" $4) print "CHANGED  " $1 ": " b[$1] " -> " $2 "|" $3 "|" $4 }
        END { for (k in b) if (!(k in seen) && !index(t, " " k " ")) print "GONE     " k }
        ' "$DIR/containers-before.txt" "$DIR/containers-after.txt")"
    if [ -n "$cdiff" ]; then
        echo "$cdiff" | sed 's/^/  /'; log "UNEXPECTED CHANGES TO OTHER CONTAINERS"; rc=1
    else
        log "All other containers unchanged (same id, StartedAt and status; $(wc -l < "$DIR/containers-before.txt" | tr -d ' ') containers total incl. targets)"
    fi

    [ -z "$failed" ] || rc=1

    {
        echo "timestamp=$TS"
        echo "sha=$SHA"
        echo "services_requested=$SERVICES"
        echo "services_deployed=${deployed# }"
        echo "service_failed=$failed"
        echo "previous_repo_head=$(git -C "$REPO" rev-parse HEAD)"
        echo "backup=$BACKUP"
        echo "rollback_dir=$DIR"
        echo "override=$OVERRIDE"
        echo "result=$([ "$rc" -eq 0 ] && echo success || echo FAILED)"
    } > "$DEPLOY_ROOT/last-deploy"

    log "--- Summary ---"
    log "Deployed      : ${deployed# }"
    [ -z "$failed" ] || log "FAILED        : $failed (rolled back; later services not deployed)"
    log "DB backup     : $BACKUP"
    log "Rollback dir  : $DIR  (rollback-<svc>-service.tgz, audit-uploads/, rows-*.txt)"
    log "Rollback image: auditsphere-rollback/<svc>-service:$TS"
    log "Override      : $OVERRIDE"
    log "Record        : $DEPLOY_ROOT/last-deploy"
    log "Result        : $([ "$rc" -eq 0 ] && echo SUCCESS || echo FAILED)"
    return "$rc"
}

main "$@" < /dev/null
REMOTE
}

# -----------------------------------------------------------------------------
# Local part
# -----------------------------------------------------------------------------
echo ""
echo "AuditSphere backend PRODUCTION deploy"
echo "  host     : $VPS_HOST ($REMOTE_REPO)"
echo "  ref      : $REF"
echo "  services : $SERVICES"
echo "  mode     : $([ "$DRY_RUN" -eq 1 ] && echo DRY RUN || echo REAL)"

section "Resolving ref"
git -C "$REPO_ROOT" fetch --quiet origin
SHA="$(git -C "$REPO_ROOT" rev-parse --verify "$REF^{commit}")" || die "cannot resolve $REF"
echo "  $REF = $SHA"
echo "  (building from git archive $SHA; the local working tree is NOT used)"

section "Server state"
SERVER_HEAD="$(rssh "git -C $REMOTE_REPO rev-parse HEAD")" || die "ssh to $VPS_HOST failed"
echo "  server repo HEAD: $SERVER_HEAD"

SVC_PATHS=""
for s in $SERVICES; do SVC_PATHS="$SVC_PATHS backend/$s-service"; done

section "What will ship ($SERVER_HEAD..$SHA, selected services only)"
if git -C "$REPO_ROOT" cat-file -e "$SERVER_HEAD^{commit}" 2>/dev/null; then
    # shellcheck disable=SC2086
    COMMITS="$(git -C "$REPO_ROOT" log --format='  %h %an %ad %s' --date=short "$SERVER_HEAD..$SHA" -- $SVC_PATHS)"
    echo "${COMMITS:-  (no commits touching these services)}"
    # shellcheck disable=SC2086
    STAT="$(git -C "$REPO_ROOT" diff --stat "$SERVER_HEAD" "$SHA" -- $SVC_PATHS)"
    echo "${STAT:-  (no source differences - containers are only rebuilt/recreated with migrate-only commands)}"
    if ! git -C "$REPO_ROOT" merge-base --is-ancestor "$SERVER_HEAD" "$SHA"; then
        echo "  WARNING: server HEAD is not an ancestor of $SHA (the diff above may include reverts)"
    fi
else
    echo "  WARNING: server HEAD $SERVER_HEAD not present locally; cannot show diff"
fi

section "Building (linux/amd64) from git archive"
mkdir -p "$TMP_DIR/src"
git -C "$REPO_ROOT" archive "$SHA" backend | tar -x -C "$TMP_DIR/src"
for s in $SERVICES; do
    cmd_pkg="./cmd"
    [ "$s" = auth ] && cmd_pkg="./cmd/auth"
    printf '  %-10s ' "$s"
    (cd "$TMP_DIR/src/backend/$s-service" &&
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "./$s" "$cmd_pkg")
    file "$TMP_DIR/src/backend/$s-service/$s" | grep -q 'ELF 64-bit.*x86-64' || die "$s binary is not linux/amd64"
    echo "ok ($(du -h "$TMP_DIR/src/backend/$s-service/$s" | cut -f1))"
done
echo "$SHA" > "$TMP_DIR/src/backend/DEPLOY_SHA"
: > "$TMP_DIR/src/backend/BINARY_SHA256"
PAYLOAD_DIRS=""
for s in $SERVICES; do
    PAYLOAD_DIRS="$PAYLOAD_DIRS $s-service"
    sum="$(shasum -a 256 "$TMP_DIR/src/backend/$s-service/$s" | cut -d' ' -f1)"
    echo "$sum $s" >> "$TMP_DIR/src/backend/BINARY_SHA256"
    echo "  sha256 $s: $sum"
done
# shellcheck disable=SC2086
tar --no-xattrs --no-mac-metadata -czf "$TMP_DIR/payload.tgz" -C "$TMP_DIR/src/backend" DEPLOY_SHA BINARY_SHA256 $PAYLOAD_DIRS
echo "  payload: $(du -h "$TMP_DIR/payload.tgz" | cut -f1)"
TAR_WARN="$(rssh 'tar -tzf - 2>&1 >/dev/null' < "$TMP_DIR/payload.tgz")" || die "server tar cannot read payload: $TAR_WARN"
[ -z "$TAR_WARN" ] || die "server tar warns about payload: $TAR_WARN"
echo "  payload readable by server tar (streamed, not stored)"

remote_script > "$TMP_DIR/deploy-remote.sh"

section "Remote pre-flight (read-only)"
rssh "bash -s -- $SHA $TS dry $SERVICES" < "$TMP_DIR/deploy-remote.sh" || die "pre-flight failed"

if [ "$DRY_RUN" -eq 1 ]; then
    echo ""
    echo "Dry run complete. Nothing was uploaded or changed on the server."
    exit 0
fi

if [ "$ASSUME_YES" -ne 1 ]; then
    [ -t 0 ] || die "not a terminal; pass --yes to deploy non-interactively"
    echo ""
    printf "Deploy %s (%s) of [%s] to PRODUCTION? Type 'deploy' to continue: " "$REF" "${SHA:0:7}" "$SERVICES"
    read -r answer
    [ "$answer" = deploy ] || die "aborted"
fi

REMOTE_DIR="$REMOTE_DEPLOY_ROOT/$TS"
section "Uploading to $REMOTE_DIR"
rssh "mkdir -p $REMOTE_DIR && chmod 700 $REMOTE_DEPLOY_ROOT $REMOTE_DIR"
scp -q "${SSH_OPTS[@]}" "$TMP_DIR/payload.tgz" "$TMP_DIR/deploy-remote.sh" "$VPS_HOST:$REMOTE_DIR/"
echo "  uploaded payload.tgz and deploy-remote.sh"

section "Deploying (runs detached on the server; log: $REMOTE_DIR/deploy.log)"
rssh "nohup setsid bash $REMOTE_DIR/deploy-remote.sh $SHA $TS real $SERVICES > $REMOTE_DIR/deploy.log 2>&1 < /dev/null &"

set +e
rssh "bash -s -- $REMOTE_DIR" <<'FOLLOW'
d="$1"
i=0
while [ ! -f "$d/pid" ] && [ "$i" -lt 30 ]; do sleep 1; i=$((i + 1)); done
[ -f "$d/pid" ] || { echo "remote deploy did not start (no pid file)"; exit 4; }
pid="$(cat "$d/pid")"
tail -n +1 -f "$d/deploy.log" &
tp=$!
while [ ! -f "$d/exit-code" ] && kill -0 "$pid" 2>/dev/null; do sleep 2; done
sleep 1
kill "$tp" 2>/dev/null; wait "$tp" 2>/dev/null
[ -f "$d/exit-code" ] || { echo "remote deploy process ended without an exit code"; exit 5; }
exit "$(cat "$d/exit-code")"
FOLLOW
RC=$?
set -e

if [ "$RC" -eq 255 ]; then
    echo ""
    echo "Lost the SSH connection while following the deploy. It keeps running on the server."
    echo "Re-attach: ssh $VPS_HOST 'tail -f $REMOTE_DIR/deploy.log; cat $REMOTE_DIR/exit-code'"
    exit 3
fi
echo ""
[ "$RC" -eq 0 ] && echo "Deploy finished successfully." || echo "Deploy FAILED (exit $RC). See $REMOTE_DIR/deploy.log on the server."
exit "$RC"
