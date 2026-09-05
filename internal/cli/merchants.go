package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"nscardprice/internal/app"
	"nscardprice/internal/config"
)

func newMerchantsCmd(opts *config.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "merchants",
		Short: "列出支持的商家",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(*opts)
			if err != nil {
				return exitError(2, err.Error())
			}
			type merchantInfo struct {
				Key  string `json:"key"`
				Name string `json:"name"`
			}
			out := make([]merchantInfo, 0, len(a.Reg))
			for _, k := range a.Reg.Keys() {
				out = append(out, merchantInfo{Key: k, Name: a.Reg[k].Name()})
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(out)
		},
	}
}
