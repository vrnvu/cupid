# Cupid

Hotel management service with HTTP API and data synchronization.

## Setup

```bash
cp .env.example .env
# Edit .env for local setup (DB_HOST=localhost)
```

## Environment Variables

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

## Development

```bash
make test                    
make build                   
make run-server            
make run-data-sync          
make start-docker           
make integration-test       
```

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/hotels/{hotelID}` - Get hotel by ID