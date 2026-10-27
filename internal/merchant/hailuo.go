package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"nscardprice/internal/game"
	"nscardprice/internal/httpx"
	"nscardprice/internal/parse"
)

const (
	hailuoDetailURL = "https://hailuo.dwzjd.com/api/product/detail/"
	hailuoListURL   = "https://hailuo.dwzjd.com/api/products"
)

type hailuo struct {
	client *httpx.Client
}

func newHailuo(c *httpx.Client) Merchant { return &hailuo{client: c} }

func (m *hailuo) Key() string  { return "hailuo" }
func (m *hailuo) Name() string { return "海螺" }

func hailuoHeaders() map[string]string {
	return map[string]string{
		"cb-lang":      "zh-CN",
		"appid":        "wx7f7b845076caaf81",
		"user-agent":   miniUserAgent,
		"xweb_xhr":     "1",
		"form-type":    "routine",
		"content-type": "application/json",
		"accept":       "*/*",
	}
}

func (m *hailuo) Fetch(ctx context.Context, entry game.MerchantEntry) (Result, error) {
	res := newResult(m.Key(), m.Name())
	res.ItemID = entry.GameID

	body, err := m.client.Get(ctx, hailuoDetailURL+entry.GameID, hailuoHeaders(), nil)
	if err != nil {
		return res, err
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return res, err
	}

	status := stringify(parse.Get(payload, "status"))
	switch status {
	case "400533":
		res.Status = StatusUnavailable
		res.Note = "海螺 商品不存在: " + stringify(parse.Get(payload, "msg"))
		return res, nil
	case "", "0", "200":
	default:
		return res, fmt.Errorf("海螺 status=%s msg=%s game_id=%s", status, stringify(parse.Get(payload, "msg")), entry.GameID)
	}

	store, _ := parse.Get(payload, "data.storeInfo").(map[string]any)
	if store == nil {
		res.Status = StatusUnavailable
		res.Note = "海螺 data.storeInfo unavailable"
		return res, nil
	}
	if id := stringify(store["id"]); id != "" {
		res.ItemID = id
	}

	sell, sellOK := parse.AsMoney(store["price"])
	if sellOK {
		res.SellPrice = f64ptr(sell)
	}
	if stringify(store["out_recycle"]) == "0" {
		res.Status = StatusUnavailable
		res.Note = "海螺 out_recycle=0"
		return res, nil
	}

	diff, diffOK := parse.AsMoney(store["out_diff"])
	if !sellOK || !diffOK {
		return res, fmt.Errorf("海螺 missing data.storeInfo.price or data.storeInfo.out_diff for game_id=%s", entry.GameID)
	}
	recycle, ok := hailuoRecyclePrice(sell, diff)
	if !ok {
		res.Status = StatusUnavailable
		res.Note = "海螺 recycle_price<=0"
		return res, nil
	}
	res.RecyclePrice = f64ptr(recycle)
	res.Note = "recycle_price=data.storeInfo.price-data.storeInfo.out_diff"
	return res, nil
}

// hailuoRecyclePrice derives the recycle price from the sell price and the
// merchant's deduction. A non-positive result means 海螺 does not actually take
// the item, so callers must report it as unavailable rather than a negative or
// zero price.
func hailuoRecyclePrice(sell, diff float64) (float64, bool) {
	price := sell - diff
	return price, price > 0
}

// hailuoPlatform maps 海螺's category ids to platforms. 1 is Switch, 41 is
// Switch 2; other categories (PS/Xbox/retro) are marked "其他" so IsNS rejects
// them even when their names carry no platform marker.
func hailuoPlatform(cateID string) string {
	parts := strings.Split(cateID, ",")
	for _, p := range parts {
		if strings.TrimSpace(p) == "41" {
			return "Nintendo Switch 2"
		}
	}
	for _, p := range parts {
		if strings.TrimSpace(p) == "1" {
			return "Nintendo Switch"
		}
	}
	return "其他"
}

// Catalog pages through 海螺's full product list.
func (m *hailuo) Catalog(ctx context.Context, limit int) ([]CatalogItem, error) {
	const pageSize = 50
	fetchPage := func(page int) ([]CatalogItem, error) {
		if err := waitCatalogPage(ctx); err != nil {
			return nil, err
		}
		params := url.Values{
			"cid":                {"0"},
			"sid":                {"0"},
			"keyword":            {""},
			"priceOrder":         {""},
			"salesOrder":         {""},
			"news":               {"0"},
			"best":               {"0"},
			"store_label_id":     {""},
			"page":               {stringify(page)},
			"limit":              {stringify(pageSize)},
			"coupon_category_id": {""},
			"productId":          {""},
		}
		body, err := m.client.Get(ctx, hailuoListURL, hailuoHeaders(), params)
		if err != nil {
			return nil, err
		}
		var payload any
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
		var items []CatalogItem
		for _, row := range firstRows(payload, "data") {
			m, ok := row.(map[string]any)
			if !ok {
				continue
			}
			id := firstString(m, "id", "productId")
			name := firstString(m, "store_name", "name", "title")
			if id == "" || name == "" {
				continue
			}
			items = append(items, CatalogItem{
				Name:     name,
				GameID:   id,
				Platform: hailuoPlatform(firstString(m, "cate_id")),
			})
		}
		return items, nil
	}
	return collectCatalog(fetchPage, limit)
}
