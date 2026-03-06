#!/usr/bin/env bash
set -e

echo "====================================="
echo "Running PostQ-Tunnel E2E Protocol Test"
echo "====================================="

export QTUN_DOMAIN=""
export GOROOT=$(go env GOROOT)

cleanup() {
    echo "Cleaning up processes..."
    kill $AGENT_PID 2>/dev/null || true
    kill $SERVER_PID 2>/dev/null || true
    kill $HTTP_PID 2>/dev/null || true
    rm -rf /tmp/qtun_test_dir
}
trap cleanup EXIT

echo "[1/5] Building binaries..."
go build -o bin/qtunnel ./cmd/server/main.go
go build -o bin/qtun ./cmd/qtun/main.go

echo "[2/5] Starting dummy backend service on port 8081..."
mkdir -p /tmp/qtun_test_dir
echo "E2E_SUCCESS_PAYLOAD" > /tmp/qtun_test_dir/index.html
cd /tmp/qtun_test_dir && python3 -m http.server 8081 &
HTTP_PID=$!
cd - > /dev/null

sleep 1

echo "[3/5] Starting Edge Server (qtunnel) on :8080 (http) and :4443 (pqc)..."
./bin/qtunnel &
SERVER_PID=$!

sleep 2

echo "[4/5] Starting Edge Agent (qtun) for subdomain 'e2e'..."
./bin/qtun -server localhost:4443 -sub e2e -local localhost:8081 &
AGENT_PID=$!

sleep 3

echo "[5/5] Sending test HTTP request routing through tunnel..."
# The proxy extracts subdomain from the Host header (e.g. e2e.localhost)
RESPONSE=$(curl -s -f -H "Host: e2e.localhost" http://localhost:8080/ || echo "CURL_FAILED")

if echo "$RESPONSE" | grep -q "E2E_SUCCESS_PAYLOAD"; then
    echo "✅ E2E Test Passed: Tunnel correctly routed and multiplexed the request."
    exit 0
else
    echo "❌ E2E Test Failed! Unexpected response:"
    echo "$RESPONSE"
    exit 1
fi
