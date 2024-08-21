#!/bin/bash

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if a port is in use
port_in_use() {
    lsof -i :$1 >/dev/null 2>&1
}

# Check if PostgreSQL is installed
if ! command_exists psql; then
    echo "PostgreSQL is not installed. Installing now..."
    {
        apt-get update
        DEBIAN_FRONTEND=noninteractive apt-get install -y postgresql postgresql-contrib
    } &>/dev/null &

    pid=$! # Process ID of the background installation
    spin='-\|/'
    i=0
    while kill -0 $pid 2>/dev/null
    do
        i=$(( (i+1) %4 ))
        printf "\r${spin:$i:1}"
        sleep .1
    done
    printf "\r"

    if command_exists psql; then
        echo "PostgreSQL installed successfully."
    else
        echo "Failed to install PostgreSQL. Please install it manually."
        exit 1
    fi
else
    echo "PostgreSQL is already installed."
fi

# PostgreSQL data directory
PGDATA="${PGDATA:-/workspace/.pgsql/data}"

# Check if PostgreSQL is already running
if pg_isready >/dev/null 2>&1; then
    echo "PostgreSQL is already running."
else
    echo "Starting PostgreSQL..."
    
    # Ensure the PostgreSQL data directory exists
    mkdir -p "$PGDATA" &>/dev/null

    # Initialize the database if it doesn't exist
    if [ ! -f "$PGDATA/PG_VERSION" ]; then
        echo "Initializing PostgreSQL database..."
        initdb -D "$PGDATA" &>/dev/null
    fi

    # Start PostgreSQL
    sudo service postgresql start &>/dev/null

    # Wait for PostgreSQL to start
    while ! pg_isready >/dev/null 2>&1; do
        echo -n "."
        sleep 1
    done
    echo # New line after dots

    echo "PostgreSQL started successfully"
fi

# Set up port forwarding for PostgreSQL
if ! port_in_use 5432; then
    gp ports create 5432 auto >/dev/null 2>&1
    echo "Port forwarding set up for PostgreSQL on port 5432"
else
    echo "Port 5432 is already in use. PostgreSQL is probably already running."
fi

echo "PostgreSQL is now running in the background. Use 'sudo service postgresql stop' to stop PostgreSQL."