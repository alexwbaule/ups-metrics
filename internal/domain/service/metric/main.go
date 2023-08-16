package metric

import (
	"context"
	"fmt"
	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"github.com/alexwbaule/ups-metrics/internal/resource/http/client"
	"time"
)

type GetMetric struct {
	log      *logger.Logger
	intv     time.Duration
	jobs     chan<- device.Metric
	client   *client.Client
	loginusr device.Login
	auth     device.Authentication
}

func NewMetric(l *application.Application, j chan<- device.Metric) *GetMetric {
	return &GetMetric{
		log:      l.Log,
		intv:     l.Config.GetInterval(),
		jobs:     j,
		client:   client.New(l.Config, fmt.Sprintf("https://%s", l.Config.GetDeviceAddress())),
		loginusr: l.Config.GetLogin(),
	}
}

func (g *GetMetric) Run(ctx context.Context) error {
	ticker := time.NewTicker(g.intv)
	defer ticker.Stop()

	err := g.login(ctx)
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
		g.log.Errorf("token error: %s", client.ErrorCodes(metrics.ResponseStatus))
		err := g.login(ctx)
		if err != nil {
			return err
		}
	}
	g.jobs <- metrics
	return nil
}

func (g *GetMetric) login(ctx context.Context) error {
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
		return err
	}
	if get.IsError() {
		return fmt.Errorf("error: %s", get.String())
	}
	if g.auth.ResponseStatus != "S001" {
		return fmt.Errorf("login error: %s", client.ErrorCodes(g.auth.ResponseStatus))
	}
	return nil
}
