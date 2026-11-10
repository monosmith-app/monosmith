package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestAnUnreadableSmithTomlIsWrappedInTheGuideNotCutInTheHeader(t *testing.T) {
	problem := "smith.toml:68:6: toml: expected '=' after key. Fix that line, then run smith again with a long tail of words that cannot fit"
	sh := NewShell(Config{Theme: theme.Default(), Problem: problem})
	banner := sh.brokenConfigBanner(40)
	if len(banner) < 3 {
		t.Fatalf("the message wraps over several rows: %q", banner)
	}
	for _, line := range banner {
		if pane.Width(line) > 40 {
			t.Errorf("row wider than 40: %q", line)
		}
	}
	if !strings.Contains(pane.Strip(strings.Join(banner, " ")), "smith.toml:68:6") {
		t.Errorf("the banner names the line: %q", banner)
	}
}
