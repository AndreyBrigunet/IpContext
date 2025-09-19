package languages

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Store keeps cached languages per country code and refreshes them periodically
// using GeoNames countryInfoJSON endpoint.
type Store struct {
	username string
	interval time.Duration
	client   *http.Client
	log      zerolog.Logger

	mu   sync.RWMutex
	data map[string][]string // countryCode -> languages (ISO 639 codes as provided by GeoNames)

	countries []string
}

func New(username string, interval time.Duration, countries []string, logger zerolog.Logger) *Store {
	return &Store{
		username:  username,
		interval:  interval,
		client:    &http.Client{Timeout: 8 * time.Second},
		log:       logger,
		data:      make(map[string][]string),
		countries: dedupSorted(countries),
	}
}

func dedupSorted(in []string) []string {
	m := map[string]struct{}{}
	for _, v := range in { m[v] = struct{}{} }
	out := make([]string, 0, len(m))
	for v := range m { out = append(out, v) }
	sort.Strings(out)
	return out
}

// Start launches a background updater.
func (s *Store) Start(stop <-chan struct{}) {
	if s.username == "" { return }
	go func() {
		s.log.Info().Msg("Starting languages updater")
		s.refreshAll()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.refreshAll()
			case <-stop:
				s.log.Info().Msg("Stopping languages updater")
				return
			}
		}
	}()
}

func (s *Store) refreshAll() {
	if s.username == "" { return }

	for _, cc := range s.countries {
		if err := s.refresh(cc); err != nil {
			s.log.Warn().Err(err).Str("country", cc).Msg("Failed to refresh languages")
		}

		time.Sleep(1100 * time.Millisecond)
	}
}

// RefreshAllOnce triggers a single full refresh cycle.
func (s *Store) RefreshAllOnce() { s.refreshAll() }

// RefreshAll is an alias for RefreshAllOnce for symmetry with neighbours.
func (s *Store) RefreshAll() { s.refreshAll() }

func (s *Store) refresh(countryCode string) error {
	url := fmt.Sprintf("http://api.geonames.org/countryInfoJSON?country=%s&username=%s", countryCode, s.username)
	resp, err := s.client.Get(url)
	if err != nil { return err }

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("geonames status %d", resp.StatusCode)
	}

	var payload struct {
		Geonames []struct {
			Languages string `json:"languages"`
		} `json:"geonames"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil { return err }
	if len(payload.Geonames) == 0 {
		return nil
	}

	langs := parseLanguages(payload.Geonames[0].Languages)
	if len(langs) == 0 {
		return nil
	}

	s.mu.Lock()
	s.data[countryCode] = langs
	s.mu.Unlock()

	return nil
}

func parseLanguages(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" { return nil }

	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}

	for _, p := range parts {
		code := strings.TrimSpace(p)
		if code == "" { continue }
		// GeoNames may include variants like pt-BR; keep as-is
		if _, ok := seen[code]; ok { continue }
		seen[code] = struct{}{}
		out = append(out, code)
	}

	return out
}

// Get returns languages for a country code
func (s *Store) Get(countryCode string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[countryCode]
}
