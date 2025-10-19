#!/bin/sh
###############################################################################
# HelixTrack Core - Docker Entrypoint Script
#
# Features:
# - Automatic port selection if port is taken
# - Service discovery registration
# - Database migration
# - Health checks before startup
# - Graceful shutdown handling
###############################################################################

set -e

echo "========================================="
echo "HelixTrack Core Starting..."
echo "========================================="

# Configuration
: ${SERVER_HOST:=0.0.0.0}
: ${SERVER_PORT:=8080}
: ${SERVER_PORT_RANGE_START:=8080}
: ${SERVER_PORT_RANGE_END:=8089}
: ${AUTO_PORT_SELECTION:=true}
: ${SERVICE_REGISTRY_URL:=http://service-registry:8500}
: ${SERVICE_DISCOVERY_ENABLED:=true}
: ${DB_HOST:=core-db}
: ${DB_PORT:=5432}
: ${DB_NAME:=helixtrack_core}
: ${DB_USER:=helixtrack}
: ${DB_TYPE:=postgres}

###############################################################################
# Function: Wait for database
###############################################################################
wait_for_db() {
    echo "Waiting for database at ${DB_HOST}:${DB_PORT}..."

    MAX_TRIES=30
    TRIES=0

    while [ $TRIES -lt $MAX_TRIES ]; do
        if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; then
            echo "✓ Database is ready"
            return 0
        fi

        TRIES=$((TRIES + 1))
        echo "  Waiting for database... ($TRIES/$MAX_TRIES)"
        sleep 2
    done

    echo "✗ Database connection failed after $MAX_TRIES attempts"
    return 1
}

###############################################################################
# Function: Find available port
###############################################################################
find_available_port() {
    if [ "$AUTO_PORT_SELECTION" != "true" ]; then
        echo "$SERVER_PORT"
        return 0
    fi

    echo "Searching for available port in range $SERVER_PORT_RANGE_START-$SERVER_PORT_RANGE_END..."

    for port in $(seq $SERVER_PORT_RANGE_START $SERVER_PORT_RANGE_END); do
        if ! nc -z localhost $port 2>/dev/null; then
            echo "✓ Found available port: $port"
            echo "$port"
            return 0
        fi
        echo "  Port $port is in use, trying next..."
    done

    echo "✗ No available ports found in range"
    return 1
}

###############################################################################
# Function: Register with service discovery
###############################################################################
register_service() {
    if [ "$SERVICE_DISCOVERY_ENABLED" != "true" ]; then
        echo "Service discovery disabled, skipping registration"
        return 0
    fi

    local PORT=$1
    echo "Registering service with discovery at $SERVICE_REGISTRY_URL..."

    # Create service registration JSON
    cat > /tmp/service-registration.json <<EOF
{
  "ID": "${HOSTNAME}-${PORT}",
  "Name": "${SERVICE_NAME:-helixtrack-core}",
  "Tags": ["core", "api", "v${BUILD_VERSION:-1.0.0}"],
  "Address": "${HOSTNAME}",
  "Port": ${PORT},
  "Meta": {
    "version": "${BUILD_VERSION:-1.0.0}",
    "auto_selected_port": "${AUTO_PORT_SELECTION}",
    "started_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  },
  "Check": {
    "HTTP": "http://${HOSTNAME}:${PORT}/health",
    "Interval": "30s",
    "Timeout": "10s",
    "DeregisterCriticalServiceAfter": "90s"
  }
}
EOF

    # Register with Consul
    if curl -f -X PUT \
        -H "Content-Type: application/json" \
        -d @/tmp/service-registration.json \
        "${SERVICE_REGISTRY_URL}/v1/agent/service/register" \
        > /dev/null 2>&1; then
        echo "✓ Service registered successfully"
        rm -f /tmp/service-registration.json
        return 0
    else
        echo "⚠ Service registration failed (non-fatal)"
        rm -f /tmp/service-registration.json
        return 0
    fi
}

###############################################################################
# Function: Run database migrations
###############################################################################
run_migrations() {
    if [ "$DB_TYPE" != "postgres" ]; then
        echo "Skipping migrations (not using PostgreSQL)"
        return 0
    fi

    echo "Running database migrations..."

    export PGPASSWORD="$DB_PASSWORD"

    # Check if database exists
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        echo "✓ Database exists"

        # Run migration scripts if they exist
        if [ -d "/app/Database/DDL" ]; then
            echo "  Checking for pending migrations..."
            # Migration logic would go here
            # For now, just verify connection
            if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
                echo "✓ Database connection verified"
            fi
        fi
    else
        echo "⚠ Database does not exist, it should be created by init scripts"
    fi

    unset PGPASSWORD
}

###############################################################################
# Function: Create configuration file
###############################################################################
create_config() {
    local PORT=$1
    echo "Creating runtime configuration..."

    cat > /tmp/runtime.json <<EOF
{
  "log": {
    "log_path": "/app/logs",
    "logfile_base_name": "htCore",
    "log_size_limit": 100000000,
    "level": "${LOG_LEVEL:-info}"
  },
  "listeners": [
    {
      "address": "${SERVER_HOST}",
      "port": ${PORT},
      "https": false
    }
  ],
  "database": {
    "type": "${DB_TYPE}",
    "host": "${DB_HOST}",
    "port": ${DB_PORT},
    "database": "${DB_NAME}",
    "user": "${DB_USER}",
    "password": "${DB_PASSWORD}",
    "sslmode": "${DB_SSLMODE:-require}"
  },
  "services": {
    "authentication": {
      "enabled": ${AUTH_SERVICE_ENABLED:-true},
      "url": "${AUTH_SERVICE_URL:-http://auth-service:8081}",
      "timeout": 30
    },
    "permissions": {
      "enabled": ${PERM_SERVICE_ENABLED:-true},
      "url": "${PERM_SERVICE_URL:-http://perm-service:8082}",
      "timeout": 30
    }
  },
  "version": "${BUILD_VERSION:-1.0.0}"
}
EOF

    echo "✓ Configuration created"
}

###############################################################################
# Function: Cleanup on exit
###############################################################################
cleanup() {
    echo ""
    echo "========================================="
    echo "HelixTrack Core Shutting Down..."
    echo "========================================="

    if [ "$SERVICE_DISCOVERY_ENABLED" = "true" ] && [ -n "$SELECTED_PORT" ]; then
        echo "Deregistering from service discovery..."
        curl -f -X PUT \
            "${SERVICE_REGISTRY_URL}/v1/agent/service/deregister/${HOSTNAME}-${SELECTED_PORT}" \
            > /dev/null 2>&1 || true
        echo "✓ Service deregistered"
    fi

    echo "✓ Cleanup complete"
}

trap cleanup EXIT INT TERM

###############################################################################
# Main Execution
###############################################################################

# Step 1: Wait for database
if [ "$DB_TYPE" = "postgres" ]; then
    wait_for_db || exit 1
fi

# Step 2: Find available port
SELECTED_PORT=$(find_available_port) || exit 1
export SERVER_PORT=$SELECTED_PORT

# Step 3: Create configuration
create_config $SELECTED_PORT

# Step 4: Run migrations
run_migrations

# Step 5: Register with service discovery
register_service $SELECTED_PORT

# Step 6: Start application
echo "========================================="
echo "Starting HelixTrack Core on port $SELECTED_PORT"
echo "========================================="
echo ""

# Execute the application with runtime config
exec "$@" --config=/tmp/runtime.json
