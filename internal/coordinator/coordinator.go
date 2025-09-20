package coordinator

import (
	"context"
	"time"

	"github.com/andreybrigunet/ipapi/internal/languages"
	"github.com/andreybrigunet/ipapi/internal/neighbours"
	"github.com/rs/zerolog"
)

type Coordinator struct {
	neighStore *neighbours.Store
	langStore  *languages.Store
	logger     zerolog.Logger
	intervals  Intervals
}

type Intervals struct {
	Neighbours time.Duration
	Languages  time.Duration
}

func New(neighStore *neighbours.Store, langStore *languages.Store, intervals Intervals, logger zerolog.Logger) *Coordinator {
	return &Coordinator{
		neighStore: neighStore,
		langStore:  langStore,
		logger:     logger,
		intervals:  intervals,
	}
}

func (c *Coordinator) Start(ctx context.Context) {
	if c.neighStore == nil && c.langStore == nil {
		c.logger.Info().Msg("No stores configured, coordinator will not run")
		return
	}

	go c.run(ctx)
}

func (c *Coordinator) run(ctx context.Context) {
	c.logger.Info().Msg("Starting GeoNames coordinator (neighbours -> languages)")
	
	now := time.Now()
	nextNeigh := now
	nextLang := now

	for {
		next := c.calculateNextUpdate(nextNeigh, nextLang)
		sleep := time.Until(next)
		if sleep < 0 {
			sleep = 0
		}

		select {
		case <-ctx.Done():
			c.logger.Info().Msg("Coordinator shutting down")
			return
		case <-time.After(sleep):
			now = time.Now()
			
			if c.shouldUpdateNeighbours(now, nextNeigh) {
				c.neighStore.RefreshAllOnce()
				nextNeigh = nextNeigh.Add(c.intervals.Neighbours)
			}
			
			if c.shouldUpdateLanguages(now, nextLang) {
				c.langStore.RefreshAllOnce()
				nextLang = nextLang.Add(c.intervals.Languages)
			}
		}
	}
}

func (c *Coordinator) calculateNextUpdate(nextNeigh, nextLang time.Time) time.Time {
	next := nextNeigh
	
	if c.langStore != nil && c.intervals.Languages > 0 && 
		(nextLang.Before(next) || next.IsZero()) {
		next = nextLang
	}
	
	return next
}

func (c *Coordinator) shouldUpdateNeighbours(now, nextNeigh time.Time) bool {
	return c.neighStore != nil && 
		   c.intervals.Neighbours > 0 && 
		   !now.Before(nextNeigh)
}

func (c *Coordinator) shouldUpdateLanguages(now, nextLang time.Time) bool {
	return c.langStore != nil && 
		   c.intervals.Languages > 0 && 
		   !now.Before(nextLang)
}
