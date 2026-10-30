# Organisation structure model plan: configurable levels and a department tree (new tenants only)

Status: FINAL SPEC (revision 3, 2026-10-28). Plan only: no application code, migration or tenant has been changed. Verified against the repo on branch `staging`. Implementation agents follow this document; section 13 lists the files each phase touches.

Hard constraint: existing tenants on the VPS get zero DDL and zero data change. The flexible structure exists only for tenants created from now on through `onboard-tenant.sh` (and the site generator that prints its command). One build of master-service, audit-service, auth-service and the frontend runs on both kinds of tenant.

## 0. Decisions (2026-10-28)

| # | Decision | Where it is specified |
|---|---|---|
| D1 | No legacy/flat preset for new clients. Presets: `simple` (Departemen only, the default), `corporate`, `bank`, `custom`. All on the flexible schema, gapped levels 10/20/30/40, `child.level > parent.level` (skipping allowed), max depth 6. | 4.1, 4.2, 7 |
| D2 | Legacy tenants: the Assignment Letter unit field becomes a dropdown of master departments (flat). The chosen department's name is saved into the existing `working_unit` text. No schema change. Validated server-side. | 5.2, 5.3 |
| D3 | One stored name per level; the frontend is bilingual. Preset levels carry a stable `name_key` (`directorate`, `division`, `department`, `section`) translated via i18n; custom or renamed levels are shown as typed. | 4.1, 8 |
| D4 | The letter's unit id is validated server-side by audit-service against the locally mirrored department refs; unknown, inactive, deleted or wrong-company ids are rejected; name/path/level name are taken from the server-side record. | 4.5, 5.3 |
| D5 | Renaming or moving a unit republishes `department.updated` for its whole subtree (bounded by depth 6). | 4.3, 6 |
| D6 | Audit team (SKAI/SPI/CAE) becomes a dropdown in both modes. Source: a server-side constant list in audit-service, served by an options endpoint and validated on save. | 5.4 |
| D7 | The bank preset demo seeds a sample tree including a level-skip and a duplicate leaf name under two parents. | 4.2.1 |
| D8 | Only roles `ADMIN` or `EXECUTIVE` may change org levels (create, rename, renumber, delete). Enforced in master-service and gated in the frontend. | 9.1 |
| D9 | Level changes go through a dry-run impact preview, then confirmation by re-entering the password (new auth-service step-up endpoint, none exists today), applied in one transaction and written to an org change log. Renumbering that breaks `child.level > parent.level` is rejected. | 9.2-9.4 |
| D10 | Still open: `department_code` uniqueness on new tenants. Default in this spec: globally unique among live rows (partial on `deleted_at`). | 12 |

## 1. Summary of the design

1. One org-unit table (`departments`) plus a per-tenant `org_levels` table and a parent link. No Division/Directorate tables.
2. Opt-in flag `ORG_STRUCTURE_MODE=flexible`, written only into a NEW tenant's compose file by `generate-tenant-stack.sh` when `--org-preset` is given (every new tenant gets one; `simple` is the default the site generator prints). Absent = legacy. The `org_levels` table is the persisted marker.
3. All new DDL is raw, gated SQL; `models.Department` and `models.AssignmentLetter` gain no persisted fields. Otherwise the shared build would alter every legacy tenant on the next migrate.
4. Frontend learns the mode from `GET /api/v1/tenant-config` (master-service).
5. Level edits are privileged (ADMIN/EXECUTIVE), previewed, password-confirmed, transactional and logged.

## 2. Current state (verified)

| Area | Fact | Reference |
|---|---|---|
| Department model | Flat. `DepartmentCode` and `DepartmentName` are both `uniqueIndex` (global, not per company, not partial on `deleted_at`). `Level int not null` is a free number with no meaning table. Has `CompanyID` (not null), optional `BusinessUnitID`, `PicID`. No parent. | `backend/master-service/models/department.go` |
| Business unit | Already has `ParentBusinessUnitID` (self-reference). Precedent for naming and shape. | `backend/master-service/models/business_unit.go` |
| Employee | `DepartmentID` not null, FK to Department (same DB, enforced). `EmployeeCode` unique. | `backend/master-service/models/employee.go` |
| Audit universe | `DepartmentID` FK inside master DB (enforced). Entity type `DEPARTMENT`. | `backend/master-service/models/audit_universe.go` |
| Annual plan activity | Many-to-many `InvolvedDepartments` inside master DB. | `backend/master-service/models/annual_audit_plan_activity.go` |
| Migration | `departmentMigration` shadow struct breaks the Department/Employee FK cycle on first create, then real `models.Department` is auto-migrated after Employee. | `backend/master-service/cmd/migrate.go` lines 24-40, 215-222 |
| Constraint to API error map | Names `idx_departments_department_code` and `idx_departments_department_name`. Any index rename or replacement must be added here or users get a raw DB error. | `backend/master-service/pkg/errors/constraints.go` lines 15-19 |
| Department API | `GET/POST /departments`, `GET/PUT/DELETE /departments/:id`, `POST /departments/events/resync`. `Delete` has no child or usage guard of its own; it relies on DB FKs (for example `fk_employees_department`). | `backend/master-service/routes/routes.go` lines 61-70, `services/department/department.go` |
| Department events | Topic `department.events`, payload `{id, department_code, department_name, is_active, company_id, business_unit_id}`. Outbox-based; resync endpoint exists for repair. | `backend/master-service/pkg/kafka/domain_publisher.go` |
| Audit-service consumer | `ApplyDepartment` upserts an `AnnualSnapshotRef{Kind: department, RefID, Name}` in the audit DB and rewrites `AnnualAuditActivity.Department` (name) by `DepartmentID`; legacy rows are linked by case-insensitive name. | `backend/audit-service/services/annual_plan_sync/sync.go` lines 84-110, `models/annual-snapshot-ref.go` |
| Annual plan snapshot | `DepartmentID string`, `Department string` (name snapshot). | `backend/audit-service/models/audit-annual-plan.go` lines 42-56 |
| Activity plan | `TargetUnitID *uuid.UUID` (soft ref to a department) and `Department string`. | `backend/audit-service/models/audit-activity.go` lines 120, 169 |
| Assignment letter | `WorkingUnit string(255)` free text; `CompanyID *uuid`, `CompanyName`; `AuditTeam string(100)` (values `SKAI`, `SPI`, `CAE`) which is also an input to `NextLetterNumber`; `MembersList []LetterMember{Name, Role}` stored as JSON. | `backend/audit-service/models/assignment-letter.go` lines 13-45, 205-212 |
| Working paper | `TeamMember{ID int64, Name, Role}`; `syncFromAssignmentLetter` copies names from `MembersList` only when the paper has none. | `backend/audit-service/models/working-paper.go` lines 13-17, 71-88 |
| Letter reads of `WorkingUnit` | Only the model, the seeder (`cmd/seed.go`: Finance, IT, Operations, Procurement, Maintenance, HR) and frontend (`types/audit.ts`, `AssignmentLetterTable.vue`, `AuditorDashboard.vue`, form, store). No backend report reads it. | grep over `backend/` and `frontend/` |
| Risk-service | `RiskProfile.DepartmentID uuid` (soft ref, accountable owner), `RiskControlMatrix.Department varchar(100)` text, `AuditeeSurvey.Department` text (audit-service). | `backend/risk-service/models/risk_profile.go`, `risk_control_matrix.go`; `backend/audit-service/models/auditee_survey.go` |
| Analytics | Reads over HTTP, owns no tenant DB; department is only an input to risk scoring. | `backend/analytics-service/services/analytics_service.go` |
| Frontend: form | Step 2 uses `store.options.workingUnit`, a hard-coded array `Production, Marketing, Finance, IT, Operations, Procurement, Maintenance`; members are `UInput` name + `USelectMenu` role from `['Person in Charge','Supervisor','Chairperson','Member']`; company select sits below the unit select. "Audit Team" (`store.options.auditTeam`) is in step 1. | `frontend/components/assignment-letter/AssignmentLetterForm.vue` lines 169-175, 218-320; `frontend/stores/assignment-letter.ts` lines 65-73 |
| Frontend: master options | `departmentOptions` label is `name (code)`, no company filter, no path. | `frontend/stores/master-options.ts` lines 44-48 |
| Frontend: department CRUD | `DepartmentForm.vue` takes `level` as a number input; `DepartmentTable.vue` shows "Level N". `types/master.ts` `Department` has no parent. | `frontend/components/master/DepartmentForm.vue` line 68, `DepartmentTable.vue` line 93 |
| Frontend: settings | Tabs are a fixed list `profile, settings, mfa, activity, permissions, datasource, faq`; there is no organisation tab. | `frontend/pages/settings/index.vue` lines 150, 185-230 |
| Site generator | UI simulation: only the Resend provisioning call is real. Step 1 "Organization" collects client name, industry, slug, ports, brand colour. It prints the `onboard-tenant.sh` command line. | `frontend/pages/site-generator/index.vue` (`siteGenerator.organization.*`, line 1113) |
| Onboarding | Order: `auth migrate up`, `auth seed`, `audit migrate up`, `master migrate`, risk catalogue seed, then demo seeders only if not `--empty-data` (default is empty). Registry columns are fixed: `slug, client_name, domain, api_domain, fe_port, kong_port, kong_admin_port, redis_db, created_at`. | `scripts/onboard-tenant.sh` lines 160-220, `backend/tenants/registry.tsv` |
| Demo seeder | Four departments, all `Level: 1`, one company. | `backend/master-service/pkg/database/seeders/seeder.go` lines 398-451 |

### 2.1 Additional findings for the decisions (verified)

| Area | Fact | Reference |
|---|---|---|
| Role names | Seeded roles: `ADMIN`, `AUDITOR`, `EXECUTIVE`, `AUDITEE`, `VIEWER`, `DEPARTMENT_HEAD`. `EXECUTIVE` is the seeded role that acts as CAE. | `backend/auth-service/pkg/database/seeders/seeder.go` lines 141-168; `frontend/composables/useRbac.ts` line 11 |
| master-service RBAC | Every `/api/v1` route only requires `Authenticate()`; no route uses `RequireRoles` today, so any logged-in user can write departments. `RequireRoles` exists but compares exact strings (case-sensitive), unlike auth-service. The JWT `roles` claim is put into the context. | `backend/master-service/routes/routes.go` lines 36-40; `pkg/middleware/auth.go` lines 15, 71, 82-110 |
| Frontend RBAC | `matchesRole` treats `executive`, `cae`, `chiefauditexecutive` as aliases, and `isExecutive` also matches viewer/board keywords (`viewer`, `direksi`, `komisaris`, ...). There is a role-preview mode (`activeRoleView`). So `isExecutive` must NOT be used to gate level edits. | `frontend/composables/useRbac.ts` lines 6, 70-90, 159 |
| Re-auth | auth-service has no verify-password / step-up endpoint. It has `POST /auth/change-password` (protected; body `old_password`, `new_password`) and MFA verify endpoints. | `backend/auth-service/routes/registry.go` lines 57-70; `models/dto.go` line 61 |
| Audit trail | No generic audit-trail table or service exists. Kafka auth events (`password.changed`, `role.assigned`, ...) are only logged; the Settings "activity" tab shows profile and trusted devices. | `backend/master-service/pkg/kafka/events.go`; `frontend/components/settings/activity.vue` |
| Audit team UI | Radio buttons over `store.options.auditTeam` = `SKAI` (Satuan Kerja Audit Internal), `SPI` (Satuan Pengawas Intern), `CAE` (Chief Audit Executive); step 1 of the form. No server-side check of the value; it feeds `NextLetterNumber`. | `AssignmentLetterForm.vue` lines 164-183; `stores/assignment-letter.ts` lines 66-70; `models/assignment-letter.go` lines 205-212 |
| Snapshot refs | `annual_snapshot_refs` holds only `kind, ref_id, year, name, risk_level, updated_at`; rows are never deleted; `department.deleted` is not applied. It cannot check active/deleted/company today. | `backend/audit-service/models/annual-snapshot-ref.go`; `services/annual_plan_sync/handler.go` line 82 |
| audit -> master client | audit-service already has `pkg/masterclient` (lists departments) and `pkg/servicetoken`; tenant compose sets `MASTER_SERVICE_URL` to the tenant's own master. | `backend/audit-service/pkg/masterclient/client.go`; `scripts/templates/docker-compose.tenant.yml.tpl` line 103 |
| Letter CRUD routes | Letter create/update live in `backend/audit-service/routes/handler.go`, which has uncommitted changes on `staging` right now (with `controllers/assignment_letter/publish.go`). | git status |

## 3. Opt-in mechanism and the zero-DDL guarantee

### 3.1 How migrate runs today (verified)

| Path | What runs | Touches tenants? |
|---|---|---|
| `.github/workflows/deploy.yml` -> `backend/deploy-prod.sh` | Starts the NON-tenant prod stack (`rb_audit`) with a migrate-only override: `./master migrate && ./master serve`, `./audit migrate up && ./audit serve` (lines 129-147); refuses seeder commands (lines 238-244). Line 30: never touches `*_dev` containers or `backend/tenants/*`. | No tenant, but re-runs migrate on the prod DB every deploy. |
| Tenant stack (`scripts/templates/docker-compose.tenant.yml.tpl`) | master command `./master serve`; no migrate at start. | Only when an operator runs it. |
| `scripts/onboard-tenant.sh` step 4 | `./master migrate`, `./audit migrate up` via `compose run`; re-onboarding an existing slug re-runs them. | Yes: the only automatic route into a tenant DB. |
| Manual `compose run ... migrate` | Same. | Operator action. |

So the new binary's migrate must be a strict no-op for this feature on the prod stack and on every legacy tenant, and re-onboarding a legacy tenant must never enable it.

### 3.2 The flag

- `ORG_STRUCTURE_MODE=flexible`, read by master-service and audit-service config (viper env binding; default empty = legacy). auth-service does not need it.
- Written by `generate-tenant-stack.sh` into the `master-service-<slug>` and `audit-service-<slug>` environment, rendered only when `--org-preset` is passed, and only when the stack for that slug is generated for the first time.
- `onboard-tenant.sh` refuses `--org-preset`/`--org-levels` when the slug already exists in `backend/tenants/registry.tsv` before this run ("converting an existing tenant is out of scope"). Without the flags it behaves exactly as today.
- Persisted marker: the `org_levels` table. Effective mode = flag on AND `org_levels` exists. Flag off with the table present: API behaves as legacy, logs a warning, migrate touches nothing. Flag on with the table missing: gated DDL runs on the next migrate (the normal new-tenant path).
- Mode is resolved once per process (config + one catalog query, cached); no per-request DDL.

### 3.3 DDL that runs only when the flag is on

master `migrate`, after the normal AutoMigrate, in `migrateOrgStructure(db)` (returns immediately unless the flag is on); every statement is `IF NOT EXISTS`/`IF EXISTS` so it is idempotent:

1. `CREATE TABLE org_levels` (4.1).
2. `CREATE TABLE org_change_log` (9.4).
3. `ALTER TABLE departments ADD COLUMN parent_department_id uuid NULL`, FK to `departments(id)` `ON DELETE RESTRICT`, index `idx_departments_parent` on `(company_id, parent_department_id)`, `CHECK (parent_department_id <> id)`.
4. `DROP INDEX idx_departments_department_name` (global unique name).
5. `CREATE UNIQUE INDEX uq_departments_company_parent_name ON departments (company_id, COALESCE(parent_department_id, '00000000-0000-0000-0000-000000000000'), lower(department_name)) WHERE deleted_at IS NULL`.
6. Default for D10: `DROP INDEX idx_departments_department_code` and `CREATE UNIQUE INDEX uq_departments_code_live ON departments (department_code) WHERE deleted_at IS NULL` (still global, but deleted codes can be reused).

audit `migrate up`, same flag:

7. `ALTER TABLE assignment_letters ADD COLUMN auditee_unit_id uuid, auditee_unit_name varchar(255), auditee_unit_path varchar(500), auditee_unit_level_name varchar(50)`, index `idx_assignment_letters_auditee_unit`.
8. `ALTER TABLE annual_snapshot_refs ADD COLUMN company_id uuid, is_active boolean, deleted_at timestamptz, path varchar(500), level int, level_name varchar(50)` (needed for D4).

Nothing else in the plan is DDL. auth-service's step-up endpoint needs no schema (one-time token ids live in Redis).

### 3.4 Keeping the legacy path byte-identical

- `models.Department` gets no persisted field. Response-only fields (`ParentDepartmentID`, `LevelName`, `Path`) are `gorm:"-"` with `json:",omitempty"`. Flexible code reads/writes the parent through a separate struct `DepartmentOrg` (table `departments`, columns `id`, `parent_department_id`; never auto-migrated) with explicit `Select`/`Updates`.
- Same for letters: `AssignmentLetter` gets only `gorm:"-"` snapshot fields, hydrated and saved through `AssignmentLetterOrg` (table `assignment_letters`, the four `auditee_unit_*` columns) in flexible mode; and `AnnualSnapshotRefOrg` for the extra ref columns.
- `OrgLevel` and `OrgChangeLog` are NOT in the AutoMigrate list; they are created by the gated SQL.
- The global unique name index is declared by the `uniqueIndex` tag on `models.Department.DepartmentName` and on the `departmentMigration` shadow struct (`cmd/migrate.go` lines 24-40). Kept, it would be re-created on every flexible migrate. Remove the tag from both and add a legacy step that runs `CREATE UNIQUE INDEX idx_departments_department_name ON departments (department_name)` only when the flag is off AND `org_levels` does not exist AND `pg_indexes` has no index of that name. On every existing tenant the index exists, so nothing executes; a from-scratch legacy DB (local dev, CI, fresh `rb_audit`) still gets it. Do the same for `idx_departments_department_code` (tag removed, legacy step re-creates it only if missing) so step 6 above is not undone. Removing a tag never makes AutoMigrate drop anything.
- The service-level duplicate-name check (`FindByName` -> `errDepartmentNameExists`) stays global in legacy, scoped to (company, parent) in flexible.

Proof: the schema-diff check (section 14, check 1) is mandatory.

## 4. Schema (flexible tenants only)

### 4.1 `org_levels` (master-service, `rb_audit_master_<slug>`)

| Column | Type | Notes |
|---|---|---|
| `id` | uuid pk | |
| `level` | int not null | Gapped number, `CHECK (level BETWEEN 1 AND 999)`, unique among live rows (`uq_org_levels_level WHERE deleted_at IS NULL`). |
| `name` | varchar(50) not null | Stored display name in the tenant's language (e.g. `Direktorat`). Unique case-insensitive among live rows. Used by backend, exports and letter snapshots. |
| `name_key` | varchar(30) null | Preset key: `directorate`, `division`, `department`, `section`. Set by the preset seeder; set to NULL by the backend whenever the name is changed by a user. `CHECK (name_key IN (...))`. |
| `is_active` | bool default true | |
| `created_at`, `updated_at`, `deleted_at` | standard | |

At most 6 live rows (constant `MaxOrgLevels = 6`). `name_key` is the minimal addition needed for D3: without it the frontend cannot tell a preset name (translate) from a typed one (show raw), for example an admin who types "Divisi" on an English tenant.

### 4.2 Level numbering and presets

| Preset | Levels (`level`, `name`, `name_key`) |
|---|---|
| `simple` (default) | 10 Departemen `department` |
| `corporate` | 10 Divisi `division`; 20 Departemen `department` |
| `bank` | 10 Direktorat `directorate`; 20 Divisi `division`; 30 Departemen `department`; 40 Seksi `section` |
| `custom` | Names from `--org-levels` in order, numbered 10, 20, 30...; `name_key` NULL |

Stored names default to Indonesian. `--org-levels` overrides the stored names for any preset (for example English names for an English-speaking client); when names are overridden, `name_key` is kept for preset levels only if the override is the preset default. The UI translates by `name_key` anyway, so the override matters for exports and letter snapshots.

Rules: `departments.level` must equal an active `org_levels.level` (service validation; no FK). `child.level > parent.level` (levels may be skipped; cycles impossible because levels strictly increase). Depth (number of ancestors + 1) <= 6. Changing a unit's level or parent must keep every existing child strictly deeper. New levels get a number between neighbours (Settings suggests the midpoint, e.g. 25), so existing units never need renumbering to insert a level.

Seeded demo departments: today `master seed` creates four departments at `Level: 1`. In flexible mode the seeder reads the tenant's levels and places them at the preset's department level (`simple` 10, `corporate` 20, `bank` 30, `custom` the deepest level). In legacy mode it is unchanged.

#### 4.2.1 Bank demo tree (D7)

Seeded only with `--with-demo-data --org-preset=bank`, idempotent by `department_code` (the seeder uses `FirstOrCreate` today), all in the seeded holding company:

```
Direktorat Utama (DIR-UTM, 10)
  Internal Audit (DEPT-003, 30)                 <- level skip: Departemen directly under Direktorat
Direktorat Keuangan (DIR-KEU, 10)
  Divisi Akuntansi (DIV-AKT, 20)
    Finance & Accounting (DEPT-001, 30)
    Departemen Pajak (DEPT-PJK-AKT, 30)          <- duplicate leaf name, parent 1
  Divisi Treasury (DIV-TRS, 20)
    Departemen Pajak (DEPT-PJK-TRS, 30)          <- duplicate leaf name, parent 2
Direktorat Operasional (DIR-OPS, 10)
  Operations (DEPT-004, 30)                     <- level skip
  Divisi Teknologi Informasi (DIV-TI, 20)
    Information Technology (DEPT-002, 30)
      Seksi Infrastruktur (SKS-INFRA, 40)
```

The four existing codes keep their employees (employees reference department ids). The walkthrough shows `Direktorat Keuangan › Divisi Akuntansi › Departemen Pajak` vs `... › Divisi Treasury › Departemen Pajak`.

### 4.3 `departments` (flexible)

`parent_department_id` and the replacement indexes from 3.3. Service rules, transactional: parent live and same `company_id`; level rules of 4.2; `BusinessUnitID` belongs to the same company (existing check; business unit stays an orthogonal axis); delete refused (409) while the unit has live children (employee and audit-universe FKs already block, mapped to readable errors). Rename or move: in the same transaction, enqueue `department.updated` through the existing outbox for the unit and all descendants (D5).

Path is derived, never stored in master: the tree is built in memory and returned as `path[] = {id, name, level, level_name, name_key}`; the UI joins with ` › `.

### 4.4 Assignment Letter (audit-service, flexible)

Columns from 3.3 step 7. `working_unit` stays and is set to the path label on every flexible write, so the table, dashboard and exports keep working. `LetterMember` JSON gains `employee_id`, `employee_code`; `name` stays authoritative for display. Working paper `TeamMember` gets the same optional fields, copied in `syncFromAssignmentLetter`. A snapshot (not a live link), because a signed, numbered letter must not change when a unit is renamed.

### 4.5 Local department mirror for validation (D4)

audit-service `ApplyDepartment` (flexible) also writes `company_id`, `is_active`, `path`, `level`, `level_name` to the ref row (via `AnnualSnapshotRefOrg`), and `department.deleted` sets `deleted_at` on the ref instead of being skipped. The event payload already carries `is_active` and `company_id`; `path`, `level`, `level_name` are the additive fields from section 6. The subtree republish of D5 keeps descendant paths fresh. Repair path: `POST /departments/events/resync` (exists) and `audit migrate backfill-annual-snapshots --from-source` (exists), extended to fill the new columns in flexible mode.

### 4.6 Cross-service integrity

Inside master: real FKs. audit/risk to master (`auditee_unit_id`, `AnnualAuditActivity.DepartmentID`, `TargetUnitID`, `risk_profiles.department_id`): business key only, not enforced, eventually consistent; D4 validation happens at write time only. Employee link: `employee_code` (the key `users.employee_id` uses) plus `employee_id`.

## 5. Behaviour by mode (same build on both kinds of tenant)

### 5.1 Matrix

| Area | Legacy (flag off) | Flexible (flag on) |
|---|---|---|
| `GET /tenant-config` (master) | `{"org_structure":{"mode":"legacy"}}` | `{"org_structure":{"mode":"flexible","max_levels":6,"levels":[{id,level,name,name_key}]}}` |
| `/departments` list/detail | Response unchanged (new keys omitted). | Adds `parent_department_id`, `level_name`, `name_key`, `path`; filters `company_id`, `level`, `parent_id`, `is_active`. |
| `POST/PUT /departments` | Unchanged. A body with `parent_department_id` returns 422 `ORG_STRUCTURE_NOT_ENABLED` (never silently ignored). Name unique globally; level a free number. | Rules of 4.2/4.3; name unique per (company, parent); subtree events on rename/move. |
| `/org-levels*`, `/departments/tree` | 404 `ORG_STRUCTURE_NOT_ENABLED`; no table access. | Available (section 9 for writes). |
| `department.*` events | Payload unchanged. | Additive `parent_department_id`, `level`, `level_name`, `path`. |
| Letter: unit | Dropdown of master departments; name saved in `working_unit`; validated via master (5.3). | Auditee unit picker (any level, path label); id validated against local refs; snapshot columns (5.3). |
| Letter: members | Free-text name + role as today. | Picked from the employee master (`employeeId`, `employeeCode`, `name`, `role`). |
| Letter: audit team | Dropdown from audit-service options (5.4). | Same. |
| Master demo seeder | `Level: 1`, unchanged. | Preset-aware; bank tree (4.2.1). |
| Settings "Organization" tab | Hidden. | Visible; level edits only for ADMIN/EXECUTIVE. |

### 5.2 How the frontend learns the mode

`GET /api/v1/tenant-config` (master-service, any authenticated user, served on every tenant, no DDL) is loaded with the master options in `stores/master-options.ts` and exposed as `orgStructure.mode` / `orgStructure.levels`. On error the UI falls back to legacy behaviour, never flexible. The `/tenant-config` prefix must be added to every Kong config.

### 5.3 Assignment Letter unit: UI and server-side validation

UI (both modes): step 2 order is company/entity, then "Auditee Unit", then audit team and members side by side. The unit select is filtered by the selected company; company becomes required once a unit is chosen. The hard-coded `options.workingUnit` list is removed.

Legacy mode (D2), no new columns:

- The dropdown lists live, active master departments of the selected company (`GET /departments`, flat, label `name (code)`). The client sends the stored text `workingUnit` and a transient `workingUnitDepartmentId` that is not persisted.
- audit-service on create, or on update when `workingUnit` differs from the stored value: requires `workingUnitDepartmentId`, fetches the department from the tenant's own master-service (`pkg/masterclient`, service token), and rejects with 422 if not found/deleted, `is_active = false`, or `company_id` differs from the letter's `companyId`. It then writes the master record's `department_name` into `working_unit` (the client's text is ignored).
- Why not the local refs here: on legacy tenants `annual_snapshot_refs` has no company/active/deleted columns and adding them would be DDL. If master is unreachable the save returns 503 "unit could not be verified, try again"; there is no unverified fallback.
- Old letters whose text matches no department: the form shows the stored text as the current value with a "not in master data" hint. Saving without touching the field keeps the text unchanged and needs no validation; changing it requires a valid pick. This also covers seeded letters (`Finance`, `IT`, ...) that never matched department names.

Flexible mode (D4):

- The picker lists departments of the selected company at any level with the path label. The client sends `auditeeUnitId` (the only unit field the server trusts).
- audit-service validates against the local mirror (`annual_snapshot_refs` kind `department`, columns of 4.5): reject 422 if no row, `deleted_at` set, `is_active = false`, or `company_id` differs from the letter's `companyId`. It writes `auditee_unit_name`, `auditee_unit_path`, `auditee_unit_level_name` and `working_unit` (path label) from the ref row; client-sent names are ignored.
- If the ref row is missing (event not yet consumed), audit-service falls back once to `masterclient` and upserts the ref from the answer; if both fail, 503.
- Unchanged unit on update is accepted even if the unit has since been deactivated (the letter is history).

### 5.4 Audit team dropdown (D6, both modes)

- Source: a constant list in audit-service (`SKAI`, `SPI`, `CAE`, same codes as today), because audit-service owns letters and uses the code in `NextLetterNumber`. Served by a new `GET /api/v1/assignment-letters/options` returning `{"audit_teams":[{"value":"SKAI"},{"value":"SPI"},{"value":"CAE"}]}`; labels come from i18n (`assignmentLetter.auditTeams.SKAI` ...), not from the server.
- Validation on create and on update when the value changes: must be in the list, else 422. Existing letters with another value keep it if untouched.
- UI: `USelectMenu` instead of radio buttons; the hard-coded `options.auditTeam` in the store is replaced by the endpoint (with the same three codes as an offline fallback).
- Per-tenant configurable teams would need a table, so they are out of scope (would be flexible-only).

## 6. API summary

master-service:

| Method and path | Mode | Auth | Notes |
|---|---|---|---|
| `GET /tenant-config` | both | authenticated | 5.1 |
| `GET /org-levels` | flexible | authenticated | list with `name_key` |
| `POST /org-levels/changes/preview` | flexible | ADMIN, EXECUTIVE | 9.2; read-only |
| `POST /org-levels/changes/apply` | flexible | ADMIN, EXECUTIVE + `X-Step-Up-Token` | 9.3; all level create/rename/renumber/delete goes through here (no separate per-level write routes) |
| `GET /org-levels/changes` | flexible | ADMIN, EXECUTIVE | change log, newest first, paginated |
| `GET /departments/tree?company_id=` | flexible | authenticated | nested `children[]` |
| `GET /departments` | both | authenticated | flexible filters; sort only by whitelist `department_name, department_code, level, created_at` checked against the GORM schema |
| `POST/PUT/DELETE /departments` | both | authenticated (unchanged) | flexible rules and subtree events |
| `department.*` events | both | n/a | additive fields in flexible |

audit-service: `GET /assignment-letters/options` (both); letter create/update validation (5.3, 5.4); `POST /assignment-letters/unit-usage` `{unit_ids}` -> count of letters referencing them (flexible, used by the preview screen).

auth-service: `POST /auth/verify-password` (9.3).

## 7. Presets, site generator and onboarding

New-tenant order in `onboard-tenant.sh` step 4:

1. `auth migrate up`, `auth seed` (unchanged).
2. `audit migrate up` (flag on: gated audit DDL).
3. `master migrate` (flag on: gated master DDL; `org_levels` empty).
4. New: `master seed-org-levels --preset simple|corporate|bank|custom [--levels "A,B,C"]`. Runs even with `--empty-data` (levels are configuration). Idempotent: no-op if `org_levels` has live rows; never overwrites admin edits.
5. Risk catalogue seed (unchanged), then demo seeders only with `--with-demo-data` (preset-aware master seeder; bank tree).

Flags: `--org-preset=<simple|corporate|bank|custom>` turns the flag on and renders it into the compose file; `--org-levels="..."` sets the names (required for `custom`). Without `--org-preset` the script behaves exactly as today and the tenant is legacy: that is the re-onboarding/recovery path. `registry.tsv` keeps its columns.

Site generator (`frontend/pages/site-generator/index.vue`): step 1 gets a required "Organisation structure" field with the four presets, `simple` preselected, preset names shown through i18n, and an editable level list for `custom`. The review step shows the choice; the printed command always includes `--org-preset=`. There is no legacy option. Honest status: `handleProvisionSite()` is still a UI simulation (only the Resend call is real); the page changes only the printed command and summary, and the real effect happens when an operator runs `onboard-tenant.sh` with those flags.

Tenant isolation is unchanged: levels live in the tenant's own database; the step-up key is derived from that stack's own `jwt.secret`, so a token from one tenant is useless on another.

## 8. Frontend scope and i18n

- `orgStructure` state from `/tenant-config`; helper `levelLabel(level)` = `name_key ? t('orgStructure.levelKeys.' + name_key) : name`. Keys: `directorate` Directorate/Direktorat, `division` Division/Divisi, `department` Department/Departemen, `section` Section/Seksi. Unit names in paths are data and shown as stored.
- `master-options.ts`: `departmentOptions` gains company filter; label is the path in flexible, `name (code)` in legacy.
- `DepartmentForm.vue` / `DepartmentTable.vue`: flexible shows a level select (translated labels), a parent select (same company, lower level), level label and path; legacy unchanged.
- Assignment Letter form and store: 5.3 and 5.4; members picker in flexible only.
- Settings "Organization" tab (flexible only; `pages/settings/index.vue` `validTabs` and sidebar): department tree (add, rename, move, deactivate; same permissions as department CRUD today) and the levels editor (section 9), with the change history.
- All new copy in `frontend/locales/en/common.json` and `frontend/locales/id/common.json`.

## 9. Changing levels: roles, preview, verification, audit log

### 9.1 Roles (D8)

- Backend: `RequireRoles("ADMIN", "EXECUTIVE")` on `/org-levels/changes/*`. Fix `RequireRoles` in `master-service/pkg/middleware/auth.go` to compare case-insensitively (`strings.EqualFold`), matching auth-service; it is unused today, so the change affects no other route. No aliases (`CAE`, `CHIEF_AUDIT_EXECUTIVE`) are accepted, per the decision.
- Frontend: new strict helper `canManageOrgLevels` in `useRbac.ts` that checks the user's actual role names (normalised to lower case, exactly `admin` or `executive`), ignores the role-preview `activeRoleView`, and does not use `matchesRole`/`isExecutive` (which also match viewer/board keywords). It only hides controls; the backend is authoritative.

### 9.2 Preview (dry run)

`POST /org-levels/changes/preview` body: the full desired level set `{levels:[{id?, level, name, name_key?}], delete_ids:[]}`. Response:

- `base_version` (hash of the current live `org_levels` rows) and `diff` (added, renamed, renumbered, deleted).
- `impact`: departments whose `level` or level name changes (count + up to 50 ids), employees in those departments (count), and the ids the frontend passes to audit-service `unit-usage` for the letter count. The preview states that letters are not modified (they keep their snapshot level name).
- `violations[]` and `can_apply`. Violations: more than 6 levels, number outside 1-999, duplicate number or name, deleting a level still used by a live department, and any existing parent/child pair where the new numbers would give `child.level <= parent.level` (departments move with their level, so this is exactly a renumber that reorders levels that are in use).

### 9.3 Verification and apply (D9)

1. User reviews the preview and enters their password.
2. Frontend calls auth-service `POST /auth/verify-password` `{password, purpose: "org_levels.change"}` (new, protected route). auth-service checks the password hash of the authenticated user; failures are rate-limited (5 per 15 minutes per user, Redis) and emitted as a security event; purpose is whitelisted. On success it returns `{step_up_token, expires_in: 300}`: a JWT with `sub`, `purpose`, `jti`, `iat`, `exp` (5 minutes), signed with a key derived from the stack secret (`HMAC-SHA256(jwt.secret, "auditsphere-step-up-v1")`) so it can never be accepted as an access token by any service. Password only (MFA step-up is not required by the decision).
3. Frontend calls `POST /org-levels/changes/apply` with the same desired set, `base_version`, and header `X-Step-Up-Token`.
4. master `RequireStepUp("org_levels.change")` verifies signature (derived key), purpose, expiry, `sub` equals the access token's user, and consumes `jti` once (`SET stepup:<jti> NX EX 300` in the tenant's Redis; Redis unavailable -> 503).
5. master recomputes the preview server-side (never trusts the client's), rejects 409 if `base_version` changed and 422 if there are violations, then in ONE transaction: moves changed level numbers to a temporary offset (+1000) and then to final values (unique indexes cannot be deferred), updates `departments.level` for moved levels, applies renames (user rename clears `name_key`), soft-deletes removed levels, inserts the `org_change_log` row, and enqueues `department.updated` for every department whose level or level name changed through the existing outbox. Commit or nothing.

### 9.4 Audit log

No audit-trail facility exists, so master gets `org_change_log` (gated DDL, flexible only): `id`, `action` (`org_levels.apply`), `actor_user_id`, `actor_username`, `actor_roles`, `ip`, `user_agent`, `base_version`, `before` jsonb, `after` jsonb, `impact` jsonb, `created_at`. Append-only: no update or delete route. Shown in the Settings tab via `GET /org-levels/changes`. auth-service additionally emits a `security.step_up` event for every verify attempt.

## 10. Out of scope: existing tenants

No backfill, no legacy-text linking, no index swap, no `org_levels` on any existing tenant. Existing tenants keep flat departments, the global unique name and free-text `working_unit` (now chosen from a department dropdown, with old text preserved).

Future, optional, NOT part of this work and requiring explicit user approval per tenant: a "convert existing tenant" runbook (duplicate pre-check per company and parent, backup, migrate with the flag on, backfill `org_levels` from distinct `departments.level` renumbered to 10/20/..., optional dry-run linking of `working_unit` text). It needs its own rollback section: the index swap is not reversible once same-name units exist under different parents.

## 11. Risks

1. Shared build: any persisted field added to `models.Department`, `models.AssignmentLetter` or `models.AnnualSnapshotRef`, or any new model added to AutoMigrate, silently alters legacy tenants and the prod stack. The schema-diff check is mandatory in CI for master and audit.
2. Removing the `uniqueIndex` tags moves legacy index creation into an explicit step; a from-scratch legacy DB must still end with both indexes.
3. Flag drift (variable removed from a flexible tenant, or hand-added to a legacy one). Hand-adding it to a legacy tenant would run the index swap on its next migrate; the runbook must forbid it.
4. Legacy letter saves that change the unit now depend on master-service being reachable (503 otherwise).
5. Department CRUD (not levels) remains open to any authenticated user, as today; only level changes are restricted by D8.
6. Subtree and renumber republish can emit many events; bounded by tree size and depth 6.
7. Kong: `/tenant-config`, `/org-levels`, `/departments/tree`, `/assignment-letters/options`, `/assignment-letters/unit-usage`, `/auth/verify-password` must be routable in every Kong config.
8. `backend/audit-service/routes/handler.go` and `controllers/assignment_letter/` have uncommitted work on `staging`; phase 3 must start after that lands or be coordinated.
9. New tenants still ship the default `admin` / `password123` account until the provisioning plan replaces it; the step-up check protects level changes but not that exposure.

## 12. Open question

1. `department_code` uniqueness on new tenants. Default in this spec: globally unique among live rows (`uq_departments_code_live`, partial on `deleted_at`). Reason: codes appear in labels and cross-company lists, and global keeps them unambiguous; the partial index only lets deleted codes be reused. Alternative: unique per company (`(company_id, department_code)`), useful only if subsidiaries reuse the same code scheme. Deciding later is cheap: it only changes gated DDL step 6 for tenants created after the decision.

## 13. Implementation phases and files

Phase 1 can start immediately; it has no dependency. New files are marked (new).

| Phase | Agent | Depends on | Files |
|---|---|---|---|
| 1. master-service core | auditsphere-backend | none | `backend/master-service/pkg/config/config.go` (bind `ORG_STRUCTURE_MODE`); `pkg/orgmode/orgmode.go` (new; resolve effective mode); `cmd/migrate.go` (remove name/code `uniqueIndex` tags from `departmentMigration`, call `migrateOrgStructure` and the legacy index step; `fresh` also drops the org tables in flexible mode); `pkg/database/orgstructure/migrate.go` (new; gated DDL 3.3 steps 1-6 and legacy step 3.4); `models/department.go` (remove tags, add `gorm:"-"` response fields); `models/department_org.go`, `models/org_level.go`, `models/org_change_log.go` (new, not auto-migrated); `repositories/department/department.go` (scoped name check, tree, subtree); `repositories/org_level/org_level.go` (new); `services/department/department.go` (rules, subtree events); `services/org_level/org_level.go` (new; preview, apply, two-phase renumber); `services/tenant_config/tenant_config.go` (new); `services/registry.go`; `controllers/department/department.go`; `controllers/org_level/org_level.go`, `controllers/tenant_config/tenant_config.go` (new); `controllers/registry.go`; `routes/routes.go` (new routes, `RequireRoles("ADMIN","EXECUTIVE")`, `RequireStepUp`); `pkg/middleware/auth.go` (case-insensitive `RequireRoles`); `pkg/middleware/step_up.go` (new; derived-key verify, Redis one-time `jti`); `pkg/kafka/domain_publisher.go` (additive fields, subtree enqueue); `pkg/errors/codes.go`, `pkg/errors/constraints.go`; `cmd/seed_org_levels.go` (new; `seed-org-levels` registered on `rootCmd`); `pkg/database/seeders/seeder.go` (preset-aware levels, bank tree); tests next to each file, including a legacy no-op migrate test |
| 2. auth-service step-up | auditsphere-backend-2 | none (contract in 9.3) | `backend/auth-service/models/dto.go` (`VerifyPasswordRequest/Response`); `services/auth/auth_service.go` (verify, rate limit, derived-key token); `controllers/auth/auth_controller.go` (`VerifyPassword`); `routes/registry.go` (`protected.POST("/verify-password")`); `pkg/kafka/events.go` (`security.step_up`); tests |
| 3. audit-service | auditsphere-backend-2 | field contract from 1; after the uncommitted `routes/handler.go` / `controllers/assignment_letter/` work lands | `backend/audit-service/pkg/config/config.go` (flag); `cmd/migrate.go` (gated DDL 3.3 steps 7-8); `models/assignment-letter.go` (`gorm:"-"` snapshot fields, `LetterMember` employee fields, audit team constant); `models/assignment_letter_org.go`, `models/annual_snapshot_ref_org.go` (new); `models/working-paper.go` (`TeamMember` fields, `syncFromAssignmentLetter`); `routes/handler.go` (letter create/update validation 5.3/5.4, `options`, `unit-usage`); `pkg/masterclient/client.go` (`GetDepartment(id)`); `services/annual_plan_sync/sync.go`, `handler.go` (extra ref columns, `department.deleted` sets `deleted_at`, flexible only); `cmd/backfill_annual_snapshots.go` (fill new ref columns in flexible); tests including legacy no-op migrate |
| 4. Tenant scripts and Kong | auditsphere-devops | images from 1-3 | `scripts/templates/docker-compose.tenant.yml.tpl` (`ORG_STRUCTURE_MODE` placeholder on master and audit); `scripts/generate-tenant-stack.sh` (`--org-preset` renders the flag on first generation only); `scripts/onboard-tenant.sh` (flags, registry guard, `seed-org-levels` call, summary line); `scripts/templates/kong-tenant.yml.tpl`, `backend/kong-gateway/kong/prod/kong.yml`, `backend/kong-gateway/kong/dev/kong-dev.yml` (new prefixes, 11.7); confirm `backend/deploy-prod.sh` override never sets the flag |
| 5. Frontend | auditsphere-frontend | API contracts (mockable) | `frontend/types/master.ts` (`OrgLevel`, department fields), `frontend/types/audit.ts` (letter fields, members); `frontend/composables/useDepartmentApi.ts`; `frontend/composables/useOrgStructureApi.ts` (new; tenant-config, levels, preview/apply, verify-password); `frontend/composables/useRbac.ts` (`canManageOrgLevels`); `frontend/stores/master-options.ts`; `frontend/stores/org-structure.ts` (new); `frontend/stores/assignment-letter.ts`; `frontend/components/assignment-letter/AssignmentLetterForm.vue`; `frontend/components/master/DepartmentForm.vue`, `DepartmentTable.vue`; `frontend/components/settings/organization.vue` (new; tree, levels editor, preview + password modal, history); `frontend/pages/settings/index.vue`; `frontend/pages/site-generator/index.vue` |
| 6. Translations | auditsphere-software-translation | key list from 5 | `frontend/locales/en/common.json`, `frontend/locales/id/common.json` (level keys, presets, auditee unit, audit teams, preview/verification copy, errors, Settings tab, site-generator field); parity check |
| 7. QA | auditsphere-qa | 1-6 | runs section 14 |
| 8. Later, optional | auditsphere-backend | 1, 3 | PKAT path display and ambiguity guard in `ApplyDepartment` (uses the gated ref `path`) |

Parallel tracks: 1 and 2 start now and are independent; 3 can start against the contract but merges after the pending `staging` work; 5 runs against mocks in parallel with 3 and 4; 6 runs alongside 5; 4 waits for images.

## 14. Verification checklist

1. Zero schema diff on legacy: restore copies of a legacy tenant's master and audit databases and of the prod `rb_audit` ones; `pg_dump --schema-only`; run the new `master migrate` and `audit migrate up` exactly as `deploy-prod.sh` and `onboard-tenant.sh` do (flag unset); dump again. Diff must be empty, row counts of `departments`, `assignment_letters`, `annual_snapshot_refs` unchanged; run twice. `idx_departments_department_name` and `idx_departments_department_code` still present.
2. From-scratch legacy DB (flag unset, empty DB): both legacy unique indexes exist after migrate.
3. Legacy behaviour: `/departments` responses have no new keys; duplicate name returns 409; `parent_department_id` in a body returns 422; `/org-levels` returns 404; `/tenant-config` says `legacy`. Letter: unit dropdown lists master departments; saving an unknown, inactive or other-company department id is rejected; the saved `working_unit` equals the master name; an old letter with unmatched text saves unchanged; audit team is a dropdown and an unknown code is rejected.
4. New tenant: onboard a throwaway slug with `--org-preset=simple --empty-data`; `org_levels` = (10 Departemen, `department`); new column and indexes present; old global name index absent; `/tenant-config` says `flexible`. Re-run onboarding: nothing changes.
5. Guard: `onboard-tenant.sh` on an existing slug with `--org-preset` refuses; without it the schema is unchanged.
6. Bank demo: `--org-preset=bank --with-demo-data`; the tree of 4.2.1 exists, including the level skip and the two `Departemen Pajak` units; no `Level 1` rows; paths are correct in `/departments/tree`.
7. Rules: same name under one parent returns 409; re-parent under a descendant rejected; delete with children 409; rename a Divisi and confirm `department.updated` for every descendant and updated ref paths in audit-service.
8. Level changes: AUDITOR gets 403 on preview/apply and sees no controls; ADMIN and EXECUTIVE can preview; apply without step-up token 401; wrong password rate-limited after 5 attempts; a used or expired step-up token rejected; a step-up token sent as `Authorization` on any service rejected; renumber 10/20/30/40 to 100/200/300/400 succeeds and moves departments; a renumber that inverts two used levels is reported as a violation and rejected; concurrent change returns 409; `org_change_log` row written with before/after/impact; the whole apply rolls back if any step fails.
9. Letter (flexible): unit picker shows paths; unknown/inactive/wrong-company `auditeeUnitId` rejected; stored name/path come from the server; renaming the unit leaves the letter unchanged; members carry `employeeId`/`employeeCode`, and working papers copy them.
10. Bilingual: switch EN/ID; preset level names translate, a renamed level shows as typed.
11. Kong: every new route reachable through the new tenant's Kong and the prod Kong.
