package migrator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func humanizeDuration(duration time.Duration, zeroValue string) string {
	if duration == 0 {
		return zeroValue
	}

	const day = 24 * time.Hour

	days := duration / day
	duration %= day
	hours := duration / time.Hour
	duration %= time.Hour
	mins := duration / time.Minute
	duration %= time.Minute
	secs := duration / time.Second
	duration %= time.Second
	millis := duration / time.Millisecond

	var parts []string

	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if mins > 0 {
		parts = append(parts, fmt.Sprintf("%dm", mins))
	}

	if secs > 0 {
		parts = append(parts, fmt.Sprintf("%ds", secs))
	}

	if millis > 0 {
		parts = append(parts, fmt.Sprintf("%dms", millis))
	}

	return strings.Join(parts, " ")
}

func Sha256ToHexStr(data string) string {
	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}

func mapKeysToSlice[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
