# 火枪手 (huoqiangshou)

## 接口列表

| 用途 | URL |
|---|---|
| 查询单张卡带价格（详情） | `https://api.xiamenhuoqiangshou.com/seller/sellerProduct/getProductDetail` |
| 查询单张卡带价格（回收问卷扣减） | `https://api.xiamenhuoqiangshou.com/seller/sellerProduct/getProductApprizeQuestion` |
| 所有卡带名称 / game id 列表 | `https://api.xiamenhuoqiangshou.com/seller/category/getProductInfoPage` |

## 公共请求头

```
terminal: WECHAT
user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 MicroMessenger/7.0.20.1781(0x6700143B) NetType/WIFI MiniProgramEnv/Windows WindowsWechat/WMPF WindowsWechat(0x63090a13) UnifiedPCWindowsWechat(0xf2541939) XWEB/19841
xweb_xhr: 1
referer: https://servicewechat.com/wx0f883cb942dd9691/630/page-frame.html
accept: */*
```

## 1. 单张卡带价格

需要两个接口配合：
1. `getProductDetail` 取 `data.receivePrice`（回收基准价）与 `data.retailPrice` 等卖价字段
2. `getProductApprizeQuestion` 取 `data.listProductQuestion` 中「别店/别家/其他」答案的 `reducePrice`（扣减价）

回收价 = `receivePrice - other_shop.reducePrice`

### 示例 curl

```bash
curl -s -X POST 'https://api.xiamenhuoqiangshou.com/seller/sellerProduct/getProductDetail' \
  -H 'terminal: WECHAT' \
  -H 'xweb_xhr: 1' \
  -H 'referer: https://servicewechat.com/wx0f883cb942dd9691/630/page-frame.html' \
  -H 'accept: */*' \
  -d 'productId=12345&lottoProduct=false'

curl -s -X POST 'https://api.xiamenhuoqiangshou.com/seller/sellerProduct/getProductApprizeQuestion' \
  -H 'terminal: WECHAT' \
  -H 'xweb_xhr: 1' \
  -H 'referer: https://servicewechat.com/wx0f883cb942dd9691/630/page-frame.html' \
  -H 'accept: */*' \
  -d 'productId=12345'
```

## 2. 卡带列表（game id 全量）

- URL: `https://api.xiamenhuoqiangshou.com/seller/category/getProductInfoPage`
- 方法: POST（form 表单），每页 30 条，`productType=CARD`
- 关键响应字段: 每条记录的 `id`/`productId`（game id）、`productName`/`name`（卡带名称）
- 注意: 翻页过快会返回 200 空 body（限流），分页请求需保持约 250ms~1s 间隔

### 示例 curl

```bash
curl -s -X POST 'https://api.xiamenhuoqiangshou.com/seller/category/getProductInfoPage' \
  -H 'terminal: WECHAT' \
  -H 'xweb_xhr: 1' \
  -H 'referer: https://servicewechat.com/wx0f883cb942dd9691/630/page-frame.html' \
  -H 'accept: */*' \
  -d 'pageNumber=1&pageNum=1&pageSize=30&brandId=&productType=CARD&productName=&monthSale=&stockNum=&supportChinese=&screenPrice=&timeSort=&startPrice=0&endPrice=999&linkStatus=&gameType='
```
