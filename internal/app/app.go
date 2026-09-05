// Package app wires the runtime options into the concrete dependencies
// (HTTP client, game repository, merchant registry) shared by the commands.
package app

import (
	nscardprice "nscardprice"
	"nscardprice/internal/config"
	"nscardprice/internal/gamesrepo"
	"nscardprice/internal/gamesrepo/jsonrepo"
	"nscardprice/internal/httpx"
	"nscardprice/internal/merchant"
)

// App holds the assembled dependencies for a command invocation.
type App struct {
	Opts config.Options
	Repo gamesrepo.Repository
	Reg  merchant.Registry
	HTTP *httpx.Client
}

// New assembles the app from options, loading the game list and building the
// merchant registry.
func New(opts config.Options) (*App, error) {
	rateGap := opts.RateGap
	if rateGap <= 0 {
		rateGap = config.DefaultRateGap
	}
	httpClient := httpx.New(httpx.Options{
		Timeout:  opts.Timeout,
		Retries:  opts.Retries,
		Backoff:  opts.Backoff,
		RateGap:  rateGap,
		Insecure: opts.Insecure,
	})
	repo, err := jsonrepo.Load(opts.GamesFile)
	if err != nil && opts.GamesFile == config.DefaultGamesFile {
		// The default catalog is missing in the working directory; fall back to
		// the embedded copy so the CLI works from anywhere.
		repo, err = jsonrepo.LoadBytes(nscardprice.GamesJSON)
	}
	if err != nil {
		return nil, err
	}
	return &App{
		Opts: opts,
		Repo: repo,
		Reg:  merchant.NewRegistry(httpClient),
		HTTP: httpClient,
	}, nil
}
