package victorialogs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/port"
	"github.com/alexwbaule/ups-metrics/internal/resource/http/client"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer/resilience"
)

// VictoriaLogsWriter implements the LogWriter interface for VictoriaLogs
type VictoriaLogsWriter struct {
	config         device.VictoriaLogs
	httpClient     *client.Client
	baseURL        string
	retryConfig    resilience.RetryConfig
	circuitBreaker *resilience.CircuitBreaker
}

// NewVictoriaLogsWriter creates a new VictoriaLogs writer instance
func NewVictoriaLogsWriter(config device.VictoriaLogs, httpClient *client.Client) *VictoriaLogsWriter {
	baseURL := fmt.Sprintf("http://%s:%s", config.Address, config.Port)

	// Configure retry logic for VictoriaLogs specific errors
	retryConfig := resilience.DefaultRetryConfig()
	retryConfig.RetryableErrors = append(retryConfig.RetryableErrors,
		"connection reset by peer",
		"broken pipe",
		"no such host",
		"i/o timeout",
		"context deadline exceeded",
	)

	// Create circuit breaker for resilience
	circuitBreaker := resilience.NewCircuitBreaker(resilience.DefaultCircuitBreakerConfig())

	return &VictoriaLogsWriter{
		config:         config,
		httpClient:     httpClient,
		baseURL:        baseURL,
		retryConfig:    retryConfig,
		circuitBreaker: circuitBreaker,
	}
}

// WriteLog implements the LogWriter interface to send logs to VictoriaLogs
func (w *VictoriaLogsWriter) WriteLog(ctx context.Context, entry port.LogEntry) error {
	// Use circuit breaker and retry logic for resilience
	return w.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return resilience.WithRetry(ctx, w.retryConfig, func(ctx context.Context) error {
			return w.writeLogInternal(ctx, entry)
		})
	})
}

// writeLogInternal performs the actual log writing operation
func (w *VictoriaLogsWriter) writeLogInternal(ctx context.Context, entry port.LogEntry) error {
	// Create VictoriaLogs compatible log entry
	logData := map[string]any{
		"_time":   entry.Timestamp.Format(time.RFC3339Nano),
		"level":   entry.Level,
		"message": entry.Message,
		"source":  entry.Source,
		"_msg":    entry.Message, // VictoriaLogs uses _msg as the main message field
	}

	// Add metadata fields
	for key, value := range entry.Metadata {
		logData[key] = value
	}

	// Convert to JSON
	jsonData, err := json.Marshal(logData)
	if err != nil {
		// JSON marshaling errors are not retryable
		return fmt.Errorf("failed to marshal log entry to JSON: %w", err)
	}

	// Create HTTP request
	req := client.Request{
		Url: "/insert/jsonline",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// Add authentication if configured
	if w.config.Username != "" && w.config.Password != "" {
		req.Headers["Authorization"] = fmt.Sprintf("Basic %s",
			basicAuth(w.config.Username, w.config.Password))
	}

	// Create context with timeout
	if w.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, w.config.Timeout)
		defer cancel()
	}

	// Send the request
	response, err := w.httpClient.Post(ctx, req, jsonData, nil)
	if err != nil {
		// Log the error for monitoring
		w.logError("HTTP request failed", err, entry)
		return fmt.Errorf("failed to send log to VictoriaLogs: %w", err)
	}

	// Check response status
	if response.StatusCode() >= 400 {
		errorMsg := fmt.Sprintf("VictoriaLogs API error: status %d, body: %s",
			response.StatusCode(), string(response.Body()))

		// Log the API error for monitoring
		w.logError("VictoriaLogs API error", fmt.Errorf(errorMsg), entry)

		// Determine if this is a retryable error
		if w.isRetryableStatusCode(response.StatusCode()) {
			return fmt.Errorf(errorMsg)
		}

		// Non-retryable error (e.g., 400 Bad Request)
		return fmt.Errorf("non-retryable VictoriaLogs API error: %s", errorMsg)
	}

	return nil
}

// Close implements the LogWriter interface
func (w *VictoriaLogsWriter) Close() error {
	// VictoriaLogs writer doesn't need explicit cleanup
	// The HTTP client handles connection pooling
	return nil
}

// basicAuth creates a basic authentication string
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// isRetryableStatusCode determines if an HTTP status code should trigger a retry
func (w *VictoriaLogsWriter) isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case 429: // Too Many Requests
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	case 408: // Request Timeout
		return true
	default:
		return false
	}
}

// logError logs errors for monitoring and debugging purposes
func (w *VictoriaLogsWriter) logError(operation string, err error, entry port.LogEntry) {
	// Get circuit breaker state for context
	state, failures, successes := w.circuitBreaker.GetStats()
	stateStr := "closed"
	switch state {
	case resilience.StateOpen:
		stateStr = "open"
	case resilience.StateHalfOpen:
		stateStr = "half-open"
	}

	// Log error with context for monitoring
	log.Printf("[VictoriaLogsWriter] %s: %v | Circuit: %s (failures: %d, successes: %d) | Entry: %s/%s | Target: %s:%s",
		operation, err, stateStr, failures, successes,
		entry.Level, entry.Source, w.config.Address, w.config.Port)
}

// GetCircuitBreakerState returns the current circuit breaker state for monitoring
func (w *VictoriaLogsWriter) GetCircuitBreakerState() (state string, failures int, successes int) {
	cbState, f, s := w.circuitBreaker.GetStats()
	switch cbState {
	case resilience.StateClosed:
		state = "closed"
	case resilience.StateOpen:
		state = "open"
	case resilience.StateHalfOpen:
		state = "half-open"
	default:
		state = "unknown"
	}
	return state, f, s
}

// ResetCircuitBreaker resets the circuit breaker (useful for manual recovery)
func (w *VictoriaLogsWriter) ResetCircuitBreaker() {
	w.circuitBreaker.Reset()
	log.Printf("[VictoriaLogsWriter] Circuit breaker reset for %s:%s", w.config.Address, w.config.Port)
}
