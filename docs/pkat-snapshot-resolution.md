# PKAT display snapshots: resolve on save, backfill existing rows

Verified against the code on 2026-10-21 (branch `staging`, uncommitted working tree). Re-check the "Current state" section before relying on it.

Owner: project lead. Implementer: `auditsphere-backend` (section 5). No frontend change.

---

## 1. Problem

On the Annual Audit Plan (PKAT), each activity row stores the ids `auditUniverseId` and `departmentId`, and three display snapshots: `department`, `riskName`, `riskLevel`. The snapshots were blank for new rows (example: PKAT-2026-001, activity `8be6606f-...`).

## 2. Current state (root causes)

- The PKAT payload never sends the snapshots: `frontend/utils/annualPlanPayload.ts` (`buildAnnualActivityPayload`) sends the ids and says so on purpose, so a stale copy does not overwrite a refreshed value.
- `controllers/annual_plan/annual_plan_activities.go` (`PrepareOnUpdate`) replaces the whole `activities` column with the rows decoded from the body. A row without snapshot fields therefore stores blank snapshots on every save.
- `PrepareOnCreate` never fills them either.
- The only writer is the Kafka sync (`services/annual_plan_sync`). It updates plans that exist when an event arrives and matches by id. A row saved after the event, or saved by a PUT that cleared it, stays blank until the next event for that entity.
- audit-service has no local copy of departments, universe entities or yearly scores. The authoritative copies are in master-service (departments) and risk-service (universe). Both are behind JWT middleware, and audit-service has no service credential, so it cannot call them from a save or a backfill without forwarding a user token.

## 3. Decision

Keep a local lookup in audit-service (`annual_snapshot_refs`), fed by the events the sync consumer already receives. Resolve snapshots from it in the save hooks, and fill existing rows with a backfill command.

Options considered:

| Option | Verdict |
|---|---|
| (a) Local lookup fed by events, resolved on save, plus backfill | **Chosen.** No new cross-service call or credential. Save latency unchanged. A missing lookup leaves a blank field and does not fail the save. |
| (b) Call master-service and risk-service on save | Rejected. Risk-service and master-service require a JWT. The backfill has no user token, so it would need a service credential. Adds two network calls to every save. |
| (c) Frontend sends the names | Rejected. The frontend deliberately omits snapshots (see section 2). It would also make the client the source of truth for names. |

Trade-off: the lookup is only as complete as the events the consumer has seen. Existing data is seeded by re-publishing the current state with the existing resync endpoints (section 6), not by replaying Kafka. The consumer group has already committed its offsets, so a replay would need a new group id.

## 4. Data model

New table `annual_snapshot_refs` (model `models.AnnualSnapshotRef`, `models/annual-snapshot-ref.go`):

| Column | Type | Notes |
|---|---|---|
| kind | varchar(20), not null | `department`, `entity`, `year_score` |
| ref_id | varchar(64), not null | other service's id, lower-case |
| year | int, not null | 0 for department and entity; the scored year for year_score |
| name | varchar(255) | department or entity name |
| risk_level | varchar(50) | year_score only |
| updated_at | timestamp | |

Unique index `idx_annual_snapshot_refs_key` on (kind, ref_id, year) is the upsert target.

- Links are business-key only (`ref_id` is master-service or risk-service id). No foreign key exists or can exist.
- A row is written by the event handler and never deleted. A deleted source keeps its last name, as plan snapshots do.
- Missing row = "not known yet". It is never an error.

Migration risk: additive. `AutoMigrate` creates an empty table. No existing column changes. `audit_annuals.activities` is still JSON, so no SQL change to plans.

## 5. Behaviour and hand-off to `auditsphere-backend`

Done in this change (reference, do not re-implement):

1. `services/annual_plan_sync/sync.go`: `ApplyDepartment`, `ApplyEntity` and `ApplyYearScore` upsert the lookup (skipping empty names or levels) before they update plans. An upsert error is returned, so Kafka redelivers the event.
2. `services/annual_plan_sync/refs.go`: `FillSnapshots(tx, year, rows, prior)`. Per field, in order: lookup value; value the row already carries; stored value of the same row if the same department or entity (and, for the level, the same year); else blank.
3. `controllers/annual_plan/annual_plan_activities.go`: `PrepareOnCreate` and `PrepareOnUpdate` call `fillSnapshots`. It runs in a savepoint and only logs on error, so a failed lookup never fails the save.
4. `services/annual_plan_sync/backfill.go`: `BackfillSnapshots(ctx, db)` fills only empty fields of live plans. Optimistic write per plan, same as the sync. Idempotent.
5. `cmd/migrate.go`: `audit migrate backfill-annual-snapshots [--dry-run]`. Runs in a transaction and rolls back on `--dry-run`, like `backfill-atr`. `annual_snapshot_refs` is added to `AutoMigrate` and to the down list.

## 6. Ops order (for `auditsphere-devops`)

1. Deploy audit-service. `audit migrate up` (or the deploy's own migrate step) creates `annual_snapshot_refs`. Until the lookup is seeded, saves still work and fill nothing.
2. Confirm the sync consumer is running (`kafka_consumer.enabled: true`, group `audit-service-annual-plan-sync`), and check its lag is 0 before step 3.
3. Re-publish current state, with a valid admin JWT:
   - `POST /api/v1/departments/events/resync` on master-service (publishes `department.updated` for every live department);
   - `POST /api/v1/audit-universe/events/resync` on risk-service (publishes entity updates and every yearly score).
   Both are idempotent for the consumer (convergent writes).
4. Wait for the consumer to catch up (step 2 check).
5. `audit migrate backfill-annual-snapshots --dry-run`, read the counts, then run it without the flag.

Do not run step 3 or the backfill against production without a dry run first. Both re-publish or write across all tenants of the environment.

## 7. Smoke checks

- `GET /annual-audit-plans/:id` for PKAT-2026-001: row `8be6606f-...` has `department`, `riskName` and `riskLevel` non-empty (or the expected blank if the universe has no 2026 score).
- Save the plan from the PKAT form without changing anything. The snapshots are unchanged.
- Change a row's department to another known one. The department name changes, and `riskName` is unchanged.
- Run the backfill twice. The second run reports `plans_changed=0 fields_filled=0`.

## 8. Not covered

- Wrong (non-empty) snapshots are not repaired by the backfill. They converge through events and the next save of that row.
- Rows that exist only as legacy rows with no id and never saved are resolved through the deterministic ids (`form.EnsureRowIDs`) on the next save, not by the backfill.
- A consumer that is disabled in an environment leaves the lookup empty there.
