# Office structure plan: company offices with configurable levels (replaces "Location as branch")

Status: approved for implementation, 2026-11-24. Related: [org-structure-model-plan.md](org-structure-model-plan.md) (same idea for departments), [office-structure-frontend-plan.md](office-structure-frontend-plan.md) (frontend detail), Auditable Entities multi-level + master link (to be written after this ships).

## 0. Decisions

| # | Decision |
|---|---|
| D1 | The `locations` table stays. Risk Profile, Risk Assessment, RCM, IC Assessment and Employee keep their `location_id`; none of them change. |
| D2 | A location becomes an **office**. It has one **owner company** (`company_id`), which places it in that company's office tree, and may be **shared with other companies** (`location_companies`), e.g. a building used by the holding and two subsidiaries. Sharing does not place the office in the other companies' trees; it lists it under "Kantor bersama". |
| D3 | Offices form a **tree** per owner company (`parent_location_id`), so a branch can sit under a regional office. |
| D4 | Level names are **configured per company** (`office_levels.company_id`). A company without its own scheme **inherits** the scheme of its nearest ancestor company that has one; the tenant default (`company_id IS NULL`) is the last fallback. There is no separate Region master: "region" is whatever level a company names so. |
| D5 | The **function** of an office (head office, R&D, operational, warehouse, data center, other) is separate from its **level**. |
| D6 | `company_type = branch` is retired. Existing branch companies are **converted automatically** into offices by the backfill (section 4). |
| D7 | Employees stay linked to an office through `work_location_id`. Employees are PIC/auditee of auditable entities, not entities. |
| D8 | Applies to every tenant. All schema changes are additive and nullable; no mode flag. |

## 1. Current state (verified)

- `locations` (master-service): name, address, city, province, postal code, country, lat/long, is_active. Flat, no company, no code, no parent. Routes `GET/POST/PUT/DELETE /locations` (`routes/routes.go:116`), writers = `masterDataWriters`.
- Used as the **branch** by `risk_profile.location_id`, `risk_assessments.location_id`, `risk_control_matrix.location_id` (risk-service, `services/risk_location.go`, `services/rcm_branch.go`), `ic_assessments.location_id` (audit-service), `employees.work_location_id`, `companies.location_id`; frontend heat map, RCM, ICA, working paper, mapped data importer.
- `Company`: `parent_id`, `company_type` holding / subsidiary / branch. No feature reads `branch`.
- `org_levels` is gated behind `ORG_STRUCTURE_MODE=flexible` and uses a preview/apply flow with step-up. Office levels are deliberately simpler: plain CRUD, all tenants.

## 2. Schema (master-service)

### 2.1 `office_levels` (new, AutoMigrate)

| Column | Type | Notes |
|---|---|---|
| id | uuid pk | |
| company_id | uuid null, fk companies | NULL = tenant default scheme. |
| level | int not null | Gapped: 10, 20, 30. Lower = higher in the tree. Unique per (company_id, level) among non-deleted rows. |
| name | varchar(100) not null | As typed, e.g. "Kantor Wilayah". |
| name_key | varchar(50) null | `head_office`, `regional_office`, `area_office`, `branch`, `sub_branch`, `unit`. NULL = show `name`. |
| is_active | bool default true | |
| created_at, updated_at, deleted_at | | |

Seed (idempotent, tenant default only): 10 Kantor Pusat `head_office`, 20 Kantor Wilayah `regional_office`, 30 Kantor Cabang `branch`.

**Effective scheme** of company C: the active levels of the first of C, C.parent, C.parent.parent, ... that has at least one non-deleted level row; else the tenant default.

A level cannot be deleted while an office of a company using that scheme sits on it (409 `OFFICE_LEVEL_IN_USE`); it can be deactivated.

### 2.2 `locations` (additive, nullable)

| Column | Type | Notes |
|---|---|---|
| company_id | uuid null, fk companies | Owner. NULL only until backfilled; required on create/update from the API. |
| parent_location_id | uuid null, fk locations | Same owner company. |
| level | int null | A level of the owner company's effective scheme. |
| office_code | varchar(30) null | Unique per owner company when set (partial unique index). |
| office_function | varchar(30) not null default 'operational' | `head_office`, `rnd`, `operational`, `warehouse`, `data_center`, `other`. |
| head_employee_id | uuid null | Kepala kantor; must be an existing employee. |

### 2.3 `location_companies` (new)

| Column | Type | Notes |
|---|---|---|
| location_id | uuid fk locations | pk part |
| company_id | uuid fk companies | pk part; never the owner company |
| created_at | | |

### 2.4 Rules (service layer, 422 with `code`)

| Code | Rule |
|---|---|
| `OFFICE_COMPANY_REQUIRED` | `company_id` missing on create/update. |
| `OFFICE_PARENT_OTHER_COMPANY` | Parent has a different owner company. |
| `OFFICE_PARENT_LEVEL` | Parent's level must be lower than the child's (levels may be skipped). Both null is allowed. |
| `OFFICE_CYCLE` | Parent is the office itself or one of its descendants. |
| `OFFICE_DEPTH` | Depth above 8. |
| `OFFICE_LEVEL_UNKNOWN` | `level` is not an active level of the owner's effective scheme. |
| `OFFICE_HEAD_OFFICE_EXISTS` | A second active `head_office` office for the same owner company. |
| `OFFICE_CODE_TAKEN` | `office_code` already used in the owner company. |
| `OFFICE_HAS_CHILDREN` | Delete while it has child offices. |

Changing the owner company of an office with children is refused (`OFFICE_HAS_CHILDREN`). Existing delete behaviour for referenced locations is kept.

### 2.5 `companies`

- `location_id` stays readable for one release, marked deprecated in JSON docs; the head office is the owner's office with `office_function = head_office`. On company create/update, if `location_id` is sent, keep accepting it (no behaviour change) — the frontend stops sending it.
- `company_type = branch` refused on create/update (422 `COMPANY_TYPE_RETIRED`) after the backfill.

## 3. API contract (master-service, behind Kong as today)

All responses use the existing envelope `{ "success": bool, "data": ..., "message"?, "code"? }`.

### 3.1 Office object (`GET /locations`, `GET /locations/:id`)

Existing fields unchanged, plus:

```json
{
  "id": "uuid", "name": "Kanwil Jawa Timur", "address": "...", "city": "Surabaya",
  "province": "Jawa Timur", "postal_code": "60111", "country": "Indonesia",
  "latitude": -7.25, "longitude": 112.75, "is_active": true,
  "company_id": "uuid|null",
  "company_name": "PT Bank X",
  "parent_location_id": "uuid|null",
  "level": 20,
  "level_name": "Kantor Wilayah",
  "level_name_key": "regional_office|null",
  "office_code": "KW-JATIM|null",
  "office_function": "operational",
  "head_employee_id": "uuid|null",
  "head_employee_name": "string|null",
  "shared_company_ids": ["uuid"],
  "path": [{ "id": "uuid", "name": "Kantor Pusat", "level": 10 }],
  "children_count": 12
}
```

`path` runs from the root office down to the parent (excludes self).

### 3.2 Endpoints

| Method | Path | Notes |
|---|---|---|
| GET | `/locations` | Query (all optional): `company_id` (offices owned by **or shared with** that company; each item gets `"shared": true` when not owned), `level`, `office_function`, `parent_id` (`root` = top level), `q`, `is_active`, `tree=true` (nested `children` arrays, roots = no parent). No params = today's flat list plus new fields. |
| GET | `/locations/:id` | |
| GET | `/locations/:id/descendants` | Flat list of all descendant offices, for "include sub-offices" filters. |
| POST / PUT | `/locations`, `/locations/:id` | Body = office object writable fields + `shared_company_ids` (replaces the set). |
| DELETE | `/locations/:id` | |
| GET | `/companies/:id/office-levels` | `{ "source_company_id": "uuid|null", "inherited": bool, "levels": [office_level] }` — the effective scheme. |
| PUT | `/companies/:id/office-levels` | Body `{ "levels": [{ "id"?, "level", "name", "name_key"?, "is_active" }] }`. Replaces the company's own scheme (creates it if it was inheriting). |
| DELETE | `/companies/:id/office-levels` | Drop the company's own scheme → inherits again. 409 `OFFICE_LEVEL_IN_USE` if any office of a company using it sits on a level the inherited scheme lacks. |
| GET / PUT | `/office-levels/default` | Tenant default scheme, same body as above. |

Writers: `masterDataWriters` (same as `/locations`). Readers: any authenticated user.

### 3.3 Kafka (phase 3)

Topic `location.events`, events `location.created|updated|deleted`, outbox + relay as `pkg/kafka/domain_publisher.go`. Data = office object of 3.1 without `children_count`.

## 4. Backfill: `master backfill-offices`

Dry run unless `--apply`. Idempotent; touches only rows with `company_id IS NULL` and companies with `company_type = 'branch'`. Prints a report per step.

1. A location referenced by `companies.location_id` → owner = that company, `office_function = head_office` (only if the company has none yet).
2. Otherwise, owner = the company of most active employees whose `work_location_id` is that location.
3. Otherwise, if the tenant has exactly one company → that company.
4. Otherwise left NULL and listed for manual assignment in the UI (list filter "Tanpa company").
5. **Convert branch companies** (D6): for each `company_type = branch` company B with parent P:
   - if B has `location_id` → that location becomes the office (owner = P, keep its address), else create an office with B's name, owner P, `office_code = B.company_code`;
   - `office_function = operational`, level = the lowest-level `branch`-keyed level of P's effective scheme, else null;
   - every employee with `company_id = B` → `company_id = P`, `work_location_id` = the office if it was empty;
   - other references to B (departments, business units, vision-mission-goals) are reported, not moved; if none, B is soft-deleted, otherwise B is deactivated (`is_active = false`).
   - B without a parent: reported, not converted.

Parent offices and levels of non-branch locations are not guessed; admins arrange them in the UI.

## 5. Frontend

See [office-structure-frontend-plan.md](office-structure-frontend-plan.md).

## 6. Consumers

- Phases 1–2: no change in risk-service or audit-service.
- Phase 3 (optional): "include sub-offices" toggle on branch filters (Risk Profile, RCM, heat map, ICA) using `/locations/:id/descendants`.
- Auditable Entities: an office becomes an entity source (`entity_type = office`, `source_id = location id`) under its owner company. Separate plan.

## 7. Phases

| Phase | Scope | Owner | Est. |
|---|---|---|---|
| 1 | Schema 2.1–2.3, rules 2.4, API 3.1–3.2, seed, backfill command with tests | backend | 2–3 d |
| 2 | Frontend plan items (Company → Kantor tab, levels settings, office list tree, pickers with path) | frontend | 3–4 d |
| 3 | `location.events`, descendants filters in consumers | backend + frontend | 2–3 d |
| 4 | Retire `companies.location_id` writes and `company_type = branch` in UI | both | 0.5 d |

## 8. Risks

- Wrong backfill guess (step 2): dry run reviewed per tenant before `--apply`.
- Branch conversion moves employees to the parent company: shown in the dry run; RBAC scoped by company should be checked by QA on a tenant copy.
- Re-parenting changes roll-ups for past risk data; reports read the current tree.
- Inherited schemes: a holding edits its levels and every inheriting subsidiary changes; the UI warns with the number of companies affected.
