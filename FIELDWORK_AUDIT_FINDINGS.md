# Audit Fieldwork Endpoints - Issues Analysis & Recommendations

## Issue 1: Missing Authorization Header on Test-Controls Endpoint

### Finding
**Status: NOT A BACKEND ISSUE** - The endpoint is correctly protected by auth middleware

#### Backend Verification
- ✅ Endpoint is registered under `/api/v1` group (routes/handler.go:243-250)
- ✅ `/api/v1` group has auth middleware applied (routes/handler.go:74)
- ✅ Auth middleware properly checks for Authorization header (pkg/middleware/auth.go:31-57)
- ✅ All fieldwork endpoints use the same CRUD handlers with identical protection
- ✅ No special skip or bypass for test-controls endpoint

#### Root Cause: Frontend Token Transmission Issue
The error "Missing authorization header" indicates the **frontend is not sending the token**, not that the backend is missing auth checks.

**Frontend flow:**
1. `frontend/plugins/api-auth.client.ts` - Global $fetch interceptor that adds Authorization header
2. Intercepts when: `authStore.token` is set AND URL matches an API base
3. **Problem**: The authStore.token may not be set when the first fieldwork request is made

**Suspect Areas:**
1. **Plugin Initialization Order** (plugins/api-auth.client.ts vs plugins/auth.client.ts)
   - `api-auth.client.ts` is synchronous and creates the $fetch interceptor immediately
   - `auth.client.ts` is async and calls `await authStore.fetchUser()` to restore token from cookies
   - If api-auth initializes before the token is restored, early requests won't have headers

2. **Token State Race Condition**
   - When `selectedAssignmentLetter` changes, watcher calls `fetchAllFieldworkData()` (line 321-325)
   - If this happens before authStore is fully hydrated, token won't be available
   - The 5 parallel $fetch calls in Promise.all() (lines 173-179) will all fail auth

3. **Missing Error Handling**
   - The store silently catches auth errors (line 195) and shows "Failed to load fieldwork data"
   - No specific handling for 401 Unauthorized responses

### Recommendations - FRONTEND FIXES

#### Quick Fix: Ensure Token Exists Before Fetching
```typescript
// In stores/audit-fieldwork.ts - fetchAllFieldworkData function
const fetchAllFieldworkData = async (assignmentLetterId: string) => {
  if (!assignmentLetterId) return
  
  // WAIT for auth to initialize before making requests
  const authStore = useAuthStore()
  if (!authStore._initialized) {
    console.warn('[Fieldwork] Waiting for auth initialization...')
    // Either wait or debounce the call
    return
  }
  
  if (!authStore.token) {
    console.error('[Fieldwork] No auth token available')
    errorMsg.value = 'Authentication required. Please refresh the page.'
    return
  }
  
  // ... rest of function
}
```

#### Recommended Fix: Add Request Interceptor at Store Level
Create a helper that ensures auth before requests:
```typescript
// In composables/useAuthenticatedApi.ts (new file)
export const useAuthenticatedApi = () => {
  const authStore = useAuthStore()
  
  const waitForAuth = async () => {
    // Wait for auth to initialize (max 5 seconds)
    const maxRetries = 50
    for (let i = 0; i < maxRetries; i++) {
      if (authStore._initialized && authStore.token) return true
      await new Promise(resolve => setTimeout(resolve, 100))
    }
    return false
  }
  
  const fetch = async (url: string, options?: any) => {
    if (!await waitForAuth()) {
      throw new Error('Authentication timeout')
    }
    return $fetch(url, options)
  }
  
  return { fetch, waitForAuth }
}
```

Then use in stores:
```typescript
const { fetch: authenticatedFetch } = useAuthenticatedApi()
const res = await authenticatedFetch(`${baseUrl}/fieldwork/test-controls?assignmentLetterId=${assignmentLetterId}`)
```

#### Ensure Plugin Load Order
Check `nuxt.config.ts` to ensure auth plugin loads before other plugins that make API calls:
```typescript
plugins: [
  '~/plugins/auth.client.ts',      // FIRST: restore token from cookies
  '~/plugins/api-auth.client.ts',   // SECOND: attach token to all requests
  '~/plugins/other-plugins.ts',
]
```

---

## Issue 2: Excessive API Calls (25 Tables, Called Every Second)

### Current Behavior Analysis
**What the code does:**
1. Watcher on `selectedAssignmentLetter` (line 321-325)
2. When assignment letter changes, calls `fetchAllFieldworkData()`
3. Makes 5 parallel requests to fetch: interviews, observations, documents, samples, test-controls
4. **This is correct and expected behavior** for initial load

**Problem identified:**
- 5 stages × 5 table types = 25 tables
- If this is calling every second, something is re-triggering the watcher excessively
- **Suspect**: Component re-renders causing `selectedAssignmentLetter` to change/re-evaluate

### Recommendations - BACKEND CACHING STRATEGY

#### 1. Add Response Caching Headers
```go
// In routes/handler.go, create a caching middleware

func cacheMiddleware(duration time.Duration) gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Header("Cache-Control", fmt.Sprintf("private, max-age=%d", int(duration.Seconds())))
    c.Header("ETag", generateETag(c)) // Generate weak ETag based on query params
    c.Next()
  }
}

// Apply to fieldwork routes that are safe to cache
fieldworkTestControls.GET("", 
  cacheMiddleware(5 * time.Minute), // Cache for 5 minutes
  crud.List(h.db, "FieldworkTestControl", ...),
)
```

#### 2. Implement Conditional Request Support (If-None-Match)
The auth middleware should recognize 304 Not Modified responses:
```go
// In middleware/auth.go - add after token validation
func (m *AuthMiddleware) CheckETag() gin.HandlerFunc {
  return func(c *gin.Context) {
    // If-None-Match handling happens in HTTP layer automatically
    // But log it for debugging
    if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" {
      c.Set("request_has_etag", true)
    }
    c.Next()
  }
}
```

#### 3. Add Redis Caching for Fieldwork Queries
The audit-service already has Redis client available. Use it:

```go
// In controllers/crud/crud_handler.go

func List(db *gorm.DB, modelName string, newSlice func() interface{}, 
          preloads ...string) gin.HandlerFunc {
  return func(c *gin.Context) {
    // Generate cache key from query params
    cacheKey := fmt.Sprintf("fieldwork:%s:%s", modelName, c.Request.URL.RawQuery)
    
    // Try to get from Redis (if available)
    if redisClient := c.MustGet("redis").(*redis.Client); redisClient != nil {
      if cached, err := redisClient.Get(c.Request.Context(), cacheKey).Result(); err == nil {
        // Return cached response
        var result interface{}
        if err := json.Unmarshal([]byte(cached), &result); err == nil {
          response.OK(c, "Cached "+modelName, result)
          return
        }
      }
    }
    
    // ... rest of List handler
    
    // Cache the result before responding
    if redisClient := c.MustGet("redis").(*redis.Client); redisClient != nil {
      responseData := gin.H{
        "items": items,
        "pagination": gin.H{
          "page": page,
          "page_size": pageSize,
          "total": total,
          "total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
        },
      }
      if data, err := json.Marshal(responseData); err == nil {
        redisClient.Set(c.Request.Context(), cacheKey, data, 5*time.Minute)
      }
    }
    
    response.OK(c, modelName+" fetched successfully", gin.H{...})
  }
}
```

#### 4. Query Parameter Optimization
The fieldwork endpoints accept `assignmentLetterId` as query param. Add indexing:

```go
// In models/audit-fieldwork.go
type FieldworkTestControl struct {
  ID                  uuid.UUID `gorm:"primaryKey;index"`
  AssignmentLetterID  string    `gorm:"index"` // Add index for filter queries
  ControlName         string
  // ... other fields
}

type FieldworkSample struct {
  ID                  uuid.UUID `gorm:"primaryKey;index"`
  AssignmentLetterID  string    `gorm:"index"`  // Add index
  DocumentName        string
  // ... other fields
}

// Similar for: FieldworkInterview, FieldworkObservation, FieldworkDocument
```

Then run migration:
```bash
cd backend/audit-service
go run cmd/seed/main.go  # Runs AutoMigrate with indexes
```

---

### Recommendations - FRONTEND OPTIMIZATION

#### 1. Debounce Assignment Letter Selection
```typescript
// In components that set selectedAssignmentLetter
import { debounce } from 'lodash-es'

const handleAssignmentLetterChange = debounce((letterId: string) => {
  selectedAssignmentLetter.value = letterId
}, 300) // Wait 300ms for user to finish selecting
```

#### 2. Add Request Deduplication
```typescript
// In useAuthenticatedApi or as composable
const pendingRequests = new Map<string, Promise<any>>()

const fetchWithDedup = async (url: string, options?: any) => {
  // Return existing pending request if one is already in flight for this URL
  if (pendingRequests.has(url)) {
    return pendingRequests.get(url)
  }
  
  const promise = authenticatedFetch(url, options)
  pendingRequests.set(url, promise)
  
  try {
    return await promise
  } finally {
    pendingRequests.delete(url)
  }
}
```

#### 3. Use Stale-While-Revalidate Pattern
```typescript
const fetchAllFieldworkData = async (assignmentLetterId: string) => {
  if (!assignmentLetterId) return
  
  // Serve stale data immediately while refreshing in background
  const cacheKey = `fieldwork_${assignmentLetterId}`
  const cached = sessionStorage.getItem(cacheKey)
  
  if (cached) {
    // Load cached data immediately
    const data = JSON.parse(cached)
    fieldworkData.value[assignmentLetterId] = data
  }
  
  // Refresh in background
  try {
    const [interviews, observations, ...] = await Promise.all([...])
    fieldworkData.value[assignmentLetterId] = { interviews, observations, ... }
    sessionStorage.setItem(cacheKey, JSON.stringify(fieldworkData.value[assignmentLetterId]))
  } catch (error) {
    if (!cached) throw error  // Only error if no cached fallback
  }
}
```

#### 4. Implement Request Cancellation
```typescript
// In stores/audit-fieldwork.ts
const abortControllers = new Map<string, AbortController>()

const fetchAllFieldworkData = async (assignmentLetterId: string) => {
  // Cancel previous request for this assignment letter
  if (abortControllers.has(assignmentLetterId)) {
    abortControllers.get(assignmentLetterId)?.abort()
  }
  
  const controller = new AbortController()
  abortControllers.set(assignmentLetterId, controller)
  
  try {
    const [interviewsRes, ...] = await Promise.all([
      $fetch(`...`, { signal: controller.signal }),
      ...
    ])
    // ...
  } catch (error) {
    if (error.name !== 'AbortError') {
      // Handle real errors, ignore abort
    }
  }
}
```

---

## Summary of Issues Found

| Issue | Root Cause | Severity | Fix Location |
|-------|-----------|----------|--------------|
| Missing Auth Header | Frontend token not sent | HIGH | Frontend: api-auth.client.ts initialization order |
| Excessive API Calls | Likely frontend component re-renders | MEDIUM | Frontend: debounce + deduplication |
| No Response Caching | Backend doesn't set Cache-Control | MEDIUM | Backend: add caching middleware |
| No Database Indexes | Slow query filters on large tables | LOW-MEDIUM | Backend: add indexes on assignment_letter_id |

---

## Verification Checklist

### Backend
- [ ] Confirm auth middleware is applied to `/api/v1` group
- [ ] Test with missing Authorization header → should get 401
- [ ] Test with valid token → should work
- [ ] Check Redis client is properly initialized (passed to routes)
- [ ] Add cache indexes to fieldwork models
- [ ] Run `go test ./...` for audit-service

### Frontend
- [ ] Check plugin load order in nuxt.config.ts
- [ ] Verify authStore is initialized before fieldwork requests
- [ ] Monitor network tab: confirm Auth header is being sent
- [ ] Check for excessive watcher triggers (React DevTools / Vue DevTools)
- [ ] Implement debouncing for assignment letter changes
- [ ] Test caching with browser DevTools Network tab

---

## Files Affected

### Backend (Audit Service)
- `/Users/a/Projects/risk-based-audit/backend/audit-service/routes/handler.go` (lines 243-250)
- `/Users/a/Projects/risk-based-audit/backend/audit-service/pkg/middleware/auth.go` (already correct)
- `/Users/a/Projects/risk-based-audit/backend/audit-service/controllers/crud/crud_handler.go` (for caching)
- `/Users/a/Projects/risk-based-audit/backend/audit-service/models/audit-fieldwork.go` (for indexes)

### Frontend (Nuxt)
- `/Users/a/Projects/risk-based-audit/frontend/plugins/api-auth.client.ts` (token timing issue)
- `/Users/a/Projects/risk-based-audit/frontend/plugins/auth.client.ts` (initialization order)
- `/Users/a/Projects/risk-based-audit/frontend/stores/audit-fieldwork.ts` (excessive calls)
- `/Users/a/Projects/risk-based-audit/frontend/composables/useApiUrl.ts` (baseline URLs)
- `/Users/a/Projects/risk-based-audit/frontend/nuxt.config.ts` (plugin order)
