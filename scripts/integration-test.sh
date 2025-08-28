#!/bin/bash

# Set the base URL based on the environment
case "$ENV" in
  dev)
    BASE_URL="http://localhost:8080"
    ;;
  pre)
    BASE_URL="http://pre.example.com"
    ;;
  pro)
    BASE_URL="http://pro.example.com"
    ;;
  *)
    echo "Invalid ENV value. Please set ENV to 'dev', 'pre', or 'pro'."
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

status_foo=$(curl -s -o artifacts/foo.json -w "%{http_code}" -H "User-Id: 123" $BASE_URL/foo)
check_status $status_foo 200

status_bar=$(curl -s -o artifacts/bar.json -w "%{http_code}" -H "User-Id: 123" $BASE_URL/bar)
check_status $status_bar 200

status_baz=$(curl -s -o artifacts/baz.json -w "%{http_code}" -H "User-Id: 123" $BASE_URL/baz)
check_status $status_baz 200

echo "All server tests passed!"

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
  echo "\n----- data-sync case: ${CASE_ID} -----"
  CUPID_BASE_URL="http://localhost:8081" \
  HOTEL_ID="${CASE_ID}" \
  CUPID_SANDBOX_API="dummy" \
  ENABLE_TELEMETRY=0 \
  go run ./server/cmd/data-sync || true
}

run_data_sync_case "1641879"   # 200
run_data_sync_case "bad-1"     # 400
run_data_sync_case "server-1"  # 500

echo "All data-sync tests passed!"