# Step-by-Step: Consolidate Executive Summary Route

## Current Situation
You have **two versions** of the Executive Summary page:

1. **Old (nested):** `/audit-result-report/executive-summary`
   - File: `/pages/audit-result-report/executive-summary.vue`
   - Hard-coded Indonesian text

2. **New (top-level):** `/executive-summary`
   - File: `/pages/executive-summary/index.vue`
   - Uses i18n translations with `t()` function

---

## Recommended Approach: Keep New + Redirect Old

This ensures:
- ✅ New cleaner URL works: `/executive-summary`
- ✅ Old links don't break: `/audit-result-report/executive-summary` → redirects
- ✅ No data loss
- ✅ Smooth user experience

---

## Step 1: Verify New Page is Complete

Check that `/pages/executive-summary/index.vue` has all functionality:

```bash
grep -E "(openNewForm|store\.|@click)" frontend/pages/executive-summary/index.vue | head -20
```

---

## Step 2: Add Route Redirects

Edit `/frontend/nuxt.config.ts` and update the `routeRules` section:

```typescript
export default defineNuxtConfig({
  // ... existing config
  
  routeRules: {
    "/**": {
      ssr: false,
      headers: {
        "X-Robots-Tag": "noindex, nofollow, noarchive, nosnippet"
      }
    },
    
    // ADD THESE REDIRECT RULES:
    "/audit-result-report/executive-summary": { redirect: "/executive-summary" },
    "/audit-result-report/executive-summary-upload": { redirect: "/executive-summary/upload" },
    
    // Existing API proxy rules...
    "/api/v1/**": {
      proxy: process.env.API_BASE_URL_SERVER || defaultProxyTarget
    },
    "/api/analytics/**": {
      proxy: process.env.ANALYTICS_API_BASE_URL_SERVER || defaultAnalyticsProxyTarget
    },
    "/uploads/**": {
      proxy: process.env.UPLOADS_BASE_URL_SERVER || defaultUploadsProxyTarget
    }
  },
  
  // ... rest of config
})
```

---

## Step 3: Update Navigation Links (if any)

Search for hardcoded links to the old route:

```bash
# Find all references to old route
grep -r "audit-result-report/executive-summary" frontend/components frontend/pages --include="*.vue"
```

Update any found links:
```vue
<!-- OLD -->
<UButton to="/audit-result-report/executive-summary" />

<!-- NEW -->
<UButton to="/executive-summary" />
```

---

## Step 4: Update Audit Result Report Page

Check `/pages/audit-result-report/index.vue` to see if it links to executive summary:

```bash
cat frontend/pages/audit-result-report/index.vue | grep -i "executive-summary"
```

If it does, update the links.

---

## Step 5: Optional - Clean Up Old Files

**Only do this if you're 100% sure the new page has all functionality:**

```bash
# Delete old files
rm frontend/pages/audit-result-report/executive-summary.vue
rm frontend/pages/audit-result-report/executive-summary-upload.vue

# Verify deletion
ls -la frontend/pages/audit-result-report/ | grep executive
```

**Note:** Keep the redirect rules in `nuxt.config.ts` even after deleting files, so old bookmarks/links still work.

---

## Step 6: Test the Changes

1. **Test new route works:**
   ```
   http://localhost:3000/executive-summary ✅
   ```

2. **Test redirect works:**
   ```
   http://localhost:3000/audit-result-report/executive-summary → redirects to /executive-summary ✅
   ```

3. **Check upload page:**
   ```
   http://localhost:3000/executive-summary/upload ✅
   ```

---

## File Summary

### Keep (New Structure)
```
/pages/executive-summary/
├── index.vue          ← Main page
└── upload.vue         ← Upload page
```

### Remove (Optional)
```
/pages/audit-result-report/
├── executive-summary.vue          ← DELETE
└── executive-summary-upload.vue   ← DELETE
```

### Keep (Updated)
```
/pages/audit-result-report/
├── index.vue          ← Keep
├── satisfaction-survey.vue
└── other files...
```

---

## Complete nuxt.config.ts Update

Here's the exact change needed:

```typescript
// BEFORE (line ~41-56)
routeRules: {
  "/**": {
    ssr: false,
    headers: {
      "X-Robots-Tag": "noindex, nofollow, noarchive, nosnippet"
    }
  },
  "/api/v1/**": {
    proxy: process.env.API_BASE_URL_SERVER || defaultProxyTarget
  },
  "/api/analytics/**": {
    proxy: process.env.ANALYTICS_API_BASE_URL_SERVER || defaultAnalyticsProxyTarget
  },
  "/uploads/**": {
    proxy: process.env.UPLOADS_BASE_URL_SERVER || defaultUploadsProxyTarget
  }
},

// AFTER (add 2 new redirect rules)
routeRules: {
  // Redirect old nested routes to new flat structure
  "/audit-result-report/executive-summary": { redirect: "/executive-summary" },
  "/audit-result-report/executive-summary-upload": { redirect: "/executive-summary/upload" },
  
  "/**": {
    ssr: false,
    headers: {
      "X-Robots-Tag": "noindex, nofollow, noarchive, nosnippet"
    }
  },
  "/api/v1/**": {
    proxy: process.env.API_BASE_URL_SERVER || defaultProxyTarget
  },
  "/api/analytics/**": {
    proxy: process.env.ANALYTICS_API_BASE_URL_SERVER || defaultAnalyticsProxyTarget
  },
  "/uploads/**": {
    proxy: process.env.UPLOADS_BASE_URL_SERVER || defaultUploadsProxyTarget
  }
},
```

---

## Verification Commands

```bash
# Check file structure
tree frontend/pages -L 2 --dirsfirst

# Find all pages
find frontend/pages -type f -name "*.vue" | sort

# Find all "executive-summary" references
grep -r "executive-summary" frontend/components frontend/pages --include="*.vue"

# Test after making changes
npm run dev
# Then visit: http://localhost:3000/executive-summary
```

---

## Before & After

### BEFORE (Current State)
```
URL: http://localhost:3000/audit-result-report/executive-summary
File: /pages/audit-result-report/executive-summary.vue
Status: Nested, hard-coded text
```

### AFTER (After Implementation)
```
URL: http://localhost:3000/executive-summary ✅
File: /pages/executive-summary/index.vue
Status: Flat structure, i18n ready

OLD URL: http://localhost:3000/audit-result-report/executive-summary
Behavior: Automatically redirects to /executive-summary ✅
```

---

## FAQ

**Q: Will old links break?**
A: No, they'll automatically redirect to the new URL.

**Q: Do I need to delete old files?**
A: No, but it's recommended to keep things clean. The redirects work regardless.

**Q: Can I do this without redirects?**
A: Yes, but then old links will 404. Not recommended.

**Q: What about bookmarks?**
A: Users with bookmarks to old URL will be redirected automatically.

**Q: Does this affect SEO?**
A: No, 301 redirects are SEO-friendly. The new URL will inherit old URL's SEO value.

---

## Next Steps

1. ✅ Read this guide
2. ⬜ Edit `nuxt.config.ts` and add redirect rules
3. ⬜ Test the changes in browser
4. ⬜ Search and update any hardcoded links
5. ⬜ Delete old files (optional, recommended)
6. ⬜ Commit and deploy

**Ready to implement? Follow the steps above!**
