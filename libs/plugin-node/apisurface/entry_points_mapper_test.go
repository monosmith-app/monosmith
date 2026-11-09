package apisurface

import (
	"slices"
	"testing"
)

func TestEntryPointsMapsExportsConditionsAndFallbacks(t *testing.T) {
	manifest := `{
  "name": "pkg",
  "main": "./dist/index.cjs",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": { "require": "./dist/index.cjs", "import": "./dist/index.mjs", "types": "./dist/index.d.ts" },
    "./cli": "./dist/cli.js",
    "./data/*": "./dist/data/*.js",
    "./package.json": "./package.json"
  }
}`
	points, private, err := EntryPoints(manifest)
	if err != nil || private {
		t.Fatalf("EntryPoints = %v, private %v, err %v", points, private, err)
	}
	got := map[string][]string{}
	for _, p := range points {
		got[p.Subpath] = p.Candidates
	}
	if want := []string{"dist/index.d.ts", "dist/index.mjs", "dist/index.cjs"}; !slices.Equal(got["."], want) {
		t.Errorf("main entry = %v, want %v", got["."], want)
	}
	if want := []string{"dist/cli.js"}; !slices.Equal(got["./cli"], want) {
		t.Errorf("./cli = %v, want %v", got["./cli"], want)
	}
	if _, ok := got["./data/*"]; ok {
		t.Error("a wildcard subpath has no fixed file and must be left out")
	}
}

func TestEntryPointsWithoutExportsUsesTypesMainAndTheDefaultIndex(t *testing.T) {
	points, _, err := EntryPoints(`{"name":"p","main":"lib/main.js","types":"lib/main.d.ts"}`)
	if err != nil || len(points) != 1 {
		t.Fatalf("EntryPoints = %v, %v", points, err)
	}
	if want := []string{"lib/main.d.ts", "lib/main.js", "index.d.ts", "index.js"}; !slices.Equal(points[0].Candidates, want) {
		t.Errorf("candidates = %v, want %v", points[0].Candidates, want)
	}
}

func TestEntryPointsOfAPrivatePackageIsEmpty(t *testing.T) {
	points, private, err := EntryPoints(`{"name":"p","private":true,"main":"x.js"}`)
	if err != nil || !private || len(points) != 0 {
		t.Errorf("got %v, private %v, err %v", points, private, err)
	}
}

func TestSourceVariantsMapABuiltPathBackToItsSource(t *testing.T) {
	got := sourceVariants("dist/esm/util/index.d.ts")
	for _, want := range []string{"src/util/index.ts", "src/util/index.tsx", "util/index.ts"} {
		if !slices.Contains(got, want) {
			t.Errorf("sourceVariants lacks %q: %v", want, got)
		}
	}
	if slices.Contains(got, "dist/esm/util/index.d.ts") {
		t.Error("the built file itself is not a variant")
	}
}
