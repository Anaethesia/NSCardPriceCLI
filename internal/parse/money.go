package parse

import (
	"encoding/json"
	"strconv"
	"strings"
)

// AsMoney coerces a loosely-typed JSON value into a float64, tolerating
// thousand separators (e.g. "1,234") which the merchant APIs sometimes emit.
// It returns false for nil, booleans and non-numeric strings.
func AsMoney(v any) (float64, bool) {
	switch t := v.(type) {
	case nil:
		return 0, false
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case string:
		s := strings.TrimSpace(strings.ReplaceAll(t, ",", ""))
		if s == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case bool:
		return 0, false
	default:
		return 0, false
	}
}

// FindMoneyByKey recursively walks a decoded JSON value and returns the first
// numeric value found under a key whose lowercased name contains one of the
// given fragments. It mirrors the Python find_money_by_key helper.
func FindMoneyByKey(data any, keys ...string) (float64, bool) {
	switch c := data.(type) {
	case map[string]any:
		for key, value := range c {
			lowered := strings.ToLower(key)
			for _, frag := range keys {
				if strings.Contains(lowered, frag) {
					if m, ok := AsMoney(value); ok {
						return m, true
					}
				}
			}
			if m, ok := FindMoneyByKey(value, keys...); ok {
				return m, true
			}
		}
	case []any:
		for _, item := range c {
			if m, ok := FindMoneyByKey(item, keys...); ok {
				return m, true
			}
		}
	}
	return 0, false
}
