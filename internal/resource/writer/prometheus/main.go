package prometheus

import (
	"context"
	"strconv"

	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer"
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
		if s, err := strconv.ParseFloat(gauge.Phases.Value, 64); err == nil {
			UPSMetricName.WithLabelValues(metric.DeployName, UPSMetricStatusLabel(gauge.Name), gauge.Unit).Set(s)
			w.log.Infof("adding phase %s (%s) == %f to gauge", gauge.Name, gauge.Phases.Value, s)
		} else {
			if gauge.Name == "Tipo" {
				var value float64
				if gauge.Phases.Value == "UPS Line Interative" {
					value = 1
				}
				UPSMetricName.WithLabelValues(metric.DeployName, UPSMetricStatusLabel(gauge.Name), gauge.Unit).Set(value)
				w.log.Infof("adding phase %s (%s) == %f to gauge", gauge.Name, gauge.Phases.Value, value)
			}
		}
	}

	for _, state := range metric.States {
		name, value := UPSMetricStateLabel(state.Name, state.Value)
		UPSMetricState.WithLabelValues(metric.DeployName, name).Set(value)
		w.log.Infof("adding state %s -> %s -> %f to gauge", state.Name, name, value)
	}
	return nil
}
