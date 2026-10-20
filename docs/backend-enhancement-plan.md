# AuditSphere Backend Enhancement Plan

Date: 2026-10-15. Status: study and plan only. No code, config, or server was changed.
Method: read-only review of `backend/**`, compose files, Kong config, deploy scripts and CI workflows.
Citations are `path:line` in this repo. Secret values were not recorded; where a secret was compared, only a hash match was checked.
Effort: S = under 1 day, M = 1 to 3 days, L = 1 to 2 weeks. Items marked "unverified" need a check on the live VPS, which this study did not touch.

## 1. Summary

The most serious problem is that **any holder of the JWT signing secret can mint a token with any user and any role, and the audit, master and risk APIs accept it.** Combined with open self-registration, a stranger on the internet can obtain a valid token and then use the audit, master and risk APIs with no role check. Blockers 2 and 3 below are exploitable from outside the network without any credential.

Blockers before wider production use:
1. The same 64-character JWT secret is committed in four service `config.yaml` files and in `backend/deploy-prod.sh:144`. Rotate it and remove the literal.
2. Self-registration is open and assigns no role, and audit, master and risk authorise any valid token (no role check).
3. `/uploads/*` and `/api/v1/media/download/:id` are served without authentication.
4. The CI deploy path (`.github/workflows/deploy.yml:98,268`) deploys `docker-compose.prod.yml` directly, and that file runs `auth seed`, `audit seed` and `master seed` on every start. This conflicts with the no-seeders-on-prod rule.
5. Published ports bind `0.0.0.0`, including the Kong admin API (`docker-compose.prod.yml:234,244`).

## 2. Verified current state

### 2.1 Identity and tokens
| Finding | Where |
|---|---|
| Same 64-char literal secret in `auth`, `audit`, `master`, `risk` `config.yaml` and in the analytics override in `deploy-prod.sh:144`. Hashes match. | `backend/*/pkg/config/config.yaml` (secret key), `backend/deploy-prod.sh:144` |
| Services fail fast on an empty secret (good), but the config files supply a non-empty default, so the fail-fast never fires in a checkout. | `backend/risk-service/middleware/auth.go:33-37`, `backend/analytics-service/middleware/auth.go:33-37`, `backend/master-service/cmd/serve.go:59` |
| Tokens are HS256 with `exp` required and algorithm pinned. Good. | `backend/auth-service/pkg/utils/jwt.go:38-43`, `backend/audit-service/pkg/middleware/auth.go:66-71` |
| Access token lifetime defaults to 24 h. No refresh endpoint is mounted, although a refresh-token repository exists. | `backend/auth-service/pkg/config/config.go:149`, `backend/auth-service/routes/registry.go:59-69`, `backend/auth-service/repositories/refresh_token_repository.go` |
| Logout deletes a Redis key, but no service checks that key (auth's own middleware does not read Redis). A logged-out token stays valid for 24 h. | `backend/auth-service/services/auth/auth_service.go:354-356`, `backend/auth-service/pkg/middleware/auth.go` |
| audit-service accepts the token from `?token=` and from `auth_token`, `auth-token` and `token` cookies. Query tokens end up in logs, see below. | `backend/audit-service/pkg/middleware/auth.go:42-54` |
| Request logger writes the raw query string, so `?token=` is logged. | `backend/audit-service/pkg/middleware/logger.go:16,25` |
| Password minimum length is 6. | `backend/auth-service/models/dto.go:12,22,54` |
| No failed-login counter, lockout, or per-account throttle on login or MFA. Only log lines. | `backend/auth-service/services/auth/auth_service.go:66-100` |
| MFA TOTP secret is stored as a plain `varchar(255)`. `json:"-"` only hides it from API output. | `backend/auth-service/models/user.go:63` |
| `/api/v1/auth/register` is public and calls `Register`, which creates an active user with no role assignment. No feature flag or invite gate. | `backend/auth-service/routes/registry.go:61`, `backend/auth-service/services/auth/auth_service.go:304-347` |
| The seeder creates six accounts with a fixed, committed default password, and `create-if-missing` runs on every start. | `backend/auth-service/pkg/database/seeders/seeder.go:223-304` (password literal at :224; do not copy it) |
| The Resend webhook is intentionally unauthenticated at middleware level and relies on controller signature checks. Confirm the controller check. | `backend/auth-service/routes/registry.go:184-188` |

### 2.2 Authorisation (RBAC)
| Finding | Where |
|---|---|
| audit, master, risk and analytics authenticate every `/api/v1` route and then do not check roles. Any valid token has full CRUD, including DELETE. | `backend/audit-service/routes/handler.go:76-82`, `backend/master-service/routes/routes.go:40`, `backend/risk-service/routes/routes.go:60`, `backend/analytics-service/routes/routes.go:18` |
| `RequireRoles` is used only inside auth-service. | `backend/auth-service/routes/registry.go:84,149,151,159` |
| Casbin enforcer and `CasbinMiddleware` exist in three services and are never mounted. The `rbac_matrix` table is seeded but only read by the seeder, so it is documentation, not enforcement. | `backend/*/pkg/permissions/casbin.go`, `backend/audit-service/pkg/middleware/casbin.go:11`, `backend/auth-service/models/rbac_matrix.go` |
| Generic CRUD (`crud.Update`, `crud.Create`, `crud.Delete`) is mounted directly on routes with no ownership check. Update accepts any JSON key that maps to a schema column (mass assignment). | `backend/audit-service/routes/handler.go:86-120`, `backend/audit-service/controllers/crud/crud_handler.go:345-420` |
| Debug route `/debug/routes` lists all routes and is registered outside auth. | `backend/audit-service/routes/handler.go:68-80` |
| Swagger UI is registered unconditionally in the audit service. | `backend/audit-service/cmd/serve.go:258` |
| Rate limiter (tollbooth) is defined in three services and never mounted. The implementation calls `c.Next()` after `Abort` (harmless but wrong). | `backend/*/pkg/middleware/security.go:13-25` |

### 2.3 Files and uploads
| Finding | Where |
|---|---|
| `/uploads/*` is served by a public handler. Only `action-taken-reports` is excluded. | `backend/audit-service/cmd/serve.go:240`, `backend/audit-service/routes/uploads.go:11-31`, Kong `backend/kong-gateway/kong/prod/kong.yml:442` |
| `/api/v1/media/download/:id` is registered on the engine, not the authenticated group. | `backend/audit-service/routes/handler.go:498-502`, Kong `kong.yml:441` |
| Several `uploaded-*` controllers take `fileName` from a form field (not the multipart header) and join it into the disk path. A value containing `../` escapes the upload directory. This is a write-anywhere-in-container primitive for any authenticated user. Likely; confirm with a test in a sandbox. | `backend/audit-service/controllers/uploaded_annual_plan_controller.go:47,84` (same pattern in ~10 controllers) |
| Client-supplied `fileType` decides the served `Content-Type`, and the `View` endpoint serves `inline`. No `nosniff`. Stored HTML or SVG would execute on the API origin. | `backend/audit-service/controllers/uploaded_annual_plan_controller.go:216`, `backend/audit-service/controllers/uploaded_performance_report_controller.go:171-175` |
| Upload size is checked only after `io.ReadAll` of the whole body. No `MaxBytesReader`. | `backend/audit-service/controllers/uploaded_annual_plan_controller.go:58-75` |
| Media upload uses `header.Filename` (Go strips the path), so it is safer. The local provider does not validate type or size itself. | `backend/audit-service/controllers/media/controller.go:27-35`, `backend/audit-service/pkg/media/local.go:20-31` |
| ATR evidence is the good pattern: extension allowlist, name sanitising, `MaxBytesReader`, `nosniff`. Reuse it. | `backend/audit-service/controllers/action_taken_report/evidence.go:22-77,137,268` |
| prod compose mounts a named volume `audit_uploads` at `/root/uploads`. The brief's "no volume" claim is not true for the repo prod file. Whether the VPS uses it is unverified. | `backend/docker-compose.prod.yml:156-158,269` |

### 2.4 Network exposure and infrastructure
| Finding | Where |
|---|---|
| Published on all interfaces: auth `8001`, analytics `8084`, audit `8002`, master `8003`, risk `8004`, python-ai `8000`, Kong `8080`, `8443`, `8444`, admin `8009`. | `backend/docker-compose.prod.yml:94,132,150,181,202,214,242-245` |
| Kong admin listens on `0.0.0.0:8001` and is published as `8009`. Admin API has no auth in this config. | `backend/docker-compose.prod.yml:234,244` |
| Monitoring stack publishes Prometheus `9090`, Grafana `9300`, node-exporter `9100`, Loki `3100` without a loopback bind. | `monitoring/docker-compose.monitoring.yml:37,62,87,106` |
| Dev compose binds `127.0.0.1` correctly. Use it as the pattern. | `backend/docker-compose.dev.yml:16,28,44,59,79` |
| Postgres, Redis, Kafka and Zookeeper have no published ports in prod (good). Redis has no `requirepass`. Postgres `sslmode` is `disable` in all four service configs. | `backend/docker-compose.prod.yml:5-70`, `backend/*/pkg/config/config.yaml:12` |
| Kafka uses PLAINTEXT on both listeners, with no SASL or TLS anywhere in Go code. | `backend/docker-compose.prod.yml:61-66`, grep of `backend/**` (no SASL or TLS config) |
| CORS: Kong prod allows only `https://auditsphere.app` with credentials (good). analytics allows `*` with credentials. python-ai allows `*`. The custom `kong-gateway` config is also `*` with credentials. | `backend/kong-gateway/kong/prod/kong.yml:31-58`, `backend/analytics-service/cmd/main.go:29-35`, `backend/python-ai/main.py:21`, `backend/kong-gateway/config/config.yaml:92-114` |
| Kong prod has no `jwt` plugin. Auth is per-service only, and analytics, python-ai and parts of the others rely on app code. Analytics and python-ai have no role check. | `backend/kong-gateway/kong/prod/kong.yml` (no jwt plugin) |
| Kong proxy-cache keeps `Authorization` in the key (good). The app-level `ResponseCache` is mounted globally before the authenticated group, so a cached 200 can be served without re-validating the token until its TTL expires. Verify the cache's position relative to `Authenticate`. | `backend/audit-service/cmd/serve.go:236`, `backend/audit-service/pkg/middleware/cache.go:30` |
| python-ai auth is an optional `X-API-Key` middleware that fails open when `API_KEY` is empty. The prod compose python-ai service does not set `API_KEY`. `/retrain/auto` is reachable through Kong. | `backend/python-ai/main.py:49-62,1274`, `backend/docker-compose.prod.yml:213-224`, Kong `kong.yml:625-630` |
| The custom `kong-gateway` Go binary has an auth middleware that checks HMAC type but does not pin HS256, and its config has a committed placeholder secret. Confirm whether it is deployed at all (compose uses `kong:3.x`). | `backend/kong-gateway/pkg/middleware/auth.go:68-73`, `backend/kong-gateway/config/config.yaml:83` |
| Dockerfiles: six of seven have no `USER` directive (run as root) and use `FROM alpine:latest` (unpinned). python-ai uses `python:3.10-slim`. | `backend/*/Dockerfile` |
| `/swagger` and `/debug/routes` are reachable from Kong if the paths are routed. Kong routes list does not include them, so exposure is internal only (verify). | `backend/audit-service/cmd/serve.go:258` |

### 2.5 Deploy, seeding, migrations
| Finding | Where |
|---|---|
| CI deploy uses `docker-compose.prod.yml` directly with `up -d --build --force-recreate --remove-orphans`. No no-seed override is applied in the workflow. | `.github/workflows/deploy.yml:98-99,268` |
| prod compose runs seeders on every start: `auth seed`, `audit seed`, `master seed`. `risk up` runs migrations. | `backend/docker-compose.prod.yml:106,159,182,203` |
| The manual path `deploy-prod.sh` uses a no-seed override kept outside the repo, takes a `pg_dumpall` before deploy, and snapshots audit uploads. This is the safe path. | `backend/deploy-prod.sh:109-141,247-287,425-430` |
| Auth RBAC matrix is re-applied on every boot: existing rows are `Updates`-d from code, overwriting admin edits. | `backend/auth-service/pkg/database/seeders/seeder.go:416-421` |
| Migrate verbs differ by service (`auth migrate up`, `audit migrate up`, `master migrate`, `risk up`). The no-seed override runs `audit migrate up && audit serve`; the repo prod compose does not run migrate for audit. | `backend/deploy-prod.sh:109-141`, `backend/audit-service/cmd/migrate.go` |
| CI runs no tests, `go vet`, `govulncheck`, or secret scanning before it deploys. | `.github/workflows/deploy.yml` (pre-compiles binaries only) |

### 2.6 Kafka and outbox
| Finding | Where |
|---|---|
| Outbox relay (risk-service): rows are sent inside one DB transaction; a row is deleted after a successful send. If the commit fails after sends succeed, rows are re-sent: at-least-once. No attempt cap or dead-letter. `WaitForAll` acks (good). | `backend/risk-service/pkg/kafka/relay.go:99-137,155` |
| Master-service uses the same pattern (`datahub_sync`, `kafka/producer.go`). Confirm its relay has the same shape. | `backend/master-service/pkg/kafka/producer.go:59-99` |
| Audit consumer: handler errors retry in place with backoff and the offset is marked only on success (good). Auto-commit is enabled, so the committed offset may trail. Consumers must be idempotent. `annual_plan_sync` has `event_id` but no dedupe table was found. | `backend/audit-service/pkg/kafka/consumer.go:83-84,221-278`, `backend/audit-service/services/annual_plan_sync/handler.go` |
| Security events (login, logout, MFA) are fire-and-forget Kafka messages with no confirmed durable sink in the repo. | `backend/auth-service/services/auth/auth_service.go:363-369` (about 38 publish calls in auth) |

### 2.7 Backups
| Finding | Where |
|---|---|
| Pre-deploy `pg_dumpall` plus gzip only. No scheduled dump in the repo. No offsite copy in the repo. The server crontab was not checked (unverified). | `backend/deploy-prod.sh:425-430`, `backend/scripts/sync-clone-db.sh:28` |

### 2.8 Tenant isolation
| Finding | Where |
|---|---|
| Per-tenant JWT secret is generated with `openssl rand`, reused on re-run, and passed as `TENANT_JWT_SECRET`. Good. | `scripts/generate-tenant-stack.sh:186-195,212,246` |
| The bai tenant compose has a literal default database password, and a bind mount to a developer's absolute path on a Mac. Both break isolation and portability. | `backend/tenants/bai/docker-compose.yml:29,97` |
| Tenant ports are loopback-bound (good). Registry has only two rows. | `backend/tenants/bai/docker-compose.yml:192-220`, `backend/tenants/registry.tsv` |

## 3. Prioritised enhancements

### 3a. Must do before wider production use (P0)

| ID | Enhancement | Why / impact | Where | Effort | Dependencies | Data migration / downtime |
|---|---|---|---|---|---|---|
| A1 | Rotate the JWT secret. Remove the literal from the four `config.yaml` files and `deploy-prod.sh:144`. Load it only from env or a Docker secret, with no default. Keep `WithValidMethods`. | Anyone with repo or deploy-script read access can forge admin tokens for every service. | `backend/*/pkg/config/config.yaml`, `backend/deploy-prod.sh:144`, `docker-compose.prod.yml:96` | S (code) + M (rollout) | A2 and A3 reduce the blast radius; do A1 first. | No data change. All sessions end at rotation: schedule it. Rotate all five services in one window. Treat the old value as compromised. |
| A2 | Close public self-registration (remove the route, or make it admin-only or invite-only). Users with no role are denied by default. | Stranger obtains a token, then uses full CRUD in audit, master, risk. | `backend/auth-service/routes/registry.go:61`, `services/auth/auth_service.go:304-347` | S | A3 for the deny-by-default part. | Count users with no `user_roles` rows before enforcing (query, not run). No downtime. |
| A3 | Enforce authorisation on every mutating route in audit, master, risk, analytics. Start with a role map per route group (deny by default). Use the Casbin enforcer already in the repo, or `RequireRoles` with a route table. Ship in log-only shadow mode for one release first. | Authenticated users can read and delete everything. Casbin and RequireRoles exist but are not wired. | `backend/*/routes/*.go`, `backend/*/pkg/permissions/casbin.go` | M-L | Role matrix decision (see section 5, decision 2). | No schema change. Shadow mode prevents breaking users. Enforcement may need a short staged rollout. |
| A4 | Stop serving files without auth. Replace `/uploads/*` and `/api/v1/media/download/:id` with authenticated download handlers, and remove the public static route. Keep the legacy path for one release with a redirect or 410 for private dirs. | Charters, working papers, reports and evidence are readable by URL, and URLs with names are guessable. | `backend/audit-service/cmd/serve.go:240`, `routes/uploads.go`, `routes/handler.go:498-502`, Kong `kong.yml:441-442` | M | Frontend must send the bearer header on downloads (it may use `?token=` today; unverified). | Stored `FilePath` values start with `/uploads/`; keep them readable through the new handler. Requires frontend change. Downtime: none if done with a dual route. |
| A5 | Bind every published port to `127.0.0.1` (or remove the publish). Kong admin: no publish, listen on the loopback only. Monitoring ports: loopback only. | Admin API and internal services reachable from the internet. | `docker-compose.prod.yml:94,132,150,181,202,214,242-245`, `docker-compose.prod.yml:234`, `monitoring/docker-compose.monitoring.yml` | S | nginx upstreams must use `127.0.0.1:<port>` (platform). | None in data. Brief container recreate per service. Coordinate with `auditsphere-platform` for nginx. |
| A6 | Remove seeders from prod start. Split the compose: runtime services run `serve` only. One-off `migrate` jobs run explicitly. Seed tooling becomes a demo-only compose. Remove the auth RBAC `Updates` on boot (seeder.go:416-421). Fix the CI path so it cannot deploy a seeding command. | Constraint: no seeders on prod. Today CI runs `seed` on every deploy, and the auth seeder creates default accounts with a known password. | `docker-compose.prod.yml:106,159,182,203`, `.github/workflows/deploy.yml:98-99,268`, `seeders/seeder.go:223-304,416-421` | M | Decide the deploy path (CI or `deploy-prod.sh`). Rotate or disable the seeded accounts on prod (ops action, unverified whether they exist). | No data change by itself. Disabling the default accounts needs an admin action. Downtime: none. |
| A7 | Harden the uploaded-document controllers: take `fileName` only from `filepath.Base`, allowlist extensions, write to a server-generated name (UUID), validate content by magic bytes (`http.DetectContentType`), cap the body with `MaxBytesReader` before reading, set `nosniff`, serve `attachment` for any type that is not PDF or image. Reuse the ATR evidence pattern. | Path traversal write and stored-XSS on the API origin. | `backend/audit-service/controllers/uploaded_*.go` (about 10 files) | M | None. | New uploads only; no backfill. Existing rows keep `FilePath`. Verify existing rows for `..` before rollout (query, not run). No downtime. |
| A8 | Backups: scheduled `pg_dumpall` (or per-DB `pg_dump -Fc`) on a timer, plus uploads snapshot, encrypted, copied offsite, with retention and a restore drill. | Only a pre-deploy dump exists. No offsite copy. | `backend/deploy-prod.sh:425-430`, new cron/timer on the VPS | M | Platform (storage target, cron), devops (script). | No downtime. Restore test is required before this counts as done. |
| A9 | Remove credentials from URLs and logs. Accept the token only in `Authorization`. Remove `?token=` and cookie fallback in audit auth (or move to an HttpOnly cookie plus CSRF in the BFF). Drop `RawQuery` from the logger, or redact `token`, `access_token`, `password`. | Tokens land in access logs and Loki. | `backend/audit-service/pkg/middleware/auth.go:42-54`, `backend/audit-service/pkg/middleware/logger.go:16,25` | S | Frontend downloads must not depend on `?token=` (check). | None. Old logs keep the tokens; rotate A1 covers that. |
| A10 | Make logout and revocation real. Short access tokens (15 min) plus refresh-token rotation with a stored hash, or a `jti` denylist in Redis checked by every service. | Logged-out or stolen tokens stay valid for 24 h. | `backend/auth-service/services/auth/auth_service.go:354-356`, all `middleware/auth.go` | M-L | Denylist option needs Redis in every service (already there). | Refresh tokens table exists; additive. No downtime if rolled out behind a flag. |
| A11 | python-ai: fail closed when `API_KEY` is empty; set it in prod; restrict `/retrain/auto` to internal callers (no Kong route). Tighten analytics and python-ai CORS to frontend origins. | Model retrain callable from the internet. | `backend/python-ai/main.py:49-62,1274`, Kong `kong.yml:625-630`, `analytics-service/cmd/main.go:29-35` | S | Set `API_KEY` in env on both sides. | None. Brief restart. |
| A12 | CI gate before deploy: `go vet`, `go test`, `govulncheck`, `gitleaks` (or `trufflehog`), image build. Non-root `USER`, pinned base image digests. | No checks at all run on the deploy path. Root containers. | `.github/workflows/deploy.yml`, `backend/*/Dockerfile` | S-M | None. | None. |

### 3b. Should do within a quarter (P1)

| ID | Enhancement | Why / impact | Where | Effort | Dependencies | Data / downtime |
|---|---|---|---|---|---|---|
| B1 | Login abuse controls: per-account failed-attempt counter with backoff, per-IP limit on `/auth/login` and `/verify-mfa-login` (mount the existing tollbooth limiter or Redis token bucket), Kong route-level limit for auth. Fix `RateLimiter` `Next` bug. | No brute-force protection on login or MFA. | `auth_service.go:66-100`, `pkg/middleware/security.go:13-25`, `kong.yml` | S-M | Redis (present). | Additive (Redis keys). None. |
| B2 | Password policy and MFA: minimum 12 characters, breached-password check, Argon2id or bcrypt cost review, mandatory MFA for ADMIN and AUDIT_MANAGER. | Six-character passwords. | `auth-service/models/dto.go:12,22,54` | S-M | Decision on MFA scope. | Existing hashes stay valid; rehash on login. None. |
| B3 | Field-level encryption for secrets and sensitive PII (see section 4). Start with the MFA secret. | Plain-text TOTP seeds in the database. | `auth-service/models/user.go:63` | M | KEK source decision (section 4). | Dual-write, batched backfill, then switch reads. No downtime. |
| B4 | Kafka security and reliability: SASL/SCRAM and TLS on brokers, per-service ACLs, idempotent producer, dedupe table in each consumer keyed by `event_id`. | Any container on the network can read or inject events. Consumers are not idempotent. | `docker-compose.prod.yml:61-66`, `pkg/kafka/*`, `audit-service/pkg/kafka/consumer.go` | L | Broker reconfiguration. | Broker restart window, plus a consumer-table migration (additive). Schedule the window. |
| B5 | Outbox hygiene: claim rows with a status and lease, send outside the long transaction, cap attempts, move poison rows to a dead-letter table, export lag and error metrics. | Long-held row locks; unbounded retries; no visibility. | `backend/risk-service/pkg/kafka/relay.go:99-137`, master equivalent | M | None. | Additive columns (`status`, `locked_until`). None. |
| B6 | Durable business and security audit trail: one `audit_events` table per service (actor, action, entity, before and after hash, IP, user agent, request id). Keep Kafka as a copy, not the source. | Security events are fire-and-forget; no who-changed-what for audit evidence. | `auth_service.go:363-369`, all services | M | Retention decision (section 5). | Additive table. None. Partition by month later. |
| B7 | Replace generic CRUD with explicit DTOs per model: field allowlist for update, sort and filter (the whitelist for user-chosen columns is a per-model constant), ownership checks where records belong to a department or engagement. | Mass assignment; sort or filter injection risk if a column comes from the request. | `audit-service/controllers/crud/crud_handler.go:345-420`, `routes/handler.go:86-120` | M-L | A3 role map. | None. Behaviour change for clients that rely on arbitrary keys (check the frontend). |
| B8 | Cache ordering: mount `ResponseCache` after `Authenticate` on authenticated groups, or key the cache by a hash of the token and keep TTL short. Confirm the Kong proxy-cache TTL (300 s) is acceptable for fieldwork. | A cached 200 can be served to an expired token within TTL. | `audit-service/cmd/serve.go:236`, `pkg/middleware/cache.go`, `kong.yml:464-475` | S | None. | None. |
| B9 | Transport and data-store auth: Redis `requirepass` from env; Postgres `sslmode=require` (or `verify-full` with the internal CA) for the four services; Postgres password rotation per tenant. | Internal-only today, but a foothold on the Docker network reads everything. | `backend/*/pkg/config/config.yaml:12,14-17`, `docker-compose.prod.yml:19-30` | M | Certificates for Postgres. | Restart. Redis password change needs all clients updated at once. |
| B10 | Move to asymmetric JWT (EdDSA or RS256). Only auth-service holds the private key; other services verify with a public key and a `kid`. Rotation becomes publish-new-key, then retire. | Removes the shared HMAC secret that today lets every service forge tokens. Makes A1 a routine operation. | `backend/*/pkg/utils/jwt.go`, `*/middleware/auth.go` | L | A1 done first; Kong or nginx have no JWT plugin, so the change is app-only. | Dual-verify window (accept HS256 and new key for one release). Short downtime-free rollout. |
| B11 | Observability: request id propagated through Kong to services; JSON logs with actor and tenant; `/metrics` on each Go service on an internal port; alerts for 401 or 403 spikes, outbox lag, consumer lag, failed backups. | No metrics in Go services; logs are not correlated. | all `cmd/serve.go`, `monitoring/prometheus*` | M | Monitoring stack (platform). | None. |
| B12 | Migration policy: migrations run as one-off jobs (one replica, advisory lock), are additive in one release and destructive only in a later one, and are never run from `serve`. Document the verb per service. | `audit serve` path runs `migrate up`; AutoMigrate ships with the deploy. | `docker-compose.prod.yml`, `audit-service/cmd/migrate.go` | M | None. | None. |
| B13 | Tenant hygiene: remove the literal DB password and the developer host-path bind from `tenants/bai/docker-compose.yml:29,97`; add an automated check that a generated tenant compose has no literal secrets and only loopback ports. | Tenant stacks copy the same mistakes. | `backend/tenants/bai/docker-compose.yml`, `scripts/generate-tenant-stack.sh` | S | Devops. | None for the bai tenant until recreated. |
| B14 | Uploads storage: object storage (S3-compatible or ImageKit or GDrive already present) with server-side encryption and short-lived signed URLs. Until then, a dedicated named volume with host disk encryption and a backup job (A8). | Local disk is the only copy; no encryption; no lifecycle. | `audit-service/pkg/media/*.go`, `docker-compose.prod.yml:269` | L | Storage contract decision (section 5). | Migrate existing `FilePath`s with a copy job and a path map. Downtime: none with a dual-read window. |
| B15 | Remove or lock down the unused custom `kong-gateway` binary (confirm first), and its committed placeholder secret and `*` CORS. | Dead code with an unsafe config is a trap. | `backend/kong-gateway/**` | S | Confirm it is unused. | None. |
| B16 | Disable Swagger and `/debug/routes` in prod (build tag or env flag). | Route and schema disclosure. | `audit-service/cmd/serve.go:258`, `routes/handler.go:68-80` | S | None. | None. |

### 3c. Nice to have (P2)

| ID | Enhancement | Why | Effort |
|---|---|---|---|
| C1 | SBOM per image (syft) and signed images (cosign) before deploy. | Supply-chain evidence. | S |
| C2 | Malware scanning on upload (ClamAV sidecar or a scan job before the file becomes downloadable). | Defence in depth for evidence files. | M |
| C3 | Data retention jobs: scheduled purge or anonymisation for employee PII and expired evidence, tied to section 5 answers. | Retention obligations. | M |
| C4 | mTLS between services inside the Docker network (or a service mesh if the fleet grows). | Stops a compromised container from calling peers unauthenticated. | L |
| C5 | Penetration test and a second review of this plan after A1 to A12. | Independent check. | M (external) |
| C6 | Tenant-level Kong consumers and per-tenant rate limits. | Noisy-neighbour control on shared VPS. | M |
| C7 | Dashboards and runbooks per service (restore, rotate secret, re-run outbox). | Operability. | M |

## 4. Encryption recommendation (not implemented)

**Encrypt in transit.** nginx TLS for public traffic (platform). Kong to services and service to Postgres, Redis and Kafka: TLS on the Docker network (B4, B9). Until then, all three stay on the internal network with no published ports (already true for Postgres, Redis and Kafka in prod).

**Encrypt at rest, layered:**
1. Host disk encryption (LUKS or provider volume encryption) for `/var/lib/docker` and backups. Lowest effort; protects stolen disks and snapshots. Platform/ops, S-M.
2. Application-level envelope encryption for a short list of fields (below). Protects against SQL dump leaks and DB-only access.
3. Uploads: object storage with SSE now, or per-file envelope encryption later (B14).

**Which fields.**
| Tier | Fields | Reason |
|---|---|---|
| 1, encrypt now | `users` MFA `secret_key` (`auth-service/models/user.go:63`), any stored API keys or webhook secrets (verify where they live), refresh tokens (store SHA-256 hash only, not encrypted) | Direct credential material. |
| 2, decide with compliance | `employees` personal fields (name, email, phone, national ID if present), interviewee names and notes in fieldwork, finding narratives and management responses | Client personal data and audit evidence. Scope depends on section 5. |
| 3, do not encrypt | `username`, `email` used for login, IDs, status, dates, RBAC data | Needed for lookup, filtering, joins; protect by access control and logs instead. |

**Key management.**
- Envelope scheme: AES-256-GCM. Each row stores `ciphertext`, `nonce`, `key_version`. Use the table, column and row id as AAD so a ciphertext cannot be moved to another row.
- Keys: one data-encryption key per service per tenant, wrapped by a key-encryption key (KEK). KEK source phase 1: a Docker secret file outside the repo, one per environment and tenant. Phase 2: a KMS or Vault transit key. The KEK is never in `config.yaml` or compose env files in the repo.
- Rotation: new `key_version`, lazy re-encryption on write, background batch re-encrypt, retire old version only when a count query shows zero rows on it. Rotation of the KEK rewraps DEKs without touching row data.
- Losing the KEK loses the data. Back up the KEK separately from the database dump (A8). Tenant `.env` files already warn about this for `JWT_SECRET`; apply the same rule.

**Search implications.**
- Encrypted fields are not searchable or sortable. Do not put Tier 2 fields in list filters.
- Where exact lookup is needed (for example login by email, or employee lookup by email for `users.employee_id` fallback), add a keyed blind index: HMAC-SHA256 of the normalised value with a separate key, stored in its own column. Equality only.

**Migration path (no downtime).**
1. Additive: add `*_enc`, `key_version` columns (nullable). Deploy with dual-write for new rows.
2. Backfill in batches of about 500 with a resume cursor, outside business hours if large. Verify with a count of rows where `*_enc` is null.
3. Switch reads to the encrypted column with a fallback to plaintext for one release.
4. Drop plaintext in a later release, after a backup taken for that purpose and a restore test.

**Constraint check.** The backfill writes data. It is a production data write, so it must go through a reviewed script with dry-run and row counts, never a seeder, and must not overwrite existing values except the designated encrypted column.

## 5. Open questions the user must answer

1. **Compliance and data classification (decision 1).** Which obligations apply to client data: Indonesian PDP Law (UU PDP), ISO 27001 controls already documented in `docs/ISO_27001/`, or client contract terms? Must Tier 2 fields be encrypted at field level, or are disk encryption, access control and audit logging sufficient? Is there a data-residency requirement?
2. **Roles and self-registration (decision 2).** Who may create, edit and delete in audit, master and risk, by role? The seeded RBAC matrix is the only written source, and it is not enforced today. Is self-registration wanted for any path (demo, trial), or should it be removed? Who approves role changes?
3. **Recovery and retention (decision 3).** RPO and RTO for production (the plan assumes RPO 24 h and RTO 4 h as a placeholder). Offsite target and retention for database dumps and uploads. Retention for audit logs, working papers and deleted records, and whether hard delete is required for employee PII.
4. Is the GitHub repository public, or has it ever been public? This decides how urgent A1 is, and whether the old secret must be treated as leaked for all time.
5. Is production live with real client data today, and is the VPS running the repo compose file, the no-seed override, or both? (Unverified; this study did not touch the server.)
6. Which Kong and nginx paths are used by real clients, so A4 and A5 can be sequenced without breaking them?

## 6. Conflicts with the constraints

- **No data overwritten in production.** Violated today by `auth seed` and `audit seed` and `master seed` in `docker-compose.prod.yml:106,159,182`, and by CI running that file directly (`deploy.yml:98-99,268`). The auth RBAC `Updates` on every boot (`seeder.go:416-421`) overwrites admin edits. A6 fixes both.
- **No seeders on prod.** Same as above. The `deploy-prod.sh` path is the only one that sets `SEED_ON_BOOT=false` and uses a no-seed override, and that override lives outside the repo.
- **Migration risk.** A1 ends every session. A3 can lock users out if roles are wrong; use shadow mode. A4 changes URLs stored in the database; keep the legacy path for one release. A10 and B10 change the token contract and need a dual-accept window. A7 changes write behaviour for uploads; existing files keep their paths.
- **Uploads and force-recreate.** `deploy.yml:268` uses `--force-recreate`. A named volume survives that. A `down -v` or a compose file without the volume would not. Confirm on the VPS which compose file is in use (unverified).

## 7. Hand-off

Order matters. Steps 1 to 3 are the blockers.

**auditsphere-backend (and backend-2):**
1. A2: close or gate `/auth/register`; deny-by-default for users with no role.
2. A9: remove `?token=` and cookie fallback in audit auth; drop query logging or redact.
3. A3 (shadow mode first): role map and Casbin or `RequireRoles` on mutating routes in audit, master, risk, analytics.
4. A7: uploaded-document controllers (sanitize, allowlist, magic bytes, `MaxBytesReader`, `nosniff`, attachment).
5. A4: authenticated download handlers; keep legacy `/uploads` for one release.
6. A10: refresh-token rotation or `jti` denylist; B1 login throttling; B2 password policy.
7. A11: python-ai fail-closed `API_KEY`; analytics and python-ai CORS.
8. A6 (code side): split runtime and seed commands; remove boot-time RBAC overwrite.
9. B5, B6, B7, B16 as the quarter plan allows. B3 (MFA secret encryption) after the KEK decision.

**auditsphere-devops:**
1. A1 rotation runbook: new secret per environment, one window, all five services, old value retired. Remove literals from four `config.yaml` files and `deploy-prod.sh:144`.
2. A6 (deploy side): make CI deploy run only `serve` services, and a separate one-off `migrate` step. Move seed tooling to a demo-only compose. Confirm the VPS compose in use.
3. A5 (compose side): loopback binds for all published ports; remove the Kong admin publish.
4. A12: CI gate (`go vet`, `go test`, `govulncheck`, `gitleaks`); non-root `USER`; pin base images.
5. A8 (script side): scheduled encrypted dumps and uploads snapshot; restore script.
6. B9: Redis `requirepass`; Postgres TLS plan. B13: clean the bai tenant compose; add a generated-compose lint.
7. B4 (broker side): SASL and TLS plan, scheduled window.

**auditsphere-platform:**
1. A5 (nginx side): upstreams to loopback; confirm no public path reaches 8001, 8002, 8003, 8004, 8000, 8009.
2. Monitoring ports (9090, 9300, 9100, 3100) to loopback or private network; Grafana auth.
3. A8 (storage side): offsite target, cron or timer, retention, encryption key custody (separate from the dump).
4. Host disk encryption decision (section 4).
5. B11: monitoring alerts (401 or 403 spikes, outbox lag, backup failure).

**Smoke checks before calling the plan done:**
- Forged-token test: a token signed with the old secret is rejected by all five services after rotation (expect 401).
- Self-registration: `POST /api/v1/auth/register` returns 404 or 403; a user with no role gets 403 on `POST /api/v1/risks` and `DELETE /api/v1/audit-charters/:id`.
- Files: `GET /uploads/<any>` and `GET /api/v1/media/download/<id>` without a token return 401 or 404.
- Ports: from outside the VPS, only 80, 443 (and 22 if allowed) answer; `8009` and `8001` do not.
- Seeders: a prod deploy log has no `seed` command; the seeded default accounts cannot log in.
- Upload: a `fileName` containing `../` is rejected; an HTML file served by download is `attachment` with `nosniff`.
- Backup: a timer-produced dump exists offsite, and a restore into a scratch Postgres succeeds with matching row counts.
- Logs: a request with `?token=` does not appear in Loki.
- CI: a deploy with a failing `govulncheck` or test does not reach the VPS.
