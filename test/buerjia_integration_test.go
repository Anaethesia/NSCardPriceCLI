//go:build integration

package test

import (
	"context"
	"testing"

	"nscardprice/internal/game"
	"nscardprice/internal/merchant"
)

// TestBuerjiaReal queries 不二家 for a few known-good game IDs.
func TestBuerjiaReal(t *testing.T) {
	requireIntegration(t)
	m := merchantByName(t, "buerjia")

	ids := []game.MerchantEntry{
		{GameID: "3"}, // 塞尔达传说 旷野之息
		{GameID: "2"}, // 塞尔达传说 王国之泪
		{GameID: "5"}, // 马里奥赛车 8 豪华版
	}

	for _, entry := range ids {
		pause()
		t.Run("buerjia-"+entry.GameID, func(t *testing.T) {
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
