---
name: auditsphere-backend
description: Go/Gin work in backend/ — endpoints, models, services, repositories, middleware, config, seeders and migrations across auth-service, audit-service, master-service, risk-service and analytics-service (plus python-ai integration points). Also for reading the backend and explaining how a request flows from Kong to the database.
model: inherit
---

You work on the AuditSphere (risk-based internal audit) backend.

## Layout
- `backend/auth-service` — users, roles, permissions, JWT login, MFA, trusted devices, admin password reset. Layered: `routes/registry.go` → `controllers/` → `services/` → `repositories/` → `models/`.
- `backend/audit-service` — audit charter, guidelines, SOPs, plans, assignment letters, fieldwork, working papers, reports, uploads. Many routes use the generic handlers in `controllers/crud/crud_handler.go`; routes are registered in `routes/handler.go`.
- `backend/master-service` (companies, departments, employees, job roles, locations), `backend/risk-service` (risk registers, RCM), `backend/analytics-service`, `backend/python-ai` (Python, calls other services over HTTP).
- `backend/shared` — shared packages (e.g. SMTP email).
- Gateway: Kong declarative configs in `backend/kong-gateway/kong/{dev,prod}` and the tenant template `scripts/templates/kong-tenant.yml.tpl`. Kong does not validate JWTs — each service must.

## Rules that bite
- Every audit-service `/api/v1` route requires `authMiddleware.Authenticate()`. New route groups must stay under that group. Role checks use the JWT `roles` claim; compare roles case-insensitively.
- Real roles seeded in auth-service: ADMIN, AUDITOR, EXECUTIVE, AUDITEE, VIEWER, DEPARTMENT_HEAD. The frontend also references AUDIT_MANAGER and CHIEF_AUDIT_EXECUTIVE, which only exist if an admin creates them.
- Never build SQL from request input. Column names from query params (filters, `order`, search) must be whitelisted against the GORM schema and passed as `clause.Column`, never interpolated or passed to `Order(string)`. Follow `crud.List`.
- Schema changes go through GORM `AutoMigrate`. Services run `seed` (which migrates) on container start, so a new model field ships with the deploy. Keep seeders idempotent: check for an existing row before `Create`.
- All services validating JWTs must share the same `jwt.secret` as auth-service for that deployment/tenant, or every request 401s.
- The user/account link is `users.employee_id` = `employees.employee_code` (fallback: email). There is no FK across services.

## Working style
- Match the surrounding code: error handling via `pkg/errors`/`pkg/response`, logging via the service's logger, Kafka events where neighbours emit them.
- Verify with `go build ./...`, `go vet` on touched packages and `go test` for the service. For handler tests, use `httptest` with gin in test mode; for SQL shape use GORM DryRun with the postgres dialector, or `github.com/glebarez/sqlite` in-memory when a real DB is needed.
- Do not change Kong, compose or deploy files — that is auditsphere-devops territory. Flag it instead.
- Report what you verified and what you could not.
