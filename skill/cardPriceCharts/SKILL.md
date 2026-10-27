---
name: card-price-charts
description: 把 data/custom.json 的卡带生成「卡带 × 6 家回收商」回收价对比表 PNG，并按行橘色高亮最高价。当用户要卡带价格对比图、回收价一览表、价格矩阵图片或问哪家回收最高时使用。不用于单款卡带查价。
---

# 卡带回收价对比图

把 `data/custom.json` 里多款 Switch 卡带在 6 家回收商的价格，做成一张表格 PNG：每行一款卡带，每列一家回收商，当行最高回收价整格橘底高亮。

**优先自己按本文风格规范生成图**；需要稳定复现或对结果没把握时，用 `scripts/render_chart.py` 兜底。两条路径输出必须一致。

## 输出约定

- 只交付 **PNG**。HTML 只是渲染中间产物，放在临时目录，不要当交付物。
- 默认路径：`<repo>/reports/card-price-chart.png`（`reports/` 已 gitignore）。
- 图片文件名用英文小写短横线，例如 `card-price-chart.png`、`card-price-chart-20.png`。

## 数据契约

- 数据源：`data/custom.json`，按**文件顺序**取前 N 条（默认 20，当前上限 30）。
- 行顺序 = custom.json 顺序，**不要按价格重排**。
- 列顺序固定，不能改：

| 列 | 商家 key | 表头 |
|---|---|---|
| 1 | laolieren | 老猎人 |
| 2 | hailuo | 海螺 |
| 3 | buerjia | 不二家 |
| 4 | hangzhouxizi | 杭州西子 |
| 5 | huoqiangshou | 火枪手 |
| 6 | mogushijian | 蘑菇时间 |

- 价格用 `recycle_price`（回收价），**不要**用 `sell_price`（售价）。
- 总共 7 列：卡带名称 + 6 家。不要加排名、均价、涨幅等额外列。

## 取数

先截取前 N 条到临时文件，再查，避免跑全量 278 条：

```bash
python3 - <<'PY'
import json
src = "data/custom.json"
dst = "/tmp/card-price-subset.json"
games = [g for g in json.load(open(src, encoding="utf-8")) if g.get("enabled", True)]
json.dump(games[:20], open(dst, "w", encoding="utf-8"), ensure_ascii=False)
PY

nscardprice query --custom --all \
  --custom-file /tmp/card-price-subset.json --log-dir ""
```

- 20 条 × 6 家约 30 秒，属正常；不要为此加缓存。
- 部分商家失败是常态：退出码 `0` 即使夹带 `error` 也继续渲染，逐条读 `status` 即可。
- 只有退出码 `2`（用法错误/未找到）、退出码 `1`（所有商家均失败）或 stdout 不是合法 JSON 才终止。
- 不要因为某家 `error` / `unavailable` / `skipped` 就退出，也不要重试整张表。

## 单元格语义（必须准确）

| status | 显示 | 参与最高价 |
|---|---|---|
| `ok` + 有价格 | 价格数字，如 `170` | 是 |
| `ok` + 价格为 null | `暂无` | 否 |
| `unavailable` | `不收` | 否 |
| `skipped` | `未收录` | 否 |
| `error` | `错误` | 否 |
| 整条游戏缺失 | `无数据` | 否 |

**失败绝不要写成 0**：0 会被读成真实回收价，而且会污染最高价高亮。

## 表格图风格规范（必须遵守）

画布与配色：

- 宽度 **1680px**，高度按内容计算；Chrome `--force-device-scale-factor=2` 导出 → PNG 宽 **3360px**。
- 页面底色 `#F7F5F0`（暖白），表格卡片 `#FFFFFF`，主文字 `#1F2328`，次要文字 `#8A8578`。
- 强调色 **`#E8730C`**，只用于最高价单元格和标题装饰条，不要大面积铺。
- 字体：`-apple-system, "PingFang SC", "Hiragino Sans GB", "Helvetica Neue", Arial, sans-serif`；价格开 `font-variant-numeric: tabular-nums`。

版式：

- 标题 `Switch 卡带回收价对比`，34px / 700，左侧一根 6×26px 橘色圆角装饰条。
- 副标题 14px，灰：**只放抓取时间** `<YYYY-MM-DD HH:MM>`（取 JSON 的 `fetched_at`），不要款数、商家数、单位等文案。
- 标题块与卡片间距 26px；卡片圆角 16px，浅投影 `0 14px 34px rgba(31,35,40,.07)`。
- 表头：高 46px，深色底 `#1F2328`，文字 `#F7F5F0`，13px / 600，字间距 `.08em`；首列左对齐，其余居中。
- 数据行：高 54px；名称列固定 500px，6 家价格列均分剩余宽度（1680 画布下每列约 182px）。
- 行分隔用 1px `#EFEBE2` 细线（用 inset box-shadow 画，不要影响行高）；隔行 `#FBFAF6` 淡底。
- 名称列：16px / 600，左对齐，`padding-left:24px`；NS2 条目在名称后加一个小徽标 `NS2`（10px、底色 `#F0EDE5`、文字色 `#8A8071`），**不要**为平台单开一列。
- 价格列：17px / 600，居中，纯数字不带 `¥`。
- 最高价：整格橘底 `#E8730C` + 白字 + 700，18px，去掉该格分隔线；**并列最高全部高亮**；该行没有任何正价格则不高亮。
- 缺数据格：14px / 400，`#B3ADA0`；`错误` 用 `#B4513C`。
- 图下方留 40px 空白收尾，**不要页脚图例、说明文字或单位标注**。

克制项：不要渐变、不要第二个强调色、不要 emoji、不要商家 logo、不要 3D/发光效果、不要导出 CSV/HTML。

## 路径 A：自己生成（首选）

1. 按「取数」拿到 JSON。
2. 按上面的契约与风格拼 HTML：`body{width:1680px; margin:0; box-sizing:border-box}`，`table{border-collapse:separate; border-spacing:0; table-layout:fixed}`。
3. 高度先算好再截图，行高/表头高都是固定像素：
   `H = 44 + 46 + 10 + 24 + 26 + 46 + N*54 + 26 + 40`
4. Chrome headless 截图（HTML 放 `/tmp`）：

```bash
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --headless=new --disable-gpu --hide-scrollbars --no-first-run \
  --no-default-browser-check --disable-extensions \
  --force-device-scale-factor=2 --window-size=1680,<H> \
  --virtual-time-budget=2000 --screenshot=<out.png> "file:///tmp/card-price-chart.html"
```

5. 用 Read 打开 PNG，按下面清单逐项校验；不合格就改 HTML 重截，不要嘴上说通过。

## 路径 B：脚本兜底

```bash
python3 skill/cardPriceCharts/scripts/render_chart.py --limit 20
# 指定输出
python3 skill/cardPriceCharts/scripts/render_chart.py --limit 20 --out reports/card-price-chart-20.png
# 复用已抓好的 JSON，不再请求商家
python3 skill/cardPriceCharts/scripts/render_chart.py --input-json /tmp/query.json
```

参数：`--limit`（默认 20）、`--custom-file`、`--cli`、`--input-json`、`--out`、`--html-out`、`--chrome`、`--timeout`。
脚本只用 Python 3 标准库 + 本机 Chrome，不依赖 PIL / matplotlib / cairosvg；stdout 输出一行 JSON 摘要（行数、高亮数、各状态计数、图片路径）。

## 校验清单

- [ ] PNG 宽度 3360px（2x），高度足够，最后一行完整、底部无截断、无多余说明文字。
- [ ] 行数 = N，列数 = 7；行顺序与 custom.json 一致。
- [ ] 表头 6 家名称和顺序正确（老猎人 / 海螺 / 不二家 / 杭州西子 / 火枪手 / 蘑菇时间）。
- [ ] 每条有正价格的行恰好高亮一个或多个橘格，且橘格数值 = 该行最大 `recycle_price`。
- [ ] 橘格是整格填充、白字，没有只高亮数字或加边框。
- [ ] 失败语义正确：`error`/`skipped`/`unavailable` 显示为「错误/未收录/不收」，没有出现假 `0`。
- [ ] 名称没有截断成省略号；长名可略微缩字号，但不能丢字。
- [ ] 副标题时间与 JSON `fetched_at` 一致。

## 注意事项

- 不要做缓存、不要落库、不要起后台服务；这个 skill 一天只跑 2~3 次，每次 30 秒左右可接受。
- 不要修改 `data/custom.json` 来完成出图；缺 ID 就如实显示「未收录」。
- 商家目录会变，价格波动正常；本图只是抓取时刻快照，副标题必须标时间。
