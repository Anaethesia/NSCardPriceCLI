//go:build integration

package test

import (
	"context"
	"testing"

	"nscardprice/internal/game"
	"nscardprice/internal/merchant"
)

// TestMogushijianReal queries 蘑菇时间 for known-good card IDs. Prices come
// from the public catalog endpoint, which needs no token.
func TestMogushijianReal(t *testing.T) {
	requireIntegration(t)
	m := merchantByName(t, "mogushijian")

	ids := []game.MerchantEntry{
		{GameID: "2500"},     // 塞尔达传说 旷野之息
		{GameID: "617"},      // 塞尔达传说 王国之泪
		{GameID: "2657"},     // NS2 赛博朋克 2077
		{GameID: "99999999"}, // 不存在，应返回 unavailable
	}

	for _, entry := range ids {
		pause()
		t.Run("mogushijian-"+entry.GameID, func(t *testing.T) {
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
