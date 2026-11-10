package discovery

import "testing"

func TestAnEnvironmentBuildTaskWritesItsOwnFolder(t *testing.T) {
	scripts := map[string]string{
		"build":     "vite build --mode prd --outDir ../../dist/apps/web/prd",
		"build:uat": "vite build --mode uat --outDir ../../dist/apps/web/uat",
		"build:lcl": "vite build --mode lcl --outDir dist/lcl",
		"start":     "vite --mode lcl",
	}
	for script, want := range map[string]string{"build": "../../dist/apps/web/prd", "build:uat": "../../dist/apps/web/uat", "build:lcl": "dist/lcl"} {
		task := TaskFor(script, "apps/web", scripts)
		if len(task.Outputs) != 1 || task.Outputs[0] != want || !task.Cache {
			t.Errorf("%s: outputs %v cache %v, want %s cached", script, task.Outputs, task.Cache, want)
		}
		if len(task.Inputs) == 0 {
			t.Errorf("%s: an environment build is cached on the project's files, the env file among them", script)
		}
	}
}

func TestAPlainBuildKeepsTheDefaultFolder(t *testing.T) {
	task := TaskFor("build", "apps/web", map[string]string{"build": "vite build"})
	if len(task.Outputs) != 1 || task.Outputs[0] != "../../dist/apps/web" {
		t.Errorf("got %v", task.Outputs)
	}
	other := TaskFor("build:docs", "apps/web", map[string]string{"build:docs": "typedoc"})
	if len(other.Outputs) != 0 || other.Cache {
		t.Errorf("a script that is not an environment build stays an uncached task: %+v", other)
	}
}
