// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// CORSMode defines the CORS behavior
type CORSMode string

const (
	CORSModeStrict      CORSMode = "strict"
	CORSModeDevelopment CORSMode = "development"
	CORSModeDisabled    CORSMode = "disabled"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Mode           CORSMode
	AllowedOrigins []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	GlobalRPS    float64
	GlobalBurst  int
	SessionRPS   float64
	SessionBurst int
}

// GetCORSConfig returns the CORS configuration from environment
func GetCORSConfig() CORSConfig {
	mode := CORSMode(getEnv("MCP_CORS_MODE", "strict"))
	originsStr := getEnv("MCP_ALLOWED_ORIGINS", "")

	var origins []string
	if originsStr != "" {
		origins = strings.Split(originsStr, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
	}

	return CORSConfig{
		Mode:           mode,
		AllowedOrigins: origins,
	}
}

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(config CORSConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		switch config.Mode {
		case CORSModeDisabled:
			// No CORS headers
		case CORSModeDevelopment:
			// Allow all origins in development
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Zabbix-Token, X-Zabbix-URL")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		case CORSModeStrict:
			// Only allow configured origins
			for _, allowed := range config.AllowedOrigins {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Zabbix-Token, X-Zabbix-URL")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					break
				}
			}
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ZabbixContextMiddleware extracts Zabbix configuration from request headers and adds to context
func ZabbixContextMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract from headers
		if url := r.Header.Get(ZabbixHeaderURL); url != "" {
			ctx = context.WithValue(ctx, contextKey(ZabbixURL), url)
		}
		if token := r.Header.Get(ZabbixHeaderToken); token != "" {
			ctx = context.WithValue(ctx, contextKey(ZabbixToken), token)
		}

		// Extract from query parameters (for URL param support)
		if url := r.URL.Query().Get("ZABBIX_URL"); url != "" {
			ctx = context.WithValue(ctx, contextKey(ZabbixURL), url)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(log.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}).Debug("HTTP request")

		next.ServeHTTP(w, r)
	})
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	global  *rate.Limiter
	session map[string]*rate.Limiter
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		global:  rate.NewLimiter(rate.Limit(config.GlobalRPS), config.GlobalBurst),
		session: make(map[string]*rate.Limiter),
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(sessionID string) bool {
	return rl.global.Allow()
}

// GetRateLimitConfig returns rate limit configuration from environment
func GetRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalRPS:    10,
		GlobalBurst:  20,
		SessionRPS:   5,
		SessionBurst: 10,
	}
}

// HealthHandler returns a health check handler
func HealthHandler(logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	}
}

// BuildMiddlewareStack builds the complete middleware stack for HTTP mode
func BuildMiddlewareStack(handler http.Handler, logger *log.Logger) http.Handler {
	corsConfig := GetCORSConfig()

	// Apply middleware from outer to inner
	handler = LoggingMiddleware(logger, handler)
	handler = ZabbixContextMiddleware(logger, handler)
	handler = CORSMiddleware(corsConfig, handler)

	return handler
}

// GetEnvWithFallback is an exported version of getEnv
func GetEnvWithFallback(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
