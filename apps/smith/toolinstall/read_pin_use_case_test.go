package toolinstall

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestParsePin(t *testing.T) {
	for name, tc := range map[string]struct {
		format, text string
		want         Requirement
	}{
		"nvmrc major":          {"text", "v22\n", Requirement{Prefix: "22"}},
		"nvmrc full":           {"text", "20.11.0", Requirement{Prefix: "20.11.0"}},
		"nvmrc alias":          {"text", "lts/*", Requirement{}},
		"go toolchain wins":    {"go-mod", "module m\n\ngo 1.23\n\ntoolchain go1.23.4\n", Requirement{Min: "1.23.4"}},
		"go directive":         {"go-mod", "module m\n\ngo 1.23\n", Requirement{Min: "1.23"}},
		"go without directive": {"go-mod", "module m\n", Requirement{}},
		"requires-python":      {"pyproject", "[project]\nrequires-python = \">=3.10,<4\"\n", Requirement{Min: "3.10"}},
		"requires-python none": {"pyproject", "[project]\nname = \"x\"\n", Requirement{}},
		"global.json":          {"global-json", `{"sdk": {"version": "8.0.100"}}`, Requirement{Prefix: "8.0", Min: "8.0.100"}},
		"global.json broken":   {"global-json", `{`, Requirement{}},
		"rust channel":         {"rust-toolchain", "[toolchain]\nchannel = \"1.80.0\"\n", Requirement{Min: "1.80.0"}},
		"rust stable":          {"rust-toolchain", "[toolchain]\nchannel = \"stable\"\n", Requirement{}},
		"rust plain file":      {"rust-toolchain", "1.79.0\n", Requirement{Min: "1.79.0"}},
		"unknown format":       {"yaml", "x", Requirement{}},
	} {
		if got := ParsePin(tc.format, tc.text); got != tc.want {
			t.Errorf("%s: got %+v, want %+v", name, got, tc.want)
		}
	}
}

func TestFindRequirementTakesTheFirstFileThatPins(t *testing.T) {
	root, project := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(project, ".nvmrc"), []byte("20\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pins := []protocol.ToolVersionPin{{File: ".nvmrc", Format: "text"}, {File: ".node-version", Format: "text"}}
	got, from := FindRequirement(pins, []string{root, project})
	if got != (Requirement{Prefix: "20"}) || from != ".nvmrc" {
		t.Fatalf("got %+v from %q", got, from)
	}
	if got, from := FindRequirement(pins, []string{root}); !got.Any() || from != "" {
		t.Fatalf("no file: got %+v from %q", got, from)
	}
}

func TestRequirementSatisfied(t *testing.T) {
	for name, tc := range map[string]struct {
		req       Requirement
		installed string
		want      bool
	}{
		"any":                 {Requirement{}, "1.0", true},
		"prefix match":        {Requirement{Prefix: "22"}, "22.3.0", true},
		"prefix not a prefix": {Requirement{Prefix: "22"}, "220.0.0", false},
		"min ok":              {Requirement{Min: "1.23.4"}, "1.25.1", true},
		"min older":           {Requirement{Min: "1.23.4"}, "1.23.3", false},
		"min short":           {Requirement{Min: "3.10"}, "3.9.18", false},
		"unknown version":     {Requirement{Min: "3.10"}, "", true},
		"both":                {Requirement{Prefix: "8.0", Min: "8.0.100"}, "8.0.50", false},
	} {
		if got := tc.req.Satisfied(tc.installed); got != tc.want {
			t.Errorf("%s: got %v", name, got)
		}
	}
}

func TestFillAndFindVersion(t *testing.T) {
	if got := Fill("Python.Python.{major.minor}", "3.13.1"); got != "Python.Python.3.13" {
		t.Errorf("fill: %q", got)
	}
	if got := Fill("dotnet-sdk-{major}.0", "10"); got != "dotnet-sdk-10.0" {
		t.Errorf("fill: %q", got)
	}
	if got := FindVersion("go version go1.25.1 windows/amd64"); got != "1.25.1" {
		t.Errorf("find: %q", got)
	}
}
