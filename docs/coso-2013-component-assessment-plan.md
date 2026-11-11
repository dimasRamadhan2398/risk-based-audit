# COSO 2013 internal control component assessment: plan

Status: DRAFT FOR DECISION (revision 1, 2026-11-04). Plan only: no application code, migration or tenant has been changed. Verified against the repo on branch `staging`, HEAD `b8ac7ba`, with uncommitted work in the tree (section 2.3). Re-check section 2 before relying on it. Open questions are in section 13; the defaults written in this spec apply until the user decides otherwise.

Hard constraint: existing databases (the prod `rb_audit` stack and tenant `bai`) get only additive DDL: new tables and new nullable columns. No column is dropped, renamed or retyped, no existing row is rewritten by a migration, and the reference data seed is insert-if-missing.

## 0. Decisions in this spec

| # | Decision | Section |
|---|---|---|
| D1 | The 5 RCM "dimensions" (design, operating, coverage, timeliness, automation) stay as control-level effectiveness criteria. They are relabelled in the UI as "Control effectiveness criteria" and are no longer called COSO components anywhere. | 3, 9.3 |
| D2 | The COSO catalogue (5 components, 17 principles) and the per-control tags live in **risk-service**, next to the RCM. The department/period assessment lives in **audit-service**. | 3 |
| D3 | The department is the primary scope (master `departments.id`). Branch (master Location) is an optional sub-scope, derived from the linked risk exactly as the RCM page does today. | 6 |
| D4 | The system proposes a 0-100 % score per component from existing data, mapped to the RCM 1-5 bands. Missing data gives no score (NULL, shown as "Insufficient data"), never 0. | 5 |
| D5 | Final values are the auditor's. A final rating that differs from the proposal, or any final rating where there was no proposal, needs a written justification. Every proposal, change and transition is kept in an append-only event log, with the evidence snapshot used. | 4.2, 7 |
| D6 | Workflow `DRAFT -> SUBMITTED -> APPROVED`, `SUBMITTED -> DRAFT` (returned), `APPROVED -> REOPENED -> SUBMITTED`. Approver must not be the submitter. Approved assessments and locked periods are read-only. | 7 |
| D7 | Overall conclusion: `EFFECTIVE` only if all 5 components are present and functioning, no principle has a major deficiency, and the approver confirms the components operate together. Any major deficiency forces `NOT_EFFECTIVE`. | 4.3 |
| D8 | One new Kong prefix: `/api/v1/ic-assessments` (audit-service). Every new risk-service route sits under the existing `/api/v1/rcm` prefix. | 8.4 |
| D9 | COSO ERM 2017 tagging of Corporate Risk Profile risks is phase 2, outline only. | 11 |

## 1. Summary

Today the RCM page and both dashboards present the 5 RCM effectiveness dimensions as if they were the 5 COSO 2013 components, and the dashboards fall back to made-up numbers (78, 80, 75, 82, 76, 74) when there is no data. This plan:

1. Adds a COSO 2013 catalogue and lets each RCM row be tagged with a primary component and the principles (1-17) it supports.
2. Adds a per-department (optionally per-branch), per-period (year, optionally quarter) assessment: a rating per principle, a rolled-up rating per component, a component conclusion and an overall conclusion.
3. Proposes a score per component from data AuditSphere already holds, keeps the auditor's final value, the override reason, who and when, and the evidence used.
4. Replaces the misleading dashboard cards with approved assessment results.

## 2. Current state (verified)

### 2.1 What exists

| Area | Fact | Reference |
|---|---|---|
| RCM model | Table `risk_control_matrix`. `risk_id` (soft link to `risk_register.id`), `department varchar(100) not null` (free text), `year int`, 5 weight/rating pairs, `total_weighted_score` (0-100: each rating / 5 x weight), `inherent_risk`, `residual_risk`. Comment says "COSO 2013 5 Dimensions". | `backend/risk-service/models/risk_control_matrix.go` |
| RCM API | `GET/POST /rcm`, `PUT/DELETE /rcm/:id`, `GET /rcm/summary`. Filters `year`, `department` (text). Missing ratings default to 3 server-side. Empty summary returns a fake 60 % "Weak". | `backend/risk-service/controllers/rcm_controller.go` lines 100-163, 362-381 |
| RCM RBAC | `Authenticate` + `DenyViewerWrites` only. Any non-viewer role (including AUDITEE) can write RCM rows. | `backend/risk-service/routes/routes.go` lines 58-64, 120-124 |
| RCM "department" | Filled by the form from the selected risk's **branch or category**, default `Head Office`; the department filter is a hard-coded list of branch names. It is not a master department. | `frontend/pages/risk-profile/risk-control-matrix/index.vue` lines 938-945, 1116; `frontend/utils/rcmForm.ts` |
| RCM branch | Derived on the client: the linked risk's assessment location for the RCM year, else the risk's own location; unlinked -> "Unassigned". | `frontend/utils/rcmBranch.ts` (untracked), `stores/rcm.ts` |
| Risk department | `risk_profile.department_id uuid not null` exists, but the CRP API writes `uuid.Nil`, and old rows carry branch sentinels in it. **Risks have no usable department.** | `backend/risk-service/services/risk_service.go` line 247, `services/risk_location.go` lines 22-35 |
| Quarterly assessment | `risk_assessment` per register and year: `impact_q1..q4`, `likelihood_q1..q4` (0 = not filled), `risk_level_q1..q4` text. | `backend/risk-service/models/risk_assessment.go` |
| Risk appetite rule | `risk_appetite` holds statements and a `threshold_limit` with no link to risks. "Within appetite" is a frontend rule: level from the 5x5 matrix is LOW or LOW_MODERATE. The matrix exists only in the frontend. | `frontend/pages/risk-appetite/index.vue` lines 507-510; `frontend/stores/risk-profile.ts` lines 407-417 |
| Mitigation | `risk_mitigation.risk_id` = `risk_register.id`. | `backend/risk-service/models/risk_mitigation.go` |
| Audit charter | `audit_charters.is_active`; `GET /audit-charters/active`. Tenant-wide. | `backend/audit-service/models/audit-charter.go`, `routes/handler.go` line 82 |
| Integrity pact | auth-service `confidentiality_agreements` (`user_id`, `agreement_type` SYSTEM/ROLE/PROJECT, `version`, `is_accepted`, `accepted_at`, `expires_at`). Only per-user `GET /confidentiality/status` and `POST /confidentiality/accept`. **No aggregate endpoint.** | `backend/auth-service/models/confidentiality_agreement.go`, `routes/registry.go` lines 198-205 |
| User to department | `users.employee_id` = `employees.employee_code` (business key) -> `employees.department_id`. `users.department` is free text. `GET /employees` has **no department filter** (only `page`, `page_size`, `search`). | `backend/auth-service/models/user.go`; `backend/master-service/controllers/employee/employee.go` line 51 |
| Department | `departments.pic_id` (nullable); `GET /tenant-config` exists in master (org structure mode). | `backend/master-service/models/department.go` line 27; `routes/routes.go` line 72 |
| LHA | `audit_result_reports.department varchar(100)` (text), `report_date`, `status` (`DRAFT`, `Final`/`Approved`/`Published`), `assignmentLetterId` (letter number or UUID). | `backend/audit-service/models/audit-result-report.go` |
| Letter period | `assignment_letters.finish_period varchar` (string date). | `backend/audit-service/models/assignment-letter.go` |
| ATR | `action_taken_reports`: `audit_result_report_id`, `status` (PLANNED, IN_PROGRESS, PENDING_REVIEW, COMPLETED, CANCELLED), `due_date`, `reviewed_at` (set on review approve, reject and cancel), overdue derived (`ComputeATROverdue`, Asia/Jakarta). No department column (via the LHA). **No timestamp of when the PIC first entered the action plan.** | `backend/audit-service/models/action-taken-report*.go`; `controllers/action_taken_report/handler.go` lines 735-745 |
| Executive summary | Per `tahun` + `quarter`, `status` Draft/Approved/Rejected, `approvedAt`. **No department.** | `backend/audit-service/models/executive-summary.go` |
| QA | master `qa_reports`: `type`, `period` (text), `result` (free text), `status`. About the internal audit function, **no department, no numeric result**. | `backend/master-service/models/quality_assurance.go` |
| Service clients | audit-service has `riskclient`, `masterclient`, `authclient` (forward the caller's JWT) and `servicetoken`; sibling URLs resolve from `KAFKA_SERVICE_NAME`, so no new env is needed. risk-service has only `masterclient` (locations, departments; `DepartmentName(id)`). | `backend/audit-service/pkg/*`; `backend/risk-service/pkg/masterclient` |
| Roles | audit-service canonical roles: ADMIN, CHIEF_AUDIT_EXECUTIVE (EXECUTIVE maps to it), AUDIT_MANAGER, AUDITOR, DEPARTMENT_HEAD, AUDITEE, VIEWER (`CanonicalRole`). Executive summary approve = ADMIN + CAE. Both services refuse VIEWER writes by default (`DenyViewerWrites`). | `backend/audit-service/services/action_taken_report/rbac.go`; `controllers/executive_summary/approve.go` |
| Dashboards | `ExecutiveDashboard.vue` lines 413-427 and `AdminDashboard.vue` lines 365-450 label the RCM dimensions as Control Environment ... Monitoring, with hard-coded fallback scores and Indonesian text outside i18n. | `frontend/components/dashboard/` |
| Migrations | Prod: `./risk up && ./risk serve`, `./audit migrate up && ./audit serve` on every deploy. Tenant stacks run `serve` only; migrate runs at onboarding or by hand. | `backend/docker-compose.prod.yml` line 230; `backend/deploy-prod.sh`; `scripts/templates/docker-compose.tenant.yml.tpl` lines 129, 180 |
| Kong | Four configs: `backend/kong-gateway/kong/dev/kong-dev.yml`, `backend/kong-gateway/kong/prod/kong.yml`, `backend/tenants/bai/kong/kong.yml`, `scripts/templates/kong-tenant.yml.tpl`. `/api/v1/rcm` is in all four. | grep |

### 2.2 Gaps this plan has to close or work around

| Gap | Handling |
|---|---|
| G1 RCM rows have no master department | Add `risk_control_matrix.department_id` (4.1). Legacy rows stay "Unassigned" until edited or backfilled by an explicit command. |
| G2 Risks have no department | A department's risk set = risks linked from its RCM rows (5.2). Proper risk ownership is phase 2. |
| G3 Risk level matrix only in the frontend | Port the 5x5 matrix to risk-service with a parity test against `stores/risk-profile.ts`. |
| G4 No integrity pact aggregate | New auth-service endpoint (8.3) + master employee `department_id` filter. |
| G5 No ATR "response entered at" | Optional additive column `action_taken_reports.action_plan_submitted_at` (4.2.6). Until filled, management-response timeliness uses a coverage proxy. |
| G6 LHA and ATR department is text | Match `audit_result_reports.department` to the master department name (trim, case-insensitive). Business key, not enforced; unmatched counts are shown in the evidence. |
| G7 Executive summary and QA have no department | Shown as entity-level context in the evidence, not scored. |

### 2.3 Uncommitted work this plan touches

`frontend/pages/risk-profile/risk-control-matrix/index.vue`, `frontend/stores/rcm.ts`, `frontend/utils/rcmBranch.ts`, `frontend/utils/rcmForm.ts` (untracked), `frontend/components/shared/SignatureBox.vue` (untracked), `backend/audit-service/cmd/serve.go`, `backend/audit-service/controllers/executive_summary/approve.go`. Frontend tasks F1 and F3 start after that work is committed.

## 3. Ownership

Recommendation: split by what each half needs to be correct.

- **risk-service owns the catalogue and the tags.** The tags must reference RCM rows; in the same database (`rb_audit_risk_<slug>`) they get real foreign keys. Control Activities and Risk Assessment evidence is local there too.
- **audit-service owns the assessment.** It is an internal audit conclusion with an approval, like the executive summary. Monitoring, Information & Communication and most Control Environment evidence is local (charter, LHA, ATR). It already has clients for risk, master and auth that forward the caller's token, and resolves sibling URLs from `KAFKA_SERVICE_NAME`, so no compose change is needed. Putting the assessment in risk-service would need new audit and auth clients and new env vars in four compose files.

Cost of the split: audit-service stores principle and component **codes** (`P01`..`P17`, `CE`..`MA`) as business keys, validated against a Go constant list in audit-service that mirrors the risk-service seed (the 17 COSO 2013 principles are fixed by the standard). No cross-service foreign key; both lists are unit-tested against the table in 4.1.1.

analytics-service: no change in phase 1.

## 4. Data model

### 4.1 risk-service (`rb_audit_risk_<slug>`)

#### 4.1.1 `coso_components` and `coso_principles` (new, reference data)

| Table | Columns |
|---|---|
| `coso_components` | `id` uuid pk; `framework` varchar(20) not null default `COSO_2013`; `code` varchar(10) not null; `sort_order` int; `name_en` varchar(100); `created_at`, `updated_at`. Unique `(framework, code)`. |
| `coso_principles` | `id` uuid pk; `framework` varchar(20) not null default `COSO_2013`; `code` varchar(10) not null; `number` int not null; `component_code` varchar(10) not null; `name_en` varchar(255); `is_active` bool default true; `created_at`, `updated_at`. Unique `(framework, code)`. |

`framework` leaves room for ERM 2017 (section 11) without new tables. Display names come from i18n (`coso.components.*`, `coso.principles.*`); `name_en` is for exports and API readers.

Seed (17 rows):

| Component | Code | Principles |
|---|---|---|
| Control Environment | `CE` | P01 commitment to integrity and ethical values; P02 board exercises oversight; P03 structure, authority and responsibility; P04 commitment to competence; P05 enforces accountability |
| Risk Assessment | `RA` | P06 specifies suitable objectives; P07 identifies and analyses risk; P08 assesses fraud risk; P09 identifies and analyses significant change |
| Control Activities | `CA` | P10 selects and develops control activities; P11 general controls over technology; P12 deploys through policies and procedures |
| Information & Communication | `IC` | P13 uses relevant information; P14 communicates internally; P15 communicates externally |
| Monitoring Activities | `MA` | P16 ongoing and/or separate evaluations; P17 evaluates and communicates deficiencies |

Seeding runs inside `risk up` (not `risk seed`, which `deploy-prod.sh` refuses and which overwrites demo data): `INSERT ... ON CONFLICT (framework, code) DO NOTHING`. It never updates or deletes an existing row, so it is safe on prod, on `bai`, on re-onboarding and on every restart. It is reference data, so it also runs for `--empty-data` tenants.

#### 4.1.2 `risk_control_matrix` (additive columns)

| Column | Type | Notes |
|---|---|---|
| `department_id` | uuid NULL, index | Master department (soft ref, not enforced). Validated on write through `masterclient.DepartmentName`; when set, the server also writes the master name into the existing `department` text. NULL on all legacy rows. |
| `coso_component_code` | varchar(10) NULL | Primary component. Must be the component of at least one tagged principle. |

Nothing else on the table changes; the 5 dimension columns keep their names.

#### 4.1.3 `rcm_coso_principles` (new join table)

`rcm_id` uuid FK -> `risk_control_matrix.id` ON DELETE CASCADE; `principle_id` uuid FK -> `coso_principles.id` ON DELETE RESTRICT; `created_at`. PK `(rcm_id, principle_id)`. Index on `principle_id`.

Implement it as an explicit model, **not** a GORM `many2many` association on `RiskControlMatrix`: `UpdateRCM` uses `db.Save(&existing)`, which would upsert associations. The tag list is a `gorm:"-"` field (`coso_principle_codes []string` in JSON), written in the same transaction as the RCM row (delete the row's tags, insert the new set). RCM uses soft delete, so tags of a soft-deleted row stay but are never read (joins filter on `deleted_at IS NULL`).

### 4.2 audit-service (`rb_audit_audit_<slug>`)

All tables are new; `AutoMigrate` creates them. Every `*_by` is stored as user id (JWT `sub`) plus display name, because users live in auth-service.

#### 4.2.1 `ic_assessments` (header)

| Column | Notes |
|---|---|
| `id` uuid pk | |
| `department_id` uuid not null | Master department, validated at create through `masterclient`. |
| `department_code`, `department_name`, `company_id` | Snapshot at create; never rewritten. |
| `location_id` uuid NULL, `location_key` varchar(36) not null default `''`, `location_name` | Optional branch scope. `location_key` = location id or `''`, so the unique index needs no expression. |
| `year` int not null, `quarter` smallint not null default 0 | 0 = annual. `CHECK (quarter BETWEEN 0 AND 4)`. |
| `period_start`, `period_end`, `as_of_date` date | Calendar periods (fiscal year = calendar year, as `utils/fiscalPeriod.ts`). `as_of_date` = min(period_end, proposal date). |
| `status` varchar(20) | `DRAFT`, `SUBMITTED`, `APPROVED`, `REOPENED`. |
| `version` int not null default 1 | Optimistic lock on every write. |
| `proposal_run` int default 0 | Increments per proposal. |
| `derived_conclusion` varchar(20) | `EFFECTIVE`, `NOT_EFFECTIVE`, `REQUIRES_JUDGMENT`, `INCOMPLETE` (4.3). |
| `final_conclusion` varchar(20) NULL | `EFFECTIVE` / `NOT_EFFECTIVE`. |
| `operate_together` bool NULL, `operate_together_note` text | Approver confirmation (COSO "operating together"). |
| `conclusion_text`, `conclusion_justification` text | Justification required when final differs from derived or derived is `REQUIRES_JUDGMENT`. |
| `created_by_*`, `submitted_by_*`, `submitted_at`, `approved_by_*`, `approved_by_role`, `approved_at`, `approval_note` | |
| `reopen_count` int, `last_reopen_reason` text | |
| `created_at`, `updated_at`, `deleted_at` | |

Unique index `uq_ic_assessments_scope` on `(department_id, location_key, year, quarter)` `WHERE deleted_at IS NULL` (GORM `uniqueIndex` tag with `where:`, same pattern as `assignment_letters`).

#### 4.2.2 `ic_assessment_components` (5 rows per assessment)

`id`; `assessment_id` FK cascade; `component_code`; `proposed_score_pct` decimal(5,2) NULL; `proposed_rating` smallint NULL (1-5); `proposed_status` NULL; `proposed_severity` NULL; `proposal_data_coverage_pct` decimal(5,2) NULL (share of signal weight that had data); `final_rating` smallint NULL; `final_status` NULL (`YES`, `PARTIAL`, `NO`); `final_severity` NULL; `derived_status`, `derived_severity` (from principles, 4.3); `conclusion` text; `override_reason` text; `is_override` bool (stored, recomputed on save); `reviewed_by_*`, `reviewed_at`. Unique `(assessment_id, component_code)`.

#### 4.2.3 `ic_assessment_principles` (17 rows per assessment)

`id`; `assessment_id` FK cascade; `component_code`; `principle_code`; `tagged_control_count` int; `tagged_control_avg_pct` decimal(5,2) NULL (evidence only); `status` (`NOT_ASSESSED`, `YES`, `PARTIAL`, `NO`, `NOT_APPLICABLE`); `severity` (`NONE`, `DEFICIENCY`, `SIGNIFICANT`, `MAJOR`); `rationale` text; `updated_by_*`, `updated_at`. Unique `(assessment_id, principle_code)`.

Validation (server): `YES` -> severity `NONE` or `DEFICIENCY`; `PARTIAL` -> `DEFICIENCY` or `SIGNIFICANT`; `NO` -> `MAJOR` (a principle that is not present and functioning is a major deficiency); `NOT_APPLICABLE` -> rationale required, severity `NONE`; `PARTIAL`, `NO`, `NOT_APPLICABLE` and any severity above `NONE` require a rationale.

#### 4.2.4 `ic_assessment_evidence` (append-only, one row per signal per proposal run)

`id`; `assessment_id`; `proposal_run`; `component_code`; `signal_key` (5.3); `principle_hint` (for example `P07`); `source_service`; `source_ref` (table or endpoint); `state` (`OK`, `NO_DATA`, `NOT_APPLICABLE`, `SOURCE_UNAVAILABLE`); `value_pct` NULL; `numerator`, `denominator` int NULL; `weight` decimal; `scope_note` (for example "department-wide, no branch data"); `detail` jsonb (counts, up to 200 record ids for drill-down, unmatched counts, parse failures); `collected_at`. The current evidence is the rows of the latest `proposal_run`; older runs stay for the audit trail.

#### 4.2.5 `ic_assessment_events` (append-only audit trail) and `ic_assessment_periods`

- `ic_assessment_events`: `id`; `assessment_id` (NULL for period events); `action` (`CREATE`, `PROPOSE`, `UPDATE`, `OVERRIDE`, `SUBMIT`, `RETURN`, `APPROVE`, `REOPEN`, `DELETE`, `PERIOD_LOCK`, `PERIOD_UNLOCK`); `from_status`, `to_status`; `actor_user_id`, `actor_name`, `actor_roles`; `reason`; `before` jsonb, `after` jsonb (changed fields only); `ip`, `user_agent`; `created_at`. No update or delete route; written in the same transaction as the change.
- `ic_assessment_periods`: `id`; `year`; `quarter` (0 = whole year); `status` (`OPEN`, `LOCKED`); `locked_by_*`, `locked_at`, `lock_note`. Unique `(year, quarter)`. No row = open.

#### 4.2.6 Optional additive column (phase 1b)

`action_taken_reports.action_plan_submitted_at timestamptz NULL`: set once, the first time the PIC saves a non-empty action plan. NULL on history; rows without it are excluded from the timeliness signal (never counted as late).

### 4.3 Roll-up rules

Severity order: `NONE < DEFICIENCY < SIGNIFICANT < MAJOR`. `NOT_APPLICABLE` principles are ignored in roll-ups.

Component (derived, recomputed on every save):
- `derived_severity` = highest principle severity.
- `derived_status` = `NO` if any principle is `NO` or `MAJOR`; `YES` if every applicable principle is `YES`; otherwise `PARTIAL`. Any `NOT_ASSESSED` principle -> component incomplete.
- Final: the auditor sets `final_status`, `final_severity` and `final_rating`. `final_severity` may be higher than derived (deficiencies across principles can add up to a major one, as COSO allows) but never lower. `final_status` may not be `YES` when derived is `NO`. Rating consistency: `MAJOR` -> rating <= 2; `YES` -> rating >= 3.
- `is_override` = `final_rating` differs from `proposed_rating`, or `proposed_rating` is NULL. Then `override_reason` is required (at least 20 characters).

Overall:
- `INCOMPLETE` while any principle is `NOT_ASSESSED` or any component has no final values. Submit is refused.
- `NOT_EFFECTIVE` if any component `final_status = NO` or any `MAJOR`.
- `EFFECTIVE` if all five are `YES` with no severity above `DEFICIENCY`.
- `REQUIRES_JUDGMENT` otherwise (some `PARTIAL`/`SIGNIFICANT`, no major).
- `final_conclusion` is set by the auditor before submit. `EFFECTIVE` is refused when there is any `MAJOR` or any component `NO`. Approve requires `operate_together` to be answered; `operate_together = false` forces `NOT_EFFECTIVE`.

## 5. Proposed score

### 5.1 Scale and general rule

- Each signal yields a value 0-100 % with numerator and denominator, or a state without a value (`NO_DATA`, `NOT_APPLICABLE`, `SOURCE_UNAVAILABLE`).
- Component score = sum(weight x value) / sum(weight), over signals with a value.
- **Score is NULL** when the weights with a value add up to less than 50 % of the component's total weight. The UI shows "Insufficient data" with the reasons; it is never shown or averaged as 0.
- Rating bands are the RCM bands (`GetEffectivenessRating`): >= 90 -> 5 Highly effective; >= 80 -> 4 Effective; >= 70 -> 3 Moderately effective; >= 60 -> 2 Weak; < 60 -> 1 Ineffective.
- Suggested status: rating 4-5 -> `YES`, 3 -> `PARTIAL`, 1-2 -> `NO`. Suggested severity: 5 -> `NONE`, 4 -> `DEFICIENCY`, 3 -> `DEFICIENCY`, 1-2 -> `SIGNIFICANT`. **The system never proposes `MAJOR`**; that is a human judgment.
- Proposal is a component-level hint only. Principles are never pre-filled, except that `tagged_control_count` and `tagged_control_avg_pct` are shown next to each principle.
- Weights are constants in audit-service in phase 1 (open question Q10).

### 5.2 Scope used for each signal

- Department D, year Y, quarter Q (0 = annual), as-of date A.
- RCM set C = live RCM rows with `department_id = D` and `year = Y`. RCM is yearly, so a quarterly assessment uses the year's rows (labelled "annual data").
- Risk set R = distinct `risk_id` of rows in C (gap G2).
- Branch scope (if `location_id` set): C and R are narrowed by the RCM branch rule (risk assessment location for the year, else the risk's location, including legacy sentinels via `resolveBranch`), ported to risk-service. Audit-side signals have no branch; they are computed department-wide and marked so.
- LHA set L = `audit_result_reports` with status Final/Approved/Published (case-insensitive), `report_date` in the period, and `department` matching D's master name (trim, case-insensitive). The number of LHAs in the period with no department match is stored in `detail`.
- ATR set T = `action_taken_reports` of reports in L (any report date), excluding `CANCELLED`, with `due_date` inside the period and on or before A.

### 5.3 Formulas per component

| Component | Signal (`signal_key`) | Value | Weight | Source | Principle hint |
|---|---|---|---|---|---|
| CE | `ce.integrity_pact` | Linked users of D's active employees with an accepted, unexpired SYSTEM agreement / linked users of D's active employees. No linked users -> `NO_DATA`. | 60 | master `GET /employees?department_id=` (new filter), auth `GET /confidentiality/acceptance-summary` (new) | P01 |
| CE | `ce.unit_accountability` | 100 if D is active and `pic_id` is set, else 0. | 20 | master `GET /departments/:id` | P03, P05 |
| CE | `ce.audit_charter` | 100 if an active charter exists, else 0. Entity-level. | 20 | audit `audit_charters.is_active` | P02 |
| RA | `ra.quarterly_coverage` | Quarters filled / quarters due over R. Due = quarters of Y that ended on or before A (up to Q if quarterly; if none ended yet, the current quarter). Filled = `impact_qN > 0 AND likelihood_qN > 0` in `risk_assessment` for Y. R empty -> `NO_DATA`. | 50 | risk `risk_assessment` | P07, P09 |
| RA | `ra.above_appetite_mitigated` | Risks in R above appetite with at least one live `risk_mitigation` row / risks in R above appetite. Above appetite = level from the latest filled quarter (else inherent likelihood x impact) is not LOW or LOW_MODERATE, using the ported matrix (G3). None above appetite -> `NOT_APPLICABLE`. | 50 | risk `risk_assessment`, `risk_register`, `risk_mitigation` | P07 |
| CA | `ca.rcm_effectiveness` | Mean `total_weighted_score` over C. C empty -> `NO_DATA`. Also returns the mean per tagged principle (P10-P17) for the principle rows. | 100 | risk `risk_control_matrix`, `rcm_coso_principles` | P10-P12 |
| IC | `ic.lha_timeliness` | LHAs in L whose `report_date` is at most 30 days after the letter's `finish_period` / LHAs in L with both dates parseable. Unparseable dates are counted in `detail`, not in the denominator. L empty -> `NO_DATA`. | 50 | audit `audit_result_reports`, `assignment_letters` | P14 |
| IC | `ic.management_response` | Phase 1: ATRs of L not in `PLANNED` / all non-cancelled ATRs of L (response coverage). Phase 1b, when 4.2.6 exists: ATRs with `action_plan_submitted_at` at most 14 days after `created_at` / ATRs with `action_plan_submitted_at` or still `PLANNED` after 14 days. | 50 | audit `action_taken_reports` | P14, P17 |
| MA | `ma.atr_completed_on_time` | ATRs in T with `status = COMPLETED` and `reviewed_at` on or before `due_date` end of day / ATRs in T. T empty -> `NO_DATA`. | 50 | audit `action_taken_reports` | P16, P17 |
| MA | `ma.atr_not_overdue` | 100 - (ATRs of L overdue at A per `ComputeATROverdue` / open ATRs of L with a due date). No open ATRs with a due date -> `NOT_APPLICABLE`. | 50 | audit `action_taken_reports` | P17 |

Context evidence, shown but not scored: executive summary for (Y, Q) approved or not (`executive_summaries`, entity-level, IC); latest QA reports with `type`, `period`, `result` text (`qa_reports`, IA function, MA); RCM row count without department (data quality); RCM rows whose dimension ratings were never set (default 3, shown as a caveat on `ca.rcm_effectiveness`).

A source that cannot be reached gives `SOURCE_UNAVAILABLE` for its signals; the proposal is saved and marked partial, and can be re-run. Proposing is allowed only in `DRAFT` or `REOPENED`.

## 6. Department and branch filters

- **Department** is the key everywhere: master `departments.id`, picked from `GET /departments` (flat on legacy tenants; path label on flexible tenants, per `docs/org-structure-model-plan.md`). Any level of a flexible tree can be assessed; roll-up of child units into a parent is phase 2.
- **Branch** is optional and narrower: an assessment with `location_id` uses only branch-attributable risk-side data (5.2). The list page filters by branch the same way the RCM page does; "Unassigned" is not a valid assessment scope.
- **RCM page**: the hard-coded branch list in the department filter is replaced by master departments (`department_id`), plus "Unassigned" for rows without one. The branch filter stays as it is (`utils/rcmBranch.ts`). The existing `department` text filter in the API stays for old clients.
- **Legacy RCM rows**: `department_id` stays NULL. An explicit command `./risk backfill-rcm-department [--apply]` (dry run by default) matches the `department` text to live master department names (trim, case-insensitive, unique match only) and reports unmatched rows. Never part of `risk up`. Today most values are branch names or categories, so expect most rows to stay unassigned.

## 7. Workflow, RBAC and locking

### 7.1 States

| From | Action | To | Who | Rules |
|---|---|---|---|---|
| (none) | create | `DRAFT` | Assessors | Period open; scope unique; department exists and is active; creates 5 component + 17 principle rows. |
| `DRAFT`, `REOPENED` | propose | same | Assessors | New `proposal_run`, evidence rows, `PROPOSE` event. Final values are untouched. |
| `DRAFT`, `REOPENED` | update | same | Assessors | `version` must match (else 409). `UPDATE`/`OVERRIDE` events with before/after. |
| `DRAFT`, `REOPENED` | submit | `SUBMITTED` | Assessors | Not `INCOMPLETE`; all override reasons and rationales present; `final_conclusion` set and allowed (4.3); at least one proposal run. |
| `SUBMITTED` | return | `DRAFT` | Approvers | Reason required. |
| `SUBMITTED` | approve | `APPROVED` | Approvers | Approver is not the submitter; `operate_together` answered; optional signature (reuse `SignatureBox` once committed). |
| `APPROVED` | reopen | `REOPENED` | Approvers | Period not locked; reason required; `reopen_count + 1`. The approved state stays in the event log. |
| `DRAFT` | delete | soft-deleted | Creator or ADMIN | Never after first submit. |

### 7.2 Roles (audit-service canonical roles)

| Capability | Roles |
|---|---|
| Read list/detail/evidence/events | All authenticated roles except DEPARTMENT_HEAD and AUDITEE; VIEWER read-only (existing `DenyViewerWrites`). |
| DEPARTMENT_HEAD, AUDITEE | Read `APPROVED` assessments of their own department only (department resolved from the JWT `employee_code` through master). |
| Assessors: create, propose, update/override, submit, delete own draft | AUDITOR, AUDIT_MANAGER, CHIEF_AUDIT_EXECUTIVE, ADMIN |
| Approvers: return, approve, reopen | CHIEF_AUDIT_EXECUTIVE (EXECUTIVE canonicalises to it), ADMIN, the same set as executive summary approve (Q1) |
| Lock/unlock period | CHIEF_AUDIT_EXECUTIVE, ADMIN |
| RCM tag write | Same as RCM writes today (any non-viewer). The existing open RCM write access is a separate finding (12, R5). |

Enforced with role checks in audit-service (`CanonicalRole`); the frontend gets strict helpers `canAssessInternalControl`, `canApproveInternalControl` in `useRbac.ts` that check actual roles and ignore the role-preview mode, like `canManageOrgLevels`.

### 7.3 Locking

- `APPROVED` is read-only (except reopen).
- Locking a period (year, or year + quarter; locking the year also locks its quarters) blocks create, propose, update, submit, reopen and delete for that period. Unlock needs a reason; both are logged.
- The evidence is a snapshot: later edits to RCM, risks, LHA or ATR never change an approved assessment. The detail page can show "source data changed since approval" by re-running the collectors read-only (no write). No cross-service lock blocks RCM edits for a locked year (Q5).

## 8. API

Responses use the existing `{success, data}` shape. All list sort fields are whitelisted against the GORM schema; filters are typed (uuid, int, enum) and never passed through as column names.

### 8.1 risk-service (under `/api/v1/rcm`, no Kong change)

| Method and path | Auth | Notes |
|---|---|---|
| `GET /rcm/coso-principles?framework=COSO_2013` | authenticated | Components with nested principles. Registered before `/rcm/:id`. |
| `GET /rcm` | authenticated | Adds `department_id`, `coso_component_code`, `coso_principle_codes` to each row. New filters `department_id` (uuid), `component` and `principle` (validated against the catalogue). |
| `POST /rcm`, `PUT /rcm/:id` | non-viewer (unchanged) | Accept `department_id`, `coso_component_code`, `coso_principle_codes`. Unknown codes or a component that matches no tagged principle -> 422. Omitted fields leave tags unchanged; `[]` clears them. |
| `GET /rcm/coso-signals?department_id=&year=&quarter=&location_id=&as_of=` | authenticated | Read-only. Returns `ra.quarterly_coverage`, `ra.above_appetite_mitigated`, `ca.rcm_effectiveness` (with per-principle means) with numerators, denominators and ids. Called by audit-service (caller's token forwarded) and by the evidence drill-down. |
| `GET /rcm/summary` | authenticated | Unchanged contract, but an empty set returns `null` scores instead of the fake 60 %. Check the dashboards (9.4) handle null. |

### 8.2 audit-service (new prefix `/api/v1/ic-assessments`)

| Method and path | Who |
|---|---|
| `GET /ic-assessments?year=&quarter=&department_id=&location_id=&status=&sort=&order=&page=&page_size=` | readers (7.2) |
| `GET /ic-assessments/summary?year=&quarter=` | readers; latest `APPROVED` per department: component statuses and ratings, overall conclusion, counts |
| `POST /ic-assessments` `{department_id, year, quarter, location_id?}` | assessors |
| `GET /ic-assessments/:id` | readers; header, components, principles, current evidence |
| `PUT /ic-assessments/:id` `{version, components[], principles[], final_conclusion, conclusion_text, conclusion_justification}` | assessors |
| `POST /ic-assessments/:id/propose` | assessors |
| `GET /ic-assessments/:id/evidence?run=` | readers |
| `GET /ic-assessments/:id/events` | readers except DEPARTMENT_HEAD/AUDITEE |
| `POST /ic-assessments/:id/submit` / `return` / `approve` / `reopen` | per 7.1 |
| `DELETE /ic-assessments/:id` | creator or ADMIN, `DRAFT` only |
| `GET /ic-assessments/periods?year=`, `POST /ic-assessments/periods/lock`, `POST /ic-assessments/periods/unlock` | readers / CAE, ADMIN |

Static paths (`summary`, `periods`) are registered before `/:id`.

### 8.3 Other services (existing prefixes)

- auth-service: `GET /api/v1/confidentiality/acceptance-summary?employee_codes=a,b,c` (max 500) -> `[{employee_code, has_user, accepted, accepted_at, version}]`. Roles: AUDITOR, AUDIT_MANAGER, CAE, ADMIN (EXECUTIVE). Counts only, no agreement content. Current version = the version the confidentiality service asks users to accept.
- master-service: `GET /employees?department_id=&is_active=` filter (additive, typed).

### 8.4 Kong

Add `/api/v1/ic-assessments` to the audit-service route in all four configs: `backend/kong-gateway/kong/dev/kong-dev.yml`, `backend/kong-gateway/kong/prod/kong.yml`, `backend/tenants/bai/kong/kong.yml`, `scripts/templates/kong-tenant.yml.tpl`. Without it the pages 404 through Kong while working against the service directly. Nothing else needs Kong: risk routes are under `/api/v1/rcm`, the auth and master additions are under existing prefixes, and the service-to-service calls go over the Docker network.

## 9. Frontend

### 9.1 Navigation

New item "Internal Control Assessment (COSO 2013)" in the Risk Management group of `frontend/layouts/default.vue`, directly after "Risk Control Matrix". Routes: `/internal-control-assessment` (list) and `/internal-control-assessment/[id]` (detail). Hidden for roles with no read access.

### 9.2 Assessment pages

- **List**: year and quarter selectors (`utils/fiscalPeriod.ts`), branch filter, status filter. A department x component grid: each cell shows status icon + text (`YES`/`PARTIAL`/`NO`/not assessed) and the final rating, never colour alone, dark mode included; last column the overall conclusion. Rows without an assessment show "Not started" and a create action for assessors. Period lock banner and lock/unlock for CAE/ADMIN.
- **Detail**: header (department path or name, branch, period, status, version, who/when); status stepper; action bar by state and role (Propose / Re-propose, Save, Submit, Return, Approve, Reopen).
  - Five component cards: proposed % and rating with data coverage, or "Insufficient data" plus reasons; final rating, status and severity; override reason, required and highlighted when the final differs from the proposal; derived status from principles.
  - Principles table grouped by component: status, severity, rationale; tagged control count and mean, linking to the RCM page filtered by department and principle.
  - Evidence drawer per component: every signal with value, numerator/denominator, weight, state, scope note, collected time, and links to the source page (risk profile, RCM, LHA list, ATR list filtered by report). Earlier proposal runs selectable.
  - Conclusion: derived vs final, "components operate together" (approver), justification.
  - History tab from `/events`.
  - Unsaved-change guard; 409 version conflict shows a reload prompt.

### 9.3 RCM changes (`pages/risk-profile/risk-control-matrix/index.vue`, `stores/rcm.ts`, `utils/rcmForm.ts`)

- Form step 1: the disabled department text becomes a required master department select (`department_id`). Existing rows with no `department_id` show their old text as a hint until a department is picked.
- New block in step 3 "COSO 2013 mapping": primary component select, principle multi-select grouped by component (labels from i18n), helper text per principle. Optional; an untagged control is valid.
- Step 3 title and the cards: "COSO 5 Dimensions Evaluation" becomes "Control effectiveness criteria"; the interpretation table stays.
- Table: principle chips column; filters for component and principle; the department filter uses master departments plus "Unassigned".
- Store: `fetchCosoCatalogue()` (cached), the new fields on `RCMItem`, filter state.

### 9.4 Dashboards

- `ExecutiveDashboard.vue` and `AdminDashboard.vue`: remove the mapping of RCM dimensions to component names and every hard-coded fallback number. The "Internal control (COSO 2013)" section reads `GET /ic-assessments/summary`: per component the number of departments `YES`/`PARTIAL`/`NO`, the overall conclusion count, and an empty state "No approved assessment for this period" with a link to the list. The RCM card stays, renamed "Control effectiveness (RCM)", and shows "No data" when the summary is null. Hard-coded Indonesian text in those sections moves to i18n.
- Auditor and auditee dashboards: no change in phase 1.

### 9.5 i18n

New namespaces in `frontend/locales/en/common.json` and `frontend/locales/id/common.json`: `coso.components.{CE,RA,CA,IC,MA}`, `coso.principles.P01..P17` (short title + description), `icAssessment.*` (statuses, severities, conclusions, signals by `signal_key`, evidence states, workflow actions, errors, empty states), and renamed `rcm.*` keys for "Control effectiveness criteria". Indonesian terms follow the standard Indonesian translation of the COSO terms (Lingkungan Pengendalian, Penilaian Risiko, Aktivitas Pengendalian, Informasi dan Komunikasi, Aktivitas Pemantauan); final wording is owned by `auditsphere-software-translation`.

## 10. Migration and rollout

| Change | Service | Kind | Safe on a running deploy |
|---|---|---|---|
| `coso_components`, `coso_principles`, seed | risk | new tables, insert-if-missing in `risk up` | Yes |
| `risk_control_matrix.department_id`, `coso_component_code` | risk | nullable columns, no default | Yes; old rows NULL |
| `rcm_coso_principles` | risk | new table | Yes |
| `ic_*` tables (6) | audit | new tables via `migrate up` | Yes |
| `action_taken_reports.action_plan_submitted_at` (1b) | audit | nullable column | Yes; never backfilled |
| employees filter, acceptance summary | master, auth | no schema | Yes |

- No data backfill runs in a migration. The only backfill (`backfill-rcm-department`) is an explicit, dry-run-by-default command.
- Order: risk-service and audit-service images can deploy in any order. The audit collectors treat a 404 from `/rcm/coso-signals` as `SOURCE_UNAVAILABLE`.
- Prod stack: migrations run on deploy (`risk up`, `audit migrate up`).
- Tenant `bai` and any other existing tenant: the stack runs `serve` only, so after pulling the new images the operator must run `./risk up` and `./audit migrate up` once for the tenant, then reload Kong. Re-running is a no-op.
- New tenants: `onboard-tenant.sh` already runs `risk up` and `audit migrate up`; the catalogue arrives with `risk up`, so `--empty-data` tenants get it too. No script change.
- Rollback: the old image ignores the new tables and columns; nothing to undo.
- Seeders (`risk seed`, `audit seed`) are not changed in phase 1. A demo assessment seed is a later, separate item (Q12).

## 11. Phase 2 outline: COSO ERM 2017 (not built now)

- Same catalogue tables with `framework = COSO_ERM_2017`: 5 components (Governance & Culture; Strategy & Objective-Setting; Performance; Review & Revision; Information, Communication & Reporting), 20 principles; insert-if-missing in `risk up`.
- New join `risk_register_erm_principles (risk_register_id, principle_id)` plus nullable `risk_register.erm_component_code`; tags edited on the Corporate Risk Profile form.
- Link to COSO 2013 through the existing `risk_control_matrix.risk_id`: a risk tagged ERM "Performance" shows its controls and their COSO 2013 principles, and the assessment evidence can show, per ERM component, which risks lack tagged controls.
- Prerequisite: proper risk ownership (`risk_profile.department_id` written by the CRP API), which also replaces the RCM-derived risk set R.
- Also phase 2: roll-up of child units into parent units on flexible tenants, and configurable signal weights.

## 12. Risks

1. R1 Department matching for LHA/ATR is by name. Renamed departments or free text break the match; unmatched counts are always shown, and flexible tenants can switch to the letter's `auditee_unit_id` later.
2. R2 RCM rows without a department give `NO_DATA` for RA and CA until departments are assigned. This is correct but will look empty on existing tenants; the list page shows the unassigned RCM count.
3. R3 Legacy RCM ratings default to 3 when unset; the CA signal shows the caveat count.
4. R4 Proposals depend on three other services; partial proposals are allowed and marked.
5. R5 (existing, out of scope) RCM writes are open to every non-viewer role, including AUDITEE. Tags inherit that. Recommend a separate fix restricting RCM writes to the audit roles.
6. R6 Kong: missing `/api/v1/ic-assessments` in any of the four configs gives a 404 on that stack only.
7. R7 Uncommitted RCM and dashboard work (2.3) conflicts with F1/F3 if started in parallel.

## 13. Open questions (defaults apply until decided)

1. Q1 Approvers: CAE + ADMIN (default, same as executive summary)? Should AUDIT_MANAGER also approve?
2. Q2 Branch-level assessments in phase 1 (default: allowed, optional), or department only?
3. Q3 Quarterly assessments (default: allowed, quarter optional) or annual only?
4. Q4 Thresholds: LHA within 30 days of the letter finish date, management response within 14 days. Constants (default) or tenant settings?
5. Q5 Should locking a year also block RCM edits for that year (cross-service check), or is the evidence snapshot plus the drift indicator enough (default)?
6. Q6 May DEPARTMENT_HEAD/AUDITEE see approved assessments of their own department (default: yes)?
7. Q7 Allow `NOT_APPLICABLE` principles with a rationale (default: yes), or must all 17 always be assessed?
8. Q8 Run `backfill-rcm-department --apply` on prod and `bai` after reviewing the dry run, or leave legacy rows unassigned?
9. Q9 Approval signature image (reuse `SignatureBox`) required, optional (default) or not used?
10. Q10 Signal weights fixed in code (default) or configurable per tenant?
11. Q11 Integrity pact: count only the SYSTEM agreement at its current version (default), or also ROLE agreements?
12. Q12 Should the demo seed (`--with-demo-data`) include tagged RCM rows and one approved assessment for the walkthrough?

## 14. Implementation tasks and sequencing

| # | Agent | Depends on | Task |
|---|---|---|---|
| B1 | auditsphere-backend | none | risk-service: models `CosoComponent`, `CosoPrinciple`, `RcmCosoPrinciple`; additive fields on `RiskControlMatrix`; `cmd/migrate.go` AutoMigrate + `seedCosoCatalogue` (insert-if-missing) in `risk up`, also in `dropAllTables`; RCM controller (fields, validation, transactional tag write, new filters); `GET /rcm/coso-principles`; `GET /rcm/coso-signals` with the ported risk matrix and branch rule; empty `/rcm/summary` returns null; `cmd/backfill_rcm_department.go` (dry run default); tests including a migrate run on a copy of an existing schema (additive diff only) and a matrix parity test. Fix the model comment "COSO 2013 5 Dimensions". |
| B2 | auditsphere-backend-2 | none | auth-service `GET /confidentiality/acceptance-summary` with role check and limits; master-service `GET /employees` `department_id`/`is_active` filters; tests. |
| B3 | auditsphere-backend-2 | contracts of B1/B2 (can start against them) | audit-service: models of 4.2 in `migrate up`; principle constant list; collectors (local charter/LHA/ATR/exec summary, `riskclient` for signals, `masterclient` for department/employees/PIC, `authclient` for acceptance); formulas 5.1-5.3; roll-up 4.3; workflow 7.1; role checks 7.2; locks 7.3; event log; endpoints 8.2 with sort whitelist; tests per signal (no data -> NULL, unreachable -> `SOURCE_UNAVAILABLE`), workflow and SoD tests, VIEWER and DEPARTMENT_HEAD scoping tests. Phase 1b: `action_plan_submitted_at` set on the first action plan save. |
| D1 | auditsphere-devops | B3 merged | Add `/api/v1/ic-assessments` to the four Kong configs; deploy; for tenant `bai` run `./risk up` and `./audit migrate up` once, reload Kong. |
| F1 | auditsphere-frontend | B1 contract; after 2.3 work is committed | RCM: department select, COSO mapping block, chips, filters, relabel; store and `utils/rcmForm.ts`; unit tests. |
| F2 | auditsphere-frontend-2 | B3 contract (mockable) | Navigation item; `stores/ic-assessment.ts`; list and detail pages; evidence drawer; history; `useRbac` helpers; unit tests for roll-up display, override-required logic and empty states. |
| F3 | auditsphere-frontend | F2 store; after 2.3 work is committed | Executive and Admin dashboards per 9.4. |
| T1 | auditsphere-software-translation | key lists from F1-F3 | EN/ID keys of 9.5, the 17 principle titles and descriptions, renamed RCM keys, hard-coded dashboard strings; parity check. |
| Q1 | auditsphere-qa | B1-B3, D1, F1-F3, T1 | Section 15. |

Parallel: B1 and B2 start now; B3 starts against their contracts; F2 runs on mocks alongside B3; F1/F3 wait for the uncommitted RCM and dashboard work; D1 after B3; T1 alongside F*.

## 15. Verification checklist

1. Additive only: dump the schema of copies of prod `rb_audit` risk/audit and `bai` databases, run the new `risk up` and `audit migrate up` twice, dump again. The diff contains only the new tables, the two new RCM columns (and the 1b ATR column); row counts of existing tables unchanged; `coso_principles` has 17 rows after both runs.
2. Seed: delete one principle row in a test DB, run `risk up`, it comes back; edit `name_en` on another, run `risk up`, the edit stays.
3. RCM: create with department and tags; unknown principle 422; component not matching a tag 422; `PUT` without tag fields keeps tags; `[]` clears; VIEWER 403; legacy rows list with `department_id: null`; empty `/rcm/summary` returns null, no 60 %.
4. Signals: a department with no RCM rows gives CA `NO_DATA` and a NULL score, shown as "Insufficient data"; stopping risk-service gives `SOURCE_UNAVAILABLE` and a partial proposal; a branch-scoped assessment narrows CA/RA and marks audit-side signals department-wide.
5. Workflow: submit refused while incomplete or a reason is missing; approver = submitter refused; `EFFECTIVE` refused with a `MAJOR`; approve without `operate_together` refused; reopen logs the reason; locked period refuses every write; every action leaves one event row with before/after.
6. RBAC: AUDITOR can create/propose/submit, not approve; CAE (seeded `EXECUTIVE`) can approve; VIEWER reads, all writes 403; DEPARTMENT_HEAD sees only approved assessments of its own department.
7. Kong: `GET /api/v1/ic-assessments` returns 200 through the dev, prod and `bai` Kong, and through a stack generated from the template.
8. Dashboards: with no approved assessment the empty state shows and no hard-coded numbers appear; EN/ID switch shows all new strings translated.
