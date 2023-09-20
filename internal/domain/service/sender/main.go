package sender

import (
	"context"
	"fmt"
	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/influx"
	"github.com/alexwbaule/ups-metrics/internal/resource/http/client"
	"strings"
)

type Worker struct {
	log    *logger.Logger
	client *client.Client
	influx device.Influx
}

func NewWorker(l *application.Application) *Worker {
	iflx := l.Config.GetMetricConfig()
	return &Worker{
		log:    l.Log,
		influx: iflx,
		client: client.New(l.Config, fmt.Sprintf("http://%s:%s", iflx.Address, iflx.Port), l.Log),
	}
}

func (w *Worker) Run(ctx context.Context, jobs <-chan device.Metric) error {
	for {
		select {
		case <-ctx.Done():
			w.log.Infof("stopping worker job...")
			return context.Canceled
		case item := <-jobs:
			w.log.Infof("sending metrics to %s on database %s", w.influx.Address, w.influx.Database)
			err := w.write(ctx, item)
			if err != nil {
				w.log.Errorf("worker error: %s", err.Error())
				return err
			}
		}
	}
}

func (w *Worker) write(ctx context.Context, metric device.Metric) error {
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
		if influx.UPSMetricName(gauge.Name) != "" {
			body.WriteString(fmt.Sprintf("%s,host=%s value=%s %d\n",
				influx.UPSMetricName(gauge.Name), metric.DeployName, gauge.Phases.Value, metric.GetAt.UnixNano()))
		}
	}

	for _, state := range metric.States {
		if influx.UPSMetricState(state.Name) != "" {
			body.WriteString(fmt.Sprintf("%s,host=%s value=\"%s\" %d\n",
				influx.UPSMetricState(state.Name), metric.DeployName, influx.UPSMetricStateValue(state.Name, state.Value), metric.GetAt.UnixNano()))
		}
	}

	get, err := w.client.Post(ctx, request, body.String(), &response)
	w.print(get)
	if err != nil {
		return err
	}
	if get.IsError() {
		return fmt.Errorf("error: %s", get.String())
	}
	if get.StatusCode() != 204 {
		return fmt.Errorf("error: %s", get.String())
	}
	return nil
}

func (w *Worker) print(get *client.Response) {
	debug := fmt.Sprintf("curl -X %s \"%s\" -d '%+v' ", get.Request.Method, get.Request.URL, get.Request.Body)
	for s, header := range get.Request.Header {
		if s == "User-Agent" {
			continue
		}
		debug += fmt.Sprintf("--header \"%s: %s\" ", s, header[0])
	}
	w.log.Debug(debug)
}
