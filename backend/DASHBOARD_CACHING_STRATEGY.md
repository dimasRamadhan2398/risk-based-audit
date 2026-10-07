# Dashboard Caching Strategy for AuditSphere

## Executive Summary

This document outlines a comprehensive caching strategy to optimize AuditSphere dashboard performance by reducing API calls, database load, and page load times. The implementation uses HTTP cache headers (Phase 1), with optional extensions for database optimization (Phase 2) and Redis-based caching (Phase 3).

**Expected Performance Improvements:**
- 40-50% reduction in API calls for frequent dashboard users
- 30-40% reduction in database query count during peak hours
- 2-3x faster dashboard load time on slow connections
- 20-25% reduction in server CPU usage

---

## Phase 1: HTTP Cache Headers (Implemented ✓)

HTTP cache headers are implemented via middleware in each service that automatically adds cache directives to GET requests.

### Implementation Details

#### Middleware Deployment

Cache middleware has been added to all dashboard-related services:

1. **Analytics Service** (`analytics-service/middleware/cache.go`)
   - Location: `/Users/a/Projects/risk-based-audit/backend/analytics-service/middleware/cache.go`
   - Enabled in: `cmd/main.go`

2. **Audit Service** (`audit-service/pkg/middleware/cache.go`)
   - Location: `/Users/a/Projects/risk-based-audit/backend/audit-service/pkg/middleware/cache.go`
   - Enabled in: `cmd/serve.go`

3. **Risk Service** (`risk-service/middleware/cache.go`)
   - Location: `/Users/a/Projects/risk-based-audit/backend/risk-service/middleware/cache.go`
   - Enabled in: `cmd/serve.go`

4. **Master Service** (`master-service/pkg/middleware/cache.go`)
   - Location: `/Users/a/Projects/risk-based-audit/backend/master-service/pkg/middleware/cache.go`
   - Enabled in: `cmd/serve.go`

#### Cache Headers Format

```
Cache-Control: {visibility}, max-age={ttl}
ETag: "{hash}"
Vary: Accept-Encoding
Last-Modified: {RFC1123-Date}
Expires: {RFC1123-Date}
```

**Example Response:**
```
Cache-Control: public, max-age=300
ETag: "a1b2c3d4e5f6g7h8"
Vary: Accept-Encoding
Last-Modified: Mon, 04 Oct 2026 14:00:00 GMT
Expires: Mon, 04 Oct 2026 14:05:00 GMT
```

#### Cache TTL Configuration

Cache configurations are defined in each middleware file as a map. TTLs are assigned based on:

1. **Update Frequency** - Data that changes frequently has shorter TTLs
2. **Critical Nature** - Critical data (findings, risks) has shorter TTLs
3. **Stability** - Master data (employees, departments) has longer TTLs

### Dashboard Endpoints & Cache Settings

#### Analytics Service Endpoints

```
GET /api/analytics/report                           → 5 min (300s)
GET /api/analytics/predict                          → 10 min (600s)
GET /api/analytics/risk-score                       → 5 min (300s)
POST /api/analytics/risk-score                      → 5 min (300s)
GET /api/analytics/risk-score/batch                 → 5 min (300s)
GET /api/analytics/anomaly                          → 10 min (600s)
POST /api/analytics/anomaly                         → 10 min (600s)
GET /api/analytics/anomaly/batch                    → 10 min (600s)
GET /api/analytics/text-analysis                    → 5 min (300s)
POST /api/analytics/text-analysis                   → 5 min (300s)
GET /api/analytics/text-analysis/batch              → 5 min (300s)
GET /api/analytics/performance-trend                → 5 min (300s)
POST /api/analytics/performance-trend               → 5 min (300s)
GET /api/analytics/performance-trend/batch          → 5 min (300s)
GET /api/analytics/caatt/full-population            → 30 min (1800s)
GET /api/analytics/caatt/duplicate-gap              → 30 min (1800s)
GET /api/analytics/caatt/benford                    → 30 min (1800s)
GET /api/analytics/caatt/stratification             → 30 min (1800s)
GET /api/analytics/caatt/reconciliation             → 30 min (1800s)
GET /api/analytics/caatt/policy-violations          → 30 min (1800s)
GET /api/analytics/caatt/data-quality               → 30 min (1800s)
```

#### Audit Service Endpoints

```
GET /api/v1/audit-result-reports/recent-findings    → 1 min (60s)
GET /api/v1/audit-result-reports/auto-findings      → 5 min (300s)
GET /api/v1/annual-audit-plans                      → 10 min (600s)
GET /api/v1/audit-assignments                       → 5 min (300s)
GET /api/v1/audit-activities                        → 5 min (300s)
GET /api/v1/strategic-plans                         → 10 min (600s)
GET /api/v1/audit-charters                          → 30 min (1800s)
GET /api/v1/audit-mandates                          → 30 min (1800s)
```

#### Risk Service Endpoints

```
GET /api/v1/risks                                   → 5 min (300s)
GET /api/v1/mitigations                             → 5 min (300s)
GET /api/v1/risk-factors/standard                   → 30 min (1800s)
GET /api/v1/risk-factors/corporate                  → 10 min (600s)
GET /api/v1/audit-universe/standard                 → 30 min (1800s)
GET /api/v1/audit-universe/corporate                → 10 min (600s)
GET /api/v1/rcm                                     → 10 min (600s)
GET /api/v1/rcm/summary                             → 5 min (300s)
```

#### Master Service Endpoints

```
GET /api/v1/companies                               → 1 hour (3600s)
GET /api/v1/business-units                          → 1 hour (3600s)
GET /api/v1/departments                             → 1 hour (3600s)
GET /api/v1/employees                               → 1 hour (3600s)
GET /api/v1/job-roles                               → 1 hour (3600s)
GET /api/v1/locations                               → 1 hour (3600s)
GET /api/v1/quality-assurance                       → 5 min (300s)
```

### Middleware Features

#### ETag Generation

ETags are generated using MD5 hash of the response body, enabling:
- Browser cache validation
- Conditional requests (If-None-Match)
- Reduced bandwidth consumption

#### Conditional Cache Behavior

- **GET Requests**: Full cache headers applied
- **HEAD Requests**: Cache headers applied (same as GET)
- **POST Requests**: Only analytics batch endpoints cached
- **PUT/DELETE Requests**: No caching (cache invalidation happens naturally)

#### CORS Compatibility

Cache headers are exposed via CORS headers:
- `Access-Control-Expose-Headers` includes cache headers
- Allows browsers to read and respect cache directives
- Fully compatible with `cache-control` client directives

---

## Phase 2: Database Query Optimization (Recommended)

While HTTP caching reduces API calls, database optimization improves query performance.

### Recommended Database Indexes

#### Audit Service Indexes

```sql
-- Speed up recent findings queries (ordered by creation date)
CREATE INDEX idx_audit_result_reports_created_at_desc 
ON audit_result_reports(created_at DESC);

-- Speed up assignment filtering
CREATE INDEX idx_audit_assignments_audit_plan_id_status 
ON audit_assignments(audit_plan_id, status);

-- Speed up activity tracking
CREATE INDEX idx_audit_activities_created_at_status 
ON audit_activities(created_at DESC, status);

-- Speed up working paper sample queries
CREATE INDEX idx_working_paper_samples_assignment_id 
ON working_paper_samples(assignment_letter_id);
```

#### Risk Service Indexes

```sql
-- Speed up risk filtering by status and risk level
CREATE INDEX idx_risks_status_risk_level 
ON risks(status, risk_level);

-- Speed up RCM lookups
CREATE INDEX idx_rcm_control_id_risk_id 
ON rcm_matrix(control_id, risk_id);

-- Speed up mitigation filtering
CREATE INDEX idx_mitigations_risk_id 
ON mitigations(risk_id);
```

#### Master Service Indexes

```sql
-- Speed up employee filtering
CREATE INDEX idx_employees_status_department_id 
ON employees(status, department_id);

-- Speed up department filtering
CREATE INDEX idx_departments_company_id_status 
ON departments(company_id, status);

-- Speed up organizational queries
CREATE INDEX idx_business_units_company_id 
ON business_units(company_id);
```

### Implementation Steps

1. **Analyze Current Queries**: Review slow query logs to identify bottlenecks
2. **Create Indexes**: Run the migration scripts above
3. **Monitor Impact**: Track query execution time before/after
4. **Tune TTLs**: Adjust cache TTLs based on query performance

---

## Phase 3: Redis-Based Caching (Optional Enhancement)

For high-traffic deployments or complex aggregations, Redis caching provides:
- Cache sharing across server instances
- Distributed cache invalidation
- Sub-millisecond lookups

### Redis Implementation Pattern

```go
// Cache a dashboard summary for 5 minutes
redisClient.Set(ctx, "dashboard:summary:user123", jsonData, 5*time.Minute)

// Retrieve cached data
cachedData := redisClient.Get(ctx, "dashboard:summary:user123")

// Invalidate specific cache
redisClient.Del(ctx, "dashboard:summary:user123")

// Invalidate by pattern (requires Redis 6.2+)
redisClient.Keys(ctx, "dashboard:*")
```

### Recommended Redis Keys Pattern

```
dashboard:summary:{user_id}          → Full dashboard data (5 min TTL)
dashboard:recent-findings:{user_id}  → Recent findings list (1 min TTL)
dashboard:kpi-forecast               → KPI forecast chart (10 min TTL)
dashboard:anomalies                  → Anomaly detection results (10 min TTL)
dashboard:risk-summary               → Risk summary stats (5 min TTL)
```

### Cache Invalidation Triggers

```go
// When a finding is created/updated
findings.Create() {
    // ... create logic ...
    redis.Del("dashboard:recent-findings:*")  // Pattern delete
    redis.Del("dashboard:summary:*")
    redis.Del("dashboard:risk-summary")
}

// When a risk is modified
risks.Update() {
    // ... update logic ...
    redis.Del("dashboard:risk-summary")
    redis.Del("dashboard:kpi-forecast")
}

// When an assignment changes
assignments.Update() {
    // ... update logic ...
    redis.Del("dashboard:*")  // Clear entire dashboard cache
}
```

---

## Cache Invalidation Strategy

### HTTP Cache Invalidation (Automatic)

With HTTP cache headers, invalidation happens automatically when:

1. **Time-Based Expiration**: Cache expires when `max-age` is exceeded
2. **ETag Mismatch**: Browser sends `If-None-Match` header, server returns 304 Not Modified
3. **Manual Invalidation**: Clients can send `Cache-Control: no-cache` header

### Cache Busting for Critical Updates

When data must be invalidated immediately (e.g., after a finding is created):

#### Option 1: Version Query Parameter (Recommended)

```javascript
// Frontend: Add version parameter to force cache refresh
const response = await fetch(
    `/api/v1/audit-result-reports/recent-findings?v=${Date.now()}`,
    { method: 'GET' }
);
```

#### Option 2: Conditional Request Headers

```javascript
// Frontend: Use If-None-Match with browser cache
const response = await fetch('/api/v1/audit-result-reports/recent-findings', {
    method: 'GET',
    headers: {
        'If-None-Match': previousETag  // If data hasn't changed, 304 returned
    }
});
```

#### Option 3: Redis Invalidation (Production)

```go
// Backend: Explicitly invalidate caches when data changes
func (s *FindingService) CreateFinding(finding *models.Finding) error {
    // Create the finding
    if err := s.repo.Create(finding); err != nil {
        return err
    }
    
    // Invalidate related caches
    if s.redis != nil {
        s.redis.Del(context.Background(), "dashboard:recent-findings:*")
        s.redis.Del(context.Background(), "dashboard:summary:*")
    }
    
    return nil
}
```

### Cache Invalidation Events

| Event | Affected Caches | TTL Until Expiration |
|-------|-----------------|----------------------|
| Finding Created | recent-findings, summary | 1 min |
| Risk Updated | risk-summary, kpi-forecast | 5-10 min |
| Assignment Changed | assignments, summary | 5 min |
| Employee Updated | employees, summary | 1 hour |
| Department Changed | departments, summary | 1 hour |
| RCM Modified | rcm, control-effectiveness | 10 min |

---

## Monitoring & Metrics

### Key Metrics to Track

```
1. Cache Hit Ratio
   - Percentage of requests served from cache
   - Goal: > 60% for dashboard endpoints
   
2. Average Response Time
   - Before: Baseline from current implementation
   - After: Should improve by 2-3x
   
3. Database Query Count
   - Measure queries during peak hours
   - Goal: 30-40% reduction
   
4. Server CPU Usage
   - Monitor CPU during dashboard loads
   - Goal: 20-25% reduction
```

### Observability

Add logging to cache middleware:

```go
// In middleware ResponseCache()
if cached {
    logger.Info("Cache HIT", 
        logger.LogField("path", c.Request.URL.Path),
        logger.LogField("ttl", config.MaxAge))
} else {
    logger.Info("Cache MISS",
        logger.LogField("path", c.Request.URL.Path))
}
```

---

## Frontend Integration

The frontend already supports HTTP caching through the browser's native fetch API.

### Browser Cache Behavior

```javascript
// Browser respects cache headers automatically
const response = await fetch('/api/v1/audit-result-reports/recent-findings');
// If server sent Cache-Control: max-age=300, browser caches for 5 minutes
// Subsequent requests within 5 minutes use cached response (200 from cache)
```

### Explicit Cache Control

```javascript
// Force fresh data (ignore cache)
const response = await fetch('/api/v1/audit-result-reports/recent-findings', {
    cache: 'no-store'  // Bypass cache entirely
});

// Use cache if available, otherwise fetch fresh
const response = await fetch('/api/v1/audit-result-reports/recent-findings', {
    cache: 'default'  // Respects Cache-Control headers
});
```

### Dashboard Refresh Behavior

The dashboard sync button should clear browser cache:

```typescript
// In dashboard page
handleSync() {
    // Add version parameter to force fresh API call
    this.isSyncing = true;
    
    // Force cache bust by adding timestamp
    const url = `/api/v1/audit-result-reports/recent-findings?t=${Date.now()}`;
    
    const response = await $fetch(url, { method: 'GET' });
    // This will bypass browser cache due to different URL
}
```

---

## Deployment Checklist

### Pre-Deployment
- [ ] Review cache TTLs with business stakeholders
- [ ] Identify critical endpoints that need shorter TTLs
- [ ] Plan Redis deployment (if using Phase 3)
- [ ] Prepare monitoring dashboards

### Deployment
- [ ] Deploy cache middleware to all services
- [ ] Enable HTTP cache headers in Kong (CORS expose headers)
- [ ] Deploy database indexes (if using Phase 2)
- [ ] Deploy Redis caching layer (if using Phase 3)

### Post-Deployment
- [ ] Monitor cache hit ratios
- [ ] Track response time improvements
- [ ] Verify database query reduction
- [ ] Monitor CPU and memory usage
- [ ] Collect user feedback on performance

### Rollback Plan
- [ ] If issues detected: Disable cache middleware in Kong
- [ ] Revert database indexes if needed
- [ ] Disable Redis caching layer

---

## Configuration Examples

### Adjusting Cache TTLs

To change TTLs for specific endpoints:

1. **Audit Service** - Edit `pkg/middleware/cache.go`:
```go
DefaultCacheConfigs["GET /api/v1/audit-result-reports/recent-findings"] = CacheConfig{
    MaxAge:   120,  // Changed from 60 to 120 seconds
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```

2. **Analytics Service** - Edit `middleware/cache.go`:
```go
DefaultCacheConfigs["GET /api/analytics/report"] = CacheConfig{
    MaxAge:   600,  // Changed from 300 to 600 seconds
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```

3. **Risk Service** - Edit `middleware/cache.go`:
```go
DefaultCacheConfigs["GET /api/v1/rcm/summary"] = CacheConfig{
    MaxAge:   600,  // Changed from 300 to 600 seconds
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```

4. **Master Service** - Edit `pkg/middleware/cache.go`:
```go
DefaultCacheConfigs["GET /api/v1/employees"] = CacheConfig{
    MaxAge:   7200, // Changed from 3600 to 7200 seconds (2 hours)
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```

### Adding New Cached Endpoints

```go
// In the appropriate middleware cache.go file
DefaultCacheConfigs["GET /api/v1/new-endpoint"] = CacheConfig{
    MaxAge:   300,  // 5 minutes
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```

---

## Troubleshooting

### Issue: Cache Headers Not Appearing

**Cause**: Middleware not enabled or route not matching

**Solution**:
1. Verify middleware is added to `router.Use()`
2. Check route pattern matches exactly
3. Restart service and verify with `curl -i`

### Issue: Stale Data in Cache

**Cause**: TTL too long for changing data

**Solution**:
1. Reduce `MaxAge` for the endpoint
2. Implement cache invalidation on write
3. Add `Cache-Control: no-cache` for critical data

### Issue: Cache Not Working in Browser

**Cause**: Browser privacy mode, CORS issues, or caching disabled

**Solution**:
1. Test with `curl -i` to verify server headers
2. Check CORS headers in response
3. Clear browser cache and retry
4. Check browser DevTools Network tab

---

## References

- [MDN: HTTP Caching](https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching)
- [RFC 9111: HTTP Caching](https://tools.ietf.org/html/rfc9111)
- [Gin Framework Middleware Guide](https://gin-gonic.com/docs/examples/using-middleware/)
- [Redis Cache Patterns](https://redis.io/docs/manual/patterns/distributed-locks/)

---

## Appendix: Middleware Implementation Details

### Cache Middleware Architecture

```
HTTP Request
    ↓
┌─────────────────────────┐
│ ResponseCache()         │
│ Middleware              │
└─────────────────────────┘
    ↓
Check Request Method (GET/HEAD/POST)
    ↓
    ├─ Not cacheable → Skip cache headers
    │
    └─ Cacheable → Look up CacheConfig
        ↓
        ├─ No config found → Skip cache headers
        │
        └─ Config found → Continue
            ↓
            Wrap response writer to capture body
            ↓
            Process request (call handler)
            ↓
            Check response status (200-299)
            ↓
            ├─ Not successful → Skip cache headers
            │
            └─ Successful → Generate cache headers
                ├─ Generate ETag (MD5 hash)
                ├─ Build Cache-Control header
                ├─ Add Vary header
                ├─ Add Last-Modified header
                └─ Add Expires header
                    ↓
            Send response with cache headers
```

### Response Writer Wrapping

```go
type responseWriter struct {
    gin.ResponseWriter
    body []byte  // Captures response body for ETag
}

func (w *responseWriter) Write(b []byte) (int, error) {
    w.body = append(w.body, b...)  // Capture
    return w.ResponseWriter.Write(b)  // Original write
}
```

This pattern allows:
- Capturing response body without affecting headers
- Computing ETag from actual response data
- No performance penalty (single buffer copy)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-10-04 | Initial implementation of Phase 1 (HTTP cache headers) |
| 2.0 | TBD | Phase 2: Database indexes |
| 3.0 | TBD | Phase 3: Redis caching |

---

**Document Status**: ✓ Phase 1 Complete  
**Last Updated**: 2026-10-04  
**Next Review**: 2026-11-04
