package pluginhost

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func install(t *testing.T, dir, name, version, command string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"name": "` + name + `", "version": "` + version + `", "command": ["` + command + `", "--serve"]}`
	if err := os.WriteFile(filepath.Join(p, ManifestFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindInstalledFollowsTheSearchPath(t *testing.T) {
	ws, extra, home := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("SMITH_PLUGIN_PATH", extra)
	t.Setenv("SMITH_HOME", home)
	local := filepath.Join(ws, ".smith", "plugins")
	install(t, local, "kiwi", "2.0.0", "./bin/kiwi-{os}-{arch}{exe}")
	install(t, extra, "kiwi", "1.0.0", "kiwi")
	install(t, extra, "melon", "1.0.0", "smith-plugin-melon")
	install(t, filepath.Join(home, "plugins"), "plum", "0.1.0", "../shared/plum")
	_ = os.MkdirAll(filepath.Join(local, "not-a-plugin"), 0o755)

	if got := SearchPath(ws); len(got) != 3 || got[0] != local || got[1] != extra || got[2] != filepath.Join(home, "plugins") {
		t.Fatalf("search path %v", got)
	}
	all, err := FindInstalled(ws)
	if err != nil || len(all) != 3 {
		t.Fatalf("%v %+v", err, all)
	}
	kiwi, _ := FindPlugin(ws, "kiwi")
	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	want := filepath.Join(local, "kiwi", "bin", "kiwi-"+runtime.GOOS+"-"+runtime.GOARCH+exe)
	if kiwi.Version != "2.0.0" || kiwi.Argv[0] != want || kiwi.Argv[1] != "--serve" {
		t.Fatalf("the workspace's kiwi wins, its command resolved: %+v", kiwi)
	}
	if melon, _ := FindPlugin(ws, "melon"); melon.Argv[0] != "smith-plugin-melon" {
		t.Fatalf("a bare command is looked up on PATH: %+v", melon)
	}
	if plum, _ := FindPlugin(ws, "plum"); plum.Argv[0] != filepath.Join(home, "plugins", "shared", "plum") {
		t.Fatalf("../ is relative to the plugin's directory: %+v", plum)
	}
	if _, err := FindPlugin(ws, "fig"); err == nil || !strings.Contains(err.Error(), "fig/plugin.json") {
		t.Fatalf("missing: %v", err)
	}
}

func TestFindInstalledReportsABadManifestAsAProblem(t *testing.T) {
	ws := t.TempDir()
	t.Setenv("SMITH_PLUGIN_PATH", "")
	t.Setenv("SMITH_HOME", t.TempDir())
	dir := filepath.Join(ws, ".smith", "plugins", "bad")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, ManifestFile), []byte(`{"name": "bad"}`), 0o644)
	all, problems, err := FindInstalledReport(ws)
	if err != nil || len(all) != 0 || len(problems) != 1 || !strings.Contains(problems[0].Problem, "command are required") {
		t.Fatalf("%v %+v %+v", err, all, problems)
	}
}

func TestSatisfies(t *testing.T) {
	for _, c := range []struct {
		version, constraint string
		want                bool
	}{
		{"1.2.3", "", true}, {"1.2.3", "*", true},
		{"1.2.3", "1.2.3", true}, {"1.2.4", "1.2.3", false}, {"v1.2.3", "=1.2.3", true},
		{"1.4.0", "^1.2", true}, {"2.0.0", "^1.2", false}, {"1.1.9", "^1.2", false},
		{"1.2.9", "~1.2", true}, {"1.3.0", "~1.2", false},
		{"3.0.0-rc.1", ">=2", true}, {"1.9.9", ">=2", false},
	} {
		got, err := Satisfies(c.version, c.constraint)
		if err != nil || got != c.want {
			t.Errorf("Satisfies(%q, %q) = %v, %v; want %v", c.version, c.constraint, got, err, c.want)
		}
	}
	if _, err := Satisfies("1.0.0", "^banana"); err == nil {
		t.Fatal("a bad constraint is an error")
	}
}

func TestOneInvalidManifestDoesNotHideTheOtherPlugins(t *testing.T) {
	ws, home := t.TempDir(), t.TempDir()
	t.Setenv("SMITH_PLUGIN_PATH", "")
	t.Setenv("SMITH_HOME", home)
	local := filepath.Join(ws, ".smith", "plugins")
	install(t, local, "kiwi", "2.0.0", "kiwi")
	if err := os.MkdirAll(filepath.Join(local, "bad"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local, "bad", ManifestFile), []byte("{ nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	all, problems, err := FindInstalledReport(ws)
	if err != nil || len(all) != 1 || all[0].Name != "kiwi" {
		t.Fatalf("%v %+v", err, all)
	}
	if len(problems) != 1 || problems[0].Dir != filepath.Join(local, "bad") || !strings.Contains(problems[0].Problem, ManifestFile) {
		t.Fatalf("problems %+v", problems)
	}
	if kiwi, err := FindPlugin(ws, "kiwi"); err != nil || kiwi.Version != "2.0.0" {
		t.Fatalf("a plugin beside the broken one still starts: %v %+v", err, kiwi)
	}
	if _, err := FindPlugin(ws, "bad"); err == nil || !strings.Contains(err.Error(), "skipped") {
		t.Fatalf("the error for the broken one says why: %v", err)
	}
}
