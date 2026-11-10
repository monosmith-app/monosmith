package choiceselect

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestLintAndTestAloneUseTheStacksOwnScaffold(t *testing.T) {
	sel, err := Solve(webOffers(), []string{"eslint", "jest"})
	if err != nil {
		t.Fatalf("lint and test apply on top of the stack's own project: %v", err)
	}
	st := sel.Stack("eslint-jest", "")
	if st.Scaffold != "" || st.Language != "node" || !slices.Equal(st.Toolchain, []string{"eslint", "jest"}) {
		t.Errorf("stack %+v: no scaffold choice, so the language plugin writes the project", st)
	}
}

func TestAChoiceThatShapesTheProjectStillNeedsOneThatWritesIt(t *testing.T) {
	tauri := offer("tauri", "shell", "tauri", func(o *Offer) { o.Layer, o.Roles = "shell", []protocol.Role{protocol.RoleToolchain} })
	for name, names := range map[string][]string{
		"style":      {"eslint", "bootstrap"},
		"library":    {"jest", "d3"},
		"desktop":    {"eslint", "tauri"},
		"fan of two": {"eslint", "vitest", "bootstrap"},
	} {
		_, err := Solve(append(webOffers(), tauri), names)
		if err == nil || !strings.Contains(err.Error(), "writes the project") && !strings.Contains(err.Error(), "provided by") {
			t.Errorf("%s: %v", name, err)
		}
	}
	_, err := Solve(append(webOffers(), tauri), []string{"eslint", "tauri"})
	if err == nil || !strings.Contains(err.Error(), "no choice in eslint, tauri writes the project") || !strings.Contains(err.Error(), "or give no --with") {
		t.Errorf("the refusal says what to do: %v", err)
	}
}
