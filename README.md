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

1. **Get MaxMind License Key**
   - Sign up at [MaxMind](https://www.maxmind.com/en/geolite2/signup)
   - Create a license key in your account settings

2. **Configure Environment**
   Create a `.env` file in the project root:
   ```
   GEOIPUPDATE_ACCOUNT_ID=your_account_id
   GEOIPUPDATE_LICENSE_KEY=your_license_key
   ```

3. **Build and Run**
   ```bash
   docker-compose up -d
   ```

4. **Verify the Service**
   ```bash
   curl "http://localhost:3280/lookup?ip=8.8.8.8"
   ```

## API Endpoints

### GET /lookup

Look up geolocation information for an IP address.

**Query Parameters:**
- `ip` - (optional) IP address to look up. If not provided, the client's IP will be used.

**Example Request:**
```bash
curl "http://localhost:3280/lookup?ip=8.8.8.8"
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
  "currency": "USD",
  "isp": "Google LLC",
  "org": "Google Public DNS",
  "as": "AS15169 Google LLC",
  "asname": "GOOGLE",
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

## Development

### Building Locally

1. Install Go 1.21 or later
2. Clone the repository
3. Build and run:
   ```bash
   go build -o ipapi ./cmd/ipapi
   ./ipapi --db-path /path/to/geoip/databases
   ```

### Running Tests

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [MaxMind](https://www.maxmind.com) for the GeoLite2 databases
- [geoip2-golang](https://github.com/oschwald/geoip2-golang) for the Go MaxMind DB reader
