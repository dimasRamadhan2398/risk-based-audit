# =============================================================================
# Tenant stack — __CLIENT_NAME__ (__SLUG__)
#
# GENERATED FILE. Do not edit by hand.
# Rendered from scripts/templates/docker-compose.tenant.yml.tpl by
# scripts/generate-tenant-stack.sh. Re-run the generator to regenerate.
#
# Shared with the control-plane stack (backend/docker-compose.yml), reached by
# network alias over the external rb_audit_network bridge:
#     postgres, redis, kafka, python-ai
#
# Per tenant (isolated): the five Go services, Kong, and the Nuxt frontend.
# Every service name carries the __SLUG__ suffix so its network alias is unique
# on the shared bridge — without that, two tenants both publishing the alias
# "auth-service" would make cross-tenant routing a DNS coin flip.
#
# Build contexts are relative to this file's location: backend/tenants/__SLUG__/
# =============================================================================

name: rb_audit_tenant___SLUG__

x-tenant-env: &tenant-env
  # Isolated database per service. The config key is `database.name`, which
  # viper's env-key replacer maps to DATABASE_NAME. DB_NAME is bound to nothing
  # and is silently ignored — it would fall back to the shared database.
  DATABASE_HOST: postgres
  DATABASE_PORT: "5432"
  DATABASE_USERNAME: postgres
  DATABASE_PASSWORD: postgres
  DATABASE_SSLMODE: disable

  # Redis is shared, so each tenant gets its own logical DB index. Sessions,
  # MFA challenges and rate-limit counters are plain keys with no tenant
  # prefix — on a shared index they would collide across tenants.
  # NOTE: Redis allows 16 logical DBs by default (0-15); index 0 belongs to the
  # control plane, so this scheme supports 15 tenants before `databases` has to
  # be raised in the Redis config.
  REDIS_HOST: redis
  REDIS_PORT: "6379"
  REDIS_DB: "__REDIS_DB__"

  # Per-tenant signing material. The stock config.yaml ships one hardcoded
  # jwt.secret shared by every service in the repo — leaving it in place would
  # mean a token minted for one tenant validates on every other tenant.
  JWT_SECRET: __JWT_SECRET__
  APP_SIGNATURE_KEY: __SIGNATURE_KEY__

  APP_APP_ENV: production

services:

  # ===========================================================================
  # AUTH SERVICE
  # ===========================================================================
  auth-service-__SLUG__:
    build:
      context: ../../auth-service
      dockerfile: Dockerfile
    image: rb_audit_auth_service:latest
    container_name: rb_audit___SLUG___auth_service
    environment:
      <<: *tenant-env
      DATABASE_NAME: rb_audit_auth___SLUG__
      KAFKA_SERVICE_NAME: auth-service-__SLUG__
      # Tenant services must not be able to mint Resend sending domains — that
      # is a control-plane privilege. The /api/v1/resend route is absent from
      # this tenant's Kong config; leaving the key empty is the second lock.
      RESEND_MASTER_API_KEY: ""
    # No `seed` here: onboard-tenant.sh already ran migrate + seed against this
    # tenant's database. Re-seeding on every container restart would resurrect
    # deleted rows and reset the admin password.
    command: ["./auth", "serve"]
    networks:
      - rb_audit_network
    restart: unless-stopped

  # ===========================================================================
  # AUDIT SERVICE
  # ===========================================================================
  audit-service-__SLUG__:
    build:
      context: ../../audit-service
      dockerfile: Dockerfile
    image: rb_audit_audit_service:latest
    container_name: rb_audit___SLUG___audit_service
    environment:
      <<: *tenant-env
      DATABASE_NAME: rb_audit_audit___SLUG__
      KAFKA_SERVICE_NAME: audit-service-__SLUG__
      # Evidence stays in this tenant's VPS silo. Google Drive is blocked by
      # architecture policy — the shared service-account credentials would put
      # every tenant's fieldwork in one Drive account.
      GDRIVE_ENABLED: "false"
    volumes:
      # Dedicated evidence silo. audit-service writes to ./uploads relative to
      # its WORKDIR (/root), and serves it back at GET /uploads.
      - __STORAGE_DIR__:/root/uploads
    command: ["./audit", "serve"]
    networks:
      - rb_audit_network
    restart: unless-stopped

  # ===========================================================================
  # MASTER SERVICE
  # ===========================================================================
  master-service-__SLUG__:
    build:
      context: ../../master-service
      dockerfile: Dockerfile
    image: rb_audit_master_service:latest
    container_name: rb_audit___SLUG___master_service
    environment:
      <<: *tenant-env
      DATABASE_NAME: rb_audit_master___SLUG__
      KAFKA_SERVICE_NAME: master-service-__SLUG__
    command: ["./master", "serve"]
    networks:
      - rb_audit_network
    restart: unless-stopped

  # ===========================================================================
  # RISK SERVICE
  # ===========================================================================
  risk-service-__SLUG__:
    build:
      context: ../../risk-service
      dockerfile: Dockerfile
    image: rb_audit_risk_service:latest
    container_name: rb_audit___SLUG___risk_service
    environment:
      <<: *tenant-env
      DATABASE_NAME: rb_audit_risk___SLUG__
      KAFKA_SERVICE_NAME: risk-service-__SLUG__
    command: ["./risk", "serve"]
    networks:
      - rb_audit_network
    restart: unless-stopped

  # ===========================================================================
  # ANALYTICS SERVICE
  #
  # Stateless: cmd/main.go starts a gin server and holds no database handle —
  # it forwards scoring requests to the shared python-ai model server. Run per
  # tenant anyway so this tenant's Kong has an upstream it owns.
  # ===========================================================================
  analytics-service-__SLUG__:
    build:
      context: ../../analytics-service
      dockerfile: Dockerfile
    image: rb_audit_analytics_service:latest
    container_name: rb_audit___SLUG___analytics_service
    environment:
      PYTHON_AI_URL: http://python-ai:8000
    networks:
      - rb_audit_network
    restart: unless-stopped

  # ===========================================================================
  # KONG GATEWAY
  # ===========================================================================
  kong-__SLUG__:
    image: kong:3.4
    container_name: rb_audit___SLUG___kong
    depends_on:
      - auth-service-__SLUG__
      - audit-service-__SLUG__
      - master-service-__SLUG__
      - risk-service-__SLUG__
      - analytics-service-__SLUG__
    volumes:
      - ./kong:/usr/local/kong/declarative:ro
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: /usr/local/kong/declarative/kong.yml
      KONG_PROXY_LISTEN: 0.0.0.0:8080, 0.0.0.0:8443 ssl http2
      KONG_ADMIN_LISTEN: 0.0.0.0:8001
      KONG_LOG_LEVEL: info
      KONG_PLUGINS: bundled
      KONG_NGINX_WORKER_PROCESSES: 1
      KONG_NGINX_PROXY_PROXY_BUFFERING: "off"
      KONG_CLIENT_MAX_BODY_SIZE: 100m
      KONG_NGINX_PROXY_CLIENT_MAX_BODY_SIZE: 100m
    ports:
      # Bound to loopback: host Nginx terminates TLS for __API_DOMAIN__ and
      # proxies here. Publishing 0.0.0.0 would expose the tenant API gateway
      # on the VPS public IP, bypassing Nginx and the certificate.
      - "127.0.0.1:__KONG_PORT__:8080"
      - "127.0.0.1:__KONG_ADMIN_PORT__:8001"
    healthcheck:
      test: ["CMD", "kong", "health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    networks:
      - rb_audit_network
    restart: unless-stopped

  # ===========================================================================
  # FRONTEND (Nuxt 4)
  # ===========================================================================
  frontend-__SLUG__:
    build:
      context: ../../../frontend
      dockerfile: Dockerfile
    image: rb_audit_frontend:latest
    container_name: rb_audit___SLUG___frontend
    depends_on:
      - kong-__SLUG__
    ports:
      - "127.0.0.1:__FE_PORT__:3000"
    environment:
      NODE_ENV: production
      HOST: 0.0.0.0
      PORT: "3000"

      # Browser-facing: same-origin /api/v1, proxied by the Nuxt server below.
      # Keeping the browser on one origin means no preflight and no third-party
      # cookie handling for the tenant domain.
      API_BASE_URL: /api/v1
      ANALYTICS_API_BASE_URL: /api/analytics
      PYTHON_AI_BASE_URL: /api/python-ai
      NUXT_PUBLIC_AUTH_SERVICE_BASE_URL: /api/v1
      NUXT_PUBLIC_AUDIT_SERVICE_BASE_URL: /api/v1
      NUXT_PUBLIC_RISK_SERVICE_BASE_URL: /api/v1
      NUXT_PUBLIC_MASTER_SERVICE_BASE_URL: /api/v1

      # Server-side proxy target: this tenant's gateway, by container alias.
      API_BASE_URL_SERVER: http://kong-__SLUG__:8080/api/v1/**
      ANALYTICS_API_BASE_URL_SERVER: http://kong-__SLUG__:8080/api/analytics/**
      PYTHON_AI_BASE_URL_SERVER: http://kong-__SLUG__:8080/api/python-ai/**
    networks:
      - rb_audit_network
    restart: unless-stopped

networks:
  rb_audit_network:
    name: rb_audit_network
    external: true
