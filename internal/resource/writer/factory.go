package writer

import (
	"fmt"

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

		// Validate VictoriaLogs configuration
		if logsConfig.Address == "" {
			return nil, fmt.Errorf("victorialogs.address is required when log_type is 'victorialogs'")
		}
		if logsConfig.Port == "" {
			return nil, fmt.Errorf("victorialogs.port is required when log_type is 'victorialogs'")
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
