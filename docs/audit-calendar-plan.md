# Audit Calendar (Kalender Audit): implementation plan

Verified against the code on 2026-10-27 (branch `staging`, HEAD `59dbf5a`, with uncommitted edits to `AuditorDashboard.vue` and `pages/dashboard/index.vue`; line numbers in those two files are from the working tree). Re-check section 2 before relying on it.

Revision 3, 2026-10-27. Product-owner decisions are recorded in section 8. Revision 2 added the Audit Kickoff and Audit Finalization milestones, their alerts, the auditee engagement view and the title pattern. Revision 3 records that **every phase date is entered by hand** (no automatic or default dates), moves phase-date entry into the MVP, and closes Q2 and Q4.

Owner: project lead. This plan is the detailed design of item **B4** (`GET /api/v1/me/calendar`) and widget **W1** (calendar of important dates) in `docs/auditor-auditee-dashboard-spec.md`, and of the calendar part of **B9/B10** (notifications and the due-soon job). It does not replace that spec: IDs B1 to B13, W0 to W12 and FE-G1 to FE-G14 refer to it. New IDs use the prefix `CAL-`.

Planning only. No application code was changed.

---

## 1. Summary

The calendar shows the audit schedule **by phase** to auditors, managers, the audited unit and viewers. Two milestones, **Audit Kickoff** and **Audit Finalization**, are highlighted and trigger alerts.

**Phase dates are manual.** The Surat Tugas team leader enters every phase date by hand. The CAE, the audit manager and the admin can override them. The system never proposes, defaults or derives a phase date. A phase with no date entered does not appear on the calendar. Instead, the missing dates show up as gaps for the team leader and for managers.

What the code gives us today:

- **Record dates exist for part of the lifecycle.** These are dates on records, not phase dates:
  - the Surat Tugas letter date and engagement period
  - fieldwork interview, observation and sample dates, which are **planned** dates
  - the PBC required date and the test-of-controls due date
  - the LHA report and signature dates
  - the ATR due date, with derived overdue days

  The calendar shows them as their own events.
- **No phase date is stored anywhere.** Kickoff (entry meeting), fieldwork, KKA review, Finalization (exit meeting) and the LHA steps have no date. `audit_executions` knows the current stage but stores no dates.
- **Most stored record dates are strings** (`varchar`). The server parses them and reports the rows it could not parse.
- **The dashboard spec names `audit_activities.planned_start/end` as a calendar source. That source is not usable**, because the RKA form never writes those columns (`backend/audit-service/services/activity_plan_form/save.go:622-641`).
- **Alerts:** email exists (Resend, `backend/audit-service/pkg/mailer`), but only as the manual ATR reminder. **No in-app notification exists**: there is no table, endpoint, store or header bell.

Recommended shape:

1. **Phase 1 (MVP):**
   - The phase-schedule table and its entry form, the "Jadwal tahapan" on the Surat Tugas. It holds **planned dates for every phase**, with Kickoff and Finalization first.
   - A publish prompt for the two milestones, and a missing-phase-dates indicator.
   - The read-only calendar endpoints `GET /api/v1/me/calendar` and `GET /api/v1/calendar`.
   - Team scoping by employee ID (B3a), which is a hard prerequisite.
   - The `/calendar` page and the dashboard widget.
2. **Phase 1b (milestone alerts), the next release:**
   - In-app notifications and a header bell.
   - H-3 and H-1 alerts for Kickoff and Finalization.
   - Email through the existing Resend sender.
   - A reminder to the team leader about missing milestone dates.
3. **Phase 2:** actual dates per phase (entered by hand), the planned-vs-actual slip view, typed letter dates and links, LHA status timestamps, and the PKAT/RKA schedule.
4. **Phases 3 and 4:** drag-to-reschedule, `.ics` export, and reminders for the non-milestone events.

---

## 2. Current state (evidence)

### 2.1 Prerequisites from the dashboard spec

| Item | Status on 2026-10-27 | Evidence |
|---|---|---|
| B1: employee link exposed to the caller | **Done** (commit `2baf95d`). `/auth/me` returns `employee_code`, `employee_id`, `department_id`. The JWT carries `employee_id` and `employee_code`, but **not** `department_id`. | `backend/auth-service/models/dto.go:39-48`; `backend/auth-service/pkg/utils/jwt.go:13-18`; `backend/audit-service/pkg/middleware/auth.go:12-21, 75-79` |
| audit-service helper for the caller's employee | Exists, but nothing uses it yet. | `backend/audit-service/pkg/base/base_controller.go:79-88` (`GetEmployeeID`) |
| B3a: stable ID on Surat Tugas team members | **Open, and a hard MVP prerequisite**: every member sees the engagement, and the team leader must be identified to edit phase dates. `LetterMember` is `{name, role}`. | `backend/audit-service/models/assignment-letter.go:13-16` |
| B2: department key on users | Open. `users.department` is free text. | `backend/auth-service/models/user.go:19` |
| B3: `/me/engagements` | Open. No `/me` route in audit-service. | `backend/audit-service/routes/handler.go:74-514` |
| FE-G1/FE-G3: auditee dashboard | **Done** (commit `818f94e`), with a calendar placeholder. | `frontend/components/dashboard/AuditeeDashboard.vue:348-353` |
| Server-side row scope pattern | ATR only (PIC sees own rows; `view_all_action_taken_report` sees all). Role aliases: `EXECUTIVE` = CAE. | `backend/audit-service/services/action_taken_report/rbac.go:21-191` |
| Letter status guard | Status changes on assignment letters are limited by role (`RequireRolesForStatusChange`). This is where Publish happens, and so where the milestone prompt hooks in. | `backend/audit-service/routes/handler.go:179-195` |

### 2.2 How engagements are linked today

- **The Surat Tugas is the engagement.** Every live letter has one `audit_executions` row with `ref` = letter number. The execution is created on save and soft-deleted with the letter (`services/audit_execution_sync/letter_link.go:14-28, 117, 269`).
- **The team leader** is the letter's `leader` (free text), else the member whose role is Chairperson (`letter_link.go:57-90`), else the linked RKA row's `team_leader_id`. Only the RKA key is an ID today, which is why B3a is needed.
- **Execution → RKA row** (`audit_executions.activity_id`) is set by the progress sync. It uses an explicit link first, then a case-insensitive title or project-code match (`services/audit_execution_sync/sync.go:227-266`). The link is fuzzy and often missing.
- **The RKA row holds** `project_code` (the Activity ID, for example `ASR-2026-003`), `team_leader_id` (a master employee UUID) and `target_unit_id` (a master department UUID) (`models/audit-activity.go:157-187`; `services/activity_plan_form/save.go:606-631`).
- **The letter has no department key** (`working_unit` is free text) and **no RKA key** (`models/assignment-letter.go:22-31`).
- Fieldwork, KKA and LHA rows point to the letter through a `varchar` `assignment_letter_id` (the letter number or the UUID; `FindAssignmentLetter`, `assignment-letter.go:145-162`). ATR points to the LHA by foreign key.

### 2.3 The dashboard stepper is not date-based

The fieldwork stepper takes its stage from which records exist: an LHA means stage 4, a KKA form means stage 3, a fieldwork item means stage 2, and otherwise the stage is 1 (`frontend/components/dashboard/AuditorDashboard.vue:1365-1390`). The letter period only drives the deadline badge (`:1835-1850`). The calendar does not take dates from the stepper.

### 2.4 Notifications and email (what exists to reuse)

| Capability | What exists | Evidence | Reuse |
|---|---|---|---|
| Email sender | Resend REST client: `RESEND_API_KEY`, `RESEND_FROM` (default `no-reply@mail.auditsphere.app`), plain text, a 12 s timeout. Returns `ErrNotConfigured` when the key is unset. | `backend/audit-service/pkg/mailer/mailer.go:18-89` | Yes, as is. |
| Recipient lookup | `authclient.Contacts` calls auth `GET /api/v1/users/contacts` **with the caller's forwarded JWT**. Filters: user `ids`, or `department_head_of=<department name>`. Limited to the assignable-user roles. | `backend/audit-service/pkg/authclient/client.go:109-175`; `backend/auth-service/routes/registry.go:135-150` | **Not usable from a background job** (no user token), and it cannot select "all users of a unit" or "by employee ID". See CAL-B15. |
| Reminder flow | Manual `POST /action-taken-reports/:id/reminder` emails the PIC and the department heads, with per-recipient skip reasons. | `backend/audit-service/controllers/action_taken_report/reminder.go:63-245` | Copy the message and skip-reason style. |
| In-app notifications | **None.** No model, endpoint, store or bell. | dashboard spec 2.5; `frontend/layouts/default.vue` | Build the calendar subset of B9 (CAL-B13). |
| Periodic job pattern | `AuditCompletionAnalyzer`: a `time.NewTicker` loop started in `serve`. | `backend/audit-service/services/audit_completion/analyzer.go:60-100` | Same pattern for CAL-B14. |
| Events from other services | audit-service consumes Kafka `department.events` only. There is no user directory. | `backend/*/pkg/kafka` | n/a |

### 2.5 Frontend stack

- `@nuxt/ui` 4.4.0 (`UCalendar` date picker with a `day` slot, `UInputDate`, `UTimeline`; **no scheduler**), `date-fns` 4.1, `vue-i18n` 11 (`frontend/package.json`).
- `frontend/components/shared/AppDatePicker.vue` already builds a month grid with date-fns.
- Deep links: only ATR (`?lha=`) and working paper (`?tab=&step=`) read query parameters. No module has an `[id]` route.
- i18n: `dashboard.auditor.calendar.*` and `dashboard.auditee.calendar.*` exist (placeholder texts). There is no `calendar.*` and no `navigation.calendar`. The fieldwork date labels are `auditFieldwork.interview.date`, `auditFieldwork.observation.date` and `auditFieldwork.sample.date` (`frontend/components/audit-fieldwork/AuditFieldwork{Interview,Observation,Sample}Modal.vue:199/143/140`).

### 2.6 Kong

Kong routes by prefix, per service, in **four** configs: `backend/kong-gateway/kong/dev/kong-dev.yml`, `backend/kong-gateway/kong/prod/kong.yml`, `backend/tenants/bai/kong/kong.yml` and `scripts/templates/kong-tenant.yml.tpl`. None has `/api/v1/me`, `/api/v1/calendar` or `/api/v1/notifications`. A missing prefix returns 404 (`no Route matched with those values`) through Kong only.

---

## 3. Phases, milestones and events

### 3.1 Phase vocabulary

Phases are aligned with the `audit_executions` canonical stages (`models/audit-execution.go:11-18`). Every event carries an icon and a text label, so colour is never the only signal. Dark-mode variants are required.

| Phase key | Display label (EN / ID) | Execution stage | Colour | Icon |
|---|---|---|---|---|
| `PLANNING` | Audit Planning / Perencanaan Audit | planning | slate | `i-lucide-clipboard-list` |
| `ASSIGNMENT` | Surat Tugas / Surat Tugas | planning | indigo | `i-lucide-file-signature` |
| `ENTRY_MEETING` | Audit Kickoff / Kickoff Audit (rapat pembukaan) | entry meeting | sky | `i-lucide-flag` (milestone) |
| `FIELDWORK` | Audit Fieldwork / Pekerjaan Lapangan Audit | fieldwork | emerald | `i-lucide-search-check` |
| `WORKING_PAPER` | KKA Review / Reviu KKA | draft findings | teal | `i-lucide-file-check` |
| `EXIT_MEETING` | Audit Finalization / Finalisasi Audit (rapat penutupan) | reporting & exit meeting | violet | `i-lucide-flag-triangle-right` (milestone) |
| `REPORTING` | Audit Reporting (LHA) / Pelaporan Audit (LHA) | reporting, completed | amber | `i-lucide-file-text` |
| `FOLLOW_UP` | Follow-up (ATR) / Tindak Lanjut (ATR) | after completed | rose (overdue: red) | `i-lucide-list-checks` |

Final wording belongs to `auditsphere-software-translation`.

### 3.2 Phase dates: entered by hand (MVP)

The team leader enters these on the Surat Tugas "Jadwal tahapan" (phase schedule, section 5.4). Each row is optional, and **a row with no date produces no calendar event**.

| Schedule row (`phase`) | Calendar event type | Shape | Important | Notes |
|---|---|---|---|---|
| `ENTRY_MEETING` | `AUDIT_KICKOFF` | day (+ optional location) | **yes** | Prompted at publish (3.4) |
| `FIELDWORK` | `PHASE_FIELDWORK` | range | no | The team's planned fieldwork window. It can differ from the Surat Tugas period. |
| `WORKING_PAPER` | `KKA_REVIEW_DUE` | day or range | no | Review chain: auditor → team leader → QA |
| `EXIT_MEETING` | `AUDIT_FINALIZATION` | day (+ optional location) | **yes** | Prompted at publish (3.4) |
| `LHA_DRAFT` | `LHA_DRAFT_DUE` | day | no | |
| `LHA_REVIEW` | `LHA_REVIEW_DUE` | day | no | |
| `LHA_PUBLISH` | `LHA_PUBLISH_DUE` | day | candidate (Q6) | |

Rules:

- **Who enters dates:** the Surat Tugas team leader. **The CAE (seeded as `EXECUTIVE`), the audit manager and the admin can override any phase date.** Every save records `updated_by_user_id` and `updated_by_name`, and the slideover shows "Set by <name>, <date>". An override by someone other than the team leader also notifies the team leader (Phase 1b, `PHASE_DATE_CHANGED`).
- **No defaults:**
  - The form opens empty for a new letter.
  - Nothing is copied from the Surat Tugas period, the RKA or the PKAT.
  - No date is derived from records (first fieldwork item, first KKA form, LHA creation).
- **Validation:** start ≤ end within a row, and Kickoff ≤ Finalization (error). A date outside the Surat Tugas period, or an LHA step before Finalization, gives a warning only.
- **Actual dates** (Phase 2) are also entered by hand ("Mark as held" / "Mark as done"). In the MVP only Kickoff and Finalization have "Mark as held", because it cancels their alerts.

### 3.3 Record events (existing dates, shown as they are)

These come from the records themselves. They are not phase dates and need no entry. "Parse" means the field is a string, parsed as `YYYY-MM-DD` and then the formats in `assignment-letter.go:79-86`.

| Event type | Phase | Source table.column | Stored as | Basis | Shape | MVP? | Auditee / dept head | VIEWER |
|---|---|---|---|---|---|---|---|---|
| `LETTER_ISSUED` | ASSIGNMENT | `assignment_letters.letter_date` | timestamp | actual | day | Yes | No | No |
| `ENGAGEMENT_PERIOD` | FIELDWORK (background band) | `assignment_letters.start_period`, `finish_period` (entered on the Surat Tugas) | varchar (parse) | planned | range | Yes | Yes | Yes |
| `INTERVIEW` | FIELDWORK | `fieldwork_interviews.date` | varchar (parse) | **planned** | day | Yes | **Yes** (activity only) | No |
| `OBSERVATION` | FIELDWORK | `fieldwork_observations.date` | varchar (parse) | **planned** | day | Yes | **Yes** (activity only) | No |
| `SAMPLE` | FIELDWORK | `fieldwork_samples.date` | varchar (parse) | **planned** | day | Yes, off by default | No | No |
| `PBC_DUE` | FIELDWORK | `fieldwork_documents.required_date` | varchar (parse) | due | day | Yes (status inferred until B6) | Yes | No |
| `TEST_CONTROL_DUE` | FIELDWORK | `fieldwork_test_controls.due_date` | varchar (parse) | due | day | Yes | No | No |
| `LHA_REPORT_DATE` | REPORTING | `audit_result_reports.report_date` | timestamp | actual | day | Yes | Published only | Published only |
| `LHA_SIGNED` | REPORTING | `audit_result_reports.signature_date` | timestamp | actual | day | Yes | Published only | Published only |
| `LHA_PUBLISHED` | REPORTING | status timestamp (CAL-B7) | n/a today | actual | day | No (Phase 2) | Yes | Yes |
| `ATR_DUE` | FOLLOW_UP | `action_taken_reports.due_date`; `is_overdue`, `overdue_days` derived | timestamp | due | day | Yes | Yes (own PIC rows) | No |
| `ATR_REVIEWED` | FOLLOW_UP | `action_taken_reports.reviewed_at` | timestamp | actual | day | Yes, off by default | Yes (own) | No |
| `PKAT_MONTH` | PLANNING | `audit_annuals.selected_months`, `activities[].timelineText` | int array / text | planned | month | No (CAL-B9) | No | No |

`audit_assignments.start_date/end_date` stay out of scope (`/audit-assignment` has no confirmed producer).

**Planned fieldwork dates.** Interviews, observations and samples are drawn as planned events: an outline chip with the label "Planned". A past date shows as "planned (past)", never as overdue. The form labels change to say so (CAL-FE9, CAL-T1).

### 3.4 Audit Kickoff and Audit Finalization

**What they map to:**

| Milestone | Lifecycle step | Indonesian term | Why |
|---|---|---|---|
| **Audit Kickoff** | Entry meeting: the formal start of the engagement with the auditee, after the Surat Tugas is issued | Rapat pembukaan (entry meeting) | The first point where the auditee unit is involved. It is not the Surat Tugas date, which is internal issuance. |
| **Audit Finalization** | Exit meeting: the close of fieldwork, where findings are discussed with the auditee before the LHA | Rapat penutupan (exit meeting) | It matches the user's example (kickoff 22 Oct, interview 25 Oct, finalization 27 Oct, all inside the fieldwork window). LHA publish is a separate, later row (Q6). |

**Where their dates come from:** only from manual entry (3.2). Nothing stores them today.

**Making sure they get entered:**

- **Publish prompt (MVP, CAL-FE4).** When a user publishes a Surat Tugas that has no Kickoff or Finalization date, a modal asks for both. It has two empty date pickers and an optional location. The user can save and publish, or publish without them. Publishing is not blocked (Q9). If they skip, the letter is flagged in the missing-dates indicator.
- **Missing-phase-dates indicator (MVP, CAL-B5 and CAL-FE11),** section 4.2 and 5.5.
- **In-app reminder to the team leader (Phase 1b):** `MILESTONE_DATE_MISSING`, once at publish and again 7 days later while still missing.
- **Mark as held (MVP):** records `actual_start`. The milestone then shows as done, and its remaining alerts are cancelled.

### 3.5 Event titles

Pattern: **`<Audit ID> – <Phase> – <Activity>`**. Parts that add nothing are dropped. The server returns the parts (`title_parts`) and a composed `title` in the request's language. The frontend localises the parts and can drop the Audit ID when the view is filtered to one engagement.

- **Audit ID:** the linked RKA `project_code` (for example `ASR-2026-003`), with the Surat Tugas number as the fallback (`audit_id_source: "rka" | "letter"`). Phase 2 (CAL-B8) stores the RKA link on the letter, which makes this exact.

| Event | Title |
|---|---|
| `AUDIT_KICKOFF` | `ASR-2026-003 – Audit Kickoff` |
| `PHASE_FIELDWORK` | `ASR-2026-003 – Audit Fieldwork` |
| `INTERVIEW` | `ASR-2026-003 – Audit Fieldwork – Interview` |
| `OBSERVATION` | `ASR-2026-003 – Audit Fieldwork – Observation` |
| `PBC_DUE` | `ASR-2026-003 – Audit Fieldwork – PBC due: <document name>` |
| `KKA_REVIEW_DUE` | `ASR-2026-003 – KKA Review` |
| `AUDIT_FINALIZATION` | `ASR-2026-003 – Audit Finalization` |
| `LHA_DRAFT_DUE` | `ASR-2026-003 – Audit Reporting – LHA draft due` |
| `ATR_DUE` | `ASR-2026-003 – Follow-up – ATR due: <finding title>` |

Interview and observation titles for an auditee contain no interviewee, interviewer, topic or notes.

### 3.6 Worked example: the auditee's unit

The team leader of `ASR-2026-003` (Published Surat Tugas, period 22 to 31 Oct) entered Kickoff on 22 Oct and Finalization on 27 Oct. An auditor planned an interview on 25 Oct. An AUDITEE or DEPARTMENT_HEAD of Cabang Bandung sees:

| Date | Title | Type | Look | Alerts (Phase 1b) |
|---|---|---|---|---|
| 22 Oct | ASR-2026-003 – Audit Kickoff | `AUDIT_KICKOFF`, important | milestone highlight | H-3 on 19 Oct, H-1 on 21 Oct |
| 22–31 Oct | ASR-2026-003 – Audit Fieldwork period | `ENGAGEMENT_PERIOD` | background band | none |
| 25 Oct | ASR-2026-003 – Audit Fieldwork – Interview | `INTERVIEW`, planned | outline chip, "Planned" | none |
| 27 Oct | ASR-2026-003 – Audit Finalization | `AUDIT_FINALIZATION`, important | milestone highlight | H-3 on 24 Oct, H-1 on 26 Oct |

If the team leader had not entered the Finalization date, the 27 Oct row would not exist and no alerts would fire. The engagement would appear in the missing-dates indicator for the team leader and the managers. The auditee would not see the gap.

The team also sees samples, test controls, PBC items and the other phase rows. Managers see all of it, including while the letter was Draft (draft style, no alerts). A VIEWER sees only the period band and the two milestones.

---

## 4. Data model, API and alerts

### 4.1 Ownership and shape

- **Owner: audit-service.** All calendar data lives in the audit database. The request path makes no call to master or risk.
- **Phase 1:**
  - one new table, `engagement_phase_schedules`, holding every phase row (manual entry)
  - read-only aggregation for record events
  - the range filter on string record dates runs after parsing, on a candidate set bounded by letter scope
  - the `requirements.MD` dashboard SLA (2 s at 100 concurrent users) is a target; CAL-B3 checks it on seeded data
- **Phase 1b:** notification tables (4.7). **Phase 2:** actual-date use of the same table, plus typed letter columns (4.6).
- All migrations are additive and run at `serve` start. `auditsphere-solution-architect` reviews them (CAL-A0 for Phases 1 and 1b, CAL-A1 for Phase 2).

### 4.2 Endpoints

| Endpoint | Who | Scope | Phase |
|---|---|---|---|
| `GET /api/v1/me/calendar?from&to&types&phases&important` | Any authenticated user | The caller's items (4.3) | 1 |
| `GET /api/v1/calendar?from&to&types&phases&important&department_id&letter&team_leader_id&overdue&include_draft` | ADMIN, EXECUTIVE / CHIEF_AUDIT_EXECUTIVE, AUDIT_MANAGER (all events); VIEWER (engagement level only, 4.3) | Org-wide. Drafts by default for managers, never for VIEWER. | 1 |
| `GET /api/v1/assignment-letters/:id/phase-schedule` | Anyone who can see the letter. The response includes `can_edit`. | One engagement, every phase row | 1 |
| `PUT /api/v1/assignment-letters/:id/phase-schedule` | Team leader of the letter; **CAE, audit manager and admin as override** | Upsert or clear rows. Planned dates, location, and `actual_start` for the milestones. | 1 |
| `GET /api/v1/me/phase-date-gaps` | Team leaders | Published (and Draft) letters the caller leads that miss any phase row. Milestones listed first. | 1 |
| `GET /api/v1/calendar/phase-date-gaps?department_id&team_leader_id` | ADMIN, CAE, AUDIT_MANAGER | The same, org-wide | 1 |
| `GET /api/v1/notifications?unread&limit`, `PUT /api/v1/notifications/:id/read`, `PUT /api/v1/notifications/read-all` | Any authenticated user | Notifications whose audience matches the caller (4.7) | 1b |
| `GET /internal/v1/contacts` (auth-service) | audit-service only, with a per-tenant service token. **Not routed by Kong.** | Email addresses for alerts | 1b |
| `PATCH /api/v1/assignment-letters/:id/phase-schedule/:phase` | Write roles above | Drag-to-reschedule | 3 |
| `GET /api/v1/calendar/feed/:token.ics` | Holder of a revocable per-user feed token | As `/me/calendar` | 4 |

Gap response, per letter: `{assignment_letter_id, letter_number, audit_id, status, team_leader_name, missing: ["ENTRY_MEETING", "EXIT_MEETING", ...], missing_milestones: 2}`. Letters with execution status `completed` are excluded.

Rules for `from` and `to`: both required, `YYYY-MM-DD`, `to >= from`, span at most 100 days. A wider span returns 400 `CALENDAR_RANGE_TOO_WIDE`. A forbidden scope returns 403 `FORBIDDEN_SCOPE`. Errors use the existing envelope.

The phase-schedule routes sit under `/assignment-letters`, and the gap routes under `/me` and `/calendar`, so there are no prefixes beyond the three in CAL-D1.

### 4.3 RBAC scoping (server-side, per the ATR pattern)

Move the role canonicalisation (`CanonicalRole`, the aliases, `IsAdmin`) from `services/action_taken_report/rbac.go` into a shared audit-service package, so ATR, the calendar, the phase schedule and notifications use one mapping.

| Role (canonical) | `/me/calendar` | `/calendar` | Phase schedule write | Gaps |
|---|---|---|---|---|
| ADMIN, CHIEF_AUDIT_EXECUTIVE (incl. seeded `EXECUTIVE`), AUDIT_MANAGER | Engagements they lead or belong to, plus ATRs where they are PIC or reviewer | **All events, including Draft letters** (draft style, `is_draft: true`) | **Override on any letter** | Org-wide |
| AUDITOR (incl. aliases), as team leader | Published letters where the caller is leader or any member, by employee ID, plus own ATRs | 403 | **Own letters as leader** | Own led letters |
| AUDITOR, as member | Same as leader (every member sees the engagement) | 403 | No (read only) | No |
| AUDITEE | Own ATRs (`pic_user_id` = caller). For Published letters of the caller's unit: `ENGAGEMENT_PERIOD`, `AUDIT_KICKOFF`, `AUDIT_FINALIZATION`, `INTERVIEW` and `OBSERVATION` (activity only), `PBC_DUE`, `LHA_*` once Published. **Never** samples, test controls, KKA, the other phase rows, or gaps. | 403 | No | No |
| DEPARTMENT_HEAD | **Same as AUDITEE**, plus the department's ATRs where the head is not PIC (read only) | 403 | No | No |
| VIEWER | Empty list | **Org-wide, engagement level only:** `ENGAGEMENT_PERIOD`, `AUDIT_KICKOFF`, `AUDIT_FINALIZATION`, `LHA_*` once Published, for Published letters. **Never** fieldwork items, PBC, KKA, ATR, other phase rows or Draft letters. Filters: phase, unit, milestones only. No alerts. | No | No |

Notes:

- **Team leader by ID:** a member whose role is Chairperson and who has an `employee_id` (B3a), else the linked RKA `team_leader_id`. A letter whose leader cannot be resolved by ID can be edited by managers only, and the gap response flags it (`leader_unresolved: true`).
- **The unit of a letter:** the linked RKA row's `target_unit_id` (through `audit_executions.activity_id`), then the letter's own `target_unit_id` once CAL-B8 adds it. A letter with neither is shown to managers only and counted in `meta.unscoped_letters`.
- **The caller's department** comes from a new `department_id` JWT claim (CAL-B1). Without it, an auditee gets ATR items only and `meta.unit_scope: "unavailable"`.
- **An account with no employee link** gets ATR items only and `meta.employee_linked: false`.
- **Legacy letters without member employee IDs** are visible to managers and to an RKA-linked leader only, until the CAL-B2 backfill runs. `meta.unlinked_members` counts them.
- **Draft letters** never reach auditors, auditees, department heads or viewers, and never trigger alerts.

### 4.4 Tenant isolation

Each tenant has its own stack and its own `rb_audit_audit_<slug>` database. The calendar reads only that database. The risks specific to this feature:

- **Kong tenant template:** if `scripts/templates/kong-tenant.yml.tpl` is not updated with the other three configs, new tenants get 404s.
- **The service token for `/internal/v1/contacts`** (Phase 1b) must be **per tenant**. `jwt.secret` is identical across tenants today, so a token from one tenant validates on every other; the service token must not repeat that.
  - Call auth by the tenant's slugged service name. All tenants share one Docker network.
- **Alert dedupe state lives in Postgres, not Redis.** Redis is shared and its keys are unprefixed.
- **The `.ics` feed token** (Phase 4) is stored per tenant database and is revocable.

### 4.5 Event payload (extends the B4 contract in the dashboard spec, section 5)

The frontend has not consumed B4 yet, so the contract can still change. `id`, `type`, `date`, `title`, `entity_type`, `entity_id`, `status` and `is_overdue` keep their meaning. `date` is the start day.

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
    "missing_phase_dates": 3
  }
}
```

Field values:

- `type`: from 3.2 and 3.3. `phase`: from 3.1.
- `basis`: `planned | actual | due`. `important`: true for milestones. `milestone`: `KICKOFF | FINALIZATION | null`. `done`: true once marked held.
- `precision`: `day | month`.
- `entity_type`: `assignment_letter | phase_schedule | fieldwork_interview | fieldwork_observation | fieldwork_sample | fieldwork_document | fieldwork_test_control | audit_result_report | action_taken_report`.
- `my_role`: `LEADER | MEMBER | PIC | REVIEWER | UNIT | null`.
- `location`: milestones only, optional. `set_by_name`: phase rows only.
- `meta.missing_phase_dates`: the number of in-scope letters with gaps. It is returned only to roles that can see gaps (4.3), and is `null` for everyone else.
- `important=true` as a query parameter returns milestones only.

The backend returns entity references, not routes. The frontend maps entities to pages (CAL-FE5).

### 4.6 Schema (for `auditsphere-solution-architect`)

- **CAL-B4 `engagement_phase_schedules`** (Phase 1, every phase):
  - Columns: `id uuid`; `assignment_letter_id uuid not null` (FK to `assignment_letters.id`, soft-deleted with the letter through the service); `phase varchar(30)` (`ENTRY_MEETING | FIELDWORK | WORKING_PAPER | EXIT_MEETING | LHA_DRAFT | LHA_REVIEW | LHA_PUBLISH`); `planned_start date not null`; `planned_end date null`; `actual_start date null`; `actual_end date null`; `location varchar(255) null`; `minutes_media_id uuid null`; `updated_by_user_id`; `updated_by_name`; `created_at`; `updated_at`; `deleted_at`.
  - Unique `(assignment_letter_id, phase)` among live rows. Indexes on `planned_start` and `actual_start`.
  - A row exists only when a date was entered, so clearing a date deletes the row.
  - No `actual_source` column: actual dates are always manual.
- **CAL-B7 LHA timestamps:** `submitted_at`, `finalized_at`, `published_at` on `audit_result_reports`, set on status change (or a status-history table; the architect decides). These are record facts, not phase dates.
- **CAL-B8 typed letter dates and links:** `start_date date null`, `finish_date date null`, `target_unit_id uuid null` and `activity_id uuid null` on `assignment_letters`, with a backfill (dry run first). Keep the string columns.
- **CAL-B9 PKAT and RKA planned schedule:** a typed planned month on `AnnualAuditActivity`, and the RKA form writing `audit_activities.planned_start/end`. These are the planner's manual entries, not calendar defaults.
- **CAL-B13 notifications** (Phase 1b, 4.7).

### 4.7 Milestone alerts (Phase 1b)

**What triggers an alert:** an **entered** Kickoff or Finalization date on a **Published** Surat Tugas. Alerts never fire for a missing date, a Draft letter, a milestone marked held, or a date in the past.

**When** (times in the tenant timezone, default `Asia/Jakarta`):

| Trigger | When | Recipients |
|---|---|---|
| `MILESTONE_DUE_SOON` H-3 | 3 calendar days before, at 08:00 | All audiences below |
| `MILESTONE_DUE_SOON` H-1 | 1 day before, at 08:00 | All audiences below |
| `MILESTONE_RESCHEDULED` | Immediately, when a date that already had an alert sent moves, or a new date is entered within 3 days | All audiences below. Pending H-3 and H-1 re-arm for the new date. |
| `MILESTONE_DATE_MISSING` | At publish, and 7 days later while still missing | Team leader only (managers see the gaps indicator) |
| `PHASE_DATE_CHANGED` | When a CAE, audit manager or admin overrides a phase date | Team leader only |

Offsets and hour are per-tenant settings (`CALENDAR_ALERT_OFFSETS_DAYS=3,1`, `CALENDAR_ALERT_HOUR=8`). Calendar days are used, with no weekend shift (Q7).

**Who receives milestone alerts:**

| Audience | In-app | Email |
|---|---|---|
| Team: leader and every member (by employee ID) | yes | yes |
| Auditee unit: AUDITEE and DEPARTMENT_HEAD users of the letter's unit | yes | yes |
| Managers: ADMIN, CAE (EXECUTIVE), AUDIT_MANAGER | yes | no (avoids one email per engagement; Q8) |
| VIEWER | no | no |

**Channels:**

- **In-app (new, CAL-B13 and CAL-FE10):** the calendar subset of B9. audit-service has no user directory, so notifications are **audience-addressed**:
  - `notifications`: `id`, `audience_kind` (`USER | EMPLOYEE | DEPARTMENT | ROLE`), `audience_key`, `type` (above), `title_parts json`, `entity_type`, `entity_id`, `assignment_letter_id`, `event_date`, `offset_days`, `dedupe_key` (unique), `created_at`.
  - `notification_reads`: `notification_id`, `user_id`, `read_at`, with unique `(notification_id, user_id)`.
  - `GET /api/v1/notifications` matches the caller's JWT `user_id`, `employee_id`, `department_id` and canonical roles. `DEPARTMENT` audiences are shown only to AUDITEE and DEPARTMENT_HEAD.
  - This differs from the dashboard spec's B9 sketch (`recipient_user_id`). ATR notifications fit the same table with `audience_kind = USER`. The architect confirms in CAL-A0.
  - UI: a header bell with an unread count and a panel linking to the calendar event.
- **Email (existing Resend sender, CAL-B14 and CAL-B15):** `pkg/mailer.Resend.Send`, plain text in the ATR-reminder style, behind `CALENDAR_EMAIL_ALERTS=true`. When `RESEND_API_KEY` is unset, in-app alerts still go out and the job logs the skipped emails.
  - Addresses come from a **new auth-service internal endpoint** (CAL-B15): `GET /internal/v1/contacts?employee_ids=&department_id=&roles=`. It authenticates with a per-tenant service token and is not routed by Kong.
  - Filtering by `department_id` needs B2: store `users.department_id`, refresh it at login from the employee link, and backfill it.
  - The job does not reuse `authclient.Contacts`, because that needs a user JWT.

**Job (CAL-B14):** `services/milestone_alerts`, a ticker loop like `AuditCompletionAnalyzer`, started in `serve`, every 15 minutes.
- Each run selects the entered milestone dates in the next 3 days for Published letters, builds the audiences, and inserts notifications with `dedupe_key = <letter_id>:<milestone>:<offset>:<event_date>`. This is idempotent across restarts, and a moved date re-arms the alerts.
- Email is logged per recipient (`notification_email_log`: `notification_id`, `email`, `status`, `error`, `sent_at`, unique `(notification_id, email)`), with up to 3 retries on later runs.
- The `requirements.MD` 6 s email target is a target, not a measurement.

**Why Phase 1b, not the MVP proper:** the MVP is where the dates come from (entry form, publish prompt, gaps). Alerts need two more tables, a job, the bell, a Kong prefix and (for email) an auth internal endpoint with a per-tenant secret. Splitting them lets the calendar and date entry ship first, so teams start entering dates. The alerts follow in the next release, ahead of all of Phase 2. In-app and email are separate tasks, so email can slip without blocking in-app.

---

## 5. Frontend

### 5.1 Library decision

All events are whole-day, so the MVP views (month, week, agenda) are built in-house with date-fns and Nuxt UI, following `AppDatePicker.vue`. This matches FE-G4 in the dashboard spec. The widget's mini-month uses `UCalendar` with its `day` slot. The decision gate before Phase 3 compares FullCalendar (the MIT daygrid, list and interaction plugins) with extending the in-house grid.

### 5.2 Calendar page (`/calendar`)

- Route `frontend/pages/calendar/index.vue`. A sidebar item directly under Dashboard (`frontend/layouts/default.vue:84-86`), `navigation.calendar`, icon `i-lucide-calendar-days`, visible to every role. The auditor branch (`default.vue:607-630`) must not hide it.
- Views: **Month** (default; range bars; up to 3 events a day plus "+n more"), **Week**, and **Agenda** (the default on small screens).
- **Milestone highlight:**
  - Listed first in a day and never collapsed into "+n more".
  - Look: a thicker chip with a phase-colour ring and a flag icon, a "Milestone" badge, and bold text.
  - A "Next milestone" strip above the grid. Done milestones show filled with a check mark.
- **Draft style (managers only):** dashed border, a muted fill and a "Draft" badge (`is_draft`). There is a filter to hide drafts.
- **Planned style:** an outline chip with "Planned" for `basis: planned`.
- **Missing-phase-dates banner** (team leaders and managers, CAL-FE11), described in 5.5.
- Filters (multi, kept in the query string): phase (legend toggle), event type (samples off by default), milestones only, unit, team leader, Surat Tugas, overdue only, "Mine / All engagements" (managers), and show drafts (managers). VIEWER gets phase, unit and milestones only.
- Click opens a `USlideover` with the details, "Set by", and an **Open** button. For phase rows, write roles also get "Edit dates", and for milestones "Mark as held".
- States (dashboard spec section 7): skeleton; empty; inline error with Retry; 403; and notes for an unlinked account, unparsed dates and unit unavailable.
- `frontend/stores/calendar.ts` (Pinia, cached by range, scope and filters). `frontend/utils/calendarPhases.ts` holds the phase→colour/icon/i18n map, the entity→route map and the title composer, with unit tests.

### 5.3 Dashboard widget (replaces both placeholders)

`frontend/components/dashboard/CalendarWidget.vue`, in `AuditorDashboard.vue` (section `:590-626`, replacing the `TODO(B4)` block and the `placeholder` badge) and in `AuditeeDashboard.vue` (`:348-353`). It shows:
- the next milestone at the top
- a `UCalendar` mini-month with phase dots and a flag on milestone days
- the next 14 days as an agenda (at most 8 rows)
- an overdue count
- for team leaders, a "**n engagements with missing phase dates**" line linking to the gaps view
- an "Open calendar" link

It wraps `AuditorWidgetState`. Remove the `notConnected` and `placeholder` keys once wired, and keep `notLinked`.

For managers, the gap count also appears on the admin dashboard as a small card linking to `/calendar?gaps=1` (part of CAL-FE11). The admin view has no calendar widget today.

### 5.4 Phase schedule on the Surat Tugas (CAL-FE4)

- **"Jadwal tahapan" (phase schedule) section** in the Surat Tugas detail (`frontend/pages/assignment-letter/index.vue` and its components), reachable at `?letter=<number>&tab=schedule`.
  - One row per phase from 3.2, with Kickoff and Finalization first and marked as milestones.
  - Each row has `AppDatePicker` fields (a single date or a range), a location for the milestones, "Set by", and a clear button.
  - All fields start **empty**. Nothing is prefilled from the period or from any record.
  - Editable when `can_edit` is true: the team leader, and the CAE, audit manager and admin (labelled "override" when they are not the leader). Read only for everyone else.
  - Validation as in 3.2.
- **Publish prompt:** on the Publish action, if Kickoff or Finalization is missing, show a modal with the two empty date pickers and the location. The choices are "Save and publish" and "Publish without dates". The second keeps the letter in the gaps list.
- The letter create form (`AssignmentLetterForm.vue`) is unchanged. Phase dates are entered after the letter exists, because the team leader may not be the person who creates the letter.

### 5.5 Missing-phase-dates indicator (CAL-FE11)

- Data from `GET /me/phase-date-gaps` (leaders) or `GET /calendar/phase-date-gaps` (managers).
- On `/calendar`: a collapsible banner, "n engagements have missing phase dates (m missing Kickoff or Finalization)". It lists each letter (Audit ID, title, leader, missing phases, milestones first) with a "Set dates" link to the phase schedule. `?gaps=1` opens it.
- On the auditor dashboard widget (team leader) and the admin dashboard card (managers), as in 5.3.
- Not shown to members, auditees, department heads or viewers.

### 5.6 Fieldwork form relabel (CAL-FE9, CAL-T1)

Relabel `auditFieldwork.interview.date`, `auditFieldwork.observation.date` and `auditFieldwork.sample.date` (and the matching list column headers):

| Field | EN | ID |
|---|---|---|
| Interview date | Planned interview date | Tanggal rencana wawancara |
| Observation date | Planned observation date | Tanggal rencana observasi |
| Sample date | Planned sampling date | Tanggal rencana pengambilan sampel |

The translation agent first checks where the keys are reused. If a key is shared with a non-planned context, the frontend adds a separate key.

### 5.7 Click-through

| Entity | Target | Query to add |
|---|---|---|
| `assignment_letter`, `phase_schedule` | `/assignment-letter` | `?letter=<number>` (`&tab=schedule` for phase rows) |
| `fieldwork_*` | `/audit-fieldwork` | `?letter=<number>&tab=interviews\|observations\|samples\|documents\|testControls` |
| `audit_result_report` | `/audit-result-report` | `?id=<uuid>` |
| `action_taken_report` | `/action-taken-report` | existing `?lha=`; add `?id=<uuid>` |

Auditees and viewers get the slideover only for fieldwork and phase events (no Open button). A record that is not found falls back to the list with a toast.

### 5.8 i18n (en / id)

New namespace `calendar.*`:
- `title`, `views.*`, `toolbar.*`, `filters.*`
- `phases.<KEY>` and `types.<TYPE>`
- `basis.*`, `badges.milestone|draft|planned|done|override`
- `nextMilestone`, `gaps.*`
- `schedule.*`: the phase schedule section, the publish prompt and validation
- `states.*`, `detail.*`

Plus `navigation.calendar` ("Calendar" / "Kalender") and `notifications.*` (bell, panel, and the alert texts for H-3, H-1, rescheduled, date missing and override, with the `{auditId}`, `{milestone}`, `{date}` and `{days}` placeholders). Alert emails are server-side plain text in Indonesian and English in one message. Client terms stay as they are: Surat Tugas, KKA, LHA, ATR, PBC, PKAT, RKA. Dates use the date-fns `id` locale, and the week starts on Monday.

---

## 6. Phased plan with hand-offs

### Phase 0: decisions

Q1 to Q5 are decided (section 8). Still open: Q6 to Q9, which are short confirmations with working defaults. None blocks Phase 1. Q7 and Q8 only affect Phase 1b. **Done when** Q6 to Q9 are answered before Phase 1b starts.

### Phase 1: MVP (manual phase dates, calendar, gaps)

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-A0 | `auditsphere-solution-architect` | Review `engagement_phase_schedules`, member `employee_id` (B3a), the `department_id` claim and the Phase 1b notification tables. | none | The review note lists the final columns, the additive migration order and the B9 alignment. |
| CAL-B1 | `auditsphere-backend` (auth) | A `department_id` JWT claim from the employee link, read in the audit middleware and base controller. | B1 (done) | A linked login's token has `department_id`. Old tokens still decode. Tests pass. |
| CAL-B2 | `auditsphere-backend` (audit) | **B3a, hard prerequisite:** `employee_id` on `LetterMember` (and `ResourceAuditor`). Add a dry-run name-match backfill report, then an admin-run apply step. | CAL-A0 | A letter round-trips member IDs. The report lists matched and unmatched members, and the apply step writes only the matched ones. |
| CAL-FE0 | `auditsphere-frontend` | Surat Tugas member picker by employee. | CAL-B2 contract | New letters store member `employee_id`, and old letters open and save. |
| CAL-B4 | `auditsphere-backend` (audit) | The `engagement_phase_schedules` table (every phase) and `GET/PUT /assignment-letters/:id/phase-schedule`, with `can_edit`, leader-by-ID resolution, manager override, the validation in 3.2 and "set by". No defaults. | CAL-A0, CAL-B2 | The leader can set, clear and mark held. CAE, manager and admin can override. A member or another auditor gets 403. kickoff > finalization returns 422. A new letter has no rows. |
| CAL-B5 | `auditsphere-backend` (audit) | `GET /me/phase-date-gaps` and `GET /calendar/phase-date-gaps`. | CAL-B4 | A published letter with no Kickoff appears with `missing_milestones >= 1` for its leader and for managers, and for nobody else. A completed engagement is excluded. |
| CAL-B3 | `auditsphere-backend` (audit) | `GET /me/calendar` and `GET /calendar` per 4.2 to 4.5: phase rows (only where entered), record events, the `important` flag, `is_draft`, `title_parts` and Audit ID, the VIEWER engagement-level filter, a shared role package, and the parser with `meta` counters. | CAL-B1, CAL-B2, CAL-B4 | Per-role table tests (4.3) pass on seeded data, including the 3.6 example and its missing-Finalization variant. A 42-day range returns in under 2 s locally. `go vet` and the tests pass. |
| CAL-D1 | `auditsphere-devops` | Add `/api/v1/me`, `/api/v1/calendar` and `/api/v1/notifications` to audit-service in all four Kong configs, in one change. `/internal/*` stays out of Kong. | none | Through Kong on dev and staging, each prefix returns 401 without a token (not 404). `/internal/v1/contacts` returns 404 through Kong. The tenant template renders all three. |
| CAL-FE4 | `auditsphere-frontend-2` | The "Jadwal tahapan" section and the publish prompt (5.4). | CAL-B4 | Dates save, clear and reload. The publish prompt shows only when a milestone is missing. Override is labelled. All fields start empty. |
| CAL-FE1 | `auditsphere-frontend` | `stores/calendar.ts`, `utils/calendarPhases.ts`, and `/calendar` with the month, week and agenda views, milestone, draft and planned styles, filters, slideover (with edit and mark-held for write roles) and states. | CAL-B3 contract (mock until it lands) | vitest covers the phase map, entity routes, the title composer and the range math. Milestones are never collapsed. Typecheck passes. |
| CAL-FE11 | `auditsphere-frontend` | Missing-phase-dates banner on `/calendar`, the line in the widget, and the admin dashboard card (5.5). | CAL-B5, CAL-FE1 | A team leader and a manager see the count and the list. Members, auditees and viewers do not. "Set dates" opens the phase schedule. |
| CAL-FE2 | `auditsphere-frontend` | `CalendarWidget.vue` in both dashboards. | CAL-FE1 | Neither dashboard shows "not connected", and the widget links to `/calendar`. |
| CAL-FE3 | `auditsphere-frontend` | Sidebar item under Dashboard for every role. | CAL-FE1 | Shows for ADMIN, AUDITOR, AUDITEE, DEPARTMENT_HEAD and VIEWER. |
| CAL-FE5 | `auditsphere-frontend-2` | Deep links (5.7), including `?letter=&tab=schedule`. | CAL-FE4 | Every link opens its record. An unknown ID falls back with a toast. Auditees and viewers get no Open button for fieldwork or phase events. |
| CAL-FE9 | `auditsphere-frontend-2` | Planned-date labels in the three fieldwork modals and lists. | CAL-T1 key decision | The labels read "Planned … date" in both languages. |
| CAL-T1 | `auditsphere-software-translation` | `calendar.*`, `navigation.calendar` and the fieldwork relabel (5.6) in en and id. Remove the obsolete placeholder keys after CAL-FE2. | CAL-FE1 and CAL-FE4 key lists | Locale parity passes, and the new files have no hardcoded strings. |
| CAL-Q1 | `auditsphere-qa` | Checks 1 to 12 in section 7 through Kong on staging, plus vitest, typecheck, `go test` and locale validation. | All of the above | Pass, with evidence. |
| CAL-D2 | `auditsphere-devops` | Release: backend via `backend/deploy-prod.sh --ref <sha>` with the no-seed override (never `docker-compose.prod.yml` as is), the frontend via the manual prod path, and Kong reloaded. Rollback named per service. | CAL-Q1 | The prod smoke `GET /api/v1/me/calendar` returns 200 for a test account, and no data was reseeded. |

Order: CAL-A0 and CAL-D1 first, in parallel. Then CAL-B1 and CAL-B2 (with CAL-FE0). Then CAL-B4, and from it CAL-B5 and CAL-FE4. Then CAL-B3, and from it CAL-FE1 → CAL-FE11, CAL-FE2, CAL-FE3. CAL-FE5, CAL-FE9 and CAL-T1 run alongside. CAL-FE1 can start against a mock as soon as section 4.5 is agreed.

### Phase 1b: milestone alerts (next release after the MVP)

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-B13 | `auditsphere-backend` (audit) | The `notifications` and `notification_reads` tables (audience-addressed, 4.7) and the `GET/PUT /api/v1/notifications` endpoints. | CAL-A0, CAL-B1 | An AUDITEE sees `DEPARTMENT` notifications for its own unit only, an auditor sees its own `EMPLOYEE` notifications only, and a VIEWER sees none. Read state is per user. |
| CAL-B14 | `auditsphere-backend` (audit) | The `services/milestone_alerts` job: H-3 and H-1, rescheduled, date missing (at publish and +7 days), override notice, the dedupe key, and per-tenant offsets and hour. Email behind `CALENDAR_EMAIL_ALERTS`. | CAL-B13, CAL-B4 | On seeded data with a fake clock: one notification per audience per offset. Nothing for a missing date, a Draft letter or a held milestone. Re-armed after a date change. No duplicates after a restart. |
| CAL-B15 | `auditsphere-backend` (auth) | B2 (`users.department_id`, refreshed at login, with a backfill) and `GET /internal/v1/contacts` with a per-tenant service token. | CAL-A0 | It returns the AUDITEE and DEPARTMENT_HEAD emails for a department and the emails for a set of employee IDs, and rejects a missing or wrong token. |
| CAL-D4 | `auditsphere-devops` | A per-tenant `INTERNAL_SERVICE_TOKEN` in `scripts/generate-tenant-stack.sh` and `onboard-tenant.sh`, set for auth and audit, plus `CALENDAR_*` and `RESEND_*` for audit-service. Confirm the slugged auth service name is used. | CAL-B15 | Two tenant stacks have different tokens, and tenant A's token is rejected by tenant B. |
| CAL-P1 | `auditsphere-platform` | Deliverability check (SPF, DKIM, DMARC) for the alert sender on each tenant domain. | CAL-B14 | Test alerts land in the inbox. |
| CAL-FE10 | `auditsphere-frontend` | Header bell, unread count and panel in `frontend/layouts/default.vue` (FE-G5, calendar subset). | CAL-B13 | The bell shows the alert types in 4.7, and mark-read works. |
| CAL-T2 | `auditsphere-software-translation` | `notifications.*` keys and the server email texts (ID and EN). | CAL-B14 text list | Parity passes, and the email texts are reviewed. |
| CAL-Q2 | `auditsphere-qa` | Checks 13 to 17 in section 7. | All of Phase 1b | Pass, with evidence. |
| CAL-D3 | `auditsphere-devops` | Release, the same path as CAL-D2. | CAL-Q2 | The job runs in prod (one log line per run), and no data was reseeded. |

In-app (CAL-B13, CAL-B14 without email, CAL-FE10) can ship before email (CAL-B15, CAL-D4, CAL-P1) if those slip.

### Phase 2: actual dates and planned vs actual

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-A1 | `auditsphere-solution-architect` | Review CAL-B6 to CAL-B9. | Phase 1 shipped | A review note with an additive migration order. |
| CAL-B6 | `auditsphere-backend` | Manual actual dates for every phase row ("Mark as done" with a date) and planned-plus-actual events. No automatic actuals. | CAL-A1 | Actuals round-trip, and the calendar shows planned (outline) next to actual (filled). |
| CAL-B7 | `auditsphere-backend` | LHA `submitted_at`, `finalized_at`, `published_at` (record facts). | CAL-A1 | Each is set once on status change, and `LHA_PUBLISHED` shows. |
| CAL-B8 | `auditsphere-backend` | Typed letter dates, plus `target_unit_id` and `activity_id` on the letter, with a backfill (dry run first). SQL range filters. Exact Audit ID. | CAL-A1 | `meta.unparsed_dates` and `meta.unscoped_letters` drop to the rows the report listed. |
| CAL-B9 | `auditsphere-backend` + `auditsphere-frontend` | Typed PKAT planned month, and the RKA form writing `planned_start/end`. | CAL-A1 | PKAT month bands show, and new RKA rows have planned dates. |
| CAL-FE6 | `auditsphere-frontend` | Planned-vs-actual columns and a slip badge ("+5 days") in the phase schedule and on the calendar. | CAL-B6 | Edits round-trip, and the slip shows. |
| CAL-T3, CAL-Q3, CAL-D5 | translation, QA, devops | Keys; write RBAC; deploy. | As above | As in Phase 1. |

### Phase 3: drag to reschedule

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-FE7 | `auditsphere-frontend` | Library decision (5.1), then drag and resize of planned phase rows and milestones for write roles, with confirm and undo. | CAL-B4 | Dragging saves for the leader and as an override. Record events are not draggable. A dragged milestone re-arms its alerts. |
| CAL-B10 | `auditsphere-backend` | `PATCH .../phase-schedule/:phase` with validation and an audit trail (who, from, to, when). | CAL-B4 | 409 on a conflict, and the trail is readable. |

### Phase 4: export and other reminders

| ID | Agent | Deliverable | Depends on | Done when |
|---|---|---|---|---|
| CAL-B11 | `auditsphere-backend` (auth + audit) | A revocable per-user feed token and `GET /calendar/feed/:token.ics` (RFC 5545). Milestones get a `VALARM` at H-1. | Phase 1 | Google Calendar and Outlook subscribe. Revoking returns 404. |
| CAL-B12 | `auditsphere-backend` | Reminders for PBC due, planned interviews and observations (for the auditee), KKA review and the LHA phase rows, plus ATR due-soon and overdue (B10), using the CAL-B13 and CAL-B14 machinery. | CAL-B14 | One notification per item per day, to the right audience only. |
| CAL-FE8 | `auditsphere-frontend` | `.ics` download and subscribe UI. | CAL-B11 | Works for an auditor and an auditee. |

---

## 7. Acceptance checks (for `auditsphere-qa`)

Phase 1:
1. A new Surat Tugas has no phase rows and no phase events. Nothing is prefilled from the period or from any record.
2. The team leader can enter, change and clear every phase date. The CAE, audit manager and admin can override, and "Set by" shows who did. A member, another auditor, an auditee or a viewer gets 403 on `PUT .../phase-schedule`.
3. Publishing a letter without Kickoff or Finalization shows the prompt. "Publish without dates" publishes and the letter appears in the gaps list for its leader and for managers only.
4. A phase with no date does not appear on the calendar for anyone.
5. A team **member** (not the leader) sees the engagement's events. An auditor on no team sees none, and gets 403 on `/calendar`.
6. A CAE or ADMIN sees **Draft** letters with `is_draft: true`. An auditor, auditee, department head or viewer never does.
7. An AUDITEE of the unit sees exactly the 3.6 example, with no interviewee, topic, sample, test control, KKA or gap data. A DEPARTMENT_HEAD of the unit sees the same.
8. A VIEWER gets 200 on `/calendar` with only the period, Kickoff, Finalization and Published LHA events.
9. Kickoff and Finalization carry `important: true`. Interview, observation and sample events carry `basis: planned` and are never overdue.
10. Titles follow `<Audit ID> – <Phase> – <Activity>`, with the letter number as the fallback Audit ID.
11. A range over 100 days returns 400. An unparseable record date is counted, and the response does not fail.
12. All three new prefixes route through Kong on dev, staging, prod and a freshly generated tenant (401, not 404). `/internal/*` is not reachable through Kong. Locale parity, vitest, typecheck, and `go vet`/`go test` for audit and auth pass.

Phase 1b:
13. With a fake clock: H-3 and H-1 notifications are created once each for the team, the auditee unit and the managers. None for a missing date, a Draft letter or a held milestone. None for VIEWER.
14. Moving the Kickoff date re-arms the alerts, and the old date's pending alerts are not sent.
15. `MILESTONE_DATE_MISSING` reaches the team leader at publish and 7 days later, and stops once the date is entered. A manager override sends `PHASE_DATE_CHANGED` to the leader.
16. Email goes to the team and the unit only (not managers). With `RESEND_API_KEY` unset, in-app still works and the job logs the skipped emails.
17. A service token from one tenant is rejected by another tenant's `/internal/v1/contacts`.

---

## 8. Decisions and open questions

Decided on 2026-10-27:

- **Q1, members:** every Surat Tugas team member sees the engagement, not only the leader. B3a is a hard MVP prerequisite (CAL-B2).
- **Q2, phase dates:** there are **no automatic or default phase dates**. The Surat Tugas team leader enters every phase date by hand, and the CAE, audit manager and admin can override. A phase with no date does not appear on the calendar. Missing dates are shown to the team leader and managers (5.5). The team leader is prompted for Kickoff and Finalization at publish. Phase-date storage and entry are in the MVP.
- **Q3, drafts:** managers see Draft Surat Tugas, marked as draft. Others do not, and drafts trigger no alerts.
- **Q4, department head:** DEPARTMENT_HEAD gets the auditee unit view.
- **Q4, VIEWER:** org-wide events at engagement level only (period, Kickoff, Finalization, LHA once Published). No fieldwork, KKA, PBC, ATR or Draft letters, and no alerts.
- **Q5, fieldwork dates:** interview, observation and sample dates are planned dates, labelled planned in the payload and the UI. The form is relabelled.
- **Auditee calendar and milestones:** the auditee calendar works at engagement level (3.6). Titles follow `<Audit ID> – <Phase> – <Activity>`. Audit Kickoff and Audit Finalization are highlighted milestones with alerts. Their dates are entered in the MVP, and their alerts ship in Phase 1b.

Still open (each has a working default, and none blocks Phase 1):

6. **Q6:** confirm that Audit Finalization = exit meeting (end of fieldwork), with LHA publish as its own row. Should LHA publish also be a highlighted milestone? (Default: no.)
7. **Q7:** are alerts on calendar days fine, or should H-3 and H-1 skip weekends and holidays? (Default: calendar days.)
8. **Q8:** managers get milestone alerts in-app only. Do they also want email, for example a daily digest? (Default: in-app only.)
9. **Q9:** should publishing a Surat Tugas be **blocked** until Kickoff and Finalization are entered? (Default: no, prompt only.)

---

## 9. Changes to other docs

- `docs/auditor-auditee-dashboard-spec.md`: B4 and W1 point to this file. This plan also covers the calendar subset of B9 and B10, and proposes audience-addressed notifications (4.7) for the architect to confirm before B9 is built.
- `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md`: add the `/calendar` route, the phase schedule on the Surat Tugas and the notification bell under sections 4.4 and 6.2 once they ship, not before.
