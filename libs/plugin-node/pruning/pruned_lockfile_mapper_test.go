package pruning

import (
	"encoding/json"
	"reflect"
	"testing"
)

const workspaceLock = `{
  "name": "acme", "lockfileVersion": 3, "requires": true,
  "packages": {
    "": { "name": "acme", "workspaces": ["apps/*", "packages/*"] },
    "apps/api": { "name": "@acme/api", "version": "1.2.0" },
    "apps/api/node_modules/ms": { "version": "3.0.0" },
    "packages/core": { "name": "@acme/core", "version": "0.1.0" },
    "packages/core/node_modules/debug": { "version": "4.0.0" },
    "packages/other": { "name": "@acme/other", "version": "0.1.0" },
    "packages/other/node_modules/chalk": { "version": "5.0.0" },
    "node_modules/@acme/api": { "resolved": "apps/api", "link": true },
    "node_modules/@acme/core": { "resolved": "packages/core", "link": true },
    "node_modules/@acme/other": { "resolved": "packages/other", "link": true },
    "node_modules/ms": { "version": "2.1.3" },
    "node_modules/express": { "version": "5.0.0" },
    "node_modules/express/node_modules/qs": { "version": "6.0.0" }
  }
}`

func TestWorkspacePackagesComeFromTheLockfilesLinks(t *testing.T) {
	got, err := WorkspacePackages([]byte(workspaceLock))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"@acme/api": "apps/api", "@acme/core": "packages/core", "@acme/other": "packages/other"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}
}

func TestSeedLockfileRootsTheAppAndMovesCopiedPackagesDependencies(t *testing.T) {
	out, err := SeedLockfile([]byte(workspaceLock), "@acme/api", "1.2.0", "apps/api", map[string]string{"@acme/core": "packages/core"})
	if err != nil {
		t.Fatal(err)
	}
	var l struct {
		Packages map[string]map[string]any `json:"packages"`
	}
	if err := json.Unmarshal(out, &l); err != nil {
		t.Fatal(err)
	}
	keys := map[string]string{}
	for k, v := range l.Packages {
		version, _ := v["version"].(string)
		keys[k] = version
	}
	want := map[string]string{
		"":                                     "1.2.0",
		"node_modules/ms":                      "3.0.0", // the app's own version wins
		"node_modules/express":                 "5.0.0",
		"node_modules/express/node_modules/qs": "6.0.0",
		"workspace_modules/@acme/core/node_modules/debug": "4.0.0",
	}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("packages = %v", keys)
	}
}
