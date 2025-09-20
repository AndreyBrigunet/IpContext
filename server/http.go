package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/andreybrigunet/ipapi/geoip"
)

type Server struct {
	server *http.Server
	geoIP  *geoip.GeoIP
	log    zerolog.Logger
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var ipStr string
	if path == "/" {
		ipStr = s.extractClientIP(r)
	} else {
		ipStr = strings.TrimPrefix(path, "/")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		s.respondError(w, "Invalid IP address", http.StatusBadRequest)
		return
	}

	resp, err := s.geoIP.LookupWithContext(r.Context(), ip.String())
	if err != nil {
		s.log.Error().Err(err).Str("ip", ipStr).Msg("Lookup failed")
		s.respondError(w, "IP lookup failed", http.StatusInternalServerError)
		return
	}
	s.respondJSON(w, resp, http.StatusOK)
}

func (s *Server) extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check CF-Connecting-IP header (Cloudflare)
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}
	
	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	
	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	s.respondJSON(w, health, http.StatusOK)
}

func NewServer(addr string, geoIP *geoip.GeoIP, logger zerolog.Logger) *Server {
	s := &Server{
		geoIP: geoIP,
		log:   logger,
	}

	r := http.NewServeMux()
	r.HandleFunc("/", s.handleRoot)
	r.HandleFunc("/health", s.handleHealth)
	
	// Apply middleware
	handler := s.corsMiddleware(s.loggingMiddleware(s.recoveryMiddleware(r)))
	
	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	s.log.Info().Str("addr", s.server.Addr).Msg("Starting HTTP server")
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) respondError(w http.ResponseWriter, message string, statusCode int) {
	resp := map[string]interface{}{
		"status":  "fail",
		"message": message,
	}

	s.respondJSON(w, resp, statusCode)
}

func (s *Server) respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// Middleware functions
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		s.log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Int("status", wrapped.statusCode).
			Dur("duration", duration).
			Str("user_agent", r.UserAgent()).
			Msg("HTTP request")
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.log.Error().
					Interface("panic", err).
					Str("path", r.URL.Path).
					Msg("Panic recovered")
				
				s.respondError(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
