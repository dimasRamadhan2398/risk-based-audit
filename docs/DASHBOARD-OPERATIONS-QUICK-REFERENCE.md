# Dashboard Load Balancing - Operations Quick Reference

## Rapid Troubleshooting

### Dashboard is Slow (High P95 Response Time)

```bash
# 1. Check if it's a cache problem
curl 'http://prometheus:9090/api/v1/query?query=rate(kong_cache_hits_total{service="analytics-service"}[5m])'

# 2. Check analytics service load
curl 'http://prometheus:9090/api/v1/query?query=sum(rate(kong_http_requests_total{service="analytics-service"}[1m]))'

# 3. Check analytics service CPU
docker stats rb_audit_analytics_service --no-stream

# 4. If CPU > 80%, check if a backend went unhealthy
curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health | jq '.[] | {target, state}'

# 5. Check database connections (if many)
docker exec rb_audit_postgres psql -U postgres -c 'SELECT count(*) FROM pg_stat_activity;'

# 6. Check Kong memory (if cache is small)
docker stats rb_audit_kong --no-stream
```

### Dashboard Returns 502 Bad Gateway

```bash
# 1. Check Kong is running
docker ps | grep kong

# 2. Check Kong logs
docker logs rb_audit_kong --tail 50 | grep -i error

# 3. Check backend service health
curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health | jq '.[] | select(.state=="unhealthy")'

# 4. If backends unhealthy, check their logs
docker logs rb_audit_analytics_service --tail 50 | grep -i error
docker logs rb_audit_audit_service --tail 50 | grep -i error

# 5. If no errors, restart Kong
docker-compose restart kong
```

### High Cache Miss Rate (< 40%)

```bash
# 1. Check if query params are changing
docker logs rb_audit_kong | grep "api/analytics" | head -10
# Look for varying query strings (cache busters)

# 2. Check cache size
docker stats rb_audit_kong --no-stream | grep rb_audit_kong
# If memory is capped, need to increase cache

# 3. Check TTL is not too short
curl http://localhost:8001/services/analytics-service/plugins | jq '.[] | select(.name=="proxy-cache") | .config.cache_ttl'

# 4. If everything looks good, cache may be cold (normal after restart)
# Monitor for 10 minutes, should improve
```

### Rate Limit Errors (429 Too Many Requests)

```bash
# 1. Check if legitimate traffic or attack
docker logs rb_audit_kong | grep "rate_limiting" | tail -10

# 2. Check rate limit configuration
curl http://localhost:8001/services/analytics-service/plugins | jq '.[] | select(.name=="rate-limiting")'

# 3. If legitimate, increase limit
# Edit Kong config and restart:
# analytics-service:
#   plugins:
#     - name: rate-limiting
#       config:
#         minute: 240  # was 120

docker-compose restart kong

# 4. If attack, block IP at firewall
iptables -A INPUT -s <attacker_ip> -j DROP
```

### Database Slow Queries

```bash
# 1. Find slow queries
docker exec rb_audit_postgres psql -U postgres -c 'SELECT query, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 5;'

# 2. Analyze query plan
docker exec rb_audit_postgres psql -U postgres -c "EXPLAIN ANALYZE SELECT COUNT(*) FROM audit_executions;"

# 3. Check table stats are up to date
docker exec rb_audit_postgres psql -U postgres -c "ANALYZE;"

# 4. Check for missing indexes
docker exec rb_audit_postgres psql -U postgres -c "SELECT schemaname, tablename FROM pg_tables WHERE schemaname='public' LIMIT 5;"
```

### Memory Leak (Kong Container Growing)

```bash
# 1. Check Kong memory usage
docker stats rb_audit_kong --no-stream

# 2. Check cache size
curl http://localhost:8001/status | jq '.memory'

# 3. Clear cache to free memory
docker exec rb_audit_kong kong cache purge

# 4. If memory still high, restart Kong
docker-compose restart kong

# 5. Monitor memory after restart
watch -n 5 'docker stats rb_audit_kong --no-stream'
```

## Health Check Commands

### Kong Health

```bash
# Is Kong running?
docker ps | grep kong

# Is Kong's admin API accessible?
curl http://localhost:8009/admin/status | jq .

# Check all upstreams are healthy
curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health | jq '.[] | {target, state}'
curl http://localhost:8009/admin/upstreams/audit-service-upstream/health | jq '.[] | {target, state}'
curl http://localhost:8009/admin/upstreams/master-service-upstream/health | jq '.[] | {target, state}'

# Check services are configured
curl http://localhost:8009/admin/services | jq '.[] | {name, upstream_id}'
```

### Backend Services Health

```bash
# Check all services health
curl http://localhost:8001/api/v1/health  # auth
curl http://localhost:8001/api/v1/health  # audit (8002 mapped to 8001)
curl http://localhost:8002/api/v1/health  # master
curl http://localhost:8002/api/v1/health  # risk (8004 mapped to 8002)
curl http://localhost:8084/api/analytics/health  # analytics
curl http://localhost:8000/health  # python-ai

# Or use docker health check
docker inspect rb_audit_analytics_service | jq '.State.Health'
```

### Database Health

```bash
# Is PostgreSQL running and accepting connections?
docker exec rb_audit_postgres pg_isready

# Check database size
docker exec rb_audit_postgres psql -U postgres -c 'SELECT pg_database.datname, pg_size_pretty(pg_database_size(pg_database.datname)) FROM pg_database;'

# Check connection count
docker exec rb_audit_postgres psql -U postgres -c 'SELECT count(*) FROM pg_stat_activity;'

# Check slow query log
docker exec rb_audit_postgres psql -U postgres -c 'SELECT query, calls, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;'
```

### Redis Health

```bash
# Is Redis running?
docker ps | grep redis

# Ping Redis
docker exec rb_audit_redis redis-cli ping

# Check memory usage
docker exec rb_audit_redis redis-cli info memory
```

## Performance Baseline Metrics

### Collect Current Performance

```bash
#!/bin/bash
# Save this as scripts/collect-dashboard-metrics.sh

TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")

echo "=== Dashboard Performance Baseline ===" 
echo "Timestamp: $TIMESTAMP"
echo ""

echo "=== Cache Hit Ratio ==="
curl -s 'http://prometheus:9090/api/v1/query?query=rate(kong_cache_hits_total{service="analytics-service"}[5m])' | jq '.data.result[0].value[1]'

echo "=== P95 Response Time (ms) ==="
curl -s 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(kong_request_upstream_duration_ms_bucket{service="analytics-service"}[5m]))by(le))' | jq '.data.result[0].value[1]'

echo "=== Error Rate ==="
curl -s 'http://prometheus:9090/api/v1/query?query=rate(kong_http_requests_total{service="analytics-service",status=~"5.."}[5m])' | jq '.data.result[0].value[1]'

echo "=== Request Rate (req/sec) ==="
curl -s 'http://prometheus:9090/api/v1/query?query=sum(rate(kong_http_requests_total{service="analytics-service"}[1m]))' | jq '.data.result[0].value[1]'

echo "=== Kong Memory Usage ==="
docker stats rb_audit_kong --no-stream | awk 'NR==2 {print $6}'

echo "=== Analytics Service CPU ==="
docker stats rb_audit_analytics_service --no-stream | awk 'NR==2 {print $3}'
```

Run before and after deployment:
```bash
bash scripts/collect-dashboard-metrics.sh > metrics-before.txt
# Deploy changes
bash scripts/collect-dashboard-metrics.sh > metrics-after.txt
diff metrics-before.txt metrics-after.txt
```

## Monitoring Queries

### Prometheus Queries for Grafana

**Cache Hit Ratio:**
```promql
sum(rate(kong_cache_hits_total{service="analytics-service"}[5m])) /
(sum(rate(kong_cache_hits_total{service="analytics-service"}[5m])) +
 sum(rate(kong_cache_misses_total{service="analytics-service"}[5m])))
```

**Request Rate by Service:**
```promql
sum(rate(kong_http_requests_total[1m])) by (service)
```

**P95 Response Time:**
```promql
histogram_quantile(0.95,
  sum(rate(kong_request_upstream_duration_ms_bucket[5m]))
  by (le, service)
)
```

**Error Rate:**
```promql
sum(rate(kong_http_requests_total{status=~"5.."}[5m])) by (service) /
sum(rate(kong_http_requests_total[5m])) by (service)
```

**Upstream Health:**
```promql
kong_upstream_target_health{state="healthy"}
```

**Connection Pool Usage:**
```promql
kong_upstream_connections_active
```

## Common Operations

### Clear Kong Cache

```bash
# Remove cached responses
docker exec rb_audit_kong kong cache purge

# Verify cache is cleared
curl -i http://localhost:8080/api/analytics/report | grep X-Cache
# Should show: X-Cache: MISS
```

### Reload Kong Config (Zero-Downtime)

```bash
# Edit Kong config file, then:
docker exec rb_audit_kong kong reload

# Verify reload succeeded
docker logs rb_audit_kong | tail -5 | grep -i reload
```

### Restart a Backend Service

```bash
# Restart analytics service (removed from load balance automatically)
docker-compose restart analytics-service

# Verify it's back in rotation (2-3 seconds for health check)
sleep 3
curl http://localhost:8009/admin/upstreams/analytics-service-upstream/health | jq '.[] | select(.target | contains("analytics")) | .state'
# Should show: "healthy"
```

### Scale Service Up

```bash
# Add new analytics-service-2
docker-compose up -d analytics-service-2

# Update Kong upstream to include new target (edit kong.yml)
# Then reload Kong
docker exec rb_audit_kong kong reload

# Verify new target is in rotation
curl http://localhost:8009/admin/upstreams/analytics-service-upstream | jq '.targets[].target'
```

### Rotate Logs

```bash
# Clear old Kong logs (keep recent 10MB)
docker logs rb_audit_kong --tail 50 > /tmp/kong-logs-backup.txt
docker exec rb_audit_kong rm /var/log/kong/*.log || true
```

## Alerting Integration

### Enable Slack Notifications

```bash
# 1. Get Slack webhook URL from your workspace
# https://api.slack.com/messaging/webhooks

# 2. Add to Prometheus config
# alertmanager:
#   slack_api_url: https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# 3. Configure alert routing
# Route dashboard alerts to #alerts-dashboard
# Route critical to #critical

# 4. Test alert
curl -X POST \
  -H 'Content-type: application/json' \
  --data '{"text":"Test alert from Prometheus"}' \
  <slack_webhook_url>
```

### Mute Alerts Temporarily

```bash
# Via Prometheus UI: http://prometheus:9090
# Click "Alerts" → Find alert → "Mute"

# Or via API
curl -X POST \
  -H 'Content-type: application/json' \
  --data '{"matchers":["instance=analytics-service"],"duration":"1h"}' \
  http://localhost:9093/api/v1/silences
```

## Documentation Links

- **Full Guide:** [DASHBOARD-LOAD-BALANCING-GUIDE.md](./DASHBOARD-LOAD-BALANCING-GUIDE.md)
- **Deployment Steps:** [DASHBOARD-DEPLOYMENT-CHECKLIST.md](./DASHBOARD-DEPLOYMENT-CHECKLIST.md)
- **Alert Rules:** `/monitoring/prometheus-dashboard-alerts.yml`
- **Kong Config (Prod):** `/backend/kong-gateway/kong/prod/kong.yml`
- **Kong Config (Dev):** `/backend/kong-gateway/kong/dev/kong-dev.yml`

## Emergency Procedures

### Complete Rollback

```bash
# 1. Stop Kong
docker-compose stop kong

# 2. Restore backup Kong config
cp /backup/kong.yml.backup backend/kong-gateway/kong/prod/kong.yml

# 3. Restart Kong with old config
docker-compose up -d kong

# 4. Verify old config is active
curl http://localhost:8009/admin/services | jq '.[] | .url' | grep -c "http://"
# Should show services with `url` property, not `upstream`

# 5. Notify team
# "Load balancing disabled, reverted to direct service URLs"
```

### Kill Stuck Requests

```bash
# If Kong is stuck processing a request:

# 1. Check running processes
docker top rb_audit_kong

# 2. Gracefully restart Kong
docker-compose restart kong --grace-period=30

# 3. If that fails, force kill
docker-compose kill -s KILL kong
docker-compose up -d kong
```

### Database Recovery

```bash
# If database is overwhelmed:

# 1. Kill long-running queries
docker exec rb_audit_postgres psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE query_start < now() - interval '5 minutes';"

# 2. Clear idle connections
docker exec rb_audit_postgres psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state='idle';"

# 3. Restart PostgreSQL
docker-compose restart postgres

# 4. Restart backend services (they'll reconnect)
docker-compose restart analytics-service audit-service master-service risk-service
```

## Contact & Escalation

- **Kong Issues:** Check Kong logs → `docker logs rb_audit_kong`
- **Backend Issues:** Check service logs → `docker logs rb_audit_analytics_service`
- **Database Issues:** Check PostgreSQL logs → `docker logs rb_audit_postgres`
- **Metrics Issues:** Check Prometheus → `http://prometheus:9090`
- **On-Call:** Alert through Slack #critical channel

