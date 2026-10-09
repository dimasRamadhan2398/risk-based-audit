# Quick Fix: Missing Authorization Header on Fieldwork Endpoints

## Problem
Fieldwork endpoints return "Missing authorization header" error when user selects an assignment letter.

**Root Cause**: Frontend's `authStore.token` is not yet loaded when the first fieldwork request is made due to async initialization timing.

## Solution: Add Auth Initialization Check to Fieldwork Store

### File: `/Users/a/Projects/risk-based-audit/frontend/stores/audit-fieldwork.ts`

**Step 1**: Add import for authStore (if not already imported)
```typescript
import { useAuthStore } from './auth'
```

**Step 2**: Modify the `fetchAllFieldworkData` function to wait for auth initialization

Replace the existing function with this version that waits for auth to be ready:

```typescript
const fetchAllFieldworkData = async (assignmentLetterId: string, skipCache = false) => {
  if (!assignmentLetterId) return

  // CRITICAL FIX: Wait for auth to initialize before making API requests
  // This prevents "Missing authorization header" errors
  const authStore = useAuthStore()
  
  // Wait up to 5 seconds for auth initialization
  let retries = 50
  while (!authStore._initialized && retries > 0) {
    await new Promise(resolve => setTimeout(resolve, 100))
    retries--
  }
  
  if (!authStore._initialized) {
    console.error('[Fieldwork] Auth initialization timeout - cannot fetch fieldwork data')
    errorMsg.value = 'Authentication initialization failed. Please refresh the page.'
    loading.value = false
    return
  }

  // Return existing pending request if one is already in flight
  if (pendingRequests.value[assignmentLetterId] && !skipCache) {
    return pendingRequests.value[assignmentLetterId]
  }

  // If data already exists and we're not forcing a refresh, skip
  if (fieldworkData.value[assignmentLetterId] && !skipCache) {
    return Promise.resolve()
  }

  loading.value = true
  errorMsg.value = ''
  const toast = useToastNotification()

  const request = (async () => {
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const results = await Promise.allSettled([
        $fetch(`${baseUrl}/fieldwork/interviews?assignmentLetterId=${assignmentLetterId}`),
        $fetch(`${baseUrl}/fieldwork/observations?assignmentLetterId=${assignmentLetterId}`),
        $fetch(`${baseUrl}/fieldwork/documents?assignmentLetterId=${assignmentLetterId}`),
        $fetch(`${baseUrl}/fieldwork/samples?assignmentLetterId=${assignmentLetterId}`),
        $fetch(`${baseUrl}/fieldwork/test-controls?assignmentLetterId=${assignmentLetterId}`)
      ])

      const [interviewsRes, observationsRes, documentsRes, samplesRes, testControlsRes] = results.map(r =>
        r.status === 'fulfilled' ? r.value : null
      )

      const interviewsList = interviewsRes?.data?.items || interviewsRes?.items || (Array.isArray(interviewsRes) ? interviewsRes : [])
      const observationsList = observationsRes?.data?.items || observationsRes?.items || (Array.isArray(observationsRes) ? observationsRes : [])
      const documentsList = documentsRes?.data?.items || documentsRes?.items || (Array.isArray(documentsRes) ? documentsRes : [])
      const samplesList = samplesRes?.data?.items || samplesRes?.items || (Array.isArray(samplesRes) ? samplesRes : [])
      const testControlsList = testControlsRes?.data?.items || testControlsRes?.items || (Array.isArray(testControlsRes) ? testControlsRes : [])

      fieldworkData.value[assignmentLetterId] = {
        interviews: Array.isArray(interviewsList) ? interviewsList : [],
        observations: Array.isArray(observationsList) ? observationsList : [],
        documents: Array.isArray(documentsList) ? documentsList : [],
        samples: Array.isArray(samplesList) ? samplesList : [],
        testControls: Array.isArray(testControlsList) ? testControlsList : []
      }
    } catch (error: any) {
      console.error('Failed to fetch fieldwork data:', error)
      const detail = extractErrorMessage(error, 'Failed to load fieldwork data.')
      errorMsg.value = detail
      toast.showError('Failed to load fieldwork data.', detail)
      fieldworkData.value[assignmentLetterId] = emptyFieldworkData()
    } finally {
      delete pendingRequests.value[assignmentLetterId]
      loading.value = false
    }
  })()

  pendingRequests.value[assignmentLetterId] = request
  return request
}
```

## Verification

After applying the fix:

1. **Open browser console** (F12)
2. **Navigate to Fieldwork page**
3. **Select an assignment letter**
4. **Check console for debug logs**:
   - Should see `[AuthStore] Session successfully restored...` 
   - Should see fieldwork data loading without 401 errors
5. **Check Network tab**:
   - `/fieldwork/test-controls` request should have `Authorization: Bearer <token>` header
   - Should get 200 response, not 401

## Alternative: Quick Workaround (If Fix Not Sufficient)

If the above fix doesn't work, it might indicate the auth plugin order issue. Add this workaround to nuxt.config.ts:

```typescript
// nuxt.config.ts
export default defineNuxtConfig({
  // ... other config
  plugins: [
    '~/plugins/auth.client.ts',      // Load first - restores token from cookie
    '~/plugins/api-auth.client.ts',   // Load second - sets up interceptor with token
    // ... other plugins
  ]
})
```

## What the Fix Does

1. **Waits for auth initialization**: Before making ANY fieldwork API calls, the store now waits for `authStore._initialized` to be true
2. **Timeout protection**: If auth takes too long (>5 seconds), shows error instead of hanging
3. **Leverages global interceptor**: Once auth is ready, the global `api-auth.client.ts` plugin will automatically add the Authorization header to all requests
4. **Maintains deduplication**: Still prevents duplicate requests and uses cached data when available

## Performance Impact

- **Negligible**: Auth initialization typically completes in <100ms, this just adds safety
- **User Experience**: If auth somehow fails, user gets a clear error message instead of silent failure
- **Network**: No additional API calls, just timing coordination

---

## Long-term Solution

Once this quick fix is confirmed working, implement the full solution from `FIELDWORK_AUDIT_FINDINGS.md`:
1. Add request deduplication at composable level
2. Implement debouncing for assignment letter selection
3. Add backend caching with Redis
4. Add database indexes on `assignment_letter_id` columns
