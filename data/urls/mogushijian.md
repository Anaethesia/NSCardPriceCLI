# 蘑菇时间 (mogushijian)

## 接口列表

| 用途 | URL |
|---|---|
| 卡带回收价目录（兼作单卡查询） | `https://api.zhualegezhua.cn/alia/used/recycle/cardsList` |

**无需 token / 登录态**：目录接口只要求 `alianame: alia2` 请求头，匿名可访问。该接口一次返回整份回收目录（每条含 `cardsId` 与 `price`），没有单独的商品详情接口，因此单卡查询时按 `cardsId` 在目录中查找，回收价直接取 `price`。

## 公共请求头

```
content-type: application/json
alianame: alia2
accept: */*
```

## 1. 卡带回收目录 / 单卡价格

- URL: `https://api.zhualegezhua.cn/alia/used/recycle/cardsList`
- 方法: POST，`Content-Type: application/json`
- 请求体: `{"page": 1}`，每页 20 条，响应 `hasMore` / `next` 指示下一页
- 关键响应字段: `cardsId`（game id）、`name`（卡带名称）、`price`（回收价）、`generationType`（1=Switch，2=Switch 2）

### 示例 curl

```bash
curl -s -X POST 'https://api.zhualegezhua.cn/alia/used/recycle/cardsList' \
  -H 'content-type: application/json' \
  -H 'alianame: alia2' \
  -H 'accept: */*' \
  -d '{"page": 1}'
```
