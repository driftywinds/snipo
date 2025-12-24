# Development Guide

This document covers building, testing, and contributing to Snipo.

## Prerequisites

- **Go 1.24+**
- **Make** (optional, for convenience commands)
- **Docker** (optional, for containerized builds)

## Building

### From Source

```bash
# Build binary
make build
# Output: ./bin/snipo

# Or directly with Go
go build -o bin/snipo ./cmd/server
```

### Docker Image

```bash
# Build local image
make docker

# Or with docker build
docker build -t snipo:local .
```

## Running

### Development Mode

```bash
# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
make dev

# Or run directly
export SNIPO_MASTER_PASSWORD="dev-password"
export SNIPO_SESSION_SECRET="dev-secret-at-least-32-characters-long" ## Generate it with: "openssl rand -hex 32
go run ./cmd/server serve
```

### With Docker Compose

```bash
# Copy example environment
cp .env.example .env
# Edit .env with your settings

docker compose up -d
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
make coverage ## we have poor coverage right now, you are welcome to improve it

# Run specific package tests
go test -v ./internal/api/handlers/...

# Run with race detection
go test -race ./...
```

## Linting

```bash
# Run linter (requires golangci-lint) - all contributions must pass this
make lint

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Configuration Reference

All configuration is via environment variables. See [`.env.example`](../.env.example) for defaults.

### Core Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_HOST` | `0.0.0.0` | Server bind address |
| `SNIPO_PORT` | `8080` | Server port |
| `SNIPO_DB_PATH` | `./data/snipo.db` | SQLite database path |
| `SNIPO_MASTER_PASSWORD` | **required** | Login password |
| `SNIPO_SESSION_SECRET` | **required** | Session signing key (32+ chars) |
| `SNIPO_SESSION_DURATION` | `168h` | Session lifetime |
| `SNIPO_TRUST_PROXY` | `false` | Trust X-Forwarded-For headers |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_RATE_LIMIT` | `100` | Login requests per window |
| `SNIPO_RATE_WINDOW` | `1m` | Rate limit window duration |
| `SNIPO_RATE_LIMIT_READ` | `1000` | API read operations (per hour) |
| `SNIPO_RATE_LIMIT_WRITE` | `500` | API write operations (per hour) |
| `SNIPO_RATE_LIMIT_ADMIN` | `100` | API admin operations (per hour) |

### API Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_ALLOWED_ORIGINS` | - | CORS allowed origins (comma-separated), use `*` for dev |
| `SNIPO_ENABLE_PUBLIC_SNIPPETS` | `true` | Enable public snippet sharing |
| `SNIPO_ENABLE_API_TOKENS` | `true` | Enable API token creation |
| `SNIPO_ENABLE_BACKUP_RESTORE` | `true` | Enable backup/restore features |

### S3 Backup

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_S3_ENABLED` | `false` | Enable S3 backup |
| `SNIPO_S3_ENDPOINT` | `s3.amazonaws.com` | S3 endpoint URL |
| `SNIPO_S3_ACCESS_KEY` | - | Access key ID |
| `SNIPO_S3_SECRET_KEY` | - | Secret access key |
| `SNIPO_S3_BUCKET` | `snipo-backups` | Bucket name |
| `SNIPO_S3_REGION` | `us-east-1` | AWS region |
| `SNIPO_S3_SSL` | `true` | Use HTTPS |

### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `SNIPO_LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `SNIPO_LOG_FORMAT` | `json` | Log format: json, text |

## Database

Snipo uses SQLite with automatic migrations. The database file is created at `SNIPO_DB_PATH` on first run.

### Migrations

Migrations are embedded in the binary and run automatically on startup. Migration files are in `migrations/`.

### Manual Database Access

```bash
sqlite3 ./data/snipo.db
```

## API Development

The API follows `RESTful` conventions. See [`docs/openapi.yaml`](openapi.yaml) for the complete specification.

### Authentication

API requests require one of:
- Session cookie (from web login) - full admin access
- Bearer token: `Authorization: Bearer <token>`
- API key header: `X-API-Key: <key>`

Create API tokens via Settings → API Tokens in the web UI.

### Token Permissions

API tokens have three permission levels:
- **read**: Can only access GET endpoints (view snippets, tags, folders)
- **write**: Can create, update, and delete snippets, tags, and folders
- **admin**: Full access including token management, settings, and backups

### Rate Limits

API endpoints are rate-limited per token:
- Read operations: 1000 requests/hour (configurable)
- Write operations: 500 requests/hour (configurable)
- Admin operations: 100 requests/hour (configurable)

Rate limit info is included in response headers:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining
- `X-RateLimit-Reset`: Unix timestamp when limit resets
- `Retry-After`: Seconds to wait (when limit exceeded)

### Response Format

All API responses use standardized envelopes:

**Single resource:**
```json
{
  "data": {...},
  "meta": {
    "request_id": "uuid",
    "timestamp": "2024-12-24T10:30:00Z",
    "version": "1.0"
  }
}
```

**List with pagination:**
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "links": {
      "self": "/api/v1/snippets?page=1",
      "next": "/api/v1/snippets?page=2",
      "prev": null
    }
  },
  "meta": {...}
}
```

### Example Requests

```bash
# Create snippet (returns {data: {...}, meta: {...}})
curl -X POST http://localhost:8080/api/v1/snippets \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Example",
    "files": [{"filename": "main.go", "content": "package main", "language": "go"}]
  }'

# List snippets with pagination
curl "http://localhost:8080/api/v1/snippets?page=1&limit=20" \
  -H "Authorization: Bearer TOKEN"

# Search snippets
curl "http://localhost:8080/api/v1/snippets/search?q=example" \
  -H "Authorization: Bearer TOKEN"

# Export backup
curl -o backup.json "http://localhost:8080/api/v1/backup/export" \
  -H "Authorization: Bearer TOKEN"

# Get API documentation
curl http://localhost:8080/api/v1/openapi.json
```

**Note:** All responses are wrapped in envelopes. Access data via `response.data` instead of directly using the response body.

## Releasing

Releases are automated via GitHub Actions when a version tag is pushed.

### Creating a Release

```bash
git checkout main
git pull origin main
git tag v1.0.0
git push origin v1.0.0
```

### Version Format

Follow [Semantic Versioning](https://semver.org/):
- **Major** (`v2.0.0`): Breaking changes (If needed)
- **Minor** (`v1.1.0`): New features, backward compatible
- **Patch** (`v1.0.1`): Bug fixes

### Release Artifacts

Each release includes:
- `snipo_linux_amd64.tar.gz` - Linux x86_64 binary
- `snipo_linux_arm64.tar.gz` - Linux ARM64 binary
- Docker images: `ghcr.io/mohamedelashri/snipo:v1.0.0`, `:v1.0`, `:v1`, `:latest`

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes and add tests
4. Run `make test` and `make lint`
5. Commit with clear messages
6. Open a pull request

### Code Style

- Follow standard Go conventions
- Run `gofmt` before committing
- Keep functions focused and testable
- Add comments for exported functions

## Keyboard Shortcuts (Web UI)

| Shortcut | Action |
|----------|--------|
| `Ctrl+K` / `Cmd+K` | Focus search |
| `Ctrl+N` / `Cmd+N` | New snippet |
| `Escape` | Close editor/modal |
