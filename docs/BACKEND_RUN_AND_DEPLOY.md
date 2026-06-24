# Backend Running and Deployment Guide

This document provides instructions on how to run the backend services locally and how to deploy them to a server.

## 🚀 Running Locally

### Prerequisites
- [Go 1.21+](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started) and [Docker Compose](https://docs.docker.com/compose/install/)
- [PostgreSQL 16](https://www.postgresql.org/)
- [Redis 7](https://redis.io/)
- [Kafka](https://kafka.apache.org/) (Confluent OSS)

### Using Docker Compose (Recommended)
The easiest way to run the entire backend infrastructure and all services is using Docker Compose.

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```

2. Start all services:
   ```bash
   docker-compose up -d
   ```
   This will start PostgreSQL, Redis, Zookeeper, Kafka, Kafka-UI, and all the microservices (`auth`, `audit`, `master`, `risk`, `analytics`).

3. Verify running containers:
   ```bash
   docker-compose ps
   ```

4. View logs:
   ```bash
   docker-compose logs -f [service_name]
   ```

### Running Individual Services (Development Mode)
If you want to run a specific service for development while keeping the infrastructure (DB, Redis, etc.) in Docker:

1. Start infrastructure only:
   ```bash
   cd backend
   ```
   ```bash
   docker-compose up -d postgres redis kafka zookeeper
   ```

2. Navigate to the service directory (e.g., `auth-service`):
   ```bash
   cd auth-service
   ```

3. Run migrations:
   ```bash
   go run ./cmd migrate up
   ```

4. Start the service:
   ```bash
   go run ./cmd serve
   ```

---

## 🛳️ Deployment to Server

### Server Information
- **IP Address**: `202.10.34.160`
- **Username**: `root`
- **Password**: `<SERVER_PASSWORD>` (See secure storage)

### Manual Deployment Steps

1. **SSH into the Server**:
   ```bash
   ssh root@202.10.34.160
   ```

2. **Prepare Environment**:
   Ensure Docker and Docker Compose are installed on the server.
   ```bash
   # Install Docker (if not installed)
   curl -fsSL https://get.docker.com -o get-docker.sh
   sh get-docker.sh
   ```

3. **Deploy the Code**:
   Clone the repository or pull the latest changes:
   ```bash
   git clone <repository_url>
   cd <repository_name>/backend
   ```

4. **Configuration**:
   Copy the example configuration and adjust for production:
   ```bash
   cp auth-service/pkg/config/config.docker.yaml auth-service/pkg/config/config.yaml
   # Repeat for other services
   ```

5. **Start Services**:
   ```bash
   docker-compose up -d --build
   ```

---

## 🤖 Automatic Deployment (CI/CD)

To automate the deployment process, you can use GitHub Actions. This allows for continuous integration and continuous deployment whenever code is pushed to the repository.

### Workflow Concept
1. **Build & Test**: Run tests on every push.
2. **Dockerize**: Build Docker images for each service.
3. **Push**: Push images to a Container Registry (e.g., Docker Hub, GitHub Container Registry).
4. **Deploy**: SSH into the server, pull the latest images, and restart containers.

### Example GitHub Actions Workflow (`.github/workflows/deploy.yml`)

```yaml
name: Deploy Backend

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build and push Auth Service
        uses: docker/build-push-action@v5
        with:
          context: ./backend/auth-service
          push: true
          tags: your-dockerhub-user/rb-audit-auth:latest

      # Repeat Build and push for other services...

      - name: Deploy to Server via SSH
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: 202.10.34.160
          username: root
          password: ${{ secrets.SERVER_PASSWORD }}
          script: |
            cd /path/to/your/project/backend
            docker compose pull
            docker compose up -d
            docker image prune -f
```

### Best Practices for Automatic Deployment
- **Use Secrets**: Never hardcode passwords or tokens in your YAML files. Use GitHub Repository Secrets.
- **Environment Variables**: Use `.env` files or environment variables in Docker Compose to manage environment-specific configurations.
- **Rollbacks**: Consider using tags (e.g., commit SHA) instead of `latest` for better versioning and easier rollbacks.
- **Health Checks**: Implement health check endpoints in your services to ensure they are running correctly after deployment.
