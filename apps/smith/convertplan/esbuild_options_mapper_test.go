package convertplan

import (
	"strings"
	"testing"
)

func TestMapEsbuildOptionsCarriesOutExtensionAndSourcemap(t *testing.T) {
	cases := []struct {
		raw, flags, sourcemap, unmapped string
	}{
		// OpenCraw's VS Code extension: main and its packaging expect .cjs.
		{`{"sourcemap": true, "outExtension": {".js": ".cjs"}}`, "--out-extension:.js=.cjs", "--sourcemap", ""},
		{`{"sourcemap": false, "outExtension": {".js": ".js"}}`, "", "", ""},
		{`{"sourcemap": "inline", "external": ["vscode"], "minify": true}`, "--external:vscode --minify=true", "--sourcemap=inline", ""},
		{`{"define": {"process.env.NODE_ENV": "\"production\""}, "banner": {"js": "x"}}`, "", "--sourcemap", "banner,define"},
		{`{}`, "", "--sourcemap", ""},
	}
	for _, c := range cases {
		tg := target(t, `{"executor": "@nx/esbuild:esbuild", "options": {"esbuildOptions": `+c.raw+`}}`)
		got := MapEsbuildOptions(tg)
		if strings.Join(got.Flags, " ") != c.flags || got.Sourcemap != c.sourcemap || strings.Join(got.Unmapped, ",") != c.unmapped {
			t.Errorf("%s: %+v", c.raw, got)
		}
	}
}

func TestEsbuildBuildCarriesOutExtension(t *testing.T) {
	tg := target(t, `{"executor": "@nx/esbuild:esbuild", "options": {"main": "apps/ext/src/extension.ts", "outputPath": "dist/apps/ext", "format": ["cjs"], "esbuildOptions": {"sourcemap": true, "outExtension": {".js": ".cjs"}}}}`)
	p := project("ext", "apps/ext", "app", nil, nil)
	line, _, ok := executorCommand(nil, p, tg)
	if !ok || line != "esbuild src/extension.ts --bundle --packages=external --platform=node --format=cjs --outdir=../../dist/apps/ext --out-extension:.js=.cjs --sourcemap" {
		t.Fatalf("line %q", line)
	}
}

func TestMapEsbuildOptionsCarriesLogOverride(t *testing.T) {
	tg := target(t, `{"executor": "@nx/esbuild:esbuild", "options": {"esbuildOptions": {"logOverride": {"require-resolve-not-external": "silent"}}}}`)
	got := MapEsbuildOptions(tg)
	if strings.Join(got.Flags, " ") != "--log-override:require-resolve-not-external=silent" || len(got.Unmapped) != 0 {
		t.Errorf("%+v", got)
	}
}
