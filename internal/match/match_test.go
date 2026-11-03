package match

import (
	"reflect"
	"testing"

	"nscardprice/internal/game"
)

// TestQueryNS2PlatformHint locks the NS2 platform hint: a query that explicitly
// asks for NS2 must not fall back to the NS1 edition, and the aliases added for
// NS2 editions must still resolve.
func TestQueryNS2PlatformHint(t *testing.T) {
	games := []game.Game{
		{Slug: "zelda-breath-of-the-wild", Name: "塞尔达传说 旷野之息", Platform: "Nintendo Switch", Enabled: true},
		{Slug: "zelda-breath-of-the-wild-ns2", Name: "NS2 旷野之息", Platform: "Nintendo Switch 2", Enabled: true},
		{Slug: "zelda-tears-of-the-kingdom", Name: "塞尔达传说 王国之泪", Platform: "Nintendo Switch", Enabled: true},
		{Slug: "zelda-tears-of-the-kingdom-ns2", Name: "NS2 王国之泪", Platform: "Nintendo Switch 2", Enabled: true},
		{Slug: "fitness-boxing-3", Name: "有氧拳击 3", Platform: "Nintendo Switch", Enabled: true},
		{Slug: "fitness-boxing-3-ns2", Name: "NS2 有氧拳击3", Platform: "Nintendo Switch 2", Enabled: true},
		{Slug: "yaoxitujian", Name: "NS2 耀西图鉴", Platform: "Nintendo Switch 2", Enabled: true,
			SearchKeyword: "耀西与不可思议的图鉴 不可思议图鉴"},
	}

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "NS2 prefix keeps only the NS2 edition",
			query: "NS2 王泪",
			want:  []string{"zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "二代卡 marker keeps only the NS2 edition",
			query: "王泪二代卡",
			want:  []string{"zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "shorter 二代 marker keeps only the NS2 edition",
			query: "王泪二代",
			want:  []string{"zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "二代 marker resolves the 旷野 alias",
			query: "旷野二代",
			want:  []string{"zelda-breath-of-the-wild-ns2"},
		},
		{
			name:  "Switch 2 spelling is also an NS2 hint",
			query: "Switch 2 王泪",
			want:  []string{"zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "spaced NS 2 spelling is also an NS2 hint",
			query: "NS 2 王泪",
			want:  []string{"zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "spaced 二代 卡 spelling is also an NS2 hint",
			query: "王泪 二代 卡",
			want:  []string{"zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "short alias without marker stays ambiguous",
			query: "王泪",
			want:  []string{"zelda-tears-of-the-kingdom", "zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "full title without marker stays ambiguous",
			query: "王国之泪",
			want:  []string{"zelda-tears-of-the-kingdom", "zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "NS2 prefix keeps only the NS2 full title",
			query: "NS2 王国之泪",
			want:  []string{"zelda-tears-of-the-kingdom-ns2"},
		},
		{
			name:  "NS2 prefix resolves the 旷野 alias",
			query: "NS2 旷野",
			want:  []string{"zelda-breath-of-the-wild-ns2"},
		},
		{
			name:  "二代卡 resolves the 野炊 alias",
			query: "野炊二代卡",
			want:  []string{"zelda-breath-of-the-wild-ns2"},
		},
		{
			name:  "旷野之息 without marker stays ambiguous",
			query: "旷野之息",
			want:  []string{"zelda-breath-of-the-wild", "zelda-breath-of-the-wild-ns2"},
		},
		{
			name:  "NS2 prefix resolves 有氧拳击3",
			query: "NS2 有氧拳击3",
			want:  []string{"fitness-boxing-3-ns2"},
		},
		{
			name:  "二代卡 resolves the 拳击3 alias",
			query: "拳击3二代卡",
			want:  []string{"fitness-boxing-3-ns2"},
		},
		{
			name:  "有氧拳击3 without marker stays ambiguous",
			query: "有氧拳击3",
			want:  []string{"fitness-boxing-3", "fitness-boxing-3-ns2"},
		},
		{
			name:  "Yoshi short alias still matches",
			query: "耀西图鉴",
			want:  []string{"yaoxitujian"},
		},
		{
			name:  "Yoshi full alias still matches",
			query: "耀西与不可思议的图鉴",
			want:  []string{"yaoxitujian"},
		},
		{
			name:  "NS2 prefix filters to the Switch 2 platform",
			query: "NS2 耀西",
			want:  []string{"yaoxitujian"},
		},
		{
			name:  "NS2 marker with no NS2 game returns nothing",
			query: "NS2 不存在的游戏",
			want:  nil,
		},
		{
			name:  "NS2 marker excludes NS1-only games",
			query: "NS2 马车8",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugs(Query(games, tt.query))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Query(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func slugs(games []game.Game) []string {
	if len(games) == 0 {
		return nil
	}
	out := make([]string, len(games))
	for i, g := range games {
		out[i] = g.Slug
	}
	return out
}
