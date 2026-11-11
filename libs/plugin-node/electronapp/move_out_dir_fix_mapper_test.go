package electronapp

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestMoveOutDirFix(t *testing.T) {
	p := protocol.Project{Name: "desk", Root: "apps/desk"}
	cases := map[string]struct {
		files map[string]string
		want  string // in the new content; "" means no fix
		file  string
	}{
		"forge in manifest": {map[string]string{"package.json": `{"devDependencies":{"@electron-forge/cli":"^7"},"config":{"forge":{"packagerConfig":{}}}}`},
			`"outDir": "../../dist/electron/apps/desk/out"`, "apps/desk/package.json"},
		"forge file, add": {map[string]string{"package.json": `{"devDependencies":{"@electron-forge/cli":"^7"}}`, "forge.config.js": "module.exports = {\n  packagerConfig: {},\n}\n"},
			"module.exports = {\n  outDir: '../../dist/electron/apps/desk/out',\n  packagerConfig", "apps/desk/forge.config.js"},
		"forge file, replace": {map[string]string{"package.json": `{"devDependencies":{"@electron-forge/cli":"^7"}}`, "forge.config.js": "module.exports = { outDir: 'build' }\n"},
			"module.exports = { outDir: '../../dist/electron/apps/desk/out' }\n", "apps/desk/forge.config.js"},
		"builder in manifest": {map[string]string{"package.json": `{"devDependencies":{"electron-builder":"^25"}}`},
			`"build": {` + "\n" + `    "directories": {` + "\n" + `      "output": "../../dist/electron/apps/desk"`, "apps/desk/package.json"},
		"builder yml": {map[string]string{"package.json": `{"devDependencies":{"electron-builder":"^25"}}`, "electron-builder.yml": "appId: x"},
			"appId: x\ndirectories:\n  output: ../../dist/electron/apps/desk\n", "apps/desk/electron-builder.yml"},
		"builder yml with directories": {map[string]string{"package.json": `{"devDependencies":{"electron-builder":"^25"}}`, "electron-builder.yml": "directories:\n  buildResources: res\n"}, "", ""},
		"builder js":                   {map[string]string{"package.json": `{"devDependencies":{"electron-builder":"^25"}}`, "electron-builder.config.js": "module.exports = {}"}, "", ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ws := t.TempDir()
			dir := filepath.Join(ws, "apps", "desk")
			writeFiles(t, dir, c.files)
			fix := MoveOutDirFix(ws, p, Detect(dir))
			if c.want == "" {
				if fix != nil {
					t.Fatalf("no fix expected: %+v", fix)
				}
				return
			}
			if fix == nil || fix.Description != MoveDescription || len(fix.Files) != 0 || len(fix.Choices) != 1 || len(fix.Choices[0].Options) != 2 {
				t.Fatalf("a presented choice: %+v", fix)
			}
			if leave := fix.Choices[0].Options[1]; len(leave.Fix.Files) != 0 || !strings.HasPrefix(leave.Label, "leave desk") {
				t.Errorf("leave changes nothing: %+v", leave)
			}
			edits := fix.Choices[0].Options[0].Fix.Files
			if len(edits) != 1 || edits[0].Path != c.file || !strings.Contains(edits[0].Content, c.want) {
				t.Errorf("edits %+v, want %q in %s", edits, c.want, c.file)
			}
		})
	}
}
