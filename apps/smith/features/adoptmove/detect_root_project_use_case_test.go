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

// testClaims are the manifests the language plugins claim in their handshake.
var testClaims = []Claim{
	{Language: "node", Manifests: []string{"package.json"}},
	{Language: "go", Manifests: []string{"go.mod"}},
	{Language: "python", Manifests: []string{"pyproject.toml"}},
	{Language: "rust", Manifests: []string{"Cargo.toml"}},
	{Language: "php", Manifests: []string{"composer.json"}},
	{Language: "dotnet", Manifests: []string{"*.csproj", "*.fsproj"}},
	{Language: "jvm", Manifests: []string{"pom.xml", "build.gradle.kts", "build.gradle"}},
	{Language: "dart", Manifests: []string{"pubspec.yaml"}},
}

func TestDetectReadsTheProjectAtTheRoot(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  Candidate
	}{
		{"node app", map[string]string{"package.json": `{"name":"@acme/WebApp","scripts":{"start":"node ."}}`}, Candidate{EcosystemNode, "package.json", "web-app", true, true}},
		{"node library", map[string]string{"package.json": `{"name":"kit","main":"index.js"}`}, Candidate{EcosystemNode, "package.json", "kit", false, true}},
		{"node private", map[string]string{"package.json": `{"name":"site","private":true}`}, Candidate{EcosystemNode, "package.json", "site", false, false}},
		{"node bin", map[string]string{"package.json": `{"name":"cli","bin":"cli.js"}`}, Candidate{EcosystemNode, "package.json", "cli", true, true}},
		{"go program", map[string]string{"go.mod": "module github.com/a/Tool\n", "main.go": "package main\n"}, Candidate{EcosystemGo, "go.mod", "tool", true, false}},
		{"go library", map[string]string{"go.mod": "module github.com/a/lib/v3\n", "lib.go": "package lib\n"}, Candidate{EcosystemGo, "go.mod", "lib", false, false}},
		{"python script", map[string]string{"pyproject.toml": "[project]\nname = \"data_tools\"\n[project.scripts]\nd = \"x:y\"\n"}, Candidate{EcosystemPython, "pyproject.toml", "data-tools", true, false}},
		{"python package", map[string]string{"pyproject.toml": "[project]\nname = \"pk\"\n"}, Candidate{EcosystemPython, "pyproject.toml", "pk", false, false}},
		{"rust binary", map[string]string{"Cargo.toml": "[package]\nname = \"rr\"\n", "src/main.rs": "fn main(){}"}, Candidate{EcosystemRust, "Cargo.toml", "rr", true, false}},
		{"composer library", map[string]string{"composer.json": `{"name":"acme/ColorValidator","type":"library"}`}, Candidate{"php", "composer.json", "color-validator", false, false}},
		{"composer project", map[string]string{"composer.json": `{"name":"acme/site","type":"project"}`}, Candidate{"php", "composer.json", "site", true, false}},
		{"dotnet web", map[string]string{"Api.csproj": `<Project Sdk="Microsoft.NET.Sdk.Web"></Project>`}, Candidate{"dotnet", "Api.csproj", "api", true, false}},
		{"dotnet console", map[string]string{"Tool.fsproj": "<Project Sdk=\"Microsoft.NET.Sdk\"><PropertyGroup><OutputType>Exe</OutputType></PropertyGroup></Project>"}, Candidate{"dotnet", "Tool.fsproj", "tool", true, false}},
		{"dotnet library", map[string]string{"Core.Lib.csproj": `<Project Sdk="Microsoft.NET.Sdk"></Project>`}, Candidate{"dotnet", "Core.Lib.csproj", "core-lib", false, false}},
		{"dart package", map[string]string{"pubspec.yaml": "name: my_widgets\n"}, Candidate{"dart", "pubspec.yaml", "my-widgets", false, false}},
	}
	for _, c := range cases {
		dir := t.TempDir()
		for name, text := range c.files {
			if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
				t.Fatal(err)
			}
			write(t, dir, name, text)
		}
		got, err := Detect(dir, testClaims)
		if err != nil || got != c.want {
			t.Errorf("%s: got %+v, %v; want %+v", c.name, got, err, c.want)
		}
	}
}

func TestDetectFindsNothingToMove(t *testing.T) {
	cases := map[string]map[string]string{
		"empty":               {},
		"node workspaces":     {"package.json": `{"name":"w","workspaces":["packages/*"]}`},
		"go.work":             {"go.mod": "module x\n", "go.work": "go 1.22\n"},
		"cargo workspace":     {"Cargo.toml": "[workspace]\nmembers = [\"a\"]\n"},
		"no plugin claims it": {"build.xml": "<project/>"},
	}
	for name, files := range cases {
		dir := t.TempDir()
		for f, text := range files {
			write(t, dir, f, text)
		}
		if _, err := Detect(dir, testClaims); !errors.Is(err, ErrNoRootProject) {
			t.Errorf("%s: err = %v, want ErrNoRootProject", name, err)
		}
	}
}

func TestDetectRefusesTwoProjectsAtTheRoot(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module x\n")
	write(t, dir, "package.json", `{"name":"tooling"}`)
	_, err := Detect(dir, testClaims)
	if err == nil || errors.Is(err, ErrNoRootProject) || !strings.Contains(err.Error(), "go.mod") || !strings.Contains(err.Error(), "package.json") {
		t.Fatalf("err = %v, want a refusal naming both manifests", err)
	}
}

func TestDetectRefusesTwoLanguagesAtTheRoot(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pom.xml", "<project/>")
	write(t, dir, "package.json", `{"name":"tooling"}`)
	_, err := Detect(dir, testClaims)
	if err == nil || errors.Is(err, ErrNoRootProject) || !strings.Contains(err.Error(), "pom.xml") || !strings.Contains(err.Error(), "package.json") {
		t.Fatalf("err = %v, want a refusal naming both manifests", err)
	}
}

func TestDetectNamesAProjectByItsFolderWhenTheManifestHasNoName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "LegacyShop")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "pom.xml", "<project/>")
	got, err := Detect(dir, testClaims)
	if err != nil || got != (Candidate{"jvm", "pom.xml", "legacy-shop", false, false}) {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestDetectKnowsOnlyWhatThePluginsClaim(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "composer.json", `{"name":"a/b"}`)
	if _, err := Detect(dir, testClaims[:4]); !errors.Is(err, ErrNoRootProject) {
		t.Fatalf("err = %v, want ErrNoRootProject without the PHP plugin", err)
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
