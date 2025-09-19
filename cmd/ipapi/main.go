package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreybrigunet/ipapi/internal/geoip"
	"github.com/andreybrigunet/ipapi/internal/neighbours"
	"github.com/andreybrigunet/ipapi/internal/languages"
	"github.com/andreybrigunet/ipapi/internal/server"
	"github.com/andreybrigunet/ipapi/internal/config"
	"github.com/andreybrigunet/ipapi/internal/logx"
)

func main() {
	// Load unified config (env + flags defaults)
	cfg := config.Load()

	// Setup logging via logx
	logger := logx.New(logx.Options{
		Level:      cfg.LogLevel,
		Format:     cfg.LogFormat,
		TimeFormat: cfg.LogTimeFmt,
	})

	logger.Info().Msg("Starting IP API service")

	var neighStore *neighbours.Store
	var langStore *languages.Store
	if username := cfg.GeoNamesUser; username != "" {
		neighInterval := 168 * time.Hour
		if v := cfg.NeighboursUpdateHours; v > 0 {
			if h := v; h > 0 {
				neighInterval = time.Duration(h) * time.Hour
			}
		}
		neighStore = neighbours.New(username, neighInterval, geoip.CountryCodes(), logger)

		// languages interval
		langInterval := 168 * time.Hour
		if v := cfg.LanguagesUpdateHours; v > 0 {
			if h := v; h > 0 {
				langInterval = time.Duration(h) * time.Hour
			}
		}
		langStore = languages.New(username, langInterval, geoip.CountryCodes(), logger)
	} else {
		logger.Info().Msg("GEONAMES_USERNAME not set; neighbours and languages will be disabled")
	}

	// Initialize GeoIP service
	geoIP, err := geoip.New(cfg.DBPath, neighStore, langStore, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize GeoIP service")
	}

	// Create and start HTTP server
	srv := server.NewServer(cfg.ListenAddr, geoIP, logger)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Coordinator to ensure neighbours and languages do not refresh at the same time
	// and languages runs after neighbours when both are due.
	type intervals struct{ neigh, lang time.Duration }
	iv := intervals{neigh: 0, lang: 0}
	if neighStore != nil { iv.neigh = 168 * time.Hour }
	if langStore != nil { iv.lang = 168 * time.Hour }
	// capture actual configured intervals
	if username := cfg.GeoNamesUser; username != "" {
		if h := cfg.NeighboursUpdateHours; h > 0 { iv.neigh = time.Duration(h) * time.Hour }
		if h := cfg.LanguagesUpdateHours; h > 0 { iv.lang = time.Duration(h) * time.Hour }
	}

	go func() {
		if neighStore == nil && langStore == nil { return }
		logger.Info().Msg("Starting GeoNames coordinator (neighbours -> languages)")
		now := time.Now()
		nextNeigh := now
		nextLang := now
		for {
			// Determine next wake-up time
			next := nextNeigh
			if langStore != nil && (langStore != nil && (iv.lang > 0) && (nextLang.Before(next) || next.IsZero())) {
				next = nextLang
			}
			sleep := time.Until(next)
			if sleep < 0 { sleep = 0 }
			select {
			case <-ctx.Done():
				return
			case <-time.After(sleep):
				// Run neighbours if due
				now = time.Now()
				if neighStore != nil && (iv.neigh > 0) && !now.Before(nextNeigh) {
					neighStore.RefreshAllOnce()
					nextNeigh = nextNeigh.Add(iv.neigh)
				}
				// Then languages if due
				now = time.Now()
				if langStore != nil && (iv.lang > 0) && !now.Before(nextLang) {
					langStore.RefreshAllOnce()
					nextLang = nextLang.Add(iv.lang)
				}
			}
		}
	}()

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
