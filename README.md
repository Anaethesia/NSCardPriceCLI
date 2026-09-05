# nscardprice

Go 编写的 Nintendo Switch 卡带回收价查询 CLI。读取 `data/games.json` 中的游戏与商家商品 ID，查询**老猎人、火枪手、不二家、杭州西子**四家回收商的回收价/售价，以 JSON 输出，方便终端使用或供 agent/脚本消费。

## 功能

- 查询任天堂 Switch 卡带回收价/售价，覆盖四家回收商
- 按游戏中文名/关键词模糊匹配，常用简称与别名自动识别
- 多关键词同时查询（最多 5 个），一次拿到多个游戏的价格
- 支持独立的自定义游戏库，与内置清单互不影响
- 一条命令同步四家商家的 NS/NS2 卡带目录，自动过滤其他平台条目
- 差异化解析各家接口：回收价、售价、暂不收购（unavailable）状态
- 并发查询，单家失败不影响其余商家
- 内置 0.2s 请求间隔制动，防止高频查询触发商家风控
- 纯 JSON 输出到 stdout，日志与错误走 stderr
- 退出码语义清晰：0 成功，1 有商家错误，2 用法错误

## 命令

```bash
nscardprice query 塞尔达              # 关键词查询全部 4 家
nscardprice query --mul 旷野 马8 双人成行  # 多关键词同时查询（最多 5 个）
nscardprice query --custom 马车8             # 仅在 data/custom.json 自定义库中查询
nscardprice query --custom --all             # 查询自定义库全部条目
nscardprice --merchant buerjia query 旷野    # 仅查指定商家（可重复）
nscardprice query --all                      # 查询全部游戏
nscardprice list laolieren 10                # 列出老猎人卡带目录前 10 条 (name+game_id)
nscardprice list buerjia --all --out data/lists   # 拉全量并落盘到 data/lists/
nscardprice sync                             # 同步 4 家 NS/NS2 目录到 data/lists
nscardprice sync -m huoqiangshou             # 只同步指定商家
nscardprice merchants                        # 列出 4 家商家 key
nscardprice --version                        # 版本号
```

常用全局 flag：`--games <path>`（默认 `data/games.json`）、`--custom-file <path>`（默认 `data/custom.json`）、`--timeout`（默认 10s）、`--retries`（默认 2）、`--concurrency`（默认 8）、`--insecure`、`--dry-run`、`--verbose`。

### data/custom.json 自定义库

独立于 `data/games.json`，适合放常查但不想混入主库的卡带（结构一致，`enabled` 缺省视为 true）。用 `query --custom` 只在此库内匹配查价：

```json
[
  {
    "slug": "mk8-custom",
    "name": "马车8 定制",
    "platform": "Nintendo Switch",
    "enabled": true,
    "merchant_ids": {
      "laolieren": {"game_id": "153"},
      "buerjia": {"game_id": "5"},
      "hangzhouxizi": {"game_id": "107", "sku_id": "148"}
    }
  }
]
```

示例输出（单游戏）：

```json
{"fetched_at":"...","slug":"...","name":"塞尔达传说 王国之泪",
 "results":[{"merchant":"laolieren","status":"ok","recycle_price":170,"sell_price":200,"currency":"CNY",...}]}
```

## 架构

```
cmd/nscardprice/    入口
internal/app/       依赖装配
internal/config/    运行选项与默认值
internal/game/      领域模型（零依赖）
internal/gamesrepo/ 游戏清单仓库（JSON 实现）
internal/merchant/  商家抽象 + 4 家实现
internal/httpx/     带重试/退避/TLS 的 HTTP 封装
internal/parse/     JSON 解析工具
internal/collector/ 并发编排与结果聚合
internal/cli/       cobra 命令
test/               真实接口集成测试（build tag 隔离，默认不跑）
data/games.json     游戏与商家商品 ID 清单
data/custom.json    自定义查询库（query --custom 使用）
data/lists/         各商家全量目录（list 全量 / sync 仅保留 NS/NS2 落盘）
logs/               错误/运行日志（nscardprice.log，追加写）
```

扩展方式：新增商家在 `internal/merchant` 加实现并注册；换数据源实现 `gamesrepo.Repository` 接口。

## 构建

安装到 PATH（推荐，之后直接用 `nscardprice`）：

```bash
go install ./cmd/nscardprice     # 装入 ~/go/bin（已在 PATH 则直接生效）
nscardprice --help
```

或仅生成本地二进制并直接运行：

```bash
make build                       # 生成 bin/nscardprice
./bin/nscardprice list
```

## 真实接口测试

```bash
NSCARD_INTEGRATION=1 go test -tags integration ./test/ -v
```

测试对每请求间隔 3s，防止触发第三方接口限流；平时 `go test ./...` 不受影响。