package geoip

import (
	"sync"
	"time"
)

type TimezoneCache struct {
	mu        sync.RWMutex
	locations map[string]*time.Location
}

var tzCache = &TimezoneCache{
	locations: make(map[string]*time.Location),
}

// GetTimezoneOffset returns the timezone offset in seconds for a given timezone string
func GetTimezoneOffset(timezone string) int {
	if timezone == "" {
		return 0
	}

	tzCache.mu.RLock()
	loc, found := tzCache.locations[timezone]
	tzCache.mu.RUnlock()

	if !found {
		var err error
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			return 0
		}

		tzCache.mu.Lock()
		tzCache.locations[timezone] = loc
		tzCache.mu.Unlock()
	}

	_, offset := time.Now().In(loc).Zone()
	return offset
}
