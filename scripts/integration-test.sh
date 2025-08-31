#!/bin/bash

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# Configuration
readonly ARTIFACTS_DIR="artifacts"
readonly WIREMOCK_URL="http://localhost:8081"
readonly WIREMOCK_RETRIES=10
readonly WIREMOCK_RETRY_DELAY=1

# Global variables
BASE_URL=""
ENV_FILE=""

# Cleanup function
function cleanup() {
  echo -e "${YELLOW}Cleaning up artifacts directory...${NC}"
  rm -rf "$ARTIFACTS_DIR"
}

# Error handling
function error_exit() {
  echo -e "${RED}Error: $1${NC}" >&2
  exit 1
}

# Success message
function success() {
  echo -e "${GREEN}$1${NC}"
}

# Warning message
function warning() {
  echo -e "${YELLOW}$1${NC}"
}

# Load environment configuration
function load_env_config() {
  case "${ENV:-local}" in
    local)
      ENV_FILE=".env"
      BASE_URL="http://localhost:8080"
      ;;
    dev|pre|pro)
      ENV_FILE=".env.${ENV}"
      BASE_URL="http://localhost:8080"
      ;;
    *)
      error_exit "Invalid ENV value. Please set ENV to 'local', 'dev', 'pre', or 'pro'."
      ;;
  esac

  echo "Testing ${ENV:-local} environment: $BASE_URL"

  if [ -f "$ENV_FILE" ]; then
    echo "Loading environment from $ENV_FILE file..."
    # shellcheck disable=SC2046
    export $(grep -v '^#' "$ENV_FILE" | xargs)
  else
    warning "Environment file $ENV_FILE not found. Using defaults."
  fi
}

# Check HTTP status code
function check_status() {
  local actual_status="$1"
  local expected_status="$2"
  local test_name="${3:-HTTP request}"

  if [ "$actual_status" -ne "$expected_status" ]; then
    error_exit "$test_name failed: expected status $expected_status but got $actual_status"
  fi
}

# Check JSON response content
function check_json_response() {
  local response="$1"
  local expected_patterns=("${@:2}")

  for pattern in "${expected_patterns[@]}"; do
    if [[ ! "$response" =~ $pattern ]]; then
      error_exit "JSON response validation failed: expected pattern '$pattern' not found in response"
    fi
  done
}

# Wait for service to be ready
function wait_for_service() {
  local url="$1"
  local service_name="$2"
  local retries="${3:-$WIREMOCK_RETRIES}"
  local delay="${4:-$WIREMOCK_RETRY_DELAY}"

  echo "Waiting for $service_name to be ready..."
  
  for ((i=1; i<=retries; i++)); do
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200"; then
      success "$service_name is ready!"
      return 0
    fi
    
    if [ "$i" -lt "$retries" ]; then
      echo "Attempt $i/$retries: $service_name not ready yet, waiting ${delay}s..."
      sleep "$delay"
    fi
  done

  error_exit "$service_name not ready after ${retries}s"
}

# Test server health endpoint
function test_health_endpoint() {
  echo "Testing server health endpoint..."
  
  local status_code
  status_code=$(curl -s -o "$ARTIFACTS_DIR/health.json" -w "%{http_code}" "$BASE_URL/health")
  
  check_status "$status_code" 200 "Health endpoint"
  
  local health_response
  health_response=$(cat "$ARTIFACTS_DIR/health.json")
  check_json_response "$health_response" "healthy" "cupid-api"
  
  success "Server health test passed!"
}

# Run data sync test case
function run_data_sync_case() {
  local case_id="$1"
  local expected_result="${2:-success}"
  
  echo ""
  echo "----- data-sync case: ${case_id} -----"
  
  if [ "${ENV:-local}" = "local" ]; then
    DB_HOST=localhost HOTEL_ID="$case_id" ./bin/data-sync || true
  else
    HOTEL_ID="$case_id" ./bin/data-sync || true
  fi
  
  echo "Data sync case $case_id completed"
}

# Test data sync functionality
function test_data_sync() {
  echo "Testing data-sync functionality..."
  
  # Wait for WireMock to be ready
  wait_for_service "$WIREMOCK_URL/__admin/mappings" "WireMock"
  
  # Run test cases
  run_data_sync_case "1641879"   # 200 - success case
  run_data_sync_case "bad-1"     # 400 - bad request
  run_data_sync_case "server-1"  # 500 - server error
  
  success "All data-sync tests passed!"
}

# Main function
function main() {
  echo "Starting integration tests..."
  
  # Set up trap for cleanup
  trap cleanup EXIT
  
  # Create artifacts directory
  mkdir -p "$ARTIFACTS_DIR"
  
  # Load environment configuration
  load_env_config
  
  # Run tests
  test_health_endpoint
  test_data_sync
  
  success "All integration tests passed!"
}

# Run main function
main "$@"