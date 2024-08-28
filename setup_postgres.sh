#!/bin/bash
set -e

trap 'echo "An error occurred. Exiting..." >&2; exit 1' ERR

# Function to check if Postgres is ready
check_postgres() {
    docker exec -i postgres-container psql -U postgres -c "SELECT 1" > /dev/null 2>&1 || return 1
}

# Function to start Postgres container and install postgres-protobuf
start_postgres() {
    if ! docker run --name postgres-container -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 postgres:14; then
        echo "Error starting postgres container" >&2
        exit 1
    fi
    export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
    
    echo "Waiting for Postgres to start..."
    while ! docker exec postgres-container pg_isready; do
        sleep 1
    done
    echo "Postgres server has started"

    # Install build dependencies and postgres-protobuf
    echo "Installing postgres-protobuf in the Postgres container..."
    if ! docker exec postgres-container bash -c "
        apt-get update && \
        apt-get install -y git make gcc postgresql-server-dev-14 libprotobuf-c-dev protobuf-c-compiler curl && \
        git clone https://github.com/mpartel/postgres-protobuf && \
        cd postgres-protobuf && \
        make && \
        make install && \
        echo 'shared_preload_libraries = '\''protobuf'\''' >> /var/lib/postgresql/data/postgresql.conf
    "; then
        echo "Failed to install postgres-protobuf in the Postgres container" >&2
        exit 1
    fi
    echo "postgres-protobuf installed successfully in the Postgres container"

    # Restart Postgres to load the new shared library
    echo "Restarting Postgres to load postgres-protobuf..."
    if ! docker restart postgres-container; then
        echo "Failed to restart Postgres container" >&2
        exit 1
    fi
    
    # Wait for Postgres to be ready again
    echo "Waiting for Postgres to restart..."
    while ! docker exec postgres-container pg_isready; do
        sleep 1
    done
    echo "Postgres server has restarted"

    # Enable the extension in the database
    echo "Enabling postgres-protobuf extension..."
    if ! docker exec -i postgres-container psql -U postgres -c "CREATE EXTENSION IF NOT EXISTS postgres_protobuf;"; then
        echo "Failed to enable postgres-protobuf extension" >&2
        exit 1
    fi
    echo "postgres-protobuf extension enabled successfully"
}

# Attempt to pull the postgres Docker image
echo "Pulling postgres image..."
if ! docker pull postgres:14; then
    echo "Failed to pull postgres image" >&2
    exit 1
fi
echo "Postgres image pulled successfully"

# Check if 'nc' command is available to determine which method to use for checking postgres availability
if ! command -v nc &> /dev/null; then
    if ! docker inspect postgres-container &> /dev/null || ! docker container inspect -f '{{.State.Running}}' postgres-container | grep -q true; then
        # Start a new postgres container if it doesn't exist or isn't running
        start_postgres
    else
        echo "Postgres container already exists and is running"
    fi
else
    # If 'nc' is available, use it to check if postgres is running on port 5432
    if ! nc -z localhost 5432; then
        # Start a new postgres container if it's not running
        start_postgres
    else
        echo "Postgres server is already running on port 5432"
    fi
fi

# Final check to ensure Postgres is fully ready
echo "Waiting for Postgres to be fully ready..."
if ! check_postgres; then
    echo "Failed to connect to Postgres" >&2
    exit 1
fi
echo "Postgres is now fully ready to accept connections"