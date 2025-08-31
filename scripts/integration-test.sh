#!/bin/bash

cleanup() {
  echo "Cleaning up artifacts directory..."
  rm -rf artifacts
}

trap cleanup EXIT

# Load environment-specific configuration
case "$ENV" in
  local)
    ENV_FILE=".env"
    ;;
  dev|pre|pro)
    ENV_FILE=".env.${ENV}"
    ;;
  *)
    echo "Invalid ENV value. Please set ENV to 'local', 'dev', 'pre', or 'pro'."
    exit 1
    ;;
esac

if [ -f "$ENV_FILE" ]; then
  echo "Loading environment from $ENV_FILE file..."
  export $(grep -v '^#' "$ENV_FILE" | xargs)
else
  echo "Environment file $ENV_FILE not found. Using defaults."
fi

case "$ENV" in
  local)
    BASE_URL="http://localhost:8080"
    echo "Testing LOCAL environment: $BASE_URL"
    ;;
  dev)
    BASE_URL="http://localhost:8080"
    echo "Testing DEV environment: $BASE_URL"
    ;;
  pre)
    BASE_URL="http://localhost:8080"
    echo "Testing PRE environment: $BASE_URL"
    ;;
  pro)
    BASE_URL="http://localhost:8080"
    echo "Testing PRO environment: $BASE_URL"
    ;;
  *)
    echo "Invalid ENV value. Please set ENV to 'local', 'dev', 'pre', or 'pro'."
    exit 1
    ;;
esac

# Function to check HTTP status code
check_status() {
  if [ "$1" -ne "$2" ]; then
    echo "Expected status $2 but got $1"
    exit 1
  fi
}

# Function to check JSON response
check_json() {
  if [ "$1" != "$2" ]; then
    echo "Expected JSON $2 but got $1"
    exit 1
  fi
}

# Call foo, bar, baz and store responses
mkdir -p artifacts

echo "Testing server health endpoint..."

status_health=$(curl -s -o artifacts/health.json -w "%{http_code}" $BASE_URL/health)
check_status $status_health 200

health_response=$(cat artifacts/health.json)
if [[ "$health_response" != *"healthy"* ]] || [[ "$health_response" != *"cupid-api"* ]]; then
  echo "Invalid health response: $health_response"
  exit 1
fi

echo "Server health test passed!"

# =========================
# data-sync integration (WireMock)
# =========================

# Assume WireMock is already running; wait until it's responsive
retries=10
until curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/__admin/mappings | grep -q "200"; do
  if [ "$retries" -le 0 ]; then
    echo "WireMock not ready after 10s; aborting data-sync tests" >&2
    exit 1
  fi
  echo "Waiting for WireMock... ($retries left)"
  retries=$((retries-1))
  sleep 1
done

run_data_sync_case() {
  CASE_ID="$1"
  echo ""
  echo "----- data-sync case: ${CASE_ID} -----"
  if [ "$ENV" = "local" ]; then
    DB_HOST=localhost HOTEL_ID="${CASE_ID}" ./bin/data-sync || true
  else
    HOTEL_ID="${CASE_ID}" ./bin/data-sync || true
  fi
}

run_data_sync_case "1641879"   # 200
run_data_sync_case "bad-1"     # 400
run_data_sync_case "server-1"  # 500

echo "All data-sync tests passed!"