package manifestfixes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAPackageThatPublishesItsOwnDistIsOfferedTheMoveAsAQuestion(t *testing.T) {
	ws, p := projectWith(t,
		`{
  "name": "@acme/kit",
  "version": "1.0.0",
  "main": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": { "types": "./dist/index.d.ts", "default": "./dist/index.js" },
    "./package.json": "./package.json"
  },
  "files": ["dist"],
  "scripts": { "build": "tsc -p tsconfig.lib.json", "test": "jest" }
}
`, `{"compilerOptions":{"outDir":"./dist","declarationDir":"./dist/types"}}`)
	p.Kind = protocol.ProjectPackage
	fix, refusal := MigrateBuildFix(ws, p, "tsconfig.json", "./dist", nil)
	if refusal != "" || fix == nil {
		t.Fatalf("a package whose whole tarball is its dist moves: %q", refusal)
	}
	if len(fix.Files) != 0 || len(fix.Choices) != 1 || len(fix.Choices[0].Options) != 2 {
		t.Fatalf("nothing is written before the answer, and 'leave' is an option: %+v", fix)
	}
	leave := fix.Choices[0].Options[1]
	if len(leave.Fix.Files) != 0 || !strings.Contains(leave.Label, "leave kit as it is") {
		t.Fatalf("leave changes nothing: %+v", leave)
	}
	move := fix.Choices[0].Options[0].Fix
	manifest := edited(&move, "package.json")
	for _, want := range []string{
		`"main": "../../dist/packages/kit/index.js"`,
		`"types": "../../dist/packages/kit/index.d.ts"`,
		`"types": "../../dist/packages/kit/index.d.ts", "default"`,
		`"./package.json": "./package.json"`,
		`"test": "jest"`,
	} {
		if !strings.Contains(strings.ReplaceAll(manifest, "\n      ", " "), want) && !strings.Contains(manifest, want) {
			t.Errorf("missing %s:\n%s", want, manifest)
		}
	}
	if strings.Contains(manifest, `"files"`) {
		t.Errorf("a files list into the old folder goes:\n%s", manifest)
	}
	config := edited(&move, "tsconfig.json")
	if strings.Contains(config, "outDir") || !strings.Contains(config, `"declarationDir":"../../dist/packages/kit/types"`) {
		t.Errorf("the declaration folder follows:\n%s", config)
	}
}

func TestAPublishedPackageWithFilesOutsideItsBuildIsRefused(t *testing.T) {
	cases := map[string]string{
		"a files entry": `{"name":"k","files":["dist","templates"],"scripts":{"build":"tsc --outDir dist"}}`,
		"a bin source":  `{"name":"k","bin":"./bin/cli.js","files":["dist"],"scripts":{"build":"tsc --outDir dist"}}`,
		"a second root": `{"name":"k","scripts":{"build":"tsc --outDir dist --declarationDir types"}}`,
	}
	for name, manifest := range cases {
		t.Run(name, func(t *testing.T) {
			ws, p := projectWith(t, manifest, `{}`)
			p.Kind = protocol.ProjectPackage
			fix, refusal := MigrateBuildFix(ws, p, "", "dist", nil)
			if fix != nil || refusal == "" {
				t.Fatalf("the move is refused with a reason, not written: %+v %q", fix, refusal)
			}
		})
	}
}

func TestAPrivateAppMovesEvenWithFilesOutsideItsBuild(t *testing.T) {
	ws, p := projectWith(t, `{"name":"app","private":true,"files":["dist","templates"],"scripts":{"build":"esbuild src/main.ts --bundle --outdir=dist","start":"node dist/main.js"}}`, `{}`)
	p.Kind = protocol.ProjectApp
	fix, refusal := MigrateBuildFix(ws, p, "", "dist", nil)
	if refusal != "" || fix == nil {
		t.Fatalf("an app is never published from dist: %q", refusal)
	}
	move := fix.Choices[0].Options[0].Fix
	manifest := edited(&move, "package.json")
	if !strings.Contains(manifest, `--outdir=../../dist/packages/kit`) || !strings.Contains(manifest, `node ../../dist/packages/kit/main.js`) || !strings.Contains(manifest, `"templates"`) {
		t.Fatalf("%s", manifest)
	}
}

func TestAProjectOthersImportByNameIsRefusedBecauseTheyWouldStopCompiling(t *testing.T) {
	ws, p := projectWith(t, `{"name":"@a/kit","private":true,"main":"./dist/index.js","types":"./dist/index.d.ts","scripts":{"build":"tsc --outDir dist"}}`, `{}`)
	fix, refusal := MigrateBuildFix(ws, p, "", "dist", []string{"web"})
	if fix != nil || !strings.Contains(refusal, "web import it by name") {
		t.Fatalf("%+v %q", fix, refusal)
	}
	// A project nobody imports, or whose entry points stay in the project, moves.
	if fix, refusal := MigrateBuildFix(ws, p, "", "dist", nil); fix == nil || refusal != "" {
		t.Fatalf("no importer, no problem: %q", refusal)
	}
}

func TestEveryScriptThatNamesTheOldFolderAsAPathFollowsAndOtherWordsStay(t *testing.T) {
	cases := map[string]string{
		"node dist/main.js":                      "node ../../dist/packages/kit/main.js",
		"rimraf ./dist":                          "rimraf ../../dist/packages/kit",
		`node -e "require('./dist/index.js')"`:   `node -e "require('../../dist/packages/kit/index.js')"`,
		"next build && cp -r assets dist/assets": "next build && cp -r assets ../../dist/packages/kit/assets",
		"tsc --outDir ../../dist/packages/kit":   "tsc --outDir ../../dist/packages/kit",
		"echo distribution dist-a":               "echo distribution dist-a",
	}
	for in, want := range cases {
		if got := rewriteScriptPaths(in, "dist", "../../dist/packages/kit"); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
	if got := rewriteScriptPaths("next build", "build", "../../dist/packages/kit"); got != "next build" {
		t.Errorf("a bare word is not a path: %q", got)
	}
}

func TestSmithTomlOutputsOverrideIsDroppedByACommandNotRewritten(t *testing.T) {
	ws, p := projectWith(t, `{"name":"kit","scripts":{"build":"tsc -p tsconfig.json"}}`, `{"compilerOptions":{"outDir":"./dist"}}`)
	_ = os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("schema = 1\n\n[projects.kit.tasks.build]\noutputs = [\"dist\"]\n"), 0o644)
	p.Tasks = map[string]protocol.Task{"build": {Outputs: []string{"dist"}}}
	fix, refusal := MigrateBuildFix(ws, p, "tsconfig.json", "./dist", nil)
	if fix == nil || refusal != "" {
		t.Fatalf("%q", refusal)
	}
	move := fix.Choices[0].Options[0].Fix
	if len(move.Steps) != 1 || strings.Join(move.Steps[0].Argv, " ") != "smith config unset projects.kit.tasks.build.outputs" {
		t.Fatalf("%+v", move.Steps)
	}
	p.Tasks = map[string]protocol.Task{"build": {Outputs: []string{"dist", "coverage"}}}
	if fix, refusal := MigrateBuildFix(ws, p, "tsconfig.json", "./dist", nil); fix != nil || !strings.Contains(refusal, "coverage") {
		t.Fatalf("an output that is not the build folder is refused: %+v %q", fix, refusal)
	}
}

func TestOutfileAndTheBuildScriptsOwnTsconfigAreLocated(t *testing.T) {
	ws, p := projectWith(t, `{"name":"k","scripts":{"build":"esbuild src/i.ts --outfile=lib/index.js"}}`, `{}`)
	config, dir, ok := LocateOutDir(ws + "/" + p.Root)
	if !ok || config != "" || dir != "lib" {
		t.Fatalf("--outfile names the folder of the file: %q %q %v", config, dir, ok)
	}
	fix := BuildIntoDistFix(ws, p, "", "lib")
	if fix == nil || !strings.Contains(edited(fix, "package.json"), `--outfile=../../dist/packages/kit/index.js`) {
		t.Fatalf("%+v", fix)
	}
}
