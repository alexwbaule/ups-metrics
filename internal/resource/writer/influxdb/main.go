package influxdb

import (
	"context"
	"fmt"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/resource/http/client"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer"
	"strings"
)

type Influx struct {
	log    *logger.Logger
	client *client.Client
	influx device.Influx
}

func NewWorker(l *logger.Logger, config *config.Config) writer.WriteMetric {
	return &Influx{
		log:    l,
		influx: config.GetMetricConfig().Influx,
		client: client.New(config, fmt.Sprintf("http://%s:%s", config.GetMetricConfig().Influx.Address, config.GetMetricConfig().Influx.Port), l),
	}
}

func (w *Influx) Write(ctx context.Context, metric device.Metric) error {
	var body strings.Builder
	var response interface{}

	request := client.Request{
		Url:            "/write",
		PathParameters: nil,
		Headers:        nil,
		QueryParameters: map[string]string{
			"db": w.influx.Database,
		},
	}

	for _, gauge := range metric.Gauges {
		if UPSMetricName(gauge.Name) != "" {
			body.WriteString(fmt.Sprintf("%s,host=%s value=%s %d\n",
				UPSMetricName(gauge.Name), metric.DeployName, gauge.Phases.Value, metric.GetAt.UnixNano()))
		}
	}

	for _, state := range metric.States {
		if UPSMetricState(state.Name) != "" {
			body.WriteString(fmt.Sprintf("%s,host=%s value=\"%s\" %d\n",
				UPSMetricState(state.Name), metric.DeployName, UPSMetricStateValue(state.Name, state.Value), metric.GetAt.UnixNano()))
		}
	}

	get, err := w.client.Post(ctx, request, body.String(), &response)
	if err != nil {
		return err
	}
	if get.IsError() {
		return fmt.Errorf("error: %s", get.String())
	}
	if get.StatusCode() != 204 {
		return fmt.Errorf("error: %s", get.String())
	}
	w.log.Infof("InfluxDB Write response: %d", get.StatusCode())
	return nil
}
