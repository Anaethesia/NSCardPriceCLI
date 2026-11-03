package collector

import (
	"context"
	"strings"
	"testing"

	"nscardprice/internal/game"
	"nscardprice/internal/merchant"
)

type stubMerchant struct {
	key  string
	name string
	res  merchant.Result
	err  error
}

func (s stubMerchant) Key() string  { return s.key }
func (s stubMerchant) Name() string { return s.name }

func (s stubMerchant) Fetch(context.Context, game.MerchantEntry) (merchant.Result, error) {
	return s.res, s.err
}

// TestFetchOneUnmappedVsUnavailable locks the distinction between "no game_id
// configured" (unmapped, a data gap) and "merchant explicitly refuses the
// cartridge" (unavailable, a merchant state).
func TestFetchOneUnmappedVsUnavailable(t *testing.T) {
	reg := merchant.Registry{
		"huoqiangshou": stubMerchant{
			key:  "huoqiangshou",
			name: "火枪手",
			res:  merchant.Result{Merchant: "huoqiangshou", Status: merchant.StatusUnavailable},
		},
	}
	tests := []struct {
		name        string
		merchantIDs map[string]game.MerchantEntry
		dryRun      bool
		wantStatus  string
		wantItemID  string
		wantNote    string
		wantError   string
	}{
		{
			name:        "merchant key absent is unmapped",
			merchantIDs: nil,
			wantStatus:  merchant.StatusUnmapped,
			wantNote:    "missing game_id",
		},
		{
			name:        "empty game_id is unmapped",
			merchantIDs: map[string]game.MerchantEntry{"huoqiangshou": {GameID: ""}},
			wantStatus:  merchant.StatusUnmapped,
			wantNote:    "missing game_id",
		},
		{
			name:        "configured game_id dry-run is ok",
			merchantIDs: map[string]game.MerchantEntry{"huoqiangshou": {GameID: "3616"}},
			dryRun:      true,
			wantStatus:  merchant.StatusOK,
			wantItemID:  "3616",
			wantNote:    "dry-run",
		},
		{
			name:        "configured game_id refused by merchant is unavailable",
			merchantIDs: map[string]game.MerchantEntry{"huoqiangshou": {GameID: "3616"}},
			wantStatus:  merchant.StatusUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.Game{Slug: "test", MerchantIDs: tt.merchantIDs}
			got := fetchOne(context.Background(), g, "huoqiangshou", reg, tt.dryRun)
			if got.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.ItemID != tt.wantItemID {
				t.Errorf("item_id = %q, want %q", got.ItemID, tt.wantItemID)
			}
			if tt.wantNote != "" && !strings.Contains(got.Note, tt.wantNote) {
				t.Errorf("note = %q, want contains %q", got.Note, tt.wantNote)
			}
			if got.Error != tt.wantError {
				t.Errorf("error = %q, want %q", got.Error, tt.wantError)
			}
		})
	}
}
