package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/andreybrigunet/ipapi/internal/geoip"
	"github.com/andreybrigunet/ipapi/internal/server"
)

var (
	listenAddr = flag.String("listen", ":3280", "Address to listen on")
	dbPath     = flag.String("db-path", "/data", "Path to GeoIP database files")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error, fatal)")
)

func main() {
	flag.Parse()

	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	lvl, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	logger := zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Logger()

	logger.Info().Msg("Starting IP API service")

	// Initialize GeoIP service
	geoIP, err := geoip.New(*dbPath, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize GeoIP service")
	}

	// Create and start HTTP server
	srv := server.NewServer(*listenAddr, geoIP, logger)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	<-ctx.Done()
	logger.Info().Msg("Shutting down...")

	if err := srv.Stop(); err != nil {
		logger.Error().Err(err).Msg("Error during server shutdown")
	}

	logger.Info().Msg("Server stopped")
}
