# Audit Activity Plan rework: match the "Formulir Rencana Kegiatan Audit"

Verified against the code on 2026-10-15 (branch `staging`, HEAD `4fde0c9` plus the uncommitted working tree). Re-check the "Current state" section before relying on it.

Owner: project lead. Implementers: `auditsphere-backend` (section 9) and `auditsphere-frontend` (section 10), working in parallel against the contract in section 6.

---

## 1. Summary

The product owner sent the paper form **FORMULIR RENCANA KEGIATAN AUDIT** (RKA). The Audit Activity Plan page (`/audit-activity-plan`) has to match it. Today the app has three overlapping places for this data, and the form's columns currently sit in the wrong one:

| Store today | What it holds | Problem |
|---|---|---|
| `audit_annuals.activities` (PKAT, JSON) | Since commit `a24ff32` (2026-10-14), **all 12 RKA columns** plus the RKA header (`title`, `activityType`) | The PKAT form became the RKA form. `docs/overview.MD:131-143` describes two different forms. |
| `activity_plans` (JSON `plannedActivities`, `resourceAuditors`, `budget`, `review`) | A generic plan: auditee, priority, budget, auditor list, approver | None of these fields are on the paper form. Category, risk, team size and mandays are copied from PKAT. |
| `audit_activities` (typed table, FK to `audit_annuals`) | Has the right columns (`justification`, `audit_purpose`, `audit_focus`, `team_size`, `total_mandays`, `team_leader_id`, `audit_universe_id`, `project_code`) | **No UI writes to it.** Only `cmd/seed.go:1698` does. But KPI work-plan realisation, audit completion and execution sync all read from it. |

### Key decisions (defaults; section 12 lists the few the product owner must confirm)

1. **Each form has one owner.**
   - The PKAT (Annual Audit Plan) owns *what is audited and with how many resources*: object, category, Audit Universe, department, schedule, team size, mandays and supervisor.
   - The RKA (Audit Activity Plan) owns *why and how*: the Jenis Kegiatan Audit header, the Judul, Kode Penugasan, Jenis Audit, Justifikasi, Tujuan, Fokus and Ketua Tim.
   - The RKA shows the PKAT-owned columns **read-only**, resolved live from the PKAT. They are never edited in two places.
2. **An RKA document is one header plus N rows.** The header is in `activity_plans` (new columns `annual_plan_id`, `activity_type`, `form_version`). Each row is an **`audit_activities`** row (new columns `activity_plan_id`, `annual_activity_id`, `sort_order`, `audit_kind`, `team_leader_name`).
   - Saving an RKA therefore feeds the KPI realisation, audit-completion and execution-status features that already read `audit_activities` but currently only ever see seed data.
3. **One PKAT row maps to at most one live RKA row.** One live RKA is allowed per (PKAT, Jenis Kegiatan Audit), which matches one paper form per activity type per year. Both rules are enforced by partial unique indexes.
4. **The header Jenis Kegiatan Audit filters the PKAT rows** by their category: Assurance↔Assurance, Audit Khusus↔Special Audit, Audit Tujuan Tertentu↔Specific Reason, Audit Investigasi↔Investigation.
   - The row column **Jenis Audit** (for example "Operasional") is a *different* vocabulary and is new. Nothing in the app stores it today.
5. **Kode Penugasan moves to the RKA row** (`audit_activities.project_code`). The user may type the client's own code (for example `PBDK-1`). If the field is left blank, the server generates `{TYPE}-{YEAR}-{NNN}` with the existing `pkg/activitycode`. The code is locked once fieldwork has started.
6. **The PKAT form loses the RKA-only fields.** These are removed from the PKAT form: Jenis Kegiatan Audit, Kode Penugasan, Justifikasi, Tujuan and Fokus.
   - The values are **not deleted**. They stay in the JSON, the server preserves them on PKAT saves, and they prefill the RKA row the first time the PKAT row is added.
7. **Legacy activity plans stay read-only.** Rows created with the old form (`form_version = 1`) can be viewed and deleted but not edited. Nothing is auto-converted, because the old fields don't map onto the paper form.
8. **The migration is additive only.** It adds columns and partial indexes and runs no seeders. It runs at `serve` start-up, because tenant stacks start with `serve` only (`cmd/serve.go:146-152`), and also in `migrate up`.

---

## 2. Current state (evidence)

### 2.1 Activity Plan (`/audit-activity-plan`)
- **Route and sidebar:** `frontend/pages/audit-activity-plan/index.vue`, sidebar `frontend/layouts/default.vue:200`. `upload.vue` handles document upload only (`/uploaded-plan-documents`). It doesn't parse anything, and it is unchanged by this plan.
- **Form:** `frontend/components/audit-activity-plan/ActivityPlanForm.vue` is a 5-step modal (`:584-590`):
  1. Basic info: planTitle, planYear (derived), period start/end, department (derived), createdBy, creationDate.
  2. Planned activities: auditName, auditee, link to an annual activity, read-only category/risk/riskLevel/duration/numberOfAuditors, priority P1-P3, estimatedSchedule (date), budgetEstimation.
  3. Resources and budget: auditor list (name, position, competence, availability), allocated budget, notes.
  4. Review and approval: creator and approver names and positions, approval date, notes.
  5. Attachments.
- **Store:** `frontend/stores/activity-plan.ts`. It links rows to PKAT rows with `linkAnnualActivity` (`:430-449`) and sends `annualPlanId`/`annualActivityId` per row (`:396-409`).
- **Backend:** `backend/audit-service/routes/handler.go:167-176` uses the generic `crud.*` handlers on `models.ActivityPlan` (`models/audit-activity.go:93-114`), with `PlannedActivity` stored as JSON (`:50-68`).
  - **Bug:** `PlannedActivity` has no `annualPlanId`/`annualActivityId` fields. `crud.CreateWithHooks` binds the body into the struct (`controllers/crud/crud_handler.go:308`), so the PKAT link is **dropped on create**. It only survives an update, because the update path is map-based (`:378`).
- **Activity code:** `controllers/activity_plan/activity_code.go` and `pkg/activitycode/activitycode.go` produce server codes such as `ASR-2026-003`, unique across `audit_activities.project_code` and `activity_plans.planned_activities`.

### 2.2 Annual Audit Plan / PKAT (`/annual-audit`)
- **Form:** `frontend/components/annual-audit/AnnualAuditForm.vue`.
  - Header (`:36-101`): year, **Jenis Kegiatan Audit** (`activityType`), title (labelled "Judul Kegiatan Audit"), status, PKAT code.
  - Rows (`:143-361`): Kode Penugasan, Objek, Jenis Audit (AuditCategory), department, Audit Universe, risk snapshot, **Justifikasi, Tujuan, Fokus**, Jadwal (text), Jumlah Tim, Total Mandays, Ketua Tim (`supervisorName`).
  - Plan-level month checklist (`selectedMonths`), notes and attachments.
- **Model:** `backend/audit-service/models/audit-annual-plan.go:17-59`. The doc comment explicitly maps the RKA columns onto `AnnualAuditActivity`. `Title` and `ActivityType` are at `:87-91`. The closed list `AnnualPlanActivityTypes` is at `models/annual_plan_status.go:76`.
- **Hooks:** `controllers/annual_plan/annual_plan_activities.go` validates `activityType` and gives each row a stable UUID `id`. `PrepareOnUpdate` decodes rows into the struct (`:63-75`), so any key the struct doesn't know, or any key the client omits, is wiped on save.
- **Kafka sync:** `services/annual_plan_sync` refreshes the `riskName`, `riskLevel` and `department` snapshots from risk-service and master-service events.
- **Delete:** the frontend decides between delete and deactivate on `plan.isUsed` (`frontend/stores/annual-audit.ts:734-769`, uncommitted). **No backend code sets `isUsed`.** A grep of `backend/` for `isUsed`/`is_used` finds nothing in audit-service, so every PKAT is hard-deleted (soft delete) even when it's in use.

### 2.3 `audit_activities`, the typed table that is unused today
- **Model:** `models/audit-activity.go:120-151`, with endpoints `/audit-activities` (`routes/handler.go:144-153`). No frontend caller exists. The only `audit-activities` hit in `frontend/` is a comment in `stores/activity-plan.ts:373`.
- **Readers:**
  - `services/audit_execution_sync/sync.go:230-266` links executions by `id`, then `title`, then `project_code`.
  - `services/audit_completion/analyzer.go:153-187` (completion rates).
  - `controllers/performance_stats_controller.go:124-133` (work-plan realisation).

---

## 3. Ownership model and invariants

```
PKAT (audit_annuals)                          RKA (activity_plans header)
  year ───────────────────────────────────────► Tahun (derived, read-only)
  activities[] (rows, stable UUID id)            activity_type  (Jenis Kegiatan Audit)
    id ◄───────── annual_activity_id ─────┐      plan_title     (Judul Kegiatan Audit)
    name, category, auditUniverseId,      │
    department, timelineText,             │    audit_activities (RKA rows)
    auditorCount, totalMandays,           └──── annual_activity_id, activity_plan_id, sort_order
    supervisorName                              project_code (Kode Penugasan), audit_kind (Jenis Audit)
                                                justification / audit_purpose / audit_focus
                                                team_leader_id + team_leader_name (Ketua Tim)
                                                snapshot cols: title, audit_type, audit_universe_id,
                                                  target_unit_id, team_size, total_mandays
```

Invariants, which the backend enforces:
- **I1.** A live `audit_activities` row has at most one live row per `annual_activity_id`.
- **I2.** A live `activity_plans` row (form_version 2) has at most one live row per (`annual_plan_id`, `activity_type`).
- **I3.** Every row of an RKA belongs to the RKA's PKAT (`audit_activities.annual_plan_id = activity_plans.annual_plan_id`), and its PKAT category maps to the RKA's `activity_type` (section 4.3).
- **I4.** The snapshot columns on `audit_activities` (`title`, `audit_type`, `audit_universe_id`, `target_unit_id`, `team_size`, `total_mandays`) are written **only by the server** from the PKAT row. They exist only because downstream SQL needs them. The API always returns the live PKAT value (section 7).

---

## 4. Field mapping

### 4.1 Header (one per form)

| Form field | Stored in | Source of truth | UI behaviour in the RKA form |
|---|---|---|---|
| Tahun | `audit_annuals.year`, through `activity_plans.annual_plan_id`. Also denormalised into `activity_plans.plan_year` for listing and filtering. | PKAT | The user picks a **PKAT** ("2026 · PKAT-2026-001 · <title> · <status>"). Tahun is shown read-only. A warning appears if the PKAT status is not `Approved`. The PKAT can't be changed after create. |
| Jenis Kegiatan Audit | `activity_plans.activity_type` (new). Closed list: `Assurance`, `Audit Khusus`, `Audit Tujuan Tertentu`, `Audit Investigasi`. | RKA | Required single select. It filters which PKAT rows can be added. It can be changed only if every existing row still matches the new type. |
| Judul Kegiatan Audit | `activity_plans.plan_title` (existing column) | RKA | Required text, at most 255 characters. |

### 4.2 Rows (one per assignment)

| # | Form column | Stored in | Source of truth | UI behaviour |
|---|---|---|---|---|
| 1 | No | `audit_activities.sort_order` | RKA (row order) | Auto-numbered 1..N. Rows can be reordered with up/down buttons. |
| 2 | Kode Penugasan | `audit_activities.project_code` (existing, unique) | **RKA** | Optional text, at most 50 characters, pattern `^[A-Za-z0-9][A-Za-z0-9._/-]*$`. Leave blank to have the server generate one. On first add it is prefilled from the legacy PKAT `activityCode` if that is set. It becomes read-only once the row is "in use" (section 7.4). |
| 3 | Objek Penugasan Audit | PKAT `activities[].name`; snapshot in `audit_activities.title` | PKAT | Read-only, labelled "Dari PKAT". |
| 4 | Jenis Audit | `audit_activities.audit_kind` (new) | **RKA** | Required. Select with suggestions (Operasional, Keuangan, Kepatuhan, Teknologi Informasi, Kinerja); custom values are allowed, at most 100 characters. |
| 5 | Audit Universe | PKAT `activities[].auditUniverseId` plus the `riskName`/`riskLevel` snapshots; snapshot in `audit_activities.audit_universe_id` | PKAT | Read-only: name plus a risk-level badge. risk-service has no universe code, so "02B - ..." is shown only if the name contains it (Q3). |
| 6 | Justifikasi Pemilihan Objek | `audit_activities.justification` (existing `text`, items joined by `\n`) | **RKA** | Numbered-list editor. At least 1 item, at most 20, each at most 1000 characters. On first add it is prefilled from the legacy PKAT `justification`. |
| 7 | Tujuan Audit | `audit_activities.audit_purpose` (existing) | **RKA** | Same as #6. Prefilled from PKAT `auditObjective`. |
| 8 | Fokus Audit | `audit_activities.audit_focus` (existing) | **RKA** | Same as #6. Prefilled from PKAT `auditFocus`. |
| 9 | Jadwal Penugasan | PKAT `activities[].timelineText` | PKAT | Read-only text, for example "Jul-Oct". |
| 10 | Jumlah Tim (Orang) | PKAT `activities[].auditorCount`; snapshot in `audit_activities.team_size` | PKAT | Read-only. |
| 11 | Total Mandays | PKAT `activities[].totalMandays`; snapshot in `audit_activities.total_mandays` | PKAT | Read-only. Header totals sum the rows. |
| 12 | Ketua Tim | `audit_activities.team_leader_id` (existing, master-service employee UUID) plus `team_leader_name` (new snapshot) | **RKA** | Optional employee picker (`useEmployeeApi().getAllEmployees()`), blank by default, which matches the paper example. The PKAT supervisor is shown beside it as a hint (Q1). |
| - | Department / auditee unit | PKAT `departmentId`; snapshot in `audit_activities.target_unit_id` | PKAT | Read-only, shown under the object in the row card. It is not a form column but is needed downstream. |

### 4.3 Header type ↔ PKAT category mapping

Compare the normalised codes from `activitycode.TypeCode` so that legacy values such as `"ASSURANCE"` also match.

| `activity_type` (RKA header) | PKAT row `category` accepted | Code prefix |
|---|---|---|
| Assurance | Assurance | ASR |
| Audit Khusus | Special Audit | SPC |
| Audit Tujuan Tertentu | Specific Reason | SPR |
| Audit Investigasi | Investigation | INV |

PKAT rows with Consulting Services, Quality Assurance Review or Follow-Up Audit can't be added to any RKA. Those categories have their own modules.

### 4.4 Fields removed

| Where | Field | Action |
|---|---|---|
| RKA form | period start/end, department, creation date, auditee, priority, estimated schedule (date), budget estimation, resource auditors, budget, review/approval | Removed from the UI. The columns and JSON stay in the DB. Old plans show them read-only in a "Data format lama" section. |
| RKA form | createdBy | Kept; set from the signed-in user by the server-side create (see 6.4). |
| RKA form | attachments | Kept as optional step 3. Useful for the signed scan of the paper form. |
| PKAT form | Jenis Kegiatan Audit (`activityType`) | Removed from the UI. The column is kept and the frontend stops sending it. |
| PKAT form | Kode Penugasan (`activities[].activityCode`), Justifikasi, Tujuan, Fokus | Removed from the UI. The **server preserves stored values** on PKAT save (section 9, B5), and they are used once as RKA prefill. |
| PKAT form | `title` | Kept but relabelled "Judul PKAT". It is the PKAT document's own name, not the RKA Judul. |
| PKAT form | `supervisorName` | Kept, still required, relabelled "Supervisor (Pengendali Teknis)" (Q1). |

---

## 5. Data model and migration

### 5.1 Model changes (`backend/audit-service/models/audit-activity.go`)

`ActivityPlan` (table `activity_plans`):
```go
AnnualPlanID  *uuid.UUID `gorm:"type:uuid;index" json:"annualPlanId"`
ActivityType  string     `gorm:"type:varchar(100)" json:"activityType"`   // Jenis Kegiatan Audit
FormVersion   int        `gorm:"type:int;not null;default:1" json:"formVersion"` // 1 = legacy, 2 = RKA form
// PlanTitle (existing) = Judul Kegiatan Audit; PlanYear (existing) = denormalised PKAT year.
```

`AuditActivity` (table `audit_activities`):
```go
ActivityPlanID   *uuid.UUID `gorm:"type:uuid;index" json:"activity_plan_id"`
AnnualActivityID *string    `gorm:"type:varchar(36);index" json:"annual_activity_id"` // AnnualAuditActivity.ID
SortOrder        int        `gorm:"type:int;not null;default:0" json:"sort_order"`
AuditKind        string     `gorm:"type:varchar(100)" json:"audit_kind"`              // Jenis Audit
TeamLeaderName   string     `gorm:"type:varchar(255)" json:"team_leader_name"`
// TeamLeaderID: change to *uuid.UUID (nullable; blank Ketua Tim is valid).
// TargetUnitID: change to *uuid.UUID and drop NOT NULL (PKAT rows may lack departmentId).
```

Move `AnnualPlanActivityTypes` to an RKA-named constant (`ActivityPlanActivityTypes`, same four values) and keep the old name as an alias until the PKAT validator stops using it.

### 5.2 Migration function `migrateActivityPlanForm(db)`

This is new, in `cmd/migrate_activity_plan_form.go`. Call it from **both** `cmd/serve.go` (after the existing AutoMigrate block, like `migrateATR`) **and** `runMigrations` in `cmd/migrate.go`. It must be idempotent and must never delete or rewrite user data.

1. `db.AutoMigrate(&models.ActivityPlan{}, &models.AuditActivity{})` adds the new columns. Existing `activity_plans` rows get `form_version = 1` from the default.
2. Run explicit SQL, wrapping each statement and logging errors without exiting, in the same style as `serve.go:118`:
   ```sql
   ALTER TABLE audit_activities ALTER COLUMN target_unit_id DROP NOT NULL;
   ALTER TABLE audit_activities ALTER COLUMN team_leader_id DROP NOT NULL;
   UPDATE audit_activities SET team_leader_id = NULL WHERE team_leader_id = '00000000-0000-0000-0000-000000000000';
   CREATE UNIQUE INDEX IF NOT EXISTS ux_audit_activities_annual_activity_live
     ON audit_activities (annual_activity_id)
     WHERE deleted_at IS NULL AND annual_activity_id IS NOT NULL AND annual_activity_id <> '';
   CREATE UNIQUE INDEX IF NOT EXISTS ux_activity_plans_annual_type_live
     ON activity_plans (annual_plan_id, activity_type)
     WHERE deleted_at IS NULL AND annual_plan_id IS NOT NULL AND form_version = 2;
   ```
3. **Best-effort link of legacy `audit_activities`.** For live rows with `annual_activity_id IS NULL`, look for exactly one row in the same `annual_plan_id`'s `activities` JSON whose `name` equals `title` (case-insensitive, trimmed). If there is exactly one match and no other live row already holds that id, set `annual_activity_id`. Log the counts (linked, ambiguous, unmatched). Don't touch `activity_plan_id`.
   - This makes sure that adding that PKAT row to an RKA later **adopts** the existing row (with its status and execution link) instead of creating a duplicate.
4. **No backfill of RKA documents.** Legacy `activity_plans` stay `form_version = 1`. Legacy PKAT `justification`/`auditObjective`/`auditFocus`/`activityCode` stay in the JSON and are used as prefill (section 6.3).
5. **No seeder changes are required.** `seedActivityPlans` and `seedAuditActivities` are insert-if-missing and prod runs migrate-only.
   - An optional form-version-2 demo seed may be added, but only as insert-if-missing, and never in the prod compose path.

**Before the prod deploy,** devops takes a `pg_dump` of each tenant's `rb_audit_audit_<slug>` and runs these read-only counts, recording the results in the deploy note:
```sql
SELECT count(*) FROM activity_plans WHERE deleted_at IS NULL;
SELECT count(*) FROM audit_activities WHERE deleted_at IS NULL;
SELECT count(*) FROM audit_annuals a, jsonb_array_elements(a.activities::jsonb) r
 WHERE a.deleted_at IS NULL AND (coalesce(r->>'justification','') <> '' OR coalesce(r->>'auditObjective','') <> '' OR coalesce(r->>'auditFocus','') <> '');
```
**Rollback:** the new columns and indexes are additive. Rolling back the binary leaves them unused. Drop the two indexes only if the old binary has to create rows that would violate them; the old binary never writes `annual_activity_id`, so it won't.

---

## 6. API contract (audit-service, behind Kong `/api/v1/activity-plans`, which needs no Kong change)

Envelope as everywhere in audit-service (`pkg/response/response.go`): `{ "success": bool, "message": string, "data": ..., "error": { "code", "message", "details" } }`.

- For 400 and 409 errors, `error.details` is a **JSON-encoded string** that the frontend parses. The `row` field is 1-based:
  `{"field":"rows[2].annualActivityId","row":3,"activityPlanId":"…","activityPlanTitle":"…","assignmentCode":"…"}`
- All IDs are UUID strings. Field names are camelCase. Dates are RFC 3339.
- These routes replace the generic `crud.*` handlers at `routes/handler.go:167-176`.

### 6.1 `GET /activity-plans`

Query parameters:

| Parameter | Default | Notes |
|---|---|---|
| `page` | 1 | |
| `page_size` | 20 | Maximum 100. |
| `search` | | ILIKE on title. |
| `year` | | int |
| `activity_type` | | |
| `annual_plan_id` | | |
| `include_legacy` | `true` | |

Sort: `year desc, activity_type, created_at desc`.

```json
{ "success": true, "data": {
  "items": [ {
    "id": "uuid", "formVersion": 2,
    "year": 2026, "activityType": "Assurance", "title": "Rencana Kegiatan Audit Assurance 2026",
    "annualPlan": { "id": "uuid", "code": "PKAT-2026-001", "title": "PKAT 2026", "status": "Approved", "isActive": true },
    "totals": { "rowCount": 3, "teamSize": 13, "totalMandays": 120 },
    "status": "PLANNED", "attachmentCount": 1,
    "createdBy": "Siti A.", "createdAt": "…", "updatedAt": "…"
  } ],
  "pagination": { "page": 1, "page_size": 20, "total": 7, "total_pages": 1 }
} }
```

For legacy items (`formVersion: 1`), `activityType` is `""`, `annualPlan` is `null`, `year` is parsed from `planYear` (0 if none), `title` is `planTitle`, and `totals.rowCount` is `len(plannedActivities)` (the other totals are 0).

### 6.2 `GET /activity-plans/:id`
```json
{ "success": true, "data": {
  "id": "uuid", "formVersion": 2,
  "annualPlanId": "uuid",
  "annualPlan": { "id": "uuid", "code": "PKAT-2026-001", "title": "PKAT 2026", "year": 2026, "status": "Approved", "isActive": true },
  "year": 2026, "activityType": "Assurance", "title": "…", "status": "PLANNED",
  "rows": [ {
    "id": "uuid",                       // audit_activities.id
    "no": 1,
    "annualActivityId": "uuid",
    "assignmentCode": "PBDK-1", "assignmentCodeEditable": true,
    "auditObject": "Audit Kehandalan Kilang",          // PKAT
    "auditCategory": "Assurance",                      // PKAT row category
    "auditKind": "Operasional",
    "auditUniverse": { "id": "uuid", "name": "02B - Pemeliharaan, utilisasi, dan kehandalan proses kilang", "riskLevel": "High" },
    "department": { "id": "uuid", "name": "Refinery Operations" },
    "justification": ["BOD & BOC Concern", "RKAP 2025", "Top risk 2024-2025: menurunnya kinerja kilang"],
    "objectives": ["Mendapatkan bukti yang memadai …", "Memperoleh insight …"],
    "focus": ["Evaluasi pelaksanaan program kerja …", "Evaluasi terhadap bad factor …"],
    "schedule": "Jul-Oct",                              // PKAT timelineText
    "teamSize": 5, "totalMandays": 44,                  // PKAT
    "teamLeader": { "employeeId": "uuid", "name": "…" },  // or null
    "pkatSupervisorName": "…",                          // PKAT, display only
    "status": "PLANNED",                                // audit_activities.status (execution-driven)
    "inUse": false,                                     // section 7.4
    "source": "pkat"                                    // "snapshot" when the PKAT row cannot be resolved
  } ],
  "totals": { "rowCount": 1, "teamSize": 5, "totalMandays": 44 },
  "attachmentCategory": "Plan", "attachments": [ { "name": "RKA-2026.pdf", "size": "120 KB", "url": "/uploads/…" } ],
  "attachmentUploadedBy": "…", "attachmentUploadDate": "2026-10-15",
  "createdBy": "…", "createdAt": "…", "updatedAt": "2026-10-15T08:00:00Z",
  "legacy": null
} }
```
- Rows are ordered by `sort_order`, then `created_at`.
- With `source: "snapshot"` (the PKAT or PKAT row was deleted), the PKAT-owned fields come from the snapshot columns, and `schedule` and `pkatSupervisorName` are `""`.
- For a legacy plan, `rows` is `[]` and `legacy` is filled:
  `{ "planTitle", "planYear", "planPeriodStart", "planPeriodEnd", "department", "creationDate", "plannedActivities": [...], "resourceAuditors": [...], "budget": {...}, "review": {...} }`. These are the stored JSON as-is.

### 6.3 `GET /activity-plans/annual-activity-options`

Lists the PKAT rows that can be picked.

Query parameters:

| Parameter | Required | Notes |
|---|---|---|
| `annual_plan_id` | yes | |
| `activity_type` | yes | |
| `exclude_plan_id` | no | The RKA being edited. Its own rows count as available. |

```json
{ "success": true, "data": { "items": [ {
  "annualActivityId": "uuid", "itemNumber": "3",
  "auditObject": "Audit Kehandalan Kilang", "auditCategory": "Assurance",
  "auditUniverse": { "id": "uuid", "name": "…", "riskLevel": "High" },
  "department": { "id": "uuid", "name": "…" },
  "schedule": "Jul-Oct", "teamSize": 5, "totalMandays": 44, "pkatSupervisorName": "…",
  "available": true,
  "usedBy": null,                  // or { "activityPlanId", "activityPlanTitle", "assignmentCode" }
  "prefill": {                     // legacy PKAT values, used once on add; empty arrays/"" if none
    "assignmentCode": "", "justification": ["…"], "objectives": ["…"], "focus": ["…"]
  },
  "existingRow": null              // or { "id", "assignmentCode", "status" } when a legacy audit_activities row is linked (adopted on save)
} ] } }
```
- Only rows whose category maps to `activity_type` (section 4.3) are returned.
- The response is 404 `ANNUAL_PLAN_NOT_FOUND` if the PKAT is missing or soft-deleted, and 400 `VALIDATION_ERROR` if `activity_type` is not one of the four values.
- `prefill` lists are split with the same normaliser as 6.4.
- When a legacy `audit_activities` row is linked (`existingRow` set), `prefill` takes its `justification`, `audit_purpose` and `audit_focus` first.

### 6.4 `POST /activity-plans` and `PUT /activity-plans/:id`

Request (the same shape for both; `updatedAt` is optional on PUT):
```json
{
  "annualPlanId": "uuid",
  "activityType": "Assurance",
  "title": "Rencana Kegiatan Audit Assurance 2026",
  "updatedAt": "2026-10-15T08:00:00Z",
  "rows": [ {
    "id": "uuid",                    // present for existing rows; omit for new
    "annualActivityId": "uuid",
    "assignmentCode": "PBDK-1",      // "" = generate (new) / keep (existing, if locked)
    "auditKind": "Operasional",
    "justification": ["…"], "objectives": ["…"], "focus": ["…"],
    "teamLeaderEmployeeId": "uuid",  // or null
    "teamLeaderName": "…"            // display snapshot sent by the client from the picker; "" when null
  } ],
  "attachmentCategory": "Plan",
  "attachments": [ { "name": "…", "size": "…", "url": "…" } ]   // full list; new files already uploaded via POST /media/upload
}
```

**Server rules** (one transaction; all-or-nothing):
- **Header**
  - `title`: required, trimmed, at most 255.
  - `activityType`: one of the four values.
  - `annualPlanId`: must be a live PKAT. The `isActive=false` PKATs are allowed on PUT but rejected on POST with 409 `ANNUAL_PLAN_INACTIVE`.
  - On PUT, `annualPlanId` must equal the stored value (400 `VALIDATION_ERROR`, field `annualPlanId`).
- **Rows:** `rows` must contain at least 1 entry and at most 200.
- **Normalising list items:** trim each item, strip a leading `^\s*\d+[.)]\s*`, and drop empty items. After that, each of `justification`, `objectives` and `focus` needs at least 1 item, at most 20 items, and each item at most 1000 characters. Store the items joined by `\n`.
- **Per row**
  - `annualActivityId` must exist in that PKAT's `activities` (400 `ANNUAL_ACTIVITY_NOT_FOUND`).
  - Its category must map to `activityType` (409 `ANNUAL_ACTIVITY_TYPE_MISMATCH`).
  - It must be unique within the request (400 `VALIDATION_ERROR`).
  - It must not be held by another live RKA's row (409 `ANNUAL_ACTIVITY_TAKEN`). The race is caught by the unique index and returns the same code.
- **`auditKind`:** required, at most 100.
- **Row identity**
  - If `id` is given, it must belong to this RKA (400).
  - Without `id`: if a live `audit_activities` row already has this `annual_activity_id` and no `activity_plan_id`, **adopt** it. Otherwise, if a soft-deleted row for this RKA has the same `annual_activity_id`, **restore** it, which keeps its code. Otherwise insert a new row.
- **Kode Penugasan**
  - Validate the pattern and length.
  - If blank on a new row, generate it with `activitycode.Next(tx, pkatRow.category, pkat.year)`.
  - If blank on an existing row, keep the stored code.
  - If it differs from the stored code while the row is `inUse`, return 409 `ASSIGNMENT_CODE_LOCKED`.
  - If the code is taken by any other `audit_activities` row (the index includes soft-deleted rows), return 409 `ASSIGNMENT_CODE_TAKEN`.
- **Server-set columns:** `annual_plan_id`, `activity_plan_id`, `sort_order` (array index + 1), the snapshot columns from the PKAT row (I4), and `status` (`PLANNED` on insert; never changed here).
- **Rows missing from a PUT** are soft-deleted. If one of them is `inUse`, return 409 `ACTIVITY_ROW_IN_USE` and write nothing.
- **Header columns set by the server:** `plan_year` (the PKAT year as a string), `form_version = 2`. On POST, `created_by` and `creation_date` come from the JWT user's name and today's date.
- **`attachmentUploadedBy` and `attachmentUploadDate`** are set by the server when `attachments` grows.
- **PUT `activityType` change:** allowed only if every remaining row still matches; otherwise 409 `ANNUAL_ACTIVITY_TYPE_MISMATCH`.
- **Uniqueness I2:** 409 `ACTIVITY_PLAN_DUPLICATE`, with `details.activityPlanId` set to the existing RKA.
- **Optimistic lock:** if `updatedAt` is given and differs from the stored value at second precision, return 409 `ACTIVITY_PLAN_STALE`.
- **Legacy plans:** PUT on a `formVersion: 1` plan returns 409 `ACTIVITY_PLAN_LEGACY_READONLY`.
- **Old clients:** a body without `rows` (the old frontend shape) returns 400 `VALIDATION_ERROR` and writes nothing. There is deliberately no compatibility shim (see section 11).

Responses: 201 (POST) or 200 (PUT) with the full 6.2 object.

### 6.5 `DELETE /activity-plans/:id`
- Soft-deletes the header and all its `audit_activities` rows.
- Returns 409 `ACTIVITY_PLAN_IN_USE` if any row is `inUse`; `details.row` names the first such row.
- Legacy plans can always be deleted.
- On success it returns 200 `{ "success": true, "data": { "id": "uuid" } }`.
- Adopted legacy rows (those that existed before the RKA) are **detached** rather than deleted: `activity_plan_id = NULL`, `sort_order = 0`.

### 6.6 Annual plan endpoints (changed behaviour, same paths)
- **`GET /annual-audit-plans` and `GET /annual-audit-plans/:id`:**
  - Each plan gets `"isUsed": bool`. It is true if a live form-v2 `activity_plans` row or a live `audit_activities` row has `annual_plan_id = plan.id`.
  - Each `activities[]` item gets `"activityPlanRef": null | { "activityPlanId", "activityPlanTitle", "auditActivityId", "assignmentCode" }`.
  - Use `crud.ListWithItems` with a decorator; it costs 2 queries per page.
- **`PUT /annual-audit-plans/:id`:** see section 7 for the new 409s and the preservation rule.
- **`DELETE /annual-audit-plans/:id`:** returns 409 `ANNUAL_PLAN_IN_USE` when `isUsed`. The frontend then offers deactivation, which is `PUT {isActive:false}` and already exists.

### 6.7 Error codes

| Code | HTTP | Meaning |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Shape, required or length errors. `details.field` points at the field. |
| `ANNUAL_PLAN_NOT_FOUND` | 404 | |
| `ANNUAL_PLAN_INACTIVE` | 409 | POST against a deactivated PKAT. |
| `ANNUAL_ACTIVITY_NOT_FOUND` | 400 | The row id is not in this PKAT. |
| `ANNUAL_ACTIVITY_TYPE_MISMATCH` | 409 | The PKAT row category doesn't match the header type. |
| `ANNUAL_ACTIVITY_TAKEN` | 409 | The PKAT row is already in another RKA. `details.activityPlanTitle` names it. |
| `ASSIGNMENT_CODE_TAKEN` | 409 | |
| `ASSIGNMENT_CODE_LOCKED` | 409 | |
| `ACTIVITY_ROW_IN_USE` | 409 | |
| `ACTIVITY_PLAN_IN_USE` | 409 | |
| `ACTIVITY_PLAN_DUPLICATE` | 409 | |
| `ACTIVITY_PLAN_STALE` | 409 | |
| `ACTIVITY_PLAN_LEGACY_READONLY` | 409 | |
| `ANNUAL_PLAN_IN_USE` | 409 | DELETE on a used PKAT. |
| `ANNUAL_PLAN_ACTIVITY_IN_USE` | 409 | A PKAT PUT removes a row that an RKA uses. `details.row` and `details.activityPlanTitle` identify it. |
| `ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT` | 409 | A PKAT PUT changes a used row's category so that it no longer matches its RKA. |

---

## 7. Sync rules with the Annual Audit Plan (PKAT)

1. **Reads are live.** Every RKA GET resolves the PKAT-owned fields from `audit_annuals.activities` at read time. The Kafka-driven snapshot refresh (`services/annual_plan_sync`) therefore shows up in the RKA automatically, and nothing new is consumed.
2. **The PKAT save refreshes the snapshots.** A new `AfterHook` on `PUT /annual-audit-plans/:id` runs in the same transaction. For each row with a linked live `audit_activities` row, it updates `title`, `audit_type`, `audit_universe_id`, `target_unit_id`, `team_size` and `total_mandays`.
   - This keeps execution-sync title matching and the KPI counts correct.
   - Issued codes are never changed, even if the year or category changes.
3. **A PKAT row removed while used** returns 409 `ANNUAL_PLAN_ACTIVITY_IN_USE`. The user removes the row from the RKA first. This applies to every PKAT status, including revisions.
4. **"In use"** means a row whose `status <> 'PLANNED'`, or one that has an `audit_executions.activity_id` pointing at it. In-use rows can't be removed from the RKA and their Kode Penugasan is locked.
5. **A PKAT row category change on a used row** returns 409 `ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT` if the new category no longer maps to its RKA's `activity_type`.
6. **Deprecated PKAT fields are preserved.** In `PrepareOnUpdate`, for each incoming row whose `id` matches a stored row, copy the stored `activityCode`, `justification`, `auditObjective` and `auditFocus` when the incoming value is empty. The PKAT header `activityType` is untouched because PUT is partial and the frontend stops sending it.
   - New PKAT rows simply have them empty.
   - The `validateActivityType` call stays, for old clients only.
7. **PKAT year change:** the RKA year follows (it is derived) and `activity_plans.plan_year` is refreshed by the same AfterHook.
8. **PKAT deactivated (`isActive=false`):** existing RKAs keep working and stay editable. A new RKA can't be created against it, and the frontend picker hides inactive PKATs.
9. **PKAT deleted:** blocked when in use (409). For PKATs that were deleted before this change, RKA rows fall back to `source: "snapshot"` and the UI shows a warning banner.
10. **PKAT status:** an RKA can be drafted against any live, active PKAT. The UI warns when the PKAT is not `Approved`. There is no server block (Q4).

---

## 8. Frontend design

### 8.1 List (`ActivityPlanTable.vue`)

Columns, with frozen first two columns as today:
- Tahun
- Judul Kegiatan Audit (with a "Format lama" badge for legacy plans)
- Jenis Kegiatan Audit
- PKAT (code)
- Jumlah Penugasan
- Total Mandays
- Status
- Lampiran
- Aksi: view, edit and delete. Edit is hidden for legacy plans.

Filters: Tahun and Jenis Kegiatan Audit (`ActivityPlanFilter.vue`, modelled on `AnnualAuditFilter.vue`). The server is paginated with the same page loop as today (`stores/activity-plan.ts:203-227`).

### 8.2 Form (`ActivityPlanForm.vue`)

Rewrite it as a modal widened to `sm:max-w-6xl` with 3 steps plus a preview.

1. **Kepala Formulir:** PKAT select, which shows Tahun read-only with a status warning; Jenis Kegiatan Audit select; Judul.
2. **Penugasan.** The button "Tambah penugasan dari PKAT" opens `AnnualActivityPicker.vue`: a multi-select over `GET /activity-plans/annual-activity-options`.
   - Taken rows are shown disabled with "Sudah dipakai di <title>".
   - When a row is added, its `prefill` values are applied.
   - Each row renders as `ActivityPlanRowCard.vue`:
     - A read-only strip, "Dari PKAT", with No, Objek, Audit Universe and risk badge, department, Jadwal, Jumlah Tim and Mandays.
     - Editable fields: Kode Penugasan (with a placeholder hint "kosongkan untuk dibuat otomatis", disabled when `assignmentCodeEditable` is false), Jenis Audit (select with create), three `NumberedListInput` fields (Justifikasi, Tujuan, Fokus), and Ketua Tim (employee `USelectMenu` with the PKAT supervisor shown as help text).
     - Remove and move up/down controls. Remove is disabled with a tooltip when `inUse`.
   - A footer shows totals: rows, Σ team and Σ mandays.
3. **Lampiran:** the existing upload flow (`uploadAttachments`, `/media/upload`), unchanged.

**Pratinjau formulir:** a button in the footer opens `ActivityPlanFormSheet.vue`, a read-only render of the paper layout. It has a header block (Tahun, Jenis Kegiatan Audit with the chosen option ticked and the others shown, Judul), then the 12-column table, with numbered lists rendered as `<ol>`. The same component is the body of `ActivityPlanViewModal.vue`.

### 8.3 `NumberedListInput.vue` (new, `components/shared/`)
- `v-model: string[]`.
- Each item is an auto-resizing `UTextarea` with a number gutter. Items can be added, removed and moved up/down.
- Pasting multi-line text into an item splits it into items and strips leading `1.` / `1)`. This uses the same normaliser as the server, put in `utils/numberedList.ts`.
- Shows a counter (`n/20`) and a per-item length check (1000).

### 8.4 Validation (client mirrors server; blocking errors in the step alert)
- **Step 1:** PKAT, Jenis Kegiatan Audit and Judul are required. Judul is at most 255.
- **Step 2:**
  - At least 1 row.
  - Each row needs `auditKind`, and at least 1 non-empty item in each of Justifikasi, Tujuan and Fokus.
  - Items are at most 1000 characters each, and at most 20 per list.
  - `assignmentCode` matches the pattern and is unique within the form.
  - No PKAT row appears twice.
- **Server 409s** are mapped by `error.code` to `auditActivityPlan.errors.<CODE>`, interpolated with the parsed `details`. The step is focused and the row is scrolled to via `details.row`.
- **Non-blocking warnings:** PKAT not Approved; a row with `source: "snapshot"`.

### 8.5 Annual Audit Plan changes
- **`AnnualAuditForm.vue`:**
  - Remove the Jenis Kegiatan Audit select (`:49-59`), the Kode Penugasan input (`:143-154`), and the Justifikasi, Tujuan and Fokus textareas (`:275-310`).
  - Relabel the title field as "Judul PKAT" and `supervisorName` as "Supervisor (Pengendali Teknis)".
  - A row with `activityPlanRef` shows a small "Dipakai di RKA: <title> (<code>)" note. Its delete button is disabled with a tooltip.
- **`utils/annualPlanPayload.ts`:** stop sending `activityType`, `activityCode`, `justification`, `auditObjective` and `auditFocus`. The server preserves them.
- **`AnnualAuditDetail.vue`:** drop the removed fields and show `activityPlanRef`.
- **`stores/annual-audit.ts`:**
  - `isUsed` now comes from the server.
  - Map the 409 codes `ANNUAL_PLAN_IN_USE`, `ANNUAL_PLAN_ACTIVITY_IN_USE` and `ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT` to i18n messages.
  - Keep the uncommitted delete/deactivate flow as it is in the working tree.

### 8.6 i18n keys (add to both `frontend/locales/en/common.json` and `frontend/locales/id/common.json`)

| Key (under `auditActivityPlan.`) | EN | ID |
|---|---|---|
| `formTitle` | Audit Activity Plan Form | Formulir Rencana Kegiatan Audit |
| `stepper.header` / `stepper.rows` / `stepper.attachments` | Form header / Assignments / Attachments | Kepala Formulir / Penugasan / Lampiran |
| `preview` | Form preview | Pratinjau formulir |
| `header.annualPlan` | Annual Audit Plan (PKAT) | Program Kerja Audit Tahunan (PKAT) |
| `header.annualPlanPlaceholder` | Select a PKAT | Pilih PKAT |
| `header.annualPlanNotApproved` | This PKAT is still {status}. You can draft the plan; it follows any later PKAT change. | PKAT ini masih berstatus {status}. Rencana dapat disusun dan akan mengikuti perubahan PKAT. |
| `header.year` | Year | Tahun |
| `header.activityType` | Audit Activity Type | Jenis Kegiatan Audit |
| `header.title` | Audit Activity Title | Judul Kegiatan Audit |
| `header.titlePlaceholder` | e.g. Assurance Activity Plan 2026 | Contoh: Rencana Kegiatan Audit Assurance 2026 |
| `activityTypes.assurance` / `.specialAudit` / `.specificPurpose` / `.investigation` | Assurance / Special Audit / Specific-Purpose Audit / Investigative Audit | Assurance / Audit Khusus / Audit Tujuan Tertentu / Audit Investigasi |
| `columns.no` | No. | No |
| `columns.assignmentCode` | Assignment Code | Kode Penugasan |
| `columns.auditObject` | Audit Engagement Object | Objek Penugasan Audit |
| `columns.auditKind` | Audit Type | Jenis Audit |
| `columns.auditUniverse` | Audit Universe | Audit Universe |
| `columns.justification` | Justification for Object Selection | Justifikasi Pemilihan Objek |
| `columns.objectives` | Audit Objectives | Tujuan Audit |
| `columns.focus` | Audit Focus | Fokus Audit |
| `columns.schedule` | Engagement Schedule | Jadwal Penugasan |
| `columns.teamSize` | Team Size (Persons) | Jumlah Tim (Orang) |
| `columns.totalMandays` | Total Mandays | Total Mandays |
| `columns.teamLeader` | Team Leader | Ketua Tim |
| `auditKinds.operational` / `.financial` / `.compliance` / `.it` / `.performance` | Operational / Financial / Compliance / Information Technology / Performance | Operasional / Keuangan / Kepatuhan / Teknologi Informasi / Kinerja |
| `rows.addFromPkat` | Add assignments from PKAT | Tambah penugasan dari PKAT |
| `rows.empty` | No assignments yet. Add them from the PKAT. | Belum ada penugasan. Tambahkan dari PKAT. |
| `rows.fromPkat` | From PKAT (edit it in the Annual Audit Plan) | Dari PKAT (ubah di Rencana Audit Tahunan) |
| `rows.assignmentCodeHint` | Leave blank to generate a code | Kosongkan untuk dibuatkan kode otomatis |
| `rows.assignmentCodeLocked` | Locked: the engagement has started | Terkunci: penugasan sudah berjalan |
| `rows.teamLeaderPlaceholder` | Select an employee | Pilih pegawai |
| `rows.pkatSupervisor` | PKAT supervisor: {name} | Supervisor PKAT: {name} |
| `rows.pkatMissing` | The PKAT row no longer exists; showing saved values | Baris PKAT sudah tidak ada; menampilkan data tersimpan |
| `rows.inUseTooltip` | Cannot remove: the engagement has started | Tidak dapat dihapus: penugasan sudah berjalan |
| `rows.moveUp` / `rows.moveDown` / `rows.remove` | Move up / Move down / Remove | Naikkan / Turunkan / Hapus |
| `list.addItem` | Add item | Tambah butir |
| `list.itemPlaceholder` | Item {num} | Butir {num} |
| `list.pasteHint` | Pasting a numbered list splits it into items | Menempel daftar bernomor akan dipecah per butir |
| `picker.title` | Select PKAT activities | Pilih kegiatan PKAT |
| `picker.usedIn` | Already in "{title}" | Sudah dipakai di "{title}" |
| `picker.noMatch` | No PKAT activities of type {type} | Tidak ada kegiatan PKAT berjenis {type} |
| `picker.confirm` | Add {count} assignment(s) | Tambahkan {count} penugasan |
| `totals.rows` / `totals.teamSize` / `totals.mandays` | Assignments / Total team / Total mandays | Jumlah penugasan / Total tim / Total mandays |
| `table.year` / `.title` / `.activityType` / `.pkat` / `.rowCount` / `.totalMandays` / `.status` | Year / Title / Activity Type / PKAT / Assignments / Total Mandays / Status | Tahun / Judul / Jenis Kegiatan / PKAT / Penugasan / Total Mandays / Status |
| `filter.year` / `filter.activityType` / `filter.all` | Year / Activity type / All | Tahun / Jenis kegiatan / Semua |
| `legacy.badge` | Old format | Format lama |
| `legacy.notice` | This plan uses the old form. It can be viewed or deleted, not edited. Create a new plan in the form format. | Rencana ini memakai formulir lama. Hanya dapat dilihat atau dihapus. Buat rencana baru sesuai formulir. |
| `legacy.section` | Old-format data | Data format lama |
| `validation.annualPlanRequired` | Select a PKAT | PKAT wajib dipilih |
| `validation.activityTypeRequired` | Select the audit activity type | Jenis Kegiatan Audit wajib dipilih |
| `validation.titleRequired` | Audit activity title is required (max 255 characters) | Judul Kegiatan Audit wajib diisi (maks. 255 karakter) |
| `validation.rowsRequired` | Add at least one assignment | Tambahkan minimal satu penugasan |
| `validation.rowAuditKind` | Assignment {num}: Audit Type is required | Penugasan {num}: Jenis Audit wajib diisi |
| `validation.rowListRequired` | Assignment {num}: {field} needs at least one item | Penugasan {num}: {field} minimal satu butir |
| `validation.rowListLimits` | Assignment {num}: {field} allows up to 20 items of 1000 characters | Penugasan {num}: {field} maksimal 20 butir @1000 karakter |
| `validation.assignmentCodeFormat` | Assignment {num}: code may use letters, digits, . _ / - (max 50) | Penugasan {num}: kode hanya huruf, angka, . _ / - (maks. 50) |
| `validation.assignmentCodeDuplicate` | Assignment {num}: code is used twice in this form | Penugasan {num}: kode dipakai dua kali di formulir ini |
| `errors.ANNUAL_PLAN_NOT_FOUND` | The PKAT no longer exists | PKAT sudah tidak ada |
| `errors.ANNUAL_PLAN_INACTIVE` | The PKAT is deactivated | PKAT sudah dinonaktifkan |
| `errors.ANNUAL_ACTIVITY_NOT_FOUND` | Assignment {row}: the PKAT row no longer exists | Penugasan {row}: baris PKAT sudah tidak ada |
| `errors.ANNUAL_ACTIVITY_TYPE_MISMATCH` | Assignment {row}: its PKAT type does not match the activity type | Penugasan {row}: jenis di PKAT tidak sesuai Jenis Kegiatan Audit |
| `errors.ANNUAL_ACTIVITY_TAKEN` | Assignment {row} is already in "{activityPlanTitle}" | Penugasan {row} sudah dipakai di "{activityPlanTitle}" |
| `errors.ASSIGNMENT_CODE_TAKEN` | Assignment {row}: code {assignmentCode} is already used | Penugasan {row}: kode {assignmentCode} sudah dipakai |
| `errors.ASSIGNMENT_CODE_LOCKED` | Assignment {row}: the code cannot change after the engagement started | Penugasan {row}: kode tidak dapat diubah setelah penugasan berjalan |
| `errors.ACTIVITY_ROW_IN_USE` | Assignment {row} has started and cannot be removed | Penugasan {row} sudah berjalan dan tidak dapat dihapus |
| `errors.ACTIVITY_PLAN_IN_USE` | This plan has started assignments and cannot be deleted | Rencana ini memiliki penugasan yang sudah berjalan dan tidak dapat dihapus |
| `errors.ACTIVITY_PLAN_DUPLICATE` | A plan of this type already exists for this PKAT | Rencana dengan jenis ini untuk PKAT tersebut sudah ada |
| `errors.ACTIVITY_PLAN_STALE` | Someone else changed this plan. Reload and try again. | Rencana ini telah diubah pengguna lain. Muat ulang lalu coba lagi. |
| `errors.ACTIVITY_PLAN_LEGACY_READONLY` | Old-format plans cannot be edited | Rencana format lama tidak dapat diubah |

Annual Audit Plan keys:

| Key | EN | ID |
|---|---|---|
| `annualAudit.form.planTitle` (value change) | PKAT title | Judul PKAT |
| `annualAudit.form.teamLead` (value change, Q1) | Supervisor | Supervisor (Pengendali Teknis) |
| `annualAudit.form.usedInActivityPlan` | Used in activity plan "{title}" ({code}) | Dipakai di Rencana Kegiatan Audit "{title}" ({code}) |
| `annualAudit.errors.ANNUAL_PLAN_IN_USE` | This PKAT is used by an activity plan and cannot be deleted | PKAT ini dipakai Rencana Kegiatan Audit dan tidak dapat dihapus |
| `annualAudit.errors.ANNUAL_PLAN_ACTIVITY_IN_USE` | Row {row} is used in "{activityPlanTitle}". Remove it there first. | Baris {row} dipakai di "{activityPlanTitle}". Hapus dari sana terlebih dahulu. |
| `annualAudit.errors.ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT` | Row {row}: its type cannot change while it is in "{activityPlanTitle}" | Baris {row}: jenisnya tidak dapat diubah selama dipakai di "{activityPlanTitle}" |

**Keys to retire** once nothing references them (the translation agent checks parity afterwards):
- `auditActivityPlan.form.*` for the removed steps: resources, budget, review, priority, auditee, estimatedSchedule, budgetEstimation, planPeriod, department, creationDate and similar.
- `auditActivityPlan.stepper.step4` and `step5`.
- `auditActivityPlan.view.*` for removed sections. Keep the ones the legacy section reuses.
- `annualAudit.form.activityType`, `activityTypePlaceholder`, `assignmentCode`, `assignmentCodePlaceholder`, `justification`, `auditObjective`, `auditFocus`.

---

## 9. Backend checklist (`auditsphere-backend`), in order

- [ ] **B1. Models and migration.**
  - Files: `models/audit-activity.go` (section 5.1); new `cmd/migrate_activity_plan_form.go` (section 5.2); call sites in `cmd/serve.go` (after line 114) and `cmd/migrate.go` `runMigrations`.
  - Accept when:
    - running it twice against a copy of a prod dump changes nothing the second time;
    - `activity_plans` and `audit_activities` row counts are unchanged;
    - both partial indexes exist;
    - existing `activity_plans` rows have `form_version = 1`;
    - `go test ./...` passes.
- [ ] **B2. RKA domain package**, `services/activity_plan_form/`, with tests:
  - list-text normaliser (`Normalize([]string) []string`, `Join`, `Split`);
  - header-type↔category matcher (`activitycode.TypeCode`, section 4.3);
  - PKAT resolver: load the plan, index rows by id, report missing ones;
  - the "in use" check (section 7.4), done in one query per plan.
  - Accept when there are table tests for the normaliser (including `"1. a\n2) b\n\n"` → `["a","b"]`), the matcher (including legacy `"ASSURANCE"`), and `inUse` against a sqlite or pg test DB.
- [ ] **B3. Endpoints** from sections 6.1-6.5 and 6.7.
  - New `controllers/activity_plan/form_controller.go`; replace `routes/handler.go:167-176`, registering `/annual-activity-options` **before** `/:id` the same way `/audit-activities/project` is done.
  - Retire `AssignCodesOnCreate`/`AssignCodesOnUpdate` from the routes, but keep `KeepCodes` and its tests until legacy data is gone.
  - Accept when there are handler tests for:
    - create with a generated code and with a user code;
    - adopt a legacy row and restore a soft-deleted row;
    - each 4xx code in section 6.7;
    - a PUT that removes an in-use row (409, nothing written);
    - concurrent duplicate add (index → 409);
    - legacy GET (the `legacy` object filled);
    - legacy PUT (409);
    - no-`rows` body (400).
- [ ] **B4. Annual plan read decorators and the delete guard** (section 6.6). Files: `routes/handler.go:155-165`, new `controllers/annual_plan/annual_plan_usage.go`. Accept when `isUsed` and `activityPlanRef` are correct for an unused, a used and a seeded PKAT, and DELETE on a used PKAT returns 409 with nothing deleted.
- [ ] **B5. Annual plan update rules** (sections 7.2, 7.3, 7.5, 7.6). Files: `controllers/annual_plan/annual_plan_activities.go` (preserve deprecated fields; 409 checks); an `AfterHook` for the snapshot refresh (`crud.UpdateWithHooks`). Update the doc comment at `models/audit-annual-plan.go:17-40` so it no longer claims the RKA columns live there. Accept when there are tests for:
  - a PKAT PUT without `justification` keeps the stored value;
  - removing a used row returns 409;
  - a category change on a used row returns 409 only when it breaks the mapping;
  - renaming a used row updates `audit_activities.title` in the same transaction.
- [ ] **B6. Guard on `/audit-activities` PUT.** For rows with `activity_plan_id` set, reject changes to `title`, `target_unit_id` and the plan dates with 409 `ACTIVITY_ROW_MANAGED_BY_PLAN`; `status` stays writable. File: `services/audit_activity/service.go`. Accept with a unit test.
- [ ] **B7. Docs.** Regenerate swagger (`docs/docs.go`) if the team keeps it current, and update the Postman "Audit Activity Plan" folder. Its URLs say `/audit-activity-plans`, which is stale; the real path is `/activity-plans`. Accept when the Postman requests run green against local.

## 10. Frontend checklist (`auditsphere-frontend`), in order

The work can start immediately against the contract. Mock it with fixtures copied from section 6.

- [ ] **F1. Types** in `frontend/types/audit.ts`:
  - add `ActivityPlanActivityType`, `ActivityPlanSummary`, `ActivityPlanDetail`, `ActivityPlanRow`, `ActivityPlanRowInput`, `AnnualActivityOption`, `ActivityPlanErrorDetails`;
  - rename the old `ActivityPlan`/`PlannedAuditActivity`/`ResourceAuditor`/`ActivityPlanBudget`/`ActivityPlanReview` to `Legacy*`, used only by the legacy view;
  - add `activityPlanRef?` and server `isUsed` to `AuditActivities`/`AnnualAuditPlan`.
  - Accept when `nuxi typecheck` is clean.
- [ ] **F2. Utility** `frontend/utils/numberedList.ts` (normalise, split on paste), with unit tests mirroring B2.
- [ ] **F3. Store rewrite** `frontend/stores/activity-plan.ts`:
  - list with filters; get; create and update with the section 6.4 payload (rows in display order, `updatedAt` echoed); delete; options;
  - error-code → i18n mapping that parses `error.details`;
  - keep `uploadAttachments`;
  - remove `linkAnnualActivity`, budget and auditor logic.
  - Accept with unit tests for: payload shape (no PKAT-owned fields sent), `assignmentCode` `""` for blank, 409 mapping, and legacy plans flagged read-only.
- [ ] **F4. Components:**
  - `components/shared/NumberedListInput.vue`;
  - `components/audit-activity-plan/AnnualActivityPicker.vue`, `ActivityPlanRowCard.vue`, `ActivityPlanFormSheet.vue`, `ActivityPlanFilter.vue`;
  - rewrite `ActivityPlanForm.vue` (section 8.2), `ActivityPlanViewModal.vue` (sheet plus legacy section) and `ActivityPlanTable.vue` (section 8.1). Keep the uncommitted sticky-column fix (`frozenTableUi`).
  - Accept with mounted-component tests for:
    - the form sheet renders the 12 columns in the paper order with numbered lists;
    - PKAT-owned fields are disabled;
    - a taken option is disabled;
    - edit is hidden for legacy plans.
- [ ] **F5. Annual plan cleanup** (section 8.5):
  - `AnnualAuditForm.vue`, `AnnualAuditDetail.vue`, `utils/annualPlanPayload.ts`, `stores/annual-audit.ts`;
  - update `tests/unit/annualPlanPayload.test.ts`, `annualPlanActivityCode.test.ts` and `annualPlanSnapshots.test.ts` to the new payload (the removed keys are not sent).
  - Accept when the payload test asserts that `activityType`, `activityCode`, `justification`, `auditObjective` and `auditFocus` are absent, and that the delete flow handles 409 `ANNUAL_PLAN_IN_USE` by offering deactivation.
- [ ] **F6. i18n:** add the section 8.6 keys to both locales and remove the retired ones. Accept when the key-parity check passes (both files have identical key sets) and there are no hardcoded strings in the touched components.
- [ ] **F7. Retire or replace the old tests:** `tests/unit/activityPlanAnnualLink.test.ts` and `activityPlanActivityCode.test.ts` assert the old behaviour. Replace them with the F3 and F4 tests; don't just delete them. Accept when `vitest run` is green.

## 11. Rollout and hand-offs

1. **`auditsphere-backend`** does B1-B6 (B7 is optional). It depends on nothing. Done when the B-tests pass and `go build ./...` and `go vet ./...` are clean for audit-service.
2. **`auditsphere-frontend`** does F1-F7 in parallel with step 1, against the section 6 fixtures. Done when vitest and typecheck are green.
3. **`auditsphere-software-translation`** reviews the ID/EN copy in section 8.6 and checks key parity after F6. It depends on F6. Done when the parity report is clean.
4. **`auditsphere-qa`** runs the Go tests, vitest, typecheck and locale validation. It then runs a local smoke test:
   - create a PKAT with 3 rows (2 Assurance, 1 Investigation);
   - create an Assurance RKA with 2 rows, one with a user code and one blank;
   - confirm the Investigation row is not offered;
   - rename a PKAT row and confirm the RKA shows the new name and `audit_activities.title` changed;
   - try to delete the PKAT (409) and to remove a used row in the PKAT (409);
   - delete the RKA, then delete the PKAT (OK).
   It depends on steps 1 and 2. Done when the evidence report is all pass.
5. **`auditsphere-devops`** deploys to prod. It depends on step 4.
   - Backend via `backend/deploy-prod.sh --ref <sha>` with the migrate-only override. **No seeders.**
   - Take the per-tenant `pg_dump` and run the section 5.2 counts before the deploy.
   - Deploy the frontend (manual build) **immediately after** the backend in the same window.
   - There is no compatibility shim. During the gap the old frontend's Activity Plan list shows blank cells, and a save is rejected with 400 without writing anything.
   - Re-run the counts after the deploy; the row counts must be unchanged.
   - Done when the counts match and the smoke test passes on one tenant.

**Behaviour change to tell the product owner:** after this ships, KPI work-plan realisation and audit-completion figures will start counting real RKA assignments. Today they only see seed data in `audit_activities`.

## 12. Open questions for the product owner (the defaults above apply until answered)

1. **Ketua Tim and the PKAT "supervisor".** Are they the same person? The default treats them as different roles. Ketua Tim is chosen on the RKA and may be blank, as in your example. The PKAT keeps its supervisor, relabelled "Supervisor (Pengendali Teknis)". If they are the same role, Ketua Tim becomes read-only from the PKAT instead.
2. **One form per activity type per year?** The default is one RKA per (PKAT, Jenis Kegiatan Audit), so all Assurance assignments of 2026 go on one form. Do you ever need two forms of the same type in one year, for example one per directorate?
3. **Audit Universe code ("02B").** The Audit Universe has no code field today. The default shows the name, so a code typed into the name will appear. Should we add a proper code to the Audit Universe? That is separate work in risk-service.
4. **PKAT approval before an RKA?** The default lets you draft an RKA on any active PKAT and shows a warning if it isn't Approved yet. Should it be blocked until the PKAT is Approved?

Also decided by default; tell us if any of these is wrong:
- Jenis Audit is a suggested list (Operasional, Keuangan, Kepatuhan, Teknologi Informasi, Kinerja) that also accepts other values.
- Jadwal, Jumlah Tim and Mandays can only be changed in the PKAT.
- Old-format activity plans stay viewable but are not converted.

## 13. Out of scope / follow-ups
- Per-row month checklist in the PKAT. `overview.MD:134` asks for "time of activity (Jan to Dec) and allow checklist" per row; today `selectedMonths` is plan-level, and Jadwal is free text (`timelineText`).
- Surat Tugas (`/assignment-letter`) picking the assignment by Kode Penugasan from RKA rows. Today executions link to `audit_activities` only by title or code matching (`audit_execution_sync/sync.go:230-266`).
- Structured Excel import of the RKA form. `/audit-activity-plan/upload` stores documents only.
- Server-side RBAC for planning writes. The JWT is only checked for validity (unchanged).
