package graylog

import (
	"context"
	"fmt"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/port"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
)

type Gelf struct {
	Address  string
	gelf     *gelf.Writer
	Hostname string
	log      *logger.Logger
}

func NewGelf(l *application.Application) *Gelf {
	cf := l.Config.GetGelfConfig()

	g, err := gelf.NewWriter(fmt.Sprintf("%s:%s", cf.Address, cf.Port))
	if err != nil {
		l.Log.Infof("Error creating Gelf Writer: %s", err.Error())
	}
	return &Gelf{
		Address:  fmt.Sprintf("%s:%s", cf.Address, cf.Port),
		gelf:     g,
		Hostname: l.Config.GetDeviceAddress(),
		log:      l.Log,
	}

}

// WriteLog implements the LogWriter interface
func (m *Gelf) WriteLog(ctx context.Context, entry port.LogEntry) error {
	// Create GELF message from LogEntry
	msg := &gelf.Message{
		Version:  "1.1",
		Host:     m.Hostname,
		Short:    entry.Message,
		TimeUnix: float64(entry.Timestamp.Unix()),
		Level:    m.mapLogLevel(entry.Level),
		Facility: "ups-metrics",
		Extra:    entry.Metadata,
	}

	// Add core fields to extra data for consistency
	if msg.Extra == nil {
		msg.Extra = make(map[string]interface{})
	}
	msg.Extra["source"] = entry.Source
	msg.Extra["level"] = entry.Level
	msg.Extra["message"] = entry.Message

	err := m.gelf.WriteMessage(msg)
	if err != nil {
		m.log.Infof("Error writing message: %s", err.Error())
		return fmt.Errorf("failed to write log to Graylog: %w", err)
	}
	m.log.Infof("Sent log: %s", entry.Message)
	return nil
}

// Close implements the LogWriter interface
func (m *Gelf) Close() error {
	return m.gelf.Close()
}

// mapLogLevel converts string log level to GELF numeric level
func (m *Gelf) mapLogLevel(level string) int32 {
	switch level {
	case "emergency":
		return 0
	case "alert":
		return 1
	case "critical":
		return 2
	case "error":
		return 3
	case "warning":
		return 4
	case "notice":
		return 5
	case "info":
		return 6
	case "debug":
		return 7
	default:
		return 6 // Default to info level
	}
}

// LogNotifications maintains backwards compatibility with existing code
func (m *Gelf) LogNotifications(not device.Notification) {
	var dt time.Time
	extraMessage := map[string]interface{}{
		"application_name": "ups-metrics",
		"id":               not.ID,
		"message":          not.Message,
		"date":             not.Date,
	}

	parse, err := time.ParseInLocation("02/01/2006 15:04:05", not.Date, time.Local)
	if err != nil {
		m.log.Infof("Error parsing date [%s]: %s", not.Date, err.Error())
		dt = time.Now()
	} else {
		dt = parse
	}

	full := fmt.Sprintf("Notification %d on %s with %s", not.ID, not.Date, not.Message)

	// Use the new WriteLog method for consistency
	entry := port.LogEntry{
		Timestamp: dt,
		Level:     "info",
		Message:   full,
		Source:    "ups-metrics",
		Metadata:  extraMessage,
	}

	err = m.WriteLog(context.Background(), entry)
	if err != nil {
		m.log.Infof("Error in LogNotifications: %s", err.Error())
	}
}

// Disconnect maintains backwards compatibility
func (m *Gelf) Disconnect() {
	_ = m.Close()
}
