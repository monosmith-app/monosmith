package workspace

import (
	"strings"
	"testing"
)

func TestNewerSmithSummaryNamesSmithUpdateAndThePin(t *testing.T) {
	got := (&NewerSmithError{}).Summary("2.55.0")
	for _, want := range []string{"This smith (2.55.0) is older than smith.toml needs", "`smith update`", "[workspace] smith"} {
		if !strings.Contains(got, want) {
			t.Errorf("summary lacks %q: %s", want, got)
		}
	}
	if strings.Contains(got, "npm install") {
		t.Errorf("the launcher keeps its version on reinstall, so npm install is not the update: %s", got)
	}
}
