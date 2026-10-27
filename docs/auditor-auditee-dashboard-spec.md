# Auditor and Auditee dashboard redesign: scope spec

Verified against the code on 2026-10-21 (branch `staging`, HEAD `f35b464`, a merge commit). Re-check section 2 before relying on it.

Status update, verified on 2026-10-27 (HEAD `59dbf5a`): B1 is done (`/auth/me` returns `employee_code`, `employee_id`, `department_id`; the JWT carries `employee_id` and `employee_code` but not `department_id`). FE-G1 and FE-G3 are done (`AuditeeDashboard.vue`). B4 and W1 (calendar) are designed in detail in `docs/audit-calendar-plan.md`, which also records that the RKA form does not write `audit_activities.planned_start/end`, so those columns are not a usable calendar source.

Owner: project lead. Implementers: `auditsphere-frontend` (sections 6.1 and 7) and `auditsphere-backend` (sections 6.2 and 7), working in parallel against the contract in section 5. `auditsphere-solution-architect` reviews the schema changes in section 6.2 before they are built. `auditsphere-software-translation` owns the new i18n keys. `auditsphere-qa` owns the RBAC smoke checks in section 8.

---

## 1. Summary

The auditor view and the admin view share one page, `frontend/pages/dashboard/index.vue`. When `isAuditorView` is true, the page renders `frontend/components/dashboard/AuditorDashboard.vue`. Everything else gets the admin view.

The redesign has three goals.

1. **Auditor view**: a working view for the fieldwork auditor and team leader. It has a calendar of execution dates, a task list with deadlines, notifications for requests and revisions, and engagement status and progress. It keeps audit statistics and planning and coverage from the admin view.
2. **Auditee view (new)**: a separate view for the `AUDITEE` role. Today an auditee lands on the admin view (section 2.1). It gets the same calendar, tasks, notifications and status, filtered to the auditee's own unit and action plans.
3. **Honest data**: each widget is either backed by an existing endpoint or marked as a gap with the owning service. Several things the current dashboards display are not backed by data (section 2.4), and the redesign must not carry them over.

The main risk is data scope. Most audit, master and risk endpoints check only that the JWT is valid, and the role is not checked (section 2.3). Until the backend adds per-user scoping, any "my engagements" widget is only a filter in the browser. The spec therefore puts scoping on the server (section 6.2, item B3).

### Decisions this spec assumes (confirm before build, section 9)

- D1. A user is linked to an employee by business key: `users.employee_id` = `employees.employee_code` (per the project convention). No foreign key.
- D2. The auditee's unit is a `department`. Today `users.department` is free text (`backend/auth-service/models/user.go:19`), so the link needs a stable key.
- D3. In v1, notifications are in-app only. Email reuses the existing ATR reminder mailer (`backend/audit-service/pkg/mailer`). No new Kafka topic.
- D4. The `DEPARTMENT_HEAD` and `EXECUTIVE`/`VIEWER` roles keep the admin view. They are out of scope for this redesign.

---

## 2. Current state (evidence)

### 2.1 Routing and role views

| Item | Evidence |
|---|---|
| Page switch: `v-if="isAuditorView"` renders `<AuditorDashboard />`, else the admin markup | `frontend/pages/dashboard/index.vue:4-7` |
| Admin `onMounted` returns early only for the auditor view, so an auditee runs every admin fetch | `frontend/pages/dashboard/index.vue:749-751` (`if (isAuditorView.value) return;`) |
| `isPureAuditor` matches only AUDITOR, AUDIT_STAFF, STAFF_AUDIT, LEAD_AUDITOR | `frontend/composables/useRbac.ts:209-221` |
| `isAuditorView` is true for pure auditors, or when the admin previews the view (`activeRoleView === 'auditor'`) | `frontend/composables/useRbac.ts:228-236` |
| There is no `isAuditeeView`. `AUDITEE` is in the role list only | `frontend/composables/useRbac.ts:151`, `frontend/types/auth.ts:23` |
| Sidebar gating: the auditor branch hides four items. Nothing hides items for `AUDITEE`, except the master-employee check | `frontend/layouts/default.vue:607-630` |
| The auditor view's stores are assignment letters, working papers, audit results (reports and recent findings), and RCM | `frontend/components/dashboard/AuditorDashboard.vue:1292-1305` (`loadInitialAuditorData`) |

**Consequence:** an `AUDITEE` user sees the admin dashboard and the full sidebar. The admin view fetches annual plans, risks, RCM, ATR summary and audit executions from endpoints that do not check role (section 2.3).

### 2.2 The current auditor view (`frontend/components/dashboard/AuditorDashboard.vue`)

| Section | Line (approx.) | Data source | Scoped to the user? |
|---|---|---|---|
| Header with role badge, role preview pill, and "Sinkronisasi Data" button | 1-60 | `authStore.user.fullName`; `useRbac` | n/a |
| Four stat cards: active fieldwork assignments, in-progress, working papers, KKA completion | 60-180 | `combinedFieldworkList` (line 907) from every assignment letter | **No.** Every letter is counted. |
| Draft findings table and severity counts | 682 (table), 1025-1120 | Draft LHA reports (`auditResultStore.reportList`), locally computed auto-findings from KKA (`getLocalAutoFindings`, `frontend/stores/audit-result-report.ts:310`), recent findings API | Partly. The selected letter filters the list. |
| PBC document tracker | 775, 1122-1155 | `fieldworkStore.fieldworkData[letter].documents` | Per selected letter. Status is inferred: "Received" if a file is attached, otherwise "Pending". There is no status field. |
| KKA form status F01 to F05 | 1157-1183 | `workingPaperStore.dataF01..F05` | Per selected letter |
| RCM quick reference (top 3) | 826-838, 1185 | `rcmStore.rcmList` (risk-service `/rcm`) | No |
| Mount | 1292-1372 | `Promise.allSettled` of five store fetches, silent. Sync button shows a toast. | n/a |

Loading and error states: there is no skeleton and no error card for the view as a whole. Only the draft-findings table has an empty row (line 682). The PBC tracker has an empty message (line 775). The mount call swallows errors with `console.warn`.

Hard-coded strings: the header, the notice and the labels are in Indonesian and English, not `t()` keys (for example line 12, "Dashboard Pelaksana Audit (Fieldwork Auditor)"). These must move to i18n.

### 2.3 Role enforcement on the backend (what the server actually checks)

| Service | Evidence | Result |
|---|---|---|
| audit-service | `backend/audit-service/routes/handler.go:71` applies `Authenticate()` to all of `/api/v1`. The only role check in its routes is the assignment-letter status guard (`handler.go:184`). | Every audit endpoint (letters, fieldwork, working papers, LHA, annual plans, audit executions, audit activities) is readable by any valid JWT. |
| audit-service ATR | `backend/audit-service/controllers/action_taken_report/handler.go` (`List`, line 361-379) filters to `pic_user_id = <me>` unless the user has `view_all_action_taken_report`. Role-to-permission mapping is in `backend/audit-service/services/action_taken_report/rbac.go`. | **The only server-side row scope in the audit service.** Reuse this pattern. |
| risk-service | `backend/risk-service/routes/routes.go`: no role middleware on routes. | Any valid JWT can read risks, RCM and the audit universe. |
| master-service | `backend/master-service/routes/routes.go`: no role middleware. | Any valid JWT can read employees and departments. |
| auth-service | `backend/auth-service/routes/registry.go:83` (`adminOnly`), `:148` (`/users/assignable`), `:150` (`/users/contacts`). | Role checks exist here only. |

Also see `docs/backend-enhancement-plan.md`. Its blocker list says audit, master and risk authorise any valid token.

### 2.4 Admin dashboard (`frontend/pages/dashboard/index.vue`): what to keep and what not to carry over

| Admin widget | Line | Status |
|---|---|---|
| Audit statistics: planned audits, open findings, execution %, ATR compliance % | 245-362 (approx.) | Keep, scoped (section 3, W5 and W6). |
| "↑ 12%" and "↑ 3" on the first two stat cards | `index.vue:58` and `:77` | **Hard-coded literals, not data.** Do not carry over. Remove the trend text until a real comparison exists. |
| Audit planning and coverage (annual plan list, audit universe, risk heat map) | `index.vue` (`annualPlanStore`, `riskProfileStore`, `RiskHeatMap`) | Keep for auditors. Hide from auditees. |
| KPI forecasting and anomaly detection (analytics, `useAiAnalytics`) | `index.vue` (`timeseriesState`, `isolationState`) | Admin only. Not in the auditor or auditee view. |
| ATR summary | `atrStore.fetchSummary()` (`index.vue:761`) | Keep. The store already scopes to the caller (section 2.3). |

### 2.5 Calendar, notifications, comments, revisions, document requests: what exists

| Capability | What exists | Evidence | Gap |
|---|---|---|---|
| Calendar | Typed dates on `audit_activities` (`planned_start`, `planned_end`), `action_taken_reports.due_date`, `audit_result_reports.report_date`, `audit_assignments.start_date`/`end_date` | `backend/audit-service/models/audit-activity.go:189-190`; `audit-assignment.go:27-28`; `action-taken-report.go` (`DueDate`); `audit-result-report.go` (`ReportDate`) | No calendar endpoint. Letter periods are free text (`assignment-letter.go`, `StartPeriod`/`FinishPeriod`). `audit_assignments` has no frontend producer. |
| Notifications | Email only: ATR reminder (`POST /action-taken-reports/:id/reminder`, sent via Resend). The header has no bell. The only frontend "notification" is the toast component. | `backend/audit-service/controllers/action_taken_report/reminder.go`; `frontend/components/shared/ToastNotification.vue`; `frontend/layouts/default.vue` (no bell) | **No in-app notification model, store, endpoint or UI.** |
| Comments | None on LHA, KKA, fieldwork or ATR. ATR has one `review_note` string (`action-taken-report.go`, `ReviewNote`). The satisfaction survey has a `comments` field. | `backend/audit-service/models/action-taken-report.go`; `frontend/pages/audit-result-report/satisfaction-survey.vue:555` | **No comment thread model or endpoint.** |
| Finding revisions | ATR "returned for revision": status goes back to `IN_PROGRESS` with `review_note` set (`frontend/components/action-taken-report/ATRActionPlanForm.vue:193`). The annual plan has `revisionHistory` (`frontend/stores/annual-audit.ts:835-850`). LHA findings have no revision state. | As cited | Only the ATR path has a revision signal. **LHA finding revisions have no model.** |
| `AuditFinding` model | `backend/audit-service/models/audit-finding.go` has `Status` (OPEN/RESOLVED/CLOSED/OVERDUE) and `TargetDate`. | File as cited | **Orphaned:** no routes, no references. Findings in use live as JSON inside `audit_result_reports.findings`. Decide to reuse or delete (solution architect). |
| Document requests (PBC) | `fieldwork_documents` has `documentName`, `description`, `requiredDate` (a string). No status, requester, recipient or response. | `backend/audit-service/models/audit-fieldwork.go` (`FieldworkDocument`) | **No request workflow.** No requested-by, requested-from, due date (typed), status or response. |
| Team membership | `assignment_letters.membersList` holds `{name, role}` only. `activity_plans.resourceAuditors[].id` is a string with no defined meaning. `audit_activities.team_leader_id` is a master-service employee UUID. | `backend/audit-service/models/assignment-letter.go` (`LetterMember`); `backend/audit-service/models/audit-activity.go:70-76, 186-187` | **No user-to-engagement link.** |
| Identity link | `users.employee_id` exists. `/auth/me` returns `models.UserInfo` (`id, username, email, full_name, phone, department, avatar_url, roles, must_change_password`), with no `employee_id`. | `backend/auth-service/models/user.go:13`; `backend/auth-service/models/dto.go:30-40`; `backend/auth-service/controllers/auth/auth_controller.go:215-218` | **The frontend cannot tell which employee the user is.** The employee lookup `GET /users/lookup` is ADMIN only (`backend/auth-service/routes/registry.go:165`; handler `backend/auth-service/controllers/user/user_controller.go:317-323`). |
| Engagement status and progress | `audit_activities.status` (`PLANNED`, `IN_PROGRESS`, `REPORTING`, `COMPLETED`). `audit_executions.progress`, `status`, `lead_auditor` (string). KKA completion is computed in the browser. | `audit-activity.go:191`; `backend/audit-service/models/audit-execution.go` | No per-engagement progress from the server. KKA completion is client-side only. |

---

## 3. Widget spec

Column "Auditor" and "Auditee" give visibility: **Full** (own data only), **Team** (own data plus team), **Hidden**. "Scoped" means the endpoint filters by the caller on the server (section 5). Until a scoped endpoint exists, the widget uses the interim source in brackets and must label its scope in the UI ("All engagements" instead of "My engagements").

| # | Widget | Auditor | Auditee | Source (target) | Source (interim) | Status |
|---|---|---|---|---|---|---|
| W0 | Header: role badge, name, sync button | Full | Full | `authStore.user` (`/auth/me`) | same | Keep. Move strings to i18n. Remove the "Matriks RBAC Baris 8 & 9" pill (it cites an RBAC matrix that is not in `docs/` or the repo). |
| W1 | **Calendar of important dates** | Team (engagements where the user is team leader or member) | Own (ATRs where the user is PIC, plus PBC due dates for the unit) | `GET /api/v1/me/calendar` (B4) | none. The widget shows an empty state until B4 exists. | **Missing** |
| W2 | **Notifications and reminders** | Own (new document requests, comments and ATR revisions waiting on the user, plus ATRs to review) | Own (new PBC requests, comments and ATRs returned for revision) | `GET /api/v1/notifications` (B9) | Derived list from `GET /action-taken-reports` (PIC scope) for ATR items only. Other types show as "not available". | **Missing** (ATR part partly derivable) |
| W3 | **Task list with deadlines** | Own tasks: KKA to finish, PBC to review, LHA draft to submit, ATR to review | Own tasks: ATR action plan to fill, evidence to upload, PBC to answer | `GET /api/v1/me/tasks` (B5) | `GET /action-taken-reports` (PIC scope, with `due_date`, `is_overdue`, `overdue_days`). Other tasks show as unavailable. | **Partial** |
| W4 | **Audit status and progress** | Team: each engagement with status, progress %, planned dates, lead | Hidden, except the engagements that produced the user's ATRs (show status only) | `GET /api/v1/me/engagements` (B3) | Client-side: `audit-activities` list (all rows, labelled "All engagements"), `audit-executions` | **Partial** |
| W5 | **Audit statistics** (kept from admin) | Team: planned audits, open findings, execution %, ATR compliance | Hidden (except ATR compliance for own ATRs) | `GET /api/v1/me/dashboard-summary` (B3) | `auditActivities`, `atrStore.fetchSummary()`, `auditResultStore` | Keep. Remove the hard-coded trend text. |
| W6 | **Audit planning and coverage** (kept from admin) | Full: PKAT coverage and audit-universe coverage | Hidden | `GET /api/v1/annual-audit-plans` (audit-service) and `GET /api/v1/audit-universe` (risk-service) | same | Keep. Auditor only. |
| W7 | Draft findings and severity counts (existing auditor widget) | Team (own letters) | Hidden | `GET /api/v1/audit-result-reports` and recent findings | same | Keep. Relabel with the scope. |
| W8 | Finding revisions needing a response | Team: LHA findings returned for revision | Own: ATR findings returned for revision, with `review_note` | Revisions in B8, plus ATR `status = IN_PROGRESS` with `review_note` | ATR only, via the list | **Partial** |
| W9 | PBC document tracker (existing) | Team (per letter) | Own: PBC items for the unit | `GET /api/v1/fieldwork/documents` with status and due date (B6) | Existing list, status inferred from the presence of a file | **Partial** |
| W10 | KKA form status F01 to F05 (existing) | Team (per letter) | Hidden | `workingPaperStore` | same | Keep for auditors. |
| W11 | RCM quick reference (existing) | Full (read) | Hidden | `GET /api/v1/rcm` (risk-service) | same | Keep as reference, lowest priority. |
| W12 | Auditee: unit and action-plan summary | Hidden | Full: own ATR list with status and progress, counts of overdue items | `GET /api/v1/action-taken-reports` (scoped) | same | New for auditee. Uses the ATR scope that already exists. |

Widgets not carried over from admin: the KPI forecast and anomaly cards (analytics, admin only), the risk heat map (admin only, auditee hidden, auditor keeps it only through W6 if needed), and the hard-coded trend text.

---

## 4. Role visibility

| Role (seeded) | Business term | View today | View after redesign | Notes |
|---|---|---|---|---|
| AUDITOR (also `AUDIT_STAFF`, `STAFF_AUDIT`, `LEAD_AUDITOR` aliases) | Fieldwork auditor, team leader | Auditor view (`AuditorDashboard.vue`) | Auditor view: W0, W1, W2, W3, W4, W5, W6, W7, W8, W9, W10, W11 | Team leader sees the team's engagements. Member sees the engagements they are on (needs B3). |
| AUDITEE | Auditee unit | Admin view, full sidebar, admin fetches | **Auditee view (new)**: W0, W1, W2, W3, W8 (own), W9 (own), W12 | Must not see admin fetches or other units' data. |
| ADMIN | Admin / preview | Admin view, or auditor view when previewing | Unchanged. Preview toggle still shows the auditor view. | The preview toggle must also be able to show the auditee view for testing. |
| EXECUTIVE (acts as CAE) | Chief audit executive | Admin view | Unchanged (out of scope, D4) | ATR review rights come from the role, not this view. |
| DEPARTMENT_HEAD | Audited unit head | Admin view | Unchanged (out of scope, D4). Notifications for ATR reminders go to the department head by e-mail (`reminder.go`). | Ask the product owner (section 9). |
| VIEWER | Read-only consumer | Admin view | Unchanged | |

---

## 5. Data contract for the new endpoints (for the frontend and backend in parallel)

All endpoints use the existing `/api/v1` base, the JWT `Authorization` header and the `ApiEnvelope` response wrapper used by the stores. Dates are ISO 8601 (`YYYY-MM-DD` for dates, RFC 3339 for times). The server resolves the caller from the JWT (`user_id`) and, for team membership, from the employee link (B1). Every field below is required unless marked optional.

**`GET /api/v1/me/calendar?from=YYYY-MM-DD&to=YYYY-MM-DD`** (audit-service)

```json
{
  "events": [
    {
      "id": "string",
      "type": "ENGAGEMENT_START | ENGAGEMENT_END | ATR_DUE | REPORT_DATE | PBC_DUE",
      "date": "YYYY-MM-DD",
      "title": "string",
      "entity_type": "audit_activity | action_taken_report | audit_result_report | fieldwork_document",
      "entity_id": "uuid",
      "status": "string",
      "is_overdue": false
    }
  ]
}
```

**`GET /api/v1/me/tasks`** (audit-service)

```json
{
  "tasks": [
    {
      "id": "string",
      "kind": "ATR_FILL_ACTION_PLAN | ATR_UPLOAD_EVIDENCE | ATR_REVIEW | ATR_REVISION | PBC_RESPOND | PBC_REVIEW | KKA_COMPLETE | LHA_SUBMIT",
      "title": "string",
      "entity_type": "string",
      "entity_id": "uuid",
      "due_date": "YYYY-MM-DD | null",
      "is_overdue": false,
      "overdue_days": 0
    }
  ]
}
```

**`GET /api/v1/me/engagements`** (audit-service). Team leader or member only, or all rows for ADMIN and EXECUTIVE.

```json
{
  "engagements": [
    {
      "id": "uuid",
      "code": "string",
      "title": "string",
      "role_in_team": "LEADER | MEMBER | AUDITEE",
      "status": "PLANNED | IN_PROGRESS | REPORTING | COMPLETED",
      "progress_percent": 0,
      "planned_start": "YYYY-MM-DD",
      "planned_end": "YYYY-MM-DD",
      "letter_number": "string | null",
      "department_id": "uuid | null"
    }
  ]
}
```

`progress_percent` is computed on the server (B3). Suggested rule: the share of KKA forms F01 to F05 completed for the linked letter. Confirm with the product owner.

**`GET /api/v1/me/dashboard-summary`** (audit-service)

```json
{
  "planned_audits": 0,
  "open_findings": 0,
  "execution_percent": 0,
  "atr_compliance_percent": 0,
  "atr_overdue": 0
}
```

**`GET /api/v1/notifications?unread=true&limit=50`** and **`PUT /api/v1/notifications/:id/read`**, **`PUT /api/v1/notifications/read-all`** (audit-service)

```json
{
  "items": [
    {
      "id": "uuid",
      "type": "PBC_REQUEST | COMMENT | REVISION_REQUESTED | ATR_ASSIGNED | ATR_RETURNED | ATR_SUBMITTED | ATR_DUE_SOON | ATR_OVERDUE",
      "title": "string",
      "body": "string",
      "entity_type": "string",
      "entity_id": "uuid",
      "due_date": "YYYY-MM-DD | null",
      "read_at": "RFC 3339 | null",
      "created_at": "RFC 3339"
    }
  ],
  "unread_count": 0
}
```

**`GET /api/v1/comments?entity_type=&entity_id=`** and **`POST /api/v1/comments`** (audit-service). Body: `{entity_type, entity_id, body, parent_id?}`. Response: the comment with `author_user_id`, `author_name`, `created_at`.

**`GET /api/v1/fieldwork/documents`**: existing list, with new fields `status` (`REQUESTED | RECEIVED | REVIEWED`), `requested_by_user_id`, `requested_from_user_id`, `due_date` (`YYYY-MM-DD`), `responded_at`. The old `requiredDate` string stays for compatibility.

Error states for all endpoints follow the existing envelope with a machine code. The frontend shows the per-widget error state (section 7) and never fails the whole page.

---

## 6. Gaps

### 6.1 Frontend (`auditsphere-frontend`, with `auditsphere-software-translation` for keys)

| ID | Gap | Where | Evidence |
|---|---|---|---|
| FE-G1 | No auditee view. `AUDITEE` falls into the admin view. | `frontend/composables/useRbac.ts` (add `isAuditeeView`); `frontend/pages/dashboard/index.vue:4-7` (add a branch) | Section 2.1 |
| FE-G2 | Admin `onMounted` runs for auditee and for any non-auditor. | `frontend/pages/dashboard/index.vue:749-751` (early-return for both new views) | Section 2.1 |
| FE-G3 | New component `AuditeeDashboard.vue` (does not exist). | `frontend/components/dashboard/` | Section 3 |
| FE-G4 | No calendar widget. Build a month and list view with Nuxt UI, without a new calendar library unless the team decides otherwise. | new component | Section 2.5 |
| FE-G5 | No notification panel and no bell in the header. | `frontend/layouts/default.vue` (header), new component | Section 2.5 |
| FE-G6 | No task list widget. | new component | Section 2.5 |
| FE-G7 | "My engagements" is labelled, but the data is all letters. Relabel until B3. | `frontend/components/dashboard/AuditorDashboard.vue:907`, `:1002` | Section 2.2 |
| FE-G8 | Hard-coded trend text on admin cards. Remove. | `frontend/pages/dashboard/index.vue:58`, `:77` | Section 2.4 |
| FE-G9 | Hard-coded strings in the auditor header, notice and labels. Move to i18n. | `frontend/components/dashboard/AuditorDashboard.vue` (header `:12`, notice `:63-71`) | Section 2.2 |
| FE-G10 | No per-widget loading, empty and error states. Only one table has an empty state. | `AuditorDashboard.vue` (all widgets) | Section 2.2 |
| FE-G11 | Admin preview toggle must also preview the auditee view. | `useRbac.ts:202-236` (`activeRoleView`) | Section 4 |
| FE-G12 | PBC "Received" is inferred from a file. Replace with the server status once B6 exists. | `AuditorDashboard.vue:1122-1155` | Section 2.2 |
| FE-G13 | Comment UI on ATR and PBC items (after B7). | `frontend/components/action-taken-report/`, PBC modal | Section 2.5 |
| FE-G14 | Sidebar for AUDITEE is not gated. Hide admin-only menus for the auditee view. | `frontend/layouts/default.vue:607-630` | Section 2.1 |

### 6.2 Backend (`auditsphere-backend`; `auditsphere-solution-architect` reviews the schema)

| ID | Owner service | Gap | Endpoint or field | Evidence |
|---|---|---|---|---|
| B1 | **auth** | Expose the employee link to the caller. | Add `employee_id` (and `department_id` once D2 is decided) to `models.UserInfo`, returned by `GET /auth/me`. | `backend/auth-service/models/dto.go:30-40`; `auth_controller.go:215-218` |
| B2 | **auth** | Department key for the auditee unit (D2). Today it is free text. | Add `department_id` (nullable, uuid) to `users`, or a stable `department_code`. Backfill is manual. | `backend/auth-service/models/user.go:19` |
| B3 | **audit** (reads master for employee and department) | No engagement scope. Membership is by name only, and the employee link is not available. | `GET /api/v1/me/engagements`, `GET /api/v1/me/dashboard-summary`. Resolve the caller's employee via B1, then match `audit_activities.team_leader_id` (master employee UUID) and member lists. The member list needs an ID, see B3a. | `backend/audit-service/models/audit-activity.go:186-187`; `assignment-letter.go` (`LetterMember`) |
| B3a | **audit** | Team members have no stable ID. `assignment_letters.membersList` is `{name, role}`. `activity_plans.resourceAuditors[].id` is undefined. | Add `employee_id` (or `user_id`) to `LetterMember` and to `ResourceAuditor`. Keep `name` for display. Decide the ID type first (D1). | `assignment-letter.go`; `audit-activity.go:70-76` |
| B4 | **audit** | No calendar aggregation. | `GET /api/v1/me/calendar?from&to` with the shape in section 5. Sources: `audit_activities.planned_start/end`, `action_taken_reports.due_date`, `audit_result_reports.report_date`, PBC `due_date` (B6). **Superseded by `docs/audit-calendar-plan.md` (2026-10-27)**, which has the verified source list, an extended payload and an org-wide `/api/v1/calendar`. | Section 2.5 |
| B5 | **audit** | No task aggregation. | `GET /api/v1/me/tasks`. Sources: ATR scope (PIC or reviewer), PBC items with `requested_from` = me, letters without a complete KKA. | Section 2.3 (ATR scope pattern) |
| B6 | **audit** | No PBC request workflow. | New fields on `FieldworkDocument` (`status`, `requested_by_user_id`, `requested_from_user_id`, `due_date`, `responded_at`); respond endpoint `PUT /api/v1/fieldwork/documents/:id/respond` for the auditee; list filters `?status=&requested_from=`. Additive migration. | `backend/audit-service/models/audit-fieldwork.go` (`FieldworkDocument`) |
| B7 | **audit** | No comment thread. | New table `comments` (`id, entity_type, entity_id, parent_id, author_user_id, author_name, body, created_at`, index on `(entity_type, entity_id)`), `GET/POST /api/v1/comments`. Entity types first: `action_taken_report`, `fieldwork_document`, `audit_result_report`. | Section 2.5 |
| B8 | **audit** | No LHA finding revision model. The ATR revision signal exists, but findings are JSON inside `audit_result_reports.findings`. | Add a `finding_revisions` table (`id, audit_result_report_id, finding_id, version, changed_by_user_id, change_summary, needs_response, responded_at, created_at`), with `POST /api/v1/audit-result-reports/:id/findings/:findingId/revisions` and `PUT .../revisions/:revId/respond`. Or extend the existing ATR review states. Decide with the solution architect. The orphaned `AuditFinding` model (`models/audit-finding.go`) is a candidate, or should be deleted. | `backend/audit-service/models/audit-finding.go`; `audit-result-report.go` |
| B9 | **audit** | No notification store, endpoints or producers. | Table `notifications` (`id, recipient_user_id, type, title, body, entity_type, entity_id, due_date, read_at, created_at`). Endpoints as in section 5. Producers: ATR assign, submit, review (return for revision), reminder; PBC request and response; comment created; finding revision requested. Email for the same events reuses `backend/audit-service/pkg/mailer`, behind a flag. | Section 2.5 |
| B10 | **audit** | No due-soon and overdue job. | A periodic job in `serve` (same pattern as the existing timers) that creates `ATR_DUE_SOON` (for example 3 days before) and `ATR_OVERDUE` notifications once per item per day. `IsOverdue` and `OverdueDays` already exist on the ATR model. | `backend/audit-service/models/action-taken-report.go` (derived fields) |
| B11 | **audit** | Engagement progress is not stored. | Computed in `GET /me/engagements` (see the progress rule in section 5). No new column. | `audit_executions.progress` exists; `audit_activities` has none |
| B12 | **risk** | No new endpoint for coverage. Auditor-only coverage (W6) reads `GET /audit-universe` and `/rcm`. Role checks are absent (section 2.3). | No new endpoint. The frontend hides W6 from auditees. A server-side role check is tracked in `docs/backend-enhancement-plan.md`, not in this spec. | `backend/risk-service/routes/routes.go` |
| B13 | **master** | The employee-to-user lookup is ADMIN only (auth side). | No change in master. B1 gives the frontend the link, and B3 does the lookup server-side. | `backend/auth-service/routes/registry.go:165`; `user_controller.go:317-323` |

---

## 7. Empty, loading and error states

Each widget has its own state. A failed widget never blanks the page.

| Widget | Loading | Empty | Error |
|---|---|---|---|
| W1 Calendar | Skeleton grid | "No important dates in this range." with the date range picker | Inline alert with Retry. The rest of the page still renders. |
| W2 Notifications | Three skeleton rows | "You have no new notifications." | Inline alert with Retry. Keep the last list. |
| W3 Tasks | Skeleton rows | "No open tasks. Nothing is waiting on you." | Inline alert with Retry. |
| W4 Engagements | Skeleton table | "You are not on any engagement yet." (auditor), or "No engagement is linked to your unit." (auditee) | Inline alert with Retry. |
| W5 Statistics | Skeleton cards | Zeros, with the subtext "No data for this period." | Card-level error text. |
| W6 Planning and coverage | Skeleton chart | "No annual plan approved yet." | Card-level error. |
| W7 Draft findings | Skeleton rows | Existing empty row | Card-level error |
| W8 Revisions needing response | Skeleton rows | "No findings are waiting for your response." | Card-level error |
| W9 PBC tracker | Skeleton rows | Existing message: "No documents requested yet." | Card-level error |
| W10 KKA status | Skeleton | "Select an assignment letter to see KKA status." | Card-level error |
| W11 RCM | Skeleton | Existing message | Card-level error |
| W12 Auditee summary | Skeleton | "No action taken reports are assigned to your unit." | Card-level error |

Interim behaviour (before the backend is ready): show the widget with the "not available yet" note, not an error. For example, "Calendar is not connected yet." This is the only state that uses that wording.

Permissions state: if the server returns 403 for a widget, show "You do not have access to this section." and do not retry.

---

## 8. RBAC and acceptance checks (for `auditsphere-qa`)

1. An AUDITEE token sees only ATRs where it is the PIC (`GET /action-taken-reports`), and `GET /me/tasks` and `GET /me/calendar` return only its own items.
2. An AUDITOR who is not on a letter's team does not see that engagement in `GET /me/engagements`.
3. An AUDITEE does not trigger the admin fetches. Check the network log on login, with the admin endpoints (`/annual-audit-plans`, `/rcm`, `/risks`) not called.
4. A notification for a PBC request reaches the requested-from user only.
5. An admin in preview mode can switch to the auditee and auditor views and back.
6. Locale parity: every new key exists in `frontend/locales/en/common.json` and `frontend/locales/id/common.json`.
7. Frontend typecheck and vitest pass. Backend `go vet` and tests pass for audit-service and auth-service.

---

## 9. Open questions (product owner)

1. D1: confirm `users.employee_id` = `employees.employee_code` as the only link. Is this already used in production data?
2. D2: is the auditee unit a master `department` (by ID) or a `department_head` relationship?
3. D3: are in-app notifications enough for v1, or is e-mail required for PBC and revision requests too?
4. Team leader and member: do members see engagements they are on, or only the leader? Defaults in section 4.
5. Progress rule for W4 (section 5).
6. Should DEPARTMENT_HEAD get the auditee view? (Default: no, D4.)
7. Is the `AuditFinding` model (`backend/audit-service/models/audit-finding.go`) the intended finding store for LHA revisions, or should it be removed?

---

## 10. Task breakdown (ordered)

Parallel work starts after step 0. Each step names the agent, the deliverable and the check that proves it is done.

**Step 0. Decisions (project lead, product owner).** Answer section 9, questions 1 to 3, and record them in this file. Done when the decisions are written here.

**Step 1 (parallel, no dependency).**

- 1a. `auditsphere-backend` (auth): B1 and B2. Add `employee_id` to `UserInfo` and the `department_id` decision. Done when `GET /api/v1/auth/me` returns `employee_id` for a seeded user with a linked employee, and the auth-service tests pass.
- 1b. `auditsphere-solution-architect`: review the schema in B3a, B6, B7, B8 and B9, and decide the `AuditFinding` question. Done when the review note lists the final columns and the migration order, with additive-only migrations.
- 1c. `auditsphere-frontend`: FE-G1, FE-G2, FE-G3 (shell only), FE-G8, FE-G11 and FE-G14. Add `isAuditeeView`, the page branch, the early-return on admin fetches, the empty `AuditeeDashboard.vue` shell with the widget slots, and the removal of hard-coded trends. Done when an AUDITEE login shows the shell and makes no admin requests.
- 1d. `auditsphere-software-translation`: move the strings in FE-G9 to i18n and add the keys for the new widgets, in both locales. Done when locale parity passes (section 8, check 6).

**Step 2 (parallel, after 1a and 1b).**

- 2a. `auditsphere-backend` (audit): B3a (member IDs on letters and resource auditors), with a backfill plan for existing letters. Done when a letter can be saved with a member employee ID and read back.
- 2b. `auditsphere-backend` (audit): B6 (PBC fields and respond endpoint), B7 (comments table and endpoints). Done when the endpoints pass CRUD tests and the existing fieldwork pages still load.
- 2c. `auditsphere-frontend`: W1, W2, W3, W4, W5 and W8 against the contract in section 5, using the interim sources in section 3 until each endpoint lands. Relabel FE-G7. Done when each widget has the four states in section 7.
- 2d. `auditsphere-frontend`: W12 (auditee summary) against the existing ATR list. Done when an AUDITEE login sees only its own ATRs.

**Step 3 (after 2a and 2b).**

- 3a. `auditsphere-backend` (audit): B3 (`/me/engagements`, `/me/dashboard-summary`) and B11 (progress rule). Done when the auditor and auditee RBAC checks in section 8, checks 1 and 2, pass against a seeded database.
- 3b. `auditsphere-backend` (audit): B4 (`/me/calendar`) and B5 (`/me/tasks`). Done when both match the contract in section 5 on seeded data.
- 3c. `auditsphere-backend` (audit): B9 (notifications store and endpoints, then producers) and B10 (due-soon job). Depends on 2b for PBC and comments. Done when an ATR submit creates a notification for the reviewer and the reminder job creates one notification per item per day.
- 3d. `auditsphere-backend` (audit): B8 (LHA finding revisions). Depends on the decision in 1b. Done when a revision can be requested and answered, and it creates a notification (3c).

**Step 4 (frontend wiring, parallel with step 3 once each contract exists).**

- 4a. `auditsphere-frontend`: wire W1, W3, W4 and W5 to 3a and 3b. Remove the "not connected" text. Relabel W4 to "My engagements" only when the scoped endpoint is live.
- 4b. `auditsphere-frontend`: notification bell and panel (FE-G5, after 3c), and the W8 revision list (after 3d).
- 4c. `auditsphere-frontend`: W9 with real PBC status (FE-G12, after 2b), and comment UI on ATR and PBC items (FE-G13, after 2b).

**Step 5 (verification and release).**

- 5a. `auditsphere-qa`: run the checks in section 8 on the staging stack. Report pass or fail with evidence.
- 5b. `auditsphere-devops`: deploy. Backend changes run migrations at `serve` start, as the activity-plan rework does. Production deploys go through `backend/deploy-prod.sh --ref`, not the seeders. The frontend has a manual prod deploy path, so follow the team runbook. Do not redeploy `docker-compose.prod.yml` as is, because it runs seeders on every start. Name the rollback for each service.
- 5c. `auditsphere-project-lead` (this file): update section 2 and the status line at the top after each step. Update `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md` for the new routes once they exist.

### Hand-off summary

| Agent | Steps | Depends on | Done when |
|---|---|---|---|
| auditsphere-backend (auth) | 1a | Step 0 | `/auth/me` returns `employee_id` |
| auditsphere-solution-architect | 1b | Step 0 | Schema review note with migration order |
| auditsphere-frontend | 1c, 2c, 2d, 4a, 4b, 4c | 1c needs nothing; 2c needs section 5 contract; 4a needs 3a and 3b; 4b needs 3c and 3d; 4c needs 2b | Each widget has four states; AUDITEE shell makes no admin requests |
| auditsphere-software-translation | 1d | Step 0 | Locale parity passes |
| auditsphere-backend (audit) | 2a, 2b, 3a, 3b, 3c, 3d | 2a needs 1a; 2b needs 1b; 3a needs 2a; 3b needs 2b; 3c needs 2b; 3d needs 1b | Section 8 checks pass on seeded data |
| auditsphere-qa | 5a | All | Section 8 passes on staging |
| auditsphere-devops | 5b | 5a | Deploy and rollback are documented |
