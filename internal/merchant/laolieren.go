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
	laolierenDetailURL = "https://api.laolieren.com/v2/game/detail"
	laolierenListURL   = "https://api.laolieren.com/v2/game/home"
)

type laolieren struct {
	client *httpx.Client
}

func newLaolieren(c *httpx.Client) Merchant { return &laolieren{client: c} }

func (m *laolieren) Key() string  { return "laolieren" }
func (m *laolieren) Name() string { return "老猎人" }

func (m *laolieren) Fetch(ctx context.Context, entry game.MerchantEntry) (Result, error) {
	res := newResult(m.Key(), m.Name())
	res.ItemID = entry.GameID

	body, err := m.client.PostJSON(ctx, laolierenDetailURL,
		map[string]string{"content-type": "application/json"},
		map[string]any{"auth": "", "id": entry.GameID},
	)
	if err != nil {
		return res, err
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return res, err
	}

	sellPrice, sellOK := parse.AsMoney(parse.First(payload, "row.price"))

	if stringify(parse.First(payload, "row.is_outside")) == "0" {
		res.Status = StatusUnavailable
		res.Note = "row.is_outside=0"
		if sellOK {
			res.SellPrice = f64ptr(sellPrice)
		}
		return res, nil
	}

	outsideDiff, ok := parse.AsMoney(parse.First(payload, "row.outside_diff"))
	if !ok || !sellOK {
		return res, fmt.Errorf("laolieren missing row.price or row.outside_diff for game_id=%s", entry.GameID)
	}
	res.SellPrice = f64ptr(sellPrice)
	res.RecyclePrice = f64ptr(sellPrice + outsideDiff)
	res.Note = "recycle_price=row.price+row.outside_diff"
	return res, nil
}

// Catalog pages through 老猎人's full product list.
func (m *laolieren) Catalog(ctx context.Context, limit int) ([]CatalogItem, error) {
	fetchPage := func(page int) ([]CatalogItem, error) {
		body, err := m.client.PostJSON(ctx, laolierenListURL,
			map[string]string{"content-type": "application/json"},
			map[string]any{
				"auth":   "",
				"page":   page,
				"app":    "weixin",
				"filter": map[string]string{"platform": "", "keyword": "", "listorder": "", "favorite": "", "genres": "", "preset": ""},
			},
		)
		if err != nil {
			return nil, err
		}
		var payload any
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
		var items []CatalogItem
		for _, row := range firstRows(payload, "rows", "data.rows", "data.list", "data") {
			m, ok := row.(map[string]any)
			if !ok {
				continue
			}
			id, name := stringify(m["id"]), stringify(m["title"])
			if id == "" || name == "" {
				continue
			}
			items = append(items, CatalogItem{Name: name, GameID: id, Platform: stringify(m["platform"])})
		}
		return items, nil
	}
	return collectCatalog(fetchPage, limit)
}
