package workspace

import "testing"

func TestValidContext(t *testing.T) {
	for c, want := range map[string]bool{
		".":                true,
		"apps/azure-host":  true,
		"apps/azure-host/": true,
		"../elsewhere":     false,
		"..":               false,
		"/abs":             false,
		`apps\api`:         false,
		"C:/x":             false,
		"apps/../..":       false,
		"apps/./api":       false,
	} {
		if got := ValidContext(c); got != want {
			t.Errorf("ValidContext(%q) = %v", c, got)
		}
	}
}
