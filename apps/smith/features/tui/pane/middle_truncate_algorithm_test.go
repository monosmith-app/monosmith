package pane

import (
	"strings"
	"testing"
)

func TestMiddleTruncateKeepsBothEndsOfAPath(t *testing.T) {
	got := MiddleTruncate("apps/markobot-vscode-extension/src/index.ts", 20)
	if Width(got) != 20 || !strings.HasPrefix(got, "apps/") || !strings.HasSuffix(got, "index.ts") {
		t.Fatalf("got %q (%d cells)", got, Width(got))
	}
}

func TestMiddleTruncateLeavesAShortStringAlone(t *testing.T) {
	if got := MiddleTruncate("apps/web", 20); got != "apps/web" {
		t.Fatal(got)
	}
}

func TestMiddleTruncateHandlesTinyWidths(t *testing.T) {
	for w := -1; w <= 3; w++ {
		if Width(MiddleTruncate("apps/markobot", w)) > max(w, 0) {
			t.Fatalf("width %d overflowed", w)
		}
	}
}
