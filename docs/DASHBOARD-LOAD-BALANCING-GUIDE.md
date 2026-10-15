# Dashboard Load Balancing & Caching Configuration Guide

## Overview

This guide documents the load balancing and caching infrastructure for the AuditSphere dashboard, which handles high concurrent API call volume. The setup uses Kong gateway upstreams with health checks, multi-tier caching, and intelligent rate limiting.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend Dashboard                         │
│           (Multiple concurrent API calls)                     │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────▼──────────┐
                    │   Kong Gateway     │
                    │   API Gateway      │
                    └─────────┬──────────┘
                              │
         ┌────────────────────┼────────────────────┐
         │                    │                    │
         ▼                    ▼                    ▼
   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
   │  Analytics   │   │ Audit Svc +  │   │ Master Data  │
   │  Service     │   │ Risk Service │   │  + Risk Svc  │
   │              │   │              │   │              │
   │ Upstream LB: │   │ Upstream LB: │   │ Upstream LB: │
   │ least_conn   │   │ least_conn   │   │ round_robin  │
   │ (1-N targets)│   │ (1-N targets)│   │ (1-N targets)│
   └──────┬───────┘   └──────┬───────┘   └──────┬───────┘
          │                  │                  │
   ┌──────▼──────────────────▼──────────────────▼───────┐
   │         Kong Response Cache Layer                   │
   │   (600-10min TTL per endpoint type)                │
   │   Cache Key: [method][path][vary_headers]          │
   │   Returns: X-Cache: HIT/MISS header               │
   └──────┬───────────────────────────────────────────┘
          │
   ┌──────▼──────────────────────────────┐
   │      PostgreSQL + Redis Cache        │
   │      (Multi-tenant database)         │
   └──────────────────────────────────────┘
```

## Configuration Files

### 1. Kong Load Balancing Configuration

**Locations:**
- Production: `/backend/kong-gateway/kong/prod/kong.yml`
- Development: `/backend/kong-gateway/kong/dev/kong-dev.yml`

### 2. Prometheus Alerts

**Location:** `/monitoring/prometheus-dashboard-alerts.yml`

### 3. Docker Compose

**Location:** `/backend/docker-compose.prod.yml`

## Load Balancing Strategy

### Upstream Algorithm Selection

The configuration uses two primary algorithms:

#### 1. **Least Connections** (Analytics & Audit Services)

```yaml
analytics-service-upstream:
  algorithm: least_connections
  slots: 25  # connection pool size per worker
```

**Best for:** Dashboard queries (long-running, complex analytics)
- Routes requests to backend with fewest active connections
- Better for uneven query duration
- Prevents overloading single slow backend

**Examples:**
- Analytics service: `/api/analytics/*` - expensive ML model queries
- Audit service: Large audit dataset aggregations

#### 2. **Round Robin** (Master & Risk Services)

```yaml
master-service-upstream:
  algorithm: round_robin
  slots: 15
```

**Best for:** Quick lookups and list queries
- Simple distribution across backends
- Good for consistent, fast queries
- Lower overhead than least-connections

**Examples:**
- Master service: Companies, departments, employees
- Risk service: Risk lists and metadata

### Scaling: Adding Service Instances

To add more instances to handle dashboard load:

**Step 1: Scale in Docker Compose**

```yaml
# backend/docker-compose.prod.yml
analytics-service-1:
  build:
    context: ./analytics-service
  container_name: rb_audit_analytics_service_1
  ports:
    - "8084:8084"

analytics-service-2:
  build:
    context: ./analytics-service
  container_name: rb_audit_analytics_service_2
  ports:
    - "8085:8084"

analytics-service-3:
  build:
    context: ./analytics-service
  container_name: rb_audit_analytics_service_3
  ports:
    - "8086:8084"
```

**Step 2: Update Kong Upstream Targets**

```yaml
# backend/kong-gateway/kong/prod/kong.yml
upstreams:
  - name: analytics-service-upstream
    algorithm: least_connections
    targets:
      - target: analytics-service-1:8084
        weight: 100
      - target: analytics-service-2:8084
        weight: 100
      - target: analytics-service-3:8084
        weight: 100
```

**Step 3: Verify Configuration**

```bash
# Check upstream targets
curl http://localhost:8009/admin/upstreams/analytics-service-upstream

# Check health status
curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health

# Test load distribution
for i in {1..10}; do
  curl -I http://localhost:8080/api/analytics/report
done
# Watch Kong logs to see which backend each request hit
```

## Health Checks

Kong actively monitors backend service health:

### Active Health Checks

```yaml
healthchecks:
  active:
    http_path: /api/v1/health
    interval: 10s
    timeout: 1s
    healthy:
      successes: 2        # 2 successful checks → mark healthy
    unhealthy:
      http_failures: 3    # 3 failures → mark unhealthy
```

**Behavior:**
- Kong probes `/api/v1/health` every 10 seconds
- If 2 consecutive probes succeed, backend re-enters rotation
- If 3 consecutive probes fail, backend removed from rotation
- Unhealthy backends do not receive new requests

### Passive Health Checks

```yaml
healthchecks:
  passive:
    healthy:
      http_statuses: [200, 201, 204, 301, 302, 304, 307, 308]
      successes: 5
    unhealthy:
      http_failures: 5
      tcp_failures: 5
      timeouts: 5
```

**Behavior:**
- Kong observes real request/response patterns
- 5 consecutive errors → mark unhealthy (remove from rotation)
- Continues active health checks to detect recovery

### Monitoring Health Status

```bash
# View upstream health in production
curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health

# Example output:
# [
#   {
#     "target": "analytics-service-1:8084",
#     "state": "healthy",
#     "upstream_id": "...",
#     "data": { ... }
#   }
# ]

# View Kong logs for health check activity
docker logs -f rb_audit_kong | grep -i health
```

## Caching Strategy

### Cache TTL by Endpoint Type

| Endpoint Type | TTL | Example Paths | Reason |
|---|---|---|---|
| Analytics (Prod) | 600s (10m) | `/api/analytics/report`, `/api/analytics/predict` | Aggregates, ML models can be stale |
| Analytics (Dev) | 120s (2m) | Same | Shorter for development iteration |
| Master Data | 600s (10m) | `/api/v1/companies`, `/api/v1/employees` | Slow-changing reference data |
| Audit Queries | 300s (5m) | `/api/v1/audit-*/`, `/api/v1/working-papers` | Medium freshness requirement |
| Risk Data | 300s (5m) | `/api/v1/risks`, `/api/v1/rcm` | Balance freshness & performance |

### Cache Key Composition

```
[METHOD][PATH][vary_headers]
```

**Example:**
```
GET /api/analytics/report [Accept-Encoding: gzip]
→ Cache key: GET:/api/analytics/report:gzip
```

**Important:** Cache respects request-level parameters:
- Different query parameters → different cache keys (not cached together)
- Same query with same params → cache hit

### Cache Invalidation

Kong does **not** auto-invalidate on backend data changes. Invalidation options:

#### Option 1: TTL Expiration (Current)
- Automatic after TTL
- No action required
- Trade-off: data freshness vs performance

#### Option 2: Manual Invalidation
```bash
# Clear entire Kong proxy cache
docker exec rb_audit_kong kong cache purge

# Or via Admin API
curl -X DELETE http://localhost:8001/cache

# Or flush Kong
docker exec rb_audit_kong kong reload
```

#### Option 3: Application-Level Cache Busting
Frontend can add cache-busting headers:
```javascript
// Disable cache for critical endpoints
fetch('/api/analytics/report', {
  headers: {
    'Cache-Control': 'no-cache'
  }
})
```

## Rate Limiting

### Dashboard Rate Limits (per IP address)

| Endpoint | Minute Limit | Hour Limit | Purpose |
|---|---|---|---|
| Analytics GET | 120 req/min | 3600 req/h | Dashboard refresh (user-initiated) |
| Analytics Batch | 120 req/min | 3600 req/h | Batch analytics operations |
| Analytics Retrain | 10 req/min | 100 req/h | Prevent model retraining abuse |
| Audit Service | 60 req/min | 1000 req/h | Fieldwork + dashboard combined |
| Master Service | (inherited) | (inherited) | Lookups typically fast |
| Risk Service | (inherited) | (inherited) | Risk queries |
| Global Limit | 100 req/min | 5000 req/h | Fallback for unmatched endpoints |

### Monitoring Rate Limits

```bash
# Check if users are hitting limits
docker logs rb_audit_kong | grep -i "rate limit exceeded"

# View Kong rate limiting metrics
curl http://localhost:8009/admin/plugins | jq '.[] | select(.name=="rate-limiting")'

# Query Prometheus for violations
curl 'http://prometheus:9090/api/v1/query?query=rate(kong_rate_limiting_exceeded_total[1m])'
```

### Adjusting Rate Limits

Edit Kong config files and reload:

```yaml
# Increase dashboard analytics limit to 200 req/min
- name: analytics-routes-cached
  plugins:
    - name: rate-limiting
      config:
        minute: 200  # was 120
        hour: 3600
```

Reload Kong:
```bash
docker-compose restart kong
# Or for zero-downtime:
docker exec rb_audit_kong kong reload
```

## Connection Pooling

### Kong → Backend Services

Kong maintains HTTP keep-alive connection pools to backends.

**Current Configuration (implicit):**
- Kong HTTP worker processes: 1 (set in docker-compose)
- Upstream connection pool per worker: managed by Kong
- Default max connections per upstream: 1000 (Kong default)

**To increase for high-concurrency dashboard:**

```yaml
# backend/docker-compose.prod.yml
kong:
  environment:
    KONG_NGINX_WORKER_PROCESSES: 4    # Match CPU cores
    KONG_UPSTREAM_KEEPALIVE: 100      # HTTP keepalive connections per worker
```

Then restart:
```bash
docker-compose restart kong
```

### Database Connection Pool

PostgreSQL connection pool is managed by backend services (configured at service level, not Kong).

**Current setup:** Each backend service has its own pool (typically 5-20 connections)

**To monitor:**
```bash
# Check active connections
docker exec rb_audit_postgres psql -U postgres -c 'SELECT count(*) FROM pg_stat_activity;'

# Check by service
docker exec rb_audit_postgres psql -U postgres -c 'SELECT application_name, count(*) FROM pg_stat_activity GROUP BY application_name;'
```

**If pool exhausted:**
```bash
# Kill idle connections (those idle > 5 min)
docker exec rb_audit_postgres psql -U postgres -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle'
  AND state_change < now() - interval '5 minutes'
  AND datname = 'rb_audit';
"
```

## Monitoring & Alerts

### Prometheus Metrics Available

Kong exports the following metrics for dashboard monitoring:

```promql
# Cache metrics
kong_cache_hits_total{service="analytics-service"}
kong_cache_misses_total{service="analytics-service"}

# Request metrics
kong_http_requests_total{service="analytics-service",status="200"}
kong_request_upstream_duration_ms{service="analytics-service"}

# Rate limiting
kong_rate_limiting_exceeded_total{service="analytics-service"}

# Upstream health
kong_upstream_target_health{upstream="analytics-service-upstream",state="healthy|unhealthy"}

# Connection pool
kong_upstream_connections_active{upstream="analytics-service-upstream"}
```

### Prometheus Queries

**Dashboard Cache Hit Ratio:**
```promql
sum(rate(kong_cache_hits_total{service="analytics-service"}[5m])) /
(sum(rate(kong_cache_hits_total{service="analytics-service"}[5m])) +
 sum(rate(kong_cache_misses_total{service="analytics-service"}[5m])))
```

**P95 Response Time:**
```promql
histogram_quantile(0.95,
  sum(rate(kong_request_upstream_duration_ms_bucket{service="analytics-service"}[5m]))
  by (le)
)
```

**Request Rate by Service:**
```promql
sum(rate(kong_http_requests_total{service=~"analytics-service|audit-service|master-service"}[1m])) by (service)
```

**Error Rate:**
```promql
sum(rate(kong_http_requests_total{service="analytics-service",status=~"5.."}[5m])) /
sum(rate(kong_http_requests_total{service="analytics-service"}[5m]))
```

### Alert Rules

**Deployed at:** `/monitoring/prometheus-dashboard-alerts.yml`

**Key Alerts:**
- `DashboardHighP95ResponseTime` - P95 > 500ms
- `DashboardCacheLowHitRatio` - Cache hit ratio < 60%
- `AnalyticsServiceCriticalLoad` - Analytics request rate > 200 req/sec
- `AnalyticsServiceUnhealthy` - Analytics upstream target unhealthy
- `DashboardServiceErrorRate` - Error rate > 5%

**Deploy alerts:**
```bash
docker cp /monitoring/prometheus-dashboard-alerts.yml rb_audit_prometheus:/etc/prometheus/rules/dashboard-alerts.yml
docker exec rb_audit_prometheus kill -HUP 1
```

**View active alerts:**
```bash
curl http://prometheus:9090/api/v1/alerts
```

## Performance Tuning

### Identify Bottleneck

1. **Check Cache Hit Ratio** (should be > 60%)
   ```bash
   curl 'http://prometheus:9090/api/v1/query?query=kong_cache_hits_total{service="analytics-service"}'
   ```

2. **Check Response Time** (P95 should be < 500ms)
   ```bash
   # Via Prometheus
   # Via Kong logs
   docker logs rb_audit_kong | grep "upstream_response_time"
   ```

3. **Check Error Rate** (should be < 1%)
   ```bash
   curl 'http://prometheus:9090/api/v1/query?query=kong_http_requests_total{service="analytics-service",status=~"5.."}'
   ```

### If Cache Hit Ratio is Low (< 60%)

**Symptom:** `X-Cache: MISS` on most requests

**Solution:**
1. Increase cache TTL (if data freshness permits)
   ```yaml
   proxy-cache:
     cache_ttl: 900  # was 600 (15 min instead of 10 min)
   ```

2. Check for cache-busting query parameters
   ```bash
   # Review Kong access logs
   docker logs rb_audit_kong | grep "?.*=" | head -20
   # Look for dynamic params (timestamps, unique IDs)
   ```

3. Increase cache size
   ```yaml
   # kong docker-compose environment
   KONG_CACHE_SIZE: 512m  # default is ~128m
   ```

### If Response Time is High (P95 > 500ms)

**Symptom:** Dashboard feels slow, users report delays

**Solution:**
1. Check cache hit ratio (low cache hits cause slow responses)
   - If low, see "If Cache Hit Ratio is Low"

2. Check analytics-service CPU
   ```bash
   docker stats rb_audit_analytics_service
   # If CPU > 80%, scale up service instances
   ```

3. Check database slow queries
   ```bash
   docker exec rb_audit_postgres psql -U postgres -c 'SELECT * FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;'
   ```

4. Enable Kong request buffering
   ```yaml
   kong:
     environment:
       KONG_NGINX_PROXY_PROXY_BUFFERING: "on"
       KONG_NGINX_PROXY_BUFFER_SIZE: 4k
       KONG_NGINX_PROXY_BUFFERS: "8 4k"
   ```

### If Error Rate is High (> 5%)

**Symptom:** Dashboard shows errors, users cannot load data

**Solution:**
1. Check backend service health
   ```bash
   curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health
   ```

2. Check service logs
   ```bash
   docker logs -f rb_audit_analytics_service | grep -i error
   ```

3. Check database connectivity
   ```bash
   docker exec rb_audit_postgres pg_isready
   ```

4. Check for OOM (out of memory)
   ```bash
   docker stats rb_audit_analytics_service --no-stream
   # If MEM% > 90%, restart service
   docker-compose restart analytics-service
   ```

## Testing & Validation

### Test Load Balancing

```bash
# Start 100 sequential requests and observe which backend handles each
for i in {1..100}; do
  echo "Request $i:"
  curl -I http://localhost:8080/api/analytics/report 2>&1 | grep -E "HTTP|Via"
  sleep 0.1
done

# Check Kong logs to see distribution
docker logs rb_audit_kong | grep "upstream" | tail -20
```

### Test Caching

```bash
# First request (cache miss)
time curl -I http://localhost:8080/api/analytics/report
# Look for: X-Cache: MISS

# Second request (cache hit)
time curl -I http://localhost:8080/api/analytics/report
# Look for: X-Cache: HIT
# Notice: second request is faster

# Verify cache is separate per query param
curl -I "http://localhost:8080/api/analytics/report?q=1"
curl -I "http://localhost:8080/api/analytics/report?q=2"
# Each gets separate cache entry
```

### Test Health Checks

```bash
# Simulate unhealthy backend
docker stop rb_audit_analytics_service

# Requests should fail-over to next healthy target
for i in {1..5}; do
  curl http://localhost:8080/api/analytics/report 2>&1
done

# Restart backend
docker start rb_audit_analytics_service

# After ~20s (2 successful health checks), backend is back in rotation
```

### Load Test

```bash
# Install Apache Bench
# brew install httpd

# Simulate 100 concurrent users making 1000 requests each
ab -n 1000 -c 100 -H "Accept: application/json" http://localhost:8080/api/analytics/report

# Key metrics to observe:
# - Requests per second (should be > 100)
# - Median response time (should be < 500ms)
# - Failed requests (should be 0)
```

## Rollback Plan

If dashboard load balancing causes issues:

### Step 1: Disable Load Balancing (use direct URLs)

```yaml
# Temporarily revert to direct service URLs
analytics-service:
  url: http://analytics-service:8084  # instead of upstream
```

```bash
docker-compose restart kong
```

### Step 2: Disable Caching

```yaml
# Remove proxy-cache plugin from analytics routes
routes:
  - name: analytics-routes-cached
    plugins: []  # remove proxy-cache
```

```bash
docker-compose restart kong
```

### Step 3: Restart Services

```bash
docker-compose restart
```

### Step 4: Verify Dashboard Works

```bash
curl http://localhost:8080/api/analytics/report
```

## Summary: Performance Gains

This load balancing & caching setup provides:

| Feature | Benefit | Implementation |
|---|---|---|
| **Load Balancing** | Distribute dashboard queries across N instances | Kong upstreams with least-connections |
| **Health Checks** | Auto-failover to healthy backends | Active + passive health checks |
| **Response Caching** | 60-70% cache hit ratio for aggregates | Kong proxy-cache (10 min TTL) |
| **Rate Limiting** | Prevent API abuse, protect from traffic spikes | Per-endpoint rate limits |
| **Connection Pooling** | Reuse connections, reduce overhead | Kong HTTP keepalive pools |
| **Monitoring** | Visibility into performance issues | Prometheus metrics + alerting |

**Expected Improvements:**
- Dashboard load time: 30-50% faster (with cache hits)
- Backend CPU usage: 40-60% lower (with load distribution)
- Error rate: Near-zero with health checks
- User experience: Smooth dashboard even under peak load

