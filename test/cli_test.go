package test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"nscardprice/internal/cli"
	"nscardprice/internal/collector"
	"nscardprice/internal/gamesrepo/jsonrepo"
	"nscardprice/internal/merchant"
)

// gamesPath returns the absolute path to data/games.json, so CLI tests run
// regardless of the working directory.
func gamesPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("no caller frame")
	}
	return filepath.Join(filepath.Dir(file), "..", "data", "games.json")
}

// runCLI executes the command tree with the given args and returns captured
// stdout, stderr and the returned error. File logging is disabled so tests do
// not touch the repository's logs/ directory.
func runCLI(args ...string) (string, string, error) {
	root := cli.NewCommand()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs(append([]string{"--log-dir", ""}, args...))
	err := root.Execute()
	return out.String(), errBuf.String(), err
}

// TestCLIMerchants checks `merchants` prints the four supported storefronts.
func TestCLIMerchants(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "merchants")
	if err != nil {
		t.Fatalf("merchants error: %v", err)
	}

	var got []struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}

	want := map[string]string{
		"laolieren":    "老猎人",
		"huoqiangshou": "火枪手",
		"buerjia":      "不二家",
		"hangzhouxizi": "杭州西子",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d merchants, want %d", len(got), len(want))
	}
	for _, m := range got {
		if want[m.Key] != m.Name {
			t.Errorf("merchant %q has name %q, want %q", m.Key, m.Name, want[m.Key])
		}
	}
}

// TestCLIVersion checks `--version` prints the version string.
func TestCLIVersion(t *testing.T) {
	stdout, _, err := runCLI("--version")
	if err != nil {
		t.Fatalf("--version error: %v", err)
	}
	if got := strings.TrimSpace(stdout); got != "nscardprice 0.1.0" {
		t.Errorf("version = %q, want %q", got, "nscardprice 0.1.0")
	}
}

// TestCLIQuerySlugAsKeyword checks a full slug string is treated as a plain
// keyword (exact slug matching is temporarily disabled): it matches every game
// whose slug contains the string.
func TestCLIQuerySlugAsKeyword(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "--dry-run", "query", "zelda-breath-of-the-wild")
	if err != nil {
		t.Fatalf("query error: %v", err)
	}

	var results []collector.GameResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("invalid JSON output (want array of games): %v\n%s", err, stdout)
	}
	want := []string{
		"zelda-breath-of-the-wild",
		"zelda-breath-of-the-wild-bundle",
		"zelda-breath-of-the-wild-ns2",
	}
	if len(results) != len(want) {
		t.Fatalf("got %d games, want %d", len(results), len(want))
	}
	for i, w := range want {
		if results[i].Slug != w {
			t.Errorf("games[%d].slug = %q, want %q", i, results[i].Slug, w)
		}
		if len(results[i].Results) != 4 {
			t.Errorf("games[%d] has %d merchant results, want 4", i, len(results[i].Results))
		}
	}
}

// TestCLIQueryKeyword checks a Chinese keyword resolves via fuzzy matching.
func TestCLIQueryKeyword(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "--dry-run", "query", "双人成行")
	if err != nil {
		t.Fatalf("query error: %v", err)
	}

	var gr collector.GameResult
	if err := json.Unmarshal([]byte(stdout), &gr); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	if gr.Slug != "it-takes-two" {
		t.Errorf("slug = %q, want it-takes-two", gr.Slug)
	}
}

// TestCLIQueryMerchantFilter checks -m restricts results to one merchant.
func TestCLIQueryMerchantFilter(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "--dry-run", "-m", "buerjia", "query", "双人成行")
	if err != nil {
		t.Fatalf("query error: %v", err)
	}

	var gr collector.GameResult
	if err := json.Unmarshal([]byte(stdout), &gr); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	if len(gr.Results) != 1 {
		t.Fatalf("got %d results, want 1", len(gr.Results))
	}
	if gr.Results[0].Merchant != "buerjia" {
		t.Errorf("merchant = %q, want buerjia", gr.Results[0].Merchant)
	}
}

// TestCLIQueryAll checks --all returns every enabled game.
func TestCLIQueryAll(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "--dry-run", "query", "--all")
	if err != nil {
		t.Fatalf("query --all error: %v", err)
	}

	var results []collector.GameResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}

	repo, _ := jsonrepo.Load(gamesPath())
	if want := len(repo.Enabled()); len(results) != want {
		t.Errorf("--all returned %d games, want %d", len(results), want)
	}
}

// TestCLIQueryNoArgs checks a bare `query` fails with a usage error.
func TestCLIQueryNoArgs(t *testing.T) {
	_, _, err := runCLI("--games", gamesPath(), "--dry-run", "query")
	ee, ok := cli.AsExitError(err)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.Code != 2 {
		t.Errorf("exit code = %d, want 2", ee.Code)
	}
}

// TestCLIQueryMulti checks `--mul` queries several keywords at once and unions
// the matches in keyword order.
func TestCLIQueryMulti(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "--dry-run", "query", "--mul", "双人成行", "马8")
	if err != nil {
		t.Fatalf("query --mul error: %v", err)
	}

	var results []collector.GameResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	want := []string{"it-takes-two", "mario-kart-8-deluxe"}
	if len(results) != len(want) {
		t.Fatalf("got %d games, want %d", len(results), len(want))
	}
	for i, w := range want {
		if results[i].Slug != w {
			t.Errorf("games[%d].slug = %q, want %q", i, results[i].Slug, w)
		}
	}
}

// TestCLIQueryMultiDedupe checks repeated keywords do not produce duplicate
// games in the output.
func TestCLIQueryMultiDedupe(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "--dry-run", "query", "--mul", "旷野", "旷野之息")
	if err != nil {
		t.Fatalf("query --mul error: %v", err)
	}

	var results []collector.GameResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	seen := map[string]bool{}
	for _, g := range results {
		if seen[g.Slug] {
			t.Errorf("duplicate slug %q in --mul output", g.Slug)
		}
		seen[g.Slug] = true
	}
	if n := len(seen); n != 3 {
		t.Errorf("got %d unique games, want 3 (基础版/同捆/NS2)", n)
	}
}

// TestCLIQueryMultiTooMany checks more than 5 keywords are rejected.
func TestCLIQueryMultiTooMany(t *testing.T) {
	_, _, err := runCLI("--games", gamesPath(), "--dry-run", "query", "--mul",
		"一", "二", "三", "四", "五", "六")
	ee, ok := cli.AsExitError(err)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.Code != 2 || !strings.Contains(ee.Msg, "最多支持 5 个关键词") {
		t.Errorf("exit = (%d, %q), want (2, 最多支持 5 个关键词...)", ee.Code, ee.Msg)
	}
}

// TestCLIQueryMultiNoArgs checks `--mul` without keywords fails.
func TestCLIQueryMultiNoArgs(t *testing.T) {
	_, _, err := runCLI("--games", gamesPath(), "--dry-run", "query", "--mul")
	ee, ok := cli.AsExitError(err)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.Code != 2 || !strings.Contains(ee.Msg, "至少 1 个关键词") {
		t.Errorf("exit = (%d, %q), want (2, ...至少 1 个关键词...)", ee.Code, ee.Msg)
	}
}

// TestCLIQueryAllMulConflict checks --all and --mul are mutually exclusive.
func TestCLIQueryAllMulConflict(t *testing.T) {
	_, _, err := runCLI("--games", gamesPath(), "--dry-run", "query", "--all", "--mul", "旷野")
	ee, ok := cli.AsExitError(err)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.Code != 2 || !strings.Contains(ee.Msg, "不能同时使用") {
		t.Errorf("exit = (%d, %q), want (2, ...不能同时使用...)", ee.Code, ee.Msg)
	}
}

// TestCLIListDryRun checks `list --dry-run` prints a name + game_id JSON
// without hitting the network.
func TestCLIListDryRun(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "--dry-run", "list", "laolieren", "10")
	if err != nil {
		t.Fatalf("list error: %v", err)
	}

	var items []merchant.CatalogItem
	if err := json.Unmarshal([]byte(stdout), &items); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	if len(items) == 0 || items[0].GameID == "" || items[0].Name == "" {
		t.Errorf("expected non-empty catalog item, got %+v", items)
	}
}

// TestCLIListUnknownMerchant checks an unknown merchant fails cleanly.
func TestCLIListUnknownMerchant(t *testing.T) {
	_, _, err := runCLI("--games", gamesPath(), "--dry-run", "list", "nope")
	ee, ok := cli.AsExitError(err)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.Code != 2 || !strings.Contains(ee.Msg, "未知商家") {
		t.Errorf("exit = (%d, %q), want (2, 未知商家...)", ee.Code, ee.Msg)
	}
}

// TestCLIListBadCount checks an invalid count fails with a usage error.
func TestCLIListBadCount(t *testing.T) {
	_, _, err := runCLI("--games", gamesPath(), "--dry-run", "list", "laolieren", "abc")
	ee, ok := cli.AsExitError(err)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.Code != 2 || !strings.Contains(ee.Msg, "无效的数量") {
		t.Errorf("exit = (%d, %q), want (2, 无效的数量...)", ee.Code, ee.Msg)
	}
}

// TestCLICustomQuery checks `query -custom` resolves only inside the custom
// library file, ignoring the main games catalog.
func TestCLICustomQuery(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "custom.json")
	// Use a name that does NOT exist in data/games.json to prove isolation.
	body := `[{"slug":"custom-kart","name":"NS 定制卡丁车","enabled":true,
		"merchant_ids":{"buerjia":{"game_id":"5"}}}]`
	if err := os.WriteFile(custom, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := runCLI("--games", gamesPath(), "--custom-file", custom, "--dry-run", "query", "-custom")
	if err == nil {
		t.Fatal("expected error: -custom is not a valid single-dash flag")
	}
	stdout, _, err = runCLI("--games", gamesPath(), "--custom-file", custom, "--dry-run", "query", "--custom", "定制卡丁车")
	if err != nil {
		t.Fatalf("query -custom error: %v", err)
	}
	var gr collector.GameResult
	if err := json.Unmarshal([]byte(stdout), &gr); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	if gr.Slug != "custom-kart" {
		t.Errorf("slug = %q, want custom-kart", gr.Slug)
	}
	if len(gr.Results) != 4 {
		t.Fatalf("got %d results, want 4 (buerjia + 3 skipped)", len(gr.Results))
	}
	for _, r := range gr.Results {
		if r.Merchant == "buerjia" && r.Status != merchant.StatusOK {
			t.Errorf("buerjia status = %q, want ok", r.Status)
		}
	}
}

// TestCLICustomEmpty checks `query -custom` on an empty library fails with a
// clear usage error.
func TestCLICustomEmpty(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(custom, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := runCLI("--games", gamesPath(), "--custom-file", custom, "--dry-run", "query", "--custom", "任意关键词")
	ee, ok := cli.AsExitError(err)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.Code != 2 || !strings.Contains(ee.Msg, "为空") {
		t.Errorf("exit = (%d, %q), want (2, ...为空...)", ee.Code, ee.Msg)
	}
}

// TestCLISyncDryRun checks `sync --dry-run` reports all four merchants via a
// JSON summary without hitting the network.
func TestCLISyncDryRun(t *testing.T) {
	stdout, _, err := runCLI("--games", gamesPath(), "--dry-run", "sync")
	if err != nil {
		t.Fatalf("sync error: %v", err)
	}
	var s struct {
		Merchants []struct {
			Key string `json:"key"`
		} `json:"merchants"`
	}
	if err := json.Unmarshal([]byte(stdout), &s); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}
	want := []string{"buerjia", "hangzhouxizi", "huoqiangshou", "laolieren"}
	if len(s.Merchants) != len(want) {
		t.Fatalf("got %d merchants, want %d", len(s.Merchants), len(want))
	}
	for i, m := range s.Merchants {
		if m.Key != want[i] {
			t.Errorf("merchant[%d] = %q, want %q", i, m.Key, want[i])
		}
	}
}
