package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/andreybrigunet/ipapi/internal/geoip"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	geoIP  *geoip.GeoIP
	log    zerolog.Logger
}

// handleRoot serves both GET / (client IP) and GET /:ip
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    var ipStr string
    if path == "/" {
        // Use requester IP
        ipStr = r.Header.Get("X-Forwarded-For")
        if ipStr == "" {
            ipStr = r.Header.Get("CF-Connecting-IP")
            if ipStr == "" {
                host, _, err := net.SplitHostPort(r.RemoteAddr)
                if err != nil {
                    s.respondError(w, "Invalid IP", http.StatusOK)
                    return
                }
                ipStr = host
            }
        } else {
            ips := strings.Split(ipStr, ",")
            ipStr = strings.TrimSpace(ips[0])
        }
    } else {
        // Expect /:ip
        ipStr = strings.TrimPrefix(path, "/")
    }

    ip := net.ParseIP(ipStr)
    if ip == nil {
        s.respondError(w, "Invalid IP", http.StatusOK)
        return
    }

    resp, err := s.geoIP.Lookup(ip.String())
    if err != nil {
        s.log.Error().Err(err).Str("ip", ipStr).Msg("Lookup failed")
        s.respondError(w, "No data", http.StatusOK)
        return
    }
    s.respondJSON(w, resp, http.StatusOK)
}

// NewServer creates a new HTTP server
func NewServer(addr string, geoIP *geoip.GeoIP, logger zerolog.Logger) *Server {
	s := &Server{
		geoIP: geoIP,
		log: logger,
	}

	r := http.NewServeMux()
	r.HandleFunc("/", s.handleRoot)
	s.server = &http.Server{
		Addr: addr,
		Handler: r,
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.log.Info().Str("addr", s.server.Addr).Msg("Starting HTTP server")
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	return s.server.Shutdown(context.Background())
}

// respondError sends a JSON error response
func (s *Server) respondError(w http.ResponseWriter, message string, statusCode int) {
	resp := map[string]interface{}{
		"status":  "fail",
		"message": message,
	}
	s.respondJSON(w, resp, http.StatusOK)
}

// respondJSON sends a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}
