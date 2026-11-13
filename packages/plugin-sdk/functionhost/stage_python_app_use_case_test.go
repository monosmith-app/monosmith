package functionhost

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func pythonProject(t *testing.T) string {
	t.Helper()
	project := t.TempDir()
	for _, f := range []string{"function_app.py", "host.json", "requirements.txt", "src/acme_api/__init__.py", "src/acme_api/__pycache__/x.pyc", "src/acme_api/greet.py"} {
		write(t, filepath.Join(project, filepath.FromSlash(f)), "x")
	}
	return project
}

func TestStagePythonAppCopiesTheDeployedFilesWithoutJunk(t *testing.T) {
	project, build := pythonProject(t), filepath.Join(t.TempDir(), "out")
	tools := Tools{Output: func(string, []string) (string, error) { return "{}\n", nil }}
	stage, err := StagePythonApp(StageOptions{ProjectDir: project, BuildDir: build, Host: Azure, Python: "py", Tools: tools})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"function_app.py", "host.json", "requirements.txt", "src/acme_api/greet.py"} {
		if _, err := os.Stat(filepath.Join(stage, filepath.FromSlash(want))); err != nil {
			t.Errorf("%s is deployed: %v", want, err)
		}
	}
	if _, err := os.Stat(filepath.Join(stage, "src", "acme_api", "__pycache__")); err == nil {
		t.Error("a cache folder is not deployed")
	}
}

func TestStagePythonAppFlattensSrcForLambdaAndShipsTheLibrariesOfTheWorkspace(t *testing.T) {
	project, build := pythonProject(t), filepath.Join(t.TempDir(), "out")
	lib := t.TempDir()
	write(t, filepath.Join(lib, "src", "core_py", "__init__.py"), "x")
	var asked []string
	tools := Tools{Output: func(dir string, argv []string) (string, error) {
		asked = argv
		out, _ := json.Marshal(map[string]string{"core-py": lib})
		return string(out), nil
	}}
	stage, err := StagePythonApp(StageOptions{ProjectDir: project, BuildDir: build, Host: AWS, Python: "py", Tools: tools})
	if err != nil {
		t.Fatal(err)
	}
	if asked[0] != "py" || asked[1] != "-c" {
		t.Errorf("the venv's python is asked: %v", asked)
	}
	for _, want := range []string{"acme_api/greet.py", "core_py/__init__.py"} {
		if _, err := os.Stat(filepath.Join(stage, filepath.FromSlash(want))); err != nil {
			t.Errorf("%s sits at the root of the zip: %v", want, err)
		}
	}
	if _, err := os.Stat(filepath.Join(stage, "function_app.py")); err == nil {
		t.Error("Lambda deploys only src/")
	}
}

func TestStagePythonAppRefusesALibraryWithoutAPackage(t *testing.T) {
	project, build := pythonProject(t), filepath.Join(t.TempDir(), "out")
	tools := Tools{Output: func(string, []string) (string, error) {
		return `{"empty-lib": "` + filepath.ToSlash(t.TempDir()) + `"}`, nil
	}}
	_, err := StagePythonApp(StageOptions{ProjectDir: project, BuildDir: build, Host: Azure, Python: "py", Tools: tools})
	if err == nil || !strings.Contains(err.Error(), "has no package to ship") {
		t.Fatalf("%v", err)
	}
}

func TestZipPythonAppZipsTheStage(t *testing.T) {
	build := t.TempDir()
	write(t, filepath.Join(build, "app", "function_app.py"), "x")
	write(t, filepath.Join(build, "app", "src", "a.py"), "x")
	zipPath, err := ZipPythonApp(build, "api")
	if err != nil {
		t.Fatal(err)
	}
	if got := zipNames(t, zipPath); !slices.Equal(got, []string{"function_app.py", "src/a.py"}) {
		t.Errorf("zip holds %v", got)
	}
	if _, err := ZipPythonApp(t.TempDir(), "api"); err == nil {
		t.Error("nothing staged is an error that names the build task")
	}
}
