package buildoutputs

import (
	"slices"
	"testing"
)

func TestRunScript(t *testing.T) {
	cases := []struct {
		command     []string
		script, pkg string
		ok          bool
	}{
		{[]string{"npm", "run", "build"}, "build", "", true},
		{[]string{"npm", "run", "build:local", "-w", "app"}, "build:local", "app", true},
		{[]string{"npm", "run", "build", "--workspace=api", "--silent"}, "build", "api", true},
		{[]string{"npm", "run", "build", "--", "--outDir", "x"}, "", "", false},
		{[]string{"npm", "run", "build", "-w"}, "", "", false},
		{[]string{"vite", "build"}, "", "", false},
	}
	for _, c := range cases {
		script, pkg, ok := RunScript(c.command)
		if script != c.script || pkg != c.pkg || ok != c.ok {
			t.Errorf("%v: %q %q %v, want %q %q %v", c.command, script, pkg, ok, c.script, c.pkg, c.ok)
		}
	}
}

func TestExpandScript(t *testing.T) {
	scripts := map[string]string{
		"build:local": "npm run typegen && vite build",
		"typegen":     "chakra typegen src/theme.ts",
		"loop":        "npm run loop",
	}
	if got, want := ExpandScript("npm run build:local", scripts), "chakra typegen src/theme.ts\nvite build"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := ExpandScript("npm run loop", scripts); got != "npm run loop" {
		t.Errorf("a script that runs itself: got %q", got)
	}
	if got := ExpandScript("npm run missing -- -x", scripts); got != "npm run missing -- -x" {
		t.Errorf("an unknown step: got %q", got)
	}
}

func TestScriptWrites(t *testing.T) {
	cases := []struct {
		script   string
		viteOut  bool
		dirs     []string
		complete bool
	}{
		{"chakra typegen src/theme.ts\nvite build", false, []string{"dist"}, true},
		{"vite build", true, nil, false},
		{"vite build --outDir ../../dist/apps/app", false, []string{"../../dist/apps/app"}, true},
		{"tsc -p tsconfig.app.json\ntsc-alias -p tsconfig.app.json", false, nil, false},
		{"tsc -p tsconfig.app.json --outDir ../../dist/apps/api", false, []string{"../../dist/apps/api"}, true},
		{"tsc --noEmit && esbuild a.ts --outdir=lib", false, []string{"lib"}, true},
		{"rollup -c", false, nil, false},
	}
	for _, c := range cases {
		dirs, complete := ScriptWrites(c.script, c.viteOut)
		if !slices.Equal(dirs, c.dirs) || complete != c.complete {
			t.Errorf("%q: %v %v, want %v %v", c.script, dirs, complete, c.dirs, c.complete)
		}
	}
}

func TestSuggested(t *testing.T) {
	cases := []struct {
		written, declared, want []string
	}{
		{[]string{"dist"}, []string{"dist", "dist-dev", "dist-preprod", "dist-prod"}, []string{"dist"}},
		{[]string{"dist/esm", "dist/types"}, []string{"dist"}, []string{"dist"}},
		{[]string{"../../dist/apps/api"}, []string{"dist"}, []string{"../../dist/apps/api"}},
	}
	for _, c := range cases {
		if got := Suggested(c.written, c.declared); !slices.Equal(got, c.want) {
			t.Errorf("Suggested(%v, %v) = %v, want %v", c.written, c.declared, got, c.want)
		}
	}
}
