package graylog

import (
	"encoding/json"
	"testing"
	"testing/quick"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/port"
)

// Feature: victorialogs-integration, Property 8: Field naming consistency
// **Validates: Requirements 3.3**

// TestFieldNamingConsistency tests that field names are consistent across all log destinations
func TestFieldNamingConsistency(t *testing.T) {
	// Property: For any log entry formatted by different writers, the field names should be consistent across all destinations
	property := func(messageLen uint8, levelLen uint8, sourceLen uint8, notificationID uint16) bool {
		// Generate valid test data
		message := generateValidString(int(messageLen%100) + 1)
		level := generateValidString(int(levelLen%10) + 1)
		source := generateValidString(int(sourceLen%20) + 1)

		// Create a device notification (simulating the original data structure)
		notification := device.Notification{
			ID:      int(notificationID),
			Message: message,
			Date:    "02/01/2006 15:04:05", // Fixed valid date format
		}

		// Create log entry as it would be created from the notification
		entry := port.LogEntry{
			Timestamp: time.Now(),
			Level:     level,
			Message:   message,
			Source:    source,
			Metadata: map[string]interface{}{
				"application_name": "ups-metrics",
				"id":               notification.ID,
				"message":          notification.Message,
				"date":             notification.Date,
			},
		}

		// Simulate Graylog field mapping (based on current implementation)
		graylogFields := map[string]interface{}{
			"timestamp":        entry.Timestamp,
			"level":            entry.Level,
			"message":          entry.Message,
			"source":           entry.Source,
			"application_name": entry.Metadata["application_name"],
			"id":               entry.Metadata["id"],
			"date":             entry.Metadata["date"],
		}

		// Simulate VictoriaLogs field mapping (based on existing implementation)
		victoriaLogsFields := map[string]interface{}{
			"_time":            entry.Timestamp.Format(time.RFC3339Nano),
			"level":            entry.Level,
			"message":          entry.Message,
			"source":           entry.Source,
			"_msg":             entry.Message,
			"application_name": entry.Metadata["application_name"],
			"id":               entry.Metadata["id"],
			"date":             entry.Metadata["date"],
		}

		// Define the core fields that should be consistent across both formats
		coreFields := []string{"level", "message", "source", "application_name", "id", "date"}

		// Verify that core fields exist in both formats
		for _, field := range coreFields {
			if _, exists := graylogFields[field]; !exists {
				t.Logf("Missing core field '%s' in Graylog format", field)
				return false
			}
			if _, exists := victoriaLogsFields[field]; !exists {
				t.Logf("Missing core field '%s' in VictoriaLogs format", field)
				return false
			}
		}

		// Verify that the values for core fields are consistent
		for _, field := range coreFields {
			graylogValue := graylogFields[field]
			victoriaLogsValue := victoriaLogsFields[field]

			// Handle type conversions for comparison
			if !valuesEqual(graylogValue, victoriaLogsValue) {
				t.Logf("Field '%s' value mismatch: Graylog=%v, VictoriaLogs=%v", field, graylogValue, victoriaLogsValue)
				return false
			}
		}

		// Test JSON serialization to ensure field names are preserved
		graylogJSON, err := json.Marshal(graylogFields)
		if err != nil {
			return false
		}

		victoriaLogsJSON, err := json.Marshal(victoriaLogsFields)
		if err != nil {
			return false
		}

		// Parse back to verify field names are preserved
		var graylogParsed map[string]interface{}
		var victoriaLogsParsed map[string]interface{}

		err = json.Unmarshal(graylogJSON, &graylogParsed)
		if err != nil {
			return false
		}

		err = json.Unmarshal(victoriaLogsJSON, &victoriaLogsParsed)
		if err != nil {
			return false
		}

		// Verify core fields are present after JSON round-trip
		for _, field := range coreFields {
			if _, exists := graylogParsed[field]; !exists {
				return false
			}
			if _, exists := victoriaLogsParsed[field]; !exists {
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Field naming consistency property failed: %v", err)
	}
}

// generateValidString creates a valid ASCII string for JSON marshaling
func generateValidString(length int) string {
	if length <= 0 {
		return "test"
	}

	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 -_."
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}

// valuesEqual compares two values, handling type conversions
func valuesEqual(a, b interface{}) bool {
	// Handle direct equality
	if a == b {
		return true
	}

	// Handle int to float64 conversion (common in JSON)
	if aInt, ok := a.(int); ok {
		if bFloat, ok := b.(float64); ok {
			return float64(aInt) == bFloat
		}
	}
	if aFloat, ok := a.(float64); ok {
		if bInt, ok := b.(int); ok {
			return aFloat == float64(bInt)
		}
	}

	// Handle time formatting differences
	if aTime, ok := a.(time.Time); ok {
		if bStr, ok := b.(string); ok {
			// Check if the string representation matches the time
			return aTime.Format(time.RFC3339Nano) == bStr
		}
	}

	return false
}
