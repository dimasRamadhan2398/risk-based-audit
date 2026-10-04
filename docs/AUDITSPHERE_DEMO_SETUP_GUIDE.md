# AuditSphere Demo Environment Setup Guide

**Version:** 1.0  
**Last Updated:** 2026-10-04  
**Target Audience:** DevOps engineers, Solution architects, Demo coordinators (beginner-friendly)

---

## 📋 Table of Contents

1. [What is a Demo Environment?](#1-what-is-a-demo-environment)
2. [Demo Architecture Overview](#2-demo-architecture-overview)
3. [Instance Configuration & Hardware Specs](#3-instance-configuration--hardware-specs)
4. [Pre-Setup Checklist](#4-pre-setup-checklist)
5. [Step-by-Step Deployment](#5-step-by-step-deployment)
6. [Multi-Tenant Demo Setup](#6-multi-tenant-demo-setup)
7. [Demo Data Seeding](#7-demo-data-seeding)
8. [Access & Credentials](#8-access--credentials)
9. [Running a Successful Demo](#9-running-a-successful-demo)
10. [Troubleshooting Common Issues](#10-troubleshooting-common-issues)
11. [Cleanup & Maintenance](#11-cleanup--maintenance)

---

## 1. What is a Demo Environment?

A **Demo Environment** is a fully functional but isolated instance of AuditSphere configured to showcase features to prospects, partners, or internal stakeholders without affecting production data or infrastructure.

### Key Characteristics:
- **Isolated**: Runs on its own VPS or local machine, separate from production
- **Pre-populated**: Contains realistic sample data (audit charters, findings, etc.)
- **Fast Deployment**: Can be spun up in hours, not days
- **Multi-tenant Ready**: Can host multiple demo client instances simultaneously
- **Reproducible**: Same setup every time using Docker & scripts
- **Resettable**: Can be wiped and re-provisioned without side effects

### Demo vs. Production

| Aspect | Demo | Production |
|--------|------|-----------|
| **Purpose** | Feature showcase, testing | Live business data |
| **Data Sensitivity** | Sample data only | Real client audit findings |
| **SLA Requirements** | Flexible | 99.9% uptime |
| **Cost Optimization** | Single small VPS (2-4 CPU, 4-8GB RAM) | Multi-instance, auto-scaling |
| **Backup Policy** | Optional (recreate from scratch) | Daily automated backups |
| **SSL Certificates** | Self-signed OK for local | Let's Encrypt production certs |
| **Scaling** | Vertical only | Horizontal + Vertical |

---

## 2. Demo Architecture Overview

AuditSphere is a **microservices-based system** with the following architecture:

```
                           ┌─────────────────────────────┐
                           │   Demo Environment VPS      │
                           │   (Single Linux Server)      │
                           └──────────────┬──────────────┘
                                          │
                    ┌─────────────────────┼────────────────────┐
                    │                     │                    │
              ┌─────▼─────┐        ┌──────▼──────┐      ┌──────▼──────┐
              │   Nginx    │        │   Nginx     │      │   Nginx     │
              │ (Reverse   │        │ (Reverse    │      │ (Reverse    │
              │  Proxy)    │        │  Proxy)     │      │  Proxy)     │
              └─────┬─────┘        └──────┬──────┘      └──────┬──────┘
                    │                     │                    │
          ┌─────────▼─────────┐   ┌──────▼──────┐      ┌──────▼──────────────┐
          │ Demo Client 1     │   │Demo Client 2│      │Docker Network (Bridge)
          │ (Frontend: :3010) │   │(:3010)      │      │
          │ (Kong API: :8090) │   │             │      ├─ Redis
          └─────────┬─────────┘   └──────┬──────┘      ├─ PostgreSQL
                    │                     │            ├─ Kafka
                    │                     │            ├─ Zookeeper
          ┌─────────▼─────────────────────▼─────────┐  ├─ Python AI
          │                                         │  └─ (Shared Services)
          │    Docker Compose Containers            │
          │                                         │
          │  ┌──────────┐  ┌──────────┐            │
          │  │ Auth Svc │  │ Audit Svc│            │
          │  └──────────┘  └──────────┘            │
          │                                         │
          │  ┌──────────┐  ┌──────────┐            │
          │  │Master Svc│  │ Risk Svc │            │
          │  └──────────┘  └──────────┘            │
          │                                         │
          │  ┌──────────┐  ┌──────────┐            │
          │  │Analytics │  │Kong API  │            │
          │  │  Service │  │Gateway   │            │
          │  └──────────┘  └──────────┘            │
          └─────────────────────────────────────────┘
```

### Components

| Component | Role | Technology |
|-----------|------|-----------|
| **Frontend** | Web UI for auditors | Nuxt 4, Vue.js, Tailwind CSS |
| **Kong API Gateway** | Routes API requests, CORS | Kong 3.4 |
| **Auth Service** | User authentication, RBAC | Go, PostgreSQL |
| **Audit Service** | Audit data & file management | Go, PostgreSQL |
| **Master Service** | Master data (clients, countries) | Go, PostgreSQL |
| **Risk Service** | Risk matrices, appetites | Go, PostgreSQL |
| **Analytics Service** | Risk scoring | Go + Python AI |
| **Database** | Data persistence | PostgreSQL 16 |
| **Cache Layer** | Sessions, rate limiting | Redis 7 |
| **Message Queue** | Event streaming | Kafka 7.6 |
| **Reverse Proxy** | HTTPS, domain routing | Nginx |

---

## 3. Instance Configuration & Hardware Specs

### Recommended Demo VPS Specifications

#### Minimum (1-2 clients, training only)
```
CPU:        2 vCPU (e.g., AWS t3.small, DigitalOcean 2GB)
RAM:        4 GB
Storage:    30 GB SSD
Bandwidth:  2 Mbps uplink
OS:         Ubuntu 22.04 LTS or 20.04 LTS
Network:    Public IPv4 with static IP
```

#### Recommended (3-5 clients, daily demos)
```
CPU:        4 vCPU (e.g., AWS t3.medium, DigitalOcean 2GB plan)
RAM:        8 GB
Storage:    80 GB SSD
Bandwidth:  5 Mbps uplink
OS:         Ubuntu 22.04 LTS
Network:    Public IPv4 with static IP + Floating IP for HA
```

#### Enterprise Demo (10+ concurrent clients, heavy load)
```
CPU:        8 vCPU (AWS t3.large or equivalent)
RAM:        16 GB
Storage:    256 GB SSD
Bandwidth:  20 Mbps uplink
OS:         Ubuntu 22.04 LTS
Network:    Public IPv4 + Private subnet for databases
Monitoring: CloudWatch, Prometheus, Grafana
Backup:     Daily snapshots to object storage
```

### VPS Provider Recommendations

| Provider | Instance Type | Monthly Cost | Notes |
|----------|---------------|--------------|-------|
| **AWS** | t3.medium | $30-40 | Elastic IPs, good for PoC |
| **DigitalOcean** | 4GB Droplet | $24-30 | Simplest setup, integrated monitoring |
| **Linode** | Linode 8GB | $40-50 | Excellent uptime, good support |
| **Hetzner** | CX31 | €8-10 | Most cost-effective in EU/Asia |
| **Local/On-Prem** | 4 vCPU bare metal | $0 | For internal training |

### Network & DNS Configuration

#### Option 1: Public Domain (Production-like)
```
Domain:          demo.auditsphere.id
Client 1:        client1.demo.auditsphere.id
Client 2:        client2.demo.auditsphere.id
API Gateway:     api-client1.demo.auditsphere.id

DNS Records:
A    demo.auditsphere.id          → 123.45.67.89 (VPS IP)
A    *.demo.auditsphere.id        → 123.45.67.89 (wildcard)

SSL Certificates: Let's Encrypt (automatic via Certbot)
```

#### Option 2: Local/Localhost (Development)
```
Domain Pattern:   *.localhost (or *.demo.local)
Client 1:        client1.localhost
Client 2:        client2.localhost

DNS:             /etc/hosts entries (Linux/Mac) or Windows Hosts file
SSL:             Self-signed certificates (browser warnings OK)
Access:          VPN or SSH tunnel from external
```

#### Option 3: Hybrid (Local backend, cloud DNS)
```
Real Domain:     demo.auditsphere.id → VPS IP
Backend Access:  VPN into VPS network
Internal Test:   client1.localhost for local development
```

---

## 4. Pre-Setup Checklist

Before you start, verify you have:

- [ ] **VPS Access**
  - [ ] SSH key pair generated
  - [ ] VPS IP address noted
  - [ ] SSH login confirmed: `ssh -i key.pem ubuntu@<IP>`

- [ ] **Domain & DNS**
  - [ ] Domain registrar credentials ready (Namecheap, Cloudflare, etc.)
  - [ ] DNS provider set up (or using registrar's DNS)
  - [ ] Wildcard DNS record or A records pointing to VPS
  - [ ] DNS propagation verified: `nslookup demo.auditsphere.id`

- [ ] **GitHub Access**
  - [ ] Personal access token or SSH key configured
  - [ ] Repository cloned successfully: `git clone https://github.com/RauMomo/risk-based-audit.git`

- [ ] **SSL/TLS (if using public domain)**
  - [ ] Email for Let's Encrypt notifications
  - [ ] Ports 80 & 443 open on VPS
  - [ ] Firewall rules configured

- [ ] **Storage (Optional)**
  - [ ] Google Drive folder created (for file uploads)
  - [ ] Google Cloud Service Account JSON key downloaded
  - [ ] Service account has Editor permissions on Drive folder

- [ ] **Tools Installed Locally**
  - [ ] `ssh` (for remote access)
  - [ ] `curl` (for API testing)
  - [ ] `git` (for cloning repo)
  - [ ] `docker` & `docker compose` (if testing locally first)

### Quick Connectivity Test

```bash
# Test SSH access
ssh -i ~/.ssh/demo-key.pem ubuntu@123.45.67.89 "echo 'SSH works!'"

# Test domain resolution
nslookup client1.demo.auditsphere.id
# Should return: 123.45.67.89

# Test HTTP connectivity
curl -I http://client1.demo.auditsphere.id
# Should return: 200 (or redirect to HTTPS)
```

---

## 5. Step-by-Step Deployment

### 5.1 Provision the VPS & Install Dependencies

#### Step 1A: SSH into the VPS

```bash
ssh -i ~/.ssh/demo-key.pem ubuntu@123.45.67.89

# Once logged in, become root (or use sudo for commands below)
sudo su -
```

#### Step 1B: Update System & Install Docker

```bash
# Update package manager
apt-get update
apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose (included in Docker Desktop 20.10+)
docker --version      # Verify Docker
docker compose version # Verify Compose

# Allow docker commands without sudo
usermod -aG docker ubuntu
su - ubuntu  # Re-login to apply group changes
```

#### Step 1C: Install Nginx & Certbot (for SSL)

```bash
apt-get install -y nginx certbot python3-certbot-nginx

# Start Nginx
systemctl start nginx
systemctl enable nginx  # Auto-start on reboot
```

#### Step 1D: Install Git & Clone AuditSphere

```bash
apt-get install -y git

# Create app directory
mkdir -p /app
cd /app

# Clone the repository (use your GitHub token if private)
git clone https://github.com/RauMomo/risk-based-audit.git
cd risk-based-audit
```

#### Step 1E: Verify Docker Connectivity

```bash
# Check that Docker daemon is running
docker ps

# Pull a test image to verify internet
docker pull hello-world
docker run hello-world

# Remove test image
docker rmi hello-world
```

### 5.2 Configure Networking & SSL

#### Step 2A: Set Up DNS Wildcard Records

On your **DNS provider** (Cloudflare, Route 53, Namecheap, etc.):

```
Record Type: A (or CNAME)
Name:        * (or *.demo)
Value:       123.45.67.89 (your VPS public IP)
TTL:         300 (or Auto)
Proxy:       DNS only (Cloudflare) - DO NOT proxy to avoid SSL issues
```

Wait 5-15 minutes for propagation. Test:

```bash
# From your VPS
nslookup client1.demo.auditsphere.id
# Should return 123.45.67.89
```

#### Step 2B: Request SSL Certificate

```bash
# For wildcard or multiple subdomains
sudo certbot certonly --nginx \
  -d demo.auditsphere.id \
  -d "*.demo.auditsphere.id" \
  --agree-tos \
  --email your-email@company.com \
  --non-interactive

# Certificates stored in: /etc/letsencrypt/live/demo.auditsphere.id/

# Verify certificate
ls -la /etc/letsencrypt/live/demo.auditsphere.id/
```

#### Step 2C: Create Nginx Reverse Proxy Template

Create `/etc/nginx/sites-available/auditsphere-demo.conf`:

```nginx
# ─────────────────────────────────────────────────────────────
# AuditSphere Demo - Reverse Proxy Template
# Dynamically routes *.demo.auditsphere.id subdomains
# ─────────────────────────────────────────────────────────────

# HTTP → HTTPS redirect
server {
    listen 80;
    listen [::]:80;
    server_name ~^(.+)\.demo\.auditsphere\.id$ ~^demo\.auditsphere\.id$;
    return 301 https://$host$request_uri;
}

# HTTPS - Frontend (Nuxt)
# Routes: client1.demo.auditsphere.id → localhost:3010
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ~^(?<client>[^.]+)\.demo\.auditsphere\.id$;

    ssl_certificate /etc/letsencrypt/live/demo.auditsphere.id/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/demo.auditsphere.id/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    client_max_body_size 100M;

    # Serve frontend from backend/tenants/{client}/frontend on port 3010
    location / {
        # Map port based on client name
        # client1 → 3010, client2 → 3011, etc.
        set $frontend_port 3010;
        if ($client = "client2") { set $frontend_port 3011; }
        if ($client = "client3") { set $frontend_port 3012; }

        proxy_pass http://127.0.0.1:$frontend_port;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# HTTPS - API Gateway (Kong)
# Routes: api-client1.demo.auditsphere.id → localhost:8090
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ~^api-(?<client>[^.]+)\.demo\.auditsphere\.id$;

    ssl_certificate /etc/letsencrypt/live/demo.auditsphere.id/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/demo.auditsphere.id/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    client_max_body_size 100M;

    # Serve Kong gateway from backend/tenants/{client}/kong on port 8090
    location / {
        # Map port based on client name
        # client1 → 8090, client2 → 8091, etc.
        set $kong_port 8090;
        if ($client = "client2") { set $kong_port 8091; }
        if ($client = "client3") { set $kong_port 8092; }

        proxy_pass http://127.0.0.1:$kong_port;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### Step 2D: Enable Nginx Configuration

```bash
# Create symlink to enable the config
sudo ln -s /etc/nginx/sites-available/auditsphere-demo.conf \
           /etc/nginx/sites-enabled/

# Test Nginx syntax
sudo nginx -t
# Output: nginx: configuration file test is successful

# Reload Nginx
sudo systemctl reload nginx
```

### 5.3 Deploy Shared Infrastructure (PostgreSQL, Redis, Kafka)

These are **shared** across all demo tenants (to save resources).

#### Step 3A: Create Docker Network

```bash
cd /app/risk-based-audit
docker network create rb_audit_network 2>/dev/null || true
```

#### Step 3B: Start Shared Services

From the repository, use the main backend docker-compose:

```bash
cd /app/risk-based-audit/backend

# Build shared images (this may take 5-10 minutes)
docker compose build

# Start shared services (PostgreSQL, Redis, Kafka, etc.)
docker compose up -d postgres redis kafka zookeeper

# Wait for services to be healthy
docker compose ps
# Should show all services as "Up (healthy)"

# Verify PostgreSQL is accepting connections
docker compose exec postgres psql -U postgres -c "SELECT version();"
```

> **Note:** The shared compose file automatically seeds databases and migrations for shared services.

---

## 6. Multi-Tenant Demo Setup

AuditSphere supports **per-tenant isolation**, meaning each client gets their own Kong gateway, frontend, and databases.

### 6.1 Overview of Tenant Structure

```
/app/risk-based-audit/
├── backend/
│   ├── tenants/
│   │   ├── client1/
│   │   │   ├── .env                    # Secrets (gitignored)
│   │   │   ├── docker-compose.yml     # Tenant stack
│   │   │   ├── kong/
│   │   │   │   └── kong.yml           # Kong routing config
│   │   │   └── uploads/               # File storage (Google Drive fallback)
│   │   ├── client2/
│   │   │   └── ...
│   │   └── registry.tsv               # Port allocation registry
│   └── scripts/
│       ├── generate-tenant-stack.sh
│       ├── onboard-tenant.sh
│       └── templates/
│           ├── docker-compose.tenant.yml.tpl
│           └── kong-tenant.yml.tpl
```

### 6.2 Generate & Onboard Demo Tenant 1

```bash
cd /app/risk-based-audit

# Make scripts executable
chmod +x backend/scripts/generate-tenant-stack.sh
chmod +x backend/scripts/onboard-tenant.sh

# Generate tenant stack for "client1"
# Arguments: <slug> <display-name> [fe_port] [kong_port] [--local] [--empty-data]
backend/scripts/generate-tenant-stack.sh client1 "Demo Client 1"

# Onboard the tenant (provision databases, run migrations, start services)
backend/scripts/onboard-tenant.sh client1 "Demo Client 1" 3010 8090 --with-demo-data

# Wait for the script to complete (takes 3-5 minutes)
```

**What the onboard script does:**
1. ✅ Generates Kong gateway config for the tenant
2. ✅ Allocates ports (3010 for frontend, 8090 for API gateway)
3. ✅ Generates `.env` file with unique JWT secrets
4. ✅ Creates PostgreSQL databases (`rb_audit_*_client1`)
5. ✅ Runs database migrations
6. ✅ Seeds demo data (sample audit charters, findings, etc.)
7. ✅ Starts all containers (Kong, frontend, auth-service, audit-service, etc.)

### 6.3 Verify Tenant Deployment

```bash
# Check if tenant containers are running
docker compose -f backend/tenants/client1/docker-compose.yml ps

# Should show:
# - kong-client1         (healthy)
# - auth-service-client1 (running)
# - audit-service-client1
# - master-service-client1
# - risk-service-client1
# - analytics-service-client1
# - <frontend service>

# Test API health
curl -I https://api-client1.demo.auditsphere.id/api/v1/auth/health
# Should return: HTTP/2 200 (after ~30 seconds for startup)

# Test frontend
curl -I https://client1.demo.auditsphere.id
# Should return: HTTP/2 200
```

### 6.4 Onboard Additional Demo Tenants

Repeat for client2, client3, etc.:

```bash
# Tenant 2
backend/scripts/onboard-tenant.sh client2 "Demo Client 2" 3011 8091 --with-demo-data

# Tenant 3
backend/scripts/onboard-tenant.sh client3 "Demo Client 3" 3012 8092 --with-demo-data

# Wait 2-3 minutes between tenants to avoid resource contention
```

Each tenant gets:
- Own Kong gateway instance
- Own frontend container
- Own set of databases (isolated, no cross-tenant access)
- Unique JWT secrets (isolation enforced)
- Dedicated Redis database index

---

## 7. Demo Data Seeding

### 7.1 What Data Is Seeded?

When you run the tenant onboarding with `--with-demo-data`, the following sample data is created:

#### Audit Service Seed Data
- **20 Sample Audit Engagements** with status progression (Planning → In Progress → Completed)
- **50 Sample Audit Activities** (audit procedures, fieldwork steps)
- **200+ Sample Audit Findings** (observations, recommendations, risk levels)
- **30 Sample Working Papers** (documentation artifacts)
- **150 File Upload Fixtures** (mock evidence documents, interview notes)

#### Master Service Seed Data
- **15 Sample Companies/Divisions**
- **10 Sample Risk Frameworks** (COSO, ISO 27001, SOX, etc.)
- **50 Sample Process Categories**
- **200 Sample Master Data Records**

#### Risk Service Seed Data
- **5 Sample Risk Registers** (enterprise + departmental)
- **100 Sample Risk Items** (with probability, impact, residual risk)
- **10 Sample Risk Matrices** (heatmaps, thresholds)
- **Risk Assessment Results** (historical scoring)

#### Authentication Service Seed Data
- **RBAC Structure**: Super Admin, CAE, Audit Lead, Auditor, Auditee roles
- **Default Admin User**: `admin` / `password123` (rotate in production!)
- **Sample Team Members**: Auditors, auditees across multiple divisions

### 7.2 Customize Seeding

#### Option A: Minimal Data (Fast Setup)

```bash
# Skip demo data, only create empty schema
backend/scripts/onboard-tenant.sh client1 "Demo Client 1" 3010 8090 --empty-data
```

Result: Empty databases with only schema and roles. Good for:
- Testing data import workflows
- Showcasing custom data loading
- Performance testing with fresh state

#### Option B: Heavy Load Testing Data

Edit `backend/audit-service/cmd/seed.go` to increase sample sizes:

```go
// Line 54 in seed.go
seedStrategicOnly: 1000  // Instead of 100 → create 1000 audit items
```

Then recompile:

```bash
cd backend/audit-service
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o audit ./cmd
cd /app/risk-based-audit
backend/scripts/onboard-tenant.sh client1 "Demo Client 1" 3010 8090
```

#### Option C: Custom Seed Data

Create a CSV or JSON file with your data:

```bash
# Generate 100 sample audit engagement records
cat > /tmp/sample-audits.json <<'EOF'
{
  "audits": [
    { "id": "AUD-2024-001", "name": "IT Control Assessment", "status": "In Progress" },
    { "id": "AUD-2024-002", "name": "Fraud Risk Review", "status": "Planning" }
  ]
}
EOF

# Load via API after tenant is running
curl -X POST https://api-client1.demo.auditsphere.id/api/v1/audit/bulk-import \
  -H "Authorization: Bearer <jwt-token>" \
  -d @/tmp/sample-audits.json
```

---

## 8. Access & Credentials

### 8.1 Default Demo Credentials

**Username:** `admin`  
**Password:** `password123`

### 8.2 Web Access

#### Frontend URLs (for each client tenant)

| Client | Frontend URL | Purpose |
|--------|--------------|---------|
| Demo 1 | https://client1.demo.auditsphere.id | Dashboard, audit plans, findings |
| Demo 2 | https://client2.demo.auditsphere.id | Parallel testing, A/B demos |
| Demo 3 | https://client3.demo.auditsphere.id | Load testing, stress testing |

#### API Gateway URLs

| Client | API URL | Purpose |
|--------|---------|---------|
| Demo 1 | https://api-client1.demo.auditsphere.id | Direct API calls, testing |
| Demo 2 | https://api-client2.demo.auditsphere.id | Parallel API calls |
| Demo 3 | https://api-client3.demo.auditsphere.id | Load testing |

#### Direct Access

```bash
# Test frontend is alive
curl -I https://client1.demo.auditsphere.id
# HTTP/2 200 OK

# Test API is responding
curl https://api-client1.demo.auditsphere.id/api/v1/auth/health
# { "status": "ok" }

# Login and get JWT token
curl -X POST https://api-client1.demo.auditsphere.id/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@company.com","password":"password123"}'
# Response: { "token": "eyJhbGc..." }
```

### 8.3 Backend Service Ports (Internal Only)

These are **NOT exposed to the internet** and only accessible from the VPS itself:

| Service | Port | Purpose |
|---------|------|---------|
| Kong Admin API | 8019 | Kong configuration (client1) |
| Kong Admin API | 8020 | Kong configuration (client2) |
| Auth Service | 8001 | User management |
| Audit Service | 8002 | Audit operations |
| Master Service | 8003 | Master data |
| Risk Service | 8004 | Risk calculations |
| Analytics Service | 8084 | Scoring & reports |
| PostgreSQL | 5432 (not exposed) | Database |
| Redis | 6379 (not exposed) | Cache |
| Kafka | 29092 (internal), 9092 (external) | Event streaming |
| Kafka UI | 8085 | Event monitoring |

### 8.4 SSH Access for Debugging

```bash
# Connect to VPS
ssh -i ~/.ssh/demo-key.pem ubuntu@demo.auditsphere.id

# Once logged in, check container logs
docker logs rb_audit_auth_service
docker logs rb_audit_postgres

# Access database directly
docker exec -it rb_audit_postgres psql -U postgres

# Inside psql:
\l                    # List databases
\c rb_audit_auth_client1  # Connect to client1 auth DB
SELECT * FROM users;  # Query users table
```

### 8.5 Secure Credential Management

#### Create a `.env.secrets` file (gitignored)

```bash
# .env.secrets (never commit this!)
ADMIN_EMAIL=demo-admin@company.com
ADMIN_PASSWORD=GenerateStrongPasswordHere!
RESEND_API_KEY=re_xxxxxxxxxxxxxxxxxxxxx
GOOGLE_DRIVE_SERVICE_ACCOUNT_JSON='{"type":"service_account",...}'
JWT_SECRET_CLIENT1=<random-32-char-string>
JWT_SECRET_CLIENT2=<random-32-char-string>
```

```bash
# Load secrets before operations
source .env.secrets

# Use in scripts
echo $ADMIN_PASSWORD
```

#### Change Default Credentials (Before Demo)

```bash
# Connect to auth service
docker exec -it rb_audit_auth_service bash

# Use internal CLI to update password
./auth --config ./pkg/config/config.yaml user update \
  --email admin@company.com \
  --password "NewSecurePassword123!"

# Exit container
exit
```

Or via API:

```bash
# Get JWT token first
TOKEN=$(curl -s -X POST https://api-client1.demo.auditsphere.id/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@company.com","password":"password123"}' | jq -r '.token')

# Update password
curl -X PUT https://api-client1.demo.auditsphere.id/api/v1/users/me/password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"new_password":"NewSecurePassword123!"}'
```

---

## 9. Running a Successful Demo

### 9.1 Pre-Demo Checklist (30 minutes before)

- [ ] **Verify All Services Are Running**
  ```bash
  docker compose -f backend/tenants/client1/docker-compose.yml ps
  docker compose -f backend/tenants/client2/docker-compose.yml ps
  # All services should be "Up (healthy)"
  ```

- [ ] **Check API Health**
  ```bash
  curl https://api-client1.demo.auditsphere.id/api/v1/auth/health
  # Should return 200 OK
  ```

- [ ] **Test Frontend Loading**
  ```bash
  # Open browser: https://client1.demo.auditsphere.id
  # Page should load in < 3 seconds
  # Login form should be visible
  ```

- [ ] **Verify Sample Data Exists**
  ```bash
  curl -X GET https://api-client1.demo.auditsphere.id/api/v1/audits \
    -H "Authorization: Bearer <token>"
  # Should return list of audit engagements
  ```

- [ ] **Check Database Connectivity**
  ```bash
  docker exec rb_audit_postgres psql -U postgres -c \
    "SELECT COUNT(*) FROM rb_audit_audit_client1.audit_engagements;"
  # Should return row count > 0
  ```

- [ ] **Test File Uploads (Optional)**
  ```bash
  # Create dummy file
  echo "Sample evidence document" > /tmp/evidence.txt

  # Upload via API
  curl -X POST https://api-client1.demo.auditsphere.id/api/v1/audit/1/upload \
    -H "Authorization: Bearer <token>" \
    -F "file=@/tmp/evidence.txt"
  # Should return 200 OK with file metadata
  ```

- [ ] **Verify Network Connectivity**
  ```bash
  # Check latency to VPS
  ping -c 3 demo.auditsphere.id

  # Check DNS resolution
  nslookup client1.demo.auditsphere.id
  ```

### 9.2 Demo Flow & Talking Points

#### 1. Authentication & RBAC (2 min)
- Login as Admin: `admin` / `password123`
- Show role-based dashboard (Super Admin view)
- Explain 5 roles: Super Admin, CAE, Audit Lead, Auditor, Auditee
- Show permission granularity (who can create vs. view vs. edit)

#### 2. Dashboard & Overview (2 min)
- Highlight KPIs: Total audits, findings, risk scores
- Show Redis-powered caching for fast dashboard load (< 2s)
- Explain real-time updates via Kafka events

#### 3. Audit Planning (3 min)
- Navigate to "Audit Engagements"
- Show 20 sample audit plans at different stages
- Explain engagement lifecycle: Planning → Scoping → Fieldwork → Reporting
- Edit an engagement to show form validation, required fields

#### 4. Fieldwork & Findings (3 min)
- Open a live audit engagement
- Show 50+ sample audit activities (procedures)
- Navigate to "Findings" tab
- Show 200+ sample findings with severity levels
- Explain risk calculation: Likelihood × Impact = Risk Score

#### 5. File Management & Google Drive (2 min)
- Show "Evidence Documents" upload interface
- Upload sample file (Word doc, PDF, image)
- Explain automatic Google Drive storage + local fallback
- Show file versioning and access control

#### 6. Risk Module (2 min)
- Navigate to "Risk Registers"
- Show enterprise + departmental risk matrices
- Display 100 sample risk items with heat map visualization
- Explain residual vs. inherent risk

#### 7. Reports & Analytics (2 min)
- Show automated report generation
- Display risk scoring powered by Python AI
- Show exportable dashboards (PDF, Excel)
- Explain audit trail and compliance tracking

#### 8. Scalability & Multi-Tenancy (2 min)
- Open second browser tab: `client2.demo.auditsphere.id`
- Demonstrate isolated data (no cross-tenant leakage)
- Show Kong gateway routing audit logs
- Explain per-tenant databases and JWT secrets

**Total Demo Time:** ~18-20 minutes + Q&A

### 9.3 Handling Common Demo Questions

| Question | Answer |
|----------|--------|
| **What if login fails?** | Password is case-sensitive. Default is `admin` / `password123`. If still failing, check API health: `curl api.../auth/health`. If unhealthy, restart service: `docker restart rb_audit_auth_service`. |
| **Why is the dashboard slow?** | First load may be slow due to large dataset seeding. Subsequent loads are cached in Redis (should be < 500ms). If slow persists, check server CPU: `docker stats`. |
| **Can I test with my own data?** | Yes! Use `--empty-data` flag during onboarding, then import via API or bulk-upload CSV. Contact engineering for custom seeding scripts. |
| **What happens if I close the browser?** | Session is stored in Redis. Reopening the tab will still have auth active until JWT expires (default 24 hours). Clearing cookies logs you out. |
| **How do I reset the demo?** | Stop the stack (`docker compose down`), delete databases (`DROP DATABASE ...`), and re-run the onboarding script. Total time: ~5 minutes. |
| **Can I use real company data?** | Not in demo. For POC with real data, provision a separate "staging" environment with data governance controls. Contact compliance team. |

---

## 10. Troubleshooting Common Issues

### Issue 1: Cannot Access Frontend (HTTPS Error)

**Symptoms:**  
```
curl: (60) SSL certificate problem: self-signed certificate
```

**Diagnosis:**
```bash
# Check if Nginx is running
systemctl status nginx
# Should show: active (running)

# Check if Let's Encrypt certificate exists
ls -la /etc/letsencrypt/live/demo.auditsphere.id/
# Should show: fullchain.pem, privkey.pem, cert.pem

# Check DNS resolution
nslookup client1.demo.auditsphere.id
# Should return your VPS IP, not 127.0.0.1
```

**Solutions:**

**A) Certificate Not Yet Issued**
```bash
# Re-request certificate
sudo certbot certonly --nginx \
  -d "demo.auditsphere.id" \
  -d "*.demo.auditsphere.id" \
  --force-renewal

# Reload Nginx
sudo systemctl reload nginx

# Test after 1-2 minutes
curl -I https://client1.demo.auditsphere.id
```

**B) DNS Not Resolving**
```bash
# Verify DNS records at registrar
# A record should point to your VPS IP

# Flush local DNS cache (Linux)
sudo systemctl restart systemd-resolved

# Flush local DNS cache (macOS)
sudo dscacheutil -flushcache

# Wait 5-10 minutes, then test
nslookup client1.demo.auditsphere.id
```

**C) Nginx Not Forwarding**
```bash
# Check Nginx config syntax
sudo nginx -t

# Check Nginx is listening on 443
sudo netstat -tlnp | grep nginx
# Should show: 0.0.0.0:443 or [::]:443

# Reload Nginx
sudo systemctl reload nginx
```

---

### Issue 2: API Returns 502 Bad Gateway

**Symptoms:**
```
curl -I https://api-client1.demo.auditsphere.id/api/v1/auth/health
# HTTP/2 502
```

**Diagnosis:**
```bash
# Check if Kong container is running
docker ps | grep kong-client1
# Should show "Up (healthy)"

# Check Kong logs
docker logs kong-client1 | tail -20

# Check if Kong is listening
docker exec kong-client1 kong health
# Should show: "OK" for database and proxy
```

**Solutions:**

**A) Kong Container Crashed**
```bash
# Restart Kong
docker restart kong-client1

# Wait 30 seconds for startup
sleep 30

# Test health again
curl https://api-client1.demo.auditsphere.id/api/v1/auth/health
```

**B) Kong Config Syntax Error**
```bash
# Validate Kong config
docker run --rm -v /app/risk-based-audit/backend/tenants/client1/kong:/cfg:ro \
  kong:3.4 kong config parse /cfg/kong.yml

# If error, check the config file
cat /app/risk-based-audit/backend/tenants/client1/kong/kong.yml

# Regenerate from template
cd /app/risk-based-audit
chmod +x backend/scripts/generate-tenant-stack.sh
backend/scripts/generate-tenant-stack.sh client1 "Demo Client 1" --force
```

**C) Upstream Services Not Running**
```bash
# Check all tenant services
docker compose -f backend/tenants/client1/docker-compose.yml ps

# If any service is "Exited", check logs
docker logs rb_audit_auth_service-client1
docker logs rb_audit_audit_service-client1

# Restart specific service
docker restart rb_audit_auth_service-client1
```

---

### Issue 3: Database Connection Error

**Symptoms:**
```
Error: connect ECONNREFUSED 127.0.0.1:5432
```

**Diagnosis:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres
# Should show: rb_audit_postgres (Up)

# Check if port is exposed (internal only)
docker port rb_audit_postgres
# Should show: No port mappings (intentional - only Docker network)

# Check logs
docker logs rb_audit_postgres
```

**Solutions:**

**A) PostgreSQL Container Crashed**
```bash
# Restart PostgreSQL
docker restart rb_audit_postgres

# Wait for startup
sleep 10

# Verify it's healthy
docker exec rb_audit_postgres psql -U postgres -c "SELECT 1;"
# Should return: 1
```

**B) Disk Space Full**
```bash
# Check disk usage
df -h

# If > 90%, clean up Docker
docker system prune -a --volumes

# If still full, delete old PostgreSQL snapshots
docker volume ls
docker volume rm <volume-name>
```

**C) Database Migration Failed**
```bash
# Check migration logs
docker compose -f backend/tenants/client1/docker-compose.yml logs auth-service-client1

# Re-run migrations
docker compose -f backend/tenants/client1/docker-compose.yml run --rm auth-service-client1 ./auth migrate up

# If migration fails due to schema, reset database
docker exec rb_audit_postgres psql -U postgres -c "DROP DATABASE rb_audit_auth_client1;"
```

---

### Issue 4: Slow Performance / High CPU

**Symptoms:**
```
Dashboard takes > 5 seconds to load
API responses > 1 second
```

**Diagnosis:**
```bash
# Check resource usage
docker stats

# Check which container is using most CPU/memory
docker top rb_audit_audit_service-client1

# Check slow queries in PostgreSQL
docker exec rb_audit_postgres psql -U postgres -c \
  "SELECT query, calls, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 5;"
```

**Solutions:**

**A) Too Many Demo Tenants Running**
```bash
# Check how many tenants are running
docker ps | grep -E "(auth|audit|master|risk).*client" | wc -l

# If > 5, consider stopping some
docker compose -f backend/tenants/client3/docker-compose.yml down

# Keep only 2-3 for demo
```

**B) PostgreSQL Cache Too Small**
```bash
# Edit PostgreSQL config
docker exec rb_audit_postgres bash -c \
  "sed -i 's/shared_buffers = 128MB/shared_buffers = 512MB/' /var/lib/postgresql/data/postgresql.conf"

# Restart PostgreSQL
docker restart rb_audit_postgres
```

**C) Redis Not Caching Properly**
```bash
# Check Redis memory usage
docker exec rb_audit_redis redis-cli INFO memory

# Flush old cache
docker exec rb_audit_redis redis-cli FLUSHALL

# Restart services (they will rebuild cache)
docker compose -f backend/tenants/client1/docker-compose.yml restart
```

---

### Issue 5: File Upload Fails

**Symptoms:**
```
Error uploading file to Google Drive
File saved locally but not visible in Drive folder
```

**Diagnosis:**
```bash
# Check if Google Drive is enabled
docker exec rb_audit_audit_service-client1 grep -i gdrive ./pkg/config/config.yaml

# Check if service account credentials exist
docker exec rb_audit_audit_service-client1 ls -la ./pkg/config/gdrive-credentials.json

# Check audit-service logs
docker logs rb_audit_audit_service-client1 | grep -i drive
```

**Solutions:**

**A) Google Drive Credentials Invalid**
```bash
# Copy new service account JSON key
cp ~/Downloads/service-account-key.json \
   backend/audit-service/pkg/config/gdrive-credentials.json

# Restart audit service
docker restart rb_audit_audit_service-client1
```

**B) Fall Back to Local Storage**
```bash
# Disable Google Drive to use local fallback
docker exec rb_audit_audit_service-client1 \
  sed -i 's/enabled: true/enabled: false/' ./pkg/config/config.yaml

# Restart
docker restart rb_audit_audit_service-client1

# Files now save to: /app/risk-based-audit/backend/tenants/client1/uploads/
```

**C) Check Firewall/Network**
```bash
# Test Google Drive API connectivity
docker exec rb_audit_audit_service-client1 \
  curl -I https://www.googleapis.com/auth/drive

# If timeout, check firewall rules
sudo ufw status
# Verify outbound HTTPS (443) is allowed
```

---

### Issue 6: Cannot Log In (Invalid Credentials)

**Symptoms:**
```json
POST /api/v1/auth/login
Response: { "error": "Invalid credentials" }
```

**Diagnosis:**
```bash
# Check if user exists in database
docker exec rb_audit_postgres psql -U postgres \
  -d rb_audit_auth_client1 \
  -c "SELECT email, password_hash FROM users;"

# Verify auth service is running
docker ps | grep auth-service-client1
```

**Solutions:**

**A) User Not Seeded**
```bash
# Re-run auth seeder
docker compose -f backend/tenants/client1/docker-compose.yml run --rm auth-service-client1 ./auth seed

# Verify users were created
docker exec rb_audit_postgres psql -U postgres \
  -d rb_audit_auth_client1 \
  -c "SELECT COUNT(*) FROM users;"
# Should return: > 0
```

**B) Wrong Password**
```bash
# Credentials are case-sensitive
# Correct: admin / password123 (no spaces, exact match)

# Reset password if forgotten
docker compose -f backend/tenants/client1/docker-compose.yml run --rm auth-service-client1 \
  ./auth user reset-password --email admin@company.com --password newpassword123
```

**C) JWT Secret Mismatch**
```bash
# Check if JWT_SECRET is consistent between services
cat backend/tenants/client1/.env | grep JWT_SECRET

# If mismatch, regenerate
cd backend
./scripts/generate-tenant-stack.sh client1 "Demo Client 1" --force

# Restart all services
docker compose -f backend/tenants/client1/docker-compose.yml restart
```

---

### Issue 7: Kong Cannot Route to Services

**Symptoms:**
```
curl -I https://api-client1.demo.auditsphere.id/api/v1/auth/status
HTTP/2 502
```

**Diagnosis:**
```bash
# Check Kong upstreams configuration
docker exec kong-client1 kong info
# Should list all services

# Check specific service status
docker exec kong-client1 curl http://localhost:8001/upstreams/auth-service-client1/health
# Should show active targets
```

**Solutions:**

**A) Upstream Service Not Responding**
```bash
# Test auth service directly from Kong container
docker exec kong-client1 \
  curl -I http://auth-service-client1:8001/health

# If timeout, service is down. Check logs
docker logs rb_audit_auth_service-client1
```

**B) Kong Config Missing Service**
```bash
# View Kong config
docker exec kong-client1 cat /usr/local/kong/declarative/kong.yml

# Regenerate config
cd /app/risk-based-audit
backend/scripts/generate-tenant-stack.sh client1 "Demo Client 1" --force

# Apply new config
docker restart kong-client1
```

---

## 11. Cleanup & Maintenance

### 11.1 Daily Maintenance Tasks

```bash
# 1. Monitor container health (run hourly)
watch docker ps

# 2. Check disk usage (should stay < 80%)
df -h

# 3. Review logs for errors (daily)
docker logs rb_audit_postgres --tail=100 | grep -i error
docker logs kong-client1 --tail=100 | grep -i error

# 4. Clear old Docker images (weekly)
docker image prune -a

# 5. Backup demo databases (optional)
mkdir -p /backups
for db in rb_audit_auth_client1 rb_audit_audit_client1; do
  docker exec rb_audit_postgres pg_dump -U postgres "$db" | \
    gzip > "/backups/${db}_$(date +%Y%m%d).sql.gz"
done
```

### 11.2 Stopping & Restarting Demo

**Stop all demo tenants (preserve databases):**
```bash
docker compose -f backend/tenants/client1/docker-compose.yml down
docker compose -f backend/tenants/client2/docker-compose.yml down
docker compose down  # Shared services remain
```

**Restart all demo tenants:**
```bash
docker compose -f backend/tenants/client1/docker-compose.yml up -d
docker compose -f backend/tenants/client2/docker-compose.yml up -d

# Wait for health checks
sleep 30
docker ps
```

**Full reset (WARNING: Destroys all data):**
```bash
# Stop everything
docker compose down
docker compose -f backend/tenants/client1/docker-compose.yml down
docker compose -f backend/tenants/client2/docker-compose.yml down

# Delete databases
docker exec rb_audit_postgres psql -U postgres <<EOF
DROP DATABASE IF EXISTS rb_audit_auth_client1;
DROP DATABASE IF EXISTS rb_audit_audit_client1;
DROP DATABASE IF EXISTS rb_audit_master_client1;
DROP DATABASE IF EXISTS rb_audit_risk_client1;
EOF

# Restart fresh
cd backend
docker compose up -d
backend/scripts/onboard-tenant.sh client1 "Demo Client 1" 3010 8090 --with-demo-data
```

### 11.3 Updating AuditSphere Code

```bash
# Pull latest from GitHub
cd /app/risk-based-audit
git pull origin main

# Rebuild Docker images
cd backend
docker compose build

# Restart services
docker compose down
docker compose up -d

# For tenants, rebuild and restart
for tenant in client1 client2 client3; do
  docker compose -f tenants/$tenant/docker-compose.yml build
  docker compose -f tenants/$tenant/docker-compose.yml restart
done
```

### 11.4 Scaling Demo to More Tenants

```bash
# For each new tenant:
cd /app/risk-based-audit

# Generate stack (port auto-increments)
backend/scripts/generate-tenant-stack.sh client4 "Demo Client 4"

# Onboard
backend/scripts/onboard-tenant.sh client4 "Demo Client 4" 3013 8093 --with-demo-data

# Add Nginx routing (if not already using wildcard)
# Update /etc/nginx/sites-available/auditsphere-demo.conf if needed
sudo systemctl reload nginx
```

### 11.5 Archiving & Offboarding Demo

```bash
# Export client data before deletion
docker exec rb_audit_postgres pg_dump -U postgres rb_audit_audit_client1 | \
  gzip > /backups/client1-export-$(date +%Y%m%d).sql.gz

# Stop tenant
docker compose -f backend/tenants/client1/docker-compose.yml down

# Delete databases
docker exec rb_audit_postgres psql -U postgres <<EOF
DROP DATABASE rb_audit_auth_client1;
DROP DATABASE rb_audit_audit_client1;
DROP DATABASE rb_audit_master_client1;
DROP DATABASE rb_audit_risk_client1;
EOF

# Remove from registry
sed -i '/client1/d' backend/tenants/registry.tsv

# Delete tenant directory
rm -rf backend/tenants/client1
```

---

## Appendix: Quick Reference Commands

### System Health

```bash
# Check all services
docker ps -a

# Check resource usage
docker stats

# Check network
docker network inspect rb_audit_network

# Check volumes
docker volume ls
```

### Database Operations

```bash
# Access PostgreSQL
docker exec -it rb_audit_postgres psql -U postgres

# List all databases
\l

# Connect to tenant database
\c rb_audit_audit_client1

# Show tables
\dt

# Query audit findings
SELECT COUNT(*) FROM audit_findings;

# Exit
\q
```

### Service Restart

```bash
# Restart single service
docker restart rb_audit_auth_service

# Restart all tenant services
docker compose -f backend/tenants/client1/docker-compose.yml restart

# Restart and view logs
docker restart rb_audit_audit_service && docker logs -f rb_audit_audit_service
```

### Logs & Debugging

```bash
# View live logs
docker logs -f rb_audit_auth_service

# Last 50 lines
docker logs rb_audit_auth_service --tail=50

# Logs with timestamps
docker logs rb_audit_auth_service -t --tail=100

# Grep for errors
docker logs rb_audit_postgres | grep -i error
```

### API Testing

```bash
# Login and get token
TOKEN=$(curl -s -X POST http://api-client1.localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@company.com","password":"password123"}' | \
  jq -r '.token')

# Use token in subsequent requests
curl -H "Authorization: Bearer $TOKEN" \
  http://api-client1.localhost/api/v1/audits

# Save token for reuse
echo $TOKEN > /tmp/demo-token.txt
```

---

## Support & Contact

For questions or issues during demo setup:

- **Engineering Team**: Contact backend lead for architectural questions
- **DevOps Team**: SSH access issues, server provisioning
- **Product Team**: Feature demo flow, sample data customization
- **GitHub Issues**: Report bugs with reproduction steps

**Repository**: https://github.com/RauMomo/risk-based-audit  
**Documentation**: `/docs/` directory in repository

---

**End of Demo Setup Guide**
