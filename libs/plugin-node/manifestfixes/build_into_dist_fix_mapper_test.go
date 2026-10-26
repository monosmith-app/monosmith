package manifestfixes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func projectWith(t *testing.T, manifest, tsconfig string) (string, protocol.Project) {
	t.Helper()
	ws := t.TempDir()
	dir := filepath.Join(ws, "packages", "kit")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(manifest), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(tsconfig), 0o644)
	return ws, protocol.Project{Name: "kit", Root: "packages/kit"}
}

func edited(f *protocol.Fix, name string) string {
	for _, e := range f.Files {
		if strings.HasSuffix(e.Path, name) {
			return e.Content
		}
	}
	return ""
}

func TestATscBuildMovesItsOutDirOutOfTheTsconfigIntoAFlag(t *testing.T) {
	ws, p := projectWith(t,
		`{
  "name": "kit",
  "main": "./dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc -p tsconfig.json",
    "start": "node dist/main.js"
  }
}
`,
		`{
  // ours
  "compilerOptions": {
    "strict": true,
    "outDir": "./dist"
  },
  "include": ["src"]
}
`)
	f := BuildIntoDistFix(ws, p, "tsconfig.json", "./dist")
	if f == nil {
		t.Fatal("a fix was expected")
	}
	config := edited(f, "tsconfig.json")
	if strings.Contains(config, "outDir") || !strings.Contains(config, `"strict": true`) || !strings.Contains(config, "// ours") {
		t.Fatalf("the tsconfig keeps everything but the output folder:\n%s", config)
	}
	manifest := edited(f, "package.json")
	for _, want := range []string{`"build": "tsc -p tsconfig.json --outDir ../../dist/packages/kit"`, `"main": "../../dist/packages/kit/index.js"`, `"types": "../../dist/packages/kit/index.d.ts"`, `"start": "node ../../dist/packages/kit/main.js"`} {
		if !strings.Contains(manifest, want) {
			t.Errorf("missing %s:\n%s", want, manifest)
		}
	}
}

func TestAnOutDirThatIsNotLastOrNotATscBuildIsEditedInPlace(t *testing.T) {
	ws, p := projectWith(t, `{"name":"kit","scripts":{"build":"vite build"}}`, `{"compilerOptions":{"outDir":"build","strict":true}}`)
	f := BuildIntoDistFix(ws, p, "tsconfig.json", "build")
	if f == nil || !strings.Contains(edited(f, "tsconfig.json"), `"outDir":"../../dist/packages/kit"`) {
		t.Fatalf("the value is pointed at dist: %+v", f)
	}
	ws, p = projectWith(t, `{"name":"kit","scripts":{"build":"tsc"}}`, `{"compilerOptions":{"outDir":"build","strict":true}}`)
	f = BuildIntoDistFix(ws, p, "tsconfig.json", "build")
	if f == nil || strings.Contains(edited(f, "tsconfig.json"), "outDir") || !strings.Contains(edited(f, "tsconfig.json"), `"strict":true`) {
		t.Fatalf("a middle property goes cleanly: %+v", f)
	}
}

func TestAFlagInsideTheSourceTreeIsRewritten(t *testing.T) {
	ws, p := projectWith(t, `{"name":"kit","scripts":{"build":"tsc --outDir lib && node x.js"}}`, `{}`)
	f := BuildIntoDistFix(ws, p, "", "lib")
	if f == nil || !strings.Contains(edited(f, "package.json"), `tsc --outDir ../../dist/packages/kit && node x.js`) {
		t.Fatalf("%+v", f)
	}
}
