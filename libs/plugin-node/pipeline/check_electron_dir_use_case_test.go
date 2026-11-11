package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCheckElectronDir(t *testing.T) {
	p := protocol.Project{Name: "desk", Root: "apps/desk"}
	cases := map[string]struct {
		manifest, forgeConfig string
		applies, ok, fix      bool
	}{
		"not electron":   {manifest: `{"devDependencies":{"vite":"^6"}}`},
		"forge in out/":  {manifest: `{"devDependencies":{"@electron-forge/cli":"^7"}}`, forgeConfig: "module.exports = {}\n", applies: true, fix: true},
		"forge moved":    {manifest: `{"devDependencies":{"@electron-forge/cli":"^7"},"config":{"forge":{"outDir":"../../dist/electron/apps/desk/out"}}}`, applies: true, ok: true},
		"forge computed": {manifest: `{"devDependencies":{"@electron-forge/cli":"^7"}}`, forgeConfig: "module.exports = { outDir: join(a) }\n", applies: true},
		"builder dist/":  {manifest: `{"devDependencies":{"electron-builder":"^25"}}`, applies: true, fix: true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ws := t.TempDir()
			dir := filepath.Join(ws, "apps", "desk")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(c.manifest), 0o644)
			if c.forgeConfig != "" {
				_ = os.WriteFile(filepath.Join(dir, "forge.config.js"), []byte(c.forgeConfig), 0o644)
			}
			f, applies := CheckElectronDir(ws, p)
			if applies != c.applies {
				t.Fatalf("applies %v: %+v", applies, f)
			}
			if !applies {
				return
			}
			if f.OK != c.ok || !f.Warning || (f.Fix != nil) != c.fix {
				t.Errorf("%+v", f)
			}
			if !c.ok && !c.fix && !strings.Contains(f.Remedy, "by hand") && !strings.Contains(f.Problem, "cannot read") {
				t.Errorf("without a fix the remedy says what to do: %+v", f)
			}
		})
	}
}
