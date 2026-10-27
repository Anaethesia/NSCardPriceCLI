package merchant

import "testing"

func TestHailuoRecyclePrice(t *testing.T) {
	tests := []struct {
		name   string
		sell   float64
		diff   float64
		want   float64
		wantOK bool
	}{
		{name: "normal deduction", sell: 165, diff: 10, want: 155, wantOK: true},
		{name: "zero deduction", sell: 165, diff: 0, want: 165, wantOK: true},
		{name: "negative result is unavailable", sell: 9, diff: 30, want: -21, wantOK: false},
		{name: "zero result is unavailable", sell: 30, diff: 30, want: 0, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := hailuoRecyclePrice(tt.sell, tt.diff)
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("hailuoRecyclePrice(%v, %v) = (%v, %v), want (%v, %v)", tt.sell, tt.diff, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
