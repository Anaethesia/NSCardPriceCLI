# 海螺 (hailuo)

## 接口列表

| 用途 | URL |
|---|---|
| 查询单张卡带价格 | `https://hailuo.dwzjd.com/api/product/detail/{game_id}` |
| 所有卡带名称 / game id 列表 | `https://hailuo.dwzjd.com/api/products` |

**无需 token / 登录态**：详情与列表接口都可直接用小程序请求头匿名访问。

## 公共请求头

```
cb-lang: zh-CN
appid: wx7f7b845076caaf81
user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 MicroMessenger/7.0.20.1781(0x6700143B) NetType/WIFI MiniProgramEnv/Windows WindowsWechat/WMPF WindowsWechat(0x63090a13) UnifiedPCWindowsWechat(0xf2541939) XWEB/19841
xweb_xhr: 1
form-type: routine
content-type: application/json
accept: */*
```

## 1. 单张卡带价格（detail）

- URL: `https://hailuo.dwzjd.com/api/product/detail/{game_id}`
- 方法: GET
- 关键响应字段（`data.storeInfo` 内）: `price`（卖价）、`out_diff`（回收扣减）、`out_recycle`（=0 表示不回收）、`cate_id`（1=Switch，41=Switch 2）
- 回收价 = `storeInfo.price - storeInfo.out_diff`
- `status=400533`（商品不存在）视为 unavailable，不重试、不报错

### 示例 curl

```bash
curl -s 'https://hailuo.dwzjd.com/api/product/detail/2060' \
  -H 'cb-lang: zh-CN' \
  -H 'appid: wx7f7b845076caaf81' \
  -H 'xweb_xhr: 1' \
  -H 'form-type: routine' \
  -H 'content-type: application/json' \
  -H 'accept: */*'
```

## 2. 卡带列表（game id 全量）

- URL: `https://hailuo.dwzjd.com/api/products`
- 方法: GET，每页最多 50 条
- 查询参数: `cid=0&sid=0&keyword=&priceOrder=&salesOrder=&news=0&best=0&store_label_id=&page=1&limit=50&coupon_category_id=&productId=`
- 关键响应字段: 每条记录的 `id`（game id）、`store_name`（卡带名称）、`cate_id`（平台分类）

### 示例 curl

```bash
curl -s 'https://hailuo.dwzjd.com/api/products?cid=0&sid=0&keyword=&priceOrder=&salesOrder=&news=0&best=0&store_label_id=&page=1&limit=50&coupon_category_id=&productId=' \
  -H 'cb-lang: zh-CN' \
  -H 'appid: wx7f7b845076caaf81' \
  -H 'xweb_xhr: 1' \
  -H 'form-type: routine' \
  -H 'content-type: application/json' \
  -H 'accept: */*'
```
