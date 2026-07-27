# Database and Docker Compose Setup

This project expects PostgreSQL to be available at:

```text
postgres://sample:sample@localhost:5432/sample?sslmode=disable
```

That connection string means:

| Setting | Value |
| --- | --- |
| Host | `localhost` |
| Port | `5432` |
| User | `sample` |
| Password | `sample` |
| Database | `sample` |
| TLS/SSL | Disabled for local development |

## Choose one PostgreSQL setup

Use either the Docker Compose database or an existing PostgreSQL installation.
Do not run both on port `5432` at the same time.

## Option 1: Docker Compose

Run these commands from the project root:

```bash
docker compose up -d postgres
```

This command:

- Downloads `postgres:17` if it is not already available.
- Creates the `sample-postgres` container.
- Creates the `sample` role and database using the settings in `compose.yaml`.
- Publishes PostgreSQL on `localhost:5432`.
- Runs the container in the background because of `-d`.

Check whether the database is running and healthy:

```bash
docker compose ps
```

Follow its logs:

```bash
docker compose logs -f postgres
```

Open a PostgreSQL shell inside the container:

```bash
docker compose exec postgres psql -U sample -d sample
```

Stop the containers without deleting the database volume:

```bash
docker compose stop
```

Stop and remove the containers and network while preserving the database volume:

```bash
docker compose down
```

Remove the containers **and permanently delete the database volume**:

```bash
docker compose down --volumes
```

Use the last command only when the local development data is no longer needed.

## Option 2: Existing local PostgreSQL

This is the setup currently used on this machine. A native PostgreSQL server was
already listening on port `5432`, but it did not have the role or database
expected by the application.

First, inspect the active PostgreSQL user and check whether the required objects
already exist:

```bash
psql -h localhost -p 5432 -d postgres \
  -c "SELECT current_user, current_database();" \
  -c "SELECT rolname FROM pg_roles WHERE rolname = 'sample';" \
  -c "SELECT datname FROM pg_database WHERE datname = 'sample';"
```

The inspection showed that the current administrative user was `leunghanxi` and
that neither `sample` object existed.

The role and database were then created with:

```bash
psql -h localhost -p 5432 -d postgres -v ON_ERROR_STOP=1 \
  -c "CREATE ROLE sample WITH LOGIN PASSWORD 'sample';" \
  -c "CREATE DATABASE sample OWNER sample;"
```

`ON_ERROR_STOP=1` tells `psql` to stop immediately if a statement fails.

Verify that the application credentials work:

```bash
PGPASSWORD=sample psql \
  -h localhost \
  -p 5432 \
  -U sample \
  -d sample \
  -c "SELECT current_user, current_database();"
```

The expected result is:

```text
 current_user | current_database
--------------+------------------
 sample       | sample
```

## Environment configuration

The default database URL is defined in the backend configuration. It can be
overridden with an environment variable:

```bash
export DATABASE_URL="postgres://sample:sample@localhost:5432/sample?sslmode=disable"
```

The same value is documented in `.env.example`.

If the project later loads a local `.env` file, create it from the example:

```bash
cp .env.example .env
```

Do not commit secrets used outside local development.

## Backend commands

Run the following from the `backend` directory.

Format all Go source files:

```bash
go fmt ./...
```

Run static analysis:

```bash
go vet ./...
```

Run all tests:

```bash
go test ./...
```

Start the API:

```bash
go run ./cmd/api
```

When startup succeeds, the API listens on port `8080`.

Verify database readiness from another terminal:

```bash
curl -i http://localhost:8080/health/ready
```

The expected response is:

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ready"}
```

## Troubleshooting commands used

Check which process is listening on PostgreSQL's port:

```bash
lsof -nP -iTCP:5432 -sTCP:LISTEN
```

Check which process is listening on the API port:

```bash
lsof -nP -iTCP:8080 -sTCP:LISTEN
```

Inspect the Docker Compose configuration and available services:

```bash
docker compose config --services
```

Inspect running Compose containers:

```bash
docker compose ps
```

The original application error was:

```text
FATAL: role "sample" does not exist (SQLSTATE 28000)
```

This meant the application successfully reached PostgreSQL, but that particular
server did not contain the configured login role. Creating the role and database
resolved the problem.

If `go run ./cmd/api` reports:

```text
listen tcp :8080: bind: address already in use
```

another API process is already using port `8080`. Locate it with the `lsof`
command above, or use the already-running process instead of starting a second
copy.
