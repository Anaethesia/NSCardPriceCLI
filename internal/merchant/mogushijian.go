package merchant

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"nscardprice/internal/game"
	"nscardprice/internal/httpx"
	"nscardprice/internal/parse"
)

const mogushijianListURL = "https://api.zhualegezhua.cn/alia/used/recycle/cardsList"

// mogushijian exposes recycle prices through a public, token-free catalog
// endpoint; there is no per-product detail call, so the catalog is fetched
// once per process and reused for both Fetch and Catalog.
type mogushijian struct {
	client *httpx.Client

	mu     sync.Mutex
	loaded bool
	items  []CatalogItem
	prices map[string]float64
}

func newMogushijian(c *httpx.Client) Merchant {
	return &mogushijian{client: c}
}

func (m *mogushijian) Key() string  { return "mogushijian" }
func (m *mogushijian) Name() string { return "蘑菇时间" }

func mogushijianHeaders() map[string]string {
	return map[string]string{
		"content-type": "application/json",
		"alianame":     "alia2",
		"accept":       "*/*",
	}
}

func (m *mogushijian) Fetch(ctx context.Context, entry game.MerchantEntry) (Result, error) {
	res := newResult(m.Key(), m.Name())
	res.ItemID = entry.GameID

	if err := m.loadCatalog(ctx); err != nil {
		return res, err
	}
	m.mu.Lock()
	price, ok := m.prices[entry.GameID]
	m.mu.Unlock()
	if !ok {
		res.Status = StatusUnavailable
		res.Note = "蘑菇时间 目录中未找到该商品"
		return res, nil
	}
	res.RecyclePrice = f64ptr(price)
	res.Note = "recycle_price=catalog_item.price"
	return res, nil
}

func (m *mogushijian) loadCatalog(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loaded {
		return nil
	}

	items := make([]CatalogItem, 0, 300)
	prices := make(map[string]float64, 300)
	seen := map[string]bool{}
	page := 1
	for page > 0 && page <= 200 {
		if err := waitCatalogPage(ctx); err != nil {
			return err
		}
		body, err := m.client.PostJSON(ctx, mogushijianListURL, mogushijianHeaders(), map[string]any{"page": page})
		if err != nil {
			return err
		}
		var payload any
		if err := json.Unmarshal(body, &payload); err != nil {
			return err
		}
		rows := firstRows(payload, "list", "data.list", "data")
		for _, row := range rows {
			card, ok := row.(map[string]any)
			if !ok {
				continue
			}
			id := firstString(card, "cardsId", "id")
			name := firstString(card, "name", "cardName")
			if id == "" || name == "" || seen[id] {
				continue
			}
			seen[id] = true
			price, priceOK := parse.AsMoney(card["price"])
			if !priceOK {
				continue
			}
			items = append(items, CatalogItem{
				Name:     name,
				GameID:   id,
				Platform: mogushijianPlatform(firstString(card, "generationType")),
			})
			prices[id] = price
		}
		hasMore := parse.Get(payload, "hasMore")
		more, _ := hasMore.(bool)
		if !more {
			break
		}
		next := stringify(parse.Get(payload, "next"))
		if next == "" || next == stringify(page) {
			page++
		} else if n, err := strconv.Atoi(next); err != nil || n <= 0 {
			page++
		} else {
			page = n
		}
	}

	m.items = items
	m.prices = prices
	m.loaded = true
	return nil
}

// mogushijianPlatform maps generationType to a platform: 1 is Switch, 2 is
// Switch 2.
func mogushijianPlatform(generation string) string {
	switch generation {
	case "2":
		return "Nintendo Switch 2"
	case "1":
		return "Nintendo Switch"
	default:
		return "其他"
	}
}

// Catalog returns 蘑菇时间's full recycle catalog.
func (m *mogushijian) Catalog(ctx context.Context, limit int) ([]CatalogItem, error) {
	if err := m.loadCatalog(ctx); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(m.items)
	if limit > 0 && limit < n {
		n = limit
	}
	return append([]CatalogItem(nil), m.items[:n]...), nil
}
