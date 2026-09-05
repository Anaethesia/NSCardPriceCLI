// Package nscardprice holds embedded default assets for the CLI. It lives at
// the module root so go:embed can reach data/games.json.
package nscardprice

import _ "embed"

// GamesJSON is the default game catalog shipped with the binary. It mirrors
// data/games.json and is used only when the default file cannot be read from
// the working directory, so explicitly passed --games paths take priority.
//
//go:embed data/games.json
var GamesJSON []byte
