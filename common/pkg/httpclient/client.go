package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a thin wrapper around http.Client that standardises base URL
// handling and request helpers across microservices.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option modifies the behaviour of Client when constructed.
type Option func(*Client)

// WithTimeout overrides the default request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient injects a custom http.Client implementation.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithTransport overrides the default HTTP transport for connection pooling tuning.
func WithTransport(t http.RoundTripper) Option {
	return func(c *Client) {
		c.httpClient.Transport = t
	}
}

// defaultTransport returns an http.Transport with connection pooling tuned
// for service-to-service communication (higher MaxIdleConnsPerHost than
// Go's default of 2).
func defaultTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}
}

// New constructs a Client with sane defaults.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: defaultTransport(),
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Do issues an HTTP request using the configured base URL and client.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader, headers http.Header) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("httpclient: nil client")
	}
	url := c.buildURL(path)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	for key, values := range headers {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}
	return c.httpClient.Do(req)
}

// DoBytes is a convenience helper that accepts a byte slice body.
func (c *Client) DoBytes(ctx context.Context, method, path string, body []byte, headers http.Header) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	return c.Do(ctx, method, path, reader, headers)
}

// DoJSON marshals payload to JSON and sets the appropriate header.
func (c *Client) DoJSON(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*http.Response, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if headers == nil {
		headers = make(http.Header)
	}
	if headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", "application/json")
	}
	return c.DoBytes(ctx, method, path, data, headers)
}

func (c *Client) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if c.baseURL == "" {
		return path
	}
	return c.baseURL + path
}
