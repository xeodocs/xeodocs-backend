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
