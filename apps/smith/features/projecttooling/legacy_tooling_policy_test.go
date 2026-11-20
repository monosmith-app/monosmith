package projecttooling

import (
	"strings"
	"testing"
)

func TestClassifyBuildNamesTheKindOfEachBuild(t *testing.T) {
	cases := []struct {
		f    Facts
		want BuildKind
	}{
		{Facts{}, BuildNone},
		{Facts{BuildScript: "react-scripts build"}, BuildCRA},
		{Facts{BuildScript: "vite build"}, BuildSmith},
		{Facts{BuildScript: "tsc", OutDir: "./lib"}, BuildTsc},
		{Facts{BuildScript: "tsc -p tsconfig.json", OutDir: "dist"}, BuildSmith},
		{Facts{BuildScript: "tsc && cp a b", OutDir: "lib"}, BuildOther},
		{Facts{BuildScript: "rollup -c"}, BuildOther},
	}
	for _, c := range cases {
		if got := ClassifyBuild(c.f); got != c.want {
			t.Errorf("%q into %q: %s, want %s", c.f.BuildScript, c.f.OutDir, got, c.want)
		}
	}
}

func TestLegacyReasonsNameESLint8CreateReactAppTscIntoLibAndBundlers(t *testing.T) {
	got := LegacyReasons(Facts{LegacyESLintFiles: []string{".eslintrc.js"}, BuildScript: "tsc", OutDir: "./lib"})
	if len(got) != 2 || !strings.Contains(got[0], ".eslintrc.js") || got[1] != "a tsc build into lib/" {
		t.Errorf("reasons %q", got)
	}
	got = LegacyReasons(Facts{BuildScript: "react-scripts build", Dependencies: map[string]bool{"react-scripts": true}, ESLintInManifest: true})
	if len(got) != 2 || !strings.Contains(got[0], "create-react-app") || !strings.Contains(got[1], "eslintConfig") {
		t.Errorf("reasons %q", got)
	}
	got = LegacyReasons(Facts{ESLintRange: "^7.32.0", BuildScript: "rollup -c"})
	if len(got) != 2 || got[0] != "ESLint 7" || got[1] != "a rollup build" {
		t.Errorf("reasons %q", got)
	}
}

func TestCurrentToolsAreNotLegacyAndSmithRecommendsKeepingThem(t *testing.T) {
	f := Facts{ESLintRange: "^10.8.1", BuildScript: "tsc -p tsconfig.build.json --outDir dist"}
	if got := LegacyReasons(f); len(got) != 0 {
		t.Errorf("reasons %q", got)
	}
	if Recommend(LegacyReasons(f)) != Keep {
		t.Error("smith recommends keeping current tools")
	}
	if Recommend([]string{"ESLint 8"}) != Smith {
		t.Error("smith recommends its tooling over a legacy tool")
	}
}
