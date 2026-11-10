package protocol

import (
	"strings"
	"testing"
)

func TestValidateChoice(t *testing.T) {
	good := Choice{Axis: "framework", Name: "react", Language: "node", Roles: []Role{RoleScaffold, RoleToolchain}, Requires: []string{"x"}, Needs: []string{"test"}}
	if err := ValidateChoice(good); err != nil {
		t.Fatalf("a well formed choice is accepted: %v", err)
	}
	cases := map[string]struct {
		mutate func(*Choice)
		want   string
	}{
		"empty axis":    {func(c *Choice) { c.Axis = "" }, "axis"},
		"upper name":    {func(c *Choice) { c.Name = "React" }, "name"},
		"language":      {func(c *Choice) { c.Language = "" }, "language"},
		"role":          {func(c *Choice) { c.Roles = []Role{RoleDeploy} }, "scaffold or toolchain"},
		"self require":  {func(c *Choice) { c.Requires = []string{"react"} }, "itself"},
		"self conflict": {func(c *Choice) { c.Conflicts = []string{"react"} }, "itself"},
		"bad needs":     {func(c *Choice) { c.Needs = []string{"A B"} }, "axis"},
	}
	for name, tc := range cases {
		c := good
		tc.mutate(&c)
		err := ValidateChoice(c)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want an error with %q", name, err, tc.want)
		}
	}
}
