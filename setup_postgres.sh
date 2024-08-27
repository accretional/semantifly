#!/usr/bin/env bash

set -euo pipefail

pull_postgres_image() {
    echo -n "Pulling postgres image "
    while ! docker pull postgres &> /dev/null; do
        for s in / - \\ \|; do
            printf "\r%s" "$s"
            sleep .1
        done
    done
    printf "\r"
    echo "Postgres image pulled successfully"
}

start_postgres_container() {
    docker run --name postgres-container -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 postgres &> /dev/null || { echo "Error starting postgres container: $?"; exit 1; }
    export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
    
    echo "Starting postgres server "
    while ! docker exec postgres-container pg_isready &> /dev/null; do
        for s in / - \\ \|; do
            printf "\r%s" "$s"
            sleep .1
        done
    done
    printf "\r"
    echo "Postgres server has started"
}

check_postgres_availability() {
    if command -v nc &> /dev/null; then
        if ! nc -z localhost 5432 &> /dev/null; then
            start_postgres_container
        else
            echo "Postgres server is already running on port 5432"
        fi
    else
        if ! docker inspect postgres-container &> /dev/null; then
            start_postgres_container
        else
            echo "Postgres container already exists"
        fi
    fi
}

install_pgvector() {
    echo "Installing pgvector"

    docker exec postgres-container apt-get update >/dev/null 2>&1
    docker exec postgres-container apt-get install -y git build-essential postgresql-server-dev-16 >/dev/null 2>&1
    
    docker exec postgres-container bash -c "cd /tmp && git clone --branch v0.7.4 https://github.com/pgvector/pgvector.git" >/dev/null 2>&1
    
    docker exec postgres-container bash -c "cd /tmp/pgvector && make && make install" >/dev/null 2>&1
    
    docker exec postgres-container psql -U postgres -c "CREATE EXTENSION vector;" >/dev/null 2>&1
}

main() {
    pull_postgres_image
    check_postgres_availability
    install_pgvector
}

main