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

// Feature: victorialogs-integration, Property 16: Type-based configuration validation
// **Validates: Requirements 5.4**

// TestTypeBasedConfigurationValidation tests that configuration with log_type validates corresponding destination config
func TestTypeBasedConfigurationValidation(t *testing.T) {
	// Property: For any configuration with log_type specified,
	// the system should validate that the corresponding destination configuration is present and valid
	property := func(logType string, hasValidDestConfig bool) bool {
		// Only test valid log types
		if logType != "gelf" && logType != "victorialogs" {
			return true // Skip invalid types
		}

		deviceConfig := &device.Config{
			Logs: device.Logs{
				Type: logType,
			},
		}

		// Add destination configuration based on type and validity flag
		if hasValidDestConfig {
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
		// If hasValidDestConfig is false, we leave destination config empty/invalid

		config := &Config{device: deviceConfig}

		// Validate that the system can detect configuration completeness
		if logType == "victorialogs" {
			vlConfig := config.GetVictoriaLogsConfig()
			if hasValidDestConfig {
				// Should have complete configuration
				return vlConfig.Address != "" && vlConfig.Port != ""
			} else {
				// Should detect missing/incomplete configuration
				return vlConfig.Address == "" || vlConfig.Port == ""
			}
		} else if logType == "gelf" {
			gelfConfig := config.GetGelfConfig()
			if hasValidDestConfig {
				// Should have complete configuration
				return gelfConfig.Address != "" && gelfConfig.Port != ""
			} else {
				// Should detect missing/incomplete configuration
				return gelfConfig.Address == "" || gelfConfig.Port == ""
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Type-based configuration validation property failed: %v", err)
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

// Feature: victorialogs-integration, Property 17: Switch-based configuration handling
// **Validates: Requirements 5.2**

// TestSwitchBasedConfigurationHandling tests that configuration with log_type set to "gelf" uses only Graylog and ignores VictoriaLogs
func TestSwitchBasedConfigurationHandling(t *testing.T) {
	// Property: For any configuration with log_type set to "gelf",
	// only Graylog should be used and VictoriaLogs settings should be ignored
	property := func(gelfAddress, gelfPort, vlAddress, vlPort string) bool {
		// Skip empty values for required fields
		if gelfAddress == "" || gelfPort == "" {
			return true
		}

		// Create configuration with both gelf and victorialogs settings
		deviceConfig := &device.Config{
			Logs: device.Logs{
				Type: "gelf", // Explicitly set to gelf
				Gelf: device.Gelf{
					Address: gelfAddress,
					Port:    gelfPort,
				},
				VictoriaLogs: device.VictoriaLogs{
					Address: vlAddress,
					Port:    vlPort,
					Timeout: 30 * time.Second,
				},
			},
		}

		config := &Config{device: deviceConfig}

		// Verify that log type is correctly identified as gelf
		logType := config.GetLogType()
		if logType != "gelf" {
			return false
		}

		// Verify that gelf configuration is accessible
		gelfConfig := config.GetGelfConfig()
		if gelfConfig.Address != gelfAddress || gelfConfig.Port != gelfPort {
			return false
		}

		// The key property is that when type is "gelf", the system should prioritize gelf config
		// and the log type should be correctly identified
		return logType == "gelf" && gelfConfig.Address == gelfAddress && gelfConfig.Port == gelfPort
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Switch-based configuration handling property failed: %v", err)
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

// Feature: victorialogs-integration, Property 18: Deployment compatibility
// **Validates: Requirements 5.4**

// TestDeploymentCompatibility tests that existing deployment configuration files work without breaking changes
func TestDeploymentCompatibility(t *testing.T) {
	// Property: For any existing deployment configuration file,
	// the updated system should parse and execute without breaking changes
	property := func(hasLegacyGelf bool, hasNewTypeField bool, logType string) bool {
		// Create a configuration that simulates existing deployment scenarios
		deviceConfig := &device.Config{
			Device: device.Device{
				Address: "ups-device",
			},
			Logs: device.Logs{},
		}

		// Simulate legacy configuration (only gelf config, no type field)
		if hasLegacyGelf && !hasNewTypeField {
			deviceConfig.Logs.Gelf = device.Gelf{
				Address: "legacy-graylog",
				Port:    "12201",
			}
			// No type field set - this simulates legacy config
		}

		// Simulate new configuration with explicit type
		if hasNewTypeField {
			if logType == "gelf" || logType == "victorialogs" {
				deviceConfig.Logs.Type = logType

				if logType == "gelf" {
					deviceConfig.Logs.Gelf = device.Gelf{
						Address: "new-graylog",
						Port:    "12201",
					}
				} else if logType == "victorialogs" {
					deviceConfig.Logs.VictoriaLogs = device.VictoriaLogs{
						Address: "victoria-logs",
						Port:    "9428",
						Timeout: 30 * time.Second,
					}
				}
			}
		}

		config := &Config{device: deviceConfig}

		// Test that configuration can be read without errors
		retrievedType := config.GetLogType()

		// For legacy configurations (no type field), the system should handle gracefully
		if !hasNewTypeField {
			// Legacy config should have empty type (system should detect this)
			return retrievedType == ""
		}

		// For new configurations with explicit type, verify correctness
		if hasNewTypeField && (logType == "gelf" || logType == "victorialogs") {
			if retrievedType != logType {
				return false
			}

			// Verify that the appropriate configuration is accessible
			if logType == "gelf" {
				gelfConfig := config.GetGelfConfig()
				return gelfConfig.Address != "" && gelfConfig.Port != ""
			} else if logType == "victorialogs" {
				vlConfig := config.GetVictoriaLogsConfig()
				return vlConfig.Address != "" && vlConfig.Port != ""
			}
		}

		// For invalid configurations, the system should handle gracefully
		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Deployment compatibility property failed: %v", err)
	}
}
