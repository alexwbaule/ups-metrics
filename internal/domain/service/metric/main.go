package metric

import (
	"context"
	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/resource/smsups"
	"time"
)

type GetMetric struct {
	log *logger.Logger
	*config.Config
	sms  *smsups.SMSUps
	jobs chan<- device.Metric
}

func NewMetric(l *application.Application, s *smsups.SMSUps, j chan<- device.Metric) *GetMetric {
	return &GetMetric{
		log:    l.Log,
		jobs:   j,
		Config: l.Config,
		sms:    s,
	}
}

func (g *GetMetric) Run(ctx context.Context) error {
	ticker := time.NewTicker(g.Config.GetInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			g.log.Infof("stopping get metric job...")
			return context.Canceled
		}
		err := g.getStats(ctx)
		if err != nil {
			g.log.Errorf("get metric error: %s", err)
			return err
		}
	}
}

func (g *GetMetric) getStats(ctx context.Context) error {
	metrics, err := g.sms.GetMeasurements(ctx)
	if err != nil {
		return err
	}
	g.jobs <- metrics
	return nil
}
