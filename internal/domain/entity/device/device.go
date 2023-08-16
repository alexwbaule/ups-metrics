package device

import "time"

type Config struct {
	Device `mapstructure:"device"`
	Influx `mapstructure:"influxdb"`
}
type Influx struct {
	Address  string `mapstructure:"address"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
}
type Device struct {
	Interval time.Duration `mapstructure:"interval"`
	Address  string        `mapstructure:"address"`
	LogLevel string        `mapstructure:"log"`
	Login    `mapstructure:"login"`
	Http     `mapstructure:"http"`
}
type Http struct {
	HttpClient `mapstructure:"client"`
}

type HttpClient struct {
	MaxIdleConns          int           `mapstructure:"max_idle_conns"`
	MaxConnsPerHost       int           `mapstructure:"max_conns_per_host"`
	MaxIdleConnsPerHost   int           `mapstructure:"max_idle_conns_per_host"`
	ResponseHeaderTimeout time.Duration `mapstructure:"response_header_timeout"`
	TLSHandshakeTimeout   time.Duration `mapstructure:"tls_handshake_timeout"`
	ExpectContinueTimeout time.Duration `mapstructure:"expect_continue_timeout"`
	DialTimeout           time.Duration `mapstructure:"dial_timeout"`
	DialKeepAlive         time.Duration `mapstructure:"dial_keep_alive"`
	RetryCount            int           `mapstructure:"retry_count"`
	RetryWaitCount        time.Duration `mapstructure:"retry_wait_count"`
	RetryMaxWaitTime      time.Duration `mapstructure:"retry_max_wait_time"`
}

type Login struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Metric struct {
	ResponseStatus string   `json:"responseStatus"`
	UPSType        string   `json:"tipoUPS"`
	Gauges         []Gauges `json:"medidores"`
	States         []States `json:"estados"`
	DeployID       string   `json:"deployId"`
	DeployName     string   `json:"deployName"`
	Alert24HState  string   `json:"alerta24hState"`
}
type Phases struct {
	Value string `json:"valor"`
	Max   string `json:"max"`
	Min   string `json:"min"`
}
type Gauges struct {
	Name   string `json:"nome"`
	Phases Phases `json:"fases"`
	Type   string `json:"tipo"`
	Unit   string `json:"unidade"`
}
type States struct {
	Name  string `json:"nome"`
	Value bool   `json:"valor"`
}

type Authentication struct {
	ResponseStatus string `json:"responseStatus"`
	Token          string `json:"token"`
	RefreshToken   string `json:"refreshToken"`
	DeployID       string `json:"deployId"`
	DeployName     string `json:"deployName"`
	Perfil         string `json:"perfil"`
	Serie          string `json:"serie"`
	Usuario        string `json:"usuario"`
	Code           string `json:"code"`
	Features       struct {
		DeviceType string `json:"deviceType"`
		LedRGB     string `json:"ledRGB"`
	} `json:"features"`
}
