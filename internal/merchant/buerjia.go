package merchant

import (
	"context"
	"encoding/json"
	"fmt"

	"nscardprice/internal/game"
	"nscardprice/internal/httpx"
	"nscardprice/internal/parse"
)

const (
	buerjiaDetailURL = "https://bejdw.crystal-cloud.top/api/games/detail"
	buerjiaListURL   = "https://bejdw.crystal-cloud.top/api/games/list"
)

type buerjia struct {
	client *httpx.Client
}

func newBuerjia(c *httpx.Client) Merchant { return &buerjia{client: c} }

func (m *buerjia) Key() string  { return "buerjia" }
func (m *buerjia) Name() string { return "不二家" }

func buerjiaHeaders() map[string]string {
	return map[string]string{
		"content-type": "application/json",
		"referer":      "https://bejdw.crystal-cloud.top/api/",
		"user-agent":   miniUserAgent,
		"xweb_xhr":     "1",
	}
}

func (m *buerjia) Fetch(ctx context.Context, entry game.MerchantEntry) (Result, error) {
	res := newResult(m.Key(), m.Name())
	res.ItemID = entry.GameID

	body, err := m.client.PostJSON(ctx, buerjiaDetailURL, buerjiaHeaders(),
		map[string]any{"id": entry.GameID})
	if err != nil {
		return res, err
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return res, err
	}

	data, _ := payload.(map[string]any)["data"].(map[string]any)
	if data == nil {
		return res, fmt.Errorf("buerjia response missing data object for game_id=%s", entry.GameID)
	}

	if id := stringify(data["id"]); id != "" {
		res.ItemID = id
	}
	nobox, noboxOK := parse.AsMoney(data["nobox"])
	if noboxOK {
		res.SellPrice = f64ptr(nobox)
	}

	box, boxOK := parse.AsMoney(data["box"])
	if !boxOK || box == 0 {
		res.Status = StatusUnavailable
		res.Note = "buerjia missing data.box recycle price"
		return res, nil
	}

	recycle := box
	note := "recycle_price=data.box"
	if notaobao, ok := parse.AsMoney(data["notaobao"]); ok {
		recycle += notaobao
		note = "recycle_price=data.box+data.notaobao"
	}
	res.RecyclePrice = f64ptr(recycle)
	res.Note = note
	return res, nil
}

// Catalog pages through 不二家's full product list.
func (m *buerjia) Catalog(ctx context.Context, limit int) ([]CatalogItem, error) {
	const pageSize = 30
	fetchPage := func(page int) ([]CatalogItem, error) {
		body, err := m.client.PostJSON(ctx, buerjiaListURL, buerjiaHeaders(), map[string]any{
			"search": "", "page": page, "pageNum": page, "pageSize": pageSize, "limit": pageSize,
		})
		if err != nil {
			return nil, err
		}
		var payload any
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
		var items []CatalogItem
		for _, row := range firstRows(payload, "data.data", "data.list", "data.rows", "list", "rows", "data") {
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
