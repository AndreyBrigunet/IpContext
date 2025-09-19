package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreybrigunet/ipapi/internal/config"
	"github.com/andreybrigunet/ipapi/internal/geoip"
	"github.com/andreybrigunet/ipapi/internal/languages"
	"github.com/andreybrigunet/ipapi/internal/logx"
	"github.com/andreybrigunet/ipapi/internal/neighbours"
	"github.com/andreybrigunet/ipapi/internal/server"
)

func main() {
	cfg := config.Load()

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
			neighInterval = time.Duration(v) * time.Hour
		}
		neighStore = neighbours.New(username, neighInterval, geoip.CountryCodes(), logger)

		langInterval := 168 * time.Hour
		if v := cfg.LanguagesUpdateHours; v > 0 {
			langInterval = time.Duration(v) * time.Hour
		}
		langStore = languages.New(username, langInterval, geoip.CountryCodes(), logger)
	} else {
		logger.Info().Msg("GEONAMES_USERNAME not set; neighbours and languages will be disabled")
	}

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

	type intervals struct{ neigh, lang time.Duration }
	iv := intervals{neigh: 0, lang: 0}
	if neighStore != nil { iv.neigh = 168 * time.Hour }
	if langStore != nil { iv.lang = 168 * time.Hour }
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
				now = time.Now()
				if neighStore != nil && (iv.neigh > 0) && !now.Before(nextNeigh) {
					neighStore.RefreshAllOnce()
					nextNeigh = nextNeigh.Add(iv.neigh)
				}
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
