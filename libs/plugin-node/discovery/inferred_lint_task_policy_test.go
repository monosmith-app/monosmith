package discovery

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestInferredLintTask(t *testing.T) {
	cases := []struct {
		name     string
		manifest string
		files    []string
		want     bool
	}{
		{"legacy eslintrc and eslint in devDependencies", `{"devDependencies":{"eslint":"^8"}}`, []string{".eslintrc.js"}, true},
		{"flat config", `{"devDependencies":{"eslint":"^9"}}`, []string{"eslint.config.mjs"}, true},
		{"eslintConfig in package.json", `{"devDependencies":{"eslint":"^8"},"eslintConfig":{"extends":"standard"}}`, nil, true},
		{"own lint script wins", `{"scripts":{"lint":"eslint src"},"devDependencies":{"eslint":"^8"}}`, []string{".eslintrc.json"}, false},
		{"config without eslint", `{}`, []string{".eslintrc.json"}, false},
		{"eslint without config", `{"devDependencies":{"eslint":"^8"}}`, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range c.files {
				_ = os.WriteFile(filepath.Join(dir, f), []byte("{}"), 0o644)
			}
			_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(c.manifest), 0o644)
			res, err := Discover(filepath.Dir(dir), []string{filepath.Base(dir) + "/package.json"})
			if err != nil {
				t.Fatal(err)
			}
			task, ok := res.Projects[0].Tasks["lint"]
			if c.want && (!ok || !slices.Equal(task.Command, []string{"npx", "eslint", "."}) || task.Fix == nil || !task.Cache) {
				t.Fatalf("want the inferred lint task, got %v %+v", ok, task)
			}
			if !c.want && ok && !slices.Contains(task.Command, "run") {
				t.Fatalf("want no inferred lint task, got %+v", task)
			}
		})
	}
}
