package cachecommand

import "fmt"

// formatSize writes a byte count the way people read it: 912 B, 3.4 MB, 1.2 GB.
func formatSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value, suffix := float64(n)/unit, "KB"
	for _, next := range []string{"MB", "GB", "TB"} {
		if value < unit {
			break
		}
		value /= unit
		suffix = next
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}
