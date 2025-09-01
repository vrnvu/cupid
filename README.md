# Cupid

Hotel management service with HTTP API and data synchronization.

## Design

### Key Technical Decisions

**Database Architecture**
- Designed normalized PostgreSQL schema with 12 tables to handle complex hotel data (rooms, photos, facilities, policies)
- Added strategic indexes on hotel_id, entity lookups, and rating fields - queries run in <10ms
- All data operations wrapped in transactions with proper rollback handling to maintain data consistency
- Implemented upsert operations with `ON CONFLICT` to handle idempotent data synchronization from external API
- Configured connection pooling for 25 max connections to handle concurrent requests

**Error Handling & Resilience**
- Built graceful degradation: Redis cache failures don't break the application, continues with database fallback
- All operations use context for cancellation and timeouts (15s for HTTP, 5s for database)
- Created custom error types (ErrHotelNotFound) for proper error handling instead of generic errors
- Implemented transaction rollback with defer cleanup patterns to prevent resource leaks

**Caching Strategy**
- Added Redis cache for frequently accessed review data with 5-minute TTL to reduce database load
- Implemented cache-aside pattern with database fallback when cache is unavailable
- Used proper cache key management (`reviews:hotel:{id}`) with expiration strategies

**Testing Approach**
- All tests run in parallel using `t.Parallel()` - test suite completes in under 30 seconds
- Used random IDs in tests to avoid conflicts between parallel test runs 
- Built three testing levels: unit tests with mocks, integration tests with real database, end-to-end API tests
- Added tests for concurrent database access and race conditions to ensure thread safety

**API Design**
- Created RESTful endpoints with proper HTTP status codes (200, 400, 404, 405, 500)
- Generated complete OpenAPI specification with request/response schemas for all endpoints
- Implemented input validation with meaningful error messages for better developer experience
- Added pagination support for list endpoints (limit/offset with max 100 items per page)
- Built multi-language support for English, French, and Spanish translations

**Data Synchronization**
- Implemented batch processing for 100 hotels with configurable hotel lists
- Used incremental updates with upsert operations to handle both new and updated data efficiently
- Made individual hotel sync failures non-blocking - entire batch continues even if some hotels fail
- Added rate limiting with 100ms delays between API calls to respect external API limits
  - Circuit breaker could be added for extra observability 
- Created separate sync processes for content, reviews, and translations to handle different data types

**Observability**
- Integrated OpenTelemetry with HoneyComb for distributed tracing in production
- Added health checks with database connectivity validation for monitoring
- Implemented structured logging throughout the application for debugging

**Deployment**
- Built multi-stage Docker containers with Alpine Linux for minimal attack surface (final image <50MB)
- Created Docker Compose setup with health checks and networking for local development
- Implemented environment-based configuration for seamless local/dev/pre/pro deployments
- Added Makefile with comprehensive build, test, and deployment commands

### Workflows
- `ci.yml`: for build, test, linting and integration testing. I build it simulating multiple dev/pre/pro environments to showcase how we re-use the integration tests, we only change the env vars.
- `data-sync.yml`: it explains a data sync strategy

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
- `GET /api/v1/hotels/{hotelID}/reviews` - Get hotel reviews
- `GET /api/v1/hotels/{hotelID}/translations/{language}` - Get hotel translations (supported languages: fr, es, en)