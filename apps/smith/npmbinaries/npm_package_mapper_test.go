package npmbinaries

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformPackageSpellsNodePlatforms(t *testing.T) {
	cases := map[[2]string]string{
		{"windows", "amd64"}: "@acme/cli-win32-x64",
		{"darwin", "arm64"}:  "@acme/cli-darwin-arm64",
		{"linux", "amd64"}:   "@acme/cli-linux-x64",
	}
	for in, want := range cases {
		if got, ok := PlatformPackage("@acme/cli", in[0], in[1]); !ok || got != want {
			t.Errorf("PlatformPackage(%v) = %q, %v; want %q", in, got, ok, want)
		}
	}
	if _, ok := PlatformPackage("cli", "plan9", "amd64"); ok {
		t.Error("plan9 has no npm platform")
	}
}

func TestPlatformManifestLimitsInstallToItsPlatform(t *testing.T) {
	r := Release{Names: []string{"@acme/cli"}, Version: "5.0.0", License: "MIT", Repository: "https://github.com/acme/cli"}
	var m map[string]any
	if err := json.Unmarshal([]byte(PlatformManifest(r, "@acme/cli-win32-x64", "windows", "amd64")), &m); err != nil {
		t.Fatal(err)
	}
	if m["os"].([]any)[0] != "win32" || m["cpu"].([]any)[0] != "x64" {
		t.Errorf("os/cpu = %v/%v", m["os"], m["cpu"])
	}
	if m["repository"].(map[string]any)["url"] != "git+https://github.com/acme/cli.git" {
		t.Errorf("repository = %v", m["repository"])
	}
}

func TestWrapperManifestDependsOnEveryPlatformAtItsVersion(t *testing.T) {
	r := Release{Names: []string{"@acme/cli", "acme"}, Binary: "acme", Version: "5.0.0"}
	platforms := map[string]string{"linux-x64": "@acme/cli-linux-x64", "win32-x64": "@acme/cli-win32-x64"}
	var m struct {
		Name     string
		Bin      map[string]string
		Optional map[string]string `json:"optionalDependencies"`
	}
	if err := json.Unmarshal([]byte(WrapperManifest(r, "acme", platforms)), &m); err != nil {
		t.Fatal(err)
	}
	if m.Name != "acme" || m.Bin["acme"] != "bin/acme.js" {
		t.Errorf("name %q bin %v", m.Name, m.Bin)
	}
	if len(m.Optional) != 2 || m.Optional["@acme/cli-linux-x64"] != "5.0.0" {
		t.Errorf("optionalDependencies = %v", m.Optional)
	}
}

func TestShimResolvesThePlatformPackage(t *testing.T) {
	shim := Shim("acme", map[string]string{"linux-x64": "@acme/cli-linux-x64"})
	for _, want := range []string{"#!/usr/bin/env node", `"linux-x64": "@acme/cli-linux-x64"`, "const NAME = 'acme'", "NAME + '.exe' : NAME", "require.resolve(pkg + '/bin/' + file)", "is not on the npm registry yet", "Install it now?", "SMITH_AUTO_FIX"} {
		if !strings.Contains(shim, want) {
			t.Errorf("shim lacks %q:\n%s", want, shim)
		}
	}
}

// TestTheLauncherLogicPassesItsNodeTests runs the node tests of the launcher's
// choices (fake npm, no clock, no terminal). It is skipped where node is not installed.
func TestTheLauncherLogicPassesItsNodeTests(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	out, err := exec.Command(node, "--test", "shim_template.test.js").CombinedOutput()
	if err != nil {
		t.Fatalf("node --test failed:\n%s", strings.TrimSpace(string(out)))
	}
	// The shim that ships is the template with its names filled in: it must parse.
	file := filepath.Join(t.TempDir(), "acme.js")
	if err := os.WriteFile(file, []byte(Shim("acme", map[string]string{"linux-x64": "@acme/cli-linux-x64"})), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(node, "--check", file).CombinedOutput(); err != nil {
		t.Fatalf("the shim does not parse:\n%s", out)
	}
}
