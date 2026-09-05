// Package parse provides small, reusable helpers for walking arbitrary JSON
// responses and coercing values, mirroring the original Python get_path /
// first_value helpers.
package parse

import (
	"strconv"
	"strings"
)

// Get returns the value at a dotted path. Numeric segments index into arrays,
// e.g. "data.rows.0.title".
func Get(data any, path string) any {
	var current = data
	for _, part := range strings.Split(path, ".") {
		switch c := current.(type) {
		case map[string]any:
			current = c[part]
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil
			}
			current = c[idx]
		default:
			return nil
		}
	}
	return current
}

// First returns the first value in the given paths that is neither nil nor
// the empty string. Missing paths simply fall through to the next candidate.
func First(data any, paths ...string) any {
	for _, p := range paths {
		if v := Get(data, p); v != nil && v != "" {
			return v
		}
	}
	return nil
}
