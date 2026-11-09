package adoptmove

import "testing"

func TestDeriveName(t *testing.T) {
	cases := []struct{ raw, fallback, want string }{
		{"@acme/WebApp", "x", "web-app"},
		{"my_package", "x", "my-package"},
		{"github.com/acme/Billing-Tool", "x", "billing-tool"},
		{"github.com/acme/billing/v2", "x", "billing"},
		{"Data.Tools", "x", "data-tools"},
		{"", "My Repo", "my-repo"},
		{"123", "api", "api"},
		{"@@@", "", "project"},
		{"", "", "project"},
	}
	for _, c := range cases {
		if got := DeriveName(c.raw, c.fallback); got != c.want {
			t.Errorf("DeriveName(%q, %q) = %q, want %q", c.raw, c.fallback, got, c.want)
		}
	}
}
