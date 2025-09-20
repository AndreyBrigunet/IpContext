#!/bin/bash

# IP API Performance Benchmark Script

echo "🚀 IP API Performance Benchmark"
echo "================================"

BASE_URL="http://localhost:3280"
TEST_IPS=("8.8.8.8" "1.1.1.1" "208.67.222.222" "9.9.9.9" "76.76.19.19")

# Check if service is running
echo "📡 Checking service availability..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo "❌ Service is not running at $BASE_URL"
    echo "Start the service with: docker compose up -d"
    exit 1
fi
echo "✅ Service is running"
echo ""

# Test health endpoint performance
echo "🏥 Health Check Performance:"
echo "----------------------------"
for i in {1..5}; do
    curl -w "Response time: %{time_total}s\n" -s "$BASE_URL/health" -o /dev/null
done
echo ""

# Test IP lookup performance (first run - cache miss)
echo "🔍 IP Lookup Performance (Cache Miss):"
echo "--------------------------------------"
for ip in "${TEST_IPS[@]}"; do
    echo "Testing $ip:"
    curl -w "  Response time: %{time_total}s\n" -s "$BASE_URL/$ip" -o /dev/null
done
echo ""

# Test IP lookup performance (second run - cache hit)
echo "⚡ IP Lookup Performance (Cache Hit):"
echo "------------------------------------"
for ip in "${TEST_IPS[@]}"; do
    echo "Testing $ip:"
    curl -w "  Response time: %{time_total}s\n" -s "$BASE_URL/$ip" -o /dev/null
done
echo ""

# Load test with concurrent requests
echo "🔥 Concurrent Load Test (100 requests, 10 concurrent):"
echo "-----------------------------------------------------"
if command -v ab &> /dev/null; then
    ab -n 100 -c 10 -q "$BASE_URL/8.8.8.8" | grep -E "(Requests per second|Time per request|Transfer rate)"
else
    echo "⚠️  Apache Bench (ab) not installed. Install with:"
    echo "   Ubuntu/Debian: sudo apt-get install apache2-utils"
    echo "   macOS: brew install apache2"
    echo ""
    echo "Running simple concurrent test instead..."
    
    # Simple concurrent test
    start_time=$(date +%s.%N)
    for i in {1..20}; do
        curl -s "$BASE_URL/8.8.8.8" > /dev/null &
    done
    wait
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc -l)
    rps=$(echo "scale=2; 20 / $duration" | bc -l)
    echo "20 concurrent requests completed in ${duration}s"
    echo "Requests per second: $rps"
fi

echo ""
echo "🎯 Performance Summary:"
echo "======================"
echo "• Health check should be < 0.5ms"
echo "• Cache miss should be 2-5ms"
echo "• Cache hit should be < 1ms"
echo "• Service should handle 1000+ req/s"
echo ""
echo "💡 Tips for better performance:"
echo "• Increase CACHE_TTL_MINUTES for better hit ratio"
echo "• Set LOG_LEVEL=warn in production"
echo "• Use LOG_FORMAT=json for structured logging"
