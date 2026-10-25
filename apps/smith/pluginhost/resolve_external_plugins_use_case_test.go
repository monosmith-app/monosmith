package pluginhost

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestResolveExternalSaysWhereAPluginComesFromAndWhatItAsksFor(t *testing.T) {
	ws, extra, home := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("SMITH_PLUGIN_PATH", extra)
	t.Setenv("SMITH_HOME", home)
	install(t, filepath.Join(ws, ".smith", "plugins"), "local", "1.0.0", "./bin/local")
	install(t, extra, "onpath", "1.0.0", "onpath")
	install(t, filepath.Join(home, "plugins"), "homed", "1.0.0", "homed")
	manifest := filepath.Join(ws, ".smith", "plugins", "local", ManifestFile)
	if err := os.WriteFile(manifest, []byte(`{"name":"local","version":"1.0.0","command":["./bin/local"],"env":["AWS_*"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	external := map[string]workspace.ExternalPlugin{
		"local":  {Env: []string{"NPM_TOKEN"}, SHA256: "ab"},
		"onpath": {},
		"homed":  {},
		"inline": {Command: []string{"/bin/x", "--serve"}},
	}
	for name, source := range map[string]string{"local": ".smith/plugins", "onpath": "SMITH_PLUGIN_PATH", "homed": "$SMITH_HOME/plugins", "inline": "smith.toml"} {
		p, err := ResolveExternal(ws, external, name)
		if err != nil || p.Source != source {
			t.Errorf("%s: source %q, %v", name, p.Source, err)
		}
	}
	p, _ := ResolveExternal(ws, external, "local")
	if !slices.Equal(p.ManifestEnv, []string{"AWS_*"}) || !slices.Equal(p.WorkspaceEnv, []string{"NPM_TOKEN"}) || p.Pin != "ab" {
		t.Errorf("env or pin not carried: %+v", p)
	}
	if _, err := ResolveExternal(ws, external, "missing"); err == nil {
		t.Error("a plugin that is installed nowhere resolved")
	}
	all, err := ResolveAllExternal(ws, map[string]workspace.ExternalPlugin{"inline": external["inline"]})
	if err != nil || len(all) != 1 {
		t.Errorf("only declared plugins are needed: %v %v", all, err)
	}
}
