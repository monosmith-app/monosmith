package tooling

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestApplyAddsOnlyWhatIsMissing(t *testing.T) {

	ws := t.TempDir()
	write(t, ws, "package.json", `{"name": "acme", "devDependencies": {"jest": "1.0.0"}}`)
	write(t, ws, "apps/web/package.json", `{"name": "@acme/web", "scripts": {"build": "tsc", "test": "mine"}}`)
	write(t, ws, "tool.config.mjs", "// mine\n")
	params := protocol.GenerateParams{Name: "web", Directory: "apps/web"}
	setup := Setup{
		RootFiles:       map[string]string{"tool.config.mjs": "// theirs\n", "other.json": "{}\n"},
		DevDependencies: map[string]string{"jest": "^30.0.0", "ts-jest": "^29.0.0"},
		Files:           map[string]string{"src/a.spec.ts": "test\n"},
		Scripts:         map[string]string{"test": "jest", "lint": "eslint ."},
		Fields:          map[string]any{"jest": map[string]any{"preset": "ts-jest"}},
	}
	res, err := Apply(ws, params, setup)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(res.Files, ","); got != "apps/web/package.json,apps/web/src/a.spec.ts,other.json,package.json" {
		t.Fatalf("files %s", got)
	}
	if !res.InstallPending {
		t.Fatal("changed root devDependencies must report InstallPending")
	}
	root := read(t, ws, "package.json")
	if !strings.Contains(root, `"jest": "1.0.0"`) || !strings.Contains(root, `"ts-jest": "^29.0.0"`) {
		t.Fatalf("root:\n%s", root)
	}
	project := read(t, ws, "apps/web/package.json")
	for _, want := range []string{`"test": "mine"`, `"lint": "eslint ."`, `"preset": "ts-jest"`} {
		if !strings.Contains(project, want) {
			t.Fatalf("project lacks %s:\n%s", want, project)
		}
	}
	if read(t, ws, "tool.config.mjs") != "// mine\n" {
		t.Fatal("an existing root file was overwritten")
	}

	res, err = Apply(ws, params, setup)
	if err != nil || len(res.Files) != 0 || res.InstallPending {
		t.Fatalf("a second apply changes nothing: %v %v", res.Files, err)
	}
}

func TestApplyNeedsTheRootManifest(t *testing.T) {
	if _, err := Apply(t.TempDir(), protocol.GenerateParams{Directory: "apps/web"}, Setup{}); err == nil {
		t.Fatal("a workspace without package.json must fail")
	}
}

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestApplyAddsRootFieldsAndEntriesTheWorkspaceLacks(t *testing.T) {
	fields := map[string]any{"overrides": map[string]any{"eslint-plugin-jsx-a11y": map[string]any{"eslint": "$eslint"}}}
	params := protocol.GenerateParams{Name: "web", Directory: "apps/web"}

	fresh := t.TempDir()
	write(t, fresh, "package.json", `{"name": "acme"}`)
	write(t, fresh, "apps/web/package.json", `{"name": "@acme/web"}`)
	if _, err := Apply(fresh, params, Setup{RootFields: fields}); err != nil {
		t.Fatal(err)
	}
	if root := read(t, fresh, "package.json"); !strings.Contains(root, `"eslint": "$eslint"`) {
		t.Fatalf("the override was not added:\n%s", root)
	}

	own := t.TempDir()
	write(t, own, "package.json", `{"name": "acme", "overrides": {"semver": "7.5.2"}}`)
	write(t, own, "apps/web/package.json", `{"name": "@acme/web"}`)
	if _, err := Apply(own, params, Setup{RootFields: fields}); err != nil {
		t.Fatal(err)
	}
	if root := read(t, own, "package.json"); !strings.Contains(root, "eslint-plugin-jsx-a11y") || !strings.Contains(root, `"semver": "7.5.2"`) {
		t.Fatalf("an object field gains the entries it lacks and keeps its own:\n%s", root)
	}

	chosen := t.TempDir()
	write(t, chosen, "package.json", `{"name": "acme", "overrides": {"eslint-plugin-jsx-a11y": "1.0.0"}}`)
	write(t, chosen, "apps/web/package.json", `{"name": "@acme/web"}`)
	if _, err := Apply(chosen, params, Setup{RootFields: fields}); err != nil {
		t.Fatal(err)
	}
	if root := read(t, chosen, "package.json"); strings.Contains(root, "$eslint") || !strings.Contains(root, `"eslint-plugin-jsx-a11y": "1.0.0"`) {
		t.Fatalf("an entry the workspace chose is left alone:\n%s", root)
	}
}
