package vite

import "testing"

func TestReactBuildPackagesNameViteAndItsReactPluginWithCatalogVersions(t *testing.T) {
	got, overrides := ReactBuildPackages()
	for _, name := range []string{"vite", "@vitejs/plugin-react"} {
		if got[name] == "" {
			t.Errorf("%s has no version: %v", name, got)
		}
	}
	if len(got) != 2 {
		t.Errorf("want only the two build packages, got %v", got)
	}
	if overrides["@babel/plugin-transform-runtime"] == "" {
		t.Errorf("the React stack's override is missing: %v", overrides)
	}
}
