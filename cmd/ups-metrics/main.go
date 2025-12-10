package main

import (
	"context"
	"net/http"

	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/domain/service/metric"
	"github.com/alexwbaule/ups-metrics/internal/domain/service/notification"
	"github.com/alexwbaule/ups-metrics/internal/resource/smsups"
	"github.com/alexwbaule/ups-metrics/internal/resource/writer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

func main() {
	app := application.NewApplication()

	app.Run(func(ctx context.Context) error {
		app.Log.Infof("Device Interval: %+v", app.Config.GetInterval())

		app.Log.SetLevel(app.Config.GetLogLevel())

		sms := smsups.MewSMSUPS(app)
		err := sms.Login(ctx, 1)
		if err != nil {
			return err
		}

		// Create log writer using factory
		logWriter, err := writer.CreateLogWriter(app)
		if err != nil {
			app.Log.Errorf("failed to create log writer: %s", err)
			return err
		}

		// Ensure graceful shutdown of log writer when context is done
		defer func() {
			if logWriter != nil {
				if closeErr := logWriter.Close(); closeErr != nil {
					app.Log.Errorf("error closing log writer during shutdown: %s", closeErr)
				}
			}
		}()

		metrics := metric.NewMetric(app, sms)
		notif := notification.NewGetNotification(app.Log, app.Config, sms, logWriter)

		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			return metrics.Run(ctx)
		})

		g.Go(func() error {
			return notif.Run(ctx)
		})

		g.Go(func() error {
			<-ctx.Done()
			return config.SaveLastIdConfig(notif.LastId())
		})

		g.Go(func() error {
			http.Handle("/metrics", promhttp.Handler())
			return http.ListenAndServe(":"+app.Config.GetMetricConfig().Prometheus.Port, nil)
		})
		return g.Wait()
	})
}
