# Audit Calendar (Kalender Audit): implementation plan

Verified against the code on 2026-10-27 (branch `staging`, HEAD `59dbf5a`, with uncommitted edits to `AuditorDashboard.vue` and `pages/dashboard/index.vue`; line numbers in those two files are from the working tree). Re-check section 2 before relying on it.

Revision 4, 2026-10-27, approved by the user. Product-owner decisions are in section 8. The revisions so far:
- **Revision 2:** Kickoff and Finalization milestones, alerts, the auditee view, and the title pattern.
- **Revision 3:** every phase date is entered by hand; phase-date entry moves into the MVP.
- **Revision 4:**
  - LHA Publication is the third milestone.
  - Alerts count calendar days.
  - Managers get email only for items on hold that need a manager (section 4.8).
  - The Surat Tugas publish gate and the publish-exception request (section 3.5).
  - The publish-role mismatch is recorded as a prerequisite (CAL-B18).
  - Q6 to Q9 are closed.

Owner: project lead. This plan is the detailed design of item **B4** (`GET /api/v1/me/calendar`) and widget **W1** (calendar of important dates) in `docs/auditor-auditee-dashboard-spec.md`, and of the calendar part of **B9/B10** (notifications and the due-soon job). It does not replace that spec: IDs B1 to B13, W0 to W12 and FE-G1 to FE-G14 refer to it. New IDs use the prefix `CAL-`.

Planning only. No application code was changed.

---

## 1. Summary

The calendar shows the audit schedule **by phase** to auditors, managers, the audited unit and viewers. Three milestones are highlighted and trigger alerts: **Audit Kickoff** (entry meeting), **Audit Finalization** (exit meeting) and **LHA Publication**.

**Phase dates are manual.** The Surat Tugas team leader enters every phase date by hand, and the CAE, audit manager and admin can override them. The system never proposes, defaults or derives a phase date. A phase with no date entered does not appear on the calendar, and the gap is shown to the team leader and managers.

**Publishing a Surat Tugas is gated.** A letter cannot be published until Kickoff and Finalization are entered. The exceptions are:
- a manager explicitly confirms publication without them, giving a reason; or
- the team leader requests an exception and a manager approves it.

The request puts the letter on hold, and the managers are emailed.

What the code gives us today:

- **Record dates exist for part of the lifecycle.** These are dates on records, not phase dates:
  - the Surat Tugas letter date and engagement period
  - fieldwork interview, observation and sample dates (planned)
  - the PBC required date and the test-of-controls due date
  - the LHA report and signature dates
  - the ATR due date, with derived overdue days
- **No phase date is stored anywhere.** `audit_executions` knows the current stage but stores no dates.
- **Most stored record dates are strings** (`varchar`), so the server parses them.
- **`audit_activities.planned_start/end` is not a usable source.** The RKA form never writes those columns (`backend/audit-service/services/activity_plan_form/save.go:622-641`).
- **Alerts:** email exists (Resend, `backend/audit-service/pkg/mailer`), but only as the manual ATR reminder. **No in-app notification exists.**
- **Approval holds that exist in code:** PKAT `Pending Review` and `Pending Approval`. The LHA and the KKA have **no** "waiting for approval" state, and ATR review is not manager-only (section 2.7).
- **Publishing a Surat Tugas is a manager action today.** The frontend and backend disagree on which roles may do it (section 2.1). The fix is a prerequisite, and `auditsphere-backend` is doing it in parallel (CAL-B18).

Recommended shape:

1. **Phase 1 (MVP):**
   - The phase-schedule table and its entry form, the "Jadwal tahapan" on the Surat Tugas.
   - The publish gate with the exception request and its manager email.
   - The missing-dates indicator.
   - The read-only calendar endpoints `GET /api/v1/me/calendar` and `GET /api/v1/calendar`.
   - Team scoping by employee ID (B3a).
   - The `/calendar` page and the dashboard widget.
2. **Phase 1b (alerts and manager holds), the next release:**
   - In-app notifications and a header bell.
   - H-3 and H-1 alerts for the three milestones, plus the LHA-published alert.
   - Email through Resend.
   - Manager hold emails for the PKAT, and for a new LHA "Pending Approval" state.
   - A daily reminder for holds still waiting.
3. **Phase 2:** actual dates per phase (entered by hand), the planned-vs-actual view, and typed letter dates and links.
4. **Phases 3 and 4:** drag-to-reschedule, `.ics` export, and reminders for the non-milestone events.

---

## 2. Current state (evidence)

### 2.1 Prerequisites and related code

| Item | Status on 2026-10-27 | Evidence |
|---|---|---|
| B1: employee link exposed to the caller | **Done** (commit `2baf95d`). `/auth/me` returns `employee_code`, `employee_id`, `department_id`. The JWT carries `employee_id` and `employee_code`, but **not** `department_id`. | `backend/auth-service/models/dto.go:39-48`; `backend/auth-service/pkg/utils/jwt.go:13-18`; `backend/audit-service/pkg/middleware/auth.go:12-21, 75-79` |
| audit-service helper for the caller's employee | Exists, but nothing uses it yet. | `backend/audit-service/pkg/base/base_controller.go:79-88` (`GetEmployeeID`) |
| B3a: stable ID on Surat Tugas team members | **Open, and a hard MVP prerequisite** (members see the engagement; the leader is identified for editing and publish requests). `LetterMember` is `{name, role}`. | `backend/audit-service/models/assignment-letter.go:13-16` |
| B2: department key on users | Open. `users.department` is free text. | `backend/auth-service/models/user.go:19` |
| B3: `/me/engagements` | Open. No `/me` route in audit-service. | `backend/audit-service/routes/handler.go:74-514` |
| FE-G1/FE-G3: auditee dashboard | **Done** (commit `818f94e`), with a calendar placeholder. | `frontend/components/dashboard/AuditeeDashboard.vue:348-353` |
| Server-side row scope pattern | ATR only (PIC sees own rows; `view_all_action_taken_report` sees all). Role aliases: `EXECUTIVE` = CAE. | `backend/audit-service/services/action_taken_report/rbac.go:21-191` |
| **Who may publish a Surat Tugas** | Backend: `RequireRolesForStatusChange(..., "Draft", "ADMIN", "AUDIT_MANAGER", "CHIEF_AUDIT_EXECUTIVE")`. Every other role can create letters only in `Draft` and cannot change the status. The roles are compared by exact name, with no aliases. Frontend: `canManageAssignmentLetter` = ADMIN or EXECUTIVE. | `backend/audit-service/routes/handler.go:183-191`; `backend/audit-service/pkg/middleware/status_guard.go:41-95`; `frontend/composables/useRbac.ts:219` |
| **Publish-role mismatch: prerequisite CAL-B18** | The seeded CAE role `EXECUTIVE` sees the publish button (`useRbac.ts:219`) but gets 403 from the backend, which expects `CHIEF_AUDIT_EXECUTIVE` or `AUDIT_MANAGER` (`handler.go:185-186`). `AUDIT_MANAGER` and `CHIEF_AUDIT_EXECUTIVE` are not seeded, and they would pass the backend without seeing the button. In practice only ADMIN can publish. **`auditsphere-backend` is fixing this in parallel.** This plan does not design the fix. The publish gate (3.5) and every "manager" check in this plan assume the role set that fix settles on. | as above |
| Who may publish an LHA | Frontend only: `canPublishLha` = ADMIN, CHIEF_AUDIT_EXECUTIVE or AUDIT_MANAGER, hidden in the auditor view. **There is no backend role check on `/audit-result-reports`.** Approving (`Final`, `Approved` or `Published`) creates one ATR per finding. | `frontend/composables/useRbac.ts:310-317`; `backend/audit-service/routes/handler.go:442-455` |

### 2.2 How engagements are linked today

- **The Surat Tugas is the engagement.** Every live letter has one `audit_executions` row with `ref` = letter number. The execution is created on save and soft-deleted with the letter (`services/audit_execution_sync/letter_link.go:14-28, 117, 269`). Any publish path added by this plan must go through the same write path (`ctrlAuditExecution.SyncLetterOnWrite`, `handler.go:190`), so the execution stays in step.
- **The team leader** is the letter's `leader` (free text), else the member whose role is Chairperson (`letter_link.go:57-90`), else the linked RKA row's `team_leader_id`. Only the RKA key is an ID today.
- **Execution → RKA row** (`audit_executions.activity_id`) is set by the progress sync: an explicit link first, then a case-insensitive title or project-code match (`services/audit_execution_sync/sync.go:227-266`). The link is fuzzy and often missing.
- **The RKA row holds** `project_code` (the Activity ID, for example `ASR-2026-003`), `team_leader_id` and `target_unit_id` (`models/audit-activity.go:157-187`).
- **The letter has no department key and no RKA key** (`models/assignment-letter.go:22-31`).
- Fieldwork, KKA and LHA rows point to the letter through a `varchar` `assignment_letter_id` (the letter number or the UUID). ATR points to the LHA by foreign key.

### 2.3 The dashboard stepper is not date-based

The fieldwork stepper takes its stage from which records exist (`frontend/components/dashboard/AuditorDashboard.vue:1365-1390`). The letter period only drives the deadline badge (`:1835-1850`). The calendar does not take dates from the stepper.

### 2.4 Notifications and email (what exists to reuse)

| Capability | What exists | Evidence | Reuse |
|---|---|---|---|
| Email sender | Resend REST client: `RESEND_API_KEY`, `RESEND_FROM`, plain text, a 12 s timeout. Returns `ErrNotConfigured` when the key is unset. | `backend/audit-service/pkg/mailer/mailer.go:18-89` | Yes, as is. |
| Recipient lookup | `authclient.Contacts` calls auth `GET /api/v1/users/contacts` **with the caller's forwarded JWT**. Filters: `ids`, or `department_head_of=<department name>`. Allowed for the assignable-user roles, which include AUDITOR. | `backend/audit-service/pkg/authclient/client.go:109-175`; `backend/auth-service/routes/registry.go:135-150` | **Event-driven emails** (a user action, so a token exists) can use it once it gains `roles` and `department_id` filters (CAL-B20). **Background jobs** cannot (no token), so they use CAL-B15. |
| Reminder flow | Manual `POST /action-taken-reports/:id/reminder` emails the PIC and the department heads, with per-recipient skip reasons. | `backend/audit-service/controllers/action_taken_report/reminder.go:63-245` | Copy the message and skip-reason style. |
| In-app notifications | **None.** | dashboard spec 2.5; `frontend/layouts/default.vue` | Build the calendar subset of B9 (CAL-B13). |
| Periodic job pattern | `AuditCompletionAnalyzer` ticker loop, started in `serve`. | `backend/audit-service/services/audit_completion/analyzer.go:60-100` | Same pattern for CAL-B14. |

### 2.5 Frontend stack

- `@nuxt/ui` 4.4.0 (`UCalendar` with a `day` slot, `UInputDate`, `UTimeline`; no scheduler), `date-fns` 4.1, `vue-i18n` 11.
- `frontend/components/shared/AppDatePicker.vue` builds a month grid with date-fns.
- Deep links: only ATR (`?lha=`) and working paper (`?tab=&step=`) read query parameters.
- i18n: `dashboard.auditor.calendar.*` and `dashboard.auditee.calendar.*` exist (placeholders). The fieldwork date labels are `auditFieldwork.interview.date`, `.observation.date` and `.sample.date` (`frontend/components/audit-fieldwork/AuditFieldwork{Interview,Observation,Sample}Modal.vue:199/143/140`).

### 2.6 Kong

There are four configs: `backend/kong-gateway/kong/dev/kong-dev.yml`, `backend/kong-gateway/kong/prod/kong.yml`, `backend/tenants/bai/kong/kong.yml` and `scripts/templates/kong-tenant.yml.tpl`. None has `/api/v1/me`, `/api/v1/calendar` or `/api/v1/notifications`. A missing prefix returns 404 through Kong only.

### 2.7 Approval and review holds in the code (input to the manager email rule)

| Item | Hold state in the code | Who can release it | Manager-only? | Evidence |
|---|---|---|---|---|
| PKAT (annual plan) | `Pending Review`: the manager approves (→ `Pending Approval`) or rejects (→ `Revised`). `Pending Approval`: the CAE approves (→ `Approved`). | Audit manager; then CAE | **Yes** | `backend/audit-service/models/annual_plan_status.go:9-13, 67-72`; `frontend/stores/annual-audit.ts:918-939` |
| LHA | **None.** `Draft` → `Final` / `Approved` / `Published` in one step. There is no "submitted for approval" status. | Managers (frontend only) | Would be, but there is no state to detect | `useRbac.ts:310-317`; `handler.go:442-455` |
| KKA (F01 to F05) | **None.** `working_paper_headers` and the form tables have no review status. Only the legacy `working_papers` table has `status` and `reviewed_by_id`. | n/a | n/a | `backend/audit-service/models/working-paper.go:24-35, 245-262` |
| ATR | `PENDING_REVIEW` | Any role with `review_action_taken_report`: CAE, audit manager **and AUDITOR** | **No** | `backend/audit-service/models/action-taken-report-status.go:18-22`; `services/action_taken_report/rbac.go:28-29, 58-60, 191-193` |
| Executive summary | `Draft` / `Approved` / `Rejected`. The approve endpoint is ADMIN and CAE only, but there is no "submitted" state, so a Draft that is still being edited looks the same as one that is waiting. | ADMIN, CAE | Yes, but no hold state | `backend/audit-service/models/executive-summary.go:18`; `controllers/executive_summary/approve.go:35-41` |
| Surat Tugas publish exception | **New** (3.5) | ADMIN, CAE, audit manager | **Yes** | this plan |

---

## 3. Phases, milestones and events

### 3.1 Phase vocabulary

Phases are aligned with the `audit_executions` canonical stages (`models/audit-execution.go:11-18`). Every event carries an icon and a text label, never colour alone. Dark-mode variants are required.

| Phase key | Display label (EN / ID) | Execution stage | Colour | Icon |
|---|---|---|---|---|
| `PLANNING` | Audit Planning / Perencanaan Audit | planning | slate | `i-lucide-clipboard-list` |
| `ASSIGNMENT` | Surat Tugas / Surat Tugas | planning | indigo | `i-lucide-file-signature` |
| `ENTRY_MEETING` | Audit Kickoff / Kickoff Audit (rapat pembukaan) | entry meeting | sky | `i-lucide-flag` (milestone) |
| `FIELDWORK` | Audit Fieldwork / Pekerjaan Lapangan Audit | fieldwork | emerald | `i-lucide-search-check` |
| `WORKING_PAPER` | KKA Review / Reviu KKA | draft findings | teal | `i-lucide-file-check` |
| `EXIT_MEETING` | Audit Finalization / Finalisasi Audit (rapat penutupan) | reporting & exit meeting | violet | `i-lucide-flag-triangle-right` (milestone) |
| `REPORTING` | Audit Reporting (LHA) / Pelaporan Audit (LHA) | reporting, completed | amber | `i-lucide-file-text`; LHA publish: `i-lucide-file-badge` (milestone) |
| `FOLLOW_UP` | Follow-up (ATR) / Tindak Lanjut (ATR) | after completed | rose (overdue: red) | `i-lucide-list-checks` |

Final wording belongs to `auditsphere-software-translation`.

### 3.2 Phase dates: entered by hand (MVP)

The team leader enters these on the Surat Tugas "Jadwal tahapan" (section 5.4). Each row is optional, and **a row with no date produces no calendar event**.

| Schedule row (`phase`) | Calendar event type | Shape | Milestone | Notes |
|---|---|---|---|---|
| `ENTRY_MEETING` | `AUDIT_KICKOFF` | day (+ optional location) | **KICKOFF** | Required to publish (3.5) |
| `FIELDWORK` | `PHASE_FIELDWORK` | range | no | The team's planned fieldwork window |
| `WORKING_PAPER` | `KKA_REVIEW_DUE` | day or range | no | Review chain: auditor → team leader → QA |
| `EXIT_MEETING` | `AUDIT_FINALIZATION` | day (+ optional location) | **FINALIZATION** | Required to publish (3.5) |
| `LHA_DRAFT` | `LHA_DRAFT_DUE` | day | no | |
| `LHA_REVIEW` | `LHA_REVIEW_DUE` | day | no | |
| `LHA_PUBLISH` | `LHA_PUBLICATION` | day | **LHA_PUBLICATION** | The planned publication date. Not required to publish the Surat Tugas (it is often unknown at that point), but it is in the missing-dates indicator. |

Rules:

- **Who enters dates:** the Surat Tugas team leader. **The CAE (seeded as `EXECUTIVE`), the audit manager and the admin can override any phase date.** Every save records `updated_by_user_id` and `updated_by_name`, and the slideover shows "Set by <name>, <date>". An override notifies the team leader (Phase 1b, `PHASE_DATE_CHANGED`).
- **No defaults:** the form opens empty, nothing is copied from the period, RKA or PKAT, and no date is derived from records.
- **Validation:** start ≤ end within a row, and Kickoff ≤ Finalization ≤ LHA publication (errors). A date outside the Surat Tugas period, or an LHA step before Finalization, gives a warning only.
- **Actual dates** (Phase 2) are entered by hand. In the MVP only Kickoff and Finalization have "Mark as held", which cancels their pending alerts. For LHA Publication, the actual publication is a record fact (3.3, `LHA_PUBLISHED`), not a phase date: it stops the planned-date alerts but writes nothing to the phase row.

### 3.3 Record events (existing dates, shown as they are)

These come from the records and need no entry. "Parse" means the field is a string, parsed as `YYYY-MM-DD` and then the formats in `assignment-letter.go:79-86`.

| Event type | Phase | Source table.column | Stored as | Basis | Shape | MVP? | Auditee / dept head | VIEWER |
|---|---|---|---|---|---|---|---|---|
| `LETTER_ISSUED` | ASSIGNMENT | `assignment_letters.letter_date` | timestamp | actual | day | Yes | No | No |
| `ENGAGEMENT_PERIOD` | FIELDWORK (band) | `assignment_letters.start_period`, `finish_period` | varchar (parse) | planned | range | Yes | Yes | Yes |
| `INTERVIEW` | FIELDWORK | `fieldwork_interviews.date` | varchar (parse) | **planned** | day | Yes | **Yes** (activity only) | No |
| `OBSERVATION` | FIELDWORK | `fieldwork_observations.date` | varchar (parse) | **planned** | day | Yes | **Yes** (activity only) | No |
| `SAMPLE` | FIELDWORK | `fieldwork_samples.date` | varchar (parse) | **planned** | day | Yes, off by default | No | No |
| `PBC_DUE` | FIELDWORK | `fieldwork_documents.required_date` | varchar (parse) | due | day | Yes | Yes | No |
| `TEST_CONTROL_DUE` | FIELDWORK | `fieldwork_test_controls.due_date` | varchar (parse) | due | day | Yes | No | No |
| `LHA_REPORT_DATE` | REPORTING | `audit_result_reports.report_date` | timestamp | actual | day | Yes | Published only | Published only |
| `LHA_SIGNED` | REPORTING | `audit_result_reports.signature_date` | timestamp | actual | day | Yes | Published only | Published only |
| `LHA_PUBLISHED` | REPORTING | `audit_result_reports.published_at` (**new**, CAL-B7), set on the first transition to an approved status (`Final` / `Approved` / `Published`, the same predicate as the ATR sync) | timestamp | actual | day | Phase 1b | **Yes** | Yes |
| `ATR_DUE` | FOLLOW_UP | `action_taken_reports.due_date`; `is_overdue`, `overdue_days` derived | timestamp | due | day | Yes | Yes (own PIC rows) | No |
| `ATR_REVIEWED` | FOLLOW_UP | `action_taken_reports.reviewed_at` | timestamp | actual | day | Yes, off by default | Yes (own) | No |
| `PKAT_MONTH` | PLANNING | `audit_annuals.selected_months`, `activities[].timelineText` | int array / text | planned | month | No (CAL-B9) | No | No |

`audit_assignments.start_date/end_date` stay out of scope. Planned fieldwork events (interviews, observations, samples) are outline chips labelled "Planned", and are never overdue. The form labels change to say so (CAL-FE9, CAL-T1).

### 3.4 The three milestones

| Milestone | Lifecycle step | Indonesian term | Date source | Visible to | Required to publish the Surat Tugas |
|---|---|---|---|---|---|
| **Audit Kickoff** | Entry meeting: formal start with the auditee | Rapat pembukaan | Manual `ENTRY_MEETING` row | Team, managers, auditee unit, VIEWER | **Yes** |
| **Audit Finalization** | Exit meeting: close of fieldwork, findings discussed before the LHA | Rapat penutupan | Manual `EXIT_MEETING` row | Team, managers, auditee unit, VIEWER | **Yes** |
| **LHA Publication** | Publication of the Laporan Hasil Audit | Penerbitan LHA | Planned: manual `LHA_PUBLISH` row. Actual: `LHA_PUBLISHED` record event. | Planned: team, managers, VIEWER (**not** the auditee). Actual: everyone in 4.3, including the auditee. | No |

Why the auditee does not see the planned LHA date: it is an internal target that often slips. Showing it would set an expectation the unit cannot track. The auditee learns of the LHA when it is actually published, which is also when its ATRs are created.

All three milestones get the same highlight (5.2). Their alerts are in 4.7.

### 3.5 Surat Tugas publish gate (decision Q9)

> **Interpretation approved by the user on 2026-10-27.** A Surat Tugas cannot be published without Kickoff and Finalization dates. The team leader is prompted to enter them. If the letter must be published without them, the team leader requests permission, and a manager (CAE, audit manager or admin) approves or rejects it. While the request waits, the letter is on hold and the managers are emailed. A manager who publishes a letter must enter the dates or explicitly confirm publication without them, with a reason.

**One fact shapes the design: team leaders cannot publish today.** Only managers may change a letter's status (2.1). The gate therefore:

- keeps publishing a manager action, with the role set fixed by prerequisite CAL-B18;
- adds a server-side precondition to publishing;
- gives the team leader a way to ask for an exception, since they cannot publish themselves.

Rules (server-enforced in audit-service; the frontend only mirrors them):

1. **Precondition.** Any write that moves a letter into `Published` (`POST` with `status: Published`, or `PUT` from `Draft`) returns 422 `MILESTONE_DATES_REQUIRED` with the missing phases, **unless**:
   - both `ENTRY_MEETING` and `EXIT_MEETING` rows exist; or
   - the letter has an **approved** publish exception; or
   - the caller is a manager and the request carries `confirm_without_milestones: true` and a non-empty `exception_reason`. This records a self-approved exception.

   Letters published before the gate are grandfathered. They are not unpublished, and they show in the missing-dates indicator.
2. **Manager publishing (their own letter or anyone's).** The publish dialog asks for the two dates, which is the normal path. "Publish without milestone dates" needs a checkbox ("I confirm…") and a reason, and it writes an exception record with `kind = MANAGER_CONFIRMED`. Managers do not need another manager's approval, but the action is logged and visible.
3. **Team leader.** The phase schedule shows a "Ready for publication" checklist (Kickoff, Finalization).
   - **With both dates:** nothing new is needed; a manager publishes as today. The letter appears in the managers' "Ready to publish" list (5.5).
   - **Without them:** the team leader can enter them now (the prompt) or click **"Request publication without milestone dates"**, which needs a reason. This creates an exception request with `status = PENDING`.
4. **On hold.** While a request is `PENDING`:
   - The letter stays `Draft`; no new letter status is added, so the dashboards, execution sync and filters that test `Published` are untouched.
   - It shows a "Waiting for manager approval" badge for the team leader, and is listed in the managers' pending list.
   - One pending request per letter.
   - **The managers are emailed at once** (4.8).
5. **Decision** (managers only; not the requester, if they are also a manager):
   - **Approve:** the request becomes `APPROVED`, and the letter is published in the same transaction, through the same write path as `PUT` (`SyncLetterOnWrite`), so the execution sync runs.
   - **Reject:** `REJECTED` with a note. The letter stays `Draft`, and the team leader is told in-app from Phase 1b (in the MVP, by the badge and the note on the letter).
6. **Superseded.** If both dates are entered while a request is pending, the request becomes `SUPERSEDED` and the letter is publishable normally. If the requester withdraws it, it becomes `WITHDRAWN`.
7. **Audit trail.** Every request and decision is stored (4.6, `letter_publish_exceptions`). It records who, when, the reason, the decision note, and which phases were missing. A letter published under an exception shows "Published without milestone dates, approved by <name> on <date>: <reason>" on the letter and in the calendar slideover. It stays in the missing-dates indicator until the dates are entered.

**Placement: MVP.** The gate has to be live before teams start publishing under the new rule, and the request is useless if nobody hears about it. So the MVP includes:
- the precondition and the exception request, approve and reject endpoints
- the team leader and manager UI
- the manager email for a new request, which is event-driven: the request is a user action, so it uses the forwarded token and the `roles` filter from CAL-B20

The daily "still waiting" reminder and the in-app bell entries come with Phase 1b.

### 3.6 Event titles

Pattern: **`<Audit ID> – <Phase> – <Activity>`**. Parts that add nothing are dropped. The server returns `title_parts` and a composed `title` in the request's language.

- **Audit ID:** the linked RKA `project_code` (for example `ASR-2026-003`), with the Surat Tugas number as the fallback (`audit_id_source: "rka" | "letter"`).

| Event | Title |
|---|---|
| `AUDIT_KICKOFF` | `ASR-2026-003 – Audit Kickoff` |
| `INTERVIEW` | `ASR-2026-003 – Audit Fieldwork – Interview` |
| `PBC_DUE` | `ASR-2026-003 – Audit Fieldwork – PBC due: <document name>` |
| `KKA_REVIEW_DUE` | `ASR-2026-003 – KKA Review` |
| `AUDIT_FINALIZATION` | `ASR-2026-003 – Audit Finalization` |
| `LHA_PUBLICATION` (planned) | `ASR-2026-003 – Audit Reporting – LHA publication` |
| `LHA_PUBLISHED` (actual) | `ASR-2026-003 – Audit Reporting – LHA published` |
| `ATR_DUE` | `ASR-2026-003 – Follow-up – ATR due: <finding title>` |

Titles shown to an auditee contain no interviewee, interviewer, topic or notes.

### 3.7 Worked example: the auditee's unit

For `ASR-2026-003` (Surat Tugas period 22 to 31 Oct), the team leader entered Kickoff on 22 Oct, Finalization on 27 Oct and a planned LHA publication on 10 Nov. A manager published the letter (the gate passed). An auditor planned an interview on 25 Oct. The LHA was actually published on 12 Nov.

An AUDITEE or DEPARTMENT_HEAD of Cabang Bandung sees:

| Date | Title | Look | Alerts to the unit (Phase 1b) |
|---|---|---|---|
| 22 Oct | ASR-2026-003 – Audit Kickoff | milestone | H-3 on 19 Oct, H-1 on 21 Oct |
| 22–31 Oct | ASR-2026-003 – Audit Fieldwork period | background band | none |
| 25 Oct | ASR-2026-003 – Audit Fieldwork – Interview | outline, "Planned" | none |
| 27 Oct | ASR-2026-003 – Audit Finalization | milestone | H-3 on 24 Oct, H-1 on 26 Oct |
| 12 Nov | ASR-2026-003 – Audit Reporting – LHA published | milestone (done) | On publication, 12 Nov |

The team and the managers also see the planned LHA publication on 10 Nov (milestone, with H-3 and H-1 alerts on 7 and 9 Nov). Next to it they see the actual publication on 12 Nov, so the two-day slip is visible. A VIEWER sees the period band, the three milestones and the published LHA, and gets no alerts.

---

## 4. Data model, API and alerts

### 4.1 Ownership and shape

- **Owner: audit-service.** All calendar data lives in the audit database. The request path makes no call to master or risk.
- **New tables:**
  - **Phase 1:** `engagement_phase_schedules` (manual phase dates) and `letter_publish_exceptions` (the gate).
  - **Phase 1b:** `notifications`, `notification_reads` and `notification_email_log`; plus `published_at` and `submitted_for_approval_at` on `audit_result_reports`.
  - **Phase 2:** typed letter columns.
- All migrations are additive and run at `serve` start. `auditsphere-solution-architect` reviews them (CAL-A0 for Phases 1 and 1b, CAL-A1 for Phase 2).
- The `requirements.MD` dashboard SLA (2 s at 100 concurrent users) and the email SLA (6 s) are targets, not measurements.

### 4.2 Endpoints

"Managers" below means the publishing role set after CAL-B18: ADMIN, the CAE (seeded `EXECUTIVE` / `CHIEF_AUDIT_EXECUTIVE`) and AUDIT_MANAGER.

| Endpoint | Who | Scope | Phase |
|---|---|---|---|
| `GET /api/v1/me/calendar?from&to&types&phases&important` | Any authenticated user | The caller's items (4.3) | 1 |
| `GET /api/v1/calendar?from&to&types&phases&important&department_id&letter&team_leader_id&overdue&include_draft` | Managers (all events); VIEWER (engagement level only) | Org-wide | 1 |
| `GET /api/v1/assignment-letters/:id/phase-schedule` | Anyone who can see the letter. Includes `can_edit`, the publish checklist and any pending exception. | One engagement | 1 |
| `PUT /api/v1/assignment-letters/:id/phase-schedule` | Team leader; **managers as override** | Upsert or clear rows | 1 |
| `PUT /api/v1/assignment-letters/:id` with `status: Published` (existing route) | Managers (per CAL-B18) | Gated by 3.5. Accepts `confirm_without_milestones` and `exception_reason`. | 1 |
| `POST /api/v1/assignment-letters/:id/publish-exception` | Team leader of the letter | Body `{reason}`. Creates a `PENDING` request and emails the managers. | 1 |
| `POST /api/v1/assignment-letters/:id/publish-exception/:reqId/approve` and `/reject` | Managers (not the requester) | Approve publishes the letter. Reject needs a note. | 1 |
| `DELETE /api/v1/assignment-letters/:id/publish-exception/:reqId` | The requester | Withdraw | 1 |
| `GET /api/v1/me/phase-date-gaps` | Team leaders | Own led letters with missing phase rows, plus own pending or rejected requests | 1 |
| `GET /api/v1/calendar/phase-date-gaps?department_id&team_leader_id` | Managers | Org-wide gaps, **pending publish requests**, and "ready to publish" letters (Draft, both milestones set) | 1 |
| `GET /api/v1/notifications?unread&limit`, `PUT /api/v1/notifications/:id/read`, `PUT /api/v1/notifications/read-all` | Any authenticated user | Audience match (4.7) | 1b |
| `POST /api/v1/audit-result-reports/:id/submit-for-approval` | The letter's team leader or member | Draft → `Pending Approval` (CAL-B16) | 1b |
| `GET /api/v1/users/contacts?roles=&department_id=` (auth, existing route, new filters) | Assignable-user roles (existing) | Event-driven email recipients | 1 (`roles`), 1b (`department_id`) |
| `GET /internal/v1/contacts` (auth) | audit-service only, with a per-tenant service token. **Not routed by Kong.** | Recipients for background-job emails | 1b |
| `PATCH /api/v1/assignment-letters/:id/phase-schedule/:phase` | Write roles | Drag-to-reschedule | 3 |
| `GET /api/v1/calendar/feed/:token.ics` | Revocable per-user feed token | As `/me/calendar` | 4 |

Gap response, per letter: `{assignment_letter_id, letter_number, audit_id, status, team_leader_name, missing: [...], missing_milestones, publish_exception: {id, status, reason, requested_by_name, requested_at} | null, ready_to_publish}`. Letters with execution status `completed` are excluded.

Range rules: `from` and `to` are required, `YYYY-MM-DD`, with a span of at most 100 days (otherwise 400 `CALENDAR_RANGE_TOO_WIDE`). A forbidden scope returns 403 `FORBIDDEN_SCOPE`.

Every letter route stays under `/assignment-letters`, which Kong already routes. The only new prefixes are the three in CAL-D1. `submit-for-approval` sits under `/audit-result-reports`, which is already routed.

### 4.3 RBAC scoping (server-side, per the ATR pattern)

Move the role canonicalisation (`CanonicalRole`, the aliases, `IsAdmin`) from `services/action_taken_report/rbac.go` into a shared audit-service package, so the calendar, phase schedule, publish gate, LHA guard (CAL-B16) and notifications all use one mapping. Align it with whatever role mapping the CAL-B18 fix introduces, rather than adding a second one.

| Role (canonical) | `/me/calendar` | `/calendar` | Phase schedule write | Publish / decide exceptions | Gaps |
|---|---|---|---|---|---|
| ADMIN, CHIEF_AUDIT_EXECUTIVE (incl. seeded `EXECUTIVE`), AUDIT_MANAGER | Engagements they lead or belong to, plus own ATRs | **All events, including Draft letters** (`is_draft: true`) | **Override on any letter** | **Yes** | Org-wide, plus pending requests and ready to publish |
| AUDITOR, as team leader | Published letters where leader or member (by employee ID), plus own ATRs | 403 | Own led letters | Request an exception only | Own led letters |
| AUDITOR, as member | Same as leader | 403 | No | No | No |
| AUDITEE | Own ATRs. For Published letters of the caller's unit: period, Kickoff, Finalization, interviews and observations (activity only), PBC, `LHA_PUBLISHED` and the published LHA dates. **Never** samples, test controls, KKA, other phase rows, the planned LHA publication, gaps or exceptions. | 403 | No | No | No |
| DEPARTMENT_HEAD | **Same as AUDITEE**, plus the department's ATRs (read only) | 403 | No | No | No |
| VIEWER | Empty | **Org-wide, engagement level only:** period, the three milestones (planned LHA publication included), `LHA_PUBLISHED`, for Published letters. **Never** fieldwork items, PBC, KKA, ATR, gaps or drafts. | No | No | No |

Notes:

- **Team leader by ID:** a member with the Chairperson role and an `employee_id` (B3a), else the linked RKA `team_leader_id`. If the leader cannot be resolved, only managers can edit or publish, and the gap response flags it (`leader_unresolved: true`).
- **The unit of a letter:** the linked RKA row's `target_unit_id`, then the letter's own `target_unit_id` once CAL-B8 adds it. A letter with neither is shown to managers only (`meta.unscoped_letters`).
- **The caller's department** comes from the `department_id` JWT claim (CAL-B1). Without it, an auditee gets ATR items only (`meta.unit_scope: "unavailable"`).
- **Draft letters** never reach auditors, auditees, department heads or viewers, and never trigger milestone alerts.

### 4.4 Tenant isolation

- **Kong tenant template:** update `scripts/templates/kong-tenant.yml.tpl` with the other three configs, or new tenants get 404s.
- **The service token for `/internal/v1/contacts`** must be **per tenant**. `jwt.secret` is identical across tenants today, and the service token must not repeat that.
  - Call auth by the tenant's slugged service name; all tenants share one Docker network.
- **Alert dedupe state lives in Postgres, not Redis.** Redis is shared and its keys are unprefixed.
- **Manager email recipients are the managers of the same tenant only.** Event-driven emails resolve them through the tenant's own auth-service with the caller's token. Job emails use the tenant's internal endpoint.
- **The `.ics` feed token** (Phase 4) is stored per tenant database and is revocable.

### 4.5 Event payload (extends the B4 contract in the dashboard spec, section 5)

The frontend has not consumed B4 yet, so the contract can still change. `id`, `type`, `date`, `title`, `entity_type`, `entity_id`, `status` and `is_overdue` keep their meaning.

```json
{
  "range": { "from": "2026-10-01", "to": "2026-10-31" },
  "scope": "me",
  "events": [
    {
      "id": "phase:9f20...:ENTRY_MEETING",
      "type": "AUDIT_KICKOFF",
      "phase": "ENTRY_MEETING",
      "basis": "planned",
      "important": true,
      "milestone": "KICKOFF",
      "done": false,
      "date": "2026-10-22",
      "start_date": "2026-10-22",
      "end_date": "2026-10-22",
      "precision": "day",
      "title": "ASR-2026-003 – Audit Kickoff",
      "title_parts": { "audit_id": "ASR-2026-003", "phase": "ENTRY_MEETING", "activity": null, "subject": null },
      "audit_id": "ASR-2026-003",
      "audit_id_source": "rka",
      "letter_number": "ST-004/SKAI/2026",
      "assignment_letter_id": "5b1e...",
      "entity_type": "phase_schedule",
      "entity_id": "9f20...",
      "unit_id": "c0d2...",
      "unit_name": "Cabang Bandung",
      "location": "Ruang rapat lt. 3",
      "team_leader_name": "string",
      "set_by_name": "string",
      "my_role": "UNIT",
      "status": "Published",
      "is_draft": false,
      "published_with_exception": false,
      "is_overdue": false,
      "overdue_days": 0
    }
  ],
  "meta": {
    "employee_linked": true,
    "unit_scope": "ok",
    "unparsed_dates": 0,
    "unscoped_letters": 0,
    "unlinked_members": 0,
    "missing_phase_dates": 3,
    "pending_publish_requests": 1
  }
}
```

Field values:

- `milestone`: `KICKOFF | FINALIZATION | LHA_PUBLICATION | null`. `important` is true when `milestone` is set, and also for `LHA_PUBLISHED`.
- `done`: true for a Kickoff or Finalization marked held, and for `LHA_PUBLISHED`.
- `published_with_exception`: true when the letter was published without milestone dates (3.5).
- `basis`: `planned | actual | due`. `precision`: `day | month`.
- `entity_type`: `assignment_letter | phase_schedule | fieldwork_interview | fieldwork_observation | fieldwork_sample | fieldwork_document | fieldwork_test_control | audit_result_report | action_taken_report`.
- `my_role`: `LEADER | MEMBER | PIC | REVIEWER | UNIT | null`.
- `meta.missing_phase_dates` and `meta.pending_publish_requests`: only for roles that can see gaps; `null` otherwise.
- `important=true` as a query parameter returns milestones only.

The backend returns entity references, not routes. The frontend maps entities to pages (CAL-FE5).

### 4.6 Schema (for `auditsphere-solution-architect`)

- **CAL-B4 `engagement_phase_schedules`** (Phase 1):
  - Columns: `id uuid`; `assignment_letter_id uuid not null` (FK, soft-deleted with the letter through the service); `phase varchar(30)` (`ENTRY_MEETING | FIELDWORK | WORKING_PAPER | EXIT_MEETING | LHA_DRAFT | LHA_REVIEW | LHA_PUBLISH`); `planned_start date not null`; `planned_end date null`; `actual_start date null`; `actual_end date null`; `location varchar(255) null`; `minutes_media_id uuid null`; `updated_by_user_id`; `updated_by_name`; `created_at`; `updated_at`; `deleted_at`.
  - Unique `(assignment_letter_id, phase)` among live rows. A row exists only when a date was entered.
- **CAL-B17 `letter_publish_exceptions`** (Phase 1):
  - Columns: `id uuid`; `assignment_letter_id uuid not null`; `kind varchar(20)` (`REQUEST | MANAGER_CONFIRMED`); `status varchar(20)` (`PENDING | APPROVED | REJECTED | WITHDRAWN | SUPERSEDED`); `reason text not null`; `missing_phases jsonb`; `requested_by_user_id`; `requested_by_name`; `requested_at`; `decided_by_user_id null`; `decided_by_name null`; `decided_at null`; `decision_note text null`; `email_sent_at null`; `email_error text null`; `created_at`; `updated_at`.
  - Partial unique index on `(assignment_letter_id) WHERE status = 'PENDING'`. Rows are never deleted.
- **CAL-B7 LHA timestamps** (Phase 1b): `submitted_for_approval_at` and `published_at` on `audit_result_reports`.
  - `published_at` is set once, on the first transition to `Final`, `Approved` or `Published`. These are record facts, not phase dates.
- **CAL-B16 LHA `Pending Approval` status** (Phase 1b):
  - A new LHA status value between `Draft` and the approved statuses, plus a backend role guard on `/audit-result-reports`. Only managers may move a report to `Final`, `Approved` or `Published`, or send it back to `Draft` from `Pending Approval`.
  - This closes the frontend-only check in 2.1.
  - The ATR sync predicate is unchanged: `Pending Approval` is not approved.
- **CAL-B8 typed letter dates and links** (Phase 2): `start_date`, `finish_date`, `target_unit_id`, `activity_id` on `assignment_letters`, with a backfill (dry run first).
- **CAL-B9 PKAT and RKA planned schedule** (Phase 2): a typed planned month on `AnnualAuditActivity`, and the RKA form writing `audit_activities.planned_start/end`.
- **CAL-B13 notifications** (Phase 1b): see 4.7.

### 4.7 Milestone alerts (Phase 1b)

Timing counts **calendar days** (decision Q7), in the tenant timezone (default `Asia/Jakarta`), at 08:00. Offsets and hour are per-tenant settings (`CALENDAR_ALERT_OFFSETS_DAYS=3,1`, `CALENDAR_ALERT_HOUR=8`). Alerts fire only on **entered** dates of **Published** letters, never for dates in the past, and never to VIEWER.

| Trigger | Milestone | When | Team (leader and members) | Auditee unit (AUDITEE, DEPARTMENT_HEAD) | Managers |
|---|---|---|---|---|---|
| `MILESTONE_DUE_SOON` | Kickoff, Finalization | H-3, H-1 | in-app + email | in-app + email | in-app |
| `MILESTONE_DUE_SOON` | LHA Publication (planned) | H-3, H-1, until the LHA is actually published | in-app + email | — (not visible to the unit) | in-app |
| `LHA_PUBLISHED` | LHA Publication (actual) | At the first transition to an approved status (event-driven) | in-app | **in-app + email** (ATRs are created at this moment) | in-app |
| `MILESTONE_RESCHEDULED` | any of the three | When a date that already had an alert sent moves, or a new date is entered within 3 days | in-app + email | in-app + email (Kickoff and Finalization only) | in-app |
| `MILESTONE_DATE_MISSING` | Kickoff, Finalization | 7 days after publication while still missing (only for letters published with an exception) | leader, in-app | — | in-app (gaps list) |
| `PHASE_DATE_CHANGED` | any phase row | When a manager overrides a date | leader, in-app | — | — |
| `PUBLISH_EXCEPTION_DECIDED` | n/a | On approve or reject | leader, in-app | — | — |

The alerts stop when:
- **Kickoff or Finalization** is marked held.
- **The planned LHA Publication** is matched by an actual publication. This uses the record fact; no phase date is written.

**In-app (CAL-B13, CAL-FE10).** Notifications are **audience-addressed**, because audit-service has no user directory:
- `notifications`: `id`, `audience_kind` (`USER | EMPLOYEE | DEPARTMENT | ROLE`), `audience_key`, `type`, `title_parts json`, `entity_type`, `entity_id`, `assignment_letter_id`, `event_date`, `offset_days`, `dedupe_key` (unique), `created_at`.
- `notification_reads`: `notification_id`, `user_id`, `read_at`, with unique `(notification_id, user_id)`.
- `GET /api/v1/notifications` matches the caller's JWT `user_id`, `employee_id`, `department_id` and canonical roles. `DEPARTMENT` audiences are shown only to AUDITEE and DEPARTMENT_HEAD.
- This differs from the dashboard spec's B9 sketch (`recipient_user_id`); the architect confirms in CAL-A0.

**Email.** All email uses `pkg/mailer.Resend.Send`, plain text, Indonesian and English in one message, behind `CALENDAR_EMAIL_ALERTS=true`. Every send is logged per recipient in `notification_email_log` (`notification_id`, `email`, `status`, `error`, `sent_at`, unique `(notification_id, email)`), with up to 3 retries. When `RESEND_API_KEY` is unset, in-app alerts still go out and the skipped emails are logged.
- **Event-driven** sends (`LHA_PUBLISHED`, `MILESTONE_RESCHEDULED`, manager holds in 4.8) happen right after the user's write commits. They resolve recipients through `/users/contacts` with the caller's forwarded token, using the new `roles` and `department_id` filters (CAL-B20).
- **Scheduled** sends (H-3, H-1, hold reminders) come from the job and resolve recipients through `/internal/v1/contacts` (CAL-B15).

**Job (CAL-B14):** `services/milestone_alerts`, a ticker loop like `AuditCompletionAnalyzer`, every 15 minutes. It selects the entered milestone dates in the next 3 days and the holds older than 1 day, then writes notifications with `dedupe_key = <entity>:<type>:<offset>:<event_date>`. This is idempotent across restarts, and a moved date re-arms the alerts.

### 4.8 Manager email: only for items on hold that need a manager (decision Q8)

Managers get **in-app only** for everything in 4.7. The **one exception**: an email when something is on hold and cannot continue until a manager acts.

From the code (2.7), the cases are:

| # | Case | Hold state | Email to | Trigger | Phase |
|---|---|---|---|---|---|
| H1 | **Surat Tugas publication without milestone dates requested** | `letter_publish_exceptions.status = PENDING` (3.5) | Managers (excluding the requester) | On request (event-driven) | **1 (MVP)** |
| H2 | **PKAT waiting for manager review** | `audit_annuals.status = "Pending Review"` | AUDIT_MANAGER (ADMIN as fallback when the tenant has no audit manager) | On the transition into the state | 1b |
| H3 | **PKAT waiting for CAE approval** | `audit_annuals.status = "Pending Approval"` | CAE (`EXECUTIVE` / `CHIEF_AUDIT_EXECUTIVE`; ADMIN as fallback) | On the transition | 1b |
| H4 | **LHA waiting for approval** | `Pending Approval` (**new**, CAL-B16; no such state today) | Managers | On `submit-for-approval` | 1b |
| — | **Still waiting** reminder for H1 to H4 | the same states, open for more than 1 calendar day | the same recipients | Daily at 08:00 (job), until released | 1b |

Not emailed to managers, with the reason from the code:

- **ATR `PENDING_REVIEW`.** Auditors hold `review_action_taken_report` too (`rbac.go:60`), so the item is not waiting on a manager. Reviewers get it in-app (with B9 and B10). If the user wants a manager escalation, for example still pending after N days, it is a one-rule addition to H-type holds.
- **KKA review.** F01 to F05 have no review state. A manager hold for the KKA can only exist after a KKA review status is added (dashboard spec B8, or a KKA review item). This is not in this plan.
- **Executive summary approval.** There is no "submitted" state, so a Draft being edited looks the same as one waiting. Add a submitted state first if this is wanted. Not in this plan.

Each hold email states the item, who is waiting, since when, and a direct link to it: the letter's pending request, the PKAT, or the LHA. A hold leaves the list as soon as it is released, and the reminders stop.

---

## 5. Frontend

### 5.1 Library decision

All events are whole-day, so the MVP views (month, week, agenda) are built in-house with date-fns and Nuxt UI, following `AppDatePicker.vue` (FE-G4). The widget's mini-month uses `UCalendar`. FullCalendar is re-evaluated before Phase 3.

### 5.2 Calendar page (`/calendar`)

- Route `frontend/pages/calendar/index.vue`. A sidebar item under Dashboard (`frontend/layouts/default.vue:84-86`), `navigation.calendar`, `i-lucide-calendar-days`, for every role. The auditor branch (`default.vue:607-630`) must not hide it.
- Views: **Month** (range bars; up to 3 events a day plus "+n more"), **Week**, **Agenda** (the default on small screens).
- **Milestone highlight (all three):** listed first and never collapsed; a thick chip with a phase-colour ring and the milestone icon; a "Milestone" badge; bold text; a "Next milestone" strip. Done milestones (held, or LHA published) show filled with a check mark. A planned LHA publication next to its actual publication shows a slip badge ("+2 days").
- **Draft style (managers only):** dashed border, muted fill, "Draft" badge.
- **Published with exception:** an amber "No milestone dates" marker on the letter's events (team and managers only).
- **Planned style:** an outline chip with "Planned".
- Filters (in the query string): phase, event type (samples off by default), milestones only, unit, team leader, Surat Tugas, overdue, "Mine / All engagements" (managers), and show drafts (managers). VIEWER gets phase, unit and milestones only.
- The slideover shows the details, "Set by", the exception note if any, and an **Open** button. For phase rows, write roles also get "Edit dates", and Kickoff and Finalization get "Mark as held".
- Banners on the page: missing phase dates and pending publish requests (5.5).
- States as in the dashboard spec section 7.
- `frontend/stores/calendar.ts`; `frontend/utils/calendarPhases.ts` (phase map, entity routes, title composer), with unit tests.

### 5.3 Dashboard widget (replaces both placeholders)

`frontend/components/dashboard/CalendarWidget.vue`, in `AuditorDashboard.vue` (section `:590-626`) and `AuditeeDashboard.vue` (`:348-353`). It shows the next milestone, a `UCalendar` mini-month (phase dots, a flag on milestone days), the next 14 days, an overdue count and an "Open calendar" link. For team leaders it also shows "**n engagements with missing phase dates**" and any of their publish requests that are pending or rejected.

A manager card on the admin dashboard shows: missing phase dates (n), **publish requests waiting for you (n)** and ready to publish (n), each linking to `/calendar?gaps=1`. Remove the old placeholder keys once wired.

### 5.4 Phase schedule and publication on the Surat Tugas (CAL-FE4, CAL-FE12)

- **"Jadwal tahapan"** in the Surat Tugas detail (`frontend/pages/assignment-letter/index.vue` and its components), reachable at `?letter=<number>&tab=schedule`.
  - One row per phase, with the three milestones first.
  - All fields start empty. "Set by" is shown, override is labelled, and validation follows 3.2.
- **"Ready for publication" checklist** at the top of the section: Kickoff ✓/✗ and Finalization ✓/✗. For the team leader:
  - **Both set:** "Ready. A manager can publish this letter."
  - **Missing:** the date fields are highlighted, with a secondary button **"Request publication without milestone dates"**. It opens a dialog that needs a reason and lists the missing dates.
  - **While pending:** a "Waiting for manager approval since <date>" badge and a Withdraw button.
  - **After a rejection:** the manager's note and the option to enter the dates.
- **Manager publish dialog**, replacing the current direct status change:
  - **Both dates set:** a normal confirm.
  - **Missing:** the two date fields (save and publish), or a "Publish without milestone dates" checkbox ("I confirm the Kickoff and Finalization dates are not known yet") with a required reason.
  - **The server rejects:** the dialog shows the 422 `MILESTONE_DATES_REQUIRED` message, not a generic error.
- **Manager decision panel** for a pending request (on the letter, and from the banner): the requester, reason, missing phases, and Approve and publish / Reject (note required).
- **Publish button visibility** follows the role set that the CAL-B18 fix settles on. This plan adds no role list of its own.

### 5.5 Missing dates and pending requests (CAL-FE11)

- Data from `GET /me/phase-date-gaps` (leaders) or `GET /calendar/phase-date-gaps` (managers).
- On `/calendar`, a collapsible banner with three groups for managers:
  - **Publish requests waiting** (first)
  - **Ready to publish**
  - **Missing phase dates**: milestones first; letters published with an exception are marked.

  Team leaders see their own missing dates and requests. Each row links to the phase schedule.
- Not shown to members, auditees, department heads or viewers.

### 5.6 LHA submit for approval (CAL-FE13, Phase 1b)

In the LHA module (`frontend/pages/audit-result-report/index.vue` and its components):
- **Team (auditor view):** a "Submit for approval" action. The report then shows "Pending approval".
- **Managers:** they keep the publish and authorize action (`canPublishLha`) and gain "Return to draft" with a note.
- The status options gain `Pending Approval`.

### 5.7 Fieldwork form relabel (CAL-FE9, CAL-T1)

| Field | EN | ID |
|---|---|---|
| `auditFieldwork.interview.date` | Planned interview date | Tanggal rencana wawancara |
| `auditFieldwork.observation.date` | Planned observation date | Tanggal rencana observasi |
| `auditFieldwork.sample.date` | Planned sampling date | Tanggal rencana pengambilan sampel |

The list column headers change too. The translation agent first checks where these keys are reused. If one is shared with a non-planned context, the frontend adds a separate key.

### 5.8 Click-through

| Entity | Target | Query |
|---|---|---|
| `assignment_letter`, `phase_schedule`, publish request | `/assignment-letter` | `?letter=<number>` (`&tab=schedule`) |
| `fieldwork_*` | `/audit-fieldwork` | `?letter=<number>&tab=…` |
| `audit_result_report` | `/audit-result-report` | `?id=<uuid>` |
| `action_taken_report` | `/action-taken-report` | existing `?lha=`; add `?id=<uuid>` |
| PKAT hold (email link) | `/annual-audit` | `?id=<uuid>` |

Auditees and viewers get the slideover only for fieldwork and phase events. A record that is not found falls back to the list with a toast.

### 5.9 i18n (en / id)

Add these keys:
- **`calendar.*`:** views, filters, phases, types, badges (milestone, draft, planned, done, override, noMilestoneDates), next milestone, gaps, `schedule.*`, `publishGate.*` (checklist, request dialog, manager dialog, decision panel, 422 message), states and detail.
- **`navigation.calendar`:** "Calendar" / "Kalender".
- **`notifications.*`:** every type in 4.7 and 4.8.
- **LHA:** `auditResultReport.status.pendingApproval` and the submit and return actions.

Email texts are server-side (ID and EN). Client terms stay as they are: Surat Tugas, KKA, LHA, ATR, PBC, PKAT, RKA.

---

## 6. Phased plan with hand-offs

### Phase 0: decisions and prerequisites

- Q1 to Q9 are decided (section 8), and the 3.5 interpretation is approved.
- **Prerequisite CAL-B18 (publish-role mismatch) is in progress with `auditsphere-backend`, in parallel with this plan.** It is referenced here, not designed here.
  - CAL-B4, CAL-B17 and CAL-FE12 depend on its outcome: which roles may publish, and how the CAE (`EXECUTIVE`) is recognised.
  - **Done when** a seeded `EXECUTIVE` can publish a letter through the backend, the publish button and the backend guard agree on the role set, and an AUDITOR still cannot change a letter's status. The project lead then records the final role set in 2.1 and 4.2.

### Phase 1: MVP (manual phase dates, publish gate, calendar, gaps)

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-B18 | `auditsphere-backend` (in progress, separate task) | **Prerequisite:** the publish-role mismatch (2.1). Not designed in this plan. | none | As in Phase 0. |
| CAL-A0 | `auditsphere-solution-architect` | Review `engagement_phase_schedules`, `letter_publish_exceptions`, member `employee_id` (B3a), the `department_id` claim, and the Phase 1b tables (notifications, LHA status and timestamps). | none | A review note with the final columns, the additive migration order and the B9 alignment. |
| CAL-B1 | `auditsphere-backend` (auth) | A `department_id` JWT claim. | B1 (done) | A linked login's token has `department_id`, and old tokens decode. |
| CAL-B20 | `auditsphere-backend` (auth) | A `roles=` filter on `GET /users/contacts` (and `department_id=` once B2 lands in Phase 1b). | none | `?roles=…` (the manager role set from CAL-B18) returns only active users with those roles. |
| CAL-B2 | `auditsphere-backend` (audit) | **B3a, hard prerequisite:** `employee_id` on `LetterMember` (and `ResourceAuditor`), with a dry-run name-match backfill report and an admin-run apply step. | CAL-A0 | A letter round-trips member IDs, and the report lists matched and unmatched members. |
| CAL-FE0 | `auditsphere-frontend` | Surat Tugas member picker by employee. | CAL-B2 contract | New letters store member `employee_id`, and old letters open and save. |
| CAL-B4 | `auditsphere-backend` (audit) | `engagement_phase_schedules` and `GET/PUT .../phase-schedule`: `can_edit`, leader by ID, manager override, validation, "set by". No defaults. Shared role package aligned with CAL-B18. | CAL-A0, CAL-B2, CAL-B18 | The leader can set, clear and mark held. Managers can override. Others get 403. Order violations return 422. A new letter has no rows. |
| CAL-B17 | `auditsphere-backend` (audit) | Publish gate (3.5): the precondition on every write into `Published`, `letter_publish_exceptions`, and request, approve, reject and withdraw. Approve publishes through `SyncLetterOnWrite`. Event-driven manager email for H1 (CAL-B20 recipients, Resend), recorded on the request row. | CAL-B4, CAL-B18, CAL-B20 | Publishing without dates returns 422. A manager with confirmation and a reason publishes, with the exception logged. A leader request creates PENDING and emails the managers of the same tenant only. Approve publishes and creates the execution. Reject keeps Draft. A second pending request returns 409. |
| CAL-B5 | `auditsphere-backend` (audit) | `GET /me/phase-date-gaps` and `GET /calendar/phase-date-gaps`, with pending requests and ready to publish. | CAL-B4, CAL-B17 | The leader and the managers see the right groups and nobody else does. Completed engagements are excluded. |
| CAL-B3 | `auditsphere-backend` (audit) | `GET /me/calendar` and `GET /calendar` (4.2 to 4.5): phase rows where entered, record events, the three milestones, `is_draft`, `published_with_exception`, titles, VIEWER and auditee rules (the planned LHA publication hidden from the auditee), and `meta`. | CAL-B1, CAL-B2, CAL-B4 | Per-role tests (4.3) pass, including the 3.7 example. A 42-day range returns in under 2 s locally. |
| CAL-D1 | `auditsphere-devops` | Add `/api/v1/me`, `/api/v1/calendar` and `/api/v1/notifications` to all four Kong configs, in one change. `/internal/*` stays out. | none | 401 (not 404) through Kong on dev and staging for each prefix. The tenant template renders all three. |
| CAL-FE4 | `auditsphere-frontend-2` | "Jadwal tahapan" (5.4). | CAL-B4 | Dates save, clear and reload. Fields start empty. Override is labelled. |
| CAL-FE12 | `auditsphere-frontend-2` | Publish gate UI (5.4): leader checklist, request dialog, pending badge and withdraw; manager publish dialog with confirm and reason; manager decision panel. Publish button visibility per CAL-B18. | CAL-B17, CAL-B18 | Every path in 3.5 works end to end, and the 422 message is shown in the dialog. |
| CAL-FE1 | `auditsphere-frontend` | Store, utilities and the `/calendar` page (5.2). | CAL-B3 contract (mock first) | vitest covers the maps, titles and range math. Milestones are never collapsed. Typecheck passes. |
| CAL-FE11 | `auditsphere-frontend` | Banners, widget lines and the admin dashboard card (5.3, 5.5). | CAL-B5, CAL-FE1 | Leaders and managers see their groups, and others see nothing. |
| CAL-FE2, CAL-FE3 | `auditsphere-frontend` | Widget in both dashboards; sidebar item. | CAL-FE1 | No "not connected" text. The item shows for every role. |
| CAL-FE5 | `auditsphere-frontend-2` | Deep links (5.8). | CAL-FE4 | Every link opens its record, with a fallback for unknown IDs. |
| CAL-FE9 | `auditsphere-frontend-2` | Planned-date labels in the fieldwork modals and lists. | CAL-T1 | The labels read "Planned … date" in both languages. |
| CAL-T1 | `auditsphere-software-translation` | `calendar.*` (including `publishGate.*`), `navigation.calendar`, and the fieldwork relabel. H1 email text (ID and EN). | CAL-FE1, CAL-FE4, CAL-FE12 key lists | Parity passes, and there are no hardcoded strings. |
| CAL-Q1 | `auditsphere-qa` | Checks 1 to 16 in section 7, through Kong on staging, plus vitest, typecheck, `go test` and locale validation. | All of the above | Pass, with evidence. |
| CAL-D2 | `auditsphere-devops` | Release via `backend/deploy-prod.sh --ref <sha>` with the no-seed override (never `docker-compose.prod.yml` as is), the manual frontend path, and a Kong reload. Confirm `RESEND_*` is set for audit-service in each tenant. Rollback named per service. | CAL-Q1 | The prod smoke returns 200 for `/me/calendar`. One test publish request emails the test manager. No data was reseeded. |

Order:
1. CAL-B18 (already running), CAL-A0, CAL-B20 and CAL-D1, in parallel.
2. CAL-B1 and CAL-B2 (with CAL-FE0).
3. CAL-B4 (once CAL-B18 lands), then CAL-B17 and CAL-FE4, then CAL-B5 and CAL-FE12.
4. CAL-B3, then CAL-FE1, then CAL-FE11, CAL-FE2 and CAL-FE3.

CAL-FE5, CAL-FE9 and CAL-T1 run alongside. CAL-FE1 can start against a mock as soon as 4.5 is agreed.

### Phase 1b: alerts and manager holds (next release)

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-B13 | `auditsphere-backend` (audit) | `notifications`, `notification_reads`, `notification_email_log`, and the endpoints. | CAL-A0, CAL-B1 | Audience matching per role (4.7). VIEWER sees none. Read state is per user. |
| CAL-B7 | `auditsphere-backend` (audit) | LHA `submitted_for_approval_at` and `published_at` (set once). The `LHA_PUBLISHED` calendar event. | CAL-A0 | The first approval sets `published_at`, and later edits do not change it. |
| CAL-B16 | `auditsphere-backend` (audit) | LHA `Pending Approval` status, `submit-for-approval`, and the backend role guard on approval statuses (4.6), using the CAL-B18 role set. | CAL-A0, CAL-B18 | An auditor can submit but cannot approve (403). A manager can approve or return. ATRs are still created only on approval. |
| CAL-B14 | `auditsphere-backend` (audit) | The `services/milestone_alerts` job and the event-driven senders: everything in 4.7 and H2 to H4 plus the daily hold reminder in 4.8. | CAL-B13, CAL-B4, CAL-B7, CAL-B16, CAL-B15 (for job email) | With a fake clock: the right audiences per row of 4.7 and 4.8. No alert for missing dates, drafts, held milestones or VIEWER. LHA planned alerts stop after publication. Re-armed on a date change. No duplicates after a restart. |
| CAL-B15 | `auditsphere-backend` (auth) | B2 (`users.department_id`, refreshed at login, with a backfill), the `department_id` filter on `/users/contacts`, and `GET /internal/v1/contacts` with a per-tenant service token. | CAL-A0 | Returns the unit's AUDITEE and DEPARTMENT_HEAD emails and the managers by role. Rejects a wrong or missing token. |
| CAL-D4 | `auditsphere-devops` | A per-tenant `INTERNAL_SERVICE_TOKEN` in `scripts/generate-tenant-stack.sh` and `onboard-tenant.sh`, `CALENDAR_*` settings, and the slugged auth service name. | CAL-B15 | Two tenants have different tokens, and tenant A's token is rejected by tenant B. |
| CAL-P1 | `auditsphere-platform` | Deliverability check (SPF, DKIM, DMARC) for the alert sender on each tenant domain. | CAL-B14 | Test alerts and hold emails land in the inbox. |
| CAL-FE10 | `auditsphere-frontend` | Header bell, unread count and panel (FE-G5, calendar subset). | CAL-B13 | Shows every type in 4.7 and 4.8, and mark-read works. |
| CAL-FE13 | `auditsphere-frontend-2` | LHA submit for approval and return to draft (5.6). | CAL-B16 | The status shows, and the actions respect roles. |
| CAL-T2 | `auditsphere-software-translation` | `notifications.*`, the LHA status keys, and all email texts (ID and EN). | CAL-B14 text list | Parity passes, and the emails are reviewed. |
| CAL-Q2 | `auditsphere-qa` | Checks 17 to 24 in section 7. | All of Phase 1b | Pass, with evidence. |
| CAL-D3 | `auditsphere-devops` | Release, the same path as CAL-D2. | CAL-Q2 | The job runs in prod (one log line per run), and no data was reseeded. |

In-app (CAL-B13, CAL-B14 without email, CAL-FE10) can ship before job email (CAL-B15, CAL-D4, CAL-P1). Event-driven emails need only CAL-B20 (plus B2 for the auditee unit).

### Phase 2: actual dates and planned vs actual

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-A1 | `auditsphere-solution-architect` | Review CAL-B6, CAL-B8 and CAL-B9. | Phase 1 shipped | Review note. |
| CAL-B6 | `auditsphere-backend` | Manual actual dates for every phase row, with planned-plus-actual events. No automatic actuals. | CAL-A1 | Actuals round-trip, and planned (outline) shows next to actual (filled). |
| CAL-B8 | `auditsphere-backend` | Typed letter dates, `target_unit_id`, `activity_id`, with a backfill (dry run first). SQL range filters. Exact Audit ID. | CAL-A1 | `meta.unparsed_dates` and `meta.unscoped_letters` drop to the listed rows. |
| CAL-B9 | `auditsphere-backend` + `auditsphere-frontend` | Typed PKAT planned month, and the RKA form writing `planned_start/end`. | CAL-A1 | PKAT bands show, and new RKA rows have planned dates. |
| CAL-FE6 | `auditsphere-frontend` | Planned-vs-actual columns and slip badges. | CAL-B6 | Edits round-trip, and the slip shows. |
| CAL-T3, CAL-Q3, CAL-D5 | translation, QA, devops | Keys; write RBAC; deploy. | As above | As in Phase 1. |

### Phase 3: drag to reschedule

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-FE7 | `auditsphere-frontend` | Library decision (5.1), then drag and resize of planned phase rows and milestones for write roles, with confirm and undo. | CAL-B4 | Saves for the leader and as an override. Record events are not draggable. A moved milestone re-arms its alerts. |
| CAL-B10 | `auditsphere-backend` | `PATCH .../phase-schedule/:phase` with validation and an audit trail. | CAL-B4 | 409 on a conflict, and the trail is readable. |

### Phase 4: export and other reminders

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-B11 | `auditsphere-backend` (auth + audit) | Revocable feed token and `GET /calendar/feed/:token.ics`. Milestones get a `VALARM` at H-1. | Phase 1 | Google Calendar and Outlook subscribe. Revoking returns 404. |
| CAL-B12 | `auditsphere-backend` | Reminders for PBC due, planned interviews and observations (for the auditee), KKA review and the LHA draft and review rows, plus ATR due-soon and overdue (B10). | CAL-B14 | One notification per item per day, to the right audience. |
| CAL-FE8 | `auditsphere-frontend` | `.ics` download and subscribe UI. | CAL-B11 | Works for an auditor and an auditee. |

---

## 7. Acceptance checks (for `auditsphere-qa`)

Phase 1:
1. A new Surat Tugas has no phase rows and no phase events. Nothing is prefilled.
2. The team leader can enter, change and clear every phase date. Managers can override, and "Set by" shows who. Members, other auditors, auditees and viewers get 403.
3. Prerequisite CAL-B18 holds: a seeded `EXECUTIVE` can publish a letter, and an AUDITOR cannot change a letter's status.
4. Publishing a letter without Kickoff and Finalization returns 422 `MILESTONE_DATES_REQUIRED`, through both `POST` and `PUT`.
5. A manager publishing with `confirm_without_milestones` and a reason succeeds. The exception (`MANAGER_CONFIRMED`) is stored with who, when and why, and the letter shows "Published without milestone dates".
6. A team leader's request creates one `PENDING` exception. The letter stays `Draft` with a "Waiting for manager approval" badge. An email goes to every manager of the same tenant (not the requester), and the send is recorded on the request.
7. A manager's approval publishes the letter and creates its audit execution. A rejection keeps it `Draft` and shows the note. Entering both dates while pending marks the request `SUPERSEDED`. A second pending request returns 409.
8. Gaps: the team leader and the managers see missing dates, pending requests and ready-to-publish letters. Members, auditees, department heads and viewers do not.
9. A phase with no date does not appear on the calendar for anyone.
10. A team member (not the leader) sees the engagement. An auditor on no team sees none and gets 403 on `/calendar`.
11. Managers see Draft letters with `is_draft: true`. Nobody else does.
12. An AUDITEE of the unit sees the 3.7 example, but **not** the planned LHA publication. A DEPARTMENT_HEAD sees the same.
13. A VIEWER sees the period, the three milestones (planned LHA publication included) and the published LHA, and nothing else.
14. Interview, observation and sample events are `basis: planned` and never overdue. The three milestones are `important: true`.
15. Titles follow `<Audit ID> – <Phase> – <Activity>`, with the letter number as the fallback.
16. Range over 100 days returns 400. Unparseable dates are counted. The three prefixes route through Kong on every config, and `/internal/*` does not. Locale parity, vitest, typecheck, and `go vet`/`go test` (audit, auth) pass.

Phase 1b:
17. With a fake clock: H-3 and H-1 for Kickoff and Finalization reach the team and the unit (in-app and email) and the managers (in-app only). There are none for missing dates, drafts, held milestones or VIEWER.
18. Planned LHA Publication: H-3 and H-1 go to the team (in-app and email) and the managers (in-app), never to the auditee, and stop once the LHA is published.
19. LHA published: the auditee unit gets in-app and email once; the team and the managers get in-app; VIEWER gets nothing. `published_at` is set once.
20. Moving a milestone date re-arms the alerts, and the old date's pending alerts are not sent.
21. Manager hold emails go out exactly for H1 to H4 (4.8), to the listed roles, with a daily reminder while the item is held. An ATR in `PENDING_REVIEW` does **not** email managers.
22. An auditor can submit an LHA for approval but cannot approve it (403 from the backend). Approving still creates the ATRs.
23. With `RESEND_API_KEY` unset, in-app alerts still work and the skipped emails are logged.
24. A service token from one tenant is rejected by another tenant's `/internal/v1/contacts`.

---

## 8. Decisions and open questions

All decided on 2026-10-27:

- **Q1, members:** every Surat Tugas team member sees the engagement. B3a is a hard MVP prerequisite (CAL-B2).
- **Q2, phase dates:** no automatic or default phase dates. The team leader enters every phase date by hand, and the CAE, audit manager and admin can override. A phase with no date does not appear. Missing dates are shown to the team leader and managers. Phase-date entry is in the MVP.
- **Q3, drafts:** managers see Draft Surat Tugas, marked as draft. Others do not, and drafts trigger no alerts.
- **Q4, department head:** the auditee unit view.
- **Q4, VIEWER:** org-wide, engagement level only, with no alerts.
- **Q5, fieldwork dates:** interview, observation and sample dates are planned dates, and the form is relabelled.
- **Q6, LHA Publication:** the third important milestone, next to Kickoff and Finalization. The planned date is the manually entered `LHA_PUBLISH` row; it alerts the team and managers at H-3 and H-1 and is hidden from the auditee. **An alert also fires when the LHA is actually published** (`published_at`): to the auditee unit (in-app and email, because its ATRs start then), the team and the managers (in-app). VIEWER gets no alerts.
- **Q7, timing:** calendar days.
- **Q8, manager email:** managers get in-app only, except an email for items on hold that need a manager. These are H1 to H4 in 4.8: the publish exception, PKAT pending review, PKAT pending approval and LHA pending approval (a new state), with a daily reminder. ATR review, KKA review and executive summary approval are excluded, for the reasons given in 4.8.
- **Q9, publish gate:** publishing a Surat Tugas requires Kickoff and Finalization dates, or a manager-confirmed or manager-approved exception (3.5). It is in the MVP. The interpretation in 3.5 was approved by the user: publishing stays a manager action, team leaders request an exception, and the letter stays `Draft` while a request waits.

Prerequisite (not a question): **CAL-B18, the publish-role mismatch** (`frontend/composables/useRbac.ts:219` allows ADMIN and `EXECUTIVE`; `backend/audit-service/routes/handler.go:185-186` allows ADMIN, `AUDIT_MANAGER` and `CHIEF_AUDIT_EXECUTIVE` by exact name). `auditsphere-backend` is fixing it in parallel. Record the final role set in 2.1 and 4.2 when it lands.

Open: none.

---

## 9. Changes to other docs

- `docs/auditor-auditee-dashboard-spec.md`: B4 and W1 point to this file. This plan also covers the calendar subset of B9 and B10, and proposes audience-addressed notifications (4.7) for the architect to confirm before B9 is built.
- `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md`: once shipped, not before, add:
  - the `/calendar` route
  - the phase schedule and the publish gate on the Surat Tugas (section 4.4)
  - the LHA `Pending Approval` status (section 6.1)
  - the notification bell (section 6.2)
