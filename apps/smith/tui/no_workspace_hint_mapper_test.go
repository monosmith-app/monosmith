package tui

import (
	"strings"
	"testing"
)

func TestNoWorkspaceHintNamesTheOneWayIn(t *testing.T) {
	got := Shell{}.noWorkspaceHint()
	if !strings.Contains(got, "smith adopt") || strings.Contains(got, "convert") {
		t.Errorf("hint = %q", got)
	}
}
