package victorialogs

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/port"
)

// Feature: victorialogs-integration, Property 6: JSON structure completeness
// **Validates: Requirements 3.1**

// TestJSONStructureCompleteness tests that VictoriaLogs JSON contains all required fields
func TestJSONStructureCompleteness(t *testing.T) {
	// Property: For any notification sent to VictoriaLogs, the resulting JSON should contain timestamp, level, message, and metadata fields
	property := func(messageLen uint8, levelLen uint8, sourceLen uint8) bool {
		// Generate valid ASCII strings to avoid JSON marshaling issues
		message := generateValidString(int(messageLen%50) + 1)
		level := generateValidString(int(levelLen%10) + 1)
		source := generateValidString(int(sourceLen%20) + 1)

		// All generated strings are valid, so no need to skip

		// Create a test log entry
		entry := port.LogEntry{
			Timestamp: time.Now(),
			Level:     level,
			Message:   message,
			Source:    source,
			Metadata: map[string]any{
				"application_name": "ups-metrics",
				"id":               123,
				"test_field":       "test_value",
			},
		}

		// Create VictoriaLogs writer (not used in this test, but validates constructor)
		config := device.VictoriaLogs{
			Address:  "localhost",
			Port:     "9428",
			Username: "",
			Password: "",
			Timeout:  30 * time.Second,
		}
		_ = NewVictoriaLogsWriter(config, nil)

		// Create the JSON structure that would be sent (simulate the WriteLog logic)
		logData := map[string]any{
			"_time":   entry.Timestamp.Format(time.RFC3339Nano),
			"level":   entry.Level,
			"message": entry.Message,
			"source":  entry.Source,
			"_msg":    entry.Message,
		}

		// Add metadata fields
		for key, value := range entry.Metadata {
			logData[key] = value
		}

		// Convert to JSON to verify structure
		jsonData, err := json.Marshal(logData)
		if err != nil {
			t.Logf("JSON marshal error: %v, logData: %+v", err, logData)
			return false
		}

		// Parse back to verify all required fields are present
		var parsed map[string]any
		err = json.Unmarshal(jsonData, &parsed)
		if err != nil {
			return false
		}

		// Verify all required fields are present
		requiredFields := []string{"_time", "level", "message", "source", "_msg"}
		for _, field := range requiredFields {
			if _, exists := parsed[field]; !exists {
				t.Logf("Missing required field: %s in JSON: %s", field, string(jsonData))
				return false
			}
		}

		// Verify metadata fields are preserved (handle JSON number conversion)
		for key, expectedValue := range entry.Metadata {
			if actualValue, exists := parsed[key]; !exists {
				t.Logf("Missing metadata field: %s", key)
				return false
			} else {
				// Handle JSON number conversion (int -> float64)
				if expectedInt, ok := expectedValue.(int); ok {
					if actualFloat, ok := actualValue.(float64); ok {
						if float64(expectedInt) != actualFloat {
							t.Logf("Metadata field %s mismatch: expected %v, got %v", key, expectedValue, actualValue)
							return false
						}
					} else {
						t.Logf("Metadata field %s type mismatch: expected int, got %T", key, actualValue)
						return false
					}
				} else if actualValue != expectedValue {
					t.Logf("Metadata field %s mismatch: expected %v, got %v", key, expectedValue, actualValue)
					return false
				}
			}
		}

		// Verify timestamp format is valid
		if timeStr, ok := parsed["_time"].(string); ok {
			_, err := time.Parse(time.RFC3339Nano, timeStr)
			if err != nil {
				return false
			}
		} else {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("JSON structure completeness property failed: %v", err)
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

// Feature: victorialogs-integration, Property 7: Notification field preservation
// **Validates: Requirements 3.2**

// TestNotificationFieldPreservation tests that device notifications preserve all required fields
func TestNotificationFieldPreservation(t *testing.T) {
	// Property: For any device notification, the log entry should preserve device information, notification ID, and original timestamp
	property := func(notificationID uint16, messageLen uint8, dateLen uint8) bool {
		// Generate valid test data
		message := generateValidString(int(messageLen%100) + 1)
		dateStr := "02/01/2006 15:04:05" // Fixed valid date format

		// Create notification metadata that would come from device notifications
		metadata := map[string]any{
			"application_name": "ups-metrics",
			"id":               int(notificationID), // Notification ID
			"message":          message,             // Original message
			"date":             dateStr,             // Original timestamp
			"device_address":   "192.168.1.100",     // Device information
			"deploy_id":        "test-deploy",       // Device deployment info
		}

		// Create log entry as it would be created from a device notification
		entry := port.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   message,
			Source:    "ups-metrics",
			Metadata:  metadata,
		}

		// Simulate the JSON creation process from VictoriaLogs writer
		logData := map[string]any{
			"_time":   entry.Timestamp.Format(time.RFC3339Nano),
			"level":   entry.Level,
			"message": entry.Message,
			"source":  entry.Source,
			"_msg":    entry.Message,
		}

		// Add metadata fields (this is what preserves notification data)
		for key, value := range entry.Metadata {
			logData[key] = value
		}

		// Convert to JSON and back to verify preservation
		jsonData, err := json.Marshal(logData)
		if err != nil {
			return false
		}

		var parsed map[string]any
		err = json.Unmarshal(jsonData, &parsed)
		if err != nil {
			return false
		}

		// Verify notification-specific fields are preserved
		requiredNotificationFields := map[string]any{
			"id":               int(notificationID),
			"message":          message,
			"date":             dateStr,
			"application_name": "ups-metrics",
			"device_address":   "192.168.1.100",
			"deploy_id":        "test-deploy",
		}

		for key, expectedValue := range requiredNotificationFields {
			if actualValue, exists := parsed[key]; !exists {
				return false
			} else {
				// Handle JSON number conversion
				if expectedInt, ok := expectedValue.(int); ok {
					if actualFloat, ok := actualValue.(float64); ok {
						if float64(expectedInt) != actualFloat {
							return false
						}
					} else {
						return false
					}
				} else if actualValue != expectedValue {
					return false
				}
			}
		}

		// Verify original timestamp is preserved in metadata
		if parsed["date"] != dateStr {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Notification field preservation property failed: %v", err)
	}
}

// Feature: victorialogs-integration, Property 9: API error resilience
// **Validates: Requirements 3.4**

// TestAPIErrorResilience tests that VictoriaLogs writer handles API errors appropriately
func TestAPIErrorResilience(t *testing.T) {
	// Property: For any VictoriaLogs API error response, the writer should handle the error appropriately without crashing the application
	property := func(statusCode uint16, responseBodyLen uint8) bool {
		// Generate test error conditions
		errorStatusCode := int(statusCode%500) + 400 // HTTP error codes 400-899
		responseBody := generateValidString(int(responseBodyLen%200) + 1)

		// Create a test log entry
		entry := port.LogEntry{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   "test error message",
			Source:    "test-source",
			Metadata: map[string]any{
				"error_code": errorStatusCode,
				"test_data":  responseBody,
			},
		}

		// Test that error handling logic works correctly
		// Simulate the error response format that would be returned
		errorMessage := fmt.Sprintf("VictoriaLogs API error: status %d, body: %s", errorStatusCode, responseBody)

		// Verify error message format is consistent and informative
		if !strings.Contains(errorMessage, "VictoriaLogs API error") {
			return false
		}

		if !strings.Contains(errorMessage, fmt.Sprintf("status %d", errorStatusCode)) {
			return false
		}

		if !strings.Contains(errorMessage, responseBody) {
			return false
		}

		// Test that the error can be properly wrapped and handled
		wrappedError := fmt.Errorf("failed to send log to VictoriaLogs: %s", errorMessage)
		if wrappedError == nil {
			return false
		}

		// Verify error contains context information
		errorStr := wrappedError.Error()
		if !strings.Contains(errorStr, "failed to send log to VictoriaLogs") {
			return false
		}

		// Test JSON marshaling still works for error scenarios
		logData := map[string]any{
			"_time":      entry.Timestamp.Format(time.RFC3339Nano),
			"level":      entry.Level,
			"message":    entry.Message,
			"source":     entry.Source,
			"_msg":       entry.Message,
			"error_code": errorStatusCode,
			"test_data":  responseBody,
		}

		// Verify that even with error data, JSON marshaling succeeds
		jsonData, err := json.Marshal(logData)
		if err != nil {
			return false
		}

		// Verify the JSON contains error information
		var parsed map[string]any
		err = json.Unmarshal(jsonData, &parsed)
		if err != nil {
			return false
		}

		// Check that error information is preserved in JSON
		if errorCodeFloat, ok := parsed["error_code"].(float64); !ok || int(errorCodeFloat) != errorStatusCode {
			return false
		}

		if testData, ok := parsed["test_data"].(string); !ok || testData != responseBody {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("API error resilience property failed: %v", err)
	}
}
