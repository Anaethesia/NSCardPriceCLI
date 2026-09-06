package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"nscardprice/internal/app"
	"nscardprice/internal/collector"
	"nscardprice/internal/config"
	"nscardprice/internal/game"
	"nscardprice/internal/logging"
	"nscardprice/internal/match"
	"nscardprice/internal/merchant"
)

// maxQueryKeywords bounds how many keywords `query --mul` accepts at once.
const maxQueryKeywords = 5

func newQueryCmd(opts *config.Options) *cobra.Command {
	var all, custom, mul, slug bool
	cmd := &cobra.Command{
		Use:   "query [关键词...]",
		Short: "查询卡带回收价",
		Args: func(cmd *cobra.Command, args []string) error {
			slugMode, _ := cmd.Flags().GetBool("slug")
			multi, _ := cmd.Flags().GetBool("mul")
			allMode, _ := cmd.Flags().GetBool("all")
			if slugMode {
				if allMode {
					return exitError(2, "--all 与 --slug 不能同时使用")
				}
				if multi {
					return exitError(2, "--slug 与 --mul 不能同时使用")
				}
				if len(args) != 1 {
					return exitError(2, "query --slug 需要恰好 1 个 slug 参数")
				}
				return nil
			}
			if !multi && len(args) > 1 {
				return exitError(2, fmt.Sprintf("query 最多接受 1 个关键词（用 --mul 可同时查询最多 %d 个）", maxQueryKeywords))
			}
			if multi && len(args) > maxQueryKeywords {
				return exitError(2, fmt.Sprintf("--mul 最多支持 %d 个关键词，当前 %d 个", maxQueryKeywords, len(args)))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if all && mul {
				return exitError(2, "--all 与 --mul 不能同时使用")
			}
			if all && slug {
				return exitError(2, "--all 与 --slug 不能同时使用")
			}
			if slug && mul {
				return exitError(2, "--slug 与 --mul 不能同时使用")
			}
			return runQuery(opts, args, all, custom, mul, slug, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "查询全部启用游戏")
	cmd.Flags().BoolVarP(&custom, "custom", "c", false, "仅在 data/custom.json 自定义库中查询")
	cmd.Flags().BoolVar(&mul, "mul", false, "多关键词查询（空格分隔，最多 5 个）")
	cmd.Flags().BoolVar(&slug, "slug", false, "按 slug 精确匹配（不区分大小写）")
	return cmd
}

// resolveQueryGames resolves the user input against base games. With --slug it
// does an exact (case-insensitive) slug match; otherwise it falls back to
// keyword/multi-keyword fuzzy matching.
func resolveQueryGames(base []game.Game, args []string, slug, mul bool) ([]game.Game, error) {
	if slug {
		return match.ExactSlug(base, strings.ToLower(strings.TrimSpace(args[0]))), nil
	}
	return queryGames(base, args, mul)
}

func runQuery(opts *config.Options, args []string, all, custom, mul, slug bool, out io.Writer) error {
	a, err := app.New(*opts)
	if err != nil {
		return exitError(2, err.Error())
	}

	var games []game.Game
	switch {
	case custom:
		games, err = loadCustomGames(opts.CustomFile)
		if err != nil {
			logging.Errorf("query custom 加载失败 file=%s: %v", opts.CustomFile, err)
			return exitError(2, err.Error())
		}
		if len(games) == 0 {
			return exitError(2, fmt.Sprintf("%s 为空（没有可查询的自定义条目）", opts.CustomFile))
		}
		if !all {
			games, err = resolveQueryGames(games, args, slug, mul)
			if err != nil {
				return err
			}
			if len(games) == 0 {
				return exitError(2, fmt.Sprintf("未找到匹配的自定义游戏: %v", args))
			}
		}
	case all:
		games = a.Repo.Enabled()
	default:
		games, err = resolveQueryGames(a.Repo.Enabled(), args, slug, mul)
		if err != nil {
			return err
		}
		if len(games) == 0 {
			return exitError(2, fmt.Sprintf("未找到匹配的游戏: %v", args))
		}
	}

	results, err := collector.Collect(context.Background(), games, a.Reg, collector.Options{
		Merchants:   opts.Merchants,
		Concurrency: opts.Concurrency,
		DryRun:      opts.DryRun,
	})
	if err != nil {
		logging.Errorf("query 收集失败: %v", err)
		return exitError(2, err.Error())
	}

	printResults(out, results)

	keyword := "(--all)"
	if len(args) > 0 {
		keyword = strings.Join(args, " | ")
	}
	errCount := 0
	for _, g := range results {
		for _, r := range g.Results {
			if r.Status == merchant.StatusError {
				errCount++
				logging.Errorf("query game=%q merchant=%s: %s", g.Slug, r.Merchant, r.Error)
			}
		}
	}
	logging.Infof("query keyword=%q games=%d errors=%d", keyword, len(results), errCount)

	for _, r := range results {
		if r.HasErrors() {
			return exitError(1, "")
		}
	}
	return nil
}

// queryGames resolves one keyword, or with --mul up to maxQueryKeywords
// keywords, against base. Multi-keyword results are deduplicated by slug and
// returned in keyword order (each keyword keeps base order).
func queryGames(base []game.Game, args []string, mul bool) ([]game.Game, error) {
	if !mul {
		if len(args) != 1 {
			return nil, exitError(2, "请提供游戏名/关键词，或使用 --all")
		}
		return match.Query(base, args[0]), nil
	}
	if len(args) == 0 {
		return nil, exitError(2, "--mul 需要至少 1 个关键词")
	}
	if len(args) > maxQueryKeywords {
		return nil, exitError(2, fmt.Sprintf("--mul 最多支持 %d 个关键词，当前 %d 个", maxQueryKeywords, len(args)))
	}
	seen := make(map[string]bool, len(base))
	var out []game.Game
	for _, kw := range args {
		for _, g := range match.Query(base, kw) {
			if !seen[g.Slug] {
				seen[g.Slug] = true
				out = append(out, g)
			}
		}
	}
	return out, nil
}

func printResults(out io.Writer, results []collector.GameResult) {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if len(results) == 1 {
		_ = enc.Encode(results[0])
	} else {
		_ = enc.Encode(results)
	}
}
