# COSO 2013 PoF assessment (QUANTITATIVE_POF): API contract

Status: contract for the backend implemented on branch `staging` (2026-11-12). The frontend is built from this file. Everything is ADDITIVE: the qualitative YES/PARTIAL/NO assessment (`method = QUALITATIVE`) keeps its behaviour and endpoints. Background and decisions: `docs/coso-pof-competitor-research.md` (section 6), `docs/coso-2013-component-assessment-plan.md`.

## 0. Conventions

- Base path: `/api/v1/ic-assessments` (audit-service, the prefix already routed by Kong in all four configs, so **no Kong change**). The catalogue endpoint is `/api/v1/rcm/coso-pof-catalogue` (risk-service, under the existing `/api/v1/rcm` prefix).
- Success envelope: `{"success": true, "message": "...", "data": ...}`.
- Error envelope: `{"success": false, "error": {"code": "...", "message": "...", "details": "..."}}`.
- Validation errors of the PoF endpoints are **422** with one message per field:
  `{"success": false, "error": {"code": "VALIDATION_ERROR", "message": "Invalid request", "fields": {"items[2].achievement": "must be one of 0, 0.25, 0.5, 0.75, 1"}}}`.
  Field keys use the request path (`items[3].achievement`, `pofs[0].explanation`, `mapping.columns.weight`).
- Other codes: 400 `BAD_REQUEST` (malformed query/body), 401, 403 `FORBIDDEN`, 404 `NOT_FOUND`, 409 (`VERSION_CONFLICT`, `INVALID_STATUS`, `PERIOD_LOCKED`, `METHOD_MISMATCH`, `ASSESSMENT_EXISTS`, `IMPORT_ALREADY_APPLIED`), 413 file too large, 415 file type, 422 `MALWARE_DETECTED`, 502 `SOURCE_UNAVAILABLE` (risk-service catalogue unreachable), 503 `SCANNER_UNAVAILABLE`.
- Numbers: weights, achievement (0 to 1) and scores are JSON numbers rounded to 4 decimals. **A node that is not fully assessed has `score: null` and `achievement: null`, never 0.** `assessed_score` / `assessed_weight` give the partial sums so the UI can show progress.
- Roles are compared case-insensitively. VIEWER: read-only (every write is 403 by `DenyViewerWrites`; routes with `evidence`, `download` or `export` in a path segment are 403 as well, `DenyViewerFiles`).
- Codes: components `CE RA CA IC MA`, principles `P01..P17`, PoF `1.1.1` (component.principle.pof), element `1.1.1.1`.

## 1. Method per branch

The method is chosen per branch (`location_id`) and stored on the assessment header when it is created. It never changes afterwards.

| Field | Values |
|---|---|
| `method` | `QUALITATIVE` (default, existing flow), `QUANTITATIVE_POF` (new) |

### GET `/branch-methods?location_id=`
Any authenticated reader. `data`:
```json
{
  "methods": ["QUALITATIVE", "QUANTITATIVE_POF"],
  "default": {"method": "QUALITATIVE", "source": "builtin"},
  "items": [{"location_id": "uuid", "location_name": "Cabang Surabaya", "method": "QUANTITATIVE_POF", "updated_by_name": "x", "updated_at": "..."}],
  "effective": {"location_id": "uuid|null", "method": "QUANTITATIVE_POF", "source": "branch|default|builtin"}
}
```
`effective` is present only with `location_id` (or `location_id=` empty for the no-branch scope). `default` is the setting of the no-branch scope (`location_id: null`).

### PUT `/branch-methods`
Approver roles (CAE, ADMIN). Body `{"location_id": "uuid|null", "method": "QUALITATIVE|QUANTITATIVE_POF"}`. `location_id: null` sets the tenant default (assessments without a branch). Affects **new** assessments only. Response `data`: the row plus `existing_assessments: {"QUALITATIVE": 2, "QUANTITATIVE_POF": 0}` for that scope so the UI can warn about mixing methods. Logged as `BRANCH_METHOD_UPDATE`.

### POST `/` (existing)
Body unchanged (`department_id, year, quarter, location_id?`); the method comes from the branch setting (the request cannot override it). For `QUANTITATIVE_POF` the server reads the current catalogue from risk-service and copies it into the assessment (80 PoF rows, about 190 element rows, weights, texts, default owners), records `catalog_version` and the thresholds in effect. Response `data` is the existing detail plus `method`, `catalog_version`, `pof_yes_threshold`, `pof_partial_threshold`. 502 `SOURCE_UNAVAILABLE` when the catalogue cannot be read (nothing is created).

For a QUANTITATIVE_POF assessment:
- `PUT /:id` accepts only `final_conclusion`, `conclusion_text`, `conclusion_justification` (sending `components` or `principles` is 409 `METHOD_MISMATCH`).
- `POST /:id/propose` is 409 `METHOD_MISMATCH` (no signal proposal).
- The principle rows (`status`, `severity`, `rationale`) and component rows (`final_rating`, `final_status`, `final_severity`) are **derived by the server** from the PoF scores each time a value is saved, so `GET /`, `GET /summary` and the dashboards keep working. They are empty (`NOT_ASSESSED`, null) until every PoF below them is assessed.
- `submit`, `return`, `approve`, `reopen`, `DELETE`: same endpoints, roles and segregation of duties as today.

## 2. Settings: thresholds and owners

### GET `/pof-settings`
Any authenticated reader.
```json
{
  "pof_yes_threshold": 0.8,
  "pof_partial_threshold": 0.4,
  "source": "stored|default",
  "owner_overrides": [{"principle_code": "P11", "owner_roles": ["HEAD_OF_IT"], "updated_by_name": "x", "updated_at": "..."}]
}
```
### PUT `/pof-settings`
ADMIN. Body (every field optional): `{"pof_yes_threshold": 0.8, "pof_partial_threshold": 0.4, "owner_overrides": [{"principle_code": "P11", "owner_roles": ["HEAD_OF_IT"]}]}`. Rules: `0 < partial < yes <= 1`, at most 3 decimals; an override with `owner_roles: []` removes the override; a label matches `^[A-Z][A-Z0-9_]{1,39}$`. Thresholds apply to assessments created afterwards (an assessment keeps the values it was created with). Logged as `POF_SETTINGS_UPDATE`.

Status from achievement `a` (PoF, principle, component or total): `a >= pof_yes_threshold` is `YES`, `a >= pof_partial_threshold` is `PARTIAL`, otherwise `NO`. Defaults 0.80 and 0.40, reasons in the research doc.

### GET `/api/v1/rcm/coso-pof-catalogue?framework=COSO_2013&version=` (risk-service)
Any authenticated role. `data`: `{"catalogue": {framework, version, name, is_current, components:[{code, number, name_en, name_id, weight, principles:[{code, number, index, name_en, name_id, weight, default_owner_roles:[], pofs:[{code, index, name_en, name_id, weight, default_owner_roles, elements:[{code, index, type: "PRESENT|FUNCTION", description_en, description_id, weight, default_methods:["D","K"], scale_type: "DISCRETE|QUESTIONNAIRE", requires_independent_assessor}]}]}]}]}, "counts": {"pofs": 81, "elements": 187}, "owner_roles": [{"label": "HEAD_OF_IT", "name_en": "...", "name_id": "...", "auth_role": ""}]}`. `auth_role` is the matching auth-service role (`EXECUTIVE`, `DEPARTMENT_HEAD`) or empty: a label with no role is assigned to a user together with the free label, no role is created. A stored catalogue that breaks a weight invariant is never served (500 `CATALOGUE_INVALID`).

Methods: `D` document, `K` questionnaire, `W` interview, `O` observation. `DISCRETE` accepts 0, 0.25, 0.5, 0.75, 1; `QUESTIONNAIRE` accepts any value in [0, 1].

## 3. Reading the assessment

### GET `/:id/pof`
Readers (assessors, approvers, VIEWER); scoped readers (DEPARTMENT_HEAD, AUDITEE) see an APPROVED assessment of their own department, or the principles/PoFs assigned to them (section 5). Query: `expand=elements` (include elements, default off), `component=CE`, `principle=P01`, `status=NOT_ASSESSED|IN_PROGRESS|YES|PARTIAL|NO` (PoF status), `assignee=me`, `needs=explanation|evidence`. Filters narrow the `pofs` lists; totals are always over everything and `filtered` says whether a filter was applied. 409 `METHOD_MISMATCH` for a qualitative assessment.

```json
{
  "assessment_id": "uuid", "method": "QUANTITATIVE_POF", "catalog_version": "2013.1", "status": "DRAFT",
  "version": 3, "editable": true, "filtered": false,
  "thresholds": {"yes": 0.8, "partial": 0.4},
  "total": {"weight": 100, "score": 83.26, "achievement": 0.8326, "status": "YES", "assessed_weight": 100, "assessed_score": 83.26,
            "progress": {"pofs_total": 81, "pofs_assessed": 81, "elements_total": 187, "elements_assessed": 187}},
  "derived_conclusion": "EFFECTIVE",
  "components": [{
    "code": "CE", "weight": 25, "score": 21.64, "achievement": 0.8656, "status": "YES", "assessed_weight": 25, "assessed_score": 21.64,
    "progress": {"pofs_total": 20, "pofs_assessed": 20, "elements_total": 55, "elements_assessed": 55},
    "principles": [{
      "code": "P01", "number": 1, "name_en": "...", "weight": 5.5, "score": null, "achievement": null, "status": "IN_PROGRESS",
      "assessed_weight": 2.75, "assessed_score": 2.5, "progress": {...},
      "default_owner_roles": ["EXECUTIVE"], "assignees": [{"id": "uuid", "user_id": "u1", "name": "A", "role_label": "EXECUTIVE", "due_date": "2026-12-01"}],
      "pofs": [{
        "code": "1.1.1", "name_en": "Tone at the top", "name_id": "Teladan dari pimpinan", "weight": 1.375,
        "basis": "NONE|POF|ELEMENTS", "direct_achievement": 0.96, "achievement": 0.96, "score": 1.32, "status": "YES",
        "assessed_weight": 1.375, "assessed_score": 1.32, "explanation": "", "recommendation": "", "methods": ["D", "W"],
        "needs_explanation": false, "evidence_count": 1, "default_owner_roles": ["EXECUTIVE"], "assignees": [],
        "progress": {"elements_total": 3, "elements_assessed": 3}, "updated_by_name": "x", "updated_at": "...",
        "elements": [{
          "code": "1.1.1.1", "type": "PRESENT", "description_en": "...", "description_id": "...", "weight": 0.459,
          "default_methods": ["D"], "methods": ["D"], "scale_type": "DISCRETE", "requires_independent_assessor": false,
          "achievement": 0.75, "score": 0.3443, "explanation": "", "recommendation": "", "needs_explanation": true,
          "evidence_count": 0, "can_edit": true, "updated_by_name": "x", "updated_at": "..."
        }]
      }]
    }]
  }]
}
```
`elements` is present only with `expand=elements`. `basis`: `ELEMENTS` as soon as one element has a value (then the PoF is computed from its elements and `direct_achievement` is ignored), `POF` when only the PoF-level value is set, `NONE` otherwise. `status` of a node that is not complete: `NOT_ASSESSED` (nothing yet) or `IN_PROGRESS` (some). `can_edit` is true when the caller may write that element now (state, lock, assignment and independence rules, section 4).

### GET `/:id/pof/:code`
One PoF (`1.1.1`) with all elements, its evidence list and assignments (same PoF object as above, always with `elements`, plus `evidence: [...]` and `assignments: [...]`).

### GET `/:id/pof-summary`
Compact roll-up equal to the "Lampiran" summary: `{"total": {...}, "components": [{code, name_en, weight, score, achievement, status, progress}], "principles": [{code, component_code, name_en, weight, score, achievement, status, progress}], "progress": {"pofs_total", "pofs_not_assessed", "pofs_in_progress", "pofs_assessed", "by_status": {"YES": 0, "PARTIAL": 0, "NO": 0}, "elements_total", "elements_assessed", "missing_explanation": 0}, "needs_attention": ["1.2.3", "..."]}`. `needs_attention` lists PoF codes that are `NO`, or `PARTIAL`, or assessed below 1 without an explanation.

## 4. Writing values

### PUT `/:id/pof-values`
Atomic bulk write (one transaction; any invalid item rejects the whole request with 422 and `fields`). Up to 500 items in total. Body:
```json
{
  "pofs": [{"pof_code": "1.1.1", "achievement": 0.96, "explanation": "...", "recommendation": "...", "methods": ["D", "W"]}],
  "elements": [{"element_code": "1.1.1.1", "achievement": 0.75, "explanation": "...", "recommendation": "...", "methods": ["D"]},
               {"element_code": "1.1.1.2", "achievement": null}]
}
```
- Every field except the code is optional; an omitted field keeps its value, `achievement: null` clears it.
- `achievement` of an element: `DISCRETE` elements accept 0, 0.25, 0.5, 0.75, 1; `QUESTIONNAIRE` elements accept 0 to 1 with at most 3 decimals. A PoF-level value accepts 0 to 1 with at most 3 decimals (it is an aggregate such as 0.96). Anything else: 422 on `elements[i].achievement` / `pofs[i].achievement`.
- A PoF-level `achievement` is refused with 409 `POF_HAS_ELEMENT_SCORES` while any of its elements has a value (elements win); clear the elements first.
- `methods` is a subset of `D K W O` (no duplicates).
- Text fields are plain text, at most 4000 characters. Every cell/field is stored as text; a leading `= + - @` is kept literally and neutralised only on export.
- Allowed while the assessment is `DRAFT` or `REOPENED`, the period is open, and the caller is an assessor (AUDITOR, AUDIT_MANAGER, CAE, ADMIN) **or** is assigned to the principle or PoF of the item (section 5). Otherwise 403 (item not allowed) / 409 (`INVALID_STATUS`, `PERIOD_LOCKED`).
- Elements with `requires_independent_assessor` cannot be written by a caller who holds the CAE role (EXECUTIVE canonicalises to CAE): 403 `FORBIDDEN` with a message naming the rule. Assign them to a user who is not the CAE (QAIP or external).
- Response `data`: `{"version": 3, "total": {...}, "affected": [{"pof": {PoF object without elements}, "principle": {score block}, "component": {score block}}], "derived_conclusion": "INCOMPLETE"}`. The header `version` is not changed by value writes (per-row last write wins; each write is an event with before/after); `submit` still needs the current header `version`.
- Events: one `POF_UPDATE` per request with the changed items (before/after), actor, IP.

### Submit rules (QUANTITATIVE_POF)
`POST /:id/submit` (body `{version, reason?}`) is refused with 422 `VALIDATION_ERROR` and `error.details` (JSON array of strings, same as the qualitative flow) while: any PoF is not fully assessed (including a PoF with only some of its elements assessed); any assessed value below 1 has no explanation; `final_conclusion` is missing, `EFFECTIVE` is chosen with a principle or component `NO`, or the justification rule fails (`conclusion_justification` needed when the final differs from the derived conclusion or the derived one is `REQUIRES_JUDGMENT`). `GET /:id` lists the same strings in `submit_issues`.

Derived conclusion: `INCOMPLETE` while any PoF is open; `NOT_EFFECTIVE` if any component is `NO`; `EFFECTIVE` if all five components are `YES` and no deficiency above DEFICIENCY; otherwise `REQUIRES_JUDGMENT`. A component is `NO` when its achievement is below the partial threshold **or** any of its principles is `NO`; `YES` only when its achievement reaches the yes threshold and every principle is `YES`. Principle severity: `YES` is `NONE`; `PARTIAL` is `DEFICIENCY` at or above the midpoint of the two thresholds, else `SIGNIFICANT`; `NO` is `MAJOR`. Component rating 1 to 5 uses the RCM bands (90, 80, 70, 60 percent), capped at 2 when the component is `NO`.

## 5. Assignment (PIC) and default owners

Every PoF carries `default_owner_roles` (labels, from the catalogue, replaced by the tenant overrides of section 2 when set; copied into the assessment at creation). The default is information only: **no permission is derived from it.** A person is assigned explicitly, per principle or per PoF.

### GET `/:id/pof-assignments`
`data`: `[{"id", "scope_type": "PRINCIPLE|POF", "scope_code": "P01|1.1.1", "assignee_user_id", "assignee_name", "role_label", "due_date": "YYYY-MM-DD|null", "note", "assigned_by_name", "created_at"}]`. Optional `assignee=me`.
### PUT `/:id/pof-assignments`
Assessors (not scoped roles), assessment editable. Body `{"scope_type", "scope_code", "assignee_user_id", "assignee_name", "role_label": "HEAD_OF_IT", "due_date": "2026-12-01", "note": ""}`; upsert by (scope, assignee). `assignee_user_id` is the auth-service user id (JWT `sub`); `role_label` is a free label (`^[A-Z][A-Z0-9_]{1,39}$`) or an auth-service role name, never a new role. Multiple assignees per scope are allowed. 422 for an unknown scope code.
### DELETE `/:id/pof-assignments/:assignment_id`
Assessors; removes the assignment.

An assignee with a scoped role (DEPARTMENT_HEAD, AUDITEE) can read only the principles and PoFs assigned to them (`GET /:id/pof`, `/:id/pof/:code`, `pof-evidence`) and write values for them; they cannot read the rest of an assessment that is not APPROVED.

## 6. Evidence

Evidence is attached to a PoF or an element: an uploaded file (stored in the database, not on disk, so it survives container recreation) or a reference (document number, link text).

- `GET /:id/pof-evidence?scope_code=1.1.1.2` list (metadata only): `[{"id", "scope_type": "POF|ELEMENT", "scope_code", "kind": "FILE|REFERENCE", "title", "doc_number", "note", "file_name", "content_type", "size_bytes", "sha256", "source": "MANUAL|IMPORT", "import_id", "uploaded_by_name", "created_at"}]`.
- `POST /:id/pof-evidence`: `multipart/form-data` with `file` plus `scope_code`, `title`, `doc_number`, `note`; or `application/json` `{"scope_code", "title", "doc_number", "note"}` for a reference. Limit 10 MB; extension allowlist (pdf, doc(x), xls(x), ppt(x), csv, txt, images), magic-byte check and ClamAV like every other upload. 413 / 415 / 422 `MALWARE_DETECTED` / 503 `SCANNER_UNAVAILABLE`. Allowed to the same callers as value writes.
- `GET /:id/pof-evidence/:evidence_id/download`: the file (attachment, nosniff). VIEWER: 403.
- `DELETE /:id/pof-evidence/:evidence_id`: assessors, assessment editable.

## 7. Excel export

- `GET /:id/pof-export?lang=id|en&blank=true|false` returns an `.xlsx` (`Content-Disposition: attachment`). `blank=true` leaves the achievement/explanation/recommendation columns empty (the template, filled with the same structure).
- `GET /pof-export-template?lang=id|en` returns the empty template for the current catalogue (no assessment needed).
- Layout: sheet `Ringkasan` / `Summary` (equal to the Lampiran: component and principle rows with weight, score, achievement, status, plus the total), one sheet per component (`CE`, `RA`, `CA`, `IC`, `MA`), and a hidden sheet `_meta` (format marker `AUDITSPHERE_COSO_POF`, version 1, catalogue version, assessment id). Component sheet columns: `No | Uraian | Tipe | Metode (D/K/W/O) | Bobot | Capaian | Skor | Penjelasan | Saran`, with PoF rows in bold (their capaian/skor are the roll-up) and principle rows as group headers. `Skor` is written as a value, never as a formula. Every text cell that starts with `=`, `+`, `-`, `@`, a tab or a carriage return is prefixed with an apostrophe-safe single quote so a spreadsheet never evaluates it.
- Only readers who may read the assessment can export; VIEWER is 403.

## 8. Excel import

Three steps: preview, optional remap, apply. Import never evaluates formulas or macros: only the cached **values** of cells are read as text (`.xlsm`, `.xlsb`, `.xls` and any file whose zip contains `vbaProject.bin` are refused). Limits: 5 MB, 25 sheets, 5000 rows and 40 columns per sheet, a 2 MB cap on the sum of the decompressed XML parts read (zip-bomb guard).

### POST `/:id/pof-import/preview`
`multipart/form-data`: `file` (required), optional `format` (`auto` default, `auditsphere`, `external`), `profile_id`, `mapping` (JSON string, overrides the profile), `sheet` (name). Assessors only, assessment editable. The original file is stored with the import record (as evidence when applied). Response `data`:
```json
{
  "import_id": "uuid", "status": "PREVIEWED", "format": "AUDITSPHERE|EXTERNAL", "file_name": "kke.xlsx", "sha256": "...",
  "sheets": [{"name": "CE", "rows": 120, "selected": true}], "sheet": "CE", "header_row": 4,
  "profile": {"id": "uuid|null", "name": "...", "matched": true},
  "mapping": {"header_row": 4, "numbering": "no", "columns": {"no": 1, "description": 2, "type": 3, "methods": 4, "weight": 5, "achievement": 6, "score": 7, "explanation": 8, "recommendation": 9}},
  "columns": [{"index": 1, "letter": "A", "header": "No", "mapped_to": "no"}],
  "summary": {"rows_read": 190, "rows_with_number": 181, "matched_by_code": 170, "matched_by_text": 6, "unmatched": 5, "no_value": 20, "will_update": 156, "conflicts": 3, "weight_mismatches": 2},
  "rows": [{"row": 12, "sheet": "CE", "no": "1.1.1.1", "description": "...", "level": "ELEMENT", "block": "PRESENT", "target": {"code": "1.1.1.1", "level": "ELEMENT"}, "match": "CODE|TEXT|NONE", "similarity": 1.0,
            "achievement": 0.75, "explanation": "...", "recommendation": "", "current": {"achievement": 0.5, "explanation": ""}, "action": "UPDATE|SAME|SKIP|CONFLICT|UNMATCHED", "issues": ["achievement 0.8 is not on the discrete scale"]}],
  "unmatched": [{"row": 47, "no": "9.9.9", "description": "...", "reason": "no catalogue code and no text match above 0.60"}],
  "warnings": ["sheet 'Lampiran' skipped: no header row found"]
}
```
- `format: auto`: a workbook with the `_meta` sheet marker is `AUDITSPHERE` and is read by the exact layout of section 7. Anything else is `EXTERNAL`.
- `EXTERNAL` (KKE-style working papers): per sheet the parser finds the header row (the first of the first 30 rows with at least 3 recognised header words such as `no`, `uraian/description/indikator`, `bobot/weight`, `capaian/hasil/nilai`, `skor/score`, `penjelasan/keterangan/explanation`, `saran/rekomendasi`, `tipe/present/function`, `metode`), maps columns by header words (Indonesian and English, case/space/punctuation-insensitive), reads the hierarchy from the numbering (`^\d+(\.\d+){0,3}$`: 1 digit group = component, 2 = principle, 3 = PoF, 4 = element), detects the `Present` / `Function` block from rows whose description is just that word (or from the type column), and carries values down as the sheet is read. Rows without a number are kept only as block markers or reported.
- Matching to the assessment: by exact number first (`match: CODE`); when a number is absent or unknown, by normalised text similarity (token Jaccard, accent/case/punctuation insensitive) to the element/PoF description in the assessment's language, accepted at or above 0.60 and unique (`match: TEXT`, listed so the user can accept or reject). A number that exists but whose text is clearly different (similarity below 0.30 on a non-empty description) is flagged `CONFLICT` and is not written unless the user accepts it. **Rows that do not match are listed in `unmatched` and never written silently.**
- `action`: `UPDATE` (will change), `SAME` (equal to the current value), `SKIP` (no value in the file), `CONFLICT`, `UNMATCHED`. `issues` carries per-row problems (value outside the scale, not a number, weight different from the catalogue). A file weight that differs from the catalogue weight is reported (`weight_mismatches`) but never applied: weights come from the catalogue version of the assessment.
- Values: decimal comma and percent are understood (`0,75`, `75%`) as text conversions; a cell that is a formula is read by its cached value; an error cell (`#N/A`) is `no_value` with an issue.

### POST `/:id/pof-import/:import_id/remap`
Body `{"mapping": {...}, "sheet": "CE", "profile_id": "uuid|null"}`: re-runs the preview on the stored file with a different column mapping or header row (the "pemetaan kolom" step). Same response as preview. 409 `IMPORT_ALREADY_APPLIED` after apply.

### POST `/:id/pof-import/:import_id/apply`
Body:
```json
{"mode": "fill_empty|overwrite", "accept_text_matches": true, "accept_conflicts": ["1.2.3.1"], "only_rows": [12, 13], "save_profile_as": "KKE format 2025"}
```
- `mode`: `fill_empty` writes only where the assessment has no value yet (default); `overwrite` replaces values. Never applies to an element the caller may not write (independence rule), those rows come back as `skipped_forbidden`.
- `accept_text_matches` (default false): rows matched by text are written only when true (or listed by `only_rows`).
- `save_profile_as`: stores the mapping that was used as a reusable profile (section 9) for the tenant; same name updates it.
- Same permission and state rules as `PUT /:id/pof-values`, atomic. Applies the discrete-scale validation: a row with an invalid value is **not written** and is returned in `rejected` with the reason (the other rows are applied).
- The original file is attached as evidence to the assessment (`source: IMPORT`). Event `POF_IMPORT` records counts, file hash and profile.
- Response `data`: `{"import_id", "status": "APPLIED", "applied": {"elements": 150, "pofs": 6}, "unchanged": 10, "skipped_forbidden": [...], "rejected": [{"row": 40, "code": "1.3.2.1", "reason": "..."}], "unmatched": [...], "profile_saved": {"id", "name"}, "total": {score block}}`.

### GET `/:id/pof-import/:import_id` and `/download`
The stored preview; `.../download` returns the original file (VIEWER 403).
### DELETE `/:id/pof-import/:import_id`
Discards a `PREVIEWED` import (file removed). Applied imports are kept as evidence.

## 9. Import mapping profiles ("the software learns the format")

A profile is a saved mapping (sheet name pattern, header row, column positions or header words, numbering mode) per tenant. Preview with `profile_id` uses it; with no profile and `format: external` the server first tries every saved profile whose `sheet_pattern` and header words fit (best fit wins, reported in `profile.matched`), then falls back to header detection. `use_count` and `last_used_at` are updated on apply.

- `GET /pof-import-profiles`: `[{"id", "name", "description", "format": "EXTERNAL", "mapping": {...}, "use_count", "last_used_at", "created_by_name", "created_at"}]` (any reader).
- `POST /pof-import-profiles`: assessors. Body `{"name", "description", "mapping": {"sheet_pattern": "^(CE|RA|CA|IC|MA)$", "header_row": 4, "numbering": "no", "columns": {"no": 1, "description": 2, ...}, "header_words": {"achievement": ["capaian", "hasil"]}}}`. Name unique per tenant (case-insensitive), 422 `fields` otherwise. Columns are 1-based indexes (0/absent = not present).
- `PUT /pof-import-profiles/:profile_id`, `DELETE /pof-import-profiles/:profile_id` (soft delete). Assessors.

## 10. Workflow summary

`DRAFT -> SUBMITTED -> APPROVED`, `SUBMITTED -> DRAFT` (return), `APPROVED -> REOPENED -> SUBMITTED`: unchanged. Values can be written only in `DRAFT` and `REOPENED`. The approver is not the submitter. A locked period refuses every write. `allowed_actions` in `GET /:id` follows the same rules and does not list value writes (they depend on the item).

## 11. Not in scope / known limits

- Elements cannot be marked not applicable (the weights would have to be redistributed); a requirement that does not apply is scored per the organisation's rule and explained.
- No automatic proposal from system signals for this method (phase 3 of the research doc).
- Catalogue weights are per catalogue version and are not edited per tenant; a new version is a new seed.
- `.xls` and `.xlsm` import is refused by design.
