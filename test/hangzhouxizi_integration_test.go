//go:build integration

package test

import (
	"context"
	"testing"

	"nscardprice/internal/game"
	"nscardprice/internal/merchant"
)

// TestHangzhouXiziReal queries 杭州西子 for a few known-good games. Unlike the
// other merchants it needs a uuid and sku_id to resolve the exact spec.
func TestHangzhouXiziReal(t *testing.T) {
	requireIntegration(t)
	m := merchantByName(t, "hangzhouxizi")

	const uuid = "6c99e77fba4f44f9ac46b767fd079f0b"
	entries := []game.MerchantEntry{
		{GameID: "50", UUID: uuid, SkuID: "786"},  // 塞尔达传说 旷野之息
		{GameID: "13", UUID: uuid, SkuID: "822"},  // 塞尔达传说 王国之泪
		{GameID: "107", UUID: uuid, SkuID: "148"}, // 马里奥赛车 8 豪华版
	}

	for _, entry := range entries {
		pause()
		t.Run("hangzhouxizi-"+entry.GameID, func(t *testing.T) {
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
