package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStage(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("packages/sdk/package.json", `{
  "name": "@acme/sdk",
  "version": "1.0.0",
  "main": "../../dist/packages/sdk/index.js",
  "types": "../../dist/packages/sdk/index.d.ts",
  "bin": { "sdk": "./../../dist/packages/sdk/cli.js" },
  "files": ["../../dist/packages/sdk"],
  "scripts": { "build": "tsc" },
  "dependencies": { "zod": "^3" }
}
`)
	write("packages/sdk/README.md", "# sdk\n")
	write("packages/sdk/package.nls.json", `{"title": "SDK"}`)
	write("packages/sdk/package.nls.pt-br.json", `{"title": "SDK"}`)
	write("dist/packages/sdk/index.js", "")
	dir, err := Stage(ws, "packages/sdk")
	if err != nil || dir != filepath.Join(ws, "dist", "packages", "sdk") {
		t.Fatal(dir, err)
	}
	staged, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	for _, want := range []string{`"main": "./index.js"`, `"types": "./index.d.ts"`, `"sdk": "./cli.js"`, `"zod": "^3"`} {
		if !strings.Contains(string(staged), want) {
			t.Errorf("staged package.json lacks %s:\n%s", want, staged)
		}
	}
	if strings.Contains(string(staged), `"files"`) || strings.Contains(string(staged), `"scripts"`) {
		t.Errorf("files and scripts do not belong to the staged package:\n%s", staged)
	}
	if _, err := os.Stat(filepath.Join(dir, "README.md")); err != nil {
		t.Error("the README goes with it")
	}
	for _, nls := range []string{"package.nls.json", "package.nls.pt-br.json"} {
		if _, err := os.Stat(filepath.Join(dir, nls)); err != nil {
			t.Errorf("%s goes with it: an extension's %%key%% texts are read from it", nls)
		}
	}
	// A package the workspace reads from its sources publishes their build.
	write("packages/kit/package.json", `{"name": "@acme/kit", "type": "module", "exports": "./src/index.ts"}`)
	write("dist/packages/kit/index.js", "")
	kit, _ := Stage(ws, "packages/kit")
	staged, _ = os.ReadFile(filepath.Join(kit, "package.json"))
	for _, want := range []string{`"default": "./index.js"`, `"types": "./index.d.ts"`, `"main": "./index.js"`, `"type": "module"`} {
		if !strings.Contains(string(staged), want) {
			t.Errorf("staged kit lacks %s:\n%s", want, staged)
		}
	}
	write("apps/plain/package.json", `{"name": "plain"}`)
	if dir, _ := Stage(ws, "apps/plain"); dir != filepath.Join(ws, "apps", "plain") {
		t.Fatalf("a project without a build is packed from its own directory: %s", dir)
	}
}
