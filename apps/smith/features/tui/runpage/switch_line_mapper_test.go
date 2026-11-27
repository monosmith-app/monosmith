package runpage

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func TestTheSwitchNamesBothPanesTheTaskOfTheLogAndItsSpans(t *testing.T) {
	line, spans := Switch(plain, "web:test", true, 78)
	text := pane.Strip(line)
	if !strings.Contains(text, "Tasks") || !strings.Contains(text, "Log: web:test") || !strings.HasSuffix(strings.TrimRight(text, " "), "tab switches") {
		t.Fatalf("switch %q", text)
	}
	if len(spans) != 2 || spans[0].ID != TasksID || spans[1].ID != LogID {
		t.Fatalf("spans %+v", spans)
	}
	runes := []rune(text)
	if got := strings.TrimSpace(string(runes[spans[1].From:spans[1].To])); got != "Log: web:test" {
		t.Fatalf("the log span covers %q", got)
	}
}
