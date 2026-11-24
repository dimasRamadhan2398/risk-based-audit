# Office structure: frontend plan

Status: in implementation, 2026-11-24. Parent spec: [office-structure-plan.md](office-structure-plan.md) (decisions D1–D8, API contract 3.1–3.2, error codes 2.4). This document covers phase 2 (frontend) and lists what waits for phase 3.

## 1. Current frontend (surveyed)

| Area | Files | Today |
|---|---|---|
| Location master | `pages/master/location/index.vue`, `components/master/LocationTable.vue`, `components/master/LocationForm.vue`, `stores/location.ts` (setup store), `composables/useLocationApi.ts`, `types/master.ts` (`Location*`) | Flat list, client-side search and paging through `TableEntities`. Form: name, address, city, province, postal code, country, status. `useLocationApi` still hand-rolls `Authorization` headers (the `api-auth` plugin already attaches the token). |
| Company master | `pages/master/company/index.vue`, `components/master/CompanyTable.vue`, `components/master/CompanyForm.vue`, `stores/company.ts`, `composables/useCompanyApi.ts` | List + modal only, **no detail page**. Form offers `HOLDING`, `SUBSIDIARY`, `BRANCH`; payload carries `location_id` (no field in the form, but it is resent from the row). |
| Shared options | `stores/master-options.ts` | Loads companies, departments, job roles, locations, employees, business units once; `locationOptions` label = `name — city`. |
| Employee | `components/master/EmployeeForm.vue` | Work location `USelectMenu` over `masterOptions.locationOptions`. |
| Branch pickers (value = `location_id`) | `components/risk-profile/RiskHeatMap.vue` (filter + add/edit modals, `USelect`), `pages/risk-profile/risk-control-matrix/index.vue` (filter `USelect` via `utils/rcmBranch.ts#buildRcmBranchOptions`, form "own branch" `USelectMenu`), `pages/internal-control-assessment/index.vue` (branch filter `USelect`) | Labels are the bare office name; `USelect` has no search. All read `stores/risk-profile.ts#locations` (`{id,name}` of active rows) except ICA which calls `useLocationApi` directly. |
| Name-based consumers (no picker change) | `components/working-paper/WorkingPaperHeaderForm.vue` + `stores/working-paper.ts` (header `location` is a **name string**, matched by name against risks), `pages/risk-profile/audit-universe/index.vue` (syncs names), `composables/useMappedDataImporter.ts` / `components/common/MappedSourceDataBanner.vue` (resolve names to ids) | Left as is: changing the label would change the stored value or the match key. |
| Org levels UI | none | `org_levels` has no screen yet. Department `level` is a plain number input. No tree component in use; Nuxt UI v4 ships `UTree` but its row slots are awkward for per-row actions, so we use a flattened tree (rows with depth) rendered in a list. |
| Map | none | No map library in `package.json` (only chart.js / nuxt-charts). |
| RBAC | `useRbac().canWriteMasterData` (`MASTER_DATA_WRITE_ROLES = ['ADMIN']`, mirrors backend `masterDataWriters`) | Used by every master table to hide add/edit/delete. VIEWER is blocked from `/master/*` by `middleware/02-viewer-readonly.global.ts`. |
| Errors | `utils/error.ts#getUserErrorMessage` | Already maps a backend `code` to `errors.codes.<CODE>`; we only add translations. |

## 2. Design decisions

- **One office form** (`LocationForm.vue`, extended) used by the master page and the company "Kantor" tab. It emits `saved`; each host reloads its own list. Store create/update no longer refetch the master list themselves.
- **Tree built client-side** from the flat list (`parent_location_id`), via `utils/officeTree.ts`. `tree=true` is not needed: the flat response already has every field, filters keep working, and an office whose parent is filtered out shows as a root. Pure functions → unit-tested.
- **Tenant default level scheme lives on Master → Kantor** (second tab "Level Kantor default"), not in Settings. Settings is per-user (profile, MFA, devices, system logs) with its own viewer tab policy; the scheme is master data written by the same role (`masterDataWriters`) and belongs next to the offices it shapes.
- **Map: no new dependency in phase 2.** A map needs a tile provider (OpenStreetMap tiles are a third-party request from every viewer's browser, with a usage policy) plus Leaflet (~40 kB gz). The "Kantor" tab shows a coordinates card per office with latitude/longitude and an "Open in map" link to `openstreetmap.org` (external, opened in a new tab; it is not an API URL so no token is involved). Leaflet + a self-hosted or commercial tile source can be added later without changing data.
- **Level label**: `name_key` set → localized preset `masterData.office.levelPresets.<key>`; else `name`. Office rows prefer `level_name_key` / `level_name` from the API; the scheme is only needed in forms.
- **Partial PUT**: update sends only fields that changed, **plus `company_id` always** (rule `OFFICE_COMPANY_REQUIRED` applies to update). `shared_company_ids` is sent only when the set changed (it replaces the set). Cleared nullable fields are sent as `null`.
- **"Tanpa company"** filter: the API has no `company_id=null` query; the page fetches without `company_id` and keeps rows with no `company_id` client-side.
- **Affected companies warning**: before saving a scheme, the store loads the effective scheme of each candidate company (descendants of the company, or all companies for the tenant default) and counts those with `inherited = true` and `source_company_id` = the scheme being edited. Requests run in parallel, only when the confirm opens.

## 3. Files

### 3.1 Types — `types/master.ts`

- `OfficeFunction` (`head_office | rnd | operational | warehouse | data_center | other`) + `OFFICE_FUNCTIONS`; `OfficeLevelNameKey` (`head_office | regional_office | area_office | branch | sub_branch | unit`) + `OFFICE_LEVEL_NAME_KEYS`.
- `Location` gains the 3.1 fields: `latitude`, `longitude`, `company_id`, `company_name`, `parent_location_id`, `level`, `level_name`, `level_name_key`, `office_code`, `office_function`, `head_employee_id`, `head_employee_name`, `shared_company_ids`, `path: OfficePathItem[]`, `children_count`, `shared`, `children`. `type Office = Location`.
- `LocationFormState`, `CreateLocationRequest`, `UpdateLocationRequest` gain the writable fields. `LocationListQuery` = the GET query params.
- `OfficeLevel`, `OfficeLevelInput`, `OfficeLevelScheme { source_company_id, inherited, levels }`.
- `CreateCompanyRequest` / `UpdateCompanyRequest` lose `location_id`; `Company.location_id` stays (read-only, deprecated).

### 3.2 API — `composables/useLocationApi.ts`, `composables/useOfficeLevelApi.ts` (new)

- `useLocationApi`: drop hand-rolled auth headers; `getLocations(query?)` passes `company_id, level, office_function, parent_id, q, is_active, tree` through ofetch `query`; add `getDescendants(id)`. `getAllLocations` unchanged (never throws).
- `useOfficeLevelApi`: `getCompanyScheme(id)`, `saveCompanyScheme(id, levels)` (PUT), `resetCompanyScheme(id)` (DELETE), `getDefaultScheme()`, `saveDefaultScheme(levels)`. Normalizes `data` being either the scheme object or a bare level array.

### 3.3 Utils — `utils/officeTree.ts` (new)

`buildOfficeTree`, `flattenOfficeTree(roots, expanded)`, `descendantIds`, `officePathNames`, `officePathLabel`, `toOfficeOption` (`{ label: path, value: id, description: office_code }`), `sortOffices`, `levelDisplayName`, `normalizeLevels`, `validateLevelScheme` (returns an i18n key), `nextLevelNumber`, `officeMapUrl`.

### 3.4 Stores

- `stores/location.ts` (setup store, extended): filters `companyFilter` (`''` all, `'__none__'` Tanpa company, or id), `levelFilter`, `functionFilter`, `statusFilter`, `viewMode` (`table|tree`), `expandedIds`; `filteredLocations` keeps client-side search (name, code, city, province, address, path); `officeTree` / `treeRows`; form with all new fields; `openCreateModal(preset?)` (company / parent preset for "add child"); snapshot of the edited row for the partial PUT; `parentCandidates` + `levelScheme` loaded when the form's company changes; `toggleActive(office)`; `branchNames` kept.
- `stores/office-levels.ts` (new, setup store): `schemes` cache by company id, `defaultScheme`, `fetchScheme`, `saveScheme`, `resetScheme`, `fetchDefault`, `saveDefault`, `countAffected(sourceCompanyId | null, companies)`, `activeLevels(companyId)`.
- `stores/master-options.ts`: `locationOptions` → path labels + code description (sorted by path); add `invalidate()` so pickers reload after office writes.
- `stores/risk-profile.ts`: keep `locations` (`{id,name}`) for the existing name logic; add `locationOptions` (path label, code) built from the same active rows.
- `stores/company.ts`: drop `location_id` from form/payloads; keep a legacy `BRANCH` value only when editing a company that still has it.

### 3.5 Components / pages

| File | Change |
|---|---|
| `pages/master/location/index.vue` | Title "Kantor"/"Offices"; `UTabs`: "Daftar Kantor" (table) and "Level Kantor default" (`OfficeLevelEditor` mode `default`). |
| `components/master/LocationTable.vue` | Filters (company incl. "Tanpa company", level — enabled when a company is chosen, from its effective scheme —, function, status), search, table/tree toggle. Columns: office (+ path), code, company, level, function, city, status, actions. Tree view uses `OfficeTreeList`. Loading / error / empty states kept. |
| `components/master/LocationForm.vue` | Sections: Kantor (company *, parent, level, code, function, head employee, shared companies), Alamat (address, city, province, postal code, country, latitude, longitude), status. Company locked when editing an office with children. Parent options = same company, owned only, minus self and descendants. Level options = active levels of the company's effective scheme above the parent's level. Emits `saved`. |
| `components/master/office/OfficeTreeList.vue` (new) | Flattened tree rows: chevron, name, code, level badge, function badge, head, status; actions (writers only): add child, edit, activate/deactivate. |
| `components/master/office/OfficeLevelEditor.vue` (new) | Editable table of levels (number, name, preset `name_key`, active), add/remove rows, validation, save with confirm that names the number of affected inheriting companies. Mode `company` shows inherited state + source company, "Buat skema sendiri" / "Kembali ikut induk". |
| `components/master/office/CompanyOfficesTab.vue` (new) | Owned office tree (add root/child, edit, deactivate, re-parent via form), "Kantor bersama" list (owner company shown), coordinates cards with map link. |
| `pages/master/company/[id].vue` (new) | Company detail: header, `UTabs` Profil / Kantor / Level Kantor. Back link to list. Edit opens the existing `CompanyForm` modal. |
| `components/master/CompanyTable.vue` | Company name links to the detail page; eye action. |
| `components/master/CompanyForm.vue` | Type options: Holding, Subsidiary only (legacy Branch shown disabled for a company that still has it). |
| `components/master/EmployeeForm.vue` | Work location picker gets path labels + code search through `masterOptions.locationOptions` (`filter-fields`). |
| `components/risk-profile/RiskHeatMap.vue` | Branch filter and add/edit branch selects → `USelectMenu` with path labels, searchable by name/code. |
| `pages/risk-profile/risk-control-matrix/index.vue` + `utils/rcmBranch.ts` | Filter and form "own branch" use path labels; `buildRcmBranchOptions` takes an optional `label`. |
| `pages/internal-control-assessment/index.vue` | Branch filter → searchable `USelectMenu` with path labels. |

Pickers **not** changed: working paper header (stores the name), audit universe sync, mapped data importer (name matching).

### 3.6 Error codes → i18n (`errors.codes.<CODE>`, both locales)

`OFFICE_COMPANY_REQUIRED`, `OFFICE_PARENT_OTHER_COMPANY`, `OFFICE_PARENT_LEVEL`, `OFFICE_CYCLE`, `OFFICE_DEPTH`, `OFFICE_LEVEL_UNKNOWN`, `OFFICE_HEAD_OFFICE_EXISTS`, `OFFICE_CODE_TAKEN`, `OFFICE_HAS_CHILDREN`, `OFFICE_LEVEL_IN_USE`, `COMPANY_TYPE_RETIRED`. `getUserErrorMessage` picks them up from `error.data.error.code` or `error.data.code`.

### 3.7 i18n

New namespaces `masterData.office.*` (form, filters, tree, function and level preset names, map, shared list), `masterData.officeLevels.*` (editor, inherited banner, confirmations), `masterData.companyDetail.*` (tabs, header); `masterData.location.*` titles reworded to "Kantor"/"Offices"; `navigation.location` → "Kantor"/"Offices". Both `locales/en/common.json` and `locales/id/common.json`, key parity tested.

### 3.8 RBAC

All write controls (add, add child, edit, (de)activate, delete, level editor save / own scheme / reset) render only when `canWriteMasterData`. Readers see the same screens read-only. The backend enforces `masterDataWriters` regardless.

## 4. Tests (vitest, `tests/unit`)

- `officeTree.test.ts`: tree build (orphans become roots, sort by level then name), flatten with expand state, descendants, path label, option label/description, level display name, scheme validation, next level number, map URL.
- `officeLevels.test.ts`: GET scheme (inherited + source), PUT body shape, DELETE then refetch, 409 `OFFICE_LEVEL_IN_USE` message, default scheme, affected company count.
- `locationMaster.test.ts` (updated): company required, create payload with new fields, PUT sends only changed fields + `company_id`, `shared_company_ids` only when changed, query params for filters, "Tanpa company" filter, error codes translated, i18n parity for new namespaces and every code.
- `companyOfficeFields.test.ts`: company payload has no `location_id`; form offers no `BRANCH`.

## 5. Risks

- **Contract drift**: built against 3.1–3.2 while the backend is in progress. Assumptions (section 6) are isolated in `useLocationApi` / `useOfficeLevelApi` normalizers.
- **Affected-company count** costs one request per candidate company; acceptable for tenant-sized company lists, revisit if a tenant has hundreds.
- **Partial PUT with `null`**: if the backend ignores `null` instead of clearing, clearing a parent/level/code/head would silently do nothing — needs a manual check.
- **Large office lists**: the tree and table are client-side over the full list; fine for hundreds, not tens of thousands.
- Pickers now show long path labels; the RCM and heat map filters are wider-content selects — checked to wrap, but needs a visual check.

## 6. API assumptions

1. Errors carry `code` either at `data.error.code` or `data.code` (both handled).
2. `GET /office-levels/default` returns either the scheme object or `{ levels }` or a bare array; PUT body `{ levels }`.
3. Explicit `null` clears `parent_location_id`, `level`, `office_code`, `head_employee_id`, `latitude`, `longitude` on PUT.
4. `GET /locations?company_id=X` marks shared rows with `shared: true`; rows without it are owned.
5. Level numbers are positive integers; the UI suggests gaps of 10.

## 7. Deferred to phase 3

- "Include sub-offices" toggle on branch filters (Risk Profile, RCM, heat map, ICA) using `GET /locations/:id/descendants` (the API method is added now, unused).
- Kafka `location.events` consumers.
- A real map (Leaflet + tile source decision).
- Drag-and-drop re-parenting in the tree.
- Phase 4: remove `Company.location_id` from types and the legacy `BRANCH` display.
