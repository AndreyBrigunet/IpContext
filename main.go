package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreybrigunet/ipapi/internal/config"
	"github.com/andreybrigunet/ipapi/internal/coordinator"
	"github.com/andreybrigunet/ipapi/internal/geoip"
	"github.com/andreybrigunet/ipapi/internal/languages"
	"github.com/andreybrigunet/ipapi/internal/logx"
	"github.com/andreybrigunet/ipapi/internal/neighbours"
	"github.com/andreybrigunet/ipapi/internal/server"
	"github.com/rs/zerolog"
)

func main() {
	cfg := config.Load()

	logger := logx.New(logx.Options{
		Level:      cfg.LogLevel,
		Format:     cfg.LogFormat,
		TimeFormat: cfg.LogTimeFmt,
	})

	logger.Info().Msg("Starting IP API service")

	neighStore, langStore := initializeStores(cfg, logger)

	geoIP, err := geoip.New(cfg.DBPath, neighStore, langStore, logger)
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("dbPath", cfg.DBPath).
			Msg("Failed to initialize GeoIP service. Ensure MaxMind databases exist at DB_PATH env (e.g., /app/data)")
	}

	srv := server.NewServer(cfg.ListenAddr, geoIP, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if neighStore != nil || langStore != nil {
		intervals := coordinator.Intervals{
			Neighbours: calculateInterval(cfg.NeighboursUpdateHours),
			Languages:  calculateInterval(cfg.LanguagesUpdateHours),
		}
		coord := coordinator.New(neighStore, langStore, intervals, logger)
		coord.Start(ctx)
	}

	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	<-ctx.Done()
	logger.Info().Msg("Shutting down...")


	if err := srv.Stop(); err != nil {
		logger.Error().Err(err).Msg("Error during server shutdown")
	} else {
		logger.Info().Msg("Server stopped gracefully")
	}
}


func initializeStores(cfg *config.Config, logger zerolog.Logger) (*neighbours.Store, *languages.Store) {
	if cfg.GeoNamesUser == "" {
		logger.Info().Msg("GEONAMES_USERNAME not set; neighbours and languages will be disabled")
		return nil, nil
	}

	countryCodes := geoip.CountryCodes()
	
	neighInterval := calculateInterval(cfg.NeighboursUpdateHours)
	neighStore := neighbours.New(cfg.GeoNamesUser, neighInterval, countryCodes, logger)

	langInterval := calculateInterval(cfg.LanguagesUpdateHours)
	langStore := languages.New(cfg.GeoNamesUser, langInterval, countryCodes, logger)

	return neighStore, langStore
}


func calculateInterval(hours int) time.Duration {
	if hours > 0 {
		return time.Duration(hours) * time.Hour
	}

	return 168 * time.Hour
}
