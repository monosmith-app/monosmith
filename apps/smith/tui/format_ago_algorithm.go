package tui

import (
	"fmt"
	"time"
)

// agoText says how long ago something happened in a few words: "just now",
// "40s ago", "3 min ago", "2 h ago", "4 d ago". A negative age counts as now.
func agoText(d time.Duration) string {
	switch {
	case d < 5*time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d/time.Second))
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d/time.Minute))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d h ago", int(d/time.Hour))
	}
	return fmt.Sprintf("%d d ago", int(d/(24*time.Hour)))
}
