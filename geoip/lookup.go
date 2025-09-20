package geoip

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/andreybrigunet/ipapi/neighbours"
	"github.com/andreybrigunet/ipapi/languages"
	"github.com/oschwald/geoip2-golang"
	"github.com/rs/zerolog"
)

type GeoIP struct {
	cityDB    *geoip2.Reader
	asnDB     *geoip2.Reader
	neigh     *neighbours.Store
	langs     *languages.Store
	logger    zerolog.Logger
	closeOnce sync.Once
}

// Response represents the IP lookup response structure
type Response struct {
	Query         string              `json:"query"`
	Status        string              `json:"status"`
	Continent     string              `json:"continent,omitempty"`
	ContinentCode string              `json:"continentCode,omitempty"`
	Country       string              `json:"country,omitempty"`
	CountryCode   string              `json:"countryCode,omitempty"`
	Region        string              `json:"region,omitempty"`
	RegionName    string              `json:"regionName,omitempty"`
	City          string              `json:"city,omitempty"`
	District      string              `json:"district,omitempty"`
	Zip           string              `json:"zip,omitempty"`
	Lat           float64             `json:"lat,omitempty"`
	Lon           float64             `json:"lon,omitempty"`
	Timezone      string              `json:"timezone,omitempty"`
	Offset        int                 `json:"offset,omitempty"`
	CurrencyCode  string              `json:"currencyCode,omitempty"`
	CurrencySymbol string              `json:"currencySymbol,omitempty"`
	ISP           string              `json:"isp,omitempty"`
	Org           string              `json:"org,omitempty"`
	AS            string              `json:"as,omitempty"`
	ASName        string              `json:"asname,omitempty"`
	Mobile        bool                `json:"mobile"`
	Proxy         bool                `json:"proxy"`
	Hosting       bool                `json:"hosting"`
	Neighbours    []neighbours.Neighbour `json:"neighbours,omitempty"`
	IsEUCountry   bool                `json:"isEUCountry"`
	Languages     []string            `json:"languages,omitempty"`
}

// New creates a new GeoIP service instance
func New(dbPath string, neigh *neighbours.Store, langs *languages.Store, logger zerolog.Logger) (*GeoIP, error) {
	cityDB, err := geoip2.Open(filepath.Join(dbPath, "GeoLite2-City.mmdb"))
	if err != nil {
		return nil, err
	}

	asnDB, err := geoip2.Open(filepath.Join(dbPath, "GeoLite2-ASN.mmdb"))
	if err != nil {
		cityDB.Close()
		return nil, err
	}

	return &GeoIP{
		cityDB: cityDB,
		asnDB:  asnDB,
		neigh:  neigh,
		langs:  langs,
		logger: logger,
	}, nil
}

// Lookup performs an IP address lookup
func (g *GeoIP) Lookup(ipStr string) (*Response, error) {
	return g.LookupWithContext(context.Background(), ipStr)
}

// LookupWithContext performs an IP address lookup with context
func (g *GeoIP) LookupWithContext(ctx context.Context, ipStr string) (*Response, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, errors.New("invalid IP address")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Lookup city data
	city, err := g.cityDB.City(ip)
	if err != nil {
		return nil, err
	}

	// Lookup ASN data
	asn, err := g.asnDB.ASN(ip)
	if err != nil {
		// Non-fatal error, we can continue without ASN data
		g.logger.Warn().Err(err).Str("ip", ipStr).Msg("Failed to lookup ASN data")
	}

	// Build response
	resp := &Response{
		Query:      ipStr,
		Status:     "success",
		Continent:  city.Continent.Names["en"],
		ContinentCode: city.Continent.Code,
		Country:    city.Country.Names["en"],
		CountryCode: city.Country.IsoCode,
		City:       city.City.Names["en"],
		Zip:        city.Postal.Code,
		Lat:        city.Location.Latitude,
		Lon:        city.Location.Longitude,
		Timezone:   city.Location.TimeZone,
		Mobile:     false,
		Proxy:      false,
		Hosting:    false,
	}

	// Add region data if available
	if len(city.Subdivisions) > 0 {
		subdiv := city.Subdivisions[0]
		resp.Region = subdiv.IsoCode
		resp.RegionName = subdiv.Names["en"]
	}

	// Add ASN data if available
	if asn != nil {
		// ip-api format example: "AS15169 Google LLC"
		resp.AS = "AS" + strconv.Itoa(int(asn.AutonomousSystemNumber)) + " " + asn.AutonomousSystemOrganization
		resp.ASName = asn.AutonomousSystemOrganization
		// Use ASN Org for ISP/Org as a reasonable approximation
		resp.ISP = asn.AutonomousSystemOrganization
		resp.Org = asn.AutonomousSystemOrganization
	}

	// Add currency information
	if code, ok := countryCurrencyMap[city.Country.IsoCode]; ok {
		resp.CurrencyCode = code

		if sym, ok := currencySymbols[code]; ok {
			resp.CurrencySymbol = sym
		} else {
			resp.CurrencySymbol = code
		}
	}

	// Attach neighbours if available
	if g.neigh != nil && resp.CountryCode != "" {
		resp.Neighbours = g.neigh.Get(resp.CountryCode)
	}

	// Determine EU membership
	if resp.CountryCode != "" {
		resp.IsEUCountry = IsEUCountry(resp.CountryCode)
	}

	// Attach languages if available
	if g.langs != nil && resp.CountryCode != "" {
		resp.Languages = g.langs.Get(resp.CountryCode)
	}

	// Compute timezone offset in seconds (relative to UTC) as in ip-api
	if resp.Timezone != "" {
		if loc, err := time.LoadLocation(resp.Timezone); err == nil {
			_, offset := time.Now().In(loc).Zone()
			resp.Offset = offset
		}
	}

	return resp, nil
}

// Close releases resources used by the GeoIP databases
func (g *GeoIP) Close() error {
	var err1, err2 error
	
	g.closeOnce.Do(func() {
		err1 = g.cityDB.Close()
		err2 = g.asnDB.Close()
	})

	if err1 != nil {
		return err1
	}
	return err2
}
