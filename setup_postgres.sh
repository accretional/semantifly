#!/bin/bash
set -e

trap 'echo "An error occurred. Exiting..." >&2; exit 1' ERR

# Function to check if Postgres is ready
check_postgres() {
    docker exec -i postgres-container psql -U postgres -c "SELECT 1" > /dev/null 2>&1 || return 1
}

# Function to start Postgres container
start_postgres() {
    if ! docker run --name postgres-container -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 postgres; then
        echo "Error starting postgres container" >&2
        exit 1
    fi
    export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
    
    echo "Waiting for Postgres to start..."
    while ! docker exec postgres-container pg_isready; do
        sleep 1
    done
    echo "Postgres server has started"
}

# Attempt to pull the postgres Docker image
echo "Pulling postgres image..."
if ! docker pull postgres; then
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