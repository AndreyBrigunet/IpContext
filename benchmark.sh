#!/bin/bash

set -euo pipefail

readonly BASE_URL="${API_URL:-http://localhost:3280}"
readonly TEST_IPS=("8.8.8.8" "1.1.1.1" "208.67.222.222" "9.9.9.9" "76.76.19.19")
readonly HEALTH_TESTS=5
readonly LOAD_REQUESTS=100
readonly LOAD_CONCURRENCY=10

error() {
    echo "[ERROR] $*" >&2
    exit 1
}

check_dependencies() {
    command -v curl >/dev/null || error "curl is required"
}

check_service() {
    echo "Checking service availability..."
    if ! curl -sf "$BASE_URL/health" >/dev/null 2>&1; then
        error "Service not available at $BASE_URL"
    fi
    echo "Service is running"
}

benchmark_health() {
    echo "Health endpoint benchmark:"
    local total=0
    for ((i=1; i<=HEALTH_TESTS; i++)); do
        local time
        time=$(curl -w "%{time_total}" -s "$BASE_URL/health" -o /dev/null)
        total=$(echo "$total + $time" | bc -l)
        printf "  Test %d: %.3fs\n" "$i" "$time"
    done
    local avg
    avg=$(echo "scale=3; $total / $HEALTH_TESTS" | bc -l)
    printf "  Average: %.3fs\n" "$avg"
}

benchmark_lookup() {
    local test_name="$1"
    echo "$test_name:"
    
    local total=0
    local count=0
    
    for ip in "${TEST_IPS[@]}"; do
        local time
        time=$(curl -w "%{time_total}" -s "$BASE_URL/$ip" -o /dev/null)
        total=$(echo "$total + $time" | bc -l)
        count=$((count + 1))
        printf "  %-15s %.3fs\n" "$ip:" "$time"
    done
    
    local avg
    avg=$(echo "scale=3; $total / $count" | bc -l)
    printf "  Average: %.3fs\n" "$avg"
}

benchmark_load() {
    echo "Load test ($LOAD_REQUESTS requests, $LOAD_CONCURRENCY concurrent):"
    
    if command -v ab >/dev/null 2>&1; then
        ab -n "$LOAD_REQUESTS" -c "$LOAD_CONCURRENCY" -q "$BASE_URL/8.8.8.8" 2>/dev/null | \
            grep -E "(Requests per second|Time per request|Transfer rate)" | \
            sed 's/^/  /'
    else
        echo "Apache Bench not found, using curl-based test..."
        local start_time end_time duration rps
        start_time=$(date +%s.%N)
        
        for ((i=1; i<=20; i++)); do
            curl -s "$BASE_URL/8.8.8.8" >/dev/null &
        done
        wait
        
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l)
        rps=$(echo "scale=2; 20 / $duration" | bc -l)
        printf "  20 requests in %.3fs (%.2f req/s)\n" "$duration" "$rps"
    fi
}

main() {
    echo "IpContext API Benchmark"
    echo "======================"
    echo
    
    check_dependencies
    check_service
    echo
    
    benchmark_health
    echo
    
    benchmark_lookup "IP lookup"
    echo
    
    benchmark_load
}

main "$@"
