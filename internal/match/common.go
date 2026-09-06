// Package-level tuning tables for match: synonyms (rewrite), noise (remove),
// matchRules (alias scoring) and rejectRules (exclusion). They are kept apart
// from match.go so the matching logic and the data tables stay easy to review.
package match

// synonyms are applied (ordered, lowercase) before punctuation/noise stripping:
// alternate user spellings are rewritten to one canonical form so any variant
// hits the same game.
var synonyms = [][2]string{
	{"马力欧", "马里奥"},
	{"荒野之息", "旷野之息"},
	{"野炊", "旷野之息"},
	{"鬼屋", "洋馆"},
	{"斯普拉遁", "斯普拉顿"},
	{"塞尔达传说2", "塞尔达王国之泪"},
	{"咚奇刚", "大金刚"},
	{"驭天飞行", "御天飞行"},
	{"异度神剑", "异度之刃"},
	{"朋友聚会", "朋友收集"},
}

// noise are words removed from normalized strings so they do not drive matches.
var noise = []string{
	"ns", "ns1", "ns2", "switch", "nintendo",
	"游戏", "卡带", "中文", "港版", "日版", "标准版", "特别版", "传说", "超级", "兄弟",
}

// matchRules maps a slug to groups of aliases. Groups are ANDed together, the
// terms inside a group are ORed.
var matchRules = map[string][][]string{
	"zelda-breath-of-the-wild":            {{"旷野之息", "荒野之息", "野炊", "旷野"}},
	"zelda-tears-of-the-kingdom":          {{"王国之泪", "王泪"}},
	"super-mario-party-jamboree":          {{"空前盛会"}, {"马里奥", "马力欧"}},
	"zelda-breath-of-the-wild-bundle":     {{"旷野之息", "荒野之息", "野炊"}, {"同捆", "dlc", "+dlc", "扩充票", "扩充版", "全dlc"}},
	"zelda-echoes-of-wisdom":              {{"智慧的再现", "智慧再现"}},
	"fitness-boxing-3":                    {{"有氧拳击3", "健身拳击3"}},
	"super-mario-odyssey":                 {{"奥德赛"}},
	"kirby-and-the-forgotten-land":        {{"探索发现"}, {"星之卡比", "卡比"}},
	"pokemon-legends-z-a":                 {{"z-a", "za"}, {"宝可梦", "口袋妖怪"}},
	"splatoon-3":                          {{"喷射战士3", "斯普拉遁", "斯普拉顿", "喷3"}},
	"super-mario-bros-wonder":             {{"惊奇"}, {"马里奥", "马力欧"}},
	"mario-kart-8-deluxe":                 {{"赛车8", "马车8", "马8"}},
	"luigis-mansion-3":                    {{"鬼屋3", "洋馆3"}},
	"it-takes-two":                        {{"双人成行"}},
	"dave-the-diver":                      {{"潜水员戴夫"}},
	"tomodachi-life-living-the-dream":     {{"朋友聚会", "朋友收集", "梦想生活"}},
	"super-mario-party-jamboree-tv-ns2":   {{"空前盛会"}, {"tv"}},
	"zelda-tears-of-the-kingdom-ns2":      {{"王国之泪"}},
	"zelda-breath-of-the-wild-ns2":        {{"旷野之息", "荒野之息"}},
	"donkey-kong-bananza":                 {{"蕉力全开"}, {"大金刚", "咚奇刚"}},
	"mario-kart-world":                    {{"赛车世界", "马车世界", "马车9"}},
	"cyberpunk-2077-ns2":                  {{"赛博朋克2077", "赛博朋克 2077"}},
	"kirby-star-world-ns2":                {{"星耀世界"}, {"星之卡比", "卡比"}},
	"pokemon-pokopia":                     {{"pokopia", "宝森", "宝可森"}},
	"hyrule-warriors-age-of-imprisonment": {{"封印战记"}, {"塞尔达无双"}},
	"kirby-air-riders":                    {{"驭天飞行", "御天飞行"}},
	"xenoblade1":                          {{"异度之刃", "异度神剑", "xenoblade"}, {"决定版", "终极版", "definitive"}},
}

// rejectRules lists terms that disqualify a game when matched via rules,
// preventing e.g. the base 旷野之息 from matching a "同捆"/"dlc" query.
var rejectRules = map[string][]string{
	"zelda-breath-of-the-wild":     {"扩充", "dlc", "同捆"},
	"super-mario-party-jamboree":   {"tv"},
	"kirby-and-the-forgotten-land": {"星耀世界"},
	"pokemon-legends-z-a":          {"朱紫", "扩充票"},
	"splatoon-3":                   {"扩充票", "dlc"},
	"mario-kart-8-deluxe":          {"扩充版", "dlc"},
	"xenoblade1":                   {"异度之刃2", "异度神剑 2", "异度神剑2", "异度之刃3", "异度神剑 3", "异度神剑3", "异度之刃x", "异度神剑x", "黄金", "伊拉", "ns2", "订购"},
}
