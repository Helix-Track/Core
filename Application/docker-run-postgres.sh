#!/bin/bash
# Run HelixTrack Core with PostgreSQL database

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "============================================"
echo "HelixTrack Core - PostgreSQL Configuration"
echo "============================================"
echo ""

# Check if .env.postgres exists
if [ ! -f ".env.postgres" ]; then
    echo "Error: .env.postgres not found"
    echo "Please copy .env.example to .env.postgres and configure"
    exit 1
fi

# Create necessary directories
mkdir -p logs Database

# Create PostgreSQL initialization script if it doesn't exist
if [ ! -f "Database/init-postgres.sql" ]; then
    echo "Creating PostgreSQL initialization script..."
    cat > Database/init-postgres.sql <<'EOF'
-- HelixTrack Core - PostgreSQL Initialization Script

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set timezone
SET timezone = 'UTC';

-- Create schema
CREATE SCHEMA IF NOT EXISTS helixtrack;

-- Grant permissions
GRANT ALL ON SCHEMA helixtrack TO helixtrack;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'HelixTrack Core database initialized successfully';
END $$;
EOF
fi

# Build and start services
echo "Building and starting services..."
docker-compose -f docker-compose.postgres.yml up --build -d

echo ""
echo "Services started successfully!"
echo ""
echo "Core API:     http://localhost:8080"
echo "Chat API:     http://localhost:9090"
echo "PostgreSQL:   localhost:5432 (Core)"
echo "Chat DB:      localhost:5433"
echo "Mock Auth:    http://localhost:8081"
echo "Mock Perm:    http://localhost:8082"
echo ""
echo "Optional services (use --profile):"
echo "  pgAdmin:    http://localhost:5050 (--profile admin-tools)"
echo "  Prometheus: http://localhost:9091 (--profile monitoring)"
echo "  Grafana:    http://localhost:3000 (--profile monitoring)"
echo ""
echo "To view logs: docker-compose -f docker-compose.postgres.yml logs -f"
echo "To stop:      docker-compose -f docker-compose.postgres.yml down"
echo ""

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
timeout=60
counter=0

while [ $counter -lt $timeout ]; do
    if docker exec helixtrack-postgres pg_isready -U helixtrack -d helixtrack > /dev/null 2>&1; then
        echo "✓ PostgreSQL is ready!"
        break
    fi
    sleep 1
    counter=$((counter + 1))
    echo -n "."
done

if [ $counter -eq $timeout ]; then
    echo "✗ PostgreSQL did not become ready within ${timeout}s"
    echo "Check logs: docker-compose -f docker-compose.postgres.yml logs postgres"
    exit 1
fi

# Wait for service to be healthy
echo "Waiting for Core service to be healthy..."
timeout=60
counter=0

while [ $counter -lt $timeout ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✓ Service is healthy!"
        break
    fi
    sleep 1
    counter=$((counter + 1))
    echo -n "."
done

if [ $counter -eq $timeout ]; then
    echo "✗ Service did not become healthy within ${timeout}s"
    echo "Check logs: docker-compose -f docker-compose.postgres.yml logs helixtrack-core"
    exit 1
fi

# Wait for Chat service to be healthy
echo ""
echo "Waiting for Chat service to be healthy..."
timeout=60
counter=0

while [ $counter -lt $timeout ]; do
    if curl -s http://localhost:9090/health > /dev/null 2>&1; then
        echo "✓ Chat service is healthy!"
        break
    fi
    sleep 1
    counter=$((counter + 1))
    echo -n "."
done

if [ $counter -eq $timeout ]; then
    echo "✗ Chat service did not become healthy within ${timeout}s"
    echo "Check logs: docker-compose -f docker-compose.postgres.yml logs chat-service"
    # Don't exit - allow Core to run even if Chat fails
fi

echo ""
echo "✓ HelixTrack (PostgreSQL) is ready!"
echo ""
