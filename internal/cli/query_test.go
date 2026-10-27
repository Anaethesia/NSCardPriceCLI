package cli

import (
	"testing"

	"nscardprice/internal/collector"
	"nscardprice/internal/merchant"
)

func TestQueryShouldExit(t *testing.T) {
	result := func(statuses ...string) collector.GameResult {
		game := collector.GameResult{Slug: "test"}
		for _, status := range statuses {
			game.Results = append(game.Results, collector.MerchantResult{Status: status})
		}
		return game
	}

	tests := []struct {
		name    string
		results []collector.GameResult
		want    bool
	}{
		{
			name:    "no results",
			results: nil,
			want:    false,
		},
		{
			name:    "all ok",
			results: []collector.GameResult{result(merchant.StatusOK, merchant.StatusOK)},
			want:    false,
		},
		{
			name:    "unavailable and skipped are usable",
			results: []collector.GameResult{result(merchant.StatusUnavailable, merchant.StatusSkipped)},
			want:    false,
		},
		{
			name:    "one merchant error is tolerated",
			results: []collector.GameResult{result(merchant.StatusOK, merchant.StatusError)},
			want:    false,
		},
		{
			name: "errors spanning multiple games are tolerated when data remains",
			results: []collector.GameResult{
				result(merchant.StatusError),
				result(merchant.StatusOK, merchant.StatusError),
			},
			want: false,
		},
		{
			name:    "single merchant failure exits non-zero",
			results: []collector.GameResult{result(merchant.StatusError)},
			want:    true,
		},
		{
			name: "all merchants failed across all games",
			results: []collector.GameResult{
				result(merchant.StatusError, merchant.StatusError),
				result(merchant.StatusError),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queryShouldExit(tt.results); got != tt.want {
				t.Fatalf("queryShouldExit() = %v, want %v", got, tt.want)
			}
		})
	}
}
