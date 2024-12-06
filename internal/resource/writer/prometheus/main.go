package prometheus

import (
	"context"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer"
	"strconv"
)

type Prometheus struct {
	log        *logger.Logger
	prometheus device.Prometheus
}

func NewWorker(l *logger.Logger, config *config.Config) writer.WriteMetric {
	return &Prometheus{
		log:        l,
		prometheus: config.GetMetricConfig().Prometheus,
	}
}

func (w *Prometheus) Write(ctx context.Context, metric device.Metric) error {
	w.log.Infof("sending metric to prometheus")

	for _, gauge := range metric.Gauges {
		if UPSMetricName(gauge.Name) != nil {
			if s, err := strconv.ParseFloat(gauge.Phases.Value, 64); err == nil {
				UPSMetricName(gauge.Name).WithLabelValues(metric.DeployName, gauge.Type, gauge.Unit).Set(s)
				w.log.Infof("adding phase %s (%s) == %f to gauge", gauge.Name, gauge.Phases.Value, s)
			}
		}
	}

	for _, state := range metric.States {
		if UPSMetricState(state.Name) != nil {
			UPSMetricState(state.Name).WithLabelValues(metric.DeployName, UPSMetricStateLabel(state.Name, state.Value)).Set(UPSMetricStateValue(state.Name, state.Value))
			w.log.Infof("adding state %s -> %s -> %f to histogram", state.Name, UPSMetricStateLabel(state.Name, state.Value), UPSMetricStateValue(state.Name, state.Value))
		}
	}
	return nil
}
