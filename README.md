# IP API Service

A high-performance IP geolocation service similar to ip-api.com, built with Go and MaxMind GeoLite2 databases.

## Features

- Fast IP geolocation lookups (<2ms typical response time)
- Supports both IPv4 and IPv6 addresses
- Automatic IP detection from X-Forwarded-For, CF-Connecting-IP headers
- Daily automatic database updates
- Containerized with Docker
- RESTful JSON API

## Prerequisites

- Docker and Docker Compose
- MaxMind GeoLite2 Account (free)

## Getting Started

### Production Deployment (Using Pre-built Images)

1. **Get MaxMind License Key**
   - Sign up at [MaxMind](https://www.maxmind.com/en/geolite2/signup)
   - Create a license key in your account settings

2. **Configure Environment**
   Create a `.env` file in the project root:
   ```
   GEOIPUPDATE_ACCOUNT_ID=your_account_id
   GEOIPUPDATE_LICENSE_KEY=your_license_key
   GEONAMES_USERNAME=your_geonames_username
   ```

3. **Deploy with Pre-built Image**
   ```bash
   docker compose build --no-cache
   docker compose up -d
   ```

4. **Verify the Service**
   ```bash
   curl "http://localhost:3280/"
   curl "http://localhost:3280/health"
   ```

### Development Setup (Local Build)

For development with local builds:
```bash
docker compose -f docker-compose.dev.yml build --no-cache
docker compose -f docker-compose.dev.yml up -d
```

## API Endpoints

### Endpoints

- `GET /` — returns info for the requester IP.
- `GET /:ip` — returns info for the specified IP (e.g., `/8.8.8.8`).
- `GET /health` — health check endpoint.

**Example Requests:**
```bash
# Requester IP
curl "http://localhost:3280/"

# Specific IP (short form)
curl "http://localhost:3280/8.8.8.8"
```

**Example Response (200 OK):**
```json
{
  "query": "8.8.8.8",
  "status": "success",
  "continent": "North America",
  "continentCode": "NA",
  "country": "United States",
  "countryCode": "US",
  "region": "CA",
  "regionName": "California",
  "city": "Mountain View",
  "district": "",
  "zip": "94043",
  "lat": 37.4192,
  "lon": -122.0574,
  "timezone": "America/Los_Angeles",
  "offset": -25200,
  "currencyCode": "USD",
  "currencySymbol": "$",
  "isEUCountry": false,
  "languages": ["en"],
  "neighbours": [
    {"countryCode": "CA", "countryName": "Canada"},
    {"countryCode": "MX", "countryName": "Mexico"}
  ],
  "isp": "Google LLC",
  "org": "Google LLC",
  "as": "AS15169 Google LLC",
  "asname": "Google LLC",
}
```

**Error Response (400 Bad Request):**
```json
{
  "status": "fail",
  "message": "Invalid IP address"
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GEOIPUPDATE_ACCOUNT_ID` | MaxMind account ID | Required |
| `GEOIPUPDATE_LICENSE_KEY` | MaxMind license key | Required |
| `GEOIPUPDATE_EDITION_IDS` | Database editions to download | `GeoLite2-City GeoLite2-ASN` |
| `GEOIPUPDATE_FREQUENCY` | Update frequency in hours | `24` |
| `GEONAMES_USERNAME` | GeoNames username to enable neighbours API | empty (disabled) |
| `NEIGHBOURS_UPDATE_HOURS` | Neighbours refresh interval in hours | `168` (weekly) |
| `LANGUAGES_UPDATE_HOURS` | Languages refresh interval in hours | `168` (weekly) |
| `CACHE_TTL_MINUTES` | In-memory cache TTL in minutes | `5` |
| `LISTEN_ADDR` | Server listen address | `:3280` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `LOG_FORMAT` | Log format (json, console) | `console` |

### Performance Tuning

For **maximum performance** in production:

```bash
# Recommended production settings
LOG_LEVEL=warn              # Reduce logging overhead
CACHE_TTL_MINUTES=10        # Longer cache for better hit ratio
LOG_FORMAT=json             # Structured logging for monitoring
```

For **development**:
```bash
# Development settings
LOG_LEVEL=debug             # Detailed logging
CACHE_TTL_MINUTES=1         # Short cache for testing
LOG_FORMAT=console          # Human-readable logs
```

## Benchmarking & Monitoring

### Performance Testing

**Quick Benchmark:**
```bash
# Run the included benchmark script
chmod +x benchmark.sh
./benchmark.sh
```

**Manual Testing:**
```bash
# Simple response time test
time curl -s "http://localhost:3280/8.8.8.8" > /dev/null

# Load testing with Apache Bench (if installed)
ab -n 1000 -c 10 http://localhost:3280/8.8.8.8

# Health check monitoring
curl -w "@curl-format.txt" -s "http://localhost:3280/health"
```

Create `curl-format.txt` for detailed timing:
```
     time_namelookup:  %{time_namelookup}s\n
        time_connect:  %{time_connect}s\n
     time_appconnect:  %{time_appconnect}s\n
    time_pretransfer:  %{time_pretransfer}s\n
       time_redirect:  %{time_redirect}s\n
  time_starttransfer:  %{time_starttransfer}s\n
                     ----------\n
          time_total:  %{time_total}s\n
```

### Cache Statistics

Monitor cache performance by checking logs for cache hit/miss patterns. In debug mode, you'll see detailed timing information.

## License

This project is licensed under the MIT License