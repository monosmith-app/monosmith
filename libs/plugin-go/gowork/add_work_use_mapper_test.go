package pipeline

import "testing"

func TestAddUse(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"go 1.26\n\nuse (\n\t./apps/api\n)\n", "go 1.26\n\nuse (\n\t./apps/api\n\t./libs/core\n)\n"},
		{"go 1.26\n\nuse ./apps/api\n", "go 1.26\n\nuse ./apps/api\n\nuse ./libs/core\n"},
		{"go 1.26", "go 1.26\n\nuse ./libs/core\n"},
	} {
		if got := AddUse(c.src, "libs/core"); got != c.want {
			t.Errorf("AddUse(%q) = %q, want %q", c.src, got, c.want)
		}
	}
}
