package scaffold

import (
	"strings"
	"testing"
)

func TestCheckPackageNameFollowsEachLanguagesRule(t *testing.T) {
	for _, c := range []struct {
		language, name string
		ok             bool
	}{
		{"node", "@pargic/core", true},
		{"node", "pargic-core", true},
		{"node", "Pargic", false},
		{"node", "_core", false},
		{"rust", "pargic", true},
		{"rust", "pargic_core-2", true},
		{"rust", "2pargic", false},
		{"rust", strings.Repeat("a", 65), false},
		{"python", "pargic.core", true},
		{"python", "-pargic", false},
		{"go", "pargic-core", true},
		{"go", "github.com/x/y", false},
		{"dotnet", "Pargic.Core", true},
		{"dotnet", "Pargic-Core", false},
		{"dart", "pargic_core", true},
		{"dart", "pargic-core", false},
		{"php", "pargic/core", true},
		{"php", "core", true},
		{"php", "Pargic/Core", false},
		{"ruby", "pargic_core", true},
		{"ruby", "pargic-core", false},
		{"jvm", "pargic-core", true},
		{"jvm", ".core", false},
		{"cpp", "pargic_core", true},
		{"cpp", "pargic-core", false},
		{"swift", "PargicCore", true},
		{"swift", "1core", false},
		{"lua", "pargic-core", true},
		{"lua", "Pargic", false},
		{"bicep", "", true},
	} {
		err := CheckPackageName(c.language, c.name)
		if (err == nil) != c.ok {
			t.Errorf("%s %q: got %v, want ok=%v", c.language, c.name, err, c.ok)
		}
	}
}

func TestCheckPackageNameNamesTheRule(t *testing.T) {
	err := CheckPackageName("rust", "2pargic")
	if err == nil || !strings.Contains(err.Error(), "Cargo crate name") || !strings.Contains(err.Error(), "starting with a letter") {
		t.Fatalf("the error must name the rule: %v", err)
	}
	err = CheckPackageName("bicep", "x")
	if err == nil || !strings.Contains(err.Error(), "does not apply to a bicep project") {
		t.Fatalf("a language without a package name must refuse: %v", err)
	}
}
