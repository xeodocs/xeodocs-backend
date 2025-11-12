# XeoDocs Backend

This repository contains the backend for XeoDocs, an AI-driven translation system for open-source technology documentation on xeodocs.com. The backend is built as a Golang monorepo with microservices architecture.

## Overview

XeoDocs automatically translates documentation from English to other languages using AI, maintaining up-to-date versions through periodic git pulls and builds.

## Architecture

- Microservices: Auth, Project, Translation, Build, Logging, Scheduler
- API Gateway for routing
- PostgreSQL for data storage
- MinIO for file storage
- Docker for containerization

## Setup

1. Clone the repository
2. Run `docker compose -f docker-compose.dev.yml up --build` for development
3. For production, use `docker-compose.prod.yml`
4. Stop the containers with `docker compose -f docker-compose.dev.yml down --volumes --rmi local`

## Services

- Gateway: API routing and middleware
- Auth: User authentication and authorization
- Project: Repository management
- Translation: AI-powered translations
- Build: Static site generation
- Logging: Centralized logging
- Scheduler: Periodic tasks

## Testing

### E2E Tests

End-to-End tests validate the full workflow through HTTP requests to the API gateway. They ensure integration across services without involving the frontend.

#### Running E2E Tests

1. Start a brand new development stack:
   ```bash
   docker-compose -f docker-compose.dev.yml down --volumes --rmi local
   docker-compose -f docker-compose.dev.yml up --build -d
   ```

2. Run the E2E tests:
   ```bash
   go test ./tests/e2e -v
   ```

3. Stop the stack:
   ```bash
   docker-compose -f docker-compose.dev.yml down --volumes --rmi local
   ```

#### E2E Test Coverage

- **Auth Service**: Login, password change, user CRUD (create, read, update, delete), role CRUD
- Tests simulate real client interactions via the gateway at `http://localhost:12020/v1`
- Assumes empty DB with tables created and default admin user seeded

### Unit Tests

Run unit tests for individual packages:
```bash
go test ./...
```

## Temporary Container Access (Docker)

You can create a temporary TCP proxy to expose a container port (e.g., 5432 for PostgreSQL) from your host machine to the private Docker network without modifying `docker-compose.yml` or restarting your services. 

This allows tools like pgAdmin (running on your host) to connect to `localhost:5432` as if the port were directly exposed.

The approach uses a lightweight temporary container (based on Alpine Linux with `socat`) that joins your existing `xeodocs-internal-net` network. It listens on the host's port 5432 and forwards traffic to the `db` service (resolving via Docker's internal DNS). When you're done, just stop the container—it auto-removes with `--rm`.

### Steps
1. **Network name**: The network name is `xeodocs-internal-net`.

2. **Start the temporary proxy**:
   ```
   docker run --rm -d \
     --network xeodocs-internal-net \
     -p 5432:5432 \
     alpine/socat \
     TCP-LISTEN:5432,fork \
     TCP:db:5432
   ```
   - `--rm`: Auto-deletes the container when stopped.
   - `-d`: Runs in the background.
   - `-p 5432:5432`: Maps host port 5432 to the container's 5432.
   - `TCP-LISTEN:5432,fork TCP:db:5432`: Forwards incoming connections to the `db` container on port 5432.
   - If port 5432 is already in use on your host, change it (e.g., `-p 5433:5432` and connect pgAdmin to `localhost:5433`).

3. **Connect with pgAdmin**:
   - Host: `localhost` (or `127.0.0.1`)
   - Port: `5432` (or whatever you mapped)
   - Database: `xeodocs_db`
   - Username: `user` 
   - Password: `password` 

4. **Verify the connection** (optional):
   - Check the proxy is running: `docker ps` (look for the socat container).
   - Test from host: `docker run --rm --network none postgres:15 psql -h localhost -p 5432 -U user -d xeodocs_db` (it should connect and show the DB prompt).

5. **Clean up**: Stop the proxy with `docker stop <container_id>` (from `docker ps`), or just kill it—it'll self-remove.

### Notes
- **Security**: This exposes the port only while the proxy runs, so it's truly temporary. Use it in a secure environment or behind a firewall.
- **Conflicts**: If your host already uses 5432 (e.g., local Postgres), pick a different port.
- This works with your existing setup since the proxy joins the same network and resolves `db` via DNS.
