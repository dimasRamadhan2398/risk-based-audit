#!/bin/bash

# Target configuration
VPS_IP="202.10.34.166"
VPS_USER="root"
TARGET_DIR="/app/rbia-frontend"

# Exit on error
set -e

echo "=== Packaging frontend (excluding node_modules & build dirs) ==="
tar --exclude='node_modules' --exclude='.nuxt' --exclude='.output' --exclude='.git' --exclude='dist' --exclude='frontend.tar.gz' -czf frontend.tar.gz .

echo "=== Uploading archive to VPS ($VPS_IP) ==="
ssh -o StrictHostKeyChecking=no $VPS_USER@$VPS_IP "mkdir -p $TARGET_DIR"
scp -o StrictHostKeyChecking=no frontend.tar.gz $VPS_USER@$VPS_IP:$TARGET_DIR/

echo "=== Extracting and building on VPS ==="
ssh -o StrictHostKeyChecking=no $VPS_USER@$VPS_IP "cd $TARGET_DIR && \
  tar -xzf frontend.tar.gz && \
  rm -f frontend.tar.gz && \
  echo 'Creating .env file...' && \
  echo \"API_BASE_URL=http://$VPS_IP:8080/api/v1\" > .env && \
  echo \"ANALYTICS_API_BASE_URL=http://$VPS_IP:8080/api/analytics\" >> .env && \
  echo \"NUXT_PUBLIC_AUTH_SERVICE_BASE_URL=http://$VPS_IP:8080/api/v1\" >> .env && \
  echo \"NUXT_PUBLIC_AUDIT_SERVICE_BASE_URL=http://$VPS_IP:8080/api/v1\" >> .env && \
  echo \"NUXT_PUBLIC_RISK_SERVICE_BASE_URL=http://$VPS_IP:8080/api/v1\" >> .env && \
  echo \"NUXT_PUBLIC_MASTER_SERVICE_BASE_URL=http://$VPS_IP:8080/api/v1\" >> .env && \
  echo 'Running docker-compose up...' && \
  docker compose -f docker-compose.prod.yml up -d --build"

# Clean up local archive
rm -f frontend.tar.gz

echo "=== Deployment Completed Successfully! ==="
echo "You can access the website at http://$VPS_IP:3000"
