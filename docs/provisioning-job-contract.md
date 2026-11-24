# Tenant Provisioning Job — Shared Contract

Status: draft, 2026-11-24. Source of truth for the script, the provisioner, the auth-service proxy and the Site Generator UI. Change this file first if the contract has to change.

## Goal

The Site Generator UI starts a real provisioning job and shows real-time progress from the system doing the work, instead of a `setTimeout` animation and a copy-paste CLI command.

## Architecture

```
Browser (/site-generator, ADMIN only)
  └─ fetch + SSE ──> Kong ──> auth-service  /api/v1/provisioning/*   (JWT + role check, proxy)
                                └─ HTTP (X-Provisioner-Token) ──> provisioner  (Go binary on the VPS HOST, systemd)
                                                                   └─ exec (args array, no shell) ──> scripts/onboard-tenant.sh
```

- The provisioner runs on the host, not in a container, because the script needs docker, `/etc/nginx` and certbot. It listens only on 127.0.0.1 / the docker bridge address. It is never exposed publicly.
- One job runs at a time. Other jobs wait as `queued`. This also serializes writes to `backend/tenants/registry.tsv`.

## Steps (fixed order, 9 total)

| index | key | title |
|---|---|---|
| 1 | generate | Generate tenant stack |
| 2 | databases | Create databases |
| 3 | build | Build images |
| 4 | migrate | Migrate and seed |
| 5 | storage | Create storage |
| 6 | email | Provision email domain |
| 7 | start | Start containers |
| 8 | health | Wait for services to be healthy |
| 9 | ingress | Configure nginx and TLS |

A step that does not apply (for example email/ingress in local mode) is reported as `skipped` with a reason in `message`.

## Script output protocol (`scripts/onboard-tenant.sh --progress`)

With `--progress`, the script writes machine-readable lines to stdout in addition to its normal log lines:

```
@@PROGRESS {"key":"build","index":3,"total":9,"status":"running","message":"Building 6 images"}
@@PROGRESS {"key":"build","index":3,"total":9,"status":"done","message":"6 images built in 24s"}
@@RESULT {"slug":"bai","frontend_url":"https://bai.example","api_url":"https://api-bai.example","admin_username":"admin","admin_email":"cae@bai.co.id","admin_initial_password":"<random>"}
```

- `status` is one of `running | done | skipped | failed`.
- On failure the script emits `failed` for the current step with the error in `message`, then exits non-zero.
- Every other stdout/stderr line is a plain log line.
- `--dry-run` emits the full step sequence (with short sleeps) and a fake `@@RESULT`, with no side effects. It is for local UI testing only and must be labelled as dry run in every message.
- The script must never print the admin password on a log line. The provisioner masks it on lines after `@@RESULT` and masks `PASSWORD=…`/`SECRET=…`/`TOKEN=…`-style values everywhere, but it cannot know the value before `@@RESULT` arrives.

### Script invocation (provisioner → `onboard-tenant.sh`)

The provisioner runs `bash <repo>/scripts/onboard-tenant.sh` with cwd `<repo>` and this argv (one element per value, no shell), i.e. the options-first form from "Script CLI details" below:

```
--progress [--dry-run] [--local | --domain <fqdn>] [--fe-port N --api-port N]
  [--admin-name <name>] --admin-email <email> (--with-demo-data | --empty-data) [--skip-email]
  [--health-timeout N] [--allow-no-tls] -- <slug> <client_name>
```

- `--local` when mode=local; `--domain` only for production with a domain. Ports only when both were given. When omitted, the script/generator allocates them at run time (serialized by the queue), so two queued jobs never get the same "next" ports.
- `generate-tenant-stack.sh --next-ports` is run the same way (`bash <repo>/scripts/generate-tenant-stack.sh --next-ports`). The provisioner uses the last stdout line that starts with `{`.

### Script CLI details (added by the infra agent, 2026-11-24)

These refine the protocol above; none of them changes its line format.

- Invocation the provisioner should use: options first, then `--`, then the positionals, so a client name that starts with `-` can never be read as a flag:
  `onboard-tenant.sh --progress [--dry-run] [--local | --domain <fqdn>] [--fe-port N --api-port N] [--admin-name X] --admin-email Y [--with-demo-data | --empty-data] [--skip-email] -- <slug> <client_name>`
  The older form (positionals first, ports as 3rd/4th positional) still works.
- `--domain` must be `<slug>.<base>`; the API is always `api-<slug>.<base>`. Anything else fails step 1.
- `--skip-email` (alias `--no-email`) reports step 6 as `skipped`.
- Every message in a dry run starts with `[dry run] `. `--dry-run-fail <key>` (dry run only) makes that step report `failed` and exit 1, for testing the UI's failure path.
- A failure before step 1 has begun (unknown option, missing argument) is reported as `failed` on step 1 without a preceding `running` event.
- Step 6 (email) is never fatal: a Resend error is reported as `skipped` with a message starting "Email provisioning FAILED".
- Step 9 (ingress) in production fails when DNS does not resolve, certbot is missing or certbot fails. With `--allow-no-tls` the step is `skipped` instead, with a message starting "TLS NOT CONFIGURED", and the `@@RESULT` URLs use `http://`.
- `@@RESULT` is printed as the last line even without `--progress`, because it is the only place the initial password appears. `admin_email` is `null` when no `--admin-email` was given. `admin_initial_password` is `null` when the tenant's auth database already had an `admin` user (re-onboarding), since the seeder does not change an existing admin's password.
- `generate-tenant-stack.sh --next-ports` prints `{"fe_port":N,"api_port":N}` on one line (registry and listening host ports skipped, read-only). This is what `GET /ports/next` runs.

### Auth seeder interface (onboard step 4)

- The script runs `./auth seed` in the tenant's auth-service container with `ADMIN_PASSWORD` always set (random, 20 chars), and `ADMIN_NAME` / `ADMIN_EMAIL` when given. They are passed through the environment (`compose run -e NAME`), never on a command line.
- The seeder must create the `admin` user (username stays `admin`) only when it is missing, using `ADMIN_PASSWORD` for the hash, `ADMIN_NAME` for the full name and `ADMIN_EMAIL` for the email. It must not change an existing admin. When a variable is absent or empty, it falls back to the current default.
- It must not log the password.
- Sample users (auditor, dept_head, new_dept_head, auditee, viewer, all `password123`): seeded only when `SEED_DEMO_USERS=true`, or when `SEED_DEMO_USERS` and `ADMIN_PASSWORD` are both unset (the dev/demo compose stacks, unchanged). So a provisioned tenant (which always sets `ADMIN_PASSWORD`) gets only `admin`, unless the script passes `-e SEED_DEMO_USERS=true` with `--with-demo-data`. `onboard-tenant.sh` does this: it passes `SEED_DEMO_USERS=true` with `--with-demo-data` and `false` with `--empty-data`. Existing sample users on existing tenants are not removed.
- Implemented in `backend/auth-service/pkg/database/seeders/seeder.go`; `./auth seed --help` mentions `ADMIN_PASSWORD` and `SEED_DEMO_USERS`.
- `./auth seed --help` must mention `ADMIN_PASSWORD`. The script checks for that string to detect an older image. On an older image, onboarding a new admin fails at step 4 instead of creating the shared default password.

## Provisioner internal API (host, `X-Provisioner-Token` header required)

| Method | Path | Purpose |
|---|---|---|
| POST | /jobs | Create a job. Returns 202 + Job. |
| GET | /jobs | List recent jobs (newest first, without secrets). |
| GET | /jobs/{id} | Get one Job. |
| GET | /jobs/{id}/events?since={seq} | SSE stream. Replays events after `seq`, then streams live, and ends after the final `status` event. |
| GET | /ports/next | Next free `{ "fe_port", "api_port" }` according to the registry and the host OS. |
| GET | /healthz | Liveness. |

### POST /jobs body

```json
{
  "client_name": "PT Bank Arta Indonesia",
  "slug": "bai",
  "mode": "local | production",
  "domain": "bai.example.com",
  "fe_port": 3015,
  "api_port": 8095,
  "admin_name": "Siti CAE",
  "admin_email": "cae@bai.co.id",
  "with_demo_data": false,
  "provision_email": true,
  "allow_no_tls": false,
  "health_timeout_seconds": 120,
  "dry_run": false
}
```

Validation (both UI and provisioner, and the script remains the final check):
- `slug`: `^[a-z][a-z0-9]{1,30}$`, and not already in the registry.
- `client_name`: 2–120 chars. It is passed as an argument, never interpolated into a shell string.
- Ports, if given, must be free in the registry and the OS. If they are omitted, the script allocates them at run time (the next free pair at that moment; `/ports/next` is only a preview for the form and reserves nothing).
- `admin_email`: valid email.

Clarifications implemented by the provisioner (backend, 2026-11-24):
- `client_name` and `admin_name` may contain quotes, `$()`, backticks, `;` etc. (they are data). They must not contain control characters / line breaks (they would corrupt `registry.tsv` and log lines) and must not start with `-` (the script would read them as a flag). `admin_name` is optional, max 120 chars.
- `mode` is required: `local` or `production`. `domain` is optional, a lowercase hostname of the form `<slug>.<base>` (the script requires that), and ignored in local mode.
- `fe_port` and `api_port` are given together or both omitted, 1024–65535, and different. "Free" means not in `registry.tsv` (fe/kong/kong-admin columns), not held by a queued/running job, and bindable on the host.
- `provision_email` defaults to `true` when omitted. `slug` is also refused while a queued/running job holds it.
- `allow_no_tls` (bool, default `false`): production only. `true` with `mode: "local"` is **rejected** with 400 (`fields.allow_no_tls`); local mode never configures TLS anyway (step 9 is always `skipped`). In production, `true` passes `--allow-no-tls`: a DNS or certbot problem at step 9 makes the step `skipped` ("TLS NOT CONFIGURED ...") instead of failing the job, and the tenant is served on plain HTTP until someone runs the certbot command shown in the step message.
- `health_timeout_seconds` (int, optional, 30–900): passed as `--health-timeout <n>` (seconds; how long step 8 waits for Kong and auth-service). Omitted or `null` means no flag, so the script default of 120 s applies. Outside 30–900 is a 400 (`fields.health_timeout_seconds`); the script itself would accept 1–1800. A non-integer is a 400 from JSON decoding.
- Script argv order: `--progress [--dry-run] [--local | --domain D] [--fe-port N --api-port N] [--admin-name X] --admin-email Y (--with-demo-data | --empty-data) [--skip-email] [--health-timeout N] [--allow-no-tls] -- <slug> <client_name>`.
- Retry (no flag needed): a slug that is in `registry.tsv` is accepted again when the latest real (non dry-run) job this provisioner ran for it is `failed` and no job for it is queued/running. The failed run left the registry row behind and the script is idempotent. The new job carries `"retry_of": "<failed job id>"`, and the ports that slug already holds in the registry count as free for it (registry and OS check). A registry slug that this provisioner never ran, or whose latest job succeeded, stays a 409.
- `created_by` is required; auth-service always overwrites it with the JWT `user_id`.
- Errors: 400 `VALIDATION_FAILED` or 409 `CONFLICT` (slug/port taken), body `{"success":false,"error":{"code","message"},"data":{"fields":{"<field>":"<reason>"}}}`.

### Job

```json
{
  "id": "uuid",
  "status": "queued | running | succeeded | failed",
  "slug": "bai",
  "mode": "production",
  "dry_run": false,
  "created_by": "user id/email from auth-service",
  "created_at": "RFC3339", "started_at": "RFC3339|null", "finished_at": "RFC3339|null",
  "current_step": "build",
  "steps": [
    {"key":"generate","index":1,"title":"Generate tenant stack","status":"done","message":"...","started_at":"...","finished_at":"..."}
  ],
  "error": null,
  "result": {"frontend_url":"...","api_url":"...","admin_username":"admin","admin_email":"...","admin_initial_password":"...","tls":true}
}
```

- `admin_initial_password` is kept in memory only and never written to disk or logs. It is returned only to the job's creator, and only until it has been read once (after that, the field is `null`).
  - The creator is identified on reads by the `X-Provisioner-User` header, which auth-service sets to the JWT `user_id`. The first `GET /jobs/{id}` by the creator after `@@RESULT` returns it; every later read (and every read by anyone else, `GET /jobs`, SSE) returns `null`. A provisioner restart also loses it. The UI must show it from that one response.
- `result.tls` is `false` when the tenant is plain HTTP: local mode, or production with step 9 `skipped` under `allow_no_tls`. When step 9 is `skipped`, the provisioner also forces `result.frontend_url` and `result.api_url` to `http://` (the script already reports `http` then; this only guards the contract). The UI should show a "TLS not configured" warning with the step 9 message when `tls` is `false` in production mode.
- `steps[]` starts with all 9 steps in status `pending` (provisioner-only state; the script never reports it).
- Extra fields the provisioner also returns: `client_name`, `domain`, `fe_port`, `api_port` (as requested; `null` when the script allocates them), `admin_name`, `admin_email`, `with_demo_data`, `provision_email`, `allow_no_tls`, `health_timeout_seconds` (`null` = script default 120 s), `exit_code`. `result.slug` is passed through from `@@RESULT`.
- `GET /jobs` returns `{"jobs":[Job, ...]}`, newest first, `?limit=` 1–500 (default 50).
- `GET /healthz` does not require the token (it returns only `{"status":"ok"}`), so systemd/monitoring can probe it.
- If the script exits 0 without `@@RESULT`, the job is `succeeded` with `result: null` and a warning log event. If it exits non-zero, `error` is `"<step title>: <message>"` from the failed step, or `"onboard-tenant.sh exited with code N"`. A `result` received before a failure is kept.
- Jobs and their event logs are persisted as files under the provisioner's data dir, so a restart keeps the history. A job that was `running` when the provisioner restarted becomes `failed` with error `"provisioner restarted"`.

### SSE events

```
id: 17
event: progress          # data = step object (same shape as steps[])
event: log               # data = {"line":"...","stream":"stdout|stderr","ts":"..."}
event: status            # data = {"status":"running|succeeded|failed","error":null}
event: result            # data = result object (password excluded; fetch it via GET /jobs/{id})
```

- `id` is a per-job sequence starting at 1. `?since=N` (or the `Last-Event-ID` header) replays events with id > N.
- The first event is `status: running` when the job starts; a queued job's stream only carries heartbeats (`: heartbeat` comment every 15s) until then. At the end: any remaining `progress`, then `result` (if any), then the final `status`, then the stream closes.
- `log` lines generated by the provisioner itself use `"stream":"provisioner"`.

## Public API (auth-service via Kong)

`/api/v1/provisioning/...` mirrors the provisioner paths above (`/jobs`, `/jobs/{id}`, `/jobs/{id}/events`, `/ports/next`).
- Requires a valid JWT. Roles: platform admin only (the highest admin role in auth-service).
- auth-service adds `created_by` and forwards to `PROVISIONER_URL` with `PROVISIONER_TOKEN` (env).
- The SSE route is streamed unbuffered (auth-service flushes per event; Kong/nginx turn buffering off and use long read timeouts on this path).
- The frontend uses `fetch` streaming with the `Authorization` header, falls back to polling `GET /jobs/{id}` every 2s, and resumes the stream with `?since=` after reconnecting.
- Role: `ADMIN` (the highest role auth-service seeds; there is no separate super/platform admin role). Every tenant also has an `ADMIN`, so tenants are kept out by configuration: only the control-plane auth-service sets `PROVISIONER_URL`, and only the control-plane Kong routes `/api/v1/provisioning`. Without `PROVISIONER_URL` the routes answer 503.
- Bodies: provisioner success bodies are relayed as-is (Job, `{"jobs":[...]}`, `{"fe_port","api_port"}`), not wrapped in `{success,data}`. Errors use the `{success:false,error:{code,message}}` envelope.
- auth-service status codes: 401/403 from the JWT/role check; 503 `PROVISIONER_UNAVAILABLE` when `PROVISIONER_URL` is unset or the provisioner cannot be reached; 502 `PROVISIONER_AUTH_FAILED` when the provisioner rejects `PROVISIONER_TOKEN` (never relayed as 401, which would sign the user out); 502 `NEXT_PORTS_FAILED` from `/ports/next` when the generator fails; others relayed from the provisioner.

## Provisioner process (host)

- Binary: `backend/provisioner-service` (`go build -o provisioner ./cmd/provisioner`), no third-party dependencies.
- Env: `PROVISIONER_ADDR` (default `127.0.0.1:8790`; wildcard and public addresses are refused at startup), `PROVISIONER_TOKEN` (required, ≥ 32 chars, same value as auth-service's `PROVISIONER_TOKEN`), `PROVISIONER_REPO_DIR` (required, must contain `scripts/onboard-tenant.sh`), `PROVISIONER_DATA_DIR` (required; `jobs/<id>/job.json` and `jobs/<id>/events.jsonl`, mode 0700/0600), `PROVISIONER_BASH` (default `bash`).
- The script gets an allow-listed environment, not the provisioner's own: `PATH`, `HOME`, `USER`, `LOGNAME`, `SHELL`, `LANG`, `LANGUAGE`, `LC_*`, `TZ`, `TERM`, `TMPDIR`, `XDG_RUNTIME_DIR`, `SSL_CERT_FILE`, `SSL_CERT_DIR`, `HTTP(S)_PROXY`/`NO_PROXY` (either case), `DOCKER_*`, `COMPOSE_*`, `BUILDKIT_*`, and the variables the scripts read: `BASE_DOMAIN`, `TENANT_STORAGE_ROOT`, `IMAGE_TAG`, `DRY_RUN_DELAY`, `AUDITSPHERE_EMAIL_ENV`, `EMAIL_SENDING_DOMAIN`, `EMAIL_FROM_LOCALPART`, `RESEND_REGION`, `RESEND_CURL_CONNECT_TIMEOUT`, `RESEND_CURL_MAX_TIME` (plus the Windows system variables on Windows). `PROVISIONER_*` (the token above all), `ADMIN_PASSWORD` and anything else are dropped. The Resend master key comes from `/etc/auditsphere/email.env`, not the environment. A new variable a script reads from the environment must be added to `scriptEnvNames` in `runner.go`.
- On SIGTERM it cancels the running script (the job becomes `failed`, "provisioner stopped while the job was running"). The script runs in its own process group (`Setpgid`); cancellation sends SIGTERM to the whole group and SIGKILL 10s later, so a foreground `docker compose build` is stopped too (bash alone would wait for it). `KillMode=control-group` in the unit is still a good backstop.
- `result.admin_email` is `null` when `@@RESULT` had `null`; `result.admin_initial_password` is always `null` when `@@RESULT` had `null` (the admin already existed), so the UI should say "existing admin, password unchanged" in that case.

## Out of scope for this iteration

- `--org-preset` (not designed yet, see `docs/org-structure-model-plan.md`).
- Enforcing MFA for the first admin.
- Moving the tenant registry from TSV to Postgres.
