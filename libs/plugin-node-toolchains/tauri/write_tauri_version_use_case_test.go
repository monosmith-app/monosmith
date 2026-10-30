package tauri

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const confWithNestedVersion = `{
  "productName": "Studio",
  "version": "0.1.0",
  "bundle": { "windows": { "wix": { "version": "9.9.9" } } }
}
`

func TestSetConfVersionChangesOnlyTheTopLevelVersion(t *testing.T) {
	out, changed := setConfVersion(confWithNestedVersion, "1.2.3")
	if !changed || !strings.Contains(out, `"version": "1.2.3"`) || !strings.Contains(out, `"version": "9.9.9"`) {
		t.Fatalf("changed=%v\n%s", changed, out)
	}
}

func TestSetConfVersionLeavesAPathToPackageJSON(t *testing.T) {
	src := `{"version": "../package.json"}`
	if out, changed := setConfVersion(src, "1.2.3"); changed || out != src {
		t.Fatalf("changed=%v %s", changed, out)
	}
}

func TestSetCargoVersionChangesOnlyThePackage(t *testing.T) {
	src := "[package]\nname = \"a\"\nversion = \"0.0.0\"\n\n[dependencies]\nserde = { version = \"1\" }\n"
	out, changed := setCargoVersion(src, "1.2.3")
	if !changed || !strings.Contains(out, "version = \"1.2.3\"") || !strings.Contains(out, `serde = { version = "1" }`) {
		t.Fatalf("changed=%v\n%s", changed, out)
	}
	if _, changed := setCargoVersion("[package]\nname = \"a\"\nversion.workspace = true\n", "1.2.3"); changed {
		t.Error("a shared version is not ours to write")
	}
}

func versionedApp(t *testing.T, conf, cargo string) (string, protocol.Project) {
	t.Helper()
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", conf)
	write(t, ws, "apps/studio/src-tauri/Cargo.toml", cargo)
	return ws, protocol.Project{Name: "studio", Root: "apps/studio", Version: "1.0.0"}
}

func TestWriteVersionStampsCargoAndConf(t *testing.T) {
	ws, p := versionedApp(t, `{"version": "0.0.0"}`, "[package]\nname = \"a\"\nversion = \"0.0.0\"\n")
	res, err := WriteVersion(ws, protocol.VersionParams{Project: p, Version: "2.0.0"})
	if err != nil || len(res.Files) != 2 {
		t.Fatalf("%+v %v", res, err)
	}
	for file, want := range map[string]string{"Cargo.toml": `version = "2.0.0"`, "tauri.conf.json": `"version": "2.0.0"`} {
		b, _ := os.ReadFile(filepath.Join(ws, "apps/studio/src-tauri", file))
		if !strings.Contains(string(b), want) {
			t.Errorf("%s: %s", file, b)
		}
	}
}

func TestWriteVersionDryRunWritesNothing(t *testing.T) {
	ws, p := versionedApp(t, `{"version": "0.0.0"}`, "[package]\nname = \"a\"\nversion = \"0.0.0\"\n")
	res, err := WriteVersion(ws, protocol.VersionParams{Project: p, Version: "2.0.0", DryRun: true})
	if err != nil || len(res.Files) != 2 {
		t.Fatalf("%+v %v", res, err)
	}
	if b, _ := os.ReadFile(filepath.Join(ws, "apps/studio/src-tauri/Cargo.toml")); !strings.Contains(string(b), "0.0.0") {
		t.Errorf("dry run wrote %s", b)
	}
}

func TestVersionFindingsWarnOnAMismatch(t *testing.T) {
	ws, p := versionedApp(t, `{"version": "0.0.0"}`, "[package]\nname = \"a\"\nversion = \"1.0.0\"\n")
	f := versionFindings(ws, []protocol.Project{p})
	if len(f) != 1 || f[0].OK || !f[0].Warning || !strings.Contains(f[0].Detail, "tauri.conf.json 0.0.0") {
		t.Fatalf("%+v", f)
	}
	ws, p = versionedApp(t, `{"version": "../package.json"}`, "[package]\nname = \"a\"\nversion = \"1.0.0\"\n")
	if f := versionFindings(ws, []protocol.Project{p}); len(f) != 1 || !f[0].OK {
		t.Fatalf("a path to package.json agrees: %+v", f)
	}
}
