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
2. Run `docker compose -f docker-compose.dev.yml up` for development
3. For production, use `docker-compose.prod.yml`

## Services

- Gateway: API routing and middleware
- Auth: User authentication and authorization
- Project: Repository management
- Translation: AI-powered translations
- Build: Static site generation
- Logging: Centralized logging
- Scheduler: Periodic tasks

## Contributing

Follow the monorepo structure. Use shared packages in `internal/shared/`.

## License

MIT License
