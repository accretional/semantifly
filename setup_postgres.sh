#!/bin/bash
set -e

trap 'echo "An error occurred. Exiting..." >&2; exit 1' ERR

CONTAINER_NAME="postgres-container"
POSTGRES_PORT=5432

# Function to check if Postgres is ready
check_postgres() {
    docker exec -i $CONTAINER_NAME psql -U postgres -c "SELECT 1" > /dev/null 2>&1 || return 1
}

# Function to remove existing container if it exists
cleanup_existing_container() {
    if docker ps -a --format '{{.Names}}' | grep -q "^$CONTAINER_NAME$"; then
        echo "Removing existing container..."
        docker rm -f $CONTAINER_NAME
    fi
}

# Function to start Postgres container
start_postgres() {
    cleanup_existing_container
    if ! docker run --name $CONTAINER_NAME -e POSTGRES_PASSWORD=postgres -d -p $POSTGRES_PORT:5432 postgres:14; then
        echo "Error starting postgres container" >&2
        exit 1
    fi
    export DATABASE_URL=postgres://postgres:postgres@localhost:$POSTGRES_PORT/postgres
    
    echo "Waiting for Postgres to start..."
    while ! docker exec $CONTAINER_NAME pg_isready; do
        sleep 1
    done
    echo "Postgres server has started"
}

# Attempt to pull the postgres Docker image
echo "Pulling postgres image..."
if ! docker pull postgres:14; then
    echo "Failed to pull postgres image" >&2
    exit 1
fi
echo "Postgres image pulled successfully"

# Check if the container exists and is running
if docker ps --format '{{.Names}}' | grep -q "^$CONTAINER_NAME$"; then
    echo "Postgres container is already running"
else
    echo "Starting new Postgres container..."
    start_postgres
fi

# Final check to ensure Postgres is fully ready
echo "Performing final checks"
if ! check_postgres; then
    echo "Failed to connect to Postgres" >&2
    exit 1
fi
echo "Postgres is now ready to accept connections"