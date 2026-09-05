//go:build integration

package test

import (
	"context"
	"testing"
	"time"

	"nscardprice/internal/game"
	"nscardprice/internal/merchant"
)

// TestLaolierenReal queries 老猎人 for a few known-good game IDs and verifies
// each returns either a recycle price or a well-defined unavailable status.
func TestLaolierenReal(t *testing.T) {
	requireIntegration(t)
	m := merchantByName(t, "laolieren")

	ids := []game.MerchantEntry{
		{GameID: "154"},  // 塞尔达传说 旷野之息
		{GameID: "3282"}, // 塞尔达传说 王国之泪
		{GameID: "153"},  // 马里奥赛车 8 豪华版
	}

	for _, entry := range ids {
		pause()
		t.Run("laolieren-"+entry.GameID, func(t *testing.T) {
			res, err := m.Fetch(context.Background(), entry)
			t.Logf("game_id=%s %s", entry.GameID, resultSummary(res))
			if err != nil {
				t.Fatalf("Fetch(game_id=%s): %v", entry.GameID, err)
			}
			expectFetchStatus(t, entry.GameID, res)
			if res.Status == merchant.StatusOK && res.RecyclePrice == nil {
				t.Errorf("status ok but recycle_price is nil for game_id=%s", entry.GameID)
			}
		})
	}
}

// TestLaolierenRealDuplicateFetch exercises two consecutive calls to the same
// ID, the hottest pattern the CLI produces for one game, and asserts the pause
// actually ran between the two requests.
func TestLaolierenRealDuplicateFetch(t *testing.T) {
	requireIntegration(t)
	m := merchantByName(t, "laolieren")
	entry := game.MerchantEntry{GameID: "154"}

	first, err := m.Fetch(context.Background(), entry)
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	t.Logf("first=%s", resultSummary(first))

	start := time.Now()
	pause() // must wait before the second request
	second, err := m.Fetch(context.Background(), entry)
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	t.Logf("second=%s (gap=%s)", resultSummary(second), time.Since(start).Round(time.Millisecond))

	if first.Status != second.Status {
		t.Logf("status changed between fetches: %s -> %s", first.Status, second.Status)
	}
	if got := time.Since(start); got < 2*time.Second {
		t.Errorf("gap between requests was %s, expected at least 2s", got)
	}
}
