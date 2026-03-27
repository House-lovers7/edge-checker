package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/House-lovers7/edge-checker/internal/profile"
)

// Response records the result of a single HTTP request.
type Response struct {
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"duration_ms"`
	BodySize   int64         `json:"body_size"`
	Timestamp  time.Time     `json:"timestamp"`
	Error      string        `json:"error,omitempty"`
}

// Client wraps net/http.Client with profile headers and extra headers.
type Client struct {
	httpClient   *http.Client
	profile      *profile.Profile
	extraHeaders map[string]string
	host         string // Host header override (for CDN routing)
}

// NewClient creates a configured HTTP client.
// If insecure is true, TLS certificate verification is skipped.
func NewClient(timeout time.Duration, prof *profile.Profile, extraHeaders map[string]string, host string, insecure bool) *Client {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested via --insecure flag
	}

	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		// Do not follow redirects — observe WAF responses directly
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Client{
		httpClient:   httpClient,
		profile:      prof,
		extraHeaders: extraHeaders,
		host:         host,
	}
}

// Do sends an HTTP request and returns the observed response.
func (c *Client) Do(ctx context.Context, method, url string) *Response {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return &Response{
			Timestamp: start,
			Duration:  time.Since(start),
			Error:     fmt.Sprintf("request creation failed: %v", err),
		}
	}

	// Apply profile headers
	if c.profile != nil {
		for k, v := range c.profile.Headers {
			req.Header.Set(k, v)
		}
	}

	// Apply extra headers (override profile)
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}

	// Override Host header for CDN testing
	if c.host != "" {
		req.Host = c.host
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &Response{
			Timestamp: start,
			Duration:  time.Since(start),
			Error:     fmt.Sprintf("request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Read and discard body to get size and ensure connection reuse
	bodySize, _ := io.Copy(io.Discard, resp.Body)

	return &Response{
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
		BodySize:   bodySize,
		Timestamp:  start,
	}
}

// DoWithHeaders sends a request with additional headers merged on top of profile + extraHeaders.
func (c *Client) DoWithHeaders(ctx context.Context, method, url string, additionalHeaders map[string]string) *Response {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return &Response{
			Timestamp: start,
			Duration:  time.Since(start),
			Error:     fmt.Sprintf("request creation failed: %v", err),
		}
	}

	if c.profile != nil {
		for k, v := range c.profile.Headers {
			req.Header.Set(k, v)
		}
	}
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range additionalHeaders {
		req.Header.Set(k, v)
	}
	if c.host != "" {
		req.Host = c.host
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &Response{
			Timestamp: start,
			Duration:  time.Since(start),
			Error:     fmt.Sprintf("request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	bodySize, _ := io.Copy(io.Discard, resp.Body)
	return &Response{
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
		BodySize:   bodySize,
		Timestamp:  start,
	}
}

// Close cleans up the client's transport.
func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}
