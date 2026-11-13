package tsconfigchain

import (
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, root, name, body string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReachesBase(t *testing.T) {
	root := t.TempDir()
	put(t, root, Base, BaseFile)
	put(t, root, "tsconfig.browser.json", `{"extends": "./tsconfig.base.json"}`)
	put(t, root, "apps/direct/tsconfig.json", `{"extends": "../../tsconfig.base.json"}`)
	put(t, root, "apps/browser/tsconfig.json", `{"extends": "../../tsconfig.browser.json"}`)
	put(t, root, "apps/array/tsconfig.json", `{"extends": ["astro/tsconfigs/strict", "../../tsconfig.browser.json"]}`)
	put(t, root, "apps/build/tsconfig.json", `{"extends": "./tsconfig.mid.json"} // comment`)
	put(t, root, "apps/build/tsconfig.mid.json", `{"extends": "../../tsconfig.base.json"}`)
	put(t, root, "apps/preset/tsconfig.json", `{"extends": "@monosmith/tsconfig/node.json"}`)
	put(t, root, "apps/none/tsconfig.json", `{"compilerOptions": {}}`)
	put(t, root, "apps/wrong/tsconfig.json", `{"extends": "../tsconfig.base.json"}`)
	for dir, want := range map[string]bool{
		"direct": true, "browser": true, "array": true, "build": true,
		"preset": false, "none": false, "wrong": false,
	} {
		got, err := ReachesBase(root, filepath.Join(root, "apps", dir, "tsconfig.json"))
		if err != nil || got != want {
			t.Errorf("%s: got %v, %v; want %v", dir, got, err, want)
		}
	}
}

func TestUnreachedListsTheFilesThatMissTheBase(t *testing.T) {
	root := t.TempDir()
	put(t, root, Base, BaseFile)
	put(t, root, "apps/ok/tsconfig.json", `{"extends": "../../tsconfig.base.json"}`)
	put(t, root, "apps/ok/tsconfig.build.json", `{"extends": "./tsconfig.json"}`)
	put(t, root, "apps/bad/tsconfig.json", `{"compilerOptions": {}}`)
	put(t, root, "apps/bad/node_modules/x/tsconfig.json", `{}`)
	got, err := Unreached(root)
	if err != nil || len(got) != 1 || got[0] != "apps/bad/tsconfig.json" {
		t.Errorf("got %v, %v", got, err)
	}
}
