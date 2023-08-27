package config

import (
	"fmt"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/spf13/viper"
	"time"
)

var (
	defaultInterval              = 10 * time.Second
	defaultMaxIdleConns          = 20
	defaultMaxConnsPerHost       = 10
	defaultMaxIdleConnsPerHost   = 10
	defaultDialTimeout           = 500 * time.Millisecond
	defaultDialKeepAlive         = 90 * time.Second
	defaultResponseHeaderTimeout = 10 * time.Second
	defaultTLSHandshakeTimeout   = 10 * time.Second
	defaultExpectContinueTimeout = 10 * time.Second
	defaultRetryCount            = 3
	defaultRetryWaitCount        = 100 * time.Millisecond
	defaultRetryMaxWaitTime      = 500 * time.Millisecond
)

type Config struct {
	device *device.Config
}

const defaultConfig = `conf/config.yaml`

func NewDefaultConfig() (*Config, error) {
	var config device.Config

	viper.SetConfigType("yaml")
	viper.SetConfigFile(defaultConfig)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config file: %w", err)
	}
	setDefaults(&config)

	return &Config{
		device: &config,
	}, err
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

func (c *Config) GetMetricConfig() device.Influx {
	return c.device.Influx
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
