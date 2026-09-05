// Package gamesrepo exposes the game list as an interface so the data source
// can be swapped (JSON file today, database or remote API later) without
// touching the layers above.
package gamesrepo

import "nscardprice/internal/game"

// Repository supplies the tracked games.
type Repository interface {
	// All returns every game, including disabled entries.
	All() []game.Game
	// Enabled returns only enabled games.
	Enabled() []game.Game
}
