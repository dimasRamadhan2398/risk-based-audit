# Backend security hardening: change record (items A to G)

Status: changes are in the working tree on `staging`. Nothing is committed or deployed.
Plan references: `docs/backend-enhancement-plan.md` (A1, A4, A9, A10, A11, the debug and swagger item).
No secret values are recorded here. Where a value was compared, only a SHA-256 digest was used.

## Operator actions required before deploy

1. **Rotate the JWT secret.** The old value is committed in git history. Follow `docs/jwt-rotation-runbook.md`. Every service now refuses to start without a `JWT_SECRET` of at least 64 bytes, and refuses the old value.
2. **python-ai** needs `API_KEY` (same value as analytics `PYTHON_AI_API_KEY`) and `INTERNAL_RETRAIN_TOKEN`. Without `API_KEY`, every analytics AI call returns 401 after this change.
3. **Frontend** file links need an Authorization header (see section 6).
4. Set `ENABLE_SWAGGER` / `ENABLE_DEBUG_ROUTES` only in dev. Both default to off.

## A. JWT literal removal and startup policy

| File | Change |
|---|---|
| `auth-service/pkg/config/config.yaml`, `config.docker.yaml` | `jwt.secret` set to empty. Comment says to use the env var. |
| `audit-service/pkg/config/config.yaml` | same |
| `master-service/pkg/config/config.yaml`, `config.local.yaml` | same |
| `risk-service/pkg/config/config.yaml` | same |
| `backend/config.docker.yaml` | same |
| `{auth,audit,master,risk}-service/pkg/config/config.go` | `viper.BindEnv("jwt.secret", "JWT_SECRET")` added, so the env value wins even when the file has no key. |
| `{auth,audit,master,risk}-service/pkg/config/jwt_secret.go` (new) | `ValidateJWTSecret`: refuses empty, under 64 bytes, or a digest on the known-public list. The error never includes the value. `Config.ValidateJWT()` wraps it. |
| `{auth,audit,master,risk}-service/pkg/config/jwt_secret_test.go` (new) | Policy cases, env overrides file, env works when the file has no key, empty env fails, digest blocklist (uses a synthetic value added to the list in the test). |
| `analytics-service/middleware/jwt_secret.go`, `jwt_secret_test.go` (new); `middleware/auth.go` | Same policy. `NewAuthMiddleware` uses it, and `cmd/main.go` already fatals on error. |
| `risk-service/middleware/jwt_secret.go`, `jwt_secret_test.go` (new); `middleware/auth.go` | Same policy in `NewAuthMiddleware`. |
| `auth-service/cmd/auth/serve.go`, `audit-service/cmd/serve.go`, `risk-service/cmd/serve.go` | Startup gate calls `cfg.ValidateJWT()` before binding. |
| `master-service/cmd/serve.go` | Same gate, after `initConfig` and before the logger exists. It returns an error. |

Verified: a throwaway `go test` in a scratch copy, fed the committed old value from `git show HEAD` through an environment variable, returns "matches a publicly known value". The value was never printed. Startup with JWT_SECRET unset or set to a 5-byte value exits 1 with a clear message on auth, audit, risk, analytics, and master (exit 1). A 71-byte synthetic value passes the gate and fails later at the database.

Not changed by this work, but flagged: `backend/deploy-prod.sh` and `backend/docker-compose.yml` contained the literal at HEAD. The working tree now shows someone else's uncommitted edits that remove it. Confirm those edits before committing. `kong-gateway/config/config.yaml` has a `jwt_secret` placeholder that is not the old value.

## B. Public self-registration closed

| File | Change |
|---|---|
| `auth-service/routes/registry.go` | `POST /auth/register` removed. |
| `auth-service/routes/auth/auth.go` | Same route removed from the legacy, unused route file. |
| `auth-service/controllers/auth/auth_controller.go`, `services/auth/auth_service.go` | `Register` handler and service method removed. |
| `auth-service/models/dto.go` | `RegisterRequest` removed (no remaining users). |
| `auth-service/services/user/user_service.go` | `CreateUser` (admin path, `POST /api/v1/users`, ADMIN only) now requires at least one non-blank role, assigns the roles, and deletes the account again if none of the names exist. |

Tests: `auth-service/routes/auth_register_test.go` (route absent, POST returns 404), `auth-service/controllers/user/create_user_roles_test.go` (no roles rejected, roles assigned, unknown role removes the account).

Frontend: no registration page or caller of `/auth/register` was found. No frontend caller of `POST /api/v1/users` was found either.

## C. Files: authenticated, confined, safe types

| File | Change |
|---|---|
| `audit-service/pkg/safefile/safefile.go`, `safefile_test.go` (new) | `ResolveUnder` confines paths to the root, including symlinks. `ContentTypeFor` is an extension allowlist. HTML, SVG, XML and script types are never returned. `ServeFile` sends inline only for safe types and attachment otherwise, adds `nosniff`, a sandbox CSP and `Cache-Control: private`. Directories are refused. `IsPrivatePath` covers `action-taken-reports`. |
| `audit-service/routes/uploads.go` | `RegisterUploads(engine, root, authenticate)`. `/uploads/*` requires a valid token (was public). Uses safefile. No directory listing. |
| `audit-service/cmd/serve.go` | Passes `authMiddleware.Authenticate()` to `RegisterUploads`. |
| `audit-service/routes/handler.go` | `GET /api/v1/media/download/:id` moved into the `/api/v1/media` group (was outside the authenticated group). |
| `audit-service/controllers/media/controller.go` | `Download` rewritten. The `path` / `filepath` parameter is confined to the uploads root (was a traversal: `uploads/../../x` was read unchecked). Private directories are skipped in the fallback search. Uses safefile. |
| `audit-service/controllers/uploaded_performance_report_controller.go` | `getDocumentContentType` ignores the client-supplied `fileType` and uses the extension allowlist. Applies to all `uploaded_*` download and view handlers. |

Tests: `pkg/safefile/safefile_test.go`, `routes/uploads_test.go` (token required, query token refused, private dir 404, traversal stays inside root, no listing), `controllers/media/controller_test.go`, `controllers/content_type_test.go`.

## D. No tokens in query strings or cookies; redacted logs

| File | Change |
|---|---|
| `audit-service/pkg/middleware/auth.go` | `?token=` and `auth_token` / `auth-token` / `token` cookies removed. Header only. |
| `risk-service/middleware/auth.go`, `analytics-service/middleware/auth.go` | Cookie fallback removed (the same pattern). |
| `*/middleware/redact.go`, `redact_test.go` (new) in auth, audit, master, risk and analytics | `RedactQuery` replaces the values of parameters whose names contain token, secret, key, password, passwd, signature or auth. An unparseable query is dropped. `RedactingLogger` is gin's access log using it. |
| `auth-service/pkg/middleware/logger.go`, `audit-service/pkg/middleware/logger.go` | Log `RedactQuery(RawQuery)`. |
| `master-service/cmd/serve.go` | `gin.Logger()` replaced by `middleware.RedactingLogger()`. |
| `risk-service/cmd/serve.go`, `analytics-service/cmd/main.go` | `gin.Default()` replaced by `gin.New()` with `Recovery` and `RedactingLogger`. |

Tests: `*/middleware/auth_test.go` in audit, risk and analytics (header accepted, `?token=` and cookies refused, missing or foreign token refused). `redact_test.go` in all five.

Frontend: no `?token=` URLs and no `auth_token` cookie usage was found, so nothing relied on them.

## E. Logout ends the session

Findings: auth-service login stores `token:<jwt>` in Redis and logout deletes it, but the protected-route middleware never checked that key. A logged-out token therefore kept working until its JWT expiry (24h). There is no refresh endpoint. `RefreshTokenRepository` exists but nothing calls it, so there is no refresh path to harden.

| File | Change |
|---|---|
| `auth-service/pkg/middleware/auth.go` | `WithSessionCheck(SessionChecker)`. When set, `Authenticate` also requires a live session. An error counts as not active (fail closed). |
| `auth-service/cmd/auth/serve.go` | `loginSessionCheck(redisClient)` checks `token:<jwt>` with `EXISTS` and is wired in. |
| `auth-service/pkg/middleware/session_test.go` (new) | Logout makes the token return 401. Redis error returns 401. With no check configured, behaviour is unchanged. |

Residual risk: audit, master, risk and analytics verify JWTs locally and do not see the session. A logged-out token still works there until expiry. Fixing that needs a shared session check or a shorter expiry (follow-up).

## F. python-ai fails closed

| File | Change |
|---|---|
| `python-ai/security.py` (new) | `token_matches` (constant time, empty never matches). `install_api_key_middleware` requires `X-API-Key` on every path except `/health`. With `API_KEY` unset, everything except `/health` returns 401. `internal_call_allowed` checks `X-Internal-Token` against `INTERNAL_RETRAIN_TOKEN`. |
| `python-ai/main.py` | Uses the helpers. Removed the old bypass for `/docs` and `/openapi.json`. `/retrain/auto` requires `X-Internal-Token`, and is refused while `INTERNAL_RETRAIN_TOKEN` is unset. |
| `python-ai/test_security.py` (new) | 10 tests with a FastAPI TestClient: missing, wrong and unset keys return 401, `/health` is public, `/retrain/auto` needs the internal token, unset internal token disables it. Result: 10 OK (Python 3.12, fastapi 0.115.6, httpx in a temporary venv). |

Not run: `main.py` itself, because importing it loads torch and the model stack. It was byte-compiled only.

## G. Debug routes and Swagger off by default

| File | Change |
|---|---|
| `audit-service/routes/debug.go` (new), `debug_test.go` | `GET /debug/routes` only when `ENABLE_DEBUG_ROUTES=true`. `SwaggerEnabled()` reads `ENABLE_SWAGGER` (strictly `true`). |
| `audit-service/routes/handler.go` | Calls `registerDebugRoutes`. The route was previously public and unconditional. |
| `audit-service/cmd/serve.go` | Mounts `/swagger/*any` only when `ENABLE_SWAGGER=true`. |
| `auth-service/cmd/auth/serve.go` | Same `ENABLE_SWAGGER` gate for auth Swagger. |

## Routes still public (after this change)

| Service | Route | Reason |
|---|---|---|
| auth | `POST /api/v1/auth/login` | Credential entry point. |
| auth | `POST /api/v1/auth/verify-mfa-login` | Second step of login, before a token exists. |
| auth | `POST /api/v1/resend/webhook` | Called by Resend. Authenticated by signature in the controller, not by JWT. |
| auth, audit, master, risk, analytics, python-ai | `GET /health` | Health probes (Kong active check). Returns a static body. |
| audit | `/swagger/*`, `/debug/routes` | Not mounted by default. They become public only when `ENABLE_SWAGGER` / `ENABLE_DEBUG_ROUTES` is `true`, so keep them off in prod. |
| python-ai | `/docs`, `/openapi.json` | No longer public. They now need `X-API-Key`. |

Every `/api/v1` route in audit, master, risk and analytics sits behind `Authenticate()` (verified in the route registration code). `/uploads/*` and `/api/v1/media/download/:id` are now behind it too.

## 6. Frontend follow-ups (not changed)

Authorization headers cannot be sent by `window.open`, `tab.location.href` or `<a href>`. The following call sites open `/uploads/...` or `/api/v1/media/download/...` this way and will get 401:

- `frontend/stores/audit-fieldwork.ts`: about lines 603-615, 651, 1476-1525 (`window.open` and `tab.location.href` to `viewUrl`, `url`, `blobUrl`). The `$fetch` path around line 1809 already sends the header and is fine.
- `frontend/stores/charter.ts:441`
- `frontend/components/audit-charter/SopList.vue:203-208`
- `frontend/components/audit-charter/GuidelineList.vue:199-204`
- `frontend/utils/documentViewer.ts:37, 125`
- `frontend/stores/consulting-service.ts:249`
- `frontend/pages/kpi-performance/index.vue:70-72` (check whether it is a file URL)

Fix for each: fetch with `Authorization: Bearer <token>` as a blob, then open the object URL. `audit-fieldwork.ts` already has this pattern. `frontend/nuxt.config.ts` proxies `/uploads/**` to Kong. It should forward the Authorization header, but this was not tested.

Frontend `<img src>` uses only data URLs (avatars, signatures), so it is unaffected.

## Devops follow-ups (not changed, per the rule on deploy files)

- python-ai: set `API_KEY` and `INTERNAL_RETRAIN_TOKEN` in the compose env. The compose default `PYTHON_AI_API_KEY=${...:-dev-ai-api-key}` is a literal fallback and should not be used in prod.
- python-ai publishes port 8000 to the host (`docker-compose.yml`, `docker-compose.prod.yml`). Remove the publish. Only analytics on the internal network needs it.
- Confirm the working-tree edits to `deploy-prod.sh` and `docker-compose.yml` that remove the JWT literal are committed. The no-seed override must pass `JWT_SECRET` from env, because every service now refuses to start without it.
- Kong Go gateway (`kong-gateway/pkg/middleware/logging.go:30`) also logs the raw query string. It is outside the listed services, so it was not changed.

## Verification

| Scope | Result |
|---|---|
| auth-service: `go build ./...`, `go vet ./...`, `go test -count=1 ./...` | PASS |
| risk-service: same | PASS |
| analytics-service: same | PASS |
| master-service: build, vet, targeted tests (`cmd`, `pkg/config`, `pkg/middleware`) | PASS. Full `go test` has 2 failures, `controllers/company` and `controllers/department`. Both reproduce on a clean HEAD export, so they predate this work. |
| audit-service | The real tree does not build: `services/audit_activity/service.go` (lines 63, 90, plus later edits) does not match the `TargetUnitID` type in `models/audit-activity.go`. Both files are someone else's uncommitted work. Verified instead on a HEAD export with only the files in this change overlaid. Build, vet and the touched-package tests pass there (`routes`, `controllers`, `controllers/media`, `pkg/safefile`, `pkg/middleware`, `pkg/config`). |
| Startup gates | Built binaries; checked unset, short and valid-length secrets on five services. |
| python-ai | `python -m unittest test_security`: 10 OK. |
| Images | Not rebuilt. |

Not verified: end-to-end login, logout and file download against running Postgres, Redis and Kong. Live `main.py` in python-ai. Frontend behaviour after the 401 change.
