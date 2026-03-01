package smsups

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

type SMSUps struct {
	log      *logger.Logger
	intv     time.Duration
	client   *client.Client
	loginusr device.Login
	auth     *device.Authentication
	maxTry   int
}

func MewSMSUPS(l *application.Application) *SMSUps {
	return &SMSUps{
		log:      l.Log,
		intv:     l.Config.GetInterval(),
		client:   client.New(l.Config, fmt.Sprintf("https://%s", l.Config.GetDeviceAddress()), l.Log),
		loginusr: l.Config.GetLogin(),
		maxTry:   l.Config.GetHttpClient().RetryCount,
	}
}

func (g *SMSUps) GetMeasurements(ctx context.Context) (device.Metric, error) {
	// Adiciona timeout de 30s para a requisição completa
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return g.medidores(reqCtx, 1)
}

func (g *SMSUps) GetNotifications(ctx context.Context) (device.Notifications, error) {
	// Adiciona timeout de 30s para a requisição completa
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return g.notifications(reqCtx, 1)
}

func (g *SMSUps) Login(ctx context.Context, retryCount int) error {
	g.log.Infof("Calling Login....[%d]", retryCount)
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
	g.print(get)
	if err != nil {
		return g.backoffLogin(ctx, retryCount, err)
	}
	if get.IsError() {
		return g.backoffLogin(ctx, retryCount, get.Error().(error))
	}
	if g.auth.ResponseStatus != "S001" {
		return fmt.Errorf("login error: %s", client.ErrorCodes(g.auth.ResponseStatus))
	}
	return nil
}

func (g *SMSUps) notifications(ctx context.Context, retryCount int) (device.Notifications, error) {
	var notifications device.Notifications

	request := client.Request{
		Url:            "/sms/mobile/beannotificacao/",
		PathParameters: nil,
		Headers: map[string]string{
			"token":    g.auth.Token,
			"deployid": g.auth.DeployID,
		},
		QueryParameters: map[string]string{
			"qtd": "1000",
		},
	}
	get, err := g.client.Get(ctx, request, &notifications)
	g.print(get)
	if err != nil {
		return g.backoffNotification(ctx, retryCount, err)
	}
	if get.IsError() {
		return g.backoffNotification(ctx, retryCount, get.Error().(error))
	}
	if notifications.ResponseStatus != "" {
		g.log.Errorf("token error: [%s]", client.ErrorCodes(notifications.ResponseStatus))
		g.log.Errorf("metrics error: [%#v]", notifications)

		err := g.Login(ctx, 1)
		if err != nil {
			return g.notifications(ctx, retryCount)
		}
	}
	return notifications, nil
}

func (g *SMSUps) medidores(ctx context.Context, retryCount int) (device.Metric, error) {
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
	g.print(get)

	if err != nil {
		return g.backoffMetric(ctx, retryCount, err)
	}
	if get.IsError() {
		return g.backoffMetric(ctx, retryCount, get.Error().(error))
	}
	if metrics.ResponseStatus != "S001" {
		g.log.Errorf("token error: [%s]", client.ErrorCodes(metrics.ResponseStatus))
		g.log.Errorf("metrics error: [%#v]", metrics)

		err := g.Login(ctx, 1)
		if err != nil {
			return g.medidores(ctx, retryCount)
		}
	}
	metrics.GetAt = time.Now()
	return metrics, nil
}

func (g *SMSUps) backoffNotification(ctx context.Context, retryCount int, err error) (device.Notifications, error) {
	var urlError *url.Error

	// Apenas 1 retry rápido - o ticker cuidará do resto
	maxQuickRetries := 2
	if retryCount >= maxQuickRetries {
		return device.Notifications{}, fmt.Errorf("notification request failed after %d attempts: %w", retryCount, err)
	}
	
	if errors.As(err, &urlError) {
		if urlError.Timeout() {
			g.log.Warnf("Notification timeout, quick retry %d/%d", retryCount, maxQuickRetries)
			time.Sleep(2 * time.Second) // Espera 2s antes do retry
			return g.notifications(ctx, retryCount+1)
		}
	}
	return device.Notifications{}, fmt.Errorf("notification request failed: %w", err)
}

func (g *SMSUps) backoffMetric(ctx context.Context, retryCount int, err error) (device.Metric, error) {
	var urlError *url.Error

	// Apenas 1 retry rápido - o ticker cuidará do resto
	maxQuickRetries := 2
	if retryCount >= maxQuickRetries {
		return device.Metric{}, fmt.Errorf("metric request failed after %d attempts: %w", retryCount, err)
	}
	
	if errors.As(err, &urlError) {
		if urlError.Timeout() {
			g.log.Warnf("Metric timeout, quick retry %d/%d", retryCount, maxQuickRetries)
			time.Sleep(2 * time.Second) // Espera 2s antes do retry
			return g.medidores(ctx, retryCount+1)
		}
	}
	return device.Metric{}, fmt.Errorf("metric request failed: %w", err)
}

func (g *SMSUps) backoffLogin(ctx context.Context, retryCount int, err error) error {
	var urlError *url.Error

	// Apenas 1 retry rápido - o ticker cuidará do resto
	maxQuickRetries := 2
	if retryCount >= maxQuickRetries {
		return fmt.Errorf("login failed after %d attempts: %w", retryCount, err)
	}
	
	if errors.As(err, &urlError) {
		if urlError.Timeout() {
			g.log.Warnf("Login timeout, quick retry %d/%d", retryCount, maxQuickRetries)
			time.Sleep(2 * time.Second) // Espera 2s antes do retry
			return g.Login(ctx, retryCount+1)
		}
	}
	return fmt.Errorf("login failed: %w", err)
}

func (g *SMSUps) print(get *client.Response) {
	debug := fmt.Sprintf("curl -X %s \"%s\" ", get.Request.Method, get.Request.URL)
	for s, header := range get.Request.Header {
		if s == "User-Agent" {
			continue
		}
		debug += fmt.Sprintf("--header \"%s: %s\" ", s, header[0])
	}
	g.log.Debug(debug)
}
