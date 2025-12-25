#!/bin/bash

# Chaos API Proxy - Example Usage Script

BASE_URL="http://localhost:8080"
API_URL="${BASE_URL}/api/v1"

echo "🌪️  Chaos API Proxy - Example Usage"
echo "===================================="
echo ""

# Check if server is running
echo "1️⃣  Checking server health..."
if ! curl -s "${BASE_URL}/health" > /dev/null; then
    echo "❌ Server is not running! Please start it first:"
    echo "   docker-compose up -d"
    echo "   OR"
    echo "   make run"
    exit 1
fi
echo "✅ Server is healthy"
echo ""

# Create a sample configuration
echo "2️⃣  Creating sample chaos configuration..."
CONFIG_RESPONSE=$(curl -s -X POST "${API_URL}/configs" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "JSONPlaceholder Test",
    "description": "Testing with JSONPlaceholder API",
    "target": "https://jsonplaceholder.typicode.com",
    "enabled": true,
    "rules": {
      "latency_ms": 500,
      "jitter": 200,
      "inject_failure_rate": 0.1,
      "error_code": 503,
      "error_body": "{\"error\": \"Service temporarily unavailable (injected by Chaos Proxy)\"}"
    }
  }')

CONFIG_ID=$(echo "$CONFIG_RESPONSE" | jq -r '.id')
echo "✅ Configuration created with ID: $CONFIG_ID"
echo ""

# List all configurations
echo "3️⃣  Listing all configurations..."
curl -s "${API_URL}/configs" | jq '.'
echo ""

# Make some test requests through the proxy
echo "4️⃣  Making test requests through the proxy..."
echo ""

for i in {1..10}; do
    echo "Request #$i:"
    START=$(date +%s%N)
    
    RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/proxy/${CONFIG_ID}/posts/1")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)
    
    END=$(date +%s%N)
    DURATION=$(( ($END - $START) / 1000000 ))
    
    if [ "$HTTP_CODE" -eq 200 ]; then
        CHAOS_HEADER=$(echo "$BODY" | grep -o "X-Chaos-Proxy")
        echo "  ✅ Success (${DURATION}ms) - HTTP $HTTP_CODE"
        echo "$BODY" | head -n 3
    else
        echo "  ⚠️  Chaos Injected (${DURATION}ms) - HTTP $HTTP_CODE"
        echo "$BODY"
    fi
    echo ""
    sleep 0.5
done

# Get the configuration details
echo "5️⃣  Getting configuration details..."
curl -s "${API_URL}/configs/${CONFIG_ID}" | jq '.'
echo ""

# Update the configuration (disable chaos)
echo "6️⃣  Disabling chaos (setting enabled=false)..."
curl -s -X PUT "${API_URL}/configs/${CONFIG_ID}" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"${CONFIG_ID}\",
    \"name\": \"JSONPlaceholder Test\",
    \"target\": \"https://jsonplaceholder.typicode.com\",
    \"enabled\": false,
    \"rules\": {
      \"latency_ms\": 500,
      \"jitter\": 200,
      \"inject_failure_rate\": 0.1
    }
  }" | jq '.'
echo ""

# Try to use disabled config
echo "7️⃣  Trying to use disabled configuration..."
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/proxy/${CONFIG_ID}/posts/1")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
echo "HTTP Code: $HTTP_CODE (should be 403 Forbidden)"
echo ""

# Clean up - delete the configuration
echo "8️⃣  Cleaning up - deleting configuration..."
curl -s -X DELETE "${API_URL}/configs/${CONFIG_ID}" | jq '.'
echo ""

echo "🎉 Demo completed!"
echo ""
echo "📚 Next steps:"
echo "  - Create your own configurations for your APIs"
echo "  - Experiment with different chaos parameters"
echo "  - Check the README.md for more examples"
echo "  - Monitor your application's resilience!"
