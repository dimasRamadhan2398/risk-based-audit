# AuditSphere production deploy: target structure

Status: plan, 2026-10-27. Written from `.github/workflows/deploy.yml`, `backend/deploy-prod.sh`, `backend/docker-compose.prod.yml`, and the root and `backend/` wrapper scripts as they are on the checked-out branch (`staging`). Nothing here is implemented yet. `auditsphere-devops` reconciles its tidy-up branch against this plan.

## What exists today, and what is wrong with it

- `deploy.yml` is manual and calls `backend/deploy-prod.sh --yes --ref <ref> <services>`. That script is the only path that is safe today: it builds from `git archive`, takes a `pg_dumpall` backup, runs migrate-only commands through an override kept on the VPS, deploys one service at a time with `--no-deps`, and compares row counts.
- Two other prod paths still exist. The root `deploy-prod.sh` (aliased by root `deploy.sh`) runs `backend/deploy-backend.sh`, which does `docker compose -f docker-compose.prod.yml up -d --build` with no no-seed override. The prod compose file runs `seed` for auth, audit and master, so that path overwrites production rows. It then runs `frontend/deploy-prod.sh`, which force-recreates the frontend container. These must be retired or fenced off.
- `backend/deploy-prod-serverbuild.sh` is a near-copy of `deploy-prod.sh` that builds inside a Go container on the server. Nothing in the repo calls it. It is a drift risk.
- A second repo also deploys prod (from memory notes, not verified here).
- ClamAV is not in `backend/docker-compose.prod.yml` on this checkout. It exists in commit `dffc8f6` and in a worktree.

## 1. Stacks and ownership

| Stack / container | Source | Who changes it | Owner step | Deploy may recreate it? |
|---|---|---|---|---|
| `rb_audit_postgres` (volume `postgres_data`) | compose project `rb_audit_backend` | devops, by hand | none | Never |
| redis, zookeeper, kafka | same project | devops, by hand | none | Never |
| auth, audit, master, risk, analytics (`rb_audit_<svc>_service`) | `backend/<svc>-service` | this workflow | `deploy-prod.sh` | Yes, one at a time, `--no-deps` |
| clamav (`rb_audit_clamav`) | same project, once merged | devops | ClamAV step (section 4) | Only in its own step |
| python-ai | same project | devops | none (not in `ALL_SERVICES`) | Never |
| kong (`rb_audit_kong`) | compose + `scripts/sync-kong.sh` | devops | Kong step, separate | Never from backend deploy |
| frontend (auditsphere.app) | `frontend/deploy-prod.sh` | manual today | open decision 1 | Not by backend deploy |
| tenant stacks `backend/tenants/<slug>/` | `generate-tenant-stack.sh`, `onboard-tenant.sh` | devops, per tenant | tenant step (section 5) | Never from prod deploy |
| `*_dev` containers | dev compose | dev workflow | none | Never |

## 2. Stateful data a deploy must not destroy

| Data | Where it lives | Backup today | Gap and required fix |
|---|---|---|---|
| Postgres: prod DBs and all tenant DBs (`rb_audit_<svc>_<slug>`) | volume `postgres_data`, one container | `pg_dumpall` to `/root/db-backups/` before every deploy, checked for the `cluster dump complete` trailer | On-host only. Platform must add an off-host copy and a retention rule. |
| Audit uploads | volume `audit_uploads` mounted at `/root/uploads` | `docker cp` into `/root/auditsphere-deploy/<ts>/audit-uploads`, restored after recreate | Same host, no off-host copy. Covered by the script. |
| Master uploads (`./uploads/qar-reports`, `./uploads/consulting`) | No volume. Relative path resolves under `/root` in the container's writable layer | None | Lost when the container is recreated. The script does not copy it. This blocks deploying `master`. Fix in section 6. |
| ClamAV signatures | volume `clamav_db` at `/var/lib/clamav` | None (re-downloadable) | Exclude from backup. Expect a slow first start while signatures download. |
| Redis, Kafka, Zookeeper, python-ai models | `redis_data`, `kafka_data`, `zookeeper_data`, `python_ai_models` | None | Not reviewed here. Devops to confirm each is cache or rebuildable before excluding it. |
| Rollback copies | `/root/auditsphere-deploy/<ts>/` (source tarball, image tag, row snapshots) | Kept on host | Needs a retention rule. The platform owner decides how many to keep. |

## 3. Per-service deploy sequence

Run `auth`, `audit`, `master`, `risk`, `analytics` in that order, with stop on first failure. The workflow takes one ref and one service subset.

1. Gate: the `production` environment requires approval. Resolve the ref to a SHA. Use a commit that has been on staging and passed QA.
2. Preflight (read-only dry run first): disk under 90%, Postgres running, lock held, each target container in compose project `rb_audit_backend` with working dir `/app/rbia-repo/backend`, ClamAV healthy if the service needs it (section 4), and the effective commands contain no seeder.
3. Backup: `pg_dumpall` must end with `cluster dump complete`. Abort if not. Snapshot row counts and other containers' state.
4. Rollback copy: source tarball and image tag `auditsphere-rollback/<svc>-service:<ts>`. Record the SHA of the currently running deploy.
5. Uploads: copy out for both `audit` and `master`. Abort if the copy fails.
6. Build from `git archive <SHA>` for linux/amd64. Ship only that service directory.
7. Recreate: `docker compose -f docker-compose.prod.yml -f <no-seed override> up -d --no-deps --build <svc>-service`. The container's command runs migrate, then serve (`auth`, `audit`, `master`: `migrate up` or `migrate`; `risk`: `risk up`). Analytics has no migrate step. Migrations are additive-only by convention; see the rollback notes.
8. Health: `/health` returns 200 twice in a row within 150 seconds (analytics: any HTTP answer). The running binary's sha256 matches the built one.
9. Restore uploads and check the file count is not lower than before.
10. Row counts: any `LOST` or `MISSING` table fails the deploy. Any change to other containers fails the deploy.
11. Record `sha=` and result in `last-deploy`, and append the SHA to a success history file (new, section 6).

Rollback, in order of preference:

- Automatic, per service: a failed health check or binary check restores the previous source tarball and rebuilds. The schema is not reverted, so this is safe only when the failed ref added no destructive migration.
- Manual, to the previous good SHA: rerun this workflow with `ref=<previous good SHA>` and the same service list. This is safe only if every migration since that SHA is additive. The current `last-deploy` `previous_repo_head` field is the server checkout's HEAD, not a deployed SHA, so it cannot be used for this. The history file in step 11 fixes that.
- Data restore: restore the `pg_dumpall` file. This rolls back every database, including all tenants, to the backup time and discards all later writes. It requires stopping the app services. It is a decision for the owner, never an automatic step.

## 4. ClamAV rules

- ClamAV is its own step, not a member of `ALL_SERVICES`. Start it with `up -d --no-deps clamav`, then wait for `PONG` on port 3310 before any `audit` or `master` recreate.
- `audit` and `master` run with `CLAMAV_MODE=required`. If ClamAV is down, uploads fail even though `/health` passes, so the health check does not catch it. The preflight must check ClamAV directly.
- No service lists ClamAV in `depends_on`. Keep it that way. A `--no-deps` recreate must never restart ClamAV as a side effect.
- Order of rollout: ship the ClamAV service and its volume first, let signatures finish downloading, then ship the `audit` and `master` refs that set `CLAMAV_MODE=required`.
- `auth`, `risk` and `analytics` never depend on ClamAV.
- Rolling `audit` or `master` back to a pre-ClamAV SHA is safe. ClamAV can keep running unused.
- Capacity: confirm the VPS has room for the 1.5 GB memory limit before the first start.

## 5. Tenant stacks

- Tenant stacks live in `backend/tenants/<slug>/`. They are generated by `generate-tenant-stack.sh` and started by `onboard-tenant.sh`. Each has its own compose file, its own JWT secret, its own databases on the shared Postgres, its own Redis DB index, and its own ports, all from `backend/tenants/registry.tsv`.
- The prod deploy must never run the generate or onboard scripts, write the registry, or change tenant containers. The existing "other containers unchanged" check in `deploy-prod.sh` enforces this for every non-target container. It also catches an unplanned Postgres restart.
- The prod JWT secret (`/root/auditsphere-deploy/jwt-secret`) must not be copied into any tenant. Each tenant keeps its own secret.
- Tenant DBs are inside the prod `pg_dumpall`. A data restore for prod therefore also rolls back every tenant.
- Not verified: whether tenant stacks build from the same `backend/<svc>-service` directories. If they do, a prod deploy does not change tenants, but a tenant rebuild picks up new code. Devops to confirm before the tenant step is written.
- Tenant migrations are not run by the prod deploy. Tenant schema changes happen only through their own stack restart or re-onboard.

## 6. Open decisions for the user

1. Frontend: keep it out of this workflow. Recommendation: keep it out for now, with its own manual or workflow step, pinned to a built SHA. Remove the frontend call from the root `deploy-prod.sh`.
2. Second repo's prod deploy: remove it. Recommendation: one prod path only.
3. Root `deploy-prod.sh`, `deploy.sh`, `backend/deploy-backend.sh`: retire. They run the seeding path. Recommendation: delete, or replace with a stub that prints "use deploy.yml".
4. `backend/deploy-prod-serverbuild.sh`: delete. Recommendation: delete, since nothing calls it.
5. Master uploads: add a named volume `master_uploads` in prod compose, and make the upload directory configurable by env (it is hard-coded to `./uploads` today). Recommendation: approve. Copy the existing files out of the master container before its first recreate with the new volume.
6. ClamAV in prod: ship now, or hold until the VPS capacity check is done. Recommendation: hold until section 4's capacity check passes.
7. Off-host backups of `/root/db-backups` and the upload copies, plus a retention rule. Owner: platform. Recommendation: approve.
8. Hard-coded fallback for `DATA_HUB_API_KEY` in `docker-compose.prod.yml`. Recommendation: remove the fallback and require it from the deploy env. Do not copy the value anywhere.
9. Success history file for deployed SHAs (step 11). Recommendation: approve. It is what makes a manual rollback possible.

## Hand-off

- `auditsphere-devops`: (a) remove the seeding path from the root wrappers and `deploy-backend.sh`, and delete `deploy-prod-serverbuild.sh`; (b) add the ClamAV step and preflight check; (c) extend the upload copy-out and restore to `master` once the volume exists; (d) add the SHA history file; (e) add a `--rollback-to <sha>` path or document the manual re-run; (f) confirm the tenant build question.
- `auditsphere-backend` (or `auditsphere-backend-2`): make the master upload directory configurable by env. No model or migration changes are needed.
- `auditsphere-platform`: off-host copy and retention for `/root/db-backups` and the upload copies; confirm the VPS-level backup of the Docker volumes; remove the second repo's prod deploy access if decision 2 is approved.

Smoke checks before calling the demo or prod ready:

- Dry run prints the effective command for every target. None contains `seed`. `auth`, `audit`, `master` contain `migrate`.
- Each service's `/health` returns 200 (analytics: any answer). `docker inspect` shows the no-seed command.
- Built and running sha256 match for each recreated service.
- `rows-compare.txt` has no `LOST` or `MISSING` lines. `containers-after` shows no unexpected changes.
- Backup file exists, is gzip-valid, and ends with `cluster dump complete`.
- Audit and master uploads file counts are equal to or higher than before.
- ClamAV answers `PONG` from inside the `audit` container. An upload of a test file is accepted, and an EICAR test file is rejected.
- Login through Kong succeeds (auth), and a tenant's health endpoint still responds after a prod deploy.
