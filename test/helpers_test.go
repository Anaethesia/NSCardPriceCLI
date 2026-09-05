//go:build integration

package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"nscardprice/internal/httpx"
	"nscardprice/internal/merchant"
)

// requireIntegration skips unless NSCARD_INTEGRATION=1 is set, so `go test
// ./...` stays free of real network calls.
func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("NSCARD_INTEGRATION") == "" {
		t.Skip("set NSCARD_INTEGRATION=1 to run real API integration tests")
	}
}

// pause keeps the gap between real requests large enough to avoid IP bans.
func pause() {
	time.Sleep(3 * time.Second)
}

// realClient builds an httpx client with generous timeouts for real networks.
func realClient() *httpx.Client {
	return httpx.New(httpx.Options{
		Timeout: 15 * time.Second,
		Retries: 2,
		Backoff: 3 * time.Second,
	})
}

// merchantByName returns the registered merchant implementation for key.
func merchantByName(t *testing.T, key string) merchant.Merchant {
	t.Helper()
	reg := merchant.NewRegistry(realClient())
	m, ok := reg[key]
	if !ok {
		t.Fatalf("merchant %q not registered", key)
	}
	return m
}

// expectFetchStatus accepts the statuses a real fetch may legitimately report.
func expectFetchStatus(t *testing.T, id string, res merchant.Result) {
	t.Helper()
	switch res.Status {
	case merchant.StatusOK, merchant.StatusUnavailable, merchant.StatusSkipped:
		// ok: a merchant may not currently buy this game.
	default:
		t.Errorf("unexpected status %q for game_id=%s (want ok/unavailable/skipped)", res.Status, id)
	}
}

// resultSummary renders a Result with dereferenced pointers so test logs show
// actual prices instead of pointer addresses.
func resultSummary(r merchant.Result) string {
	recycle, sell := "nil", "nil"
	if r.RecyclePrice != nil {
		recycle = trimFloat(*r.RecyclePrice)
	}
	if r.SellPrice != nil {
		sell = trimFloat(*r.SellPrice)
	}
	return "merchant=" + r.Merchant +
		" status=" + r.Status +
		" recycle_price=" + recycle +
		" sell_price=" + sell +
		" item_id=" + r.ItemID +
		" note=" + r.Note +
		" error=" + r.Error
}

// trimFloat renders a float without trailing zeros.
func trimFloat(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", v), "0"), ".")
}
