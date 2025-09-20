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
   docker compose -f docker-compose.dev.yml build --no-cache
   docker compose -f docker-compose.dev.yml up -d
   ```

4. **Verify the Service**
   ```bash
   curl "http://localhost:3280/"
   curl "http://localhost:3280/health"
   ```

### Development Setup (Local Build)

For development with local builds:
```bash
docker compose -f docker-compose.dev.yml up -d
```

### GitHub Actions CI/CD

This project uses GitHub Actions for automated building and publishing of Docker images:

1. **Automatic Builds**: Every push to `main`/`master` branch triggers a build
2. **Multi-platform**: Builds for both `linux/amd64` and `linux/arm64`
3. **Security Scanning**: Includes Trivy vulnerability scanning
4. **Container Registry**: Images are published to GitHub Container Registry (ghcr.io)

**Available Image Tags:**
- `ghcr.io/andreybrigunet/ipapi:latest` - Latest stable build
- `ghcr.io/andreybrigunet/ipapi:main` - Latest main branch build
- `ghcr.io/andreybrigunet/ipapi:v1.0.0` - Specific version tags

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
  "mobile": false,
  "proxy": false,
  "hosting": true
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

## Development

### Building Locally

1. Install Go 1.21 or later
2. Clone the repository
3. Build and run:
   ```bash
   go build -o ipapi ./cmd/ipapi
   ./ipapi --db-path /path/to/geoip/databases
   ```


## License

This project is licensed under the MIT License