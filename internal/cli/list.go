package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"nscardprice/internal/app"
	"nscardprice/internal/config"
	"nscardprice/internal/logging"
	"nscardprice/internal/merchant"
)

func newListCmd(opts *config.Options) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "list <merchant> [count]",
		Short: "列出某商家卡带目录 (name + game_id)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return exitError(2, "请指定商家，例如: nscardprice list laolieren 10")
			}
			if len(args) > 2 {
				return exitError(2, "参数过多: list 最多接收 <merchant> [count] 两个参数")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts, args, all, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "拉取全部（忽略 count）")
	cmd.Flags().StringVar(&opts.ListsDir, "out", config.DefaultListsDir, "落盘目录")
	return cmd
}

func runList(opts *config.Options, args []string, all bool, out, errOut io.Writer) error {
	a, err := app.New(*opts)
	if err != nil {
		return exitError(2, err.Error())
	}
	key, ok := a.Reg.ResolveKey(args[0])
	if !ok {
		return exitError(2, fmt.Sprintf("未知商家: %s", args[0]))
	}
	cataloger, ok := a.Reg.Catalog(key)
	if !ok {
		return exitError(2, fmt.Sprintf("商家 %s 不支持目录查询", key))
	}

	limit := 10
	if all {
		limit = 0
	} else if len(args) == 2 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 0 {
			return exitError(2, fmt.Sprintf("无效的数量: %q", args[1]))
		}
		limit = n
	}

	if opts.DryRun {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode([]merchant.CatalogItem{
			{Name: "塞尔达传说 旷野之息", GameID: "154", Platform: "Nintendo Switch"},
		})
		return nil
	}

	items, err := cataloger.Catalog(context.Background(), limit)
	if err != nil {
		logging.Errorf("list 拉取目录 merchant=%s: %v", key, err)
		return exitError(1, err.Error())
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	_ = enc.Encode(items)

	logging.Infof("list merchant=%s items=%d limit=%d", key, len(items), limit)
	return writeCatalog(opts.ListsDir, key, items, errOut)
}

func writeCatalog(dir, key string, items []merchant.CatalogItem, errOut io.Writer) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return exitError(1, err.Error())
	}
	path := filepath.Join(dir, key+".json")
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return exitError(1, err.Error())
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		logging.Errorf("list 写盘 %s: %v", path, err)
		return exitError(1, err.Error())
	}
	fmt.Fprintf(errOut, "已写入 %s (%d 条)\n", path, len(items))
	logging.Infof("wrote %s (%d 条)", path, len(items))
	return nil
}
