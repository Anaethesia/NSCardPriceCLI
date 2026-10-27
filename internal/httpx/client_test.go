package httpx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestRateGapPacing checks that parallel requests issued through one client are
// spaced out by at least RateGap, so bursty merchant queries stay below
// anti-scraping limits.
func TestRateGapPacing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "{}")
	}))
	defer server.Close()

	c := New(Options{Timeout: time.Second, RateGap: 40 * time.Millisecond})
	ctx := context.Background()

	const n = 6
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Get(ctx, server.URL, nil, nil); err != nil {
				t.Errorf("Get: %v", err)
			}
		}()
	}
	wg.Wait()

	if want := (n - 1) * c.opts.RateGap; time.Since(start) < want {
		t.Errorf("elapsed %v < min pacing %v, requests sent too bursty", time.Since(start), want)
	}
}

// TestTransportConfig guards the HTTP transport setup: it must keep the
// environment proxy support of http.DefaultTransport, and --insecure must only
// affect this client, not the process-wide default transport.
func TestTransportConfig(t *testing.T) {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	tlsBefore := defaultTransport.TLSClientConfig

	tests := []struct {
		name         string
		insecure     bool
		wantInsecure bool
	}{
		{name: "default keeps environment proxy support", insecure: false, wantInsecure: false},
		{name: "insecure disables TLS verification", insecure: true, wantInsecure: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(Options{Insecure: tt.insecure})
			tr, ok := c.hc.Transport.(*http.Transport)
			if !ok {
				t.Fatalf("transport type = %T, want *http.Transport", c.hc.Transport)
			}
			if tr.Proxy == nil {
				t.Fatal("transport.Proxy is nil; HTTP_PROXY/HTTPS_PROXY would be ignored")
			}
			gotInsecure := tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify
			if gotInsecure != tt.wantInsecure {
				t.Fatalf("InsecureSkipVerify = %v, want %v", gotInsecure, tt.wantInsecure)
			}
			if after := http.DefaultTransport.(*http.Transport).TLSClientConfig; after != tlsBefore {
				t.Fatal("New mutated http.DefaultTransport TLS config")
			}
		})
	}
}
