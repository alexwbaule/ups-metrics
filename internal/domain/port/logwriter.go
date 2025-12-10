package port

import (
	"context"
	"time"
)

// LogWriter defines the contract for writing logs to different destinations
type LogWriter interface {
	WriteLog(ctx context.Context, entry LogEntry) error
	Close() error
}

// LogEntry represents a structured log entry with all necessary fields
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// LogWriterFactory defines the contract for creating log writers based on configuration
type LogWriterFactory interface {
	CreateLogWriter(config interface{}) (LogWriter, error)
}
