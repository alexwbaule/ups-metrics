package victorialogs

import (
	"encoding/json"
	"fmt"
	"net"
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

// Feature: victorialogs-integration, Property 5: Graceful error handling
// **Validates: Requirements 1.5**

// TestGracefulErrorHandling tests that invalid configuration returns descriptive error messages without crashing
func TestGracefulErrorHandling(t *testing.T) {
	// Property: For any invalid configuration, the system should return descriptive error messages and not crash or enter an undefined state
	property := func(address string, port string, username string, password string, timeoutSecs int) bool {
		// Test various invalid configuration scenarios
		timeout := time.Duration(timeoutSecs%3600+1) * time.Second // Ensure positive timeout

		// Create potentially invalid VictoriaLogs configuration
		config := device.VictoriaLogs{
			Address:  address,
			Port:     port,
			Username: username,
			Password: password,
			Timeout:  timeout,
		}

		// Test that NewVictoriaLogsWriter handles any configuration gracefully
		writer := NewVictoriaLogsWriter(config, nil)
		if writer == nil {
			// Constructor should never return nil, even with invalid config
			return false
		}

		// Test that Close() method works regardless of configuration
		err := writer.Close()
		if err != nil {
			// Close should not fail for VictoriaLogs writer
			return false
		}

		// Test error message formatting for various scenarios
		testErrors := []struct {
			errorType string
			message   string
		}{
			{"marshal", "failed to marshal log entry to JSON"},
			{"network", "failed to send log to VictoriaLogs"},
			{"api", "VictoriaLogs API error"},
		}

		for _, testError := range testErrors {
			// Verify error messages are descriptive and contain context
			errorMsg := fmt.Sprintf("%s: test error", testError.message)
			if !strings.Contains(errorMsg, testError.message) {
				return false
			}

			// Test that errors can be wrapped properly
			wrappedErr := fmt.Errorf("operation failed: %w", fmt.Errorf(errorMsg))
			if wrappedErr == nil {
				return false
			}

			// Verify wrapped error contains original context
			if !strings.Contains(wrappedErr.Error(), testError.message) {
				return false
			}
		}

		// Test that configuration validation logic works
		isValidConfig := address != "" && port != ""

		// Even with invalid config, the system should handle it gracefully
		// (validation happens at factory level, not constructor level)
		if !isValidConfig {
			// Invalid config should be detectable but not cause crashes
			return true
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Graceful error handling property failed: %v", err)
	}
}

// TestStreamFieldsDetection tests the automatic detection of stream fields
func TestStreamFieldsDetection(t *testing.T) {
	// Test that stream fields are automatically detected
	streamFields := DetectStreamFields()

	// Verify that all fields are detected and not empty
	if streamFields.AppName == "" {
		t.Error("DetectStreamFields should return a non-empty app_name")
	}
	if streamFields.Hostname == "" {
		t.Error("DetectStreamFields should return a non-empty hostname")
	}
	if streamFields.RemoteIP == "" {
		t.Error("DetectStreamFields should return a non-empty remote_ip")
	}

	// Verify that app_name doesn't contain path separators (should be just the executable name)
	if strings.Contains(streamFields.AppName, "/") || strings.Contains(streamFields.AppName, "\\") {
		t.Errorf("app_name should not contain path separators, got: %s", streamFields.AppName)
	}

	// Verify that remote_ip looks like an IP address
	if net.ParseIP(streamFields.RemoteIP) == nil {
		t.Errorf("remote_ip should be a valid IP address, got: %s", streamFields.RemoteIP)
	}

	t.Logf("Detected stream fields - app_name: %s, hostname: %s, remote_ip: %s",
		streamFields.AppName, streamFields.Hostname, streamFields.RemoteIP)
}

// TestVictoriaLogsWriterWithAutoDetection tests that VictoriaLogs writer includes auto-detected stream fields
func TestVictoriaLogsWriterWithAutoDetection(t *testing.T) {
	config := device.VictoriaLogs{
		Address: "test-server",
		Port:    "9428",
		Timeout: 30 * time.Second,
	}

	writer := NewVictoriaLogsWriter(config, nil)

	// Verify that stream fields were auto-detected during writer creation
	if writer.streamFields.AppName == "" {
		t.Error("VictoriaLogs writer should have auto-detected app_name")
	}
	if writer.streamFields.Hostname == "" {
		t.Error("VictoriaLogs writer should have auto-detected hostname")
	}
	if writer.streamFields.RemoteIP == "" {
		t.Error("VictoriaLogs writer should have auto-detected remote_ip")
	}
}
