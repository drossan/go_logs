// Package http provides HTTP endpoints for dynamic log level control.
//
// This package enables runtime log level changes via HTTP endpoints, which is
// essential for debugging production issues without redeploying. It includes
// security features like authentication, rate limiting, and IP whitelisting.
//
// Features:
//   - GET /log-level - Get current log level
//   - PUT /log-level - Set log level
//   - GET /log-level/metrics - Get logging metrics
//   - Bearer token authentication
//   - Rate limiting
//   - IP whitelisting (exact and CIDR)
//
// Example:
//
//	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
//	handler := http.NewDynamicLevelHandler(logger, http.Config{
//	    Endpoint:  "/debug/level",
//	    AuthToken: "secret-token",
//	})
//	http.Handle("/debug/", handler)
//	http.ListenAndServe(":8080", nil)
package http

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	go_logs "github.com/drossan/go_logs"
)

// Config holds configuration for the dynamic level handler.
type Config struct {
	// Endpoint is the base path for the handler (default: "/log-level")
	Endpoint string

	// AuthToken is the Bearer token for authentication (optional)
	AuthToken string

	// RateLimit is the maximum requests per second (0 = unlimited)
	RateLimit int

	// AllowedIPs is a list of allowed IPs or CIDR ranges (empty = all allowed)
	AllowedIPs []string
}

// LevelResponse represents the response for GET /log-level
type LevelResponse struct {
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
}

// LevelRequest represents the request for PUT /log-level
type LevelRequest struct {
	Level string `json:"level"`
}

// MetricsResponse represents the response for GET /log-level/metrics
type MetricsResponse struct {
	Total   int64            `json:"total"`
	ByLevel map[string]int64 `json:"by_level"`
	Dropped int64            `json:"dropped"`
}

// DynamicLevelHandler handles HTTP requests for dynamic log level control.
type DynamicLevelHandler struct {
	logger go_logs.Logger
	config Config

	// Rate limiting
	requests    map[string]int
	rateMu      sync.Mutex
	lastCleanup time.Time

	// Parsed CIDR networks for IP whitelisting
	allowedNets []*net.IPNet
}

// NewDynamicLevelHandler creates a new dynamic level handler.
func NewDynamicLevelHandler(logger go_logs.Logger, config Config) *DynamicLevelHandler {
	h := &DynamicLevelHandler{
		logger:      logger,
		config:      config,
		requests:    make(map[string]int),
		lastCleanup: time.Now(),
	}

	// Parse CIDR networks
	for _, ip := range config.AllowedIPs {
		if strings.Contains(ip, "/") {
			_, network, err := net.ParseCIDR(ip)
			if err == nil {
				h.allowedNets = append(h.allowedNets, network)
			}
		} else {
			// Single IP, convert to /32
			_, network, err := net.ParseCIDR(ip + "/32")
			if err == nil {
				h.allowedNets = append(h.allowedNets, network)
			}
		}
	}

	return h
}

// ServeHTTP implements http.Handler
func (h *DynamicLevelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Check IP whitelist
	if !h.isIPAllowed(r) {
		h.writeError(w, http.StatusForbidden, "IP not allowed")
		return
	}

	// Check authentication
	if !h.isAuthenticated(r) {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Check rate limit
	if !h.checkRateLimit(r) {
		h.writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	// Route based on path and method
	path := r.URL.Path
	method := r.Method

	// Handle metrics endpoint
	if strings.HasSuffix(path, "/metrics") {
		if method != http.MethodGet {
			h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handleMetrics(w, r)
		return
	}

	// Handle log level endpoint
	switch method {
	case http.MethodGet:
		h.handleGetLevel(w, r)
	case http.MethodPut:
		h.handleSetLevel(w, r)
	default:
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleGetLevel handles GET requests to get the current log level
func (h *DynamicLevelHandler) handleGetLevel(w http.ResponseWriter, r *http.Request) {
	level := h.logger.GetLevel()

	resp := LevelResponse{
		Level:     level.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(resp)
}

// handleSetLevel handles PUT requests to set the log level
func (h *DynamicLevelHandler) handleSetLevel(w http.ResponseWriter, r *http.Request) {
	var req LevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Parse level
	level, err := parseLevel(req.Level)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Set level
	h.logger.SetLevel(level)

	resp := LevelResponse{
		Level:     level.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleMetrics handles GET requests to get logging metrics
func (h *DynamicLevelHandler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Get metrics from logger
	mg, ok := h.logger.(interface{ GetMetrics() *go_logs.Metrics })
	if !ok {
		h.writeError(w, http.StatusInternalServerError, "metrics not available")
		return
	}

	metrics := mg.GetMetrics()
	snapshot := metrics.Snapshot()

	byLevel := make(map[string]int64)
	for level, count := range snapshot.ByLevel {
		byLevel[level.String()] = count
	}

	resp := MetricsResponse{
		Total:   snapshot.Total,
		ByLevel: byLevel,
		Dropped: metrics.Dropped(),
	}

	json.NewEncoder(w).Encode(resp)
}

// isIPAllowed checks if the request IP is in the whitelist
func (h *DynamicLevelHandler) isIPAllowed(r *http.Request) bool {
	if len(h.allowedNets) == 0 {
		return true
	}

	// Get client IP
	ipStr := r.RemoteAddr
	if idx := strings.LastIndex(ipStr, ":"); idx != -1 {
		ipStr = ipStr[:idx]
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check against allowed networks
	for _, network := range h.allowedNets {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// isAuthenticated checks if the request has valid authentication
func (h *DynamicLevelHandler) isAuthenticated(r *http.Request) bool {
	if h.config.AuthToken == "" {
		return true
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}

	// Check Bearer token
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	return parts[1] == h.config.AuthToken
}

// checkRateLimit checks if the request is within rate limits
func (h *DynamicLevelHandler) checkRateLimit(r *http.Request) bool {
	if h.config.RateLimit <= 0 {
		return true
	}

	// Get client IP for rate limiting
	ip := r.RemoteAddr

	h.rateMu.Lock()
	defer h.rateMu.Unlock()

	// Cleanup old entries every minute
	if time.Since(h.lastCleanup) > time.Minute {
		h.requests = make(map[string]int)
		h.lastCleanup = time.Now()
	}

	// Check and increment
	if h.requests[ip] >= h.config.RateLimit {
		return false
	}
	h.requests[ip]++

	return true
}

// writeError writes an error response
func (h *DynamicLevelHandler) writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// parseLevel parses a level string into a Level
func parseLevel(s string) (go_logs.Level, error) {
	switch strings.ToLower(s) {
	case "trace":
		return go_logs.TraceLevel, nil
	case "debug":
		return go_logs.DebugLevel, nil
	case "info":
		return go_logs.InfoLevel, nil
	case "warn", "warning":
		return go_logs.WarnLevel, nil
	case "error":
		return go_logs.ErrorLevel, nil
	case "fatal":
		return go_logs.FatalLevel, nil
	case "silent":
		return go_logs.SilentLevel, nil
	default:
		return go_logs.InfoLevel, &invalidLevelError{s}
	}
}

// invalidLevelError represents an invalid level error
type invalidLevelError struct {
	level string
}

func (e *invalidLevelError) Error() string {
	return "invalid log level: " + e.level
}
