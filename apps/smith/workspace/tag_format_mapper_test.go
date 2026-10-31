package workspace

import "testing"

func TestExpandTagFormat(t *testing.T) {
	cases := []struct{ format, scope, name, want string }{
		{"", "@acme", "kit", "kit@"},
		{"{name}@", "@acme", "kit", "kit@"},
		{"{scope}/{name}@", "@acme", "kit", "@acme/kit@"},
		{"{name}-v", "", "kit", "kit-v"},
		{"release/{name}/", "@acme", "kit", "release/kit/"},
	}
	for _, c := range cases {
		if got := ExpandTagFormat(c.format, c.scope, c.name); got != c.want {
			t.Errorf("%q with scope %q, name %q: got %q, want %q", c.format, c.scope, c.name, got, c.want)
		}
	}
}
