// Package collector orchestrates concurrent price fetching across games and
// merchants, aggregating results in a deterministic order.
package collector

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"nscardprice/internal/game"
	"nscardprice/internal/merchant"
)

// MerchantResult is the JSON-able form of one merchant's result for a game.
type MerchantResult struct {
	Merchant     string   `json:"merchant"`
	MerchantName string   `json:"merchant_name"`
	Status       string   `json:"status"`
	RecyclePrice *float64 `json:"recycle_price"`
	SellPrice    *float64 `json:"sell_price"`
	Currency     string   `json:"currency"`
	ItemID       string   `json:"item_id,omitempty"`
	Note         string   `json:"note,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// GameResult aggregates all merchant results for one game.
type GameResult struct {
	FetchedAt string           `json:"fetched_at"`
	Slug      string           `json:"slug"`
	Name      string           `json:"name"`
	Platform  string           `json:"platform"`
	Results   []MerchantResult `json:"results"`
}

// Options controls which merchants to query and how many requests run at once.
type Options struct {
	Merchants   []string // empty means all supported merchants
	Concurrency int
	DryRun      bool
}

// HasErrors reports whether any result ended in an error status.
func (r GameResult) HasErrors() bool {
	for _, m := range r.Results {
		if m.Status == merchant.StatusError {
			return true
		}
	}
	return false
}

// Collect fetches, concurrently, every (game × merchant) pair and returns the
// results grouped by game, in input order.
func Collect(ctx context.Context, games []game.Game, reg merchant.Registry, opts Options) ([]GameResult, error) {
	keys, err := filteredKeys(reg, opts.Merchants)
	if err != nil {
		return nil, err
	}
	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	results := make([][]merchant.Result, len(games))
	for i := range results {
		results[i] = make([]merchant.Result, len(keys))
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	for gi, g := range games {
		for mi, key := range keys {
			wg.Add(1)
			go func(gi, mi int, g game.Game, key string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				results[gi][mi] = fetchOne(ctx, g, key, reg, opts.DryRun)
			}(gi, mi, g, key)
		}
	}
	wg.Wait()

	fetchedAt := time.Now().Format(time.RFC3339)
	out := make([]GameResult, 0, len(games))
	for gi, g := range games {
		gr := GameResult{FetchedAt: fetchedAt, Slug: g.Slug, Name: g.Name, Platform: g.Platform}
		for mi, key := range keys {
			gr.Results = append(gr.Results, toMerchantResult(results[gi][mi], key, reg))
		}
		out = append(out, gr)
	}
	return out, nil
}

func fetchOne(ctx context.Context, g game.Game, key string, reg merchant.Registry, dryRun bool) merchant.Result {
	m := reg[key]
	entry, ok := g.Merchant(key)
	if !ok {
		return merchant.Result{
			Merchant:     key,
			MerchantName: m.Name(),
			Status:       merchant.StatusSkipped,
			Currency:     "CNY",
			Note:         "missing game_id",
			Error:        "missing game_id",
		}
	}
	if dryRun {
		return merchant.Result{
			Merchant:     key,
			MerchantName: m.Name(),
			Status:       merchant.StatusOK,
			Currency:     "CNY",
			ItemID:       entry.GameID,
			Note:         "dry-run; remote API not called",
		}
	}
	r, err := m.Fetch(ctx, entry)
	if err != nil {
		r.Error = err.Error()
		r.Status = merchant.StatusError
	}
	return r
}

func toMerchantResult(r merchant.Result, key string, reg merchant.Registry) MerchantResult {
	name := r.MerchantName
	if name == "" && reg[key] != nil {
		name = reg[key].Name()
	}
	return MerchantResult{
		Merchant:     key,
		MerchantName: name,
		Status:       r.Status,
		RecyclePrice: r.RecyclePrice,
		SellPrice:    r.SellPrice,
		Currency:     r.Currency,
		ItemID:       r.ItemID,
		Note:         r.Note,
		Error:        r.Error,
	}
}

func filteredKeys(reg merchant.Registry, requested []string) ([]string, error) {
	if len(requested) == 0 {
		return reg.Keys(), nil
	}
	seen := map[string]bool{}
	keys := make([]string, 0, len(requested))
	for _, r := range requested {
		key, ok := reg.ResolveKey(r)
		if !ok {
			return nil, fmt.Errorf("未知商家: %s", r)
		}
		if !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys, nil
}
