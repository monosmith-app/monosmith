package pruning

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPrunedManifestPointsWorkspaceDependenciesAtTheirCopies(t *testing.T) {
	raw := []byte(`{
  "name": "@acme/api",
  "version": "1.2.0",
  "main": "dist/main.js",
  "dependencies": { "@acme/core": ">=0.0.1", "express": "^5.0.0" },
  "peerDependencies": { "@acme/util": "*", "zod": "^4" },
  "peerDependenciesMeta": { "@acme/util": { "optional": true } },
  "devDependencies": { "jest": "^30" },
  "workspaces": ["x"]
}`)
	ws := map[string]bool{"@acme/core": true, "@acme/util": true}
	out, used, err := PrunedManifest(raw, func(n string) bool { return ws[n] }, func(n string) string { return "file:./workspace_modules/" + n })
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"@acme/core", "@acme/util"}; !reflect.DeepEqual(used, want) {
		t.Errorf("used = %v", used)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	deps := got["dependencies"].(map[string]any)
	if deps["@acme/core"] != "file:./workspace_modules/@acme/core" || deps["@acme/util"] != "file:./workspace_modules/@acme/util" || deps["express"] != "^5.0.0" {
		t.Errorf("dependencies = %v", deps)
	}
	if peers := got["peerDependencies"].(map[string]any); len(peers) != 1 || peers["zod"] != "^4" {
		t.Errorf("peerDependencies = %v", peers)
	}
	for _, gone := range []string{"devDependencies", "workspaces", "peerDependenciesMeta"} {
		if _, ok := got[gone]; ok {
			t.Errorf("%s kept", gone)
		}
	}
	if got["main"] != "dist/main.js" {
		t.Errorf("main = %v", got["main"])
	}
}

func TestPrunedManifestLeavesAManifestWithoutWorkspaceDependencies(t *testing.T) {
	raw := []byte(`{ "name": "a", "dependencies": { "ms": "^2" } }`)
	out, used, err := PrunedManifest(raw, func(string) bool { return false }, func(string) string { return "" })
	if err != nil || len(used) != 0 {
		t.Fatalf("used %v, %v", used, err)
	}
	var got map[string]any
	_ = json.Unmarshal(out, &got)
	if got["dependencies"].(map[string]any)["ms"] != "^2" {
		t.Errorf("got %s", out)
	}
}
