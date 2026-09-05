package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nscardprice/internal/game"
	"nscardprice/internal/httpx"
	"nscardprice/internal/parse"
)

const (
	huoqiangshouDetailURL  = "https://api.huoqiangshou.cn/seller/sellerProduct/getProductDetail"
	huoqiangshouApprizeURL = "https://api.huoqiangshou.cn/seller/sellerProduct/getProductApprizeQuestion"
	huoqiangshouListURL    = "https://api.huoqiangshou.cn/seller/category/getProductInfoPage"
)

type huoqiangshou struct {
	client *httpx.Client
}

func newHuoqiangshou(c *httpx.Client) Merchant { return &huoqiangshou{client: c} }

func (m *huoqiangshou) Key() string  { return "huoqiangshou" }
func (m *huoqiangshou) Name() string { return "火枪手" }

func huoqiangshouHeaders() map[string]string {
	return map[string]string{
		"terminal":   "WECHAT",
		"user-agent": miniUserAgent,
		"xweb_xhr":   "1",
		"referer":    "https://servicewechat.com/wx0f883cb942dd9691/630/page-frame.html",
		"accept":     "*/*",
	}
}

func (m *huoqiangshou) Fetch(ctx context.Context, entry game.MerchantEntry) (Result, error) {
	res := newResult(m.Key(), m.Name())
	res.ItemID = entry.GameID

	detailBody, err := m.client.PostForm(ctx, huoqiangshouDetailURL, huoqiangshouHeaders(),
		map[string][]string{"productId": {entry.GameID}, "lottoProduct": {"false"}})
	if err != nil {
		return res, err
	}
	apprizeBody, err := m.client.PostForm(ctx, huoqiangshouApprizeURL, huoqiangshouHeaders(),
		map[string][]string{"productId": {entry.GameID}})
	if err != nil {
		return res, err
	}

	var detail, apprize any
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		return res, err
	}
	if err := json.Unmarshal(apprizeBody, &apprize); err != nil {
		return res, err
	}

	receive, receiveOK := parse.AsMoney(parse.First(detail, "data.receivePrice"))
	if !receiveOK {
		return res, fmt.Errorf("火枪手 missing detail.data.receivePrice for game_id=%s", entry.GameID)
	}
	if sell, ok := parse.AsMoney(parse.First(detail, "data.retailPrice", "data.sellPrice", "data.salePrice", "data.price")); ok {
		res.SellPrice = f64ptr(sell)
	}

	reduce, reduceOK := otherShopReduce(apprize)
	if !reduceOK {
		res.Status = StatusUnavailable
		res.Note = "火枪手 no other-shop recycle option"
		return res, nil
	}
	res.RecyclePrice = f64ptr(receive - reduce)
	res.Note = "recycle_price=detail.data.receivePrice-apprize.other_shop.reducePrice"
	return res, nil
}

// otherShopReduce finds the "别店/别家/其他" question's reducePrice in the
// apprize response.
func otherShopReduce(apprize any) (float64, bool) {
	questions, _ := parse.Get(apprize, "data.listProductQuestion").([]any)
	for _, q := range questions {
		qm, ok := q.(map[string]any)
		if !ok {
			continue
		}
		answers, _ := qm["answers"].([]any)
		for _, a := range answers {
			am, ok := a.(map[string]any)
			if !ok {
				continue
			}
			text := strings.ToLower(stringify(am["name"]) + stringify(am["valueType"]))
			if strings.Contains(text, "别店") || strings.Contains(text, "别家") ||
				strings.Contains(text, "其他") || strings.Contains(text, "other_shop") {
				if v, ok := parse.AsMoney(am["reducePrice"]); ok {
					return v, true
				}
			}
		}
	}
	return 0, false
}

// Catalog pages through 火枪手's full product list.
func (m *huoqiangshou) Catalog(ctx context.Context, limit int) ([]CatalogItem, error) {
	const pageSize = 30
	fetchPage := func(page int) ([]CatalogItem, error) {
		// 火枪手 throttles rapid pagination with empty 2xx bodies; pacing the
		// requests avoids the point where pages start coming back empty.
		select {
		case <-time.After(1000 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		form := map[string][]string{
			"pageNumber":     {stringify(page)},
			"pageNum":        {stringify(page)},
			"pageSize":       {stringify(pageSize)},
			"brandId":        {""},
			"productType":    {"CARD"},
			"productName":    {""},
			"monthSale":      {""},
			"stockNum":       {""},
			"supportChinese": {""},
			"screenPrice":    {""},
			"timeSort":       {""},
			"startPrice":     {"0"},
			"endPrice":       {"999"},
			"linkStatus":     {""},
			"gameType":       {""},
		}
		body, err := m.client.PostForm(ctx, huoqiangshouListURL, huoqiangshouHeaders(), form)
		if err != nil {
			return nil, err
		}
		var payload any
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
		var items []CatalogItem
		for _, row := range firstRows(payload, "data.rows", "rows", "data.list", "data") {
			m, ok := row.(map[string]any)
			if !ok {
				continue
			}
			id := firstString(m, "id", "productId")
			name := firstString(m, "productName", "name", "title")
			if id == "" || name == "" {
				continue
			}
			// 火枪手列表接口没有平台字段（只有 brandId/categoryName 等，
			// 不是平台），NS/NS2 过滤交给 IsNS 按名称正则处理。
			items = append(items, CatalogItem{Name: name, GameID: id})
		}
		return items, nil
	}
	return collectCatalog(fetchPage, limit)
}
