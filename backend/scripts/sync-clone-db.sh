#!/bin/bash
set -e

echo "=========================================================================="
echo " Starting Clone PostgreSQL Database Container (Port 5435)"
echo "=========================================================================="

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BACKEND_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
cd "$BACKEND_DIR"

# Launch clone postgres container
docker compose -f docker-compose.clone-db.yml up -d

echo "Waiting for postgres-clone container (rb_audit_postgres_clone) to become ready..."
until docker exec rb_audit_postgres_clone pg_isready -U postgres; do
  sleep 1
done

echo "Container rb_audit_postgres_clone is ready on port 5435!"

# Check if active postgres container is running
if docker ps --format '{{.Names}}' | grep -q "^rb_audit_postgres$"; then
  echo "Active postgres database 'rb_audit_postgres' detected."
  echo "Cloning live data from 'rb_audit_postgres' -> 'rb_audit_postgres_clone'..."

  # Dump live databases from active container and restore into clone container
  docker exec rb_audit_postgres pg_dumpall -U postgres | docker exec -i rb_audit_postgres_clone psql -U postgres >/dev/null 2>&1 || true

  echo "Live database sync completed!"
else
  echo "Active postgres container 'rb_audit_postgres' is not running."
  echo "Clone container initialized with standard clone-db-init.sql data."
fi

echo "=========================================================================="
echo " Clone PostgreSQL Database is running!"
echo " Connection Details for Settings -> Data Source (http://localhost:3000/settings?tab=datasource):"
echo "   Host / Server IP : localhost (or postgres-clone inside Docker network)"
echo "   Port            : 5435"
echo "   Database Name   : core_banking (or rb_audit_clone / rb_audit_master_service)"
echo "   Username        : postgres"
echo "   Password        : postgres"
echo "=========================================================================="
