package buildoutputs

import (
	"slices"
	"strings"
	"testing"
)

func TestWrittenDirs(t *testing.T) {
	cases := map[string][]string{
		"esbuild src/index.ts --bundle --outfile=dist/index.esm.js --sourcemap && tsc -p tsconfig.lib.json --declarationDir dist/src --tsBuildInfoFile dist/src/x.tsbuildinfo && smith declaration-specifiers dist/src": {"dist", "dist/src"},
		"esbuild src/main.ts --bundle --outdir=dist":                             {"dist"},
		"rollup -c -o build/bundle.js":                                           {"build"},
		"tsc --outDir ../../dist/packages/x":                                     {"../../dist/packages/x"},
		"esbuild a.ts --outfile=bundle.js":                                       nil,
		"esbuild a.ts --outdir=$OUT":                                             nil,
		`node -e "require('esbuild').buildSync({outdir:'out',platform:'node'})"`: {"out"},
		"jest": nil,
	}
	for script, want := range cases {
		if got := WrittenDirs(script); !slices.Equal(got, want) {
			t.Errorf("%q: %v, want %v", script, got, want)
		}
	}
}

func TestGapOf(t *testing.T) {
	script := "esbuild src/index.ts --bundle --outfile=dist/index.esm.js && tsc --declarationDir dist/src"
	gap, ok := GapOf("auth", "packages/auth", script, []string{"../../dist/packages/auth"}, false)
	if !ok || !slices.Equal(gap.Missing, []string{"dist"}) || !slices.Equal(gap.Outputs, []string{"dist"}) {
		t.Fatalf("%+v", gap)
	}
	if !strings.Contains(gap.Problem(), "writes packages/auth/dist but its task caches dist/packages/auth") || !strings.Contains(gap.Remedy(), `outputs = ["dist"] under [projects.auth.tasks.build]`) {
		t.Errorf("%q %q", gap.Problem(), gap.Remedy())
	}
	if _, ok := GapOf("auth", "packages/auth", script, []string{"dist"}, false); ok {
		t.Error("dist covers dist/src and dist")
	}
	kept, _ := GapOf("auth", "packages/auth", script, []string{"coverage"}, true)
	if !slices.Equal(kept.Outputs, []string{"dist", "coverage"}) {
		t.Errorf("a hand-declared output stays: %v", kept.Outputs)
	}
}

func TestDeclareOutputs(t *testing.T) {
	src := "[projects.auth]\ntag_prefix = \"a@\"\n"
	got := DeclareOutputs(src, []Gap{{Project: "auth", Outputs: []string{"dist"}}, {Project: "ms.sql", Outputs: []string{"dist"}}})
	want := src + "\n[projects.auth.tasks.build]\noutputs = [\"dist\"]\n\n[projects.\"ms.sql\".tasks.build]\noutputs = [\"dist\"]\n"
	if got != want {
		t.Errorf("%q", got)
	}
}
