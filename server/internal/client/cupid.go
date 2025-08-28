package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// / Client is an HTTP client for Cupid
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	userAgent  string
	forceClose bool
}

// Option configures the Client.
type Option func(*Client)

// WithHTTPClient allows providing a custom *http.Client (e.g., custom Transport).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithTimeout configures timeout on the underlying *http.Client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient != nil {
			c.httpClient.Timeout = timeout
		}
	}
}

// WithConnectionClose forces requests to close connections and disables keep-alives when possible.
func WithConnectionClose() Option {
	return func(c *Client) {
		c.forceClose = true
		if tr, ok := c.httpClient.Transport.(*http.Transport); ok {
			tr.DisableKeepAlives = true
		}
	}
}

// New constructs a new Client.
func New(baseURL string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("baseURL is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid baseURL: %w", err)
	}
	c := &Client{
		baseURL:    parsed,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		userAgent:  "cupid-client/1.0",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Error represents 4xx or 5xx HTTP responses.
type Error struct {
	StatusCode int
	RequestID  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: status=%d request_id=%s", e.StatusCode, e.RequestID)
}

// Do issues an HTTP request and returns the response body for 2xx codes.
// For 4xx/5xx, it returns a typed error containing status and request id.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader, headers http.Header) ([]byte, *http.Response, error) {
	// Compose URL with minimal assumptions. Caller is responsible for correct path.
	fullURL := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		// naive join: base + path
		base := strings.TrimRight(c.baseURL.String(), "/")
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		fullURL = base + path
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, nil, err
	}
	if c.forceClose {
		req.Close = true
	}

	// Apply per-call headers only; caller owns headers.
	for k, vv := range headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	requestID := resp.Header.Get("X-Request-Id")

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, resp, nil
	}
	if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
		return nil, resp, &Error{StatusCode: resp.StatusCode, RequestID: requestID}
	}
	if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		return nil, resp, &Error{StatusCode: resp.StatusCode, RequestID: requestID}
	}
	return nil, resp, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}
