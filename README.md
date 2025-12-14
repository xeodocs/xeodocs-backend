# XeoDocs Backend

The backend service for the **XeoDocs** translation orchestration system. This application manages projects, languages, file synchronization states, and coordinates the translation workflow between the XeoDocs CLI and the database.

## Architecture

This project follows a **Modular Monolith** architecture in Go. It is structured to separate concerns into distinct modules while running as a single deployable unit.

### Key Modules
- **Auth**: Authentication using JWT (for Admin) and API Keys (for CLI).
- **Projects**: Management of documentation projects (repositories).
- **Languages**: Target languages configuration for each project.
- **Files**: Tracking of file states (`pending`, `translated`, `outdated`) and checksums.
- **Workflow**: Logic for CLI interaction (`next` file to translate, `submit` translations, status checks).
- **Configurations**: System-wide key/value settings.
- **UserPreferences**: Admin-specific UI preferences.
- **Paths**: Management of ignored and special paths for projects.

## Tech Stack

- **Language**: Go 1.23+
- **Router**: [Chi](https://github.com/go-chi/chi) (v5)
- **Database**: PostgreSQL (via `database/sql` + `lib/pq`)
- **Auth**: `golang-jwt/jwt/v5` & `bcrypt`
- **Config**: `joho/godotenv`

## Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL database
- Git

### Configuration

Create a `.env` file in the root directory (copy from `.env.example` if available, or use the reference below):

```bash
PORT=12020
DATABASE_URL=postgres://user:password@localhost:5432/xeodocs?sslmode=disable
JWT_SECRET=your-super-secret-key-change-me
ENVIRONMENT=development
```

### Database Setup

1. **Migrations**
   This project uses [Goose](https://github.com/pressly/goose) for database migrations.

   The application does NOT run migrations automatically on development environments. You must run them manually using `goose`.

   First, install `goose`:
   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

   Or on macOS:

   ```bash
   brew install goose
   ```

   Then, execute the migrations using the `DATABASE_URL` from your local environment file:
   ```bash
   # Load environment variables from .env
   export $(grep -v '^#' .env | xargs)
   
   # Run migrations
   goose -dir migrations postgres "$DATABASE_URL" up
   ```

2. **Create Admin User**
   You must insert the first admin user directly via `psql`. The password must be hashed using bcrypt. Since the migrations enable the `pgcrypto` extension, you can use it to generate the hash.

   ```bash
   # Load environment variables from .env
   export $(grep -v '^#' .env | xargs)
    
   # Run psql command to create admin user
   psql "$DATABASE_URL" -c "
   INSERT INTO users (name, email, password_hash) 
   VALUES (
       'Admin', 
       'admin@xeodocs.com', 
       crypt('admin123', gen_salt('bf'))
   );"
   ```

### Build and Run

1. **Install Dependencies**
   ```bash
   go mod download
   ```

2. **Run the Server**
   ```bash
   go run cmd/api/main.go
   ```

3. **Hot Reload (Development)**
   For a better development experience with hot-reloading, use [Air](https://github.com/air-verse/air).

   Install Air:
   ```bash
   go install github.com/air-verse/air@latest
   ```

   Run with Air:
   ```bash
   air
   ```

4. **Build Binary**
   ```bash
   go build -o bin/api ./cmd/api
   ```

## API Documentation

The API follows the OpenAPI 3.0 specification.
- **Base URL**: `/v1`
- **Health Check**: `GET /v1/health`

### Main Endpoints

- **Auth**: `POST /v1/auth/login`
- **Projects**: `GET /v1/projects`, `POST /v1/projects`, ...
- **Workflow (CLI)**:
  - `GET /v1/projects/{projectId}/languages/{languageId}/next-file`
  - `POST /v1/projects/{projectId}/languages/{languageId}/submissions`
  - `GET /v1/projects/{projectId}/languages/{languageId}/status`

## Project Structure

```
.
├── cmd/
│   └── api/            # Application entrypoint
├── internal/
│   ├── modules/        # Domain modules (auth, projects, etc.)
│   └── shared/         # Shared code (database, config, response)
├── go.mod              # Go module definition
└── README.md           # Project documentation
```

## License

MIT