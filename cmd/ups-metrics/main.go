package main

import (
	"context"
	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/domain/service/metric"
	"github.com/alexwbaule/ups-metrics/internal/domain/service/sender"
	"golang.org/x/sync/errgroup"
)

func main() {
	app := application.NewApplication()

	jobs := make(chan device.Metric)

	app.Run(func(ctx context.Context) error {
		app.Log.Infof("Device Interval: %+v", app.Config.GetInterval())

		worker := sender.NewWorker(app)
		metrics := metric.NewMetric(app, jobs)

		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			return worker.Run(ctx, jobs)
		})

		g.Go(func() error {
			return metrics.Run(ctx)
		})

		g.Go(func() error {
			<-ctx.Done()
			close(jobs)
			return nil
		})
		return g.Wait()
	})
}
