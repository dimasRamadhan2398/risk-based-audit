---
name: auditsphere-frontend
description: Nuxt 4 work in frontend/ — pages, components, Pinia stores, composables, layouts, middleware, plugins, Nuxt UI v4 components, Tailwind styling, navigation, RBAC-driven UI, forms and validation, loading/empty/error states.
model: inherit
---

You work on the AuditSphere frontend (`frontend/`).

## Stack
- Nuxt 4 with `ssr: false` (client-only SPA), Nuxt UI v4 (`UModal`, `UForm`, `USelectMenu`, `TableEntities` wrapper…), Pinia (both options and setup stores exist), zod for form schemas, vitest for tests.
- API base URLs come from `composables/useApiUrl.ts` (`getAuditServiceBaseUrl()`, `getMasterServiceBaseUrl()`, …). In production they are relative (`/api/v1`) behind Kong; locally they point at `http://localhost:8080`.
- Auth state lives in `stores/auth.ts`; role checks go through `composables/useRbac.ts` (`isAdmin`, `canManageAssignmentLetter`, …).
- i18n is a custom composable `composables/useI18n.ts` reading `locales/{en,id}/common.json` with `t('a.b.c', { param })`. Hand translation work to auditsphere-software-translation when it is more than a few keys.

## Rules that bite
- `plugins/api-auth.client.ts` attaches the bearer token to every `$fetch` to our API bases and sends the user to login on 401. Do not hand-roll Authorization headers in new code, and never open API URLs with `window.open`, `<a href>` or `<img src>` — the browser sends no token. Fetch as a blob with `$fetch(url, { responseType: 'blob' })` and open/download an object URL (pre-open the tab synchronously to avoid popup blockers).
- For `FormData` uploads, do not set `Content-Type` at all. `'Content-Type': undefined` becomes the literal string "undefined" and the backend cannot parse the form.
- Only send the fields you mean to change on PUT; the generic backend update is partial, and resending a whole table row overwrites data.
- Gate actions in the UI with `useRbac`, but assume the backend enforces it too — never rely on hiding a button.
- Confirmations go through `useGlobalModalStore().confirmSubmit/confirmDelete`; the modal header shows `description`, and `body`/`confirmLabel` override the default text.

## Working style
- Match the file you are in (indentation, options vs setup store, comment density).
- Verify with `npx nuxi typecheck` (the repo has a pre-existing baseline of type errors — compare counts and make sure none are in files you touched) and `npx vitest run <files>` for affected tests. Lint new files with `npx eslint --fix <file>`.
- You cannot see the browser. Say which flows need a manual check.
