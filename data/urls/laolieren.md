# 老猎人 (laolieren)

## 接口列表

| 用途 | URL |
|---|---|
| 查询单张卡带价格 | `https://api.laolieren.com/v2/game/detail` |
| 所有卡带名称 / game id 列表 | `https://api.laolieren.com/v2/game/home` |

## 1. 单张卡带价格（detail）

- URL: `https://api.laolieren.com/v2/game/detail`
- 方法: POST，`Content-Type: application/json`
- 请求体: `{"auth": "", "id": "<game_id>"}`
- 关键响应字段: `row.price`（卖价）、`row.outside_diff`（回收加价，回收价 = `row.price + row.outside_diff`）、`row.is_outside`（=0 表示不回收）

### 示例 curl

```bash
curl -s -X POST 'https://api.laolieren.com/v2/game/detail' \
  -H 'content-type: application/json' \
  -d '{"auth": "", "id": "12345"}'
```

## 2. 卡带列表（game id 全量）

- URL: `https://api.laolieren.com/v2/game/home`
- 方法: POST，`Content-Type: application/json`
- 请求体（分页）: `{"auth": "", "page": 1, "app": "weixin", "filter": {"platform": "", "keyword": "", "listorder": "", "favorite": "", "genres": "", "preset": ""}}`
- 关键响应字段: 每条记录的 `id`（game id）、`title`（卡带名称）、`platform`

### 示例 curl

```bash
curl -s -X POST 'https://api.laolieren.com/v2/game/home' \
  -H 'content-type: application/json' \
  -d '{"auth": "", "page": 1, "app": "weixin", "filter": {"platform": "", "keyword": "", "listorder": "", "favorite": "", "genres": "", "preset": ""}}'
```
