package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds application configuration loaded from env and flags.
type Config struct {
	ListenAddr string
	DBPath     string
	LogLevel   string
	LogFormat  string // json | console
	LogTimeFmt string // Go layout or aliases handled by logx

	GeoNamesUser         string
	NeighboursUpdateHours int
	LanguagesUpdateHours  int
	CacheTTLMinutes      int
}

// Load reads environment variables and flags, applying sane defaults.
// Flags override environment variables when provided.
func Load() *Config {
	cfg := &Config{
		ListenAddr:            getEnv("LISTEN_ADDR", ":3280"),
		DBPath:                getEnv("DB_PATH", "/data"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		LogFormat:             getEnv("LOG_FORMAT", "console"),
		LogTimeFmt:            getEnv("LOG_TIME_FORMAT", "2006-01-02 15:04:05"),
		GeoNamesUser:          getEnv("GEONAMES_USERNAME", ""),
		NeighboursUpdateHours: getEnvInt("NEIGHBOURS_UPDATE_HOURS", 168),
		LanguagesUpdateHours:  getEnvInt("LANGUAGES_UPDATE_HOURS", 168),
		CacheTTLMinutes:       getEnvInt("CACHE_TTL_MINUTES", 5),
	}

	// Define flags that can override env
	listen := flag.String("listen", cfg.ListenAddr, "Address to listen on")
	dbPath := flag.String("db-path", cfg.DBPath, "Path to GeoIP database files")
	logLevel := flag.String("log-level", cfg.LogLevel, "Log level (debug, info, warn, error, fatal)")
	flag.Parse()

	cfg.ListenAddr = *listen
	cfg.DBPath = *dbPath
	cfg.LogLevel = *logLevel

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
