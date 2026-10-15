# Quick Cache Reference

One-page summary of the HTTP caching implementation for AuditSphere dashboard.

## What Was Done

✓ HTTP cache headers added to 38 dashboard endpoints across 4 services
✓ Cache TTLs optimized based on data freshness requirements
✓ ETag support for conditional requests and bandwidth savings
✓ All services building successfully, ready for deployment

## How to Verify It's Working

```bash
# Test endpoint with cache
curl -i http://localhost:8002/api/v1/audit-result-reports/recent-findings

# Look for these headers:
# Cache-Control: public, max-age=60
# ETag: "hash_value"
# Last-Modified: date
# Expires: future_date
```

## Most Important Endpoints

| Endpoint | TTL | Use Case |
|----------|-----|----------|
| `/audit-result-reports/recent-findings` | 1 min | Critical findings |
| `/analytics/report` | 5 min | Dashboard summary |
| `/rcm/summary` | 5 min | Control effectiveness |
| `/risks` | 5 min | Risk metrics |
| `/employees` | 1 hr | Master data |

## Expected Performance Gains

- **API Calls**: 40-50% reduction
- **Database Queries**: 30-40% reduction  
- **Response Time**: 2-3x faster
- **Server CPU**: 20-25% reduction

## Files to Know

```
DASHBOARD_CACHING_STRATEGY.md      ← Read this for details
CACHE_MIDDLEWARE_GUIDE.md          ← Developer how-to
DATABASE_INDEXES_FOR_CACHING.sql   ← Phase 2 optimization
```

## To Change Cache TTL

Example: Increase recent findings from 1 min to 2 min

1. Edit: `audit-service/pkg/middleware/cache.go`
2. Find: `DefaultCacheConfigs["GET /api/v1/audit-result-reports/recent-findings"]`
3. Change: `MaxAge: 60` → `MaxAge: 120`
4. Rebuild: `cd audit-service && go build ./...`
5. Redeploy

## To Add New Cached Endpoint

1. Add route normally: `api.GET("/api/v1/new-endpoint", handler)`
2. Edit middleware cache.go for that service
3. Add to DefaultCacheConfigs:
```go
DefaultCacheConfigs["GET /api/v1/new-endpoint"] = CacheConfig{
    MaxAge:   300,  // seconds
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```
4. Rebuild and redeploy

## Common TTL Patterns

```
60s   → Recently updated data (findings, activities)
300s  → Dashboard metrics (stats, summary, trends)
600s  → Plan data (audit plans, assignments)
1800s → Stable data (charter, mandates)
3600s → Master data (employees, departments)
```

## Browser Cache Behavior

1. **First request**: Server returns data + Cache-Control header
2. **Browser caches**: Response stored for `max-age` seconds
3. **Next request within TTL**: Browser serves from cache (instant)
4. **After TTL expires**: Browser requests fresh data

## If Cache Causes Issues

Option 1: Add version parameter
```javascript
const url = `/api/v1/audit-result-reports/recent-findings?v=${Date.now()}`;
```

Option 2: Use Cache-Control header
```javascript
fetch(url, { cache: 'no-store' })  // Bypass cache
```

Option 3: Reduce TTL for that endpoint (in middleware)

## Testing in Browser

1. Open Chrome DevTools (F12)
2. Go to Network tab
3. Refresh page
4. Look for `Cache-Control` in response headers
5. Make same request again
6. Status should show `(from cache)` or `304 Not Modified`

## Architecture

```
Browser/Client
    ↓ ↑
    Cache Middleware ← adds headers
    ↓
    Handler (business logic)
    ↓
    Database
```

## When to Use Phase 2 (Database Indexes)

If you see:
- Queries still slow despite caching
- Database CPU high during peak hours
- Need to reduce TTLs but still get slow responses

Then: Run `DATABASE_INDEXES_FOR_CACHING.sql`

## When to Use Phase 3 (Redis)

If you see:
- Cache hit ratio < 40%
- Multiple server instances fighting for cache
- Need sub-millisecond response times
- Distributed cache invalidation needed

Then: Plan Redis implementation

## Key Concepts

**Cache Hit**: Client gets response from browser cache (instant)  
**Cache Miss**: Client makes request to server (normal latency)  
**ETag**: Hash of response body, enables 304 Not Modified responses  
**TTL**: Time-to-live, how long browser keeps response  
**Cache Control**: HTTP header telling browser how to cache  
**Vary**: Indicates what affects cache (e.g., Accept-Encoding)  

## Support

- **Questions about implementation**: See `CACHE_MIDDLEWARE_GUIDE.md`
- **Performance metrics**: See `DASHBOARD_CACHING_STRATEGY.md`
- **Troubleshooting**: See `CACHE_MIDDLEWARE_GUIDE.md` troubleshooting section
- **Database optimization**: See `DATABASE_INDEXES_FOR_CACHING.sql`

---

**Status**: Ready for deployment  
**All Services**: Building successfully  
**Documentation**: Complete
