package runhistory

import (
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// Ago says how long ago something happened in a few words: "just now",
// "40s ago", "3 min ago", "2 h ago", "4 d ago". A negative age counts as now.
func Ago(d time.Duration) string {
	switch {
	case d < 5*time.Second:
		return words.Say("tui.format_ago.just_now")
	case d < time.Minute:
		return words.Say("tui.format_ago.s_ago", "int", int(d/time.Second))
	case d < time.Hour:
		return words.Say("tui.format_ago.min_ago", "int", int(d/time.Minute))
	case d < 24*time.Hour:
		return words.Say("tui.format_ago.h_ago", "int", int(d/time.Hour))
	}
	return words.Say("tui.format_ago.d_ago", "int", int(d/(24*time.Hour)))
}
