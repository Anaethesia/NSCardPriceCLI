// Package match implements fuzzy keyword matching between a user query and the
// tracked games. It ports the MATCH_RULES / REJECT_RULES / SYNONYMS / NOISE
// tables from the original Python project so aliases like 马车8 and the
// distinction between a base game and its DLC 同捆 edition resolve correctly.
// The tuning tables (synonyms, noise, matchRules, rejectRules) live in
// common.go.
package match

import (
	"regexp"
	"strings"

	"nscardprice/internal/game"
)

var stripRe = regexp.MustCompile(`[\s\-_/·:：,，。()（）\[\]【】+]+`)

// Normalize lowercases, applies synonyms, strips punctuation and removes noise
// words. It is exported so tests can assert on the expected normalization.
func Normalize(s string) string {
	v := strings.ToLower(s)
	for _, pair := range synonyms {
		v = strings.ReplaceAll(v, pair[0], pair[1])
	}
	v = stripRe.ReplaceAllString(v, "")
	for _, w := range noise {
		v = strings.ReplaceAll(v, w, "")
	}
	return v
}

// Query returns the enabled games matching the keyword, preserving input order
// (games are already sorted by slug by the repository). Slug exact matching is
// temporarily disabled, so a slug-shaped keyword is matched like any other
// keyword (see ExactSlug, kept for future re-enabling).
func Query(games []game.Game, q string) []game.Game {
	ql := strings.ToLower(strings.TrimSpace(q))
	if ql == "" {
		return nil
	}
	qn := Normalize(q)

	var out []game.Game
	for _, g := range games {
		if !g.Enabled {
			continue
		}
		if matches(g, ql, qn) {
			out = append(out, g)
		}
	}
	return out
}

func matches(g game.Game, ql, qn string) bool {
	// 1. Raw case-insensitive substring on slug/name/search keyword.
	for _, raw := range []string{g.Slug, g.Name, g.SearchKeyword} {
		if raw != "" && strings.Contains(strings.ToLower(raw), ql) {
			return true
		}
	}

	// 2. Normalized containment, handling synonyms and spacing differences.
	for _, raw := range []string{g.Name, g.SearchKeyword} {
		n := Normalize(raw)
		if n == "" {
			continue
		}
		if strings.Contains(n, qn) || strings.Contains(qn, n) {
			return true
		}
	}

	// 3. Alias rules: all groups must match, and no reject term may be present.
	if groups, ok := matchRules[g.Slug]; ok && len(groups) > 0 {
		if groupsMatch(qn, groups) && !groupRejected(g.Slug, qn) {
			return true
		}
	}
	return false
}

func groupsMatch(qn string, groups [][]string) bool {
	for _, group := range groups {
		ok := false
		for _, term := range group {
			if t := Normalize(term); t != "" && strings.Contains(qn, t) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func groupRejected(slug, qn string) bool {
	for _, term := range rejectRules[slug] {
		if t := Normalize(term); t != "" && strings.Contains(qn, t) {
			return true
		}
	}
	return false
}

// ExactSlug returns the single enabled game whose slug equals ql
// (case-insensitive), or nil. It is not called by Query right now: the CLI
// temporarily treats slugs as plain keywords, and this exact-slug fast path is
// kept so the behavior can be re-enabled later.
func ExactSlug(games []game.Game, ql string) []game.Game {
	for _, g := range games {
		if g.Enabled && g.Slug != "" && strings.ToLower(g.Slug) == ql {
			return []game.Game{g}
		}
	}
	return nil
}
