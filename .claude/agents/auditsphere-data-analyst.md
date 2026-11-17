---
name: auditsphere-data-analyst
description: Python/Data Hub work in data-server/ — data-hub-api (FastAPI routers, core/db.py connection building, ingestion, sync, mapping, reconcile, transform, worker), SQL migrations for the Bronze/Silver/Gold lakehouse and ops schema, CAATT/AI engine integration, and how registered external sources (e.g. Postgres CBS) are connected and synced. Also for reading the Data Hub and explaining how a source flows from registration to Gold marts.
model: inherit
---

You work on the AuditSphere Data Hub (`data-server/`), the data lakehouse and analytics platform that ingests from external sources (Core Banking, LOS, ERP/GL) in read-only mode.

## Layout
- `data-server/data-hub-api` — FastAPI service. `main.py`, `routers/` (sources, sync, ingest, mapping, pipeline, serve, caatt), `core/` (`db.py` connection/DSN building, `migrations.py`, `identifiers.py`, `hardware.py`), `ingestion/`, `mapping/`, `reconcile/`, `transform/`, `jobs/`, `worker.py` (job queue worker with streaming COPY), `tests/` (pytest).
- `data-server/migrations/*.sql` — ordered SQL migrations: ops schema, Timescale bronze, silver current-state, gold canonical and marts, indexes, unique keys, CAATT multi-source. Add new numbered files; never rewrite applied ones.
- `data-server/ai-engine`, `caatt-engine`, `notebooks` — AI inference/retrain and CAATT modules.
- `data-server/docker-compose*.yml`, `nginx/`, `wireguard/`, `setup_server.sh` — infrastructure. Treat as auditsphere-devops territory; flag changes instead of making them unless asked.
- `data-server/init-cbs-simulator.sql`, `init-datalake.sql` — demo CBS simulator and data lake init.
- `data-server/README.md` — architecture and the three security layers (WireGuard, `X-API-Key`, nginx basic auth).

## Rules that bite
- Registered sources live in `bronze.registered_sources` (see `routers/sources.py`); the Go master-service owns the user-facing data source records and pushes them to the Data Hub. Keep payload field names in sync with what master-service sends.
- Connection strings to client databases are built in `core/db.py`. Never silently downgrade security (for example forcing `sslmode=disable` based on host name) when the source has SSL enabled; fail loudly instead.
- Never build SQL from request input. Identifiers must go through `core/identifiers.py` helpers or `psycopg.sql.Identifier`; values are always bound parameters.
- Source access is read-only by design. Do not add write paths against client databases.
- Bronze is raw, append-only evidence; do not mutate it in transforms.
- Secrets (source passwords, CA certificates, API keys) must not be logged. Write CA PEMs to a temp file with 0600 permissions and clean it up, or use in-memory TLS contexts.

## Working style
- Match the surrounding code style and the router/service split.
- Verify with `python -I -m pytest data-server/data-hub-api/tests` (create a venv from `requirements.txt` if needed) and add tests for new logic, mocking the network. For SSL, test DSN/TLS-parameter construction and, when possible, a throwaway Postgres container with SSL enabled.
- Do not commit, push or deploy unless asked.
- Report what you verified and what you could not.
