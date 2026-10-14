# Execution progress derivation: backend deploy, backfill, frontend build

Scope: audit-service execution progress is now derived from the annual plan, assignment letter, fieldwork (test controls, mitigation plan), working paper (findings, action plans, recommendations) and LHA. Sync runs on save of each module (new hooks in `backend/audit-service/routes/handler.go`) instead of on page load. `POST /api/v1/audit-executions/sync` recalculates every execution once (the backfill).

Production target: auditsphere.app (VPS 202.10.34.166, repo `/app/rbia-repo`). Nothing in this runbook is run automatically. Each step needs an explicit go-ahead.

Hard rules:
- Backend: only via `backend/deploy-prod.sh` with an explicit `--ref`. Never `deploy.yml` (its `up --force-recreate` wipes `/root/uploads`) and never a whole-stack `docker compose up`. Never run the seeders (`seed`), they overwrite production rows.
- No schema migration is added. `fieldwork_test_controls.mitigation_plan` and `working_paper_plans` already exist and are AutoMigrated.
- The backfill is a WRITE. It rewrites `progress`, `status`, `status_detail` and `activity_id` on every `audit_executions` row and pushes status to `audit_activities`.

## 0. Pre-checks (local, read-only)

1. The ref to ship is pushed and contains the hooks and the new derivation:
   `git log --oneline origin/staging -5 -- backend/audit-service/routes/handler.go backend/audit-service/services/audit_execution_sync`
   Today `handler.go` has only the `POST /sync` route. The save hooks are not in the tree yet. Do not start if they are missing from the ref.
2. Tests pass for the changed packages:
   `cd backend/audit-service && go vet ./... && go test ./services/audit_execution_sync/... ./controllers/audit_execution/... ./services/audit_completion/...`
3. Frontend on the same ref removes the sync-on-load call (`grep -rn "audit-executions/sync" frontend` should show no call in page load code).
4. Announce the change to users (see section 7) before step 2.
5. Tools: `gh` is not installed here. Check release status by hand, not by guessing.

## 1. Backend deploy (audit-service only, with save hooks)

Dry run first. It makes no server changes, but it does SSH to the VPS and read container state:

```
cd /Users/a/Projects/risk-based-audit
backend/deploy-prod.sh --dry-run --ref origin/staging audit
```

Check the output: the server HEAD, the target container command (must show `migrate up && serve`, not `seed`), the compose project name, uploads file count, and the list of databases.

Real deploy. Do not pass `--yes`. The script asks you to type `deploy`, after showing the commit SHA:

```
backend/deploy-prod.sh --ref origin/staging audit
```

What the script does (so you know what to expect): validates the payload, runs `pg_dumpall` to `/root/db-backups/pg_dumpall-<TS>-pre-backend-deploy.sql.gz` and verifies it, snapshots row counts and hashes, copies `/root/uploads` out and back, recreates only `rb_audit_audit_service`, checks health and binary SHA256, and rolls back on failure. The row-count check exits non-zero on any LOST or MISSING table.

Stop if the row check reports LOST or MISSING, or if the script rolls back.

## 2. Backup and before-snapshot (before the backfill)

The `pg_dumpall` from step 1 is the full-cluster backup. Also take a targeted dump of the audit database and a CSV of the rows the backfill will change. Run this on the VPS, which is read-only for the database:

```
TS=$(date -u +%Y%m%dT%H%M%SZ)
ssh root@202.10.34.166 "docker exec rb_audit_postgres pg_dump -U postgres -Fc rb_audit_audit_service > /root/db-backups/rb_audit_audit_service-$TS-pre-backfill.dump && docker exec rb_audit_postgres pg_restore -l /root/db-backups/rb_audit_audit_service-$TS-pre-backfill.dump >/dev/null && echo dump-ok"

ssh root@202.10.34.166 "docker exec rb_audit_postgres psql -U postgres -d rb_audit_audit_service -X --csv -c \"SELECT id, ref, name, activity_id, progress, status, status_detail FROM audit_executions ORDER BY ref\" > /root/db-backups/audit_executions-before-$TS.csv"

ssh root@202.10.34.166 "docker exec rb_audit_postgres psql -U postgres -d rb_audit_audit_service -X --csv -c \"SELECT id, title, project_code, status FROM audit_activities ORDER BY id\" > /root/db-backups/audit_activities-before-$TS.csv"
```

Note: the `audit_activities` status column is named `status`. Confirm with `\d audit_activities` if the query errors.

Count before (write these numbers down):

```
ssh root@202.10.34.166 "docker exec rb_audit_postgres psql -U postgres -d rb_audit_audit_service -X -c \"SELECT status, count(*) FROM audit_executions GROUP BY 1 ORDER BY 1\""
```

## 3. Dry run of the backfill on a scratch copy (local)

Do this before production. Restore the targeted dump into a scratch database on the local Docker Postgres (container `rb_audit_postgres`, currently running), run the new audit-service against it, then compare.

The local `rb_audit_audit_service` database has 0 executions and 0 test controls, so it cannot prove anything. Use the restored prod dump. Whether the local service can be pointed at a different database name is not verified: `pkg/config/config.go` calls `viper.AutomaticEnv()`, but the env key mapping for `database.name` is not confirmed. Check that before relying on it.

## 4. Backfill on production (WRITE)

Go-ahead required. Run once, after step 1 is confirmed healthy. Use a token from an account with access to the executions module, held in an environment variable that is not echoed. Do not paste tokens into chat or into files.

Count before (section 2) must be recorded. Then:

```
curl -sS -X POST "https://auditsphere.app/api/v1/audit-executions/sync" \
  -H "Authorization: Bearer $AUDIT_TOKEN" -H "Content-Type: application/json" \
  -o /root/backfill-response.json -w "%{http_code}\n"
```

Confirm the public base URL first: the frontend proxies `/api/v1` to Kong, so this path is expected to work, but it is not verified from here. Unauthenticated, the endpoint should return 401.

The response has `synced` (count), `items` (per-execution progress and status) and `unmatched_source_rows` (rows whose letter reference matches no assignment letter). Save the response.

Count after: repeat the `GROUP BY status` query from section 2 and take the new CSV snapshot (same two commands with `-after-`).

## 5. Verification query (per-execution changes, before committing)

Load both CSVs into a local scratch database (not production) and list every change:

```
CREATE TEMP TABLE before_exec (id uuid, ref text, name text, activity_id uuid, progress int, status text, status_detail text);
\copy before_exec FROM 'audit_executions-before.csv' CSV HEADER
CREATE TEMP TABLE after_exec (LIKE before_exec);
\copy after_exec FROM 'audit_executions-after.csv' CSV HEADER

SELECT b.ref, b.name,
       b.progress AS progress_before, a.progress AS progress_after,
       b.status AS status_before, a.status AS status_after,
       b.status_detail AS detail_before, a.status_detail AS detail_after,
       (b.activity_id IS DISTINCT FROM a.activity_id) AS activity_linked_changed
FROM before_exec b FULL JOIN after_exec a ON a.id = b.id
WHERE b.id IS NULL OR a.id IS NULL
   OR b.progress IS DISTINCT FROM a.progress
   OR b.status IS DISTINCT FROM a.status
   OR b.status_detail IS DISTINCT FROM a.status_detail
   OR b.activity_id IS DISTINCT FROM a.activity_id
ORDER BY b.ref;
```

Rows with `id` missing on one side mean an execution was created or deleted during the run. Review every status change with the module owners before the backfill is accepted.

## 6. Frontend build (last)

Frontend is not built by CI or run by compose. Follow the manual procedure used for the earlier frontend deploys:

1. Export the ref: `git archive origin/staging frontend`.
2. Build in a local `node:20-alpine` container with NO `frontend/.env` present. A server-side `.env` bakes the wrong API URL into the build.
3. Rsync `.output` to the server. Do not build on the server (2 CPUs, 3.8 GB RAM, swap in use).
4. Dump the current container env, rename the running `rbia-frontend-prod` to `-old`, and start the new one on `rb_audit_network` with `--env-file`. The container is not compose-managed, so do not run `compose up` on it.
5. Health check. If it fails, swap the old container back.
6. Copy the new `_nuxt` chunks into `/var/www/auditsphere-chunk-archive/_nuxt/` so open tabs keep working.

## 7. Verification (smoke checklist)

- Create or edit each module (annual plan, assignment letter, fieldwork test control with mitigation, working paper finding/plan, LHA). The execution progress and status should update on save, with no page-load sync call in the network tab.
- Open the executions list. Progress and status match the snapshot diff for the rows you expected to change.
- `unmatched_source_rows` from the backfill response is zero, or each entry is explained.
- `docker logs --tail 200 rb_audit_audit_service` shows no `audit_execution_sync` errors.
- Uploads: `docker exec rb_audit_audit_service sh -c 'find /root/uploads -type f | wc -l'` matches the number recorded before step 1.

## 8. Rollback

Option A, targeted (preferred): restore the before values for the changed rows only. Load `audit_executions-before-<TS>.csv` into a temp table, then in a transaction run `UPDATE audit_executions e SET progress = b.progress, status = b.status, status_detail = b.status_detail, activity_id = b.activity_id FROM before b WHERE e.id = b.id;`. Do the same for `audit_activities.status` from its CSV. Review the row count before committing.

Option B, backend: redeploy the previous ref with the same script, `backend/deploy-prod.sh --ref <previous-sha> audit`. The previous image is also tagged `auditsphere-rollback/audit-service:<TS>`, and the script keeps `rollback-audit-service.tgz` under `/root/auditsphere-deploy/<TS>/`. Note that this does not undo data. Redeploying the old code does not restore old progress values, so use option A for data.

Option C, last resort: restore the targeted dump with `pg_restore --clean` into `rb_audit_audit_service`, with audit-service stopped. This discards every audit-service write made since the dump. Do not use it without the user's explicit decision. `pg_dumpall` restore affects every database and is worse.

## 9. Decisions required before production is touched

1. Environment: confirm the target is production (auditsphere.app) and that step 3 (scratch dry run) is accepted as the gate.
2. Timing: a window for the audit-service recreate (a few seconds of downtime) and for the backfill, with a named operator.
3. Scope of the backfill: confirm it may change `status`, `progress` and `activity_id` on existing executions, and push status to `audit_activities`. Review the step 5 diff with the module owners first.
4. Communication: whether existing executions whose status moves will be announced to users, and who writes the notice.
5. Frontend: whether the frontend ships in the same window as the backend, given that the old frontend's sync-on-load call still works against the new endpoint.
6. Uploads: whether to accept the current per-recreate copy-out (deploy-prod.sh) or first fix the missing volume.
