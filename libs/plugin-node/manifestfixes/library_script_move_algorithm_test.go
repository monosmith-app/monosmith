package manifestfixes

import "testing"

func TestMoveLibraryScriptRewritesEveryPathItWrites(t *testing.T) {
	script := "esbuild src/index.ts --bundle --packages=external --format=esm --outfile=dist/index.esm.js --sourcemap && tsc -p tsconfig.lib.json --emitDeclarationOnly --declaration --declarationDir dist/src --tsBuildInfoFile dist/src/tsconfig.lib.tsbuildinfo && smith declaration-specifiers dist/src"
	want := "esbuild src/index.ts --bundle --packages=external --format=esm --outfile=../../dist/packages/kit/index.esm.js --sourcemap && tsc -p tsconfig.lib.json --emitDeclarationOnly --declaration --declarationDir ../../dist/packages/kit/src --tsBuildInfoFile ../../dist/packages/kit/src/tsconfig.lib.tsbuildinfo && smith declaration-specifiers ../../dist/packages/kit/src"
	got, reason := moveLibraryScript(script, "dist", "../../dist/packages/kit")
	if reason != "" || got != want {
		t.Fatalf("reason %q\n got %s\nwant %s", reason, got, want)
	}
}

func TestMoveLibraryScriptKeepsQuotesAndEnvironmentAssignments(t *testing.T) {
	got, reason := moveLibraryScript(`NODE_ENV=production esbuild src/main.ts --outdir="./dist"`, "dist", "../../dist/packages/kit")
	if reason != "" || got != `NODE_ENV=production esbuild src/main.ts --outdir="../../dist/packages/kit"` {
		t.Fatalf("%q %q", got, reason)
	}
}

func TestMoveLibraryScriptRefusesWhatItCannotReadWithCertainty(t *testing.T) {
	for name, script := range map[string]string{
		"an unknown command": "rimraf dist && tsc --outDir dist",
		"a pipe":             "tsc --outDir dist | tee build.log",
		"a redirect":         "node gen.js > dist/out.txt",
		"a subshell":         "tsc --outDir dist --pretty $(date)",
	} {
		t.Run(name, func(t *testing.T) {
			if got, reason := moveLibraryScript(script, "dist", "../../dist/packages/kit"); reason == "" || got != script {
				t.Fatalf("refused with a reason and the script unchanged: %q %q", got, reason)
			}
		})
	}
}

func TestMoveLibraryScriptLeavesACommandThatNamesNoOldFolderAlone(t *testing.T) {
	got, reason := moveLibraryScript("eslint . && tsc --outDir dist", "dist", "../../dist/packages/kit")
	if reason != "" || got != "eslint . && tsc --outDir ../../dist/packages/kit" {
		t.Fatalf("%q %q", got, reason)
	}
}
