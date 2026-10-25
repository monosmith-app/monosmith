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
	out, changed, err := completeRootManifest([]byte(raw), "acme", map[string]string{"vitest": "^5.0.0", "eslint": "^9.0.0"}, nil)
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
	out, changed, err := completeRootManifest([]byte(raw), "a", map[string]string{"vitest": "^4.0.0"}, nil)
	if err != nil || changed || string(out) != raw {
		t.Fatalf("%q changed %v err %v", out, changed, err)
	}
}

func TestCompleteRootManifestStartsANewOne(t *testing.T) {
	out, changed, err := completeRootManifest(nil, "Acme", map[string]string{"vitest": "^5.0.0"}, nil)
	if err != nil || !changed {
		t.Fatalf("changed %v, err %v", changed, err)
	}
	for _, want := range []string{`"name": "acme-workspace"`, `"private": true`, `"vitest": "^5.0.0"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("missing %s in\n%s", want, out)
		}
	}
}

func TestCompleteRootManifestAddsOverridesBesideTheOnesThere(t *testing.T) {
	raw := `{"name":"a","workspaces":["apps/*"],"overrides":{"eslint-plugin-jsx-a11y":{"eslint":"$eslint"},"semver":"7.5.2"}}`
	out, changed, err := completeRootManifest([]byte(raw), "a", nil, map[string]string{"@babel/plugin-transform-runtime": "^7.29.0", "semver": "9.9.9"})
	if err != nil || !changed {
		t.Fatalf("changed %v, err %v", changed, err)
	}
	got := string(out)
	for _, want := range []string{`"@babel/plugin-transform-runtime": "^7.29.0"`, `"eslint": "$eslint"`, `"semver": "7.5.2"`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in\n%s", want, got)
		}
	}
	if strings.Contains(got, "9.9.9") {
		t.Errorf("an override the workspace chose was replaced:\n%s", got)
	}
	if _, changed, _ = completeRootManifest(out, "a", nil, map[string]string{"@babel/plugin-transform-runtime": "^7.29.0"}); changed {
		t.Error("a second pass changed it again")
	}
}

func TestCompleteRootManifestStartsOverridesWhenThereAreNone(t *testing.T) {
	out, changed, err := completeRootManifest([]byte(`{"name":"a","workspaces":["apps/*"]}`), "a", nil, map[string]string{"x": "^1.0.0"})
	want := "\"overrides\": {\n    \"x\": \"^1.0.0\"\n  }"
	if err != nil || !changed || !strings.Contains(string(out), want) {
		t.Fatalf("changed %v, err %v\n%s", changed, err, out)
	}
}
