// Package merchant defines the abstraction over the four storefronts: each
// merchant knows how to fetch one game's recycle price and, for the list
// command, how to page through its full product catalog.
package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"nscardprice/internal/game"
	"nscardprice/internal/httpx"
	"nscardprice/internal/parse"
)

// Result statuses reported to callers.
const (
	StatusOK          = "ok"
	StatusUnavailable = "unavailable"
	StatusSkipped     = "skipped"
	StatusError       = "error"
	StatusReady       = "ready"
)

// miniUserAgent is the WeChat mini-program user-agent used by these
// storefronts' mini program APIs.
const miniUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 MicroMessenger/7.0.20.1781(0x6700143B) NetType/WIFI MiniProgramEnv/Windows WindowsWechat/WMPF WindowsWechat(0x63090a13) UnifiedPCWindowsWechat(0xf2541939) XWEB/19841"

// Result is a single merchant's parsed price for one game.
type Result struct {
	Merchant     string
	MerchantName string
	Status       string
	RecyclePrice *float64
	SellPrice    *float64
	Currency     string
	ItemID       string
	Note         string
	Error        string
}

func newResult(key, name string) Result {
	return Result{Merchant: key, MerchantName: name, Currency: "CNY", Status: StatusOK}
}

// Merchant fetches one game's price from a single storefront.
type Merchant interface {
	Key() string
	Name() string
	Fetch(ctx context.Context, entry game.MerchantEntry) (Result, error)
}

// CatalogItem is one product from a merchant's full catalog.
type CatalogItem struct {
	Name     string `json:"name"`
	GameID   string `json:"game_id"`
	Platform string `json:"platform,omitempty"`
}

// nonNSRe matches catalog rows that are not Nintendo Switch cartridges (PS
// consoles, handhelds, PC, peripherals/周边, shipping-only fee links).
var nonNSRe = regexp.MustCompile(`(?i)(ps\d|xbox|x360|psv|playstation|nds|3ds|steam|pc版|zhoubian|邮费)`)

// IsNS reports whether a catalog item is a Nintendo Switch / Switch 2
// cartridge. When the API provides an authoritative platform field it wins;
// otherwise rows whose names carry a non-NS marker (PS/Xbox/PC/周边/邮费…)
// are rejected and the rest are kept, since the four catalogs label
// cross-platform titles in the name but leave NS titles bare.
func IsNS(it CatalogItem) bool {
	if it.Platform != "" {
		p := strings.ToLower(it.Platform)
		return strings.Contains(p, "switch") || strings.Contains(p, "ns")
	}
	return !nonNSRe.MatchString(it.Name)
}

// Cataloger lists a merchant's catalog. limit<=0 means "all pages".
type Cataloger interface {
	Catalog(ctx context.Context, limit int) ([]CatalogItem, error)
}

// Registry maps merchant keys to implementations.
type Registry map[string]Merchant

// NewRegistry returns all supported merchants registered by key.
func NewRegistry(c *httpx.Client) Registry {
	r := Registry{}
	for _, m := range []Merchant{
		newLaolieren(c),
		newHuoqiangshou(c),
		newBuerjia(c),
		newHangzhouXizi(c),
	} {
		r[m.Key()] = m
	}
	return r
}

// Catalog returns the cataloger for a key, or false if the merchant does not
// support full-catalog listing.
func (r Registry) Catalog(key string) (Cataloger, bool) {
	m, ok := r[key]
	if !ok {
		return nil, false
	}
	c, ok := m.(Cataloger)
	return c, ok
}

// ResolveKey resolves a user-supplied merchant key or Chinese name to a
// registry key.
func (r Registry) ResolveKey(input string) (string, bool) {
	if _, ok := r[input]; ok {
		return input, true
	}
	for key, m := range r {
		if m.Name() == input {
			return key, true
		}
	}
	return "", false
}

// Keys returns the registered keys in sorted order.
func (r Registry) Keys() []string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func f64ptr(v float64) *float64 { return &v }

// stringify converts a loosely-typed JSON scalar into its string form,
// avoiding scientific notation for integer-valued floats.
func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprint(v)
	}
}

// firstString returns the first present, non-empty value among the keys.
func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s := stringify(m[k]); s != "" {
			return s
		}
	}
	return ""
}

// firstRows returns the first array found at one of the dotted paths.
func firstRows(payload any, paths ...string) []any {
	for _, p := range paths {
		if v := parse.Get(payload, p); v != nil {
			if rows, ok := v.([]any); ok {
				return rows
			}
		}
	}
	return nil
}

// collectCatalog drives pagination for a merchant. fetch(page) returns the
// items of that page. limit<=0 means keep going until an empty page.
func collectCatalog(fetch func(page int) ([]CatalogItem, error), limit int) ([]CatalogItem, error) {
	const maxPages = 1000
	var out []CatalogItem
	seen := map[string]bool{}

	for page := 1; page <= maxPages; page++ {
		items, err := fetch(page)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}
		for _, it := range items {
			if it.GameID == "" || seen[it.GameID] {
				continue
			}
			seen[it.GameID] = true
			out = append(out, it)
			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
		if limit <= 0 {
			continue
		}
	}
	return out, nil
}
