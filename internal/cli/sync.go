package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"nscardprice/internal/app"
	"nscardprice/internal/config"
	"nscardprice/internal/logging"
	"nscardprice/internal/merchant"
)

// SyncEntry summarizes one merchant's catalog in a sync run.
type SyncEntry struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Total       int    `json:"total"`
	Kept        int    `json:"ns_total"`
	WrittenPath string `json:"written_path,omitempty"`
}

// SyncSummary is the JSON written to stdout after a sync run.
type SyncSummary struct {
	SyncedAt  string      `json:"synced_at"`
	Merchants []SyncEntry `json:"merchants"`
}

func newSyncCmd(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "同步 4 家商家 NS/NS2 卡带目录到 data/lists",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(opts, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&opts.ListsDir, "out", config.DefaultListsDir, "落盘目录")
	return cmd
}

func runSync(opts *config.Options, out, errOut io.Writer) error {
	a, err := app.New(*opts)
	if err != nil {
		return exitError(2, err.Error())
	}

	keys := opts.Merchants
	if len(keys) == 0 {
		keys = a.Reg.Keys()
	} else {
		resolved := make([]string, 0, len(keys))
		for _, k := range keys {
			rk, ok := a.Reg.ResolveKey(k)
			if !ok {
				return exitError(2, fmt.Sprintf("未知商家: %s", k))
			}
			resolved = append(resolved, rk)
		}
		keys = resolved
	}

	summary := SyncSummary{SyncedAt: time.Now().Format(time.RFC3339)}
	for _, key := range keys {
		cataloger, ok := a.Reg.Catalog(key)
		if !ok {
			fmt.Fprintf(errOut, "跳过 %s：不支持目录查询\n", key)
			continue
		}
		entry := SyncEntry{Key: key, Name: a.Reg[key].Name()}
		if opts.DryRun {
			fmt.Fprintf(errOut, "[dry-run] 跳过 %s 的网络请求\n", key)
		} else {
			items, err := cataloger.Catalog(context.Background(), 0)
			if err != nil {
				logging.Errorf("sync 拉取目录 merchant=%s: %v", key, err)
				return exitError(1, fmt.Sprintf("%s: %v", key, err))
			}
			ns := make([]merchant.CatalogItem, 0, len(items))
			for _, it := range items {
				if merchant.IsNS(it) {
					ns = append(ns, it)
				}
			}
			entry.Total = len(items)
			entry.Kept = len(ns)
			if err := writeCatalog(opts.ListsDir, key, ns, errOut); err != nil {
				logging.Errorf("sync 写盘 merchant=%s: %v", key, err)
				return err
			}
			entry.WrittenPath = filepath.Join(opts.ListsDir, key+".json")
			logging.Infof("sync merchant=%s total=%d ns=%d path=%s", key, len(items), len(ns), entry.WrittenPath)
		}
		summary.Merchants = append(summary.Merchants, entry)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(summary)
}
