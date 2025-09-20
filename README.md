# 🌍 IpContext - Open-Source IP Metadata API

[![Go Report Card](https://goreportcard.com/badge/github.com/andreybrigunet/IpContext)](https://goreportcard.com/report/github.com/andreybrigunet/IpContext)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/andreybrigunet/IpContext)](https://golang.org/)
[![Build Status](https://img.shields.io/github/actions/workflow/status/andreybrigunet/IpContext/docker-build.yml?branch=master)](https://github.com/andreybrigunet/IpContext/actions)

**IpContext** is a high-performance, open-source IP geolocation and metadata API service built with Go. Get comprehensive information about any IP address including geolocation, ISP details, timezone, currency, neighboring countries, and supported languages - all in a single, lightning-fast API call.

🚀 **Production-ready** • 🐳 **Docker-native** • 🔄 **Auto-updating** • ⚡ **Sub-millisecond responses**

## ✨ Features

### 🌐 **Comprehensive IP Intelligence**
- **Geolocation Data**: Country, region, city, coordinates, timezone
- **ISP & ASN Information**: Internet Service Provider, Organization, Autonomous System details
- **Currency Information**: Currency codes, symbols
- **Country Neighbors**: List of neighboring countries with names
- **Language Support**: Supported languages for each country
- **EU Membership**: European Union membership status

### ⚡ **High Performance**
- **Sub-millisecond lookups** with intelligent caching
- **Concurrent processing** with Go's goroutines
- **Memory-efficient** data structures
- **Built-in health monitoring**

### 🔄 **Auto-Updating**
- **MaxMind GeoLite2** database integration
- **Automatic daily updates** via geoipupdate
- **Zero-downtime updates**
- **Configurable update intervals**

### 🐳 **Production Ready**
- **Docker & Docker Compose** support
- **Graceful shutdown** handling
- **Structured logging** with zerolog
- **CORS support** for web applications
- **Health check endpoints**

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository:**
```bash
git clone https://github.com/andreybrigunet/IpContext.git
cd IpContext
```

2. **Set up environment:**
```bash
cp .env.example .env
# Edit .env with your MaxMind credentials (see Configuration section)
```

3. **Start the service:**
```bash
docker compose up -d
```


4. **Test the API:**
```bash
curl http://localhost:3280/8.8.8.8
```

The API will be available at `http://localhost:3280` 🎉

## 📡 API Usage

### **Health Check**
```bash
curl http://localhost:3280/health
```

### **Get Your IP Information**
```bash
curl http://localhost:3280/
```

### **Query Specific IP**
```bash
curl http://localhost:3280/8.8.8.8
```

### **Response Format**

```json
{
  "query": "8.8.8.8",
  "status": "success",
  "continent": "North America",
  "continentCode": "NA",
  "country": "United States",
  "countryCode": "US",
  "region": "California",
  "regionName": "California",
  "city": "Mountain View",
  "zip": "94043",
  "lat": 37.751,
  "lon": -97.822,
  "timezone": "America/Chicago",
  "offset": -18000,
  "currencyCode": "USD",
  "currencySymbol": "$",
  "isp": "GOOGLE",
  "org": "GOOGLE",
  "as": "AS15169 GOOGLE",
  "asname": "GOOGLE",
  "neighbours": [
    {
      "countryCode": "CA",
      "countryName": "Canada"
    },
    {
      "countryCode": "CU",
      "countryName": "Cuba"
    },
    {
      "countryCode": "MX", 
      "countryName": "Mexico"
    }
  ],
  "languages": [
    {
      "code": "en",
      "name": "English",
      "native": "English"
    }
  ],
  "isEUCountry": false
}
```

## ⚙️ Configuration

Configure IpContext via environment variables or command-line flags:

| Environment Variable | Flag | Default | Description |
|---------------------|------|---------|-------------|
| `LISTEN_ADDR` | `-listen` | `:3280` | Server listen address |
| `DB_PATH` | `-db-path` | `/data` | Path to MaxMind database files |
| `LOG_LEVEL` | `-log-level` | `info` | Log level (debug, info, warn, error, fatal) |
| `LOG_FORMAT` | | `console` | Log format (console, json) |
| `LOG_TIME_FORMAT` | | `2006-01-02 15:04:05` | Log timestamp format |
| `GEONAMES_USERNAME` | | | GeoNames username for enhanced features |
| `NEIGHBOURS_UPDATE_HOURS` | | `168` | Hours between neighbor data updates |
| `LANGUAGES_UPDATE_HOURS` | | `168` | Hours between language data updates |
| `CACHE_TTL_MINUTES` | | `5` | Response cache TTL in minutes |

### **Required MaxMind Setup**

1. **Create MaxMind Account**: Sign up at [MaxMind](https://www.maxmind.com/en/geolite2/signup)
2. **Generate License Key**: Create a license key in your account dashboard
3. **Configure Environment**:

```bash
# .env file
GEOIPUPDATE_ACCOUNT_ID=your_account_id
GEOIPUPDATE_LICENSE_KEY=your_license_key

# Optional: Enhanced features
GEONAMES_USERNAME=your_geonames_username
```

## 🐳 Docker Deployment

### **Production Setup with Auto-Updates**

The included `docker-compose.yml` provides a complete production setup:

```yaml
services:
  geoip-app:
    image: ghcr.io/andreybrigunet/ipcontext:latest
    pull_policy: always
    ports:
      - "3280:3280"
    volumes:
      - geoip_data:/app/data
    environment:
      - TZ=UTC
      - DB_PATH=/app/data
      - GEONAMES_USERNAME=${GEONAMES_USERNAME:-}
      - NEIGHBOURS_UPDATE_HOURS=${NEIGHBOURS_UPDATE_HOURS:-168}
      - LANGUAGES_UPDATE_HOURS=${LANGUAGES_UPDATE_HOURS:-168}
      - CACHE_TTL_MINUTES=${CACHE_TTL_MINUTES:-5}
      - LOG_LEVEL=${LOG_LEVEL:-info}
      - LOG_FORMAT=${LOG_FORMAT:-console}
    depends_on:
      - geoip-update
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://localhost:3280/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  geoip-update:
    image: maxmindinc/geoipupdate:latest
    env_file:
      - .env
    environment:
      - GEOIPUPDATE_ACCOUNT_ID=${MM_ACCOUNT_ID}
      - GEOIPUPDATE_LICENSE_KEY=${MM_LICENSE_KEY}
      - GEOIPUPDATE_EDITION_IDS=GeoLite2-City GeoLite2-ASN
      - GEOIPUPDATE_FREQUENCY=24
      - GEOIPUPDATE_DB_DIR=/geoip_data
    volumes:
      - geoip_data:/geoip_data
    restart: unless-stopped

volumes:
  geoip_data:
```

### **Development Setup**

```bash
# Clone and build
git clone https://github.com/andreybrigunet/IpContext.git
cd IpContext

# Use development compose file
docker compose -f docker-compose.dev.yml build --no-cache
docker compose -f docker-compose.dev.yml up -d

# View logs
docker compose logs -f geoip-app
```

### **Performance Testing**

```bash
# Load test with included script
./benchmark.sh
```

## 🗺️ Roadmap

### **Upcoming Features**
- [ ] **Rate Limiting**: Configurable rate limiting per IP/API key
- [ ] **Metrics & Monitoring**: Prometheus metrics endpoint
- [ ] **API Authentication**: Optional API key system
- [ ] **Batch Processing**: Multiple IP lookups in single request


## 📊 Performance Benchmarks

| Metric | Value |
|--------|-------|
| **Response Time** | 1.827 [ms] |
| **Throughput** | 5474.95 [#/sec] |
| **Memory Usage** | < 50MB |
| **Database Size** | ~100MB (GeoLite2) |
| **Cold Start** | < 2 seconds |

## 🤝 Contributing

We love contributions! Here's how you can help:

- 🐛 **Report bugs** via [GitHub Issues](https://github.com/andreybrigunet/IpContext/issues)
- 📖 **Improve documentation**
- 🧪 **Add tests** for better coverage
- 🚀 **Submit pull requests**

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **[MaxMind](https://www.maxmind.com/)** for providing GeoLite2 databases
- **[GeoNames](https://www.geonames.org/)** for geographical data APIs
- **[oschwald/geoip2-golang](https://github.com/oschwald/geoip2-golang)** for MaxMind Go integration
- **[rs/zerolog](https://github.com/rs/zerolog)** for high-performance logging

## 📞 Support & Community

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/andreybrigunet/IpContext/issues)

---

<div align="center">

**⭐ Star this project if you find it useful!**

**Keywords**: `go` `golang` `ip-geolocation` `ip-api` `ipinfo` `geoip` `maxmind` `mmdb` `docker` `api` `microservice` `geolocation` `asn` `isp` `timezone` `currency` `neighbors` `languages` `high-performance` `open-source`

</div>