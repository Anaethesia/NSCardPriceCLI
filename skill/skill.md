---
name: "nscardprice-cli"
description: "使用 nscardprice CLI 查询任天堂 Switch 卡带回收价/售价。Invoke when user asks for game cart prices, 卡带价格, or wants to run nscardprice commands."
---

# 功能说明

本 skill 教 agent 如何调用 `nscardprice` CLI 查询任天堂 Switch 卡带价格。程序读取 `data/games.json` 中的游戏与商家商品 ID，并发查询**四家**回收商：**laolieren（老猎人）、huoqiangshou（火枪手）、buerjia（不二家）、hangzhouxizi（杭州西子）**。

## 核心规则（必须遵守）

1. **游戏匹配**：`query` 参数按中文名/关键词模糊匹配（大小写不敏感），常用简称与别名自动识别；精确 slug 匹配暂未启用。
2. **批量查询优先 `--mul`**：一次要查多个游戏时，优先用 `nscardprice query --mul 词1 词2 词3`（最多 5 个）一次查完；若 `--mul` 退出码非 0（如关键词未命中），再逐个 `query` 分开查询。 
3. **商家指定**：`-m/--merchant` 可多次使用，接受 key 或中文名（如 `-m buerjia` 或 `-m 不二家`）；不传 = 查询全部 4 家。
4. **输出格式**：所有命令的正式输出是 JSON，写到 **stdout**；日志与错误信息写到 **stderr**。解析数据只看 stdout。
5. **退出码**：`0` = 成功（含 unavailable/skipped 状态）；`1` = 有商家返回 error；`2` = 用法错误/未找到。判断失败必须以退出码为准，`非 0 即失败`。


## 命令一览

| 命令 | 功能 |
|------|------|
| `nscardprice query [关键词]` | 查一个游戏的回收价（默认 4 家全查） |
| `nscardprice query --slug <slug>` | 按 slug 精确查询（不区分大小写），返回唯一游戏；与 `--all/--mul` 互斥 |
| `nscardprice query --mul 词1 词2 词3` | 查多个游戏的回收价（空格分隔，最多 5 个，结果按关键词顺序去重合并） |
| `nscardprice query -c/--custom 关键词` | 只在 `data/custom.json` 自定义库中匹配查价 |
| `nscardprice query --custom --all` | 查询自定义库全部条目 |
| `nscardprice query --all` | 全量查询 games.json 中所有启用游戏 |
| `nscardprice list <merchant> [count]` | 拉取某商家卡带目录（name + game_id），默认 10 条 |
| `nscardprice list <merchant> --all` | 拉取某商家全量目录 |
| `nscardprice sync` | 同步 4 家商家全量目录并过滤出 NS/NS2 后覆盖到 `data/lists/`（可用 `-m` 指定商家） |
| `nscardprice merchants` | 列出支持的商家 key 和中文名 |
| `nscardprice --version` | 输出版本号 |
| `nscardprice --help` | 帮助 |

常用 flag：`--games <path>`（默认 `data/games.json`）、`--custom-file <path>`（默认 `data/custom.json`）、`--log-dir <dir>`（默认 `logs/`，空字符串=不写日志文件）、`--timeout`（默认 10s）、`--retries`（默认 2）、`--concurrency`（默认 8）、`--dry-run`（不发真实请求，只校验）、`--insecure`、`--verbose`。运行与错误日志追加写到 `logs/nscardprice.log`（stderr 另会打印错误摘要，日志文件作为留档排查）。所有真实请求经 httpx 全局 **0.2s 间隔限速**（同一 client 串行排队），避免查询多家/多商品过快触发商家风控。火枪手目录翻页另有 1s/页间隔；集成测试仍按 3s/请求。

## 参考数据

`references/games-catalog.json` 收录全部游戏目录（每条含 `slug`/`name`/`platform`，可选 `search_keyword`，无商家 ID）。用途：

- 用户口述的卡带名需要映射到库内条目时，按 `name`/`search_keyword` 查找，避免模糊匹配命中多个版本（如基础版 vs DLC同捆 vs NS2 版）。
- 用户用非标准叫法/关键词过短没有命中时， 你可以在关键词基础上补充信息并在`references/games-catalog.json`关联对应游戏的 `slug`，再 `nscardprice query --slug <slug>` 精确查询。

## custom.json 自定义库

`data/custom.json` 是独立于主库的自定义查询库（结构同 games.json，`enabled` 缺省视为 true），供 `query -c/--custom` 专用，适合自维护常查卡带的精确 game_id/sku_id：

## 结果 JSON 结构（query）

单游戏返回对象，多游戏返回数组：

```json
{
  "fetched_at": "2026-09-05T10:00:00+08:00",
  "slug": "zelda-tears-of-the-kingdom",
  "name": "塞尔达传说 王国之泪",
  "platform": "Nintendo Switch",
  "results": [
    {
      "merchant": "laolieren",
      "merchant_name": "老猎人",
      "status": "ok",
      "recycle_price": 170,
      "sell_price": 200,
      "currency": "CNY",
      "item_id": "3282",
      "note": "recycle_price=row.price+row.outside_diff"
    }
  ]
}
```

`status` 取值：`ok`（有回收价）、`unavailable`（商家当前不收）、`skipped`（该游戏此商家无 game_id）、`error`（请求/解析失败）。

## Few-shot 示例

### 1. 关键词查价（全部商家）

```
$ nscardprice query 塞尔达旷野
```
→ 输出 1~2 个游戏（旷野之息 + 可能含同捆），每个含 4 家回收价。**注意**：`塞尔达` 会匹配多个游戏；若要唯一结果，用更精确的中文名（如 `双人成行`）。

### 2. 只查自定义库

```
$ nscardprice query -c 马车8
```
→ 只在 `data/custom.json` 内匹配，适合查自维护的常用卡带；自定义库为空或未命中时退出码 2。

### 3. 只查一家回收商价格（火枪手）

```
$ nscardprice query -m huoqiangshou 王国之泪
```
→ 单游戏 JSON，`results` 只有 1 项，`merchant` 为 `huoqiangshou`。

### 4. 查多张卡带价格

```
$ nscardprice query --mul 旷野 马8 双人成行
```
→ 返回 3 个游戏的回收价，每个含 4 家。**注意**：`query --mul` 会按关键词顺序去重合并，若关键词重复，结果会包含重复项。

### 5. 一个关键词对应多款拆分的场景

**塞尔达旷野之息有 3 个条目**：基础版、DLC同捆、NS2 版（实测确认）：

```
$ nscardprice query 旷野之息
# -> 3 条全部命中（基础版 / DLC同捆 / NS2版）

$ nscardprice query 旷野之息同捆
# -> 命中 DLC同捆 + NS2版（NS2版名字含"旷野之息"且无拒绝词），
#    仍不是单一结果

$ nscardprice query zelda-breath-of-the-wild
# -> 仍按关键词处理，3 条全部命中

$ nscardprice query --slug zelda-breath-of-the-wild
# -> 精确命中 1 条（不区分大小写），适合需唯一结果的场景
```

**教训**：中文关键词常返回多个条目。当用户问"旷野之息多少钱"而未说明版本时，应**展示所有命中条目并让用户确认**；需要唯一结果就用 `query --slug <slug>` 或精确中文名，或把这些卡带维护进 `custom.json` 再 `query -c` 查。

### 6. 全量查询

```
$ nscardprice query --all
```
→ 返回数组，长度 = games.json 启用游戏数。适合做价格快照/巡检。

### 7. list / sync 拉目录（补数据源）

```
$ nscardprice list laolieren 10
$ nscardprice list buerjia --all --out data/lists
$ nscardprice sync         # 一条命令同步 4 家，仅保留 NS/NS2，覆盖 data/lists
```
→ stdout 输出 `[{name, game_id}, ...]` JSON；同时写入 `data/lists/<merchant>.json`（默认目录），stderr 提示落盘路径。拉回的新 `game_id` 可用于补充 `data/games.json`。

### 8. 错误处理（必须检查退出码）

```
$ nscardprice query 不存在游戏xyz
# 退出码 2，stderr: 未找到匹配的游戏: "不存在游戏xyz"

$ nscardprice list 火星卖家
# 退出码 2，stderr: 未知商家: 火星卖家

$ nscardprice query
# 退出码 2，stderr: 请提供游戏名/关键词，或使用 --all
```

### 9. agent 实际操作模式（推荐）

```bash
# 查价后解析 JSON：
nscardprice query 双人成行 | python3 -c \
  "import json,sys; d=json.load(sys.stdin); \
   [print(r['merchant_name'], r.get('recycle_price')) for r in d['results']]"

# 判断失败：用退出码而不是 stderr 文本：
nscardprice query xxx >/dev/null 2>&1 || echo "查询失败"
```

## 模糊匹配规则简述

- **模糊匹配为主**：slug 字符串按普通关键词处理（精确 slug 匹配暂未启用），中文简称/别名由规则表覆盖。
- **别名映射**：马力欧→马里奥、野炊/荒野之息→旷野之息、斯普拉遁→斯普拉顿 等自动归一。
- **分组规则**：`空前盛会`+`马里奥` 这类多关键词须同时命中；`同捆/dlc/扩充票` 只匹配 DLC同捆条目。
- **拒绝规则**：查基础版时若带 `dlc`/`同捆` 字样会排除基础版，防止混入。

总之：**阅读 stdout 的结构化 JSON，以退出码判断成败，商家结果逐个读取，忽略 stderr 内容（仅日志）**。