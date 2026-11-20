package core

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	telemetryerrors "github.com/e2engine/telemetry-go/pkg/errors"
)

const (
	defaultTimeout = time.Second * 2

	maxEventsBytes       = 64 * 1024
	maxRegistrationBytes = 4096
)

type Service struct {
	baseURL string
	http    *http.Client

	initializeOnce sync.Once
	initializeErr  error

	mu    sync.Mutex
	token string
}

func NewService(baseURL string) *Service {
	return newService(baseURL, getHTTPClient(defaultTimeout))
}

func newService(baseURL string, httpClient *http.Client) *Service {
	return &Service{
		baseURL: baseURL,
		http:    httpClient,
	}
}

func getHTTPClient(timeout time.Duration) *http.Client {
	// Fresh HTTP/1 connections avoid transport replays on stale pooled connections
	// and HTTP/2 stream retries. Best-effort delivery makes one network attempt.
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: timeout}).DialContext,
		TLSHandshakeTimeout: timeout,
		DisableKeepAlives:   true,
		Protocols:           protocols,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (s *Service) request(ctx context.Context, path, token string, body []byte) (*http.Response, error) {
	u, err := url.JoinPath(s.baseURL, path)
	if err != nil {
		return nil, errors.Join(telemetryerrors.ErrInvalidURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Join(telemetryerrors.ErrConstructRequest, err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, errors.Join(telemetryerrors.ErrDoRequest, err)
	}

	return resp, nil
}
