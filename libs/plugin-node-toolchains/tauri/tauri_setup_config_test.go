package tauri

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

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
	if !Detect(ws, "apps/studio") || len(res.Files) < 10 {
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
	if Detect(ws, "apps/lib") {
		t.Error("nothing to wrap, no src-tauri")
	}
}
