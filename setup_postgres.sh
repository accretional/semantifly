docker pull postgres
docker run --name postgres-container -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 postgres
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres