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
