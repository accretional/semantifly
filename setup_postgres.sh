# Attempt to pull the postgres Docker image with a loading animation
echo -n "Pulling postgres image "
while ! docker pull postgres > /dev/null 2>&1; do
    for s in / - \\ \|; do
        printf "\r%s" "$s"
        sleep .1
    done
done
printf "\r"
echo "Postgres image pulled successfully"
echo "Starting postgres server "

# Check if 'nc' command is available to determine which method to use for checking postgres availability
if ! command -v nc > /dev/null 2>&1; then
    if ! docker inspect postgres-container > /dev/null 2>&1; then

        # Start a new postgres container if it doesn't exist
        docker run --name postgres-container -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 postgres > /dev/null 2>&1 || { echo "Error starting postgres container: $?"; exit 1; }
        export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
        
        while ! docker exec postgres-container pg_isready > /dev/null 2>&1; do
            for s in / - \\ \|; do
                printf "\r%s" "$s"
                sleep .1
            done
        done
        printf "\r"
        echo "Postgres server has started"
    else
        echo "Postgres container already exists"
    fi
else
    # If 'nc' is available, use it to check if postgres is running on port 5432
    if ! nc -z localhost 5432 > /dev/null 2>&1; then
        # Start a new postgres container if it's not running
        docker run --name postgres-container -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 postgres > /dev/null 2>&1 || { echo "Error starting postgres container: $?"; exit 1; }
        export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
        
        while ! docker exec postgres-container pg_isready > /dev/null 2>&1; do
            for s in / - \\ \|; do
                printf "\r%s" "$s"
                sleep .1
            done
        done
        printf "\r"
        echo -e "\nPostgres server has started"
    else
        echo "Postgres server is already running on port 5432"
    fi
fi