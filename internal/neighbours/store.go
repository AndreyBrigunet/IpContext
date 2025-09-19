package neighbours

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Neighbour struct {
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
}

// RefreshAllOnce triggers a single full refresh cycle.
func (s *Store) RefreshAllOnce() {
    s.refreshAll()
}

type Store struct {
	username string
	interval time.Duration
	client   *http.Client
	log      zerolog.Logger

	mu   sync.RWMutex
	data map[string][]Neighbour

	countries []string
}

func New(username string, interval time.Duration, countries []string, logger zerolog.Logger) *Store {
	return &Store{
		username:  username,
		interval:  interval,
		client:    &http.Client{Timeout: 8 * time.Second},
		log:       logger,
		data:      make(map[string][]Neighbour),
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

// Start launches a background updater. Safe to call even if username is empty (no-op).
func (s *Store) Start(stop <-chan struct{}) {
	if s.username == "" { return }

	go func() {
		s.log.Info().Msg("Starting neighbours updater")

		s.refreshAll()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.refreshAll()
			case <-stop:
				s.log.Info().Msg("Stopping neighbours updater")
				return
			}
		}
	}()
}

// RefreshAll triggers a single full refresh cycle.
func (s *Store) RefreshAll() {
	s.refreshAll()
}

func (s *Store) refreshAll() {
	if s.username == "" { return }

	for _, cc := range s.countries {
		if err := s.refresh(cc); err != nil {
			s.log.Warn().Err(err).Str("country", cc).Msg("Failed to refresh neighbours")
		}

		time.Sleep(1100 * time.Millisecond)
	}

	s.log.Info().Msg("Neighbours updated")
}

func (s *Store) refresh(countryCode string) error {
	url := fmt.Sprintf("http://api.geonames.org/neighboursJSON?country=%s&username=%s", countryCode, s.username)

	resp, err := s.client.Get(url)
	if err != nil { return err }

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("geonames status %d", resp.StatusCode)
	}

	var payload struct{
		Geonames []struct{
			CountryCode string `json:"countryCode"`
			CountryName string `json:"countryName"`
		} `json:"geonames"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil { return err }
	items := make([]Neighbour, 0, len(payload.Geonames))
	for _, g := range payload.Geonames {
		if g.CountryCode == "" || g.CountryName == "" { continue }
		items = append(items, Neighbour{ CountryCode: g.CountryCode, CountryName: g.CountryName })
	}

	s.mu.Lock()
	s.data[countryCode] = items
	s.mu.Unlock()

	return nil
}

func (s *Store) Get(countryCode string) []Neighbour {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[countryCode]
}
