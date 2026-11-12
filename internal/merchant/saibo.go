package merchant

import (
	"context"

	"nscardprice/internal/game"
	"nscardprice/internal/httpx"
)

// saiboShopURL is 赛博电玩's Taobao storefront. The store is a Taobao shop
// (shop id 255640440, based in Jiaxing) and does not expose a public
// recycle-price API like the other six merchants, so Fetch reports
// unavailable and the URL is kept for reference.
const saiboShopURL = "https://shop255640440.taobao.com"

type saibo struct {
	client *httpx.Client
}

func newSaibo(c *httpx.Client) Merchant { return &saibo{client: c} }

func (m *saibo) Key() string  { return "saibo" }
func (m *saibo) Name() string { return "赛博电玩" }

func (m *saibo) Fetch(ctx context.Context, entry game.MerchantEntry) (Result, error) {
	res := newResult(m.Key(), m.Name())
	res.ItemID = entry.GameID
	res.Status = StatusUnavailable
	res.Note = "赛博电玩 为淘宝店铺，无公开回收价 API；店铺 " + saiboShopURL
	return res, nil
}
