#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="$PROJECT_DIR/bin/linux/amd64/doorman-keepassxc"
CONTAINER_NAME="doorman-keepassxc-test-$$"
IMAGE="docker.io/library/ubuntu:24.04"

pass=0
fail=0

cleanup() {
    echo "--- Cleaning up container $CONTAINER_NAME"
    podman rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
}
trap cleanup EXIT

assert_eq() {
    local label="$1" expected="$2" actual="$3"
    if [[ "$expected" == "$actual" ]]; then
        echo "  PASS: $label"
        pass=$((pass + 1))
    else
        echo "  FAIL: $label (expected '$expected', got '$actual')"
        fail=$((fail + 1))
    fi
}

if [[ ! -x "$BINARY" ]]; then
    echo "Binary not found at $BINARY — run 'make build-linux' first."
    exit 2
fi

echo "=== Starting integration test ==="

echo "--- Starting container and installing keepassxc-cli"
podman run -d \
    --name "$CONTAINER_NAME" \
    -v "$BINARY:/usr/local/bin/doorman-keepassxc:ro,z" \
    "$IMAGE" \
    sleep infinity

podman exec "$CONTAINER_NAME" bash -c "
    apt-get update -qq && apt-get install -y -qq keepassxc >/dev/null 2>&1
"

TEST_PASSWORD="testmaster123"
TEST_DB="/tmp/test.kdbx"

echo "--- Creating test KeePassXC database"
podman exec "$CONTAINER_NAME" bash -c "
    printf '%s\n%s\n' '${TEST_PASSWORD}' '${TEST_PASSWORD}' | keepassxc-cli db-create -p '${TEST_DB}'
"

echo "--- Creating groups and adding test entries"
podman exec "$CONTAINER_NAME" bash -c "
    echo '${TEST_PASSWORD}' | keepassxc-cli mkdir '${TEST_DB}' myapp
    echo '${TEST_PASSWORD}' | keepassxc-cli mkdir '${TEST_DB}' service
"
# Add entry with -p (password prompt): stdin = master password + entry password
podman exec "$CONTAINER_NAME" bash -c "
    printf '%s\n%s\n' '${TEST_PASSWORD}' 'supersecret42' | keepassxc-cli add -p '${TEST_DB}' myapp/db-password -u admin
"
podman exec "$CONTAINER_NAME" bash -c "
    printf '%s\n%s\n' '${TEST_PASSWORD}' 'api-key-value-99' | keepassxc-cli add -p '${TEST_DB}' service/api-key -u service-account
"

# Test 1: info command
echo "--- Test: info command"
info_output=$(podman exec "$CONTAINER_NAME" doorman-keepassxc info)
if echo "$info_output" | grep -q '"doorman-keepassxc"'; then
    echo "  PASS: info command"
    pass=$((pass + 1))
else
    echo "  FAIL: info command"
    fail=$((fail + 1))
fi

# Test 2: get a secret
echo "--- Test: get myapp/db-password"
secret=$(podman exec -e KEEPASSXC_DB="$TEST_DB" -e KEEPASSXC_PASSWORD="$TEST_PASSWORD" \
    "$CONTAINER_NAME" doorman-keepassxc get "myapp/db-password" 2>/dev/null || echo "GET_ERROR")
assert_eq "get myapp/db-password" "supersecret42" "$secret"

# Test 3: get another secret
echo "--- Test: get service/api-key"
secret2=$(podman exec -e KEEPASSXC_DB="$TEST_DB" -e KEEPASSXC_PASSWORD="$TEST_PASSWORD" \
    "$CONTAINER_NAME" doorman-keepassxc get "service/api-key" 2>/dev/null || echo "GET_ERROR")
assert_eq "get service/api-key" "api-key-value-99" "$secret2"

# Test 4: missing entry returns non-zero
echo "--- Test: missing entry"
set +e
podman exec -e KEEPASSXC_DB="$TEST_DB" -e KEEPASSXC_PASSWORD="$TEST_PASSWORD" \
    "$CONTAINER_NAME" doorman-keepassxc get "nonexistent/entry" >/dev/null 2>&1
exit_code=$?
set -e
if [[ "$exit_code" -ne 0 ]]; then
    echo "  PASS: missing entry exits non-zero ($exit_code)"
    pass=$((pass + 1))
else
    echo "  FAIL: missing entry should exit non-zero"
    fail=$((fail + 1))
fi

# Test 5: missing KEEPASSXC_DB exits non-zero
echo "--- Test: missing KEEPASSXC_DB"
set +e
podman exec "$CONTAINER_NAME" doorman-keepassxc get "anything" >/dev/null 2>&1
exit_code=$?
set -e
if [[ "$exit_code" -ne 0 ]]; then
    echo "  PASS: missing KEEPASSXC_DB exits non-zero ($exit_code)"
    pass=$((pass + 1))
else
    echo "  FAIL: missing KEEPASSXC_DB should exit non-zero"
    fail=$((fail + 1))
fi

echo ""
echo "=== Results: $pass passed, $fail failed ==="
if [[ $fail -gt 0 ]]; then
    exit 1
fi
