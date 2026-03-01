package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	go_logs "github.com/drossan/go_logs"
)

// TestDynamicLevelHandler_GetLevel tests getting the current log level
func TestDynamicLevelHandler_GetLevel(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{})

	req := httptest.NewRequest(http.MethodGet, "/log-level", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp LevelResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Level.String() returns uppercase
	if !strings.EqualFold(resp.Level, "info") {
		t.Errorf("Expected level 'info', got '%s'", resp.Level)
	}
}

// TestDynamicLevelHandler_SetLevel tests setting the log level
func TestDynamicLevelHandler_SetLevel(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{})

	// Set to debug
	body := `{"level": "debug"}`
	req := httptest.NewRequest(http.MethodPut, "/log-level", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Verify level was changed
	if logger.GetLevel() != go_logs.DebugLevel {
		t.Errorf("Expected level DebugLevel, got %v", logger.GetLevel())
	}
}

// TestDynamicLevelHandler_InvalidLevel tests setting an invalid level
func TestDynamicLevelHandler_InvalidLevel(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{})

	body := `{"level": "invalid"}`
	req := httptest.NewRequest(http.MethodPut, "/log-level", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

// TestDynamicLevelHandler_Auth tests authentication
func TestDynamicLevelHandler_Auth(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{
		AuthToken: "secret-token",
	})

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"wrong token", "wrong-token", http.StatusUnauthorized},
		{"correct token", "secret-token", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/log-level", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

// TestDynamicLevelHandler_MethodNotAllowed tests invalid HTTP methods
func TestDynamicLevelHandler_MethodNotAllowed(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{})

	methods := []string{http.MethodPost, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/log-level", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Method %s: expected status 405, got %d", method, rec.Code)
		}
	}
}

// TestDynamicLevelHandler_AllLevels tests all valid log levels
func TestDynamicLevelHandler_AllLevels(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{})

	levels := map[string]go_logs.Level{
		"trace": go_logs.TraceLevel,
		"debug": go_logs.DebugLevel,
		"info":  go_logs.InfoLevel,
		"warn":  go_logs.WarnLevel,
		"error": go_logs.ErrorLevel,
	}

	for name, level := range levels {
		t.Run(name, func(t *testing.T) {
			body := `{"level": "` + name + `"}`
			req := httptest.NewRequest(http.MethodPut, "/log-level", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}

			if logger.GetLevel() != level {
				t.Errorf("Expected level %v, got %v", level, logger.GetLevel())
			}
		})
	}
}

// TestDynamicLevelHandler_CustomEndpoint tests custom endpoint path
func TestDynamicLevelHandler_CustomEndpoint(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{
		Endpoint: "/debug/level",
	})

	// Test custom endpoint
	req := httptest.NewRequest(http.MethodGet, "/debug/level", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// TestDynamicLevelHandler_Metrics tests that metrics endpoint works
func TestDynamicLevelHandler_Metrics(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

	// Log some messages
	logger.Info("test message 1")
	logger.Info("test message 2")
	logger.Error("error message")

	handler := NewDynamicLevelHandler(logger, Config{})

	// Get metrics
	req := httptest.NewRequest(http.MethodGet, "/log-level/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp MetricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Total != 3 {
		t.Errorf("Expected total 3, got %d", resp.Total)
	}
}

// TestDynamicLevelHandler_RateLimit tests rate limiting (if enabled)
func TestDynamicLevelHandler_RateLimit(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{
		RateLimit: 2, // 2 requests per second
	})

	// Make 3 rapid requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/log-level", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if i < 2 && rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i, rec.Code)
		}
		if i >= 2 && rec.Code != http.StatusTooManyRequests {
			t.Errorf("Request %d: expected status 429, got %d", i, rec.Code)
		}
	}
}

// TestDynamicLevelHandler_IPWhitelist tests IP whitelisting
func TestDynamicLevelHandler_IPWhitelist(t *testing.T) {
	logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
	handler := NewDynamicLevelHandler(logger, Config{
		AllowedIPs: []string{"192.168.1.100", "10.0.0.0/8"},
	})

	tests := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{"allowed exact", "192.168.1.100:1234", http.StatusOK},
		{"allowed CIDR", "10.0.0.1:1234", http.StatusOK},
		{"not allowed", "8.8.8.8:1234", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/log-level", nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
