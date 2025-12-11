package writer

import (
	"fmt"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/port"
	"github.com/alexwbaule/ups-metrics/internal/resource/graylog"
	"github.com/alexwbaule/ups-metrics/internal/resource/http/client"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer/victorialogs"
)

// ConfigProvider defines the interface for configuration access
type ConfigProvider interface {
	GetLogType() string
	GetVictoriaLogsConfig() device.VictoriaLogs
	GetGelfConfig() device.Gelf
	GetDeviceAddress() string
	GetHttpClient() device.HttpClient
}

// CreateLogWriter creates the appropriate log writer based on configuration
func CreateLogWriter(app *application.Application) (port.LogWriter, error) {
	return CreateLogWriterWithConfig(app.Config, app)
}

// CreateLogWriterWithConfig creates the appropriate log writer with a config provider
func CreateLogWriterWithConfig(config ConfigProvider, app *application.Application) (port.LogWriter, error) {
	logType := config.GetLogType()

	// Validate that log_type is specified
	if logType == "" {
		return nil, fmt.Errorf("log_type must be explicitly specified in configuration (supported values: 'gelf', 'victorialogs')")
	}

	switch logType {
	case "gelf":
		// Create Graylog writer
		gelfWriter := graylog.NewGelf(app)
		return gelfWriter, nil

	case "victorialogs":
		// Create VictoriaLogs writer
		logsConfig := config.GetVictoriaLogsConfig()

		// Enhanced validation for VictoriaLogs configuration
		if err := validateVictoriaLogsConfig(logsConfig); err != nil {
			return nil, fmt.Errorf("VictoriaLogs configuration validation failed: %w", err)
		}

		// Create HTTP client for VictoriaLogs
		baseURL := fmt.Sprintf("http://%s:%s", logsConfig.Address, logsConfig.Port)
		httpClient := client.New(app.Config, baseURL, app.Log)
		vlWriter := victorialogs.NewVictoriaLogsWriter(logsConfig, httpClient)
		return vlWriter, nil

	default:
		return nil, fmt.Errorf("invalid log_type '%s': supported values are 'gelf' or 'victorialogs'", logType)
	}
}

// validateVictoriaLogsConfig performs comprehensive validation of VictoriaLogs configuration
func validateVictoriaLogsConfig(config device.VictoriaLogs) error {
	var errors []string

	// Validate required fields
	if config.Address == "" {
		errors = append(errors, "address is required")
	}
	if config.Port == "" {
		errors = append(errors, "port is required")
	}

	// Validate timeout if specified
	if config.Timeout < 0 {
		errors = append(errors, "timeout must be positive")
	}
	if config.Timeout > 0 && config.Timeout < time.Second {
		errors = append(errors, "timeout should be at least 1 second for reliable operation")
	}

	// Validate authentication consistency
	if (config.Username == "") != (config.Password == "") {
		errors = append(errors, "both username and password must be provided together, or both left empty")
	}

	// Note: Stream fields (app_name, hostname, remote_ip) are automatically detected
	// No validation needed as they are generated at runtime

	// Return combined error if any validation failed
	if len(errors) > 0 {
		return fmt.Errorf("VictoriaLogs configuration errors: %s", joinErrors(errors))
	}

	return nil
}

// joinErrors joins multiple error messages into a readable format
func joinErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	if len(errors) == 1 {
		return errors[0]
	}

	result := ""
	for i, err := range errors {
		if i == 0 {
			result = err
		} else if i == len(errors)-1 {
			result += " and " + err
		} else {
			result += ", " + err
		}
	}
	return result
}
