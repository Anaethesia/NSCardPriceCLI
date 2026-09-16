# 不二家 (buerjia)

## 接口列表

| 用途 | URL |
|---|---|
| 查询单张卡带价格 | `https://bejdw.crystal-cloud.top/api/games/detail` |
| 所有卡带名称 / game id 列表 | `https://bejdw.crystal-cloud.top/api/games/list` |

## 公共请求头

```
content-type: application/json
referer: https://bejdw.crystal-cloud.top/api/
user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 MicroMessenger/7.0.20.1781(0x6700143B) NetType/WIFI MiniProgramEnv/Windows WindowsWechat/WMPF WindowsWechat(0x63090a13) UnifiedPCWindowsWechat(0xf2541939) XWEB/19841
xweb_xhr: 1
```

## 1. 单张卡带价格（detail）

- URL: `https://bejdw.crystal-cloud.top/api/games/detail`
- 方法: POST，`Content-Type: application/json`
- 请求体: `{"id": "<game_id>"}`
- 关键响应字段（`data` 对象内）: `box`（盒装回收价）、`notaobao`（无淘宝凭证扣减）、`nobox`（卖价）
- 回收价 = `data.box`（若存在 `data.notaobao` 则为 `data.box + data.notaobao`）；`data.box` 为空/0 表示不回收

### 示例 curl

```bash
curl -s -X POST 'https://bejdw.crystal-cloud.top/api/games/detail' \
  -H 'content-type: application/json' \
  -H 'referer: https://bejdw.crystal-cloud.top/api/' \
  -H 'xweb_xhr: 1' \
  -d '{"id": "12345"}'
```

## 2. 卡带列表（game id 全量）

- URL: `https://bejdw.crystal-cloud.top/api/games/list`
- 方法: POST，`Content-Type: application/json`，每页 30 条
- 关键响应字段: 每条记录的 `id`/`commodityId`/`goodsId`（game id）、`name`/`commodityName`/`goodsName`（卡带名称）

### 示例 curl

```bash
curl -s -X POST 'https://bejdw.crystal-cloud.top/api/games/list' \
  -H 'content-type: application/json' \
  -H 'referer: https://bejdw.crystal-cloud.top/api/' \
  -H 'xweb_xhr: 1' \
  -d '{"search": "", "page": 1, "pageNum": 1, "pageSize": 30, "limit": 30}'
```
