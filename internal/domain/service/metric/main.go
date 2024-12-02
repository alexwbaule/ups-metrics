package metric

import (
	"context"
	"fmt"
	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/resource/smsups"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer/influxdb"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer/prometheus"
	"time"
)

type GetMetric struct {
	log *logger.Logger
	*config.Config
	sms *smsups.SMSUps
}

func NewMetric(l *application.Application, s *smsups.SMSUps) *GetMetric {
	return &GetMetric{
		log:    l.Log,
		Config: l.Config,
		sms:    s,
	}
}

func (g *GetMetric) Run(ctx context.Context) error {
	ticker := time.NewTicker(g.Config.GetInterval())
	var metricWriter writer.WriteMetric

	defer ticker.Stop()

	if g.Config.GetMetricConfig().Prometheus.Enabled {
		g.log.Infof("Starting Prometheus metrics collection")
		metricWriter = prometheus.NewWorker(g.log, g.Config)
	} else if g.Config.GetMetricConfig().Influx.Enabled {
		g.log.Infof("Starting InfluxDB metrics collection")
		metricWriter = influxdb.NewWorker(g.log, g.Config)
	} else {
		e := fmt.Errorf("no metric configuration found")
		return e
	}

	for {
		select {
		case <-ctx.Done():
			g.log.Infof("stopping get metric job...")
			return context.Canceled
		case <-ticker.C:
		}
		metric, err := g.getStats(ctx)
		if err != nil {
			g.log.Errorf("get metric error: %s", err)
			return err
		}
		err = metricWriter.Write(ctx, metric)
		if err != nil {
			g.log.Errorf("writing metric error: %s", err)
			return err
		}
	}
}

func (g *GetMetric) getStats(ctx context.Context) (device.Metric, error) {
	return g.sms.GetMeasurements(ctx)
}
