package merchant

import "testing"

func TestBuerjiaUnavailable(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "日版无中文不收", want: true},
		{name: "NS1 女神异闻录P5R 皇家版 P5R（日版无中文，不收）", want: true},
		{name: "NS1 怪物猎人 崛起 rise（欧美版不收）", want: true},
		{name: "NS1 火焰纹章 风花雪月", want: false},
		{name: "NS1 宝可梦 皮卡丘", want: false},
		{name: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buerjiaUnavailable(tt.name); got != tt.want {
				t.Fatalf("buerjiaUnavailable(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
