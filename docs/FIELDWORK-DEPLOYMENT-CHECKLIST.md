# Fieldwork API Performance Optimization - Deployment Checklist

**Deployment Date:** __________  
**Deployed By:** __________  
**Environment:** [ ] Dev  [ ] Staging  [ ] Production  

---

## Pre-Deployment (24 hours before)

- [ ] Read full documentation in `docs/AUDIT-FIELDWORK-PERFORMANCE-OPTIMIZATION.md`
- [ ] Review Kong configuration changes:
  ```bash
  git diff backend/kong-gateway/kong/*/kong.yml
  ```
- [ ] Review frontend store changes:
  ```bash
  git diff frontend/stores/audit-fieldwork.ts
  ```
- [ ] Notify team in Slack channel: #auditsphere-deployments
- [ ] Backup current Kong configuration:
  ```bash
  cp backend/kong-gateway/kong/prod/kong.yml \
     backend/kong-gateway/kong/prod/kong.yml.backup.$(date +%s)
  ```
- [ ] Backup current deployment state:
  ```bash
  docker-compose config > deployment-state-$(date +%s).yml
  ```

---

## Backend Deployment (Kong Gateway)

### Step 1: Verify Kong Configuration
- [ ] Check Kong is running: `docker ps | grep kong`
- [ ] Verify current config: `docker exec rb_audit_kong kong config`
- [ ] Test Kong API endpoint:
  ```bash
  curl -i http://localhost:8000/api/v1/fieldwork/interviews \
    -H "Authorization: Bearer YOUR_TEST_TOKEN"
  ```
  Expected: `200 OK` or `401 Unauthorized` (not `502 Bad Gateway`)

### Step 2: Deploy Kong Configuration
- [ ] Copy updated Kong config to container:
  ```bash
  docker cp backend/kong-gateway/kong/prod/kong.yml \
    rb_audit_kong:/etc/kong/kong.yml
  ```
- [ ] Reload Kong configuration:
  ```bash
  docker exec rb_audit_kong kong config load
  ```
  Expected output: `Configuration loaded`

- [ ] Verify plugins are loaded:
  ```bash
  curl http://localhost:8001/plugins | jq '.data[] | {name: .name, config: .config}'
  ```
  Expected to see:
  - `proxy-cache` plugin with `cache_ttl: 300`
  - `rate-limiting` plugin with limits configured

### Step 3: Verify Kong Changes
- [ ] Test rate limiting headers:
  ```bash
  curl -v http://localhost:8000/api/v1/fieldwork/interviews \
    -H "Authorization: Bearer YOUR_TEST_TOKEN" | grep "X-RateLimit"
  ```
  Expected: Headers showing `X-RateLimit-Limit-Minute` and `X-RateLimit-Remaining-Minute`

- [ ] Test cache headers on repeated requests:
  ```bash
  # First request (cache miss)
  curl -v http://localhost:8000/api/v1/fieldwork/interviews?assignmentLetterId=test-1 \
    -H "Authorization: Bearer YOUR_TEST_TOKEN" | grep "X-Cache"
  # Second request (cache hit)
  curl -v http://localhost:8000/api/v1/fieldwork/interviews?assignmentLetterId=test-1 \
    -H "Authorization: Bearer YOUR_TEST_TOKEN" | grep "X-Cache"
  ```
  Expected: First shows `X-Cache: MISS`, second shows `X-Cache: HIT`

- [ ] Monitor Kong logs for errors:
  ```bash
  docker logs -f rb_audit_kong --tail 50
  ```
  Expected: No error messages

### Step 4: Rollback Plan (if needed)
If Kong configuration causes issues:
```bash
# Restore from backup
docker cp deployment-state-*.yml rb_audit_kong:/etc/kong/kong.yml
docker exec rb_audit_kong kong config load
docker-compose restart kong-gateway
```

---

## Frontend Deployment

### Step 1: Build Frontend
- [ ] Install dependencies:
  ```bash
  cd frontend && npm install
  ```
- [ ] Run type checking:
  ```bash
  npm run typecheck
  ```
  Expected: No TypeScript errors in `stores/audit-fieldwork.ts`

- [ ] Run unit tests (if exist):
  ```bash
  npm run test
  ```

- [ ] Build production bundle:
  ```bash
  npm run build
  ```
  Expected: No build errors, `.output/public/` directory created

### Step 2: Test Frontend Changes Locally (optional)
- [ ] Start dev server:
  ```bash
  npm run dev
  ```
- [ ] Navigate to audit fieldwork page
- [ ] Open DevTools → Network tab
- [ ] Select an assignment letter
- [ ] Verify:
  - [ ] Only 5 concurrent GET requests (not doubled)
  - [ ] Response headers show `X-Cache: MISS` first time
  - [ ] Switch away and back to same letter
  - [ ] Response headers show from browser cache (instant)
  - [ ] Edit and save a record
  - [ ] Only 1 POST request (not 2)

### Step 3: Deploy Frontend Bundle
- [ ] Copy build to deployment location:
  ```bash
  docker cp frontend/.output/public/* \
    rb_audit_frontend:/usr/share/nginx/html/
  ```
  OR via CI/CD:
  ```bash
  npm run deploy  # if CI/CD configured
  ```

- [ ] Verify frontend is serving:
  ```bash
  curl -i https://auditsphere.app/
  ```
  Expected: `200 OK` with HTML content

- [ ] Clear browser cache (Cmd+Shift+Delete) for testing

### Step 4: Test Frontend in Production
- [ ] Open production app in incognito window
- [ ] Login to your account
- [ ] Navigate to Audit Fieldwork module
- [ ] Open DevTools → Network tab → filter "fieldwork"
- [ ] Select an assignment letter
- [ ] Verify API call counts:
  - [ ] 5 initial API calls (interviews, observations, documents, samples, test-controls)
  - [ ] No duplicate calls
  - [ ] All requests complete in < 1 second
- [ ] Check response headers:
  - [ ] Look for `X-Cache: MISS` in responses
  - [ ] Look for `X-RateLimit-*` headers
  - [ ] Look for `Cache-Control: public, max-age=300` headers

---

## Backend API Changes (if needed)

### Step 1: Optional Backend Optimization
If you want to implement HTTP caching headers on backend:

- [ ] Update fieldwork route handlers to add cache headers
- [ ] Add database indexes on `assignment_letter_id`:
  ```sql
  CREATE INDEX idx_fieldwork_interviews_assignment_letter_id 
    ON fieldwork_interviews(assignment_letter_id);
  CREATE INDEX idx_fieldwork_observations_assignment_letter_id 
    ON fieldwork_observations(assignment_letter_id);
  CREATE INDEX idx_fieldwork_documents_assignment_letter_id 
    ON fieldwork_documents(assignment_letter_id);
  CREATE INDEX idx_fieldwork_samples_assignment_letter_id 
    ON fieldwork_samples(assignment_letter_id);
  CREATE INDEX idx_fieldwork_test_controls_assignment_letter_id 
    ON fieldwork_test_controls(assignment_letter_id);
  ```

- [ ] Rebuild audit-service:
  ```bash
  docker-compose build audit-service
  docker-compose restart audit-service
  ```

- [ ] Verify startup:
  ```bash
  docker logs -f rb_audit_audit_service --tail 100
  ```
  Expected: No startup errors, service healthy

### Step 2: Database Index Verification
- [ ] Connect to database:
  ```bash
  docker exec -it rb_audit_postgres psql -U postgres -d rb_audit
  ```
- [ ] Check indexes created:
  ```sql
  \d fieldwork_interviews
  \d fieldwork_samples
  ```
  Expected: Indexes listed

---

## Monitoring Setup

### Step 1: Deploy Prometheus Alerts
- [ ] Copy alert rules:
  ```bash
  docker cp monitoring/prometheus-fieldwork-alerts.yml \
    rb_audit_prometheus:/etc/prometheus/rules/fieldwork-alerts.yml
  ```

- [ ] Reload Prometheus:
  ```bash
  docker exec rb_audit_prometheus kill -HUP 1
  ```

- [ ] Verify alerts loaded:
  ```bash
  curl http://localhost:9090/api/v1/rules | jq '.data.groups[] | select(.name=="auditsphere-fieldwork-performance")'
  ```

### Step 2: Test Alert Rules
- [ ] Generate high load to test alerts:
  ```bash
  # Rate limit test
  for i in {1..100}; do
    curl -s http://localhost:8000/api/v1/fieldwork/interviews &
  done
  wait
  ```

- [ ] Check Prometheus for alert firing:
  ```
  http://localhost:9090/alerts
  ```

- [ ] Check Alertmanager (if configured):
  ```
  http://localhost:9093/#/alerts
  ```

### Step 3: Grafana Dashboard (optional)
- [ ] Import dashboard JSON if available
- [ ] Verify metrics visible:
  - [ ] Request rate graph
  - [ ] Cache hit ratio
  - [ ] Response time percentiles
  - [ ] Error rate

---

## Smoke Tests

### Test 1: Basic Fieldwork API
```bash
BEARER_TOKEN="your_token_here"

# Test interviews endpoint
curl -v "http://localhost:8000/api/v1/fieldwork/interviews?assignmentLetterId=Letter-2024-001" \
  -H "Authorization: Bearer $BEARER_TOKEN" | jq '.data | length'

# Test with different letter (should be fresh)
curl -v "http://localhost:8000/api/v1/fieldwork/interviews?assignmentLetterId=Letter-2024-002" \
  -H "Authorization: Bearer $BEARER_TOKEN" | jq '.data | length'

# Test create
curl -X POST "http://localhost:8000/api/v1/fieldwork/interviews" \
  -H "Authorization: Bearer $BEARER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "assignmentLetterId": "Letter-2024-001",
    "interviewee": "Test User",
    "intervieweePosition": "Manager",
    "interviewer": "Auditor",
    "interviewerPosition": "Senior Auditor",
    "date": "2024-10-04",
    "topic": "Internal Controls"
  }' | jq '.data.id'
```
Expected: Valid JSON responses with data

### Test 2: Rate Limiting
```bash
# Generate 70 requests in rapid succession (should hit limit)
for i in {1..70}; do
  curl -s "http://localhost:8000/api/v1/fieldwork/interviews?assignmentLetterId=test-$i" \
    -H "Authorization: Bearer $BEARER_TOKEN" \
    -o /dev/null -w "%{http_code}\n" &
done
wait | sort | uniq -c

# Expected output:
# ~60 requests with 200
# ~10 requests with 429 (rate limited)
```

### Test 3: Cache Behavior
```bash
# First request (cache miss)
time curl -s "http://localhost:8000/api/v1/fieldwork/interviews?assignmentLetterId=cache-test" \
  -H "Authorization: Bearer $BEARER_TOKEN" | wc -l

# Wait 1 second
sleep 1

# Second request (cache hit)  
time curl -s "http://localhost:8000/api/v1/fieldwork/interviews?assignmentLetterId=cache-test" \
  -H "Authorization: Bearer $BEARER_TOKEN" | wc -l

# Cache hit should be significantly faster
```

### Test 4: Frontend Request Deduplication
1. Open browser DevTools → Network tab
2. Filter to "fieldwork" requests only
3. Go to Audit Fieldwork page
4. Rapidly click through different assignment letters
5. Expected: Each assignment letter selection = 5 requests (not 10 or 15)

---

## Performance Baseline

Record baseline metrics before proceeding:

| Metric | Baseline | Target | Actual |
|--------|----------|--------|--------|
| API calls per page load | 5+ | 5 (1st), 0 (2nd) | |
| Page load time (1st) | ~2s | ~1s | |
| Page load time (2nd, cached) | ~2s | ~100ms | |
| Cache hit ratio | 0% | >60% | |
| API calls per minute | >100 | <30 | |
| Rate limit violations | Frequent | None | |
| P95 response time | >500ms | <300ms | |

---

## Rollback Procedure

If issues occur after deployment:

### Step 1: Identify Issue
```bash
# Check Kong logs
docker logs rb_audit_kong | grep -i error | tail -20

# Check audit-service logs
docker logs rb_audit_audit_service | grep -i error | tail -20

# Check performance
curl -w "\nTime: %{time_total}s\n" http://localhost:8000/api/v1/fieldwork/interviews
```

### Step 2: Rollback Kong
```bash
# Restore previous Kong config
docker cp backend/kong-gateway/kong/prod/kong.yml.backup \
  rb_audit_kong:/etc/kong/kong.yml

# Reload
docker exec rb_audit_kong kong config load

# Verify
docker logs rb_audit_kong --tail 20
```

### Step 3: Rollback Frontend
```bash
# If using Docker:
docker-compose restart frontend

# Or redeploy previous version from git:
git checkout HEAD~1 -- frontend/stores/audit-fieldwork.ts
npm run build
npm run deploy
```

### Step 4: Verify Rollback
```bash
# Should see increased API calls again
curl -v http://localhost:8000/api/v1/fieldwork/interviews
```

---

## Post-Deployment (24-48 hours after)

- [ ] Monitor Prometheus alerts for any anomalies
- [ ] Check Grafana dashboard for normal metrics
- [ ] Review Kong logs for errors:
  ```bash
  docker logs rb_audit_kong --since 24h | grep -i error
  ```
- [ ] Review audit-service logs for errors:
  ```bash
  docker logs rb_audit_audit_service --since 24h | grep -i error
  ```
- [ ] Test user feedback - ask team if they notice improvements
- [ ] Document actual performance improvements in DEPLOYMENT_RESULTS.md

---

## Sign-Off

- [ ] All tests passed
- [ ] No errors in logs
- [ ] Performance metrics within targets
- [ ] Team notified of changes
- [ ] Documentation updated
- [ ] Monitoring alerts configured

**Deployment Status:** [ ] Success [ ] Partial [ ] Rolled Back

**Notes:**
```
_________________________________________________________________________

_________________________________________________________________________

_________________________________________________________________________
```

**Sign-Off Date:** __________  
**Verified By:** __________

---

## Support & Troubleshooting

For issues post-deployment:

1. **High API calls continue:** Check browser console for errors
2. **Cache not working:** Verify `X-Cache` header in responses
3. **Rate limiting too strict:** Adjust Kong config and reload
4. **Slow responses:** Check database indexes and slow queries
5. **Memory issues:** Check container resource limits

Contact: #auditsphere-devops Slack channel

