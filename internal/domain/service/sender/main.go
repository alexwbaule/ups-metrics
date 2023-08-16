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
	"time"
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
		client: client.New(l.Config, fmt.Sprintf("http://%s:%s", iflx.Address, iflx.Port)),
	}
}

func (w *Worker) Run(ctx context.Context, jobs <-chan device.Metric) error {
	for {
		select {
		case <-ctx.Done():
			w.log.Infof("Stopping worker job...")
			return context.Canceled
		case item := <-jobs:
			w.log.Infof("Sending metrics to %s on database %s", w.influx.Address, w.influx.Database)
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
			body.WriteString(fmt.Sprintf("%s,host=%s value=%s %d\n", influx.UPSMetricName(gauge.Name), metric.DeployName, gauge.Phases.Value, time.Now().UnixNano()))
		}
	}

	for _, state := range metric.States {
		if influx.UPSMetricState(state.Name) != "" {
			body.WriteString(fmt.Sprintf("ups_status,state=%s,host=%s value=%v %d\n", influx.UPSMetricState(state.Name), metric.DeployName, state.Value, time.Now().UnixNano()))
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
	return nil
}
