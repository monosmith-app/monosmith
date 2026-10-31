package tauri

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/extras"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

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
	rel := res.Projects["studio"].Release
	if rel == nil || rel.Build != "smith run tauri-build -p studio" || strings.Join(rel.Platforms, ",") != "windows-x64,macos-arm64,macos-x64,linux-x64" ||
		rel.Assets[0] != "dist/tauri/apps/studio/release/bundle/**/*.{msi,exe,dmg,deb,rpm,AppImage}" || !rel.Checksums || rel.AssetName == "" {
		t.Errorf("release defaults %+v", rel)
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

func TestCIReusesTheWorkspacesPinnedRust(t *testing.T) {
	app := protocol.CIParams{Projects: []protocol.Project{{Name: "studio", Root: "apps/studio"}}}
	for name, tc := range map[string]struct{ file, body, want string }{
		"no pin":         {"", "", RustVersion},
		"toolchain file": {"rust-toolchain.toml", "[toolchain]\nchannel = \"1.99.1\"\n", "1.99.1"},
		"rust-version":   {"Cargo.toml", "[workspace.package]\nrust-version = \"1.88\"\n", "1.88"},
	} {
		ws := t.TempDir()
		write(t, ws, "apps/studio/src-tauri/tauri.conf.json", "{}")
		if tc.file != "" {
			write(t, ws, tc.file, tc.body)
		}
		got := CI(ws, app).Toolchains
		if len(got) != 1 || got[0].Name != "rust" || got[0].Version != tc.want || got[0].GitHub.Inputs[0].Value != tc.want {
			t.Errorf("%s: want one rust toolchain at %s, got %+v", name, tc.want, got)
			continue
		}
		if !strings.Contains(got[0].Install, "--default-toolchain "+tc.want+"\n") {
			t.Errorf("%s: install command: %q", name, got[0].Install)
		}
	}
}

func TestCISetupNamesItsStepInTheLog(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", "{}")
	got := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Name: "studio", Root: "apps/studio"}}})
	if !strings.Contains(got.Setup[0].Command[2], "echo 'installing the system libraries Tauri needs (apt-get)'") {
		t.Errorf("setup command: %q", got.Setup[0].Command[2])
	}
}

func TestContributeWithExtras(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", "{}")
	settings := map[string]any{
		"sidecars":  []any{map[string]any{"name": "videocraw", "project": "cli"}},
		"resources": []any{"assets/models"},
	}
	project := protocol.Project{Name: "studio", Root: "apps/studio", Settings: map[string]map[string]any{Name: settings}}
	c := Contribute(ws, protocol.ContributeParams{Projects: []protocol.Project{project}}).Projects["studio"]
	build, prepare, verify := c.Replace[TaskName], c.Replace[extras.PrepareTask], c.Replace[extras.VerifyTask]
	if !strings.Contains(strings.Join(build.Command, " "), "--config {workspaceRoot}/dist/tauri-prepare/apps/studio/tauri.config.json") {
		t.Errorf("build command: %v", build.Command)
	}
	if !build.Cache || !prepare.Cache || len(build.DependsOn) != 2 || build.DependsOn[1] != extras.PrepareTask || c.Replace[DevTask].DependsOn[0] != extras.PrepareTask {
		t.Errorf("build %+v prepare %+v", build, prepare)
	}
	if prepare.Env[SettingsEnv] == "" || verify.DependsOn[0] != TaskName || !strings.Contains(c.Release.Build, "tauri-build tauri-verify") {
		t.Errorf("verify %+v release %q", verify, c.Release.Build)
	}
	// Anything the cache key cannot see turns the cache off for the prepare and the build.
	for _, r := range []string{"$OTHER/x", "dist/out", "/abs/path", "../sibling/x"} {
		settings["resources"] = []any{r}
		c = Contribute(ws, protocol.ContributeParams{Projects: []protocol.Project{project}}).Projects["studio"]
		if c.Replace[TaskName].Cache || c.Replace[extras.PrepareTask].Cache {
			t.Errorf("%s should not be cached", r)
		}
	}
	// No settings: tasks are as before.
	plain := Contribute(ws, protocol.ContributeParams{Projects: []protocol.Project{{Name: "studio", Root: "apps/studio"}}}).Projects["studio"]
	if _, ok := plain.Replace[extras.PrepareTask]; ok || strings.Contains(strings.Join(plain.Replace[TaskName].Command, " "), "--config") || strings.Contains(plain.Release.Build, "verify") {
		t.Errorf("plain: %+v", plain)
	}
}
func TestGenerateWritesSrcTauriForAWebApp(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "package.json", `{"name":"ws","private":true}`)
	write(t, ws, "Cargo.toml", "[workspace]\nmembers = [\"apps/*/src-tauri\"]\n")
	write(t, ws, "apps/studio/package.json", `{"name":"studio","dependencies":{"react":"^19.0.0"}}`)
	write(t, ws, "apps/studio/index.html", "<div id=\"app\"></div>")
	write(t, ws, "apps/studio/src-tauri/build.rs", "// mine\n")
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	res, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "my-studio", Directory: "apps/studio", Workspace: protocol.WorkspaceInfo{Name: "Acme"}})
	if err != nil {
		t.Fatal(err)
	}
	read := func(rel string) string {
		b, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		return string(b)
	}
	conf := read("apps/studio/src-tauri/tauri.conf.json")
	for _, want := range []string{`"frontendDist": "../../../dist/apps/studio"`, `"identifier": "com.acme.my-studio"`} {
		if !strings.Contains(conf, want) {
			t.Errorf("tauri.conf.json lacks %s: %s", want, conf)
		}
	}
	if strings.Contains(conf, "beforeBuildCommand") {
		t.Errorf("smith builds the web app first: %s", conf)
	}
	if got := read("apps/studio/src-tauri/build.rs"); got != "// mine\n" {
		t.Errorf("an existing file is kept: %q", got)
	}
	if got := read("apps/studio/src-tauri/src/main.rs"); !strings.Contains(got, "my_studio_lib::run()") {
		t.Errorf("main.rs: %s", got)
	}
	if got := read("apps/studio/src-tauri/Cargo.toml"); !strings.Contains(got, "\n[workspace]\n") {
		t.Errorf("Cargo.toml needs its own workspace: %s", got)
	}
	if got := read("apps/studio/package.json"); !strings.Contains(got, "@tauri-apps/api") || !strings.Contains(got, `"react"`) {
		t.Errorf("package.json: %s", got)
	}
	if got := read("Cargo.toml"); !strings.Contains(got, `exclude = ["apps/studio/src-tauri"]`) {
		t.Errorf("root Cargo.toml: %s", got)
	}
	if !tauriapp.Detect(ws, "apps/studio") || len(res.Files) < 10 {
		t.Errorf("files %v", res.Files)
	}
	icon := read("apps/studio/src-tauri/icons/icon.ico")
	if len(icon) < 30 || icon[23:26] != "PNG" {
		t.Errorf("icon.ico is not an ICO holding a PNG")
	}
}

func TestGenerateWritesNoSrcTauriWithoutAWebPage(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "package.json", `{"name":"ws","private":true}`)
	write(t, ws, "apps/lib/package.json", `{"name":"lib"}`)
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := p.Generate(protocol.GenerateParams{Name: "lib", Directory: "apps/lib"}); err != nil {
		t.Fatal(err)
	}
	if tauriapp.Detect(ws, "apps/lib") {
		t.Error("nothing to wrap, no src-tauri")
	}
}
