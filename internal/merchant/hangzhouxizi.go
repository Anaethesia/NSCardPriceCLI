package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nscardprice/internal/game"
	"nscardprice/internal/httpx"
	"nscardprice/internal/parse"
)

const (
	xiziDetailURL = "https://xcx.hzxzdwsc.com/api/index/detail"
	xiziGuigeURL  = "https://xcx.hzxzdwsc.com/api/index/guige"
	xiziListURL   = "https://xcx.hzxzdwsc.com/api/index/goods"
)

// xiziRecycleKeys is the ordered list of candidate recycle-price fields in the
// guige response, mirroring the original XIZI_RECYCLE_PRICE_KEYS.
var xiziRecycleKeys = []string{
	"hs_price_2", "hs_price_1", "hs_price", "recycle_price", "recyclePrice", "receivePrice", "price",
}

type hangzhouxizi struct {
	client *httpx.Client
}

func newHangzhouXizi(c *httpx.Client) Merchant { return &hangzhouxizi{client: c} }

func (m *hangzhouxizi) Key() string  { return "hangzhouxizi" }
func (m *hangzhouxizi) Name() string { return "杭州西子" }

func xiziHeaders() map[string]string {
	return map[string]string{
		"content-type": "application/json",
		"referer":      "https://xcx.hzxzdwsc.com",
		"user-agent":   miniUserAgent,
		"xweb_xhr":     "1",
	}
}

func (m *hangzhouxizi) Fetch(ctx context.Context, entry game.MerchantEntry) (Result, error) {
	res := newResult(m.Key(), m.Name())
	res.ItemID = entry.GameID

	detailBody, err := m.client.PostJSON(ctx, xiziDetailURL, xiziHeaders(), map[string]any{"id": entry.GameID})
	if err != nil {
		return res, err
	}
	var detail any
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		return res, err
	}

	if msg := xiziError(res, detail); msg != "" {
		return res, fmt.Errorf("%s", msg)
	}

	bianma := xiziBianma(detail, entry.SkuID)
	if bianma == "" {
		return xiziGoodsFallback(res, detail, entry.GameID), nil
	}

	guigeBody, err := m.client.PostJSON(ctx, xiziGuigeURL, xiziHeaders(),
		map[string]any{"id": entry.GameID, "guige": bianma})
	if err != nil {
		return res, err
	}
	var guige any
	if err := json.Unmarshal(guigeBody, &guige); err != nil {
		return res, err
	}

	return xiziParseGuige(res, detail, guige, entry.GameID)
}

// xiziError inspects known error codes and returns a message, or "" if none.
func xiziError(res Result, detail any) string {
	if code := stringify(parse.Get(detail, "code")); code == "201" && strings.Contains(strings.ToLower(stringify(parse.Get(detail, "message"))), "uuid") {
		return "杭州西子 requires a valid uuid"
	}
	if code := stringify(parse.Get(detail, "code")); code == "99999" && strings.Contains(strings.ToLower(stringify(parse.Get(detail, "message"))), "token") {
		return "杭州西子 rejected the request token"
	}
	return ""
}

// xiziBianma picks the guige bianma to request, preferring the game's sku_id.
func xiziBianma(detail any, preferredSkuID string) string {
	groups, _ := parse.Get(detail, "data.guige").([]any)
	var candidates []map[string]any
	for _, group := range groups {
		gm, ok := group.(map[string]any)
		if !ok {
			continue
		}
		values, _ := gm["gg_value"].([]any)
		for _, v := range values {
			vm, ok := v.(map[string]any)
			if !ok || stringify(vm["bianma"]) == "" {
				continue
			}
			candidates = append(candidates, vm)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	if preferredSkuID != "" {
		for _, c := range candidates {
			if stringify(c["id"]) == preferredSkuID || stringify(c["sku_id"]) == preferredSkuID || stringify(c["bianma"]) == preferredSkuID {
				return stringify(c["bianma"])
			}
		}
	}
	for _, c := range candidates {
		title := stringify(c["title"])
		if strings.Contains(title, "二手") && strings.Contains(title, "盒") {
			return stringify(c["bianma"])
		}
	}
	return stringify(candidates[0]["bianma"])
}

func xiziParseGuige(res Result, detail, guige any, gameID string) (Result, error) {
	goods, _ := parse.Get(detail, "data.goods").(map[string]any)
	if goods != nil {
		if n := stringify(goods["id"]); n != "" {
			res.ItemID = n
		}
	}

	gp := parse.Get(guige, "data.price")
	sell, sellOK := firstMoneyAny(gp, "price")
	if !sellOK && goods != nil {
		sell, sellOK = parse.AsMoney(goods["price"])
	}
	if sellOK {
		res.SellPrice = f64ptr(sell)
	}

	var recycle float64
	var recycleOK bool
	recycle, recycleOK = firstMoneyAny(gp, xiziRecycleKeys...)
	if !recycleOK && goods != nil {
		recycle, recycleOK = parse.AsMoney(goods["price"])
	}
	if !recycleOK {
		return res, fmt.Errorf("hangzhouxizi missing guige.data.price recycle field for game_id=%s", gameID)
	}
	res.RecyclePrice = f64ptr(recycle)
	res.Note = "recycle_price=guige.data.price.hs_price_2"
	return res, nil
}

func xiziGoodsFallback(res Result, detail any, gameID string) Result {
	goods, _ := parse.Get(detail, "data.goods").(map[string]any)
	if goods == nil {
		res.Status = StatusUnavailable
		res.Note = "杭州西子 goods unavailable"
		return res
	}
	if n := stringify(goods["id"]); n != "" {
		res.ItemID = n
	}
	if p, ok := parse.AsMoney(goods["price"]); ok {
		res.SellPrice = f64ptr(p)
		res.RecyclePrice = f64ptr(p)
		res.Note = "recycle_price=detail.data.goods.price fallback"
	}
	return res
}

// firstMoneyAny returns the first numeric value among the keys, descending
// into lists when the value is an array.
func firstMoneyAny(v any, keys ...string) (float64, bool) {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range keys {
			if m, ok := parse.AsMoney(t[k]); ok {
				return m, true
			}
		}
	case []any:
		for _, item := range t {
			if m, ok := firstMoneyAny(item, keys...); ok {
				return m, true
			}
		}
	}
	return 0, false
}

// Catalog pages through 杭州西子's full product list.
func (m *hangzhouxizi) Catalog(ctx context.Context, limit int) ([]CatalogItem, error) {
	const pageSize = 30
	fetchPage := func(page int) ([]CatalogItem, error) {
		body, err := m.client.PostJSON(ctx, xiziListURL, xiziHeaders(), map[string]any{
			"name": "", "keyword": "", "page": page, "pageNum": page, "pageSize": pageSize,
		})
		if err != nil {
			return nil, err
		}
		var payload any
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
		var items []CatalogItem
		for _, row := range firstRows(payload, "data.goods", "data.list", "data.rows", "rows", "data") {
			m, ok := row.(map[string]any)
			if !ok {
				continue
			}
			id := firstString(m, "id", "commodityId", "goodsId")
			name := firstString(m, "name", "commodityName", "goodsName", "title")
			if id == "" || name == "" {
				continue
			}
			items = append(items, CatalogItem{Name: name, GameID: id})
		}
		return items, nil
	}
	return collectCatalog(fetchPage, limit)
}
