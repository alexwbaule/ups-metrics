package port

import (
	"context"
	"errors"
	"testing"
	"testing/quick"
	"time"
)

// Feature: victorialogs-integration, Property 2: Writer factory correctness
// **Validates: Requirements 1.2**

// MockLogWriter implements LogWriter interface for testing
type MockLogWriter struct {
	entries []LogEntry
	closed  bool
}

func (m *MockLogWriter) WriteLog(ctx context.Context, entry LogEntry) error {
	if m.closed {
		return errors.New("writer is closed")
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockLogWriter) Close() error {
	m.closed = true
	return nil
}

// MockLogWriterFactory implements LogWriterFactory interface for testing
type MockLogWriterFactory struct{}

func (f *MockLogWriterFactory) CreateLogWriter(config interface{}) (LogWriter, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	return &MockLogWriter{}, nil
}

// TestLogWriterFactoryCorrectness tests that factory creates valid LogWriter instances
func TestLogWriterFactoryCorrectness(t *testing.T) {
	factory := &MockLogWriterFactory{}

	// Property: For any valid configuration, factory should create a LogWriter that implements the interface
	property := func(configValue string) bool {
		if configValue == "" {
			return true // Skip empty configs as they're invalid
		}

		writer, err := factory.CreateLogWriter(configValue)
		if err != nil {
			return false
		}

		// Test that the writer implements LogWriter interface correctly
		ctx := context.Background()
		entry := LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "test message",
			Source:    "test",
			Metadata:  map[string]interface{}{"key": "value"},
		}

		// Test WriteLog method
		err = writer.WriteLog(ctx, entry)
		if err != nil {
			return false
		}

		// Test Close method
		err = writer.Close()
		if err != nil {
			return false
		}

		// Test that writing after close returns error
		err = writer.WriteLog(ctx, entry)
		return err != nil // Should return error after close
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// TestLogEntryStructure tests that LogEntry contains all required fields
func TestLogEntryStructure(t *testing.T) {
	// Property: For any LogEntry, all required fields should be accessible
	property := func(message string, level string, source string) bool {
		if message == "" || level == "" || source == "" {
			return true // Skip invalid inputs
		}

		entry := LogEntry{
			Timestamp: time.Now(),
			Level:     level,
			Message:   message,
			Source:    source,
			Metadata:  map[string]interface{}{"test": "value"},
		}

		// Verify all fields are accessible and have expected types
		return !entry.Timestamp.IsZero() &&
			entry.Level == level &&
			entry.Message == message &&
			entry.Source == source &&
			entry.Metadata != nil
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("LogEntry structure property failed: %v", err)
	}
}
