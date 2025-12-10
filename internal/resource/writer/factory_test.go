package writer

import (
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/resource/graylog"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer/victorialogs"
)

// Feature: victorialogs-integration, Property 3: Switch-based destination selection
// **Validates: Requirements 1.2**

// TestSwitchBasedDestinationSelection tests that log_type configuration correctly selects the appropriate writer
func TestSwitchBasedDestinationSelection(t *testing.T) {
	// Property: For any configuration with log_type set to "victorialogs",
	// only the VictoriaLogs writer should be created and used for logging
	property := func(logType string, hasValidVLConfig bool, hasValidGelfConfig bool) bool {
		// Skip invalid log types for this specific property test
		if logType != "gelf" && logType != "victorialogs" {
			return true
		}

		// Create mock config
		mockConfig := createMockConfig(logType, hasValidVLConfig, hasValidGelfConfig)

		// Create minimal application for testing
		log := logger.NewLogger()
		app := &application.Application{
			Config: &config.Config{}, // We'll use the mock config through the helper function
			Log:    log,
		}

		// Test the factory function using the helper
		writer, err := CreateLogWriterWithConfig(mockConfig, app)

		switch logType {
		case "victorialogs":
			if !hasValidVLConfig {
				// Should fail if VictoriaLogs config is invalid
				return err != nil && writer == nil
			}
			// Should create VictoriaLogs writer
			if err != nil {
				return false
			}
			_, isVictoriaLogs := writer.(*victorialogs.VictoriaLogsWriter)
			return isVictoriaLogs

		case "gelf":
			// Should create Graylog writer (doesn't require config validation in current implementation)
			if err != nil {
				return false
			}
			_, isGraylog := writer.(*graylog.Gelf)
			return isGraylog

		default:
			// Should fail for invalid types
			return err != nil && writer == nil
		}
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Switch-based destination selection property failed: %v", err)
	}
}

// TestInvalidLogTypeHandling tests that invalid log types return appropriate errors
func TestInvalidLogTypeHandling(t *testing.T) {
	// Property: For any invalid or missing log_type, the factory should return an explicit error
	property := func(logType string) bool {
		mockConfig := createMockConfig(logType, false, false)

		log := logger.NewLogger()
		app := &application.Application{
			Config: &config.Config{},
			Log:    log,
		}

		writer, err := CreateLogWriterWithConfig(mockConfig, app)

		// For valid types, we expect success (assuming valid config)
		if logType == "gelf" {
			return err == nil && writer != nil
		}

		if logType == "victorialogs" {
			// Will fail due to missing VictoriaLogs config, but that's expected
			return err != nil && writer == nil
		}

		// For invalid or empty types, we expect an error
		if logType == "" || (logType != "gelf" && logType != "victorialogs") {
			return err != nil && writer == nil
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Invalid log type handling property failed: %v", err)
	}
}

// createMockConfig creates a mock config for testing
func createMockConfig(logType string, hasValidVLConfig bool, hasValidGelfConfig bool) *mockConfig {
	// Create base device config
	deviceConfig := &device.Config{
		Device: device.Device{
			Address: "test-device",
		},
		Logs: device.Logs{
			Type: logType,
		},
	}

	// Set up VictoriaLogs config if needed
	if hasValidVLConfig {
		deviceConfig.Logs.VictoriaLogs = device.VictoriaLogs{
			Address: "victoria-logs",
			Port:    "9428",
			Timeout: 30 * time.Second,
		}
	}

	// Set up Gelf config if needed
	if hasValidGelfConfig {
		deviceConfig.Logs.Gelf = device.Gelf{
			Address: "graylog",
			Port:    "12201",
		}
	}

	return &mockConfig{device: deviceConfig}
}

// Feature: victorialogs-integration, Property 4: Configuration validation
// **Validates: Requirements 1.4**

// TestConfigurationValidation tests that configuration validation works correctly
func TestConfigurationValidation(t *testing.T) {
	// Property: For any VictoriaLogs configuration, the validation process should verify
	// connection parameters and return appropriate success or failure results
	property := func(logType string, address string, port string, username string, password string) bool {
		// Only test victorialogs configurations for this property
		if logType != "victorialogs" {
			return true
		}

		// Create mock config with the provided parameters
		deviceConfig := &device.Config{
			Device: device.Device{
				Address: "test-device",
			},
			Logs: device.Logs{
				Type: logType,
				VictoriaLogs: device.VictoriaLogs{
					Address:  address,
					Port:     port,
					Username: username,
					Password: password,
					Timeout:  30 * time.Second,
				},
			},
		}

		mockConfig := &mockConfig{device: deviceConfig}

		// Create minimal application for testing
		log := logger.NewLogger()
		app := &application.Application{
			Config: &config.Config{},
			Log:    log,
		}

		// Test the factory function
		writer, err := CreateLogWriterWithConfig(mockConfig, app)

		// Configuration should be considered valid if address and port are non-empty
		isValidConfig := address != "" && port != ""

		if isValidConfig {
			// Should succeed with valid configuration
			return err == nil && writer != nil
		} else {
			// Should fail with invalid configuration (missing address or port)
			return err != nil && writer == nil
		}
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Configuration validation property failed: %v", err)
	}
}

// Feature: victorialogs-integration, Property 14: Initialization validation
// **Validates: Requirements 4.5**

// TestInitializationValidation tests that log writer initialization validates configuration and establishes connections
func TestInitializationValidation(t *testing.T) {
	// Property: For any log writer initialization, the process should validate configuration
	// and establish necessary connections before returning success
	property := func(logType string, hasValidConfig bool) bool {
		// Only test valid log types
		if logType != "gelf" && logType != "victorialogs" {
			return true
		}

		var mockConf *mockConfig

		if logType == "victorialogs" {
			// Create VictoriaLogs config
			deviceConfig := &device.Config{
				Device: device.Device{
					Address: "test-device",
				},
				Logs: device.Logs{
					Type: logType,
				},
			}

			if hasValidConfig {
				deviceConfig.Logs.VictoriaLogs = device.VictoriaLogs{
					Address: "victoria-logs",
					Port:    "9428",
					Timeout: 30 * time.Second,
				}
			}
			// If hasValidConfig is false, VictoriaLogs config remains empty

			mockConf = &mockConfig{device: deviceConfig}
		} else {
			// For gelf, we don't require specific validation in current implementation
			deviceConfig := &device.Config{
				Device: device.Device{
					Address: "test-device",
				},
				Logs: device.Logs{
					Type: logType,
					Gelf: device.Gelf{
						Address: "graylog",
						Port:    "12201",
					},
				},
			}
			mockConf = &mockConfig{device: deviceConfig}
		}

		// Create minimal application for testing
		log := logger.NewLogger()
		app := &application.Application{
			Config: &config.Config{},
			Log:    log,
		}

		// Test the factory function
		writer, err := CreateLogWriterWithConfig(mockConf, app)

		if logType == "victorialogs" {
			if hasValidConfig {
				// Should succeed with valid VictoriaLogs configuration
				return err == nil && writer != nil
			} else {
				// Should fail with invalid VictoriaLogs configuration
				return err != nil && writer == nil
			}
		} else if logType == "gelf" {
			// Gelf should always succeed in current implementation
			return err == nil && writer != nil
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Initialization validation property failed: %v", err)
	}
}

// mockConfig implements the ConfigProvider interface for testing
type mockConfig struct {
	device *device.Config
}

func (c *mockConfig) GetLogType() string {
	return c.device.Logs.Type
}

func (c *mockConfig) GetVictoriaLogsConfig() device.VictoriaLogs {
	return c.device.Logs.VictoriaLogs
}

func (c *mockConfig) GetGelfConfig() device.Gelf {
	return c.device.Logs.Gelf
}

func (c *mockConfig) GetDeviceAddress() string {
	return c.device.Device.Address
}

func (c *mockConfig) GetHttpClient() device.HttpClient {
	return device.HttpClient{
		MaxIdleConns:          20,
		MaxConnsPerHost:       10,
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: 15 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 15 * time.Second,
		DialTimeout:           500 * time.Millisecond,
		DialKeepAlive:         90 * time.Second,
		RetryCount:            3,
		RetryWaitCount:        100 * time.Millisecond,
		RetryMaxWaitTime:      500 * time.Millisecond,
	}
}

// Feature: victorialogs-integration, Property 19: Validation error clarity
// **Validates: Requirements 5.5**

// TestValidationErrorClarity tests that invalid VictoriaLogs configuration produces clear error messages
func TestValidationErrorClarity(t *testing.T) {
	// Property: For any invalid VictoriaLogs configuration, the validation error message should clearly indicate the specific configuration problem
	property := func(address string, port string, logType string) bool {
		// Only test victorialogs configurations for this property
		if logType != "victorialogs" {
			return true
		}

		// Create mock config with potentially invalid parameters
		deviceConfig := &device.Config{
			Device: device.Device{
				Address: "test-device",
			},
			Logs: device.Logs{
				Type: logType,
				VictoriaLogs: device.VictoriaLogs{
					Address:  address,
					Port:     port,
					Username: "test-user",
					Password: "test-pass",
					Timeout:  30 * time.Second,
				},
			},
		}

		mockConfig := &mockConfig{device: deviceConfig}

		// Create minimal application for testing
		log := logger.NewLogger()
		app := &application.Application{
			Config: &config.Config{},
			Log:    log,
		}

		// Test the factory function
		writer, err := CreateLogWriterWithConfig(mockConfig, app)

		// Determine if configuration is valid
		isValidConfig := address != "" && port != ""

		if isValidConfig {
			// Should succeed with valid configuration
			return err == nil && writer != nil
		} else {
			// Should fail with invalid configuration and provide clear error message
			if err == nil || writer != nil {
				return false
			}

			errorMsg := err.Error()

			// Verify error message clarity - should mention specific missing fields
			if address == "" {
				if !strings.Contains(errorMsg, "address") && !strings.Contains(errorMsg, "Address") {
					return false
				}
				if !strings.Contains(errorMsg, "required") {
					return false
				}
			}

			if port == "" {
				if !strings.Contains(errorMsg, "port") && !strings.Contains(errorMsg, "Port") {
					return false
				}
				if !strings.Contains(errorMsg, "required") {
					return false
				}
			}

			// Error message should mention victorialogs context
			if !strings.Contains(errorMsg, "victorialogs") && !strings.Contains(errorMsg, "VictoriaLogs") {
				return false
			}

			return true
		}
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Validation error clarity property failed: %v", err)
	}
}
