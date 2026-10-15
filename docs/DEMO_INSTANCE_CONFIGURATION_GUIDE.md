# AuditSphere Demo - Instance Configuration & Hardware Guide

**Choose the right infrastructure for your demo scenario**

---

## 1. Quick Selection Matrix

| Scenario | Tenants | Users/Tenant | Recommendation | Cost | Setup Time |
|----------|---------|--------------|-----------------|------|------------|
| **Single Demo** | 1 | 5-10 | Laptop/Local Docker | $0 | 30 min |
| **Sales POC** | 1-2 | 10-20 | Cloud Starter | $30-50/mo | 1 hour |
| **Extended POC** | 2-3 | 20-50 | Cloud Standard | $50-100/mo | 1.5 hours |
| **Multi-Client Demo** | 5+ | 50+ | Cloud Pro | $200-500/mo | 2-3 hours |
| **Internal Training** | 3-5 | 100+ | On-Prem | Variable | 2-4 hours |
| **Load Testing** | 10+ | 1000+ | Cloud Enterprise | $500-2000/mo | 4+ hours |

---

## 2. Scenario-Based Hardware Recommendations

### Scenario A: Local Development Machine

**Best For:**
- Internal team demos
- Quick feature showcase
- Development/testing
- Training sessions

**Hardware Requirements:**
```
CPU:        4 cores (8 virtual)
RAM:        8 GB minimum, 16 GB recommended
Storage:    50-100 GB SSD available
Network:    WiFi or Ethernet
OS:         macOS 12+, Ubuntu 20.04+, Windows 10+
```

**Tools to Install:**
```bash
# macOS
brew install docker docker-compose git nginx certbot

# Ubuntu/Debian
sudo apt-get install docker.io docker-compose git nginx certbot

# Windows
# Download: Docker Desktop, Git Bash, WSL2
```

**Setup Steps:**
```bash
# Clone repository
git clone https://github.com/RauMomo/risk-based-audit.git
cd risk-based-audit

# Build shared services
cd backend
docker compose -f docker-compose.yml build

# Generate tenant
../scripts/generate-tenant-stack.sh demo "Internal Demo"

# Onboard with local settings
../scripts/onboard-tenant.sh demo "Internal Demo" 3010 8090 \
  --local --with-demo-data

# Access at http://demo.localhost:3010
```

**Pros:**
- ✅ No cloud costs
- ✅ Fully offline capable
- ✅ Instant spin-up
- ✅ Full control

**Cons:**
- ❌ Not accessible externally
- ❌ Limited to local network
- ❌ Requires good laptop
- ❌ Can be slow with multiple tenants

**Testing Checklist:**
- [ ] Docker daemon running
- [ ] All services healthy: `docker ps`
- [ ] Frontend loads: `http://demo.localhost:3010`
- [ ] Login works: `admin` / `password123`

---

### Scenario B: Cloud Starter (Sales POC)

**Best For:**
- Initial customer POC
- 1-2 week demos
- Limited concurrent users
- Cost-sensitive deployments

**Provider Recommendations:**

#### DigitalOcean (Simplest)
```
Droplet:     2GB RAM / 1 vCPU
Storage:     50 GB SSD
Bandwidth:   2 TB/month
Price:       $12-15/month
Region:      Closest to customer (Singapore, Tokyo, etc.)
```

**Setup:**
```bash
# 1. Create Droplet
# - Image: Ubuntu 22.04 LTS
# - Region: Singapore/Tokyo
# - Auth: SSH Key (not password)

# 2. Get IP address from DigitalOcean console
# Example: 192.0.2.100

# 3. SSH in
ssh -i ~/.ssh/do-key root@192.0.2.100

# 4. Follow main guide section 5 (Deployment Steps)
```

#### AWS (More Control)
```
Instance:    t3.small (2 vCPU, 2 GB RAM)
Storage:     30 GB gp3 SSD
Region:      ap-southeast-1 (Singapore)
Price:       $20-30/month
Network:     Elastic IP (static)
```

**Setup Differences:**
```bash
# Security group (firewall) must allow:
# - Port 22 (SSH)
# - Port 80 (HTTP)
# - Port 443 (HTTPS)
# - Port 3000-3020 (frontend range)
# - Port 8000-8100 (API range)

# Use Elastic IP for stable address
# DNS → Elastic IP
```

#### Linode (Best Balance)
```
Nanode 1GB:  1 vCPU, 1 GB RAM (minimum, not recommended)
Linode 2GB:  1 vCPU, 2 GB RAM
Linode 4GB:  2 vCPU, 4 GB RAM (RECOMMENDED)
Storage:     80 GB SSD
Price:       $6-20/month
Region:      Singapore, Tokyo
```

**Setup:**
```bash
# Similar to AWS but simpler networking
# Linode's dashboard is cleaner than AWS
```

**Network Configuration:**
```
Domain:              demo.auditsphere.id
Registrar:           Namecheap or Cloudflare
DNS Config:          A record → IP address
Subdomain Pattern:   *.demo.auditsphere.id → IP
SSL Certificate:     Let's Encrypt (free, auto-renew)
```

**Expected Costs:**
```
VPS/Droplet:         $12-20/month
Domain:              $10-15/year
Bandwidth (excess):  $0-5/month
Total:               ~$15-25/month
```

**Performance Expectations:**
- Dashboard load: 2-3 seconds (initial), < 1s (cached)
- API response: 200-500ms
- Concurrent users: 10-20
- Concurrent tenants: 1-2

**Testing Checklist:**
- [ ] SSH access working
- [ ] DNS resolution correct: `nslookup demo.auditsphere.id`
- [ ] SSL certificate valid: `curl -v https://demo.auditsphere.id`
- [ ] All services healthy
- [ ] Frontend and API accessible externally
- [ ] Performance acceptable for demo

---

### Scenario C: Cloud Standard (Extended POC)

**Best For:**
- 2-4 week customer POC
- Multiple concurrent demos
- 20-50 daily active users
- Better performance required

**Provider: AWS or DigitalOcean**

#### DigitalOcean
```
Droplet:     4GB RAM / 2 vCPU
Storage:     80 GB SSD
Bandwidth:   4 TB/month
Price:       $24-30/month
Regions:     Singapore, Tokyo (App Platform for managed databases optional)
```

#### AWS
```
Instance:    t3.medium (2 vCPU, 4 GB RAM)
Storage:     80 GB gp3 SSD
Bandwidth:   Included in region
Price:       $30-45/month
```

#### Hetzner (Best Value)
```
CX31:        2 vCPU, 4 GB RAM
Storage:     80 GB SSD
Bandwidth:   20 TB/month
Price:       €8-10/month (~$9-11)
Region:      EU, Asia
```

**Network Setup:**
```bash
# Multiple tenants on one VPS
# Client 1: 3010 (frontend) / 8090 (API)
# Client 2: 3011 (frontend) / 8091 (API)
# Client 3: 3012 (frontend) / 8092 (API)
# DNS:      *.demo.auditsphere.id → VPS IP
# SSL:      Wildcard cert for *.demo.auditsphere.id
```

**Database Optimization:**
```yaml
# PostgreSQL tuning for 2-3 tenants
shared_buffers: 1GB          # 25% of RAM
effective_cache_size: 3GB    # 75% of RAM
maintenance_work_mem: 256MB
work_mem: 32MB
```

**Expected Costs:**
```
VPS:         $25-30/month
Domain:      $10-15/year
Backups:     $5-10/month
Monitoring:  $0-5/month
Total:       ~$40-50/month
```

**Performance Expectations:**
- Dashboard: 1-2 seconds
- API response: 100-300ms
- Concurrent users: 30-50
- Concurrent tenants: 2-3

---

### Scenario D: Cloud Pro (Multi-Client Demo)

**Best For:**
- 1-3 month customer POCs
- 5+ concurrent clients
- Heavy load testing
- Production-like environment

**Recommended Setup:**

#### Primary VPS (App Server)
```
Provider:    AWS t3.large or DigitalOcean 8GB
CPU:         4 vCPU
RAM:         8 GB
Storage:     200 GB SSD
Price:       $60-100/month
```

#### Optional: Separate Database VPS (Recommended)
```
Provider:    AWS t3.medium or DigitalOcean 4GB
CPU:         2 vCPU
RAM:         4 GB
Purpose:     Dedicated PostgreSQL (lower app-tier contention)
Price:       $30-50/month
Network:     Private network, app-tier connects via private IP
```

**Architecture:**
```
                    ┌─────────────────────────┐
                    │   CDN / DDoS Protection │
                    │   (Cloudflare, Akamai)  │
                    └────────────┬────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │                         │
          ┌─────────▼──────────┐    ┌───────▼────────┐
          │  App VPS (Nginx)   │    │  LB / Failover │
          │ - Kong Gateway     │    │  (Optional)     │
          │ - Frontends        │    │                 │
          │ - Microservices    │    │                 │
          └────────────┬───────┘    └────────────────┘
                       │
          ┌────────────┴────────────┐
          │                         │
    ┌─────▼──────┐          ┌─────▼──────┐
    │   Cache    │          │  Database  │
    │   (Redis)  │          │(PostgreSQL)│
    │            │          │            │
    │ (On App)   │          │(Separate)  │
    └────────────┘          └────────────┘
```

**Database Optimization:**
```yaml
# PostgreSQL for heavy load
shared_buffers: 2GB
effective_cache_size: 6GB
maintenance_work_mem: 512MB
work_mem: 64MB
max_connections: 200
```

**Multi-Tenant Configuration:**
```
Tenant 1: 3010 / 8090
Tenant 2: 3011 / 8091
Tenant 3: 3012 / 8092
Tenant 4: 3013 / 8093
Tenant 5: 3014 / 8094
Tenant 6: 3015 / 8095

DNS Wildcard: *.demo.auditsphere.id → App VPS IP
```

**Backup & Recovery:**
```bash
# Daily automated backups
*/2 * * * * /opt/scripts/backup-demo.sh

# S3 backup (AWS)
aws s3 sync /backups s3://auditsphere-backups/demo/

# Point-in-time recovery available
```

**Monitoring & Alerting:**
```
Tools:
- Prometheus (metrics)
- Grafana (dashboards)
- ELK Stack (logs)
- PagerDuty (alerts)

Metrics to Monitor:
- CPU usage > 70%
- Memory usage > 80%
- Disk usage > 85%
- API latency > 1s
- Error rate > 1%
```

**Expected Costs:**
```
App VPS:         $60-80/month
Database VPS:    $30-50/month (optional)
Domain:          $10-15/year
Backups (S3):    $5-20/month
Monitoring:      $10-50/month
CDN (optional):  $20-100/month
Total:           ~$135-315/month
```

**Performance Expectations:**
- Dashboard: < 1 second
- API response: 50-200ms
- Concurrent users: 100-200
- Concurrent tenants: 5+
- 99.5% uptime SLA

---

### Scenario E: On-Premises (Internal Training)

**Best For:**
- Internal team training
- High security requirements
- No internet dependency
- Permanent training lab

**Hardware (Bare Metal or VM):**
```
CPU:        8-16 cores
RAM:        16-32 GB
Storage:    500 GB SSD
Network:    Dedicated 10 Mbps uplink
OS:         Ubuntu Server 22.04 LTS
```

**Networking (Internal):**
```
Domain:     demo.auditsphere.local (or demo.<company>.local)
DNS:        Corporate DNS server
SSL:        Self-signed or corporate CA
Access:     VPN or internal network only
```

**Setup:**
```bash
# 1. Install on bare metal or VM
# Follow main deployment guide (section 5)

# 2. Configure internal DNS
# Add A records to corporate DNS:
#   demo.auditsphere.local          → IP
#   *.demo.auditsphere.local        → IP

# 3. Generate self-signed SSL (or corporate cert)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/nginx/ssl/demo.key \
  -out /etc/nginx/ssl/demo.crt

# 4. Configure Nginx for internal IPs only
# In /etc/nginx/sites-available/auditsphere-demo.conf:
#   server { listen 192.168.1.100:443 ssl http2; ... }
#   NOT: listen 0.0.0.0:443 (external exposure)

# 5. Setup VPN for remote access (optional)
#   - OpenVPN
#   - WireGuard
```

**Access Control:**
```
Network Firewall:
  - Outbound only (training doesn't need internet)
  - Internal network access
  - VPN access for remote (optional)

Ports Allowed:
  - 22 (SSH) - admin only
  - 80 (HTTP) - redirects to 443
  - 443 (HTTPS) - internal subnet
  - 5432 (PostgreSQL) - localhost only
  - 6379 (Redis) - localhost only
```

**Expected Costs:**
```
Hardware:      $2,000-5,000 (one-time)
Networking:    $500-2,000/year
Maintenance:   $0-2,000/year
Total:         ~$2,500-7,000 initial + $500-2,000/year
```

**Performance Expectations:**
- Dashboard: < 1 second
- API response: 50-100ms
- Concurrent users: 200+
- Concurrent tenants: 10+
- No cloud latency

---

### Scenario F: Enterprise / Load Testing

**Best For:**
- Production-like staging
- 1000+ concurrent users
- 10+ simultaneous tenants
- High availability required

**Architecture:**
```
                    ┌─────────────────────┐
                    │  DDoS Protection    │
                    │  (Cloudflare)       │
                    └──────────┬──────────┘
                               │
                    ┌──────────┴──────────┐
                    │                     │
              ┌─────▼─────┐        ┌─────▼─────┐
              │ LB 1      │        │ LB 2      │
              │ (Active)  │        │ (Standby) │
              └─────┬─────┘        └─────┬─────┘
                    │                    │
        ┌───────────┴────────────┬──────┴──────────┐
        │                        │                  │
    ┌───▼──┐              ┌──────▼────┐       ┌─────▼───┐
    │App 1 │              │  App 2    │       │ App 3   │
    │      │              │           │       │         │
    │ (4vCPU│              │ (4vCPU    │       │(4vCPU   │
    │ 8GB)  │              │  8GB)     │       │ 8GB)    │
    └───┬──┘              └──────┬────┘       └─────┬───┘
        │                        │                  │
        └────────────┬───────────┴──────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
    ┌───▼──────┐          ┌──────▼─────┐
    │PostgreSQL│          │ Redis      │
    │Cluster   │          │ Cluster    │
    │(Primary +│          │(3 nodes)   │
    │ Replica) │          │            │
    └──────────┘          └────────────┘
```

**Infrastructure Costs:**
```
App Tier (3x):           $150-300/month
Load Balancer:           $50-100/month
Database (HA):           $200-400/month
Cache (Cluster):         $100-200/month
CDN/DDoS:                $50-300/month
Monitoring/Logging:      $100-500/month
Backup (S3):             $20-100/month
Total:                   ~$700-2,000/month
```

**Performance SLA:**
```
API Response:            < 100ms (p95)
Dashboard Load:          < 500ms (p95)
Concurrent Users:        1,000+
Concurrent Tenants:      10-20
Uptime Target:           99.9%
```

---

## 3. Detailed Hardware Specifications

### CPU Selection

**Important Metrics:**
```
vCPU Count:    Number of virtual processors
ECU Rating:    Compute power (AWS metric)
Turbo Boost:   Speed boost under load

PostgreSQL needs:
  - Multi-threaded (parallel queries)
  - Good single-core performance
  - Minimal context switching

Recommendation:
  - Avoid burst-only instances (t2, t3 without unlimited)
  - Use m5, m6i, c5 for stable performance
  - For cloud: on-demand (not spot) for reliability
```

**Processor Comparison:**
```
AWS t3.small    │ 2 vCPU, 2GB RAM   │ Burst (not ideal for 24/7)
AWS t3.medium   │ 2 vCPU, 4GB RAM   │ Burst (better)
AWS m5.large    │ 2 vCPU, 8GB RAM   │ Stable (recommended)

DigitalOcean 2GB  │ 1 vCPU, 2GB RAM │ Entry-level
DigitalOcean 4GB  │ 2 vCPU, 4GB RAM │ Recommended
DigitalOcean 8GB  │ 4 vCPU, 8GB RAM │ Good performance

Hetzner CX31    │ 2 vCPU, 4GB RAM  │ Best value
Hetzner CX41    │ 4 vCPU, 8GB RAM  │ High performance
```

### RAM Allocation

**Docker Memory Breakdown (4GB instance):**
```
Available:              4 GB
┌─────────────────────────┐
│ PostgreSQL: 1.5 GB     │  (shared_buffers=512MB, effective_cache=1.5GB)
│ Redis: 256 MB          │  (maxmemory=256MB)
│ Kafka: 512 MB          │  (xms=256m, xmx=512m)
│ Kong: 256 MB           │  (per instance)
│ Microservices: 512 MB  │  (auth, audit, master, risk, analytics)
│ Nginx: 128 MB          │  (lightweight)
│ System/Overhead: 384 MB│
└─────────────────────────┘
```

**Memory Tuning:**
```bash
# For 4 GB instance
docker run -e POSTGRES_SHARED_BUFFERS=512MB \
           -e POSTGRES_EFFECTIVE_CACHE_SIZE=1500MB \
           postgres:16

# For 8 GB instance
POSTGRES_SHARED_BUFFERS=2GB
POSTGRES_EFFECTIVE_CACHE_SIZE=6GB

# For 16 GB instance
POSTGRES_SHARED_BUFFERS=4GB
POSTGRES_EFFECTIVE_CACHE_SIZE=12GB
```

### Storage Recommendations

**SSD vs HDD:**
```
SSD (RECOMMENDED for demo):
  - Fast I/O (crucial for database)
  - Lower latency
  - Better for random access
  - Cost: ~$0.10/GB/month

HDD (NOT recommended):
  - Slow for database workloads
  - 5-10x slower for random access
  - Only use for backups
  - Cost: ~$0.02/GB/month
```

**Storage Size by Scenario:**
```
1 Tenant + Demo Data:     30-50 GB
2-3 Tenants:              80-120 GB
5+ Tenants:               200-300 GB
+Backups (weekly):        +100-500 GB
+Log Aggregation:         +50-200 GB
```

**Disk I/O Monitoring:**
```bash
# Check IOPS usage
iostat -x 1

# Monitor disk utilization
df -h
du -sh /var/lib/docker/volumes/*

# Identify large files
du -sh /* | sort -hr | head -10
```

---

## 4. Network Configuration Patterns

### Pattern 1: Public Domain with Wildcard DNS

**Best For:** Cloud deployments, accessible demos

```
Domain:              demo.auditsphere.id
DNS Provider:        Cloudflare / Route 53
SSL:                 Let's Encrypt (free, auto-renew)
Subdomains:          *.demo.auditsphere.id
Access:              Internet accessible

DNS Configuration:
┌─────────────────────────────────┐
│ Type  │ Name  │ Value            │
├─────────────────────────────────┤
│ A     │ @     │ 123.45.67.89     │
│ A     │ *     │ 123.45.67.89     │
│ CNAME │ api   │ demo.auditsphere │
└─────────────────────────────────┘

Nginx Reverse Proxy:
  client1.demo.auditsphere.id  →  127.0.0.1:3010 (frontend)
  api-client1.demo.auditsphere.id  →  127.0.0.1:8090 (Kong)
```

**Certbot SSL Setup:**
```bash
sudo certbot certonly --nginx \
  -d demo.auditsphere.id \
  -d '*.demo.auditsphere.id' \
  --preferred-challenges dns \
  --agree-tos \
  --email admin@company.com \
  --non-interactive

# Auto-renewal every 60 days
# Logs to: /etc/letsencrypt/
```

### Pattern 2: Local Hostname with Self-Signed SSL

**Best For:** Internal networks, training labs

```
Domain:              demo.local (or demo.auditsphere.local)
DNS:                 /etc/hosts or corporate DNS
SSL:                 Self-signed or corporate CA
Access:              Internal network + VPN

/etc/hosts Configuration:
┌──────────────────────────────────────┐
│ 192.168.1.100  demo.local            │
│ 192.168.1.100  client1.demo.local    │
│ 192.168.1.100  api-client1.demo.local│
└──────────────────────────────────────┘

Self-Signed Certificate:
openssl req -x509 -nodes -days 365 \
  -newkey rsa:2048 \
  -keyout /etc/nginx/ssl/demo.key \
  -out /etc/nginx/ssl/demo.crt

# Subjects:
#   Common Name: *.demo.local
#   Alt Names: demo.local, *.demo.local
```

### Pattern 3: IP-Based Access (Minimal Setup)

**Best For:** Quick internal testing

```
Access:              http://123.45.67.89:3010 (no HTTPS)
                     http://123.45.67.89:8090 (no HTTPS)

Nginx (Optional):
  - Not required
  - Services exposed directly via ports
  - Good for load testing APIs
```

**Disadvantages:**
- ❌ Not HTTPS (data not encrypted)
- ❌ Not production-like
- ❌ Difficult with multiple tenants (need unique IPs)
- ❌ Browser warnings about SSL

---

## 5. Scaling Checklist

### When to Scale Up

**Vertical Scaling (bigger single server):**
- CPU utilization consistently > 70%
- Memory utilization > 80%
- Disk I/O waiting > 20%

**Action:**
```bash
# Upgrade to larger instance size
# For AWS t3.medium → t3.large
# For DO 4GB → 8GB droplet
```

**Horizontal Scaling (multiple servers):**
- Need to add 5+ more tenants
- Concurrent users > 200
- Multiple demo regions required

**Action:**
```bash
# Set up load balancing
# Deploy database replication
# Use CDN for static assets
# Separate app and database tiers
```

### Monitoring Thresholds

```
CPU Usage:
  Green:    0-60%
  Yellow:   60-80%    → Monitor closely
  Red:      80%+      → Scale immediately

Memory Usage:
  Green:    0-70%
  Yellow:   70-85%    → Monitor closely
  Red:      85%+      → Scale or reduce tenants

Disk Usage:
  Green:    0-70%
  Yellow:   70-85%    → Clean old logs
  Red:      85%+      → Clean or expand

API Response Time:
  Green:    < 200ms
  Yellow:   200-500ms
  Red:      500ms+    → Investigate

Error Rate:
  Green:    < 0.1%
  Yellow:   0.1-1%
  Red:      1%+
```

---

## 6. Cost Optimization Tips

### Reduce Costs

1. **Use Reserved Instances (If 3+ months)**
   ```
   AWS:       30-40% discount (1-year commitments)
   DO:        No reservation, but stable pricing
   Hetzner:   Already 50% cheaper than AWS
   ```

2. **Remove Unused Tenants**
   ```bash
   # Stop tenant to save resources
   docker compose -f backend/tenants/client3/docker-compose.yml down
   
   # Saves ~300-500MB RAM per tenant
   ```

3. **Auto-Scale Database**
   ```bash
   # Close idle connections
   docker exec rb_audit_postgres \
     psql -U postgres -c "SELECT pg_terminate_backend(pid) \
       FROM pg_stat_activity WHERE datname='rb_audit_audit_client1' \
       AND state='idle';"
   ```

4. **Optimize Backups**
   ```bash
   # Incremental backups (not full)
   # Store to cheaper S3 regions
   # Archive old backups to Glacier
   ```

### Compare Providers (Monthly Costs)

```
Scenario: 2-3 Tenants, 4GB RAM

AWS:
  t3.medium: $30
  Storage:   $5
  Data Out:  $5
  Total:     ~$40/month

DigitalOcean:
  4GB Droplet: $24
  Storage:     $2
  Bandwidth:   $0 (first 2TB free)
  Total:       ~$26/month

Hetzner:
  CX31: €8
  Storage: €1
  Total: ~€9 (~$10/month)

Linode:
  4GB: $20
  Storage: $2
  Total: ~$22/month
```

---

## 7. Disaster Recovery & High Availability

### Backup Strategy

**Frequency:**
```
Demo (no critical data):    Weekly
POC (customer evaluation):  Daily
Extended POC (production):  Multiple daily
```

**Backup Script:**
```bash
#!/bin/bash
# /opt/scripts/backup-demo.sh

BACKUP_DIR="/backups/demo-$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# PostgreSQL full backup
docker exec rb_audit_postgres pg_dumpall -U postgres | \
  gzip > $BACKUP_DIR/postgres-full.sql.gz

# Tenant data (selected)
for tenant in client1 client2; do
  docker exec rb_audit_postgres pg_dump -U postgres \
    rb_audit_audit_$tenant | \
    gzip > $BACKUP_DIR/audit-$tenant.sql.gz
done

# Upload to S3
aws s3 sync $BACKUP_DIR s3://demo-backups/
```

### Point-in-Time Recovery

```bash
# 1. Stop affected container
docker stop rb_audit_audit_service

# 2. Restore from backup
docker exec rb_audit_postgres \
  psql -U postgres rb_audit_audit_client1 < backup.sql

# 3. Restart services
docker restart rb_audit_audit_service
```

### High Availability Setup (Advanced)

```
┌─────────────────────────────────────┐
│  Load Balancer (AWS ELB or Nginx)   │
└──────────────┬──────────────────────┘
               │
      ┌────────┴────────┐
      │                 │
  ┌───▼──┐          ┌───▼──┐
  │ App1 │          │ App2 │
  │ (active)       │ (warm)
  └───┬──┘          └───┬──┘
      │                 │
  ┌───┴─────────────────┴───┐
  │    Shared Database      │
  │    (replication)        │
  └─────────────────────────┘
```

---

## Quick Decision Tree

```
START
  │
  ├─→ [Internal Team Demo?]
  │     └─→ YES: Local Machine (Scenario A)
  │     └─→ NO: Continue
  │
  ├─→ [Budget < $30/month?]
  │     └─→ YES: DigitalOcean 2GB or Hetzner (Scenario B)
  │     └─→ NO: Continue
  │
  ├─→ [Customer POC (1-4 weeks)?]
  │     └─→ YES: Cloud Standard 4GB (Scenario C)
  │     └─→ NO: Continue
  │
  ├─→ [Multiple Concurrent Demos?]
  │     └─→ YES: Cloud Pro 8GB+ (Scenario D)
  │     └─→ NO: Continue
  │
  ├─→ [High Security / On-Prem Required?]
  │     └─→ YES: Bare Metal (Scenario E)
  │     └─→ NO: Continue
  │
  └─→ DEFAULT: Cloud Standard 4GB (most versatile)
```

---

**Next Steps:**

1. Choose your scenario from Section 1
2. Select a provider (DigitalOcean recommended for simplicity)
3. Follow the Quick Start Checklist (DEMO_QUICK_START_CHECKLIST.md)
4. Run the main setup guide (AUDITSPHERE_DEMO_SETUP_GUIDE.md)

---
