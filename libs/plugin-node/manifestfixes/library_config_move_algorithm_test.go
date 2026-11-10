package manifestfixes

import (
	"strings"
	"testing"
)

func TestMoveConfigOutputsMovesOnlyTheOldFolder(t *testing.T) {
	config := "{\n  // ours\n  \"compilerOptions\": {\n    \"outDir\": \"dist\",\n    \"tsBuildInfoFile\": \"dist/tsconfig.lib.tsbuildinfo\",\n    \"declarationDir\": \"./dist/src\"\n  }\n}\n"
	got := moveConfigOutputs(config, "dist", "../../dist/packages/kit")
	for _, want := range []string{`"outDir": "../../dist/packages/kit"`, `"tsBuildInfoFile": "../../dist/packages/kit/tsconfig.lib.tsbuildinfo"`, `"declarationDir": "../../dist/packages/kit/src"`, "// ours"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in\n%s", want, got)
		}
	}
	spec := `{"compilerOptions": {"outDir": "./out-tsc/jest"}}`
	if moveConfigOutputs(spec, "dist", "../../dist/packages/kit") != spec {
		t.Error("a folder that is not the build folder stays")
	}
}

func TestAddDeclarationDirGoesBeforeOutDirWithItsIndent(t *testing.T) {
	got := addDeclarationDir("{\r\n  \"compilerOptions\": {\r\n    \"rootDir\": \"src\",\r\n    \"outDir\": \"x\"\r\n  }\r\n}\r\n", "x/src")
	want := "    \"declarationDir\": \"x/src\",\r\n    \"outDir\": \"x\""
	if !strings.Contains(got, want) || strings.Contains(got, "\r\r") {
		t.Fatalf("%q", got)
	}
	if !configHasOption(got, "declarationDir") || configHasOption("{\n // \"declarationDir\": 1\n}", "declarationDir") {
		t.Error("an option in a comment is not set")
	}
}

func TestAddPathAliasesCreatesThePathsObject(t *testing.T) {
	base := "{\n  \"compilerOptions\": {\n    \"composite\": true,\n    \"strict\": true\n  }\n}\n"
	got, conflicts, err := addPathAliases(base, map[string]string{"@a/two": "./packages/two/src/index.ts", "@a/one": "./packages/one/src/index.ts"})
	if err != nil || len(conflicts) != 0 {
		t.Fatal(err, conflicts)
	}
	want := "{\n  \"compilerOptions\": {\n    \"paths\": {\n      \"@a/one\": [\"./packages/one/src/index.ts\"],\n      \"@a/two\": [\"./packages/two/src/index.ts\"]\n    },\n    \"composite\": true,\n    \"strict\": true\n  }\n}\n"
	if got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	again, _, _ := addPathAliases(got, map[string]string{"@a/one": "./packages/one/src/index.ts"})
	if again != got {
		t.Error("an alias that is there already changes nothing")
	}
}

func TestAddPathAliasesAddsToExistingPathsAndReportsConflicts(t *testing.T) {
	base := "{\n  \"compilerOptions\": {\n    \"paths\": {\n      \"@a/one\": [\"elsewhere/index.ts\"],\n      \"@a/old\": [\"x.ts\"]\n    }\n  }\n}\n"
	got, conflicts, err := addPathAliases(base, map[string]string{"@a/one": "./packages/one/src/index.ts", "@a/two": "./packages/two/src/index.ts"})
	if err != nil || conflicts["@a/one"] != "elsewhere/index.ts" || len(conflicts) != 1 {
		t.Fatal(err, conflicts)
	}
	if !strings.Contains(got, "\"paths\": {\n      \"@a/two\": [\"./packages/two/src/index.ts\"],\n      \"@a/one\"") || strings.Contains(got, "./packages/one") {
		t.Fatalf("the new alias goes in, the conflicting one stays as it was:\n%s", got)
	}
	empty, _, _ := addPathAliases("{\"compilerOptions\": {\"paths\": {}}}", map[string]string{"@a/x": "x.ts"})
	if !strings.Contains(empty, `"@a/x": ["x.ts"]`) {
		t.Fatalf("an empty paths object takes the entry: %s", empty)
	}
}

func TestAddPathAliasesRefusesWhatItCannotEditWithCertainty(t *testing.T) {
	for name, config := range map[string]string{
		"a baseUrl elsewhere": `{"compilerOptions": {"baseUrl": "src"}}`,
		"no compilerOptions":  `{"extends": "./x.json"}`,
		"not JSON":            `[`,
	} {
		if _, _, err := addPathAliases(config, map[string]string{"@a/x": "x.ts"}); err == nil {
			t.Errorf("%s is refused", name)
		}
	}
}
