package functionhost

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCompiledEntry(t *testing.T) {
	cases := []struct {
		name     string
		files    []string
		main     string
		wantFail string
	}{
		{name: "main file present", files: []string{"dist/index.js"}, main: "dist/index.js"},
		{name: "main with ./ prefix", files: []string{"dist/index.js"}, main: "./dist/index.js"},
		{name: "main glob of the v4 model", files: []string{"dist/src/functions/a.js"}, main: "dist/src/functions/*.js"},
		{name: "main missing", files: []string{"host.json"}, main: "dist/index.js", wantFail: "dist/index.js"},
		{name: "main glob matches nothing", files: []string{"dist/other.txt"}, main: "dist/src/functions/*.js", wantFail: "no compiled code"},
		{name: "no main, compiled code in dist", files: []string{"dist/functions/a.js"}},
		{name: "no main, no dist", files: []string{"host.json", "package.json"}, wantFail: "no .js file"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stage := t.TempDir()
			for _, f := range c.files {
				write(t, filepath.Join(stage, filepath.FromSlash(f)), "x")
			}
			manifest := map[string]any{"name": "api"}
			if c.main != "" {
				manifest["main"] = c.main
			}
			err := CheckCompiledEntry(stage, manifest, "dist/apps/api")
			if c.wantFail == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantFail != "" && (err == nil || !strings.Contains(err.Error(), c.wantFail)) {
				t.Fatalf("want an error with %q, got %v", c.wantFail, err)
			}
		})
	}
}

func TestPackFunctionAppFailsWithoutCompiledCode(t *testing.T) {
	project, build := t.TempDir(), filepath.Join(t.TempDir(), "out")
	write(t, filepath.Join(build, "package.json"), `{"name":"api","main":"dist/src/functions/*.js"}`)
	write(t, filepath.Join(build, "host.json"), `{}`)
	tools := Tools{Run: func(string, []string) error { return nil }}
	if _, err := PackFunctionApp(PackOptions{ProjectDir: project, BuildDir: build, Name: "api", Tools: tools}); err == nil || !strings.Contains(err.Error(), "no compiled code") {
		t.Fatalf("a build without compiled code fails before the zip: %v", err)
	}
}
