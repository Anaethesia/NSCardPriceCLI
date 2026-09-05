package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"nscardprice/internal/config"
	"nscardprice/internal/logging"
)

const version = "0.1.0"

// NewCommand builds the root cobra command tree.
func NewCommand() *cobra.Command {
	opts := &config.Options{}

	root := &cobra.Command{
		Use:           "nscardprice",
		Short:         "Nintendo Switch 卡带回收价查询",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("nscardprice {{.Version}}\n")
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return exitError(2, err.Error())
	})
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := logging.Setup(opts.LogDir); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "警告: 日志目录不可用:", err)
		}
		logging.Infof("start cmd=%s", cmd.Name())
		return nil
	}

	pf := root.PersistentFlags()
	pf.StringVar(&opts.GamesFile, "games", config.DefaultGamesFile, "games.json 路径")
	pf.StringVar(&opts.CustomFile, "custom-file", config.DefaultCustomFile, "custom.json 路径")
	pf.StringVar(&opts.LogDir, "log-dir", config.DefaultLogDir, "日志目录（空字符串=不写日志文件）")
	pf.StringArrayVarP(&opts.Merchants, "merchant", "m", nil, "仅查询指定商家（可重复，key 或中文名）")
	pf.DurationVar(&opts.Timeout, "timeout", config.DefaultTimeout, "单请求超时")
	pf.IntVar(&opts.Retries, "retries", config.DefaultRetries, "每请求重试次数")
	pf.DurationVar(&opts.Backoff, "backoff", config.DefaultBackoff, "重试退避间隔")
	pf.IntVar(&opts.Concurrency, "concurrency", config.DefaultConcurrency, "并发上限")
	pf.BoolVar(&opts.Insecure, "insecure", false, "跳过 TLS 校验")
	pf.BoolVar(&opts.DryRun, "dry-run", false, "不发起真实请求")
	pf.BoolVar(&opts.Verbose, "verbose", false, "输出调试日志到 stderr")

	root.AddCommand(newQueryCmd(opts), newListCmd(opts), newMerchantsCmd(opts), newSyncCmd(opts))
	return root
}

// Run executes the command tree and returns the process exit code.
func Run() int {
	if err := NewCommand().Execute(); err != nil {
		if ee, ok := AsExitError(err); ok && ee.Msg != "" {
			fmt.Fprintln(os.Stderr, "错误:", ee.Msg)
		}
		return ExitCode(err)
	}
	return 0
}
