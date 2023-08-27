package metric

import (
	"context"
	"errors"
	"fmt"
	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/resource/http/client"
	"net/url"
	"time"
)

type GetMetric struct {
	log      *logger.Logger
	intv     time.Duration
	jobs     chan<- device.Metric
	client   *client.Client
	loginusr device.Login
	auth     device.Authentication
	maxTry   int
}

func NewMetric(l *application.Application, j chan<- device.Metric) *GetMetric {
	return &GetMetric{
		log:      l.Log,
		intv:     l.Config.GetInterval(),
		jobs:     j,
		client:   client.New(l.Config, fmt.Sprintf("https://%s", l.Config.GetDeviceAddress()), l.Log),
		loginusr: l.Config.GetLogin(),
		maxTry:   l.Config.GetHttpClient().RetryCount,
	}
}

func (g *GetMetric) Run(ctx context.Context) error {
	ticker := time.NewTicker(g.intv)
	defer ticker.Stop()

	err := g.login(ctx, 1)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			g.log.Infof("Stopping GpuStat job...")
			return context.Canceled
		}
		g.log.Infof("Getting metrics from %s with user %s", g.auth.DeployName, g.auth.Usuario)

		err := g.getStats(ctx)
		if err != nil {
			g.log.Errorf("run error: %s", err)
			return err
		}
	}
}

func (g *GetMetric) getStats(ctx context.Context) error {
	var metrics device.Metric
	request := client.Request{
		Url:            "/sms/mobile/medidores/",
		PathParameters: nil,
		Headers: map[string]string{
			"token":    g.auth.Token,
			"deployid": g.auth.DeployID,
		},
		QueryParameters: nil,
	}
	get, err := g.client.Get(ctx, request, &metrics)
	if err != nil {
		return err
	}
	if get.IsError() {
		return fmt.Errorf("error: %s", get.String())
	}
	if metrics.ResponseStatus != "S001" {
		g.log.Errorf("token error: [%s]", client.ErrorCodes(metrics.ResponseStatus))
		g.log.Errorf("metrics error: [%#v]", metrics)

		err := g.login(ctx, 1)
		if err != nil {
			return err
		}
	}
	metrics.GetAt = time.Now()
	g.jobs <- metrics
	return nil
}

func (g *GetMetric) login(ctx context.Context, retryCount int) error {
	request := client.Request{
		Url:            "/sms/mobile/login/",
		PathParameters: nil,
		Headers:        nil,
		QueryParameters: map[string]string{
			"username": g.loginusr.Username,
			"password": g.loginusr.Password,
			"iddevice": "22",
			"sodevice": "android",
		},
	}
	get, err := g.client.Post(ctx, request, nil, &g.auth)
	if err != nil {
		return g.backoff(ctx, retryCount, err)
	}
	if get.IsError() {
		return g.backoff(ctx, retryCount, get.Error().(error))
	}
	if g.auth.ResponseStatus != "S001" {
		return fmt.Errorf("login error: %s", client.ErrorCodes(g.auth.ResponseStatus))
	}
	return nil
}

func (g *GetMetric) backoff(ctx context.Context, retryCount int, err error) error {
	var urlError *url.Error

	if retryCount == g.maxTry {
		return err
	}
	if errors.As(err, &urlError) {
		if urlError.Timeout() {
			g.log.Infof("Trying again (%d)(%d)...", retryCount, g.maxTry)
			return g.login(ctx, retryCount+1)
		}
	}
	return err
}
