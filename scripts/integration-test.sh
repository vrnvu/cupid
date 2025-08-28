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

echo "All tests passed!"