package scaffolding

import (
	"strings"
	"testing"
)

func TestCompleteRootManifestKeepsWhatIsThereAsWritten(t *testing.T) {
	raw := `{
  "name": "acme",
  "scripts": {
    "affected": "smith run lint --affected && smith run test --affected",
    "build": "smith run build"
  },
  "devDependencies": {
    "eslint": "^10.0.0",
    "zod": "^4.0.0"
  },
  "private": true
}
`
	out, changed, err := completeRootManifest([]byte(raw), "acme", map[string]string{"vitest": "^5.0.0", "eslint": "^9.0.0"})
	if err != nil || !changed {
		t.Fatalf("changed %v, err %v", changed, err)
	}
	got := string(out)
	for _, want := range []string{
		`"affected": "smith run lint --affected && smith run test --affected"`,
		`"eslint": "^10.0.0"`,
		`"vitest": "^5.0.0"`,
		"\"workspaces\": [\n    \"apps/*\",\n    \"libs/*\",\n    \"packages/*\"\n  ]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in\n%s", want, got)
		}
	}
	if strings.Contains(got, "\\u0026") {
		t.Errorf("& was escaped:\n%s", got)
	}
	if strings.Index(got, `"name"`) > strings.Index(got, `"scripts"`) || strings.Index(got, `"scripts"`) > strings.Index(got, `"private"`) {
		t.Errorf("keys moved:\n%s", got)
	}
	if strings.Index(got, `"vitest"`) < strings.Index(got, `"eslint"`) || strings.Index(got, `"vitest"`) > strings.Index(got, `"zod"`) {
		t.Errorf("sorted devDependencies lost their order:\n%s", got)
	}
}

func TestCompleteRootManifestLeavesAFinishedOneUntouched(t *testing.T) {
	raw := `{"name":"a","workspaces":["packages/*"],"devDependencies":{"vitest":"^5.0.0"}}`
	out, changed, err := completeRootManifest([]byte(raw), "a", map[string]string{"vitest": "^4.0.0"})
	if err != nil || changed || string(out) != raw {
		t.Fatalf("%q changed %v err %v", out, changed, err)
	}
}

func TestCompleteRootManifestStartsANewOne(t *testing.T) {
	out, changed, err := completeRootManifest(nil, "Acme", map[string]string{"vitest": "^5.0.0"})
	if err != nil || !changed {
		t.Fatalf("changed %v, err %v", changed, err)
	}
	for _, want := range []string{`"name": "acme-workspace"`, `"private": true`, `"vitest": "^5.0.0"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("missing %s in\n%s", want, out)
		}
	}
}
