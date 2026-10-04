# AuditSphere Demo - Quick Start Checklist

**Complete setup in ~2-3 hours on a fresh VPS**

---

## Phase 1: Infrastructure (30 minutes)

### 1.1 SSH into VPS
```bash
ssh -i ~/.ssh/demo-key.pem ubuntu@123.45.67.89
sudo su -
```
- [ ] Connected successfully

### 1.2 Install Docker & Dependencies
```bash
curl -fsSL https://get.docker.com | sh
apt-get install -y nginx certbot python3-certbot-nginx git
docker --version && docker compose version
```
- [ ] Docker installed
- [ ] Docker Compose installed
- [ ] Nginx installed

### 1.3 Clone Repository
```bash
mkdir -p /app && cd /app
git clone https://github.com/RauMomo/risk-based-audit.git
cd risk-based-audit
```
- [ ] Repository cloned

---

## Phase 2: DNS & SSL (15 minutes)

### 2.1 Configure DNS
At your DNS provider (Cloudflare, Route 53, etc.):
```
A record:  *.demo.auditsphere.id  →  123.45.67.89
A record:  demo.auditsphere.id    →  123.45.67.89
```
- [ ] DNS records created
- [ ] Propagation confirmed: `nslookup client1.demo.auditsphere.id`

### 2.2 Request SSL Certificate
```bash
sudo certbot certonly --nginx \
  -d demo.auditsphere.id \
  -d "*.demo.auditsphere.id" \
  --agree-tos \
  --email admin@company.com \
  --non-interactive
```
- [ ] SSL certificate issued

### 2.3 Configure Nginx Reverse Proxy
```bash
# Copy the reverse proxy config from guide section 5.2C
sudo nano /etc/nginx/sites-available/auditsphere-demo.conf

# Enable it
sudo ln -s /etc/nginx/sites-available/auditsphere-demo.conf \
           /etc/nginx/sites-enabled/

# Test & reload
sudo nginx -t && sudo systemctl reload nginx
```
- [ ] Nginx config created and enabled
- [ ] Nginx reloaded without errors

---

## Phase 3: Shared Infrastructure (20 minutes)

### 3.1 Create Docker Network
```bash
cd /app/risk-based-audit
docker network create rb_audit_network 2>/dev/null || true
```
- [ ] Docker network created

### 3.2 Start Shared Services
```bash
cd backend
docker compose build
docker compose up -d postgres redis kafka zookeeper
docker compose ps
```
- [ ] All services showing "Up (healthy)"
- [ ] Postgres accepting connections: 
  ```bash
  docker compose exec postgres psql -U postgres -c "SELECT 1;"
  ```
  - [ ] Returns "1"

---

## Phase 4: Tenant 1 Setup (30 minutes)

### 4.1 Generate Tenant Stack
```bash
cd /app/risk-based-audit
chmod +x backend/scripts/generate-tenant-stack.sh
backend/scripts/generate-tenant-stack.sh client1 "Demo Client 1"
```
- [ ] Stack generated in `backend/tenants/client1/`

### 4.2 Onboard Tenant 1
```bash
chmod +x backend/scripts/onboard-tenant.sh
backend/scripts/onboard-tenant.sh client1 "Demo Client 1" 3010 8090 --with-demo-data
```

This script:
- Creates databases
- Runs migrations
- Seeds demo data
- Starts containers
- Takes ~5-10 minutes

- [ ] Script completed without errors
- [ ] All containers running: 
  ```bash
  docker compose -f backend/tenants/client1/docker-compose.yml ps
  ```
  - [ ] Shows all services as "Up"

### 4.3 Verify Client 1
```bash
# Wait 30 seconds for startup
sleep 30

# Test health
curl -I https://api-client1.demo.auditsphere.id/api/v1/auth/health
# Should return: HTTP/2 200

# Test frontend
curl -I https://client1.demo.auditsphere.id
# Should return: HTTP/2 200
```
- [ ] API responding (200)
- [ ] Frontend responding (200)

---

## Phase 5: Additional Tenants (15 minutes each)

### 5.1 Tenant 2
```bash
cd /app/risk-based-audit
backend/scripts/onboard-tenant.sh client2 "Demo Client 2" 3011 8091 --with-demo-data
```
- [ ] Client 2 running
- [ ] API: `curl -I https://api-client2.demo.auditsphere.id/api/v1/auth/health`
- [ ] Frontend: `curl -I https://client2.demo.auditsphere.id`

### 5.2 Tenant 3 (Optional)
```bash
backend/scripts/onboard-tenant.sh client3 "Demo Client 3" 3012 8092 --with-demo-data
```
- [ ] Client 3 running

---

## Phase 6: Demo Readiness (15 minutes)

### 6.1 Pre-Demo Checks
```bash
# Health check all tenants
for tenant in client1 client2; do
  echo "Checking $tenant..."
  curl -s https://api-$tenant.demo.auditsphere.id/api/v1/auth/health
  echo ""
done

# Check databases have data
docker exec rb_audit_postgres psql -U postgres \
  -d rb_audit_audit_client1 \
  -c "SELECT COUNT(*) FROM audit_engagements;" \
  # Should return: > 0

# Check frontend loads
curl -s https://client1.demo.auditsphere.id | head -50
# Should show HTML with <title>AuditSphere</title>
```
- [ ] All APIs responding
- [ ] Demo data present in databases
- [ ] Frontend loads successfully

### 6.2 Credentials Ready
- [ ] Default login: `admin` / `password123`
- [ ] Note: Change these before going to production!

### 6.3 Demo URLs Ready
```
Frontend 1:   https://client1.demo.auditsphere.id
Frontend 2:   https://client2.demo.auditsphere.id
API 1:        https://api-client1.demo.auditsphere.id
API 2:        https://api-client2.demo.auditsphere.id
```
- [ ] All URLs accessible
- [ ] SSL certificates valid (no browser warnings)

---

## Phase 7: Running the Demo

### 7.1 Demo Walkthrough (20 minutes)
1. **Authentication** (2 min)
   - Login: `admin` / `password123`
   - Show dashboard

2. **Audit Planning** (3 min)
   - Show 20 sample audit engagements
   - Show engagement lifecycle

3. **Fieldwork & Findings** (3 min)
   - Show 50+ audit activities
   - Show 200+ findings with risk levels

4. **Risk Module** (2 min)
   - Show risk registers
   - Show 100+ risk items in heat map

5. **Multi-Tenancy** (2 min)
   - Open client2 in another tab
   - Show data isolation

6. **Reports & Analytics** (3 min)
   - Show report generation
   - Show AI-powered scoring

7. **Q&A** (5 min)

### 7.2 Key Talking Points
- **Isolation**: Per-tenant databases, no cross-contamination
- **Speed**: Dashboard < 2s (Redis caching)
- **Scalability**: Add tenants in 15 minutes
- **Compliance**: ISO 27001, GDPR, UU PDP ready
- **Integration**: Google Drive for files, Kafka for events

---

## Troubleshooting

### API Returning 502
```bash
# Check Kong is healthy
docker ps | grep kong-client1

# If down, restart
docker restart kong-client1

# Wait 30 seconds
sleep 30
curl -I https://api-client1.demo.auditsphere.id/api/v1/auth/health
```

### Frontend Not Loading
```bash
# Check frontend container
docker ps | grep frontend-client1

# Check Nginx forwarding
sudo curl -I http://127.0.0.1:3010

# Check Nginx config
sudo nginx -t
```

### Database Connection Error
```bash
# Restart PostgreSQL
docker restart rb_audit_postgres

# Verify it's healthy
docker exec rb_audit_postgres psql -U postgres -c "SELECT 1;"

# Restart services
docker compose -f backend/tenants/client1/docker-compose.yml restart
```

### Slow Performance
```bash
# Check resource usage
docker stats

# Reduce number of tenants if needed
docker compose -f backend/tenants/client3/docker-compose.yml down

# Clear cache
docker exec rb_audit_redis redis-cli FLUSHALL
```

---

## Post-Demo Cleanup

### Option A: Keep Running (for extended POC)
```bash
# Daily maintenance
docker system prune -a

# Check logs
docker logs rb_audit_postgres --tail=50

# Monitor performance
watch docker stats
```

### Option B: Stop Demo (preserve data)
```bash
docker compose -f backend/tenants/client1/docker-compose.yml down
docker compose -f backend/tenants/client2/docker-compose.yml down

# Shared services remain
docker compose ps
```

### Option C: Full Reset (start over)
```bash
# Stop everything
docker compose down
for tenant in client1 client2 client3; do
  docker compose -f backend/tenants/$tenant/docker-compose.yml down 2>/dev/null
done

# Delete databases
docker exec rb_audit_postgres psql -U postgres <<'EOF'
DROP DATABASE IF EXISTS rb_audit_auth_client1;
DROP DATABASE IF EXISTS rb_audit_audit_client1;
DROP DATABASE IF EXISTS rb_audit_master_client1;
DROP DATABASE IF EXISTS rb_audit_risk_client1;
EOF

# Start fresh
cd /app/risk-based-audit/backend
docker compose up -d postgres redis kafka zookeeper
```

---

## Success Indicators

You've successfully set up AuditSphere demo when:

- [ ] `docker ps` shows all services as "Up"
- [ ] `curl https://api-client1.demo.auditsphere.id/api/v1/auth/health` returns 200
- [ ] Frontend loads at `https://client1.demo.auditsphere.id` in < 3 seconds
- [ ] Can login with `admin` / `password123`
- [ ] Dashboard shows audit KPIs (total audits, findings count)
- [ ] Sample data is visible (20+ audit engagements, 200+ findings)
- [ ] Multi-tenant demo works (client2 has different data than client1)
- [ ] File uploads work (save to Google Drive or local fallback)

---

## Support

| Issue | Command |
|-------|---------|
| **Check all containers** | `docker ps -a` |
| **View container logs** | `docker logs <container-name>` |
| **Access PostgreSQL** | `docker exec -it rb_audit_postgres psql -U postgres` |
| **Restart a service** | `docker restart <container-name>` |
| **Check resource usage** | `docker stats` |
| **Validate Nginx config** | `sudo nginx -t` |
| **Test API health** | `curl https://api-client1.demo.auditsphere.id/api/v1/auth/health` |

---

**Estimated Total Setup Time: 2-3 hours** (including DNS propagation wait)

**Next Step:** Run the demo walkthrough from Phase 7.1!
