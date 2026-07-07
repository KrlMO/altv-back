# AGENTS.md

## Running the app

```bash
# Start PostgreSQL
docker compose up -d

# Run server (generates swagger docs first)
make run

# Debug with Delve
make debug

# Generate swagger docs only
make swagger
```

## Entry point

`cmd/api/main.go` - loads `.env` with godotenv, connects PostgreSQL, sets up Gin router.

## Architecture

Domain-driven structure under `internal/domain/{domain}/`:
- `handler.go` - HTTP handlers
- `service.go` - business logic
- `repository.go` - database queries
- `model.go` - domain models
- `dto.go` - request/response DTOs
- `module.go` - route registration

## Database

PostgreSQL via `pgx/v5`. Migrations in `migrations/` folder. Database: `arena_latam`, user: `postgres`, pass: `123456`.

## Key packages

- `pkg/middleware/` - auth, cors, roles
- `pkg/utils/` - helpers (jwt, hash, pointer)
- `internal/domain/identity/` - guards, global roles, constants
- `internal/domain/teams/rules/` - permissions, validation, context

## Testing

No test directory visible. Tests likely colocated or absent.