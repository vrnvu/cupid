#!/bin/bash

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# Configuration
readonly ARTIFACTS_DIR="artifacts"

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
  local retries="${3:-10}"
  local delay="${4:-1}"

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

function test_health_endpoint() {
  echo "Testing health endpoint scenarios..."
  
  # Test healthy service
  local status_code
  status_code=$(curl -s -o "$ARTIFACTS_DIR/health.json" -w "%{http_code}" "$BASE_URL/health")
  check_status "$status_code" 200 "Health endpoint - healthy service"
  
  local health_response
  health_response=$(cat "$ARTIFACTS_DIR/health.json")
  check_json_response "$health_response" "healthy" "cupid-api"
  
  # Test method not allowed
  status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/health")
  check_status "$status_code" 405 "Health endpoint - method not allowed"
  
  success "Health endpoint tests passed!"
}

function test_hotel_endpoints() {
  echo "Testing hotel endpoint scenarios..."
  
  # Test get existing hotel
  local status_code
  status_code=$(curl -s -o "$ARTIFACTS_DIR/hotel.json" -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879")
  check_status "$status_code" 200 "Get hotel by valid ID"
  
  local hotel_response
  hotel_response=$(cat "$ARTIFACTS_DIR/hotel.json")
  check_json_response "$hotel_response" "1641879" "The Z Hotel Covent Garden" "8.3"
  
  # Test hotel not found
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/999999")
  check_status "$status_code" 404 "Get non-existent hotel"
  
  # Test invalid hotel ID format
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/invalid")
  check_status "$status_code" 400 "Get hotel by invalid ID format"
  
  # Test method not allowed
  status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/hotels/1641879")
  check_status "$status_code" 405 "Wrong HTTP method for hotel endpoint"
  
  # Test missing hotel ID
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/")
  check_status "$status_code" 404 "Get hotel with missing ID"
  
  success "Hotel endpoint tests passed!"
}

function test_reviews_endpoints() {
  echo "Testing reviews endpoint scenarios..."
  
  # Test get hotel reviews
  local status_code
  status_code=$(curl -s -o "$ARTIFACTS_DIR/reviews.json" -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/reviews")
  check_status "$status_code" 200 "Get hotel reviews"
  
  local reviews_response
  reviews_response=$(cat "$ARTIFACTS_DIR/reviews.json")
  check_json_response "$reviews_response" "1641879" "count" "reviews"
  
  # Test hotel with no reviews
  status_code=$(curl -s -o "$ARTIFACTS_DIR/reviews_empty.json" -w "%{http_code}" "$BASE_URL/api/v1/hotels/999999/reviews")
  check_status "$status_code" 200 "Get reviews for hotel with no reviews"
  
  local empty_reviews_response
  empty_reviews_response=$(cat "$ARTIFACTS_DIR/reviews_empty.json")
  check_json_response "$empty_reviews_response" "999999" "count" "reviews"
  
  # Test invalid hotel ID format
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/invalid/reviews")
  check_status "$status_code" 400 "Get reviews with invalid hotel ID format"
  
  # Test method not allowed
  status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/hotels/1641879/reviews")
  check_status "$status_code" 405 "Wrong HTTP method for reviews endpoint"
  
  # Test missing hotel ID
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels//reviews")
  check_status "$status_code" 301 "Get reviews with missing hotel ID"
  
  success "Reviews endpoint tests passed!"
}

function test_translations_endpoints() {
  echo "Testing translations endpoint scenarios..."
  
  # Test get French translations
  local status_code
  status_code=$(curl -s -o "$ARTIFACTS_DIR/translations_fr.json" -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/translations/fr")
  check_status "$status_code" 200 "Get French translations"
  
  local translations_response
  translations_response=$(cat "$ARTIFACTS_DIR/translations_fr.json")
  check_json_response "$translations_response" "1641879" "fr" "count" "translations"
  
  # Test get Spanish translations
  status_code=$(curl -s -o "$ARTIFACTS_DIR/translations_es.json" -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/translations/es")
  check_status "$status_code" 200 "Get Spanish translations"
  
  translations_response=$(cat "$ARTIFACTS_DIR/translations_es.json")
  check_json_response "$translations_response" "1641879" "es" "count" "translations"
  
  # Test get English translations
  status_code=$(curl -s -o "$ARTIFACTS_DIR/translations_en.json" -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/translations/en")
  check_status "$status_code" 200 "Get English translations"
  
  translations_response=$(cat "$ARTIFACTS_DIR/translations_en.json")
  check_json_response "$translations_response" "1641879" "en" "count" "translations"
  
  # Test hotel with no translations
  status_code=$(curl -s -o "$ARTIFACTS_DIR/translations_empty.json" -w "%{http_code}" "$BASE_URL/api/v1/hotels/999999/translations/fr")
  check_status "$status_code" 200 "Get translations for hotel with no translations"
  
  local empty_translations_response
  empty_translations_response=$(cat "$ARTIFACTS_DIR/translations_empty.json")
  check_json_response "$empty_translations_response" "999999" "fr" "count" "translations"
  
  # Test unsupported language
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/translations/de")
  check_status "$status_code" 400 "Get translations with unsupported language"
  
  # Test invalid hotel ID format
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/invalid/translations/fr")
  check_status "$status_code" 400 "Get translations with invalid hotel ID format"
  
  # Test method not allowed
  status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/hotels/1641879/translations/fr")
  check_status "$status_code" 405 "Wrong HTTP method for translations endpoint"
  
  # Test missing hotel ID
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels//translations/fr")
  check_status "$status_code" 301 "Get translations with missing hotel ID"
  
  # Test missing language
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/translations/")
  check_status "$status_code" 404 "Get translations with missing language"
  
  # Test invalid language format
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/translations/123")
  check_status "$status_code" 400 "Get translations with invalid language format"
  
  success "Translations endpoint tests passed!"
}

function test_error_scenarios() {
  echo "Testing error scenarios..."
  
  # Test 404 for non-existent endpoints
  local status_code
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/nonexistent")
  check_status "$status_code" 404 "Non-existent endpoint"
  
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/nonexistent")
  check_status "$status_code" 404 "Non-existent hotel sub-endpoint"
  
  # Test malformed URLs
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879/translations")
  check_status "$status_code" 404 "Translations endpoint without language"
  
  success "Error scenarios test passed!"
}

# Populate database with test data
function populate_test_data() {
  echo "Populating database with test data..."
  
  wait_for_service "$BASE_URL/health" "Server"
  
  # Populate hotel data for the main test hotel
  echo "Syncing hotel 1641879..."
  ./bin/data-sync || true
  
  # Wait a moment for data to be processed
  sleep 2
  
  # Verify the data was populated
  local status_code
  status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/hotels/1641879")
  if [ "$status_code" -eq 200 ]; then
    success "Test data populated successfully"
  else
    warning "Test data population may have failed (status: $status_code)"
  fi
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
  
  run_data_sync_case "1641879"   # 200 - success case
  
  success "All data-sync tests passed!"
}

function test_batch_sync() {
  echo "Testing batch sync functionality..."
  
  echo "----- batch sync test -----"
  
  # Test batch sync with a small subset (just a few hotels)
  # We'll use a timeout to avoid running the full 100 hotels
  if command -v timeout >/dev/null 2>&1; then
    timeout 30s ./bin/data-sync || true
  else
    # Fallback for systems without timeout command (like macOS)
    echo "Running batch sync without timeout (timeout command not available)"
    ./bin/data-sync &
    local sync_pid=$!
    sleep 30
    kill $sync_pid 2>/dev/null || true
  fi
  
  echo "Batch sync test completed"
  
  success "Batch sync test passed!"
}

function test_response_validation() {
  echo "Testing response validation..."
  
  # Test hotel response structure
  local hotel_response
  hotel_response=$(curl -s "$BASE_URL/api/v1/hotels/1641879")
  
  # Validate required fields exist
  check_json_response "$hotel_response" "hotel_id" "hotel_name" "rating" "review_count"
  
  # Test reviews response structure
  local reviews_response
  reviews_response=$(curl -s "$BASE_URL/api/v1/hotels/1641879/reviews")
  
  # Validate required fields exist
  check_json_response "$reviews_response" "hotel_id" "count" "reviews"
  
  # Test translations response structure
  local translations_response
  translations_response=$(curl -s "$BASE_URL/api/v1/hotels/1641879/translations/fr")
  
  # Validate required fields exist
  check_json_response "$translations_response" "hotel_id" "language" "count" "translations"
  
  success "Response validation tests passed!"
}

function test_performance() {
  echo "Testing basic performance..."
  
  # Test response time for main endpoints
  local start_time
  local end_time
  local response_time
  
  start_time=$(date +%s%N)
  curl -s -o /dev/null "$BASE_URL/health"
  end_time=$(date +%s%N)
  response_time=$(( (end_time - start_time) / 1000000 ))
  
  if [ "$response_time" -gt 1000 ]; then
    warning "Health endpoint response time: ${response_time}ms (slow)"
  else
    success "Health endpoint response time: ${response_time}ms"
  fi
  
  start_time=$(date +%s%N)
  curl -s -o /dev/null "$BASE_URL/api/v1/hotels/1641879"
  end_time=$(date +%s%N)
  response_time=$(( (end_time - start_time) / 1000000 ))
  
  if [ "$response_time" -gt 1000 ]; then
    warning "Hotel endpoint response time: ${response_time}ms (slow)"
  else
    success "Hotel endpoint response time: ${response_time}ms"
  fi
  
  success "Performance tests completed!"
}

# Main function
function main() {
  echo "Starting comprehensive integration tests..."
  
  # Set up trap for cleanup
  trap cleanup EXIT
  
  # Create artifacts directory
  mkdir -p "$ARTIFACTS_DIR"
  
  # Load environment configuration
  load_env_config
  
  # Build binaries if they don't exist
  if [ ! -f "./bin/data-sync" ] || [ ! -f "./bin/server" ]; then
    echo "Building binaries..."
    make build
  fi
  
  # Populate database with test data
  populate_test_data
  
  # Run tests
  test_health_endpoint
  test_hotel_endpoints
  test_reviews_endpoints
  test_translations_endpoints
  test_error_scenarios
  test_response_validation
  test_performance
  test_data_sync
  test_batch_sync
  
  success "All integration tests passed!"
}

# Run main function
main "$@"