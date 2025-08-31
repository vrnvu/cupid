# Cupid

Hotel management service with HTTP API and data synchronization.

## Project Structure

### Server Application
- `server/` - Main Go application
  - `cmd/server/` - HTTP API server entry point
  - `cmd/data-sync/` - Data synchronization tool that fetches hotel data from Cupid API
  - `internal/` - Application internals:
    - `client/` - HTTP client for Cupid API calls
    - `database/` - Database connection and repository layer
    - `handlers/` - HTTP request handlers
    - `telemetry/` - OpenTelemetry configuration with HoneyComb

### Scripts and Testing
- `scripts/integration-test.sh` - Integration test script that tests both server endpoints and data-sync functionality. This is re-used in DEV/PRE and PRO environments.
- `wiremock/` - Mock server for testing external API calls

### Database
- `migrations/` - Database schema migrations

## Getting Started

```bash
make start-docker
make test
make integration-test
```

## Environment Setup

Copy the example environment file and configure it:
```bash
cp .env.example .env
```

### Environment Variables
- `PORT` - Server port (8080)
- `ENABLE_TELEMETRY` - OpenTelemetry (0)
- `DB_HOST` - Database host (localhost)
- `DB_PORT` - Database port (5432)
- `DB_USER` - Database user (cupid)
- `DB_PASSWORD` - Database password (cupid123)
- `DB_NAME` - Database name (cupid)
- `DB_SSLMODE` - Database SSL mode (disable)
- `CUPID_BASE_URL` - Cupid API base URL
- `CUPID_SANDBOX_API` - Cupid API key
- `HOTEL_ID` - Default hotel ID for data sync
- `ENV` - Environment (dev/pre/pro)

## Make Commands

- `make test` - Run unit tests
- `make build` - Build both server and data-sync binaries
- `make run-server` - Start the HTTP server
- `make run-data-sync` - Run data synchronization
- `make start-docker` - Start PostgreSQL database with Docker
- `make integration-test` - Run integration tests against local environment
- `make lint` - Run code linting

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/hotels/{hotelID}` - Get hotel by ID