package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/spf13/viper"
)

var (
	defaultInterval              = 10 * time.Second
	defaultMaxIdleConns          = 20
	defaultMaxConnsPerHost       = 10
	defaultMaxIdleConnsPerHost   = 10
	defaultDialTimeout           = 500 * time.Millisecond
	defaultDialKeepAlive         = 90 * time.Second
	defaultResponseHeaderTimeout = 15 * time.Second
	defaultTLSHandshakeTimeout   = 15 * time.Second
	defaultExpectContinueTimeout = 15 * time.Second
	defaultRetryCount            = 3
	defaultRetryWaitCount        = 100 * time.Millisecond
	defaultRetryMaxWaitTime      = 500 * time.Millisecond
)

type Config struct {
	device       *device.Config
	stateManager *StateManager
}

const defaultConfig = `conf/config.yaml`
const defaultCountConfig = `conf/count.yaml`

func NewDefaultConfig() (*Config, error) {
	v := viper.New()
	var config device.Config

	v.SetConfigType("yaml")
	v.SetConfigFile(defaultConfig)
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	err = v.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config file: %w", err)
	}
	setDefaults(&config)

	// Initialize state manager
	stateManager := NewStateManager()
	if err := stateManager.LoadState(); err != nil {
		return nil, fmt.Errorf("error loading application state: %w", err)
	}

	return &Config{
		device:       &config,
		stateManager: stateManager,
	}, err
}

// SaveLastIdConfig saves the last notification ID (legacy function for backward compatibility)
func SaveLastIdConfig(id int) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(defaultCountConfig)

	// Try to read existing config first to preserve other settings
	_ = v.ReadInConfig() // Ignore error if file doesn't exist

	v.Set("last", id)
	v.Set("updated_at", time.Now().Format(time.RFC3339))

	// Ensure directory exists
	dir := filepath.Dir(defaultCountConfig)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Try SafeWriteConfig first, fallback to WriteConfig
	if err := v.SafeWriteConfig(); err != nil {
		// If file exists, WriteConfig will update it
		return v.WriteConfig()
	}
	return nil
}

func (c *Config) GetLastKnowId() int {
	return c.stateManager.GetLastNotificationId()
}

// UpdateLastNotificationId updates the last notification ID and saves state
func (c *Config) UpdateLastNotificationId(id int) error {
	return c.stateManager.UpdateLastNotificationId(id)
}

// SaveState saves the current application state
func (c *Config) SaveState() error {
	return c.stateManager.SaveState()
}

func (c *Config) GetLogLevel() string {
	return c.device.LogLevel
}
func (c *Config) GetInterval() time.Duration {
	return c.device.Interval
}

func (c *Config) GetLogin() device.Login {
	return c.device.Login
}

func (c *Config) GetMetricConfig() device.Metrics {
	return c.device.Metrics
}

func (c *Config) GetGelfConfig() device.Gelf {
	return c.device.Logs.Gelf
}

func (c *Config) GetLogType() string {
	return c.device.Logs.Type
}

func (c *Config) GetVictoriaLogsConfig() device.VictoriaLogs {
	return c.device.Logs.VictoriaLogs
}

func (c *Config) GetLogsConfig() device.Logs {
	return c.device.Logs
}

func (c *Config) GetDeviceAddress() string {
	return c.device.Device.Address
}

func (c *Config) GetHttpClient() device.HttpClient {
	return c.device.Http.HttpClient
}

func setDefaults(cfg *device.Config) {
	if cfg.Interval == 0 {
		cfg.Interval = defaultInterval
	}
	if cfg.HttpClient.MaxIdleConns == 0 {
		cfg.HttpClient.MaxIdleConns = defaultMaxIdleConns
	}
	if cfg.HttpClient.MaxConnsPerHost == 0 {
		cfg.HttpClient.MaxConnsPerHost = defaultMaxConnsPerHost
	}
	if cfg.HttpClient.MaxIdleConnsPerHost == 0 {
		cfg.HttpClient.MaxIdleConnsPerHost = defaultMaxIdleConnsPerHost
	}
	if cfg.HttpClient.ResponseHeaderTimeout == 0 {
		cfg.HttpClient.ResponseHeaderTimeout = defaultResponseHeaderTimeout
	}
	if cfg.HttpClient.TLSHandshakeTimeout == 0 {
		cfg.HttpClient.TLSHandshakeTimeout = defaultTLSHandshakeTimeout
	}
	if cfg.HttpClient.ExpectContinueTimeout == 0 {
		cfg.HttpClient.ExpectContinueTimeout = defaultExpectContinueTimeout
	}
	if cfg.HttpClient.DialTimeout == 0 {
		cfg.HttpClient.DialTimeout = defaultDialTimeout
	}
	if cfg.HttpClient.DialKeepAlive == 0 {
		cfg.HttpClient.DialKeepAlive = defaultDialKeepAlive
	}
	if cfg.HttpClient.RetryCount == 0 {
		cfg.HttpClient.RetryCount = defaultRetryCount
	}
	if cfg.HttpClient.RetryWaitCount == 0 {
		cfg.HttpClient.RetryWaitCount = defaultRetryWaitCount
	}
	if cfg.HttpClient.RetryMaxWaitTime == 0 {
		cfg.HttpClient.RetryMaxWaitTime = defaultRetryMaxWaitTime
	}
}
