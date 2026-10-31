package extras

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestParseExtras(t *testing.T) {
	got, err := ParseExtras(map[string]any{
		"sidecars":  []any{map[string]any{"name": "videocraw", "project": "cli", "bin": "vc"}, map[string]any{"name": "x", "from": "dist/x"}},
		"resources": []any{"assets/models", map[string]any{"from": "$P/dist", "to": "plugins/"}},
	})
	if err != nil || len(got.Sidecars) != 2 || len(got.Resources) != 2 || got.Resources[1].To != "plugins/" || got.Sidecars[0].Bin != "vc" {
		t.Fatalf("%+v %v", got, err)
	}
	for name, settings := range map[string]map[string]any{
		"typo":          {"sidecar": []any{}},
		"both sources":  {"sidecars": []any{map[string]any{"name": "a", "project": "p", "from": "f"}}},
		"no source":     {"sidecars": []any{map[string]any{"name": "a"}}},
		"name with exe": {"sidecars": []any{map[string]any{"name": "a.exe", "from": "f"}}},
		"bin of a file": {"sidecars": []any{map[string]any{"name": "a", "from": "f", "bin": "b"}}},
		"empty from":    {"resources": []any{map[string]any{"to": "x"}}},
	} {
		if _, err := ParseExtras(settings); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestHostTripleOf(t *testing.T) {
	if got := hostTripleOf("rustc 1.99\nhost: aarch64-apple-darwin\nrelease: 1.99\n"); got != "aarch64-apple-darwin" {
		t.Errorf("got %q", got)
	}
	if exeSuffix("x86_64-pc-windows-msvc") != ".exe" || exeSuffix("x86_64-unknown-linux-gnu") != "" {
		t.Error("suffix")
	}
}

func TestPrepareExtras(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", `{"bundle":{"externalBin":["binaries/mine"],"resources":{"keep/":"keep/"}}}`)
	write(t, ws, "tools/bin/prog.exe", "binary")
	write(t, ws, "tools/bin/mine", "mine")
	write(t, ws, "assets/models/a.txt", "a")
	write(t, ws, "assets/models/sub/b.txt", "b")
	t.Setenv("EXTRA_DIR", filepath.Join(ws, "outside"))
	write(t, ws, "outside/p.dll", "p")
	req := PrepareRequest{
		WorkspaceRoot: ws, ProjectRoot: "apps/studio", Triple: "x86_64-pc-windows-msvc",
		Extras: Extras{
			Sidecars:  []Sidecar{{Name: "prog", From: "tools/bin/prog"}, {Name: "mine", From: "tools/bin/mine"}, {Name: "built", Project: "cli"}},
			Resources: []Resource{{From: "assets/models"}, {From: "$EXTRA_DIR", To: "plugins/"}, {From: "assets/models/*.txt", To: "flat/"}},
		},
		Program: func(s Sidecar) (string, error) { return filepath.Join(ws, "tools/bin/prog.exe"), nil },
	}
	if _, err := PrepareExtras(req); err != nil {
		t.Fatal(err)
	}
	prep := filepath.Join(ws, "dist/tauri-prepare/apps/studio")
	for _, f := range []string{
		"binaries/prog-x86_64-pc-windows-msvc.exe", "binaries/built-x86_64-pc-windows-msvc.exe",
		"resources/models/a.txt", "resources/models/sub/b.txt", "resources/plugins/p.dll", "resources/flat/a.txt",
	} {
		if _, err := os.Stat(filepath.Join(prep, f)); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	// A sidecar the developer already lists goes to the path they gave, and is not listed twice.
	if _, err := os.Stat(filepath.Join(ws, "apps/studio/src-tauri/binaries/mine-x86_64-pc-windows-msvc.exe")); err != nil {
		t.Errorf("developer's path: %v", err)
	}
	conf, _ := os.ReadFile(filepath.Join(prep, "tauri.config.json"))
	got := string(conf)
	for _, want := range []string{`"binaries/mine"`, `../../../dist/tauri-prepare/apps/studio/binaries/prog"`, `"keep/": "keep/"`, `studio/resources": "."`} {
		if !strings.Contains(got, want) {
			t.Errorf("tauri.config.json lacks %s: %s", want, got)
		}
	}
	if strings.Count(got, "binaries/mine") != 1 {
		t.Errorf("mine listed twice: %s", got)
	}
}

func TestPrepareExtrasFailsClearly(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", "{}")
	write(t, ws, "empty/.keep", "")
	_ = os.Remove(filepath.Join(ws, "empty", ".keep"))
	for name, tc := range map[string]struct {
		extras Extras
		want   string
	}{
		"unset variable": {Extras{Resources: []Resource{{From: "$NOT_SET_ANYWHERE/x"}}}, "NOT_SET_ANYWHERE"},
		"missing folder": {Extras{Resources: []Resource{{From: "nope"}}}, "nope"},
		"empty folder":   {Extras{Resources: []Resource{{From: "empty"}}}, "nothing to bundle"},
		"no match":       {Extras{Resources: []Resource{{From: "assets/*.png"}}}, "nothing matches"},
		"missing file":   {Extras{Sidecars: []Sidecar{{Name: "a", From: "dist/a"}}}, "does not exist"},
	} {
		_, err := PrepareExtras(PrepareRequest{WorkspaceRoot: ws, ProjectRoot: "apps/studio", Triple: "x86_64-unknown-linux-gnu", Extras: tc.extras})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v, want %q", name, err, tc.want)
		}
	}
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", `{"bundle":{"resources":["a/*"]}}`)
	write(t, ws, "assets/x.txt", "x")
	_, err := PrepareExtras(PrepareRequest{WorkspaceRoot: ws, ProjectRoot: "apps/studio", Triple: "t", Extras: Extras{Resources: []Resource{{From: "assets"}}}})
	if err == nil || !strings.Contains(err.Error(), "array") {
		t.Errorf("array resources: %v", err)
	}
}

func TestPickProgram(t *testing.T) {
	two := []string{"/t/a.exe", "/t/cli"}
	for name, tc := range map[string]struct {
		programs []string
		s        Sidecar
		want     string
	}{
		"bin":       {two, Sidecar{Name: "x", Bin: "a"}, "/t/a.exe"},
		"only one":  {[]string{"/t/only"}, Sidecar{Name: "x"}, "/t/only"},
		"by name":   {two, Sidecar{Name: "cli"}, "/t/cli"},
		"ambiguous": {two, Sidecar{Name: "x"}, ""},
		"no bin":    {two, Sidecar{Name: "x", Bin: "z"}, ""},
	} {
		got, err := pickProgram(tc.programs, tc.s, "crate")
		if (err == nil) != (tc.want != "") || got != tc.want {
			t.Errorf("%s: %q %v", name, got, err)
		}
	}
}

func TestExtrasDoctor(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", `{"bundle":{"externalBin":["binaries/legacy"]}}`)
	write(t, ws, "assets/ok.txt", "x")
	old := HostTriple
	HostTriple = func() (string, error) { return "x86_64-unknown-linux-gnu", nil }
	defer func() { HostTriple = old }()
	settings := map[string]any{
		"sidecars":  []any{map[string]any{"name": "a", "from": "dist/cargo/release/a"}, map[string]any{"name": "b", "from": "tools/b"}},
		"resources": []any{"assets/ok.txt", "assets/missing"},
	}
	project := protocol.Project{Name: "studio", Root: "apps/studio", Settings: map[string]map[string]any{tauriapp.Name: settings}}
	failed := map[string]bool{}
	for _, f := range Findings(ws, []protocol.Project{project}) {
		if !f.OK {
			failed[f.Check] = f.Warning
		}
	}
	want := map[string]bool{ // true: only a warning
		"the sidecar a of studio exists":                         true,
		"the sidecar b of studio exists":                         false,
		"the resource assets/missing of studio exists":           false,
		"the sidecar legacy of studio is named for this machine": true,
	}
	if len(failed) != len(want) {
		t.Fatalf("failed: %v", failed)
	}
	for check, warn := range want {
		if w, ok := failed[check]; !ok || w != warn {
			t.Errorf("%s: failed=%v warning=%v", check, ok, w)
		}
	}
	project.Settings[tauriapp.Name] = map[string]any{"sidecar": []any{}}
	if got := Findings(ws, []protocol.Project{project}); len(got) != 1 || got[0].OK {
		t.Errorf("a typo in the table must fail: %+v", got)
	}
}

func TestVerifyExtras(t *testing.T) {
	ws := t.TempDir()
	old := HostTriple
	HostTriple = func() (string, error) { return "x86_64-unknown-linux-gnu", nil }
	defer func() { HostTriple = old }()
	extras := Extras{Sidecars: []Sidecar{{Name: "vc", From: "x"}}}
	if _, err := VerifyExtras(ws, "apps/studio", extras); err == nil {
		t.Fatal("no build: want an error")
	}
	release := "dist/tauri/apps/studio/release/"
	write(t, ws, release+"vc", "x")
	if _, err := VerifyExtras(ws, "apps/studio", extras); err != nil {
		t.Fatal(err)
	}
	write(t, ws, release+"bundle/deb/studio_1_amd64/data/usr/bin/studio", "x")
	if _, err := VerifyExtras(ws, "apps/studio", extras); err == nil || !strings.Contains(err.Error(), "deb") {
		t.Fatalf("a deb without the sidecar must fail: %v", err)
	}
}

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
