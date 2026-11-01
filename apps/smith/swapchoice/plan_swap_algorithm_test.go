package swapchoice

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func offer(axis, name string, roles ...protocol.Role) choiceselect.Offer {
	return choiceselect.Offer{Choice: protocol.Choice{Axis: axis, Name: name, Language: "node", Roles: roles}, Plugin: "node"}
}

func offers() []choiceselect.Offer {
	return []choiceselect.Offer{
		offer("framework", "react", protocol.RoleScaffold),
		offer("test", "jest", protocol.RoleToolchain),
		offer("test", "vitest", protocol.RoleToolchain),
		offer("style", "bootstrap"),
	}
}

func TestPlanSwapReplacesTheChoiceOnTheAxis(t *testing.T) {
	s, err := PlanSwap(offers(), []string{"react", "jest"}, "test", "vitest")
	if err != nil {
		t.Fatal(err)
	}
	if s.From != "jest" || s.To != "vitest" || strings.Join(s.Added, ",") != "vitest" || strings.Join(s.Removed, ",") != "jest" {
		t.Fatalf("swap %+v", s)
	}
}

func TestPlanSwapRefusesWithASentenceThatSaysWhatToDo(t *testing.T) {
	for name, c := range map[string]struct {
		current  []string
		axis, to string
		want     string
	}{
		"no choice on the axis": {[]string{"react"}, "test", "vitest", "no choice on the test axis"},
		"unknown choice":        {[]string{"react", "jest"}, "test", "mocha", "no choice is named"},
		"other axis":            {[]string{"react", "jest"}, "test", "bootstrap", "not on the test axis"},
		"same choice":           {[]string{"react", "jest"}, "test", "jest", "already uses jest"},
	} {
		if _, err := PlanSwap(offers(), c.current, c.axis, c.to); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error = %v, want %q", name, err, c.want)
		}
	}
}
