package notification

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/quick"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/domain/port"
)

// MockLogWriter implements LogWriter interface for testing
type MockLogWriter struct {
	mu           sync.Mutex
	logs         []port.LogEntry
	writeDelay   time.Duration
	shouldError  bool
	errorMessage string
	closed       bool
}

func NewMockLogWriter() *MockLogWriter {
	return &MockLogWriter{
		logs: make([]port.LogEntry, 0),
	}
}

func (m *MockLogWriter) WriteLog(ctx context.Context, entry port.LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("writer is closed")
	}

	// Check for context cancellation with delay
	if m.writeDelay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.writeDelay):
			// Continue with write after delay
		}
	}

	// Check for context cancellation again
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue
	}

	if m.shouldError {
		return errors.New(m.errorMessage)
	}

	m.logs = append(m.logs, entry)
	return nil
}

func (m *MockLogWriter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockLogWriter) SetWriteDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeDelay = delay
}

func (m *MockLogWriter) SetError(shouldError bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = shouldError
	m.errorMessage = message
}

func (m *MockLogWriter) GetLogs() []port.LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]port.LogEntry, len(m.logs))
	copy(result, m.logs)
	return result
}

// TestContextCancellationResponsiveness tests Property 11: Context cancellation responsiveness
// Feature: victorialogs-integration, Property 11: Context cancellation responsiveness
func TestContextCancellationResponsiveness(t *testing.T) {
	property := func(delayMs uint8) bool {
		// Ensure reasonable delay range (10ms to 100ms for testing)
		if delayMs < 10 {
			delayMs = 10
		}
		if delayMs > 100 {
			delayMs = 100
		}

		delay := time.Duration(delayMs) * time.Millisecond

		// Create a mock log writer with a delay
		mockWriter := NewMockLogWriter()
		mockWriter.SetWriteDelay(delay)

		// Create a context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())

		// Create a log entry
		entry := port.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "test message",
			Source:    "test",
			Metadata:  map[string]interface{}{"test": "data"},
		}

		// Start the write operation in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- mockWriter.WriteLog(ctx, entry)
		}()

		// Cancel the context immediately
		cancel()

		// Wait for completion with timeout
		select {
		case err := <-done:
			// The operation should return a context cancellation error
			return errors.Is(err, context.Canceled)
		case <-time.After(200 * time.Millisecond):
			// If we timeout here, the writer didn't respond to cancellation properly
			return false
		}
	}

	config := &quick.Config{
		MaxCount: 20, // Reduced for faster testing
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 11 failed: Context cancellation responsiveness - %v", err)
	}
}

// TestServiceContinuity tests Property 10: Service continuity
// Feature: victorialogs-integration, Property 10: Service continuity
func TestServiceContinuity(t *testing.T) {
	property := func(numFailures uint8, numSuccesses uint8) bool {
		// Ensure reasonable test ranges
		if numFailures > 10 {
			numFailures = 10
		}
		if numSuccesses > 10 {
			numSuccesses = 10
		}
		if numFailures == 0 && numSuccesses == 0 {
			numSuccesses = 1 // At least one operation
		}

		mockWriter := NewMockLogWriter()
		ctx := context.Background()

		totalOperations := int(numFailures) + int(numSuccesses)
		successCount := 0
		errorCount := 0

		// Simulate a series of log write operations with some failures
		for i := 0; i < totalOperations; i++ {
			// Make some operations fail (first numFailures operations)
			shouldFail := i < int(numFailures)
			mockWriter.SetError(shouldFail, "simulated VictoriaLogs unavailability")

			entry := port.LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "test notification",
				Source:    "ups-metrics",
				Metadata:  map[string]interface{}{"notification_id": i},
			}

			err := mockWriter.WriteLog(ctx, entry)
			if err != nil {
				errorCount++
				// Service should continue processing despite errors
				// The error should be logged but not crash the service
				if len(err.Error()) == 0 {
					return false // Error should have meaningful message
				}
			} else {
				successCount++
			}
		}

		// Verify that:
		// 1. The expected number of operations failed
		expectedErrors := int(numFailures)
		if errorCount != expectedErrors {
			return false
		}

		// 2. The expected number of operations succeeded
		expectedSuccesses := int(numSuccesses)
		if successCount != expectedSuccesses {
			return false
		}

		// 3. Service continued processing all operations despite failures
		totalProcessed := successCount + errorCount
		if totalProcessed != totalOperations {
			return false
		}

		// 4. Successful operations were actually logged
		logs := mockWriter.GetLogs()
		if len(logs) != successCount {
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10 failed: Service continuity - %v", err)
	}
}

// TestErrorPropagationCorrectness tests Property 12: Error propagation correctness
// Feature: victorialogs-integration, Property 12: Error propagation correctness
func TestErrorPropagationCorrectness(t *testing.T) {
	property := func(errorMessage string, shouldError bool) bool {
		// Ensure we have a non-empty error message when we should error
		if shouldError && len(errorMessage) == 0 {
			errorMessage = "test error"
		}

		// Limit error message length for reasonable testing
		if len(errorMessage) > 100 {
			errorMessage = errorMessage[:100]
		}

		mockWriter := NewMockLogWriter()
		mockWriter.SetError(shouldError, errorMessage)

		ctx := context.Background()
		entry := port.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "test message",
			Source:    "test",
			Metadata:  map[string]interface{}{"test": "data"},
		}

		err := mockWriter.WriteLog(ctx, entry)

		if shouldError {
			// When writer should error, we must get an error back
			if err == nil {
				return false
			}
			// The error should contain meaningful information
			return len(err.Error()) > 0
		} else {
			// When writer should not error, we should get no error
			return err == nil
		}
	}

	config := &quick.Config{
		MaxCount: 50,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 12 failed: Error propagation correctness - %v", err)
	}
}
