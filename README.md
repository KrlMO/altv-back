# Arena API

Backend API for Arena Latam - a team management and gaming platform.

## Requirements

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 16 (via Docker)

## Quick Start

```bash
# Copy environment variables
cp .env.example .env

# Start database
docker compose up -d postgres

# Run the server (generates swagger docs first)
make run

# Or run everything with Docker
make docker-up
```

The API will be available at `http://localhost:8080`

## Development

```bash
# Run with hot reload (requires air)
make run

# Run tests
make test

# Run tests with coverage
make test-cover

# Generate swagger docs
make swagger

# Build binary
make build

# Run with Delve debugger
make debug
```

## API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/health` | Health check | No |
| GET | `/swagger/*` | Swagger UI | No |
| POST | `/api/v1/auth/register` | Register user | No |
| POST | `/api/v1/auth/login` | Login | No |
| GET | `/api/v1/countries/` | List countries | No |
| GET | `/api/v1/users/:id` | Get user profile | No |
| GET | `/api/v1/teams/` | List teams | No |
| GET | `/api/v1/teams/:id` | Get team | No |
| POST | `/api/v1/teams/` | Create team | Yes |
| PATCH | `/api/v1/teams/:id` | Update team | Yes |
| POST | `/api/v1/teams/:id/members` | Add member | Yes |
| PATCH | `/api/v1/teams/:id/members/:userId` | Update member role | Yes |
| DELETE | `/api/v1/teams/:id/members/:userId` | Remove member | Yes |
| POST | `/api/v1/teams/:id/games` | Create team game | Yes |
| GET | `/api/v1/team-games/:id/members` | List game members | No |
| POST | `/api/v1/team-games/:id/members` | Add game members | Yes |
| DELETE | `/api/v1/team-games/:id/members` | Remove game members | Yes |
| POST | `/api/v1/team-games/:id/leave` | Leave team game | Yes |

## Project Structure

```
cmd/api/          - Entry point
internal/
  config/         - Configuration loader
  database/       - PostgreSQL connection
  domain/         - Business logic (DDD)
    auth/         - Authentication & authorization
    users/        - User management
    teams/        - Team management
    team_games/   - Team game management
    countries/    - Countries catalog
    identity/     - Role constants & guards
  infrastructure/ - External services (mail)
  router/         - Route wiring
pkg/
  middleware/      - HTTP middleware (auth, cors)
  response/       - Generic response types
  utils/          - Helpers (jwt, hash)
migrations/       - Database migrations
```

## Environment Variables

See `.env.example` for all required variables.

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `JWT_SECRET` | Yes | Secret for JWT signing |
| `SERVER_PORT` | No | Server port (default: 8080) |
| `CORS_ORIGIN` | No | Allowed CORS origin |
| `FRONTEND_URL` | No | Frontend URL for redirects |

## License

MIT
