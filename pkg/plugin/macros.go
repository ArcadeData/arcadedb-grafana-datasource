package plugin

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

var timeFilterRegex = regexp.MustCompile(`\$__timeFilter\((\w+)\)`)

// ExpandMacros replaces Grafana macros in a query string.
func ExpandMacros(query string, timeRange backend.TimeRange, interval time.Duration) string {
	fromMs := timeRange.From.UnixMilli()
	toMs := timeRange.To.UnixMilli()

	// $__timeFrom -> epoch ms
	query = replaceAll(query, "$__timeFrom", strconv.FormatInt(fromMs, 10))

	// $__timeTo -> epoch ms
	query = replaceAll(query, "$__timeTo", strconv.FormatInt(toMs, 10))

	// $__timeFilter(column) -> column >= fromMs AND column <= toMs
	query = timeFilterRegex.ReplaceAllStringFunc(query, func(match string) string {
		submatch := timeFilterRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		col := submatch[1]
		return fmt.Sprintf("%s >= %d AND %s <= %d", col, fromMs, col, toMs)
	})

	// $__interval -> duration string
	query = replaceAll(query, "$__interval", formatDuration(interval))

	return query
}

// replaceAll replaces all occurrences of old with new in s (literal, not regex).
func replaceAll(s, old, new string) string {
	result := s
	for {
		i := indexOf(result, old)
		if i < 0 {
			break
		}
		result = result[:i] + new + result[i+len(old):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// formatDuration formats a duration as a Grafana-friendly string (e.g., "60s", "5m", "1h").
func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
