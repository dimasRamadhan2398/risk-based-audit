# Fieldwork API - Rate Limiting & Caching Quick Reference

## What Changed?

### Kong Gateway Changes
- ✅ **Response Caching:** 5-minute TTL for fieldwork GET endpoints
- ✅ **Rate Limiting:** Per-endpoint limits (60 req/min on prod, 300 on dev)

### Frontend Store Changes (`/frontend/stores/audit-fieldwork.ts`)
- ✅ **Request Deduplication:** Multiple calls to same endpoint within ~1s return same Promise
- ✅ **Smart Refetch:** Only refetch after CREATE operations (UPDATE/DELETE use cache invalidation)
- ✅ **Promise.allSettled():** Continue loading if 1-2 endpoints fail

---

## Impact on Your Code

### No Breaking Changes ✅
All function signatures remain the same:
```typescript
store.fetchAllFieldworkData(assignmentLetterId)  // Still works the same way
store.fetchInterviews(assignmentLetterId)        // No changes
store.saveInterview()                             // No changes
store.deleteInterview()                           // No changes
```

### What's Different (Internals)
- Fewer API calls being made to backend
- More responses coming from Kong cache
- Cache headers in responses: `X-Cache: HIT | MISS`

---

## Using the Store

### Basic Usage (No Changes)
```typescript
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'

const store = useAuditFieldworkStore()

// This still triggers 5 concurrent API calls
await store.fetchAllFieldworkData('Letter-2024-001')

// These still work the same
await store.saveInterview()
await store.deleteInterview(0)
```

### New Capability: Cache Invalidation
For advanced use cases where you need to force a refresh:

```typescript
// Invalidate cache for specific assignment letter
store.invalidateFieldworkDataCache('Letter-2024-001')

// Or invalidate all caches
store.invalidateFieldworkDataCache()

// Then fetch fresh data
await store.fetchAllFieldworkData('Letter-2024-001', true) // skipCache = true
```

---

## Rate Limit Response Headers

### What You'll See
When you make API requests, you'll get headers like:

```
X-RateLimit-Limit-Minute: 60
X-RateLimit-Remaining-Minute: 45
X-Ratelimit-Reset: 1728057600
```

### If Rate Limited (429 Too Many Requests)
```
HTTP/1.1 429 Too Many Requests
Retry-After: 45
X-RateLimit-Remaining-Minute: 0
```

**Action:** Wait 45 seconds before retrying. The frontend handles this automatically.

---

## Cache Behavior

### Cache Hit (Response from Kong, not backend)
```
HTTP/1.1 200 OK
X-Cache: HIT
Age: 23
Cache-Control: public, max-age=300
Content-Type: application/json
```

**Note:** Response is from Kong cache, not from PostgreSQL. Very fast!

### Cache Miss (Fresh from backend)
```
HTTP/1.1 200 OK
X-Cache: MISS
Cache-Control: public, max-age=300
Content-Type: application/json
```

**Note:** Backend query executed. Normal speed (50-200ms).

### No Cache (POST/PUT/DELETE)
```
HTTP/1.1 201 Created
X-Cache: BYPASS
Cache-Control: no-cache
Content-Type: application/json
```

**Note:** These operations bypass caching (correct behavior).

---

## Performance Expectations

### Page Load (First Time)
```
Time: ~500-800ms
Requests: 5 concurrent (interviews, observations, documents, samples, test-controls)
Cache: All 5 miss (first time)
```

### Page Load (Second Time, Same Letter)
```
Time: ~50-100ms
Requests: 0 (browser cache + Kong cache)
Cache: All 5 hit
```

### Page Load (Different Letter)
```
Time: ~500-800ms
Requests: 5 concurrent
Cache: All 5 miss (new letter ID)
```

### Save Interview
```
Time: ~200-400ms
Requests: 2
  1. POST /fieldwork/interviews (new record)
     → Expect 1 API to fetch new list
  2. GET /fieldwork/interviews (refetch for new ID)
     → Cache miss (fresh data)
Cache: 1 miss, 1 hit
```

### Update Interview
```
Time: ~100-200ms
Requests: 1
  1. PUT /fieldwork/interviews/{id}
     → No refetch (cache invalidated locally)
Cache: 1 miss
```

---

## Debugging

### Check Cache Status
Open DevTools → Network tab:
```
GET /api/v1/fieldwork/interviews?assignmentLetterId=...
Response Headers:
  X-Cache: HIT              ← Cache hit from Kong
  Age: 45                   ← Seconds since cached
  Cache-Control: max-age=300
```

### Check Rate Limit Status
```javascript
// In browser console
fetch('/api/v1/fieldwork/interviews')
  .then(r => {
    console.log('Rate Limit Remaining:', r.headers.get('X-RateLimit-Remaining-Minute'))
    console.log('Cache:', r.headers.get('X-Cache'))
  })
```

### Monitor API Calls
```javascript
// In browser console, monitor network activity
const observer = new PerformanceObserver((list) => {
  list.getEntries().forEach((entry) => {
    if (entry.name.includes('/fieldwork/')) {
      console.log(entry.name, entry.duration + 'ms')
    }
  })
})

observer.observe({ entryTypes: ['resource'] })
```

---

## Troubleshooting

### Problem: 429 Too Many Requests
**Cause:** Hit rate limit (60 req/min)  
**Solution:** 
- Wait for `Retry-After` seconds
- Check Network tab to count requests
- Verify no infinite loops or rapid clicks

### Problem: Stale Data After Update
**Cause:** Cache not invalidated properly  
**Solution:**
- Manually call `store.invalidateFieldworkDataCache(letterId)`
- Refresh page (clears all caches)
- Wait 5 minutes for cache TTL to expire

### Problem: 304 Not Modified
**Cause:** ETag/If-None-Match headers  
**Solution:**
- This is normal and good (saves bandwidth)
- Browser handles automatically
- No action needed

### Problem: 503 Service Unavailable
**Cause:** Backend might be restarting after deploy  
**Solution:**
- Wait 30-60 seconds
- Check backend service health: `docker ps | grep audit-service`
- Check logs: `docker logs rb_audit_audit_service`

---

## Testing Locally

### Load Test Script
```bash
# Test rate limiting (should hit limit around 60 requests)
for i in {1..100}; do
  curl -s -i http://localhost:8000/api/v1/fieldwork/interviews \
    -H "Authorization: Bearer your_token" \
    | grep -E "X-RateLimit|X-Cache" &
done
wait
```

### Cache Test
```javascript
// In browser console
async function testCache() {
  const letterid = 'Letter-2024-001'
  
  // First call - should miss
  const t1 = performance.now()
  const r1 = await fetch(`/api/v1/fieldwork/interviews?assignmentLetterId=${letterid}`)
  const t1_end = performance.now()
  console.log(`Call 1 (${t1_end - t1}ms):`, r1.headers.get('X-Cache'))
  
  // Immediate second call - should hit
  const t2 = performance.now()
  const r2 = await fetch(`/api/v1/fieldwork/interviews?assignmentLetterId=${letterid}`)
  const t2_end = performance.now()
  console.log(`Call 2 (${t2_end - t2}ms):`, r2.headers.get('X-Cache'))
}

testCache()
```

---

## Common Questions

**Q: Can I disable caching?**  
A: For testing, yes: `store.fetchAllFieldworkData(id, skipCache = true)`

**Q: Will my updates be seen immediately?**  
A: Yes! CREATE/UPDATE/DELETE bypass cache. GET requests see new data after 5 minutes or manual refresh.

**Q: What if another user updates the same record?**  
A: You'll see stale data for up to 5 minutes. Manual refresh gets fresh data immediately.

**Q: Can I clear the cache before 5 minutes?**  
A: Call `store.invalidateFieldworkDataCache()` to clear all pending requests.

**Q: Does this work offline?**  
A: No. Caching only works for HTTP requests. No service worker is implemented.

**Q: Is there risk of data loss?**  
A: No. All modifications go through API. Caching is read-only.

---

## Support

For issues or questions:
1. Check Prometheus metrics: `http://localhost:9090/graph`
2. Check Kong metrics: `http://localhost:8001/metrics`
3. Review logs: `docker logs rb_audit_kong` or `docker logs rb_audit_audit_service`
4. Ask in #auditsphere-dev Slack channel

---

## See Also
- Full documentation: `docs/AUDIT-FIELDWORK-PERFORMANCE-OPTIMIZATION.md`
- Kong Configuration: `backend/kong-gateway/kong/prod/kong.yml`
- Frontend Store: `frontend/stores/audit-fieldwork.ts`
- Page Component: `frontend/pages/audit-fieldwork/index.vue`
