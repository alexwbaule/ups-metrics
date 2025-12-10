package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/port"
	"github.com/alexwbaule/ups-metrics/internal/resource/smsups"
)

type GetNotification struct {
	log       *logger.Logger
	config    *config.Config
	sms       *smsups.SMSUps
	logWriter port.LogWriter
	last      int
}

func NewGetNotification(log *logger.Logger, cfg *config.Config, sms *smsups.SMSUps, logWriter port.LogWriter) *GetNotification {
	return &GetNotification{
		log:       log,
		config:    cfg,
		sms:       sms,
		logWriter: logWriter,
		last:      cfg.GetLastKnowId(),
	}
}

func (g *GetNotification) Run(ctx context.Context) error {
	ticker := time.NewTicker(g.config.GetInterval())
	defer ticker.Stop()

	// Ensure log writer is properly closed when context is done
	defer func() {
		if g.logWriter != nil {
			if err := g.logWriter.Close(); err != nil {
				g.log.Errorf("error closing log writer: %s", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			g.log.Infof("stopping get notifications job...")
			return ctx.Err()
		case <-ticker.C:
		}
		err := g.getStats(ctx)
		if err != nil {
			g.log.Errorf("get notifications error: %s", err)
			return err
		}
	}
}

func (g *GetNotification) getStats(ctx context.Context) error {
	n, err := g.sms.GetNotifications(ctx)
	if err != nil {
		return err
	}
	g.log.Infof("sending notifications bigger than %d", g.last)

	s := len(n.Notifications) - 1

	for i := s; i >= 0; i-- {
		notification := n.Notifications[i]
		if notification.ID > g.last {
			g.log.Infof("sending notifications id: %d", notification.ID)

			// Convert notification to LogEntry and use LogWriter interface
			err := g.writeNotificationLog(ctx, notification)
			if err != nil {
				g.log.Errorf("error writing notification log: %s", err)
				// Continue processing other notifications even if one fails
				continue
			}

			g.last = notification.ID
		}
	}
	return nil
}

func (g *GetNotification) LastId() int {
	return g.last
}

// writeNotificationLog converts a device notification to a LogEntry and writes it using the LogWriter
func (g *GetNotification) writeNotificationLog(ctx context.Context, notification device.Notification) error {
	// Parse the notification date
	var timestamp time.Time
	parse, err := time.ParseInLocation("02/01/2006 15:04:05", notification.Date, time.Local)
	if err != nil {
		g.log.Infof("Error parsing date [%s]: %s", notification.Date, err.Error())
		timestamp = time.Now()
	} else {
		timestamp = parse
	}

	// Create metadata with device information and notification details
	metadata := map[string]interface{}{
		"application_name": "ups-metrics",
		"id":               notification.ID,
		"message":          notification.Message,
		"date":             notification.Date,
		"device_address":   g.config.GetDeviceAddress(),
	}

	// Create the full log message
	fullMessage := fmt.Sprintf("Notification %d on %s with %s", notification.ID, notification.Date, notification.Message)

	// Create LogEntry
	entry := port.LogEntry{
		Timestamp: timestamp,
		Level:     "info",
		Message:   fullMessage,
		Source:    "ups-metrics",
		Metadata:  metadata,
	}

	// Write log using the LogWriter interface with context support
	return g.logWriter.WriteLog(ctx, entry)
}
