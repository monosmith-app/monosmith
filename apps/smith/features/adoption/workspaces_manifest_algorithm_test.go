package adoption

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestWorkspacePatternsCoverTopLevelAndNestedProjects(t *testing.T) {
	got := workspacePatterns([]string{"web", "packages/a", "packages/b", "apps/site"})
	want := []string{"apps/*", "packages/*", "web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestWithWorkspacesKeepsTheManifestAsWritten(t *testing.T) {
	raw := []byte("{\n  \"name\": \"acme\",\n  \"scripts\": {\"build\": \"a && b\"}\n}\n")
	out, changed, err := withWorkspaces(raw, "acme", []string{"web"})
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	if !strings.Contains(string(out), `"build": "a && b"`) || !strings.Contains(string(out), `"workspaces": ["web"]`) {
		t.Fatalf("manifest:\n%s", out)
	}
	var probe map[string]any
	if err := json.Unmarshal(out, &probe); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out)
	}
}

func TestWithWorkspacesStartsAManifestAndLeavesAnExistingListAlone(t *testing.T) {
	out, changed, err := withWorkspaces(nil, "acme", []string{"web", "packages/*"})
	if err != nil || !changed || !strings.Contains(string(out), `"private": true`) || !strings.Contains(string(out), `"web", "packages/*"`) {
		t.Fatalf("new manifest: %v %v\n%s", changed, err, out)
	}
	own := []byte(`{"name":"a","workspaces":["x/*"]}`)
	if got, changed, _ := withWorkspaces(own, "acme", []string{"web"}); changed || string(got) != string(own) {
		t.Fatalf("an existing workspaces list is the user's: %s", got)
	}
	if got, changed, _ := withWorkspaces([]byte("{}"), "acme", []string{"web"}); !changed || !strings.Contains(string(got), `"workspaces": ["web"]`) {
		t.Fatalf("empty object: %s", got)
	}
}
