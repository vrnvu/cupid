# Cupid

<img width="1150" height="642" alt="Screenshot 2025-09-01 at 20 28 18" src="https://github.com/user-attachments/assets/720027b9-67ed-4359-8665-a0035d7248ae" />


<img width="734" height="768" alt="Screenshot 2025-09-01 at 20 00 01" src="https://github.com/user-attachments/assets/c59fb37c-03bd-4b48-bef0-1f305d2975b4" />

<img width="1173" height="768" alt="Screenshot 2025-09-01 at 20 02 05" src="https://github.com/user-attachments/assets/68b001b2-85fd-45ee-9d89-f3fd31265ae8" />


<img width="1321" height="803" alt="Screenshot 2025-09-01 at 19 58 55" src="https://github.com/user-attachments/assets/6b77b8f2-2ee7-4802-a1af-48708c4a1576" />

<img width="1306" height="831" alt="Screenshot 2025-09-01 at 19 59 17" src="https://github.com/user-attachments/assets/1f64e8ba-b819-4c94-80de-2e8992aeee56" />


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

## Tests

```
➜  server git:(main) ✗ cd .. && make test
        github.com/vrnvu/cupid/cmd/embedding-generator          coverage: 0.0% of statements
        github.com/vrnvu/cupid/cmd/data-sync            coverage: 0.0% of statements
        github.com/vrnvu/cupid/cmd/server               coverage: 0.0% of statements
ok      github.com/vrnvu/cupid/internal/ai      0.184s  coverage: 88.1% of statements
ok      github.com/vrnvu/cupid/internal/cache   0.549s  coverage: 85.7% of statements
ok      github.com/vrnvu/cupid/internal/client  0.535s  coverage: 53.0% of statements
        github.com/vrnvu/cupid/internal/telemetry               coverage: 0.0% of statements
ok      github.com/vrnvu/cupid/internal/database        0.911s  coverage: 77.4% of statements
ok      github.com/vrnvu/cupid/internal/handlers        0.227s  coverage: 42.6% of statements
➜  cupid git:(main) ✗
```

```
➜  cupid git:(main) ✗ make integration-test
Starting comprehensive integration tests...
Testing local environment: http://localhost:8080
Loading environment from .env file...
Populating database with test data...
Waiting for Server to be ready...
Server is ready!
Syncing hotel 1641879...
2025/09/01 19:29:13 Starting sync for hotel 1641879
2025/09/01 19:29:14 Completed sync for hotel 1641879
2025/09/01 19:29:15 Flushing traces to Honeycomb...
Test data populated successfully
Testing health endpoint scenarios...
Health endpoint tests passed!
Testing hotel endpoint scenarios...
Hotel endpoint tests passed!
Testing reviews endpoint scenarios...
Reviews endpoint tests passed!
Testing translations endpoint scenarios...
Translations endpoint tests passed!
Testing error scenarios...
Error scenarios test passed!
Testing response validation...
Response validation tests passed!
Testing basic performance...
Health endpoint response time: 10ms
Hotel endpoint response time: 10ms
Performance tests completed!
Testing data-sync functionality...

----- data-sync case: 1641879 -----
2025/09/01 19:29:17 Starting sync for hotel 1641879
2025/09/01 19:29:18 Completed sync for hotel 1641879
2025/09/01 19:29:18 Flushing traces to Honeycomb...
Data sync case 1641879 completed
All data-sync tests passed!
Testing batch sync functionality...
----- batch sync test -----
Running batch sync without timeout (timeout command not available)
2025/09/01 19:29:19 Starting sync for hotel 1641879
2025/09/01 19:29:20 Completed sync for hotel 1641879
2025/09/01 19:29:20 Flushing traces to Honeycomb...
Batch sync test completed
Batch sync test passed!
All integration tests passed!
Cleaning up artifacts directory...
➜  cupid git:(main) ✗
```

```
➜  cupid git:(main) ✗ make test-ai-integration
=== RUN   TestAIService_Integration
=== PAUSE TestAIService_Integration
=== RUN   TestNewService
=== PAUSE TestNewService
=== RUN   TestGetModelInfo
=== PAUSE TestGetModelInfo
=== RUN   TestGenerateEmbedding_Success
=== PAUSE TestGenerateEmbedding_Success
=== RUN   TestGenerateEmbedding_EmptyText
=== PAUSE TestGenerateEmbedding_EmptyText
=== RUN   TestGenerateEmbedding_WhitespaceOnly
=== PAUSE TestGenerateEmbedding_WhitespaceOnly
=== RUN   TestGenerateEmbeddings_Success
=== PAUSE TestGenerateEmbeddings_Success
=== RUN   TestGenerateEmbeddings_EmptyInput
=== PAUSE TestGenerateEmbeddings_EmptyInput
=== RUN   TestGenerateEmbeddings_FiltersEmptyTexts
=== PAUSE TestGenerateEmbeddings_FiltersEmptyTexts
=== RUN   TestGenerateEmbeddings_APIError
=== PAUSE TestGenerateEmbeddings_APIError
=== RUN   TestGenerateEmbeddings_InvalidJSON
=== PAUSE TestGenerateEmbeddings_InvalidJSON
=== RUN   TestGenerateEmbeddings_MismatchedResponse
=== PAUSE TestGenerateEmbeddings_MismatchedResponse
=== CONT  TestAIService_Integration
=== CONT  TestGenerateEmbeddings_Success
=== NAME  TestAIService_Integration
    integration_test.go:16: OPENAI_API_KEY not set, skipping integration test
=== CONT  TestGenerateEmbedding_Success
--- SKIP: TestAIService_Integration (0.00s)
=== CONT  TestGetModelInfo
--- PASS: TestGetModelInfo (0.00s)
=== CONT  TestNewService
--- PASS: TestNewService (0.00s)
=== CONT  TestGenerateEmbeddings_APIError
=== CONT  TestGenerateEmbedding_WhitespaceOnly
--- PASS: TestGenerateEmbedding_WhitespaceOnly (0.00s)
=== CONT  TestGenerateEmbeddings_FiltersEmptyTexts
=== CONT  TestGenerateEmbedding_EmptyText
--- PASS: TestGenerateEmbedding_EmptyText (0.00s)
=== CONT  TestGenerateEmbeddings_EmptyInput
=== CONT  TestGenerateEmbeddings_MismatchedResponse
--- PASS: TestGenerateEmbeddings_EmptyInput (0.00s)
=== CONT  TestGenerateEmbeddings_InvalidJSON
--- PASS: TestGenerateEmbeddings_APIError (0.00s)
--- PASS: TestGenerateEmbeddings_InvalidJSON (0.00s)
--- PASS: TestGenerateEmbeddings_MismatchedResponse (0.00s)
--- PASS: TestGenerateEmbeddings_FiltersEmptyTexts (0.00s)
--- PASS: TestGenerateEmbedding_Success (0.00s)
--- PASS: TestGenerateEmbeddings_Success (0.00s)
PASS
ok      github.com/vrnvu/cupid/internal/ai      0.328s
➜  cupid git:(main) ✗
```
