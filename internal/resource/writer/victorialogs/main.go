package victorialogs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/port"
	"github.com/alexwbaule/ups-metrics/internal/resource/http/client"
)

// VictoriaLogsWriter implements the LogWriter interface for VictoriaLogs
type VictoriaLogsWriter struct {
	config     device.VictoriaLogs
	httpClient *client.Client
	baseURL    string
}

// NewVictoriaLogsWriter creates a new VictoriaLogs writer instance
func NewVictoriaLogsWriter(config device.VictoriaLogs, httpClient *client.Client) *VictoriaLogsWriter {
	baseURL := fmt.Sprintf("http://%s:%s", config.Address, config.Port)
	return &VictoriaLogsWriter{
		config:     config,
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// WriteLog implements the LogWriter interface to send logs to VictoriaLogs
func (w *VictoriaLogsWriter) WriteLog(ctx context.Context, entry port.LogEntry) error {
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
		return fmt.Errorf("failed to send log to VictoriaLogs: %w", err)
	}

	// Check response status
	if response.StatusCode() >= 400 {
		return fmt.Errorf("VictoriaLogs API error: status %d, body: %s",
			response.StatusCode(), string(response.Body()))
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
