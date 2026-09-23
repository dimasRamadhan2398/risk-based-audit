# Nuxt Routing Configuration Guide

## Overview
Nuxt uses **file-based routing**, meaning your URL structure is determined by your folder structure in the `pages/` directory.

---

## Current Routing Structure

### File-Based Routes (Automatic)
```
/pages/
├── index.vue                              → /
├── dashboard/
│   └── index.vue                          → /dashboard
├── audit-result-report/
│   ├── index.vue                          → /audit-result-report
│   ├── executive-summary.vue              → /audit-result-report/executive-summary ❌
│   ├── executive-summary-upload.vue       → /audit-result-report/executive-summary-upload
│   └── satisfaction-survey.vue            → /audit-result-report/satisfaction-survey
├── executive-summary/                     
│   ├── index.vue                          → /executive-summary ✅
│   └── upload.vue                         → /executive-summary/upload
└── master/
    └── employee/
        └── index.vue                      → /master/employee
```

---

## 3 Methods to Change Routes

### Method 1: File Structure (RECOMMENDED)
**Simply move the file to change the URL.**

#### Current State:
```
/pages/audit-result-report/executive-summary.vue  →  /audit-result-report/executive-summary
```

#### To Move to:
```
/pages/executive-summary/index.vue  →  /executive-summary
```

**Steps:**
1. The top-level `/executive-summary` folder already exists with files
2. Check if the content needs to be updated or merged
3. Files can be deleted after verification

---

### Method 2: Route Rules (Route Redirects in nuxt.config.ts)
**Add automatic redirects from old URLs to new ones.**

Edit `/frontend/nuxt.config.ts`:

```typescript
export default defineNuxtConfig({
  routeRules: {
    // Redirect old nested route to new flat route
    '/audit-result-report/executive-summary': { redirect: '/executive-summary' },
    '/audit-result-report/executive-summary-upload': { redirect: '/executive-summary/upload' },
    
    // API proxy rules
    '/api/v1/**': {
      proxy: process.env.API_BASE_URL_SERVER || 'http://localhost:8080/api/v1/**'
    },
  },
  
  // ... rest of config
})
```

**Pros:** 
- ✅ Old links still work (redirects to new URL)
- ✅ SEO friendly (301 redirects)
- ✅ Backward compatible

---

### Method 3: Dynamic Routes (For Parameterized URLs)
**Use square brackets for dynamic segments.**

#### Example:
```
/pages/
├── master/
│   ├── employee/
│   │   ├── index.vue              → /master/employee
│   │   └── [id].vue               → /master/employee/:id
```

In the `[id].vue` file:
```vue
<script setup lang="ts">
const route = useRoute()
const employeeId = route.params.id
</script>

<template>
  <div>Employee ID: {{ employeeId }}</div>
</template>
```

**Usage:** 
- `/master/employee/emp-123` - Routes to `[id].vue` with `id='emp-123'`

---

## Quick Reference: Common Routes

| File Path | URL | Method |
|-----------|-----|--------|
| `pages/index.vue` | `/` | File-based |
| `pages/dashboard/index.vue` | `/dashboard` | File-based |
| `pages/dashboard/[id].vue` | `/dashboard/:id` | Dynamic |
| `pages/admin/settings.vue` | `/admin/settings` | File-based |
| `pages/[...slug].vue` | `/any/nested/path` | Catch-all |

---

## Implementation for Your Case

### Option A: Keep Both (with Redirect)
1. Keep new structure: `/pages/executive-summary/index.vue`
2. Add redirect rule in `nuxt.config.ts`:

```typescript
routeRules: {
  '/audit-result-report/executive-summary': { redirect: '/executive-summary' },
  '/audit-result-report/executive-summary-upload': { redirect: '/executive-summary/upload' },
}
```

### Option B: Replace Old with New
1. Delete `/pages/audit-result-report/executive-summary.vue`
2. Delete `/pages/audit-result-report/executive-summary-upload.vue`
3. Verify `/pages/executive-summary/` has the latest content
4. Update all navigation links in components

---

## How to Check Current Routes

### Method 1: Check File Structure
```bash
find pages -type f -name "*.vue" | sort
```

### Method 2: Check Generated Routes (at runtime)
In browser console:
```javascript
// If using Vue Router (though Nuxt has file-based routing)
console.log(useRouter().getRoutes())
```

### Method 3: Nuxt DevTools
- Install Nuxt DevTools: `npm install -D @nuxt/devtools`
- Open DevTools in browser to see all routes

---

## Best Practices

1. **Use kebab-case for files:** `dashboard-settings.vue` ✅ not `dashboardSettings.vue`
2. **Use index.vue for folder routes:** `folder/index.vue` → `/folder` ✅
3. **Keep related pages together:** Group by feature, not by page type
4. **Use route prefixes consistently:** `/admin/*`, `/settings/*`, `/master/*`
5. **Document route changes:** Add comments in nuxt.config.ts for redirects

---

## Common Patterns

### Nested Routes (Layouts)
```
/pages/
├── admin/
│   ├── index.vue                  → /admin
│   ├── users.vue                  → /admin/users
│   ├── settings/
│   │   └── index.vue              → /admin/settings
```

### Grouped Routes (with Parentheses)
Files in `(group)` folder don't appear in URL:
```
/pages/
├── (auth)/
│   ├── login.vue                  → /login
│   └── signup.vue                 → /signup
```

---

## Useful Commands

```bash
# Check page structure
find frontend/pages -type f -name "*.vue" | sort

# Navigate to check specific page
cat frontend/pages/audit-result-report/executive-summary.vue

# Check if routes are being used
grep -r "audit-result-report/executive-summary" frontend/
```

---

## Summary

For your **Executive Summary** case:

**Recommended:** Use Option A (Keep both + redirect)
- ✅ New URL works: `/executive-summary`
- ✅ Old URL redirects: `/audit-result-report/executive-summary`
- ✅ No broken links for existing users
- ✅ Better URL hierarchy (flattens navigation)

This approach is safer and provides a smooth transition.
