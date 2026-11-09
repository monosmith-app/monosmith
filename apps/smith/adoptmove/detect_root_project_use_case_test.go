package adoptmove

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectReadsTheProjectAtTheRoot(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  Candidate
	}{
		{"node app", map[string]string{"package.json": `{"name":"@acme/WebApp","scripts":{"start":"node ."}}`}, Candidate{EcosystemNode, "package.json", "web-app", true}},
		{"node library", map[string]string{"package.json": `{"name":"kit","main":"index.js"}`}, Candidate{EcosystemNode, "package.json", "kit", false}},
		{"node bin", map[string]string{"package.json": `{"name":"cli","bin":"cli.js"}`}, Candidate{EcosystemNode, "package.json", "cli", true}},
		{"go program", map[string]string{"go.mod": "module github.com/a/Tool\n", "main.go": "package main\n"}, Candidate{EcosystemGo, "go.mod", "tool", true}},
		{"go library", map[string]string{"go.mod": "module github.com/a/lib/v3\n", "lib.go": "package lib\n"}, Candidate{EcosystemGo, "go.mod", "lib", false}},
		{"python script", map[string]string{"pyproject.toml": "[project]\nname = \"data_tools\"\n[project.scripts]\nd = \"x:y\"\n"}, Candidate{EcosystemPython, "pyproject.toml", "data-tools", true}},
		{"python package", map[string]string{"pyproject.toml": "[project]\nname = \"pk\"\n"}, Candidate{EcosystemPython, "pyproject.toml", "pk", false}},
		{"rust binary", map[string]string{"Cargo.toml": "[package]\nname = \"rr\"\n", "src/main.rs": "fn main(){}"}, Candidate{EcosystemRust, "Cargo.toml", "rr", true}},
	}
	for _, c := range cases {
		dir := t.TempDir()
		for name, text := range c.files {
			if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
				t.Fatal(err)
			}
			write(t, dir, name, text)
		}
		got, err := Detect(dir)
		if err != nil || got != c.want {
			t.Errorf("%s: got %+v, %v; want %+v", c.name, got, err, c.want)
		}
	}
}

func TestDetectFindsNothingToMove(t *testing.T) {
	cases := map[string]map[string]string{
		"empty":           {},
		"node workspaces": {"package.json": `{"name":"w","workspaces":["packages/*"]}`},
		"go.work":         {"go.mod": "module x\n", "go.work": "go 1.22\n"},
		"cargo workspace": {"Cargo.toml": "[workspace]\nmembers = [\"a\"]\n"},
		"only a pom":      {"pom.xml": "<project/>"},
	}
	for name, files := range cases {
		dir := t.TempDir()
		for f, text := range files {
			write(t, dir, f, text)
		}
		if _, err := Detect(dir); !errors.Is(err, ErrNoRootProject) {
			t.Errorf("%s: err = %v, want ErrNoRootProject", name, err)
		}
	}
}

func TestDetectRefusesTwoProjectsAtTheRoot(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module x\n")
	write(t, dir, "package.json", `{"name":"tooling"}`)
	_, err := Detect(dir)
	if err == nil || errors.Is(err, ErrNoRootProject) || !strings.Contains(err.Error(), "go.mod") || !strings.Contains(err.Error(), "package.json") {
		t.Fatalf("err = %v, want a refusal naming both manifests", err)
	}
}

func TestNeedsAnswerNamesTheExactCommands(t *testing.T) {
	msg := NeedsAnswer(Candidate{Manifest: "package.json", Name: "web", App: true}).Error()
	for _, want := range []string{"smith adopt --init --move apps/web", "--move <folder>", "--move none", "does not choose the folder"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message lacks %q:\n%s", want, msg)
		}
	}
}
