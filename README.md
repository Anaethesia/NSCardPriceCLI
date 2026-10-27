English | [简体中文](README.zh-CN.md)

# NSCardPriceCLI

A Go CLI for querying Nintendo Switch cartridge recycle prices. It reads games and merchant item IDs from `data/games.json`, queries **laolieren (老猎人), huoqiangshou (火枪手), buerjia (不二家), hangzhouxizi (杭州西子), hailuo (海螺), and mogushijian (蘑菇时间)** for recycle/sell prices, and outputs JSON that is easy to consume from a terminal, agent, or script.

## Features

- Query Nintendo Switch cartridge recycle/sell prices across six recyclers
- Fuzzy matching by Chinese game name/keyword; common abbreviations and aliases are recognized automatically
- Exact slug lookup (`query --slug`, case-insensitive)
- Query multiple keywords at once (up to 5) and get several games in one call
- Independent custom game library that does not affect the built-in catalog
- Sync the NS/NS2 cartridge catalogs of all six merchants with one command, filtering out other platforms
- Per-merchant response parsing: recycle price, sell price, and unavailable status
- Concurrent queries; a failure at one merchant does not affect the others
- Built-in 0.2s request pacing to avoid triggering merchant rate limits
- Pure JSON on stdout; logs and errors on stderr
- Clear exit codes: 0 = success (including partial merchant failures or unavailable/skipped results), 1 = all merchants failed, 2 = usage error

## Commands

```bash
nscardprice query 塞尔达              # keyword query across all 6 merchants
nscardprice query --slug zelda-breath-of-the-wild  # exact slug lookup (case-insensitive)
nscardprice query --custom --slug mk8-custom       # exact slug lookup in the custom library
nscardprice query --mul 旷野 马8 双人成行  # query multiple keywords (up to 5)
nscardprice query --custom 马车8             # query only the data/custom.json library
nscardprice query --custom --all             # query every entry in the custom library
nscardprice --merchant buerjia query 旷野    # query a single merchant (repeatable)
nscardprice query -m 老猎人,不二家 旷野      # comma-separated merchant list
nscardprice query --all                      # query all games
nscardprice list laolieren 10                # list the first 10 entries of the laolieren catalog (name + game_id)
nscardprice list buerjia --all --out data/lists   # fetch the full catalog and save it to data/lists/
nscardprice sync                             # sync all six NS/NS2 catalogs to data/lists
nscardprice sync -m huoqiangshou             # sync only the specified merchant
nscardprice merchants                        # list the six merchant keys
nscardprice --version                        # print the version
```

Common global flags: `--games <path>` (default `data/games.json`), `--custom-file <path>` (default `data/custom.json`), `--timeout` (default 10s), `--retries` (default 2), `--concurrency` (default 8), `--insecure`, `--dry-run`, `--verbose`.

### data/custom.json custom library

Independent of `data/games.json`, this library is for cartridges you query often but do not want to mix into the main catalog (same structure; a missing `enabled` field is treated as true). Use `query --custom` to match and query only within this library:

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

Example output (single game):

```json
{"fetched_at":"...","slug":"...","name":"塞尔达传说 王国之泪",
 "results":[{"merchant":"laolieren","status":"ok","recycle_price":170,"sell_price":200,"currency":"CNY",...}]}
```

## Architecture

```
cmd/nscardprice/    entry point
internal/app/       dependency wiring
internal/config/    runtime options and defaults
internal/game/      domain models (zero dependencies)
internal/gamesrepo/ game catalog repository (JSON implementation)
internal/merchant/  merchant abstraction + six implementations
internal/httpx/     HTTP wrapper with retry/backoff/TLS support
internal/parse/     JSON parsing helpers
internal/collector/ concurrent orchestration and result aggregation
internal/cli/       cobra commands
test/               live-API integration tests (build-tag isolated, not run by default)
data/games.json     game and merchant item ID catalog
data/custom.json    custom query library (used by query --custom)
data/lists/         full merchant catalogs (list --all / sync keeps only NS/NS2)
logs/               error/runtime logs (nscardprice.log, append-only)
```

Extension: add a merchant implementation in `internal/merchant` and register it; replace the data source by implementing the `gamesrepo.Repository` interface.

## Build

Install to PATH (recommended; afterwards use `nscardprice` directly):

```bash
go install ./cmd/nscardprice     # installs to ~/go/bin (takes effect if already on PATH)
nscardprice --help
```

Or build a local binary and run it directly:

```bash
make build                       # produces bin/nscardprice
./bin/nscardprice list
```

## Live API integration tests

```bash
NSCARD_INTEGRATION=1 go test -tags integration ./test/ -v
```

The tests pause 3s between requests to avoid triggering third-party rate limits; `go test ./...` is unaffected.
