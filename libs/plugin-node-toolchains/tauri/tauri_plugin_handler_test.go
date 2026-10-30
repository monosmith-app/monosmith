package tauri

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func write(t *testing.T, ws, rel, body string) {
	t.Helper()
	p := filepath.Join(ws, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetect(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/conf/src-tauri/tauri.conf.json", "{}")
	write(t, ws, "apps/cargo/src-tauri/Cargo.toml", "[package]\nname=\"a\"\n\n[dependencies]\ntauri = { version = \"2\" }\n")
	write(t, ws, "apps/plain/src-tauri/Cargo.toml", "[package]\nname=\"a\"\n\n[dependencies]\nserde = \"1\"\n")
	write(t, ws, "apps/web/package.json", "{}")
	for root, want := range map[string]bool{"apps/conf": true, "apps/cargo": true, "apps/plain": false, "apps/web": false} {
		if got := Detect(ws, root); got != want {
			t.Errorf("%s: %v, want %v", root, got, want)
		}
	}
}

func TestContribute(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", "{}")
	res := Contribute(ws, protocol.ContributeParams{Projects: []protocol.Project{
		{Name: "studio", Root: "apps/studio", Tasks: map[string]protocol.Task{"build": {}}},
		{Name: "web", Root: "apps/web"},
	}})
	if _, ok := res.Projects["web"]; ok || len(res.Projects) != 1 {
		t.Fatalf("only the Tauri project: %+v", res.Projects)
	}
	task := res.Projects["studio"].Replace[TaskName]
	if !task.Cache || task.Outputs[0] != "../../dist/tauri/apps/studio" || strings.Join(task.DependsOn, ",") != "^build,build" {
		t.Errorf("tauri-build %+v", task)
	}
	if got := task.Command[len(task.Command)-1]; got != "{workspaceRoot}/dist/tauri/apps/studio" {
		t.Errorf("target dir must be absolute: %q", got)
	}
	if dev := res.Projects["studio"].Replace[DevTask]; dev.Cache || dev.Command[len(dev.Command)-1] != "dev" {
		t.Errorf("tauri-dev %+v", dev)
	}
}

func TestDoctor(t *testing.T) {
	realLook, realPlatform, realPkg := lookPath, platform, pkgConfigHas
	t.Cleanup(func() { lookPath, platform, pkgConfigHas = realLook, realPlatform, realPkg })
	platform = "linux"
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	pkgConfigHas = func(string) bool { return false }

	ws := t.TempDir()
	projects := []protocol.Project{{Name: "studio", Root: "studio"}}
	if got := Doctor(ws, protocol.DoctorParams{Projects: projects}); len(got.Findings) != 0 {
		t.Fatalf("no Tauri app, nothing to check: %+v", got.Findings)
	}
	write(t, ws, "studio/src-tauri/tauri.conf.json", "{}")
	res := Doctor(ws, protocol.DoctorParams{Projects: projects})
	if len(res.Findings) != 4 {
		t.Fatalf("findings %+v", res.Findings)
	}
	for _, f := range res.Findings {
		if strings.Contains(f.Check, "Cargo workspace") {
			continue // passes: the test workspace has no root Cargo.toml
		}
		if f.OK || f.Remedy == "" {
			t.Errorf("%q should fail with a remedy: %+v", f.Check, f)
		}
	}

	write(t, ws, "node_modules/@tauri-apps/cli/package.json", "{}")
	lookPath = func(string) (string, error) { return "x", nil }
	pkgConfigHas = func(string) bool { return true }
	for _, f := range Doctor(ws, protocol.DoctorParams{Projects: projects}).Findings {
		if !f.OK {
			t.Errorf("%q should pass: %+v", f.Check, f)
		}
	}
	platform = "windows"
	if got := Doctor(ws, protocol.DoctorParams{Projects: projects}); len(got.Findings) != 3 {
		t.Errorf("system libraries are a Linux check: %+v", got.Findings)
	}
}

func TestCIGivesRustAndLinuxLibsOnlyForATauriApp(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", "{}")
	if got := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}}}); len(got.Toolchains) != 0 || len(got.Setup) != 0 {
		t.Fatalf("no Tauri app, nothing to set up: %+v", got)
	}
	got := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Name: "studio", Root: "apps/studio"}}})
	if len(got.Toolchains) != 1 || got.Toolchains[0].GitHub == nil || len(got.Setup) != 1 {
		t.Fatalf("toolchain and one setup step: %+v", got)
	}
	if cmd := got.Setup[0].Command[2]; !strings.Contains(cmd, "libwebkit2gtk-4.1-dev") || !strings.Contains(cmd, "libayatana-appindicator3-dev") {
		t.Errorf("apt packages: %q", cmd)
	}
}

func TestWorkspaceMembers(t *testing.T) {
	ws := t.TempDir()
	roots := []string{"apps/studio"}
	if got := workspaceMembers(ws, roots); got != nil {
		t.Fatalf("no root Cargo.toml: %v", got)
	}
	for body, want := range map[string]int{
		"[workspace]\nmembers = [\"apps/*/src-tauri\"]\n":                                        1,
		"[workspace]\nmembers = [\"apps/*/src-tauri\"]\nexclude = [\"apps/studio/src-tauri\"]\n": 0,
		"[workspace]\nmembers = [\"libs/*\", \"apps/*\"]\n":                                      0,
		"[package]\nname = \"x\"\n":                                                              0,
	} {
		write(t, ws, "Cargo.toml", body)
		if got := workspaceMembers(ws, roots); len(got) != want {
			t.Errorf("%q: %v, want %d", body, got, want)
		}
	}
}

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "package.json", `{"name":"ws","private":true,"workspaces":["apps/*"]}`)
	write(t, ws, "apps/studio/package.json", `{"name":"studio","version":"0.0.0"}`)
	p := New("1.0.0")
	if _, err := p.Describe(protocol.InitializeParams{WorkspaceRoot: ws}); err != nil {
		t.Fatal(err)
	}
	params := protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "studio", Directory: "apps/studio", Convention: true}
	if res, err := p.Generate(params); err != nil || len(res.Files) != 0 {
		t.Fatalf("a convention adds nothing: %+v %v", res, err)
	}
	params.Convention = false
	if _, err := p.Generate(params); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(ws, "apps", "studio", "package.json"))
	if !strings.Contains(string(data), `"tauri": "tauri"`) {
		t.Errorf("no tauri script: %s", data)
	}
	root, _ := os.ReadFile(filepath.Join(ws, "package.json"))
	if !strings.Contains(string(root), "@tauri-apps/cli") {
		t.Errorf("no CLI dependency: %s", root)
	}
}
