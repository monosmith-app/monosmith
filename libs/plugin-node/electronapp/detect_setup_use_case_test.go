package electronapp

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFiles writes files (path relative to dir) for a test.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetect(t *testing.T) {
	const forge = `{"devDependencies":{"@electron-forge/cli":"^7"}}`
	const builder = `{"devDependencies":{"electron-builder":"^25"}}`
	cases := map[string]struct {
		files map[string]string
		want  Setup
	}{
		"not electron":      {map[string]string{"package.json": `{"devDependencies":{"vite":"^6"}}`}, Setup{}},
		"smith-made":        {map[string]string{"package.json": forge, "scripts/assemble.mjs": ""}, Setup{Tool: SmithForge}},
		"forge in manifest": {map[string]string{"package.json": `{"devDependencies":{"@electron-forge/cli":"^7"},"config":{"forge":{"outDir":"build/out"}}}`}, Setup{Tool: Forge, OutDir: "build/out", Config: "package.json", Editable: true}},
		"forge file, default": {map[string]string{"package.json": forge, "forge.config.js": "module.exports = {\n  packagerConfig: {},\n}\n"},
			Setup{Tool: Forge, OutDir: "out", Config: "forge.config.js", Editable: true}},
		"forge file, outDir": {map[string]string{"package.json": forge, "forge.config.ts": "const config: ForgeConfig = { outDir: '../../dist/electron/apps/desk/out' }\n"},
			Setup{Tool: Forge, OutDir: "../../dist/electron/apps/desk/out", Config: "forge.config.ts", Editable: true}},
		"forge file, computed": {map[string]string{"package.json": forge, "forge.config.js": "module.exports = { outDir: path.join(a, 'b') }\n"},
			Setup{Tool: Forge, Config: "forge.config.js"}},
		"no forge config": {map[string]string{"package.json": forge}, Setup{Tool: Forge, OutDir: "out"}},
		"builder in manifest": {map[string]string{"package.json": `{"devDependencies":{"electron-builder":"^25"},"build":{"directories":{"output":"release"}}}`},
			Setup{Tool: Builder, OutDir: "release", Config: "package.json", Editable: true}},
		"builder yml": {map[string]string{"package.json": builder, "electron-builder.yml": "appId: x\ndirectories:\n  output: release\n"},
			Setup{Tool: Builder, OutDir: "release", Config: "electron-builder.yml"}},
		"builder yml, default": {map[string]string{"package.json": builder, "electron-builder.yml": "appId: x\n"},
			Setup{Tool: Builder, OutDir: "dist", Config: "electron-builder.yml", Editable: true}},
		"builder macro": {map[string]string{"package.json": builder, "electron-builder.json": `{"directories":{"output":"dist/${os}"}}`},
			Setup{Tool: Builder, Config: "electron-builder.json", Editable: true}},
		"builder js": {map[string]string{"package.json": builder, "electron-builder.config.js": "module.exports = {}"},
			Setup{Tool: Builder, Config: "electron-builder.config.js"}},
		"builder, no config": {map[string]string{"package.json": builder}, Setup{Tool: Builder, OutDir: "dist", Config: "package.json", Editable: true}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, c.files)
			if got := Detect(dir); got != c.want {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
		})
	}
}
