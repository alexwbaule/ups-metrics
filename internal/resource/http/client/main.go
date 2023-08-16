package client

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/go-resty/resty/v2"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	Url             string
	PathParameters  map[string]string
	QueryParameters map[string]string
	Headers         map[string]string
}

type Client struct {
	*resty.Client
}

type Response struct {
	*resty.Response
}

func New(cfg *config.Config, baseUrl string) *Client {
	client := resty.New()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = cfg.GetHttpClient().MaxIdleConns
	transport.MaxConnsPerHost = cfg.GetHttpClient().MaxConnsPerHost
	transport.MaxIdleConnsPerHost = cfg.GetHttpClient().MaxIdleConnsPerHost
	transport.ResponseHeaderTimeout = cfg.GetHttpClient().ResponseHeaderTimeout
	transport.TLSHandshakeTimeout = cfg.GetHttpClient().TLSHandshakeTimeout
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	transport.ExpectContinueTimeout = cfg.GetHttpClient().ExpectContinueTimeout
	transport.DialContext = (&net.Dialer{
		Timeout:   cfg.GetHttpClient().DialTimeout,
		KeepAlive: cfg.GetHttpClient().DialKeepAlive,
	}).DialContext

	client.
		EnableTrace().
		SetBaseURL(baseUrl).
		SetTransport(transport).
		SetRetryAfter(
			func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
				return 0, errors.New("quota exceeded")
			}).
		SetRetryCount(cfg.GetHttpClient().RetryCount).
		SetRetryWaitTime(cfg.GetHttpClient().RetryWaitCount).
		SetRetryMaxWaitTime(cfg.GetHttpClient().RetryMaxWaitTime)

	return &Client{client}
}

func (c *Client) Get(ctx context.Context, request Request, result any) (*Response, error) {

	Url, err := url.Parse(request.Url)

	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, "path", Url.Path)

	r := c.R().
		SetPathParams(request.PathParameters).
		SetHeaders(request.Headers).
		SetQueryParams(request.QueryParameters).
		SetContext(ctx).
		SetResult(result)

	res, err := r.Get(request.Url)

	return &Response{res}, err
}

func (c *Client) Post(ctx context.Context, request Request, body any, result any) (*Response, error) {

	Url, err := url.Parse(request.Url)

	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, "path", Url.Path)

	r := c.R().
		SetPathParams(request.PathParameters).
		SetHeaders(request.Headers).
		SetQueryParams(request.QueryParameters).
		SetContext(ctx).
		SetResult(result).
		SetBody(body)

	res, err := r.Post(request.Url)

	return &Response{res}, err
}
