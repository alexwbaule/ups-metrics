package notification

import (
	"context"
	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/resource/graylog"
	"github.com/alexwbaule/ups-metrics/internal/resource/smsups"
	"time"
)

type GetNotification struct {
	log *logger.Logger
	*config.Config
	sms     *smsups.SMSUps
	graylog *graylog.Gelf
	last    int
}

func NewGetNotification(l *application.Application, s *smsups.SMSUps) *GetNotification {
	return &GetNotification{
		log:     l.Log,
		Config:  l.Config,
		sms:     s,
		last:    l.Config.GetLastKnowId(),
		graylog: graylog.NewGelf(l),
	}
}

func (g *GetNotification) Run(ctx context.Context) error {
	ticker := time.NewTicker(g.Config.GetInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			g.log.Infof("stopping get notifications job...")
			return context.Canceled
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
	g.log.Infof("sending notifications bigger than %d to %s for hostname %s", g.last, g.graylog.Address, g.graylog.Hostname)

	s := len(n.Notifications) - 1

	for i := s; i >= 0; i-- {
		notification := n.Notifications[i]
		if notification.ID > g.last {
			g.log.Infof("sending notifications id: %d", notification.ID)
			g.graylog.LogNotifications(notification)
			g.last = notification.ID
		}
	}
	return nil
}

func (g *GetNotification) LastId() int {
	return g.last
}
