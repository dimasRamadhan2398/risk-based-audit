# Cache Middleware Usage Guide

Quick reference for developers on how to use and configure the dashboard caching system.

## Overview

The caching system uses HTTP cache headers to optimize dashboard performance. It's automatically enabled on all services and requires no code changes to start working.

## Files Modified

### Cache Middleware Files Created

```
analytics-service/middleware/cache.go
audit-service/pkg/middleware/cache.go
risk-service/middleware/cache.go
master-service/pkg/middleware/cache.go
```

### Main Service Files Updated

```
analytics-service/cmd/main.go           - Added middleware import and usage
audit-service/cmd/serve.go              - Added middleware import and usage
risk-service/cmd/serve.go               - Added middleware import and usage
master-service/cmd/serve.go             - Added middleware import and usage
```

## How It Works

1. **Automatic Detection**: When a GET request is made to the service
2. **Config Lookup**: Middleware checks if the endpoint has a cache configuration
3. **Processing**: Request is processed normally by the handler
4. **Header Addition**: If successful (200-299), cache headers are added to response
5. **Client Caching**: Browser/client caches response for the specified TTL

## Verifying Cache is Working

### Using curl

```bash
# Check that cache headers are present
curl -i "http://localhost:8002/api/v1/audit-result-reports/recent-findings"

# Look for these headers in response:
# Cache-Control: public, max-age=60
# ETag: "a1b2c3d4e5f6g7h8"
# Last-Modified: Mon, 04 Oct 2026 14:00:00 GMT
# Expires: Mon, 04 Oct 2026 14:01:00 GMT
```

### Using Browser DevTools

1. Open Chrome DevTools (F12)
2. Go to Network tab
3. Make a request to a dashboard API
4. Click the request in the network log
5. Go to Response Headers tab
6. Look for `Cache-Control` header with `max-age` value

### Making Conditional Request

```bash
# First request (gets full response)
curl -i "http://localhost:8002/api/v1/audit-result-reports/recent-findings" > first.txt

# Extract ETag from headers
ETAG=$(grep -i "^etag:" first.txt | cut -d' ' -f2)

# Second request with If-None-Match (should return 304)
curl -i -H "If-None-Match: $ETAG" "http://localhost:8002/api/v1/audit-result-reports/recent-findings"

# Expected: 304 Not Modified (client uses cached version)
```

## Adding Cache to a New Endpoint

### Step 1: Identify the Endpoint

```go
// Example: A new GET endpoint in audit-service
api.GET("/api/v1/my-new-endpoint", handler)
```

### Step 2: Add Configuration

Edit the service's middleware/cache.go file and add to `DefaultCacheConfigs`:

```go
// For audit-service/pkg/middleware/cache.go
DefaultCacheConfigs["GET /api/v1/my-new-endpoint"] = CacheConfig{
    MaxAge:   300,  // 5 minutes
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```

### Step 3: Rebuild and Test

```bash
# Rebuild the service
cd /Users/a/Projects/risk-based-audit/backend/audit-service
go build ./...

# Test the endpoint
curl -i "http://localhost:8002/api/v1/my-new-endpoint"

# Verify cache headers are present
```

## Changing Cache TTL

To modify the cache time-to-live (TTL) for an endpoint:

### Find the Endpoint

Locate the endpoint in the appropriate middleware file:

```go
// In audit-service/pkg/middleware/cache.go
DefaultCacheConfigs["GET /api/v1/audit-result-reports/recent-findings"] = CacheConfig{
    MaxAge:   60,  // Currently 1 minute
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```

### Modify the MaxAge Value

```go
// Change from 60 to 120 seconds (2 minutes)
DefaultCacheConfigs["GET /api/v1/audit-result-reports/recent-findings"] = CacheConfig{
    MaxAge:   120,  // Changed from 60 to 120
    IsPublic: true,
    VaryBy:   "Accept-Encoding",
}
```

### Rebuild and Deploy

```bash
cd /Users/a/Projects/risk-based-audit/backend/audit-service
go build ./...
# Deploy the updated service
```

## Cache Configuration Options

### CacheConfig Structure

```go
type CacheConfig struct {
    MaxAge   int    // Time-to-live in seconds
    IsPublic bool   // true = "public" (any cache can store)
                    // false = "private" (user-specific data)
    VaryBy   string // Header to vary cache by (usually "Accept-Encoding")
}
```

### Common TTL Values

| TTL | Use Case | Example |
|-----|----------|---------|
| 60s | Frequently updated data | Recent findings |
| 300s (5 min) | Dashboard stats | Risk summary, metrics |
| 600s (10 min) | Periodic updates | Audit plans, RCM data |
| 1800s (30 min) | Stable data | Charter, mandates |
| 3600s (1 hour) | Master data | Employees, departments |

### IsPublic Flag

```go
// Use IsPublic: true for aggregate/public data
CacheConfig{
    MaxAge:   300,
    IsPublic: true,  // Can be cached by browsers, proxies, CDNs
}

// Use IsPublic: false for user-specific data
CacheConfig{
    MaxAge:   300,
    IsPublic: false, // Only browser cache, not shared proxies
}
```

## Disabling Cache for an Endpoint

To exclude an endpoint from caching, simply don't add it to `DefaultCacheConfigs`:

```go
// This endpoint won't be cached (no entry in DefaultCacheConfigs)
api.GET("/api/v1/sensitive-endpoint", handler)
```

If you need to explicitly prevent caching on a cached endpoint, modify the middleware:

```go
// In middleware, before processing
if c.Request.URL.Path == "/api/v1/special-case" {
    c.Next()  // Skip caching for this request
    return
}
```

## Performance Tuning

### Checking Cache Hit Ratio

Add logging to the middleware to track cache effectiveness:

```go
// In ResponseCache() middleware
if /* found cache config */ {
    logger.Info("Cacheable endpoint",
        logger.LogField("path", c.Request.URL.Path),
        logger.LogField("ttl", config.MaxAge),
    )
}
```

### Monitoring Query Performance

Before and after implementing caching:

```bash
# Monitor database queries
tail -f /var/log/postgres/slow_query.log

# Query response times in browser (DevTools > Network)
# Compare API response times before and after
```

### Adjusting TTLs Based on Load

```
High Traffic Period → Longer TTLs (reduce DB hits)
   300s → 600s

Data Freshness Critical → Shorter TTLs (more accurate data)
   300s → 60s

Peak Traffic Time → Longer TTLs
   Regular TTL → 2x or 3x TTL
```

## Cache Invalidation

### Automatic Invalidation (HTTP)

Cache automatically expires when TTL is reached. No action needed.

### Manual Invalidation (Browser)

```javascript
// Force fresh data from server
const response = await fetch('/api/v1/audit-result-reports/recent-findings', {
    cache: 'no-store'  // Bypass cache entirely
});

// Or add a version parameter
const url = `/api/v1/audit-result-reports/recent-findings?v=${Date.now()}`;
const response = await fetch(url);
```

### Server-Side Invalidation (Future)

For Redis-based caching (Phase 3):

```go
// When data changes, invalidate cache
func (s *Service) UpdateData(data *Model) error {
    // Update the data
    if err := s.repo.Update(data); err != nil {
        return err
    }
    
    // Invalidate cache
    if s.redis != nil {
        s.redis.Del(context.Background(), "cache:key:pattern:*")
    }
    
    return nil
}
```

## Testing Cache Headers

### Basic Test Script

```bash
#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test endpoint
URL="http://localhost:8002/api/v1/audit-result-reports/recent-findings"

echo "Testing cache headers for: $URL"
echo "======================================="

# Make request and capture headers
RESPONSE=$(curl -s -i "$URL")

echo "$RESPONSE" | head -20
echo ""

# Check for cache headers
if echo "$RESPONSE" | grep -q "Cache-Control"; then
    echo -e "${GREEN}✓ Cache-Control header found${NC}"
else
    echo -e "${RED}✗ Cache-Control header missing${NC}"
fi

if echo "$RESPONSE" | grep -q "ETag"; then
    echo -e "${GREEN}✓ ETag header found${NC}"
else
    echo -e "${RED}✗ ETag header missing${NC}"
fi

if echo "$RESPONSE" | grep -q "Last-Modified"; then
    echo -e "${GREEN}✓ Last-Modified header found${NC}"
else
    echo -e "${RED}✗ Last-Modified header missing${NC}"
fi

if echo "$RESPONSE" | grep -q "Expires"; then
    echo -e "${GREEN}✓ Expires header found${NC}"
else
    echo -e "${RED}✗ Expires header missing${NC}"
fi
```

## Troubleshooting

### Cache Headers Not Appearing

**Problem**: API response doesn't have Cache-Control headers

**Checklist**:
- [ ] Middleware is imported in main.go/serve.go
- [ ] Middleware is added to `router.Use()` or `engine.Use()`
- [ ] Endpoint is in the `DefaultCacheConfigs` map
- [ ] Route path matches exactly (including method)
- [ ] Response status code is 2xx (200-299)

### Cache Not Persisting in Browser

**Problem**: Browser doesn't cache the response

**Possible Causes**:
- [ ] Response status is not 2xx
- [ ] CORS headers prevent caching
- [ ] Browser privacy mode is enabled
- [ ] Browser cache is disabled in DevTools
- [ ] max-age is set to 0

**Solution**:
```bash
# Test with curl (bypass browser)
curl -i "http://localhost:8002/api/v1/audit-result-reports/recent-findings"

# Compare Vary header value
# Should be "Accept-Encoding" typically
```

### Stale Cache Preventing Updates

**Problem**: Data hasn't changed but cache is being served

**Solution**:
1. Add version parameter to force refresh: `?v=1.2.3`
2. Wait for TTL to expire naturally
3. Clear browser cache manually
4. Implement cache invalidation in handler

### CORS Issues with Caching

**Problem**: Cache headers not exposed to browser

**Solution**: Ensure CORS middleware exposes cache headers:

```go
// In CORS middleware
c.Writer.Header().Set("Access-Control-Expose-Headers", 
    "Content-Length, Cache-Control, ETag, Last-Modified, Expires")
```

## Performance Impact

### Expected Improvements

| Metric | Expected Improvement |
|--------|----------------------|
| API Calls | -40 to -50% |
| DB Queries | -30 to -40% |
| Response Time | 2-3x faster |
| Server CPU | -20 to -25% |

### Measuring Impact

```bash
# Before optimization (baseline)
ab -c 10 -n 1000 http://localhost:8002/api/v1/audit-result-reports/recent-findings

# After optimization (with caching)
# Should see:
# - Faster response times
# - Higher requests per second
# - Lower CPU usage on server
```

## Advanced Usage

### Adding Custom Cache Logic

For endpoints with complex caching needs:

```go
// In the endpoint handler
middleware.SetCacheControl(c, 300, true)  // 5 minutes, public
middleware.SetETag(c, customData)         // Custom ETag generation
```

### Conditional Caching

Cache based on query parameters:

```go
// In middleware, check query parameters
if c.Query("nocache") == "true" {
    c.Next()  // Skip cache for this request
    return
}
```

## References

- [MDN: HTTP Caching](https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching)
- [RFC 9111: HTTP Caching](https://datatracker.ietf.org/doc/html/rfc9111)
- [Gin Middleware Documentation](https://gin-gonic.com/docs/examples/using-middleware/)

---

**Last Updated**: 2026-10-04  
**Status**: Ready for use
