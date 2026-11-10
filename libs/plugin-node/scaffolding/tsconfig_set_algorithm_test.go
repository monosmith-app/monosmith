package scaffolding

import "testing"

func TestTsconfigSetFollowsTheSiblings(t *testing.T) {
	for _, c := range []struct {
		files, build string
		spec         bool
	}{
		{"", TsconfigBuild, false},
		{"none", TsconfigBuild, false},
		{"tsconfig.build.json,tsconfig.json", TsconfigBuild, false},
		{"tsconfig.build.json,tsconfig.json,tsconfig.spec.json", TsconfigBuild, true},
		{"tsconfig.json,tsconfig.lib.json,tsconfig.spec.json", TsconfigLib, true},
		{"tsconfig.json,tsconfig.lib.json", TsconfigLib, false},
	} {
		build, spec := TsconfigSet(map[string]string{"tsconfig.files": c.files})
		if build != c.build || spec != c.spec {
			t.Errorf("%q: got %s %v, want %s %v", c.files, build, spec, c.build, c.spec)
		}
	}
}
