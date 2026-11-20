package workspace

import (
	"strings"
	"testing"
)

func TestNewerSmithSummaryNamesSmithUpdate(t *testing.T) {
	got := (&NewerSmithError{}).Summary("2.55.0")
	if want := "This smith (2.55.0) is older than smith.toml needs. Run `smith update`."; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if strings.Contains(got, "npm install") {
		t.Errorf("the launcher keeps its version on reinstall, so npm install is not the update: %s", got)
	}
}
