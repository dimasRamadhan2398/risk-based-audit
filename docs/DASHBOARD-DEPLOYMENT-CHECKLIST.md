# Dashboard Load Balancing - Deployment Checklist

## Pre-Deployment Verification

- [ ] **Review Kong Configuration Changes**
  - [ ] Check `/backend/kong-gateway/kong/prod/kong.yml` for syntax errors
  - [ ] Check `/backend/kong-gateway/kong/dev/kong-dev.yml` for syntax errors
  - [ ] Verify all services have upstream references
  - [ ] Verify health check endpoints are correct for each service

  ```bash
  # Validate YAML syntax
  docker run -i --rm mikefarah/yq . < backend/kong-gateway/kong/prod/kong.yml
  docker run -i --rm mikefarah/yq . < backend/kong-gateway/kong/dev/kong-dev.yml
  ```

- [ ] **Verify Backend Services Have Health Endpoints**
  - [ ] Auth service: `GET /api/v1/health`
  - [ ] Audit service: `GET /api/v1/health`
  - [ ] Master service: `GET /api/v1/health`
  - [ ] Risk service: `GET /api/v1/health`
  - [ ] Analytics service: `GET /api/analytics/health`
  - [ ] Python AI: `GET /health`

  ```bash
  # Test each endpoint
  curl http://localhost:8001/api/v1/health
  curl http://localhost:8002/api/v1/health
  curl http://localhost:8084/api/analytics/health
  curl http://localhost:8000/health
  ```

- [ ] **Current Load Test Baseline**
  - [ ] Document current dashboard response time (P95, P99)
  - [ ] Document current cache hit ratio (if applicable)
  - [ ] Document current error rate
  - [ ] Take screenshot of Prometheus metrics

- [ ] **Backup Current Configuration**
  ```bash
  cp backend/kong-gateway/kong/prod/kong.yml backend/kong-gateway/kong/prod/kong.yml.backup
  cp backend/kong-gateway/kong/dev/kong-dev.yml backend/kong-gateway/kong/dev/kong-dev.yml.backup
  ```

## Development Environment Deployment

### Step 1: Deploy Dev Kong Configuration

- [ ] **Restart Kong with new configuration**
  ```bash
  docker-compose restart kong
  ```

- [ ] **Verify Kong starts without errors**
  ```bash
  docker logs -f rb_audit_kong --tail 50
  # Wait for: "Server started successfully"
  ```

- [ ] **Verify upstreams are created**
  ```bash
  curl http://localhost:8009/admin/upstreams
  # Should see: auth-service-upstream, analytics-service-upstream, etc.
  ```

### Step 2: Test Load Balancing (Dev)

- [ ] **Check upstream health status**
  ```bash
  curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health
  # All targets should show "healthy"
  ```

- [ ] **Test routing to analytics service**
  ```bash
  for i in {1..5}; do
    curl -I http://localhost:8080/api/analytics/report
    sleep 0.1
  done
  
  docker logs rb_audit_kong | grep "upstream" | tail -5
  # Should see requests successfully routed
  ```

- [ ] **Verify cache is working**
  ```bash
  # First request (cache miss)
  curl -i http://localhost:8080/api/analytics/report | grep X-Cache
  
  # Second request (cache hit)
  curl -i http://localhost:8080/api/analytics/report | grep X-Cache
  # Should show: X-Cache: HIT (or similar)
  ```

### Step 3: Test Prometheus Alerts (Dev)

- [ ] **Deploy alert rules to Prometheus**
  ```bash
  docker cp monitoring/prometheus-dashboard-alerts.yml rb_audit_prometheus:/etc/prometheus/rules/dashboard-alerts.yml
  docker exec rb_audit_prometheus kill -HUP 1
  ```

- [ ] **Verify alerts are loaded**
  ```bash
  curl http://localhost:9090/api/v1/alerts | jq '.data.groups[] | select(.name | contains("dashboard"))'
  # Should see alert groups from dashboard-alerts.yml
  ```

- [ ] **Trigger test alert (optional)**
  ```bash
  # Simulate high load for 5 minutes to trigger dashboard alerts
  ab -n 1000 -c 50 http://localhost:8080/api/analytics/report &
  sleep 300
  kill %1
  
  # Check Prometheus for triggered alerts
  curl http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.endpoint=="dashboard")'
  ```

### Step 4: Run Dashboard Smoke Test (Dev)

- [ ] **Dashboard loads without errors**
  - [ ] Visit http://localhost:3000/analytics (or dev URL)
  - [ ] Dashboard renders all cards
  - [ ] No console errors in browser dev tools
  - [ ] Network tab shows successful responses (200, 304)

- [ ] **API responses are correct**
  ```bash
  # Verify analytics endpoints respond
  curl http://localhost:8080/api/analytics/report -s | jq '.' | head -20
  curl http://localhost:8080/api/analytics/predict -s | jq '.' | head -20
  curl http://localhost:8080/api/v1/risks -s | jq '.' | head -20
  ```

### Step 5: Monitor for Issues (Dev)

- [ ] **Watch Kong logs for errors**
  ```bash
  docker logs -f rb_audit_kong --tail 20
  ```

- [ ] **Watch Prometheus alerts**
  ```bash
  curl http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.state=="firing")'
  # Should be empty (no alerts firing)
  ```

- [ ] **Check Docker resource usage**
  ```bash
  docker stats --no-stream
  # Kong memory should not increase significantly
  # Analytics service CPU should be normal
  ```

### Step 6: Rollback Decision (Dev)

- [ ] **If any test fails:**
  ```bash
  # Restore backup
  cp backend/kong-gateway/kong/dev/kong-dev.yml.backup backend/kong-gateway/kong/dev/kong-dev.yml
  
  # Restart Kong
  docker-compose restart kong
  
  # Investigate issue and retry
  ```

- [ ] **If all tests pass:** Proceed to production

## Production Environment Deployment

### Step 1: Schedule Maintenance Window (if needed)

- [ ] **Determine deployment timing**
  - [ ] Off-peak hours (e.g., 2 AM - 4 AM)
  - [ ] Notify users of brief API downtime
  - [ ] Have rollback plan ready

### Step 2: Pre-Deployment Checks

- [ ] **Verify production is healthy**
  ```bash
  # SSH to production VPS
  ssh root@auditsphere.app
  
  # Check Docker containers
  docker ps | grep -E "kong|audit|analytics|master"
  
  # Check backend services respond
  curl http://analytics-service:8084/api/analytics/health
  curl http://audit-service:8001/api/v1/health
  curl http://master-service:8002/api/v1/health
  ```

- [ ] **Backup production Kong config**
  ```bash
  ssh root@auditsphere.app
  cp /app/rbia-repo/backend/kong-gateway/kong/prod/kong.yml \
     /app/rbia-repo/backend/kong-gateway/kong/prod/kong.yml.backup.$(date +%s)
  ```

- [ ] **Verify git is clean**
  ```bash
  cd /app/rbia-repo
  git status
  # Should show: "working tree clean" (after this commit)
  ```

### Step 3: Deploy Kong Configuration

- [ ] **Commit and push changes**
  ```bash
  git add backend/kong-gateway/kong/prod/kong.yml
  git add backend/kong-gateway/kong/dev/kong-dev.yml
  git add monitoring/prometheus-dashboard-alerts.yml
  git add docs/DASHBOARD-LOAD-BALANCING-GUIDE.md
  git add docs/DASHBOARD-DEPLOYMENT-CHECKLIST.md
  
  git commit -m "feat: add dashboard load balancing and caching configuration

  - Kong upstreams for analytics, audit, master, risk services
  - Least-connections load balancing for analytics (long queries)
  - Round-robin for master data (fast queries)
  - Health checks (active & passive) for auto-failover
  - Aggressive caching (10m TTL) for dashboard aggregates
  - Dashboard-specific rate limiting (120 req/min)
  - Prometheus alerts for dashboard performance
  
  Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>"
  
  git push origin main
  ```

- [ ] **Deploy via GitHub Actions**
  - [ ] Go to https://github.com/RauMomo/risk-based-audit/actions
  - [ ] Click "AuditSphere CI/CD Pipeline"
  - [ ] Click "Run workflow" → select branch "main"
  - [ ] Monitor deployment progress

  ```bash
  # Or manually trigger if using workflow_dispatch
  gh workflow run deploy.yml --ref main
  ```

- [ ] **Wait for deployment to complete**
  - [ ] Expected time: 5-10 minutes
  - [ ] Check GitHub Actions UI for green checkmark
  - [ ] SSH into VPS and verify Kong restarted

### Step 4: Verify Production Deployment

- [ ] **Kong is running with new config**
  ```bash
  ssh root@auditsphere.app
  docker logs rb_audit_kong --tail 20 | grep -E "started|listening|error"
  ```

- [ ] **Upstreams are created**
  ```bash
  curl http://localhost:8009/admin/upstreams | jq '.[] | .name'
  # Should list: auth-service-upstream, analytics-service-upstream, etc.
  ```

- [ ] **All upstream targets are healthy**
  ```bash
  for upstream in auth-service analytics-service audit-service master-service risk-service; do
    echo "=== $upstream ==="
    curl http://localhost:8009/admin/upstreams/${upstream}-upstream/health | jq '.[] | {target, state}'
  done
  ```

- [ ] **Analytics cache is working**
  ```bash
  # First request (cache miss)
  curl -i https://auditsphere.app/api/analytics/report 2>&1 | grep -E "X-Cache|HTTP"
  
  # Second request (cache hit)
  curl -i https://auditsphere.app/api/analytics/report 2>&1 | grep -E "X-Cache|HTTP"
  # Should show: X-Cache: HIT
  ```

### Step 5: Test Production Dashboard

- [ ] **Dashboard loads successfully**
  - [ ] Visit https://auditsphere.app
  - [ ] Login with test account
  - [ ] Navigate to Analytics/Dashboard
  - [ ] All metrics load without errors
  - [ ] Page feels responsive

- [ ] **Monitor production metrics**
  ```bash
  # SSH to VPS with Prometheus
  ssh root@auditsphere.app
  curl http://localhost:9090/api/v1/query?query='rate(kong_http_requests_total{service="analytics-service"}[5m])'
  curl http://localhost:9090/api/v1/query?query='kong_cache_hits_total{service="analytics-service"}'
  ```

- [ ] **Check for production alerts**
  ```bash
  curl http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.state=="firing")'
  # Should be empty (no alerts)
  ```

### Step 6: Deploy Prometheus Alert Rules

- [ ] **Copy alert rules to Prometheus**
  ```bash
  ssh root@auditsphere.app
  docker cp monitoring/prometheus-dashboard-alerts.yml rb_audit_prometheus:/etc/prometheus/rules/dashboard-alerts.yml
  docker exec rb_audit_prometheus kill -HUP 1
  ```

- [ ] **Verify alerts are loaded**
  ```bash
  curl http://localhost:9090/api/v1/rules | jq '.data.groups[] | select(.name | contains("dashboard"))'
  # Should see all dashboard alert rules
  ```

### Step 7: Monitor for 1 Hour

- [ ] **Check Kong logs** (no errors)
  ```bash
  ssh root@auditsphere.app
  docker logs rb_audit_kong --since 10m | grep -i error
  # Should be empty
  ```

- [ ] **Monitor dashboard response times**
  - [ ] Prometheus: P95 response time should be < 500ms
  - [ ] Prometheus: Cache hit ratio should be > 60%
  - [ ] Prometheus: Error rate should be < 1%

- [ ] **Check user reports**
  - [ ] No user-reported dashboard slowness
  - [ ] No user-reported API errors
  - [ ] All audit teams reporting normal dashboard performance

### Step 8: Rollback Decision

- [ ] **If all checks pass:** Proceed to scaling (see below)

- [ ] **If issues detected:** Rollback

  ```bash
  ssh root@auditsphere.app
  cd /app/rbia-repo
  
  # Restore backup
  cp backend/kong-gateway/kong/prod/kong.yml.backup.* backend/kong-gateway/kong/prod/kong.yml
  
  # Restart Kong
  docker-compose restart kong
  
  # Verify rollback
  docker logs rb_audit_kong --tail 20 | grep -E "started|listening"
  curl http://localhost:8009/admin/services | jq '.[] | .upstream' | grep -c "null"
  # Should show services using `url` not `upstream`
  
  # Force re-deploy previous version
  git revert HEAD --no-edit
  git push origin main
  # Trigger GitHub Actions redeploy
  ```

## Post-Deployment: Scaling (Optional)

### When to Scale

Add more service instances if:
- Analytics service CPU consistently > 70%
- Cache hit ratio < 60% (indicates misses due to high volume)
- P95 response time > 500ms under normal load
- Dashboard users report slowness during peak hours

### Scale Analytics Service to 3 Instances

- [ ] **Update docker-compose.prod.yml**
  ```yaml
  analytics-service-1:
    build:
      context: ./analytics-service
    container_name: rb_audit_analytics_service_1
    depends_on:
      postgres:
        condition: service_started
      redis:
        condition: service_healthy
      kafka:
        condition: service_healthy
      python-ai:
        condition: service_started
    ports:
      - "8084:8084"
    networks:
      - rb_audit_network
    restart: unless-stopped

  analytics-service-2:
    build:
      context: ./analytics-service
    container_name: rb_audit_analytics_service_2
    depends_on:
      postgres:
        condition: service_started
      redis:
        condition: service_healthy
      kafka:
        condition: service_healthy
      python-ai:
        condition: service_started
    ports:
      - "8085:8084"
    networks:
      - rb_audit_network
    restart: unless-stopped

  analytics-service-3:
    build:
      context: ./analytics-service
    container_name: rb_audit_analytics_service_3
    depends_on:
      postgres:
        condition: service_started
      redis:
        condition: service_healthy
      kafka:
        condition: service_healthy
      python-ai:
        condition: service_started
    ports:
      - "8086:8084"
    networks:
      - rb_audit_network
    restart: unless-stopped

  kong:
    # Update depends_on to include all instances
    depends_on:
      - auth-service
      - audit-service
      - master-service
      - risk-service
      - analytics-service-1
      - analytics-service-2
      - analytics-service-3
      - python-ai
  ```

- [ ] **Update Kong upstream targets**
  ```yaml
  analytics-service-upstream:
    algorithm: least_connections
    targets:
      - target: analytics-service-1:8084
        weight: 100
      - target: analytics-service-2:8084
        weight: 100
      - target: analytics-service-3:8084
        weight: 100
  ```

- [ ] **Start new instances**
  ```bash
  docker-compose up -d --build analytics-service-1 analytics-service-2 analytics-service-3
  docker-compose restart kong
  ```

- [ ] **Verify all instances are healthy**
  ```bash
  curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health | jq '.[] | {target, state}'
  # All should show "healthy"
  ```

- [ ] **Monitor load distribution**
  ```bash
  # Run 30 requests and check distribution
  for i in {1..30}; do
    curl -i http://localhost:8080/api/analytics/report 2>&1 | grep -E "HTTP/|upstream_addr" &
  done
  wait
  
  docker logs rb_audit_kong | grep "upstream_addr" | tail -30 | sort | uniq -c
  # Should see roughly 10 requests to each instance
  ```

## Monitoring & Alerting Setup

- [ ] **Grafana Dashboard**
  - [ ] Create new dashboard for "Dashboard Performance"
  - [ ] Add panels:
    - [ ] Cache hit ratio over time
    - [ ] P95 response time per service
    - [ ] Request rate per service
    - [ ] Error rate per service
    - [ ] Upstream health status

- [ ] **Alert Channels**
  - [ ] Configure Slack notifications for critical alerts
  - [ ] Configure email for warning alerts
  - [ ] Test alert delivery

## Documentation

- [ ] **Update team wiki**
  - [ ] Link to DASHBOARD-LOAD-BALANCING-GUIDE.md
  - [ ] Post alert runbooks for team
  - [ ] Document scaling procedure

- [ ] **Share with team**
  - [ ] Slack notification about new load balancing
  - [ ] Performance improvements (before/after metrics)
  - [ ] How to troubleshoot dashboard issues

## Sign-Off

- [ ] **Engineering Lead Approval**
  - [ ] Reviewed Kong configuration
  - [ ] Approved deployment timing
  - [ ] Verified rollback plan

- [ ] **DevOps Lead Verification**
  - [ ] Production deployment successful
  - [ ] All health checks passing
  - [ ] Alerts configured and tested
  - [ ] Monitoring in place

- [ ] **Product/Support Sign-Off**
  - [ ] Dashboard performance improvement confirmed
  - [ ] No user-facing issues
  - [ ] Support team notified of new alerts

## Post-Deployment: 24-Hour Monitoring

- [ ] **Monitor error rates**
  ```bash
  curl 'http://prometheus:9090/api/v1/query?query=rate(kong_http_requests_total{status=~"5.."}[5m])'
  # Should be < 1%
  ```

- [ ] **Monitor response times**
  ```bash
  curl 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,kong_request_upstream_duration_ms)'
  # Should be < 500ms
  ```

- [ ] **Monitor cache effectiveness**
  ```bash
  curl 'http://prometheus:9090/api/v1/query?query=rate(kong_cache_hits_total[5m])'
  # Should be high relative to total requests
  ```

- [ ] **Monitor alert firing**
  ```bash
  curl http://localhost:9090/api/v1/alerts?active=true
  # Should be minimal/none
  ```

If any issues detected within 24 hours, be prepared to rollback.

