# Employee code migration plan: standardise on EMP-0001

Status: PLAN ONLY. Nothing in this document has been executed. No writes to any
database (local, staging or the production VPS). Every query below is a read-only
SELECT and must be run inside a read-only session.

Canonical format: `EMP-` followed by a zero-padded number of at least four digits
(`EMP-0001`, `EMP-0004`, `EMP-10000`). This is what `services/employee/employee.go`
generates (`employeeCodePrefix = "EMP-"`, padding in `employeeCodeDigits`) and what
`models/employee.go` documents (`// e.g. "EMP-0001"`).

---

## 1. Current state (verified in the repo)

### 1.1 Formats in use

| Format | Where | Example |
|---|---|---|
| `EMP-NNNN` (canonical) | master-service generator; master seeder rows at `pkg/database/seeders/seeder.go` lines 509-594 and every later `getByName(... "EMP-000n")` call | `EMP-0001` Ahmad Wijaya ... `EMP-0006` Budi Santoso |
| `EMPNNN` (legacy, no dash) | master seeder rows at lines 611, 628, 645, 662, 679, 696; auth seeder at `pkg/database/seeders/seeder.go` lines 229-234 | `EMP001` System Administrator, `EMP004` Bob Auditee, `EMP007` New Department Head |
| `EMP-NNN` (3 digits, with dash) | `backend/auth-service/scripts/seed_admin.sql` line 23 | `EMP-001` |

### 1.2 The seeded data already collides by person, not by string

The master seeder creates two different sets of people:

- `EMP-0001`..`EMP-0006`: Ahmad Wijaya, Citra Dewi, Dian Pratama, Fitri Handayani, Hendra Wijaya, Budi Santoso.
- `EMP001`..`EMP007`: System Administrator, John Auditor, Alice Department Head, Bob Auditee, Carol Viewer, New Department Head.

So a digit-based rewrite is wrong:

- `EMP001` -> `EMP-0001` would merge System Administrator into Ahmad Wijaya.
- `EMP004` -> `EMP-0004` would merge Bob Auditee into Fitri Handayani (in the seed data).
- `EMP005` -> `EMP-0005` would merge Carol Viewer into Hendra Wijaya.

The mapping must follow the person (email, full name, and the user link), never the digits.

### 1.3 Link between the two services

- `users.employee_id` (auth-service, `models/user.go` line 13, `uniqueIndex`) = `employees.employee_code` (master-service). There is no FK across databases.
- auth-service resolves the link through master-service `GET /api/v1/employees/by-code/{code}` (`pkg/masterclient/employee.go`, `url.PathEscape(code)`). `/auth/me` and the `employee_code` / `employee_id` / `department_id` fields in the response depend on this call.
- The JWT carries `employee_code` (`pkg/utils/jwt.go`, claim `employee_code,omitempty`). audit-service reads it in `pkg/middleware/auth.go` and sets `employee_code` on the context.
- Both `employees.employee_code` and `users.employee_id` have unique indexes, so a rename cannot create a transient duplicate in one statement (see section 4).

### 1.4 Generator

`services/employee/employee.go`:

- `nextEmployeeCode(existing []string)` only reads codes that start with `EMP-` followed by digits. Legacy `EMP004` is ignored. Gaps are not filled.
- `createWithGeneratedCode` and `NextEmployeeCode` both call `repositories/employee` `FindCodesWithPrefix("EMP-")`, which uses `LIKE 'EMP-%'`, including soft-deleted rows (the unique index counts them).
- `Create` accepts an explicit `employee_code` with no format check. An operator can still create `EMP004` by hand.

Note: `services/employee/employee.go`, its tests and the repository/controller files are already modified in the working tree (`git status`). The generator change in section 5 must be merged with that pending work.

### 1.5 Other places that store or show employee codes

| Location | Stores code? | Breaks if codes change? | Action |
|---|---|---|---|
| master `employees.employee_code` | yes (source) | - | rename (section 4) |
| auth `users.employee_id` | yes (link) | yes, `/auth/me` degrades to "link unavailable" until updated | rename in lockstep (section 4) |
| auth `masterclient` `FindByCode` | calls by code | yes during the window only | none after migration |
| JWT `employee_code` claim | yes (issued at login) | stale tokens keep the old code until expiry or re-login | ask users to re-login after cut-over |
| audit-service `pkg/base/base_controller.go` `GetEmployeeID` | uses `employee_id` UUID from token | no | none |
| audit-service `models/assignment-letter.go` `LetterMember` | `Name`, `Role` only (free text) | no | none |
| audit-service `cmd/seed.go` team/letter member names | names only | no | none |
| master `services/datahub_sync.go` `EmployeeSync.EmployeeIDNumber` | field declared, no caller in repo populates it | unknown on the Data Hub side | confirm with the Data Hub owner what they expect (open question) |
| frontend (`types/master.ts`, `types/auth.ts`, `stores/employee.ts`, `components/master/Employee*.vue`, `composables/useEmployeeApi.ts`, `composables/useUserAccountApi.ts`) | free-text field, no hard-coded EMP literals found | no | none, unless a format validator is added (section 5) |
| `docs/API_Testing_Guide.md`, `docs/API_Testing_Guide.html`, `docs/Postman_Collection_RBIA_Master_Service.json` | examples use `EMP001` | docs only | update examples |
| `backend/auth-service/scripts/seed_admin.sql` line 23 | `EMP-001` (third format) | yes if the script is run | update to the chosen code |
| Go test fixtures (master `services/employee/*_test.go`, `repositories/employee/employee_test.go`, `controllers/employee/employee_test.go`, audit-service `pkg/middleware/auth_test.go`, auth-service `pkg/masterclient/employee_test.go`) | fixtures only | tests only | update when the seeder/generator changes |
| `docs/pkat-snapshot-resolution.md`, `scripts/check-kong-routes.sh` (untracked) | no EMP references found | no | none |
| Kong, compose, deploy files | none | none | none (devops territory, not touched) |

Searches covered `backend/`, `frontend/` (excluding node_modules, .nuxt, .output) and `docs/`. The whole-repo grep for `EMP[-_]?\d` found no hits in python-ai, risk-service or analytics-service.

---

## 2. Read-only inventory (run on prod VPS, both databases)

Run each block in a session that refuses writes. Do not run any seeder.

```sql
-- Session guard (read-only transaction for the whole session)
SET default_transaction_read_only = on;
```

### 2.1 Master database: `rb_audit_master_service`

```sql
-- A. Format census
SELECT
  CASE
    WHEN employee_code ~ '^EMP-[0-9]{4,}$' THEN 'canonical'
    WHEN employee_code ~ '^EMP[0-9]+$'     THEN 'legacy_no_dash'
    WHEN employee_code ~ '^EMP-[0-9]+$'    THEN 'dash_short'
    ELSE 'other'
  END AS fmt,
  COUNT(*) AS n
FROM employees
GROUP BY 1
ORDER BY 1;

-- B. Every code, including soft-deleted rows (the unique index counts them)
SELECT id, employee_code, full_name, email, department_id, job_role_id,
       manager_id, is_active, deleted_at, created_at
FROM employees
ORDER BY employee_code;

-- C. Codes with whitespace or lowercase (would not match a plain string compare)
SELECT id, '[' || employee_code || ']' AS code
FROM employees
WHERE employee_code <> UPPER(BTRIM(employee_code));

-- D. Duplicate emails (identity key for the mapping)
SELECT LOWER(BTRIM(email)) AS email_key, COUNT(*) AS n, ARRAY_AGG(employee_code) AS codes
FROM employees
GROUP BY 1
HAVING COUNT(*) > 1;

-- E. Current max for each numeric family (for allocating targets)
SELECT MAX(CAST(SUBSTRING(employee_code FROM '[0-9]+$') AS INT)) AS max_num
FROM employees
WHERE employee_code ~ '^EMP-?[0-9]+$';
```

### 2.2 Auth database: `rb_audit_auth_service`

```sql
-- F. Every login link, including inactive and soft-deleted accounts
SELECT id, username, email, employee_id, is_active, deleted_at, created_at
FROM users
ORDER BY employee_id NULLS LAST, username;

-- G. Format census of links
SELECT
  CASE
    WHEN employee_id IS NULL                  THEN 'unlinked'
    WHEN employee_id ~ '^EMP-[0-9]{4,}$'      THEN 'canonical'
    WHEN employee_id ~ '^EMP[0-9]+$'          THEN 'legacy_no_dash'
    ELSE 'other'
  END AS fmt,
  COUNT(*) AS n
FROM users
GROUP BY 1
ORDER BY 1;

-- H. Users linked by email fallback only (employee_id empty) - these resolve by email, not code
SELECT id, username, email
FROM users
WHERE employee_id IS NULL OR BTRIM(employee_id) = '';
```

### 2.3 Cross-database check (done in a local script, not in SQL)

Export the results of B and F as CSV with `\copy (...) TO STDOUT CSV HEADER` from each database, then compare locally:

- Every `users.employee_id` that is not NULL must equal exactly one `employees.employee_code`. Report any orphan (link with no master row) and any master row with no user (informational).
- Expected at least this link (user report): `users.employee_id = 'EMP004'` for username `auditee`, which resolves to the master row Bob Auditee, Finance & Accounting.
- Record `department_id` for the auditee row. It is needed for the verification in section 7.

Keep the CSV exports outside the repository (for example `/tmp/...`) and do not commit them.

---

## 3. Mapping rule

### 3.1 Principle

Map by person identity, then by number. Identity is the first matching key of:

1. the user's `email` in auth matches `employees.email` in master (case-insensitive, trimmed), or
2. the `users.employee_id` link already points to the row.

The digits are only used to pick a target code, never to decide that two rows are the same person.

### 3.2 Target code

For each legacy code `EMP<N>` (N digits, as stored), the default target is `EMP-` plus N zero-padded to 4 digits (`EMP004` -> `EMP-0004`). Use the default only if all of these hold:

a. No row in master (including soft-deleted) currently has that target code, or the row that has it is the same person (identity match). A same-person duplicate is a separate decision (section 6), not an automatic merge.
b. No other legacy row maps to the same target.
c. The target is not reserved by a later code (see 3.4).

If any check fails, allocate the next free number instead: `max(canonical numeric) + 1`, then incrementing, in ascending order of the legacy code. Write the allocation into the mapping file.

### 3.3 Mapping file (local, not in the database)

Build one CSV per migration run, reviewed by the user before any write:

```
legacy_code,new_code,master_employee_id,email,full_name,user_id_or_none,rule
EMP004,EMP-0004,<uuid>,auditee@company.com,Bob Auditee,<user uuid>,default_target
EMP001,EMP-0007,<uuid>,admin@company.com,System Administrator,<user uuid>,allocated_collision
```

The example rows show the expected shape only. Whether prod needs `EMP-0004` depends on the section 2 output: if prod already has `EMP-0004` for a different person, the default target is blocked and `EMP004` must be allocated a new number, and the auditee link follows that new number.

### 3.4 Reserved numbers

Before allocating, check the seeder and generator collisions: `EMP-0001`..`EMP-0006` are used by the seed people, and any allocated number must not reuse one of them for a different person. A reserved set should be written down before the run.

### 3.5 Uniqueness

After mapping, all new codes must be distinct and must match `^EMP-[0-9]{4,}$`. Check this in the local script before the first write (no duplicates, no NULLs, no legacy values left in the mapping output).

---

## 4. Order of operations (write phase, only after approval)

There is no cross-database transaction in PostgreSQL for two separate databases, so the link cannot be made atomic. The design is: one transaction per database, master first, a short window, and a reverse mapping for rollback.

### Step 0: Decisions (section 6 must be answered first)

### Step 1: Backup (before any change)

On the VPS, as a backup only, with the output stored off the host:

```bash
pg_dump -Fc -d rb_audit_master_service -f rb_audit_master_service_$(date +%Y%m%d_%H%M).dump
pg_dump -Fc -d rb_audit_auth_service   -f rb_audit_auth_service_$(date +%Y%m%d_%H%M).dump
sha256sum *.dump > backups.sha256
```

Also save the section 2 query output and the mapping CSV next to the dumps. Test-restore both dumps into a scratch database (not prod) and run the verification queries there.

### Step 2: Rehearsal on a restored copy

Restore the dumps to a local or scratch Postgres (never to a prod database). Run the mapping, both transactions and the verification in section 7 on the copy. The output must match the expected mapping file.

### Step 3: Ship the code that reads both formats (safe first)

Deploy the generator change (section 5.1) before the data change. It reads legacy and canonical codes, so it is safe against both states. Deploy with seeding disabled, using the existing no-seed procedure (see the team memory notes on backend prod deploy). Do not run the seeders on prod.

### Step 4: Freeze writes (short window)

Ask devops to stop or pause write traffic to master-service and auth-service for the window (logins, employee create and update). This is an ops action and is not done by this plan. Without a freeze, a create during the window can take a legacy-looking target.

### Step 5: Master transaction (first)

One transaction, two passes, because `employee_code` has a non-deferrable unique index and a direct rename can collide with a row that has not been renamed yet:

```sql
BEGIN;
-- Pass 1: move every legacy code out of the way
UPDATE employees
SET employee_code = 'TMP-' || id::text
WHERE employee_code IN (/* legacy codes from mapping file */);

-- Pass 2: assign the new codes (VALUES list generated from the reviewed mapping file)
UPDATE employees AS e
SET employee_code = m.new_code
FROM (VALUES
  ('<employee uuid>'::uuid, 'EMP-0004')
  -- ... one row per mapping entry
) AS m(id, new_code)
WHERE e.id = m.id;

-- Expect: pass 1 = N rows, pass 2 = N rows. Stop if not.
SELECT COUNT(*) FROM employees WHERE employee_code ~ '^EMP[0-9]+$';   -- expect 0 for mapped rows
COMMIT;   -- only if counts match
```

The VALUES list is generated by a local script from the reviewed CSV. No user input goes into SQL.

### Step 6: Auth transaction (second, same window)

```sql
BEGIN;
UPDATE users SET employee_id = 'TMP-' || id::text
WHERE employee_id IN (/* legacy codes */);

UPDATE users AS u
SET employee_id = m.new_code
FROM (VALUES
  ('<user uuid>'::uuid, 'EMP-0004')
  -- ... one row per linked user in the mapping file
) AS m(id, new_code)
WHERE u.id = m.id;

SELECT COUNT(*) FROM users WHERE employee_id ~ '^EMP[0-9]+$';  -- expect 0
COMMIT;
```

Then run the cross-database check from section 2.3 again. Every link must resolve.

### Step 7: Release the freeze and verify (section 7)

### Why master first

Either order leaves a short window where the link is broken: with master first, auth users still point at the legacy code while master has the new code (`/auth/me` degrades for linked users); with auth first, the reverse. Master is the record of truth for the employee, so it is changed first and checked before auth is touched. The window must be seconds, which is why the steps are scripted and the mapping is reviewed beforehand.

### Rollback

- If master commits and auth fails: run the reverse mapping (`new_code -> legacy_code`) on master, in the same two-pass form, and keep the freeze.
- If auth commits and master fails: the auth transaction is already rolled back (it is in its own transaction); nothing to do but fix the cause and retry.
- If the problem appears after both commits: reverse both (auth first, then master) with the reverse mapping. This is safe only if no new employees or users were created in the window that use the new codes. The freeze prevents that.
- The dumps are the last resort. A full `pg_restore` would overwrite any writes made after the backup, so it is only for disaster recovery with the freeze and a user decision.

---

## 5. Code changes (after the data is migrated, except 5.1)

### 5.1 Generator: account for legacy codes (ship first, step 3)

Current behaviour (`services/employee/employee.go`):

- `FindCodesWithPrefix("EMP-")` reads only `EMP-%`.
- `nextEmployeeCode` takes `max(EMP-<n>) + 1`.

Change:

- Read `employee_code` values matching `EMP%` (both forms), for example `LIKE 'EMP%'` with the same `ESCAPE` discipline, or a second query. Keep the pattern as a bound parameter; never build it from input.
- `nextEmployeeCode` parses both `EMP-<n>` and `EMP<n>` with `^EMP-?([0-9]+)$`, takes the max over both, and returns `EMP-%04d` of `max+1`. This prevents the generator from issuing `EMP-0007` while a legacy `EMP007` is still waiting to be mapped to `EMP-0007`.
- Keep soft-deleted rows in the input, as now.
- Tests: extend the existing `nextEmployeeCode` tests with `EMP007` (expects `EMP-0008`), mixed input, `EMP-` only, and overflow.

This change does not change any stored data and is safe before the migration.

### 5.2 Seeders (after the data migration, in the same release)

- Master seeder `pkg/database/seeders/seeder.go`, lines 611-696: change the six legacy codes to canonical codes that do not collide with `EMP-0001`..`EMP-0006`. Proposed (needs user decision, section 6): `EMP-0007` System Administrator, `EMP-0008` John Auditor, `EMP-0009` Alice Department Head, `EMP-0010` Bob Auditee, `EMP-0011` Carol Viewer, `EMP-0012` New Department Head.
- Auth seeder `pkg/database/seeders/seeder.go`, lines 229-234: same codes, and the `users` links must use the same values. Change both in one commit so the seeded links are never broken.
- The seeders use `FirstOrCreate` by `employee_code` (master line 715). A seeder run against a database that still has the legacy rows would create duplicate people under the new codes. That is one reason seeders must not run on prod, and why the code change must be deployed with seeding disabled before the data migration.
- `scripts/seed_admin.sql` line 23: `EMP-001` -> the chosen code.
- Keep the seeders idempotent (check before create), as the repo rules require.

### 5.3 Optional: format validation on create and update

- Add a format check in `normalizeAndValidate` (`^EMP-[0-9]{4,}$`), so new rows cannot be created in legacy form. This is a behaviour change: edit forms for any row that is still legacy would fail until fixed. Ship it only after the migration is verified, and decide with the user.
- Frontend `EmployeeForm.vue` can show the same hint. Copy goes through the i18n locale files (en and id).

### 5.4 Docs and tests

- Update `docs/API_Testing_Guide.md`, `docs/API_Testing_Guide.html` and `docs/Postman_Collection_RBIA_Master_Service.json` examples from `EMP001` to the chosen canonical code.
- Update the Go test fixtures listed in section 1.5 to the canonical codes.

---

## 6. Risks and decisions needed before any write

### 6.1 Decisions for the user

1. Canonical target for each legacy person, especially where the default `EMP-000N` is taken by another person (section 3.2). Confirm the proposed numbers in 5.2.
2. Freeze: may devops pause write traffic to master-service and auth-service for a short window? Who runs it, and for how long?
3. Scope of prod: are the seeded test accounts (`admin`, `auditor`, `dept_head`, `new_dept_head`, `auditee`, `viewer`) real accounts on prod, or only dev data? If prod has only the auditee/prod users, the mapping is small.
4. Same-person duplicates: if one person has two master rows (for example `EMP004` and `EMP-0004` with the same email), do we merge (re-point data and soft-delete one row), or keep both and link only the one used by the login? Merging is out of this plan's scope unless confirmed.
5. Data Hub: what does the Data Hub expect in `employee_id_number`? Is it the employee code, and does it need to be told about the rename? (`EmployeeSync.EmployeeIDNumber` is declared but not populated in this repo.)
6. Re-login: after the cut-over, active sessions keep the old `employee_code` claim until the access token expires. Accept a forced re-login for everyone, or wait for expiry? Check the configured token expiry first.
7. Format validation (5.3): ship now, later, or never?
8. Does prod allow a short login/`/auth/me` test with the real `auditee` password? Each login writes to the auth database (session, trusted device or audit records), so it is a write. The verification in section 7 should use a dedicated test account or a user-approved login.

### 6.2 Technical risks

- Broken link window: `/auth/me` and the `employee_code` claim degrade for linked users between the two commits. Mitigation: freeze, scripted run, reviewed mapping.
- Unique index collisions in a single UPDATE: avoided by the two-pass rename in section 4.
- Seeder or generator deployed with seeding enabled would reseed prod. Mitigation: no-seed deploy, as in the team memory notes.
- Stale JWT claims after cut-over: see 6.1 item 6. Audit-service uses the `employee_id` UUID for most checks, so the impact should be limited, but confirm the audit-service `employee_code` uses.
- Soft-deleted rows keep their codes in the unique index. The mapping must include them (section 2.1 B).
- Whitespace or case variants (section 2.1 C) would not match the link. Fix them as part of the mapping, not by a LIKE in production.
- Working-tree changes to `services/employee/employee.go` and tests (see 1.4) may conflict with the generator change. Merge them first.
- Docs and Postman examples will be stale until updated; this is cosmetic but visible to new testers.
- No Kong, compose or deploy change is needed. If the devops freeze needs a container stop, that is an ops action owned by auditsphere-devops.

---

## 7. Verification after migration

### 7.1 Data checks (read-only, both DBs)

```sql
-- master: no legacy codes left, no duplicates, canonical format only
SELECT COUNT(*) FROM employees WHERE employee_code !~ '^EMP-[0-9]{4,}$';           -- expect 0 for active rows (soft-deleted rows also listed)
SELECT employee_code, COUNT(*) FROM employees GROUP BY 1 HAVING COUNT(*) > 1;     -- expect 0 rows

-- auth: every link resolves in master (run the cross-DB check from section 2.3)
SELECT COUNT(*) FROM users WHERE employee_id !~ '^EMP-[0-9]{4,}$' AND employee_id IS NOT NULL;  -- expect 0
```

Cross-DB: every `users.employee_id` must match one `employees.employee_code`. Record orphans.

### 7.2 Service checks

1. Master: `GET /api/v1/employees/by-code/EMP-0004` (or the code the mapping chose for Bob Auditee). Expect 200 with `full_name` Bob Auditee and `department_id` of Finance & Accounting. Also `GET /api/v1/employees/by-code/EMP004` must now return 404 (legacy code gone).
2. Auth login as the `auditee` user: `POST /api/v1/auth/login` (auth-service, no Kong JWT check). Use the approved account (section 6.1 item 8).
3. `GET /api/v1/auth/me` with the token from step 2. Expect:
   - `employee_code` = the new code (`EMP-0004` or the allocated code),
   - `employee_id` = Bob Auditee's master UUID,
   - `department_id` = the Finance & Accounting department UUID (the same value as the master row).
4. Audit-service: call one authenticated `/api/v1` endpoint with the same token and confirm a 200 (the token must pass `Authenticate()`).
5. Generator: `GET /api/v1/employees/next-code` (or the route that wraps `NextEmployeeCode`) returns the next canonical code, and it is greater than every migrated code.
6. Frontend: open the employee list, search `EMP-0004`, and check the auditee's account page shows the new code.

### 7.3 Expected negatives

- Old JWTs issued before the cut-over may still show the legacy `employee_code` until they expire or the user logs in again. Expected, not a failure.
- `/auth/me` for a linked user returns "link unavailable" only during the window. Should not persist after step 7.

---

## 8. Summary of what is NOT in this plan

- No write of any kind (no migration run, no seeder run, no database change).
- No Kong, compose, or deploy change (auditsphere-devops).
- No merge of same-person duplicates (needs the user decision in 6.1 item 4).
- No change to the uncommitted working tree.
