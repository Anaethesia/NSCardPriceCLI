// Package game defines the core domain model for games and their merchant
// identifiers. It has no external dependencies so it can be reused by every
// layer without creating import cycles.
package game

import "sort"

// MerchantEntry holds the identifiers a game uses at one merchant.
type MerchantEntry struct {
	// GameID is the merchant product id. Empty means the game is not tracked
	// at that merchant (query is skipped).
	GameID string `json:"game_id"`
	// UUID is required by Huozhou Xizi (杭州西子) and can fall back to a
	// default value.
	UUID string `json:"uuid,omitempty"`
	// SkuID is used by Huozhou Xizi to resolve the exact spec (bianma).
	SkuID string `json:"sku_id,omitempty"`
}

// Game represents a tracked Nintendo cartridge game.
type Game struct {
	Slug        string                   `json:"slug"`
	Name        string                   `json:"name"`
	Platform    string                   `json:"platform"`
	Enabled     bool                     `json:"enabled"`
	MerchantIDs map[string]MerchantEntry `json:"merchant_ids"`
	// SearchKeyword is an optional extra alias used by fuzzy matching, e.g.
	// "动物森友会" for the game whose name is "捡树枝".
	SearchKeyword string `json:"search_keyword,omitempty"`
}

// MerchantEntry returns the entry for a given merchant key and a bool
// indicating whether an entry exists.
func (g *Game) Merchant(key string) (MerchantEntry, bool) {
	if g == nil || g.MerchantIDs == nil {
		return MerchantEntry{}, false
	}
	entry, ok := g.MerchantIDs[key]
	// A key present but with an empty game_id is treated as "not tracked".
	if ok && entry.GameID == "" {
		return MerchantEntry{}, false
	}
	return entry, ok
}

// EnabledMerchants returns the keys of merchants that have a usable game_id,
// in sorted order for deterministic output.
func (g *Game) EnabledMerchants() []string {
	keys := make([]string, 0, len(g.MerchantIDs))
	for key := range g.MerchantIDs {
		if _, ok := g.Merchant(key); ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
