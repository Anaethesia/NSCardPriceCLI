// Package config holds the runtime options and their defaults, shared by the
// CLI layer and the application assembly.
package config

import "time"

const (
	// DefaultGamesFile is the JSON list of tracked games and merchant IDs.
	DefaultGamesFile = "data/games.json"
	// DefaultListsDir is where the list/sync commands write merchant catalogs.
	DefaultListsDir = "data/lists"
	// DefaultCustomFile is the optional user-maintained custom library queried
	// via `query -custom`.
	DefaultCustomFile = "data/custom.json"
	// DefaultLogDir is where Errorf/Infof traces are appended as log files.
	DefaultLogDir = "logs"

	DefaultTimeout     = 10 * time.Second
	DefaultRetries     = 2
	DefaultBackoff     = 1 * time.Second
	DefaultConcurrency = 8
	// DefaultRateGap spaces consecutive API requests to avoid triggering the
	// merchants' rate limiting when querying several stores/products at once.
	DefaultRateGap = 200 * time.Millisecond
)

// Options are the resolved runtime settings passed from the CLI into the app.
type Options struct {
	GamesFile   string
	ListsDir    string
	CustomFile  string
	LogDir      string
	Merchants   []string // empty means "all supported merchants"
	Timeout     time.Duration
	Retries     int
	Backoff     time.Duration
	Concurrency int
	RateGap     time.Duration // <=0 means use DefaultRateGap
	Insecure    bool
	DryRun      bool
	Verbose     bool
}
