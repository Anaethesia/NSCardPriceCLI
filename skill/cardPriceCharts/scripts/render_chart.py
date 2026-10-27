#!/usr/bin/env python3
"""Render a Switch cartridge recycle-price comparison table as a PNG.

Data comes from data/custom.json: the first N entries by default, or only the
cartridges the user names. Rendering is plain HTML + headless Chrome, so no
third-party Python packages are required (no PIL / matplotlib / cairosvg).

Usage:
    python3 render_chart.py --limit 30
    python3 render_chart.py --names "捡树枝,双人成行"
    python3 render_chart.py --limit 30 --out reports/card-price-chart.png
    python3 render_chart.py --input-json /tmp/query.json
"""

import argparse
import html
import json
import os
import re
import shutil
import signal
import subprocess
import sys
import tempfile
import time
import unicodedata
from pathlib import Path

# Fixed column order: (merchant key, display name).
MERCHANTS = [
    ("laolieren", "老猎人"),
    ("hailuo", "海螺"),
    ("buerjia", "不二家"),
    ("hangzhouxizi", "杭州西子"),
    ("huoqiangshou", "火枪手"),
    ("mogushijian", "蘑菇时间"),
]

# Layout constants; keep in sync with the CSS below so the screenshot height
# matches the rendered content exactly.
PAGE_TOP = 44
TITLE_LH = 46
SUB_MT = 10
SUB_LH = 24
CARD_MT = 26
HEAD_H = 46
ROW_H = 54
CARD_PB = 26
PAGE_BOTTOM = 40
WIDTH = 1680
NAME_COL_W = 500
DEFAULT_LIMIT = 30
MAX_GAMES = 30

CHROME_CANDIDATES = [
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
    "/Applications/Chromium.app/Contents/MacOS/Chromium",
    "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
    "google-chrome",
    "chromium",
    "chromium-browser",
    "chrome",
]


def repo_root():
    # Inside the repo the script lives in skill/cardPriceCharts/scripts.
    # Installed skills live outside it, so also honor NSCARD_HOME, the default
    # clone location, and walk up from the script/cwd looking for data/custom.json.
    here = Path(__file__).resolve()
    candidates = []
    env = os.environ.get("NSCARD_HOME")
    if env:
        candidates.append(Path(env).expanduser())
    candidates.append(Path.home() / "NSCardPriceCLI")
    candidates.extend(here.parents)
    candidates.extend([Path.cwd(), *Path.cwd().parents])
    for candidate in candidates:
        if (candidate / "data" / "custom.json").is_file():
            return candidate.resolve()
    return here.parents[3]


def load_custom_games(custom_file):
    data = json.loads(Path(custom_file).read_text(encoding="utf-8"))
    if not isinstance(data, list):
        raise SystemExit("custom.json 必须是数组: %s" % custom_file)
    return [g for g in data if g.get("enabled", True)]


def norm(text):
    value = unicodedata.normalize("NFKC", str(text or "")).lower()
    return re.sub(r"\s+", "", value)


def split_names(raw):
    return [part.strip() for part in re.split(r"[,，、;；]+", raw or "") if part.strip()]


def cli_matches(cli, custom_file, name, timeout):
    """Resolve one cartridge name with the CLI matcher; dry-run never hits the network."""
    cmd = [
        cli,
        "query",
        "--custom",
        name,
        "--custom-file",
        str(custom_file),
        "--log-dir",
        "",
        "--dry-run",
    ]
    try:
        proc = subprocess.run(
            cmd,
            cwd=str(repo_root()),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            timeout=timeout,
        )
    except (OSError, subprocess.TimeoutExpired):
        return None
    if proc.returncode != 0:
        return None
    try:
        return normalize(json.loads(proc.stdout.decode("utf-8")))
    except ValueError:
        return None


def local_matches(games, name):
    query = norm(name)
    if not query:
        return []
    for field in ("slug", "name", "search_keyword"):
        exact = [g for g in games if norm(g.get(field)) == query]
        if exact:
            return exact
    return [
        g
        for g in games
        if query in norm(g.get("name")) or query in norm(g.get("search_keyword"))
    ]


def resolve_named_games(names, games, cli, custom_file, timeout):
    """Map user-supplied names to custom.json entries, preserving the given order."""
    by_slug = {g.get("slug"): g for g in games}
    resolved = []
    seen = set()
    for name in names:
        hits = cli_matches(cli, custom_file, name, timeout) if cli else None
        if not hits:
            hits = local_matches(games, name)
        if not hits:
            raise SystemExit("没有找到卡带「%s」：请换一个更准确的名称，或直接给 slug" % name)
        if len(hits) > 1:
            exact = [
                g
                for g in hits
                if norm(g.get("name")) == norm(name) or norm(g.get("slug")) == norm(name)
            ]
            if len(exact) == 1:
                hits = exact
            else:
                labels = "、".join("%s (%s)" % (g.get("name"), g.get("slug")) for g in hits)
                raise SystemExit(
                    "卡带名「%s」匹配到多款，请让用户确认或用 slug 指定：%s" % (name, labels)
                )
        slug = hits[0].get("slug")
        game = by_slug.get(slug)
        if game is None:
            raise SystemExit("slug %s 不在 custom.json 中（可能已停用）" % slug)
        if slug in seen:
            continue
        seen.add(slug)
        resolved.append(game)
    return resolved


def resolve_cli(explicit):
    if explicit:
        return explicit
    local = repo_root() / "bin" / "nscardprice"
    if local.is_file():
        return str(local)
    found = shutil.which("nscardprice")
    if found:
        return found
    raise SystemExit(
        "找不到 nscardprice：请按 skill 的「安装 CLI」执行 "
        "`go install ./cmd/nscardprice`，或用 --cli 指定可执行文件路径"
    )


def fetch_prices(games, cli, timeout):
    tmpdir = tempfile.mkdtemp(prefix="card-price-")
    subset = Path(tmpdir) / "subset.json"
    subset.write_text(json.dumps(games, ensure_ascii=False), encoding="utf-8")

    cmd = [
        cli,
        "query",
        "--custom",
        "--all",
        "--custom-file",
        str(subset),
        "--log-dir",
        "",
    ]
    sys.stderr.write("查询 %d 款 × %d 家：%s\n" % (len(games), len(MERCHANTS), " ".join(cmd)))
    try:
        proc = subprocess.run(
            cmd,
            cwd=str(repo_root()),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            timeout=timeout,
        )
    except subprocess.TimeoutExpired:
        raise SystemExit("nscardprice 查询超时（%ss）" % timeout)

    stderr = proc.stderr.decode("utf-8", "replace").strip()
    if proc.returncode == 2:
        raise SystemExit("CLI 用法错误/未找到（退出码 2）：%s" % (stderr or "无 stderr"))
    if proc.returncode == 1:
        raise SystemExit("所有商家均查询失败（退出码 1）：%s" % (stderr[-500:] or "无 stderr"))

    try:
        payload = json.loads(proc.stdout.decode("utf-8"))
    except ValueError:
        raise SystemExit(
            "CLI 输出不是合法 JSON（退出码 %s）：%s" % (proc.returncode, stderr[-500:] or "无 stderr")
        )

    if proc.returncode != 0:
        sys.stderr.write("警告：CLI 退出码 %s，仍使用已返回的数据\n" % proc.returncode)
    return normalize(payload)


def normalize(payload):
    if isinstance(payload, dict):
        return [payload]
    if isinstance(payload, list):
        return payload
    raise SystemExit("无法识别的查询结果结构: %s" % type(payload).__name__)


def fmt_price(value):
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        return None
    number = float(value)
    if number.is_integer():
        return "%d" % int(number)
    return "%.1f" % number


def cell_for(result):
    """Map one merchant result to (kind, text, numeric price or None)."""
    if result is None:
        return "na", "无数据", None
    status = str(result.get("status") or "").lower()
    if status == "ok":
        price = result.get("recycle_price")
        text = fmt_price(price)
        if text is None:
            return "na", "暂无", None
        return "price", text, float(price)
    if status == "unavailable":
        return "na", "不收", None
    if status == "skipped":
        return "na", "未收录", None
    if status == "error":
        return "err", "错误", None
    return "na", "—", None


def build_rows(games, payload):
    by_slug = {}
    for game in payload:
        slug = game.get("slug")
        if slug:
            by_slug[slug] = game

    rows = []
    missing_merchants = set()
    for game in games:
        found = by_slug.get(game.get("slug"))
        results = {}
        if found:
            for item in found.get("results") or []:
                key = item.get("merchant")
                if key:
                    results[key] = item
        cells = []
        for key, _label in MERCHANTS:
            if found is None:
                cells.append(("na", "无数据", None))
                continue
            cells.append(cell_for(results.get(key)))
        for key, _label in MERCHANTS:
            if found is not None and key not in results:
                missing_merchants.add(key)

        positive = [c[2] for c in cells if c[2] is not None and c[2] > 0]
        top = max(positive) if positive else None
        rows.append(
            {
                "name": found.get("name") if found else game.get("name"),
                "platform": found.get("platform") if found else game.get("platform"),
                "cells": cells,
                "top": top,
                "fetched_at": found.get("fetched_at") if found else None,
            }
        )
    return rows, missing_merchants


def response_height(row_count):
    return (
        PAGE_TOP
        + TITLE_LH
        + SUB_MT
        + SUB_LH
        + CARD_MT
        + HEAD_H
        + row_count * ROW_H
        + CARD_PB
        + PAGE_BOTTOM
    )


def format_time(fetched_at):
    if fetched_at and len(fetched_at) >= 16:
        return fetched_at[:16].replace("T", " ")
    return fetched_at or "未知时间"


CSS = """
  * { box-sizing: border-box; }
  html, body { margin: 0; padding: 0; background: #F7F5F0; }
  body {
    width: %(width)dpx;
    padding: %(pt)dpx 44px %(pb)dpx 44px;
    color: #1F2328;
    font-family: -apple-system, "PingFang SC", "Hiragino Sans GB", "Helvetica Neue", Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
  }
  .title {
    font-size: 34px;
    line-height: %(title_lh)dpx;
    font-weight: 700;
    letter-spacing: .01em;
  }
  .title::before {
    content: "";
    display: inline-block;
    width: 6px;
    height: 26px;
    border-radius: 3px;
    background: #E8730C;
    margin-right: 12px;
    vertical-align: -3px;
  }
  .subtitle {
    margin-top: %(sub_mt)dpx;
    font-size: 14px;
    line-height: %(sub_lh)dpx;
    color: #7C7668;
    font-variant-numeric: tabular-nums;
  }
  .card {
    margin-top: %(card_mt)dpx;
    padding-bottom: %(card_pb)dpx;
    background: #FFFFFF;
    border-radius: 16px;
    overflow: hidden;
    box-shadow: 0 14px 34px rgba(31,35,40,.07);
  }
  table {
    width: 100%%;
    border-collapse: separate;
    border-spacing: 0;
    table-layout: fixed;
  }
  col.name-col { width: %(name_col_w)dpx; }
  thead th {
    height: %(head_h)dpx;
    padding: 0;
    background: #1F2328;
    color: #F7F5F0;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: .08em;
    text-align: center;
  }
  thead th.name { text-align: left; padding-left: 24px; }
  tbody td {
    height: %(row_h)dpx;
    padding: 0;
    text-align: center;
    font-size: 17px;
    font-variant-numeric: tabular-nums;
    color: #1F2328;
    box-shadow: inset 0 1px 0 #EFEBE2;
  }
  tbody tr:first-child td { box-shadow: none; }
  tbody tr.alt td { background: #FBFAF6; }
  td.name {
    text-align: left;
    padding: 0 18px 0 24px;
    font-size: 16px;
    font-weight: 600;
  }
  td.name .tag {
    display: inline-block;
    margin-left: 8px;
    padding: 1px 7px;
    border-radius: 5px;
    background: #F0EDE5;
    color: #8A8071;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: .06em;
    vertical-align: 2px;
  }
  td.price { font-weight: 600; }
  td.na { color: #B3ADA0; font-size: 14px; font-weight: 400; }
  td.err { color: #B4513C; font-size: 14px; font-weight: 400; }
  tbody tr.alt td.hi,
  tbody td.hi {
    background: #E8730C;
    color: #FFFFFF;
    font-size: 18px;
    font-weight: 700;
    box-shadow: none;
  }
"""

def cell_html(kind, text, is_top):
    classes = [kind]
    if is_top:
        classes.append("hi")
    return '<td class="%s">%s</td>' % (" ".join(classes), html.escape(text))


def build_html(rows):
    css = CSS % {
        "width": WIDTH,
        "pt": PAGE_TOP,
        "pb": PAGE_BOTTOM,
        "title_lh": TITLE_LH,
        "sub_mt": SUB_MT,
        "sub_lh": SUB_LH,
        "card_mt": CARD_MT,
        "card_pb": CARD_PB,
        "head_h": HEAD_H,
        "row_h": ROW_H,
        "name_col_w": NAME_COL_W,
    }

    fetched_at = next((r["fetched_at"] for r in rows if r["fetched_at"]), None)
    subtitle = format_time(fetched_at)

    body_rows = []
    for index, row in enumerate(rows):
        alt = " alt" if index % 2 == 1 else ""
        name = html.escape(str(row["name"] or ""))
        tag = '<span class="tag">NS2</span>' if "Switch 2" in str(row["platform"] or "") else ""
        cells = []
        for kind, text, value in row["cells"]:
            is_top = row["top"] is not None and value is not None and value == row["top"]
            cells.append(cell_html(kind, text, is_top))
        body_rows.append(
            '<tr class="%s"><td class="name">%s%s</td>%s</tr>'
            % (alt.strip(), name, tag, "".join(cells))
        )

    price_cols = "<col>" * len(MERCHANTS)
    headers = "".join("<th>%s</th>" % label for _key, label in MERCHANTS)

    return """<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<style>%(css)s</style>
</head>
<body>
  <div class="title">Switch 卡带回收价对比</div>
  <div class="subtitle">%(subtitle)s</div>
  <div class="card">
    <table>
      <colgroup><col class="name-col">%(price_cols)s</colgroup>
      <thead><tr><th class="name">卡带名称</th>%(headers)s</tr></thead>
      <tbody>
        %(body_rows)s
      </tbody>
    </table>
  </div>
</body>
</html>
""" % {
        "css": css,
        "subtitle": html.escape(subtitle),
        "price_cols": price_cols,
        "headers": headers,
        "body_rows": "\n        ".join(body_rows),
    }


def find_chrome(explicit):
    if explicit:
        return explicit
    env = os.environ.get("CHROME_PATH")
    if env:
        return env
    for candidate in CHROME_CANDIDATES:
        if os.path.isabs(candidate):
            if os.path.isfile(candidate):
                return candidate
            continue
        found = shutil.which(candidate)
        if found:
            return found
    raise SystemExit("找不到 Chrome/Chromium：请用 --chrome 指定路径或设置 CHROME_PATH")


def stop_process_tree(proc):
    try:
        os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
    except (AttributeError, ProcessLookupError, PermissionError):
        proc.terminate()
    try:
        proc.wait(timeout=3)
        return
    except subprocess.TimeoutExpired:
        pass
    try:
        os.killpg(os.getpgid(proc.pid), signal.SIGKILL)
    except (AttributeError, ProcessLookupError, PermissionError):
        proc.kill()


def wait_for_png(proc, out_path, deadline):
    """Chrome may keep running after writing the screenshot; poll the file."""
    last_size = -1
    stable = 0
    while time.time() < deadline:
        if out_path.is_file():
            size = out_path.stat().st_size
            if size > 10000:
                if size == last_size:
                    stable += 1
                    if stable >= 2:
                        return True
                else:
                    stable = 0
            last_size = size
        if proc.poll() is not None:
            break
        time.sleep(0.2)
    return out_path.is_file() and out_path.stat().st_size > 10000


def screenshot(chrome, html_path, out_path, height):
    out_path.parent.mkdir(parents=True, exist_ok=True)
    if out_path.exists():
        out_path.unlink()
    profile = tempfile.mkdtemp(prefix="card-price-chrome-")
    args = [
        "--disable-gpu",
        "--hide-scrollbars",
        "--no-first-run",
        "--no-default-browser-check",
        "--disable-extensions",
        "--force-device-scale-factor=2",
        "--window-size=%d,%d" % (WIDTH, height),
        "--virtual-time-budget=2000",
        "--user-data-dir=%s" % profile,
        "--screenshot=%s" % out_path,
        "file://%s" % html_path,
    ]

    last_error = ""
    for headless in ("--headless=new", "--headless"):
        proc = subprocess.Popen(
            [chrome, headless] + args,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            start_new_session=True,
        )
        try:
            if wait_for_png(proc, out_path, time.time() + 60):
                return
        finally:
            stop_process_tree(proc)
        last_error = "未生成图片（%s）" % headless
    raise SystemExit("Chrome 截图失败：%s" % last_error)


def main():
    parser = argparse.ArgumentParser(description="生成 Switch 卡带回收价对比表 PNG")
    parser.add_argument(
        "--limit", type=int, default=DEFAULT_LIMIT, help="默认取 custom.json 前 N 条（默认 30）"
    )
    parser.add_argument(
        "--names", default=None, help="用户点名的卡带，逗号/顿号分隔（优先生效，忽略 --limit）"
    )
    parser.add_argument("--custom-file", default=None, help="custom.json 路径（默认 <repo>/data/custom.json）")
    parser.add_argument("--cli", default=None, help="nscardprice 可执行文件路径")
    parser.add_argument("--input-json", default=None, help="跳过查询，复用已有的 query 输出 JSON")
    parser.add_argument("--out", default=None, help="PNG 输出路径（默认 <repo>/reports/card-price-chart.png）")
    parser.add_argument("--html-out", default=None, help="保留中间 HTML 到指定路径")
    parser.add_argument("--chrome", default=None, help="Chrome/Chromium 可执行文件路径")
    parser.add_argument("--timeout", type=int, default=600, help="查询超时秒数（默认 600）")
    args = parser.parse_args()

    root = repo_root()
    custom_file = args.custom_file or str(root / "data" / "custom.json")
    games = load_custom_games(custom_file)
    if not games:
        raise SystemExit("custom.json 没有可用条目: %s" % custom_file)

    names = split_names(args.names)
    if names:
        if len(names) > MAX_GAMES:
            raise SystemExit("一次最多点名 %d 款卡带，当前 %d 款" % (MAX_GAMES, len(names)))
        try:
            cli_for_names = resolve_cli(args.cli)
        except SystemExit:
            cli_for_names = None
        games = resolve_named_games(
            names, games, cli_for_names, custom_file, min(args.timeout, 60)
        )
    else:
        if args.limit < 1:
            raise SystemExit("--limit 必须 >= 1")
        games = games[: args.limit]

    if args.input_json:
        payload = normalize(json.loads(Path(args.input_json).read_text(encoding="utf-8")))
    else:
        payload = fetch_prices(games, resolve_cli(args.cli), args.timeout)

    rows, missing_merchants = build_rows(games, payload)
    if missing_merchants:
        sys.stderr.write(
            "警告：结果中缺少商家 %s，可能是 CLI 版本过旧（应为 6 家）\n"
            % ", ".join(sorted(missing_merchants))
        )
    if not any(kind != "err" for row in rows for kind, _text, _value in row["cells"]):
        raise SystemExit("所有商家均查询失败，没有可展示的价格")

    html_text = build_html(rows)
    if args.html_out:
        html_path = Path(args.html_out).resolve()
        html_path.parent.mkdir(parents=True, exist_ok=True)
        html_path.write_text(html_text, encoding="utf-8")
        cleanup = False
    else:
        handle = tempfile.NamedTemporaryFile(
            "w", suffix=".html", prefix="card-price-chart-", delete=False, encoding="utf-8"
        )
        handle.write(html_text)
        handle.close()
        html_path = Path(handle.name)
        cleanup = True

    out_path = Path(args.out) if args.out else root / "reports" / "card-price-chart.png"
    out_path = out_path.resolve()
    height = response_height(len(rows))

    try:
        screenshot(find_chrome(args.chrome), html_path, out_path, height)
    finally:
        if cleanup:
            html_path.unlink(missing_ok=True)

    statuses = {}
    highlighted = 0
    for row in rows:
        if row["top"] is not None:
            highlighted += 1
        for kind, text, _value in row["cells"]:
            label = {"price": "ok", "na": text, "err": "error"}[kind]
            statuses[label] = statuses.get(label, 0) + 1

    summary = {
        "out": str(out_path),
        "rows": len(rows),
        "merchants": len(MERCHANTS),
        "rows_highlighted": highlighted,
        "cells": statuses,
        "fetched_at": next((r["fetched_at"] for r in rows if r["fetched_at"]), None),
        "width": WIDTH * 2,
        "height": height * 2,
    }
    print(json.dumps(summary, ensure_ascii=False))


if __name__ == "__main__":
    main()
