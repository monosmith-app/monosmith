package adoptsplit

import "testing"

func TestKeepsOwnVersion(t *testing.T) {
	cases := []struct {
		root, project string
		want          bool
	}{
		{"^4.9.5", "^5.9.3", true},
		{"^5.0.0", "~5.9.3", false},
		{"", "^5.9.3", false},
		{">=18", "^20.1.0", true},
		{"latest", "^5.0.0", false},
		{"^1.6.1", "1.x", false},
	}
	for _, c := range cases {
		if got := KeepsOwnVersion(c.root, c.project); got != c.want {
			t.Errorf("KeepsOwnVersion(%q, %q) = %v, want %v", c.root, c.project, got, c.want)
		}
	}
}
