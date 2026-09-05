package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"

	"nscardprice/internal/game"
)

// customItem is the JSON shape of data/custom.json. It mirrors game.Game,
// except that "enabled" is optional and defaults to true so entries take
// effect without an explicit flag.
type customItem struct {
	Slug          string                        `json:"slug"`
	Name          string                        `json:"name"`
	Platform      string                        `json:"platform"`
	Enabled       *bool                         `json:"enabled"`
	SearchKeyword string                        `json:"search_keyword,omitempty"`
	MerchantIDs   map[string]game.MerchantEntry `json:"merchant_ids"`
}

// loadCustomGames reads a custom library file into games. An empty library
// yields an empty slice (not an error) so `query -custom` can report "为空"
// cleanly; a missing file is an error.
func loadCustomGames(path string) ([]game.Game, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var items []customItem
	if len(bytes.TrimSpace(b)) != 0 {
		if err := json.Unmarshal(b, &items); err != nil {
			return nil, err
		}
	}
	games := make([]game.Game, 0, len(items))
	for _, it := range items {
		if it.Enabled != nil && !*it.Enabled {
			continue
		}
		games = append(games, game.Game{
			Slug:          it.Slug,
			Name:          it.Name,
			Platform:      it.Platform,
			Enabled:       true,
			SearchKeyword: it.SearchKeyword,
			MerchantIDs:   it.MerchantIDs,
		})
	}
	sort.Slice(games, func(i, j int) bool { return games[i].Slug < games[j].Slug })
	return games, nil
}
