# 杭州西子 (hangzhouxizi)

## 接口列表

| 用途 | URL |
|---|---|
| 查询单张卡带价格（详情） | `https://xcx.hzxzdwsc.com/api/index/detail` |
| 查询单张卡带价格（规格回收价） | `https://xcx.hzxzdwsc.com/api/index/guige` |
| 所有卡带名称 / game id 列表 | `https://xcx.hzxzdwsc.com/api/index/goods` |

## 公共请求头

```
content-type: application/json
referer: https://xcx.hzxzdwsc.com
user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 MicroMessenger/7.0.20.1781(0x6700143B) NetType/WIFI MiniProgramEnv/Windows WindowsWechat/WMPF WindowsWechat(0x63090a13) UnifiedPCWindowsWechat(0xf2541939) XWEB/19841
xweb_xhr: 1
```

## 1. 单张卡带价格

需要两个接口配合：
1. `detail` 取 `data.guige[].gg_value[].bianma`（规格编码），优先选 title 含「二手」和「盒」的规格（或 game 数据中的 sku_id）
2. `guige` 用 `{"id": <game_id>, "guige": <bianma>}` 取回收价

关键响应字段: `guige.data.price.hs_price_2`（回收价，候选字段顺序 `hs_price_2` > `hs_price_1` > `hs_price` > `recycle_price` > `recyclePrice` > `receivePrice` > `price`）；卖价取 `guige.data.price.price` 或 `detail.data.goods.price`。若 detail 无规格编码，回退使用 `detail.data.goods.price`

### 示例 curl

```bash
curl -s -X POST 'https://xcx.hzxzdwsc.com/api/index/detail' \
  -H 'content-type: application/json' \
  -H 'referer: https://xcx.hzxzdwsc.com' \
  -H 'xweb_xhr: 1' \
  -d '{"id": "12345"}'

curl -s -X POST 'https://xcx.hzxzdwsc.com/api/index/guige' \
  -H 'content-type: application/json' \
  -H 'referer: https://xcx.hzxzdwsc.com' \
  -H 'xweb_xhr: 1' \
  -d '{"id": "12345", "guige": "6"}'
```

## 2. 卡带列表（game id 全量）

- URL: `https://xcx.hzxzdwsc.com/api/index/goods`
- 方法: POST，`Content-Type: application/json`，每页 30 条
- 关键响应字段: 每条记录的 `id`/`commodityId`/`goodsId`（game id）、`name`/`commodityName`/`goodsName`（卡带名称）

### 示例 curl

```bash
curl -s -X POST 'https://xcx.hzxzdwsc.com/api/index/goods' \
  -H 'content-type: application/json' \
  -H 'referer: https://xcx.hzxzdwsc.com' \
  -H 'xweb_xhr: 1' \
  -d '{"name": "", "keyword": "", "page": 1, "pageNum": 1, "pageSize": 30}'
```
