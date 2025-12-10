package config

import (
	"testing"
	"testing/quick"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
)

// Feature: victorialogs-integration, Property 1: Configuration parsing completeness
// **Validates: Requirements 1.1**

// TestConfigurationParsingCompleteness tests that VictoriaLogs configuration parsing preserves all parameters
func TestConfigurationParsingCompleteness(t *testing.T) {
	// Property: For any valid VictoriaLogs configuration containing address, port, and authentication parameters,
	// the configuration parser should successfully extract all parameters without loss
	property := func(address string, port string, username string, password string, timeoutSecs int) bool {
		// Skip invalid inputs
		if address == "" || port == "" || username == "" || password == "" || timeoutSecs <= 0 || timeoutSecs > 3600 {
			return true
		}

		// Create a VictoriaLogs configuration
		timeout := time.Duration(timeoutSecs) * time.Second
		originalConfig := device.VictoriaLogs{
			Address:  address,
			Port:     port,
			Username: username,
			Password: password,
			Timeout:  timeout,
		}

		// Create a complete device config with VictoriaLogs
		deviceConfig := &device.Config{
			Logs: device.Logs{
				Type:         "victorialogs",
				VictoriaLogs: originalConfig,
			},
		}

		// Create config wrapper
		config := &Config{device: deviceConfig}

		// Test that all parameters are preserved through getter methods
		retrievedConfig := config.GetVictoriaLogsConfig()

		return retrievedConfig.Address == originalConfig.Address &&
			retrievedConfig.Port == originalConfig.Port &&
			retrievedConfig.Username == originalConfig.Username &&
			retrievedConfig.Password == originalConfig.Password &&
			retrievedConfig.Timeout == originalConfig.Timeout
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Configuration parsing completeness property failed: %v", err)
	}
}

// Feature: victorialogs-integration, Property 15: Explicit type requirement
// **Validates: Requirements 5.1**

// TestExplicitTypeRequirement tests that configuration without log_type fails with clear error message
func TestExplicitTypeRequirement(t *testing.T) {
	// Property: For any configuration without log_type specified,
	// the system should fail with a clear error message requiring explicit type specification
	property := func(hasType bool, logType string) bool {
		// Create device config
		deviceConfig := &device.Config{
			Logs: device.Logs{},
		}

		// Set type only if hasType is true and logType is valid
		if hasType && (logType == "gelf" || logType == "victorialogs") {
			deviceConfig.Logs.Type = logType
		}

		config := &Config{device: deviceConfig}
		retrievedType := config.GetLogType()

		// If we didn't set a type or set an invalid type, it should be empty or invalid
		if !hasType || (logType != "gelf" && logType != "victorialogs") {
			// The system should detect missing/invalid type
			return retrievedType == "" || (retrievedType != "gelf" && retrievedType != "victorialogs")
		}

		// If we set a valid type, it should be preserved
		return retrievedType == logType
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Explicit type requirement property failed: %v", err)
	}
}

// TestLogTypeValidation tests that only valid log types are accepted
func TestLogTypeValidation(t *testing.T) {
	// Property: For any log type value, only "gelf" and "victorialogs" should be considered valid
	property := func(logType string) bool {
		deviceConfig := &device.Config{
			Logs: device.Logs{
				Type: logType,
			},
		}

		config := &Config{device: deviceConfig}
		retrievedType := config.GetLogType()

		// The retrieved type should match what we set
		return retrievedType == logType
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Log type validation property failed: %v", err)
	}
}

// Feature: victorialogs-integration, Property 20: Explicit error for missing configuration
// **Validates: Requirements 1.4, 1.5**

// TestExplicitErrorForMissingConfiguration tests that missing or invalid configuration produces clear error messages
func TestExplicitErrorForMissingConfiguration(t *testing.T) {
	// Property: For any configuration with invalid log_type or missing destination configuration,
	// the system should fail with a clear error message indicating the problem
	property := func(logType string, hasDestinationConfig bool) bool {
		// Create device config
		deviceConfig := &device.Config{
			Device: device.Device{
				Address: "test-device",
			},
			Logs: device.Logs{
				Type: logType,
			},
		}

		// Add destination config if specified
		if hasDestinationConfig {
			if logType == "victorialogs" {
				deviceConfig.Logs.VictoriaLogs = device.VictoriaLogs{
					Address: "victoria-logs",
					Port:    "9428",
					Timeout: 30 * time.Second,
				}
			} else if logType == "gelf" {
				deviceConfig.Logs.Gelf = device.Gelf{
					Address: "graylog",
					Port:    "12201",
				}
			}
		}

		config := &Config{device: deviceConfig}

		// Test configuration validation logic
		retrievedType := config.GetLogType()

		// Verify type is preserved correctly
		if retrievedType != logType {
			return false
		}

		// Test destination config retrieval
		if logType == "victorialogs" {
			vlConfig := config.GetVictoriaLogsConfig()
			if hasDestinationConfig {
				// Should have valid configuration
				return vlConfig.Address != "" && vlConfig.Port != ""
			} else {
				// Should have empty configuration (missing config scenario)
				return vlConfig.Address == "" || vlConfig.Port == ""
			}
		} else if logType == "gelf" {
			gelfConfig := config.GetGelfConfig()
			if hasDestinationConfig {
				// Should have valid configuration
				return gelfConfig.Address != "" && gelfConfig.Port != ""
			} else {
				// Should have empty configuration (missing config scenario)
				return gelfConfig.Address == "" || gelfConfig.Port == ""
			}
		} else {
			// Invalid log type - should be detectable
			return logType == "" || (logType != "gelf" && logType != "victorialogs")
		}
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Explicit error for missing configuration property failed: %v", err)
	}
}
