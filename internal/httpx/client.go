// Package httpx wraps http.Client with retries, backoff and optional TLS
// verification, decoupled from any specific merchant.
package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Options configure a Client.
type Options struct {
	Timeout  time.Duration
	Retries  int
	Backoff  time.Duration
	Insecure bool
	// RateGap is the minimum gap between two consecutive HTTP requests,
	// enforced across all calls on the client (concurrent queries to several
	// merchants/products, catalog paging, ...). <=0 disables throttling.
	RateGap time.Duration
}

// Client performs JSON/form requests with retries.
type Client struct {
	opts Options
	hc   *http.Client

	mu   sync.Mutex
	last time.Time // time of the most recent request, for RateGap pacing
}

// New builds a Client from the given options, applying sensible defaults for
// zero values.
func New(o Options) *Client {
	if o.Timeout <= 0 {
		o.Timeout = 10 * time.Second
	}
	if o.Retries < 0 {
		o.Retries = 0
	}
	if o.Backoff <= 0 {
		o.Backoff = time.Second
	}
	transport := &http.Transport{
		// Honor HTTP(S)_PROXY from the environment (e.g. sandbox egress proxy);
		// without this the client dials directly and can hang/time out where
		// direct external access is blocked.
		Proxy: http.ProxyFromEnvironment,
	}
	if o.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user opted in via --insecure
	}
	return &Client{
		opts: o,
		hc:   &http.Client{Timeout: o.Timeout, Transport: transport},
	}
}

// PostJSON sends a JSON-encoded body and returns the raw response body.
func (c *Client) PostJSON(ctx context.Context, u string, headers map[string]string, body any) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	build := func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(buf.Bytes()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req, nil
	}
	return c.do(ctx, build)
}

// PostForm sends an application/x-www-form-urlencoded body.
func (c *Client) PostForm(ctx context.Context, u string, headers map[string]string, form url.Values) ([]byte, error) {
	body := form.Encode()
	build := func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req, nil
	}
	return c.do(ctx, build)
}

// Get performs a GET request with optional query params.
func (c *Client) Get(ctx context.Context, u string, headers map[string]string, params url.Values) ([]byte, error) {
	build := func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		if len(params) > 0 {
			req.URL.RawQuery = params.Encode()
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req, nil
	}
	return c.do(ctx, build)
}

// do rebuilds the request on every attempt and returns the body for the first
// 2xx response. Non-2xx responses are retried up to the configured limit.
func (c *Client) do(ctx context.Context, build func() (*http.Request, error)) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.opts.Retries; attempt++ {
		req, err := build()
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)

		c.throttle(ctx)

		resp, err := c.hc.Do(req)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				// A 2xx response with an empty body is usually throttling or a
				// gateway hiccup; treat it as retryable so paginated catalog
				// pulls (e.g. 火枪手) survive transient empty pages.
				if len(body) == 0 {
					lastErr = fmt.Errorf("empty response body")
				} else {
					return body, nil
				}
			} else if readErr != nil {
				lastErr = readErr
			} else {
				lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(body))
			}
		} else {
			lastErr = err
		}

		if attempt < c.opts.Retries {
			d := c.opts.Backoff * time.Duration(attempt+1)
			select {
			case <-time.After(d):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, lastErr
}

// throttle enforces the configured RateGap: every request waits until at least
// RateGap has passed since the previous request issued from this client. The
// wait happens while holding the lock so concurrent callers line up in order
// and the actual request stream is strictly spaced (protects against merchant
// anti-scraping limits under bursty multi-store/multi-product queries).
func (c *Client) throttle(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.opts.RateGap <= 0 {
		return
	}
	if wait := time.Until(c.last.Add(c.opts.RateGap)); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
		}
	}
	c.last = time.Now()
}

func truncate(b []byte) string {
	const max = 200
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}
