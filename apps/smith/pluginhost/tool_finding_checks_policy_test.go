package pluginhost

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// withoutArticle is the way smith doctor compares the name of a check with
// "<title> is installed": case and a leading "the ", "a " or "an " do not count.
func withoutArticle(check string) string {
	c := strings.ToLower(strings.TrimSpace(check))
	for _, article := range []string{"the ", "a ", "an "} {
		c = strings.TrimPrefix(c, article)
	}
	return c
}

// TestAnInstallFindingOfAPluginNamesATool keeps the rule that a doctor finding
// that says a tool is not installed carries the tools choice: smith doctor adds
// the fix to a finding whose check is "<title of a tool the plugin declares> is
// installed". The checks that stay out of this table have a written reason in
// docs/plugins/using/doctor-fixes.md (Docker is running, Ninja, Maven, ...).
func TestAnInstallFindingOfAPluginNamesATool(t *testing.T) {
	findings := map[string][]string{
		"go":     {"Go is installed"},
		"jvm":    {"a JDK is installed"},
		"dotnet": {"the .NET SDK is installed"},
		"dart":   {"the Dart SDK is installed"},
		"ruby":   {"Ruby is installed"},
		"php":    {"PHP is installed"},
		"cpp":    {"CMake is installed", "a C and C++ compiler is installed"},
		"rust":   {"Rust (cargo) is installed", "cargo-audit is installed"},
		"tauri":  {"Rust (cargo) is installed"},
	}
	for name, checks := range findings {
		ctor, ok := Builtin(name)
		if !ok {
			t.Fatalf("no built-in plugin %q", name)
		}
		info, err := ctor("test").Describe(protocol.InitializeParams{})
		if err != nil {
			t.Fatal(err)
		}
		for _, check := range checks {
			found := false
			for _, tool := range info.Tools {
				found = found || withoutArticle(tool.Title+" is installed") == withoutArticle(check)
			}
			if !found {
				t.Errorf("%s: the finding %q matches no tool the plugin declares, so smith doctor gives it no install choice", name, check)
			}
		}
	}
}
