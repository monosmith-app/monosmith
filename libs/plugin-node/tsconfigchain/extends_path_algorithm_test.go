package tsconfigchain

import "testing"

func TestExtendsFollowsTheDepthOfTheProject(t *testing.T) {
	for _, c := range []struct{ dir, want string }{
		{"apps/api", "../../tsconfig.base.json"},
		{"apps/web/", "../../tsconfig.base.json"},
		{"apps/a/b", "../../../tsconfig.base.json"},
		{"api", "../tsconfig.base.json"},
		{"", "./tsconfig.base.json"},
	} {
		if got := Extends(c.dir, Base); got != c.want {
			t.Errorf("Extends(%q) = %q, want %q", c.dir, got, c.want)
		}
	}
}
