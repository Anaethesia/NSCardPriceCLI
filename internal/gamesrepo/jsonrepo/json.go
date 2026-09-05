// Package jsonrepo loads the game list from a JSON file (data/games.json by
// default).
package jsonrepo

import (
	"encoding/json"
	"os"
	"sort"

	"nscardprice/internal/game"
)

// Repo is a gamesrepo.Repository backed by a JSON file, loaded once.
type Repo struct {
	games []game.Game
}

// Load reads and decodes the JSON file. Games are sorted by slug so output is
// deterministic regardless of file ordering.
func Load(path string) (*Repo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadBytes(b)
}

// LoadBytes decodes the game list from raw JSON bytes, sorted by slug.
func LoadBytes(b []byte) (*Repo, error) {
	var games []game.Game
	if err := json.Unmarshal(b, &games); err != nil {
		return nil, err
	}
	sort.Slice(games, func(i, j int) bool { return games[i].Slug < games[j].Slug })
	return &Repo{games: games}, nil
}

// All returns every game.
func (r *Repo) All() []game.Game {
	return r.games
}

// Enabled returns only enabled games.
func (r *Repo) Enabled() []game.Game {
	var out []game.Game
	for _, g := range r.games {
		if g.Enabled {
			out = append(out, g)
		}
	}
	return out
}
